// Package configview renders read-only configuration documents for the
// View config surface (RC-004F). Every document passes through Redact before
// it is rendered, so no secret value leaves Core through this surface.
package configview

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Marker replaces a redacted value when no *_env reference names its source.
const Marker = "<redacted>"

// envNamePattern is the shape of an environment-variable reference name.
var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// sensitiveSegments are key segments that mark a secret wherever they appear
// in a snake-cased key (for example db_password, client_secret).
var sensitiveSegments = map[string]struct{}{
	"password": {}, "passwd": {}, "secret": {}, "secrets": {}, "credential": {},
	"credentials": {}, "authorization": {}, "bearer": {}, "apikey": {}, "privatekey": {},
	"passphrase": {}, "cookie": {}, "cookies": {},
}

// sensitiveSuffixes match the compacted key (lowercase, separators removed),
// which also covers camelCase and SHOUTING forms such as OPENAIAPIKEY.
var sensitiveSuffixes = []string{
	"apikey", "privatekey", "clientsecret", "password", "passwd", "secret",
	"credentials", "credential", "authorization", "bearer", "token", "passphrase",
}

// referenceSegments end a key that names where a secret lives rather than
// holding it (api_key_env, secret_ref). Their values stay visible unless the
// value itself looks like a secret.
var referenceSegments = map[string]struct{}{"env": {}, "ref": {}}

// Result is a redacted document plus the paths that were redacted.
type Result struct {
	Document   any
	Redactions []string
}

// Redact returns a deep copy of doc with secret values replaced. Typed values
// (structs, typed maps) are normalized through YAML first so nothing is left
// unwalked. Output maps are always map[string]any so JSON encoding works.
func Redact(doc any) (Result, error) {
	r := &redactor{}
	out, err := r.walk(doc, "")
	if err != nil {
		return Result{}, err
	}
	sort.Strings(r.paths)
	if r.paths == nil {
		r.paths = []string{}
	}
	return Result{Document: out, Redactions: r.paths}, nil
}

type redactor struct{ paths []string }

func (r *redactor) mark(path string) { r.paths = append(r.paths, path) }

func (r *redactor) walk(v any, path string) (any, error) {
	switch t := v.(type) {
	case nil, bool, int, int32, int64, uint, uint32, uint64, float32, float64, time.Time:
		return t, nil
	case string:
		return r.scrubString(t, path), nil
	case map[string]any:
		return r.walkMap(t, path)
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[fmt.Sprint(k)] = val
		}
		return r.walkMap(m, path)
	case []any:
		return r.walkList(t, path)
	case []string:
		list := make([]any, len(t))
		for i, s := range t {
			list[i] = s
		}
		return r.walkList(list, path)
	default:
		generic, err := normalize(t)
		if err != nil {
			return nil, fmt.Errorf("normalize %s: %w", displayPath(path), err)
		}
		return r.walk(generic, path)
	}
}

func (r *redactor) walkMap(m map[string]any, path string) (any, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make(map[string]any, len(m))
	for _, k := range keys {
		child := joinPath(path, k)
		val := m[k]
		switch {
		case IsReferenceKey(k):
			out[k] = r.scrubReference(k, val, child)
		case IsSensitiveKey(k) && hasSecretMaterial(val):
			out[k] = replacement(m, k)
			r.mark(child)
		default:
			cleaned, err := r.walk(val, child)
			if err != nil {
				return nil, err
			}
			out[k] = cleaned
		}
	}
	return out, nil
}

func (r *redactor) walkList(list []any, path string) (any, error) {
	out := make([]any, len(list))
	for i := 0; i < len(list); i++ {
		child := path + "[" + strconv.Itoa(i) + "]"
		// A sensitive CLI flag ("--api-key", "--token") hides the next argument.
		if i > 0 {
			if flag, ok := list[i-1].(string); ok && isSensitiveFlag(flag) {
				if s, ok := list[i].(string); ok && s != "" {
					out[i] = Marker
					r.mark(child)
					continue
				}
			}
		}
		cleaned, err := r.walk(list[i], child)
		if err != nil {
			return nil, err
		}
		out[i] = cleaned
	}
	return out, nil
}

// scrubReference keeps an env/ref name visible. A value that is not a plain
// reference (a nested block, or a string that looks like a secret or, for
// *_env keys, is not an env-var name) is redacted.
func (r *redactor) scrubReference(key string, val any, path string) any {
	s, ok := val.(string)
	if !ok {
		if val == nil {
			return nil
		}
		if _, scalar := val.(bool); scalar {
			return val
		}
		r.mark(path)
		return Marker
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	envKey := lastSegment(key) == "env"
	if LooksSecret(s) || (envKey && !envNamePattern.MatchString(s)) {
		r.mark(path)
		return Marker
	}
	return s
}

// scrubString redacts secret-looking tokens inside a free string value.
func (r *redactor) scrubString(s, path string) string {
	if s == "" {
		return s
	}
	if strings.Contains(s, "-----BEGIN") && strings.Contains(strings.ToUpper(s), "PRIVATE KEY") {
		r.mark(path)
		return Marker
	}
	if isBearerValue(s) {
		r.mark(path)
		return Marker
	}
	fields := strings.Fields(s)
	replaced := make([]string, 0, len(fields))
	changed := false
	for i, f := range fields {
		// "... Bearer <token>" inside a longer header or command string.
		if i > 0 && strings.EqualFold(strings.Trim(fields[i-1], `"':`), "bearer") && looksLikeHeaderToken(f) {
			replaced = append(replaced, Marker)
			changed = true
			continue
		}
		if repl, ok := scrubField(f); ok {
			replaced = append(replaced, repl)
			changed = true
			continue
		}
		replaced = append(replaced, f)
	}
	if !changed {
		return s
	}
	r.mark(path)
	return strings.Join(replaced, " ")
}

// replacement is ${env:NAME} when a sibling <key>_env names the source,
// otherwise the fixed marker.
func replacement(m map[string]any, key string) string {
	for _, suffix := range []string{"_env", "_env_var", "Env"} {
		if ref, ok := m[key+suffix].(string); ok {
			ref = strings.TrimSpace(ref)
			if envNamePattern.MatchString(ref) {
				return "${env:" + ref + "}"
			}
		}
	}
	return Marker
}

// hasSecretMaterial reports whether a value under a sensitive key must be
// hidden. Empty strings, nil, and booleans (flags such as `secret: true`)
// carry no secret material; everything else does, including numbers, maps,
// and lists.
func hasSecretMaterial(v any) bool {
	switch t := v.(type) {
	case nil, bool:
		return false
	case string:
		return strings.TrimSpace(t) != ""
	default:
		return true
	}
}

func normalize(v any) (any, error) {
	raw, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func joinPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func displayPath(path string) string {
	if path == "" {
		return "document"
	}
	return path
}

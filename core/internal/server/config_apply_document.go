package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/configview"
	"gopkg.in/yaml.v3"
)

// maxConfigDocumentBytes bounds a submitted config document (RC-004G).
const maxConfigDocumentBytes = 1 << 20

// envPlaceholderPrefix is how View config renders a secret that has an env
// reference. A document carrying it in a secret field was copied from a
// redacted view and must not be written back.
const envPlaceholderPrefix = "${env:"

// parseConfigDocument parses a YAML or JSON document into a JSON-normalized
// object, so the same document in either format has the same hash.
func parseConfigDocument(text, format string) (map[string]any, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("document is required")
	}
	parsedFormat, err := configview.ParseFormat(format)
	if err != nil {
		return nil, err
	}
	var raw any
	if parsedFormat == configview.FormatJSON {
		err = json.Unmarshal([]byte(text), &raw)
	} else {
		err = yaml.Unmarshal([]byte(text), &raw)
	}
	if err != nil {
		return nil, fmt.Errorf("document is not valid %s: %v", parsedFormat, err)
	}
	return normalizeConfigObject(raw)
}

// normalizeConfigObject round-trips v through JSON so maps, numbers, and
// times have one representation. The result must be an object.
func normalizeConfigObject(v any) (map[string]any, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("document must be an object with string keys: %v", err)
	}
	var out any
	if err := json.Unmarshal(encoded, &out); err != nil {
		return nil, err
	}
	obj, ok := out.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document must be an object")
	}
	return obj, nil
}

// canonicalConfigJSON returns the canonical JSON (sorted keys) of doc and its
// SHA-256. Confirm tokens bind to this hash.
func canonicalConfigJSON(doc map[string]any) ([]byte, string, error) {
	encoded, err := json.Marshal(doc)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(encoded)
	return encoded, hex.EncodeToString(sum[:]), nil
}

// configWriteBackViolations lists values that must never be written: the
// View config redaction marker, an ${env:...} placeholder in a secret field,
// and literal secret material (secrets belong in *_env or *_ref references).
func configWriteBackViolations(doc map[string]any) []string {
	var out []string
	walkWriteBack(doc, "", false, &out)
	if redacted, err := configview.Redact(doc); err == nil {
		for _, path := range redacted.Redactions {
			out = append(out, path+": literal secret value (use a *_env or *_ref reference)")
		}
	}
	sort.Strings(out)
	return dedupeStrings(out)
}

func walkWriteBack(v any, path string, secret bool, out *[]string) {
	switch t := v.(type) {
	case string:
		if strings.Contains(t, configview.Marker) {
			*out = append(*out, displayConfigPath(path)+": redaction marker "+configview.Marker)
		} else if secret && strings.Contains(t, envPlaceholderPrefix) {
			*out = append(*out, displayConfigPath(path)+": ${env:...} placeholder in a secret field")
		}
	case map[string]any:
		for key, child := range t {
			childSecret := secret || (configview.IsSensitiveKey(key) && !configview.IsReferenceKey(key))
			walkWriteBack(child, joinConfigPath(path, key), childSecret, out)
		}
	case []any:
		for i, child := range t {
			walkWriteBack(child, path+"["+strconv.Itoa(i)+"]", secret, out)
		}
	}
}

// configDiff is the structured dry-run diff: dotted paths only, never values,
// so a preview cannot leak what it compares.
type configDiff struct {
	Changed []string `json:"changed"`
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
}

func diffConfigDocuments(before, after any) configDiff {
	left, right := map[string]string{}, map[string]string{}
	flattenConfig(before, "", left)
	flattenConfig(after, "", right)
	diff := configDiff{Changed: []string{}, Added: []string{}, Removed: []string{}}
	for path, value := range right {
		old, ok := left[path]
		switch {
		case !ok:
			diff.Added = append(diff.Added, path)
		case old != value:
			diff.Changed = append(diff.Changed, path)
		}
	}
	for path := range left {
		if _, ok := right[path]; !ok {
			diff.Removed = append(diff.Removed, path)
		}
	}
	sort.Strings(diff.Changed)
	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)
	return diff
}

func flattenConfig(v any, path string, out map[string]string) {
	switch t := v.(type) {
	case map[string]any:
		if len(t) == 0 && path != "" {
			out[path] = "{}"
		}
		for key, child := range t {
			flattenConfig(child, joinConfigPath(path, key), out)
		}
	case []any:
		if len(t) == 0 {
			out[displayConfigPath(path)] = "[]"
		}
		for i, child := range t {
			flattenConfig(child, path+"["+strconv.Itoa(i)+"]", out)
		}
	default:
		encoded, _ := json.Marshal(t)
		out[displayConfigPath(path)] = string(encoded)
	}
}

func joinConfigPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func displayConfigPath(path string) string {
	if path == "" {
		return "(root)"
	}
	return path
}

func dedupeStrings(items []string) []string {
	out := items[:0]
	for i, item := range items {
		if i == 0 || item != items[i-1] {
			out = append(out, item)
		}
	}
	return out
}

package protocol

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigDocumentRedactedValue replaces any value the shared secret classifier
// refuses to show. It matches the MCP config redaction marker.
const ConfigDocumentRedactedValue = "[redacted]"

type ConfigDocumentExportFormat string

const (
	ConfigDocumentExportYAML ConfigDocumentExportFormat = "yaml"
	ConfigDocumentExportJSON ConfigDocumentExportFormat = "json"
)

// ConfigDocumentExport is a read-only, deterministic view of one stored
// revision. When RedactionApplied is true the content no longer hashes to the
// stored digest and must not be treated as a re-importable copy.
type ConfigDocumentExport struct {
	DocumentID       string                     `json:"document_id"`
	Kind             string                     `json:"kind"`
	Version          string                     `json:"version"`
	Scope            ConfigDocumentScope        `json:"scope"`
	Format           ConfigDocumentExportFormat `json:"format"`
	Content          string                     `json:"content"`
	RedactionApplied bool                       `json:"redaction_applied"`
	RedactedPaths    []string                   `json:"redacted_paths"`
	ExportSHA256     string                     `json:"export_sha256"`
}

// ParseConfigDocumentExportFormat accepts only the two supported formats.
// An empty value selects YAML.
func ParseConfigDocumentExportFormat(raw string) (ConfigDocumentExportFormat, error) {
	switch ConfigDocumentExportFormat(raw) {
	case "", ConfigDocumentExportYAML:
		return ConfigDocumentExportYAML, nil
	case ConfigDocumentExportJSON:
		return ConfigDocumentExportJSON, nil
	}
	return "", fmt.Errorf("unsupported config document export format")
}

// RedactConfigDocumentForExport returns the full envelope as a generic tree
// with every classifier-flagged value replaced, plus the redacted paths in
// deterministic order. It applies the strict (new-write) rules of the single
// classifier so rows admitted under older rules are still safe to display.
func RedactConfigDocumentForExport(document ConfigDocument) (map[string]any, []string, error) {
	metadataJSON, err := json.Marshal(document.Metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal config document metadata: %w", err)
	}
	metadata, err := decodeConfigDocumentExportJSON(metadataJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("decode config document metadata: %w", err)
	}
	spec, err := decodeConfigDocumentExportJSON(document.Spec)
	if err != nil {
		return nil, nil, fmt.Errorf("decode config document spec: %w", err)
	}
	envelope := map[string]any{
		"apiVersion": document.APIVersion,
		"kind":       string(document.Kind),
		"metadata":   metadata,
		"spec":       spec,
	}
	paths := make([]string, 0)
	redacted, _ := redactConfigDocumentExportValue(envelope, "", &paths).(map[string]any)
	return redacted, paths, nil
}

// RenderConfigDocumentExport renders the redacted envelope with sorted keys.
// The same document always yields byte-identical content.
func RenderConfigDocumentExport(document ConfigDocument, format ConfigDocumentExportFormat) (ConfigDocumentExport, error) {
	if _, err := ParseConfigDocumentExportFormat(string(format)); err != nil || format == "" {
		return ConfigDocumentExport{}, fmt.Errorf("unsupported config document export format")
	}
	envelope, paths, err := RedactConfigDocumentForExport(document)
	if err != nil {
		return ConfigDocumentExport{}, err
	}
	var content []byte
	if format == ConfigDocumentExportJSON {
		content, err = renderConfigDocumentExportJSON(envelope)
	} else {
		content, err = renderConfigDocumentExportYAML(envelope)
	}
	if err != nil {
		return ConfigDocumentExport{}, err
	}
	sum := sha256.Sum256(content)
	metadata, _ := envelope["metadata"].(map[string]any)
	scope, _ := metadata["scope"].(map[string]any)
	return ConfigDocumentExport{
		DocumentID: exportString(metadata, "id"),
		Kind:       exportString(envelope, "kind"),
		Version:    exportString(metadata, "version"),
		Scope: ConfigDocumentScope{
			Kind: ConfigDocumentScopeKind(exportString(scope, "kind")),
			Ref:  exportString(scope, "ref"),
		},
		Format:           format,
		Content:          string(content),
		RedactionApplied: len(paths) > 0,
		RedactedPaths:    paths,
		ExportSHA256:     "sha256:" + hex.EncodeToString(sum[:]),
	}, nil
}

func redactConfigDocumentExportValue(value any, path string, paths *[]string) any {
	redact := func(at string) any {
		*paths = append(*paths, at)
		return ConfigDocumentRedactedValue
	}
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(typed))
		for _, key := range keys {
			child := typed[key]
			outKey := key
			if looksLikeRawConfigDocumentSecret(key, configDocumentSecretsStrict) {
				outKey = uniqueRedactedExportKey(out)
				out[outKey] = redact(joinExportPath(path, outKey))
				continue
			}
			childPath := joinExportPath(path, key)
			_, isText := child.(string)
			switch {
			case configDocumentSecretRefField(key, configDocumentSecretsStrict):
				out[outKey] = redactConfigDocumentExportRef(child, childPath, redact)
			case configDocumentSensitiveField(key), isText && configDocumentStrictSensitiveField(key):
				if keepConfigDocumentSensitiveExportValue(child) {
					out[outKey] = child
				} else {
					out[outKey] = redact(childPath)
				}
			default:
				out[outKey] = redactConfigDocumentExportValue(child, childPath, paths)
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, child := range typed {
			out[index] = redactConfigDocumentExportValue(child, fmt.Sprintf("%s[%d]", path, index), paths)
		}
		return out
	case string:
		if looksLikeRawConfigDocumentSecret(typed, configDocumentSecretsStrict) {
			return redact(path)
		}
	}
	return value
}

// A sensitive field may show a managed reference (a name, not a value) or an
// empty/null value. Everything else is hidden.
func keepConfigDocumentSensitiveExportValue(value any) bool {
	if value == nil {
		return true
	}
	text, ok := value.(string)
	return ok && (text == "" || isConfigDocumentSecretRef(text))
}

func redactConfigDocumentExportRef(value any, path string, redact func(string) any) any {
	switch typed := value.(type) {
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			if keepConfigDocumentSensitiveExportValue(item) {
				out[index] = item
			} else {
				out[index] = redact(fmt.Sprintf("%s[%d]", path, index))
			}
		}
		return out
	default:
		if keepConfigDocumentSensitiveExportValue(typed) {
			return typed
		}
		return redact(path)
	}
}

func uniqueRedactedExportKey(existing map[string]any) string {
	candidate := ConfigDocumentRedactedValue
	for suffix := 2; ; suffix++ {
		if _, taken := existing[candidate]; !taken {
			return candidate
		}
		candidate = fmt.Sprintf("[redacted-%d]", suffix)
	}
}

func joinExportPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

func exportString(values map[string]any, key string) string {
	text, _ := values[key].(string)
	return text
}

func decodeConfigDocumentExportJSON(raw []byte) (any, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, err := decodeConfigDocumentJSONValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("config document export expects exactly one JSON value")
	}
	return value, nil
}

func renderConfigDocumentExportJSON(envelope map[string]any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(envelope); err != nil {
		return nil, fmt.Errorf("render config document export JSON: %w", err)
	}
	return buffer.Bytes(), nil
}

func renderConfigDocumentExportYAML(envelope map[string]any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(configDocumentExportYAMLNode(envelope)); err != nil {
		return nil, fmt.Errorf("render config document export YAML: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("render config document export YAML: %w", err)
	}
	return buffer.Bytes(), nil
}

// configDocumentExportYAMLNode builds an explicit node tree so key order is
// sorted and JSON numbers keep their exact text instead of becoming strings.
func configDocumentExportYAMLNode(value any) *yaml.Node {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, key := range keys {
			node.Content = append(node.Content, yamlScalar("!!str", key), configDocumentExportYAMLNode(typed[key]))
		}
		return node
	case []any:
		node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, item := range typed {
			node.Content = append(node.Content, configDocumentExportYAMLNode(item))
		}
		return node
	case string:
		return yamlScalar("!!str", typed)
	case json.Number:
		if strings.ContainsAny(typed.String(), ".eE") {
			return yamlScalar("!!float", typed.String())
		}
		return yamlScalar("!!int", typed.String())
	case bool:
		if typed {
			return yamlScalar("!!bool", "true")
		}
		return yamlScalar("!!bool", "false")
	default:
		return yamlScalar("!!null", "null")
	}
}

func yamlScalar(tag, value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
}

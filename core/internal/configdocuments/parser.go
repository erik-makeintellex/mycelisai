package configdocuments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
	"gopkg.in/yaml.v3"
)

// ParseDocument gives direct JSON/YAML authoring the same strict envelope as
// Soma-authored configuration. Family-specific validation remains in protocol.
func ParseDocument(raw []byte, format string) (protocol.ConfigDocument, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = detectFormat(raw)
	}
	var canonical []byte
	var err error
	switch format {
	case "json":
		canonical = raw
	case "yaml", "yml":
		canonical, err = yamlToJSON(raw)
		if err != nil {
			return protocol.ConfigDocument{}, err
		}
	default:
		return protocol.ConfigDocument{}, fmt.Errorf("unsupported config document format %q", format)
	}

	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	var document protocol.ConfigDocument
	if err := decoder.Decode(&document); err != nil {
		return protocol.ConfigDocument{}, fmt.Errorf("decode config document: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return protocol.ConfigDocument{}, err
	}
	return document, nil
}

// LoadDocumentFile loads only a file below the configured durable root.
func LoadDocumentFile(root, path string) (protocol.ConfigDocument, error) {
	resolved, err := resolveConfigPath(root, path)
	if err != nil {
		return protocol.ConfigDocument{}, err
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return protocol.ConfigDocument{}, fmt.Errorf("read config document: %w", err)
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(resolved)), ".")
	return ParseDocument(raw, ext)
}

// ConfiguredRoot returns the durable file-authoring root shared by the API and
// Soma tools. By default it stays inside the governed workspace sandbox.
func ConfiguredRoot() string {
	if root := strings.TrimSpace(os.Getenv("MYCELIS_CONFIG_ROOT")); root != "" {
		return root
	}
	workspace := strings.TrimSpace(os.Getenv("MYCELIS_WORKSPACE"))
	if workspace == "" {
		workspace = "./workspace"
	}
	return filepath.Join(workspace, "config")
}

func resolveConfigPath(root, path string) (string, error) {
	root = strings.TrimSpace(root)
	path = strings.TrimSpace(path)
	if root == "" {
		return "", fmt.Errorf("config root is required")
	}
	if path == "" {
		return "", fmt.Errorf("config document path is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve config root: %w", err)
	}
	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(absRoot, candidate)
	}
	absPath, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve config document path: %w", err)
	}
	relative, err := filepath.Rel(absRoot, absPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("config document path must stay within the configured root")
	}
	return absPath, nil
}

func detectFormat(raw []byte) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return "json"
	}
	return "yaml"
}

func yamlToJSON(raw []byte) ([]byte, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode YAML config document: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("config document must contain exactly one YAML document")
		}
		return nil, fmt.Errorf("decode YAML config document: %w", err)
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("convert YAML config document: %w", err)
	}
	return canonical, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("config document must contain exactly one JSON value")
		}
		return fmt.Errorf("decode config document: %w", err)
	}
	return nil
}

package configview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Output formats.
const (
	FormatYAML = "yaml"
	FormatJSON = "json"
)

// ParseFormat normalizes the ?format= query value. Empty means YAML.
func ParseFormat(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "yaml", "yml":
		return FormatYAML, nil
	case "json":
		return FormatJSON, nil
	}
	return "", fmt.Errorf("unsupported format %q (expected yaml or json)", raw)
}

// Rendered is a redacted document rendered as text.
type Rendered struct {
	Format     string
	Document   string
	Redactions []string
}

// Render redacts doc and renders it in format. Rendering never happens on an
// unredacted value.
func Render(doc any, format string) (Rendered, error) {
	res, err := Redact(doc)
	if err != nil {
		return Rendered{}, err
	}
	var text string
	switch format {
	case FormatJSON:
		raw, err := json.MarshalIndent(res.Document, "", "  ")
		if err != nil {
			return Rendered{}, err
		}
		text = string(raw) + "\n"
	case FormatYAML:
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(res.Document); err != nil {
			return Rendered{}, err
		}
		_ = enc.Close()
		text = buf.String()
	default:
		return Rendered{}, fmt.Errorf("unsupported format %q", format)
	}
	return Rendered{Format: format, Document: text, Redactions: res.Redactions}, nil
}

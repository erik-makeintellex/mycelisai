package configdocuments

import (
	"os"
	"path/filepath"
	"testing"
)

const validDocumentYAML = `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: browser-app
  name: Browser app
  version: "1"
  owner_id: operator-1
  scope:
    kind: workspace
    ref: workspace-1
  enabled: true
  source:
    kind: file
    ref: templates/browser-app.yaml
  governance:
    risk_level: low
    approval_posture: required
spec:
  defaults:
    delivery_form: browser_app
`

func TestParseDocumentAcceptsStrictYAMLAndJSON(t *testing.T) {
	for _, test := range []struct {
		name   string
		format string
		raw    string
	}{
		{name: "yaml", format: "yaml", raw: validDocumentYAML},
		{name: "auto yaml", raw: validDocumentYAML},
		{name: "json", format: "json", raw: `{"apiVersion":"mycelis.ai/v1","kind":"OutcomeTemplate","metadata":{"id":"browser-app","name":"Browser app","version":"1","owner_id":"operator-1","scope":{"kind":"workspace","ref":"workspace-1"},"enabled":true,"source":{"kind":"api","ref":"request:test"},"governance":{"risk_level":"low","approval_posture":"required"}},"spec":{"defaults":{"delivery_form":"browser_app"}}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			document, err := ParseDocument([]byte(test.raw), test.format)
			if err != nil {
				t.Fatalf("ParseDocument: %v", err)
			}
			if document.Metadata.ID != "browser-app" || len(document.Spec) == 0 {
				t.Fatalf("unexpected document: %+v", document)
			}
		})
	}
}

func TestParseDocumentRejectsUnknownAndMultipleDocuments(t *testing.T) {
	for _, raw := range []string{
		validDocumentYAML + "unknown: true\n",
		validDocumentYAML + "---\n" + validDocumentYAML,
	} {
		if _, err := ParseDocument([]byte(raw), "yaml"); err == nil {
			t.Fatalf("expected invalid document to fail: %q", raw)
		}
	}
}

func TestLoadDocumentFileStaysInsideConfiguredRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "templates", "browser-app.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(validDocumentYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	document, err := LoadDocumentFile(root, filepath.Join("templates", "browser-app.yaml"))
	if err != nil || document.Metadata.ID != "browser-app" {
		t.Fatalf("LoadDocumentFile = %+v, %v", document, err)
	}
	if _, err := LoadDocumentFile(root, filepath.Join("..", "outside.yaml")); err == nil {
		t.Fatal("expected path traversal to fail")
	}
}

func TestConfiguredRootDefaultsInsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_CONFIG_ROOT", "")
	t.Setenv("MYCELIS_WORKSPACE", workspace)
	want := filepath.Join(workspace, "config")
	if got := ConfiguredRoot(); got != want {
		t.Fatalf("ConfiguredRoot() = %q, want %q", got, want)
	}
}

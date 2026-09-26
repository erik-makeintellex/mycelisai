package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// exportSecretMarker appears in every seeded fake secret so one substring
// check proves no raw value escapes. These are test fixtures, not credentials.
const exportSecretMarker = "EXPORTTESTSECRET"

func seededSecretExportDocument() ConfigDocument {
	return ConfigDocument{
		APIVersion: ConfigDocumentAPIVersionV1,
		Kind:       ConfigDocumentKindWorkerProfile,
		Metadata: ConfigDocumentMetadata{
			ID: "seeded-profile", Name: "Seeded profile", Version: "1", OwnerID: "operator-1",
			Scope:      ConfigDocumentScope{Kind: ConfigDocumentScopeWorkspace, Ref: "primary"},
			Enabled:    true,
			Source:     ConfigDocumentSource{Kind: ConfigDocumentSourceAPI, Ref: "seed"},
			SecretRefs: []string{"env:OPENAI_API_KEY", "sk-EXPORTTESTSECRET0007"},
			Governance: ConfigDocumentGovernance{RiskLevel: ConfigDocumentRiskLow, ApprovalPosture: ApprovalPostureOptional},
		},
		// Written as a raw row that bypassed the validator.
		Spec: json.RawMessage(`{
			"api_key": "sk-live-EXPORTTESTSECRET0001",
			"api_key_ref": "env:OPENAI_API_KEY",
			"connection": {"headers": {"Authorization": "Bearer EXPORTTESTSECRET0002xyz", "Accept": "application/json"}},
			"args": ["--verbose", "ghp_EXPORTTESTSECRET0003abcdef", "--flag=xoxb-EXPORTTESTSECRET0004"],
			"notes": "first line\n-----BEGIN RSA PRIVATE KEY-----\nMIIEXPORTTESTSECRET0005\n-----END RSA PRIVATE KEY-----\n",
			"summary": "rotate weekly\nuse key sk-EXPORTTESTSECRET0006 only in dev",
			"sk-EXPORTTESTSECRET0008": "value",
			"token_refs": ["secret://team/token", "EXPORTTESTSECRET0009 raw"],
			"retries": 3,
			"ratio": 0.25,
			"enabled": true,
			"role": "reviewer"
		}`),
	}
}

func TestConfigDocumentRedactSeededSecretsNeverRender(t *testing.T) {
	document := seededSecretExportDocument()
	for _, format := range []ConfigDocumentExportFormat{ConfigDocumentExportYAML, ConfigDocumentExportJSON} {
		export, err := RenderConfigDocumentExport(document, format)
		if err != nil {
			t.Fatalf("RenderConfigDocumentExport(%s) error = %v", format, err)
		}
		encoded, _ := json.Marshal(export)
		if strings.Contains(string(encoded), exportSecretMarker) {
			t.Fatalf("%s export leaked a seeded secret: %s", format, encoded)
		}
		if !export.RedactionApplied {
			t.Fatalf("%s export redaction_applied = false", format)
		}
		want := []string{
			"metadata.secret_refs[1]",
			"spec.api_key",
			"spec.args[1]",
			"spec.args[2]",
			"spec.connection.headers.Authorization",
			"spec.notes",
			"spec.[redacted]",
			"spec.summary",
			"spec.token_refs[1]",
		}
		if !reflect.DeepEqual(export.RedactedPaths, want) {
			t.Fatalf("%s redacted_paths = %#v, want %#v", format, export.RedactedPaths, want)
		}
		for _, kept := range []string{"env:OPENAI_API_KEY", "secret://team/token", "application/json", "--verbose"} {
			if !strings.Contains(export.Content, kept) {
				t.Fatalf("%s export dropped non-secret value %q:\n%s", format, kept, export.Content)
			}
		}
		if export.DocumentID != "seeded-profile" || export.Kind != "WorkerProfile" || export.Version != "1" || export.Scope.Ref != "primary" {
			t.Fatalf("%s export identity = %+v", format, export)
		}
	}
}

func TestConfigDocumentExportIsDeterministicAndHashed(t *testing.T) {
	document := seededSecretExportDocument()
	for _, format := range []ConfigDocumentExportFormat{ConfigDocumentExportYAML, ConfigDocumentExportJSON} {
		first, err := RenderConfigDocumentExport(document, format)
		if err != nil {
			t.Fatalf("first render: %v", err)
		}
		for i := 0; i < 5; i++ {
			next, err := RenderConfigDocumentExport(document, format)
			if err != nil {
				t.Fatalf("repeat render: %v", err)
			}
			if next.Content != first.Content || next.ExportSHA256 != first.ExportSHA256 {
				t.Fatalf("%s export is not deterministic", format)
			}
		}
		sum := sha256.Sum256([]byte(first.Content))
		if first.ExportSHA256 != "sha256:"+hex.EncodeToString(sum[:]) {
			t.Fatalf("%s export_sha256 does not hash content", format)
		}
	}
}

func TestConfigDocumentExportRendersCanonicalEnvelope(t *testing.T) {
	document := seededSecretExportDocument()
	yamlExport, err := RenderConfigDocumentExport(document, ConfigDocumentExportYAML)
	if err != nil {
		t.Fatalf("yaml render: %v", err)
	}
	if !strings.HasPrefix(yamlExport.Content, "apiVersion: mycelis.ai/v1\nkind: WorkerProfile\nmetadata:\n") {
		t.Fatalf("yaml envelope order unexpected:\n%s", yamlExport.Content)
	}
	var fromYAML map[string]any
	if err := yaml.Unmarshal([]byte(yamlExport.Content), &fromYAML); err != nil {
		t.Fatalf("yaml export does not parse: %v", err)
	}
	spec := fromYAML["spec"].(map[string]any)
	if spec["retries"] != 3 || spec["ratio"] != 0.25 || spec["enabled"] != true {
		t.Fatalf("yaml scalar types not preserved: %#v", spec)
	}
	jsonExport, err := RenderConfigDocumentExport(document, ConfigDocumentExportJSON)
	if err != nil {
		t.Fatalf("json render: %v", err)
	}
	var fromJSON map[string]any
	if err := json.Unmarshal([]byte(jsonExport.Content), &fromJSON); err != nil {
		t.Fatalf("json export does not parse: %v", err)
	}
	if fromJSON["kind"] != "WorkerProfile" || fromJSON["spec"].(map[string]any)["retries"] != float64(3) {
		t.Fatalf("json envelope unexpected: %#v", fromJSON)
	}
}

func TestConfigDocumentExportCleanDocumentIsNotRedacted(t *testing.T) {
	document := seededSecretExportDocument()
	document.Metadata.SecretRefs = []string{"env:OPENAI_API_KEY"}
	document.Spec = json.RawMessage(`{"role":"reviewer","api_key":"env:OPENAI_API_KEY","password":"","headers":{"Authorization":"vault://team/auth"}}`)
	export, err := RenderConfigDocumentExport(document, ConfigDocumentExportJSON)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if export.RedactionApplied || len(export.RedactedPaths) != 0 {
		t.Fatalf("clean document redacted: %#v", export.RedactedPaths)
	}
	if !strings.Contains(export.Content, "vault://team/auth") {
		t.Fatalf("managed ref hidden:\n%s", export.Content)
	}
}

func TestConfigDocumentExportRejectsUnknownFormat(t *testing.T) {
	for _, raw := range []string{"xml", "YAML", "yml", " json", "toml"} {
		if _, err := ParseConfigDocumentExportFormat(raw); err == nil {
			t.Fatalf("ParseConfigDocumentExportFormat(%q) accepted", raw)
		}
		if _, err := RenderConfigDocumentExport(seededSecretExportDocument(), ConfigDocumentExportFormat(raw)); err == nil {
			t.Fatalf("RenderConfigDocumentExport(%q) accepted", raw)
		}
	}
	if format, err := ParseConfigDocumentExportFormat(""); err != nil || format != ConfigDocumentExportYAML {
		t.Fatalf("empty format = %q, %v; want yaml default", format, err)
	}
}

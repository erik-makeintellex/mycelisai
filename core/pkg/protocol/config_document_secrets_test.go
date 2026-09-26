package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func classifierTestDocument(spec string) ConfigDocument {
	document := seededSecretExportDocument()
	// OutcomeTemplate keeps family decoders out of classifier assertions.
	document.Kind = ConfigDocumentKindOutcomeTemplate
	document.Metadata.SecretRefs = nil
	document.Spec = json.RawMessage(spec)
	return document
}

// Every value below is a fake fixture carrying exportSecretMarker.
var strictOnlySecretSpecs = map[string]string{
	"array token":          `{"args":["--verbose","ghp_EXPORTTESTSECRET0003abcdef"]}`,
	"embedded flag":        `{"command":"run --token=xoxb-EXPORTTESTSECRET0004"}`,
	"authorization header": `{"headers":{"Authorization":"Basic EXPORTTESTSECRET"}}`,
	"bearer in text":       `{"note":"curl -H 'Authorization: Bearer EXPORTTESTSECRET0002xyz'"}`,
	"rsa private key":      `{"notes":"-----BEGIN RSA PRIVATE KEY-----\nEXPORTTESTSECRET\n-----END RSA PRIVATE KEY-----"}`,
	"pgp private block":    `{"notes":"-----BEGIN PGP PRIVATE KEY BLOCK-----\nEXPORTTESTSECRET\n-----END PGP PRIVATE KEY BLOCK-----"}`,
	"aws key token":        `{"note":"use AKIAEXPORTTESTSECRET here"}`,
	"secret-looking key":   `{"sk-EXPORTTESTSECRET0008":"value"}`,
	"url userinfo":         `{"dsn":"postgres://app:EXPORTTESTSECRETpw@db:5432/app"}`,
	"url in text":          `{"note":"connect with redis://:EXPORTTESTSECRET@cache:6379/0 today"}`,
	"env assignment":       `{"env":["OPENAI_API_KEY=EXPORTTESTSECRETvalue"]}`,
	"uppercase assignment": `{"script":"export DB_PASSWORD=EXPORTTESTSECRET"}`,
	"query assignment":     `{"url":"https://api.example.test/v1?access_token=EXPORTTESTSECRET01&x=1"}`,
	"stripe live":          `{"note":"key sk_live_EXPORTTESTSECRET01"}`,
	"stripe test":          `{"note":"key sk_test_EXPORTTESTSECRET01"}`,
	"github server token":  `{"note":"ghs_EXPORTTESTSECRETabcdef012345"}`,
	"github oauth token":   `{"note":"x gho_EXPORTTESTSECRETabcdef012345"}`,
	"github user token":    `{"note":"x ghu_EXPORTTESTSECRETabcdef012345"}`,
	"github refresh token": `{"note":"x ghr_EXPORTTESTSECRETabcdef012345"}`,
	"gitlab token":         `{"note":"x glpat-EXPORTTESTSECRET0123"}`,
	"google api key":       `{"note":"x AIzaEXPORTTESTSECRET0123456789abcd"}`,
	"hugging face token":   `{"note":"x hf_EXPORTTESTSECRET0123456789abcd"}`,
	"jwt":                  `{"note":"eyJhbGciOiJub25lIn0.eyJzdWIiOiJ0ZXN0In0.EXPORTTESTSECRETsig"}`,
	"passphrase key":       `{"tls":{"passphrase":"EXPORTTESTSECRET phrase"}}`,
	"auth string key":      `{"auth":"EXPORTTESTSECRETvalue"}`,
	"signing key":          `{"jwt":{"signing_key":"EXPORTTESTSECRET-signing"}}`,
}

func TestConfigDocumentRedactStrictRulesRejectNewWritesAndHideExports(t *testing.T) {
	for name, spec := range strictOnlySecretSpecs {
		document := classifierTestDocument(spec)
		issues := ValidateNewConfigDocument(document)
		if !hasConfigDocumentIssue(issues, "spec.raw_secret") {
			t.Fatalf("%s: strict issues = %+v, want spec.raw_secret", name, issues)
		}
		for _, issue := range issues {
			if strings.Contains(issue.Field+issue.Message, exportSecretMarker) {
				t.Fatalf("%s: validation issue echoed a secret: %+v", name, issue)
			}
		}
		if DryRunConfigDocument(document).Valid {
			t.Fatalf("%s: dry-run accepted a raw secret", name)
		}
		for _, format := range []ConfigDocumentExportFormat{ConfigDocumentExportYAML, ConfigDocumentExportJSON} {
			export, err := RenderConfigDocumentExport(document, format)
			if err != nil {
				t.Fatalf("%s: render %s: %v", name, format, err)
			}
			if strings.Contains(export.Content, exportSecretMarker) || !export.RedactionApplied {
				t.Fatalf("%s: %s export leaked or did not flag:\n%s", name, format, export.Content)
			}
		}
	}
}

func TestConfigDocumentRedactStoredRulesKeepLegacyRevisionsValid(t *testing.T) {
	// Admitted under the stored rules: must still revalidate and digest at read,
	// compile, activation, and rollback even though strict rules reject it now.
	for name, spec := range strictOnlySecretSpecs {
		document := classifierTestDocument(spec)
		if issues := ValidateConfigDocument(document); len(issues) != 0 {
			t.Fatalf("%s: stored-rule issues = %+v, want none", name, issues)
		}
		if _, err := CanonicalConfigDocumentDigest(document); err != nil {
			t.Fatalf("%s: stored-rule digest failed: %v", name, err)
		}
	}
}

func TestConfigDocumentRedactStrictIsSupersetOfStoredRules(t *testing.T) {
	storedRejected := []string{
		`{"api_key":"sk-live-not-a-ref"}`,
		`{"provider_value":"ghp_not-a-ref"}`,
		`{"value":"sk-learn"}`,
		`{"value":"AKIA-short"}`,
		`{"value":"-----BEGIN PRIVATE KEY-----"}`,
		`{"credentials":{"user":"a"}}`,
		`{"token_ref":"not a ref"}`,
	}
	for _, spec := range storedRejected {
		document := classifierTestDocument(spec)
		if len(ValidateConfigDocument(document)) == 0 {
			t.Fatalf("stored rules accepted %s", spec)
		}
		if len(ValidateNewConfigDocument(document)) == 0 {
			t.Fatalf("strict rules accepted a stored-rule rejection %s", spec)
		}
	}
}

func TestConfigDocumentRedactBenignTextStaysValid(t *testing.T) {
	benign := `{
		"role": "reviewer",
		"notes": "Use sk-learn models and read the xoxb-docs page. Bearer tokens rotate weekly; basic authentication is off.",
		"region": "Asia Pacific",
		"tags": ["desk-top", "sketch", "task-runner-for-nightly-builds"],
		"auth": {"token_ref": "env:OPENAI_API_KEY", "mode": "oauth"},
		"authorization": "",
		"limits": "max_tokens=100 auth=true timeout=30",
		"links": ["postgres://app@db:5432/app", "postgres://app:${DB_PASSWORD}@db/app", "https://example.test/a:b@c"],
		"env": ["TOKEN=$API_TOKEN", "API_KEY=env:OPENAI_API_KEY", "PASSWORD=<set-in-env>", "SECRET=***"],
		"passphrase_ref": "secret://tls/passphrase"
	}`
	document := classifierTestDocument(benign)
	for _, issue := range ValidateNewConfigDocument(document) {
		t.Fatalf("benign text flagged: %+v", issue)
	}
	export, err := RenderConfigDocumentExport(document, ConfigDocumentExportJSON)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if export.RedactionApplied {
		t.Fatalf("benign text redacted at %v", export.RedactedPaths)
	}
}

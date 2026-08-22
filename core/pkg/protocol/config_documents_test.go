package protocol

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func validConfigDocument() ConfigDocument {
	return ConfigDocument{
		APIVersion: ConfigDocumentAPIVersionV1,
		Kind:       ConfigDocumentKindOutcomeTemplate,
		Metadata: ConfigDocumentMetadata{
			ID:      "outcome-template-1",
			Name:    "Browser package",
			Version: "1.0.0",
			OwnerID: "operator-1",
			Scope: ConfigDocumentScope{
				Kind: ConfigDocumentScopeWorkspace,
				Ref:  "workspace-1",
			},
			Enabled: true,
			Source: ConfigDocumentSource{
				Kind: ConfigDocumentSourceFile,
				Ref:  "templates/browser-package.json",
			},
			SecretRefs: []string{"env:OPENAI_API_KEY", "vault:mycelis/search/token"},
			Governance: ConfigDocumentGovernance{
				RiskLevel:       ConfigDocumentRiskMedium,
				ApprovalPosture: ApprovalPostureRequired,
			},
		},
		Spec: json.RawMessage(`{"deliverable":{"format":"browser_package"},"auth":{"token_ref":"env:OPENAI_API_KEY"}}`),
	}
}

func hasConfigDocumentIssue(issues []ConfigDocumentValidationIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func TestValidateConfigDocumentAcceptsV1Document(t *testing.T) {
	document := validConfigDocument()
	if issues := ValidateConfigDocument(document); len(issues) != 0 {
		t.Fatalf("ValidateConfigDocument() issues = %+v, want none", issues)
	}

	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"apiVersion":"mycelis.ai/v1"`, `"kind":"OutcomeTemplate"`, `"owner_id"`, `"secret_refs"`, `"governance"`, `"spec"`} {
		if !strings.Contains(string(payload), field) {
			t.Fatalf("marshaled document = %s, missing %s", payload, field)
		}
	}
}

func TestValidateConfigDocumentFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ConfigDocument)
		code   string
	}{
		{name: "unsupported api version", mutate: func(document *ConfigDocument) { document.APIVersion = "mycelis.ai/v2" }, code: "config.unsupported_api_version"},
		{name: "unsupported kind", mutate: func(document *ConfigDocument) { document.Kind = "TeamManifest" }, code: "config.unsupported_kind"},
		{name: "missing id", mutate: func(document *ConfigDocument) { document.Metadata.ID = "" }, code: "metadata.missing_id"},
		{name: "invalid id", mutate: func(document *ConfigDocument) { document.Metadata.ID = "Not Stable" }, code: "metadata.invalid_id"},
		{name: "missing name", mutate: func(document *ConfigDocument) { document.Metadata.Name = " " }, code: "metadata.missing_name"},
		{name: "missing version", mutate: func(document *ConfigDocument) { document.Metadata.Version = "" }, code: "metadata.missing_version"},
		{name: "invalid version", mutate: func(document *ConfigDocument) { document.Metadata.Version = "version one" }, code: "metadata.invalid_version"},
		{name: "missing owner", mutate: func(document *ConfigDocument) { document.Metadata.OwnerID = "" }, code: "metadata.missing_owner"},
		{name: "unsupported scope", mutate: func(document *ConfigDocument) { document.Metadata.Scope.Kind = "global" }, code: "metadata.unsupported_scope"},
		{name: "missing scoped ref", mutate: func(document *ConfigDocument) { document.Metadata.Scope.Ref = "" }, code: "metadata.missing_scope_ref"},
		{name: "built in scope ref", mutate: func(document *ConfigDocument) {
			document.Metadata.Scope = ConfigDocumentScope{Kind: ConfigDocumentScopeBuiltIn, Ref: "unexpected"}
		}, code: "metadata.invalid_scope_ref"},
		{name: "unsupported source", mutate: func(document *ConfigDocument) { document.Metadata.Source.Kind = "webhook" }, code: "metadata.unsupported_source"},
		{name: "missing provenance", mutate: func(document *ConfigDocument) { document.Metadata.Source.Ref = "" }, code: "metadata.missing_source_ref"},
		{name: "unsupported risk", mutate: func(document *ConfigDocument) { document.Metadata.Governance.RiskLevel = "critical" }, code: "metadata.unsupported_risk_level"},
		{name: "unsupported approval", mutate: func(document *ConfigDocument) { document.Metadata.Governance.ApprovalPosture = "implicit" }, code: "metadata.unsupported_approval_posture"},
		{name: "raw metadata secret", mutate: func(document *ConfigDocument) { document.Metadata.SecretRefs = []string{"sk-live-not-a-ref"} }, code: "metadata.invalid_secret_ref"},
		{name: "duplicate metadata secret ref", mutate: func(document *ConfigDocument) {
			document.Metadata.SecretRefs = []string{"env:OPENAI_API_KEY", "env:OPENAI_API_KEY"}
		}, code: "metadata.duplicate_secret_ref"},
		{name: "missing spec", mutate: func(document *ConfigDocument) { document.Spec = nil }, code: "spec.invalid_json"},
		{name: "empty spec", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`{}`) }, code: "spec.empty"},
		{name: "null spec", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`null`) }, code: "spec.invalid_json"},
		{name: "array spec", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`[]`) }, code: "spec.invalid_json"},
		{name: "malformed spec", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`{"name":`) }, code: "spec.invalid_json"},
		{name: "trailing spec value", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`{"name":"one"} {"name":"two"}`) }, code: "spec.invalid_json"},
		{name: "duplicate spec key", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`{"name":"one","name":"two"}`) }, code: "spec.invalid_json"},
		{name: "raw secret field", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`{"api_key":"sk-live-not-a-ref"}`) }, code: "spec.raw_secret"},
		{name: "prefixed raw secret field", mutate: func(document *ConfigDocument) {
			document.Spec = json.RawMessage(`{"database_password":"plain-password"}`)
		}, code: "spec.raw_secret"},
		{name: "raw secret looking value", mutate: func(document *ConfigDocument) { document.Spec = json.RawMessage(`{"provider_value":"ghp_not-a-ref"}`) }, code: "spec.raw_secret"},
		{name: "invalid nested secret ref", mutate: func(document *ConfigDocument) {
			document.Spec = json.RawMessage(`{"auth":{"token_ref":"plain-token"}}`)
		}, code: "spec.invalid_secret_ref"},
		{name: "invalid compound secret ref", mutate: func(document *ConfigDocument) {
			document.Spec = json.RawMessage(`{"auth":{"api_key_secret_ref":"plain-token"}}`)
		}, code: "spec.invalid_secret_ref"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document := validConfigDocument()
			tt.mutate(&document)
			issues := ValidateConfigDocument(document)
			if !hasConfigDocumentIssue(issues, tt.code) {
				t.Fatalf("ValidateConfigDocument() issues = %+v, want code %q", issues, tt.code)
			}
		})
	}
}

func TestValidateConfigDocumentAcceptsSupportedScopesAndSecretRefs(t *testing.T) {
	tests := []struct {
		name       string
		scope      ConfigDocumentScope
		secretRefs []string
	}{
		{name: "built in", scope: ConfigDocumentScope{Kind: ConfigDocumentScopeBuiltIn}, secretRefs: []string{"OPENAI_API_KEY"}},
		{name: "organization", scope: ConfigDocumentScope{Kind: ConfigDocumentScopeOrganization, Ref: "org-1"}, secretRefs: []string{"env:OPENAI_API_KEY"}},
		{name: "workspace", scope: ConfigDocumentScope{Kind: ConfigDocumentScopeWorkspace, Ref: "workspace-1"}, secretRefs: []string{"secret:mycelis/openai"}},
		{name: "operator", scope: ConfigDocumentScope{Kind: ConfigDocumentScopeOperator, Ref: "operator-1"}, secretRefs: []string{"secret://mycelis/operator", "sm://projects/mycelis/openai"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document := validConfigDocument()
			document.Metadata.Scope = tt.scope
			document.Metadata.SecretRefs = tt.secretRefs
			document.Spec = json.RawMessage(`{"credential":"env:OPENAI_API_KEY"}`)
			if issues := ValidateConfigDocument(document); len(issues) != 0 {
				t.Fatalf("ValidateConfigDocument() issues = %+v, want none", issues)
			}
		})
	}
}

func TestValidateConfigDocumentDoesNotTreatOrdinaryRefsAsSecrets(t *testing.T) {
	document := validConfigDocument()
	document.Spec = json.RawMessage(`{"payload_schema_ref":"schemas/outcome-v1","template_ref":"template-1"}`)

	if issues := ValidateConfigDocument(document); len(issues) != 0 {
		t.Fatalf("ValidateConfigDocument() issues = %+v, want ordinary refs accepted", issues)
	}
}

func TestCanonicalConfigDocumentDigestIgnoresSpecObjectKeyOrder(t *testing.T) {
	first := validConfigDocument()
	first.Spec = json.RawMessage(`{"z":3,"nested":{"b":2,"a":1},"items":[{"y":2,"x":1}]}`)
	second := validConfigDocument()
	second.Spec = json.RawMessage("{\n  \"items\": [{\"x\": 1, \"y\": 2}],\n  \"nested\": {\"a\": 1, \"b\": 2},\n  \"z\": 3\n}")

	firstDigest, err := CanonicalConfigDocumentDigest(first)
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := CanonicalConfigDocumentDigest(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstDigest != secondDigest {
		t.Fatalf("digests differ across object key order: %q != %q", firstDigest, secondDigest)
	}
	if !strings.HasPrefix(firstDigest, "sha256:") || len(firstDigest) != len("sha256:")+64 {
		t.Fatalf("digest = %q, want sha256-prefixed hex", firstDigest)
	}
}

func TestCanonicalConfigDocumentDigestChangesWithAuthoritativeContent(t *testing.T) {
	first := validConfigDocument()
	second := validConfigDocument()
	second.Metadata.Version = "1.0.1"

	firstDigest, err := CanonicalConfigDocumentDigest(first)
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := CanonicalConfigDocumentDigest(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstDigest == secondDigest {
		t.Fatal("digest did not change with document version")
	}

	second = validConfigDocument()
	second.Spec = json.RawMessage(`{"deliverable":{"format":"pdf"}}`)
	secondDigest, err = CanonicalConfigDocumentDigest(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstDigest == secondDigest {
		t.Fatal("digest did not change with spec content")
	}
}

func TestCanonicalConfigDocumentDigestRejectsInvalidDocument(t *testing.T) {
	document := validConfigDocument()
	document.Metadata.Version = ""
	if digest, err := CanonicalConfigDocumentDigest(document); err == nil || digest != "" {
		t.Fatalf("CanonicalConfigDocumentDigest() = %q, %v; want empty digest and error", digest, err)
	}
}

func TestDryRunConfigDocumentDescribesEffectWithoutMutation(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		action  ConfigDocumentEffectAction
	}{
		{name: "activate", enabled: true, action: ConfigDocumentEffectActivate},
		{name: "deactivate", enabled: false, action: ConfigDocumentEffectDeactivate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document := validConfigDocument()
			document.Metadata.Enabled = tt.enabled
			before := ConfigDocument{
				APIVersion: document.APIVersion,
				Kind:       document.Kind,
				Metadata:   document.Metadata,
				Spec:       append(json.RawMessage(nil), document.Spec...),
			}
			before.Metadata.SecretRefs = append([]string(nil), document.Metadata.SecretRefs...)

			result := DryRunConfigDocument(document)
			if !result.Valid || result.Digest == "" || len(result.Issues) != 0 {
				t.Fatalf("DryRunConfigDocument() = %+v, want valid digest and no issues", result)
			}
			if result.Effect.Action != tt.action || result.Effect.DocumentID != document.Metadata.ID || result.Effect.Enabled != tt.enabled {
				t.Fatalf("dry-run effect = %+v, want action %q for %q", result.Effect, tt.action, document.Metadata.ID)
			}
			if !reflect.DeepEqual(document, before) {
				t.Fatalf("DryRunConfigDocument() mutated input:\n got  %+v\n want %+v", document, before)
			}
		})
	}
}

func TestDryRunConfigDocumentInvalidHasNoDigestOrEffect(t *testing.T) {
	document := validConfigDocument()
	document.Spec = json.RawMessage(`{"password":"raw-password"}`)

	result := DryRunConfigDocument(document)
	if result.Valid || result.Digest != "" || result.Effect.Action != ConfigDocumentEffectNone {
		t.Fatalf("DryRunConfigDocument() = %+v, want invalid result with no digest or effect", result)
	}
	if !hasConfigDocumentIssue(result.Issues, "spec.raw_secret") {
		t.Fatalf("DryRunConfigDocument() issues = %+v, want spec.raw_secret", result.Issues)
	}
}

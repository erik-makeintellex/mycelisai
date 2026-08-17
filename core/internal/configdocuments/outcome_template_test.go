package configdocuments

import (
	"encoding/json"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestCompileOutcomeTemplateDocumentUsesEnvelopeSnapshotAndPrecedence(t *testing.T) {
	document := parsedDocumentForCompile(t)
	result, err := CompileOutcomeTemplateDocument(
		document,
		protocol.MinimumSufficientBrief{TargetOutcome: "Operator outcome"},
		protocol.MinimumSufficientBrief{Audience: "Policy audience"},
	)
	if err != nil {
		t.Fatalf("CompileOutcomeTemplateDocument: %v", err)
	}
	if !result.Ready || result.WorkIntent == nil {
		t.Fatalf("expected ready WorkIntent: %+v", result)
	}
	if result.Brief.TargetOutcome != "Operator outcome" || result.Brief.Audience != "Policy audience" {
		t.Fatalf("unexpected precedence: %+v", result.Brief)
	}
	if result.TemplateSnapshot.ID != document.Metadata.ID || result.TemplateSnapshot.Version != document.Metadata.Version {
		t.Fatalf("unexpected snapshot: %+v", result.TemplateSnapshot)
	}
	wantDigest, _ := protocol.CanonicalConfigDocumentDigest(document)
	if result.TemplateSnapshot.Digest != wantDigest {
		t.Fatalf("snapshot digest = %q, want %q", result.TemplateSnapshot.Digest, wantDigest)
	}
}

func parsedDocumentForCompile(t *testing.T) protocol.ConfigDocument {
	t.Helper()
	spec, err := json.Marshal(map[string]any{
		"intent_kind": "project",
		"defaults": map[string]any{
			"target_outcome":      "Template outcome",
			"audience":            "Template audience",
			"essential_behavior":  []string{"Work end to end"},
			"quality_bar":         "Production ready",
			"delivery_form":       "browser_app",
			"constraints":         []string{"Self contained"},
			"acceptance_evidence": []string{"Interactive proof"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindOutcomeTemplate,
		Metadata: protocol.ConfigDocumentMetadata{
			ID: "browser-app", Name: "Browser app", Version: "1", OwnerID: "operator-1", Enabled: true,
			Scope:  protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeWorkspace, Ref: "workspace-1"},
			Source: protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceAPI, Ref: "request:test"},
			Governance: protocol.ConfigDocumentGovernance{
				RiskLevel: protocol.ConfigDocumentRiskLow, ApprovalPosture: protocol.ApprovalPostureRequired,
			},
		},
		Spec: spec,
	}
}

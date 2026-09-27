package configdocuments

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

const tokenBudgetSeedFile = "../../config/documents/templates/token-budget-defaults.yaml"

func TestTokenBudgetSeed_IsBuiltInAndEqualsCompiledDefaults(t *testing.T) {
	raw, err := os.ReadFile(tokenBudgetSeedFile)
	if err != nil {
		t.Fatal(err)
	}
	document, err := ParseDocument(raw, "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if document.Kind != protocol.ConfigDocumentKindTokenBudgetPolicy || document.Metadata.ID != protocol.TokenBudgetDefaultsDocumentID ||
		document.Metadata.Scope.Kind != protocol.ConfigDocumentScopeBuiltIn || document.Metadata.Source.Ref != "core/config/documents/templates/token-budget-defaults.yaml" {
		t.Fatalf("seed metadata = %+v", document.Metadata)
	}
	if issues := protocol.ValidateNewConfigDocument(document); len(issues) != 0 {
		t.Fatalf("seed invalid: %+v", issues)
	}
	compiled, err := CompileDocument(document, protocol.MinimumSufficientBrief{}, protocol.MinimumSufficientBrief{})
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := compiled.(protocol.TokenBudgetPolicySpec)
	if !ok || !reflect.DeepEqual(spec, protocol.DefaultTokenBudgetPolicySpec()) {
		t.Fatalf("seed spec = %+v, want the compiled defaults", compiled)
	}
	assertNoRawSecretsOrSwarmSubjects(t, "token-budget-defaults.yaml", raw)
}

func TestTokenBudgetPolicy_PublicStoreAndActivationPathsAreClosed(t *testing.T) {
	spec, _ := json.Marshal(protocol.DefaultTokenBudgetPolicySpec())
	document := protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindTokenBudgetPolicy,
		Metadata: protocol.ConfigDocumentMetadata{
			ID: protocol.TokenBudgetOverridesDocumentID, Name: "Forged budgets", Version: "9", OwnerID: "attacker",
			Scope: protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOperator, Ref: "default"}, Enabled: true,
			Source:     protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceAPI, Ref: "api:/api/v1/config-documents"},
			Governance: protocol.ConfigDocumentGovernance{RiskLevel: protocol.ConfigDocumentRiskLow, ApprovalPosture: protocol.ApprovalPostureAutoAllowed},
		},
		Spec: spec,
	}
	if err := guardPublicStore("user-1", document); !errors.Is(err, ErrTokenBudgetPolicyReserved) {
		t.Fatalf("public store err = %v, want reserved", err)
	}
	if err := guardPublicActivation(RevisionRecord{RecordID: "r", Document: document}); !errors.Is(err, ErrTokenBudgetPolicyReserved) {
		t.Fatalf("public activation err = %v, want reserved", err)
	}
	if err := guardTokenBudgetStore("system:bootstrap", document); err == nil {
		t.Fatal("reserved actor accepted on the budget path")
	}
	if err := guardTokenBudgetStore("user-1", document); err != nil {
		t.Fatalf("budget path refused a valid operator revision: %v", err)
	}
}

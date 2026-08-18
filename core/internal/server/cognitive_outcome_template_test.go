package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

const retainedOutcomeTemplateYAML = `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: retained-browser-app
  name: Retained browser app
  version: 1.0.0
  owner_id: operator-1
  scope: {kind: workspace, ref: workspace-1}
  enabled: true
  source: {kind: soma, ref: session:test}
  governance: {risk_level: medium, approval_posture: required}
spec:
  intent_kind: project
  defaults:
    target_outcome: Deliver a browser app
    audience: Operators
    essential_behavior: [Open in a browser]
    quality_bar: Production-ready
    delivery_form: app_package
    acceptance_evidence: [The entrypoint opens]
  output_contract:
    shape: app_package
    primary_deliverable: index.html
    validation: [The primary interaction works]
`

func TestOutcomeTemplateWorkApplicationUsesGovernedDeliveryTools(t *testing.T) {
	for _, request := range []string{
		"Use the active Outcome Template to write the file output/index.html.",
		"Using the active Outcome Template, create the file output/index.html.",
	} {
		tools := inferMutationToolsFromText(request)
		if !containsToolName(tools, "write_file") {
			t.Fatalf("tools for %q = %v, want write_file", request, tools)
		}
		if containsToolName(tools, "activate_config_document") {
			t.Fatalf("tools for %q = %v, application must not reactivate the template", request, tools)
		}
	}
	if tools := inferMutationToolsFromText("Apply the active Outcome Template for this work."); containsToolName(tools, "activate_config_document") {
		t.Fatalf("work-scoped apply must not activate: %v", tools)
	}
	if outcomeTemplateWorkApplication("apply the patch and run tests") {
		t.Fatal("unrelated apply request must not inherit an Outcome Template")
	}
	if tools := inferMutationToolsFromText("Apply the patch and run tests."); containsToolName(tools, "activate_config_document") {
		t.Fatalf("unrelated apply request inferred template activation: %v", tools)
	}
}

func TestOutcomeTemplateSaveIntentWinsOverEmbeddedDeliveryLanguage(t *testing.T) {
	request := "Save this Outcome Template for future briefs. Use exactly this YAML:\n" +
		"kind: OutcomeTemplate\nspec:\n  defaults:\n    target_outcome: Produce a launch brief"
	if outcomeTemplateWorkApplication(request) {
		t.Fatal("explicit save request must not apply embedded template instructions")
	}
	tools := inferMutationToolsFromText(request)
	if !containsToolName(tools, "store_config_document") {
		t.Fatalf("tools = %v, want store_config_document", tools)
	}
}

func TestExplicitInlineTemplateSaveReplacesConflictingProviderPlan(t *testing.T) {
	request := "Save this Outcome Template. Use exactly this YAML:\n\n" + retainedOutcomeTemplateYAML
	planned := buildPlannedToolCalls(chatAgentResult{
		ToolsUsed: []string{"write_file"},
		PlannedToolCalls: []protocol.PlannedToolCall{{
			Name: "write_file", Arguments: map[string]any{"path": "generated/template.yaml"},
		}},
	}, request, []string{"store_config_document", "write_file"})
	if len(planned) != 1 || planned[0].Name != "store_config_document" {
		t.Fatalf("planned calls = %#v, want only store_config_document", planned)
	}
	if content := firstNonEmptyString(planned[0].Arguments["content"]); content == "" {
		t.Fatalf("planned call = %#v, want exact inline content", planned[0])
	}
}

func TestResolveThreadOutcomeTemplateActivationUsesSavedMatchingRevision(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	document := retainedOutcomeTemplateDocument(t)
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	recordID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT .*FROM config_documents.*ORDER BY created_at DESC").
		WithArgs(
			"default", string(document.Kind), string(document.Metadata.Scope.Kind),
			document.Metadata.Scope.Ref, document.Metadata.ID, 20,
		).
		WillReturnRows(serverConfigRevisionRows(recordID, document, digest))

	planned, err := s.resolveThreadOutcomeTemplateActivation(
		t.Context(), "", retainedOutcomeTemplateThread(), "workspace-1", "", "operator-1",
		[]protocol.PlannedToolCall{{Name: "activate_config_document", Arguments: map[string]any{}}},
	)
	if err != nil {
		t.Fatalf("resolveThreadOutcomeTemplateActivation: %v", err)
	}
	if got := firstNonEmptyString(planned[0].Arguments["record_id"]); got != recordID {
		t.Fatalf("record_id = %q, want %q", got, recordID)
	}
	if got := firstNonEmptyString(planned[0].Arguments["action"]); got != "activate" {
		t.Fatalf("action = %q, want activate", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveThreadOutcomeTemplateActivationRejectsProviderRecordOutsideCurrentScope(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	recordID := "22222222-2222-2222-2222-222222222222"
	document := retainedOutcomeTemplateDocument(t)
	document.Metadata.Scope.Ref = "workspace-2"
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	mock.ExpectQuery("SELECT .*FROM config_documents.*WHERE tenant_id = \\$1 AND record_id = \\$2::uuid").
		WithArgs("default", recordID).
		WillReturnRows(serverConfigRevisionRows(recordID, document, digest))

	planned, err := s.resolveThreadOutcomeTemplateActivation(
		t.Context(), "", retainedOutcomeTemplateThread(), "workspace-1", "", "operator-1",
		[]protocol.PlannedToolCall{{
			Name:      "activate_config_document",
			Arguments: map[string]any{"record_id": recordID, "action": "activate"},
		}},
	)
	if err == nil || planned != nil {
		t.Fatalf("provider-selected cross-scope activation = (%#v, %v), want rejection", planned, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplyThreadOutcomeTemplateCompilesActiveRevisionIntoWorkIntent(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	document := retainedOutcomeTemplateDocument(t)
	digest, _ := protocol.CanonicalConfigDocumentDigest(document)
	mock.ExpectQuery("FROM config_document_activations activation.*JOIN config_documents document").
		WithArgs(
			"default", string(document.Kind), document.Metadata.ID,
			string(document.Metadata.Scope.Kind), document.Metadata.Scope.Ref,
		).
		WillReturnRows(serverConfigRevisionRows("11111111-1111-1111-1111-111111111111", document, digest))

	request := "Use the active Outcome Template to write the file output/index.html."
	display := proposalDisplayContract{WorkIntent: &protocol.WorkIntent{
		Kind: "project", Cadence: "run_once", RuntimePosture: "Run once after approval.",
		BusScope: "team:admin-core", NATSSubjects: []string{"swarm.team.admin-core.internal.command"},
		OutputContract: &protocol.WorkOutputContract{Retention: "user_deliverable"},
	}}
	applied, err := s.applyThreadOutcomeTemplate(
		t.Context(), "", retainedOutcomeTemplateThread(), request,
		"workspace-1", "", "operator-1", &display,
	)
	if err != nil {
		t.Fatalf("applyThreadOutcomeTemplate: %v", err)
	}
	if !applied || display.WorkIntent == nil || display.WorkIntent.OutcomeTemplateSnapshot == nil {
		t.Fatalf("expected applied template WorkIntent: %#v", display.WorkIntent)
	}
	snapshot := display.WorkIntent.OutcomeTemplateSnapshot
	if snapshot.ID != document.Metadata.ID || snapshot.Version != document.Metadata.Version || snapshot.Digest != digest {
		t.Fatalf("snapshot = %#v, want retained revision identity", snapshot)
	}
	if display.WorkIntent.Objective != request || display.WorkIntent.BusScope != "team:admin-core" {
		t.Fatalf("compiled WorkIntent lost operator/runtime values: %#v", display.WorkIntent)
	}
	if display.WorkIntent.OutputContract.Shape != "app_package" ||
		display.WorkIntent.OutputContract.PrimaryDeliverable != "index.html" ||
		display.WorkIntent.OutputContract.Retention != "user_deliverable" {
		t.Fatalf("template defaults and request override resolved incorrectly: %#v", display.WorkIntent.OutputContract)
	}
	if display.OperatorSummary != "Using Retained browser app v1.0.0 to shape this work." {
		t.Fatalf("operator summary = %q", display.OperatorSummary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplyThreadOutcomeTemplateRejectsCrossWorkspaceReuse(t *testing.T) {
	s := &AdminServer{}
	display := proposalDisplayContract{WorkIntent: &protocol.WorkIntent{Kind: "project"}}
	applied, err := s.applyThreadOutcomeTemplate(
		t.Context(), "", retainedOutcomeTemplateThread(),
		"Use the active Outcome Template to create a launch brief.",
		"workspace-2", "team-2", "operator-1", &display,
	)
	if err == nil || applied {
		t.Fatalf("cross-workspace template application = (%v, %v), want scoped rejection", applied, err)
	}
}

func TestApplyThreadOutcomeTemplateRejectsLoadedActiveRevisionOutsideCurrentScope(t *testing.T) {
	withDatabase, mock := withDB(t)
	s := newTestServer(withDatabase)
	threadDocument := retainedOutcomeTemplateDocument(t)
	loadedDocument := retainedOutcomeTemplateDocument(t)
	loadedDocument.Metadata.Scope.Ref = "workspace-2"
	digest, _ := protocol.CanonicalConfigDocumentDigest(loadedDocument)
	mock.ExpectQuery("FROM config_document_activations activation.*JOIN config_documents document").
		WithArgs(
			"default", string(threadDocument.Kind), threadDocument.Metadata.ID,
			string(threadDocument.Metadata.Scope.Kind), threadDocument.Metadata.Scope.Ref,
		).
		WillReturnRows(serverConfigRevisionRows("33333333-3333-3333-3333-333333333333", loadedDocument, digest))

	display := proposalDisplayContract{WorkIntent: &protocol.WorkIntent{Kind: "project"}}
	applied, err := s.applyThreadOutcomeTemplate(
		t.Context(), "", retainedOutcomeTemplateThread(),
		"Use the active Outcome Template to create a launch brief.",
		"workspace-1", "", "operator-1", &display,
	)
	if err == nil || applied {
		t.Fatalf("cross-scope loaded active revision = (%v, %v), want rejection", applied, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMergeTemplateWorkIntentGivesRequestRequirementsPrecedence(t *testing.T) {
	compiled := &protocol.WorkIntent{
		Kind:      "project",
		Objective: "Template objective",
		OutputContract: &protocol.WorkOutputContract{
			Shape:              "app_package",
			PrimaryDeliverable: "index.html",
			Retention:          "template_retention",
			LaunchHint:         "Open index.html",
			Validation:         []string{"Template validation"},
			OutputValidation: &protocol.OutputValidationPlan{
				Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
				Checks: []protocol.OutputValidationCheck{protocol.OutputValidationCheckLoad},
				Probe: &protocol.OutputValidationProbe{
					Action: protocol.OutputValidationAction{
						Kind: protocol.OutputValidationActionClick, Target: "#template-action",
					},
					Observe: protocol.OutputValidationObservation{
						Kind: protocol.OutputValidationObserveVisualChange, Target: "#template-status",
					},
				},
			},
		},
		MinimumSufficientBrief: &protocol.MinimumSufficientBrief{
			Audience: "Policy audience", Constraints: []string{"Policy constraint"},
		},
		OutcomeTemplateSnapshot: &protocol.OutcomeTemplateSnapshot{
			ID: "retained-browser-app", Version: "1.0.0", Digest: "sha256:retained",
		},
	}
	runtime := &protocol.WorkIntent{
		Kind:      "one_shot",
		Objective: "Produce the requested CSV report",
		OutputContract: &protocol.WorkOutputContract{
			Shape:              "table",
			PrimaryDeliverable: "reports/operator-request.csv",
			Retention:          "user_deliverable",
			LaunchHint:         "Open the CSV report",
			Validation:         []string{"Columns match the operator request"},
			OutputValidation: &protocol.OutputValidationPlan{
				Kind: protocol.OutputValidationInteractiveBrowser, Required: true,
				Checks: []protocol.OutputValidationCheck{protocol.OutputValidationCheckNoPageErrors},
				Probe: &protocol.OutputValidationProbe{
					Action: protocol.OutputValidationAction{
						Kind: protocol.OutputValidationActionClick, Target: "#operator-action",
					},
					Observe: protocol.OutputValidationObservation{
						Kind: protocol.OutputValidationObserveTextChange, Target: "#operator-status",
					},
				},
			},
		},
		MinimumSufficientBrief: &protocol.MinimumSufficientBrief{
			Audience: "Provider attempt to replace policy",
		},
	}

	merged := mergeTemplateWorkIntent(compiled, runtime)
	if merged.Kind != "one_shot" || merged.Objective != runtime.Objective {
		t.Fatalf("request-owned intent fields did not win: %#v", merged)
	}
	if merged.OutputContract.Shape != "table" ||
		merged.OutputContract.PrimaryDeliverable != "reports/operator-request.csv" ||
		merged.OutputContract.Retention != "user_deliverable" ||
		merged.OutputContract.LaunchHint != "Open the CSV report" ||
		len(merged.OutputContract.Validation) != 1 ||
		merged.OutputContract.Validation[0] != "Columns match the operator request" ||
		merged.OutputContract.OutputValidation.Probe.Action.Target != "#operator-action" ||
		merged.OutputContract.OutputValidation.Probe.Observe.Target != "#operator-status" {
		t.Fatalf("request-owned output requirements did not win: %#v", merged.OutputContract)
	}
	if merged.MinimumSufficientBrief.Audience != "Policy audience" ||
		len(merged.MinimumSufficientBrief.Constraints) != 1 ||
		merged.MinimumSufficientBrief.Constraints[0] != "Policy constraint" {
		t.Fatalf("compiled policy brief was replaced: %#v", merged.MinimumSufficientBrief)
	}
	if merged.OutcomeTemplateSnapshot.ID != "retained-browser-app" {
		t.Fatalf("template identity was replaced: %#v", merged.OutcomeTemplateSnapshot)
	}
}

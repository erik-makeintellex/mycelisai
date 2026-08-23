package server

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestInferMutationToolsFromText_RecognizesRetainedEvocationContinuation(t *testing.T) {
	request := strings.Join([]string{
		"Use the retained team evocation brief at groups/mixed-output-team-b8066/planning/TEAM_EVOCATION.md now.",
		"Do research and council prep, then have the team build the actual playable browser game package.",
		"Return direct launch path and proof notes.",
	}, " ")

	tools := inferMutationToolsFromText(request)
	if !containsToolName(tools, "write_file") || !containsToolName(tools, "delegate_task") {
		t.Fatalf("tools = %#v, want write_file and delegate_task", tools)
	}
}

func TestProjectPackageResultContractPreservesJSONDecodedValidationPlan(t *testing.T) {
	content := contentContractForTeamRequest("Build an interactive browser application.")
	payload, err := json.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}

	result := projectPackageResultContract("delivery-team", decoded, "Build an interactive browser application.")
	plan, ok := result["output_validation"].(*protocol.OutputValidationPlan)
	if !ok {
		t.Fatalf("output_validation = %#v, want decoded typed plan", result["output_validation"])
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("decoded result plan is not runnable: %v", err)
	}
}

func TestDeterministicGovernedMutationResult_BuildsTeamEvocationContinuation(t *testing.T) {
	request := strings.Join([]string{
		"Use the retained team evocation brief at groups/mixed-output-team-b8066/planning/TEAM_EVOCATION.md now.",
		"Do research and council prep, then have the team build the actual playable browser game package.",
		"Return direct launch path and proof notes.",
	}, " ")
	mutTools := inferMutationToolsFromText(request)

	result, ok := deterministicGovernedMutationResult(request, mutTools)
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}
	if !containsToolName(result.ToolsUsed, "write_file") || !containsToolName(result.ToolsUsed, "delegate_task") {
		t.Fatalf("tools_used = %#v, want write_file and delegate_task", result.ToolsUsed)
	}

	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	requirePlannedCallNames(t, calls, "write_file", "delegate_task")
	if calls[0].Arguments["path"] != "groups/mixed-output-team-b8066/planning/RESEARCH_COUNCIL_HANDOFF.md" {
		t.Fatalf("handoff path = %#v", calls[0].Arguments["path"])
	}
	content, _ := calls[0].Arguments["content"].(string)
	for _, want := range []string{"Research And Council Handoff", "Delivery lane responsibilities", "openable browser game package"} {
		if !strings.Contains(content, want) {
			t.Fatalf("handoff content missing %q: %.200q", want, content)
		}
	}
	if calls[1].Arguments["team_id"] != "mixed-output-team-b8066" {
		t.Fatalf("delegate team_id = %#v", calls[1].Arguments["team_id"])
	}
	ask, ok := calls[1].Arguments["ask"].(map[string]any)
	if !ok {
		t.Fatalf("delegate ask = %#v, want map", calls[1].Arguments["ask"])
	}
	if ask["ask_kind"] != "implementation" || ask["lane_role"] != "implementer" {
		t.Fatalf("delegate ask routing = %#v", ask)
	}
	capabilities := confirmedActionStringSlice(ask["required_capabilities"])
	if containsToolName(capabilities, "research_for_blueprint") || containsToolName(capabilities, "consult_council") {
		t.Fatalf("delivery capabilities repeat completed preparation work: %#v", capabilities)
	}
	if !containsToolName(capabilities, "write_file") || !containsToolName(capabilities, "store_artifact") {
		t.Fatalf("delivery capabilities = %#v, want retained output tools", capabilities)
	}
	constraints := confirmedActionStringSlice(ask["constraints"])
	if !containsToolName(constraints, "Read every retained user-facing entrypoint back after writing it and validate the requested behavior before reporting completion.") {
		t.Fatalf("constraints = %#v, want retained output readback", constraints)
	}
	context, ok := ask["context"].(map[string]any)
	if !ok {
		t.Fatalf("delegate context = %#v, want map", ask["context"])
	}
	if context["team_evocation_brief"] != "groups/mixed-output-team-b8066/planning/TEAM_EVOCATION.md" {
		t.Fatalf("evocation brief context = %#v", context["team_evocation_brief"])
	}
	if context["research_council_handoff"] != "groups/mixed-output-team-b8066/planning/RESEARCH_COUNCIL_HANDOFF.md" {
		t.Fatalf("research handoff context = %#v", context["research_council_handoff"])
	}
	resultContract := mapArgument(context["result_contract"])
	if resultContract["validation_mode"] != "readback_against_exit_criteria" {
		t.Fatalf("result contract = %#v, want readback validation mode", resultContract)
	}
	if resultContract["package_folder"] != "groups/mixed-output-team-b8066/generated/package" {
		t.Fatalf("package_folder = %#v", resultContract["package_folder"])
	}
	if resultContract["package_entrypoint"] != "groups/mixed-output-team-b8066/generated/package/index.html" {
		t.Fatalf("package_entrypoint = %#v", resultContract["package_entrypoint"])
	}
	for _, want := range []string{"index.html", "README.md", "PROOF.md", "project-package.json"} {
		if !containsToolName(confirmedActionStringSlice(resultContract["files_required"]), want) {
			t.Fatalf("files_required = %#v, missing %q", resultContract["files_required"], want)
		}
	}
	validationPlan, ok := resultContract["output_validation"].(*protocol.OutputValidationPlan)
	if !ok {
		t.Fatalf("result contract output_validation = %#v, want typed plan", resultContract["output_validation"])
	}
	if err := validationPlan.Validate(); err != nil {
		t.Fatalf("result contract output validation is not runnable: %v", err)
	}
	exitCriteria := confirmedActionStringSlice(ask["exit_criteria"])
	for _, want := range []string{"playable controls respond in browser", "direct launch or view path is provided for the user or another agent"} {
		if !containsToolName(exitCriteria, want) {
			t.Fatalf("exit criteria = %#v, missing %q", exitCriteria, want)
		}
	}
	proof := confirmedActionStringSlice(ask["evidence_required"])
	if !containsToolName(proof, "headed gameplay proof or equivalent interaction proof") {
		t.Fatalf("proof = %#v, want gameplay proof", proof)
	}
}

func TestInferTeamEvocationContinuationPlan_RequiresTeamContext(t *testing.T) {
	request := "Use the retained team evocation brief now and make the work happen."
	if calls, ok := inferTeamEvocationContinuationPlanFromRequest(request); ok || len(calls) != 0 {
		t.Fatalf("calls = %#v, want no continuation plan without target team context", calls)
	}
}

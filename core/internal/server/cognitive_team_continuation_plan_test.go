package server

import (
	"strings"
	"testing"
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

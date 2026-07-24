package swarm

import (
	"strings"
	"testing"
)

func TestNormalizeTeamTriggerInput_PlainTextPassesThrough(t *testing.T) {
	got := normalizeTeamTriggerInput([]byte("inspect gate state"))
	if got != "inspect gate state" {
		t.Fatalf("trigger input = %q", got)
	}
}

func TestNormalizeTeamTriggerInput_StructuredAskRendersPrompt(t *testing.T) {
	got := normalizeTeamTriggerInput([]byte(`{
		"ask_kind":"research",
		"lane_role":"researcher",
		"goal":"Find the best documented approach.",
		"constraints":["Use primary sources only."],
		"exit_criteria":["Return one recommended path."],
		"evidence_required":["source_links"]
	}`))

	for _, want := range []string{
		"You have received a structured team ask.",
		"Use the ask to stay aligned on mission, scope, and proof needs.",
		"Do not force your response into a rigid template unless the ask explicitly requires one.",
		"Deliver the best output for the job while making sure it satisfies the ask goal, constraints, and required evidence.",
		"Ask kind: research",
		"Lane role: researcher",
		"Goal: Find the best documented approach.",
		"Constraints:",
		"- Use primary sources only.",
		"Exit criteria:",
		"- Return one recommended path.",
		"Evidence required:",
		"- source_links",
		"Complete the ask within scope, match the mission, and report the outcome clearly.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered prompt missing %q:\n%s", want, got)
		}
	}
}

func TestNormalizeTeamTriggerInput_CompactsDuplicatedExecutionContext(t *testing.T) {
	got := normalizeTeamTriggerInput([]byte(`{
		"ask_kind":"implementation",
		"goal":"Build the retained package.",
		"context":{
			"run_id":"run-1",
			"contract_id":"contract-1",
			"intent_proof_id":"proof-1",
			"research_council_handoff":"groups/team/planning/RESEARCH_COUNCIL_HANDOFF.md",
			"operator_request":"A very long operator request already represented by the goal.",
			"result_contract":{"acceptance_criteria":["already rendered separately"]}
		}
	}`))

	for _, want := range []string{"run-1", "contract-1", "proof-1", "RESEARCH_COUNCIL_HANDOFF.md"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered prompt missing compact context value %q:\n%s", want, got)
		}
	}
	for _, omitted := range []string{"A very long operator request", "result_contract", "already rendered separately"} {
		if strings.Contains(got, omitted) {
			t.Fatalf("rendered prompt retained duplicated context %q:\n%s", omitted, got)
		}
	}
}

func TestNormalizeTeamTriggerInput_RendersActionableResultContract(t *testing.T) {
	got := normalizeTeamTriggerInput([]byte(`{
		"ask_kind":"implementation",
		"goal":"Build the retained package.",
		"context":{
			"result_contract":{
				"kind":"project_package",
				"files_required":["README.md","PROOF.md","project-package.json"],
				"entrypoint_required":true,
				"folder_required":true,
				"validation_required":true,
				"proof_ref_required":true,
				"repair_channel":"soma"
			}
		}
	}`))

	for _, want := range []string{
		"Output contract:",
		"Kind: project_package",
		"Required files: README.md, PROOF.md, project-package.json",
		"Return a direct entrypoint.",
		"Read the retained output back and validate it against every exit criterion before reporting completion.",
		"Return a proof reference backed by the validation performed.",
		"If validation fails, report the blocker through soma instead of claiming completion.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered prompt missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "result_contract") {
		t.Fatalf("rendered prompt exposed raw result contract:\n%s", got)
	}
}

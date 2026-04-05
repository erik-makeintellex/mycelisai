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

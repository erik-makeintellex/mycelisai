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
		"Ask kind: research",
		"Lane role: researcher",
		"Goal: Find the best documented approach.",
		"Constraints:",
		"- Use primary sources only.",
		"Exit criteria:",
		"- Return one recommended path.",
		"Evidence required:",
		"- source_links",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered prompt missing %q:\n%s", want, got)
		}
	}
}

package swarm

import (
	"strings"
	"testing"
)

func TestShouldCouncilPreflight(t *testing.T) {
	if !shouldCouncilPreflight("create_team") || !shouldCouncilPreflight("delegate_task") {
		t.Fatal("expected team operations to require preflight")
	}
	if !shouldCouncilPreflight("local_command") || shouldCouncilPreflight("read_file") {
		t.Fatal("unexpected local command or read preflight posture")
	}
}

func TestCouncilPreflightMember(t *testing.T) {
	if got := councilPreflightMember("create_team"); got != "council-architect" {
		t.Fatalf("got %q", got)
	}
	if got := councilPreflightMember("delegate_task"); got != "council-architect" {
		t.Fatalf("got %q", got)
	}
	if got := councilPreflightMember("local_command"); got != "council-coder" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatCouncilPreflightQuestion(t *testing.T) {
	call := &toolCallPayload{Name: "create_team", Arguments: map[string]any{"team_id": "x"}}
	got := formatCouncilPreflightQuestion("make a team", call)
	if got == "" || !strings.Contains(got, "create_team") || !strings.Contains(got, "make a team") {
		t.Fatalf("unexpected question: %q", got)
	}
}

func TestToolCallFingerprint(t *testing.T) {
	got := toolCallFingerprint(&toolCallPayload{
		Name:      "local_command",
		Arguments: map[string]any{"command": "echo hello"},
	})
	if !strings.Contains(got, "local_command") || !strings.Contains(got, "echo hello") {
		t.Fatalf("unexpected fingerprint: %q", got)
	}
}

func TestPreferDirectDraftResponse(t *testing.T) {
	if !preferDirectDraftResponse("create a simple hello letter for me") {
		t.Fatal("expected direct draft preference for simple letter request")
	}
	if !preferDirectDraftResponse("Summarize the current Workspace V8 design objectives.") {
		t.Fatal("expected informational summary request to prefer a direct answer")
	}
	if preferDirectDraftResponse("write the result to workspace/hello.txt") {
		t.Fatal("did not expect direct draft preference for explicit file write")
	}
	if preferDirectDraftResponse("Explain the contents of workspace/logs/hello.txt") {
		t.Fatal("did not expect direct draft preference for explicit workspace file inspection")
	}
}

func TestShouldAvoidToolsForDirectDraft(t *testing.T) {
	if !shouldAvoidToolsForDirectDraft("write_file") {
		t.Fatal("expected write_file to be blocked for direct draft requests")
	}
	for _, name := range []string{"get_system_status", "list_teams", "read_signals", "consult_council", "generate_image"} {
		if shouldAvoidToolsForDirectDraft(name) {
			t.Fatalf("did not expect %s to be blocked by direct draft guard", name)
		}
	}
}

package swarm

import (
	"strings"
	"testing"
)

func TestResponseSuggestsUnexecutedAction(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "detects step style delegation planning",
			text: "To have the team provide updates, we need to delegate a specific task. Step 1: Delegate Task",
			want: true,
		},
		{
			name: "detects permission-seeking phrasing",
			text: "Would you like me to do that?",
			want: true,
		},
		{
			name: "detects example input narration",
			text: "Example Input:\n{\n  \"operation\": \"consult_council\",\n  \"arguments\": {\"member\": \"council-architect\"}\n}\nThis will route your request to the Architect.",
			want: true,
		},
		{
			name: "ignores normal result response",
			text: "Task delegated to team admin-core.",
			want: false,
		},
		{
			name: "ignores empty",
			text: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := responseSuggestsUnexecutedAction(tt.text)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseToolCall_FallbackOperationPayload(t *testing.T) {
	got := parseToolCall("{\"operation\":\"consult_council\",\"arguments\":{\"member\":\"council-architect\",\"question\":\"What API should we use?\"}}")
	if got == nil {
		t.Fatal("expected parsed tool call")
	}
	if got.Name != "consult_council" {
		t.Fatalf("name = %q, want consult_council", got.Name)
	}
	if got.Arguments["member"] != "council-architect" {
		t.Fatalf("member = %v, want council-architect", got.Arguments["member"])
	}
}

func TestParseToolCall_PrefersToolCallPayload(t *testing.T) {
	text := "{\"tool_call\":{\"name\":\"delegate_task\",\"arguments\":{\"team_id\":\"admin-core\",\"task\":\"x\"}}}\n{\"operation\":\"consult_council\",\"arguments\":{\"member\":\"council-architect\"}}"
	got := parseToolCall(text)
	if got == nil {
		t.Fatal("expected parsed tool call")
	}
	if got.Name != "delegate_task" {
		t.Fatalf("name = %q, want delegate_task", got.Name)
	}
}

func TestAutofillToolArguments_ConsultCouncilQuestion(t *testing.T) {
	call := &toolCallPayload{
		Name:      "consult_council",
		Arguments: map[string]any{"member": "council-architect"},
	}
	autofillToolArguments(call, "Use image API recommendations for this plan.")
	if call.Arguments["question"] != "Use image API recommendations for this plan." {
		t.Fatalf("question = %v", call.Arguments["question"])
	}
}

func TestAutofillToolArguments_DoesNotOverrideQuestion(t *testing.T) {
	call := &toolCallPayload{
		Name: "consult_council",
		Arguments: map[string]any{
			"member":   "council-architect",
			"question": "existing",
		},
	}
	autofillToolArguments(call, "new input")
	if call.Arguments["question"] != "existing" {
		t.Fatalf("question overwritten: %v", call.Arguments["question"])
	}
}

func TestAutofillToolArguments_ReadSignalsTopicPatternAlias(t *testing.T) {
	call := &toolCallPayload{
		Name:      "read_signals",
		Arguments: map[string]any{"topic_pattern": "swarm.team.admin-core.signal.status"},
	}
	autofillToolArguments(call, "check signals")
	if call.Arguments["subject"] != "swarm.team.admin-core.signal.status" {
		t.Fatalf("subject = %v", call.Arguments["subject"])
	}
}

func TestNormalizeCouncilMember(t *testing.T) {
	tests := map[string]string{
		"Architect":         "council-architect",
		"council architect": "council-architect",
		"council-coder":     "council-coder",
		"Sentry":            "council-sentry",
	}
	for in, want := range tests {
		if got := normalizeCouncilMember(in); got != want {
			t.Fatalf("normalizeCouncilMember(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAutofillToolArguments_ReadSignalsExtractsSubjectFromInput(t *testing.T) {
	call := &toolCallPayload{
		Name:      "read_signals",
		Arguments: map[string]any{},
	}
	autofillToolArguments(call, "read swarm.team.OpenClawSearchTeam.signal.status and report back")
	if call.Arguments["subject"] != "swarm.team.OpenClawSearchTeam.signal.status" {
		t.Fatalf("subject = %v", call.Arguments["subject"])
	}
}

func TestExtractNATSSubject(t *testing.T) {
	got := extractNATSSubject("please inspect (swarm.team.admin-core.signal.status). now")
	if got != "swarm.team.admin-core.signal.status" {
		t.Fatalf("got %q", got)
	}
}

func TestParseLooseToolCall(t *testing.T) {
	text := `{"tool_call":{"name":"consult_council","arguments":{"member":"council-architect"}}`
	got := parseLooseToolCall(text)
	if got == nil {
		t.Fatal("expected loose tool call parse")
	}
	if got.Name != "consult_council" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestAutofillToolArguments_ConsultCouncilInfersMember(t *testing.T) {
	call := &toolCallPayload{
		Name:      "consult_council",
		Arguments: map[string]any{},
	}
	autofillToolArguments(call, "Consult the architect about API strategy")
	if call.Arguments["member"] != "council-architect" {
		t.Fatalf("member = %v", call.Arguments["member"])
	}
	if call.Arguments["question"] == nil {
		t.Fatal("expected inferred question")
	}
}

func TestAutofillToolArguments_DelegateTaskPromotesCreateTeam(t *testing.T) {
	call := &toolCallPayload{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_name":  "OpenClawSearchTeam",
			"agent_type": "coder",
		},
	}
	autofillToolArguments(call, "create team")
	if call.Name != "create_team" {
		t.Fatalf("name = %q, want create_team", call.Name)
	}
	if call.Arguments["team_id"] != "OpenClawSearchTeam" {
		t.Fatalf("team_id = %v", call.Arguments["team_id"])
	}
	if call.Arguments["role"] != "coder" {
		t.Fatalf("role = %v", call.Arguments["role"])
	}
}

func TestShouldCouncilPreflight(t *testing.T) {
	if !shouldCouncilPreflight("create_team") {
		t.Fatal("expected create_team preflight")
	}
	if !shouldCouncilPreflight("delegate_task") {
		t.Fatal("expected delegate_task preflight")
	}
	if !shouldCouncilPreflight("local_command") {
		t.Fatal("expected local_command preflight")
	}
	if shouldCouncilPreflight("read_file") {
		t.Fatal("did not expect read_file preflight")
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
	if preferDirectDraftResponse("write the result to workspace/hello.txt") {
		t.Fatal("did not expect direct draft preference for explicit file write")
	}
}

func TestShouldAvoidToolsForDirectDraft(t *testing.T) {
	if !shouldAvoidToolsForDirectDraft("write_file") {
		t.Fatal("expected write_file to be blocked for direct draft requests")
	}
	if shouldAvoidToolsForDirectDraft("generate_image") {
		t.Fatal("did not expect generate_image to be blocked by direct draft guard")
	}
}

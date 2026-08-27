package swarm

import "testing"

func TestParseToolCallForExecutionRejectsTruncatedMutation(t *testing.T) {
	call, failure := parseToolCallForExecution(`{"tool_call":{"name":"write_file","arguments":{"path":"groups/team/generated/app/index.html","content":"<html>`)
	if call != nil {
		t.Fatalf("unsafe call parsed for execution: %+v", call)
	}
	if failure == nil || failure.ToolName != "write_file" {
		t.Fatalf("failure = %+v, want typed write_file failure", failure)
	}
}

func TestParseToolCallForExecutionAcceptsFencedCompleteMutation(t *testing.T) {
	call, failure := parseToolCallForExecution("```json\n{\"tool_call\":{\"name\":\"write_file\",\"arguments\":{\"path\":\"groups/team/generated/app/index.html\",\"content\":\"<html></html>\"}}}\n```")
	if failure != nil {
		t.Fatalf("unexpected failure: %v", failure)
	}
	if call == nil || call.Name != "write_file" || call.Arguments["content"] != "<html></html>" {
		t.Fatalf("call = %+v", call)
	}
}

func TestValidateMutationToolCallRejectsMissingArguments(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
	}{
		{name: "path", args: map[string]any{"content": "ready"}},
		{name: "content", args: map[string]any{"path": "groups/team/generated/app/index.html"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			failure := validateMutationToolCall(&toolCallPayload{Name: "write_file", Arguments: test.args})
			if failure == nil {
				t.Fatal("expected typed validation failure")
			}
		})
	}
}

func TestParseToolCallForExecutionPreservesLooseReadOnlyRecovery(t *testing.T) {
	call, failure := parseToolCallForExecution(`{"tool_call":{"name":"consult_council","arguments":{"member":"council-architect"}}`)
	if failure != nil || call == nil || call.Name != "consult_council" {
		t.Fatalf("call=%+v failure=%v", call, failure)
	}
}

func TestToolLoopCorrectsUnsafeWriteWithoutExecutingIt(t *testing.T) {
	tests := []struct {
		name  string
		first string
	}{
		{
			name:  "truncated",
			first: `{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/note.txt","content":"broken`,
		},
		{
			name:  "missing content",
			first: `{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/note.txt"}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &resultContractProvider{responses: []string{
				test.first,
				`{"tool_call":{"name":"write_file","arguments":{"path":"groups/delivery-team/generated/note.txt","content":"ready"}}}`,
				"The note is ready.",
			}}
			executor := &resultContractToolExecutor{}
			agent := resultContractTestAgent(provider, executor)

			result := agent.processMessageStructuredWithPosture("Write the note file.", nil, false)

			if len(executor.calls) != 1 || executor.calls[0] != "write_file" {
				t.Fatalf("executor calls = %v, want only the corrected write", executor.calls)
			}
			if executor.files["groups/delivery-team/generated/note.txt"] != "ready" {
				t.Fatalf("written files = %v", executor.files)
			}
			if result.Text != "The note is ready." {
				t.Fatalf("result text = %q", result.Text)
			}
		})
	}
}

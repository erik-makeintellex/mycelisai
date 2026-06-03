package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func wrongBlueprintAgentResult() chatAgentResult {
	return chatAgentResult{
		Text: `{"tool_call":{"name":"generate_blueprint","arguments":{"topic":"wrong"}}}`,
	}
}

func plannedCallsFromWrongBlueprint(request string, tools []string) []protocol.PlannedToolCall {
	return buildPlannedToolCalls(wrongBlueprintAgentResult(), request, tools)
}

func requirePlannedCallNames(t *testing.T, calls []protocol.PlannedToolCall, names ...string) {
	t.Helper()
	if len(calls) != len(names) {
		t.Fatalf("planned calls = %#v, want %v", calls, names)
	}
	for i, name := range names {
		if calls[i].Name != name {
			t.Fatalf("planned call %d = %q, want %q", i, calls[i].Name, name)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

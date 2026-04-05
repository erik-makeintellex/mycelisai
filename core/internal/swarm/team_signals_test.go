package swarm

import (
	"encoding/json"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestNormalizeCommandPayload_PreservesStructuredTeamAskPayload(t *testing.T) {
	raw, err := protocol.WrapSignalPayloadWithMeta(
		protocol.SourceKindInternalTool,
		"internal_tool.delegate_task",
		protocol.PayloadKindCommand,
		"run-1",
		"alpha",
		"soma-admin",
		[]byte(`{"schema_version":"v1","ask_kind":"research","goal":"Collect source evidence"}`),
	)
	if err != nil {
		t.Fatalf("wrap payload: %v", err)
	}

	got := normalizeCommandPayload(raw)
	var ask protocol.TeamAsk
	if err := json.Unmarshal(got, &ask); err != nil {
		t.Fatalf("decode normalized team ask: %v", err)
	}
	if ask.Goal != "Collect source evidence" {
		t.Fatalf("goal = %q", ask.Goal)
	}
	if ask.AskKind != protocol.TeamAskKindResearch {
		t.Fatalf("ask_kind = %q", ask.AskKind)
	}
}

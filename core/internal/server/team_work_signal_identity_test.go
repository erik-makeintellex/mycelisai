package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestProjectedResultEventIdentityIsStableAcrossReplay(t *testing.T) {
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	item := protocol.TeamWorkItem{
		TeamID:     "restart-safe-team",
		WorkItemID: "11111111-1111-1111-1111-111111111111",
		RunID:      "run-restart-1",
		State:      protocol.TeamWorkStateOutputReady,
	}
	env := protocol.SignalEnvelope{
		Meta: protocol.SignalMeta{
			Timestamp:     now,
			SourceKind:    protocol.SourceKindInternalTool,
			SourceChannel: "swarm.team.restart-safe-team.internal.trigger",
			PayloadKind:   protocol.PayloadKindResult,
			RunID:         item.RunID,
			TeamID:        item.TeamID,
		},
		Payload: json.RawMessage(`{"work_item_id":"11111111-1111-1111-1111-111111111111","idempotency_key":"confirm-action:proof-restart-1","summary":"Package ready"}`),
	}
	payload := map[string]any{
		"work_item_id":    item.WorkItemID,
		"idempotency_key": "confirm-action:proof-restart-1",
		"summary":         "Package ready",
	}

	subject := "swarm.team.restart-safe-team.signal.result"
	signalKey := projectedTeamSignalKey(subject, protocol.PayloadKindResult, payload, env.Payload)
	first := projectedTeamSignalReceiptID(item, signalKey)
	second := projectedTeamSignalReceiptID(item, signalKey)

	if first == "" {
		t.Fatal("projected result event identity must not be empty")
	}
	if first != second {
		t.Fatalf("replayed result changed identity: %q != %q", first, second)
	}
}

func expectProjectedSignalReceipt(mock sqlmock.Sqlmock, teamID, workID, sourceChannel string) {
	mock.ExpectExec("INSERT INTO team_signal_receipts").
		WithArgs(sqlmock.AnyArg(), teamID, workID, sqlmock.AnyArg(), sourceChannel).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

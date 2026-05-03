package conversations

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestLogTurn_NilToolArgs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(),
			"run-002",
			"sess-003",
			"default",
			"coder",
			"",
			0,
			"user",
			"do stuff",
			"",
			"",
			"",
			sqlmock.AnyArg(),
			"",
			"",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = store.LogTurn(context.Background(), protocol.ConversationTurnData{
		RunID:     "run-002",
		SessionID: "sess-003",
		TenantID:  "default",
		AgentID:   "coder",
		Role:      "user",
		Content:   "do stuff",
		ToolArgs:  nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogTurn_PopulatedToolArgs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(),
			"run-003",
			"sess-004",
			"default",
			"coder",
			"team-1",
			2,
			"tool_call",
			"calling read_file",
			"ollama",
			"qwen2.5",
			"read_file",
			[]byte(`{"path":"/workspace/foo.txt"}`),
			"",
			"",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = store.LogTurn(context.Background(), protocol.ConversationTurnData{
		RunID:      "run-003",
		SessionID:  "sess-004",
		TenantID:   "default",
		AgentID:    "coder",
		TeamID:     "team-1",
		TurnIndex:  2,
		Role:       "tool_call",
		Content:    "calling read_file",
		ProviderID: "ollama",
		ModelUsed:  "qwen2.5",
		ToolName:   "read_file",
		ToolArgs:   map[string]interface{}{"path": "/workspace/foo.txt"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

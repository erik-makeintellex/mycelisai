package conversations

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestLogTurn_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(),
			"run-001",
			"sess-001",
			"default",
			"admin",
			"admin-core",
			0,
			"user",
			"hello world",
			"ollama",
			"qwen2.5",
			"",
			sqlmock.AnyArg(),
			"",
			"",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := store.LogTurn(context.Background(), protocol.ConversationTurnData{
		RunID:      "run-001",
		SessionID:  "sess-001",
		TenantID:   "default",
		AgentID:    "admin",
		TeamID:     "admin-core",
		TurnIndex:  0,
		Role:       "user",
		Content:    "hello world",
		ProviderID: "ollama",
		ModelUsed:  "qwen2.5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogTurn_UUIDColumnsUseCasts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO conversation_turns
			(id, run_id, session_id, tenant_id, agent_id, team_id, turn_index, role, content,
			 provider_id, model_used, tool_name, tool_args, parent_turn_id, consultation_of, created_at)
		VALUES ($1::uuid, NULLIF($2,'')::uuid, $3::uuid, $4, $5, NULLIF($6,''), $7, $8, $9,
		        NULLIF($10,''), NULLIF($11,''), NULLIF($12,''), $13, NULLIF($14,'')::uuid, NULLIF($15,''), $16)
	`)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = store.LogTurn(context.Background(), protocol.ConversationTurnData{
		RunID:     "11111111-1111-1111-1111-111111111111",
		SessionID: "22222222-2222-2222-2222-222222222222",
		AgentID:   "admin",
		Role:      "user",
		Content:   "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogTurn_NilDB(t *testing.T) {
	store := NewStore(nil)

	_, err := store.LogTurn(context.Background(), protocol.ConversationTurnData{
		SessionID: "sess-001",
		AgentID:   "admin",
		Role:      "user",
		Content:   "hello",
	})
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
	if err.Error() != "conversations: database not available" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLogTurn_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err = store.LogTurn(context.Background(), protocol.ConversationTurnData{
		SessionID: "sess-001",
		TenantID:  "default",
		AgentID:   "admin",
		Role:      "user",
		Content:   "hello",
	})
	if err == nil {
		t.Fatal("expected error from DB failure")
	}
	if err.Error() != "conversations: persist failed: connection refused" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLogTurn_DefaultsTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(),
			"",
			"sess-002",
			"default",
			"agent-x",
			"",
			1,
			"assistant",
			"hi there",
			"",
			"",
			"",
			sqlmock.AnyArg(),
			"",
			"",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := store.LogTurn(context.Background(), protocol.ConversationTurnData{
		SessionID: "sess-002",
		TenantID:  "",
		AgentID:   "agent-x",
		TurnIndex: 1,
		Role:      "assistant",
		Content:   "hi there",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}
}

func TestLogTurn_DefaultsSessionID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(),
			"",
			sqlmock.AnyArg(),
			"default",
			"operator",
			"",
			0,
			"interjection",
			"hold on",
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
		AgentID: "operator",
		Role:    "interjection",
		Content: "hold on",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

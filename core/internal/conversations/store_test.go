package conversations

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

// ════════════════════════════════════════════════════════════════════
// LogTurn
// ════════════════════════════════════════════════════════════════════

func TestLogTurn_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(), // id (uuid)
			"run-001",        // run_id
			"sess-001",       // session_id
			"default",        // tenant_id
			"admin",          // agent_id
			"admin-core",     // team_id
			0,                // turn_index
			"user",           // role
			"hello world",    // content
			"ollama",         // provider_id
			"qwen2.5",        // model_used
			"",               // tool_name
			sqlmock.AnyArg(), // tool_args (nil []byte — driver may coerce)
			"",               // parent_turn_id
			"",               // consultation_of
			sqlmock.AnyArg(), // created_at
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

	// Expect tenant_id = "default" even though we pass empty string
	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(), // id
			"",               // run_id (empty)
			"sess-002",       // session_id
			"default",        // tenant_id — defaulted
			"agent-x",        // agent_id
			"",               // team_id
			1,                // turn_index
			"assistant",      // role
			"hi there",       // content
			"",               // provider_id
			"",               // model_used
			"",               // tool_name
			sqlmock.AnyArg(), // tool_args (nil []byte — driver may coerce)
			"",               // parent_turn_id
			"",               // consultation_of
			sqlmock.AnyArg(), // created_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := store.LogTurn(context.Background(), protocol.ConversationTurnData{
		SessionID: "sess-002",
		TenantID:  "", // empty — should default to "default"
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

func TestLogTurn_NilToolArgs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(), // id
			"run-002",        // run_id
			"sess-003",       // session_id
			"default",        // tenant_id
			"coder",          // agent_id
			"",               // team_id
			0,                // turn_index
			"user",           // role
			"do stuff",       // content
			"",               // provider_id
			"",               // model_used
			"",               // tool_name
			sqlmock.AnyArg(), // tool_args — nil []byte when ToolArgs is nil
			"",               // parent_turn_id
			"",               // consultation_of
			sqlmock.AnyArg(), // created_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = store.LogTurn(context.Background(), protocol.ConversationTurnData{
		RunID:     "run-002",
		SessionID: "sess-003",
		TenantID:  "default",
		AgentID:   "coder",
		Role:      "user",
		Content:   "do stuff",
		ToolArgs:  nil, // explicitly nil
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

	// When ToolArgs is populated, it should be JSON-marshaled
	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(
			sqlmock.AnyArg(), // id
			"run-003",        // run_id
			"sess-004",       // session_id
			"default",        // tenant_id
			"coder",          // agent_id
			"team-1",         // team_id
			2,                // turn_index
			"tool_call",      // role
			"calling read_file", // content
			"ollama",         // provider_id
			"qwen2.5",        // model_used
			"read_file",      // tool_name
			[]byte(`{"path":"/workspace/foo.txt"}`), // tool_args JSON
			"",               // parent_turn_id
			"",               // consultation_of
			sqlmock.AnyArg(), // created_at
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

// ════════════════════════════════════════════════════════════════════
// GetRunConversation
// ════════════════════════════════════════════════════════════════════

func turnColumns() []string {
	return []string{
		"id", "run_id", "session_id", "tenant_id", "agent_id",
		"team_id", "turn_index", "role", "content",
		"provider_id", "model_used",
		"tool_name", "tool_args",
		"parent_turn_id", "consultation_of",
		"created_at",
	}
}

func TestGetRunConversation_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
	now := time.Now()

	rows := sqlmock.NewRows(turnColumns()).
		AddRow("turn-1", "run-001", "sess-1", "default", "admin", "admin-core", 0, "user", "hello", "ollama", "qwen2.5", "", "", "", "", now).
		AddRow("turn-2", "run-001", "sess-1", "default", "admin", "admin-core", 1, "assistant", "hi back", "ollama", "qwen2.5", "", "", "", "", now)

	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE run_id = \\$1 ORDER BY").
		WithArgs("run-001").
		WillReturnRows(rows)

	turns, err := store.GetRunConversation(context.Background(), "run-001", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	if turns[0].ID != "turn-1" {
		t.Errorf("expected turn-1, got %s", turns[0].ID)
	}
	if turns[1].Role != "assistant" {
		t.Errorf("expected assistant role, got %s", turns[1].Role)
	}
	if turns[0].AgentID != "admin" {
		t.Errorf("expected agent_id=admin, got %s", turns[0].AgentID)
	}
}

func TestGetRunConversation_WithAgentFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
	now := time.Now()

	rows := sqlmock.NewRows(turnColumns()).
		AddRow("turn-3", "run-001", "sess-1", "default", "coder", "council-core", 0, "user", "write code", "", "", "", "", "", "", now)

	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE run_id = \\$1 AND agent_id = \\$2 ORDER BY").
		WithArgs("run-001", "coder").
		WillReturnRows(rows)

	turns, err := store.GetRunConversation(context.Background(), "run-001", "coder")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	if turns[0].AgentID != "coder" {
		t.Errorf("expected agent_id=coder, got %s", turns[0].AgentID)
	}
}

func TestGetRunConversation_NilDB(t *testing.T) {
	store := NewStore(nil)

	_, err := store.GetRunConversation(context.Background(), "run-001", "")
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
	if err.Error() != "conversations: database not available" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetRunConversation_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	rows := sqlmock.NewRows(turnColumns()) // no rows
	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE run_id = \\$1 ORDER BY").
		WithArgs("run-nonexistent").
		WillReturnRows(rows)

	turns, err := store.GetRunConversation(context.Background(), "run-nonexistent", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if turns == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(turns) != 0 {
		t.Errorf("expected 0 turns, got %d", len(turns))
	}
}

// ════════════════════════════════════════════════════════════════════
// GetSessionTurns
// ════════════════════════════════════════════════════════════════════

func TestGetSessionTurns_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
	now := time.Now()

	rows := sqlmock.NewRows(turnColumns()).
		AddRow("turn-10", "", "sess-standalone", "default", "architect", "", 0, "user", "design this", "", "", "", "", "", "", now).
		AddRow("turn-11", "", "sess-standalone", "default", "architect", "", 1, "assistant", "here is the design", "ollama", "qwen2.5", "", "", "", "", now)

	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE session_id = \\$1 ORDER BY").
		WithArgs("sess-standalone").
		WillReturnRows(rows)

	turns, err := store.GetSessionTurns(context.Background(), "sess-standalone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	if turns[0].SessionID != "sess-standalone" {
		t.Errorf("expected session_id=sess-standalone, got %s", turns[0].SessionID)
	}
	if turns[1].Content != "here is the design" {
		t.Errorf("unexpected content: %s", turns[1].Content)
	}
}

func TestGetSessionTurns_NilDB(t *testing.T) {
	store := NewStore(nil)

	_, err := store.GetSessionTurns(context.Background(), "sess-001")
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
	if err.Error() != "conversations: database not available" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetSessionTurns_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	rows := sqlmock.NewRows(turnColumns()) // no rows
	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE session_id = \\$1 ORDER BY").
		WithArgs("sess-ghost").
		WillReturnRows(rows)

	turns, err := store.GetSessionTurns(context.Background(), "sess-ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if turns == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(turns) != 0 {
		t.Errorf("expected 0 turns, got %d", len(turns))
	}
}

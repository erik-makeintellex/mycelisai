package conversations

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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

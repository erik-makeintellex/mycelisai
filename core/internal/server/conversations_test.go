package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/internal/events"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// ── Shared helpers for conversation handler tests ────────────────

// withConversations creates a conversations.Store backed by sqlmock.
func withConversations(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	store := conversations.NewStore(db)
	return func(s *AdminServer) {
		s.Conversations = store
	}, mock
}

// withNATS starts an embedded NATS server and wires NC into AdminServer.
func withNATS(t *testing.T) func(*AdminServer) {
	t.Helper()
	opts := &natsserver.Options{Host: "127.0.0.1", Port: -1}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		t.Fatalf("nats connect: %v", err)
	}
	t.Cleanup(func() {
		nc.Close()
		srv.Shutdown()
	})
	return func(s *AdminServer) {
		s.NC = nc
	}
}

// convTurnColumns returns the column names returned by conversation_turns SELECT queries.
func convTurnColumns() []string {
	return []string{
		"id", "run_id", "session_id", "tenant_id", "agent_id",
		"team_id", "turn_index", "role", "content",
		"provider_id", "model_used",
		"tool_name", "tool_args",
		"parent_turn_id", "consultation_of",
		"created_at",
	}
}

// ════════════════════════════════════════════════════════════════════
// HandleGetRunConversation — GET /api/v1/runs/{id}/conversation
// ════════════════════════════════════════════════════════════════════

func TestHandleGetRunConversation_HappyPath(t *testing.T) {
	convOpt, mock := withConversations(t)
	s := newTestServer(convOpt)
	now := time.Now()

	rows := sqlmock.NewRows(convTurnColumns()).
		AddRow("t-1", "run-100", "sess-1", "default", "admin", "admin-core", 0, "user", "hello", "ollama", "qwen2.5", "", "", "", "", now).
		AddRow("t-2", "run-100", "sess-1", "default", "admin", "admin-core", 1, "assistant", "hi", "ollama", "qwen2.5", "", "", "", "", now)

	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE run_id = \\$1 ORDER BY").
		WithArgs("run-100").
		WillReturnRows(rows)

	mux := setupMux(t, "GET /api/v1/runs/{id}/conversation", s.HandleGetRunConversation)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/run-100/conversation", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["run_id"] != "run-100" {
		t.Errorf("expected run_id=run-100, got %v", data["run_id"])
	}
	turns, ok := data["turns"].([]interface{})
	if !ok {
		t.Fatalf("expected turns array, got %T", data["turns"])
	}
	if len(turns) != 2 {
		t.Errorf("expected 2 turns, got %d", len(turns))
	}
}

func TestHandleGetRunConversation_MissingRunID(t *testing.T) {
	convOpt, _ := withConversations(t)
	s := newTestServer(convOpt)

	// Use a route without {id} so PathValue returns ""
	mux := setupMux(t, "GET /api/v1/runs/{id}/conversation", s.HandleGetRunConversation)
	// An empty path segment will cause PathValue("id") to be ""
	rr := doRequest(t, mux, "GET", "/api/v1/runs//conversation", "")

	// Go 1.22+ mux won't match a pattern with empty path value — it returns 404
	// So we test with a direct handler call instead
	rr = doRequest(t, http.HandlerFunc(s.HandleGetRunConversation), "GET", "/api/v1/runs/conversation", "")
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleGetRunConversation_NilConversations(t *testing.T) {
	s := newTestServer() // no Conversations wired

	mux := setupMux(t, "GET /api/v1/runs/{id}/conversation", s.HandleGetRunConversation)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/run-100/conversation", "")

	assertStatus(t, rr, http.StatusServiceUnavailable)
}

// ════════════════════════════════════════════════════════════════════
// HandleGetSessionConversation — GET /api/v1/conversations/{session_id}
// ════════════════════════════════════════════════════════════════════

func TestHandleGetSessionConversation_HappyPath(t *testing.T) {
	convOpt, mock := withConversations(t)
	s := newTestServer(convOpt)
	now := time.Now()

	rows := sqlmock.NewRows(convTurnColumns()).
		AddRow("t-10", "", "sess-standalone", "default", "architect", "", 0, "user", "design this", "", "", "", "", "", "", now)

	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE session_id = \\$1 ORDER BY").
		WithArgs("sess-standalone").
		WillReturnRows(rows)

	mux := setupMux(t, "GET /api/v1/conversations/{session_id}", s.HandleGetSessionConversation)
	rr := doRequest(t, mux, "GET", "/api/v1/conversations/sess-standalone", "")

	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["session_id"] != "sess-standalone" {
		t.Errorf("expected session_id=sess-standalone, got %v", data["session_id"])
	}
}

func TestHandleGetSessionConversation_MissingSessionID(t *testing.T) {
	convOpt, _ := withConversations(t)
	s := newTestServer(convOpt)

	// Direct handler call — no mux means PathValue("session_id") returns ""
	rr := doRequest(t, http.HandlerFunc(s.HandleGetSessionConversation), "GET", "/api/v1/conversations/", "")
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleGetSessionConversation_NilConversations(t *testing.T) {
	s := newTestServer() // no Conversations wired

	mux := setupMux(t, "GET /api/v1/conversations/{session_id}", s.HandleGetSessionConversation)
	rr := doRequest(t, mux, "GET", "/api/v1/conversations/sess-001", "")

	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestHandleGetRunConversation_SynthesizesFromEventsWhenTurnsEmpty(t *testing.T) {
	convOpt, convMock := withConversations(t)
	evDB, evMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { evDB.Close() })
	s := newTestServer(convOpt)
	s.Events = events.NewStore(evDB, nil)
	now := time.Now()

	convMock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE run_id = \\$1 ORDER BY").
		WithArgs("run-events-1").
		WillReturnRows(sqlmock.NewRows(convTurnColumns()))

	evRows := sqlmock.NewRows([]string{
		"id", "run_id", "tenant_id", "event_type", "severity",
		"source_agent", "source_team", "payload", "audit_event_id", "emitted_at",
	}).
		AddRow("ev-1", "run-events-1", "default", "tool.completed", "info", "soma", "market-research-team", `{"team_id":"market-research-team","team_name":"Market Research Team"}`, "", now).
		AddRow("ev-2", "run-events-1", "default", "artifact.created", "info", "admin", "governance", `{"package_kind":"team_execution_bundle","base_path":"deliverables/teams/market-research-team"}`, "", now).
		AddRow("ev-3", "run-events-1", "default", "mission.completed", "info", "admin", "governance", `{"execution_state":"verified"}`, "", now)
	evMock.ExpectQuery("SELECT .+ FROM mission_events WHERE run_id = \\$1 ORDER BY").
		WithArgs("run-events-1").
		WillReturnRows(evRows)

	mux := setupMux(t, "GET /api/v1/runs/{id}/conversation", s.HandleGetRunConversation)
	rr := doRequest(t, mux, "GET", "/api/v1/runs/run-events-1/conversation", "")
	assertStatus(t, rr, http.StatusOK)

	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	data, _ := resp["data"].(map[string]interface{})
	turns, _ := data["turns"].([]interface{})
	if len(turns) != 3 {
		t.Fatalf("expected 3 synthesized turns from mission_events, got %d (%+v)", len(turns), turns)
	}
}

// ════════════════════════════════════════════════════════════════════
// HandleRunInterject — POST /api/v1/runs/{id}/interject
// ════════════════════════════════════════════════════════════════════

package server

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/conversations"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var serverTestNATSPort int32 = 15100 + int32(time.Now().UnixNano()%1000)
var serverTestNATSClientPort int32 = 25000 + int32(time.Now().UnixNano()%5000)

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
	opts := &natsserver.Options{Host: "127.0.0.1", Port: nextServerTestNATSPort(t)}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	nc, err := nats.Connect(srv.ClientURL(), nats.Dialer(&net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: nextServerTestNATSClientPort(t),
		},
		Timeout: 2 * time.Second,
	}))
	if err != nil {
		srv.Shutdown()
		t.Fatalf("nats connect: %v", err)
	}
	t.Cleanup(func() {
		nc.Close()
		srv.Shutdown()
		srv.WaitForShutdown()
	})
	return func(s *AdminServer) {
		s.NC = nc
	}
}

func nextServerTestNATSPort(t *testing.T) int {
	return nextServerTestTCPPort(t, &serverTestNATSPort)
}

func nextServerTestNATSClientPort(t *testing.T) int {
	return nextServerTestTCPPort(t, &serverTestNATSClientPort)
}

func nextServerTestTCPPort(t *testing.T, counter *int32) int {
	t.Helper()
	for range 200 {
		port := int(atomic.AddInt32(counter, 1))
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			_ = ln.Close()
			return port
		}
	}
	t.Fatal("no available low NATS test port")
	return 0
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

// ════════════════════════════════════════════════════════════════════
// HandleRunInterject — POST /api/v1/runs/{id}/interject
// ════════════════════════════════════════════════════════════════════

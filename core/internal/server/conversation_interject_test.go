package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func testInterjectSoma() *swarm.Soma {
	return swarm.NewTestSoma([]*swarm.TeamManifest{
		{
			ID:   "admin-core",
			Name: "Admin",
			Type: swarm.TeamTypeAction,
			Members: []protocol.AgentManifest{
				{ID: "admin", Role: "admin"},
			},
		},
		{
			ID:   "council-core",
			Name: "Council",
			Type: swarm.TeamTypeAction,
			Members: []protocol.AgentManifest{
				{ID: "council-architect", Role: "architect"},
				{ID: "council-coder", Role: "coder"},
			},
		},
	})
}

func TestHandleRunInterject_TargetedAgent(t *testing.T) {
	natsOpt := withNATS(t)
	convOpt, mock := withConversations(t)
	s := newTestServer(natsOpt, convOpt, func(s *AdminServer) {
		s.Soma = testInterjectSoma()
	})
	mock.ExpectExec("INSERT INTO conversation_turns").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"message": "stop what you are doing", "agent_id": "admin"}`
	mux := setupMux(t, "POST /api/v1/runs/{id}/interject", s.HandleRunInterject)
	rr := doRequest(t, mux, "POST", "/api/v1/runs/run-200/interject", body)

	assertStatus(t, rr, http.StatusOK)
	assertAgentsReached(t, rr, 1)
}

func TestHandleRunInterject_Subtests(t *testing.T) {
	natsOpt := withNATS(t)

	t.Run("Broadcast", func(t *testing.T) {
		convOpt, mock := withConversations(t)
		s := newTestServer(natsOpt, convOpt, func(s *AdminServer) {
			s.Soma = testInterjectSoma()
		})
		mock.ExpectExec("INSERT INTO conversation_turns").WillReturnResult(sqlmock.NewResult(1, 1))

		body := `{"message": "attention everyone"}`
		mux := setupMux(t, "POST /api/v1/runs/{id}/interject", s.HandleRunInterject)
		rr := doRequest(t, mux, "POST", "/api/v1/runs/run-200/interject", body)

		assertStatus(t, rr, http.StatusOK)
		assertAgentsReached(t, rr, 3)
	})

	t.Run("MissingRunID", func(t *testing.T) {
		s := newTestServer(natsOpt)
		rr := doRequest(t, http.HandlerFunc(s.HandleRunInterject), "POST", "/api/v1/runs/interject", `{"message":"hi"}`)
		assertStatus(t, rr, http.StatusBadRequest)
	})

	t.Run("EmptyMessage", func(t *testing.T) {
		s := newTestServer(natsOpt)
		mux := setupMux(t, "POST /api/v1/runs/{id}/interject", s.HandleRunInterject)
		rr := doRequest(t, mux, "POST", "/api/v1/runs/run-200/interject", `{"message": ""}`)
		assertStatus(t, rr, http.StatusBadRequest)
	})
}

func TestHandleRunInterject_NilNC(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/runs/{id}/interject", s.HandleRunInterject)
	rr := doRequest(t, mux, "POST", "/api/v1/runs/run-200/interject", `{"message": "please respond"}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
}

func assertAgentsReached(t *testing.T, rr *httptest.ResponseRecorder, want float64) {
	t.Helper()
	var resp map[string]interface{}
	assertJSON(t, rr, &resp)
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got %v", resp["ok"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	reached, ok := data["agents_reached"].(float64)
	if !ok {
		t.Fatalf("expected agents_reached number, got %T", data["agents_reached"])
	}
	if reached != want {
		t.Errorf("expected agents_reached=%v, got %v", want, reached)
	}
}

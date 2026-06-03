package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_PersistsSelectedTeamSessionConversation(t *testing.T) {
	wireNATS := withNATS(t)
	convOpt, mock := withConversations(t)
	s := newTestServer(wireNATS, convOpt)
	s.Cognitive = &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{
				"mock": {Type: "mock", Enabled: true, ModelID: "test-model"},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"mock": cognitiveTestProvider{},
		},
	}

	sessionID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE session_id = \\$1 ORDER BY").
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows(convTurnColumns()))
	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(sqlmock.AnyArg(), "", sessionID, "default", "admin", "marketing-team", 0, "user", "What is this team's launch focus?", "", "", "", sqlmock.AnyArg(), "", "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(sqlmock.AnyArg(), "", sessionID, "default", "admin", "marketing-team", 1, "assistant", "Marketing is focused on launch planning.", "mock", "test-model", "", sqlmock.AnyArg(), "", "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	subject := "swarm.council.admin.request"
	_, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		resp, _ := json.Marshal(map[string]any{
			"text":        "Marketing is focused on launch planning.",
			"provider_id": "mock",
			"model_used":  "test-model",
		})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"session_id":"22222222-2222-2222-2222-222222222222","team_id":"marketing-team","messages":[{"role":"user","content":"What is this team's launch focus?"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("conversation expectations: %v", err)
	}
}

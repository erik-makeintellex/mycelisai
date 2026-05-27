package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/nats-io/nats.go"
)

func TestHandleChat_PrependsWorkspaceContextForSelectedTeam(t *testing.T) {
	wireNATS := withNATS(t)
	s := newTestServer(wireNATS)
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
	s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:      "org-123",
			Name:    "Northstar Labs",
			Purpose: "Launch the new product line.",
		},
		Departments: []OrganizationDepartmentSummary{
			{ID: "marketing-team", Name: "Marketing"},
			{ID: "product-team", Name: "Product"},
		},
	})

	subject := "swarm.council.admin.request"
	forwarded := make(chan []chatRequestMessage, 1)
	_, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		var turns []chatRequestMessage
		if err := json.Unmarshal(msg.Data, &turns); err != nil {
			t.Errorf("decode forwarded messages: %v", err)
		} else {
			forwarded <- turns
		}
		resp, _ := json.Marshal(map[string]any{
			"text": "Marketing is focused on launch planning.",
		})
		msg.Respond(resp)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"Do you see it?"}],"organization_id":"org-123","team_id":"marketing-team","team_name":"Marketing"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	select {
	case turns := <-forwarded:
		if len(turns) < 3 {
			t.Fatalf("forwarded turn count = %d, want at least 3", len(turns))
		}
		if turns[0].Role != "system" {
			t.Fatalf("attunement role = %q, want system", turns[0].Role)
		}
		if !strings.Contains(turns[0].Content, somaAttunementContextHeader) {
			t.Fatalf("attunement content = %q, want attunement marker", turns[0].Content)
		}
		if turns[1].Role != "user" {
			t.Fatalf("workspace role = %q, want user", turns[1].Role)
		}
		if !strings.Contains(turns[1].Content, "[WORKSPACE CONTEXT]") {
			t.Fatalf("workspace content = %q, want workspace context marker", turns[1].Content)
		}
		if !strings.Contains(turns[1].Content, "Organization: Northstar Labs.") {
			t.Fatalf("workspace content missing organization: %q", turns[1].Content)
		}
		if !strings.Contains(turns[1].Content, "Visible departments/teams in this organization: Marketing and Product.") {
			t.Fatalf("workspace content missing department summary: %q", turns[1].Content)
		}
		if !strings.Contains(turns[1].Content, "Current team focus: Marketing (id: marketing-team).") {
			t.Fatalf("workspace content missing current team: %q", turns[1].Content)
		}
		last := turns[len(turns)-1].Content
		if !strings.Contains(last, directAnswerRoutePrefix) {
			t.Fatalf("last user content missing direct answer route hint: %q", last)
		}
		if !strings.Contains(last, "Original request:\nDo you see it?") {
			t.Fatalf("last user content missing preserved request: %q", last)
		}
	default:
		t.Fatal("expected forwarded messages to be captured")
	}
}

func TestHandleChat_ReplaysAndPersistsSessionConversation(t *testing.T) {
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

	sessionID := "11111111-1111-1111-1111-111111111111"
	now := time.Now()
	rows := sqlmock.NewRows(convTurnColumns()).
		AddRow("turn-1", "", sessionID, "default", "admin", "admin-core", 0, "user", "Remember the launch marker.", "", "", "", "", "", "", now).
		AddRow("turn-2", "", sessionID, "default", "admin", "admin-core", 1, "assistant", "I have the launch marker in this session context.", "mock", "test-model", "", "", "", "", now)
	mock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE session_id = \\$1 ORDER BY").
		WithArgs(sessionID).
		WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(sqlmock.AnyArg(), "", sessionID, "default", "admin", "admin-core", 2, "user", "What was the launch marker?", "", "", "", sqlmock.AnyArg(), "", "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO conversation_turns").
		WithArgs(sqlmock.AnyArg(), "", sessionID, "default", "admin", "admin-core", 3, "assistant", "The prior session marker was launch.", "mock", "test-model", "", sqlmock.AnyArg(), "", "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	subject := "swarm.council.admin.request"
	forwarded := make(chan []chatRequestMessage, 1)
	_, err := s.NC.Subscribe(subject, func(msg *nats.Msg) {
		var turns []chatRequestMessage
		if err := json.Unmarshal(msg.Data, &turns); err != nil {
			t.Errorf("decode forwarded messages: %v", err)
		} else {
			forwarded <- turns
		}
		resp, _ := json.Marshal(map[string]any{
			"text":        "The prior session marker was launch.",
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

	reqBody := bytes.NewBufferString(`{"session_id":"11111111-1111-1111-1111-111111111111","messages":[{"role":"user","content":"What was the launch marker?"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	select {
	case turns := <-forwarded:
		if len(turns) != 4 {
			t.Fatalf("forwarded turn count = %d, want 4", len(turns))
		}
		if turns[0].Role != "system" || !strings.Contains(turns[0].Content, somaAttunementContextHeader) {
			t.Fatalf("first forwarded turn = %#v, want system attunement", turns[0])
		}
		if turns[1].Content != "Remember the launch marker." {
			t.Fatalf("second forwarded content = %q", turns[1].Content)
		}
		if turns[2].Role != "assistant" || turns[2].Content != "I have the launch marker in this session context." {
			t.Fatalf("third forwarded turn = %#v", turns[2])
		}
		if !strings.Contains(turns[3].Content, directAnswerRoutePrefix) {
			t.Fatalf("latest forwarded content missing direct-answer route: %q", turns[3].Content)
		}
		if !strings.Contains(turns[3].Content, "Original request:\nWhat was the launch marker?") {
			t.Fatalf("latest forwarded content missing original request: %q", turns[3].Content)
		}
	default:
		t.Fatal("expected forwarded messages to be captured")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("conversation expectations: %v", err)
	}
}

func TestHandleChat_RejectsInvalidSessionID(t *testing.T) {
	s := newTestServer()
	reqBody := bytes.NewBufferString(`{"session_id":"not-a-uuid","messages":[{"role":"user","content":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

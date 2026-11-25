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
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func setupMatrixTestServer(t *testing.T, onAdminRequest func(turns []chatRequestMessage) map[string]any) (*AdminServer, sqlmock.Sqlmock) {
	t.Helper()
	wireNATS := withNATS(t)
	dbOpt, mock := withDB(t)
	convOpt, _ := withConversations(t)
	mock.MatchExpectationsInOrder(false)
	for i := 0; i < 64; i++ {
		mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO intent_proofs").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO confirm_tokens").WillReturnResult(sqlmock.NewResult(0, 1))
	}

	s := newTestServer(wireNATS, dbOpt, convOpt)
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
	s.Soma = swarm.NewTestSoma([]*swarm.TeamManifest{
		{
			ID:   "alpha-team",
			Name: "Alpha Team",
			Members: []protocol.AgentManifest{
				{ID: "alpha-lead"},
			},
		},
	})
	_, err := s.NC.Subscribe("swarm.council.admin.request", func(msg *nats.Msg) {
		var turns []chatRequestMessage
		_ = json.Unmarshal(msg.Data, &turns)
		reply := map[string]any{
			"text":        "Soma processed the request cleanly.",
			"provider_id": "mock",
			"model_used":  "test-model",
		}
		if onAdminRequest != nil {
			reply = onAdminRequest(turns)
		}
		raw, _ := json.Marshal(reply)
		_ = msg.Respond(raw)
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := s.NC.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	return s, mock
}

func executeSomaChatMatrixTurn(t *testing.T, s *AdminServer, sessionID string, messages []chatRequestMessage) (protocol.CTSEnvelope, protocol.ChatResponsePayload) {
	t.Helper()
	payloadMap := map[string]any{"messages": messages}
	if sessionID != "" {
		payloadMap["session_id"] = sessionID
	}
	body, err := json.Marshal(payloadMap)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode APIResponse: %v", err)
	}
	if !resp.OK {
		t.Fatalf("resp.OK = false, error = %q", resp.Error)
	}
	data, _ := json.Marshal(resp.Data)
	var env protocol.CTSEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if strings.Contains(strings.ToLower(payload.Text), "can't deliver") || strings.Contains(strings.ToLower(payload.Text), "cannot deliver") {
		t.Fatalf("unexpected refusal text in payload: %q", payload.Text)
	}
	return env, payload
}

func TestSomaInteractionMatrix_Category1_CasualEverydayBuildRequests(t *testing.T) {
	prompts := []string{
		"make me a quick timer",
		"can you build a todo list?",
		"let's build a habit tracker",
		"please create a landing page for our startup",
		"help me make a calculator",
		"i'd like a simple kanban board",
		"could you create a weather widget?",
		"whip up a quick note taking tool",
	}
	s, _ := setupMatrixTestServer(t, nil)
	for _, prompt := range prompts {
		t.Run(prompt, func(t *testing.T) {
			env, payload := executeSomaChatMatrixTurn(t, s, "", []chatRequestMessage{{Role: "user", Content: prompt}})
			if env.Mode != protocol.ModeProposal {
				t.Fatalf("mode = %q, want proposal for %q", env.Mode, prompt)
			}
			if payload.Proposal == nil || strings.TrimSpace(payload.Proposal.ConfirmToken) == "" {
				t.Fatalf("expected proposal with confirm_token for %q", prompt)
			}
		})
	}
}

func TestSomaInteractionMatrix_Category2_SimpleGameAndArcadeRequests(t *testing.T) {
	prompts := []string{
		"make a snake game",
		"build a tic tac toe game",
		"create a memory game",
		"create a playable puzzle game",
		"put a team together to create a Street Fighter II Tetris game",
	}
	s, _ := setupMatrixTestServer(t, nil)
	for _, prompt := range prompts {
		t.Run(prompt, func(t *testing.T) {
			env, payload := executeSomaChatMatrixTurn(t, s, "", []chatRequestMessage{{Role: "user", Content: prompt}})
			if env.Mode != protocol.ModeProposal || payload.Proposal == nil {
				t.Fatalf("expected proposal mode for %q, got %q", prompt, env.Mode)
			}
			if len(payload.Proposal.AffectedResources) == 0 || !strings.Contains(strings.Join(payload.Proposal.AffectedResources, " "), "index.html") {
				t.Fatalf("expected HTML5 game deliverables in affected resources for %q: %v", prompt, payload.Proposal.AffectedResources)
			}
		})
	}
}

func TestSomaInteractionMatrix_Category3_DocumentsScriptsChecklistsAndRunbooks(t *testing.T) {
	prompts := []string{
		"create a markdown document for our Q3 release notes",
		"write a python script to clean up temp logs",
		"generate a deployment checklist",
		"draft an incident response runbook",
		"create a project roadmap document",
	}
	s, _ := setupMatrixTestServer(t, nil)
	for _, prompt := range prompts {
		t.Run(prompt, func(t *testing.T) {
			env, payload := executeSomaChatMatrixTurn(t, s, "", []chatRequestMessage{{Role: "user", Content: prompt}})
			if env.Mode != protocol.ModeProposal || payload.Proposal == nil {
				t.Fatalf("expected proposal mode for %q, got %q", prompt, env.Mode)
			}
		})
	}
}

func TestSomaInteractionMatrix_Category4_TeamsMultiAppAndMCPAdditions(t *testing.T) {
	prompts := []string{
		"create a team to build an inventory service",
		"put a team together to try and find out what's going on with market competitors and produce a research report",
		"Build a storefront app and have a team work on another application for fulfillment that interacts with the initial ask",
		"Add Standard Deployment MCP Filesystem (mcp_id: filesystem, mcp_group: standard, mcp_title: Filesystem) to team Alpha (team_id: alpha)",
		"Plan for team Alpha (team_id: alpha) and execute next steps",
	}
	s, _ := setupMatrixTestServer(t, nil)
	for _, prompt := range prompts {
		t.Run(prompt, func(t *testing.T) {
			env, payload := executeSomaChatMatrixTurn(t, s, "", []chatRequestMessage{{Role: "user", Content: prompt}})
			if env.Mode != protocol.ModeProposal || payload.Proposal == nil {
				t.Fatalf("expected proposal mode for %q, got %q", prompt, env.Mode)
			}
			for _, res := range payload.Proposal.AffectedResources {
				if strings.Contains(res, "starter-web-app") {
					t.Fatalf("team prompt %q must not fall back to starter-web-app, got affected resources: %v", prompt, payload.Proposal.AffectedResources)
				}
			}
		})
	}
}

func TestSomaInteractionMatrix_Category5_RuntimePlaybookSearchAndPoisonedHistoryRecovery(t *testing.T) {
	directPrompts := []string{
		"state?",
		"what's the current state of mycelis?",
		"what's been delivered through testing lately?",
		"what teams currently exist?",
		"are you able to make web requests?",
		"consult the architect and sentry on security boundaries",
		"audit our current trusted memory and governed doctrine",
	}
	s, _ := setupMatrixTestServer(t, nil)
	for _, prompt := range directPrompts {
		t.Run(prompt, func(t *testing.T) {
			env, payload := executeSomaChatMatrixTurn(t, s, "", []chatRequestMessage{{Role: "user", Content: prompt}})
			if env.Mode != protocol.ModeAnswer {
				t.Fatalf("expected answer mode for %q, got %q", prompt, env.Mode)
			}
			if strings.TrimSpace(payload.Text) == "" {
				t.Fatalf("expected non-empty direct answer for %q", prompt)
			}
		})
	}

	t.Run("multi-turn poisoned cant deliver history recovery", func(t *testing.T) {
		var capturedTurns []chatRequestMessage
		sPoisoned, _ := setupMatrixTestServer(t, func(turns []chatRequestMessage) map[string]any {
			capturedTurns = turns
			return map[string]any{
				"text":        "Here is the clean architecture overview.",
				"provider_id": "mock",
				"model_used":  "test-model",
			}
		})
		sessionID := "22222222-2222-2222-2222-222222222222"
		now := time.Now()
		_, convMock := withConversations(t)
		convOpt, convMock := withConversations(t)
		convOpt(sPoisoned)
		rows := sqlmock.NewRows(convTurnColumns()).
			AddRow("t-1", "", sessionID, "default", "admin", "admin-core", 0, "user", "Healthy turn", "", "", "", "", "", "", now).
			AddRow("t-2", "", sessionID, "default", "admin", "admin-core", 1, "assistant", "Healthy response", "mock", "test-model", "", "", "", "", now).
			AddRow("t-3", "", sessionID, "default", "admin", "admin-core", 2, "user", "Build something earlier", "", "", "", "", "", "", now).
			AddRow("t-4", "", sessionID, "default", "admin", "admin-core", 3, "assistant", "I'm sorry, I can't deliver that request right now.", "mock", "test-model", "", "", "", "", now)
		convMock.ExpectQuery("SELECT .+ FROM conversation_turns WHERE session_id = \\$1 ORDER BY").
			WithArgs(sessionID).
			WillReturnRows(rows)
		convMock.ExpectExec("INSERT INTO conversation_turns").WillReturnResult(sqlmock.NewResult(1, 1))
		convMock.ExpectExec("INSERT INTO conversation_turns").WillReturnResult(sqlmock.NewResult(1, 1))

		env, payload := executeSomaChatMatrixTurn(t, sPoisoned, sessionID, []chatRequestMessage{
			{Role: "user", Content: "How does the council review proposals?"},
		})
		if env.Mode != protocol.ModeAnswer || !strings.Contains(payload.Text, "clean architecture overview") {
			t.Fatalf("unexpected direct recovery result: mode=%q text=%q", env.Mode, payload.Text)
		}
		for _, turn := range capturedTurns {
			if strings.Contains(strings.ToLower(turn.Content), "can't deliver") || strings.Contains(turn.Content, "Build something earlier") {
				t.Fatalf("expected poisoned assistant+user pair to be stripped from session history, got turns=%+v", capturedTurns)
			}
		}

		poisonedBuild := []chatRequestMessage{
			{Role: "user", Content: "Make a widget"},
			{Role: "assistant", Content: "I'm sorry, I can't deliver that request right now."},
			{Role: "user", Content: "make me a quick timer"},
		}
		buildEnv, buildPayload := executeSomaChatMatrixTurn(t, sPoisoned, "", poisonedBuild)
		if buildEnv.Mode != protocol.ModeProposal || buildPayload.Proposal == nil || buildPayload.Proposal.ConfirmToken == "" {
			t.Fatalf("expected proposal recovery after poisoned turn, got mode=%q", buildEnv.Mode)
		}
	})
}

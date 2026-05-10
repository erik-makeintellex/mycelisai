package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/pkg/protocol"
)

func TestRespondSearchChatPayload_DirectSearchIncludesCompletedExecutionSummary(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", nil)
	rr := httptest.NewRecorder()

	s.respondSearchChatPayload(
		rr,
		req,
		"Direct web search",
		"latest news updates regarding ai agent products",
		"I used web_search through local_api. Current results:\n1. Agent product release",
		[]string{"web_search"},
		protocol.ExecutionStatusCompleted,
		"",
	)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromAPIResponse(t, rr)
	if !containsString(payload.ToolsUsed, "web_search") {
		t.Fatalf("tools_used = %v, want web_search", payload.ToolsUsed)
	}
	if !strings.Contains(payload.Text, "Agent product release") {
		t.Fatalf("payload.text = %q, want search result", payload.Text)
	}
	if payload.ExecutionSummary == nil {
		t.Fatal("expected execution_summary")
	}
	if payload.ExecutionSummary.Execution.Shape != protocol.ExecutionShapeToolAssistedWork {
		t.Fatalf("execution_summary.execution.shape = %q", payload.ExecutionSummary.Execution.Shape)
	}
	if payload.ExecutionSummary.Execution.Status != protocol.ExecutionStatusCompleted {
		t.Fatalf("execution_summary.execution.status = %q", payload.ExecutionSummary.Execution.Status)
	}
	if payload.ExecutionSummary.Proof.RunClass != protocol.ExecutionRunClassNoRun || payload.ExecutionSummary.Proof.NoRunReason == "" {
		t.Fatalf("execution_summary.proof = %+v", payload.ExecutionSummary.Proof)
	}
}

func TestHandleChat_DirectSearchBlockerHasBlockedExecutionSummary(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"look up recent Mycelis release notes"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromAPIResponse(t, rr)
	if !containsString(payload.ToolsUsed, "web_search") {
		t.Fatalf("tools_used = %v, want web_search", payload.ToolsUsed)
	}
	if payload.ExecutionSummary == nil {
		t.Fatal("expected execution_summary")
	}
	if payload.ExecutionSummary.Execution.Shape != protocol.ExecutionShapeToolAssistedWork {
		t.Fatalf("execution_summary.execution.shape = %q", payload.ExecutionSummary.Execution.Shape)
	}
	if payload.ExecutionSummary.Execution.Status != protocol.ExecutionStatusBlocked {
		t.Fatalf("execution_summary.execution.status = %q", payload.ExecutionSummary.Execution.Status)
	}
	if payload.ExecutionSummary.Proof.RunClass != protocol.ExecutionRunClassNoRun {
		t.Fatalf("execution_summary.proof.run_class = %q", payload.ExecutionSummary.Proof.RunClass)
	}
}

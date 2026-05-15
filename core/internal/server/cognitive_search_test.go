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
		"Notice: web_search via local_api; no confirmation; external results are leads, verify before relying.\nResults:\n1. Agent product release",
		[]string{"web_search"},
		protocol.ExecutionStatusCompleted,
		"",
		nil,
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
	if !strings.Contains(payload.Text, "no confirmation") {
		t.Fatalf("payload.text = %q, want no-confirm disclosure", payload.Text)
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
	if payload.ExecutionSummary.AuditRecovery.Degradation == nil {
		t.Fatal("expected degradation metadata for blocked search")
	}
	if payload.ExecutionSummary.AuditRecovery.Degradation.Code != "search_provider_disabled" {
		t.Fatalf("degradation.code = %q", payload.ExecutionSummary.AuditRecovery.Degradation.Code)
	}
	if !payload.ExecutionSummary.AuditRecovery.Degradation.RequiresAttention {
		t.Fatal("expected degradation to require operator attention")
	}
}

func TestDirectSearchNoticeNamesLocalSourcesAsRetainedContext(t *testing.T) {
	resp := searchcap.Response{
		Provider: searchcap.ProviderLocalSources,
		Metadata: map[string]any{
			"approval_mode": "notify",
		},
	}

	notice := directSearchNotice(resp)

	if !strings.Contains(notice, "governed local-source results") {
		t.Fatalf("notice = %q, want local-source trust boundary", notice)
	}
	if strings.Contains(notice, "external results are leads") {
		t.Fatalf("notice = %q, should not call local-source results external", notice)
	}
}

func TestSearchCapabilityQuestionDoesNotMatchResearchTeamPrompt(t *testing.T) {
	prompt := "i need an indepth ai research team that can take on various aspects of current research to understand optimal agentry architecture"

	if isSearchCapabilityQuestion(prompt) {
		t.Fatal("research team prompt should not be treated as a search capability question")
	}
}

func TestDirectSearchDoesNotStealTeamCreationResearchPrompt(t *testing.T) {
	prompt := "create a team to look up latest AI agent architecture research"

	if query, ok := shouldHandleDirectSearch(prompt); ok {
		t.Fatalf("team creation prompt routed to direct search query %q", query)
	}
}

func TestDirectSearchStillHandlesPlainLookupPrompt(t *testing.T) {
	prompt := "look up latest AI agent architecture research"

	if query, ok := shouldHandleDirectSearch(prompt); !ok || query != prompt {
		t.Fatalf("plain lookup direct search = (%q, %v), want original query and ok", query, ok)
	}
}

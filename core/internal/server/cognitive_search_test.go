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
		"Notice: Soma Search checked the public web. External results are leads; verify before relying.\nResults:\n1. Agent product release",
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
	if strings.Contains(payload.Text, "local_api") || strings.Contains(payload.Text, "web_search via") {
		t.Fatalf("payload.text = %q, should keep provider IDs out of the default reply", payload.Text)
	}
	if !strings.Contains(payload.Text, "Soma Search checked the public web") {
		t.Fatalf("payload.text = %q, want user-readable search boundary", payload.Text)
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
	if len(payload.ExecutionSummary.CapabilityUse) != 1 || !strings.Contains(payload.ExecutionSummary.CapabilityUse[0].Reason, "External or public web provider") {
		t.Fatalf("capability reason = %+v, want external search source provenance", payload.ExecutionSummary.CapabilityUse)
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

func TestHandleChat_DirectSearchConfiguredWebUsesWebByDefault(t *testing.T) {
	searchAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source_scope") != "web" {
			t.Fatalf("source_scope = %q, want web", r.URL.Query().Get("source_scope"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"title":"Frameworks","url":"https://example.test/frameworks","snippet":"Framework news"}]}`))
	}))
	t.Cleanup(searchAPI.Close)

	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalAPI, LocalAPIEndpoint: searchAPI.URL}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"search on what's the latest popular multi agent framework?"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromAPIResponse(t, rr)
	if strings.Contains(payload.Text, "public web research is not enabled") {
		t.Fatalf("payload.text = %q, should use configured web search", payload.Text)
	}
	if !strings.Contains(payload.Text, "Soma Search checked the public web") || !strings.Contains(payload.Text, "Frameworks") {
		t.Fatalf("payload.text = %q, want configured web result", payload.Text)
	}
	if payload.ExecutionSummary == nil {
		t.Fatal("expected execution summary")
	}
	if payload.ExecutionSummary.Execution.Status != protocol.ExecutionStatusCompleted {
		t.Fatalf("execution status = %q, want completed", payload.ExecutionSummary.Execution.Status)
	}
	if len(payload.ExecutionSummary.CapabilityUse) != 1 || !strings.Contains(payload.ExecutionSummary.CapabilityUse[0].Reason, "External or public web provider") {
		t.Fatalf("capability reason = %+v, want external web boundary", payload.ExecutionSummary.CapabilityUse)
	}
}

func TestHandleChat_DirectSearchExplicitPublicWebBlocksWhenUnconfigured(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalSources}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"search the public web for the latest popular multi agent framework"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromAPIResponse(t, rr)
	if !strings.Contains(payload.Text, "public web research is not enabled") {
		t.Fatalf("payload.text = %q, want public web setup blocker", payload.Text)
	}
	if payload.ExecutionSummary == nil || payload.ExecutionSummary.Execution.Status != protocol.ExecutionStatusBlocked {
		t.Fatalf("execution_summary = %+v, want blocked", payload.ExecutionSummary)
	}
	if payload.ExecutionSummary.AuditRecovery.Degradation == nil || payload.ExecutionSummary.AuditRecovery.Degradation.Code != "web_provider_not_configured" {
		t.Fatalf("degradation = %+v, want web_provider_not_configured", payload.ExecutionSummary.AuditRecovery.Degradation)
	}
}

func TestHandleChat_CanYouSearchOnRunsDirectSearch(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalSources}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"can you search on what's the latest popular multi agent framework?"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromAPIResponse(t, rr)
	if strings.Contains(payload.Text, "Soma Search is available") {
		t.Fatalf("payload.text = %q, should execute search instead of explaining capability", payload.Text)
	}
	if !strings.Contains(payload.ExecutionSummary.Intent.Original, "what's the latest popular multi agent framework") {
		t.Fatalf("intent.original = %q, want extracted search topic", payload.ExecutionSummary.Intent.Original)
	}
	if payload.ExecutionSummary.Execution.Status != protocol.ExecutionStatusBlocked {
		t.Fatalf("execution status = %q, want blocked public-web boundary for local-only provider", payload.ExecutionSummary.Execution.Status)
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

	if !strings.Contains(notice, "approved local data and mounted sources") {
		t.Fatalf("notice = %q, want local-source trust boundary", notice)
	}
	if strings.Contains(notice, "external results are leads") {
		t.Fatalf("notice = %q, should not call local-source results external", notice)
	}
}

func TestDirectSearchMissingWebScopeDetectsPartialMixedCoverage(t *testing.T) {
	resp := searchcap.Response{
		Metadata: map[string]any{
			"partial_source_scope": "local_sources_only",
			"missing_source_scope": "web",
		},
	}

	if !directSearchMissingWebScope(resp) {
		t.Fatalf("directSearchMissingWebScope(%+v) = false, want true", resp.Metadata)
	}
}

func TestBuildSearchExecutionSummaryNamesLocalSourceProvenance(t *testing.T) {
	summary := buildSearchExecutionSummary(
		"what is your latest research",
		"Notice: Soma Search checked approved local data and mounted sources.",
		"audit-123",
		[]string{"web_search"},
		protocol.ExecutionStatusCompleted,
		"",
		nil,
	)

	if summary == nil || len(summary.CapabilityUse) != 1 {
		t.Fatalf("summary capability use = %+v", summary)
	}
	if summary.CapabilityUse[0].Reason != "Search source: Local Mycelis context" {
		t.Fatalf("capability reason = %q, want local-source provenance", summary.CapabilityUse[0].Reason)
	}
}

func TestBuildSearchExecutionSummaryNamesPublicWebBlocker(t *testing.T) {
	summary := buildSearchExecutionSummary(
		"search latest public agent frameworks",
		"Notice: public web research was requested, but Soma Search is currently limited to approved local data and mounted sources.\nBlocked: public web research is not enabled for this workspace.",
		"audit-123",
		[]string{"web_search"},
		protocol.ExecutionStatusBlocked,
		"Public web search is not configured.",
		searchDegradation("web_provider_not_configured", "Public web search is not configured.", "Configure public web search or search local sources only."),
	)

	if summary == nil || len(summary.CapabilityUse) != 1 {
		t.Fatalf("summary capability use = %+v", summary)
	}
	if summary.CapabilityUse[0].Reason != "Search source: Public web requested; no public-web provider configured" {
		t.Fatalf("capability reason = %q, want public-web blocker provenance", summary.CapabilityUse[0].Reason)
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

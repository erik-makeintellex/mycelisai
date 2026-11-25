package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleChat_UsesSearchCapabilityForFreshnessQuestion(t *testing.T) {
	searchAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); !strings.Contains(got, "latest news") {
			t.Fatalf("query = %q, want freshness prompt", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Agent product release","url":"https://example.test/agents","snippet":"A current result about agent products."}]}`))
	}))
	t.Cleanup(searchAPI.Close)
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalAPI, LocalAPIEndpoint: searchAPI.URL}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"what's been the latest news updates regarding ai agent products"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !containsString(payload.ToolsUsed, "web_search") {
		t.Fatalf("tools_used = %v, want web_search", payload.ToolsUsed)
	}
	if !strings.Contains(payload.Text, "Agent product release") {
		t.Fatalf("payload.text = %q, want search result", payload.Text)
	}
}

func TestDirectSearchQuery_DoesNotHijackInternalRecentOrTodayQuestions(t *testing.T) {
	internalPrompts := []string{
		"what teams did we create today?",
		"summarize our recent proposals",
		"what did Soma build recently?",
		"show the latest deliverables in our workspace",
	}
	for _, prompt := range internalPrompts {
		if q, ok := directSearchQuery(prompt); ok {
			t.Fatalf("expected directSearchQuery(%q) = false, got true (%q)", prompt, q)
		}
	}

	webPrompts := []string{
		"search the web for golang 1.26 release notes",
		"what's been the latest news updates regarding ai agent products",
		"web_search(query=\"kubernetes 1.32 release\")",
	}
	for _, prompt := range webPrompts {
		if _, ok := directSearchQuery(prompt); !ok {
			t.Fatalf("expected directSearchQuery(%q) = true", prompt)
		}
	}
}

func TestHandleChat_RetryTurnReExecutesPriorUserPrompt(t *testing.T) {
	searchAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); !strings.Contains(got, "golang") {
			t.Fatalf("query = %q, want retried golang search prompt", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Go Programming Language","url":"https://go.dev","snippet":"Build simple, secure, scalable systems with Go."}]}`))
	}))
	t.Cleanup(searchAPI.Close)

	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalAPI, LocalAPIEndpoint: searchAPI.URL}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"search the web for golang"},{"role":"assistant","content":"I tried to use web_search, but search is blocked right now."},{"role":"user","content":"try again"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", rr.Code, rr.Body.String())
	}
	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, _ := json.Marshal(resp.Data)
	var envelope protocol.CTSEnvelope
	_ = json.Unmarshal(data, &envelope)
	var payload protocol.ChatResponsePayload
	_ = json.Unmarshal(envelope.Payload, &payload)
	if !containsString(payload.ToolsUsed, "web_search") {
		t.Fatalf("tools_used = %v, want web_search on 'try again'", payload.ToolsUsed)
	}
	if !strings.Contains(payload.Text, "Go Programming Language") {
		t.Fatalf("payload.text = %q, want search result on 'try again'", payload.Text)
	}
	if !strings.Contains(payload.Text, "## Web Search Results") || !strings.Contains(payload.Text, "**[Go Programming Language](https://go.dev)**") {
		t.Fatalf("payload.text = %q, want structured markdown headings and bold links", payload.Text)
	}

	repeatBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"search the web for golang"},{"role":"assistant","content":"Prior search results"},{"role":"user","content":"show me list again"}]}`)
	repeatReq := httptest.NewRequest(http.MethodPost, "/api/v1/chat", repeatBody)
	repeatRR := httptest.NewRecorder()
	http.HandlerFunc(s.HandleChat).ServeHTTP(repeatRR, repeatReq)
	if repeatRR.Code != http.StatusOK {
		t.Fatalf("repeat status = %d, want 200 body=%s", repeatRR.Code, repeatRR.Body.String())
	}
	var repeatResp protocol.APIResponse
	_ = json.NewDecoder(repeatRR.Body).Decode(&repeatResp)
	repeatData, _ := json.Marshal(repeatResp.Data)
	var repeatEnv protocol.CTSEnvelope
	_ = json.Unmarshal(repeatData, &repeatEnv)
	var repeatPayload protocol.ChatResponsePayload
	_ = json.Unmarshal(repeatEnv.Payload, &repeatPayload)
	if !strings.Contains(repeatPayload.Text, "## Web Search Results") || !strings.Contains(repeatPayload.Text, "- **Summary:**") {
		t.Fatalf("repeatPayload.text = %q, want formatted markdown list on 'show me list again'", repeatPayload.Text)
	}
}

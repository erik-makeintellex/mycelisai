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

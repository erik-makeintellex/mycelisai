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

func TestHandleChat_DirectSearchMixedScopeDisclosesWebOnlyCoverage(t *testing.T) {
	searchAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source_scope") != "web" {
			t.Fatalf("source_scope = %q, want web for public side of mixed search", r.URL.Query().Get("source_scope"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"title":"Public result","url":"https://example.test/public","snippet":"Public snippet"}]}`))
	}))
	t.Cleanup(searchAPI.Close)

	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{
			Provider:         searchcap.ProviderLocalAPI,
			LocalAPIEndpoint: searchAPI.URL,
		}, nil, nil)
	})

	reqBody := bytes.NewBufferString(`{"messages":[{"role":"user","content":"search for agent orchestration in internal and public sources"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", reqBody)
	rr := httptest.NewRecorder()

	http.HandlerFunc(s.HandleChat).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	payload := decodeChatPayloadFromAPIResponse(t, rr)
	if !strings.Contains(payload.Text, "approved local data and mounted-source coverage was unavailable") {
		t.Fatalf("payload.text = %q, want mixed-source partial warning", payload.Text)
	}
	if !strings.Contains(payload.Text, "Coverage warning:") {
		t.Fatalf("payload.text = %q, want explicit coverage warning", payload.Text)
	}
	if payload.ExecutionSummary == nil || payload.ExecutionSummary.Execution.Status != protocol.ExecutionStatusCompleted {
		t.Fatalf("execution_summary = %+v, want completed partial result", payload.ExecutionSummary)
	}
}

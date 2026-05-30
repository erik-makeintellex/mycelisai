package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

type cognitiveStatusProbe struct {
	calls   int
	healthy bool
}

func (p *cognitiveStatusProbe) Infer(context.Context, string, cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "ok", Provider: "probe", ModelUsed: "probe-model"}, nil
}

func (p *cognitiveStatusProbe) Probe(context.Context) (bool, error) {
	p.calls++
	return p.healthy, nil
}

func TestHandleCognitiveStatus_SkipsDisabledTextProviders(t *testing.T) {
	disabled := &cognitiveStatusProbe{healthy: true}
	s := &AdminServer{Cognitive: &cognitive.Router{
		Config: &cognitive.BrainConfig{Providers: map[string]cognitive.ProviderConfig{
			"disabled-local": {Type: "openai_compatible", Endpoint: "http://127.0.0.1:1234/v1", ModelID: "disabled-model", Enabled: false},
		}},
		Adapters: map[string]cognitive.LLMProvider{"disabled-local": disabled},
	}}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cognitive/status", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(s.HandleCognitiveStatus).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if disabled.calls != 0 {
		t.Fatalf("disabled provider probe calls = %d, want 0", disabled.calls)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if text := resp["text"].(map[string]any); text["status"] == "online" {
		t.Fatalf("disabled provider should not make text online: %#v", text)
	}
}

package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// b1rbCountingProvider records every provider call so a test can prove that
// no inference ran.
type b1rbCountingProvider struct{ calls int }

func (p *b1rbCountingProvider) Infer(_ context.Context, _ string, o cognitive.InferOptions) (*cognitive.InferResponse, error) {
	p.calls++
	return &cognitive.InferResponse{Text: "ok", PromptTokens: 100, CompletionTokens: o.MaxTokens, TokensUsed: 100 + o.MaxTokens}, nil
}

func (p *b1rbCountingProvider) Probe(context.Context) (bool, error) { return true, nil }

// B1R-B F2: the raw POST /api/v1/cognitive/infer surface had no product
// caller, decoded Correlation from the body and ran as uncorrelated system
// inference (no team or agent day caps). It is removed: no caller, whatever
// its role or body, can reach a provider through it or charge any team.
func TestB1rbRawInferRouteIsRemoved(t *testing.T) {
	callers := []struct {
		name     string
		identity *RequestIdentity
	}{
		{"anonymous", nil},
		{"standard user", &RequestIdentity{UserID: "u-1", Role: "user"}},
		{"admin without cognitive scope", &RequestIdentity{UserID: "u-2", Role: "admin", Scopes: []string{"config_documents:write"}}},
		{"root admin", localAdminIdentityForTest()},
	}
	bodies := []string{
		`{"profile":"chat","prompt":"hi"}`,
		`{"profile":"chat","prompt":"hi","Correlation":{"TeamID":"victim-team","AgentID":"a","RunID":"r"},"Meter":null}`,
	}
	for _, caller := range callers {
		for _, body := range bodies {
			t.Run(caller.name, func(t *testing.T) {
				router := tokenBudgetRouter()
				provider := &b1rbCountingProvider{}
				router.Adapters["local"] = provider
				s := newTestServer(func(s *AdminServer) { s.Cognitive = router })
				mux := http.NewServeMux()
				s.RegisterRoutes(mux)
				for _, method := range []string{http.MethodPost, http.MethodGet} {
					var status int
					if caller.identity == nil {
						status = doRequest(t, mux, method, "/api/v1/cognitive/infer", body).Code
					} else {
						status = doAuthenticatedRequestAs(t, mux, method, "/api/v1/cognitive/infer", body, caller.identity).Code
					}
					if status != http.StatusNotFound {
						t.Fatalf("%s /api/v1/cognitive/infer = %d, want 404 (route removed)", method, status)
					}
				}
				if provider.calls != 0 {
					t.Fatalf("provider called %d times through the removed route", provider.calls)
				}
				used := router.Budgets.Usage(context.Background(), "", protocol.TokenBudgetScopeTeamDay, "victim-team",
					router.Budgets.Limits(cognitive.BudgetSubject{TeamID: "victim-team"})).Used
				if used != 0 {
					t.Fatalf("spoofed team charged %d tokens", used)
				}
			})
		}
	}
}

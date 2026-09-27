package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func tokenBudgetRouter() *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles: map[string]string{"chat": "local", "coder": "cloud"},
			Providers: map[string]cognitive.ProviderConfig{
				"local": {Type: "openai_compatible", ModelID: "qwen3:14b", Enabled: true, DataBoundary: cognitive.DataBoundaryLocalOnly},
				"cloud": {Type: "anthropic", ModelID: "claude", Enabled: true, DataBoundary: cognitive.DataBoundaryLeavesOrg},
			},
		},
		Adapters: map[string]cognitive.LLMProvider{},
		Budgets:  cognitive.NewBudgetGovernor(protocol.DefaultTokenBudgetPolicySpec(), nil),
	}
}

func tokenBudgetMux(s *AdminServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/cognitive/budgets", s.HandleGetTokenBudgets)
	mux.HandleFunc("GET /api/v1/cognitive/budgets/usage", s.HandleGetTokenBudgetUsage)
	mux.HandleFunc("PUT /api/v1/cognitive/budgets/overrides/{level}/{ref}", s.HandlePutTokenBudgetOverride)
	mux.HandleFunc("DELETE /api/v1/cognitive/budgets/overrides/{level}/{ref}", s.HandleDeleteTokenBudgetOverride)
	return mux
}

func tokenBudgetData(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	return env.Data
}

func TestTokenBudgetOverrides_AuthorityDeniesBeforeAnyWrite(t *testing.T) {
	callers := []struct {
		name     string
		identity *RequestIdentity
		want     int
		scoped   bool
	}{
		{"anonymous", nil, http.StatusUnauthorized, false},
		{"standard user", &RequestIdentity{UserID: "u-1", Role: "user", Scopes: []string{"cognitive:write"}}, http.StatusForbidden, false},
		{"admin without cognitive:write", &RequestIdentity{UserID: "u-2", Role: "admin", Scopes: []string{"cognitive:read", "config_documents:write"}}, http.StatusForbidden, true},
	}
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		for _, caller := range callers {
			t.Run(method+" "+caller.name, func(t *testing.T) {
				dbOpt, mock := withDirectDB(t)
				s := newTestServer(dbOpt, func(s *AdminServer) { s.Cognitive = tokenBudgetRouter() })
				path, body := "/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_execution":2048}`
				var rr *httptest.ResponseRecorder
				if caller.identity == nil {
					rr = doRequest(t, tokenBudgetMux(s), method, path, body)
				} else {
					rr = doAuthenticatedRequestAs(t, tokenBudgetMux(s), method, path, body, caller.identity)
				}
				assertStatus(t, rr, caller.want)
				if caller.want == http.StatusForbidden {
					data := tokenBudgetData(t, rr)
					if data["code"] != codeAdminRequired || (data["required_scope"] == "cognitive:write") != caller.scoped {
						t.Fatalf("denial data = %v", data)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("db touched: %v", err)
				}
				if spec := s.Cognitive.Budgets.Policy(); len(spec.Overrides.Team) != 0 {
					t.Fatalf("rejected call changed policy: %+v", spec.Overrides)
				}
			})
		}
	}
}

func TestTokenBudgetOverrides_AdversarialBodiesAreRejected(t *testing.T) {
	cases := []struct{ path, body string }{
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_execution":0}`},
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_run":-1}`},
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_team_day":5000001}`},
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_execution":1023}`},
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_execution":300000}`},
		{"/api/v1/cognitive/budgets/overrides/agent/coder", `{"per_execution":90000,"per_run":80000}`},
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{}`},
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_execution":"lots"}`},
		{"/api/v1/cognitive/budgets/overrides/team/team-a", `{"warn_pct":100}`},
		{"/api/v1/cognitive/budgets/overrides/tenant/default", `{"per_execution":2048}`},
		{"/api/v1/cognitive/budgets/overrides/class/unlimited", `{"per_execution":2048}`},
		{"/api/v1/cognitive/budgets/overrides/team/" + strings.Repeat("x", 129), `{"per_execution":2048}`},
	}
	for _, tc := range cases {
		t.Run(tc.path+" "+tc.body, func(t *testing.T) {
			dbOpt, mock := withDirectDB(t)
			s := newTestServer(dbOpt, func(s *AdminServer) { s.Cognitive = tokenBudgetRouter() })
			rr := doAuthenticatedRequest(t, tokenBudgetMux(s), http.MethodPut, tc.path, tc.body)
			assertStatus(t, rr, http.StatusBadRequest)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("db touched: %v", err)
			}
		})
	}
}

func TestTokenBudgetOverrides_WithoutDatabaseIsHonest503(t *testing.T) {
	s := newTestServer(func(s *AdminServer) { s.Cognitive = tokenBudgetRouter() })
	rr := doAuthenticatedRequest(t, tokenBudgetMux(s), http.MethodPut, "/api/v1/cognitive/budgets/overrides/team/team-a", `{"per_execution":2048}`)
	assertStatus(t, rr, http.StatusServiceUnavailable)
	if spec := s.Cognitive.Budgets.Policy(); len(spec.Overrides.Team) != 0 {
		t.Fatalf("policy changed without durable storage: %+v", spec.Overrides)
	}
}

func TestGetTokenBudgets_EffectivePolicyAndAdminOnlyProvenance(t *testing.T) {
	s := newTestServer(func(s *AdminServer) { s.Cognitive = tokenBudgetRouter() })
	spec := protocol.DefaultTokenBudgetPolicySpec()
	spec.Overrides.Team = map[string]protocol.TokenBudgetLimits{"team-a": {PerExecution: 2048}}
	s.Cognitive.Budgets.SetPolicy(spec)

	user := doAuthenticatedRequestAs(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets", "", &RequestIdentity{UserID: "u-1", Role: "user"})
	assertStatus(t, user, http.StatusOK)
	data := tokenBudgetData(t, user)
	classes := data["classes"].(map[string]any)
	if classes["local_large"].(map[string]any)["per_execution"] != float64(64000) || data["default_class"] != "local_large" {
		t.Fatalf("user view = %v", data)
	}
	if _, leaked := data["overrides"]; leaked || data["can_edit"] != false {
		t.Fatalf("override provenance leaked to a non-admin: %v", data)
	}
	providers := data["providers"].([]any)
	if len(providers) != 2 {
		t.Fatalf("providers = %v", providers)
	}

	admin := doAuthenticatedRequest(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets", "")
	adminData := tokenBudgetData(t, admin)
	overrides := adminData["overrides"].(map[string]any)
	if overrides["team"].(map[string]any)["team-a"].(map[string]any)["per_execution"] != float64(2048) || adminData["can_edit"] != true {
		t.Fatalf("admin view = %v", adminData)
	}
	anonymous := doRequest(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets", "")
	assertStatus(t, anonymous, http.StatusUnauthorized)
}

func TestGetTokenBudgetUsage_ReportsSinceRestartWhenLedgerUnavailable(t *testing.T) {
	provider := &draftUsageProvider{used: 1500}
	router := tokenBudgetRouter()
	router.Adapters["local"] = provider
	s := newTestServer(func(s *AdminServer) { s.Cognitive = router })
	correlation := cognitive.InferenceCorrelation{RunID: "run-9", TeamID: "team-a", AgentID: "coder"}
	ctx := cognitive.WithExecutionMeter(context.Background(), cognitive.NewExecutionMeter(cognitive.ExecutionKindAgentTurn, correlation))
	if _, err := router.InferWithContract(ctx, cognitive.InferRequest{Profile: "chat", Prompt: "x", Correlation: correlation}); err != nil {
		t.Fatal(err)
	}
	rr := doAuthenticatedRequestAs(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets/usage?team_id=team-a", "", &RequestIdentity{UserID: "u-1", Role: "user"})
	assertStatus(t, rr, http.StatusOK)
	data := tokenBudgetData(t, rr)
	if data["scope"] != "team_day" || data["ref"] != "team-a" || data["used"] != float64(1500) || data["limit"] != float64(2000000) ||
		data["remaining"] != float64(1998500) || data["period"] != "since_restart" || data["usage_reported"] != true || data["warn"] != false {
		t.Fatalf("usage = %v", data)
	}
	run := tokenBudgetData(t, doAuthenticatedRequest(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets/usage?run_id=run-9", ""))
	if run["scope"] != "run" || run["used"] != float64(1500) || run["limit"] != float64(256000) || run["period"] != "since_restart" {
		t.Fatalf("run usage = %v", run)
	}
	for _, query := range []string{"", "?team_id=a&agent_id=b", "?tenant=x"} {
		bad := doAuthenticatedRequest(t, tokenBudgetMux(s), http.MethodGet, "/api/v1/cognitive/budgets/usage"+query, "")
		assertStatus(t, bad, http.StatusBadRequest)
	}
}

func TestStructuredChatBlocker_AgentBudgetStopIsHonest429(t *testing.T) {
	rr := httptest.NewRecorder()
	respondStructuredChatBlocker(rr, chatAgentResult{Availability: &cognitive.ExecutionAvailability{
		Code: cognitive.TokenBudgetExhaustedCode, Summary: "This work stopped because it reached its token budget (execution: 2048 of 2048 tokens).",
	}})
	assertStatus(t, rr, http.StatusTooManyRequests)
	if strings.Contains(rr.Body.String(), "confirm_token") || !strings.Contains(rr.Body.String(), cognitive.TokenBudgetExhaustedCode) {
		t.Fatalf("body = %s", rr.Body.String())
	}
}

var _ = sqlmock.AnyArg

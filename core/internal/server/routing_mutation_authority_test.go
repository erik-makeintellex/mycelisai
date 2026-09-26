package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// countingAdapter records probes so tests can prove a rejected call made no
// outbound egress.
type countingAdapter struct {
	probes atomic.Int64
}

func (c *countingAdapter) Infer(_ context.Context, _ string, _ cognitive.InferOptions) (*cognitive.InferResponse, error) {
	return &cognitive.InferResponse{Text: "stub"}, nil
}

func (c *countingAdapter) Probe(_ context.Context) (bool, error) {
	c.probes.Add(1)
	return true, nil
}

type routingRoute struct{ name, method, path, body string }

// routingMutationRoutes is every route S6d gates.
var routingMutationRoutes = []routingRoute{
	{"toggle", http.MethodPut, "/api/v1/brains/ollama/toggle", `{"enabled":false}`},
	{"policy", http.MethodPut, "/api/v1/brains/ollama/policy", `{"usage_policy":"cloud_ok","roles_allowed":["all"]}`},
	{"add", http.MethodPost, "/api/v1/brains", `{"id":"evil","type":"openai_compatible","endpoint":"http://attacker.invalid/v1","model_id":"x","enabled":true}`},
	{"update", http.MethodPut, "/api/v1/brains/ollama", `{"type":"openai_compatible","endpoint":"http://attacker.invalid/v1","model_id":"evil","enabled":true}`},
	{"delete", http.MethodDelete, "/api/v1/brains/ollama", ""},
	{"probe", http.MethodPost, "/api/v1/brains/ollama/probe", ""},
	{"profile create", http.MethodPost, "/api/v1/mission-profiles", `{"name":"x","role_providers":{"chat":"ollama"}}`},
	{"profile update", http.MethodPut, "/api/v1/mission-profiles/p-1", `{"name":"x","role_providers":{"chat":"ollama"}}`},
	{"profile delete", http.MethodDelete, "/api/v1/mission-profiles/p-1", ""},
	{"profile activate", http.MethodPost, "/api/v1/mission-profiles/p-1/activate", ""},
}

type routingFixture struct {
	s      *AdminServer
	mock   sqlmock.Sqlmock
	mux    *http.ServeMux
	probe  *countingAdapter
	before *cognitive.BrainConfig
}

// newRoutingFixture wires every brains and mission-profile route onto the
// S6c profile-override router (all execution profiles on root vllm; ollama
// enabled and unbound). The ollama adapter counts probes.
func newRoutingFixture(t *testing.T) *routingFixture {
	t.Helper()
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) { s.Cognitive = profileOverrideRouter(t) })
	probe := &countingAdapter{}
	s.Cognitive.Adapters["ollama"] = probe
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/brains", s.HandleListBrains)
	mux.HandleFunc("PUT /api/v1/brains/{id}/toggle", s.HandleToggleBrain)
	mux.HandleFunc("PUT /api/v1/brains/{id}/policy", s.HandleUpdateBrainPolicy)
	mux.HandleFunc("POST /api/v1/brains", s.HandleAddBrain)
	mux.HandleFunc("PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	mux.HandleFunc("DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	mux.HandleFunc("POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)
	mux.HandleFunc("POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	mux.HandleFunc("PUT /api/v1/mission-profiles/{id}", s.HandleUpdateMissionProfile)
	mux.HandleFunc("DELETE /api/v1/mission-profiles/{id}", s.HandleDeleteMissionProfile)
	mux.HandleFunc("POST /api/v1/mission-profiles/{id}/activate", s.HandleActivateMissionProfile)
	return &routingFixture{s: s, mock: mock, mux: mux, probe: probe, before: s.Cognitive.ConfigSnapshot()}
}

// assertRoutingUnchanged proves no provider, profile, adapter, YAML file or
// DB change and no probe egress happened.
func (f *routingFixture) assertRoutingUnchanged(t *testing.T) {
	t.Helper()
	after := f.s.Cognitive.ConfigSnapshot()
	if !reflect.DeepEqual(f.before.Providers, after.Providers) {
		t.Fatalf("providers changed:\nbefore=%+v\nafter=%+v", f.before.Providers, after.Providers)
	}
	if !reflect.DeepEqual(f.before.Profiles, after.Profiles) || !reflect.DeepEqual(f.before.ProfileOverrideOrigins, after.ProfileOverrideOrigins) {
		t.Fatalf("profiles changed: before=%v after=%v origins=%v", f.before.Profiles, after.Profiles, after.ProfileOverrideOrigins)
	}
	if _, ok := f.s.Cognitive.AdapterSnapshot("evil"); ok {
		t.Fatal("adapter added by rejected call")
	}
	if adapter, _ := f.s.Cognitive.AdapterSnapshot("ollama"); adapter != cognitive.LLMProvider(f.probe) {
		t.Fatal("ollama adapter replaced by rejected call")
	}
	if data, err := os.ReadFile(f.s.Cognitive.ConfigPath); err != nil || string(data) != "# test\n" {
		t.Fatalf("cognitive.yaml written by rejected call: %q err=%v", data, err)
	}
	if n := f.probe.probes.Load(); n != 0 {
		t.Fatalf("rejected call probed the provider %d times", n)
	}
	if err := f.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
}

func TestRoutingMutationAuthority_RejectsAnonymousStandardAndUnscopedAdmin(t *testing.T) {
	callers := []struct {
		name     string
		identity *RequestIdentity
		want     int
	}{
		{"anonymous", nil, http.StatusUnauthorized},
		{"standard soma:work user", &RequestIdentity{UserID: "u-1", Role: "user", Scopes: []string{"soma:work"}}, http.StatusForbidden},
		{"admin without cognitive:write", &RequestIdentity{UserID: "u-2", Role: "admin", Scopes: []string{"cognitive:read", "config_documents:write"}}, http.StatusForbidden},
		{"non-admin with cognitive:write", &RequestIdentity{UserID: "u-3", Role: "user", Scopes: []string{"cognitive:write"}}, http.StatusForbidden},
		{"non-admin with wildcard", &RequestIdentity{UserID: "u-4", Role: "user", Scopes: []string{"*"}}, http.StatusForbidden},
	}
	for _, route := range routingMutationRoutes {
		for _, caller := range callers {
			t.Run(route.name+"/"+caller.name, func(t *testing.T) {
				f := newRoutingFixture(t)
				var rr *httptest.ResponseRecorder
				if caller.identity == nil {
					rr = doRequest(t, f.mux, route.method, route.path, route.body)
				} else {
					rr = doAuthenticatedRequestAs(t, f.mux, route.method, route.path, route.body, caller.identity)
				}
				assertStatus(t, rr, caller.want)
				if strings.Contains(rr.Body.String(), `"data"`) {
					t.Fatalf("rejected call leaked data: %s", rr.Body.String())
				}
				f.assertRoutingUnchanged(t)
			})
		}
	}
}

// expectActivationLoad queues the activation SELECT for profile p-1.
func expectActivationLoad(mock sqlmock.Sqlmock, roleProviders string) {
	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM mission_profiles WHERE").
		WillReturnRows(sqlmock.NewRows(profileColumns).
			AddRow("p-1", "Research", sql.NullString{}, []byte(roleProviders), []byte(`[]`),
				"fresh", false, false, "default", now, now))
}

func TestRoutingMutationAuthority_AuditFailureChangesNothing(t *testing.T) {
	for _, route := range routingMutationRoutes {
		t.Run(route.name, func(t *testing.T) {
			f := newRoutingFixture(t)
			if route.name == "profile activate" {
				expectActivationLoad(f.mock, `{"coder":"ollama"}`)
			}
			f.mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
			rr := doAuthenticatedRequest(t, f.mux, route.method, route.path, route.body)
			assertStatus(t, rr, http.StatusServiceUnavailable)
			if !strings.Contains(rr.Body.String(), "Audit unavailable") {
				t.Fatalf("body = %s", rr.Body.String())
			}
			f.assertRoutingUnchanged(t)
		})
	}
}

// Without a database the audit id is empty; brains mutations fail closed.
func TestRoutingMutationAuthority_NoAuditStoreFailsClosed(t *testing.T) {
	for _, route := range routingMutationRoutes[:6] {
		t.Run(route.name, func(t *testing.T) {
			f := newRoutingFixture(t)
			f.s.DB = nil
			assertStatus(t, doAuthenticatedRequest(t, f.mux, route.method, route.path, route.body), http.StatusServiceUnavailable)
			f.assertRoutingUnchanged(t)
		})
	}
}

func TestRoutingMutationAuthority_AdminWithScopeIsAuditedFirst(t *testing.T) {
	admin := &RequestIdentity{UserID: "root-1", Role: "admin", Scopes: []string{"cognitive:write"}}
	t.Run("toggle unbound provider", func(t *testing.T) {
		f := newRoutingFixture(t)
		f.mock.ExpectExec("INSERT INTO log_entries").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", routingMutationAuditSource, sqlmock.AnyArg(), "Cognitive provider toggled", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		rr := doAuthenticatedRequestAs(t, f.mux, http.MethodPut, "/api/v1/brains/ollama/toggle", `{"enabled":false}`, admin)
		assertStatus(t, rr, http.StatusOK)
		if p, _ := f.s.Cognitive.ProviderSnapshot("ollama"); p.Enabled {
			t.Fatal("ollama still enabled")
		}
		if err := f.mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql: %v", err)
		}
	})
	t.Run("probe", func(t *testing.T) {
		f := newRoutingFixture(t)
		expectAudit(f.mock)
		assertStatus(t, doAuthenticatedRequestAs(t, f.mux, http.MethodPost, "/api/v1/brains/ollama/probe", "", admin), http.StatusOK)
		if f.probe.probes.Load() != 1 {
			t.Fatalf("probes = %d, want 1", f.probe.probes.Load())
		}
	})
	t.Run("policy", func(t *testing.T) {
		f := newRoutingFixture(t)
		expectAudit(f.mock)
		assertStatus(t, doAuthenticatedRequestAs(t, f.mux, http.MethodPut, "/api/v1/brains/ollama/policy", `{"usage_policy":"cloud_ok"}`, admin), http.StatusOK)
		if p, _ := f.s.Cognitive.ProviderSnapshot("ollama"); p.UsagePolicy != "cloud_ok" {
			t.Fatalf("policy = %q", p.UsagePolicy)
		}
	})
}

// Validation rejections happen before the audit, so they write nothing.
func TestRoutingMutationAuthority_RejectionsBeforeAuditWriteNothing(t *testing.T) {
	cases := []struct {
		route routingRoute
		want  int
	}{
		{routingRoute{"toggle missing", http.MethodPut, "/api/v1/brains/ghost/toggle", `{"enabled":false}`}, http.StatusNotFound},
		{routingRoute{"add duplicate", http.MethodPost, "/api/v1/brains", `{"id":"ollama","type":"openai_compatible"}`}, http.StatusConflict},
		{routingRoute{"add raw key", http.MethodPost, "/api/v1/brains", `{"id":"x","type":"openai","api_key":"sk-live"}`}, http.StatusBadRequest},
		{routingRoute{"probe missing", http.MethodPost, "/api/v1/brains/ghost/probe", ""}, http.StatusNotFound},
		{routingRoute{"profile create unnamed", http.MethodPost, "/api/v1/mission-profiles", `{"role_providers":{}}`}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.route.name, func(t *testing.T) {
			f := newRoutingFixture(t)
			assertStatus(t, doAuthenticatedRequest(t, f.mux, tc.route.method, tc.route.path, tc.route.body), tc.want)
			f.assertRoutingUnchanged(t)
		})
	}
}

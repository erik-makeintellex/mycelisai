package cognitive

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// overrideTestConfig ships ollama defaults for six execution profiles
// (overseer has no YAML default) plus a disabled local-ollama-dev. Endpoints
// point at a local test server so startup probes answer immediately.
const overrideTestConfig = `
providers:
  vllm:
    type: openai_compatible
    endpoint: ENDPOINT
    model_id: m-vllm
    enabled: true
    data_boundary: local_only
  ollama:
    type: ollama
    endpoint: ENDPOINT
    model_id: m-ollama
    enabled: true
    data_boundary: local_only
  local-ollama-dev:
    type: ollama
    endpoint: ENDPOINT
    model_id: m-dev
    enabled: false
profiles:
  admin: ollama
  architect: ollama
  chat: ollama
  coder: ollama
  creative: ollama
  sentry: ollama
`

// newRouterWithRoleRows boots a Router through NewRouter with the given
// system_config role.* rows (or a role-query error).
func newRouterWithRoleRows(t *testing.T, rows map[string]string, roleErr error) *Router {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectQuery("SELECT id, driver, base_url, api_key_env_var, config FROM llm_providers").
		WillReturnRows(sqlmock.NewRows([]string{"id", "driver", "base_url", "api_key_env_var", "config"}))
	roleQuery := mock.ExpectQuery("SELECT key, value FROM system_config WHERE key LIKE 'role\\.%'")
	if roleErr != nil {
		roleQuery.WillReturnError(roleErr)
	} else {
		result := sqlmock.NewRows([]string{"key", "value"})
		for key, value := range rows {
			result.AddRow("role."+key, value)
		}
		roleQuery.WillReturnRows(result)
	}
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(engine.Close)
	config := strings.ReplaceAll(overrideTestConfig, "ENDPOINT", engine.URL+"/v1")
	router, err := NewRouter(writeTestCognitiveConfig(t, config), db)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return router
}

func TestProfileOverride_DBRowToUnusableProviderFailsClosed(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")
	router := newRouterWithRoleRows(t, map[string]string{
		"coder": "local-ollama-dev", "architect": "ghost", "sentry": "  ",
	}, nil)

	for profile, wantCode := range map[string]string{"coder": ExecutionProviderDisabled, "architect": ExecutionProviderMissing} {
		got := router.ExecutionAvailability(profile, "")
		if got.Available || got.Code != wantCode || got.FallbackApplied {
			t.Fatalf("%s availability = %+v, want fail-closed %s with no substitute", profile, got, wantCode)
		}
		if !strings.Contains(got.AdminAction, "Reset the "+profile+" override to root") || !strings.Contains(got.AdminAction, "enable provider") || strings.Contains(got.RecommendedAction, "/api/") {
			t.Fatalf("%s recommended_action = %q", profile, got.RecommendedAction)
		}
		if _, err := router.InferWithContract(context.Background(), InferRequest{Profile: profile, Prompt: "x"}); err == nil {
			t.Fatalf("%s inference must fail closed", profile)
		}
	}
	assertBinding(t, router.Config, "coder", "local-ollama-dev", ProfileSourceOverride)
	// A blank row is treated as absent: root still applies.
	assertBinding(t, router.Config, "sentry", "vllm", ProfileSourceRoot)
	if state := router.ProfileOverrideState("sentry"); state.DBRowPresent || state.Origin != "" {
		t.Fatalf("blank row recorded as override: %+v", state)
	}
	if state := router.ProfileOverrideState("coder"); !state.DBRowPresent || state.Origin != ProfileOriginDB {
		t.Fatalf("coder state = %+v, want db origin with row", state)
	}
}

func TestProfileOverride_RoleQueryErrorSetsOverlayError(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")
	router := newRouterWithRoleRows(t, nil, errors.New("relation system_config does not exist"))
	if !router.Config.OverlayError {
		t.Fatal("OverlayError = false after a role query failure")
	}
	assertBinding(t, router.Config, "chat", "vllm", ProfileSourceRoot)
	if routes, overlay := router.ProfileRoutes(); !overlay || ProfileRouteHealth(routes, overlay) != ProfileRouteHealthDegraded {
		t.Fatalf("overlay error must degrade route health, got overlay=%v", overlay)
	}
}

// TestClearProfileOverride_EqualsRestart proves the reset invariant: after
// clearing a DB override, bindings equal a fresh NewRouter without the row.
func TestClearProfileOverride_EqualsRestart(t *testing.T) {
	cases := []struct {
		name, root, profile, wantProvider, wantSource string
	}{
		{"root set", "vllm", "coder", "vllm", ProfileSourceRoot},
		{"no root uses yaml default", "", "coder", "ollama", ProfileSourceDefault},
		{"no root and no default is unbound", "", "overseer", "", ProfileSourceUnbound},
		{"non-execution profile without default is removed", "vllm", "reviewer", "", ProfileSourceUnbound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clearProviderRoutingEnv(t)
			if tc.root != "" {
				t.Setenv("MYCELIS_ROOT_PROVIDER", tc.root)
			}
			withRow := newRouterWithRoleRows(t, map[string]string{tc.profile: "local-ollama-dev", "creative": "vllm"}, nil)
			cleared := withRow.ClearProfileOverride(tc.profile)
			if !cleared.Changed || cleared.Effective.ProviderID != tc.wantProvider || cleared.Effective.Source != tc.wantSource {
				t.Fatalf("clear = %+v, want %s/%s", cleared, tc.wantProvider, tc.wantSource)
			}
			restarted := newRouterWithRoleRows(t, map[string]string{"creative": "vllm"}, nil)
			got, want := withRow.Config.EffectiveProfileBindings(), restarted.Config.EffectiveProfileBindings()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("reset bindings differ from restart:\n reset   %+v\n restart %+v", got, want)
			}
			if second := withRow.ClearProfileOverride(tc.profile); second.Changed {
				t.Fatalf("second clear changed = true: %+v", second)
			}
		})
	}
}

func TestClearProfileOverride_EnvOverrideSurvivesDBClear(t *testing.T) {
	clearProviderRoutingEnv(t)
	t.Setenv("MYCELIS_ROOT_PROVIDER", "vllm")
	t.Setenv("MYCELIS_PROFILE_CODER_PROVIDER", "ollama")
	router := newRouterWithRoleRows(t, map[string]string{"coder": "local-ollama-dev"}, nil)

	assertBinding(t, router.Config, "coder", "ollama", ProfileSourceOverride)
	routes, overlay := router.ProfileRoutes()
	if coder := routes["coder"]; coder.OverrideOrigin != ProfileOriginEnv || !coder.DBRowPresent || !coder.Available {
		t.Fatalf("coder route = %+v, want env origin with hidden db row", coder)
	}
	if got := ProfileRouteHealth(routes, overlay); got != ProfileRouteHealthDegraded {
		t.Fatalf("hidden db row under env = %s, want degraded", got)
	}

	cleared := router.ClearProfileOverride("coder")
	if !cleared.Changed || cleared.RemainingOverrideOrigin != ProfileOriginEnv || cleared.Effective.ProviderID != "ollama" {
		t.Fatalf("clear = %+v, want env pin kept", cleared)
	}
	if state := router.ProfileOverrideState("coder"); state.DBRowPresent || state.Origin != ProfileOriginEnv {
		t.Fatalf("after clear state = %+v", state)
	}
	var rejection *ProfileOverrideError
	if err := router.SetProfileOverride("coder", "vllm", ProfileOriginDB); !errors.As(err, &rejection) || rejection.Code != OverrideEnvPinned {
		t.Fatalf("override of env-pinned profile = %v, want %s", err, OverrideEnvPinned)
	}
}

func overrideValidationRouter() *Router {
	return &Router{
		Config: &BrainConfig{
			RootProvider: "vllm",
			Providers: map[string]ProviderConfig{
				"vllm":     {Type: "openai_compatible", ModelID: "m", Enabled: true, DataBoundary: DataBoundaryLocalOnly},
				"cloud":    {Type: "openai", ModelID: "c", Enabled: true, DataBoundary: DataBoundaryLeavesOrg},
				"disabled": {Type: "ollama", ModelID: "d", Enabled: false},
				"nomodel":  {Type: "ollama", Enabled: true},
				"noadapt":  {Type: "ollama", ModelID: "n", Enabled: true},
			},
			Profiles:         map[string]string{"chat": "vllm", "coder": "vllm", "embed": "vllm"},
			ProfileSources:   map[string]string{"chat": ProfileSourceRoot, "coder": ProfileSourceRoot},
			ProfileFallbacks: map[string][]string{"architect": {"vllm"}},
		},
		Adapters: map[string]LLMProvider{"vllm": &routeProbeStub{healthy: true}, "cloud": &routeProbeStub{}, "disabled": &routeProbeStub{}, "nomodel": &routeProbeStub{}},
	}
}

func TestSetProfileOverrides_RejectsUnrunnableAndIsAllOrNothing(t *testing.T) {
	cases := map[string]struct {
		profile, provider, code string
	}{
		"disabled":          {"coder", "disabled", ExecutionProviderDisabled},
		"unknown":           {"coder", "ghost", ExecutionProviderMissing},
		"model-less":        {"coder", "nomodel", ExecutionModelMissing},
		"no adapter":        {"coder", "noadapt", ExecutionProviderOffline},
		"blank provider":    {"coder", " ", ExecutionProviderMissing},
		"key-shaped name":   {"role.x", "vllm", OverrideInvalidProfile},
		"traversal name":    {"../x", "vllm", OverrideInvalidProfile},
		"33 characters":     {"a" + strings.Repeat("b", 32), "vllm", OverrideInvalidProfile},
		"unknown profile":   {"reviewer", "vllm", OverrideUnknownProfile},
		"crosses fallbacks": {"architect", "cloud", OverrideBoundaryMismatch},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			router := overrideValidationRouter()
			err := router.SetProfileOverrides(map[string]string{"chat": "vllm", tc.profile: tc.provider}, ProfileOriginDB)
			var rejection *ProfileOverrideError
			if !errors.As(err, &rejection) || rejection.Code != tc.code {
				t.Fatalf("err = %v, want code %s", err, tc.code)
			}
			if router.Config.ProfileSource("chat") != ProfileSourceRoot || len(router.Config.dbProfileRows) != 0 {
				t.Fatalf("rejected batch changed state: chat=%s rows=%v", router.Config.ProfileSource("chat"), router.Config.dbProfileRows)
			}
		})
	}
	router := overrideValidationRouter()
	if err := router.SetProfileOverrides(map[string]string{"embed": "vllm", "coder": "vllm"}, ProfileOriginDB); err != nil {
		t.Fatalf("valid overrides rejected: %v", err)
	}
	if state := router.ProfileOverrideState("coder"); state.Source != ProfileSourceOverride || state.Origin != ProfileOriginDB || !state.DBRowPresent {
		t.Fatalf("coder state = %+v", state)
	}
}

func TestPersistableProfiles_NeverEmitsOverrideValues(t *testing.T) {
	clearProviderRoutingEnv(t)
	router := newRouterWithRoleRows(t, map[string]string{"coder": "vllm", "reviewer": "vllm"}, nil)
	router.Config.SetProfileOverride("creative", "vllm")
	persisted := router.Config.persistableProfiles()
	for _, profile := range []string{"coder", "creative"} {
		if persisted[profile] != "ollama" {
			t.Fatalf("%s persisted as %q, want shipped default ollama", profile, persisted[profile])
		}
	}
	if _, ok := persisted["reviewer"]; ok {
		t.Fatalf("override-only profile leaked into YAML: %v", persisted)
	}
}

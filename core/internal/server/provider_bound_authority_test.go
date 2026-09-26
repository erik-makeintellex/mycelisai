package server

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

var allExecutionProfiles = []string{"admin", "architect", "chat", "coder", "creative", "overseer", "sentry"}

func decodeProviderBound(t *testing.T, f *routingFixture, method, path, body string) providerBoundRejection {
	t.Helper()
	rr := doAuthenticatedRequest(t, f.mux, method, path, body)
	assertStatus(t, rr, http.StatusConflict)
	var got providerBoundRejection
	decodeOverrideData(t, rr, &got)
	if got.Code != providerBoundCode || got.RecommendedAction == "" {
		t.Fatalf("rejection = %+v", got)
	}
	return got
}

// The security-QA repro: disabling the provider chat resolves to must not
// leave chat provider_disabled. Every door (toggle, full update, delete)
// is refused and routing stays exactly as it was.
func TestProviderBound_RefusesDisablingOrDeletingTheChatProvider(t *testing.T) {
	doors := []struct{ name, method, path, body string }{
		{"toggle", http.MethodPut, "/api/v1/brains/vllm/toggle", `{"enabled":false}`},
		{"update omitting enabled", http.MethodPut, "/api/v1/brains/vllm", `{"type":"openai_compatible","endpoint":"http://127.0.0.1:9/v1","model_id":"m-vllm"}`},
		{"delete", http.MethodDelete, "/api/v1/brains/vllm", ""},
	}
	for _, door := range doors {
		t.Run(door.name, func(t *testing.T) {
			f := newRoutingFixture(t)
			got := decodeProviderBound(t, f, door.method, door.path, door.body)
			if got.ProviderID != "vllm" || !reflect.DeepEqual(got.Profiles, allExecutionProfiles) {
				t.Fatalf("rejection = %+v, want vllm bound to %v", got, allExecutionProfiles)
			}
			f.assertRoutingUnchanged(t)
			if a := f.s.Cognitive.ExecutionAvailability("chat", ""); !a.Available || a.ProviderID != "vllm" {
				t.Fatalf("chat availability = %+v", a)
			}
		})
	}
}

func TestProviderBound_FollowsOverridesAndClearsAfterRepoint(t *testing.T) {
	f := newRoutingFixture(t)
	if err := f.s.Cognitive.SetProfileOverride("coder", "ollama", cognitive.ProfileOriginDB); err != nil {
		t.Fatalf("pin coder: %v", err)
	}
	f.before = f.s.Cognitive.ConfigSnapshot()

	got := decodeProviderBound(t, f, http.MethodPut, "/api/v1/brains/ollama/toggle", `{"enabled":false}`)
	if !reflect.DeepEqual(got.Profiles, []string{"coder"}) {
		t.Fatalf("profiles = %v, want [coder]", got.Profiles)
	}
	decodeProviderBound(t, f, http.MethodDelete, "/api/v1/brains/ollama", "")
	f.assertRoutingUnchanged(t)

	// Re-point coder back to root; ollama is unbound and may be disabled.
	f.s.Cognitive.ClearProfileOverride("coder")
	expectAudit(f.mock)
	assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/ollama/toggle", `{"enabled":false}`), http.StatusOK)
	if p, _ := f.s.Cognitive.ProviderSnapshot("ollama"); p.Enabled {
		t.Fatal("ollama still enabled after re-point")
	}
}

func TestProviderBound_AllowsUnboundAndEnableChanges(t *testing.T) {
	t.Run("delete unbound provider", func(t *testing.T) {
		f := newRoutingFixture(t)
		expectAudit(f.mock)
		assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodDelete, "/api/v1/brains/cloud", ""), http.StatusOK)
		if _, ok := f.s.Cognitive.ProviderSnapshot("cloud"); ok {
			t.Fatal("cloud not deleted")
		}
	})
	t.Run("enable a disabled provider", func(t *testing.T) {
		f := newRoutingFixture(t)
		expectAudit(f.mock)
		assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/local-ollama-dev/toggle", `{"enabled":true}`), http.StatusOK)
		if p, _ := f.s.Cognitive.ProviderSnapshot("local-ollama-dev"); !p.Enabled {
			t.Fatal("local-ollama-dev not enabled")
		}
	})
	t.Run("update bound provider that stays enabled", func(t *testing.T) {
		f := newRoutingFixture(t)
		expectAudit(f.mock)
		assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/vllm",
			`{"type":"openai_compatible","endpoint":"http://127.0.0.1:9/v1","model_id":"m-vllm-2","enabled":true}`), http.StatusOK)
		if p, _ := f.s.Cognitive.ProviderSnapshot("vllm"); p.ModelID != "m-vllm-2" {
			t.Fatalf("vllm model = %q", p.ModelID)
		}
	})
}

// ── Mission-profile activation: all or nothing ──────────────────────────

func coderAndChatState(f *routingFixture) (cognitive.ProfileOverrideState, cognitive.ProfileOverrideState) {
	return f.s.Cognitive.ProfileOverrideState("coder"), f.s.Cognitive.ProfileOverrideState("chat")
}

func TestMissionProfileActivation_OneInvalidPairAppliesNothing(t *testing.T) {
	cases := []struct {
		name, roleProviders, code string
		want                      int
	}{
		{"disabled provider", `{"coder":"ollama","chat":"local-ollama-dev"}`, cognitive.ExecutionProviderDisabled, http.StatusBadRequest},
		{"missing provider", `{"coder":"ollama","chat":"ghost"}`, cognitive.ExecutionProviderMissing, http.StatusBadRequest},
		{"model-less provider", `{"coder":"ollama","chat":"nomodel"}`, cognitive.ExecutionModelMissing, http.StatusBadRequest},
		{"unknown profile", `{"coder":"ollama","reviewer":"ollama"}`, cognitive.OverrideUnknownProfile, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newRoutingFixture(t)
			expectActivationLoad(f.mock, tc.roleProviders)
			rr := doAuthenticatedRequest(t, f.mux, http.MethodPost, "/api/v1/mission-profiles/p-1/activate", "")
			assertStatus(t, rr, tc.want)
			var got profileOverrideRejection
			decodeOverrideData(t, rr, &got)
			if got.Code != tc.code {
				t.Fatalf("code = %q, want %q (%s)", got.Code, tc.code, rr.Body.String())
			}
			f.assertRoutingUnchanged(t)
			if coder, _ := coderAndChatState(f); coder.ProviderID != "vllm" || coder.Origin != "" {
				t.Fatalf("coder moved by rejected activation: %+v", coder)
			}
		})
	}
	t.Run("malformed role_providers", func(t *testing.T) {
		f := newRoutingFixture(t)
		expectActivationLoad(f.mock, `["coder","ollama"]`)
		assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPost, "/api/v1/mission-profiles/p-1/activate", ""), http.StatusBadRequest)
		f.assertRoutingUnchanged(t)
	})
}

func TestMissionProfileActivation_AppliesEveryPairAsRuntime(t *testing.T) {
	f := newRoutingFixture(t)
	expectActivationLoad(f.mock, `{"coder":"ollama","chat":"ollama"}`)
	expectAudit(f.mock)
	f.mock.ExpectBegin()
	f.mock.ExpectExec("UPDATE mission_profiles SET is_active=false").WillReturnResult(sqlmock.NewResult(0, 1))
	f.mock.ExpectExec("UPDATE mission_profiles SET is_active=true").WillReturnResult(sqlmock.NewResult(0, 1))
	f.mock.ExpectCommit()
	assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPost, "/api/v1/mission-profiles/p-1/activate", ""), http.StatusOK)
	coder, chat := coderAndChatState(f)
	for _, st := range []cognitive.ProfileOverrideState{coder, chat} {
		if st.ProviderID != "ollama" || st.Origin != cognitive.ProfileOriginRuntime || st.DBRowPresent {
			t.Fatalf("state = %+v, want ollama runtime without a DB row", st)
		}
	}
	if err := f.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
	if data, _ := os.ReadFile(f.s.Cognitive.ConfigPath); string(data) != "# test\n" {
		t.Fatalf("activation wrote cognitive.yaml: %q", data)
	}
}

func TestMissionProfileActivation_DBFailureAppliesNoRouting(t *testing.T) {
	f := newRoutingFixture(t)
	expectActivationLoad(f.mock, `{"coder":"ollama"}`)
	expectAudit(f.mock)
	f.mock.ExpectBegin().WillReturnError(errors.New("db down"))
	assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPost, "/api/v1/mission-profiles/p-1/activate", ""), http.StatusInternalServerError)
	f.assertRoutingUnchanged(t)
}

func TestMissionProfileActivation_OfflineRouterRefusesRouting(t *testing.T) {
	f := newRoutingFixture(t)
	f.s.Cognitive = nil
	expectActivationLoad(f.mock, `{"coder":"ollama"}`)
	assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPost, "/api/v1/mission-profiles/p-1/activate", ""), http.StatusServiceUnavailable)
	if err := f.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
}

// C1: a full PUT that keeps enabled=true but blanks the model leaves a
// bound provider non-executable (model_missing), so it is refused too.
func TestProviderBound_RefusesNonExecutableUpdateOfBoundProvider(t *testing.T) {
	for _, model := range []string{"", "   "} {
		t.Run("model="+strconv.Quote(model), func(t *testing.T) {
			f := newRoutingFixture(t)
			body := `{"type":"openai_compatible","endpoint":"http://127.0.0.1:9/v1","model_id":` + strconv.Quote(model) + `,"enabled":true}`
			got := decodeProviderBound(t, f, http.MethodPut, "/api/v1/brains/vllm", body)
			if got.ProviderID != "vllm" || !reflect.DeepEqual(got.Profiles, allExecutionProfiles) {
				t.Fatalf("rejection = %+v", got)
			}
			f.assertRoutingUnchanged(t)
			if a := f.s.Cognitive.ExecutionAvailability("chat", ""); !a.Available {
				t.Fatalf("chat availability = %+v", a)
			}
		})
	}
	t.Run("unbound provider may be blanked", func(t *testing.T) {
		f := newRoutingFixture(t)
		expectAudit(f.mock)
		assertStatus(t, doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/ollama",
			`{"type":"openai_compatible","endpoint":"http://127.0.0.1:9/v1","model_id":"","enabled":true}`), http.StatusOK)
		if p, _ := f.s.Cognitive.ProviderSnapshot("ollama"); p.ModelID != "" || !p.Enabled {
			t.Fatalf("ollama = %+v, want enabled with blank model", p)
		}
	})
}

// Pins providerConfigExecutable to the resolver's own predicate.
func TestProviderConfigExecutable_MatchesResolver(t *testing.T) {
	cases := []cognitive.ProviderConfig{
		{Type: "openai_compatible", ModelID: "m", Enabled: true},
		{Type: "openai_compatible", ModelID: "m", Enabled: false},
		{Type: "openai_compatible", ModelID: "", Enabled: true},
		{Type: "openai_compatible", ModelID: " \t", Enabled: true},
		{Type: "openai_compatible", ModelID: "", Enabled: false},
	}
	for i, cfg := range cases {
		router := &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Providers: map[string]cognitive.ProviderConfig{"p": cfg},
				Profiles:  map[string]string{"chat": "p"},
			},
			Adapters: map[string]cognitive.LLMProvider{"p": &stubAdapter{healthy: true}},
		}
		if got, want := providerConfigExecutable(cfg), router.ExecutionAvailability("chat", "").Available; got != want {
			t.Errorf("case %d %+v: providerConfigExecutable=%v, resolver available=%v", i, cfg, got, want)
		}
	}
}

// failedAfterCommitAudit matches an audit context JSON carrying the F2 fields.
type failedAfterCommitAudit struct{}

func (failedAfterCommitAudit) Match(v driver.Value) bool {
	b, ok := v.([]byte)
	s := string(b)
	return ok && strings.Contains(s, `"result_status":"failed_after_commit"`) &&
		strings.Contains(s, `"mission_profile_id":"p-1"`) && strings.Contains(s, `"recovery_action"`) &&
		strings.Contains(s, `"action":"mission_profile_activated"`)
}

// F2: a provider disabled between the commit and the runtime apply yields
// 500 mission_profile_routing_not_applied plus a failure audit event.
func TestMissionProfileActivation_PostCommitFailureIsAudited(t *testing.T) {
	f := newRoutingFixture(t)
	expectActivationLoad(f.mock, `{"coder":"ollama"}`)
	expectAudit(f.mock)
	f.mock.ExpectBegin()
	f.mock.ExpectExec("UPDATE mission_profiles SET is_active=false").WillDelayFor(300 * time.Millisecond).WillReturnResult(sqlmock.NewResult(0, 1))
	f.mock.ExpectExec("UPDATE mission_profiles SET is_active=true").WillReturnResult(sqlmock.NewResult(0, 1))
	f.mock.ExpectCommit()
	f.mock.ExpectExec("INSERT INTO log_entries").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", routingMutationAuditSource, sqlmock.AnyArg(), sqlmock.AnyArg(), failedAfterCommitAudit{}).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// An unserialized writer disables ollama inside the delayed commit.
	go func() {
		time.Sleep(100 * time.Millisecond)
		p, _ := f.s.Cognitive.ProviderSnapshot("ollama")
		p.Enabled = false
		f.s.Cognitive.StoreProviderConfig("ollama", p)
	}()
	rr := doAuthenticatedRequest(t, f.mux, http.MethodPost, "/api/v1/mission-profiles/p-1/activate", "")
	assertStatus(t, rr, http.StatusInternalServerError)
	var got struct {
		Code              string `json:"code"`
		RecommendedAction string `json:"recommended_action"`
	}
	decodeOverrideData(t, rr, &got)
	if got.Code != missionProfileRoutingNotApplied || got.RecommendedAction == "" {
		t.Fatalf("data = %+v", got)
	}
	if st := f.s.Cognitive.ProfileOverrideState("coder"); st.ProviderID != "vllm" || st.Origin != "" {
		t.Fatalf("coder moved: %+v", st)
	}
	if err := f.mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql (failure audit missing?): %v", err)
	}
}

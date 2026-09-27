package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// profileOverrideRouter: root vllm (healthy), an enabled ollama, a disabled
// local-ollama-dev, a model-less provider, and a leaves_org cloud provider.
func profileOverrideRouter(t *testing.T) *cognitive.Router {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "cognitive.yaml")
	if err := os.WriteFile(cfgPath, []byte("# test\n"), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	profiles := map[string]string{}
	sources := map[string]string{}
	for _, p := range []string{"admin", "architect", "chat", "coder", "creative", "overseer", "sentry"} {
		profiles[p], sources[p] = "vllm", cognitive.ProfileSourceRoot
	}
	return &cognitive.Router{
		ConfigPath: cfgPath,
		Config: &cognitive.BrainConfig{
			RootProvider: "vllm",
			Providers: map[string]cognitive.ProviderConfig{
				"vllm":             {Type: "openai_compatible", Endpoint: "http://127.0.0.1:9/v1", ModelID: "m-vllm", Enabled: true},
				"ollama":           {Type: "ollama", Endpoint: "http://127.0.0.1:9/v1", ModelID: "m-ollama", Enabled: true},
				"local-ollama-dev": {Type: "ollama", Endpoint: "http://127.0.0.1:9/v1", ModelID: "m-dev", Enabled: false},
				"nomodel":          {Type: "ollama", Endpoint: "http://127.0.0.1:9/v1", Enabled: true},
				"cloud":            {Type: "openai", ModelID: "c", Enabled: true, DataBoundary: cognitive.DataBoundaryLeavesOrg},
			},
			Profiles:         profiles,
			ProfileSources:   sources,
			ProfileFallbacks: map[string][]string{"architect": {"ollama"}},
		},
		Adapters: map[string]cognitive.LLMProvider{
			"vllm": &stubAdapter{healthy: true}, "ollama": &stubAdapter{healthy: true},
			"local-ollama-dev": &stubAdapter{healthy: true}, "nomodel": &stubAdapter{}, "cloud": &stubAdapter{},
		},
	}
}

func profileOverrideServer(t *testing.T) (*AdminServer, sqlmock.Sqlmock, *http.ServeMux) {
	t.Helper()
	dbOpt, mock := withDirectDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) { s.Cognitive = profileOverrideRouter(t) })
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/cognitive/profiles", s.HandleUpdateProfiles)
	mux.HandleFunc("DELETE /api/v1/cognitive/profiles/{profile}/override", s.HandleClearProfileOverride)
	mux.HandleFunc("PUT /api/v1/cognitive/providers/{id}", s.HandleUpdateProvider)
	mux.HandleFunc("GET /api/v1/cognitive/status", s.HandleCognitiveStatus)
	return s, mock, mux
}

func expectAudit(mock sqlmock.Sqlmock) {
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
}

func decodeOverrideData(t *testing.T, rr *httptest.ResponseRecorder, into any) {
	t.Helper()
	var env struct {
		OK   bool            `json:"ok"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v body=%s", err, rr.Body.String())
	}
	if err := json.Unmarshal(env.Data, into); err != nil {
		t.Fatalf("decode data: %v body=%s", err, rr.Body.String())
	}
}

func TestCognitiveProfileAuthority_RejectsAnonymousStandardAndUnscopedAdmin(t *testing.T) {
	routes := []struct{ method, path, body string }{
		{http.MethodDelete, "/api/v1/cognitive/profiles/coder/override", ""},
		{http.MethodPut, "/api/v1/cognitive/profiles", `{"profiles":{"coder":"ollama"}}`},
		{http.MethodPut, "/api/v1/cognitive/providers/vllm", `{"endpoint":"http://attacker.invalid/v1","model_id":"evil"}`},
	}
	callers := []struct {
		name     string
		identity *RequestIdentity
		want     int
	}{
		{"anonymous", nil, http.StatusUnauthorized},
		{"standard user", &RequestIdentity{UserID: "u-1", Role: "user", Scopes: []string{"soma:work"}}, http.StatusForbidden},
		{"admin without cognitive:write", &RequestIdentity{UserID: "u-2", Role: "admin", Scopes: []string{"config_documents:write", "cognitive:read"}}, http.StatusForbidden},
		{"scoped non-admin", &RequestIdentity{UserID: "u-3", Role: "user", Scopes: []string{"cognitive:write"}}, http.StatusForbidden},
	}
	for _, route := range routes {
		for _, caller := range callers {
			t.Run(route.method+" "+route.path+" "+caller.name, func(t *testing.T) {
				s, mock, mux := profileOverrideServer(t)
				s.Cognitive.SetProfileOverride("coder", "ollama", cognitive.ProfileOriginDB)
				var rr *httptest.ResponseRecorder
				if caller.identity == nil {
					rr = doRequest(t, mux, route.method, route.path, route.body)
				} else {
					rr = doAuthenticatedRequestAs(t, mux, route.method, route.path, route.body, caller.identity)
				}
				assertStatus(t, rr, caller.want)
				if denialLeaksData(rr.Body.String()) {
					t.Fatalf("rejected call leaked data: %s", rr.Body.String())
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("db touched: %v", err)
				}
				if state := s.Cognitive.ProfileOverrideState("coder"); state.ProviderID != "ollama" || !state.DBRowPresent {
					t.Fatalf("coder changed by rejected call: %+v", state)
				}
				if p, _ := s.Cognitive.ProviderSnapshot("vllm"); p.Endpoint != "http://127.0.0.1:9/v1" {
					t.Fatalf("provider changed by rejected call: %+v", p)
				}
			})
		}
	}
}

func TestCognitiveProfileOverride_DeleteRejectsBadNamesWithoutDB(t *testing.T) {
	for _, name := range []string{"..%2Fx", "role.x", "reviewer", "Chat", strings.Repeat("a", 33)} {
		t.Run(name, func(t *testing.T) {
			_, mock, mux := profileOverrideServer(t)
			rr := doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/profiles/"+name+"/override", "")
			assertStatus(t, rr, http.StatusNotFound)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("db touched: %v", err)
			}
		})
	}
}

func TestCognitiveProfileOverride_DeleteResetsToRootAndIsIdempotent(t *testing.T) {
	s, mock, mux := profileOverrideServer(t)
	s.Cognitive.SetProfileOverride("coder", "ollama", cognitive.ProfileOriginDB)

	expectAudit(mock)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM system_config").WithArgs("coder").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	rr := doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/profiles/coder/override", "")
	assertStatus(t, rr, http.StatusOK)
	var first profileOverrideResult
	decodeOverrideData(t, rr, &first)
	if !first.Changed || first.Effective.ProviderID != "vllm" || first.Effective.Source != cognitive.ProfileSourceRoot || !first.Effective.Available {
		t.Fatalf("first delete = %+v", first)
	}

	expectAudit(mock)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM system_config").WithArgs("coder").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	rr = doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/profiles/coder/override", "")
	assertStatus(t, rr, http.StatusOK)
	var second profileOverrideResult
	decodeOverrideData(t, rr, &second)
	if second.Changed || second.Effective.ProviderID != "vllm" {
		t.Fatalf("second delete = %+v, want changed=false", second)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
}

func TestCognitiveProfileOverride_DeleteFailsClosedOnAuditTxAndNilDB(t *testing.T) {
	t.Run("audit failure", func(t *testing.T) {
		s, mock, mux := profileOverrideServer(t)
		s.Cognitive.SetProfileOverride("coder", "ollama", cognitive.ProfileOriginDB)
		mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
		assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/profiles/coder/override", ""), http.StatusServiceUnavailable)
		if state := s.Cognitive.ProfileOverrideState("coder"); state.ProviderID != "ollama" || !state.DBRowPresent {
			t.Fatalf("memory changed after audit failure: %+v", state)
		}
	})
	t.Run("tx failure", func(t *testing.T) {
		s, mock, mux := profileOverrideServer(t)
		s.Cognitive.SetProfileOverride("coder", "ollama", cognitive.ProfileOriginDB)
		expectAudit(mock)
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM system_config").WillReturnError(errors.New("tx down"))
		mock.ExpectRollback()
		assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/profiles/coder/override", ""), http.StatusServiceUnavailable)
		if state := s.Cognitive.ProfileOverrideState("coder"); state.ProviderID != "ollama" || !state.DBRowPresent {
			t.Fatalf("memory changed after tx failure: %+v", state)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql: %v", err)
		}
	})
	t.Run("nil db", func(t *testing.T) {
		s := newTestServer(func(s *AdminServer) { s.Cognitive = profileOverrideRouter(t) })
		mux := setupMux(t, "DELETE /api/v1/cognitive/profiles/{profile}/override", s.HandleClearProfileOverride)
		assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodDelete, "/api/v1/cognitive/profiles/coder/override", ""), http.StatusServiceUnavailable)
		mux = setupMux(t, "PUT /api/v1/cognitive/profiles", s.HandleUpdateProfiles)
		assertStatus(t, doAuthenticatedRequest(t, mux, http.MethodPut, "/api/v1/cognitive/profiles", `{"profiles":{"coder":"ollama"}}`), http.StatusServiceUnavailable)
		if state := s.Cognitive.ProfileOverrideState("coder"); state.Source != cognitive.ProfileSourceRoot {
			t.Fatalf("nil-DB PUT changed memory: %+v", state)
		}
	})
}

func TestCognitiveProfileOverride_PutPersistsToSystemConfig(t *testing.T) {
	s, mock, mux := profileOverrideServer(t)
	expectAudit(mock)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO system_config").WithArgs("role.chat", "ollama").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO system_config").WithArgs("role.coder", "ollama").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	rr := doAuthenticatedRequest(t, mux, http.MethodPut, "/api/v1/cognitive/profiles", `{"profiles":{"coder":"ollama","chat":"ollama"}}`)
	assertStatus(t, rr, http.StatusOK)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql: %v", err)
	}
	state := s.Cognitive.ProfileOverrideState("coder")
	if state.ProviderID != "ollama" || state.Origin != cognitive.ProfileOriginDB || !state.DBRowPresent {
		t.Fatalf("coder state = %+v", state)
	}
	if saved, _ := os.ReadFile(s.Cognitive.ConfigPath); strings.Contains(string(saved), "ollama") {
		t.Fatalf("PUT wrote profiles to YAML:\n%s", saved)
	}
}

func TestCognitiveProfileOverride_PutRejectsUnrunnableAndWritesNothing(t *testing.T) {
	cases := map[string]struct {
		body string
		want int
		code string
	}{
		"disabled":            {`{"profiles":{"coder":"local-ollama-dev"}}`, http.StatusBadRequest, cognitive.ExecutionProviderDisabled},
		"unknown":             {`{"profiles":{"coder":"ghost"}}`, http.StatusBadRequest, cognitive.ExecutionProviderMissing},
		"model-less":          {`{"profiles":{"coder":"nomodel"}}`, http.StatusBadRequest, cognitive.ExecutionModelMissing},
		"crosses fallbacks":   {`{"profiles":{"architect":"cloud"}}`, http.StatusBadRequest, cognitive.OverrideBoundaryMismatch},
		"one bad in batch":    {`{"profiles":{"chat":"ollama","coder":"local-ollama-dev"}}`, http.StatusBadRequest, cognitive.ExecutionProviderDisabled},
		"key-shaped name":     {`{"profiles":{"role.x":"ollama"}}`, http.StatusBadRequest, cognitive.OverrideInvalidProfile},
		"unknown profile":     {`{"profiles":{"reviewer":"ollama"}}`, http.StatusBadRequest, cognitive.OverrideUnknownProfile},
		"env pinned conflict": {`{"profiles":{"sentry":"ollama"}}`, http.StatusConflict, cognitive.OverrideEnvPinned},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, mock, mux := profileOverrideServer(t)
			s.Cognitive.Config.ProfileOverrideOrigins = map[string]string{"sentry": cognitive.ProfileOriginEnv}
			rr := doAuthenticatedRequest(t, mux, http.MethodPut, "/api/v1/cognitive/profiles", tc.body)
			assertStatus(t, rr, tc.want)
			var rejection profileOverrideRejection
			decodeOverrideData(t, rr, &rejection)
			if rejection.Code != tc.code {
				t.Fatalf("code = %q, want %q", rejection.Code, tc.code)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("db touched: %v", err)
			}
			for _, p := range []string{"chat", "coder", "architect"} {
				if state := s.Cognitive.ProfileOverrideState(p); state.Source != cognitive.ProfileSourceRoot {
					t.Fatalf("%s changed by rejected PUT: %+v", p, state)
				}
			}
		})
	}
}

func TestCognitiveStatus_ChatMisroutedIsOfflineEvenWhenVLLMIsUp(t *testing.T) {
	s, _, mux := profileOverrideServer(t)
	s.Cognitive.Config.Profiles["chat"] = "local-ollama-dev"
	s.Cognitive.Config.ProfileSources["chat"] = cognitive.ProfileSourceOverride
	s.Cognitive.Config.ProfileOverrideOrigins = map[string]string{"chat": cognitive.ProfileOriginDB}

	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/cognitive/status", "")
	assertStatus(t, rr, http.StatusOK)
	var resp struct {
		Text               cognitiveTextStatus               `json:"text"`
		Profiles           map[string]cognitive.ProfileRoute `json:"profiles"`
		ProfileRouteHealth string                            `json:"profile_route_health"`
		OverlayError       *bool                             `json:"overlay_error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Text.Status == "online" || !strings.Contains(resp.Text.RecommendedAction, "Reset the chat override") {
		t.Fatalf("text = %+v, want offline with reset hint", resp.Text)
	}
	chat := resp.Profiles["chat"]
	if chat.Available || chat.Code != cognitive.ExecutionProviderDisabled || chat.OverrideOrigin != "db" || chat.Reachable != nil {
		t.Fatalf("chat route = %+v", chat)
	}
	if coder := resp.Profiles["coder"]; !coder.Available || coder.Reachable == nil || !*coder.Reachable || coder.Source != "root" {
		t.Fatalf("coder route = %+v", coder)
	}
	if resp.ProfileRouteHealth != cognitive.ProfileRouteHealthFailed || resp.OverlayError == nil || *resp.OverlayError {
		t.Fatalf("health = %s overlay=%v", resp.ProfileRouteHealth, resp.OverlayError)
	}
	profilesJSON, _ := json.Marshal(resp.Profiles)
	if strings.Contains(string(profilesJSON), "http") || strings.Contains(string(profilesJSON), "api_key") {
		t.Fatalf("profiles leak endpoints or secrets: %s", profilesJSON)
	}
}

func TestCognitiveStatus_AllRootReachableIsOK(t *testing.T) {
	_, _, mux := profileOverrideServer(t)
	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/cognitive/status", "")
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["profile_route_health"] != "ok" || resp["text"].(map[string]any)["status"] != "online" {
		t.Fatalf("status = %v text=%v", resp["profile_route_health"], resp["text"])
	}
}

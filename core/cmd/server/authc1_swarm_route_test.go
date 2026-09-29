package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/swarm"
)

// TestAuthC1SomaRuntimeMountsNoRawTeamSpawn proves, through the real mux that
// startSomaRuntime builds, that Soma no longer mounts its unauthenticated
// manifest-spawn handler (AUTH-C1 A1). The only /api/swarm/teams handler is
// the root-admin route AdminServer.RegisterRoutes adds, and a request that
// would have spawned a team here reaches no handler and starts nothing.
func TestAuthC1SomaRuntimeMountsNoRawTeamSpawn(t *testing.T) {
	nc := legStartTestNATS(t)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /authc1-sentinel", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	core := &coreRuntime{NC: nc, Guard: &governance.Guard{}}
	selection := &bootstrap.StartupSelection{Bundle: &bootstrap.TemplateBundle{ID: "authc1-route-test"}}
	soma := startSomaRuntime(t.Context(), mux, core, selection, swarm.NewRegistry(t.TempDir()), productServices{})
	if soma == nil {
		t.Fatal("expected Soma to start with a live NATS connection")
	}
	t.Cleanup(soma.Shutdown)

	srv := httptest.NewServer(mux)
	defer srv.Close()
	if resp, err := http.Get(srv.URL + "/authc1-sentinel"); err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("sentinel route did not resolve: %v %v", resp, err)
	}
	for _, method := range []string{http.MethodPost, http.MethodGet} {
		req, _ := http.NewRequest(method, srv.URL+"/api/swarm/teams", strings.NewReader(`{"id":"authc1-raw","name":"Raw","type":"action"}`))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s /api/swarm/teams: %v", method, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s /api/swarm/teams on the Soma mux = %d, want 404", method, resp.StatusCode)
		}
	}
	for _, team := range soma.ListTeams() {
		if team.ID == "authc1-raw" {
			t.Fatal("raw POST spawned a runtime team")
		}
	}
}

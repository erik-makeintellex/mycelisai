package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

// providerInheritingManifest relies on Start filling the member provider from
// the team provider, which mutates the running manifest.
func providerInheritingManifest() *TeamManifest {
	manifest := lifecycleManifest("inherit-team")
	manifest.Provider = "team-provider"
	manifest.Members = []protocol.AgentManifest{{ID: "inherit-lead", Role: "lead"}}
	return manifest
}

func inheritingBrain() *cognitive.Router {
	return &cognitive.Router{
		Config: &cognitive.BrainConfig{Providers: map[string]cognitive.ProviderConfig{
			"team-provider": {Enabled: true, ModelID: "model-a", Location: "local"},
		}},
		Adapters: map[string]cognitive.LLMProvider{"team-provider": teamProviderStub{}},
	}
}

func TestRestartThenIdenticalRetryIsIdempotent(t *testing.T) {
	store := newLockedDurableTeamStore()
	_, nc := startTestNATS(t)
	first := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), inheritingBrain(), nil, nil, nil)
	first.SetDurableTeamStore(store)
	if err := first.SpawnTeamContext(context.Background(), providerInheritingManifest()); err != nil {
		t.Fatalf("first spawn: %v", err)
	}
	first.Shutdown()
	persisted, _ := store.LoadRuntimeTeams(context.Background())
	if len(persisted) != 1 || persisted[0].Members[0].Provider != "" {
		t.Fatalf("persisted manifest is not the approved pre-start form: %#v", persisted)
	}

	restarted := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), inheritingBrain(), nil, nil, nil)
	restarted.SetDurableTeamStore(store)
	if err := restarted.Start(); err != nil {
		t.Fatalf("restart: %v", err)
	}
	t.Cleanup(restarted.Shutdown)
	if err := restarted.SpawnTeamContext(context.Background(), providerInheritingManifest()); err != nil {
		t.Fatalf("identical retry after restart: %v", err)
	}
	if store.saveCount() != 1 {
		t.Fatalf("saves = %d, want 1", store.saveCount())
	}
	changed := providerInheritingManifest()
	changed.Description = "different"
	if err := restarted.SpawnTeamContext(context.Background(), changed); !errors.Is(err, ErrRuntimeTeamManifestConflict) {
		t.Fatalf("conflicting retry after restart = %v", err)
	}
}

type failingSaveStore struct{ *lockedDurableTeamStore }

func (failingSaveStore) SaveRuntimeTeam(context.Context, *TeamManifest) error {
	return errors.New("pq: connection to 10.0.0.7 refused (internal detail)")
}

func TestHandleCreateTeamNormalizesFailures(t *testing.T) {
	cases := []struct {
		name   string
		store  DurableTeamStore
		seed   string
		body   string
		status int
		leak   string
	}{
		{"conflict", newLockedDurableTeamStore(), `{"id":"t1","name":"T1","description":"a"}`,
			`{"id":"t1","name":"T1","description":"b"}`, http.StatusConflict, "t1"},
		{"persistence", failingSaveStore{newLockedDurableTeamStore()}, "",
			`{"id":"t2","name":"T2"}`, http.StatusServiceUnavailable, "10.0.0.7"},
		{"runtime", newLockedDurableTeamStore(), "",
			`{"id":"t3","name":"T3","inputs":["swarm.team.t3.extra."]}`, http.StatusServiceUnavailable, "subject"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			soma, _ := newLifecycleSoma(t, tc.store)
			post := func(body string) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				soma.HandleCreateTeam(rec, httptest.NewRequest(http.MethodPost, "/api/swarm/teams", strings.NewReader(body)))
				return rec
			}
			if tc.seed != "" {
				if rec := post(tc.seed); rec.Code != http.StatusCreated {
					t.Fatalf("seed = %d", rec.Code)
				}
			}
			rec := post(tc.body)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tc.status, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), tc.leak) {
				t.Fatalf("response leaked internal detail %q: %s", tc.leak, rec.Body.String())
			}
		})
	}
}

func TestRestorationHealthSummarizesDegradations(t *testing.T) {
	soma := NewTestSoma(nil)
	if status, _ := soma.RestorationHealth(); status != "online" {
		t.Fatalf("clean status = %s", status)
	}
	soma.recordRestorationDegradation("bad-a", "pq: secret internal reason")
	soma.recordRestorationDegradation("bad-b", "digest")
	status, detail := soma.RestorationHealth()
	if status != "degraded" || !strings.Contains(detail, "2 runtime team(s)") ||
		!strings.Contains(detail, "bad-a") || strings.Contains(detail, "secret") {
		t.Fatalf("health = %s %q", status, detail)
	}
	raw, _ := json.Marshal(soma.RestorationDegradations())
	if !strings.Contains(string(raw), `"team_id":"bad-b"`) {
		t.Fatalf("degradations = %s", raw)
	}
}

func TestActivateBlueprintAndSpawnSameIDNeverOrphanATeam(t *testing.T) {
	_, nc := startTestNATS(t)
	bp := &protocol.MissionBlueprint{MissionID: "m1", Teams: []protocol.BlueprintTeam{{
		Name: "builder", Role: "build", Agents: []protocol.AgentManifest{{ID: "m1-builder-lead", Role: "lead"}},
	}}}
	baseline := nc.NumSubscriptions()
	for i := 0; i < 30; i++ {
		soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromRuntimeOrganization(&RuntimeOrganization{}), nil, nil, nil, nil)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); soma.ActivateBlueprint(bp, nil) }()
		go func() {
			defer wg.Done()
			_ = soma.SpawnTeamContext(context.Background(), lifecycleManifest("m1.builder"))
		}()
		wg.Wait()
		if n := len(soma.ListTeams()); n != 1 {
			t.Fatalf("iteration %d: running teams = %d", i, n)
		}
		soma.StopTeam("m1.builder")
		soma.Shutdown()
		if got := nc.NumSubscriptions(); got != baseline {
			t.Fatalf("iteration %d: orphaned subscriptions %d, want %d", i, got, baseline)
		}
	}
}

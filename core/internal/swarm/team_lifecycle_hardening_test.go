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
	"time"

	"github.com/mycelis/core/internal/governance"
	"github.com/nats-io/nats.go"
)

func newLifecycleSoma(t *testing.T, store DurableTeamStore) (*Soma, *nats.Conn) {
	t.Helper()
	_, nc := startTestNATS(t)
	soma := NewSoma(nc, &governance.Guard{}, NewRegistryFromManifests(nil), nil, nil, nil, nil)
	if store != nil {
		soma.SetDurableTeamStore(store)
	}
	t.Cleanup(soma.Shutdown)
	return soma, nc
}

func lifecycleManifest(teamID string) *TeamManifest {
	return &TeamManifest{ID: teamID, Name: teamID, Type: TeamTypeAction, Description: "Lifecycle proof team."}
}

func teamIDs(soma *Soma) []string {
	ids := []string{}
	for _, manifest := range soma.ListTeams() {
		ids = append(ids, manifest.ID)
	}
	return ids
}

func TestSpawnTeamStartFailureReleasesSubscriptions(t *testing.T) {
	store := newLockedDurableTeamStore()
	soma, nc := newLifecycleSoma(t, store)
	baseline := nc.NumSubscriptions()

	broken := lifecycleManifest("broken-team")
	// The canonical command subject subscribes first; the trailing-dot input then fails.
	broken.Inputs = []string{"swarm.team.broken-team.extra."}
	if err := soma.SpawnTeamContext(context.Background(), broken); err == nil {
		t.Fatal("expected Start failure")
	}
	if got := nc.NumSubscriptions(); got != baseline {
		t.Fatalf("live subscriptions after failed start = %d, want %d", got, baseline)
	}
	if ids := teamIDs(soma); len(ids) != 0 {
		t.Fatalf("failed team registered: %v", ids)
	}
	if store.saveCount() != 0 {
		t.Fatalf("failed team persisted %d times", store.saveCount())
	}
	if err := soma.SpawnTeamContext(context.Background(), lifecycleManifest("broken-team")); err != nil {
		t.Fatalf("same id after failed start: %v", err)
	}
}

func TestSomaBootStartFailureIsIsolatedAndNotRegistered(t *testing.T) {
	broken := lifecycleManifest("boot-broken")
	broken.Inputs = []string{"swarm.team.boot-broken.extra."}
	store := newLockedDurableTeamStore(broken, lifecycleManifest("boot-good"))
	soma, _ := newLifecycleSoma(t, store)
	if err := soma.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if ids := teamIDs(soma); len(ids) != 1 || ids[0] != "boot-good" {
		t.Fatalf("restored teams = %v, want [boot-good]", ids)
	}
	degraded := soma.RestorationDegradations()
	if len(degraded) != 1 || degraded[0].TeamID != "boot-broken" {
		t.Fatalf("degradations = %#v", degraded)
	}
}

func TestSpawnTeamIdenticalRetryIsIdempotent(t *testing.T) {
	store := newLockedDurableTeamStore()
	soma, nc := newLifecycleSoma(t, store)
	if err := soma.SpawnTeamContext(context.Background(), completeDurableManifest()); err != nil {
		t.Fatalf("first spawn: %v", err)
	}
	afterFirst := nc.NumSubscriptions()
	if err := soma.SpawnTeamContext(context.Background(), completeDurableManifest()); err != nil {
		t.Fatalf("identical retry: %v", err)
	}
	if got := nc.NumSubscriptions(); got != afterFirst {
		t.Fatalf("identical retry changed subscriptions %d -> %d", afterFirst, got)
	}
	if store.saveCount() != 1 || len(soma.ListTeams()) != 1 {
		t.Fatalf("saves=%d teams=%d, want 1/1", store.saveCount(), len(soma.ListTeams()))
	}
}

func TestSpawnTeamConflictingRetryFailsClosed(t *testing.T) {
	store := newLockedDurableTeamStore()
	soma, _ := newLifecycleSoma(t, store)
	if err := soma.SpawnTeamContext(context.Background(), completeDurableManifest()); err != nil {
		t.Fatalf("first spawn: %v", err)
	}
	changed := completeDurableManifest()
	changed.Description = "A different approved delivery."
	err := soma.SpawnTeamContext(context.Background(), changed)
	if !errors.Is(err, ErrRuntimeTeamManifestConflict) {
		t.Fatalf("err = %v, want ErrRuntimeTeamManifestConflict", err)
	}
	running := soma.ListTeams()
	if len(running) != 1 || running[0].Description != completeDurableManifest().Description {
		t.Fatalf("first team changed: %#v", running)
	}
	if store.saveCount() != 1 {
		t.Fatalf("saves = %d, want 1", store.saveCount())
	}
}

func TestHandleCreateTeamMapsRetryOutcomes(t *testing.T) {
	soma, _ := newLifecycleSoma(t, newLockedDurableTeamStore())
	post := func(body string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/swarm/teams", strings.NewReader(body))
		rec := httptest.NewRecorder()
		soma.HandleCreateTeam(rec, req)
		return rec.Code
	}
	body := `{"id":"http-team","name":"HTTP Team","description":"first"}`
	if got := post(body); got != http.StatusCreated {
		t.Fatalf("first = %d", got)
	}
	if got := post(body); got != http.StatusOK {
		t.Fatalf("identical retry = %d, want 200", got)
	}
	if got := post(`{"id":"http-team","name":"HTTP Team","description":"second"}`); got != http.StatusConflict {
		t.Fatalf("conflicting retry = %d, want 409", got)
	}
}

func TestTeamStartStopConcurrentDoesNotLeakSubscriptions(t *testing.T) {
	_, nc := startTestNATS(t)
	baseline := nc.NumSubscriptions()
	for i := 0; i < 25; i++ {
		team := NewTeam(lifecycleManifest("race-team"), nc, nil, nil)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _ = team.Start() }()
		go func() { defer wg.Done(); team.Stop() }()
		wg.Wait()
		if got := nc.NumSubscriptions(); got != baseline {
			t.Fatalf("iteration %d leaked subscriptions: %d, want %d", i, got, baseline)
		}
	}
}

func TestStopTeamDurablyDoesNotBlockReadsDuringDelete(t *testing.T) {
	store := newLockedDurableTeamStore()
	soma, _ := newLifecycleSoma(t, store)
	if err := soma.SpawnTeamContext(context.Background(), lifecycleManifest("slow-delete")); err != nil {
		t.Fatal(err)
	}
	store.blockDelete = make(chan struct{})
	store.deleteEntered = make(chan struct{})
	var release sync.Once
	unblock := func() { release.Do(func() { close(store.blockDelete) }) }
	t.Cleanup(unblock)
	done := make(chan error, 1)
	go func() { _, err := soma.StopTeamDurably("slow-delete"); done <- err }()
	<-store.deleteEntered
	read := make(chan int, 1)
	go func() { read <- len(soma.ListTeams()) }()
	select {
	case n := <-read:
		if n != 1 {
			t.Fatalf("team removed before durable delete committed: %d", n)
		}
	case <-time.After(time.Second):
		t.Fatal("ListTeams blocked while durable delete was in flight")
	}
	unblock()
	if err := <-done; err != nil {
		t.Fatalf("StopTeamDurably: %v", err)
	}
	if ids := teamIDs(soma); len(ids) != 0 {
		t.Fatalf("teams after stop = %v", ids)
	}
}

func TestStopTeamDurablyBoundsDeleteAndKeepsTeamOnFailure(t *testing.T) {
	store := newLockedDurableTeamStore()
	soma, nc := newLifecycleSoma(t, store)
	if err := soma.SpawnTeamContext(context.Background(), lifecycleManifest("timeout-team")); err != nil {
		t.Fatal(err)
	}
	subs := nc.NumSubscriptions()
	store.requireDeadline = true
	found, err := soma.StopTeamDurably("timeout-team")
	if !found || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("StopTeamDurably = (%v, %v), want (true, deadline exceeded)", found, err)
	}
	if !store.sawDeadline() {
		t.Fatal("durable delete ran without a bounded context")
	}
	if ids := teamIDs(soma); len(ids) != 1 || nc.NumSubscriptions() != subs || !store.has("timeout-team") {
		t.Fatalf("team not kept running and durable: ids=%v subs=%d store=%v", ids, nc.NumSubscriptions(), store.has("timeout-team"))
	}
}

func TestConcurrentStopAndSpawnSameIDKeepRuntimeAndDurableStateAligned(t *testing.T) {
	store := newLockedDurableTeamStore()
	soma, _ := newLifecycleSoma(t, store)
	for i := 0; i < 20; i++ {
		if err := soma.SpawnTeamContext(context.Background(), lifecycleManifest("contended")); err != nil {
			t.Fatalf("seed spawn: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _, _ = soma.StopTeamDurably("contended") }()
		go func() {
			defer wg.Done()
			_ = soma.SpawnTeamContext(context.Background(), lifecycleManifest("contended"))
		}()
		wg.Wait()
		running := len(soma.ListTeams())
		if running > 1 || (running == 1) != store.has("contended") {
			t.Fatalf("iteration %d: running=%d durable=%v", i, running, store.has("contended"))
		}
		_, _ = soma.StopTeamDurably("contended")
	}
}

func TestSomaHasTeam(t *testing.T) {
	soma := NewTestSoma([]*TeamManifest{{ID: "known"}})
	if !soma.HasTeam(" known ") || soma.HasTeam("unknown") || soma.HasTeam("") {
		t.Fatal("HasTeam did not reflect the runtime registry")
	}
}

type lockedDurableTeamStore struct {
	mu              sync.Mutex
	manifests       map[string][]byte
	order           []string
	saves           int
	requireDeadline bool
	deadlineSeen    bool
	blockDelete     chan struct{}
	deleteEntered   chan struct{}
}

func newLockedDurableTeamStore(seed ...*TeamManifest) *lockedDurableTeamStore {
	store := &lockedDurableTeamStore{manifests: map[string][]byte{}}
	for _, manifest := range seed {
		raw, _ := json.Marshal(manifest)
		store.manifests[manifest.ID] = raw
		store.order = append(store.order, manifest.ID)
	}
	return store
}

func (s *lockedDurableTeamStore) LoadRuntimeTeams(context.Context) ([]*TeamManifest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []*TeamManifest{}
	for _, id := range s.order {
		if raw, ok := s.manifests[id]; ok {
			var manifest TeamManifest
			_ = json.Unmarshal(raw, &manifest)
			out = append(out, &manifest)
		}
	}
	return out, nil
}

func (s *lockedDurableTeamStore) SaveRuntimeTeam(_ context.Context, manifest *TeamManifest) error {
	raw, _ := json.Marshal(manifest)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saves++
	if existing, ok := s.manifests[manifest.ID]; ok && string(existing) != string(raw) {
		return errors.New("different approved manifest")
	}
	if _, ok := s.manifests[manifest.ID]; !ok {
		s.order = append(s.order, manifest.ID)
	}
	s.manifests[manifest.ID] = raw
	return nil
}

func (s *lockedDurableTeamStore) DeleteRuntimeTeam(ctx context.Context, teamID string) error {
	if s.deleteEntered != nil {
		close(s.deleteEntered)
		<-s.blockDelete
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.requireDeadline {
		_, s.deadlineSeen = ctx.Deadline()
		return context.DeadlineExceeded
	}
	delete(s.manifests, teamID)
	return nil
}

func (s *lockedDurableTeamStore) saveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saves
}

func (s *lockedDurableTeamStore) has(teamID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.manifests[teamID]
	return ok
}

func (s *lockedDurableTeamStore) sawDeadline() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deadlineSeen
}

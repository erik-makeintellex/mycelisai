package cognitive

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func snapshotTestRouter() *Router {
	return &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local":  {Type: "openai_compatible", ModelID: "l", Enabled: true, DataBoundary: DataBoundaryLocalOnly, RolesAllowed: []string{"all"}},
				"backup": {Type: "openai_compatible", ModelID: "b", Enabled: true, DataBoundary: DataBoundaryLocalOnly},
				"off":    {Type: "openai_compatible", ModelID: "o", Enabled: false, DataBoundary: DataBoundaryLocalOnly},
			},
			Profiles: map[string]string{
				"chat": "local", "coder": "off", "architect": "off", "admin": "missing", "research": "backup",
			},
			ProfileFallbacks: map[string][]string{"coder": {"backup"}},
		},
		Adapters: map[string]LLMProvider{"local": &MockAdapter{}, "backup": &MockAdapter{}, "off": &MockAdapter{}},
	}
}

func TestProfileProviderSnapshot_MatchesProfileRoutes(t *testing.T) {
	router := snapshotTestRouter()
	routes, _ := router.ProfileRoutes()
	for profile, route := range routes {
		providerID, cfg, ok := router.ProfileProviderSnapshot(profile)
		if providerID != route.ProviderID {
			t.Fatalf("%s: snapshot provider %q, ProfileRoutes %q", profile, providerID, route.ProviderID)
		}
		if ok && cfg.ModelID != route.ModelID {
			t.Fatalf("%s: snapshot model %q, ProfileRoutes %q", profile, cfg.ModelID, route.ModelID)
		}
	}
	cases := []struct {
		profile, provider string
		ok                bool
	}{
		{"chat", "local", true},      // bound and available
		{"coder", "backup", true},    // disabled primary, explicit fallback applied
		{"architect", "off", true},   // disabled, no fallback: still the bound provider
		{"admin", "missing", false},  // bound to an unconfigured provider
		{"creative", "", false},      // execution profile with no binding
		{"research", "backup", true}, // non-execution bound profile
		{"unknown", "", false},       // not listed by ProfileRoutes
		{"  ", "", false},            // blank
	}
	for _, tc := range cases {
		providerID, _, ok := router.ProfileProviderSnapshot(tc.profile)
		if providerID != tc.provider || ok != tc.ok {
			t.Fatalf("%q: got (%q, %v), want (%q, %v)", tc.profile, providerID, ok, tc.provider, tc.ok)
		}
	}
	var nilRouter *Router
	if id, _, ok := nilRouter.ProfileProviderSnapshot("chat"); id != "" || ok {
		t.Fatal("nil router must return nothing")
	}
	if adapter, ok := router.AdapterSnapshot("local"); !ok || adapter == nil {
		t.Fatal("AdapterSnapshot(local) missing")
	}
	if _, ok := router.AdapterSnapshot("missing"); ok {
		t.Fatal("AdapterSnapshot(missing) must be absent")
	}
	if _, ok := nilRouter.AdapterSnapshot("local"); ok {
		t.Fatal("nil router AdapterSnapshot must be absent")
	}
}

func TestProfileProviderSnapshot_ReturnsCopy(t *testing.T) {
	router := snapshotTestRouter()
	_, cfg, ok := router.ProfileProviderSnapshot("chat")
	if !ok {
		t.Fatal("chat snapshot missing")
	}
	cfg.ModelID = "tampered"
	cfg.Enabled = false
	cfg.RolesAllowed[0] = "tampered"
	live := router.Config.Providers["local"]
	if live.ModelID != "l" || !live.Enabled || live.RolesAllowed[0] != "all" {
		t.Fatalf("snapshot shares memory with live config: %#v", live)
	}
}

// TestProfileProviderSnapshot_ConcurrentMutationsRaceFree runs every router
// mutation path alongside the locked readers; -race is the assertion.
func TestProfileProviderSnapshot_ConcurrentMutationsRaceFree(t *testing.T) {
	router := snapshotTestRouter()
	router.ConfigPath = filepath.Join(t.TempDir(), "cognitive.yaml")
	var wg sync.WaitGroup
	run := func(fn func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				fn(i)
			}
		}()
	}
	run(func(i int) {
		target := []string{"local", "backup"}[i%2]
		_ = router.SetProfileOverrides(map[string]string{"chat": target, "coder": target}, ProfileOriginRuntime)
	})
	run(func(int) { router.ClearProfileOverride("chat") })
	run(func(i int) {
		cfg, _ := router.ProviderSnapshot("off")
		cfg.Enabled = i%2 == 0
		router.StoreProviderConfig("off", cfg)
	})
	run(func(i int) {
		id := fmt.Sprintf("extra%d", i%3)
		_ = router.AddProvider(id, ProviderConfig{Type: "openai_compatible", ModelID: "x", Enabled: true})
		_ = router.UpdateProvider(id, ProviderConfig{Type: "openai_compatible", ModelID: "y", Enabled: true})
		_ = router.RemoveProvider(id)
	})
	run(func(int) {
		for _, profile := range []string{"chat", "coder", "architect", "research"} {
			router.ProfileProviderSnapshot(profile)
		}
		router.AdapterSnapshot("extra0")
		router.ProfileRoutes()
		router.ConfigSnapshot()
	})
	wg.Wait()
}

// TestUpdateProvider_NeverResurrectsConcurrentRemove: whatever order an
// update and a remove interleave in, the provider ends up removed.
func TestUpdateProvider_NeverResurrectsConcurrentRemove(t *testing.T) {
	router := snapshotTestRouter()
	router.ConfigPath = filepath.Join(t.TempDir(), "cognitive.yaml")
	for i := 0; i < 100; i++ {
		if err := router.AddProvider("gone", ProviderConfig{Type: "openai_compatible", ModelID: "g", Enabled: true}); err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = router.UpdateProvider("gone", ProviderConfig{Type: "openai_compatible", ModelID: "h", Enabled: true})
		}()
		go func() { defer wg.Done(); _ = router.RemoveProvider("gone") }()
		wg.Wait()
		if _, ok := router.ProviderSnapshot("gone"); ok {
			// Update ran fully before Remove only if Remove then deleted it,
			// so a surviving provider means Update resurrected it.
			t.Fatalf("iteration %d: provider resurrected after concurrent remove", i)
		}
	}
	if err := router.UpdateProvider("gone", ProviderConfig{Type: "openai_compatible", ModelID: "z"}); err == nil {
		t.Fatal("updating a removed provider must fail")
	}
}

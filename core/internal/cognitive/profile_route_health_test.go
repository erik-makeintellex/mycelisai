package cognitive

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type routeProbeStub struct {
	healthy bool
	calls   atomic.Int32
}

func (s *routeProbeStub) Infer(context.Context, string, InferOptions) (*InferResponse, error) {
	return &InferResponse{Text: "ok"}, nil
}

func (s *routeProbeStub) Probe(context.Context) (bool, error) {
	s.calls.Add(1)
	return s.healthy, nil
}

func TestProfileRouteHealth_SeverityTable(t *testing.T) {
	ok := ProfileRoute{Available: true, Code: ExecutionAvailable, Source: ProfileSourceRoot, ExecutionProfile: true, Reachable: boolPtr(true)}
	with := func(mutate func(*ProfileRoute)) ProfileRoute {
		route := ok
		mutate(&route)
		return route
	}
	cases := []struct {
		name    string
		route   ProfileRoute
		overlay bool
		want    string
	}{
		{"all root and reachable", ok, false, ProfileRouteHealthOK},
		{"override available", with(func(r *ProfileRoute) { r.Source, r.OverrideOrigin = ProfileSourceOverride, ProfileOriginDB }), false, ProfileRouteHealthOK},
		{"hosted not probed is a note", with(func(r *ProfileRoute) { r.Reachable = nil }), false, ProfileRouteHealthOK},
		{"execution disabled", with(func(r *ProfileRoute) { r.Available, r.Code = false, ExecutionProviderDisabled }), false, ProfileRouteHealthFailed},
		{"execution unbound", with(func(r *ProfileRoute) { r.Available, r.Code = false, ExecutionNoProviders }), false, ProfileRouteHealthFailed},
		{"execution unreachable", with(func(r *ProfileRoute) { r.Reachable = boolPtr(false) }), false, ProfileRouteHealthFailed},
		{"other profile disabled", with(func(r *ProfileRoute) { r.ExecutionProfile, r.Available = false, false }), false, ProfileRouteHealthDegraded},
		{"other profile unreachable", with(func(r *ProfileRoute) { r.ExecutionProfile, r.Reachable = false, boolPtr(false) }), false, ProfileRouteHealthDegraded},
		{"explicit fallback", with(func(r *ProfileRoute) { r.Source = ProfileSourceFallback }), false, ProfileRouteHealthDegraded},
		{"overlay error", ok, true, ProfileRouteHealthDegraded},
		{"db row hidden under env", with(func(r *ProfileRoute) { r.OverrideOrigin, r.DBRowPresent = ProfileOriginEnv, true }), false, ProfileRouteHealthDegraded},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ProfileRouteHealth(map[string]ProfileRoute{"p": tc.route}, tc.overlay); got != tc.want {
				t.Fatalf("health = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestProfileRoutes_ProbesOncePerProviderAndSkipsHosted(t *testing.T) {
	local := &routeProbeStub{healthy: true}
	down := &routeProbeStub{healthy: false}
	hosted := &routeProbeStub{healthy: true}
	router := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"vllm":   {Type: "openai_compatible", Endpoint: "http://127.0.0.1:9/v1", ModelID: "m", Enabled: true},
				"down":   {Type: "ollama", Endpoint: "http://127.0.0.1:9/v1", ModelID: "d", Enabled: true},
				"hosted": {Type: "anthropic", Endpoint: "https://example.invalid", ModelID: "h", Enabled: true},
				"off":    {Type: "ollama", Endpoint: "http://127.0.0.1:9/v1", ModelID: "o", Enabled: false},
			},
			Profiles: map[string]string{"chat": "vllm", "admin": "vllm", "coder": "down", "creative": "hosted", "sentry": "off"},
		},
		Adapters: map[string]LLMProvider{"vllm": local, "down": down, "hosted": hosted, "off": &routeProbeStub{healthy: true}},
	}
	routes, _ := router.ProfileRoutes()
	ProbeProfileRoutes(context.Background(), routes, time.Second)

	if local.calls.Load() != 1 || down.calls.Load() != 1 || hosted.calls.Load() != 0 {
		t.Fatalf("probe calls vllm=%d down=%d hosted=%d, want 1/1/0", local.calls.Load(), down.calls.Load(), hosted.calls.Load())
	}
	if r := routes["chat"]; r.Reachable == nil || !*r.Reachable || !r.Available {
		t.Fatalf("chat route = %+v", r)
	}
	if r := routes["coder"]; r.Reachable == nil || *r.Reachable {
		t.Fatalf("coder route = %+v, want unreachable", r)
	}
	if r := routes["creative"]; r.Reachable != nil || !r.Available {
		t.Fatalf("hosted route = %+v, want not probed", r)
	}
	if r := routes["sentry"]; r.Available || r.Code != ExecutionProviderDisabled || r.Reachable != nil {
		t.Fatalf("disabled route = %+v, want unavailable and not probed", r)
	}
	if got := ProfileRouteHealth(routes, false); got != ProfileRouteHealthFailed {
		t.Fatalf("health = %s, want failed", got)
	}
}

// TestRouterProfileMap_ConcurrentSetClearResolve hammers every locked
// reader and writer of the profile map; run with -race.
func TestRouterProfileMap_ConcurrentSetClearResolve(t *testing.T) {
	router := overrideValidationRouter()
	router.ConfigPath = filepath.Join(t.TempDir(), "cognitive.yaml")
	router.Config.yamlProfileDefaults = map[string]string{"chat": "vllm"}

	var wg sync.WaitGroup
	run := func(fn func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				fn(i)
			}
		}()
	}
	run(func(i int) { _ = router.SetProfileOverride("coder", "vllm", ProfileOriginDB) })
	run(func(i int) { router.ClearProfileOverride("coder") })
	run(func(i int) { _ = router.SetProfileOverrides(map[string]string{"chat": "vllm"}, ProfileOriginRuntime) })
	run(func(i int) { router.ExecutionAvailability("coder", "") })
	run(func(i int) {
		_, _ = router.InferWithContract(context.Background(), InferRequest{Profile: "coder", Prompt: "x"})
	})
	run(func(i int) {
		routes, overlay := router.ProfileRoutes()
		_ = ProfileRouteHealth(routes, overlay)
	})
	run(func(i int) { _ = router.ConfigSnapshot() })
	run(func(i int) { _ = router.ProfileOverrideState("coder") })
	run(func(i int) { router.StoreProviderConfig("extra", ProviderConfig{Type: "ollama"}) })
	run(func(i int) {
		if i%20 == 0 {
			_ = router.SaveConfig()
		}
	})
	run(func(i int) { _ = router.KnownProfile("embed") })
	wg.Wait()

	if got := router.ExecutionAvailability("coder", ""); !got.Available || got.ProviderID != "vllm" {
		t.Fatalf("coder after hammer = %+v", got)
	}
}

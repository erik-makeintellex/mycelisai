package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// Regression for the security-QA DATA RACE: brains handlers wrote
// Config.Providers without Router.mu while inference and snapshots read it.
// Run under -race: concurrent toggles, policy/full updates, probes and list
// against inference, route snapshots and profile overrides must not race.
func TestRoutingMutationRace_ToggleVersusInferenceAndSnapshots(t *testing.T) {
	const rounds = 40
	f := newRoutingFixture(t)
	f.mock.MatchExpectationsInOrder(false)
	// Upper bound on audited handler calls below; 409s audit nothing.
	for i := 0; i < rounds*5; i++ {
		f.mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	}

	var wg sync.WaitGroup
	run := func(fn func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				fn(i)
			}
		}()
	}
	status := func(rr int, allowed ...int) {
		for _, a := range allowed {
			if rr == a {
				return
			}
		}
		t.Errorf("unexpected status %d (allowed %v)", rr, allowed)
	}

	// Writers through the gated handlers.
	run(func(i int) {
		body := fmt.Sprintf(`{"enabled":%v}`, i%2 == 1)
		status(doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/ollama/toggle", body).Code, http.StatusOK, http.StatusConflict)
	})
	run(func(i int) {
		body := fmt.Sprintf(`{"enabled":%v}`, i%2 == 0)
		status(doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/vllm/toggle", body).Code, http.StatusOK, http.StatusConflict)
	})
	run(func(i int) {
		body := fmt.Sprintf(`{"max_output_tokens":%d}`, 512+i)
		status(doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/ollama/policy", body).Code, http.StatusOK)
	})
	run(func(i int) {
		// Disabled so the list never probes the rebuilt (real) adapter.
		body := fmt.Sprintf(`{"type":"openai_compatible","endpoint":"http://127.0.0.1:9/v1","model_id":"c-%d","data_boundary":"leaves_org","enabled":false}`, i)
		status(doAuthenticatedRequest(t, f.mux, http.MethodPut, "/api/v1/brains/cloud", body).Code, http.StatusOK)
	})
	run(func(int) {
		status(doAuthenticatedRequest(t, f.mux, http.MethodPost, "/api/v1/brains/ollama/probe", "").Code, http.StatusOK)
	})
	// A direct runtime override writer (the S6c/activation apply path).
	run(func(i int) {
		provider := "ollama"
		if i%2 == 0 {
			provider = "vllm"
		}
		_ = f.s.Cognitive.SetProfileOverrides(map[string]string{"coder": provider}, cognitive.ProfileOriginRuntime)
	})

	// Readers: inference, route and config snapshots, the swarm accessor,
	// availability and the brains list.
	run(func(int) {
		_, _ = f.s.Cognitive.InferWithContract(context.Background(), cognitive.InferRequest{Profile: "chat", Prompt: "hi"})
		_, _ = f.s.Cognitive.InferWithContract(context.Background(), cognitive.InferRequest{Profile: "coder", Provider: "ollama", Prompt: "hi"})
	})
	run(func(int) {
		_, _ = f.s.Cognitive.ProfileRoutes()
		_ = f.s.Cognitive.ConfigSnapshot()
		_, _, _ = f.s.Cognitive.ProfileProviderSnapshot("coder")
		_ = f.s.Cognitive.ExecutionAvailability("coder", "")
	})
	run(func(int) {
		status(doRequest(t, f.mux, http.MethodGet, "/api/v1/brains", "").Code, http.StatusOK)
	})
	wg.Wait()

	// vllm is root for every execution profile, so it can never be disabled.
	if p, _ := f.s.Cognitive.ProviderSnapshot("vllm"); !p.Enabled {
		t.Fatal("bound root provider vllm was disabled")
	}
	if a := f.s.Cognitive.ExecutionAvailability("chat", ""); !a.Available {
		t.Fatalf("chat unavailable after concurrent mutations: %+v", a)
	}
}

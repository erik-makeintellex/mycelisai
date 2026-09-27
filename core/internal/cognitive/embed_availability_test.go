package cognitive

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type countingEmbedProvider struct {
	calls atomic.Int32
	fail  bool
	dims  int
}

func (p *countingEmbedProvider) Infer(context.Context, string, InferOptions) (*InferResponse, error) {
	return &InferResponse{Text: "ok"}, nil
}
func (p *countingEmbedProvider) Probe(context.Context) (bool, error) { return true, nil }
func (p *countingEmbedProvider) Embed(context.Context, string, string) ([]float64, error) {
	p.calls.Add(1)
	if p.fail {
		return nil, errors.New("embedding failed")
	}
	return make([]float64, p.dims), nil
}

func embedTestRouter(p *countingEmbedProvider) *Router {
	return &Router{
		Config:   &BrainConfig{Profiles: map[string]string{"chat": "stub"}, Providers: map[string]ProviderConfig{"stub": {Enabled: true}}},
		Adapters: map[string]LLMProvider{"stub": p},
	}
}

func TestEmbeddingAvailable_NegativeCacheMeansZeroEmbedCallsWithinTTL(t *testing.T) {
	provider := &countingEmbedProvider{fail: true}
	router := embedTestRouter(provider)
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	router.embedState.now = func() time.Time { return now }

	if router.EmbeddingAvailable(context.Background()) {
		t.Fatal("failing provider reported available")
	}
	if provider.calls.Load() != 1 {
		t.Fatalf("probe calls = %d, want 1", provider.calls.Load())
	}
	for i := 0; i < 5; i++ {
		if router.EmbeddingAvailable(context.Background()) {
			t.Fatal("negative cache ignored")
		}
		if _, err := router.Embed(context.Background(), "weekend special", ""); !errors.Is(err, ErrEmbeddingUnavailable) {
			t.Fatalf("Embed within TTL err = %v, want ErrEmbeddingUnavailable", err)
		}
	}
	if provider.calls.Load() != 1 {
		t.Fatalf("Embed calls within TTL = %d, want 1 (the probe)", provider.calls.Load())
	}
	status := router.EmbeddingStatus()
	if status.Available || status.Reason == "" || status.CheckedAt.IsZero() {
		t.Fatalf("status must report the failure honestly: %+v", status)
	}

	now = now.Add(EmbeddingNegativeCacheTTL + time.Second)
	provider.fail = false
	provider.dims = EmbeddingDimensions
	if !router.EmbeddingAvailable(context.Background()) {
		t.Fatal("recovered provider must be re-probed after TTL")
	}
	if provider.calls.Load() != 2 {
		t.Fatalf("calls after TTL = %d, want 2", provider.calls.Load())
	}
}

func TestEmbeddingAvailable_WrongDimensionIsUnavailable(t *testing.T) {
	router := embedTestRouter(&countingEmbedProvider{dims: 2})
	if router.EmbeddingAvailable(context.Background()) {
		t.Fatal("a 2-dim embedder cannot serve the 768-dim store")
	}
}

func TestEmbeddingAvailable_ChatOnlyAdapterAndNilRouter(t *testing.T) {
	var nilRouter *Router
	if nilRouter.EmbeddingAvailable(context.Background()) {
		t.Fatal("nil router reported available")
	}
	chatOnly := &Router{
		Config:   &BrainConfig{Profiles: map[string]string{"chat": "c"}},
		Adapters: map[string]LLMProvider{"c": chatOnlyProvider{}},
	}
	if chatOnly.EmbeddingAvailable(context.Background()) {
		t.Fatal("chat-only adapter reported embedding available")
	}
}

type chatOnlyProvider struct{}

func (chatOnlyProvider) Infer(context.Context, string, InferOptions) (*InferResponse, error) {
	return &InferResponse{Text: "ok"}, nil
}
func (chatOnlyProvider) Probe(context.Context) (bool, error) { return true, nil }

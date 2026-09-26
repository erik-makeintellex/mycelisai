package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// TestServiceStatusesRace_ConcurrentBrainToggle is F3: buildServiceStatuses
// reads Cognitive.Config.Providers and Cognitive.Adapters through the locked
// ProviderSnapshot/AdapterSnapshot/ConfigSnapshot accessors. A concurrent
// brains toggle (StoreProviderConfig) must never race with those reads. Run
// with -race.
func TestServiceStatusesRace_ConcurrentBrainToggle(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "ollama", Endpoint: "http://127.0.0.1:9/v1", ModelID: "m", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			prov, ok := s.Cognitive.ProviderSnapshot("ollama")
			if !ok {
				continue
			}
			prov.Enabled = i%2 == 0
			s.Cognitive.StoreProviderConfig("ollama", prov)
		}
	}()

	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/services/status", nil)
	adminReq = adminReq.WithContext(context.WithValue(adminReq.Context(), ctxKeyIdentity, localAdminIdentityForTest()))
	standardReq := httptest.NewRequest(http.MethodGet, "/api/v1/services/status", nil)

	for i := 0; i < 200; i++ {
		if i%2 == 0 {
			s.buildServiceStatuses(adminReq)
		} else {
			s.buildServiceStatuses(standardReq)
		}
	}
	close(stop)
	wg.Wait()
}

// TestOutputCatalogRace_ConcurrentBrainToggle is F3: listLocalOllamaModelIDs
// reads the ollama provider through ProviderSnapshot. A concurrent brains
// toggle must never race with that read. The endpoint is left blank so the
// function returns before any outbound call, keeping this a pure race test
// on the config read rather than a networking test. Run with -race.
func TestOutputCatalogRace_ConcurrentBrainToggle(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "ollama", ModelID: "m", Enabled: true},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			prov, ok := s.Cognitive.ProviderSnapshot("ollama")
			if !ok {
				continue
			}
			prov.Enabled = i%2 == 0
			s.Cognitive.StoreProviderConfig("ollama", prov)
		}
	}()

	for i := 0; i < 200; i++ {
		s.listLocalOllamaModelIDs()
	}
	close(stop)
	wg.Wait()
}

// TestTransportProvenanceRace_ConcurrentBrainToggle is F3: applyBrainProvenance
// reads the resolved provider through ProviderSnapshot. A concurrent brains
// toggle must never race with that read. Run with -race.
func TestTransportProvenanceRace_ConcurrentBrainToggle(t *testing.T) {
	cogOpt := withCognitive(t,
		map[string]cognitive.ProviderConfig{
			"ollama": {Type: "ollama", Endpoint: "http://127.0.0.1:9/v1", ModelID: "m", Enabled: true, Location: "local"},
		},
		map[string]cognitive.LLMProvider{
			"ollama": &stubAdapter{healthy: true},
		},
	)
	s := newTestServer(cogOpt)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			prov, ok := s.Cognitive.ProviderSnapshot("ollama")
			if !ok {
				continue
			}
			prov.Enabled = i%2 == 0
			s.Cognitive.StoreProviderConfig("ollama", prov)
		}
	}()

	for i := 0; i < 200; i++ {
		payload := &protocol.ChatResponsePayload{}
		applyBrainProvenance(s, payload, chatAgentResult{ProviderID: "ollama", ModelUsed: "m"})
	}
	close(stop)
	wg.Wait()
}

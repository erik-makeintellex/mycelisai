package swarm

import (
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func TestTeamNormalizeRuntimeProviderRoutingFallsBackExplicitDefaults(t *testing.T) {
	team := &Team{
		Manifest: &TeamManifest{
			ID: "admin-core", Name: "Soma", Provider: "local-ollama-dev",
			Members: []protocol.AgentManifest{
				{ID: "admin", Role: "admin"},
				{ID: "council-coder", Role: "coder", Provider: "local-ollama-dev"},
			},
		},
		brain: &cognitive.Router{
			Config: &cognitive.BrainConfig{Providers: map[string]cognitive.ProviderConfig{
				"ollama":           {Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local"},
				"local-ollama-dev": {Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local"},
			}},
			Adapters: map[string]cognitive.LLMProvider{"ollama": teamProviderStub{}},
		},
	}

	team.normalizeRuntimeProviderRouting()
	if team.Manifest.Provider != "ollama" || team.Manifest.Members[0].Provider != "ollama" || team.Manifest.Members[1].Provider != "ollama" {
		t.Fatalf("provider routing did not converge on ollama: %#v", team.Manifest)
	}
}

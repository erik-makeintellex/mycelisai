package cognitive

import "testing"

func TestExecutionAvailability_NoProvidersConfigured(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{},
			Profiles:  map[string]string{},
		},
		Adapters: map[string]LLMProvider{},
	}

	availability := r.ExecutionAvailability("chat", "")
	if availability.Available {
		t.Fatal("expected unavailable execution")
	}
	if availability.Code != ExecutionNoProviders {
		t.Fatalf("code = %q, want %q", availability.Code, ExecutionNoProviders)
	}
	if !availability.SetupRequired {
		t.Fatal("expected setup required")
	}
}

func TestExecutionAvailability_ProviderMissing(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{},
			Profiles: map[string]string{
				"chat": "missing-provider",
			},
		},
		Adapters: map[string]LLMProvider{},
	}

	availability := r.ExecutionAvailability("chat", "")
	if availability.Available {
		t.Fatal("expected unavailable execution")
	}
	if availability.Code != ExecutionProviderMissing {
		t.Fatalf("code = %q, want %q", availability.Code, ExecutionProviderMissing)
	}
}

func TestExecutionAvailability_SuccessfulBinding(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama": {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b"},
			},
			Profiles: map[string]string{
				"chat": "ollama",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama": &startupProbeStub{healthy: true},
		},
	}

	availability := r.ExecutionAvailability("chat", "")
	if !availability.Available {
		t.Fatalf("expected available execution, got %+v", availability)
	}
	if availability.Code != ExecutionAvailable {
		t.Fatalf("code = %q, want %q", availability.Code, ExecutionAvailable)
	}
}

func TestExecutionAvailability_ExplicitDisabledProviderFallsBack(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":           {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local"},
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama": &startupProbeStub{healthy: true},
		},
	}

	availability := r.ExecutionAvailability("chat", "local-ollama-dev")
	if !availability.Available {
		t.Fatalf("expected available execution via fallback, got %+v", availability)
	}
	if availability.Code != ExecutionAvailable {
		t.Fatalf("code = %q, want %q", availability.Code, ExecutionAvailable)
	}
	if availability.ProviderID != "ollama" {
		t.Fatalf("provider_id = %q, want ollama", availability.ProviderID)
	}
	if !availability.FallbackApplied {
		t.Fatal("expected fallback applied")
	}
	if !availability.SetupRequired {
		t.Fatal("expected setup required when fallback is masking bad default")
	}
}

func TestEnsureDefaultProfileBindings_UsesFallbackProvider(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":           {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local"},
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama": &startupProbeStub{healthy: true},
		},
	}

	rebound := r.EnsureDefaultProfileBindings()
	if rebound["chat"] != "ollama" {
		t.Fatalf("chat rebound to %q, want ollama", rebound["chat"])
	}
	if got := r.Config.Profiles["chat"]; got != "ollama" {
		t.Fatalf("chat profile = %q, want ollama", got)
	}
}

func TestResolveExecutionAvailability_ScopesAndFallback(t *testing.T) {
	t.Run("ready primary provider", func(t *testing.T) {
		r := &Router{
			Config: &BrainConfig{
				Providers: map[string]ProviderConfig{
					"ollama": {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b"},
				},
				Profiles: map[string]string{
					"architect": "ollama",
				},
			},
			Adapters: map[string]LLMProvider{
				"ollama": &startupProbeStub{healthy: true},
			},
		}

		status := r.ResolveExecutionAvailability(ScopeCouncil, "architect")
		if !status.Available {
			t.Fatalf("expected available status, got %+v", status)
		}
		if status.Scope != ScopeCouncil {
			t.Fatalf("scope = %q, want %q", status.Scope, ScopeCouncil)
		}
		if status.ResolvedProviderID != "ollama" {
			t.Fatalf("resolved_provider_id = %q, want ollama", status.ResolvedProviderID)
		}
		if status.FallbackApplied {
			t.Fatal("expected fallback_applied = false")
		}
	})

	t.Run("fallback provider rebinding", func(t *testing.T) {
		r := &Router{
			Config: &BrainConfig{
				Providers: map[string]ProviderConfig{
					"ollama":           {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local"},
					"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local"},
				},
				Profiles: map[string]string{
					"coder": "local-ollama-dev",
				},
			},
			Adapters: map[string]LLMProvider{
				"ollama": &startupProbeStub{healthy: true},
			},
		}

		status := r.ResolveExecutionAvailability(ScopeTeam, "coder")
		if !status.Available {
			t.Fatalf("expected available status via fallback, got %+v", status)
		}
		if !status.FallbackApplied {
			t.Fatal("expected fallback_applied = true")
		}
		if status.ResolvedProviderID != "ollama" {
			t.Fatalf("resolved_provider_id = %q, want ollama", status.ResolvedProviderID)
		}
	})

	t.Run("no provider setup guidance across scopes", func(t *testing.T) {
		r := &Router{
			Config: &BrainConfig{
				Providers: map[string]ProviderConfig{},
				Profiles:  map[string]string{},
			},
			Adapters: map[string]LLMProvider{},
		}

		for _, scope := range []ExecutionScope{ScopeTeam, ScopeAutomation, ScopeCouncil} {
			status := r.ResolveExecutionAvailability(scope, "")
			if status.Available {
				t.Fatalf("expected unavailable for scope %q", scope)
			}
			if status.Code != "cognitive_engine_unavailable" {
				t.Fatalf("code = %q, want cognitive_engine_unavailable", status.Code)
			}
			if status.SetupGuidance == "" || status.RecommendedAction == "" {
				t.Fatalf("expected actionable setup guidance for scope %q, got %+v", scope, status)
			}
		}
	})
}

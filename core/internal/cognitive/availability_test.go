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

// --- Positive: explicit same-boundary ProfileFallbacks list is honored. ---

func TestExecutionAvailability_ExplicitSameBoundaryFallbackApplied(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":           {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			ProfileFallbacks: map[string][]string{
				"chat": {"ollama"},
			},
		},
		Adapters: map[string]LLMProvider{
			"ollama": &startupProbeStub{healthy: true},
		},
	}

	availability := r.ExecutionAvailability("chat", "local-ollama-dev")
	if !availability.Available {
		t.Fatalf("expected available execution via explicit fallback, got %+v", availability)
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

// --- Negative: no explicit list configured -> fails closed, no substitution. ---

func TestExecutionAvailability_NoExplicitFallbackFailsClosed(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":           {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			// ProfileFallbacks intentionally omitted: default is no fallback.
		},
		Adapters: map[string]LLMProvider{
			"ollama": &startupProbeStub{healthy: true},
		},
	}

	availability := r.ExecutionAvailability("chat", "local-ollama-dev")
	if availability.Available {
		t.Fatalf("expected unavailable execution with no explicit fallback list, got %+v", availability)
	}
	if availability.FallbackApplied {
		t.Fatal("expected no silent fallback without an explicit ProfileFallbacks entry")
	}
	if availability.Code != ExecutionProviderDisabled {
		t.Fatalf("code = %q, want %q", availability.Code, ExecutionProviderDisabled)
	}
}

// --- Negative: explicit list exists but the entry crosses the data boundary. ---

func TestExecutionAvailability_CrossBoundaryFallbackDenied(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
				"openai":           {Type: "openai", Enabled: true, ModelID: "gpt-4.1-mini", Location: "remote", DataBoundary: "leaves_org"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			ProfileFallbacks: map[string][]string{
				// Misconfigured: a local_only profile must never fall over to
				// a leaves_org provider, even when explicitly listed.
				"chat": {"openai"},
			},
		},
		Adapters: map[string]LLMProvider{
			"openai": &startupProbeStub{healthy: true},
		},
	}

	availability := r.ExecutionAvailability("chat", "local-ollama-dev")
	if availability.Available {
		t.Fatalf("expected unavailable execution: cross-boundary fallback must be denied, got %+v", availability)
	}
	if availability.FallbackApplied {
		t.Fatal("expected no fallback applied across data boundaries")
	}
	if availability.ProviderID == "openai" {
		t.Fatal("must never route a local_only profile to a leaves_org provider")
	}
}

func TestEnsureDefaultProfileBindings_UsesExplicitFallbackOnly(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":           {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			ProfileFallbacks: map[string][]string{
				"chat": {"ollama"},
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

func TestEnsureDefaultProfileBindings_LeavesUnboundWithoutExplicitList(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"ollama":           {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			// No ProfileFallbacks: must not silently pick "ollama".
		},
		Adapters: map[string]LLMProvider{
			"ollama": &startupProbeStub{healthy: true},
		},
	}

	rebound := r.EnsureDefaultProfileBindings()
	if _, ok := rebound["chat"]; ok {
		t.Fatalf("expected chat to remain unbound without an explicit fallback list, got rebound=%v", rebound)
	}
	if got := r.Config.Profiles["chat"]; got != "local-ollama-dev" {
		t.Fatalf("chat profile changed to %q without operator configuration", got)
	}
}

// --- Security QA fixes: boundary must be checked, and empty/unknown ---
// --- DataBoundary must normalize to local_only, not bypass the check. ---

// TestEnsureDefaultProfileBindings_DeniesCrossBoundaryFallback proves the
// HIGH finding at (former) availability.go:91: startup rebinding must
// compare against the boundary of the profile's primary/configured
// provider, not skip the check entirely.
func TestEnsureDefaultProfileBindings_DeniesCrossBoundaryFallback(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: "local_only"},
				"openai":           {Type: "openai", Enabled: true, ModelID: "gpt-4.1-mini", Location: "remote", DataBoundary: "leaves_org"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			ProfileFallbacks: map[string][]string{
				// Misconfigured: local_only primary, leaves_org candidate.
				"chat": {"openai"},
			},
		},
		Adapters: map[string]LLMProvider{
			"openai": &startupProbeStub{healthy: true},
		},
	}

	rebound := r.EnsureDefaultProfileBindings()
	if got, ok := rebound["chat"]; ok {
		t.Fatalf("chat must not be rebound across data boundaries, got %q", got)
	}
	if got := r.Config.Profiles["chat"]; got != "local-ollama-dev" {
		t.Fatalf("chat profile changed to %q across data boundaries", got)
	}
}

// TestExecutionAvailability_EmptyPrimaryBoundaryDeniesLeavesOrgFallback
// proves the HIGH finding at (former) availability.go:~223: an empty/unknown
// DataBoundary on the primary provider (as shipped in
// core/config/cognitive.yaml for local-ollama-dev/local-sovereign) must
// normalize to local_only and still refuse a leaves_org fallback candidate —
// not be treated as "no boundary to enforce".
func TestExecutionAvailability_EmptyPrimaryBoundaryDeniesLeavesOrgFallback(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				// DataBoundary intentionally empty, matching the shipped
				// local-ollama-dev/local-sovereign entries.
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: ""},
				"openai":           {Type: "openai", Enabled: true, ModelID: "gpt-4.1-mini", Location: "remote", DataBoundary: "leaves_org"},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			ProfileFallbacks: map[string][]string{
				"chat": {"openai"},
			},
		},
		Adapters: map[string]LLMProvider{
			"openai": &startupProbeStub{healthy: true},
		},
	}

	availability := r.ExecutionAvailability("chat", "")
	if availability.Available {
		t.Fatalf("expected unavailable: empty-boundary primary must normalize to local_only and refuse leaves_org fallback, got %+v", availability)
	}
	if availability.FallbackApplied || availability.ProviderID == "openai" {
		t.Fatalf("must never fall back from an empty-boundary (local_only) primary to a leaves_org provider, got %+v", availability)
	}
}

// TestExecutionAvailability_EmptyBothBoundariesStillMatch proves empty vs.
// empty (both normalize to local_only) is a legitimate match, not itself
// denied — the fix must not make every empty-boundary provider unusable.
func TestExecutionAvailability_EmptyBothBoundariesStillMatch(t *testing.T) {
	r := &Router{
		Config: &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local-ollama-dev": {Type: "openai_compatible", Enabled: false, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: ""},
				"local-sovereign":  {Type: "openai_compatible", Enabled: true, ModelID: "qwen2.5-coder:7b", Location: "local", DataBoundary: ""},
			},
			Profiles: map[string]string{
				"chat": "local-ollama-dev",
			},
			ProfileFallbacks: map[string][]string{
				"chat": {"local-sovereign"},
			},
		},
		Adapters: map[string]LLMProvider{
			"local-sovereign": &startupProbeStub{healthy: true},
		},
	}

	availability := r.ExecutionAvailability("chat", "")
	if !availability.Available || availability.ProviderID != "local-sovereign" {
		t.Fatalf("expected fallback to local-sovereign (both normalize to local_only), got %+v", availability)
	}
}

// TestValidateProfileFallbackBoundaries proves the MED finding: a
// profile_fallbacks entry whose boundary differs from the primary
// provider's (after the empty-to-local_only rule) is a config error, not a
// silent skip.
func TestValidateProfileFallbackBoundaries(t *testing.T) {
	t.Run("cross boundary is rejected", func(t *testing.T) {
		config := &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local-ollama-dev": {DataBoundary: "local_only"},
				"openai":           {DataBoundary: "leaves_org"},
			},
			Profiles:         map[string]string{"chat": "local-ollama-dev"},
			ProfileFallbacks: map[string][]string{"chat": {"openai"}},
		}
		if err := validateProfileFallbackBoundaries(config); err == nil {
			t.Fatal("expected a config validation error for a cross-boundary fallback entry")
		}
	})

	t.Run("empty primary boundary still rejects a leaves_org candidate", func(t *testing.T) {
		config := &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local-ollama-dev": {DataBoundary: ""},
				"openai":           {DataBoundary: "leaves_org"},
			},
			Profiles:         map[string]string{"chat": "local-ollama-dev"},
			ProfileFallbacks: map[string][]string{"chat": {"openai"}},
		}
		if err := validateProfileFallbackBoundaries(config); err == nil {
			t.Fatal("expected a config validation error: empty primary boundary must normalize to local_only")
		}
	})

	t.Run("same boundary is accepted", func(t *testing.T) {
		config := &BrainConfig{
			Providers: map[string]ProviderConfig{
				"local-ollama-dev": {DataBoundary: "local_only"},
				"ollama":           {DataBoundary: "local_only"},
			},
			Profiles:         map[string]string{"chat": "local-ollama-dev"},
			ProfileFallbacks: map[string][]string{"chat": {"ollama"}},
		}
		if err := validateProfileFallbackBoundaries(config); err != nil {
			t.Fatalf("expected no error for a same-boundary fallback entry, got %v", err)
		}
	})

	t.Run("no ProfileFallbacks configured is accepted", func(t *testing.T) {
		config := &BrainConfig{
			Providers: map[string]ProviderConfig{"ollama": {DataBoundary: "local_only"}},
			Profiles:  map[string]string{"chat": "ollama"},
		}
		if err := validateProfileFallbackBoundaries(config); err != nil {
			t.Fatalf("expected no error with no ProfileFallbacks, got %v", err)
		}
	})
}

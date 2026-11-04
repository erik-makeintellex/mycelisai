package cognitive

import "testing"

func TestNormalizeMediaConfig_LegacyLocalEndpointDefaultsToLocalProvider(t *testing.T) {
	cfg := NormalizeMediaConfig(MediaConfig{
		Endpoint: "http://127.0.0.1:8001/v1",
		ModelID:  "stable-diffusion-xl",
	})

	if cfg.Provider.ProviderID != DefaultMediaProviderID {
		t.Fatalf("provider_id = %q, want %q", cfg.Provider.ProviderID, DefaultMediaProviderID)
	}
	if cfg.Provider.Type != DefaultMediaProviderType {
		t.Fatalf("type = %q, want %q", cfg.Provider.Type, DefaultMediaProviderType)
	}
	if cfg.Provider.Location != DefaultMediaLocalLocation {
		t.Fatalf("location = %q, want %q", cfg.Provider.Location, DefaultMediaLocalLocation)
	}
	if cfg.Provider.DataBoundary != DefaultMediaLocalBoundary {
		t.Fatalf("data_boundary = %q, want %q", cfg.Provider.DataBoundary, DefaultMediaLocalBoundary)
	}
	if cfg.Provider.UsagePolicy != DefaultMediaLocalUsagePolicy {
		t.Fatalf("usage_policy = %q, want %q", cfg.Provider.UsagePolicy, DefaultMediaLocalUsagePolicy)
	}
	if !cfg.Provider.IsEnabled() {
		t.Fatal("expected provider to be enabled when endpoint/model are configured")
	}
}

func TestNormalizeMediaConfig_ExplicitHostedProviderPreservesHostedMetadata(t *testing.T) {
	cfg := NormalizeMediaConfig(MediaConfig{
		Provider: MediaProviderConfig{
			ProviderID:   "replicate",
			Type:         "hosted_api",
			Endpoint:     "https://api.replicate.com/v1",
			ModelID:      "sdxl",
			Location:     "remote",
			DataBoundary: "leaves_org",
			UsagePolicy:  "require_approval",
			Enabled:      boolPtr(true),
		},
	})

	if cfg.Provider.ProviderID != "replicate" {
		t.Fatalf("provider_id = %q, want replicate", cfg.Provider.ProviderID)
	}
	if cfg.Provider.Type != "hosted_api" {
		t.Fatalf("type = %q, want hosted_api", cfg.Provider.Type)
	}
	if cfg.Provider.Location != "remote" {
		t.Fatalf("location = %q, want remote", cfg.Provider.Location)
	}
	if cfg.Provider.DataBoundary != "leaves_org" {
		t.Fatalf("data_boundary = %q, want leaves_org", cfg.Provider.DataBoundary)
	}
	if cfg.Provider.UsagePolicy != "require_approval" {
		t.Fatalf("usage_policy = %q, want require_approval", cfg.Provider.UsagePolicy)
	}
	if !cfg.Provider.IsEnabled() {
		t.Fatal("expected hosted provider to remain enabled")
	}
}

func TestNormalizeMediaConfig_ExplicitFalseRemainsDisabled(t *testing.T) {
	cfg := NormalizeMediaConfig(MediaConfig{
		Provider: MediaProviderConfig{
			Endpoint: "http://127.0.0.1:8001/v1",
			ModelID:  "stable-diffusion-xl",
			Enabled:  boolPtr(false),
		},
	})

	if cfg.Provider.IsEnabled() {
		t.Fatal("expected explicit enabled=false to remain disabled")
	}
}

func TestNormalizeMediaConfig_PartialConfigIsNotGenerationReady(t *testing.T) {
	cfg := NormalizeMediaConfig(MediaConfig{
		Endpoint: "http://127.0.0.1:8001/v1",
	})

	if cfg.Provider.IsEnabled() {
		t.Fatal("expected provider to stay disabled until endpoint and model are both configured")
	}
	if cfg.IsConfigured() {
		t.Fatal("expected endpoint-only media config to be incomplete")
	}
	if cfg.IsReadyForGeneration() {
		t.Fatal("expected endpoint-only media config to be not ready for generation")
	}
}

func TestMediaProviderConfig_GenerationEndpointURLAndApprovalPolicy(t *testing.T) {
	cases := []struct {
		name         string
		provider     MediaProviderConfig
		wantEndpoint string
		wantApproval bool
	}{
		{
			name: "default local openai_compatible appends images/generations",
			provider: MediaProviderConfig{
				Type:     MediaProviderTypeOpenAICompat,
				Endpoint: "http://127.0.0.1:8001/v1/",
				ModelID:  "sdxl",
			},
			wantEndpoint: "http://127.0.0.1:8001/v1/images/generations",
			wantApproval: false,
		},
		{
			name: "openai_hosted preserves existing images/generations path and requires approval",
			provider: MediaProviderConfig{
				Type:     MediaProviderTypeOpenAIHosted,
				Endpoint: "https://api.openai.com/v1/images/generations/",
				ModelID:  "gpt-image-1",
			},
			wantEndpoint: "https://api.openai.com/v1/images/generations",
			wantApproval: true,
		},
		{
			name: "stability_compatible normalizes base endpoint and requires approval on leaves_org",
			provider: MediaProviderConfig{
				Type:         MediaProviderTypeStabilityCompat,
				Endpoint:     "https://api.stability.ai/v1",
				ModelID:      "sd3",
				DataBoundary: DefaultMediaHostedBoundary,
				UsagePolicy:  DefaultMediaLocalUsagePolicy,
			},
			wantEndpoint: "https://api.stability.ai/v1/images/generations",
			wantApproval: true,
		},
		{
			name: "replicate_compatible preserves v1/predictions endpoint and requires approval on policy",
			provider: MediaProviderConfig{
				Type:         MediaProviderTypeReplicateCompat,
				Endpoint:     "https://api.replicate.com/v1/predictions/",
				ModelID:      "flux-schnell",
				Location:     DefaultMediaLocalLocation,
				DataBoundary: DefaultMediaLocalBoundary,
				UsagePolicy:  DefaultMediaHostUsagePolicy,
			},
			wantEndpoint: "https://api.replicate.com/v1/predictions",
			wantApproval: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NormalizeMediaConfig(MediaConfig{Provider: tc.provider}).Provider
			if got := cfg.GenerationEndpointURL(); got != tc.wantEndpoint {
				t.Fatalf("GenerationEndpointURL() = %q, want %q", got, tc.wantEndpoint)
			}
			if got := cfg.RequiresApprovalByPolicy(); got != tc.wantApproval {
				t.Fatalf("RequiresApprovalByPolicy() = %v, want %v", got, tc.wantApproval)
			}
		})
	}
}

func boolPtr(value bool) *bool {
	return &value
}

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

func boolPtr(value bool) *bool {
	return &value
}

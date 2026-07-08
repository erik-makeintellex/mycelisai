package searchcap

import "testing"

func TestConfigFromEnvAcceptsSelfHostedLocalAPI(t *testing.T) {
	t.Setenv("MYCELIS_SEARCH_PROVIDER", "self_hosted")
	t.Setenv("MYCELIS_SEARCH_LOCAL_API_ENDPOINT", "http://search.local/api/search")
	t.Setenv("MYCELIS_SEARCH_MAX_RESULTS", "3")
	t.Setenv("MYCELIS_SEARCH_ONLINE_ALLOWED", "true")
	t.Setenv("MYCELIS_SEARCH_APPROVAL_MODE", "notify")
	t.Setenv("MYCELIS_SEARCH_DISCLOSURE_MODE", "notice_and_interpretation")

	cfg := ConfigFromEnv()

	if cfg.Provider != ProviderLocalAPI {
		t.Fatalf("Provider = %q, want %q", cfg.Provider, ProviderLocalAPI)
	}
	if cfg.LocalAPIEndpoint != "http://search.local/api/search" {
		t.Fatalf("LocalAPIEndpoint = %q", cfg.LocalAPIEndpoint)
	}
	if cfg.MaxResults != 3 {
		t.Fatalf("MaxResults = %d, want 3", cfg.MaxResults)
	}
	if !cfg.OnlineAllowed || !cfg.OnlineAllowedSet || cfg.ApprovalMode != "notify" || cfg.DisclosureMode != "notice_and_interpretation" {
		t.Fatalf("search governance config = %+v", cfg)
	}
}

func TestConfigFromEnvDefaultsToBuiltinWeb(t *testing.T) {
	t.Setenv("MYCELIS_SEARCH_PROVIDER", "")

	cfg := ConfigFromEnv()

	if cfg.Provider != ProviderBuiltinWeb {
		t.Fatalf("Provider = %q, want %q", cfg.Provider, ProviderBuiltinWeb)
	}
	if !cfg.OnlineAllowed || !cfg.OnlineAllowedSet || cfg.ApprovalMode != "notify" || cfg.DisclosureMode != "notice_and_interpretation" {
		t.Fatalf("search governance config = %+v", cfg)
	}
}

func TestConfigFromEnvAcceptsBuiltinWebAliases(t *testing.T) {
	t.Setenv("MYCELIS_SEARCH_PROVIDER", "local_web")

	cfg := ConfigFromEnv()

	if cfg.Provider != ProviderBuiltinWeb {
		t.Fatalf("Provider = %q, want %q", cfg.Provider, ProviderBuiltinWeb)
	}
}

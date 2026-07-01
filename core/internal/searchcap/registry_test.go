package searchcap

import "testing"

func TestServiceSearchSourceRegistryListsConfiguredAndAddedSources(t *testing.T) {
	svc := NewService(Config{Provider: ProviderSearXNG, SearXNGEndpoint: "http://searxng.local", MaxResults: 5}, nil, nil)

	added, err := svc.AddSource(SourceInput{
		Name:             "Research Team Search",
		Provider:         "local_api",
		Endpoint:         "http://search.local/api/search",
		Scope:            "Group",
		ScopeRef:         "research-team",
		Boundary:         "operator-owned research index",
		AuthScheme:       "api_token",
		SecretRef:        "MYCELIS_RESEARCH_SEARCH_TOKEN",
		Mode:             "live",
		SensitivityClass: "configured",
		TrustClass:       "bounded_external",
		Status:           "available",
	})
	if err != nil {
		t.Fatalf("AddSource: %v", err)
	}
	if added.SecretRef != "MYCELIS_RESEARCH_SEARCH_TOKEN" || added.ScopeKind != "group" || added.ScopeRef != "research-team" {
		t.Fatalf("added source = %+v", added)
	}
	if added.Endpoint != "http://search.local/api/search" || added.BaseURL != added.Endpoint {
		t.Fatalf("endpoint/base_url = %q/%q", added.Endpoint, added.BaseURL)
	}

	sources := svc.ListSources()
	if len(sources) != 2 {
		t.Fatalf("sources = %+v, want configured source plus added registry source", sources)
	}
	if sources[0].ID != "searxng" || sources[0].Endpoint != "http://searxng.local" {
		t.Fatalf("configured source = %+v", sources[0])
	}
	if sources[1].Name != "Research Team Search" || sources[1].AuthScheme != "api_token" {
		t.Fatalf("registry source = %+v", sources[1])
	}
}

func TestServiceSearchSourceRegistryRejectsRawCredentialShape(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)

	if _, err := svc.AddSource(SourceInput{
		Name:       "Hosted Search",
		Provider:   "brave",
		AuthScheme: "api_token",
		SecretRef:  "sk-this-is-a-raw-token",
	}); err == nil {
		t.Fatalf("expected raw-looking secret_ref to be rejected")
	}
	if _, err := svc.AddSource(SourceInput{
		Name:       "Embedded Credentials",
		Provider:   "local_api",
		Endpoint:   "https://user:pass@example.test/search",
		AuthScheme: "none",
	}); err == nil {
		t.Fatalf("expected endpoint credentials to be rejected")
	}
	if _, err := svc.AddSource(SourceInput{
		Name:       "Missing Endpoint",
		Provider:   "local_api",
		AuthScheme: "none",
	}); err == nil {
		t.Fatalf("expected API-backed source without endpoint to be rejected")
	}
}

func TestServiceSearchSourceRegistryAcceptsSecretRefAuthAlias(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)

	source, err := svc.AddSource(SourceInput{
		Name:       "Docs Search",
		Provider:   "local_api",
		Endpoint:   "https://docs.example.test/search",
		AuthScheme: "secret_ref",
		SecretRef:  "DOCS_SEARCH_TOKEN",
	})
	if err != nil {
		t.Fatalf("AddSource: %v", err)
	}
	if source.AuthScheme != "api_token" || source.SecretRef != "DOCS_SEARCH_TOKEN" {
		t.Fatalf("source auth = %+v, want api_token with secret ref", source)
	}
}

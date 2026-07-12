package searchcap

import "testing"

func TestServiceSearchSourceRegistryListsConfiguredAndAddedSources(t *testing.T) {
	svc := NewService(Config{Provider: ProviderSearXNG, SearXNGEndpoint: "http://searxng.local", MaxResults: 5}, nil, nil)

	added := seedManagedSource(t, svc, SourceInput{
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

func TestServiceSearchSourceRegistryAcceptsMountedFolderPath(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)

	source := seedManagedSource(t, svc, SourceInput{
		Name:       "Local client docs",
		Provider:   "mounted_folder",
		SourceType: "mounted_folder",
		Endpoint:   "workspace/client-docs",
		Boundary:   "operator-approved client documents",
		AuthScheme: "none",
	})
	if source.Endpoint != "workspace/client-docs" || source.SourceType != "mounted_folder" {
		t.Fatalf("mounted source = %+v", source)
	}
}

func TestServiceSearchSourceRegistryAcceptsSecretRefAuthAlias(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)

	source := seedManagedSource(t, svc, SourceInput{
		Name:       "Docs Search",
		Provider:   "local_api",
		Endpoint:   "https://docs.example.test/search",
		AuthScheme: "secret_ref",
		SecretRef:  "DOCS_SEARCH_TOKEN",
	})
	if source.AuthScheme != "api_token" || source.SecretRef != "DOCS_SEARCH_TOKEN" {
		t.Fatalf("source auth = %+v, want api_token with secret ref", source)
	}
}

func TestServiceSearchSourceRegistryRequiresPostgresStoreForMutations(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)
	if _, err := svc.AddSource(SourceInput{
		Name:       "Docs Search",
		Provider:   "local_api",
		Endpoint:   "https://docs.example.test/search",
		Boundary:   "docs index",
		AuthScheme: "none",
	}); err != errSourceStoreUnavailable {
		t.Fatalf("AddSource error = %v, want errSourceStoreUnavailable", err)
	}

	source := seedManagedSource(t, svc, SourceInput{
		Name:       "Docs Search",
		Provider:   "local_api",
		Endpoint:   "https://docs.example.test/search",
		Boundary:   "docs index",
		AuthScheme: "none",
	})
	updated, err := svc.UpdateSourceWithContext(t.Context(), source.ID, SourceInput{
		Name:       "Docs Search v2",
		Provider:   "local_api",
		Endpoint:   "https://docs.example.test/v2",
		Boundary:   "approved docs index",
		AuthScheme: "none",
		Status:     "available",
		Mode:       "live",
	})
	if err != errSourceStoreUnavailable {
		t.Fatalf("UpdateSourceWithContext error = %v, want errSourceStoreUnavailable", err)
	}
	if updated.ID != "" {
		t.Fatalf("updated = %+v", updated)
	}

	if err := svc.DeleteSourceWithContext(t.Context(), source.ID); err != errSourceStoreUnavailable {
		t.Fatalf("DeleteSourceWithContext error = %v, want errSourceStoreUnavailable", err)
	}
	sources := svc.ListSources()
	if len(sources) != 1 {
		t.Fatalf("sources after failed delete = %+v", sources)
	}
}

package searchcap

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceSelectedSourceRoutesToRegisteredLocalAPIEndpoint(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled, MaxResults: 5}, nil, nil)
	source := seedManagedSource(t, svc, SourceInput{
		Name:       "Team research",
		Provider:   "local_api",
		Endpoint:   "http://selected-search.local/query",
		Scope:      "group",
		ScopeRef:   "research-team",
		Boundary:   "approved research index",
		AuthScheme: "none",
		Status:     "available",
		Mode:       "live",
	})
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "selected-search.local" || r.URL.Path != "/query" {
			t.Fatalf("selected endpoint = %s", r.URL.String())
		}
		if r.URL.Query().Get("q") != "release research" {
			t.Fatalf("query params = %q", r.URL.RawQuery)
		}
		if r.URL.Query().Get("source_scope") != "" {
			t.Fatalf("selected source unexpectedly inherited source_scope = %q", r.URL.Query().Get("source_scope"))
		}
		if r.Header.Get("X-Mycelis-Team-ID") != "research-team" {
			t.Fatalf("team header = %q", r.Header.Get("X-Mycelis-Team-ID"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"results":[{"title":"Selected","url":"https://example.test/selected","snippet":"Selected result"}]}`)),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "release research", SourceID: source.ID, TeamID: "research-team"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Metadata["selected_source_id"] != source.ID || resp.Metadata["selected_source_boundary"] != "approved research index" {
		t.Fatalf("metadata = %+v", resp.Metadata)
	}
}

func TestServiceSelectedSourceBlocksOutOfScopeAndMissingSecret(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)
	groupSource := seedManagedSource(t, svc, SourceInput{
		Name:       "Group docs",
		Provider:   "local_api",
		Endpoint:   "https://search.example.test/api",
		Scope:      "group",
		ScopeRef:   "research-team",
		Boundary:   "research group only",
		AuthScheme: "none",
		Status:     "available",
	})
	resp, err := svc.Search(context.Background(), Request{Query: "docs", SourceID: groupSource.ID, TeamID: "marketing-team"})
	if err != nil {
		t.Fatalf("Search scope: %v", err)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "search_source_out_of_scope" {
		t.Fatalf("scope blocker = %+v", resp.Blocker)
	}

	authSource := seedManagedSource(t, svc, SourceInput{
		Name:       "Private API",
		Provider:   "local_api",
		Endpoint:   "https://private.example.test/api",
		Boundary:   "private search",
		AuthScheme: "api_token",
		SecretRef:  "PRIVATE_SEARCH_TOKEN",
		Status:     "available",
	})
	resp, err = svc.Search(context.Background(), Request{Query: "docs", SourceID: authSource.ID})
	if err != nil {
		t.Fatalf("Search auth: %v", err)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "search_source_secret_missing" {
		t.Fatalf("auth blocker = %+v", resp.Blocker)
	}
}

func TestServiceSelectedSourceAppliesBearerSecretRefForLocalAPI(t *testing.T) {
	t.Setenv("PRIVATE_SEARCH_TOKEN", "test-private-token")
	svc := NewService(Config{Provider: ProviderDisabled, MaxResults: 5}, nil, nil)
	source := seedManagedSource(t, svc, SourceInput{
		Name:       "Private API",
		Provider:   "local_api",
		Endpoint:   "http://private.example.test/api",
		Boundary:   "private search",
		AuthScheme: "api_token",
		SecretRef:  "env:PRIVATE_SEARCH_TOKEN",
		Status:     "available",
	})
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-private-token" {
			t.Fatalf("Authorization = %q", got)
		}
		if strings.Contains(r.URL.String(), "test-private-token") {
			t.Fatalf("secret leaked into url: %s", r.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"results":[{"title":"Private","url":"https://example.test/private","snippet":"Private result"}]}`)),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "private docs", SourceID: source.ID})
	if err != nil {
		t.Fatalf("Search auth: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if body := strings.Join([]string{resp.Results[0].Title, resp.Results[0].URL, resp.Results[0].Snippet}, " "); strings.Contains(body, "test-private-token") {
		t.Fatalf("secret leaked into response: %s", body)
	}
}

func TestServiceSelectedSourceBlocksBearerSecretRefForSearXNG(t *testing.T) {
	t.Setenv("PRIVATE_SEARCH_TOKEN", "test-private-token")
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)
	source := seedManagedSource(t, svc, SourceInput{
		Name:       "Private SearXNG",
		Provider:   "searxng",
		Endpoint:   "http://searxng.example.test",
		Boundary:   "private web search",
		AuthScheme: "bearer_token",
		SecretRef:  "PRIVATE_SEARCH_TOKEN",
		Status:     "available",
	})
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("SearXNG should not be called when source auth is unsupported")
		return nil, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "private docs", SourceID: source.ID})
	if err != nil {
		t.Fatalf("Search auth: %v", err)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "search_source_auth_adapter_required" {
		t.Fatalf("auth blocker = %+v", resp.Blocker)
	}
}

func TestServiceSelectedMountedFolderSearchesLiveFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "client-notes.md"), []byte("Hermes agent design requires offline message handoff proof."), 0o600); err != nil {
		t.Fatalf("write mounted file: %v", err)
	}
	svc := NewService(Config{Provider: ProviderDisabled, MaxResults: 5}, nil, nil)
	source := seedManagedSource(t, svc, SourceInput{
		Name:             "Client docs mount",
		Provider:         "mounted_folder",
		SourceType:       "mounted_folder",
		Endpoint:         root,
		Boundary:         "operator-approved client docs folder",
		AuthScheme:       "none",
		Mode:             "live",
		SensitivityClass: "restricted",
		TrustClass:       "trusted_internal",
		Status:           "available",
	})

	resp, err := svc.Search(context.Background(), Request{Query: "Hermes offline proof", SourceID: source.ID})
	if err != nil {
		t.Fatalf("Search mount: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Results[0].SourceKind != "mounted_folder" || resp.Results[0].Title != "client-notes.md" {
		t.Fatalf("mounted result = %+v", resp.Results[0])
	}
	if !strings.Contains(resp.Results[0].Snippet, "offline message handoff") {
		t.Fatalf("snippet = %q", resp.Results[0].Snippet)
	}
	if resp.Metadata["selected_source_boundary"] != "operator-approved client docs folder" {
		t.Fatalf("metadata = %+v", resp.Metadata)
	}
}

func TestServiceSelectedCodeContextSearchesRepositoryFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "engine.go"), []byte("package engine\n\nfunc BuildOutcomeMap() {}\n"), 0o600); err != nil {
		t.Fatalf("write code file: %v", err)
	}
	svc := NewService(Config{Provider: ProviderDisabled, MaxResults: 5}, nil, nil)
	source := seedManagedSource(t, svc, SourceInput{
		Name:       "Runtime repository map",
		Provider:   ProviderCodeContext,
		SourceType: ProviderCodeContext,
		Endpoint:   root,
		Boundary:   "operator-approved repository",
		Status:     "available",
	})

	resp, err := svc.Search(context.Background(), Request{Query: "BuildOutcomeMap", SourceID: source.ID})
	if err != nil {
		t.Fatalf("Search code context: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Provider != ProviderCodeContext {
		t.Fatalf("provider = %q", resp.Provider)
	}
	if resp.Results[0].SourceKind != ProviderCodeContext || resp.Results[0].Title != "engine.go" {
		t.Fatalf("code context result = %+v", resp.Results[0])
	}
	if resp.Results[0].ProviderMetadata["source_type"] != ProviderCodeContext {
		t.Fatalf("metadata = %+v", resp.Results[0].ProviderMetadata)
	}
	if resp.Metadata["interpretation"] != "code_context_results_are_operator_configured_repository_or_code_folder_refs" {
		t.Fatalf("interpretation = %+v", resp.Metadata)
	}
}

func TestServiceLocalSourcesIncludesAvailableMountedFolders(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "research.txt"), []byte("Mounted data search should be live to Soma."), 0o600); err != nil {
		t.Fatalf("write mounted file: %v", err)
	}
	svc := NewService(Config{Provider: ProviderLocalSources, MaxResults: 5}, nil, nil)
	seedManagedSource(t, svc, SourceInput{
		Name:       "Shared research mount",
		Provider:   "mounted_folder",
		SourceType: "mounted_folder",
		Endpoint:   root,
		Boundary:   "approved shared research folder",
		Status:     "available",
	})

	resp, err := svc.Search(context.Background(), Request{Query: "mounted data search"})
	if err != nil {
		t.Fatalf("Search default local sources: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Results[0].SourceKind != "mounted_folder" {
		t.Fatalf("source kind = %+v", resp.Results[0])
	}
}

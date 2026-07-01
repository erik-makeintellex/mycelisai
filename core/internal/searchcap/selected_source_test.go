package searchcap

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestServiceSelectedSourceRoutesToRegisteredLocalAPIEndpoint(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled, MaxResults: 5}, nil, nil)
	source, err := svc.AddSource(SourceInput{
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
	if err != nil {
		t.Fatalf("AddSource: %v", err)
	}
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "selected-search.local" || r.URL.Path != "/query" {
			t.Fatalf("selected endpoint = %s", r.URL.String())
		}
		if r.URL.Query().Get("q") != "release research" {
			t.Fatalf("query params = %q", r.URL.RawQuery)
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
	groupSource, err := svc.AddSource(SourceInput{
		Name:       "Group docs",
		Provider:   "local_api",
		Endpoint:   "https://search.example.test/api",
		Scope:      "group",
		ScopeRef:   "research-team",
		Boundary:   "research group only",
		AuthScheme: "none",
		Status:     "available",
	})
	if err != nil {
		t.Fatalf("AddSource group: %v", err)
	}
	resp, err := svc.Search(context.Background(), Request{Query: "docs", SourceID: groupSource.ID, TeamID: "marketing-team"})
	if err != nil {
		t.Fatalf("Search scope: %v", err)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "search_source_out_of_scope" {
		t.Fatalf("scope blocker = %+v", resp.Blocker)
	}

	authSource, err := svc.AddSource(SourceInput{
		Name:       "Private API",
		Provider:   "local_api",
		Endpoint:   "https://private.example.test/api",
		Boundary:   "private search",
		AuthScheme: "api_token",
		SecretRef:  "PRIVATE_SEARCH_TOKEN",
		Status:     "available",
	})
	if err != nil {
		t.Fatalf("AddSource auth: %v", err)
	}
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
	source, err := svc.AddSource(SourceInput{
		Name:       "Private API",
		Provider:   "local_api",
		Endpoint:   "http://private.example.test/api",
		Boundary:   "private search",
		AuthScheme: "api_token",
		SecretRef:  "env:PRIVATE_SEARCH_TOKEN",
		Status:     "available",
	})
	if err != nil {
		t.Fatalf("AddSource auth: %v", err)
	}
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
	source, err := svc.AddSource(SourceInput{
		Name:       "Private SearXNG",
		Provider:   "searxng",
		Endpoint:   "http://searxng.example.test",
		Boundary:   "private web search",
		AuthScheme: "bearer_token",
		SecretRef:  "PRIVATE_SEARCH_TOKEN",
		Status:     "available",
	})
	if err != nil {
		t.Fatalf("AddSource auth: %v", err)
	}
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

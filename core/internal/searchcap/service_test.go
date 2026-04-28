package searchcap

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestServiceDisabledReturnsStructuredBlocker(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)

	resp, err := svc.Search(context.Background(), Request{Query: "can you search the web?", SourceScope: "web"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "blocked" {
		t.Fatalf("Status = %q, want blocked", resp.Status)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "search_provider_disabled" {
		t.Fatalf("Blocker = %+v", resp.Blocker)
	}
}

func TestServiceStatusExplainsTokenFreeSelfHostedPath(t *testing.T) {
	svc := NewService(Config{Provider: ProviderSearXNG, SearXNGEndpoint: "http://searxng.local", MaxResults: 5}, nil, nil)

	status := svc.Status()

	if !status.Enabled || !status.Configured {
		t.Fatalf("status = %+v, want enabled and configured", status)
	}
	if !status.SupportsPublicWeb {
		t.Fatalf("SupportsPublicWeb = false, want true")
	}
	if status.RequiresHostedAPIToken {
		t.Fatalf("RequiresHostedAPIToken = true, want false")
	}
	if status.SomaToolName != "web_search" || !status.DirectSomaInteraction {
		t.Fatalf("soma direct status = %+v", status)
	}
}

func TestServiceSearXNGNormalizesJSONResults(t *testing.T) {
	svc := NewService(Config{Provider: ProviderSearXNG, SearXNGEndpoint: "http://searxng.local", MaxResults: 5}, nil, nil)
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "mycelis search" {
			t.Fatalf("q = %q", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("format") != "json" {
			t.Fatalf("format = %q", r.URL.Query().Get("format"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"results":[{"title":"Result A","url":"https://example.test/a","content":"Snippet A"}]}`)),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "mycelis search", SourceScope: "web"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Results[0].SourceKind != "searxng" || resp.Results[0].TrustClass != "bounded_external" {
		t.Fatalf("result = %+v", resp.Results[0])
	}
}

func TestServiceSearXNGForbiddenExplainsJSONFormat(t *testing.T) {
	svc := NewService(Config{Provider: ProviderSearXNG, SearXNGEndpoint: "http://searxng.local"}, nil, nil)
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "mycelis search", SourceScope: "web"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "searxng_json_disabled" {
		t.Fatalf("Blocker = %+v", resp.Blocker)
	}
}

func TestConfigFromEnvAcceptsSelfHostedLocalAPI(t *testing.T) {
	t.Setenv("MYCELIS_SEARCH_PROVIDER", "self_hosted")
	t.Setenv("MYCELIS_SEARCH_LOCAL_API_ENDPOINT", "http://search.local/api/search")
	t.Setenv("MYCELIS_SEARCH_MAX_RESULTS", "3")

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
}

func TestServiceLocalAPINormalizesJSONResults(t *testing.T) {
	svc := NewService(Config{Provider: ProviderLocalAPI, LocalAPIEndpoint: "http://search.local/api/search", MaxResults: 5}, nil, nil)
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/search" {
			t.Fatalf("path = %q, want /api/search", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "mycelis search" || r.URL.Query().Get("query") != "mycelis search" {
			t.Fatalf("query params = %q", r.URL.RawQuery)
		}
		if r.URL.Query().Get("max_results") != "2" {
			t.Fatalf("max_results = %q", r.URL.Query().Get("max_results"))
		}
		if r.URL.Query().Get("allowed_domains") != "example.test" || r.URL.Query().Get("blocked_domains") != "blocked.test" {
			t.Fatalf("domain filters = %q", r.URL.RawQuery)
		}
		if r.Header.Get("X-Mycelis-Team-ID") != "team-1" || r.Header.Get("X-Mycelis-Run-ID") != "run-1" {
			t.Fatalf("missing scope headers: %+v", r.Header)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"results":[{"title":"Result A","url":"https://example.test/a","snippet":"Snippet A","score":0.75,"provider":"local","token":"secret"},{"title":"Result B","url":"https://example.test/b","content":"Snippet B"},{"title":"Result C","url":"https://example.test/c"}]}`)),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{
		Query: "mycelis search", SourceScope: "web", MaxResults: 2,
		AllowedDomains: []string{"example.test"}, BlockedDomains: []string{"blocked.test"},
		TeamID: "team-1", RunID: "run-1",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 2 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Results[0].SourceKind != "local_api" || resp.Results[0].TrustClass != "bounded_external" || resp.Results[0].SensitivityClass != "public" {
		t.Fatalf("result = %+v", resp.Results[0])
	}
	if resp.Results[0].Score != 0.75 {
		t.Fatalf("Score = %v, want 0.75", resp.Results[0].Score)
	}
	if _, ok := resp.Results[0].ProviderMetadata["token"]; ok {
		t.Fatalf("ProviderMetadata leaked raw token: %+v", resp.Results[0].ProviderMetadata)
	}
	if resp.Results[1].Snippet != "Snippet B" {
		t.Fatalf("Snippet = %q, want Snippet B", resp.Results[1].Snippet)
	}
}

func TestServiceLocalAPIRejectsRelativeEndpoint(t *testing.T) {
	svc := NewService(Config{Provider: ProviderLocalAPI, LocalAPIEndpoint: "/api/search"}, nil, nil)

	if _, err := svc.Search(context.Background(), Request{Query: "mycelis search", SourceScope: "web"}); err == nil {
		t.Fatalf("expected relative local API endpoint to be rejected")
	}
}

func TestServiceLocalAPIMissingEndpointBlocks(t *testing.T) {
	svc := NewService(Config{Provider: ProviderLocalAPI}, nil, nil)

	resp, err := svc.Search(context.Background(), Request{Query: "mycelis search", SourceScope: "web"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "missing_local_api_endpoint" {
		t.Fatalf("Blocker = %+v", resp.Blocker)
	}
}

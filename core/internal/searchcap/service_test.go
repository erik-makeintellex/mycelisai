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

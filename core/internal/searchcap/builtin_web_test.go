package searchcap

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestServiceBuiltinWebNormalizesHTMLResults(t *testing.T) {
	svc := NewService(Config{Provider: ProviderBuiltinWeb, MaxResults: 5}, nil, nil)
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "html.duckduckgo.com" {
			t.Fatalf("host = %q, want html.duckduckgo.com", r.URL.Host)
		}
		if r.URL.Query().Get("q") != "mycelis search" {
			t.Fatalf("q = %q", r.URL.Query().Get("q"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`
				<a rel="nofollow" class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.test%2Fa">Result <b>A</b></a>
				<a class="result__snippet">Snippet <b>A</b></a>
			`)),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "mycelis search", SourceScope: "web"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Results[0].SourceKind != ProviderBuiltinWeb || resp.Results[0].TrustClass != "bounded_external" {
		t.Fatalf("result = %+v", resp.Results[0])
	}
	if resp.Results[0].Title != "Result A" || resp.Results[0].URL != "https://example.test/a" || resp.Results[0].Snippet != "Snippet A" {
		t.Fatalf("result = %+v", resp.Results[0])
	}
}

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

func TestServiceAllScopeWithBuiltinWebIncludesMountedFolderAndWebResults(t *testing.T) {
	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "internal-notes.md"), []byte("internal public architecture comparison from mounted data"), 0o600)
	if err != nil {
		t.Fatalf("write mount: %v", err)
	}
	svc := NewService(Config{Provider: ProviderBuiltinWeb, MaxResults: 4}, nil, nil)
	seedManagedSource(t, svc, SourceInput{
		Name:       "Mounted research",
		Provider:   ProviderMountedFolder,
		SourceType: ProviderMountedFolder,
		Endpoint:   root,
		Boundary:   "approved mounted research",
		Status:     "available",
	})
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("q") != "internal public architecture" {
			t.Fatalf("q = %q", r.URL.Query().Get("q"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`
				<a class="result__a" href="https://example.test/public">Public result</a>
				<a class="result__snippet">Public web snippet</a>
			`)),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "internal public architecture", SourceScope: "all"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 2 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Metadata["source_coverage"] != "local_sources_and_web" {
		t.Fatalf("metadata = %+v, want full coverage", resp.Metadata)
	}
	if resp.Results[0].SourceKind != ProviderMountedFolder || resp.Results[1].SourceKind != ProviderBuiltinWeb {
		t.Fatalf("results = %+v, want mounted then web result", resp.Results)
	}
	if resp.Metadata["local_source_count"] != 1 || resp.Metadata["web_source_count"] != 1 {
		t.Fatalf("metadata = %+v, want local and web counts", resp.Metadata)
	}
}

func TestServiceAllScopeWithPublicProviderReportsWebOnlyWhenLocalUnavailable(t *testing.T) {
	svc := NewService(Config{Provider: ProviderBuiltinWeb, MaxResults: 2}, nil, nil)
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`
				<a class="result__a" href="https://example.test/web">Web only</a>
				<a class="result__snippet">Web-only snippet</a>
			`)),
		}, nil
	})}

	resp, err := svc.Search(context.Background(), Request{Query: "internal and public", SourceScope: "all"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Metadata["partial_source_scope"] != "web_only" || resp.Metadata["missing_source_scope"] != "local_sources" {
		t.Fatalf("metadata = %+v, want web-only partial scope", resp.Metadata)
	}
	if !strings.Contains(resp.Metadata["scope_warning"].(string), "Approved local data") {
		t.Fatalf("metadata = %+v, want local-source warning", resp.Metadata)
	}
}

func TestServiceAllScopeWithPublicProviderReportsLocalOnlyWhenWebBlocked(t *testing.T) {
	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "internal-notes.md"), []byte("internal public comparison from local data"), 0o600)
	if err != nil {
		t.Fatalf("write mount: %v", err)
	}
	svc := NewService(Config{
		Provider:         ProviderBuiltinWeb,
		MaxResults:       2,
		OnlineAllowed:    false,
		OnlineAllowedSet: true,
	}, nil, nil)
	seedManagedSource(t, svc, SourceInput{
		Name:       "Mounted research",
		Provider:   ProviderMountedFolder,
		SourceType: ProviderMountedFolder,
		Endpoint:   root,
		Boundary:   "approved mounted research",
		Status:     "available",
	})

	resp, err := svc.Search(context.Background(), Request{Query: "internal public", SourceScope: "all"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "ok" || resp.Count != 1 {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.Metadata["partial_source_scope"] != "local_sources_only" || resp.Metadata["missing_source_scope"] != "web" {
		t.Fatalf("metadata = %+v, want local-only partial scope", resp.Metadata)
	}
	if resp.Results[0].SourceKind != ProviderMountedFolder {
		t.Fatalf("results = %+v, want mounted folder result", resp.Results)
	}
}

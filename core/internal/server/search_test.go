package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/searchcap"
)

func TestHandleSearchDisabledProvider(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	})
	mux := setupMux(t, "POST /api/v1/search", s.HandleSearch)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/search", `{"query":"can you search the web?","source_scope":"web"}`)

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["ok"] != false {
		t.Fatalf("ok = %v, want false", resp["ok"])
	}
	data := resp["data"].(map[string]any)
	if data["status"] != "blocked" {
		t.Fatalf("status = %v", data["status"])
	}
	blocker := data["blocker"].(map[string]any)
	if blocker["code"] != "search_provider_disabled" {
		t.Fatalf("blocker = %+v", blocker)
	}
}

func TestHandleSearchStatusReportsDirectSomaPath(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderSearXNG, SearXNGEndpoint: "http://searxng.local", MaxResults: 6}, nil, nil)
	})
	mux := setupMux(t, "GET /api/v1/search/status", s.HandleSearchStatus)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/search/status", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["provider"] != "searxng" {
		t.Fatalf("provider = %v", data["provider"])
	}
	if data["direct_soma_interaction"] != true || data["soma_tool_name"] != "web_search" {
		t.Fatalf("soma status = %+v", data)
	}
	if data["requires_hosted_api_token"] != false {
		t.Fatalf("requires_hosted_api_token = %v, want false", data["requires_hosted_api_token"])
	}
	sources := data["sources"].([]any)
	if len(sources) != 1 {
		t.Fatalf("sources = %+v, want one configured source", sources)
	}
	source := sources[0].(map[string]any)
	if source["name"] != "Self-hosted public web" || source["scope_kind"] != "all" || source["auth_scheme"] != "none" {
		t.Fatalf("source = %+v, want user-readable SearXNG source", source)
	}
}

func TestHandleSearchRejectsBadJSON(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/search", s.HandleSearch)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/search", `{bad`)
	assertStatus(t, rr, http.StatusBadRequest)
}

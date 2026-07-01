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
	if source["endpoint"] != "http://searxng.local" || source["base_url"] != "http://searxng.local" {
		t.Fatalf("source endpoint = %+v", source)
	}
}

func TestHandleSearchSourcesListsConfiguredSource(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderLocalAPI, LocalAPIEndpoint: "http://search.local/api/search"}, nil, nil)
	})
	mux := setupMux(t, "GET /api/v1/search/sources", s.HandleSearchSources)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/search/sources", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	sources := resp["data"].([]any)
	if len(sources) != 1 {
		t.Fatalf("sources = %+v, want one configured source", sources)
	}
	source := sources[0].(map[string]any)
	if source["id"] != "local_api" || source["endpoint"] != "http://search.local/api/search" || source["auth_scheme"] != "service_managed" {
		t.Fatalf("source = %+v, want token-free configured local API source", source)
	}
}

func TestHandleSearchSourcesAddsGovernedSourceWithoutRawToken(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	})
	mux := setupMux(t, "POST /api/v1/search/sources", s.HandleSearchSources)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/search/sources", `{
		"name":"Group Research Search",
		"provider":"local_api",
		"endpoint":"http://search.local/api/search",
		"scope":"Group",
		"scope_ref":"research-team",
		"boundary":"operator-owned research index",
		"auth_scheme":"api_token",
		"secret_ref":"MYCELIS_RESEARCH_SEARCH_TOKEN",
		"sensitivity":"configured",
		"trust":"bounded_external",
		"mode":"live",
		"status":"available"
	}`)

	assertStatus(t, rr, http.StatusCreated)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	source := resp["data"].(map[string]any)
	if source["secret_ref"] != "MYCELIS_RESEARCH_SEARCH_TOKEN" || source["scope_kind"] != "group" || source["scope_ref"] != "research-team" {
		t.Fatalf("source = %+v", source)
	}
	if _, leaked := source["token"]; leaked {
		t.Fatalf("source leaked token field: %+v", source)
	}

	sources := s.Search.ListSources()
	if len(sources) != 1 || sources[0].Name != "Group Research Search" {
		t.Fatalf("registered sources = %+v", sources)
	}
}

func TestHandleSearchSourcesRejectsRawCredentialFields(t *testing.T) {
	s := newTestServer(func(s *AdminServer) {
		s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	})
	mux := setupMux(t, "POST /api/v1/search/sources", s.HandleSearchSources)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/search/sources", `{"name":"Bad","provider":"brave","auth_scheme":"api_token","api_key":"raw-token"}`)

	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleSearchRejectsBadJSON(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/search", s.HandleSearch)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/search", `{bad`)
	assertStatus(t, rr, http.StatusBadRequest)
}

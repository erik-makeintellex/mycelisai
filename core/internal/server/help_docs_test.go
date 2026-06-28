package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestHandleDocsListReadAndSearch(t *testing.T) {
	s := newTestServer()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/docs", s.HandleDocsList)
	mux.HandleFunc("GET /api/v1/docs/search", s.HandleDocsSearch)
	mux.HandleFunc("GET /api/v1/docs/{slug}", s.HandleDocsRead)

	list := doRequest(t, mux, "GET", "/api/v1/docs", "")
	assertStatus(t, list, http.StatusOK)
	if !strings.Contains(list.Body.String(), "mycelis-canonical-prd") {
		t.Fatalf("expected canonical PRD in docs list: %s", list.Body.String())
	}

	read := doRequest(t, mux, "GET", "/api/v1/docs/soma-chat", "")
	assertStatus(t, read, http.StatusOK)
	if !strings.Contains(read.Body.String(), "docs/user/soma-chat.md") || !strings.Contains(read.Body.String(), "Using Soma Chat") {
		t.Fatalf("expected citable Soma docs content: %s", read.Body.String())
	}

	search := doRequest(t, mux, "GET", "/api/v1/docs/search?q=governed+context&limit=2", "")
	assertStatus(t, search, http.StatusOK)
	if !strings.Contains(search.Body.String(), `"results"`) || !strings.Contains(search.Body.String(), `"slug"`) {
		t.Fatalf("expected citable search results: %s", search.Body.String())
	}
}

func TestHandleDocsReadUnknownSlug(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/docs/{slug}", s.HandleDocsRead)

	rr := doRequest(t, mux, "GET", "/api/v1/docs/missing-doc", "")
	assertStatus(t, rr, http.StatusNotFound)
}

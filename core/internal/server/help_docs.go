package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/mycelis/core/internal/helpdocs"
)

func (s *AdminServer) HandleDocsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	store, err := helpdocs.NewStore("")
	if err != nil {
		respondError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{"sections": store.Sections()}})
}

func (s *AdminServer) HandleDocsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slug := strings.TrimSpace(r.PathValue("slug"))
	store, err := helpdocs.NewStore("")
	if err != nil {
		respondError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	doc, err := store.Read(slug)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	respondJSON(w, map[string]any{"ok": true, "data": doc})
}

func (s *AdminServer) HandleDocsSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	store, err := helpdocs.NewStore("")
	if err != nil {
		respondError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	results, err := store.Search(query, limit)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondJSON(w, map[string]any{"ok": true, "data": map[string]any{"results": results}})
}

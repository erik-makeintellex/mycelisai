package server

import (
	"encoding/json"
	"net/http"

	"github.com/mycelis/core/internal/searchcap"
)

func (s *AdminServer) searchCapabilityStatus() searchcap.Status {
	searchSvc := s.Search
	if searchSvc == nil {
		searchSvc = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	}
	return searchSvc.Status()
}

func (s *AdminServer) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req searchcap.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid search request", http.StatusBadRequest)
		return
	}
	searchSvc := s.Search
	if searchSvc == nil {
		searchSvc = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	}
	resp, err := searchSvc.Search(r.Context(), req)
	if err != nil {
		respondError(w, "search failed", http.StatusInternalServerError)
		return
	}
	respondJSON(w, map[string]any{
		"ok":   resp.Status == "ok",
		"data": resp,
	})
}

func (s *AdminServer) HandleSearchStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	respondJSON(w, map[string]any{
		"ok":   true,
		"data": s.searchCapabilityStatus(),
	})
}

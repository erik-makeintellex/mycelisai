package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/searchcap"
)

func (s *AdminServer) searchCapabilityStatus() searchcap.Status {
	return s.searchService().Status()
}

func (s *AdminServer) searchService() *searchcap.Service {
	if s.Search != nil {
		return s.Search
	}
	s.Search = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	return s.Search
}

// Search failures use the blocker envelope (MEM-LANES LOW); nothing was
// searched in either case.
const codeSearchFailed = "search_failed"

var (
	searchMemoryAccessBlocker = roleBlockerText{
		User: blockerText{"Saved memory access could not be checked, so nothing was searched.",
			"Try again in a moment. If it keeps happening, ask an admin to check the memory store."},
		Admin: blockerText{Action: "Check the Core database and team membership lookups, then try again. Nothing was searched."},
	}
	searchFailedBlocker = roleBlockerText{
		User: blockerText{"Search failed, so there are no results.",
			"Try again in a moment. If it keeps happening, ask an admin to check the search provider."},
		Admin: blockerText{Action: "Check the configured search provider and the Core logs, then try again."},
	}
)

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
	// Local sources return only saved memory the caller may read (SRU).
	reader, err := s.memoryReader(r)
	if err != nil {
		log.Printf("search: memory access check failed: %v", err)
		respondBlockerText(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, searchMemoryAccessBlocker, "memory access check failed: "+err.Error(), nil)
		return
	}
	req.Reader = &reader
	resp, err := s.searchService().Search(r.Context(), req)
	if err != nil {
		log.Printf("search: failed: %v", err)
		respondBlockerText(w, r, http.StatusInternalServerError, codeSearchFailed, searchFailedBlocker, "search failed: "+err.Error(), nil)
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

func (s *AdminServer) HandleSearchSources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respondJSON(w, map[string]any{
			"ok":   true,
			"data": s.searchService().ListSources(),
		})
	case http.MethodPost:
		var req searchcap.SourceInput
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			respondError(w, "invalid search source request", http.StatusBadRequest)
			return
		}
		source, err := s.searchService().AddSourceWithContext(r.Context(), req)
		if err != nil {
			respondError(w, err.Error(), searchcap.SourceErrorStatus(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"data": source,
		})
	default:
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *AdminServer) HandleSearchSource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	switch r.Method {
	case http.MethodPatch:
		var req searchcap.SourceInput
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			respondError(w, "invalid search source request", http.StatusBadRequest)
			return
		}
		source, err := s.searchService().UpdateSourceWithContext(r.Context(), id, req)
		if err != nil {
			respondError(w, err.Error(), searchcap.SourceErrorStatus(err))
			return
		}
		respondJSON(w, map[string]any{
			"ok":   true,
			"data": source,
		})
	case http.MethodDelete:
		if err := s.searchService().DeleteSourceWithContext(r.Context(), id); err != nil {
			respondError(w, err.Error(), searchcap.SourceErrorStatus(err))
			return
		}
		respondJSON(w, map[string]any{
			"ok": true,
			"data": map[string]any{
				"id":      id,
				"deleted": true,
			},
		})
	default:
		respondError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

package server

import (
	"net/http"

	"github.com/mycelis/core/pkg/protocol"
)

// RegisterRunRoutes registers run list, event timeline, stage projection, and chain endpoints.
func (s *AdminServer) RegisterRunRoutes(mux *http.ServeMux) {
	if mux == nil {
		return
	}
	mux.HandleFunc("GET /api/v1/runs", s.handleListRuns)
	mux.HandleFunc("GET /api/v1/runs/{id}/events", s.handleGetRunEvents)
	mux.HandleFunc("GET /api/v1/runs/{id}/stages", s.handleGetRunStages)
	mux.HandleFunc("GET /api/v1/runs/{id}/chain", s.handleGetRunChain)
}

// handleGetRunEvents returns the chronological event timeline for a run.
// GET /api/v1/runs/{id}/events
func (s *AdminServer) handleGetRunEvents(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		respondError(w, "run_id is required", http.StatusBadRequest)
		return
	}
	if s.Events == nil {
		respondError(w, "event store not initialized", http.StatusServiceUnavailable)
		return
	}

	timeline, err := s.Events.GetRunTimeline(r.Context(), runID)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, timeline)
}

// handleListRuns returns recent runs across all missions for the default tenant, newest first.
// GET /api/v1/runs
func (s *AdminServer) handleListRuns(w http.ResponseWriter, r *http.Request) {
	if s.Runs == nil {
		respondAPIError(w, "run store not initialized", http.StatusServiceUnavailable)
		return
	}

	recentRuns, err := s.Runs.ListRecentRuns(r.Context(), "default", 20)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(recentRuns))
}

// handleGetRunChain returns the causal chain for a run:
// the run itself + all sibling runs for the same mission (chain context).
// GET /api/v1/runs/{id}/chain
func (s *AdminServer) handleGetRunChain(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		respondError(w, "run_id is required", http.StatusBadRequest)
		return
	}
	if s.Runs == nil {
		respondError(w, "run store not initialized", http.StatusServiceUnavailable)
		return
	}

	// Fetch the run itself to get mission_id
	run, err := s.Runs.GetRun(r.Context(), runID)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return all runs for the mission (chain context for Team E's ViewChain component)
	chain, err := s.Runs.ListRunsForMission(r.Context(), run.MissionID, 20)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.Events != nil {
		for i := range chain {
			events, evErr := s.Events.GetRunTimeline(r.Context(), chain[i].ID)
			if evErr == nil && len(events) > 0 {
				if chain[i].Metadata == nil {
					chain[i].Metadata = map[string]interface{}{}
				}
				for _, ev := range events {
					if len(chain) > 1 && string(ev.EventType) == "signal.published" {
						continue
					}
					if ev.SourceTeam != "" && ev.SourceTeam != "governance" {
						chain[i].Metadata["team_id"] = ev.SourceTeam
					}
					for _, key := range []string{"team_id", "team_name", "listens_on", "speaks_on"} {
						if val := firstNonEmptyString(ev.Payload[key]); val != "" {
							chain[i].Metadata[key] = val
						}
					}
				}
			}
		}
	}

	respondJSON(w, map[string]interface{}{
		"run_id":     runID,
		"mission_id": run.MissionID,
		"chain":      chain,
	})
}

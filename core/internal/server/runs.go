package server

import (
	"net/http"
)

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

	respondJSON(w, map[string]interface{}{
		"run_id":     runID,
		"mission_id": run.MissionID,
		"chain":      chain,
	})
}

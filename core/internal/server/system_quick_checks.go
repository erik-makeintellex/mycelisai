package server

import (
	"net/http"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

type SystemQuickCheckResult struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Status    string    `json:"status"`
	Detail    string    `json:"detail,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// HandleSystemQuickCheck runs focused operator checks that are more specific
// than the broad services snapshot.
func (s *AdminServer) HandleSystemQuickCheck(w http.ResponseWriter, r *http.Request) {
	switch r.PathValue("id") {
	case "scheduler":
		s.handleSchedulerQuickCheck(w)
	default:
		respondAPIError(w, "Unknown system quick check", http.StatusNotFound)
	}
}

func (s *AdminServer) handleSchedulerQuickCheck(w http.ResponseWriter) {
	result := SystemQuickCheckResult{
		ID:        "scheduler",
		Label:     "Automation timing",
		Status:    "failure",
		Detail:    "Review-loop scheduler is not initialized",
		CheckedAt: time.Now(),
	}
	if s.LoopScheduler != nil {
		if s.LoopScheduler.isRunning() {
			result.Status = "healthy"
			result.Detail = "Review-loop scheduler is running"
		} else {
			result.Status = "degraded"
			result.Detail = "Review-loop scheduler exists but is not running"
		}
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

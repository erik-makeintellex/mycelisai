package server

import (
	"encoding/json"
	"net/http"
)

// GetMemoryStream handles GET /api/v1/memory/stream
func (s *AdminServer) GetMemoryStream(w http.ResponseWriter, r *http.Request) {
	if s.Mem == nil {
		http.Error(w, "Memory system not initialized", http.StatusServiceUnavailable)
		return
	}

	logs, err := s.Mem.ListRecent(50)
	if err != nil {
		http.Error(w, "Memory access failure: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

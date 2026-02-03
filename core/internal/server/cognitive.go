package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/cognitive"
)

// POST /api/v1/cognitive/infer
func (s *AdminServer) handleInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	var req cognitive.InferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	resp, err := s.Cognitive.Infer(req)
	if err != nil {
		log.Printf("Inference Failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, resp)
}

// GET /api/v1/cognitive/matrix
// Returns the current Brain Configuration (Profiles + Providers)
func (s *AdminServer) handleGetMatrix(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Cognitive == nil || s.Cognitive.Config == nil {
		http.Error(w, "Cognitive Matrix Offline", http.StatusServiceUnavailable)
		return
	}

	// Return the raw config struct
	respondJSON(w, s.Cognitive.Config)
}

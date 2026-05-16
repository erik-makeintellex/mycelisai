package server

import (
	"encoding/json"
	"log"
	"net/http"
)

// POST /api/v1/provision/draft
func (s *AdminServer) HandleProvisionDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Intent string `json:"intent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	if s.Provisioner == nil {
		http.Error(w, "Provisioning Engine Offline", http.StatusServiceUnavailable)
		return
	}

	manifest, err := s.Provisioner.Draft(r.Context(), payload.Intent)
	if err != nil {
		log.Printf("Provision Draft Failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, manifest)
}

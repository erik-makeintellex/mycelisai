package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/provisioning"
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

// POST /api/v1/provision/deploy
func (s *AdminServer) HandleProvisionDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var manifest provisioning.ServiceManifest
	if err := json.NewDecoder(r.Body).Decode(&manifest); err != nil {
		http.Error(w, "Bad Manifest JSON", http.StatusBadRequest)
		return
	}

	// TODO (Phase 20): Save to DB and Publish NATS event
	log.Printf("DEPLOYING AGENT: %s", manifest.Name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "deployed", "id": "placeholder-uuid"}`))
}

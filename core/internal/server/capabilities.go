package server

import (
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) HandleListCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Capabilities == nil {
		respondAPIError(w, "capability manifest service not initialized", http.StatusServiceUnavailable)
		return
	}
	snap, err := s.Capabilities.List(r.Context())
	if err != nil {
		respondAPIError(w, "Failed to list capability manifests: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(snap))
}

func (s *AdminServer) HandleGetCapability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Capabilities == nil {
		respondAPIError(w, "capability manifest service not initialized", http.StatusServiceUnavailable)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "Capability ID is required", http.StatusBadRequest)
		return
	}
	manifest, err := s.Capabilities.Get(r.Context(), id)
	if err != nil {
		respondAPIError(w, "Failed to get capability manifest: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if manifest == nil {
		respondAPIError(w, "Capability manifest not found", http.StatusNotFound)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(manifest))
}

func (s *AdminServer) HandleRefreshCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondAPIError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Capabilities == nil {
		respondAPIError(w, "capability manifest service not initialized", http.StatusServiceUnavailable)
		return
	}
	snap, err := s.Capabilities.Refresh(r.Context())
	if err != nil {
		respondAPIError(w, "Failed to refresh capability manifests: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(snap))
}

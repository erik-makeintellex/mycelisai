package server

import (
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) handleTriggerLoop(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	loopID := strings.TrimSpace(r.PathValue("loopId"))
	if orgID == "" || loopID == "" {
		respondAPIError(w, "organization id and loop id are required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(orgID)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	s.loopProfileStore().EnsureDefaults(home)
	profile, ok := s.loopProfileStore().Get(orgID, loopID)
	if !ok {
		respondAPIError(w, "loop profile not found", http.StatusNotFound)
		return
	}

	result, err := s.executeReviewLoop(home, profile, "manual")
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

func (s *AdminServer) handleListLoopResults(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	if orgID == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	if _, ok := s.organizationStore().Get(orgID); !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(s.loopResultStore().List(orgID)))
}

func (s *AdminServer) handleListLoopActivity(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	if orgID == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	if _, ok := s.organizationStore().Get(orgID); !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(recentLoopActivity(s.loopResultStore().List(orgID), 8)))
}

func (s *AdminServer) handleListLearningInsights(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	if orgID == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	if _, ok := s.organizationStore().Get(orgID); !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(organizationLearningInsights(s.loopResultStore().List(orgID), 6)))
}

func (s *AdminServer) handleListAutomations(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	if orgID == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(orgID)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	home = normalizeOrganizationHome(home)
	s.loopProfileStore().EnsureDefaults(home)
	items := organizationAutomations(home, s.loopProfileStore().ListByOrganization(orgID), s.loopResultStore().List(orgID))
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}

package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) handleUpdateDepartmentAIEngine(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, organizationsWriteScope); !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	departmentID := strings.TrimSpace(r.PathValue("departmentId"))
	if id == "" || departmentID == "" {
		respondAPIError(w, "organization id and department id are required", http.StatusBadRequest)
		return
	}

	var req DepartmentAIEngineUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Department AI Engine update request", http.StatusBadRequest)
		return
	}

	profileID := strings.TrimSpace(req.ProfileID)
	if req.RevertToOrganizationDefault {
		profileID = ""
	} else {
		if profileID == "" {
			respondAPIError(w, "profile_id is required unless reverting to the organization default", http.StatusBadRequest)
			return
		}
		if _, ok := lookupOrganizationAIEngineProfile(profileID); !ok {
			respondAPIError(w, "profile_id must be one of the guided AI Engine options", http.StatusBadRequest)
			return
		}
	}

	departmentFound := false
	updated, err := s.organizationStore().UpdateChecked(r.Context(), id, func(home OrganizationHomePayload) (OrganizationHomePayload, error) {
		home = normalizeOrganizationHome(home)
		for index, department := range home.Departments {
			if department.ID != departmentID {
				continue
			}
			departmentFound = true
			if profileID == "" || profileID == home.AIEngineProfileID {
				department.AIEngineOverrideProfileID = ""
				department.AIEngineOverrideSummary = ""
			} else {
				department.AIEngineOverrideProfileID = profileID
				department.AIEngineOverrideSummary = organizationAIEngineSummaryForProfile(profileID)
			}
			home.Departments[index] = department
			break
		}
		return organizationProfileTargetResult(home, departmentFound, true)
	})
	if err != nil {
		respondOrganizationStoreError(w, err)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateAgentTypeAIEngine(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, organizationsWriteScope); !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	departmentID := strings.TrimSpace(r.PathValue("departmentId"))
	agentTypeID := strings.TrimSpace(r.PathValue("agentTypeId"))
	if id == "" || departmentID == "" || agentTypeID == "" {
		respondAPIError(w, "organization id, department id, and agent type id are required", http.StatusBadRequest)
		return
	}

	var req AgentTypeAIEngineUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Agent Type AI Engine update request", http.StatusBadRequest)
		return
	}

	profileID := strings.TrimSpace(req.ProfileID)
	if req.UseTeamDefault {
		profileID = ""
	} else {
		if profileID == "" {
			respondAPIError(w, "profile_id is required unless returning to the Team default", http.StatusBadRequest)
			return
		}
		if _, ok := lookupOrganizationAIEngineProfile(profileID); !ok {
			respondAPIError(w, "profile_id must be one of the guided AI Engine options", http.StatusBadRequest)
			return
		}
	}

	departmentFound := false
	agentTypeFound := false
	updated, err := s.organizationStore().UpdateChecked(r.Context(), id, func(home OrganizationHomePayload) (OrganizationHomePayload, error) {
		home = normalizeOrganizationHome(home)
		for departmentIndex, department := range home.Departments {
			if department.ID != departmentID {
				continue
			}
			departmentFound = true
			for profileIndex, profile := range department.AgentTypeProfiles {
				if profile.ID != agentTypeID {
					continue
				}
				agentTypeFound = true
				if profileID == "" || profileID == department.AIEngineEffectiveProfileID {
					profile.AIEngineBindingProfileID = ""
				} else {
					profile.AIEngineBindingProfileID = profileID
				}
				department.AgentTypeProfiles[profileIndex] = profile
				break
			}
			home.Departments[departmentIndex] = department
			break
		}
		return organizationProfileTargetResult(home, departmentFound, agentTypeFound)
	})
	if err != nil {
		respondOrganizationStoreError(w, err)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateAgentTypeResponseContract(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, organizationsWriteScope); !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	departmentID := strings.TrimSpace(r.PathValue("departmentId"))
	agentTypeID := strings.TrimSpace(r.PathValue("agentTypeId"))
	if id == "" || departmentID == "" || agentTypeID == "" {
		respondAPIError(w, "organization id, department id, and agent type id are required", http.StatusBadRequest)
		return
	}

	var req AgentTypeResponseContractUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Agent Type Response Style update request", http.StatusBadRequest)
		return
	}

	profileID := strings.TrimSpace(req.ProfileID)
	if req.UseOrganizationOrTeamDefault {
		profileID = ""
	} else {
		if profileID == "" {
			respondAPIError(w, "profile_id is required unless returning to the Organization / Team default", http.StatusBadRequest)
			return
		}
		if _, ok := lookupResponseContractProfile(profileID); !ok {
			respondAPIError(w, "profile_id must be one of the guided Response Style options", http.StatusBadRequest)
			return
		}
	}

	departmentFound := false
	agentTypeFound := false
	updated, err := s.organizationStore().UpdateChecked(r.Context(), id, func(home OrganizationHomePayload) (OrganizationHomePayload, error) {
		home = normalizeOrganizationHome(home)
		defaultProfileID := strings.TrimSpace(home.ResponseContractProfileID)
		if defaultProfileID == "" {
			defaultProfileID = string(defaultResponseContractProfile().ID)
		}
		for departmentIndex, department := range home.Departments {
			if department.ID != departmentID {
				continue
			}
			departmentFound = true
			for profileIndex, profile := range department.AgentTypeProfiles {
				if profile.ID != agentTypeID {
					continue
				}
				agentTypeFound = true
				if profileID == "" || profileID == defaultProfileID {
					profile.ResponseContractBindingProfileID = ""
				} else {
					profile.ResponseContractBindingProfileID = profileID
				}
				department.AgentTypeProfiles[profileIndex] = profile
				break
			}
			home.Departments[departmentIndex] = department
			break
		}
		return organizationProfileTargetResult(home, departmentFound, agentTypeFound)
	})
	if err != nil {
		respondOrganizationStoreError(w, err)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

var (
	errOrganizationDepartmentNotFound = errors.New("department not found")
	errOrganizationAgentTypeNotFound  = errors.New("agent type profile not found")
)

// organizationProfileTargetResult aborts the locked update (no write) when the
// addressed department or agent type does not exist.
func organizationProfileTargetResult(home OrganizationHomePayload, departmentFound, agentTypeFound bool) (OrganizationHomePayload, error) {
	if !departmentFound {
		return home, errOrganizationDepartmentNotFound
	}
	if !agentTypeFound {
		return home, errOrganizationAgentTypeNotFound
	}
	return normalizeOrganizationHome(home), nil
}

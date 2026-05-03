package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) handleUpdateDepartmentAIEngine(w http.ResponseWriter, r *http.Request) {
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
	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
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
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}
	if !departmentFound {
		respondAPIError(w, "department not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateAgentTypeAIEngine(w http.ResponseWriter, r *http.Request) {
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
	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
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
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}
	if !departmentFound {
		respondAPIError(w, "department not found", http.StatusNotFound)
		return
	}
	if !agentTypeFound {
		respondAPIError(w, "agent type profile not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateAgentTypeResponseContract(w http.ResponseWriter, r *http.Request) {
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
	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
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
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}
	if !departmentFound {
		respondAPIError(w, "department not found", http.StatusNotFound)
		return
	}
	if !agentTypeFound {
		respondAPIError(w, "agent type profile not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

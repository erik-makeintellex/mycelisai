package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleUpdateAgentTypeResponseContract_SetsBinding(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           2,
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "delivery-specialist", Name: "Delivery Specialist", HelpsWith: "Carries execution work."},
				},
			},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/response-contract", s.handleUpdateAgentTypeResponseContract)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/agent-types/delivery-specialist/response-contract", `{"profile_id":"warm_supportive"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object response, got %T", resp.Data)
	}
	department := data["departments"].([]any)[0].(map[string]any)
	profile := department["agent_type_profiles"].([]any)[0].(map[string]any)
	if profile["inherits_default_response_contract"] != false || profile["response_contract_effective_summary"] != "Warm & Supportive" {
		t.Fatalf("expected type-specific Response Style binding, got %+v", profile)
	}
}

func TestHandleUpdateAgentTypeResponseContract_RevertsToOrganizationOrTeamDefault(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           2,
			ResponseContractProfileID: "concise_direct",
			ResponseContractSummary:   "Concise & Direct",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "delivery-specialist", Name: "Delivery Specialist", HelpsWith: "Carries execution work.", ResponseContractBindingProfileID: "warm_supportive"},
				},
			},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/response-contract", s.handleUpdateAgentTypeResponseContract)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/agent-types/delivery-specialist/response-contract", `{"use_organization_or_team_default":true}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object response, got %T", resp.Data)
	}
	department := data["departments"].([]any)[0].(map[string]any)
	profile := department["agent_type_profiles"].([]any)[0].(map[string]any)
	if profile["inherits_default_response_contract"] != true || profile["response_contract_effective_summary"] != "Concise & Direct" {
		t.Fatalf("expected agent type to inherit Organization / Team default after revert, got %+v", profile)
	}
}

func TestHandleUpdateAgentTypeResponseContract_RejectsInvalidProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           2,
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "delivery-specialist", Name: "Delivery Specialist", HelpsWith: "Carries execution work."},
				},
			},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/response-contract", s.handleUpdateAgentTypeResponseContract)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/agent-types/delivery-specialist/response-contract", `{"profile_id":"raw_prompt_override"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestAgentTypeResponseContractInheritance_PersistsAcrossOrganizationUpdates(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           2,
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "planner", Name: "Planner", HelpsWith: "Keeps priorities clear.", ResponseContractBindingProfileID: "structured_analytical"},
					{ID: "delivery-specialist", Name: "Delivery Specialist", HelpsWith: "Carries execution work."},
				},
			},
		},
	})

	updateMux := http.NewServeMux()
	updateMux.HandleFunc("PATCH /api/v1/organizations/{id}/response-contract", s.handleUpdateResponseContract)
	updateMux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)

	updateRR := doRequest(t, updateMux, "PATCH", "/api/v1/organizations/"+created.ID+"/response-contract", `{"profile_id":"warm_supportive"}`)
	assertStatus(t, updateRR, http.StatusOK)

	homeRR := doRequest(t, updateMux, "GET", "/api/v1/organizations/"+created.ID+"/home", "")
	assertStatus(t, homeRR, http.StatusOK)

	var homeResp protocol.APIResponse
	assertJSON(t, homeRR, &homeResp)
	data, ok := homeResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected home object, got %T", homeResp.Data)
	}
	department := data["departments"].([]any)[0].(map[string]any)
	profiles := department["agent_type_profiles"].([]any)
	planner := profiles[0].(map[string]any)
	delivery := profiles[1].(map[string]any)
	if planner["response_contract_effective_summary"] != "Structured & Analytical" || planner["inherits_default_response_contract"] != false {
		t.Fatalf("expected type-bound Response Style to stay stable, got %+v", planner)
	}
	if delivery["response_contract_effective_summary"] != "Warm & Supportive" || delivery["inherits_default_response_contract"] != true {
		t.Fatalf("expected inheriting agent type to follow updated Organization / Team default, got %+v", delivery)
	}
}

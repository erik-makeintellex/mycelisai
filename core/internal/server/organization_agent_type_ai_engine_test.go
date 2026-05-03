package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleUpdateAgentTypeAIEngine_SetsBinding(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                      "org-123",
			Name:                    "Northstar Labs",
			Purpose:                 "Ship a focused AI engineering organization",
			StartMode:               OrganizationStartModeTemplate,
			TeamLeadLabel:           "Team Lead",
			DepartmentCount:         1,
			SpecialistCount:         2,
			AIEngineProfileID:       "balanced",
			AIEngineSettingsSummary: "Balanced",
			Status:                  "ready",
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
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/ai-engine", s.handleUpdateAgentTypeAIEngine)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/agent-types/delivery-specialist/ai-engine", `{"profile_id":"high_reasoning"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object response, got %T", resp.Data)
	}
	departments, ok := data["departments"].([]any)
	if !ok || len(departments) != 1 {
		t.Fatalf("expected one department, got %+v", data)
	}
	department := departments[0].(map[string]any)
	profiles, ok := department["agent_type_profiles"].([]any)
	if !ok || len(profiles) != 1 {
		t.Fatalf("expected one agent type profile, got %+v", department)
	}
	profile := profiles[0].(map[string]any)
	if profile["inherits_department_ai_engine"] != false || profile["ai_engine_effective_summary"] != "High Reasoning" {
		t.Fatalf("expected type-specific agent type binding, got %+v", profile)
	}
}

func TestHandleUpdateAgentTypeAIEngine_RevertsToTeamDefault(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                      "org-123",
			Name:                    "Northstar Labs",
			Purpose:                 "Ship a focused AI engineering organization",
			StartMode:               OrganizationStartModeTemplate,
			TeamLeadLabel:           "Team Lead",
			DepartmentCount:         1,
			SpecialistCount:         2,
			AIEngineProfileID:       "balanced",
			AIEngineSettingsSummary: "Balanced",
			Status:                  "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:                        "platform",
				Name:                      "Platform Department",
				SpecialistCount:           2,
				AIEngineOverrideProfileID: "fast_lightweight",
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "delivery-specialist", Name: "Delivery Specialist", HelpsWith: "Carries execution work.", AIEngineBindingProfileID: "high_reasoning"},
				},
			},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/ai-engine", s.handleUpdateAgentTypeAIEngine)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/agent-types/delivery-specialist/ai-engine", `{"use_team_default":true}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object response, got %T", resp.Data)
	}
	department := data["departments"].([]any)[0].(map[string]any)
	profile := department["agent_type_profiles"].([]any)[0].(map[string]any)
	if profile["inherits_department_ai_engine"] != true || profile["ai_engine_effective_summary"] != "Fast & Lightweight" {
		t.Fatalf("expected agent type to inherit Team default after revert, got %+v", profile)
	}
}

func TestHandleUpdateAgentTypeAIEngine_RejectsInvalidProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                      "org-123",
			Name:                    "Northstar Labs",
			Purpose:                 "Ship a focused AI engineering organization",
			StartMode:               OrganizationStartModeTemplate,
			TeamLeadLabel:           "Team Lead",
			DepartmentCount:         1,
			SpecialistCount:         2,
			AIEngineProfileID:       "balanced",
			AIEngineSettingsSummary: "Balanced",
			Status:                  "ready",
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
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/ai-engine", s.handleUpdateAgentTypeAIEngine)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/agent-types/delivery-specialist/ai-engine", `{"profile_id":"llama3.2"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestAgentTypeAIEngineInheritance_PersistsAcrossTeamUpdates(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                      "org-123",
			Name:                    "Northstar Labs",
			Purpose:                 "Ship a focused AI engineering organization",
			StartMode:               OrganizationStartModeTemplate,
			TeamLeadLabel:           "Team Lead",
			DepartmentCount:         1,
			SpecialistCount:         2,
			AIEngineProfileID:       "balanced",
			AIEngineSettingsSummary: "Balanced",
			Status:                  "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:                        "platform",
				Name:                      "Platform Department",
				SpecialistCount:           2,
				AIEngineOverrideProfileID: "fast_lightweight",
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "planner", Name: "Planner", HelpsWith: "Keeps priorities clear.", AIEngineBindingProfileID: "high_reasoning"},
					{ID: "delivery-specialist", Name: "Delivery Specialist", HelpsWith: "Carries execution work."},
				},
			},
		},
	})

	departmentMux := http.NewServeMux()
	departmentMux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/ai-engine", s.handleUpdateDepartmentAIEngine)
	departmentMux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)

	updateRR := doRequest(t, departmentMux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/ai-engine", `{"profile_id":"deep_planning"}`)
	assertStatus(t, updateRR, http.StatusOK)

	homeRR := doRequest(t, departmentMux, "GET", "/api/v1/organizations/"+created.ID+"/home", "")
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
	if planner["ai_engine_effective_summary"] != "High Reasoning" || planner["inherits_department_ai_engine"] != false {
		t.Fatalf("expected type-bound agent type to stay stable, got %+v", planner)
	}
	if delivery["ai_engine_effective_summary"] != "Deep Planning" || delivery["inherits_department_ai_engine"] != true {
		t.Fatalf("expected inheriting agent type to follow updated Team default, got %+v", delivery)
	}
}

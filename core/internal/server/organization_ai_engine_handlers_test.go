package server

import (
	"net/http"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleUpdateOrganizationAIEngine_StoresCuratedProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                      "org-123",
			Name:                    "Northstar Labs",
			Purpose:                 "Ship a focused AI engineering organization",
			StartMode:               OrganizationStartModeTemplate,
			TemplateName:            "Engineering Starter",
			TeamLeadLabel:           "Team Lead",
			AdvisorCount:            1,
			DepartmentCount:         1,
			SpecialistCount:         2,
			AIEngineProfileID:       "starter_defaults",
			AIEngineSettingsSummary: "Starter defaults included",
			Status:                  "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/ai-engine", s.handleUpdateOrganizationAIEngine)
	mux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)

	updateRR := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/ai-engine", `{"profile_id":"high_reasoning"}`)
	assertStatus(t, updateRR, http.StatusOK)

	var updateResp protocol.APIResponse
	assertJSON(t, updateRR, &updateResp)
	data, ok := updateResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", updateResp.Data)
	}
	if data["ai_engine_profile_id"] != "high_reasoning" {
		t.Fatalf("unexpected updated profile id: %+v", data)
	}
	if data["ai_engine_settings_summary"] != "High Reasoning" {
		t.Fatalf("unexpected updated profile summary: %+v", data)
	}

	homeRR := doRequest(t, mux, "GET", "/api/v1/organizations/"+created.ID+"/home", "")
	assertStatus(t, homeRR, http.StatusOK)

	var homeResp protocol.APIResponse
	assertJSON(t, homeRR, &homeResp)
	homeData, ok := homeResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object home data, got %T", homeResp.Data)
	}
	if homeData["ai_engine_profile_id"] != "high_reasoning" || homeData["ai_engine_settings_summary"] != "High Reasoning" {
		t.Fatalf("unexpected persisted home payload: %+v", homeData)
	}
	results := s.loopResultStore().List(created.ID)
	if len(results) != 1 || results[0].Trigger != "event:organization_ai_engine_changed" {
		t.Fatalf("expected event-driven review activity after AI Engine update, got %+v", results)
	}
}

func TestHandleUpdateOrganizationAIEngine_RejectsInvalidProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                      "org-123",
			Name:                    "Northstar Labs",
			Purpose:                 "Ship a focused AI engineering organization",
			StartMode:               OrganizationStartModeEmpty,
			TeamLeadLabel:           "Team Lead",
			AIEngineSettingsSummary: "Set up later in Advanced mode",
			Status:                  "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/ai-engine", s.handleUpdateOrganizationAIEngine)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/ai-engine", `{"profile_id":"llama3.2"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdateDepartmentAIEngine_SetsOverride(t *testing.T) {
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
			{ID: "platform", Name: "Platform Department", SpecialistCount: 2},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/ai-engine", s.handleUpdateDepartmentAIEngine)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/ai-engine", `{"profile_id":"high_reasoning"}`)
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
	department, ok := departments[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object department, got %T", departments[0])
	}
	if department["inherits_organization_ai_engine"] != false {
		t.Fatalf("expected department override, got %+v", department)
	}
	if department["ai_engine_effective_summary"] != "High Reasoning" {
		t.Fatalf("unexpected effective summary: %+v", department)
	}
}

func TestHandleUpdateDepartmentAIEngine_RevertsToOrganizationDefault(t *testing.T) {
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
			{ID: "platform", Name: "Platform Department", SpecialistCount: 2, AIEngineOverrideProfileID: "high_reasoning"},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/departments/{departmentId}/ai-engine", s.handleUpdateDepartmentAIEngine)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/departments/platform/ai-engine", `{"revert_to_organization_default":true}`)
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
	department, ok := departments[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object department, got %T", departments[0])
	}
	if department["inherits_organization_ai_engine"] != true {
		t.Fatalf("expected inherited state, got %+v", department)
	}
	if department["ai_engine_effective_summary"] != "Balanced" {
		t.Fatalf("expected inherited summary, got %+v", department)
	}
}

func TestDepartmentAIEngineInheritance_PersistsAcrossOrganizationUpdates(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                      "org-123",
			Name:                    "Northstar Labs",
			Purpose:                 "Ship a focused AI engineering organization",
			StartMode:               OrganizationStartModeTemplate,
			TeamLeadLabel:           "Team Lead",
			DepartmentCount:         2,
			SpecialistCount:         4,
			AIEngineProfileID:       "balanced",
			AIEngineSettingsSummary: "Balanced",
			Status:                  "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{ID: "planning", Name: "Planning Department", SpecialistCount: 2, AIEngineOverrideProfileID: "high_reasoning"},
			{ID: "delivery", Name: "Delivery Department", SpecialistCount: 2},
		},
	})

	updateMux := http.NewServeMux()
	updateMux.HandleFunc("PATCH /api/v1/organizations/{id}/ai-engine", s.handleUpdateOrganizationAIEngine)
	updateMux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)

	updateRR := doRequest(t, updateMux, "PATCH", "/api/v1/organizations/"+created.ID+"/ai-engine", `{"profile_id":"fast_lightweight"}`)
	assertStatus(t, updateRR, http.StatusOK)

	homeRR := doRequest(t, updateMux, "GET", "/api/v1/organizations/"+created.ID+"/home", "")
	assertStatus(t, homeRR, http.StatusOK)

	var homeResp protocol.APIResponse
	assertJSON(t, homeRR, &homeResp)
	data, ok := homeResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected home object, got %T", homeResp.Data)
	}
	departments, ok := data["departments"].([]any)
	if !ok || len(departments) != 2 {
		t.Fatalf("expected two departments, got %+v", data)
	}

	planning := departments[0].(map[string]any)
	delivery := departments[1].(map[string]any)
	if planning["ai_engine_effective_summary"] != "High Reasoning" || planning["inherits_organization_ai_engine"] != false {
		t.Fatalf("expected planning override to persist, got %+v", planning)
	}
	if delivery["ai_engine_effective_summary"] != "Fast & Lightweight" || delivery["inherits_organization_ai_engine"] != true {
		t.Fatalf("expected delivery to inherit updated organization default, got %+v", delivery)
	}
}

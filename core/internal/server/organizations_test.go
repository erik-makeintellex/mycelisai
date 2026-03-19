package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

const testOrganizationStarterBundle = `id: engineering-starter
name: Engineering Starter
description: Guided AI Organization for engineering work
template_version: v1alpha1
kernel:
  mode: adaptive_delivery
council:
  mode: advisory
provider_policy:
  provider: ollama
teams:
  - id: platform
    name: Platform Department
    strategy: deliver
    model: llama3.2
    inputs: [intent]
    deliveries: [execution_result]
    members:
      - id: team-lead
        role: lead
        model: llama3.2
        system_prompt: Lead the work.
        inputs: [intent]
        outputs: [plan]
        tools: [read_file]
      - id: builder
        role: builder
        model: llama3.2
        system_prompt: Build the work.
        inputs: [plan]
        outputs: [artifact]
        tools: [write_file]
`

func writeStarterBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "engineering-starter.yaml")
	if err := os.WriteFile(path, []byte(testOrganizationStarterBundle), 0o644); err != nil {
		t.Fatalf("write starter bundle: %v", err)
	}
	return dir
}

func withTemplateBundlesPath(path string) func(*AdminServer) {
	return func(s *AdminServer) {
		s.TemplateBundlesPath = path
	}
}

func TestHandleListTemplates_OrganizationStarters(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	rr := doRequest(t, http.HandlerFunc(s.handleListTemplatesAPI), "GET", "/api/v1/templates?view=organization-starters", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}

	templates, ok := resp.Data.([]any)
	if !ok {
		t.Fatalf("expected array data, got %T", resp.Data)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 starter template, got %d", len(templates))
	}

	first, ok := templates[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object template, got %T", templates[0])
	}
	if first["name"] != "Engineering Starter" {
		t.Fatalf("unexpected starter template name: %+v", first)
	}
	if first["department_count"] != float64(1) {
		t.Fatalf("expected 1 department, got %+v", first)
	}
	if first["specialist_count"] != float64(2) {
		t.Fatalf("expected 2 specialists, got %+v", first)
	}
}

func TestHandleCreateOrganization_FromTemplateAndGetHome(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/ai-engine", s.handleUpdateOrganizationAIEngine)
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)
	mux.HandleFunc("POST /api/v1/organizations", s.handleCreateOrganization)

	body := `{"name":"Northstar Labs","purpose":"Ship a focused AI engineering organization","start_mode":"template","template_id":"engineering-starter"}`
	rr := doRequest(t, mux, "POST", "/api/v1/organizations", body)
	assertStatus(t, rr, http.StatusCreated)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatal("expected ok=true")
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", resp.Data)
	}
	id, _ := data["id"].(string)
	if id == "" {
		t.Fatalf("expected created organization id, got %+v", data)
	}
	if data["template_name"] != "Engineering Starter" {
		t.Fatalf("expected template_name Engineering Starter, got %+v", data)
	}

	homeRR := doRequest(t, mux, "GET", "/api/v1/organizations/"+id+"/home", "")
	assertStatus(t, homeRR, http.StatusOK)

	var homeResp protocol.APIResponse
	assertJSON(t, homeRR, &homeResp)
	homeData, ok := homeResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object home payload, got %T", homeResp.Data)
	}
	if homeData["name"] != "Northstar Labs" {
		t.Fatalf("unexpected home payload: %+v", homeData)
	}
	if homeData["department_count"] != float64(1) || homeData["specialist_count"] != float64(2) {
		t.Fatalf("unexpected home counts: %+v", homeData)
	}
	if homeData["ai_engine_profile_id"] != "starter_defaults" {
		t.Fatalf("unexpected home AI Engine profile: %+v", homeData)
	}
	departments, ok := homeData["departments"].([]any)
	if !ok || len(departments) != 1 {
		t.Fatalf("expected one department payload, got %+v", homeData)
	}

	actionRR := doRequest(t, mux, "POST", "/api/v1/organizations/"+id+"/workspace/actions", `{"action":"plan_next_steps"}`)
	assertStatus(t, actionRR, http.StatusOK)

	var actionResp protocol.APIResponse
	assertJSON(t, actionRR, &actionResp)
	actionData, ok := actionResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object action payload, got %T", actionResp.Data)
	}
	if actionData["request_label"] != "Plan next steps for this organization" {
		t.Fatalf("unexpected action payload: %+v", actionData)
	}
	steps, ok := actionData["priority_steps"].([]any)
	if !ok || len(steps) == 0 {
		t.Fatalf("expected priority steps, got %+v", actionData)
	}
}

func TestHandleCreateOrganization_StartEmpty(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	rr := doRequest(t, http.HandlerFunc(s.handleCreateOrganization), "POST", "/api/v1/organizations", `{"name":"Blank Canvas","purpose":"Shape a new AI Organization","start_mode":"empty"}`)
	assertStatus(t, rr, http.StatusCreated)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", resp.Data)
	}
	if data["department_count"] != float64(0) || data["specialist_count"] != float64(0) {
		t.Fatalf("unexpected empty-start counts: %+v", data)
	}
	if data["ai_engine_settings_summary"] != "Set up later in Advanced mode" {
		t.Fatalf("unexpected empty-start AI Engine summary: %+v", data)
	}
}

func TestHandleListOrganizations_ReturnsCreatedSummaries(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	createBody := `{"name":"Atlas","purpose":"Resume me later","start_mode":"empty"}`
	createRR := doRequest(t, http.HandlerFunc(s.handleCreateOrganization), "POST", "/api/v1/organizations", createBody)
	assertStatus(t, createRR, http.StatusCreated)

	rr := doRequest(t, http.HandlerFunc(s.handleListOrganizations), "GET", "/api/v1/organizations?view=summary", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	list, ok := resp.Data.([]any)
	if !ok {
		raw, _ := json.Marshal(resp.Data)
		t.Fatalf("expected list data, got %T (%s)", resp.Data, raw)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 organization summary, got %d", len(list))
	}
}

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

func TestHandleTeamLeadGuidedAction_RejectsUnknownAction(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:            "org-123",
			Name:          "Northstar Labs",
			Purpose:       "Ship a focused AI engineering organization",
			StartMode:     OrganizationStartModeEmpty,
			TeamLeadLabel: "Team Lead",
			Status:        "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"launch_agents"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleTeamLeadGuidedAction_RejectsMalformedRequest(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:            "org-123",
			Name:          "Northstar Labs",
			Purpose:       "Ship a focused AI engineering organization",
			StartMode:     OrganizationStartModeEmpty,
			TeamLeadLabel: "Team Lead",
			Status:        "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleTeamLeadGuidedAction_ReturnsNotFoundForMissingOrganization(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/org-missing/workspace/actions", `{"action":"plan_next_steps"}`)
	assertStatus(t, rr, http.StatusNotFound)
}

func TestBuildTeamLeadGuidance_UsesReadableFallbacksForPartialHome(t *testing.T) {
	response, err := buildTeamLeadGuidance(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			StartMode: OrganizationStartModeEmpty,
			Status:    "ready",
		},
	}, TeamLeadGuidedActionPlanNextSteps)
	if err != nil {
		t.Fatalf("buildTeamLeadGuidance returned error: %v", err)
	}

	if response.Headline != "Team Lead plan for this AI Organization" {
		t.Fatalf("unexpected fallback headline: %+v", response)
	}
	if response.Summary != "Team Lead recommends moving this AI Organization from setup into a focused first delivery loop." {
		t.Fatalf("unexpected fallback summary: %+v", response)
	}
	if len(response.PrioritySteps) != 3 {
		t.Fatalf("expected fallback priority steps, got %+v", response)
	}
	if response.PrioritySteps[0] != "Align the first outcome with this purpose: the current AI Organization priorities." {
		t.Fatalf("unexpected fallback priority steps: %+v", response.PrioritySteps)
	}
	if len(response.SuggestedFollowUps) != 3 {
		t.Fatalf("expected fallback follow-ups, got %+v", response)
	}
}

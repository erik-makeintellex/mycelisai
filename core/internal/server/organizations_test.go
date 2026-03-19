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

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
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

func mustResolveStarterTemplate(t *testing.T, s *AdminServer, id string) *OrganizationTemplateSummary {
	t.Helper()
	template, err := s.resolveStarterTemplate(id)
	if err != nil {
		t.Fatalf("resolve starter template %s: %v", id, err)
	}
	return template
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
	if homeData["response_contract_profile_id"] != "clear_balanced" || homeData["response_contract_summary"] != "Clear & Balanced" {
		t.Fatalf("unexpected response contract default: %+v", homeData)
	}
	departments, ok := homeData["departments"].([]any)
	if !ok || len(departments) != 1 {
		t.Fatalf("expected one department payload, got %+v", homeData)
	}
	department, ok := departments[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object department payload, got %T", departments[0])
	}
	profiles, ok := department["agent_type_profiles"].([]any)
	if !ok || len(profiles) != 2 {
		t.Fatalf("expected two agent type profiles, got %+v", department)
	}
	planner, ok := profiles[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object agent type profile, got %T", profiles[0])
	}
	if planner["name"] != "Planner" || planner["ai_engine_effective_summary"] != "High Reasoning" {
		t.Fatalf("unexpected planner profile payload: %+v", planner)
	}
	if planner["inherits_department_ai_engine"] != false || planner["inherits_default_response_contract"] != false {
		t.Fatalf("expected Planner to expose type-specific bindings, got %+v", planner)
	}
	if planner["output_type_id"] != "research_reasoning" || planner["output_model_effective_summary"] != "Qwen2.5 Coder 7B" {
		t.Fatalf("expected default single-model output routing on planner, got %+v", planner)
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

func TestHandleGetOrganizationOutputModelRouting_ListsInstalledAndPopularSelfHostedModels(t *testing.T) {
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]any{
				{"name": "qwen3:8b"},
				{"name": "qwen3:14b"},
				{"name": "llava:7b"},
				{"name": "qwen2.5-coder:14b"},
				{"name": "qwen2.5-coder:7b-instruct"},
			},
		})
	}))
	defer ollamaServer.Close()

	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)), func(s *AdminServer) {
		s.Cognitive = &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Providers: map[string]cognitive.ProviderConfig{
					"ollama": {Endpoint: ollamaServer.URL + "/v1"},
				},
			},
		}
	})
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/organizations/{id}/output-model-routing", s.handleGetOrganizationOutputModelRouting)

	rr := doRequest(t, mux, "GET", "/api/v1/organizations/"+created.ID+"/output-model-routing", "")
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object payload, got %T", resp.Data)
	}
	if data["routing_mode"] != "single_model" {
		t.Fatalf("expected default single-model routing, got %+v", data)
	}
	recommended, ok := data["recommended_models"].([]any)
	if !ok || len(recommended) < 2 {
		t.Fatalf("expected at least two recommended models, got %+v", data)
	}
	firstRecommended, _ := recommended[0].(map[string]any)
	secondRecommended, _ := recommended[1].(map[string]any)
	if firstRecommended["model_id"] != "llama3.1:8b" && firstRecommended["model_id"] != "qwen3:8b" {
		t.Fatalf("expected a popular self-hosted model first, got %+v", firstRecommended)
	}
	if secondRecommended["model_id"] != "llama3.1:8b" && secondRecommended["model_id"] != "qwen3:8b" {
		t.Fatalf("expected a popular self-hosted model second, got %+v", secondRecommended)
	}
	available, ok := data["available_models"].([]any)
	if !ok || len(available) == 0 {
		t.Fatalf("expected available model inventory, got %+v", data)
	}
	foundInstalledVision := false
	for _, item := range available {
		entry, _ := item.(map[string]any)
		if entry["model_id"] == "llava:7b" && entry["installed"] == true {
			foundInstalledVision = true
			break
		}
	}
	if !foundInstalledVision {
		t.Fatalf("expected installed llava:7b in available models, got %+v", available)
	}
	if data["review_permission_prompt"] == "" {
		t.Fatalf("expected owner review permission prompt, got %+v", data)
	}
	criteria, ok := data["automatic_selection_criteria"].([]any)
	if !ok || len(criteria) == 0 {
		t.Fatalf("expected automatic selection criteria, got %+v", data)
	}
	reviewCandidates, ok := data["review_candidates"].([]any)
	if !ok || len(reviewCandidates) == 0 {
		t.Fatalf("expected model review candidates, got %+v", data)
	}
	var foundCodeCandidate bool
	for _, item := range reviewCandidates {
		entry, _ := item.(map[string]any)
		if entry["output_type_id"] == "code_generation" {
			foundCodeCandidate = true
			if entry["model_id"] != "qwen2.5-coder:14b" || entry["installed"] != true {
				t.Fatalf("expected installed higher-capacity code candidate, got %+v", entry)
			}
			candidateCriteria, ok := entry["review_criteria"].([]any)
			if !ok || len(candidateCriteria) == 0 {
				t.Fatalf("expected code candidate criteria, got %+v", entry)
			}
		}
	}
	if !foundCodeCandidate {
		t.Fatalf("expected code generation review candidate, got %+v", reviewCandidates)
	}
}

func TestHandleUpdateOrganizationOutputModelRouting_AppliesDetectedRoleModels(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/output-model-routing", s.handleUpdateOrganizationOutputModelRouting)

	body := `{
		"routing_mode":"detected_output_types",
		"default_model_id":"qwen3:8b",
		"bindings":[
			{"output_type_id":"research_reasoning","model_id":"llama3.1:8b"},
			{"output_type_id":"code_generation","model_id":"qwen2.5-coder:7b"},
			{"output_type_id":"general_text","use_organization_default":true},
			{"output_type_id":"vision_analysis","model_id":"llava:7b"}
		]
	}`
	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/output-model-routing", body)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object payload, got %T", resp.Data)
	}
	if data["output_model_routing_mode"] != "detected_output_types" {
		t.Fatalf("expected detected routing mode, got %+v", data)
	}

	departments, ok := data["departments"].([]any)
	if !ok || len(departments) != 1 {
		t.Fatalf("expected departments payload, got %+v", data)
	}
	department, _ := departments[0].(map[string]any)
	profiles, ok := department["agent_type_profiles"].([]any)
	if !ok || len(profiles) < 2 {
		t.Fatalf("expected agent type profiles, got %+v", department)
	}

	var planner map[string]any
	var delivery map[string]any
	for _, item := range profiles {
		profile, _ := item.(map[string]any)
		switch profile["id"] {
		case "planner":
			planner = profile
		case "delivery-specialist":
			delivery = profile
		}
	}
	if planner["output_type_id"] != "research_reasoning" || planner["output_model_effective_summary"] != "Llama 3.1 8B" {
		t.Fatalf("expected planner to use research model, got %+v", planner)
	}
	if delivery["output_type_id"] != "code_generation" || delivery["output_model_effective_summary"] != "Qwen2.5 Coder 7B" {
		t.Fatalf("expected delivery specialist to use code model, got %+v", delivery)
	}
}

func TestHandleTeamLeadGuidedAction_AddsNativeTeamExecutionContractForImageRequests(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"plan_next_steps","request_context":"Create a creative team to generate a launch hero image."}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object action payload, got %T", resp.Data)
	}
	executionContract, ok := data["execution_contract"].(map[string]any)
	if !ok {
		t.Fatalf("expected execution contract, got %+v", data)
	}
	if executionContract["execution_mode"] != "native_team" {
		t.Fatalf("expected native_team execution mode, got %+v", executionContract)
	}
	if executionContract["team_name"] != "Creative Delivery Team" {
		t.Fatalf("expected creative team name, got %+v", executionContract)
	}
	outputs, ok := executionContract["target_outputs"].([]any)
	if !ok || len(outputs) < 1 {
		t.Fatalf("expected target outputs, got %+v", executionContract)
	}
	workflowGroup, ok := executionContract["workflow_group"].(map[string]any)
	if !ok {
		t.Fatalf("expected workflow group draft, got %+v", executionContract)
	}
	if workflowGroup["name"] != "Creative Delivery Team temporary workflow" {
		t.Fatalf("expected workflow group draft name, got %+v", workflowGroup)
	}
	if workflowGroup["work_mode"] != "propose_only" {
		t.Fatalf("expected propose_only workflow group mode, got %+v", workflowGroup)
	}
}

func TestHandleTeamLeadGuidedAction_AddsNativeTeamExecutionContractForMarketingRequests(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"plan_next_steps","request_context":"Create a temporary marketing launch team for a new product rollout."}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object action payload, got %T", resp.Data)
	}
	executionContract, ok := data["execution_contract"].(map[string]any)
	if !ok {
		t.Fatalf("expected execution contract, got %+v", data)
	}
	if executionContract["execution_mode"] != "native_team" {
		t.Fatalf("expected native_team execution mode, got %+v", executionContract)
	}
	if executionContract["team_name"] != "Marketing Launch Team" {
		t.Fatalf("expected marketing team name, got %+v", executionContract)
	}
	if executionContract["coordination_model"] != "compact_team" {
		t.Fatalf("expected compact team coordination, got %+v", executionContract)
	}
	if executionContract["recommended_team_member_limit"] != float64(6) {
		t.Fatalf("expected compact team member limit, got %+v", executionContract)
	}
	outputs, ok := executionContract["target_outputs"].([]any)
	if !ok || len(outputs) != 3 {
		t.Fatalf("expected three marketing outputs, got %+v", executionContract)
	}
	workflowGroup, ok := executionContract["workflow_group"].(map[string]any)
	if !ok {
		t.Fatalf("expected workflow group draft, got %+v", executionContract)
	}
	if workflowGroup["coordinator_profile"] != "Marketing Launch Team lead" {
		t.Fatalf("expected marketing workflow lead, got %+v", workflowGroup)
	}
	if workflowGroup["recommended_member_limit"] != float64(6) {
		t.Fatalf("expected recommended member limit on workflow group, got %+v", workflowGroup)
	}
}

func TestHandleTeamLeadGuidedAction_SplitsBroadRequestsIntoSmallTeamOrchestration(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"plan_next_steps","request_context":"Create a company-wide product launch program across marketing, sales, support, docs, and engineering so the organization can coordinate several workstreams at once."}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object action payload, got %T", resp.Data)
	}
	executionContract, ok := data["execution_contract"].(map[string]any)
	if !ok {
		t.Fatalf("expected execution contract, got %+v", data)
	}
	if executionContract["execution_mode"] != "native_team" {
		t.Fatalf("expected native_team execution mode, got %+v", executionContract)
	}
	if executionContract["coordination_model"] != "multi_team_orchestration" {
		t.Fatalf("expected multi-team orchestration guidance, got %+v", executionContract)
	}
	if executionContract["recommended_team_count"] != float64(3) {
		t.Fatalf("expected several small teams, got %+v", executionContract)
	}
	if executionContract["recommended_team_member_limit"] != float64(5) {
		t.Fatalf("expected compact per-team member limit, got %+v", executionContract)
	}
	if executionContract["recommended_team_shape"] == "" {
		t.Fatalf("expected recommended team shape guidance, got %+v", executionContract)
	}
	summary, _ := executionContract["summary"].(string)
	if !strings.Contains(strings.ToLower(summary), "several compact teams") {
		t.Fatalf("expected summary to describe small coordinated teams, got %q", summary)
	}
	workflowGroup, ok := executionContract["workflow_group"].(map[string]any)
	if !ok {
		t.Fatalf("expected workflow group draft, got %+v", executionContract)
	}
	if workflowGroup["recommended_member_limit"] != float64(5) {
		t.Fatalf("expected compact member cap in workflow group, got %+v", workflowGroup)
	}
	if workflowGroup["work_mode"] != "propose_only" {
		t.Fatalf("expected propose_only workflow group mode, got %+v", workflowGroup)
	}
}

func TestHandleTeamLeadGuidedAction_AddsExternalWorkflowContractForN8NRequests(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"plan_next_steps","request_context":"Create an n8n workflow contract for inbound leads."}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object action payload, got %T", resp.Data)
	}
	executionContract, ok := data["execution_contract"].(map[string]any)
	if !ok {
		t.Fatalf("expected execution contract, got %+v", data)
	}
	if executionContract["execution_mode"] != "external_workflow_contract" {
		t.Fatalf("expected external workflow execution mode, got %+v", executionContract)
	}
	if executionContract["external_target"] != "n8n workflow contract" {
		t.Fatalf("expected n8n external target, got %+v", executionContract)
	}
}

func TestHandleTeamLeadGuidedAction_AddsContinuityResumeContractForRetainedPackage(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	requestContext := "Retained package for release readiness after reboot."
	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"resume_retained_package","request_context":"`+requestContext+`"}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object action payload, got %T", resp.Data)
	}
	if data["action"] != "resume_retained_package" {
		t.Fatalf("expected resume action echo, got %+v", data)
	}
	if data["request_label"] != "Resume retained package continuity" {
		t.Fatalf("expected resume request label, got %+v", data)
	}

	executionContract, ok := data["execution_contract"].(map[string]any)
	if !ok {
		t.Fatalf("expected execution contract, got %+v", data)
	}
	if executionContract["execution_mode"] != "continuity_resume" {
		t.Fatalf("expected continuity resume mode, got %+v", executionContract)
	}
	if executionContract["continuity_label"] != "Retained package continuity" {
		t.Fatalf("expected retained package continuity label, got %+v", executionContract)
	}
	if executionContract["resume_checkpoint"] != "Continue from the last retained package after reload or reboot." {
		t.Fatalf("expected explicit resume checkpoint, got %+v", executionContract)
	}
	if executionContract["summary"] != "Resume the retained package for Northstar Labs, confirm completed work, and keep the remaining steps reviewable after a reboot or reload." {
		t.Fatalf("expected continuity summary, got %+v", executionContract)
	}
	outputs, ok := executionContract["target_outputs"].([]any)
	if !ok || len(outputs) != 3 {
		t.Fatalf("expected retained package outputs, got %+v", executionContract)
	}
	if outputs[0] != "Retained package continuity summary" || outputs[1] != "Completed work snapshot" || outputs[2] != "Remaining work checklist" {
		t.Fatalf("unexpected retained package outputs: %+v", outputs)
	}

	workflowGroup, ok := executionContract["workflow_group"].(map[string]any)
	if !ok {
		t.Fatalf("expected workflow group draft, got %+v", executionContract)
	}
	if workflowGroup["name"] != "Retained Package Continuity temporary workflow" {
		t.Fatalf("expected continuity workflow group name, got %+v", workflowGroup)
	}
	if workflowGroup["goal_statement"] != requestContext {
		t.Fatalf("expected request context to anchor workflow goal, got %+v", workflowGroup)
	}
	if workflowGroup["work_mode"] != "resume_continuity" {
		t.Fatalf("expected resume continuity workflow mode, got %+v", workflowGroup)
	}
	if workflowGroup["coordinator_profile"] != "Retained Package Continuity lead" {
		t.Fatalf("expected continuity coordinator profile, got %+v", workflowGroup)
	}
	if workflowGroup["recommended_member_limit"] != float64(4) {
		t.Fatalf("expected bounded member limit, got %+v", workflowGroup)
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

func TestHandleCreateOrganization_TriggersEventDrivenReviews(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))

	rr := doRequest(t, http.HandlerFunc(s.handleCreateOrganization), "POST", "/api/v1/organizations", `{"name":"Northstar Labs","purpose":"Ship a focused AI engineering organization","start_mode":"template","template_id":"engineering-starter"}`)
	assertStatus(t, rr, http.StatusCreated)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", resp.Data)
	}
	id, _ := data["id"].(string)
	if id == "" {
		t.Fatalf("expected organization id, got %+v", data)
	}

	results := s.loopResultStore().List(id)
	if len(results) != 2 {
		t.Fatalf("expected default event-driven review activity after create, got %+v", results)
	}
	for _, result := range results {
		if result.Trigger != "event:organization_created" {
			t.Fatalf("expected organization-created trigger label, got %+v", result)
		}
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

func TestHandleUpdateResponseContract_StoresCuratedProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           1,
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 1,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{ID: "planner", Name: "Planner", HelpsWith: "Keeps priorities clear."},
				},
			},
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/response-contract", s.handleUpdateResponseContract)
	mux.HandleFunc("GET /api/v1/organizations/{id}/home", s.handleGetOrganizationHome)

	updateRR := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/response-contract", `{"profile_id":"warm_supportive"}`)
	assertStatus(t, updateRR, http.StatusOK)

	var updateResp protocol.APIResponse
	assertJSON(t, updateRR, &updateResp)
	data, ok := updateResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", updateResp.Data)
	}
	if data["response_contract_profile_id"] != "warm_supportive" || data["response_contract_summary"] != "Warm & Supportive" {
		t.Fatalf("unexpected updated response contract: %+v", data)
	}

	homeRR := doRequest(t, mux, "GET", "/api/v1/organizations/"+created.ID+"/home", "")
	assertStatus(t, homeRR, http.StatusOK)

	var homeResp protocol.APIResponse
	assertJSON(t, homeRR, &homeResp)
	homeData, ok := homeResp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object home data, got %T", homeResp.Data)
	}
	if homeData["response_contract_profile_id"] != "warm_supportive" || homeData["response_contract_summary"] != "Warm & Supportive" {
		t.Fatalf("unexpected persisted response contract: %+v", homeData)
	}
	results := s.loopResultStore().List(created.ID)
	if len(results) != 2 {
		t.Fatalf("expected event-driven review activity after Response Style update, got %+v", results)
	}
	for _, result := range results {
		if result.Trigger != "event:response_contract_changed" {
			t.Fatalf("expected response-style event trigger label, got %+v", results)
		}
	}
}

func TestHandleUpdateResponseContract_RejectsInvalidProfile(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			ResponseContractProfileID: "clear_balanced",
			ResponseContractSummary:   "Clear & Balanced",
			Status:                    "ready",
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/organizations/{id}/response-contract", s.handleUpdateResponseContract)

	rr := doRequest(t, mux, "PATCH", "/api/v1/organizations/"+created.ID+"/response-contract", `{"profile_id":"raw_prompt_override"}`)
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

func TestNormalizeOrganizationHome_AgentTypeProfilesResolveInheritanceAndTypeBindings(t *testing.T) {
	home := normalizeOrganizationHome(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			DepartmentCount:           1,
			SpecialistCount:           2,
			AIEngineProfileID:         "balanced",
			AIEngineSettingsSummary:   "Balanced",
			ResponseContractProfileID: "warm_supportive",
			ResponseContractSummary:   "Warm & Supportive",
			Status:                    "ready",
		},
		Departments: []OrganizationDepartmentSummary{
			{
				ID:              "platform",
				Name:            "Platform Department",
				SpecialistCount: 2,
				AgentTypeProfiles: []OrganizationAgentTypeProfileSummary{
					{
						ID:                               "planner",
						Name:                             "Planner",
						HelpsWith:                        "Keeps priorities and sequencing clear.",
						AIEngineBindingProfileID:         "high_reasoning",
						ResponseContractBindingProfileID: "structured_analytical",
					},
					{
						ID:        "reviewer",
						Name:      "Reviewer",
						HelpsWith: "Checks quality before work moves forward.",
					},
				},
			},
		},
	})

	if len(home.Departments) != 1 || len(home.Departments[0].AgentTypeProfiles) != 2 {
		t.Fatalf("expected normalized agent type profiles, got %+v", home.Departments)
	}

	planner := home.Departments[0].AgentTypeProfiles[0]
	if planner.InheritsDepartmentAIEngine || planner.AIEngineEffectiveSummary != "High Reasoning" {
		t.Fatalf("expected planner AI Engine binding to stay type-specific, got %+v", planner)
	}
	if planner.InheritsDefaultResponseContract || planner.ResponseContractEffectiveSummary != "Structured & Analytical" {
		t.Fatalf("expected planner Response Style binding to stay type-specific, got %+v", planner)
	}

	reviewer := home.Departments[0].AgentTypeProfiles[1]
	if !reviewer.InheritsDepartmentAIEngine || reviewer.AIEngineEffectiveSummary != "Balanced" {
		t.Fatalf("expected reviewer AI Engine to inherit department default, got %+v", reviewer)
	}
	if !reviewer.InheritsDefaultResponseContract || reviewer.ResponseContractEffectiveSummary != "Warm & Supportive" {
		t.Fatalf("expected reviewer Response Style to inherit organization default, got %+v", reviewer)
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

func TestHandleTeamLeadGuidedAction_TriggersEventDrivenReview(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterBundle(t)))
	created := s.organizationStore().Save(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        "org-123",
			Name:                      "Northstar Labs",
			Purpose:                   "Ship a focused AI engineering organization",
			StartMode:                 OrganizationStartModeTemplate,
			TeamLeadLabel:             "Team Lead",
			AdvisorCount:              1,
			DepartmentCount:           1,
			SpecialistCount:           2,
			AIEngineProfileID:         "balanced",
			AIEngineSettingsSummary:   "Balanced",
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
					{ID: "planner", Name: "Planner", HelpsWith: "Keeps priorities clear."},
				},
			},
		},
	})
	s.loopProfileStore().EnsureDefaults(created)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"plan_next_steps"}`)
	assertStatus(t, rr, http.StatusOK)

	results := s.loopResultStore().List(created.ID)
	if len(results) != 1 || results[0].Trigger != "event:team_lead_action_completed" {
		t.Fatalf("expected event-driven review activity after Team Lead action, got %+v", results)
	}
}

func TestBuildTeamLeadGuidance_UsesReadableFallbacksForPartialHome(t *testing.T) {
	response, err := buildTeamLeadGuidance(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			StartMode: OrganizationStartModeEmpty,
			Status:    "ready",
		},
	}, TeamLeadGuidedActionPlanNextSteps, "")
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

func TestBuildTeamLeadGuidance_ResumeRetainedPackageUsesReadableFallbacksForPartialHome(t *testing.T) {
	response, err := buildTeamLeadGuidance(OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			StartMode:     OrganizationStartModeEmpty,
			TeamLeadLabel: "Team Lead",
		},
	}, TeamLeadGuidedActionResumeRetainedPackage, "")
	if err != nil {
		t.Fatalf("buildTeamLeadGuidance returned error: %v", err)
	}

	if response.Headline != "Retained package continuity for this AI Organization" {
		t.Fatalf("unexpected fallback headline: %+v", response)
	}
	if response.Summary != "Team Lead resumes the retained package for this AI Organization so completed work stays durable and the next step stays explicit after a reboot or reload." {
		t.Fatalf("unexpected fallback summary: %+v", response)
	}
	if len(response.PrioritySteps) != 3 {
		t.Fatalf("expected fallback priority steps, got %+v", response)
	}
	if response.PrioritySteps[0] != "Open the retained package and confirm the latest durable outputs." {
		t.Fatalf("unexpected fallback priority steps: %+v", response.PrioritySteps)
	}
	if response.ExecutionContract == nil {
		t.Fatal("expected execution contract")
	}
	if response.ExecutionContract.ExecutionMode != TeamLeadExecutionModeContinuityResume {
		t.Fatalf("expected continuity resume execution mode, got %+v", response.ExecutionContract)
	}
	if response.ExecutionContract.ContinuityLabel != "Retained package continuity" {
		t.Fatalf("expected continuity label, got %+v", response.ExecutionContract)
	}
	if response.ExecutionContract.WorkflowGroup == nil {
		t.Fatal("expected workflow group draft")
	}
	if response.ExecutionContract.WorkflowGroup.GoalStatement != "Resume the retained package for this AI Organization after a reboot or reload." {
		t.Fatalf("unexpected fallback goal statement: %+v", response.ExecutionContract.WorkflowGroup)
	}
}

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

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

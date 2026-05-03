package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

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
	workstreams, ok := executionContract["workstreams"].([]any)
	if !ok || len(workstreams) != 3 {
		t.Fatalf("expected creative workstreams, got %+v", executionContract)
	}
	firstWorkstream, ok := workstreams[0].(map[string]any)
	if !ok || firstWorkstream["label"] != "Creative direction lane" {
		t.Fatalf("expected creative direction workstream, got %+v", workstreams)
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
	workstreams, ok := executionContract["workstreams"].([]any)
	if !ok || len(workstreams) != 3 {
		t.Fatalf("expected compact workstreams, got %+v", executionContract)
	}
	firstWorkstream, ok := workstreams[0].(map[string]any)
	if !ok || firstWorkstream["label"] != "Planning lane" || firstWorkstream["status"] != "ACTIVE" {
		t.Fatalf("expected planning lane workstream, got %+v", workstreams)
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
	workstreams, ok := executionContract["workstreams"].([]any)
	if !ok || len(workstreams) != 3 {
		t.Fatalf("expected orchestration workstreams, got %+v", executionContract)
	}
	reviewLane, ok := workstreams[2].(map[string]any)
	if !ok || reviewLane["label"] != "Review lane" || reviewLane["owner_label"] != "Review lane lead" {
		t.Fatalf("expected review lane handoff, got %+v", workstreams)
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

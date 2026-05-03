package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func organizationCollaborationGroupColumns() []string {
	return []string{
		"id", "tenant_id", "name", "goal_statement", "work_mode",
		"allowed_capabilities", "member_user_ids", "team_ids",
		"coordinator_profile", "approval_policy_ref", "status", "created_by",
		"expiry", "created_audit_event_id", "updated_audit_event_id",
		"created_at", "updated_at",
	}
}

func organizationArtifactColumns() []string {
	return []string{
		"id", "mission_id", "team_id", "agent_id", "trace_id", "artifact_type",
		"title", "content_type", "content", "file_path", "file_size_bytes",
		"metadata", "trust_score", "status", "created_at",
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
	workstreams, ok := executionContract["workstreams"].([]any)
	if !ok || len(workstreams) != 3 {
		t.Fatalf("expected continuity workstreams, got %+v", executionContract)
	}
	completedLane, ok := workstreams[0].(map[string]any)
	if !ok || completedLane["label"] != "Completed work lane" || completedLane["status"] != "COMPLETE" {
		t.Fatalf("expected completed work lane, got %+v", workstreams)
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

func TestHandleTeamLeadGuidedAction_ResumeRetainedPackageUsesLatestGroupOutputs(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(
		dbOpt,
		withTemplateBundlesPath(writeStarterBundle(t)),
		func(s *AdminServer) {
			s.Artifacts = artifacts.NewService(s.DB, "")
		},
	)
	created := s.organizationStore().Save(s.buildOrganizationHome(OrganizationCreateRequest{
		Name:       "Northstar Labs",
		Purpose:    "Ship a focused AI engineering organization",
		StartMode:  OrganizationStartModeTemplate,
		TemplateID: "engineering-starter",
	}, mustResolveStarterTemplate(t, s, "engineering-starter")))

	groupUpdatedAt := time.Now().UTC()
	groupExpiry := groupUpdatedAt.Add(4 * time.Hour)
	teamID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WillReturnRows(sqlmock.NewRows(organizationCollaborationGroupColumns()).
			AddRow(
				"group-retained",
				"default",
				"Release Readiness Workflow",
				"Resume release readiness after reboot.",
				"resume_continuity",
				[]byte(`["artifact.review","team.coordinate"]`),
				[]byte(`["owner"]`),
				[]byte(`["11111111-1111-1111-1111-111111111111"]`),
				"release-workflow-coordinator",
				"",
				groupStatusArchived,
				"test-user-001",
				groupExpiry,
				"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				groupUpdatedAt.Add(-2*time.Hour),
				groupUpdatedAt,
			))

	mock.ExpectQuery("(?s)SELECT id, mission_id, team_id, agent_id, trace_id, artifact_type,.*FROM artifacts.*WHERE team_id = \\$1.*LIMIT \\$2").
		WithArgs(teamID, 8).
		WillReturnRows(sqlmock.NewRows(organizationArtifactColumns()).
			AddRow(
				"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				nil,
				teamID.String(),
				"review-lead",
				"",
				"document",
				"Review lane summary",
				"text/markdown",
				"review summary",
				"",
				nil,
				[]byte(`{}`),
				nil,
				"approved",
				groupUpdatedAt.Add(-15*time.Minute),
			).
			AddRow(
				"cccccccc-cccc-cccc-cccc-cccccccccccc",
				nil,
				teamID.String(),
				"validation-lead",
				"",
				"document",
				"Validation lane checklist",
				"text/markdown",
				"validation checklist",
				"",
				nil,
				[]byte(`{}`),
				nil,
				"approved",
				groupUpdatedAt.Add(-30*time.Minute),
			))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations/{id}/workspace/actions", s.handleTeamLeadGuidedAction)

	rr := doRequest(t, mux, "POST", "/api/v1/organizations/"+created.ID+"/workspace/actions", `{"action":"resume_retained_package","request_context":"Resume the retained package for Northstar Labs after a reboot or reload."}`)
	assertStatus(t, rr, http.StatusOK)

	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object action payload, got %T", resp.Data)
	}
	headline, _ := data["headline"].(string)
	if !strings.Contains(headline, "Release Readiness Workflow") {
		t.Fatalf("expected retained workflow headline, got %+v", data)
	}
	summary, _ := data["summary"].(string)
	if !strings.Contains(summary, "2 retained outputs") {
		t.Fatalf("expected retained output count in summary, got %+v", data)
	}

	executionContract, ok := data["execution_contract"].(map[string]any)
	if !ok {
		t.Fatalf("expected execution contract, got %+v", data)
	}
	if executionContract["owner_label"] != "Release Workflow Coordinator" {
		t.Fatalf("expected humanized coordinator owner, got %+v", executionContract)
	}
	if executionContract["continuity_label"] != "Release Readiness Workflow" {
		t.Fatalf("expected retained package label from state, got %+v", executionContract)
	}
	if executionContract["resume_checkpoint"] != "Open Release Readiness Workflow and continue with Review Lead after reviewing Review lane summary." {
		t.Fatalf("expected retained checkpoint from artifacts, got %+v", executionContract)
	}

	outputs, ok := executionContract["target_outputs"].([]any)
	if !ok || len(outputs) != 2 {
		t.Fatalf("expected retained output titles, got %+v", executionContract)
	}
	if outputs[0] != "Review lane summary" || outputs[1] != "Validation lane checklist" {
		t.Fatalf("unexpected retained output titles: %+v", outputs)
	}

	workstreams, ok := executionContract["workstreams"].([]any)
	if !ok || len(workstreams) != 3 {
		t.Fatalf("expected retained workstreams, got %+v", executionContract)
	}
	handoffLane, ok := workstreams[2].(map[string]any)
	if !ok || handoffLane["owner_label"] != "Review Lead" {
		t.Fatalf("expected next owner in handoff lane, got %+v", workstreams)
	}

	workflowGroup, ok := executionContract["workflow_group"].(map[string]any)
	if !ok {
		t.Fatalf("expected workflow group draft, got %+v", executionContract)
	}
	if workflowGroup["group_id"] != "group-retained" {
		t.Fatalf("expected retained group id, got %+v", workflowGroup)
	}
	if workflowGroup["name"] != "Release Readiness Workflow" {
		t.Fatalf("expected retained group name, got %+v", workflowGroup)
	}
	if workflowGroup["recommended_member_limit"] != float64(1) {
		t.Fatalf("expected member limit from retained team count, got %+v", workflowGroup)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

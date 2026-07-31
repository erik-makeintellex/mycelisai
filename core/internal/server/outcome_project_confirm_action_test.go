package server

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestEnsureOutcomeOwnershipForConfirmedActionCreatesProjectAndRegistry(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	link := testConfirmedActionTeamWorkLink(&protocol.ScopeValidation{
		AffectedResources: []string{"teams", "workspace"},
	})
	refs := []confirmActionTeamWorkRef{
		{
			WorkItemID: "44444444-4444-4444-4444-444444444444",
			TeamID:     "qa-team",
			State:      protocol.TeamWorkStateNew,
			RunID:      link.RunID,
		},
		{
			WorkItemID: "55555555-5555-5555-5555-555555555555",
			TeamID:     "qa-team",
			State:      protocol.TeamWorkStateOutputReady,
			RunID:      link.RunID,
			OutputRefs: []protocol.TeamOutputRef{{
				OutputID:   "output-1",
				TeamID:     "qa-team",
				WorkItemID: "55555555-5555-5555-5555-555555555555",
				RunID:      link.RunID,
				Kind:       "code",
				Label:      "Playable browser game",
				StorageRef: "groups/qa-team/generated/first-game",
				Entrypoint: "groups/qa-team/generated/first-game/index.html",
			}},
		},
	}

	mock.ExpectQuery("SELECT id::text, outcome_id").
		WithArgs(link.RunID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	expectOutcomeProjectInsert(mock, protocol.OutcomeProjectStatusOutputReady, "Playable browser game outcome", now)
	expectTeamRegistryEntryInsert(mock, "qa-team", "lead", now)

	project, err := s.ensureOutcomeOwnershipForConfirmedAction(t.Context(), link, refs)
	if err != nil {
		t.Fatalf("ensureOutcomeOwnershipForConfirmedAction: %v", err)
	}
	if project == nil {
		t.Fatalf("project is nil")
	}
	if project.Status != protocol.OutcomeProjectStatusOutputReady {
		t.Fatalf("status = %q, want output_ready", project.Status)
	}
	if project.WorkspaceFolder != "groups/qa-team/generated" {
		t.Fatalf("workspace_folder = %q", project.WorkspaceFolder)
	}
	if len(project.TeamRegistryRefs) != 1 {
		t.Fatalf("team registry refs = %#v", project.TeamRegistryRefs)
	}
	if len(project.OutputRefs) != 1 || project.OutputRefs[0].StorageRef != "groups/qa-team/generated/first-game" {
		t.Fatalf("output refs = %#v", project.OutputRefs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestEnsureOutcomeOwnershipForConfirmedActionReusesExistingOutcome(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	link := testConfirmedActionTeamWorkLink(&protocol.ScopeValidation{})
	projectID := "66666666-6666-4666-8666-666666666666"

	mock.ExpectQuery("SELECT id::text, outcome_id").
		WithArgs(link.RunID).
		WillReturnRows(sqlmock.NewRows(outcomeProjectScanColumns()).AddRow(
			projectID, link.RunID, "Existing outcome", "Retained work", "project", "groups/qa-team",
			"active", link.RunID, link.ProofID, link.ContractID, "", []byte(`["work-1"]`),
			[]byte(`[]`), []byte(`[]`), []byte(`[]`), "retained", now, now, "v1",
		))
	mock.ExpectQuery("SELECT id::text, project_id::text").
		WithArgs(projectID, 50).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	project, err := s.ensureOutcomeOwnershipForConfirmedAction(t.Context(), link, []confirmActionTeamWorkRef{{
		WorkItemID: "work-1", TeamID: "qa-team", State: protocol.TeamWorkStateQueued,
	}})
	if err != nil {
		t.Fatalf("ensureOutcomeOwnershipForConfirmedAction: %v", err)
	}
	if project == nil || project.ProjectID != projectID {
		t.Fatalf("project = %#v, want existing %s", project, projectID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func outcomeProjectScanColumns() []string {
	return []string{
		"id", "outcome_id", "title", "purpose", "execution_mode", "workspace_folder",
		"status", "run_id", "intent_proof_id", "contract_id", "proof_id", "work_item_refs",
		"output_refs", "proof_refs", "recovery_refs", "retention_policy", "created_at", "updated_at", "version",
	}
}

func expectOutcomeProjectInsert(mock sqlmock.Sqlmock, status protocol.OutcomeProjectStatus, title string, now time.Time) {
	mock.ExpectQuery("INSERT INTO outcome_projects").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), title, sqlmock.AnyArg(), "project",
			"groups/qa-team/generated", string(status), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"contract-1", "proof-artifact-1", jsonContainsArg("44444444-4444-4444-4444-444444444444"),
			jsonContainsArg("groups/qa-team/generated/first-game"), jsonContainsArg("contract-1"),
			sqlmock.AnyArg(), "retained", "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
}

func expectTeamRegistryEntryInsert(mock sqlmock.Sqlmock, teamID, role string, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_registry_entries").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), "", role, teamID, "",
			sqlmock.AnyArg(), true, nil, "active", "v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
}

func TestEnsureOutcomeOwnershipForConfirmedActionNoDBIsNoop(t *testing.T) {
	s := newTestServer()
	project, err := s.ensureOutcomeOwnershipForConfirmedAction(t.Context(), testConfirmedActionTeamWorkLink(&protocol.ScopeValidation{}), []confirmActionTeamWorkRef{{
		WorkItemID: "44444444-4444-4444-4444-444444444444",
		TeamID:     "qa-team",
		State:      protocol.TeamWorkStateQueued,
	}})
	if err != nil {
		t.Fatalf("ensureOutcomeOwnershipForConfirmedAction: %v", err)
	}
	if project != nil {
		t.Fatalf("project = %#v, want nil", project)
	}
}

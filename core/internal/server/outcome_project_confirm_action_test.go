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

package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleGroupOutputs_ProjectsTeamOutputRefs(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) {
		s.Artifacts = artifacts.NewService(s.DB, "")
	})
	mux := setupMux(t, "GET /api/v1/groups/{id}/outputs", s.HandleGroupOutputs)

	now := time.Now().UTC()
	teamID := "qa-delivery-team"
	workID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-output-ref").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).AddRow(
			"group-output-ref",
			"default",
			"Reference-only delivery",
			"Surface durable team output refs",
			"propose_only",
			[]byte(`["write_file"]`),
			[]byte(`[]`),
			[]byte(`["qa-delivery-team"]`),
			"groups/qa-delivery-team",
			"team lead",
			"confirmed-chat-proposal",
			groupStatusActive,
			"test-user-001",
			nil,
			"",
			"",
			now,
			now,
		))
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE agent_id = \\$1").
		WithArgs(teamID, 8).
		WillReturnRows(sqlmock.NewRows(artifactColumns()))
	outputRefs := []protocol.TeamOutputRef{{
		OutputID:    "playable-package",
		TeamID:      teamID,
		WorkItemID:  workID,
		Kind:        "project_package",
		OutputClass: protocol.OutputClassUserDeliverable,
		Label:       "Playable package",
		StorageRef:  "groups/qa-delivery-team/generated/first-game",
		Entrypoint:  "index.html",
		ProofRef:    "proof-playable-package",
		CreatedAt:   now.Add(-time.Minute),
	}}
	mock.ExpectQuery("FROM team_work_items").
		WithArgs(teamID, 8).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, teamID, "", "", "", "", "Create playable package", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`null`), []byte(`["project package"]`), []byte(`["proof"]`), []byte(`[]`),
			"confirmed", string(protocol.TeamWorkStateOutputReady), []byte(`null`), false, "",
			[]byte(`[]`), jsonArray(outputRefs), []byte(`["proof-playable-package"]`), []byte(`[]`), now.Add(-2*time.Minute), now, "v1",
		))

	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/groups/group-output-ref/outputs?limit=8", "")
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data := payload["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 projected output, got %#v", data)
	}
	first := data[0].(map[string]any)
	if first["title"] != "Playable package" || first["artifact_type"] != "project_package" {
		t.Fatalf("unexpected projected output: %#v", first)
	}
	if first["file_path"] != "groups/qa-delivery-team/generated/first-game" {
		t.Fatalf("file_path = %v", first["file_path"])
	}
	metadata := first["metadata"].(map[string]any)
	if metadata["folder"] != "groups/qa-delivery-team/generated/first-game" || metadata["entrypoint"] != "index.html" {
		t.Fatalf("package metadata = %#v", metadata)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectEmptyGroupTeamWorkOutputs(mock sqlmock.Sqlmock, teamID string, limit int) {
	mock.ExpectQuery("FROM team_work_items").
		WithArgs(teamID, limit).
		WillReturnRows(teamWorkItemRows())
}

package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
)

func TestHandleGroupOutputs_FiltersPlanningArtifactsByDefault(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) {
		s.Artifacts = artifacts.NewService(s.DB, "")
	})
	mux := setupMux(t, "GET /api/v1/groups/{id}/outputs", s.HandleGroupOutputs)

	now := time.Now().UTC()
	teamID := uuid.New()
	expectPlannedGroupOutputQuery(mock, teamID, now)
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE team_id = \\$1").
		WithArgs(teamID, 8).
		WillReturnRows(sqlmock.NewRows(artifactColumns()).
			AddRow(
				uuid.New(), nil, &teamID, "mixed-output-team", nil, "file",
				"groups/mixed-output-team/planning/TEAM_EVOCATION.md",
				"text/markdown", "# Team evocation",
				"groups/mixed-output-team/planning/TEAM_EVOCATION.md",
				nil, []byte(`{"output_class":"planning"}`), nil,
				"approved", now.Add(-time.Minute),
			).
			AddRow(
				uuid.New(), nil, &teamID, "mixed-output-team", nil,
				"project_package", "Playable package", "text/html", "",
				"groups/mixed-output-team/generated/first-game/index.html",
				nil, []byte(`{"output_class":"user_deliverable"}`), nil,
				"approved", now,
			))
	expectEmptyGroupTeamWorkOutputs(mock, teamID.String(), 8)

	rr := doAuthenticatedRequest(
		t,
		mux,
		"GET",
		"/api/v1/groups/group-planned/outputs?limit=8",
		"",
	)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data := payload["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected only user deliverable output, got %#v", data)
	}
	if first := data[0].(map[string]any); first["title"] != "Playable package" {
		t.Fatalf("unexpected output title: %#v", first["title"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleGroupOutputs_IncludeInternalReturnsPlanningArtifacts(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) {
		s.Artifacts = artifacts.NewService(s.DB, "")
	})
	mux := setupMux(t, "GET /api/v1/groups/{id}/outputs", s.HandleGroupOutputs)

	now := time.Now().UTC()
	teamID := uuid.New()
	expectPlannedGroupOutputQuery(mock, teamID, now)
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE team_id = \\$1").
		WithArgs(teamID, 8).
		WillReturnRows(sqlmock.NewRows(artifactColumns()).AddRow(
			uuid.New(), nil, &teamID, "mixed-output-team", nil, "file",
			"groups/mixed-output-team/planning/TEAM_EVOCATION.md",
			"text/markdown", "# Team evocation",
			"groups/mixed-output-team/planning/TEAM_EVOCATION.md",
			nil, []byte(`{"output_class":"planning"}`), nil, "approved", now,
		))
	expectEmptyGroupTeamWorkOutputs(mock, teamID.String(), 8)

	rr := doAuthenticatedRequest(
		t,
		mux,
		"GET",
		"/api/v1/groups/group-planned/outputs?limit=8&include_internal=true",
		"",
	)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data := payload["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected planning output with include_internal, got %#v", data)
	}
	if first := data[0].(map[string]any); first["title"] !=
		"groups/mixed-output-team/planning/TEAM_EVOCATION.md" {
		t.Fatalf("unexpected output title: %#v", first["title"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectPlannedGroupOutputQuery(
	mock sqlmock.Sqlmock,
	teamID uuid.UUID,
	now time.Time,
) {
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-planned").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).AddRow(
			"group-planned", "default", "Mixed Output Team",
			"Prepare and build a mixed output", "propose_only",
			[]byte(`["write_file"]`), []byte(`[]`),
			[]byte(fmt.Sprintf(`["%s"]`, teamID.String())),
			"groups/mixed-output-team", "team lead", "confirmed-chat-proposal",
			groupStatusActive, "test-user-001", nil, "", "", now, now,
		))
}

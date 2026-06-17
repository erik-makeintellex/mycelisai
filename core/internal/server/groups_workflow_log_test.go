package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleGroupWorkflowLog_ReturnsNormalizedTimeline(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) {
		s.Artifacts = artifacts.NewService(s.DB, "")
		s.GroupBus = NewGroupBusMonitor()
		s.GroupBus.RecordSuccess("group-game", "admin", "Please finish the playable browser build.", []string{"swarm.group.group-game.collab"})
	})
	mux := setupMux(t, "GET /api/v1/groups/{id}/workflow-log", s.HandleGroupWorkflowLog)

	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-2222-2222-222222222222"
	artifactID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	outputRefs := fmt.Sprintf(`[{
		"output_id":"game-output",
		"team_id":"game-team",
		"work_item_id":"%s",
		"run_id":"%s",
		"kind":"project_package",
		"label":"Playable browser game",
		"storage_ref":"groups/game-team/generated/first-game",
		"entrypoint":"groups/game-team/generated/first-game/index.html",
		"proof_ref":"proof-artifact-1",
		"audit_refs":["audit-work"],
		"created_at":"%s"
	}]`, workID, runID, now.Add(-time.Minute).Format(time.RFC3339Nano))

	expectWorkflowLogGroup(mock, now)
	mock.ExpectQuery("FROM team_work_items").
		WithArgs("game-team", 20).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, "game-team", runID, "", "contract-1", "proof-artifact-1", "Build a playable browser game", []byte(`["browser app"]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), []byte(`["index.html"]`), []byte(`["manual play proof"]`), []byte(`["workspace.write"]`),
			"confirmed-chat-proposal", string(protocol.TeamWorkStateOutputReady),
			[]byte(`{"headline":"Playable game package is ready","confidence_posture":"verified","source_kind":"internal_tool","source_channel":"soma.team_work","payload_kind":"team_status"}`),
			false, "", []byte(`[]`), []byte(outputRefs), []byte(`["proof-artifact-1"]`), []byte(`["audit-work"]`), now.Add(-2*time.Minute), now.Add(-time.Minute), "v1",
		))
	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE agent_id = \\$1").
		WithArgs("game-team", 20).
		WillReturnRows(sqlmock.NewRows(artifactColumns()).
			AddRow(
				artifactID,
				nil,
				nil,
				"game-team",
				nil,
				"project_package",
				"First Game Package",
				"text/html",
				"",
				"groups/game-team/generated/first-game/index.html",
				nil,
				[]byte(`{"audit_refs":["audit-artifact"]}`),
				nil,
				"approved",
				now,
			))

	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/groups/group-game/workflow-log?limit=20&include_outputs=true&include_audit=true", "")
	assertStatus(t, rr, http.StatusOK)

	var payload struct {
		Data groupWorkflowLogResponse `json:"data"`
	}
	assertJSON(t, rr, &payload)
	if payload.Data.Group.ID != "group-game" {
		t.Fatalf("group id = %q", payload.Data.Group.ID)
	}
	if payload.Data.Lifecycle.Recommendation != "archive_completed" {
		t.Fatalf("lifecycle recommendation = %q", payload.Data.Lifecycle.Recommendation)
	}
	if len(payload.Data.TeamWork) != 1 {
		t.Fatalf("team work count = %d", len(payload.Data.TeamWork))
	}
	if len(payload.Data.Outputs) != 1 {
		t.Fatalf("output count = %d", len(payload.Data.Outputs))
	}
	assertWorkflowLogHasKind(t, payload.Data.Timeline, "team_work")
	assertWorkflowLogHasKind(t, payload.Data.Timeline, "team_output_ref")
	assertWorkflowLogHasKind(t, payload.Data.Timeline, "retained_artifact")
	assertWorkflowLogHasKind(t, payload.Data.Timeline, "proof_cue")
	assertWorkflowLogHasKind(t, payload.Data.Timeline, "broadcast")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleGroupWorkflowLog_DegradesWhenOutputsUnavailable(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "GET /api/v1/groups/{id}/workflow-log", s.HandleGroupWorkflowLog)

	now := time.Now().UTC()
	expectWorkflowLogGroup(mock, now)
	mock.ExpectQuery("FROM team_work_items").
		WithArgs("game-team", 10).
		WillReturnRows(teamWorkItemRows())

	rr := doAuthenticatedRequest(t, mux, http.MethodGet, "/api/v1/groups/group-game/workflow-log?limit=10&include_outputs=true", "")
	assertStatus(t, rr, http.StatusOK)

	var payload struct {
		Data groupWorkflowLogResponse `json:"data"`
	}
	assertJSON(t, rr, &payload)
	if len(payload.Data.Degraded) != 1 {
		t.Fatalf("degraded cues = %d, want 1", len(payload.Data.Degraded))
	}
	if payload.Data.Degraded[0].Kind != "outputs_unavailable" {
		t.Fatalf("degraded kind = %q", payload.Data.Degraded[0].Kind)
	}
	assertWorkflowLogHasKind(t, payload.Data.Timeline, "degraded")
	assertWorkflowLogHasKind(t, payload.Data.Timeline, "group_brief")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func expectWorkflowLogGroup(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-game").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-game",
				"default",
				"Game Team",
				"Create a playable browser game package",
				"propose_only",
				[]byte(`["workspace.write"]`),
				[]byte(`["owner"]`),
				[]byte(`["game-team"]`),
				"groups/game-team",
				"game-lead",
				"confirmed-chat-proposal",
				groupStatusActive,
				"test-user-001",
				nil,
				"audit-created",
				"audit-updated",
				now.Add(-time.Hour),
				now.Add(-30*time.Minute),
			))
}

func assertWorkflowLogHasKind(t *testing.T, entries []groupWorkflowLogEntry, kind string) {
	t.Helper()
	for _, entry := range entries {
		if entry.Kind == kind {
			return
		}
	}
	t.Fatalf("workflow log missing entry kind %q in %#v", kind, entries)
}

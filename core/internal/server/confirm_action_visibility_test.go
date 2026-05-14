package server

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestEnsureGroupForCreatedTeamMirrorsConfirmedCreateTeam(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(true)

	auditID := "44444444-4444-4444-4444-444444444444"
	mock.ExpectQuery("FROM collaboration_groups").
		WithArgs("Research Team").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("INSERT INTO collaboration_groups").
		WithArgs(
			sqlmock.AnyArg(), "default", "Research Team",
			"Map optimal agentry architecture.",
			"propose_only",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"researcher lead", "confirmed-chat-proposal", groupStatusActive,
			"test-user", nil, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(time.Now(), time.Now()))

	err := s.ensureGroupForCreatedTeam(t.Context(), auditID, "test-user", map[string]any{
		"team_id": "research-team",
		"name":    "Research Team",
		"role":    "researcher",
		"goal":    "Map optimal agentry architecture.",
	})
	if err != nil {
		t.Fatalf("ensure group for created team: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

func TestExecutionOutputsFromToolResultsRetainsTeamAndCodeFile(t *testing.T) {
	outputs := executionOutputsFromToolResults([]plannedToolExecutionResult{
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id": "qa-browser-game-team",
				"name":    "QA Browser Game Team",
			},
			Output: "team created",
		},
		{
			Name: "write_file",
			Arguments: map[string]any{
				"path": "workspace/logs/qa_team_game.html",
			},
			Output: "wrote game file",
		},
	})

	if len(outputs) != 2 {
		t.Fatalf("outputs = %#v, want 2", outputs)
	}
	if outputs[0].Kind != "team" || outputs[0].ID != "qa-browser-game-team" || outputs[0].Href != "/groups" {
		t.Fatalf("team output = %#v", outputs[0])
	}
	if outputs[0].Retained == nil || !*outputs[0].Retained {
		t.Fatalf("team retained = %#v", outputs[0].Retained)
	}
	if outputs[1].Kind != "code" || outputs[1].ID != "workspace/logs/qa_team_game.html" {
		t.Fatalf("file output = %#v", outputs[1])
	}
	if outputs[1].Title != "workspace/logs/qa_team_game.html" {
		t.Fatalf("file title = %q", outputs[1].Title)
	}
	if outputs[1].Href != "/api/v1/workspace/files/view?path=workspace%2Flogs%2Fqa_team_game.html" {
		t.Fatalf("file href = %q", outputs[1].Href)
	}
	if outputs[1].Retained == nil || !*outputs[1].Retained {
		t.Fatalf("file retained = %#v", outputs[1].Retained)
	}
	if outputs[1].RetentionClass != protocol.ExecutionRetentionClassRetained {
		t.Fatalf("file retention = %q", outputs[1].RetentionClass)
	}
}

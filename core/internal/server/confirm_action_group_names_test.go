package server

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEnsureGroupForCreatedTeamKeepsNewTeamLaneWhenNameRepeats(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mock.MatchExpectationsInOrder(true)

	now := time.Now()
	auditID := "44444444-4444-4444-4444-444444444444"
	mock.ExpectQuery("FROM collaboration_groups").
		WithArgs("First Demo Game Team").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-first-demo", "default", "First Demo Game Team",
				"Prior first demo team.",
				"propose_only",
				`["team.coordinate","artifact.review","broadcast"]`,
				`[]`,
				`["first-demo-game-team-old"]`,
				"groups/first-demo-game-team-old",
				"worker lead", "confirmed-chat-proposal", groupStatusActive,
				"admin", nil, auditID, auditID, now, now,
			))
	mock.ExpectQuery("FROM collaboration_groups").
		WithArgs("First Demo Game Team - first-demo-game-team-new").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("INSERT INTO collaboration_groups").
		WithArgs(
			sqlmock.AnyArg(), "default", "First Demo Game Team - first-demo-game-team-new",
			"Runtime team First Demo Game Team - first-demo-game-team-new created through confirmed Soma proposal.",
			"propose_only",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "groups/first-demo-game-team-new",
			"worker lead", "confirmed-chat-proposal", groupStatusActive,
			"test-user", nil, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(time.Now(), time.Now()))

	err := s.ensureGroupForCreatedTeam(t.Context(), auditID, "test-user", map[string]any{
		"team_id": "first-demo-game-team-new",
		"name":    "First Demo Game Team",
		"role":    "worker",
	})
	if err != nil {
		t.Fatalf("ensure group for repeat team name: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet db expectations: %v", err)
	}
}

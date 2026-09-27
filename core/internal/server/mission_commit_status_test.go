package server

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// T1 truth follow-up: persistMissionBlueprint always inserted the missions
// row with the literal status "active", even when Soma was nil and nothing
// activated. commitActivationStatus computed the honest value for the JSON
// response, but the DB row never reflected it. This proves the stored row
// itself now carries the real post-activation status (PH-D follow-up).
func TestIntentCommit_MissionRowPersistsRealActivationStatus(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt) // Soma nil: nothing can activate
	bp := tierBlueprint(1, "read_file")

	expectCommitToken(t, mock, buildScopeFromBlueprint(bp), bp, "u-std")
	mock.ExpectExec("UPDATE confirm_tokens SET consumed = TRUE").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO missions").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO teams").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO service_manifests").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	// The row inserted above must be corrected to the real status once
	// activation is known, not left as the "active" placeholder.
	mock.ExpectExec("UPDATE missions SET status").
		WithArgs(commitStatusPersistedNotActivated, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE intent_proofs SET status = 'confirmed'").WillReturnResult(sqlmock.NewResult(0, 1))
	a := sqlmock.AnyArg()
	mock.ExpectExec("INSERT INTO log_entries").WithArgs(a, a, a, a, a, a, a, a).WillReturnResult(sqlmock.NewResult(0, 1))

	rr := commitAs(t, s, bp, standardUserIdentity())
	assertStatus(t, rr, http.StatusOK)

	var resp CommitResponse
	assertJSON(t, rr, &resp)
	if resp.Status != commitStatusPersistedNotActivated {
		t.Fatalf("response status = %q, want %q", resp.Status, commitStatusPersistedNotActivated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("missions row must be updated with the real activation status: %v", err)
	}
}

// buildTeamMissionLookup must not drop teams belonging to a partially
// activated mission: some of its teams are genuinely running, so team-lookup
// consumers (e.g. teams_detail) must still resolve their mission context.
func TestBuildTeamMissionLookup_IncludesPartiallyActiveMissions(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)

	mock.ExpectQuery(`(?s)SELECT t\.name, m\.id, m\.directive.*WHERE m\.status IN \('active', 'partially_active'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"name", "id", "directive"}).
			AddRow("alpha-team", "m-1", "Ship it"))

	lookup := s.buildTeamMissionLookup()
	if info, ok := lookup["alpha-team"]; !ok || info.missionID != "m-1" {
		t.Fatalf("partially_active mission's team must resolve: %+v", lookup)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet DB expectations: %v", err)
	}
}

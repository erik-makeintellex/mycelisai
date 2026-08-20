package server

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func doTeamWorkAction(t *testing.T, s *AdminServer, workID, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := setupMux(t, "POST /api/v1/teams/{id}/work/{workItemId}/actions", s.HandleTeamWorkAction)
	return doRequest(t, mux, http.MethodPost, "/api/v1/teams/research-team/work/"+workID+"/actions", body)
}

func mockTeamWorkItem(mock sqlmock.Sqlmock, teamID, workID string, state protocol.TeamWorkState, needsOperator bool, degradation string, now time.Time) {
	mock.ExpectQuery("SELECT id::text, team_id").
		WithArgs(teamID, workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, teamID, "", "", "", "", "Draft release proof", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`null`), []byte(`["release proof"]`), []byte(`["run proof"]`), []byte(`[]`),
			"auto_approved", string(state), []byte(`null`), needsOperator, degradation,
			[]byte(`["retry"]`), []byte(`[]`), []byte(`["proof-1"]`), []byte(`["audit-1"]`), now, now, "v1",
		))
}

func expectTeamWorkActionPersistence(mock sqlmock.Sqlmock, now time.Time) {
	expectTeamWorkActionWrites(mock, now)
	mock.ExpectCommit()
}

func expectTeamWorkActionWrites(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("UPDATE team_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_interactions").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func expectExternalOutcomeVerificationPersistence(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("UPDATE team_work_items").
		WithArgs(
			sqlmock.AnyArg(), string(protocol.TeamWorkStateOutputReady), sqlmock.AnyArg(), false, "",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), externalVerificationWorkIntentMatch{},
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_interactions").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectCommit()
}

type externalVerificationWorkIntentMatch struct{}

func (externalVerificationWorkIntentMatch) Match(value driver.Value) bool {
	var raw []byte
	switch typed := value.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		return false
	}
	var intent protocol.WorkIntent
	if err := json.Unmarshal(raw, &intent); err != nil || intent.SideEffect == nil || intent.SideEffect.Verification == nil {
		return false
	}
	return intent.SideEffect.SideEffectState == protocol.WorkSideEffectCommitted &&
		intent.SideEffect.Verification.Result == protocol.WorkExternalOutcomeCommitted &&
		intent.SideEffect.Verification.ActorRef == "local-user" &&
		intent.SideEffect.Verification.Summary == "The customer record contains the requested update."
}

func teamWorkItemRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "team_id", "run_id", "intent_proof_id", "contract_id", "proof_id",
		"objective", "scope", "owner", "execution_shape", "execution_mode", "work_intent", "expected_outputs", "expected_proof",
		"capability_requirements", "governance_posture", "state", "last_event", "needs_operator",
		"degradation_state", "recovery_options", "output_refs", "proof_refs", "audit_refs",
		"created_at", "updated_at", "version",
	})
}

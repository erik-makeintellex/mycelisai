package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleTeamWorkAction_PauseRecordsLifecycle(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectTeamWorkActionPersistence(mock, now)

	rr := doTeamWorkAction(t, s, workID, `{
		"action":"pause",
		"summary":"Operator paused this while checking acceptance proof."
	}`)

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["state"] != string(protocol.TeamWorkStatePaused) {
		t.Fatalf("state = %v", data["state"])
	}
	lastEvent := data["last_event"].(map[string]any)
	if lastEvent["headline"] != "Team work paused" {
		t.Fatalf("headline = %v", lastEvent["headline"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAction_RejectsInvalidTransition(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateOutputReady, false, "", now)

	rr := doTeamWorkAction(t, s, workID, `{"action":"start_work"}`)

	assertStatus(t, rr, http.StatusBadRequest)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAction_RejectsSteerWithoutGuidance(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	workID := "11111111-1111-1111-1111-111111111111"

	rr := doTeamWorkAction(t, s, workID, `{"action":"steer"}`)

	assertStatus(t, rr, http.StatusBadRequest)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAction_SteerPreservesState(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectTeamWorkActionPersistence(mock, now)

	rr := doTeamWorkAction(t, s, workID, `{
		"action":"steer",
		"summary":"Focus the proof review on deployment readiness."
	}`)

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["state"] != string(protocol.TeamWorkStateRunning) {
		t.Fatalf("state = %v", data["state"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAction_RecoverQueuesDegradedWork(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateDegraded, true, "provider_timeout", now)
	expectTeamWorkActionPersistence(mock, now)

	rr := doTeamWorkAction(t, s, workID, `{
		"action":"recover",
		"summary":"Retry with the retained proof package as context."
	}`)

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["state"] != string(protocol.TeamWorkStateQueued) {
		t.Fatalf("state = %v", data["state"])
	}
	if data["needs_operator"] != false {
		t.Fatalf("needs_operator = %v", data["needs_operator"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

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
			string(protocol.TeamExecutionShapeDeliverable), []byte(`["release proof"]`), []byte(`["run proof"]`), []byte(`[]`),
			"auto_approved", string(state), []byte(`null`), needsOperator, degradation,
			[]byte(`["retry"]`), []byte(`[]`), []byte(`["proof-1"]`), []byte(`["audit-1"]`), now, now, "v1",
		))
}

func expectTeamWorkActionPersistence(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery("INSERT INTO team_status_events").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("UPDATE team_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_interactions").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
}

func teamWorkItemRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "team_id", "run_id", "intent_proof_id", "contract_id", "proof_id",
		"objective", "scope", "owner", "execution_shape", "expected_outputs", "expected_proof",
		"capability_requirements", "governance_posture", "state", "last_event", "needs_operator",
		"degradation_state", "recovery_options", "output_refs", "proof_refs", "audit_refs",
		"created_at", "updated_at", "version",
	})
}

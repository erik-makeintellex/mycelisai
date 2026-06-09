package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleCreateTeamWork_DefaultsCreateTeamToNew(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO team_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	mux := setupMux(t, "POST /api/v1/teams/{id}/work", s.HandleCreateTeamWork)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/research-team/work", `{
		"execution_shape":"create_team",
		"objective":"Create the research team shell",
		"expected_outputs":["team roster"],
		"expected_proof":["brief recorded"]
	}`)

	assertStatus(t, rr, http.StatusCreated)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["team_id"] != "research-team" {
		t.Fatalf("team_id = %v", data["team_id"])
	}
	if data["state"] != "new" {
		t.Fatalf("state = %v, want new", data["state"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleCreateTeamWork_RejectsRunningCreateTeam(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	mux := setupMux(t, "POST /api/v1/teams/{id}/work", s.HandleCreateTeamWork)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/research-team/work", `{
		"execution_shape":"create_team",
		"state":"running",
		"objective":"Create the research team shell"
	}`)

	assertStatus(t, rr, http.StatusBadRequest)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleListTeamWork_ExcludesArchivedWhenRequested(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("state <> 'archived'").
		WithArgs("research-team", 10).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, "research-team", "", "", "", "", "Review failed proof", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDelegatedWork), []byte(`["review"]`), []byte(`["proof"]`), []byte(`[]`),
			"auto_approved", string(protocol.TeamWorkStateDegraded), []byte(`null`), true, "missing_execution_plan",
			[]byte(`["archive stale item"]`), []byte(`[]`), []byte(`["proof-1"]`), []byte(`["audit-1"]`), now, now, "v1",
		))

	mux := setupMux(t, "GET /api/v1/teams/{id}/work", s.HandleListTeamWork)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/teams/research-team/work?limit=10&include_archived=false", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	items := resp["data"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 work item, got %d", len(items))
	}
	first := items[0].(map[string]any)
	if first["state"] != string(protocol.TeamWorkStateDegraded) {
		t.Fatalf("state = %v", first["state"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleCreateTeamInteraction_PersistsDurableSourceContract(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	mock.ExpectQuery("INSERT INTO team_interactions").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))

	workID := "11111111-1111-1111-1111-111111111111"
	mux := setupMux(t, "POST /api/v1/teams/{id}/work/{workItemId}/interactions", s.HandleCreateTeamInteraction)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/teams/research-team/work/"+workID+"/interactions", `{
		"source_kind":"workspace_ui",
		"source_channel":"soma.team_work",
		"actor_ref":"soma",
		"verb":"brief",
		"summary":"Briefed the team on the deliverable",
		"payload_kind":"team_brief",
		"audit_refs":["audit-1"]
	}`)

	assertStatus(t, rr, http.StatusCreated)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["work_item_id"] != workID {
		t.Fatalf("work_item_id = %v", data["work_item_id"])
	}
	if data["source_kind"] != "workspace_ui" {
		t.Fatalf("source_kind = %v", data["source_kind"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleListTeamStatusEvents_ReturnsTimeline(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT id::text, team_id, work_item_id::text").
		WithArgs("research-team", workID, 5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "team_id", "work_item_id", "run_id", "intent_proof_id", "contract_id", "proof_id",
			"state", "headline", "details", "confidence_posture", "blocked_by", "next_action",
			"source_kind", "source_channel", "payload_kind", "audit_refs", "timestamp", "version",
		}).AddRow(
			"22222222-2222-2222-2222-222222222222",
			"research-team",
			workID,
			"",
			"",
			"contract-1",
			"proof-1",
			string(protocol.TeamWorkStateRunning),
			"Work running",
			"Team accepted the delegated work.",
			"verified",
			[]byte(`["waiting_on_review"]`),
			"Watch for output",
			"workspace_ui",
			"soma.team_work",
			"team_status",
			[]byte(`["audit-1"]`),
			now,
			"v1",
		))

	mux := setupMux(t, "GET /api/v1/teams/{id}/work/{workItemId}/status-events", s.HandleListTeamStatusEvents)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/teams/research-team/work/"+workID+"/status-events?limit=5", "")

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	items := resp["data"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 status event, got %d", len(items))
	}
	first := items[0].(map[string]any)
	if first["headline"] != "Work running" {
		t.Fatalf("headline = %v", first["headline"])
	}
	if first["state"] != string(protocol.TeamWorkStateRunning) {
		t.Fatalf("state = %v", first["state"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

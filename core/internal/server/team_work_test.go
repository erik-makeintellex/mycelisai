package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

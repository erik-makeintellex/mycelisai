package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleTeamWorkAction_PauseRecordsLifecycle(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectBegin()
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
	mock.ExpectBegin()
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateOutputReady, false, "", now)
	mock.ExpectRollback()

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
	s.DispatchOutbox = dispatchoutbox.NewStore(s.getDB())
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	runID := "22222222-2222-4222-8222-222222222222"
	proofID := "33333333-3333-4333-8333-333333333333"
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, team_id").WithArgs("research-team", workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, "research-team", runID, proofID, "44444444-4444-4444-8444-444444444444", "", "Draft release proof", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`null`), []byte(`["release proof"]`), []byte(`["run proof"]`), []byte(`[]`),
			"auto_approved", string(protocol.TeamWorkStateRunning), []byte(`null`), false, "",
			[]byte(`["retry"]`), []byte(`[]`), []byte(`["proof-1"]`), []byte(`["audit-1"]`), now, now, "v1",
		))
	mock.ExpectQuery("INSERT INTO team_status_events").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("INSERT INTO mission_events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE team_work_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO team_interactions").
		WillReturnRows(sqlmock.NewRows([]string{"timestamp"}).AddRow(now))
	mock.ExpectExec("INSERT INTO execution_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("UPDATE execution_dispatch_outbox").WillReturnResult(sqlmock.NewResult(0, 1))

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

func TestHandleTeamWorkAction_LegacySteerRemainsRecordable(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectBegin()
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateRunning, false, "", now)
	expectTeamWorkActionPersistence(mock, now)

	rr := doTeamWorkAction(t, s, workID, `{
		"action":"steer",
		"summary":"Keep the legacy work record focused on deployment proof."
	}`)

	assertStatus(t, rr, http.StatusOK)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAction_RecoverQueuesDegradedWork(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectBegin()
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

func TestHandleTeamWorkAction_RejectsUnknownExternalMutationRecovery(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, team_id").
		WithArgs("research-team", workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, "research-team", "", "", "", "", "Update the customer system", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`{"side_effect":{"effect_kind":"external_mutation","retry_safety":"safe","idempotency_key":"customer-update-1","side_effect_state":"unknown"}}`), []byte(`["customer update"]`), []byte(`["external verification"]`), []byte(`[]`),
			"required", string(protocol.TeamWorkStateDegraded), []byte(`null`), true, "external_mutation_outcome_unknown",
			[]byte(`["Ask Soma to verify the external system before deciding whether any retry is safe."]`), []byte(`[]`), []byte(`["proof-1"]`), []byte(`["audit-1"]`), now, now, "v1",
		))
	mock.ExpectRollback()

	rr := doTeamWorkAction(t, s, workID, `{
		"action":"recover",
		"summary":"Retry without verification."
	}`)

	assertStatus(t, rr, http.StatusBadRequest)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAction_VerifiesUnknownExternalOutcomeFromPayload(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id::text, team_id").
		WithArgs("research-team", workID).
		WillReturnRows(teamWorkItemRows().AddRow(
			workID, "research-team", "", "", "", "", "Update the customer system", []byte(`[]`), "Soma",
			string(protocol.TeamExecutionShapeDeliverable), "", []byte(`{"side_effect":{"effect_kind":"external_mutation","retry_safety":"safe","idempotency_key":"customer-update-1","side_effect_state":"unknown"}}`), []byte(`["customer update"]`), []byte(`["external verification"]`), []byte(`[]`),
			"required", string(protocol.TeamWorkStateDegraded), []byte(`null`), true, protocol.TeamWorkDegradationExternalMutationUnknown,
			[]byte(`["Verify the external result."]`), []byte(`[]`), []byte(`["proof-1"]`), []byte(`["audit-1"]`), now, now, "v1",
		))
	expectExternalOutcomeVerificationPersistence(mock, now)

	rr := doTeamWorkAction(t, s, workID, `{
		"action":"verify_external_outcome",
		"summary":"The customer record contains the requested update.",
		"source_kind":"system",
		"source_channel":"spoofed.channel",
		"payload_kind":"telemetry",
		"audit_refs":["spoofed-audit"],
		"payload":{"result":"committed","evidence_refs":[]}
	}`)

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["state"] != string(protocol.TeamWorkStateOutputReady) || data["needs_operator"] != false {
		t.Fatalf("state/operator = %v/%v", data["state"], data["needs_operator"])
	}
	workIntent := data["work_intent"].(map[string]any)
	sideEffect := workIntent["side_effect"].(map[string]any)
	if sideEffect["side_effect_state"] != protocol.WorkSideEffectCommitted {
		t.Fatalf("side_effect_state = %v", sideEffect["side_effect_state"])
	}
	verification := sideEffect["verification"].(map[string]any)
	if verification["result"] != protocol.WorkExternalOutcomeCommitted || verification["actor_ref"] != "local-user" {
		t.Fatalf("verification = %#v", verification)
	}
	lastEvent := data["last_event"].(map[string]any)
	if lastEvent["source_kind"] != string(protocol.SourceKindWorkspaceUI) ||
		lastEvent["source_channel"] != "teams.external_outcome_verification" ||
		lastEvent["payload_kind"] != "external_outcome_verification" {
		t.Fatalf("last_event provenance = %#v", lastEvent)
	}
	for _, auditRef := range lastEvent["audit_refs"].([]any) {
		if auditRef == "spoofed-audit" {
			t.Fatalf("client-controlled audit ref persisted: %#v", lastEvent["audit_refs"])
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleTeamWorkAction_ArchiveClearsReviewQueue(t *testing.T) {
	opt, mock := withDB(t)
	s := newTestServer(opt)
	now := time.Now().UTC()
	workID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectBegin()
	mockTeamWorkItem(mock, "research-team", workID, protocol.TeamWorkStateDegraded, true, "missing_execution_plan", now)
	expectTeamWorkActionPersistence(mock, now)

	rr := doTeamWorkAction(t, s, workID, `{"action":"archive"}`)

	assertStatus(t, rr, http.StatusOK)
	var resp map[string]any
	assertJSON(t, rr, &resp)
	data := resp["data"].(map[string]any)
	if data["state"] != string(protocol.TeamWorkStateArchived) {
		t.Fatalf("state = %v", data["state"])
	}
	if data["needs_operator"] != false {
		t.Fatalf("needs_operator = %v", data["needs_operator"])
	}
	lastEvent := data["last_event"].(map[string]any)
	if lastEvent["headline"] != "Team work archived" {
		t.Fatalf("headline = %v", lastEvent["headline"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func collaborationGroupColumns() []string {
	return []string{
		"id", "tenant_id", "name", "goal_statement", "work_mode",
		"allowed_capabilities", "member_user_ids", "team_ids",
		"coordinator_profile", "approval_policy_ref", "status", "created_by",
		"expiry", "created_audit_event_id", "updated_audit_event_id",
		"created_at", "updated_at",
	}
}

func artifactColumns() []string {
	return []string{
		"id", "mission_id", "team_id", "agent_id", "trace_id", "artifact_type",
		"title", "content_type", "content", "file_path", "file_size_bytes",
		"metadata", "trust_score", "status", "created_at",
	}
}

func TestHandleCreateGroup_Unauthorized(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/groups", s.HandleCreateGroup)

	body := `{"name":"ops","goal_statement":"run ops","work_mode":"propose_only"}`
	rr := doRequest(t, mux, "POST", "/api/v1/groups", body)
	assertStatus(t, rr, http.StatusUnauthorized)
}

func TestHandleCreateGroup_ScopeDenied(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/groups", s.HandleCreateGroup)

	body := `{"name":"ops","goal_statement":"run ops","work_mode":"propose_only"}`
	identity := &RequestIdentity{
		UserID:   "test-user-001",
		Username: "admin",
		Role:     "admin",
		Scopes:   []string{"groups:read"},
	}
	rr := doAuthenticatedRequestAs(t, mux, "POST", "/api/v1/groups", body, identity)
	assertStatus(t, rr, http.StatusForbidden)
}

func TestHandleCreateAndListGroups_HappyPath_DB(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	createMux := setupMux(t, "POST /api/v1/groups", s.HandleCreateGroup)
	listMux := setupMux(t, "GET /api/v1/groups", s.HandleListGroups)

	now := time.Now().UTC()

	mock.ExpectExec("INSERT INTO log_entries").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO collaboration_groups").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	body := `{
		"name":"ops-alpha",
		"goal_statement":"coordinate daily reliability runs",
		"work_mode":"propose_only",
		"allowed_capabilities":["runs.read","runs.propose"],
		"member_user_ids":["u1","u2"],
		"team_ids":["admin-core","council-core"],
		"coordinator_profile":"ops-profile",
		"approval_policy_ref":"policy.ops"
	}`

	createRR := doAuthenticatedRequest(t, createMux, "POST", "/api/v1/groups", body)
	assertStatus(t, createRR, http.StatusCreated)

	auditID := "11111111-1111-1111-1111-111111111111"
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				"default",
				"ops-alpha",
				"coordinate daily reliability runs",
				"propose_only",
				[]byte(`["runs.read","runs.propose"]`),
				[]byte(`["u1","u2"]`),
				[]byte(`["admin-core","council-core"]`),
				"ops-profile",
				"policy.ops",
				"active",
				"test-user-001",
				nil,
				auditID,
				auditID,
				now,
				now,
			))

	listRR := doAuthenticatedRequest(t, listMux, "GET", "/api/v1/groups", "")
	assertStatus(t, listRR, http.StatusOK)

	var listed map[string]any
	assertJSON(t, listRR, &listed)
	data, ok := listed["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", listed["data"])
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 group, got %d", len(data))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleCreateGroup_HighImpact_RequiresApproval(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "POST /api/v1/groups", s.HandleCreateGroup)

	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO intent_proofs").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO confirm_tokens").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"name":"host-ops",
		"goal_statement":"manage host actuations",
		"work_mode":"execute_bounded",
		"allowed_capabilities":["host.execute","comms.send"]
	}`
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/groups", body)
	assertStatus(t, rr, http.StatusAccepted)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected envelope data object, got %T", payload["data"])
	}
	if data["requires_approval"] != true {
		t.Fatalf("requires_approval = %v, want true", data["requires_approval"])
	}
	if _, ok := data["confirm_token"]; !ok {
		t.Fatal("expected confirm_token in approval response")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleCreateGroup_InvalidWorkMode(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/groups", s.HandleCreateGroup)

	body := `{"name":"ops","goal_statement":"run ops","work_mode":"wild-west"}`
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/groups", body)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleUpdateGroup_NotFound(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "PUT /api/v1/groups/{id}", s.HandleUpdateGroup)

	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE collaboration_groups").
		WillReturnResult(sqlmock.NewResult(0, 0))

	body := `{"name":"ops","goal_statement":"run ops","work_mode":"propose_only"}`
	rr := doAuthenticatedRequest(t, mux, "PUT", "/api/v1/groups/missing-id", body)
	assertStatus(t, rr, http.StatusNotFound)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleUpdateGroupStatus_ArchivesTemporaryGroup(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "PATCH /api/v1/groups/{id}/status", s.HandleUpdateGroupStatus)

	now := time.Now().UTC()
	auditID := "33333333-3333-3333-3333-333333333333"
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE collaboration_groups").
		WithArgs("group-temp", groupStatusArchived, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Temp Campaign",
				"Produce one campaign package",
				"execute_with_approval",
				[]byte(`["write_file"]`),
				[]byte(`["owner"]`),
				[]byte(`["11111111-1111-1111-1111-111111111111"]`),
				"marketing-lead",
				"",
				groupStatusArchived,
				"test-user-001",
				now.Add(2*time.Hour),
				auditID,
				auditID,
				now,
				now,
			))

	rr := doAuthenticatedRequest(t, mux, "PATCH", "/api/v1/groups/group-temp/status", `{"status":"archived"}`)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", payload["data"])
	}
	if data["status"] != groupStatusArchived {
		t.Fatalf("status = %v, want %q", data["status"], groupStatusArchived)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleGroupOutputs_ReturnsRetainedArtifacts(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt, func(s *AdminServer) {
		s.Artifacts = artifacts.NewService(s.DB, "/data/artifacts")
	})
	mux := setupMux(t, "GET /api/v1/groups/{id}/outputs", s.HandleGroupOutputs)

	now := time.Now().UTC()
	teamID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Temp Campaign",
				"Produce one campaign package",
				"execute_with_approval",
				[]byte(`["write_file"]`),
				[]byte(`["owner"]`),
				[]byte(`["11111111-1111-1111-1111-111111111111"]`),
				"marketing-lead",
				"",
				groupStatusArchived,
				"test-user-001",
				now.Add(2*time.Hour),
				"44444444-4444-4444-4444-444444444444",
				"55555555-5555-5555-5555-555555555555",
				now,
				now,
			))
	mock.ExpectQuery("SELECT .+ FROM artifacts WHERE team_id").
		WithArgs(teamID, 8).
		WillReturnRows(sqlmock.NewRows(artTestColumns).
			AddRow(
				"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				nil,
				teamID,
				"marketing-lead",
				nil,
				"document",
				"Launch Brief",
				"text/markdown",
				"Campaign summary",
				nil,
				nil,
				[]byte(`{}`),
				0.9,
				"approved",
				now,
			))

	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/groups/group-temp/outputs?limit=8", "")
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	items, ok := payload["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", payload["data"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 output, got %d", len(items))
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first output map, got %T", items[0])
	}
	if first["title"] != "Launch Brief" {
		t.Fatalf("title = %v, want Launch Brief", first["title"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleUpdateGroupStatus_ArchivesGroup(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	mux := setupMux(t, "PATCH /api/v1/groups/{id}/status", s.HandleUpdateGroupStatus)

	now := time.Now().UTC()
	auditID := "33333333-3333-3333-3333-333333333333"

	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE collaboration_groups").
		WithArgs("group-temp", groupStatusArchived, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Launch lane",
				"coordinate launch follow-through",
				"propose_only",
				[]byte(`["runs.read"]`),
				[]byte(`["u1"]`),
				[]byte(`["11111111-1111-1111-1111-111111111111"]`),
				"launch-profile",
				"policy.launch",
				groupStatusArchived,
				"test-user-001",
				nil,
				auditID,
				auditID,
				now,
				now,
			))

	rr := doAuthenticatedRequest(t, mux, "PATCH", "/api/v1/groups/group-temp/status", `{"status":"archived"}`)
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", payload["data"])
	}
	if got := data["status"]; got != groupStatusArchived {
		t.Fatalf("status = %v, want %s", got, groupStatusArchived)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleUpdateGroupStatus_InvalidStatus(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "PATCH /api/v1/groups/{id}/status", s.HandleUpdateGroupStatus)

	rr := doAuthenticatedRequest(t, mux, "PATCH", "/api/v1/groups/group-temp/status", `{"status":"gone"}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleGroupOutputs_ReturnsArtifactsForArchivedGroup(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(
		dbOpt,
		func(s *AdminServer) {
			s.Artifacts = artifacts.NewService(s.DB, "")
		},
	)
	mux := setupMux(t, "GET /api/v1/groups/{id}/outputs", s.HandleGroupOutputs)

	now := time.Now().UTC()
	teamID1 := uuid.New()
	teamID2 := uuid.New()
	artifactID1 := uuid.New()
	artifactID2 := uuid.New()

	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-temp").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-temp",
				"default",
				"Launch lane",
				"coordinate launch follow-through",
				"propose_only",
				[]byte(`["runs.read"]`),
				[]byte(`["u1"]`),
				[]byte(fmt.Sprintf(`["%s","%s"]`, teamID1.String(), teamID2.String())),
				"launch-profile",
				"policy.launch",
				groupStatusArchived,
				"test-user-001",
				nil,
				"",
				"",
				now,
				now,
			))

	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE team_id = \\$1").
		WithArgs(teamID1, 3).
		WillReturnRows(sqlmock.NewRows(artifactColumns()).
			AddRow(
				artifactID1,
				nil,
				&teamID1,
				"launch-lead",
				nil,
				"document",
				"Launch summary",
				"text/markdown",
				"# Launch summary",
				nil,
				nil,
				[]byte(`{}`),
				nil,
				"approved",
				now.Add(-2*time.Minute),
			))

	mock.ExpectQuery("SELECT .+ FROM artifacts\\s+WHERE team_id = \\$1").
		WithArgs(teamID2, 3).
		WillReturnRows(sqlmock.NewRows(artifactColumns()).
			AddRow(
				artifactID2,
				nil,
				&teamID2,
				"review-lead",
				nil,
				"document",
				"Review checklist",
				"text/markdown",
				"# Review checklist",
				nil,
				nil,
				[]byte(`{}`),
				nil,
				"approved",
				now.Add(-1*time.Minute),
			))

	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/groups/group-temp/outputs?limit=3", "")
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", payload["data"])
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(data))
	}

	first, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first artifact object, got %T", data[0])
	}
	second, ok := data[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second artifact object, got %T", data[1])
	}
	if first["title"] != "Review checklist" || second["title"] != "Launch summary" {
		t.Fatalf("unexpected artifact order: %#v", data)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleGroupBroadcast_FanoutParallel_DB(t *testing.T) {
	natsOpt := withNATS(t)
	dbOpt, mock := withDB(t)
	s := newTestServer(natsOpt, dbOpt, func(s *AdminServer) {
		s.GroupBus = NewGroupBusMonitor()
	})

	now := time.Now().UTC()
	auditID := "22222222-2222-2222-2222-222222222222"
	mock.ExpectQuery("SELECT id::text, tenant_id, name, goal_statement, work_mode").
		WithArgs("group-ops").
		WillReturnRows(sqlmock.NewRows(collaborationGroupColumns()).
			AddRow(
				"group-ops",
				"default",
				"ops",
				"parallel collaboration",
				"execute_with_approval",
				[]byte(`["runs.propose"]`),
				[]byte(`["u1","u2"]`),
				[]byte(`["admin-core","council-core"]`),
				"ops-profile",
				"policy.ops",
				groupStatusActive,
				"test-user-001",
				nil,
				auditID,
				auditID,
				now,
				now,
			))
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))

	groupSubj := fmt.Sprintf(protocol.TopicGroupCollabFmt, "group-ops")
	teamSubj1 := fmt.Sprintf(protocol.TopicTeamInternalCommand, "admin-core")
	teamSubj2 := fmt.Sprintf(protocol.TopicTeamInternalCommand, "council-core")

	groupCh := make(chan []byte, 1)
	team1Ch := make(chan []byte, 1)
	team2Ch := make(chan []byte, 1)

	if _, err := s.NC.Subscribe(groupSubj, func(msg *nats.Msg) { groupCh <- msg.Data }); err != nil {
		t.Fatalf("subscribe group: %v", err)
	}
	if _, err := s.NC.Subscribe(teamSubj1, func(msg *nats.Msg) { team1Ch <- msg.Data }); err != nil {
		t.Fatalf("subscribe team1: %v", err)
	}
	if _, err := s.NC.Subscribe(teamSubj2, func(msg *nats.Msg) { team2Ch <- msg.Data }); err != nil {
		t.Fatalf("subscribe team2: %v", err)
	}
	s.NC.Flush()

	mux := setupMux(t, "POST /api/v1/groups/{id}/broadcast", s.HandleGroupBroadcast)
	rr := doAuthenticatedRequest(t, mux, "POST", "/api/v1/groups/group-ops/broadcast", `{"message":"kickoff sync"}`)
	assertStatus(t, rr, http.StatusAccepted)

	select {
	case payload := <-groupCh:
		var decoded map[string]any
		if err := json.Unmarshal(payload, &decoded); err != nil {
			t.Fatalf("group payload decode: %v", err)
		}
		if decoded["group_id"] != "group-ops" {
			t.Fatalf("group_id = %v, want group-ops", decoded["group_id"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for group subject payload")
	}

	select {
	case <-team1Ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for team1 payload")
	}

	select {
	case <-team2Ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for team2 payload")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestHandleGroupMonitor_ReturnsSnapshot(t *testing.T) {
	s := newTestServer(withNATS(t), func(s *AdminServer) {
		s.GroupBus = NewGroupBusMonitor()
		s.GroupBus.RecordSuccess("group-1", "admin", "sync", []string{"swarm.group.group-1.collab"})
	})

	mux := setupMux(t, "GET /api/v1/groups/monitor", s.HandleGroupMonitor)
	rr := doAuthenticatedRequest(t, mux, "GET", "/api/v1/groups/monitor", "")
	assertStatus(t, rr, http.StatusOK)

	var payload map[string]any
	assertJSON(t, rr, &payload)
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", payload["data"])
	}
	if data["published_count"] == nil {
		t.Fatalf("expected published_count in snapshot, got %v", data)
	}
}

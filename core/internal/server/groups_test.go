package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

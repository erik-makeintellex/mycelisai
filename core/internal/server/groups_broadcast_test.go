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

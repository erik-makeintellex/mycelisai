package reactive

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// startTestNATS spins up an embedded NATS server for isolated testing.
func startTestNATS(t *testing.T) (*natsserver.Server, *nats.Conn) {
	t.Helper()
	opts := &natsserver.Options{Port: -1}
	srv, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		t.Fatalf("nats connect: %v", err)
	}
	return srv, nc
}

// ── New ───────────────────────────────────────────────────────────

func TestNew_NilNATS(t *testing.T) {
	handler := func(profileID, topic string, msg []byte) {}
	e := New(nil, handler)
	if e == nil {
		t.Fatal("expected non-nil Engine")
	}
	if e.nc != nil {
		t.Error("expected nil NATS connection")
	}
	if e.subs == nil {
		t.Error("expected non-nil subs map")
	}
	if e.handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestNew_WithNATS(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	handler := func(profileID, topic string, msg []byte) {}
	e := New(nc, handler)
	if e == nil {
		t.Fatal("expected non-nil Engine")
	}
	if e.nc == nil {
		t.Error("expected non-nil NATS connection")
	}
	if e.subs == nil {
		t.Error("expected non-nil subs map")
	}
}

func TestNew_NilHandler(t *testing.T) {
	e := New(nil, nil)
	if e == nil {
		t.Fatal("expected non-nil Engine even with nil handler")
	}
	if e.handler != nil {
		t.Error("expected nil handler")
	}
}

// ── Connected ─────────────────────────────────────────────────────

func TestConnected_NilNC(t *testing.T) {
	e := New(nil, nil)
	if e.Connected() {
		t.Error("expected Connected() == false with nil nc")
	}
}

func TestConnected_NilEngine(t *testing.T) {
	var e *Engine
	if e.Connected() {
		t.Error("expected Connected() == false on nil Engine")
	}
}

func TestConnected_LiveConnection(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	if !e.Connected() {
		t.Error("expected Connected() == true with live NATS")
	}
}

func TestConnected_ClosedConnection(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()

	e := New(nc, nil)
	nc.Close()
	if e.Connected() {
		t.Error("expected Connected() == false after NATS close")
	}
}

// ── Subscribe ─────────────────────────────────────────────────────

func TestSubscribe_NilNC(t *testing.T) {
	e := New(nil, nil)
	// Subscribe with nil nc should not return an error — it defers gracefully.
	err := e.Subscribe("profile-1", []ProfileSubscription{
		{Topic: "swarm.team.research.*"},
	})
	if err != nil {
		t.Errorf("expected nil error for deferred subscribe, got: %v", err)
	}
	// No subscriptions should be created.
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 active subs with nil nc, got %d", e.ActiveSubscriptionCount())
	}
}

func TestSubscribe_ValidTopics(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	called := make(chan string, 10)
	handler := func(profileID, topic string, msg []byte) {
		called <- profileID + ":" + topic
	}

	e := New(nc, handler)
	defer e.Close()

	err := e.Subscribe("prof-alpha", []ProfileSubscription{
		{Topic: "swarm.team.alpha.*"},
		{Topic: "swarm.team.beta.>"},
	})
	if err != nil {
		t.Fatalf("Subscribe error: %v", err)
	}
	if e.ActiveSubscriptionCount() != 2 {
		t.Errorf("expected 2 active subs, got %d", e.ActiveSubscriptionCount())
	}

	// Publish a message to one of the subscribed topics and verify the handler fires.
	nc.Publish("swarm.team.alpha.telemetry", []byte(`{"hello":"world"}`))
	nc.Flush()

	select {
	case got := <-called:
		if got != "prof-alpha:swarm.team.alpha.telemetry" {
			t.Errorf("unexpected handler call: %s", got)
		}
	case <-time.After(2 * time.Second):
		t.Error("handler was not called within timeout")
	}
}

func TestSubscribe_SkipsEmptyTopic(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	defer e.Close()

	err := e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: ""},                        // should be skipped
		{Topic: "swarm.team.gamma.events"}, // should be created
	})
	if err != nil {
		t.Fatalf("Subscribe error: %v", err)
	}
	if e.ActiveSubscriptionCount() != 1 {
		t.Errorf("expected 1 active sub (empty topic skipped), got %d", e.ActiveSubscriptionCount())
	}
}

func TestSubscribe_ReplacesExisting(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	defer e.Close()

	// First subscription set.
	err := e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: "old.topic.1"},
		{Topic: "old.topic.2"},
	})
	if err != nil {
		t.Fatalf("Subscribe (first) error: %v", err)
	}
	if e.ActiveSubscriptionCount() != 2 {
		t.Errorf("expected 2 subs after first Subscribe, got %d", e.ActiveSubscriptionCount())
	}

	// Replace with a single new topic.
	err = e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: "new.topic.1"},
	})
	if err != nil {
		t.Fatalf("Subscribe (replace) error: %v", err)
	}
	if e.ActiveSubscriptionCount() != 1 {
		t.Errorf("expected 1 sub after replacement, got %d", e.ActiveSubscriptionCount())
	}
}

func TestSubscribe_AllEmpty(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	defer e.Close()

	err := e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: ""},
		{Topic: ""},
	})
	if err != nil {
		t.Fatalf("Subscribe error: %v", err)
	}
	// All topics empty — nothing stored.
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs when all topics are empty, got %d", e.ActiveSubscriptionCount())
	}
}

// ── Unsubscribe ───────────────────────────────────────────────────

func TestUnsubscribe_ExistingProfile(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	defer e.Close()

	_ = e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: "topic.a"},
		{Topic: "topic.b"},
	})
	if e.ActiveSubscriptionCount() != 2 {
		t.Fatalf("expected 2 subs before unsubscribe, got %d", e.ActiveSubscriptionCount())
	}

	e.Unsubscribe("prof-1")
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs after unsubscribe, got %d", e.ActiveSubscriptionCount())
	}
}

func TestUnsubscribe_NonExistentProfile(t *testing.T) {
	e := New(nil, nil)
	// Should not panic.
	e.Unsubscribe("does-not-exist")
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs, got %d", e.ActiveSubscriptionCount())
	}
}

func TestUnsubscribe_OnlyTargetProfile(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	defer e.Close()

	_ = e.Subscribe("prof-a", []ProfileSubscription{{Topic: "topic.a"}})
	_ = e.Subscribe("prof-b", []ProfileSubscription{{Topic: "topic.b"}})
	if e.ActiveSubscriptionCount() != 2 {
		t.Fatalf("expected 2 subs across two profiles, got %d", e.ActiveSubscriptionCount())
	}

	e.Unsubscribe("prof-a")
	if e.ActiveSubscriptionCount() != 1 {
		t.Errorf("expected 1 sub remaining after unsubscribing prof-a, got %d", e.ActiveSubscriptionCount())
	}
}

// ── ActiveSubscriptionCount ───────────────────────────────────────

func TestActiveSubscriptionCount_Empty(t *testing.T) {
	e := New(nil, nil)
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs on fresh engine, got %d", e.ActiveSubscriptionCount())
	}
}

func TestActiveSubscriptionCount_MultipleProfiles(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	defer e.Close()

	_ = e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: "topic.1a"},
		{Topic: "topic.1b"},
	})
	_ = e.Subscribe("prof-2", []ProfileSubscription{
		{Topic: "topic.2a"},
	})

	if e.ActiveSubscriptionCount() != 3 {
		t.Errorf("expected 3 total subs, got %d", e.ActiveSubscriptionCount())
	}
}

// ── Close ─────────────────────────────────────────────────────────

func TestClose_EmptyEngine(t *testing.T) {
	e := New(nil, nil)
	// Should not panic with no subscriptions.
	e.Close()
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs after Close, got %d", e.ActiveSubscriptionCount())
	}
}

func TestClose_WithActiveSubscriptions(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)

	_ = e.Subscribe("prof-1", []ProfileSubscription{{Topic: "topic.1"}})
	_ = e.Subscribe("prof-2", []ProfileSubscription{{Topic: "topic.2"}, {Topic: "topic.3"}})
	if e.ActiveSubscriptionCount() != 3 {
		t.Fatalf("expected 3 subs before Close, got %d", e.ActiveSubscriptionCount())
	}

	e.Close()
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs after Close, got %d", e.ActiveSubscriptionCount())
	}
}

func TestClose_Idempotent(t *testing.T) {
	e := New(nil, nil)
	// Multiple closes should not panic.
	e.Close()
	e.Close()
	e.Close()
}

// ── SetDB ─────────────────────────────────────────────────────────

func TestSetDB(t *testing.T) {
	e := New(nil, nil)
	if e.db != nil {
		t.Error("expected nil db initially")
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e.SetDB(db)
	if e.db == nil {
		t.Error("expected non-nil db after SetDB")
	}
}

func TestSetDB_Nil(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nil, nil)
	e.SetDB(db)
	if e.db == nil {
		t.Fatal("expected db to be set")
	}

	// Setting nil clears the db.
	e.SetDB(nil)
	if e.db != nil {
		t.Error("expected nil db after SetDB(nil)")
	}
}

// ── ReactivateFromDB ──────────────────────────────────────────────

func TestReactivateFromDB_NilDB(t *testing.T) {
	e := New(nil, nil)
	// Both db and nc are nil — should return nil immediately.
	err := e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Errorf("expected nil error when db is nil, got: %v", err)
	}
}

func TestReactivateFromDB_NilNC(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nil, nil)
	e.SetDB(db)
	// nc is nil — should return nil immediately without querying.
	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Errorf("expected nil error when nc is nil, got: %v", err)
	}
}

func TestReactivateFromDB_WithActiveProfiles(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	handler := func(profileID, topic string, msg []byte) {}
	e := New(nc, handler)
	e.SetDB(db)
	defer e.Close()

	subs1, _ := json.Marshal([]ProfileSubscription{
		{Topic: "swarm.team.research.*"},
	})
	subs2, _ := json.Marshal([]ProfileSubscription{
		{Topic: "swarm.team.ops.>"},
		{Topic: "swarm.events.mission.*"},
	})

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscriptions"}).
			AddRow("prof-aaa", subs1).
			AddRow("prof-bbb", subs2))

	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Fatalf("ReactivateFromDB error: %v", err)
	}
	// prof-aaa: 1 sub, prof-bbb: 2 subs = 3 total
	if e.ActiveSubscriptionCount() != 3 {
		t.Errorf("expected 3 active subs after reactivation, got %d", e.ActiveSubscriptionCount())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

func TestReactivateFromDB_EmptySubscriptions(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nc, nil)
	e.SetDB(db)
	defer e.Close()

	emptySubs, _ := json.Marshal([]ProfileSubscription{})

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscriptions"}).
			AddRow("prof-empty", emptySubs))

	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Fatalf("ReactivateFromDB error: %v", err)
	}
	// Empty subscriptions array — nothing created.
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs for empty subscriptions array, got %d", e.ActiveSubscriptionCount())
	}
}

func TestReactivateFromDB_InvalidJSON(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nc, nil)
	e.SetDB(db)
	defer e.Close()

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscriptions"}).
			AddRow("prof-bad", []byte(`not valid json`)))

	err = e.ReactivateFromDB(context.Background())
	if err != nil {
		t.Fatalf("ReactivateFromDB should not error on bad JSON (skips row), got: %v", err)
	}
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs for invalid JSON, got %d", e.ActiveSubscriptionCount())
	}
}

func TestReactivateFromDB_QueryError(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	e := New(nc, nil)
	e.SetDB(db)
	defer e.Close()

	mock.ExpectQuery("SELECT .+ FROM mission_profiles").
		WillReturnError(sql.ErrConnDone)

	err = e.ReactivateFromDB(context.Background())
	if err == nil {
		t.Error("expected error from query failure")
	}
}

// ── Handler invocation ────────────────────────────────────────────

func TestSubscribe_NilHandler_NoPanic(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	// Engine with nil handler — messages should not panic.
	e := New(nc, nil)
	defer e.Close()

	err := e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: "swarm.safe.topic"},
	})
	if err != nil {
		t.Fatalf("Subscribe error: %v", err)
	}

	// Publish a message — the nil handler guard in the closure should prevent panic.
	nc.Publish("swarm.safe.topic", []byte(`test`))
	nc.Flush()

	// Give the async handler time to (not) panic.
	time.Sleep(200 * time.Millisecond)
}

// ── ProfileSubscription JSON ──────────────────────────────────────

func TestProfileSubscription_JSON(t *testing.T) {
	tests := []struct {
		name  string
		input ProfileSubscription
	}{
		{
			name:  "topic only",
			input: ProfileSubscription{Topic: "swarm.team.alpha.*"},
		},
		{
			name:  "topic with condition",
			input: ProfileSubscription{Topic: "swarm.events.>", Condition: "severity=critical"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			var decoded ProfileSubscription
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if decoded.Topic != tt.input.Topic {
				t.Errorf("Topic: got %q, want %q", decoded.Topic, tt.input.Topic)
			}
			if decoded.Condition != tt.input.Condition {
				t.Errorf("Condition: got %q, want %q", decoded.Condition, tt.input.Condition)
			}
		})
	}
}

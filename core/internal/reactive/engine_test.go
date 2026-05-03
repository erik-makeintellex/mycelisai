package reactive

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

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

func TestClose_EmptyEngine(t *testing.T) {
	e := New(nil, nil)
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
	e.Close()
	e.Close()
	e.Close()
}

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

	e.SetDB(nil)
	if e.db != nil {
		t.Error("expected nil db after SetDB(nil)")
	}
}

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

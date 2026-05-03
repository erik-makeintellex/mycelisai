package reactive

import (
	"testing"
	"time"
)

func TestSubscribe_NilNC(t *testing.T) {
	e := New(nil, nil)
	err := e.Subscribe("profile-1", []ProfileSubscription{
		{Topic: "swarm.team.research.*"},
	})
	if err != nil {
		t.Errorf("expected nil error for deferred subscribe, got: %v", err)
	}
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
		{Topic: ""},
		{Topic: "swarm.team.gamma.events"},
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
	if e.ActiveSubscriptionCount() != 0 {
		t.Errorf("expected 0 subs when all topics are empty, got %d", e.ActiveSubscriptionCount())
	}
}

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

func TestSubscribe_NilHandler_NoPanic(t *testing.T) {
	srv, nc := startTestNATS(t)
	defer srv.Shutdown()
	defer nc.Close()

	e := New(nc, nil)
	defer e.Close()

	err := e.Subscribe("prof-1", []ProfileSubscription{
		{Topic: "swarm.safe.topic"},
	})
	if err != nil {
		t.Fatalf("Subscribe error: %v", err)
	}

	nc.Publish("swarm.safe.topic", []byte(`test`))
	nc.Flush()

	time.Sleep(200 * time.Millisecond)
}

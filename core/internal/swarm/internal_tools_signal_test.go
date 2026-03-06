package swarm

import (
	"context"
	"strings"
	"testing"
	"time"

	server "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func TestHandleDelegateTask_PublishesToInternalCommand(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc})

	done := make(chan string, 1)
	if _, err := nc.Subscribe("swarm.team.alpha.internal.command", func(msg *nats.Msg) {
		done <- string(msg.Data)
	}); err != nil {
		t.Fatalf("subscribe command: %v", err)
	}
	nc.Flush()

	out, err := reg.handleDelegateTask(context.Background(), map[string]any{
		"team_id": "alpha",
		"task":    "inspect gate state",
	})
	if err != nil {
		t.Fatalf("delegate_task error: %v", err)
	}
	if !strings.Contains(out, "alpha") {
		t.Fatalf("unexpected delegate output: %s", out)
	}

	select {
	case got := <-done:
		if got != "inspect gate state" {
			t.Fatalf("task payload = %q", got)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for delegated task")
	}
}

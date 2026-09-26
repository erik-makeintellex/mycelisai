package router

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/state"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func TestIsHeartbeatEnvelope(t *testing.T) {
	if !isHeartbeatEnvelope(protocol.TopicGlobalHeartbeat, &pb.MsgEnvelope{}) {
		t.Fatal("expected global heartbeat subject to be treated as heartbeat traffic")
	}
	if !isHeartbeatEnvelope("swarm.team.alpha.signal.status", &pb.MsgEnvelope{
		Payload: &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: "agent.heartbeat"}},
	}) {
		t.Fatal("expected heartbeat event payload to be treated as heartbeat traffic")
	}
	if isHeartbeatEnvelope("swarm.team.alpha.signal.result", &pb.MsgEnvelope{
		Payload: &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: "task.completed"}},
	}) {
		t.Fatal("did not expect non-heartbeat event to be treated as heartbeat traffic")
	}
}

func startTestNATS(t *testing.T) *nats.Conn {
	t.Helper()
	srv, err := natsserver.NewServer(&natsserver.Options{Host: "127.0.0.1", Port: -1})
	if err != nil {
		t.Fatalf("nats server: %v", err)
	}
	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		t.Fatal("nats server not ready")
	}
	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		srv.Shutdown()
		t.Fatalf("nats connect: %v", err)
	}
	t.Cleanup(func() { nc.Close(); srv.Shutdown() })
	return nc
}

// reactions counts Core reactions: audit traces and approval requests.
func reactions(t *testing.T, nc *nats.Conn) func() int64 {
	t.Helper()
	var n int64
	for _, subject := range []string{protocol.TopicAuditTrace, "swarm.governance.needed"} {
		if _, err := nc.Subscribe(subject, func(*nats.Msg) { atomic.AddInt64(&n, 1) }); err != nil {
			t.Fatal(err)
		}
	}
	if err := nc.Flush(); err != nil {
		t.Fatal(err)
	}
	return func() int64 {
		time.Sleep(150 * time.Millisecond) // audit trace publishes asynchronously
		_ = nc.Flush()
		return atomic.LoadInt64(&n)
	}
}

func deliver(t *testing.T, r *Router, subject, agent, eventType string) {
	t.Helper()
	data, err := proto.Marshal(&pb.MsgEnvelope{
		SourceAgentId: agent,
		TeamId:        "alpha",
		Payload:       &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: eventType}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r.handleMessage(&nats.Msg{Subject: subject, Data: data})
}

func agentSeen(id string) bool {
	for _, agent := range state.GlobalRegistry.GetActiveAgents() {
		if agent.ID == id {
			return true
		}
	}
	return false
}

func TestRouterFailsClosedWithoutLoadedPolicy(t *testing.T) {
	for name, guard := range map[string]*governance.Guard{
		"nil":      nil,
		"degraded": governance.NewDegradedGuard(errors.New("missing policy")),
	} {
		t.Run(name, func(t *testing.T) {
			nc := startTestNATS(t)
			count := reactions(t, nc)
			r := NewRouter(nc, guard)
			hb := "hb-" + name + "-" + time.Now().Format("150405.000000")
			deliver(t, r, protocol.TopicGlobalHeartbeat, hb, "agent.heartbeat")
			if !agentSeen(hb) {
				t.Fatal("heartbeat must still update the registry")
			}
			worker := "work-" + name + "-" + time.Now().Format("150405.000000")
			deliver(t, r, "swarm.team.alpha.signal.result", worker, "task.completed")
			deliver(t, r, "swarm.team.alpha.signal.result", worker, "k8s.delete.cluster")
			if got := count(); got != 0 {
				t.Fatalf("expected no audit trace or approval publish, got %d", got)
			}
		})
	}
}

func TestRouterReactsWithLoadedPolicy(t *testing.T) {
	nc := startTestNATS(t)
	count := reactions(t, nc)
	guard := governance.NewDegradedGuard(errors.New("start degraded"))
	guard.UpdatePolicyConfig(&governance.PolicyConfig{
		Groups: []governance.PolicyGroup{{
			Name: "g", Targets: []string{"*"},
			Rules: []governance.PolicyRule{{Intent: "k8s.delete.cluster", Action: governance.ActionRequireApproval}},
		}},
		Defaults: governance.DefaultConfig{DefaultAction: governance.ActionAllow},
	})
	r := NewRouter(nc, guard)
	deliver(t, r, "swarm.team.alpha.signal.result", "w", "task.completed")
	deliver(t, r, "swarm.team.alpha.signal.result", "w", "k8s.delete.cluster")
	if got := count(); got != 2 {
		t.Fatalf("expected one audit trace and one approval request after recovery, got %d", got)
	}
}

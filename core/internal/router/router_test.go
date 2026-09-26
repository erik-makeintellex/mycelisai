package router

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/state"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func TestIsHeartbeatEnvelopeIsExact(t *testing.T) {
	hb := func(agent, eventType string) *pb.MsgEnvelope {
		return &pb.MsgEnvelope{SourceAgentId: agent,
			Payload: &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: eventType}}}
	}
	if !isHeartbeatEnvelope(protocol.TopicGlobalHeartbeat, hb("a", "agent.heartbeat")) {
		t.Fatal("canonical subject + agent.heartbeat + source agent is a heartbeat")
	}
	for name, c := range map[string]struct {
		subject string
		env     *pb.MsgEnvelope
	}{
		"nil envelope":          {protocol.TopicGlobalHeartbeat, nil},
		"no event":              {protocol.TopicGlobalHeartbeat, &pb.MsgEnvelope{SourceAgentId: "a"}},
		"wrong event":           {protocol.TopicGlobalHeartbeat, hb("a", "task.completed")},
		"no source agent":       {protocol.TopicGlobalHeartbeat, hb("", "agent.heartbeat")},
		"substring subject":     {"swarm.team.x.agent.y.heartbeat", hb("a", "agent.heartbeat")},
		"event on other subj":   {"swarm.team.alpha.signal.status", hb("a", "agent.heartbeat")},
		"padded subject":        {" " + protocol.TopicGlobalHeartbeat, hb("a", "agent.heartbeat")},
		"heartbeat-ish subject": {protocol.TopicGlobalHeartbeat + ".x", hb("a", "agent.heartbeat")},
	} {
		if isHeartbeatEnvelope(c.subject, c.env) {
			t.Errorf("%s: must not be treated as a heartbeat", name)
		}
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
			known := "hb-" + name + "-" + time.Now().Format("150405.000000")
			state.GlobalRegistry.UpdateHeartbeat(known, "alpha", "", state.StatusIdle)
			deliver(t, r, protocol.TopicGlobalHeartbeat, known, "agent.heartbeat")
			if !agentSeen(known) {
				t.Fatal("heartbeat must still refresh a registered agent")
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

func deliverEnv(t *testing.T, r *Router, subject string, env *pb.MsgEnvelope) {
	t.Helper()
	data, err := proto.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	r.handleMessage(&nats.Msg{Subject: subject, Data: data})
}

func heartbeatEnv(agent, team, sourceURI string) *pb.MsgEnvelope {
	env := &pb.MsgEnvelope{SourceAgentId: agent, TeamId: team,
		Payload: &pb.MsgEnvelope_Event{Event: &pb.EventPayload{EventType: "agent.heartbeat"}}}
	if sourceURI != "" {
		env.SwarmContext = &structpb.Struct{Fields: map[string]*structpb.Value{
			"source_uri": structpb.NewStringValue(sourceURI)}}
	}
	return env
}

// A2b item 3: while governance is degraded, heartbeats only refresh agents
// already registered; spoofed subjects, non-heartbeat subjects and unknown
// agents never change the registry.
func TestDegradedHeartbeatOnlyRefreshesKnownAgents(t *testing.T) {
	for name, guard := range map[string]*governance.Guard{
		"nil":      nil,
		"degraded": governance.NewDegradedGuard(errors.New("missing policy")),
	} {
		t.Run(name, func(t *testing.T) {
			nc := startTestNATS(t)
			r := NewRouter(nc, guard)
			suffix := name + "-" + time.Now().Format("150405.000000")

			spoof := "spoof-" + suffix
			deliverEnv(t, r, "swarm.team.x.agent.y.heartbeat", heartbeatEnv(spoof, "evil", "evil://"))
			deliverEnv(t, r, "swarm.team.alpha.signal.status", heartbeatEnv(spoof, "evil", "evil://"))
			deliverEnv(t, r, protocol.TopicGlobalHeartbeat, heartbeatEnv(spoof, "evil", "evil://"))
			if _, ok := state.GlobalRegistry.Get(spoof); ok {
				t.Fatal("degraded heartbeat must never create an agent")
			}

			known := "known-" + suffix
			state.GlobalRegistry.UpdateHeartbeat(known, "alpha", "swarm:base", state.StatusIdle)
			before, _ := state.GlobalRegistry.Get(known)
			time.Sleep(2 * time.Millisecond)
			deliverEnv(t, r, "swarm.team.x.agent."+known+".heartbeat", heartbeatEnv(known, "evil", "evil://"))
			if got, _ := state.GlobalRegistry.Get(known); !got.LastHeartbeat.Equal(before.LastHeartbeat) {
				t.Fatal("a substring heartbeat subject must be ignored")
			}
			deliverEnv(t, r, protocol.TopicGlobalHeartbeat, heartbeatEnv(known, "evil", "evil://"))
			after, _ := state.GlobalRegistry.Get(known)
			if !after.LastHeartbeat.After(before.LastHeartbeat) {
				t.Fatal("canonical heartbeat must advance last seen for a known agent")
			}
			if after.TeamID != "alpha" || after.SourceURI != "swarm:base" {
				t.Fatalf("degraded heartbeat must not rewrite team/source: %+v", after)
			}
		})
	}
}

func TestHealthyHeartbeatRegistersOnlyCanonical(t *testing.T) {
	nc := startTestNATS(t)
	guard := governance.NewDegradedGuard(errors.New("start degraded"))
	guard.UpdatePolicyConfig(&governance.PolicyConfig{
		Defaults: governance.DefaultConfig{DefaultAction: governance.ActionAllow},
	})
	r := NewRouter(nc, guard)
	suffix := time.Now().Format("150405.000000")
	spoof := "healthy-spoof-" + suffix
	deliverEnv(t, r, "swarm.team.x.agent.y.heartbeat", heartbeatEnv(spoof, "beta", ""))
	if _, ok := state.GlobalRegistry.Get(spoof); ok {
		t.Fatal("a substring heartbeat subject must not register an agent")
	}
	agent := "healthy-" + suffix
	deliverEnv(t, r, protocol.TopicGlobalHeartbeat, heartbeatEnv(agent, "beta", "swarm:base"))
	got, ok := state.GlobalRegistry.Get(agent)
	if !ok || got.TeamID != "beta" || got.SourceURI != "swarm:base" {
		t.Fatalf("healthy canonical heartbeat must register the agent: %+v ok=%v", got, ok)
	}
}

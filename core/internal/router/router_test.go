package router

import (
	"testing"

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

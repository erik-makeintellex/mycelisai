package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func main() {
	log.Println("🛡️  Starting Governance Smoke Test (Go Edition)...")

	// 1. Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("❌ NATS Connection Failed: %v", err)
	}
	defer nc.Close()
	log.Println("✅ Connected to NATS")

	// 2. Test 1: Immediate Block (System Shutdown)
	log.Println("\n🧪 Test 1: Immediate Block (system.shutdown)")

	shutdownEnv := &pb.MsgEnvelope{
		Id:            "test-id-1",
		SourceAgentId: "smoke-tester",
		TeamId:        "devops",
		Timestamp:     timestamppb.Now(),
		Payload: &pb.MsgEnvelope_Event{
			Event: &pb.EventPayload{
				EventType: "system.shutdown",
			},
		},
	}

	send(nc, shutdownEnv)
	log.Println("   Sent 'system.shutdown'. Monitor Core logs for DENY.")
	time.Sleep(1 * time.Second)

	// 3. Test 2: Require Approval (Payment > 50)
	log.Println("\n🧪 Test 2: Require approval (payment.create > 50) is observed, not parked")

	// Construct complex context/data
	// We need to match the Policy: 'amount > 50'
	// The Guard checks:
	// a) SwarmContext (headers)
	// b) Event.Data (payload)

	// Let's put 'amount' in Event Data
	dataStruct, _ := structpb.NewStruct(map[string]interface{}{
		"amount":   100.0,
		"currency": "USD",
	})

	payEnv := &pb.MsgEnvelope{
		Id:            "test-id-2",
		SourceAgentId: "finance-bot",
		TeamId:        "finance",
		Timestamp:     timestamppb.Now(),
		Payload: &pb.MsgEnvelope_Event{
			Event: &pb.EventPayload{
				EventType: "payment.create",
				Data:      dataStruct,
			},
		},
	}

	send(nc, payEnv)
	log.Println("   Sent 'payment.create' ($100). Core records one policy_approval_required_observed audit row (or logs audit_unavailable); nothing is parked or published.")

	// This command only publishes probes. It does not read Core's audit log or
	// logs, so it certifies neither outcome: check the Core log and the Audit
	// tab. Governed approval is proven on the confirm-action/proposal path.
	log.Println("\nℹ️  Publish-only probe: verify DENY and the policy_approval_required_observed record in Core's log or Automations > Approvals > Audit.")
}

func send(nc *nats.Conn, env *pb.MsgEnvelope) {
	data, err := proto.Marshal(env)
	if err != nil {
		log.Fatalf("Marshal failed: %v", err)
	}
	// The router subscribes to every product subject (protocol.TopicSwarmWild)
	// and runs the governance guard on it; publish on the envelope team's
	// machine telemetry lane so the probe never lands on an operator channel.
	if err := nc.Publish(fmt.Sprintf(protocol.TopicTeamTelemetryFmt, env.TeamId), data); err != nil {
		log.Fatalf("Publish failed: %v", err)
	}
	if err := nc.Flush(); err != nil {
		log.Fatalf("Flush failed: %v", err)
	}
}

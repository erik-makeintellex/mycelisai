package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/mycelis/core/pkg/pb/swarm"
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
	log.Println("\n🧪 Test 2: Park Request (payment.create > 50)")

	// Construct complex context/data
	// We need to match the Policy: 'amount > 50'
	// The Gatekeeper checks:
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
	log.Println("   Sent 'payment.create' ($100). Expect Pending Approval.")
	time.Sleep(2 * time.Second) // Give it time to process and park

	// 4. Check API
	log.Println("\n🧪 Test 3: Checking Admin API for Approvals...")
	resp, err := http.Get("http://localhost:8080/admin/approvals")
	if err != nil {
		log.Fatalf("❌ API Failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("❌ API Returned %d", resp.StatusCode)
	}

	var approvals []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&approvals); err != nil {
		log.Fatalf("❌ Failed to decode JSON: %v", err)
	}

	log.Printf("✅ API Response: Found %d pending approvals.", len(approvals))
	if len(approvals) > 0 {
		log.Println("🎉 SUCCESS! Gatekeeper logic verified.")
	} else {
		log.Println("⚠️  WARNING: No approvals found. Gatekeeper might have allowed it or failed to park.")
	}
}

func send(nc *nats.Conn, env *pb.MsgEnvelope) {
	data, err := proto.Marshal(env)
	if err != nil {
		log.Fatalf("Marshal failed: %v", err)
	}
	// Publish to a topic that triggers the router interception?
	// Router subscribes to "swarm.>"
	// We send to "swarm.smoke.test"
	if err := nc.Publish("swarm.smoke.test", data); err != nil {
		log.Fatalf("Publish failed: %v", err)
	}
}

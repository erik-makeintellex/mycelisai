package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/mycelis/core/pkg/pb/swarm"
)

func main() {
	log.Println("Starting Probe (Fake Radiant)...")

	// 1. Connect
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Drain()

	// 2. Construct Message
	id := "radiant-qa-001"
	msg := &pb.MsgEnvelope{
		Id:            "msg-" + time.Now().String(),
		Timestamp:     timestamppb.Now(),
		SourceAgentId: id,
		Type:          pb.MessageType_MESSAGE_TYPE_EVENT,
		Payload: &pb.MsgEnvelope_Event{
			Event: &pb.EventPayload{
				EventType: "heartbeat",
				StreamId:  "main",
			},
		},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalf("Marshal error: %v", err)
	}

	// 3. Publish
	subject := "swarm.prod.agent." + id + ".heartbeat"
	if err := nc.Publish(subject, data); err != nil {
		log.Fatalf("Publish error: %v", err)
	}

	log.Printf("STIMULUS SENT: Heartbeat -> %s", subject)

	// Give it a moment to flush
	nc.Flush()
}

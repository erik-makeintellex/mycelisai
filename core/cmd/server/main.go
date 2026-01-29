package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/mycelis/core/internal/state"
	mycelis_nats "github.com/mycelis/core/internal/transport/nats"
	pb "github.com/mycelis/core/pkg/pb/swarm"
)

func main() {
	log.Println("Starting Mycelis Core [Brain]...")

	// 1. Config
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL // nats://localhost:4222
	}

	// 2. Connect to NATS
	log.Printf("Connecting to NATS at %s...", natsURL)
	ncWrapper, err := mycelis_nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer ncWrapper.Drain()

	nc := ncWrapper.Conn

	// 3. Subscribe to Heartbeats
	// Subject: swarm.prod.agent.*.heartbeat
	// We use a wildcard to capture all agents
	subject := "swarm.prod.agent.*.heartbeat"

	_, err = nc.Subscribe(subject, func(msg *nats.Msg) {
		// Log receipt (debug)
		// log.Printf("Received heartbeat on %s", msg.Subject)

		// 4. React (Parse & Update Registry)
		var envelope pb.MsgEnvelope
		if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
			log.Printf("Error unmarshaling heartbeat: %v", err)
			return
		}

		// Look for EventPayload inside OneOf
		// In V1 usage, we might expect specific event types.
		// For simplicity, we assume presence indicates "Alive/Idle" unless specified.
		// TODO: Parse status from payload if rich heartbeat.

		agentID := envelope.SourceAgentId
		if agentID == "" {
			return
		}

		// Update Registry
		state.GlobalRegistry.UpdateHeartbeat(agentID, state.StatusIdle)
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to heartbeats: %v", err)
	}

	log.Printf("Listening for heartbeats on %s", subject)

	// 5. Serve HTTP (Health & Metrics)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		agents := state.GlobalRegistry.GetActiveAgents()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"active_agents\": %d}", len(agents))
		// JSON serialization of full list left as exercise for V1.1
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("HTTP Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP Server failed: %v", err)
	}
}

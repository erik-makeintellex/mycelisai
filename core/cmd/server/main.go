package main

import (
	"encoding/json"
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
		// 4. React (Parse & Update Registry)
		var envelope pb.MsgEnvelope
		if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
			log.Printf("Error unmarshaling heartbeat: %v", err)
			return
		}

		agentID := envelope.SourceAgentId
		if agentID == "" {
			return
		}

		// EXTRACT: SourceURI from Context (System Standard Phase 3)
		// Logic: If "source_uri" key exists in swarm_context Struct, use it.
		sourceURI := ""
		if envelope.SwarmContext != nil && envelope.SwarmContext.Fields != nil {
			if val, ok := envelope.SwarmContext.Fields["source_uri"]; ok {
				sourceURI = val.GetStringValue()
			}
		}

		// Fallback: Check Payload if it's an Event (e.g. "agent.startup")
		if sourceURI == "" {
			if evt := envelope.GetEvent(); evt != nil {
				// Sometimes transmitted in event data
				if val, ok := evt.Data.Fields["source_uri"]; ok {
					sourceURI = val.GetStringValue()
				}
			}
		}

		// Update Registry
		state.GlobalRegistry.UpdateHeartbeat(agentID, envelope.TeamId, sourceURI, state.StatusIdle)
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

		response := map[string]interface{}{
			"active_agents": len(agents),
			"agents":        agents,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("JSON Encode Error: %v", err)
		}
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

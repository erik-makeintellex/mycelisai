package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/mycelis/core/internal/governance"
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

	// 1b. Load Governance Policy
	policyPath := "core/policy/policy.yaml" // Assuming we run from root or handle path
	// Fix path for container vs local? For now, relative.

	gk, err := governance.NewGatekeeper(policyPath)
	if err != nil {
		log.Printf("⚠️ Governance Policy not loaded: %v. Allowing all.", err)
		gk = nil
	} else {
		log.Println("🛡️ Gatekeeper Active.")
	}

	// 2. Connect to NATS
	log.Printf("Connecting to NATS at %s...", natsURL)
	ncWrapper, err := mycelis_nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer ncWrapper.Drain()

	nc := ncWrapper.Conn

	// 3. Subscribe to EVERYTHING (Router Mode)
	// Subject: swarm.>
	// We capture specific traffic for Heartbeats AND Governance
	subject := "swarm.>"

	_, err = nc.Subscribe(subject, func(msg *nats.Msg) {
		// A. PROTO UNMARSHAL
		var envelope pb.MsgEnvelope
		if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
			// Not a proto message? Ignore.
			// log.Printf("Error unmarshaling: %v", err)
			return
		}

		// B. GOVERNANCE INTERCEPT
		if gk != nil {
			allowed, action, reqID := gk.Intercept(&envelope)
			if !allowed {
				if action == governance.ActionRequireApproval {
					// Publish to Governance Request Channel for User Team
					// Subject: swarm.governance.request
					// Payload: ApprovalRequest struct (not impl here fully, just mock)
					log.Printf("📢 Published Approval Request %s", reqID)
				}
				return // STOP PROCESSING
			}
		}

		// C. CORE LOGIC (Heartbeats, Registry)
		// Only process heartbeats if the subject matches or if it's an event
		// Current logic filters by internal checks

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
		// Ideally we only update registry on heartbeat events, but specifically filtering by 'heartbeat' in subject was previous logic.
		// We should check intent or event type.
		isHeartbeat := strings.Contains(msg.Subject, "heartbeat") || (envelope.GetEvent() != nil && envelope.GetEvent().EventType == "agent.heartbeat")

		if isHeartbeat {
			state.GlobalRegistry.UpdateHeartbeat(agentID, envelope.TeamId, sourceURI, state.StatusIdle)
		}
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

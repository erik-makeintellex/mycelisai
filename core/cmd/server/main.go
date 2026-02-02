package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/server"
	mycelis_nats "github.com/mycelis/core/internal/transport/nats"
)

func main() {
	log.Println("Starting Mycelis Core [Brain]...")

	// 1. Config
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL // nats://localhost:4222
	}

	// 1a. Load Cognitive Brain
	brainPath := "core/config/brain.yaml"
	cogRouter, err := cognitive.NewRouter(brainPath)
	if err != nil {
		log.Printf("⚠️ Brain Config not loaded: %v. Cognitive Matrix Disabled.", err)
		cogRouter = nil
	} else {
		log.Println("🧠 Cognitive Matrix Active.")
	}

	// 1b. Load Governance Policy
	policyPath := "core/config/policy.yaml" // Assuming we run from root or handle path
	// Fix path for container vs local? For now, relative.

	gk, err := governance.NewGatekeeper(policyPath)
	if err != nil {
		log.Printf("⚠️ Governance Policy not loaded: %v. Allowing all.", err)
		gk = nil
	} else {
		log.Println("🛡️ Gatekeeper Active.")
	}

	// 1c. Load Archivist (Memory)
	// Default to values.yaml if env missing (but security context might block raw strings?)
	// Using Env.
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "mycelis-core-postgresql"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		"mycelis",  // User (pinned in values.yaml)
		"password", // Password (pinned in values.yaml)
		dbHost,
		"5432",
		"cortex",
	)

	archivist, err := memory.NewArchivist(dbURL)
	if err != nil {
		log.Printf("⚠️ Memory System Failed: %v. Continuing without persistence.", err)
		archivist = nil
	} else {
		log.Println("🧠 Hippocampus Connected.")
		go archivist.Start(context.Background())
	}

	// 2. Connect to NATS
	var ncWrapper *mycelis_nats.Client
	var connErr error

	// Simple retry loop
	for i := 0; i < 10; i++ {
		log.Printf("Connecting to NATS at %s (Attempt %d/10)...", natsURL, i+1)
		ncWrapper, connErr = mycelis_nats.Connect(natsURL)
		if connErr == nil {
			break
		}
		log.Printf("NATS connection failed: %v. Retrying in 2s...", connErr)
		time.Sleep(2 * time.Second)
	}

	if connErr != nil {
		log.Fatalf("Failed to connect to NATS after retries: %v", connErr)
	}
	defer ncWrapper.Drain()

	nc := ncWrapper.Conn

	// 3. Start Router
	r := router.NewRouter(nc, gk)
	if err := r.Start(); err != nil {
		log.Fatalf("Failed to start Router: %v", err)
	}

	// 4. Start Memory Subscriber (The Archivist)
	if archivist != nil {
		_, err := nc.Subscribe("swarm.>", func(msg *nats.Msg) {
			var envelope pb.MsgEnvelope
			if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
				// Ignore malformed for now
				return
			}

			// Map Envelope to LogEntry
			// We use default Go timestamp in Archivist if generic, or extract from Envelope?
			// Envelope has timestamp *timestamppb.Timestamp.
			ts := time.Now()
			if envelope.Timestamp != nil {
				ts = envelope.Timestamp.AsTime()
			}

			// Extract Message Content
			msgBody := ""
			intent := "event"
			level := "INFO"
			ctxMap := make(map[string]any) // Should be generic map

			switch p := envelope.Payload.(type) {
			case *pb.MsgEnvelope_Text:
				msgBody = p.Text.Content
				intent = p.Text.Intent
				if intent == "" {
					intent = "text"
				}
			case *pb.MsgEnvelope_Event:
				msgBody = fmt.Sprintf("Event: %s", p.Event.EventType)
				intent = p.Event.EventType
			case *pb.MsgEnvelope_ToolCall:
				msgBody = fmt.Sprintf("Tool Call: %s", p.ToolCall.ToolName)
				intent = "tool_call"
			case *pb.MsgEnvelope_ToolResult:
				msgBody = fmt.Sprintf("Tool Result: %s", p.ToolResult.CallId)
				if p.ToolResult.IsError {
					level = "ERROR"
					msgBody = fmt.Sprintf("Error: %s", p.ToolResult.ErrorMessage)
				}
				intent = "tool_result"
			}

			logEntry := &memory.LogEntry{
				TraceId:   envelope.TraceId,
				Timestamp: ts,
				Source:    envelope.SourceAgentId,
				Intent:    intent,
				Message:   msgBody,
				Level:     level,
				Context:   ctxMap,
			}

			archivist.Push(logEntry)

		})
		if err != nil {
			log.Printf("Failed to subscribe Archivist: %v", err)
		}
	}

	// 5. Http Server
	mux := http.NewServeMux()

	// Create Admin Server
	adminSrv := server.NewAdminServer(r, gk, archivist, cogRouter)
	adminSrv.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("HTTP Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("HTTP Server failed: %v", err)
	}
}

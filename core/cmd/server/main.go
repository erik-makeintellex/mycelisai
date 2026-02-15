package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx driver
	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/provisioning"
	"github.com/mycelis/core/internal/registry"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/server"
	"github.com/mycelis/core/internal/swarm"
	mycelis_nats "github.com/mycelis/core/internal/transport/nats"
)

func main() {
	log.Println("Starting Mycelis Core [System]...")

	// 1. Config
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL // nats://localhost:4222
	}

	// 1a. Initialize Database (Shared)
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

	// Open Shared DB Connection
	sharedDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Printf("WARN: Shared DB Init Failed: %v", err)
	} else if err := sharedDB.Ping(); err != nil {
		log.Printf("WARN: Shared DB Ping Failed: %v", err)
	} else {
		log.Println("Shared Database Connection Active.")
	}

	// 1b. Load Cognitive Engine
	cogPath := "core/config/cognitive.yaml"
	cogRouter, err := cognitive.NewRouter(cogPath, sharedDB)
	if err != nil {
		log.Printf("WARN: Cognitive Config not loaded: %v. Cognitive Engine Disabled.", err)
		cogRouter = nil
	} else {
		log.Println("Cognitive Engine Active.")
	}

	// 1b. Load Governance Policy
	policyPath := "core/config/policy.yaml" // Assuming we run from root or handle path

	guard, err := governance.NewGuard(policyPath)
	if err != nil {
		log.Printf("WARN: Governance Policy not loaded: %v. Allowing all.", err)
		guard = nil
	} else {
		log.Println("Governance Guard Active.")
	}

	// 1c. Load Memory Service (The Cortex Memory)
	memService, err := memory.NewService(dbURL)
	if err != nil {
		log.Printf("WARN: Memory System Failed: %v. Continuing without persistence.", err)
		memService = nil
	} else {
		log.Println("Cortex Memory Connected.")
		go memService.Start(context.Background())
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
	r := router.NewRouter(nc, guard)
	if err := r.Start(); err != nil {
		log.Fatalf("Failed to start Router: %v", err)
	}

	// 4. Start Memory Subscriber (The Memory Service)
	if memService != nil {
		_, err := nc.Subscribe("swarm.>", func(msg *nats.Msg) {
			var envelope pb.MsgEnvelope
			if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
				// Ignore malformed for now
				return
			}

			// Map Envelope to LogEntry
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

			memService.Push(logEntry)

		})
		if err != nil {
			log.Printf("Failed to subscribe Memory Service: %v", err)
		}
	}

	// 5. Http Server
	mux := http.NewServeMux()

	// 1d. Initialize Archivist (Intelligence)
	var archivist *memory.Archivist
	if memService != nil && cogRouter != nil {
		archivist = memory.NewArchivist(memService, cogRouter)
		log.Println("Archivist Engine Active.")
	}

	// 5a. Initialize Registry Service (Reuse shared DB)
	var regService *registry.Service
	if sharedDB != nil {
		regService = registry.NewService(sharedDB)
		log.Println("Registry Service Active.")
	}

	// 5b. Initialize Provisioning Engine
	provEngine := provisioning.NewEngine(cogRouter)

	// 5c. Initialize Bootstrap Service
	bootstrapSrv := bootstrap.NewService(sharedDB, nc)
	bootstrapSrv.Start()

	// 5d. Initialize Swarm Intelligence (Soma)
	// Guard is already initialized as 'guard'
	// Registry path?
	teamConfigPath := "core/config/teams"
	swarmReg := swarm.NewRegistry(teamConfigPath)
	soma := swarm.NewSoma(nc, guard, swarmReg, cogRouter)
	if err := soma.Start(); err != nil {
		log.Fatalf("Failed to start Soma: %v", err)
	}

	// Routes
	mux.HandleFunc("/api/v1/nodes/pending", bootstrapSrv.HandlePendingNodes)

	// Archivist Routes
	if archivist != nil {
		mux.HandleFunc("/api/v1/memory/sitrep", func(w http.ResponseWriter, r *http.Request) {
			// POST to generate, GET to list?
			// Simple trigger for now
			// team_id param, duration param
			// Defaults: Mycelis Core, 24h
			if r.Method == "POST" {
				err := archivist.GenerateSitRep(r.Context(), "22222222-2222-2222-2222-222222222222", 24*time.Hour)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				w.Write([]byte(`{"status": "SitRep Generated"}`))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
	}

	// Create Admin Server
	// Note: We pass nil for unused services if they failed init, which is handled inside AdminServer hopefully.
	adminSrv := server.NewAdminServer(r, guard, memService, cogRouter, provEngine, regService)
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

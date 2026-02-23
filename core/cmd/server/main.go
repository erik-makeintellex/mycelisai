package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	os_signal "os/signal"
	"time"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx driver
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/events"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/overseer"
	"github.com/mycelis/core/internal/provisioning"
	"github.com/mycelis/core/internal/registry"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/server"
	mycelis_signal "github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/internal/swarm"
	mycelis_nats "github.com/mycelis/core/internal/transport/nats"
)

func main() {
	log.Println("Starting Mycelis Core [System]...")

	// Root context with signal-based cancellation for graceful shutdown
	ctx, stop := os_signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Phase 0 security: fail-closed — refuse to start without auth key
	apiKey := os.Getenv("MYCELIS_API_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: MYCELIS_API_KEY not set. Server refuses to start without authentication.")
	}

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
	cogPath := "config/cognitive.yaml"
	cogRouter, err := cognitive.NewRouter(cogPath, sharedDB)
	if err != nil {
		log.Printf("WARN: Cognitive Config not loaded: %v. Cognitive Engine Disabled.", err)
		cogRouter = nil
	} else {
		log.Println("Cognitive Engine Active.")
	}

	// 1b. Load Governance Policy
	policyPath := "config/policy.yaml"

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
		go memService.Start(ctx)
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

	// Graceful degradation: if NATS is unreachable, run in degraded mode
	var nc *nats.Conn
	if connErr != nil {
		log.Printf("WARN: NATS unreachable after retries: %v. Running in DEGRADED mode (no messaging).", connErr)
	} else {
		defer ncWrapper.Drain()
		nc = ncWrapper.Conn
	}

	// 3. Start Router (requires NATS)
	var r *router.Router
	if nc != nil {
		r = router.NewRouter(nc, guard)
		if err := r.Start(); err != nil {
			log.Printf("WARN: Failed to start Router: %v", err)
		}
	} else {
		log.Println("WARN: Router disabled (no NATS connection).")
	}

	// 4. Start Memory Subscriber (The Memory Service — requires NATS)
	if memService != nil && nc != nil {
		_, err := nc.Subscribe(protocol.TopicSwarmWild, func(msg *nats.Msg) {
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

	// Phase 7.0: MCP Ingress
	var mcpService *mcp.Service
	var mcpPool *mcp.ClientPool
	if sharedDB != nil {
		mcpService = mcp.NewService(sharedDB)
		mcpPool = mcp.NewClientPool(mcpService)
		// Reconnect persisted servers on startup (best-effort)
		if servers, err := mcpService.List(ctx); err == nil {
			mcpPool.ReconnectAll(ctx, servers)
		} else {
			log.Printf("WARN: Failed to list MCP servers for reconnect: %v", err)
		}
		log.Println("MCP Ingress Active.")
	}

	// Phase 7.5: Agent Catalogue
	var catService *catalogue.Service
	if sharedDB != nil {
		catService = catalogue.NewService(sharedDB)
		log.Println("Agent Catalogue Active.")
	}

	// Phase 7.5: Artifacts (Agent Output Persistence)
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data/artifacts"
	}
	var artService *artifacts.Service
	if sharedDB != nil {
		artService = artifacts.NewService(sharedDB, dataDir)
		log.Println("Artifacts Service Active.")
	}

	// 5b. Initialize Provisioning Engine
	provEngine := provisioning.NewEngine(cogRouter)

	// 5c. Initialize Bootstrap Service (requires NATS)
	var bootstrapSrv *bootstrap.Service
	if nc != nil {
		bootstrapSrv = bootstrap.NewService(sharedDB, nc)
		bootstrapSrv.Start()
	} else {
		log.Println("WARN: Bootstrap Service disabled (no NATS connection).")
	}

	// Initialize Signal Stream (SSE) — always created (no NATS dependency for handler itself)
	streamHandler := mycelis_signal.NewStreamHandler()

	// Create Meta-Architect (may be nil if cogRouter is nil)
	var metaArchitect *cognitive.MetaArchitect
	if cogRouter != nil {
		metaArchitect = cognitive.NewMetaArchitect(cogRouter)
		log.Println("Meta-Architect Active.")
	}

	// 5d. Initialize Swarm Intelligence (Soma — requires NATS)
	// Build MCP tool executor adapter for agent ReAct loop
	var toolExec swarm.MCPToolExecutor
	if mcpService != nil && mcpPool != nil {
		toolExec = mcp.NewToolExecutorAdapter(mcpService, mcpPool)
	}

	// Build Internal Tool Registry (built-in tools for agents)
	// Hoisted outside NATS block so MetaArchitect can read tool descriptions even in degraded mode.
	var internalTools *swarm.InternalToolRegistry
	if nc != nil {
		internalTools = swarm.NewInternalToolRegistry(swarm.InternalToolDeps{
			NC:        nc,
			Brain:     cogRouter,
			Mem:       memService,
			Architect: metaArchitect,
			Catalogue: catService,
			DB:        sharedDB,
		})
	}

	// V7 Event Spine (Team A): initialize event store + run manager before Soma
	var eventStore *events.Store
	var runsManager *runs.Manager
	if sharedDB != nil {
		runsManager = runs.NewManager(sharedDB)
		eventStore = events.NewStore(sharedDB, nc) // nc may be nil (degraded mode ok)
		log.Println("V7 Event Spine Active. (runs + events stores ready)")
	}

	var soma *swarm.Soma
	if nc != nil {
		teamConfigPath := "config/teams"
		swarmReg := swarm.NewRegistry(teamConfigPath)

		soma = swarm.NewSoma(nc, guard, swarmReg, cogRouter, streamHandler, toolExec, internalTools)
		// V7: wire event spine into Soma so ActivateBlueprint creates runs + emits events
		if runsManager != nil {
			soma.SetRunsManager(runsManager)
		}
		if eventStore != nil {
			soma.SetEventEmitter(eventStore)
		}
		if err := soma.Start(); err != nil {
			log.Printf("WARN: Failed to start Soma: %v", err)
		}
		// 5e. Swarm Routes
		mux.HandleFunc("/api/swarm/teams", soma.HandleCreateTeam)
		mux.HandleFunc("/api/swarm/command", soma.HandleCommand)
		mux.HandleFunc("/api/v1/swarm/broadcast", soma.HandleBroadcast)
	} else {
		log.Println("WARN: Soma (Swarm) disabled (no NATS connection).")
	}

	// Routes (Bootstrap — guarded)
	if bootstrapSrv != nil {
		mux.HandleFunc("/api/v1/nodes/pending", bootstrapSrv.HandlePendingNodes)
	}

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

	// Start Archivist Background Loop (DB-based periodic SitReps)
	if archivist != nil {
		go archivist.StartLoop(ctx, 5*time.Minute, "22222222-2222-2222-2222-222222222222")
	}

	// Start Archivist Daemon (NATS-based sliding window buffer — requires NATS)
	if archivist != nil && nc != nil {
		go archivist.StartDaemon(ctx, nc, "22222222-2222-2222-2222-222222222222")
	}

	// Phase 5.2: Initialize Overseer (Trust Economy Governance Valve)
	var overseerEngine *overseer.Engine
	if nc != nil {
		overseerEngine = overseer.NewEngine(nc)

		// Wire governance callback: route halted envelopes to SSE stream
		// for Zone D (Deliverables Tray) human review
		overseerEngine.SetGovernanceCallback(func(env *protocol.CTSEnvelope) {
			if streamHandler != nil {
				data, _ := json.Marshal(map[string]interface{}{
					"type":        "governance_halt",
					"source":      env.Meta.SourceNode,
					"trust_score": env.TrustScore,
					"timestamp":   env.Meta.Timestamp.Format(time.RFC3339),
					"trace_id":    env.Meta.TraceID,
				})
				streamHandler.Broadcast(string(data))
			}
		})

		if err := overseerEngine.Start(); err != nil {
			log.Printf("WARN: Overseer failed to start: %v", err)
		} else {
			log.Println("Overseer Engine Online. Trust Economy active.")
		}
	} else {
		log.Println("WARN: Overseer disabled (no NATS connection).")
	}

	// Phase 7.7: Load MCP Library (curated server registry)
	var mcpLibrary *mcp.Library
	mcpLibPath := "config/mcp-library.yaml"
	if lib, err := mcp.LoadLibrary(mcpLibPath); err != nil {
		log.Printf("WARN: MCP Library not loaded: %v", err)
	} else {
		mcpLibrary = lib
		log.Println("MCP Library Active.")
	}

	// Phase 18: Bootstrap default MCP servers (filesystem, fetch) on first run
	if mcpService != nil && mcpPool != nil && mcpLibrary != nil {
		mcpService.BootstrapDefaults(ctx, mcpLibrary, mcpPool)
	}

	// Wire system capabilities into Meta-Architect (tools, MCP servers, library)
	if metaArchitect != nil {
		caps := buildSystemCapabilities(ctx, internalTools, mcpService, mcpLibrary)
		metaArchitect.SetCapabilities(caps)
		log.Println("Meta-Architect capabilities wired.")
	}

	// Create Admin Server (V7: pass eventStore + runsManager for Event Spine routes)
	adminSrv := server.NewAdminServer(r, guard, memService, cogRouter, provEngine, regService, soma, nc, streamHandler, metaArchitect, overseerEngine, archivist, mcpService, mcpPool, mcpLibrary, catService, artService, eventStore, runsManager)
	adminSrv.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// CORS Middleware
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	// Phase 0: auth middleware wraps the mux, CORS wraps the authed mux
	authedMux := server.AuthMiddleware(apiKey, mux)
	corsMux := http.HandlerFunc(func(w http.ResponseWriter, hr *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if hr.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		authedMux.ServeHTTP(w, hr)
	})

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutdown signal received. Draining...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if soma != nil {
			soma.Shutdown()
		}

		if mcpPool != nil {
			mcpPool.ShutdownAll()
		}

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}
	}()

	log.Printf("HTTP Server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP Server failed: %v", err)
	}

	log.Println("Mycelis Core shutdown complete.")
}

// buildSystemCapabilities assembles the live system context that the
// Meta-Architect uses to generate resource-aware mission blueprints.
func buildSystemCapabilities(ctx context.Context, tools *swarm.InternalToolRegistry, mcpSvc *mcp.Service, lib *mcp.Library) *cognitive.SystemCapabilities {
	caps := &cognitive.SystemCapabilities{}

	// 1. Internal tool descriptions
	if tools != nil {
		caps.InternalTools = tools.ListDescriptions()
	}

	// 2. Installed MCP servers + their discovered tools
	if mcpSvc != nil {
		servers, err := mcpSvc.List(ctx)
		if err == nil {
			for _, srv := range servers {
				toolNames := []string{}
				if defs, err := mcpSvc.ListTools(ctx, srv.ID); err == nil {
					for _, t := range defs {
						toolNames = append(toolNames, t.Name)
					}
				}
				caps.MCPServers = append(caps.MCPServers, cognitive.MCPServerCapability{
					Name:   srv.Name,
					Status: "installed",
					Tools:  toolNames,
				})
			}
		}
	}

	// 3. Library entries (available for install)
	if lib != nil {
		for _, cat := range lib.Categories {
			for _, entry := range cat.Servers {
				reqEnv := make([]string, 0, len(entry.Env))
				for k := range entry.Env {
					reqEnv = append(reqEnv, k)
				}
				caps.MCPServers = append(caps.MCPServers, cognitive.MCPServerCapability{
					Name:        entry.Name,
					Description: entry.Description,
					Status:      "available",
					RequiredEnv: reqEnv,
				})
			}
		}
	}

	return caps
}

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/events"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/overseer"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/server"
	mycelisSignal "github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/internal/triggers"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

func startSomaRuntime(
	ctx context.Context,
	mux *http.ServeMux,
	core *coreRuntime,
	selection *bootstrap.StartupSelection,
	registry *swarm.Registry,
	services productServices,
) *swarm.Soma {
	if core.NC == nil {
		log.Println("WARN: Soma (Swarm) disabled (no NATS connection).")
		return nil
	}

	log.Printf("Soma startup instantiating runtime organization from bootstrap template bundle %s", selection.Bundle.ID)
	soma := swarm.NewSoma(core.NC, core.Guard, registry, core.CogRouter, services.Stream, services.ToolExecutor, services.InternalTools)
	startupRouting := resolveStartupProviderRouting(
		selection,
		os.Getenv("MYCELIS_TEAM_PROVIDER_MAP"),
		os.Getenv("MYCELIS_AGENT_PROVIDER_MAP"),
	)
	if !startupRouting.Policy.IsEmpty() {
		soma.SetProviderPolicy(startupRouting.Policy)
		log.Printf("Provider routing active from runtime organization policy: teams=%d agents=%d", len(startupRouting.Policy.Teams), len(startupRouting.Policy.Agents))
	}
	if startupRouting.IgnoredLegacyEnvMaps {
		log.Printf("WARN: ignoring MYCELIS_TEAM_PROVIDER_MAP / MYCELIS_AGENT_PROVIDER_MAP; provider routing now comes only from the instantiated organization and the env-map compatibility path is retired")
	}
	if services.RunsManager != nil {
		soma.SetRunsManager(services.RunsManager)
	}
	if services.EventStore != nil {
		soma.SetEventEmitter(services.EventStore)
	}
	if services.ConversationLog != nil {
		soma.SetConversationLogger(services.ConversationLog)
	}
	wireSomaMCPDescriptions(ctx, soma, services.MCP)
	if err := soma.Start(); err != nil {
		log.Printf("WARN: Failed to start Soma: %v", err)
	}
	mux.HandleFunc("/api/swarm/teams", soma.HandleCreateTeam)
	mux.HandleFunc("/api/swarm/command", soma.HandleCommand)
	mux.HandleFunc("/api/v1/swarm/broadcast", soma.HandleBroadcast)
	return soma
}

func wireSomaMCPDescriptions(ctx context.Context, soma *swarm.Soma, mcpService *mcp.Service) {
	if soma == nil || mcpService == nil {
		return
	}
	servers, err := mcpService.List(ctx)
	if err != nil {
		return
	}
	serverNames := make(map[uuid.UUID]string, len(servers))
	toolDescs := make(map[string]string)
	for _, srv := range servers {
		serverNames[srv.ID] = srv.Name
		if tools, err := mcpService.ListTools(ctx, srv.ID); err == nil {
			for _, t := range tools {
				toolDescs[t.Name] = t.Description
			}
		}
	}
	soma.SetMCPServerNames(serverNames)
	soma.SetMCPToolDescs(toolDescs)
}

func registerBootstrapRoutes(mux *http.ServeMux, bootstrapSrv *bootstrap.Service) {
	if bootstrapSrv != nil {
		mux.HandleFunc("/api/v1/nodes/pending", bootstrapSrv.HandlePendingNodes)
	}
}

func startArchivistRuntime(ctx context.Context, mux *http.ServeMux, observerNC *nats.Conn, archivist *memory.Archivist) {
	if archivist == nil {
		return
	}
	mux.HandleFunc("/api/v1/memory/sitrep", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := archivist.GenerateSitRep(r.Context(), "22222222-2222-2222-2222-222222222222", 24*time.Hour); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status": "SitRep Generated"}`))
	})
	go archivist.StartLoop(ctx, 5*time.Minute, "22222222-2222-2222-2222-222222222222")
	if observerNC != nil {
		go archivist.StartDaemon(ctx, observerNC, "22222222-2222-2222-2222-222222222222")
	}
}

func startOverseerEngine(nc *nats.Conn, streamHandler *mycelisSignal.StreamHandler) *overseer.Engine {
	if nc == nil {
		log.Println("WARN: Overseer disabled (no NATS connection).")
		return nil
	}
	overseerEngine := overseer.NewEngine(nc)
	overseerEngine.SetGovernanceCallback(func(env *protocol.CTSEnvelope) {
		if streamHandler == nil {
			return
		}
		data, _ := json.Marshal(map[string]interface{}{
			"type":        "governance_halt",
			"source":      env.Meta.SourceNode,
			"trust_score": env.TrustScore,
			"timestamp":   env.Meta.Timestamp.Format(time.RFC3339),
			"trace_id":    env.Meta.TraceID,
		})
		streamHandler.Broadcast(string(data))
	})
	if err := overseerEngine.Start(); err != nil {
		log.Printf("WARN: Overseer failed to start: %v", err)
	} else {
		log.Println("Overseer Engine Online. Trust Economy active.")
	}
	return overseerEngine
}

func loadMCPLibrary(ctx context.Context, mcpService *mcp.Service, mcpPool *mcp.ClientPool) *mcp.Library {
	lib, err := mcp.LoadLibrary("config/mcp-library.yaml")
	if err != nil {
		log.Printf("WARN: MCP Library not loaded: %v", err)
		return nil
	}
	log.Println("MCP Library Active.")
	if mcpService != nil && mcpPool != nil {
		if defaultMCPBootstrapEnabled() {
			mcpService.BootstrapDefaults(ctx, lib, mcpPool)
		} else {
			log.Println("MCP bootstrap: default server auto-install disabled by runtime environment.")
		}
	}
	return lib
}

func wireAdminServices(ctx context.Context, mux *http.ServeMux, core *coreRuntime, adminSrv *server.AdminServer, services productServices) {
	adminSrv.Comms = services.Comms
	adminSrv.Search = services.Search
	adminSrv.Conversations = services.ConversationLog
	adminSrv.Inception = services.Inception
	adminSrv.MCPToolSets = services.MCPToolSets
	adminSrv.RegisterRoutes(mux)
	adminSrv.StartLoopScheduler(ctx)
	startTriggerEngine(ctx, core.SharedDB, core.NC, adminSrv, services.EventStore, services.RunsManager)
	startReactiveEngine(ctx, core.SharedDB, adminSrv)
}

func startTriggerEngine(ctx context.Context, sharedDB *sql.DB, nc *nats.Conn, adminSrv *server.AdminServer, eventStore *events.Store, runsManager *runs.Manager) {
	if sharedDB == nil || eventStore == nil || runsManager == nil {
		log.Println("WARN: Trigger Engine disabled (missing DB, events, or runs).")
		return
	}
	triggerStore := triggers.NewStore(sharedDB)
	triggerEngine := triggers.NewEngine(triggerStore, eventStore, runsManager, nc)
	adminSrv.Triggers = triggerStore
	adminSrv.TriggerEngine = triggerEngine
	go func() {
		if err := triggerEngine.Start(ctx); err != nil {
			log.Printf("[triggers] engine start error: %v", err)
		}
	}()
	log.Println("V7 Trigger Engine Active.")
}

func startReactiveEngine(ctx context.Context, sharedDB *sql.DB, adminSrv *server.AdminServer) {
	if adminSrv.Reactive == nil || sharedDB == nil {
		return
	}
	adminSrv.Reactive.SetDB(sharedDB)
	go func() {
		if err := adminSrv.Reactive.ReactivateFromDB(ctx); err != nil {
			log.Printf("[reactive] startup reactivation: %v", err)
		} else {
			log.Println("[reactive] active profile subscriptions restored.")
		}
	}()
}

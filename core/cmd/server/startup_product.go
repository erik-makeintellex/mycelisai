package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/internal/events"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/provisioning"
	"github.com/mycelis/core/internal/registry"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/internal/server"
	mycelisSignal "github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/internal/swarm"
)

type productServices struct {
	Archivist       *memory.Archivist
	Registry        *registry.Service
	MCP             *mcp.Service
	MCPPool         *mcp.ClientPool
	MCPToolSets     *mcp.ToolSetService
	Catalogue       *catalogue.Service
	Artifacts       *artifacts.Service
	Exchange        *exchange.Service
	Provisioning    *provisioning.Engine
	Bootstrap       *bootstrap.Service
	Stream          *mycelisSignal.StreamHandler
	MetaArchitect   *cognitive.MetaArchitect
	ToolExecutor    swarm.MCPToolExecutor
	Inception       *inception.Store
	Comms           *comms.Gateway
	Search          *searchcap.Service
	InternalTools   *swarm.InternalToolRegistry
	EventStore      *events.Store
	RunsManager     *runs.Manager
	ConversationLog *conversations.Store
}

func startProductRuntime(ctx context.Context, mux *http.ServeMux, core *coreRuntime) *productRuntime {
	if core == nil {
		core = &coreRuntime{}
	}

	selection, registry := loadStartupRuntimeSelection()
	services := startProductServices(ctx, core)
	soma := startSomaRuntime(ctx, mux, core, selection, registry, services)
	registerBootstrapRoutes(mux, services.Bootstrap)
	startArchivistRuntime(ctx, mux, core.ObserverNC, services.Archivist)
	overseerEngine := startOverseerEngine(core.NC, services.Stream)

	mcpLibrary := loadMCPLibrary(ctx, services.MCP, services.MCPPool)
	if services.MetaArchitect != nil {
		caps := buildSystemCapabilities(ctx, services.InternalTools, services.MCP, mcpLibrary)
		services.MetaArchitect.SetCapabilities(caps)
		log.Println("Meta-Architect capabilities wired.")
	}

	adminSrv := server.NewAdminServer(
		core.Router,
		core.Guard,
		core.MemService,
		core.SharedDB,
		core.CogRouter,
		services.Provisioning,
		services.Registry,
		soma,
		core.NC,
		services.Stream,
		services.MetaArchitect,
		overseerEngine,
		services.Archivist,
		services.MCP,
		services.MCPPool,
		mcpLibrary,
		services.Catalogue,
		services.Artifacts,
		services.Exchange,
		services.EventStore,
		services.RunsManager,
	)
	wireAdminServices(ctx, mux, core, adminSrv, services)

	return &productRuntime{Admin: adminSrv, Soma: soma, MCPPool: services.MCPPool}
}

func loadStartupRuntimeSelection() (*bootstrap.StartupSelection, *swarm.Registry) {
	selection, registry, err := loadStartupBundleRegistry("config/templates")
	if err != nil {
		log.Fatalf("FATAL: Bootstrap startup selection failed: %v", err)
	}
	if selection.Bundle == nil || selection.Organization == nil {
		log.Fatal("FATAL: Bootstrap startup selection did not return a bundle-backed runtime organization")
	}
	log.Printf("Bootstrap Template Bundle Active: %s", selection.Bundle.ID)
	log.Printf("Bootstrap Runtime Organization Active: %s", selection.Organization.ID)
	return selection, registry
}

func startProductServices(ctx context.Context, core *coreRuntime) productServices {
	sharedDB := core.SharedDB
	cogRouter := core.CogRouter
	memService := core.MemService
	services := productServices{
		Provisioning: provisioning.NewEngine(cogRouter),
		Stream:       mycelisSignal.NewStreamHandler(),
		Comms:        comms.NewGatewayFromEnv(),
		Search:       searchcap.NewService(searchcap.ConfigFromEnv(), cogRouter, memService),
	}
	if memService != nil && cogRouter != nil {
		services.Archivist = memory.NewArchivist(memService, cogRouter)
		log.Println("Archivist Engine Active.")
	}
	if sharedDB != nil {
		services.Registry = registry.NewService(sharedDB)
		services.Catalogue = catalogue.NewService(sharedDB)
		services.Inception = inception.NewStore(sharedDB)
		services.RunsManager = runs.NewManager(sharedDB)
		services.EventStore = events.NewStore(sharedDB, core.NC)
		services.ConversationLog = conversations.NewStore(sharedDB)
		log.Println("Registry Service Active.")
		log.Println("Agent Catalogue Active.")
		log.Println("V7 Inception Recipe Store Active.")
		log.Println("V7 Event Spine Active. (runs + events stores ready)")
		log.Println("V7 Conversation Store Active.")
		services.MCP, services.MCPPool, services.MCPToolSets = startMCPRuntime(ctx, sharedDB)
		services.Artifacts = startArtifactRuntime(ctx, sharedDB)
		services.Exchange = startExchangeRuntime(ctx, sharedDB, cogRouter, memService)
	}
	if cogRouter != nil {
		services.MetaArchitect = cognitive.NewMetaArchitect(cogRouter)
		log.Println("Meta-Architect Active.")
	}
	if services.MCP != nil && services.MCPPool != nil {
		services.ToolExecutor = mcp.NewToolExecutorAdapter(services.MCP, services.MCPPool)
	}
	if providers := services.Comms.ListProviders(); len(providers) > 0 {
		ready := 0
		for _, p := range providers {
			if p.Configured {
				ready++
			}
		}
		log.Printf("Communications Gateway Active. %d/%d providers configured.", ready, len(providers))
	}
	log.Printf("Mycelis Search capability provider: %s", services.Search.Provider())
	if core.NC != nil {
		services.InternalTools = swarm.NewInternalToolRegistry(swarm.InternalToolDeps{
			NC:        core.NC,
			Brain:     cogRouter,
			Mem:       memService,
			Architect: services.MetaArchitect,
			Catalogue: services.Catalogue,
			Inception: services.Inception,
			Comms:     services.Comms,
			DB:        sharedDB,
			Exchange:  services.Exchange,
			Search:    services.Search,
		})
	}
	if core.NC != nil {
		services.Bootstrap = bootstrap.NewService(sharedDB, core.NC)
		services.Bootstrap.Start()
	} else {
		log.Println("WARN: Bootstrap Service disabled (no NATS connection).")
	}
	return services
}

func startMCPRuntime(ctx context.Context, sharedDB *sql.DB) (*mcp.Service, *mcp.ClientPool, *mcp.ToolSetService) {
	mcpService := mcp.NewService(sharedDB)
	mcpToolSets := mcp.NewToolSetService(sharedDB)
	mcpService.ToolSets = mcpToolSets
	mcpPool := mcp.NewClientPool(mcpService)
	if servers, err := mcpService.List(ctx); err == nil {
		mcpPool.ReconnectAll(ctx, servers)
	} else {
		log.Printf("WARN: Failed to list MCP servers for reconnect: %v", err)
	}
	log.Println("MCP Ingress Active.")
	return mcpService, mcpPool, mcpToolSets
}

func startArtifactRuntime(ctx context.Context, sharedDB *sql.DB) *artifacts.Service {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data/artifacts"
	}
	if err := ensureStorageLayout(resolveWorkspaceRoot(), dataDir); err != nil {
		log.Printf("WARN: Failed to prepare storage layout: %v", err)
	}
	artService := artifacts.NewService(sharedDB, dataDir)
	log.Println("Artifacts Service Active.")
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n, err := artService.DeleteExpiredCachedImages(ctx, time.Hour)
				if err != nil {
					log.Printf("WARN: image cache cleanup failed: %v", err)
				} else if n > 0 {
					log.Printf("Image cache cleanup: removed %d expired unsaved image artifact(s).", n)
				}
			}
		}
	}()
	return artService
}

func startExchangeRuntime(ctx context.Context, sharedDB *sql.DB, cogRouter *cognitive.Router, memService *memory.Service) *exchange.Service {
	var embedFn exchange.EmbedFunc
	if cogRouter != nil {
		embedFn = func(ctx context.Context, content string) ([]float64, error) {
			return cogRouter.Embed(ctx, content, "")
		}
	}
	exchangeService := exchange.NewService(sharedDB, embedFn, memService)
	if err := exchangeService.Bootstrap(ctx); err != nil {
		log.Printf("WARN: Managed Exchange bootstrap failed: %v", err)
	} else {
		log.Println("Managed Exchange Active.")
	}
	return exchangeService
}

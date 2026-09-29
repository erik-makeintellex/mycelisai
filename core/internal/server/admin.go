package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/capabilities"
	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/codecontext"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/events"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/inputs"
	"github.com/mycelis/core/internal/invocation"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/outputvalidation"
	"github.com/mycelis/core/internal/provisioning"
	"github.com/mycelis/core/internal/reactive"
	"github.com/mycelis/core/internal/registry"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/internal/triggers"
	"github.com/mycelis/core/internal/workerauthority"
	"github.com/mycelis/core/internal/workers"
	"github.com/nats-io/nats.go"
)

// AdminServer handles governance and system endpoints
type AdminServer struct {
	Router        *router.Router
	Guard         *governance.Guard
	Mem           *memory.Service
	DB            *sql.DB // direct DB for context snapshots + mission profiles
	Cognitive     *cognitive.Router
	Provisioner   *provisioning.Engine
	Registry      *registry.Service
	Soma          *swarm.Soma
	NC            *nats.Conn // NATS for chat request-reply routing
	Stream        *signal.StreamHandler
	MetaArchitect *cognitive.MetaArchitect
	Archivist     *memory.Archivist    // Phase 5.3: RAG Persistence
	Proposals     *ProposalStore       // Phase 5.3: Team Manifestation
	MCP           *mcp.Service         // Phase 7.0: MCP Ingress
	MCPPool       *mcp.ClientPool      // Phase 7.0: MCP Ingress
	MCPLibrary    *mcp.Library         // Phase 7.7: Curated MCP Library
	Catalogue     *catalogue.Service   // Phase 7.5: Agent Catalogue
	Artifacts     *artifacts.Service   // Phase 7.5: Agent Outputs
	Exchange      *exchange.Service    // Managed exchange channels, threads, and artifacts
	Comms         *comms.Gateway       // External communication providers (whatsapp/telegram/slack/etc.)
	Events        *events.Store        // V7: persistent mission event audit trail
	Runs          *runs.Manager        // V7: mission run lifecycle management
	Reactive      *reactive.Engine     // watches NATS topics for active profiles
	Triggers      *triggers.Store      // trigger rule CRUD + in-memory cache
	TriggerEngine *triggers.Engine     // evaluates rules against CTS events
	Conversations *conversations.Store // full-fidelity agent conversation turns
	Inception     *inception.Store     // inception recipe CRUD + search
	MCPToolSets   *mcp.ToolSetService  // tool set CRUD
	// Root-admin collaboration groups (DB-backed), with live bus monitor for status UI.
	GroupBus *GroupBusMonitor
	// V8 AI Organization entry flow support.
	Organizations                  *OrganizationStore
	LoopProfiles                   *LoopProfileStore
	LoopResults                    *LoopResultStore
	LoopExecution                  *LoopExecutionTracker
	LoopScheduler                  *LoopScheduler
	TemplateBundlesPath            string
	Search                         *searchcap.Service
	CodeContext                    *codecontext.Service
	Inputs                         *inputs.Service
	Capabilities                   *capabilities.Service
	MCPToolExecutor                swarm.MCPToolExecutor
	WorkerBackend                  workers.WorkerBackend
	Invocations                    *invocation.Store
	InvocationAdmissionUnavailable bool
	DispatchOutbox                 *dispatchoutbox.Store
	WorkerAuthority                *workerauthority.Store
	OutputValidator                outputvalidation.Validator
}

func NewAdminServer(r *router.Router, guard *governance.Guard, mem *memory.Service, db *sql.DB, cog *cognitive.Router, prov *provisioning.Engine, reg *registry.Service, soma *swarm.Soma, nc *nats.Conn, stream *signal.StreamHandler, architect *cognitive.MetaArchitect, arch *memory.Archivist, mcpSvc *mcp.Service, mcpPool *mcp.ClientPool, mcpLib *mcp.Library, cat *catalogue.Service, art *artifacts.Service, ex *exchange.Service, evStore *events.Store, runsManager *runs.Manager) *AdminServer {
	// Reactive engine: routes subscribed NATS messages to Soma for evaluation.
	// nc may be nil (NATS offline); engine degrades gracefully.
	reactiveEngine := reactive.New(nc, func(profileID, topic string, msg []byte) {
		// Log the reactive event — Soma routing is a future enhancement
		// (requires server reference; wired post-construction if needed).
		log.Printf("[reactive] profile %s received msg on %s (%d bytes)", profileID, topic, len(msg))
	})
	var mcpToolExecutor swarm.MCPToolExecutor
	if mcpSvc != nil && mcpPool != nil {
		mcpToolExecutor = mcp.NewToolExecutorAdapter(mcpSvc, mcpPool)
	}

	return &AdminServer{
		Router:              r,
		Guard:               guard,
		Mem:                 mem,
		DB:                  db,
		Cognitive:           cog,
		Provisioner:         prov,
		Registry:            reg,
		Soma:                soma,
		NC:                  nc,
		Stream:              stream,
		MetaArchitect:       architect,
		Archivist:           arch,
		Proposals:           NewProposalStore(),
		MCP:                 mcpSvc,
		MCPPool:             mcpPool,
		MCPLibrary:          mcpLib,
		Catalogue:           cat,
		Artifacts:           art,
		Exchange:            ex,
		Events:              evStore,
		Runs:                runsManager,
		Reactive:            reactiveEngine,
		GroupBus:            NewGroupBusMonitor(),
		Organizations:       NewOrganizationStore(db),
		LoopProfiles:        NewLoopProfileStore(),
		LoopResults:         NewLoopResultStore(),
		LoopExecution:       NewLoopExecutionTracker(),
		LoopScheduler:       nil,
		TemplateBundlesPath: "config/templates",
		MCPToolExecutor:     mcpToolExecutor,
		WorkerBackend:       configuredWorkerExecutionBackend(),
		Invocations:         invocation.NewStore(db),
		DispatchOutbox:      dispatchoutbox.NewStore(db),
		WorkerAuthority:     workerauthority.NewStore(db),
		OutputValidator:     configuredOutputValidator(),
	}
}

// GET /admin/approvals
func (s *AdminServer) handleApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, ok := requireRootAdminScope(w, r, scopeGovernanceRead); !ok {
		return
	}
	if s.Guard == nil {
		http.Error(w, "Governance disabled", http.StatusNotImplemented)
		return
	}

	pending := s.Guard.ListPending()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pending)
}

// POST /admin/approvals/{id} (legacy alias; root admin, approvals:decide)
func (s *AdminServer) handleApprovalAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireApprover(w, r); !ok {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/approvals/")
	if id == "" || id == r.URL.Path {
		respondAPIError(w, "missing approval request ID", http.StatusBadRequest)
		return
	}
	action, ok := decodeApprovalDecision(w, r)
	if !ok {
		return
	}
	s.resolveGuardApproval(w, r, id, action, "/admin/approvals")
}

// GET /agents
func (s *AdminServer) handleAgents(w http.ResponseWriter, r *http.Request) {
	agents := state.GlobalRegistry.GetActiveAgents()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_agents": len(agents),
		"agents":        agents,
	})
}

// GET /healthz
func (s *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

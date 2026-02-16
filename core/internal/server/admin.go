package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/overseer"
	"github.com/mycelis/core/internal/provisioning"
	"github.com/mycelis/core/internal/registry"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/internal/swarm"
	"github.com/nats-io/nats.go"
)

// AdminServer handles governance and system endpoints
type AdminServer struct {
	Router        *router.Router
	Guard         *governance.Guard
	Mem           *memory.Service
	Cognitive     *cognitive.Router
	Provisioner   *provisioning.Engine
	Registry      *registry.Service
	Soma          *swarm.Soma
	NC            *nats.Conn          // NATS for chat request-reply routing
	Stream        *signal.StreamHandler
	MetaArchitect *cognitive.MetaArchitect
	Overseer      *overseer.Engine    // Phase 5.2: Trust Economy
	Archivist     *memory.Archivist   // Phase 5.3: RAG Persistence
	Proposals     *ProposalStore      // Phase 5.3: Team Manifestation
	MCP           *mcp.Service        // Phase 7.0: MCP Ingress
	MCPPool       *mcp.ClientPool     // Phase 7.0: MCP Ingress
	MCPLibrary    *mcp.Library        // Phase 7.7: Curated MCP Library
	Catalogue     *catalogue.Service  // Phase 7.5: Agent Catalogue
	Artifacts     *artifacts.Service  // Phase 7.5: Agent Outputs
}

func NewAdminServer(r *router.Router, guard *governance.Guard, mem *memory.Service, cog *cognitive.Router, prov *provisioning.Engine, reg *registry.Service, soma *swarm.Soma, nc *nats.Conn, stream *signal.StreamHandler, architect *cognitive.MetaArchitect, ov *overseer.Engine, arch *memory.Archivist, mcpSvc *mcp.Service, mcpPool *mcp.ClientPool, mcpLib *mcp.Library, cat *catalogue.Service, art *artifacts.Service) *AdminServer {
	return &AdminServer{
		Router:        r,
		Guard:         guard,
		Mem:           mem,
		Cognitive:     cog,
		Provisioner:   prov,
		Registry:      reg,
		Soma:          soma,
		NC:            nc,
		Stream:        stream,
		MetaArchitect: architect,
		Overseer:      ov,
		Archivist:     arch,
		Proposals:     NewProposalStore(),
		MCP:           mcpSvc,
		MCPPool:       mcpPool,
		MCPLibrary:    mcpLib,
		Catalogue:     cat,
		Artifacts:     art,
	}
}

// RegisterRoutes adds handlers to the mux
func (s *AdminServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/approvals", s.handleApprovals)
	mux.HandleFunc("/admin/approvals/", s.handleApprovalAction) // Trailing slash for ID parsing
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/healthz", s.handleHealth)

	// Stream API (guarded — Stream may be nil in degraded mode)
	if s.Stream != nil {
		mux.HandleFunc("/api/v1/stream", s.Stream.HandleStream)
	} else {
		mux.HandleFunc("/api/v1/stream", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"SSE stream handler not initialized — infrastructure offline"}`, http.StatusServiceUnavailable)
		})
	}

	// Memory API
	mux.HandleFunc("/api/v1/memory/stream", s.GetMemoryStream)

	// Cognitive API (Real)
	mux.HandleFunc("/api/v1/cognitive/infer", s.handleInfer)
	mux.HandleFunc("/api/v1/cognitive/config", s.HandleCognitiveConfig)
	mux.HandleFunc("/api/v1/cognitive/matrix", s.HandleCognitiveConfig) // Alias: UI calls /matrix
	mux.HandleFunc("GET /api/v1/cognitive/status", s.HandleCognitiveStatus)
	mux.HandleFunc("PUT /api/v1/cognitive/profiles", s.HandleUpdateProfiles)
	mux.HandleFunc("PUT /api/v1/cognitive/providers/{id}", s.HandleUpdateProvider)
	mux.HandleFunc("/api/v1/chat", s.HandleChat)

	// Identity API
	mux.HandleFunc("/api/v1/user/me", s.HandleMe)
	mux.HandleFunc("/api/v1/teams", s.HandleTeams)
	mux.HandleFunc("/api/v1/user/settings", s.HandleUpdateSettings)

	// Missions API (Dashboard)
	mux.HandleFunc("GET /api/v1/missions", s.handleListMissions)

	// Provisioning API
	mux.HandleFunc("/api/v1/provision/draft", s.HandleProvisionDraft)
	mux.HandleFunc("/api/v1/provision/deploy", s.HandleProvisionDeploy)

	// Registry API
	mux.HandleFunc("/api/v1/registry/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			s.handleListTemplates(w, r)
		} else if r.Method == "POST" {
			s.handleRegisterTemplate(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	// Wildcard matching for connectors? Mux doesn't support wildcards well in stdlib 1.21- (but Go 1.22 does).
	// Assuming Go 1.22+ "POST /api/v1/teams/{id}/connectors".
	// Since we are using basic mux, we might rely on prefix matching if we register "/api/v1/teams/"
	// But HandleTeams is already on "/api/v1/teams".
	// We need to be careful with overlaps.
	// Let's register a specific path for connectors if possible, or handle inside HandleTeams?
	// HandleTeams probably handles GET /teams.
	// Let's register "/api/v1/teams/" as a prefix catcher if needed, OR just rely on specific pattern if Go 1.22.
	// Go 1.25 is in go.mod. So we HAVE path params!

	mux.HandleFunc("POST /api/v1/teams/{id}/connectors", s.handleInstallConnector)
	mux.HandleFunc("GET /api/v1/teams/{id}/wiring", s.handleGetWiring)

	// Intent Negotiation API
	mux.HandleFunc("POST /api/v1/intent/negotiate", s.handleIntentNegotiate)

	// Intent Commit (Instantiate Mission into Ledger + Activate in Soma)
	mux.HandleFunc("POST /api/v1/intent/commit", s.handleIntentCommit)

	// Phase 6.0: Symbiotic Seed (built-in sensor team, no LLM required)
	mux.HandleFunc("POST /api/v1/intent/seed/symbiotic", s.handleSymbioticSeed)

	// Phase 5.2: Telemetry & Trust Economy
	mux.HandleFunc("GET /api/v1/telemetry/compute", s.HandleTelemetry)
	mux.HandleFunc("/api/v1/trust/threshold", s.HandleTrustThreshold)

	// Phase 5.3: RAG Memory & Sensory
	mux.HandleFunc("GET /api/v1/memory/search", s.HandleMemorySearch)
	mux.HandleFunc("GET /api/v1/memory/sitreps", s.HandleListSitReps)
	mux.HandleFunc("GET /api/v1/sensors", s.HandleSensors)

	// Phase 5.3: Team Manifestation Proposals
	mux.HandleFunc("/api/v1/proposals", s.HandleProposals)
	mux.HandleFunc("POST /api/v1/proposals/{id}/approve", s.HandleProposalApprove)
	mux.HandleFunc("POST /api/v1/proposals/{id}/reject", s.HandleProposalReject)

	// Phase 7.0: MCP Ingress API
	mux.HandleFunc("POST /api/v1/mcp/install", s.handleMCPInstall)
	mux.HandleFunc("GET /api/v1/mcp/servers", s.handleMCPList)
	mux.HandleFunc("DELETE /api/v1/mcp/servers/{id}", s.handleMCPDelete)
	mux.HandleFunc("POST /api/v1/mcp/servers/{id}/tools/{tool}/call", s.handleMCPToolCall)
	mux.HandleFunc("GET /api/v1/mcp/tools", s.handleMCPToolsList)

	// Phase 7.7: MCP Library (curated server registry)
	mux.HandleFunc("GET /api/v1/mcp/library", s.handleMCPLibrary)
	mux.HandleFunc("POST /api/v1/mcp/library/install", s.handleMCPLibraryInstall)

	// Phase 7.7: Governance Policy API
	mux.HandleFunc("GET /api/v1/governance/policy", s.handleGetPolicy)
	mux.HandleFunc("PUT /api/v1/governance/policy", s.handleUpdatePolicy)
	mux.HandleFunc("GET /api/v1/governance/pending", s.handleGetPendingApprovals)
	mux.HandleFunc("POST /api/v1/governance/resolve/{id}", s.handleResolveApproval)

	// Phase 7.5: Agent Catalogue API
	mux.HandleFunc("GET /api/v1/catalogue/agents", s.handleListCatalogue)
	mux.HandleFunc("POST /api/v1/catalogue/agents", s.handleCreateCatalogue)
	mux.HandleFunc("PUT /api/v1/catalogue/agents/{id}", s.handleUpdateCatalogue)
	mux.HandleFunc("DELETE /api/v1/catalogue/agents/{id}", s.handleDeleteCatalogue)

	// Phase 7.5: Artifacts API (Agent Outputs)
	mux.HandleFunc("GET /api/v1/artifacts", s.handleListArtifacts)
	mux.HandleFunc("GET /api/v1/artifacts/{id}", s.handleGetArtifact)
	mux.HandleFunc("POST /api/v1/artifacts", s.handleStoreArtifact)
	mux.HandleFunc("PUT /api/v1/artifacts/{id}/status", s.handleUpdateArtifactStatus)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "JSON Encode Error", http.StatusInternalServerError)
	}
}

// GET /admin/approvals
func (s *AdminServer) handleApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

// POST /admin/approvals/{id}
func (s *AdminServer) handleApprovalAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	// /admin/approvals/req-123
	id := r.URL.Path[len("/admin/approvals/"):]
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	var payload struct {
		Action string `json:"action"` // APPROVE or DENY
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	approved := payload.Action == "APPROVE"

	msg, err := s.Guard.Resolve(id, approved, "admin-api")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if approved && msg != nil {
		// Re-inject into the system
		if err := s.Router.PublishDirect(msg); err != nil {
			log.Printf("Failed to re-publish approved msg: %v", err)
			http.Error(w, "Failed to re-publish", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"resolved"}`))
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

// POST /api/v1/intent/negotiate
func (s *AdminServer) handleIntentNegotiate(w http.ResponseWriter, r *http.Request) {
	if s.MetaArchitect == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Cognitive engine offline — MetaArchitect not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Intent string `json:"intent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if req.Intent == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"intent is required"}`, http.StatusBadRequest)
		return
	}

	blueprint, err := s.MetaArchitect.GenerateBlueprint(r.Context(), req.Intent)
	if err != nil {
		log.Printf("Intent negotiation failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error":"Cognitive engine error: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}

	respondJSON(w, blueprint)
}

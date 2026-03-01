package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/internal/events"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/overseer"
	"github.com/mycelis/core/internal/provisioning"
	"github.com/mycelis/core/internal/reactive"
	"github.com/mycelis/core/internal/registry"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/internal/state"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/internal/triggers"
	"github.com/mycelis/core/pkg/protocol"
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
	Overseer      *overseer.Engine   // Phase 5.2: Trust Economy
	Archivist     *memory.Archivist  // Phase 5.3: RAG Persistence
	Proposals     *ProposalStore     // Phase 5.3: Team Manifestation
	MCP           *mcp.Service       // Phase 7.0: MCP Ingress
	MCPPool       *mcp.ClientPool    // Phase 7.0: MCP Ingress
	MCPLibrary    *mcp.Library       // Phase 7.7: Curated MCP Library
	Catalogue     *catalogue.Service // Phase 7.5: Agent Catalogue
	Artifacts     *artifacts.Service // Phase 7.5: Agent Outputs
	Comms         *comms.Gateway     // External communication providers (whatsapp/telegram/slack/etc.)
	// V7 Event Spine (Team A)
	Events *events.Store // V7: persistent mission event audit trail
	Runs   *runs.Manager // V7: mission run lifecycle management
	// Mission Profiles & Reactive Subscriptions
	Reactive *reactive.Engine // watches NATS topics for active profiles
	// V7 Team B: Trigger Engine
	Triggers      *triggers.Store  // trigger rule CRUD + in-memory cache
	TriggerEngine *triggers.Engine // evaluates rules against CTS events
	// V7 Conversation Log
	Conversations *conversations.Store // full-fidelity agent conversation turns
	// V7 Inception Recipes — structured prompt patterns for RAG recall
	Inception *inception.Store // inception recipe CRUD + search
	// MCP Tool Sets — agent-scoped MCP tool bundles
	MCPToolSets *mcp.ToolSetService // tool set CRUD
}

func NewAdminServer(r *router.Router, guard *governance.Guard, mem *memory.Service, db *sql.DB, cog *cognitive.Router, prov *provisioning.Engine, reg *registry.Service, soma *swarm.Soma, nc *nats.Conn, stream *signal.StreamHandler, architect *cognitive.MetaArchitect, ov *overseer.Engine, arch *memory.Archivist, mcpSvc *mcp.Service, mcpPool *mcp.ClientPool, mcpLib *mcp.Library, cat *catalogue.Service, art *artifacts.Service, evStore *events.Store, runsManager *runs.Manager) *AdminServer {
	// Reactive engine: routes subscribed NATS messages to Soma for evaluation.
	// nc may be nil (NATS offline); engine degrades gracefully.
	reactiveEngine := reactive.New(nc, func(profileID, topic string, msg []byte) {
		// Log the reactive event — Soma routing is a future enhancement
		// (requires server reference; wired post-construction if needed).
		log.Printf("[reactive] profile %s received msg on %s (%d bytes)", profileID, topic, len(msg))
	})

	return &AdminServer{
		Router:        r,
		Guard:         guard,
		Mem:           mem,
		DB:            db,
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
		Events:        evStore,
		Runs:          runsManager,
		Reactive:      reactiveEngine,
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

	// Council Chat API — standardized, CTS-enveloped council interaction
	mux.HandleFunc("POST /api/v1/council/{member}/chat", s.HandleCouncilChat)
	mux.HandleFunc("GET /api/v1/council/members", s.HandleListCouncilMembers)

	// Identity API
	mux.HandleFunc("/api/v1/user/me", s.HandleMe)
	mux.HandleFunc("/api/v1/teams", s.HandleTeams)
	mux.HandleFunc("GET /api/v1/teams/detail", s.HandleTeamsDetail)
	mux.HandleFunc("/api/v1/user/settings", s.HandleUpdateSettings)

	// Missions API (Dashboard + Phase 9: Neural Wiring Edit/Delete)
	mux.HandleFunc("GET /api/v1/missions", s.handleListMissions)
	mux.HandleFunc("GET /api/v1/missions/{id}", s.handleGetMission)
	mux.HandleFunc("PUT /api/v1/missions/{id}/agents/{name}", s.handleUpdateMissionAgent)
	mux.HandleFunc("DELETE /api/v1/missions/{id}/agents/{name}", s.handleDeleteMissionAgent)
	mux.HandleFunc("DELETE /api/v1/missions/{id}", s.handleDeleteMission)

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
	mux.HandleFunc("POST /api/v1/intent/confirm-action", s.HandleConfirmAction)

	// CE-1: Orchestration Templates & Intent Proofs
	mux.HandleFunc("GET /api/v1/templates", s.handleListTemplatesAPI)
	mux.HandleFunc("GET /api/v1/intent/proof/{id}", s.handleGetIntentProof)

	// Phase 6.0: Symbiotic Seed (built-in sensor team, no LLM required)
	mux.HandleFunc("POST /api/v1/intent/seed/symbiotic", s.handleSymbioticSeed)

	// Phase 5.2: Telemetry & Trust Economy
	mux.HandleFunc("GET /api/v1/telemetry/compute", s.HandleTelemetry)
	mux.HandleFunc("/api/v1/trust/threshold", s.HandleTrustThreshold)

	// Phase 5.3: RAG Memory & Sensory
	mux.HandleFunc("GET /api/v1/memory/search", s.HandleMemorySearch)
	mux.HandleFunc("GET /api/v1/memory/sitreps", s.HandleListSitReps)
	mux.HandleFunc("/api/v1/memory/temp", s.HandleTempMemory)
	mux.HandleFunc("GET /api/v1/sensors", s.HandleSensors)
	mux.HandleFunc("GET /api/v1/comms/providers", s.HandleCommsProviders)
	mux.HandleFunc("POST /api/v1/comms/send", s.HandleCommsSend)
	mux.HandleFunc("POST /api/v1/comms/inbound/{provider}", s.HandleCommsInbound)

	// Phase 5.3: Team Manifestation Proposals
	mux.HandleFunc("/api/v1/proposals", s.HandleProposals)
	mux.HandleFunc("POST /api/v1/proposals/{id}/approve", s.HandleProposalApprove)
	mux.HandleFunc("POST /api/v1/proposals/{id}/reject", s.HandleProposalReject)

	// Phase 7.0: MCP Ingress API — raw install disabled (Phase 0 security)
	mux.HandleFunc("POST /api/v1/mcp/install", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Raw MCP install disabled. Use POST /api/v1/mcp/library/install with a curated server name."}`))
	})
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

	// Phase 19: Brains API (Provider Management)
	mux.HandleFunc("GET /api/v1/brains", s.HandleListBrains)
	mux.HandleFunc("PUT /api/v1/brains/{id}/toggle", s.HandleToggleBrain)
	mux.HandleFunc("PUT /api/v1/brains/{id}/policy", s.HandleUpdateBrainPolicy)
	// Provider CRUD + probe (hot-reload, no restart required)
	mux.HandleFunc("POST /api/v1/brains", s.HandleAddBrain)
	mux.HandleFunc("PUT /api/v1/brains/{id}", s.HandleUpdateBrain)
	mux.HandleFunc("DELETE /api/v1/brains/{id}", s.HandleDeleteBrain)
	mux.HandleFunc("POST /api/v1/brains/{id}/probe", s.HandleProbeBrain)

	// Context Snapshots (save/restore conversation context on profile switch)
	mux.HandleFunc("GET /api/v1/context/snapshots", s.HandleListSnapshots)
	mux.HandleFunc("POST /api/v1/context/snapshot", s.HandleCreateSnapshot)
	mux.HandleFunc("GET /api/v1/context/snapshots/{id}", s.HandleGetSnapshot)

	// Mission Profiles (role→provider routing + reactive NATS subscriptions)
	mux.HandleFunc("GET /api/v1/mission-profiles", s.HandleListMissionProfiles)
	mux.HandleFunc("POST /api/v1/mission-profiles", s.HandleCreateMissionProfile)
	mux.HandleFunc("PUT /api/v1/mission-profiles/{id}", s.HandleUpdateMissionProfile)
	mux.HandleFunc("DELETE /api/v1/mission-profiles/{id}", s.HandleDeleteMissionProfile)
	mux.HandleFunc("POST /api/v1/mission-profiles/{id}/activate", s.HandleActivateMissionProfile)

	// V7 Event Spine (Team A): Run Timeline + Causal Chain
	mux.HandleFunc("GET /api/v1/runs", s.handleListRuns)
	mux.HandleFunc("GET /api/v1/runs/{id}/events", s.handleGetRunEvents)
	mux.HandleFunc("GET /api/v1/runs/{id}/chain", s.handleGetRunChain)

	// V7 Team B: Trigger Rules Engine
	mux.HandleFunc("GET /api/v1/triggers", s.HandleListTriggers)
	mux.HandleFunc("POST /api/v1/triggers", s.HandleCreateTrigger)
	mux.HandleFunc("PUT /api/v1/triggers/{id}", s.HandleUpdateTrigger)
	mux.HandleFunc("DELETE /api/v1/triggers/{id}", s.HandleDeleteTrigger)
	mux.HandleFunc("POST /api/v1/triggers/{id}/toggle", s.HandleToggleTrigger)
	mux.HandleFunc("GET /api/v1/triggers/{id}/history", s.HandleTriggerHistory)

	// Service health dashboard
	mux.HandleFunc("GET /api/v1/services/status", s.HandleServicesStatus)

	// V7 Conversation Log: agent transcript browsing + user interjection
	mux.HandleFunc("GET /api/v1/runs/{id}/conversation", s.HandleGetRunConversation)
	mux.HandleFunc("GET /api/v1/conversations/{session_id}", s.HandleGetSessionConversation)
	mux.HandleFunc("POST /api/v1/runs/{id}/interject", s.HandleRunInterject)

	// V7 Inception Recipes: structured prompt patterns for RAG recall
	mux.HandleFunc("GET /api/v1/inception/contracts", s.HandleInceptionContracts)
	mux.HandleFunc("GET /api/v1/inception/recipes", s.HandleListInceptionRecipes)
	mux.HandleFunc("GET /api/v1/inception/recipes/search", s.HandleSearchInceptionRecipes)
	mux.HandleFunc("GET /api/v1/inception/recipes/{id}", s.HandleGetInceptionRecipe)
	mux.HandleFunc("POST /api/v1/inception/recipes", s.HandleCreateInceptionRecipe)
	mux.HandleFunc("PATCH /api/v1/inception/recipes/{id}/quality", s.HandleUpdateRecipeQuality)

	// MCP Tool Sets: agent-scoped MCP tool bundles
	mux.HandleFunc("GET /api/v1/mcp/toolsets", s.handleListToolSets)
	mux.HandleFunc("POST /api/v1/mcp/toolsets", s.handleCreateToolSet)
	mux.HandleFunc("PUT /api/v1/mcp/toolsets/{id}", s.handleUpdateToolSet)
	mux.HandleFunc("DELETE /api/v1/mcp/toolsets/{id}", s.handleDeleteToolSet)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "JSON Encode Error", http.StatusInternalServerError)
	}
}

// respondError writes a JSON error response with proper escaping.
// Never use fmt.Sprintf to build JSON — raw error text can contain quotes.
func respondError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
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
// Phase 10: Routes through Admin agent's ReAct loop for researched blueprint
// generation. Falls back to direct MetaArchitect if Admin agent is unavailable.
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

	// Try routing through Admin agent for research-enriched blueprint
	var blueprint *protocol.MissionBlueprint
	if s.NC != nil {
		bp, err := s.negotiateViaAdmin(r.Context(), req.Intent)
		if err == nil && bp != nil {
			blueprint = bp
		} else {
			log.Printf("Admin-routed negotiate failed, falling back to direct MetaArchitect: %v", err)
		}
	}

	// Fallback: direct MetaArchitect (single-shot, no research)
	if blueprint == nil {
		bp, err := s.MetaArchitect.GenerateBlueprint(r.Context(), req.Intent)
		if err != nil {
			log.Printf("Intent negotiation failed: %v", err)
			respondError(w, "Cognitive engine error: "+err.Error(), http.StatusBadGateway)
			return
		}
		blueprint = bp
	}

	// CE-1: Build scope validation, create intent proof + confirm token
	scope := buildScopeFromBlueprint(blueprint)
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "negotiate",
		fmt.Sprintf("Blueprint negotiation: %s", req.Intent),
		map[string]any{"intent": req.Intent, "teams": len(blueprint.Teams), "scope": scope},
	)

	proof, _ := s.createIntentProof(protocol.TemplateChatToProposal, req.Intent, scope, auditEventID)
	var confirmToken *protocol.ConfirmToken
	if proof != nil {
		confirmToken, _ = s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
	}

	templateSpec := protocol.TemplateRegistry[protocol.TemplateChatToProposal]
	respondJSON(w, protocol.NegotiateResponse{
		Blueprint:    blueprint,
		IntentProof:  proof,
		ConfirmToken: confirmToken,
		Template:     &templateSpec,
	})
}

// negotiateViaAdmin sends the intent to the Admin agent via NATS request-reply.
// The Admin agent uses research_for_blueprint → generate_blueprint in its ReAct loop.
func (s *AdminServer) negotiateViaAdmin(ctx context.Context, intent string) (*protocol.MissionBlueprint, error) {
	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, "admin")
	directive := fmt.Sprintf(
		"The user wants to negotiate a mission blueprint. Their intent is:\n\n%s\n\n"+
			"Use research_for_blueprint to gather context first, then use generate_blueprint "+
			"to create the mission blueprint. Return ONLY the blueprint JSON — no commentary.",
		intent,
	)

	reqCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	msg, err := s.NC.RequestWithContext(reqCtx, subject, []byte(directive))
	if err != nil {
		return nil, fmt.Errorf("admin agent did not respond: %w", err)
	}

	// Extract blueprint JSON from the admin's response
	return extractBlueprintFromResponse(string(msg.Data))
}

// extractBlueprintFromResponse parses a MissionBlueprint from an agent's text response.
// The agent may wrap the JSON in markdown fences or include commentary.
func extractBlueprintFromResponse(response string) (*protocol.MissionBlueprint, error) {
	text := response

	// Strip markdown fences if present
	if idx := strings.Index(text, "```json"); idx >= 0 {
		text = text[idx+7:]
		if end := strings.Index(text, "```"); end >= 0 {
			text = text[:end]
		}
	} else if idx := strings.Index(text, "```"); idx >= 0 {
		text = text[idx+3:]
		if end := strings.Index(text, "```"); end >= 0 {
			text = text[:end]
		}
	}

	// Find first { and last } for bare JSON extraction
	text = strings.TrimSpace(text)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("no JSON object found in admin response")
	}
	text = text[start : end+1]

	var bp protocol.MissionBlueprint
	if err := json.Unmarshal([]byte(text), &bp); err != nil {
		return nil, fmt.Errorf("failed to parse blueprint JSON: %w", err)
	}

	if bp.MissionID == "" || len(bp.Teams) == 0 {
		return nil, fmt.Errorf("invalid blueprint: missing mission_id or teams")
	}

	return &bp, nil
}

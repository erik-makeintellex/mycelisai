package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/provisioning"
	"github.com/mycelis/core/internal/registry"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/state"
)

// AdminServer handles governance and system endpoints
type AdminServer struct {
	Router      *router.Router
	GK          *governance.Gatekeeper
	Mem         *memory.Archivist
	Cognitive   *cognitive.Router
	Provisioner *provisioning.Engine
	Registry    *registry.Service
}

func NewAdminServer(r *router.Router, gk *governance.Gatekeeper, mem *memory.Archivist, cog *cognitive.Router, prov *provisioning.Engine, reg *registry.Service) *AdminServer {
	return &AdminServer{
		Router:      r,
		GK:          gk,
		Mem:         mem,
		Cognitive:   cog,
		Provisioner: prov,
		Registry:    reg,
	}
}

// RegisterRoutes adds handlers to the mux
func (s *AdminServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/approvals", s.handleApprovals)
	mux.HandleFunc("/admin/approvals/", s.handleApprovalAction) // Trailing slash for ID parsing
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/healthz", s.handleHealth)

	// Memory API
	mux.HandleFunc("/api/v1/memory/stream", s.GetMemoryStream)

	// Cognitive API (Real)
	mux.HandleFunc("/api/v1/cognitive/infer", s.handleInfer)
	mux.HandleFunc("/api/v1/brain/config", s.HandleBrainConfig)
	mux.HandleFunc("/api/v1/chat", s.HandleChat)

	// Identity API
	mux.HandleFunc("/api/v1/user/me", s.HandleMe)
	mux.HandleFunc("/api/v1/teams", s.HandleTeams)
	mux.HandleFunc("/api/v1/user/settings", s.HandleUpdateSettings)

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

	if s.GK == nil {
		http.Error(w, "Governance disabled", http.StatusNotImplemented)
		return
	}

	pending := s.GK.ListPending()
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

	msg, err := s.GK.Resolve(id, approved, "admin-api")
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

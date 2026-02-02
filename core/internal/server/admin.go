package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/router"
	"github.com/mycelis/core/internal/state"
)

// AdminServer handles governance and system endpoints
type AdminServer struct {
	Router *router.Router
	GK     *governance.Gatekeeper
	Mem    *memory.Archivist
}

func NewAdminServer(r *router.Router, gk *governance.Gatekeeper, mem *memory.Archivist) *AdminServer {
	return &AdminServer{
		Router: r,
		GK:     gk,
		Mem:    mem,
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

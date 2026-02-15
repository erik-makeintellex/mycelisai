package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ── Team Manifestation Proposals ─────────────────────────────
// In-memory store for team proposals from the Meta-Architect.
// Proposals are transient — they regenerate on restart.

type ProposalStatus string

const (
	ProposalPending  ProposalStatus = "pending"
	ProposalApproved ProposalStatus = "approved"
	ProposalRejected ProposalStatus = "rejected"
)

type ProposedAgent struct {
	ID           string `json:"id"`
	Role         string `json:"role"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	Model        string `json:"model,omitempty"`
}

type TeamProposal struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Role      string          `json:"role"`
	Agents    []ProposedAgent `json:"agents"`
	Reason    string          `json:"reason"`
	Status    ProposalStatus  `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

// ProposalStore is a thread-safe in-memory store for team proposals.
type ProposalStore struct {
	mu        sync.RWMutex
	proposals map[string]*TeamProposal
}

func NewProposalStore() *ProposalStore {
	ps := &ProposalStore{
		proposals: make(map[string]*TeamProposal),
	}
	ps.seedDefaults()
	return ps
}

// seedDefaults populates the store with initial Mother Brain proposals.
// These represent the autonomous team suggestions generated at boot.
func (ps *ProposalStore) seedDefaults() {
	ps.Add(&TeamProposal{
		Name: "Signal Analytics Squad",
		Role: "data_analyst",
		Agents: []ProposedAgent{
			{ID: "analyst-lead", Role: "lead_analyst", Model: "qwen2.5-coder:7b-instruct"},
			{ID: "metric-collector", Role: "collector", Model: "qwen2.5-coder:7b-instruct"},
		},
		Reason: "Telemetry volume exceeds manual review capacity. Recommend autonomous anomaly detection across NATS streams.",
		Status: ProposalPending,
	})
	ps.Add(&TeamProposal{
		Name: "Knowledge Curator Team",
		Role: "curator",
		Agents: []ProposedAgent{
			{ID: "curator-primary", Role: "knowledge_curator", Model: "qwen2.5-coder:7b-instruct"},
			{ID: "indexer-agent", Role: "indexer", Model: "nomic-embed-text"},
			{ID: "summarizer", Role: "summarizer", Model: "qwen2.5-coder:7b-instruct"},
		},
		Reason: "SitRep archive growing. Propose team to maintain semantic index and generate weekly intelligence briefs.",
		Status: ProposalPending,
	})
}

func (ps *ProposalStore) Add(p *TeamProposal) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.Status == "" {
		p.Status = ProposalPending
	}
	ps.proposals[p.ID] = p
}

func (ps *ProposalStore) List() []*TeamProposal {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	result := make([]*TeamProposal, 0, len(ps.proposals))
	for _, p := range ps.proposals {
		result = append(result, p)
	}
	return result
}

func (ps *ProposalStore) Get(id string) (*TeamProposal, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.proposals[id]
	return p, ok
}

func (ps *ProposalStore) UpdateStatus(id string, status ProposalStatus) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	p, ok := ps.proposals[id]
	if !ok {
		return fmt.Errorf("proposal %s not found", id)
	}
	p.Status = status
	return nil
}

// ── HTTP Handlers ────────────────────────────────────────────

// GET /api/v1/proposals — list all proposals
// POST /api/v1/proposals — create a new proposal (from Meta-Architect or manual)
func (s *AdminServer) HandleProposals(w http.ResponseWriter, r *http.Request) {
	if s.Proposals == nil {
		http.Error(w, `{"error":"proposal store not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case "GET":
		proposals := s.Proposals.List()
		respondJSON(w, map[string]any{
			"proposals": proposals,
			"count":     len(proposals),
		})

	case "POST":
		var p TeamProposal
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if p.Name == "" {
			http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
			return
		}
		s.Proposals.Add(&p)
		w.WriteHeader(http.StatusCreated)
		respondJSON(w, p)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /api/v1/proposals/{id}/approve — approve and manifest team
func (s *AdminServer) HandleProposalApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"proposal ID required"}`, http.StatusBadRequest)
		return
	}

	proposal, ok := s.Proposals.Get(id)
	if !ok {
		http.Error(w, `{"error":"proposal not found"}`, http.StatusNotFound)
		return
	}

	if proposal.Status != ProposalPending {
		http.Error(w, fmt.Sprintf(`{"error":"proposal already %s"}`, proposal.Status), http.StatusConflict)
		return
	}

	if err := s.Proposals.UpdateStatus(id, ProposalApproved); err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{
		"status":   "approved",
		"proposal": proposal,
	})
}

// POST /api/v1/proposals/{id}/reject — reject proposal
func (s *AdminServer) HandleProposalReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"proposal ID required"}`, http.StatusBadRequest)
		return
	}

	if _, ok := s.Proposals.Get(id); !ok {
		http.Error(w, `{"error":"proposal not found"}`, http.StatusNotFound)
		return
	}

	if err := s.Proposals.UpdateStatus(id, ProposalRejected); err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]any{
		"status": "rejected",
		"id":     id,
	})
}

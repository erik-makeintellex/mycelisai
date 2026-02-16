package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Soma is the "Executive Cell Body" of the Swarm.
// It acts as the User Proxy, receiving external inputs and high-level directives.
// It delegates execution to "Axon" (The Messenger) and Action/Expression Cores.
type Soma struct {
	id            string
	nc            *nats.Conn
	guard         *governance.Guard
	axon          *Axon
	teams         map[string]*Team
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	registry      *Registry
	brain         *cognitive.Router
	toolExecutor  MCPToolExecutor // composite (internal + MCP)
	internalTools *InternalToolRegistry
}

// NewSoma creates a new Executive instance.
// internalTools may be nil; mcpExec may be nil. Both are composed into a
// CompositeToolExecutor that is passed to all teams and agents.
func NewSoma(nc *nats.Conn, guard *governance.Guard, registry *Registry, brain *cognitive.Router, stream *signal.StreamHandler, mcpExec MCPToolExecutor, internalTools *InternalToolRegistry) *Soma {
	// Build composite tool executor (internal first, MCP fallback)
	var composite MCPToolExecutor
	if internalTools != nil || mcpExec != nil {
		composite = NewCompositeToolExecutor(internalTools, mcpExec)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Soma{
		id:            "soma-core",
		nc:            nc,
		guard:         guard,
		registry:      registry,
		brain:         brain,
		toolExecutor:  composite,
		internalTools: internalTools,
		teams:         make(map[string]*Team),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Wire Soma back-reference into internal tools (for list_teams, BuildContext, etc.)
	if internalTools != nil {
		internalTools.SetSoma(s)
	}

	// Axon is Soma's Assistant
	s.axon = NewAxon(nc, s, stream)
	return s
}

// Start brings the Soma online, listening to the global bus.
func (s *Soma) Start() error {
	log.Printf("🧠 Soma [%s] Online. Listening for User Intent...", s.id)

	// 0. Build tool descriptions for prompt injection
	var toolDescs map[string]string
	if s.internalTools != nil {
		toolDescs = s.internalTools.ListDescriptions()
	}

	// 1. Load Teams from Registry
	manifests, err := s.registry.LoadManifests()
	if err != nil {
		log.Printf("WARN: Failed to load team manifests: %v", err)
	}
	for _, m := range manifests {
		team := NewTeam(m, s.nc, s.brain, s.toolExecutor)
		if len(toolDescs) > 0 {
			team.SetToolDescriptions(toolDescs)
		}
		if s.internalTools != nil {
			team.SetInternalTools(s.internalTools)
		}
		s.teams[m.ID] = team
		if err := team.Start(); err != nil {
			log.Printf("ERR: Failed to start team %s: %v", m.ID, err)
		}
	}

	// 1. Subscribe to Global User Input (GUI, CLI, Sensors)
	if _, err = s.nc.Subscribe(protocol.TopicGlobalInputWild, s.handleGlobalInput); err != nil {
		return fmt.Errorf("failed to subscribe to global input: %w", err)
	}

	// 2. Start Axon
	if err := s.axon.Start(); err != nil {
		return fmt.Errorf("failed to start Axon: %w", err)
	}

	return nil
}

// handleGlobalInput processes raw external signals.
// CRITICAL: ALL inputs must pass Guard validation.
func (s *Soma) handleGlobalInput(msg *nats.Msg) {
	// 1. Security Check (Guard)
	if err := s.guard.ValidateIngress(msg.Subject, msg.Data); err != nil {
		log.Printf("🛡️ Soma Shield Blocked Input: %v", err)
		return
	}

	// 2. Parse Intent
	// Ideally this uses an LLM to classify, for now we use simple routing or pass to Axon.
	log.Printf("🧠 Soma Received Input on [%s]: %s", msg.Subject, string(msg.Data))

	// 3. Delegate to Axon for Optimization & Routing
	// Soma decides "What", Axon decides "How/Where"
	s.axon.ProcessSignal(msg)
}

// SpawnTeam dynamically creates and starts a new Team.
func (s *Soma) SpawnTeam(manifest *TeamManifest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.teams[manifest.ID]; exists {
		return fmt.Errorf("team %s already exists", manifest.ID)
	}

	// Persist? For now, just runtime spawn.
	// TODO: Save to Registry (YAML)

	team := NewTeam(manifest, s.nc, s.brain, s.toolExecutor)
	if s.internalTools != nil {
		team.SetToolDescriptions(s.internalTools.ListDescriptions())
	}
	if s.internalTools != nil {
		team.SetInternalTools(s.internalTools)
	}
	if err := team.Start(); err != nil {
		return err
	}

	s.teams[manifest.ID] = team
	log.Printf("Soma Spawned New Team: %s", manifest.ID)
	return nil
}

// ListTeams returns a snapshot of active teams.
func (s *Soma) ListTeams() []*TeamManifest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*TeamManifest
	for _, t := range s.teams {
		list = append(list, t.Manifest)
	}
	return list
}

// REST Handlers (For Router)

// HandleCreateTeam processes HTTP POST /api/swarm/teams
// and GET /api/swarm/teams (List)
func (s *Soma) HandleCreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.HandleListTeams(w, r)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var manifest TeamManifest
	if err := json.NewDecoder(r.Body).Decode(&manifest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic Validation
	if manifest.ID == "" || manifest.Name == "" {
		http.Error(w, "Missing ID or Name", http.StatusBadRequest)
		return
	}

	// Sanatize/Default
	if manifest.Type == "" {
		manifest.Type = TeamTypeAction
	}

	if err := s.SpawnTeam(&manifest); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "spawned", "id": manifest.ID})
}

// HandleListTeams returns the list of active teams.
func (s *Soma) HandleListTeams(w http.ResponseWriter, r *http.Request) {
	teams := s.ListTeams()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// HandleCommand processes HTTP POST /api/swarm/command
// It injects a user command into the Swarm Global Input bus.
func (s *Soma) HandleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Content string `json:"content"`
		Source  string `json:"source"` // e.g. "mission-control"
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.Content == "" {
		http.Error(w, "Missing content", http.StatusBadRequest)
		return
	}

	// 1. Construct Message
	// Topic: swarm.global.input.user
	subject := protocol.TopicGlobalInputUser
	if payload.Source != "" {
		subject = fmt.Sprintf("swarm.global.input.%s", payload.Source)
	}

	// 2. Publish
	if err := s.nc.Publish(subject, []byte(payload.Content)); err != nil {
		log.Printf("Failed to publish command: %v", err)
		http.Error(w, "Failed to inject command", http.StatusInternalServerError)
		return
	}
	s.nc.Flush()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent", "subject": subject})
}

// HandleBroadcast processes HTTP POST /api/v1/swarm/broadcast
// It fans out a directive message to ALL active teams' internal trigger topics.
func (s *Soma) HandleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Content string `json:"content"`
		Source  string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.Content == "" {
		http.Error(w, "Missing content", http.StatusBadRequest)
		return
	}

	if payload.Source == "" {
		payload.Source = "mission-control"
	}

	// Fan out to all active teams
	s.mu.RLock()
	teamCount := 0
	for id := range s.teams {
		subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, id)
		if err := s.nc.Publish(subject, []byte(payload.Content)); err != nil {
			log.Printf("Broadcast failed for team [%s]: %v", id, err)
			continue
		}
		teamCount++
		log.Printf("📡 Broadcast to team [%s] on [%s]", id, subject)
	}
	s.mu.RUnlock()
	s.nc.Flush()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "broadcast",
		"teams_hit": teamCount,
		"source":    payload.Source,
	})
}

// Shutdown stops the Soma, all teams, and its Axon.
func (s *Soma) Shutdown() {
	s.mu.RLock()
	for id, t := range s.teams {
		log.Printf("Soma shutting down Team [%s]", id)
		t.Stop()
	}
	s.mu.RUnlock()
	s.cancel()
}

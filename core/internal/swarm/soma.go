package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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
	// V7 Event Spine: optional run tracking + event audit (interfaces from pkg/protocol).
	// Both may be nil — degraded mode: teams still activate, events just aren't recorded.
	runsManager  protocol.RunsManager  // creates mission_run records
	eventEmitter protocol.EventEmitter // persists + publishes events
	// V7 Conversation Log: optional full-fidelity turn logger.
	conversationLogger protocol.ConversationLogger
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
		// V7: wire conversation logger into standing teams
		if s.conversationLogger != nil {
			team.SetConversationLogger(s.conversationLogger)
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
	// V7: wire conversation logger into dynamically spawned teams
	if s.conversationLogger != nil {
		team.SetConversationLogger(s.conversationLogger)
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

// BroadcastReply holds one team's response to a broadcast.
type BroadcastReply struct {
	TeamID  string `json:"team_id"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// HandleBroadcast processes HTTP POST /api/v1/swarm/broadcast
// It fans out a directive to ALL active teams via NATS request-reply,
// waits for each team's response, and returns them to the caller.
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

	// Snapshot team IDs under lock
	s.mu.RLock()
	teamIDs := make([]string, 0, len(s.teams))
	for id := range s.teams {
		teamIDs = append(teamIDs, id)
	}
	s.mu.RUnlock()

	// Fan out request-reply concurrently, one per team
	var wg sync.WaitGroup
	replies := make([]BroadcastReply, len(teamIDs))
	timeout := 60 * time.Second

	for i, id := range teamIDs {
		wg.Add(1)
		go func(idx int, teamID string) {
			defer wg.Done()
			subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID)
			log.Printf("📡 Broadcast request to team [%s] on [%s]", teamID, subject)

			msg, err := s.nc.Request(subject, []byte(payload.Content), timeout)
			if err != nil {
				log.Printf("Broadcast: team [%s] did not respond: %v", teamID, err)
				replies[idx] = BroadcastReply{TeamID: teamID, Error: err.Error()}
				return
			}
			replies[idx] = BroadcastReply{TeamID: teamID, Content: string(msg.Data)}
		}(i, id)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "broadcast",
		"teams_hit": len(teamIDs),
		"source":    payload.Source,
		"replies":   replies,
	})
}

// DeactivateMission stops and removes all teams belonging to a mission.
// Team IDs follow the pattern "{missionID}.{sanitized-team-name}".
func (s *Soma) DeactivateMission(missionID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := missionID + "."
	stopped := 0
	for id, team := range s.teams {
		if strings.HasPrefix(id, prefix) {
			team.Stop()
			delete(s.teams, id)
			stopped++
		}
	}
	if stopped > 0 {
		log.Printf("DeactivateMission: stopped %d teams for mission %s", stopped, missionID)
	}
	return stopped
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

// SetRunsManager wires the V7 run tracking manager into Soma.
// Must be called before ActivateBlueprint for run records to be created.
func (s *Soma) SetRunsManager(rm protocol.RunsManager) { s.runsManager = rm }

// SetEventEmitter wires the V7 event store into Soma for audit trail emission.
// Must be called before ActivateBlueprint for tool events to be recorded.
func (s *Soma) SetEventEmitter(emitter protocol.EventEmitter) { s.eventEmitter = emitter }

// SetConversationLogger wires the V7 conversation logger into Soma.
// Must be called before Start() so standing teams receive the logger.
func (s *Soma) SetConversationLogger(logger protocol.ConversationLogger) {
	s.conversationLogger = logger
}

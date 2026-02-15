package swarm

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/nats-io/nats.go"
)

// Soma is the "Executive Cell Body" of the Swarm.
// It acts as the User Proxy, receiving external inputs and high-level directives.
// It delegates execution to "Axon" (The Messenger) and Action/Expression Cores.
type Soma struct {
	id       string
	nc       *nats.Conn
	guard    *governance.Guard
	axon     *Axon
	teams    map[string]*Team
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	registry *Registry
	brain    *cognitive.Router
}

// NewSoma creates a new Executive instance.
func NewSoma(nc *nats.Conn, guard *governance.Guard, registry *Registry, brain *cognitive.Router) *Soma {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Soma{
		id:       "soma-core",
		nc:       nc,
		guard:    guard,
		registry: registry,
		brain:    brain,
		teams:    make(map[string]*Team),
		ctx:      ctx,
		cancel:   cancel,
	}
	// Axon is Soma's Assistant
	s.axon = NewAxon(nc, s)
	return s
}

// Start brings the Soma online, listening to the global bus.
func (s *Soma) Start() error {
	log.Printf("🧠 Soma [%s] Online. Listening for User Intent...", s.id)

	// 0. Load Teams from Registry
	manifests, err := s.registry.LoadManifests()
	if err != nil {
		log.Printf("WARN: Failed to load team manifests: %v", err)
	}
	for _, m := range manifests {
		team := NewTeam(m, s.nc, s.brain)
		s.teams[m.ID] = team
		if err := team.Start(); err != nil {
			log.Printf("ERR: Failed to start team %s: %v", m.ID, err)
		}
	}

	// 1. Subscribe to Global User Input (GUI, CLI, Sensors)
	if _, err = s.nc.Subscribe("swarm.global.input.>", s.handleGlobalInput); err != nil {
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

// Shutdown stops the Soma and its Axon.
func (s *Soma) Shutdown() {
	s.cancel()
	// Shutdown teams?
}

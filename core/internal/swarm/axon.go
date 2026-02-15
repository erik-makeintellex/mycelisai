package swarm

import (
	"log"

	"github.com/nats-io/nats.go"
)

// Axon is Soma's "Messenger" and Signal Optimizer.
// It handles the "How" of communication: routing, protocol translation, and optimization.
type Axon struct {
	nc   *nats.Conn
	soma *Soma
}

// NewAxon creates a new Messenger instance.
func NewAxon(nc *nats.Conn, soma *Soma) *Axon {
	return &Axon{
		nc:   nc,
		soma: soma,
	}
}

// Start brings Axon online.
func (a *Axon) Start() error {
	log.Println("⚡ Axon Messenger Online. Ready to route signals.")
	// Subscribe to internal signals if needed
	return nil
}

// ProcessSignal analyzes a raw signal from Soma and dispatches it to the appropriate Team(s).
func (a *Axon) ProcessSignal(msg *nats.Msg) {
	// 1. Optimize Signal (e.g., compress, format, enrich)
	// For now, pass through.

	// 2. Determine Route based on Intent
	// This would use the Cognitive Registry to find the "Best Team" for the job.
	// Hardcoded for V6.2 MVP:
	targetTopic := "swarm.team.genesis.internal.command" // Default to Genesis Team

	if string(msg.Data) == "system_status" {
		targetTopic = "swarm.team.telemetry.signal.status" // Expression Team
	}

	// 3. Dispatch
	log.Printf("⚡ Axon Routing Signal to [%s]", targetTopic)
	a.nc.Publish(targetTopic, msg.Data)
}

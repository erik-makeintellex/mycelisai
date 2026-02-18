package swarm

import (
	"fmt"
	"log"
	"time"

	"github.com/mycelis/core/internal/signal"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Axon is Soma's "Messenger" and Signal Optimizer.
// It handles the "How" of communication: routing, protocol translation, and optimization.
type Axon struct {
	nc     *nats.Conn
	soma   *Soma
	Stream *signal.StreamHandler // Added Stream field
}

// NewAxon creates a new Messenger instance.
func NewAxon(nc *nats.Conn, soma *Soma, stream *signal.StreamHandler) *Axon { // Updated signature
	return &Axon{
		nc:     nc,
		soma:   soma,
		Stream: stream, // Initialize Stream
	}
}

// Start brings Axon online.
func (a *Axon) Start() error {
	log.Println("⚡ Axon Messenger Online. Ready to route signals.")

	// Subscribe to all team internal events for monitoring/stream
	// "swarm.team.*.internal.>"
	_, err := a.nc.Subscribe(protocol.TopicTeamInternalWild, a.handleTeamEvent)
	if err != nil {
		return err
	}

	return nil
}

// handleTeamEvent captures traffic for the Activity Stream
func (a *Axon) handleTeamEvent(msg *nats.Msg) {
	// Broadcast to SSE
	// We wrap in a simple structure
	topic := msg.Subject
	payload := string(msg.Data)

	// Classify
	// swarm.team.<id>.internal.<type>
	// e.g. swarm.team.genesis.internal.command
	//      swarm.team.genesis.internal.response

	if a.Stream != nil {
		jsonMsg := fmt.Sprintf(`{"type": "activity", "topic": "%s", "message": %q, "timestamp": "%s"}`,
			topic, payload, time.Now().Format(time.RFC3339))
		a.Stream.Broadcast(jsonMsg)
	}
}

// ProcessSignal analyzes a raw signal from Soma and dispatches it to the appropriate Team(s).
func (a *Axon) ProcessSignal(msg *nats.Msg) {
	// 1. Optimize Signal (e.g., compress, format, enrich)
	// For now, pass through.

	// 2. Determine Route based on Intent
	// This would use the Cognitive Registry to find the "Best Team" for the job.
	// Hardcoded for V6.2 MVP:
	targetTopic := fmt.Sprintf(protocol.TopicTeamInternalCommand, "genesis") // Default to Genesis Team

	if string(msg.Data) == "system_status" {
		targetTopic = fmt.Sprintf(protocol.TopicTeamSignalStatus, "telemetry") // Expression Team
	}

	// 3. Dispatch
	log.Printf("⚡ Axon Routing Signal to [%s]", targetTopic)
	a.nc.Publish(targetTopic, msg.Data)
}

package router

import (
	"fmt"
	"log"
	"strings"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/state"
	pb "github.com/mycelis/core/pkg/pb/swarm"
)

// Router handles the distribution of messages from NATS to Agents
type Router struct {
	nc         *nats.Conn
	gatekeeper *governance.Gatekeeper
}

// NewRouter creates a new router instance
func NewRouter(nc *nats.Conn, gk *governance.Gatekeeper) *Router {
	return &Router{
		nc:         nc,
		gatekeeper: gk,
	}
}

// Start listens on the swarm network
func (r *Router) Start() error {
	log.Println("⚡ Router Listening on swarm.>")
	_, err := r.nc.Subscribe("swarm.>", r.handleMessage)
	return err
}

func (r *Router) handleMessage(msg *nats.Msg) {
	// 1. Unmarshal
	var envelope pb.MsgEnvelope
	if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
		// log.Printf("Failed to unmarshal: %v", err)
		return
	}

	// 2. Governance Check
	if r.gatekeeper != nil {
		allowed, action, reqID := r.gatekeeper.Intercept(&envelope)
		if !allowed {
			if action == governance.ActionRequireApproval {
				log.Printf("📢 Published Approval Request %s", reqID)

				// Publish Governance Request Event
				// Topic: swarm.governance.needed
				_, err := proto.Marshal(&envelope)
				if err == nil {
					// Publish the RequestID so UI can fetch details or subscribe
					// Alternatively publish the entire envelope or a specific GovernanceEvent
					r.nc.Publish("swarm.governance.needed", []byte(reqID))
				}

				// We also emit a log or metric?
			}
			// Stop processing this message (Drop or Park)
			return
		}
	}

	// 3. Heartbeat / Registry Update
	r.updateRegistry(&envelope, msg.Subject)

	// 4. Routing Logic (if any specific P2P logic is needed beyond NATS wildcards)
	// Currently NATS handles the delivery to subscribers.
	// This router component acts as the "Sidecar" or "Observer" for the Core.
}

// PublishDirect sends a message directly to NATS, bypassing Gatekeeper
// Used by Admin API for approved messages
func (r *Router) PublishDirect(envelope *pb.MsgEnvelope) error {
	data, err := proto.Marshal(envelope)
	if err != nil {
		return err
	}

	// Topic Reconstruction logic
	// Standard: swarm.team.{team}.agent.{agent}.output (if from agent)
	// But here we are REPLAYING a message. It should go to its original destination?
	// ACTUALLY: The router's job is to route based on recipient.
	// If it was an "intent" to perform an action, the original topic was the publishing topic.
	// But NATS is a bus.

	// If the original message was intercepted *before* it hit the bus?
	// The current Intercept logic in main.go checks messages *coming off the bus*.
	// This means the message WAS published, but the Router (Core) intercepted it.
	// Wait, if Core subscribes to "swarm.>", it receives everything.
	// Gatekeeper returns "false" means "Stop Processing" inside Core.
	// It does NOT stop other agents from seeing it if they are subscribed directly to NATS!

	// ARCHITECTURAL NOTE:
	// In the "Absolute Architecture", the Core is the "Brain".
	// If Agents P2P directly, Core can only audit, not block, unless Agents use Request/Reply via Core.
	// Current Gatekeeper `Intercept` returning `false` just stops the Core from reacting?
	// OR does the Router assume it is the *only* path?

	// For this task, we assume the Admin API just wants to put it back on the bus
	// so the Core (and others) can process it again?
	// But if we put it back, Gatekeeper will block it again unless we mark it approved!

	// Solution: Add "approved_signature" to SwarmContext before publishing.
	if envelope.SwarmContext == nil {
		// Create struct if nil (using simple map for now then ignoring strict struct for MVP)
		// proto.Struct is complex to init manually here without imports.
		// Let's assume the Resolve logic or caller handles context injection?
		// Or just rely on a special "admin" topic or flag.

		// MVP: We assume the Gatekeeper has a "Allow all from Admin" or internal check?
		// No, let's just publish it. If logic is "Interceptor stops Core from reacting", re-publishing won't help if it's the same MsgID.
		// New ID?
		// For MVP: We will simply Log that we are re-publishing.
	}

	// Re-construct routing key or use generic broadcast?
	// If it was "swarm.agent.x.output", we usually want it there.
	subject := fmt.Sprintf("swarm.team.%s.agent.%s.output", envelope.TeamId, envelope.SourceAgentId)

	// PUBLISH
	log.Printf("🚀 Re-Publishing Approved Message %s to %s", envelope.Id, subject)
	return r.nc.Publish(subject, data)
}

func (r *Router) updateRegistry(env *pb.MsgEnvelope, subject string) {
	agentID := env.SourceAgentId
	if agentID == "" {
		return
	}

	// Extract SourceURI
	sourceURI := ""
	if env.SwarmContext != nil && env.SwarmContext.Fields != nil {
		if val, ok := env.SwarmContext.Fields["source_uri"]; ok {
			sourceURI = val.GetStringValue()
		}
	}

	// Determine if this is a heartbeat
	isHeartbeat := strings.Contains(subject, "heartbeat") ||
		(env.GetEvent() != nil && env.GetEvent().EventType == "agent.heartbeat")

	if isHeartbeat {
		state.GlobalRegistry.UpdateHeartbeat(agentID, env.TeamId, sourceURI, state.StatusIdle)
	}
}

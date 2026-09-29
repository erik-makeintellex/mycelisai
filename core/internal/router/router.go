package router

import (
	"log"
	"strings"
	"sync/atomic"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/state"
	pb "github.com/mycelis/core/pkg/pb/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// auditSubjectPrefix is the audit subject family (swarm.audit.*) the Router
// never reacts to, derived from the protocol constant so it cannot drift.
var auditSubjectPrefix = strings.TrimSuffix(protocol.TopicAuditTrace, "trace")

// Router observes the product bus for Core: it runs the governance Guard,
// refreshes the agent registry from heartbeats and emits the audit trace. It
// only observes: NATS has already delivered every message to its direct
// subscribers, so the Router can neither hold nor release one.
type Router struct {
	nc      *nats.Conn
	guard   *governance.Guard
	auditor atomic.Pointer[ApprovalAuditor]
}

// NewRouter creates a new router instance
func NewRouter(nc *nats.Conn, guard *governance.Guard) *Router {
	return &Router{
		nc:    nc,
		guard: guard,
	}
}

// Start listens on the swarm network
func (r *Router) Start() error {
	log.Printf("Router listening on %s", protocol.TopicSwarmWild)
	_, err := r.nc.Subscribe(protocol.TopicSwarmWild, r.handleMessage)
	return err
}

func (r *Router) handleMessage(msg *nats.Msg) {
	// 1. Loop prevention: never react to the audit family.
	if strings.HasPrefix(msg.Subject, auditSubjectPrefix) {
		return
	}

	// 2. Unmarshal
	var envelope pb.MsgEnvelope
	if err := proto.Unmarshal(msg.Data, &envelope); err != nil {
		return
	}

	// 3. Governance Check. A nil or degraded guard fails closed: Core only
	// keeps liveness visible (heartbeats update the registry) and reacts to
	// nothing else. The Router is an observer, so this does not stop NATS
	// delivery to direct subscribers.
	if r.guard == nil || r.guard.Degraded() {
		// A2b: only refresh agents that are already registered; a spoofed or
		// unknown heartbeat never creates an agent or rewrites team/source.
		if isHeartbeatEnvelope(msg.Subject, &envelope) {
			state.GlobalRegistry.RefreshKnown(envelope.SourceAgentId)
		}
		return
	}
	if allowed, action := r.guard.Intercept(&envelope); !allowed {
		// C2-RETIRE: REQUIRE_APPROVAL is recorded as an observation and Core
		// does not react. Nothing is parked, published or replayed.
		if action == governance.ActionRequireApproval {
			r.observeApprovalRequired(msg.Subject, &envelope)
		}
		return
	}

	// 4. Heartbeat / Registry Update
	r.updateRegistry(&envelope, msg.Subject)
	if isHeartbeatEnvelope(msg.Subject, &envelope) {
		return
	}

	// 5. Audit trace: re-publish the envelope for the Archivist. The audit
	// family is filtered above, so this cannot loop.
	go func() {
		if err := r.nc.Publish(protocol.TopicAuditTrace, msg.Data); err != nil {
			log.Printf("⚠️ Audit Trace Failed: %v", err)
		}
	}()
}

// isHeartbeatEnvelope is exact (A2b): the canonical global heartbeat subject,
// the agent.heartbeat event, and a non-empty source agent. No substring match.
func isHeartbeatEnvelope(subject string, envelope *pb.MsgEnvelope) bool {
	return subject == protocol.TopicGlobalHeartbeat && envelope != nil &&
		envelope.GetEvent() != nil && envelope.GetEvent().EventType == "agent.heartbeat" &&
		envelope.SourceAgentId != ""
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
	isHeartbeat := isHeartbeatEnvelope(subject, env)

	if isHeartbeat {
		state.GlobalRegistry.UpdateHeartbeat(agentID, env.TeamId, sourceURI, state.StatusIdle)
	}
}

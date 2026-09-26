package governance

import (
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Guard intercepts and manages approvals. Engine.Config is swapped only under
// mu; readers take a snapshot under mu.RLock. A nil Config means degraded.
type Guard struct {
	Engine        *Engine
	PendingBuffer map[string]*pb.ApprovalRequest
	mu            sync.RWMutex
	applyMu       sync.Mutex // serializes ReplacePolicy (persist + swap)
}

func NewGuard(policyPath string) (*Guard, error) {
	engine, err := NewEngine(policyPath)
	if err != nil {
		return nil, err
	}

	return &Guard{
		Engine:        engine,
		PendingBuffer: make(map[string]*pb.ApprovalRequest),
	}, nil
}

// Intercept evaluates a message and returns (proceed bool, action string, requestID string)
func (g *Guard) Intercept(msg *pb.MsgEnvelope) (bool, string, string) {
	// Extract Context
	ctx := make(map[string]interface{})
	if msg.SwarmContext != nil {
		for k, v := range msg.SwarmContext.Fields {
			// Basic type mapping for the naive parser
			if _, ok := v.Kind.(*structpb.Value_NumberValue); ok {
				ctx[k] = v.GetNumberValue()
			} else if _, ok := v.Kind.(*structpb.Value_StringValue); ok {
				ctx[k] = v.GetStringValue()
			} else if _, ok := v.Kind.(*structpb.Value_BoolValue); ok {
				ctx[k] = v.GetBoolValue()
			}
		}
	}

	// Also mix in Event Data for context?
	// For "amount > 50", usually amount is in the event data, not the header context.
	// The policy engine might want both.
	// For this loop, let's also pull data from EventPayload if present.
	if msg.GetEvent() != nil && msg.GetEvent().Data != nil {
		for k, v := range msg.GetEvent().Data.Fields {
			if _, ok := v.Kind.(*structpb.Value_NumberValue); ok {
				ctx[k] = v.GetNumberValue()
			}
		}
	}

	// Extract intent
	intent := ""
	if msg.GetEvent() != nil {
		intent = msg.GetEvent().EventType
	}

	cfg := g.policySnapshot()
	if cfg == nil {
		// Degraded: no loaded policy is authority, so nothing is allowed.
		return false, ActionDeny, ""
	}
	action := (&Engine{Config: cfg}).Evaluate(msg.TeamId, msg.SourceAgentId, intent, ctx)

	if action == ActionAllow {
		return true, action, ""
	}

	if action == ActionDeny {
		log.Printf("DENY: Guard blocked: %s from %s", intent, msg.SourceAgentId)
		return false, action, ""
	}

	if action == ActionRequireApproval {
		reqID := g.createApprovalRequest(msg, "Policy Triggered")
		log.Printf("HALT: Guard paused: %s. Request ID: %s", intent, reqID)
		return false, action, reqID
	}

	// Unknown action (validation should make this unreachable): fail closed.
	log.Printf("DENY: Guard saw unknown policy action %q for %s", action, intent)
	return false, ActionDeny, ""
}

// policySnapshot returns the live policy pointer under the read lock. Policy
// configs are never mutated after they are swapped in, so the snapshot is safe
// to evaluate without holding the lock.
func (g *Guard) policySnapshot() *PolicyConfig {
	if g == nil {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.Engine == nil {
		return nil
	}
	return g.Engine.Config
}

func (g *Guard) createApprovalRequest(msg *pb.MsgEnvelope, reason string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	reqID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	req := &pb.ApprovalRequest{
		RequestId:       reqID,
		OriginalMessage: msg,
		Reason:          reason,
		ExpiresAt:       timestamppb.New(time.Now().Add(1 * time.Hour)),
	}

	g.PendingBuffer[reqID] = req
	return reqID
}

// ListPending returns a snapshot of all pending requests
func (g *Guard) ListPending() []*pb.ApprovalRequest {
	g.mu.RLock()
	defer g.mu.RUnlock()

	list := make([]*pb.ApprovalRequest, 0, len(g.PendingBuffer))
	for _, req := range g.PendingBuffer {
		list = append(list, req)
	}
	return list
}

// PendingRequest returns one pending request without resolving it.
func (g *Guard) PendingRequest(reqID string) (*pb.ApprovalRequest, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	req, ok := g.PendingBuffer[reqID]
	return req, ok
}

// Resolve manually approves or denies a request
func (g *Guard) Resolve(reqID string, approved bool, user string) (*pb.MsgEnvelope, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	req, exists := g.PendingBuffer[reqID]
	if !exists {
		return nil, fmt.Errorf("request %s not found", reqID)
	}

	delete(g.PendingBuffer, reqID)

	if approved {
		log.Printf("APPROVED: Request %s MANUALLY APPROVED by %s", reqID, user)
		return req.OriginalMessage, nil
	}

	log.Printf("DENIED: Request %s MANUALLY DENIED by %s", reqID, user)
	return nil, nil // Nil message means nothing to forward
}

// ValidateIngress checks raw NATS messages before they enter the Soma processing loop.
// It enforces size limits and subject allowlists.
func (g *Guard) ValidateIngress(subject string, data []byte) error {
	// 1. Size Limit (e.g., 1MB)
	if len(data) > 1024*1024 {
		return fmt.Errorf("payload too large: %d bytes", len(data))
	}

	// 2. Subject Allowlists (Basic Guard Rails)
	// Only allow specific global input channels
	// swarm.global.input.gui -> User Interface
	// swarm.global.input.sensor -> Hardware Sensors
	// swarm.global.input.cli -> Command Line
	// allowed := []string{"swarm.global.input.gui", "swarm.global.input.sensor", "swarm.global.input.cli"}
	// For now, just check prefix
	if len(subject) < 18 || subject[:18] != "swarm.global.input" {
		return fmt.Errorf("invalid ingress subject: %s", subject)
	}

	return nil
}

// GetPolicyConfig returns the current policy configuration (nil when degraded).
func (g *Guard) GetPolicyConfig() *PolicyConfig {
	return g.policySnapshot()
}

// UpdatePolicyConfig replaces the in-memory policy configuration. A non-nil
// config clears the degraded state.
func (g *Guard) UpdatePolicyConfig(cfg *PolicyConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Engine == nil {
		g.Engine = &Engine{}
	}
	g.Engine.Config = cfg
}

// ReplacePolicy persists cfg first and swaps it into memory only when persist
// succeeds, so a failed write leaves the live policy unchanged. Concurrent
// replacements are serialized so file and memory cannot diverge.
func (g *Guard) ReplacePolicy(cfg *PolicyConfig, persist func() error) error {
	if cfg == nil {
		return fmt.Errorf("policy config is required")
	}
	g.applyMu.Lock()
	defer g.applyMu.Unlock()
	if persist != nil {
		if err := persist(); err != nil {
			return err
		}
	}
	g.UpdatePolicyConfig(cfg)
	return nil
}

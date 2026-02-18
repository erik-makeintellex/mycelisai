package governance

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	pb "github.com/mycelis/core/pkg/pb/swarm"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// Guard intercepts and manages approvals
type Guard struct {
	Engine        *Engine
	PendingBuffer map[string]*pb.ApprovalRequest
	mu            sync.RWMutex
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

	action := g.Engine.Evaluate(msg.TeamId, msg.SourceAgentId, intent, ctx)

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

	return true, ActionAllow, ""
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

// GetPolicyConfig returns the current policy configuration.
func (g *Guard) GetPolicyConfig() *PolicyConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Engine.Config
}

// UpdatePolicyConfig replaces the in-memory policy configuration.
func (g *Guard) UpdatePolicyConfig(cfg *PolicyConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Engine.Config = cfg
}

// SavePolicyToFile writes the current policy configuration back to a YAML file.
func (g *Guard) SavePolicyToFile(path string) error {
	g.mu.RLock()
	cfg := g.Engine.Config
	g.mu.RUnlock()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal policy config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}

	return nil
}

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

// Gatekeeper intercepts and manages approvals
type Gatekeeper struct {
	Engine        *Engine
	PendingBuffer map[string]*pb.ApprovalRequest
	mu            sync.RWMutex
}

func NewGatekeeper(policyPath string) (*Gatekeeper, error) {
	engine, err := NewEngine(policyPath)
	if err != nil {
		return nil, err
	}

	return &Gatekeeper{
		Engine:        engine,
		PendingBuffer: make(map[string]*pb.ApprovalRequest),
	}, nil
}

// Intercept evaluates a message and returns (proceed bool, action string, requestID string)
func (g *Gatekeeper) Intercept(msg *pb.MsgEnvelope) (bool, string, string) {
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
	if msg.GetEvent() != nil {
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

func (g *Gatekeeper) createApprovalRequest(msg *pb.MsgEnvelope, reason string) string {
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
func (g *Gatekeeper) ListPending() []*pb.ApprovalRequest {
	g.mu.RLock()
	defer g.mu.RUnlock()

	list := make([]*pb.ApprovalRequest, 0, len(g.PendingBuffer))
	for _, req := range g.PendingBuffer {
		list = append(list, req)
	}
	return list
}

// Resolve manually approves or denies a request
func (g *Gatekeeper) Resolve(reqID string, approved bool, user string) (*pb.MsgEnvelope, error) {
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

// ProcessSignal handles an admin's decision (Legacy / Event based)
func (g *Gatekeeper) ProcessSignal(signal *pb.ApprovalSignal) *pb.MsgEnvelope {
	// Re-using Resolve logic to keep DRY
	msg, _ := g.Resolve(signal.RequestId, signal.Approved, signal.UserSignature)
	return msg
}

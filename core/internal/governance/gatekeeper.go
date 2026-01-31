package governance

import (
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/mycelis/core/pkg/pb/swarm"
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
	// Extract Context (Convert Struct to Map not fully impl here, passing empty for MVP)
	// In real impl, would marshal struct to map
	ctx := make(map[string]interface{})

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
		log.Printf("⛔ Gatekeeper DENIED: %s from %s", intent, msg.SourceAgentId)
		return false, action, ""
	}

	if action == ActionRequireApproval {
		reqID := g.createApprovalRequest(msg, "Policy Triggered")
		log.Printf("✋ Gatekeeper Halted: %s. Request ID: %s", intent, reqID)
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

// ProcessSignal handles an admin's decision
func (g *Gatekeeper) ProcessSignal(signal *pb.ApprovalSignal) *pb.MsgEnvelope {
	g.mu.Lock()
	defer g.mu.Unlock()

	req, exists := g.PendingBuffer[signal.RequestId]
	if !exists {
		return nil
	}

	delete(g.PendingBuffer, signal.RequestId)

	if signal.Approved {
		log.Printf("✅ Request %s APPROVED by %s", signal.RequestId, signal.UserSignature)
		return req.OriginalMessage
	}

	log.Printf("❌ Request %s REJECTED by %s", signal.RequestId, signal.UserSignature)
	return nil
}

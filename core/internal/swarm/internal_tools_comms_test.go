package swarm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mycelis/core/internal/comms"
)

type fakeCommsProvider struct{}

func (p *fakeCommsProvider) Info() comms.ProviderInfo {
	return comms.ProviderInfo{Name: "slack", Channel: "chat", Description: "Slack", Configured: true}
}

func (p *fakeCommsProvider) Send(_ context.Context, req comms.SendRequest) (comms.SendResult, error) {
	return comms.SendResult{
		Provider: "slack",
		Status:   "sent",
		Metadata: map[string]any{"recipient": req.Recipient},
	}, nil
}

func TestHandleSendExternalMessage(t *testing.T) {
	g := comms.NewGateway()
	g.Register(&fakeCommsProvider{})

	r := NewInternalToolRegistry(InternalToolDeps{Comms: g})
	out, err := r.handleSendExternalMessage(context.Background(), map[string]any{
		"provider":  "slack",
		"recipient": "#ops",
		"message":   "health check",
	})
	if err != nil {
		t.Fatalf("handleSendExternalMessage: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if parsed["status"] != "sent" {
		t.Fatalf("status = %v", parsed["status"])
	}
}

func TestHandleSendExternalMessage_NoGateway(t *testing.T) {
	r := NewInternalToolRegistry(InternalToolDeps{})
	if _, err := r.handleSendExternalMessage(context.Background(), map[string]any{
		"provider": "slack",
		"message":  "x",
	}); err == nil {
		t.Fatal("expected error when comms gateway unavailable")
	}
}

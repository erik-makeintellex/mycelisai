package provisioning

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mycelis/core/internal/cognitive"
)

// ServiceManifest represents the immutable config for a Service/Agent
type ServiceManifest struct {
	Name         string                 `json:"name"`
	Runtime      map[string]interface{} `json:"runtime"`      // Docker config, Schedule, etc.
	Connectivity ConnectivityConfig     `json:"connectivity"` // Pub/Sub wiring
	Permissions  []string               `json:"permissions"`  // MCP Scopes
}

type ConnectivityConfig struct {
	Subscriptions []string `json:"subscriptions"`
	Publications  []string `json:"publications"`
}

// Engine converts intent into manifests
type Engine struct {
	Cognitive *cognitive.Router
}

func NewEngine(router *cognitive.Router) *Engine {
	return &Engine{Cognitive: router}
}

// Draft generates a ServiceManifest from natural language intent
func (e *Engine) Draft(ctx context.Context, intent string) (*ServiceManifest, error) {
	// 1. Construct the Prompt
	prompt := fmt.Sprintf(`You are THE ARCHITECT. Your goal is to design a Microservice Manifest based on User Intent.
    
INTENT: "%s"

OUTPUT SCHEMA (Strict JSON):
{
  "name": "svc-<name>",
  "runtime": { "schedule": "..." },
  "connectivity": {
    "subscriptions": ["swarm.telemetry.>"],
    "publications": ["swarm.events.>"]
  },
  "permissions": ["mcp.<tool>.<verb>"]
}

Return ONLY value valid JSON. No markdown.`, intent)

	// 2. Call Cognitive Layer (Architect Profile)
	req := cognitive.InferRequest{
		Profile: "architect",
		Prompt:  prompt,
	}

	resp, err := e.Cognitive.InferWithContract(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cognitive failure: %w", err)
	}

	// 3. Unmarshal & Validate
	var manifest ServiceManifest
	if err := json.Unmarshal([]byte(resp.Text), &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest generated: %w", err)
	}

	return &manifest, nil
}

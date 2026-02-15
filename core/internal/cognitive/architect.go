package cognitive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// MetaArchitect decomposes high-level intent into a MissionBlueprint
// by leveraging the cognitive Router for LLM inference.
type MetaArchitect struct {
	brain *Router
}

// NewMetaArchitect creates a MetaArchitect wired to the given Router.
func NewMetaArchitect(brain *Router) *MetaArchitect {
	return &MetaArchitect{brain: brain}
}

// GenerateBlueprint takes a natural-language intent and returns a structured
// MissionBlueprint by prompting the LLM and parsing the JSON response.
func (m *MetaArchitect) GenerateBlueprint(ctx context.Context, intent string) (*protocol.MissionBlueprint, error) {
	// 1. Construct prompt instructing strict JSON output matching MissionBlueprint
	prompt := fmt.Sprintf(`You are THE META-ARCHITECT. Decompose the following intent into a Mission Blueprint.

INTENT: "%s"

OUTPUT SCHEMA (Strict JSON, no markdown):
{
  "mission_id": "mission-<short-id>",
  "intent": "<the original intent>",
  "teams": [
    {
      "name": "<team-name>",
      "role": "<team purpose>",
      "agents": [
        {
          "id": "<agent-id>",
          "role": "<specialist role>",
          "system_prompt": "<instructions>",
          "inputs": ["<nats topics>"],
          "outputs": ["<nats topics>"]
        }
      ]
    }
  ],
  "constraints": ["<any constraints>"]
}

Return ONLY valid JSON. No markdown fences.`, intent)

	req := InferRequest{
		Profile: "architect",
		Prompt:  prompt,
	}

	resp, err := m.brain.InferWithContract(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("meta-architect inference failed: %w", err)
	}

	// 2. Strip markdown code fences (same pattern as provisioning/engine.go)
	text := resp.Text
	if idx := strings.Index(text, "```json"); idx != -1 {
		text = text[idx+7:]
		if end := strings.LastIndex(text, "```"); end != -1 {
			text = text[:end]
		}
	} else if idx := strings.Index(text, "```"); idx != -1 {
		text = text[idx+3:]
		if end := strings.LastIndex(text, "```"); end != -1 {
			text = text[:end]
		}
	}
	text = strings.TrimSpace(text)

	// 3. Unmarshal and validate
	var blueprint protocol.MissionBlueprint
	if err := json.Unmarshal([]byte(text), &blueprint); err != nil {
		return nil, fmt.Errorf("invalid blueprint JSON: %w (raw: %s)", err, text)
	}

	if blueprint.MissionID == "" {
		blueprint.MissionID = fmt.Sprintf("mission-%d", time.Now().UnixMilli())
	}
	blueprint.Intent = intent

	return &blueprint, nil
}

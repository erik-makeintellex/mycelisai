package cognitive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// MCPServerCapability describes an installed or installable MCP server for architect context.
type MCPServerCapability struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`                // "installed", "available" (in library)
	Tools       []string `json:"tools,omitempty"`        // discovered tool names (installed only)
	RequiredEnv []string `json:"required_env,omitempty"` // env vars needed (e.g. GITHUB_TOKEN)
}

// SystemCapabilities provides the Meta-Architect with awareness of what tools,
// MCP servers, and credentials are available in the organism. Populated by
// the caller (main.go) to avoid circular dependencies.
type SystemCapabilities struct {
	InternalTools map[string]string     // tool name → description
	MCPServers    []MCPServerCapability // installed + library entries
}

// MetaArchitect decomposes high-level intent into a MissionBlueprint
// by leveraging the cognitive Router for LLM inference.
type MetaArchitect struct {
	brain        *Router
	capabilities *SystemCapabilities
}

// NewMetaArchitect creates a MetaArchitect wired to the given Router.
func NewMetaArchitect(brain *Router) *MetaArchitect {
	return &MetaArchitect{brain: brain}
}

// SetCapabilities provides the architect with live system context for blueprint generation.
func (m *MetaArchitect) SetCapabilities(caps *SystemCapabilities) {
	m.capabilities = caps
}

// GenerateBlueprint takes a natural-language intent and returns a structured
// MissionBlueprint by prompting the LLM and parsing the JSON response.
func (m *MetaArchitect) GenerateBlueprint(ctx context.Context, intent string) (*protocol.MissionBlueprint, error) {
	// 1. Build system context block for the prompt
	capBlock := m.buildCapabilitiesBlock()

	// 2. Construct prompt instructing strict JSON output matching MissionBlueprint
	prompt := fmt.Sprintf(`You are THE META-ARCHITECT. Decompose the following intent into a Mission Blueprint.

INTENT: "%s"
%s
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
          "model": "<cognitive profile or model name>",
          "system_prompt": "<instructions>",
          "tools": ["<tool names from available tools>"],
          "inputs": ["<nats topics>"],
          "outputs": ["<nats topics>"]
        }
      ]
    }
  ],
  "constraints": ["<any constraints>"],
  "requirements": [
    {
      "type": "<mcp_server|api_key|env_var|credential>",
      "name": "<identifier>",
      "description": "<why it is needed>",
      "required": true
    }
  ]
}

RULES:
1. Agents MUST reference tools from the Available Tools list. Do NOT invent tool names.
2. If the mission needs an external API (GitHub, Slack, email, etc.), add a "requirements" entry
   specifying the MCP server to install and any API keys / env vars needed.
3. If an MCP server is already installed, agents can use its tools directly — no requirement needed.
4. For sensor/polling agents, set role to "sensory". For LLM-reasoning agents, use "cognitive".
5. NATS topics follow the pattern: topic.<domain>.<action> (e.g., topic.weather.raw, topic.email.sent).
6. Keep teams LEAN: 2-4 agents per team maximum. Each agent must have a distinct role.
   Prefer fewer, more capable agents over many narrow specialists.
7. Assign a "model" to each agent based on task complexity:
   - Heavy reasoning/architecture/code-gen: use the largest available model
   - Routine summarization/formatting/polling: use smaller, faster models
   - If unsure, leave model as "" (system default will be used)
8. Each team should have clear input/output contracts via NATS topics. Avoid circular dependencies between teams.

Return ONLY valid JSON. No markdown fences.`, intent, capBlock)

	req := InferRequest{
		Profile: "architect",
		Prompt:  prompt,
	}

	resp, err := m.brain.InferWithContract(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("meta-architect inference failed: %w", err)
	}

	// 3. Extract JSON from LLM response (strip fences, preamble, trailing text)
	text := extractJSON(resp.Text)

	// 4. Unmarshal and validate
	var blueprint protocol.MissionBlueprint
	if err := json.Unmarshal([]byte(text), &blueprint); err != nil {
		// Truncate raw text in error to avoid breaking downstream JSON serialization
		raw := text
		if len(raw) > 200 {
			raw = raw[:200] + "..."
		}
		return nil, fmt.Errorf("invalid blueprint JSON: %w (truncated: %s)", err, raw)
	}

	if blueprint.MissionID == "" {
		blueprint.MissionID = fmt.Sprintf("mission-%d", time.Now().UnixMilli())
	}
	blueprint.Intent = intent

	// 5. Mark requirements that are already satisfied
	m.markInstalledRequirements(&blueprint)

	return &blueprint, nil
}

// buildCapabilitiesBlock generates the system context section for the architect prompt.
func (m *MetaArchitect) buildCapabilitiesBlock() string {
	if m.capabilities == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n## SYSTEM CAPABILITIES\n\n")

	// Internal tools
	if len(m.capabilities.InternalTools) > 0 {
		sb.WriteString("### Available Internal Tools\n")
		for name, desc := range m.capabilities.InternalTools {
			sb.WriteString(fmt.Sprintf("- `%s`: %s\n", name, desc))
		}
		sb.WriteString("\n")
	}

	// MCP servers (installed + library)
	if len(m.capabilities.MCPServers) > 0 {
		var installed, available []MCPServerCapability
		for _, s := range m.capabilities.MCPServers {
			if s.Status == "installed" {
				installed = append(installed, s)
			} else {
				available = append(available, s)
			}
		}

		if len(installed) > 0 {
			sb.WriteString("### Installed MCP Servers (tools ready to use)\n")
			for _, s := range installed {
				if len(s.Tools) > 0 {
					sb.WriteString(fmt.Sprintf("- **%s**: tools=[%s]\n", s.Name, strings.Join(s.Tools, ", ")))
				} else {
					sb.WriteString(fmt.Sprintf("- **%s**: (no tools discovered)\n", s.Name))
				}
			}
			sb.WriteString("\n")
		}

		if len(available) > 0 {
			sb.WriteString("### MCP Servers Available for Installation\n")
			sb.WriteString("These can be installed if the mission requires external service access.\n")
			for _, s := range available {
				envNote := ""
				if len(s.RequiredEnv) > 0 {
					envNote = fmt.Sprintf(" (requires: %s)", strings.Join(s.RequiredEnv, ", "))
				}
				sb.WriteString(fmt.Sprintf("- **%s**: %s%s\n", s.Name, s.Description, envNote))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// markInstalledRequirements checks blueprint requirements against installed MCP servers.
func (m *MetaArchitect) markInstalledRequirements(bp *protocol.MissionBlueprint) {
	if m.capabilities == nil || len(bp.Requirements) == 0 {
		return
	}

	installed := make(map[string]bool)
	for _, s := range m.capabilities.MCPServers {
		if s.Status == "installed" {
			installed[s.Name] = true
		}
	}

	for i := range bp.Requirements {
		if bp.Requirements[i].Type == "mcp_server" && installed[bp.Requirements[i].Name] {
			bp.Requirements[i].Installed = true
		}
	}
}

// extractJSON strips markdown fences, preamble text, and trailing text from
// an LLM response to isolate a JSON object. Falls back to finding the first
// '{' and last '}' if fence stripping doesn't yield valid JSON boundaries.
func extractJSON(raw string) string {
	text := raw

	// 1. Strip markdown code fences
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

	// 2. If still not starting with '{', find the first '{' and last '}'
	if len(text) > 0 && text[0] != '{' {
		start := strings.Index(text, "{")
		if start != -1 {
			end := strings.LastIndex(text, "}")
			if end > start {
				text = text[start : end+1]
			}
		}
	}

	return strings.TrimSpace(text)
}

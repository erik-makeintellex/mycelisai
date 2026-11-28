package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func isDeploymentMCPAdditionRequest(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "standard deployment mcp") ||
		strings.Contains(lower, "custom developed mcp") ||
		strings.Contains(lower, "mcp addition") ||
		strings.Contains(lower, "(mcp_group:") ||
		strings.Contains(lower, "(mcp_id:") {
		return true
	}
	return requestContainsAny(lower, []string{"add standard mcp", "attach standard mcp", "scaffold custom mcp", "bind custom mcp"}) &&
		requestContainsAny(lower, []string{"team", "deployment"})
}

func parseDeploymentMCPAdditionMeta(text string) (teamID, teamName, mcpID, mcpTitle, mcpGroup string) {
	teamID, teamName = extractTargetTeamFromRequest(text)
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)

	mcpID = extractTagValue(trimmed, lower, "(mcp_id:")
	mcpGroup = extractTagValue(trimmed, lower, "(mcp_group:")
	mcpTitle = extractTagValue(trimmed, lower, "(mcp_title:")

	if mcpGroup == "" {
		if strings.Contains(lower, "custom developed") || strings.Contains(lower, "custom mcp") || strings.HasPrefix(mcpID, "custom-") {
			mcpGroup = "custom"
		} else {
			mcpGroup = "standard"
		}
	}
	if mcpID == "" {
		for _, candidate := range []string{
			"custom-python-fastmcp", "custom-ts-mcp-sdk", "custom-http-sse-mcp", "custom-nats-service-bridge",
			"filesystem", "github", "fetch", "postgres", "sqlite", "memory", "puppeteer",
			"git", "docker-kubernetes", "sequential-thinking", "slack", "brave-search",
		} {
			if strings.Contains(lower, candidate) {
				mcpID = candidate
				break
			}
		}
	}
	if mcpID == "" {
		if mcpGroup == "custom" {
			mcpID = "custom-python-fastmcp"
		} else {
			mcpID = "filesystem"
		}
	}
	if mcpTitle == "" {
		mcpTitle = titleCaseTeamID(mcpID)
	}
	return teamID, teamName, mcpID, mcpTitle, mcpGroup
}

func extractTagValue(trimmed, lower, marker string) string {
	idx := strings.Index(lower, marker)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(trimmed[idx+len(marker):])
	for i, r := range rest {
		if r == ')' || r == ',' || r == ';' {
			rest = rest[:i]
			break
		}
	}
	return strings.Trim(strings.TrimSpace(rest), "\"'`")
}

func inferDeploymentMCPAdditionTeamCalls(latestRequest string) []protocol.PlannedToolCall {
	teamID, teamName, mcpID, mcpTitle, mcpGroup := parseDeploymentMCPAdditionMeta(latestRequest)
	groupLabel := "Standard Deployment MCP"
	if mcpGroup == "custom" {
		groupLabel = "Custom Developed MCP"
	}
	mcpToolName := "mcp_" + strings.ReplaceAll(mcpID, "-", "_")
	return []protocol.PlannedToolCall{
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id":     teamID,
				"name":        teamName,
				"description": fmt.Sprintf("Binds %s (%s / %s) to %s with Ideal Delivery Reviewer verification.", mcpTitle, mcpID, groupLabel, teamName),
				"inputs":      []string{fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID), protocol.TopicMCPOutputsRegistered, protocol.TopicGlobalBroadcast},
				"deliveries":  []string{fmt.Sprintf(protocol.TopicTeamSignalResult, teamID), fmt.Sprintf(protocol.TopicTeamSignalStatus, teamID)},
				"members": map[string]any{
					"architect":         "council-architect",
					"developer":         "prime-development-agent",
					"mcp-integrator":    "agui-design-architect-agent",
					"delivery-reviewer": "council-sentry",
				},
				"tools": []string{"write_file", "publish_signal", "read_file", mcpToolName},
			},
		},
	}
}

func deploymentMCPAdditionBundleSpecs(latestRequest string) ([]deliverableBundleFileSpec, string) {
	teamID, teamName, mcpID, mcpTitle, mcpGroup := parseDeploymentMCPAdditionMeta(latestRequest)
	slug := sanitizeCodexBundleSlug(teamID + "-" + mcpID)
	groupLabel := "Standard Deployment MCP"
	if mcpGroup == "custom" {
		groupLabel = "Custom Developed MCP"
	}

	files := []deliverableBundleFileSpec{
		{
			name:        "mcp_binding_manifest.yaml",
			title:       fmt.Sprintf("%s — %s Binding Manifest (YAML)", teamName, mcpTitle),
			contentType: "text/yaml",
			content: "team_id: " + strconvQuoteSafe(teamID) + "\n" +
				"team_name: " + strconvQuoteSafe(teamName) + "\n" +
				"mcp_addition:\n" +
				"  id: " + strconvQuoteSafe(mcpID) + "\n" +
				"  title: " + strconvQuoteSafe(mcpTitle) + "\n" +
				"  grouping: " + strconvQuoteSafe(mcpGroup) + "\n" +
				"  category_label: " + strconvQuoteSafe(groupLabel) + "\n" +
				"  signal_input: " + strconvQuoteSafe(protocol.TopicMCPOutputsRegistered) + "\n" +
				"  team_result_channel: " + strconvQuoteSafe(fmt.Sprintf(protocol.TopicTeamSignalResult, teamID)) + "\n" +
				"  reviewer_gate: \"council-sentry\"\n",
		},
		{
			name:        "mcp_integration_guide.md",
			title:       fmt.Sprintf("%s (%s) — Deployment Integration Guide", mcpTitle, groupLabel),
			contentType: "text/markdown",
			content: "# " + mcpTitle + " (" + groupLabel + ") — Deployment Addition\n\n" +
				"## Target Deployment\n" +
				"- **Team:** `" + teamName + "` (`" + teamID + "`)\n" +
				"- **MCP Identifier:** `" + mcpID + "`\n" +
				"- **Grouping:** `" + groupLabel + "` (`" + mcpGroup + "`)\n\n" +
				"## Governed Execution & Verification\n" +
				"- Bound to team member toolset as `mcp_" + strings.ReplaceAll(mcpID, "-", "_") + "`.\n" +
				"- Normalized outputs route through `" + fmt.Sprintf(protocol.TopicTeamSignalResult, teamID) + "` and are audited by `council-sentry`.\n",
		},
	}

	if mcpGroup == "custom" {
		files = append(files, deliverableBundleFileSpec{
			name:        "custom_mcp_server_scaffold.py",
			title:       fmt.Sprintf("%s — Custom Developed MCP Server Scaffold", mcpTitle),
			contentType: "text/markdown",
			content: "\"\"\"Custom Developed MCP Server Scaffold for " + teamName + " (" + mcpID + ")\"\"\"\n" +
				"import json\nimport sys\n\n" +
				"TOOLS = [\n" +
				"    {\n" +
				"        \"name\": \"" + strings.ReplaceAll(mcpID, "-", "_") + "_execute\",\n" +
				"        \"description\": \"Custom developed MCP tool bound to " + teamName + "\",\n" +
				"        \"inputSchema\": {\"type\": \"object\", \"properties\": {\"payload\": {\"type\": \"string\"}}},\n" +
				"    }\n" +
				"]\n\n" +
				"if __name__ == \"__main__\":\n" +
				"    print(json.dumps({\"mcp_server\": \"" + mcpID + "\", \"team_id\": \"" + teamID + "\", \"tools\": TOOLS}))\n",
		})
	}
	return files, slug
}

package capabilities

import (
	"encoding/json"
	"strings"

	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/mcp"
	"github.com/mycelis/core/internal/searchcap"
)

func manifestFromExchangeCapability(cap exchange.CapabilityDefinition) Manifest {
	return Manifest{
		ID:                  cap.ID,
		DisplayName:         cap.Label,
		Kind:                "exchange_capability",
		Source:              cap.Source,
		Status:              "available",
		RiskClass:           cap.RiskClass,
		Description:         cap.Description,
		DefaultAllowedRoles: append([]string(nil), cap.DefaultAllowedRoles...),
		AuditRequired:       cap.AuditRequired,
		ApprovalRequired:    cap.ApprovalRequired,
		Metadata:            map[string]any{"projection": "exchange_capability_registry"},
	}
}

func manifestFromSearchStatus(status searchcap.Status) Manifest {
	state := "disabled"
	if status.Enabled && status.Configured {
		state = "enabled"
	} else if status.Enabled {
		state = "blocked"
	}
	metadata := map[string]any{
		"provider":                    status.Provider,
		"configured":                  status.Configured,
		"supports_local_sources":      status.SupportsLocalSources,
		"supports_public_web":         status.SupportsPublicWeb,
		"direct_soma_interaction":     status.DirectSomaInteraction,
		"requires_hosted_api_token":   status.RequiresHostedAPIToken,
		"searxng_endpoint_configured": status.SearXNGEndpointConfigured,
		"max_results":                 status.MaxResults,
		"next_actions":                status.NextActions,
	}
	if status.Blocker != nil {
		metadata["blocker"] = status.Blocker
	}
	tool := strings.TrimSpace(status.SomaToolName)
	if tool == "" {
		tool = "web_search"
	}
	return Manifest{
		ID:                  "search:" + tool,
		DisplayName:         "Mycelis Search",
		Kind:                "search_capability",
		Source:              "searchcap",
		Status:              state,
		RiskClass:           "medium-risk",
		Description:         "Governed search capability exposed to Soma through the standard web_search tool.",
		ToolRefs:            []string{tool},
		DefaultAllowedRoles: []string{"soma", "team_lead", "specialist"},
		AuditRequired:       true,
		ApprovalRequired:    false,
		Metadata:            metadata,
	}
}

func manifestFromMCPServer(srv mcp.ServerConfig) Manifest {
	return Manifest{
		ID:                  "mcp_server:" + srv.ID.String(),
		DisplayName:         srv.Name,
		Kind:                "mcp_server",
		Source:              "mcp",
		Status:              normalizeStatus(srv.Status, "installed"),
		RiskClass:           "high-risk",
		Description:         "Installed MCP server registered in the runtime MCP registry.",
		DefaultAllowedRoles: []string{"soma", "team_lead", "mcp"},
		AuditRequired:       true,
		ApprovalRequired:    false,
		Metadata: map[string]any{
			"server_id": srv.ID.String(),
			"transport": srv.Transport,
			"command":   srv.Command,
			"url":       srv.URL,
		},
	}
}

func manifestFromMCPTool(tool mcp.ToolDef) Manifest {
	metadata := map[string]any{
		"tool_id":   tool.ID.String(),
		"server_id": tool.ServerID.String(),
	}
	if tool.ServerName != "" {
		metadata["server_name"] = tool.ServerName
	}
	if len(tool.InputSchema) > 0 && json.Valid(tool.InputSchema) {
		var schema any
		if err := json.Unmarshal(tool.InputSchema, &schema); err == nil {
			metadata["input_schema"] = schema
		}
	}
	ref := "mcp:" + tool.ServerName + "/" + tool.Name
	if strings.TrimSpace(tool.ServerName) == "" {
		ref = "mcp:" + tool.ServerID.String() + "/" + tool.Name
	}
	return Manifest{
		ID:                  "mcp_tool:" + tool.ID.String(),
		DisplayName:         tool.Name,
		Kind:                "mcp_tool",
		Source:              "mcp",
		Status:              "installed",
		RiskClass:           "high-risk",
		Description:         tool.Description,
		ToolRefs:            []string{ref},
		DefaultAllowedRoles: []string{"soma", "team_lead", "mcp"},
		AuditRequired:       true,
		ApprovalRequired:    false,
		Metadata:            metadata,
	}
}

func manifestFromMCPLibraryEntry(category string, entry mcp.LibraryEntry) Manifest {
	return Manifest{
		ID:                  "mcp_library:" + entry.Name,
		DisplayName:         displayName(entry.Title, entry.Name),
		Kind:                "mcp_library_entry",
		Source:              "mcp_library",
		Status:              "available",
		RiskClass:           riskForLibraryEntry(entry),
		Description:         entry.Description,
		ToolRefs:            []string{"mcp:" + entry.Name + "/*"},
		DefaultAllowedRoles: []string{"admin", "soma", "team_lead", "mcp"},
		AuditRequired:       true,
		ApprovalRequired:    entry.HasRequiredSecretEnvVar(),
		Metadata: map[string]any{
			"category":            category,
			"transport":           entry.Transport,
			"tags":                append([]string(nil), entry.Tags...),
			"required_env":        entry.DeclaredEnvKeys(),
			"tool_set":            entry.ToolSet,
			"deployment_boundary": entry.DeploymentBoundary,
			"repository":          entry.Repository,
			"homepage":            entry.Homepage,
		},
	}
}

func riskForInternalTool(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "local_command"):
		return "high-risk"
	case strings.Contains(lower, "publish"), strings.Contains(lower, "exchange"), strings.Contains(lower, "write"):
		return "medium-risk"
	default:
		return "low-risk"
	}
}

func riskForLibraryEntry(entry mcp.LibraryEntry) string {
	if entry.HasRequiredSecretEnvVar() {
		return "high-risk"
	}
	for _, tag := range entry.Tags {
		switch strings.ToLower(tag) {
		case "browser", "search", "web", "api", "github":
			return "medium-risk"
		}
	}
	return "medium-risk"
}

func normalizeStatus(raw, fallback string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	return raw
}

func displayName(title, fallback string) string {
	if strings.TrimSpace(title) != "" {
		return title
	}
	return fallback
}

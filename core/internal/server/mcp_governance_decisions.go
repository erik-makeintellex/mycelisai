package server

import "github.com/mycelis/core/internal/mcp"

func buildMCPLibraryGovernanceDecision(entry *mcp.LibraryEntry, ctx mcpGovernanceContext) mcpGovernanceDecision {
	locality := mcpLibraryLocality(entry)
	riskLevel := mcpLibraryRiskLevel(entry)
	deploymentBoundary := mcpLibraryDeploymentBoundary(entry)
	credentialBoundary := mcpLibraryCredentialBoundary(entry)

	switch {
	case deploymentBoundary == "external_saas" && credentialBoundary == "secret_required":
		return requiredMCPDecision(ctx, "credentialed_external_mcp_config", []string{
			"Credentialed external SaaS MCPs require an explicit approval gate even when launched through a local stdio wrapper.",
			"Curated-library install remains the only governed install path for this MCP configuration.",
		})
	case deploymentBoundary == "external_saas":
		return requiredMCPDecision(ctx, "external_service_mcp_config", []string{
			"External service MCP configuration still requires an explicit approval gate.",
			"Curated-library install remains the only governed install path for this MCP configuration.",
		})
	case locality == "remote":
		return requiredMCPDecision(ctx, "remote_mcp_config", []string{
			"Remote MCP configuration still requires an explicit approval gate.",
			"Owner-scoped settings installs auto-allow only local-first MCP configuration.",
		})
	case isOwnedGroupMCPConfig(ctx):
		return allowedMCPDecision(ctx, "user_owned_mcp_config", []string{
			"Owned MCP configuration from the MCP settings page is auto-allowed for the current root/owner user group.",
		})
	case riskLevel == "high":
		return requiredMCPDecision(ctx, "capability_risk", []string{
			"This MCP configuration crosses a higher-risk boundary and still needs an explicit approval gate.",
		})
	default:
		return allowedMCPDecision(ctx, "curated_mcp_config", []string{
			"Curated local-first MCP configuration is within the default policy boundary.",
		})
	}
}

func buildMCPConfigGovernanceDecision(ctx mcpGovernanceContext, locality, riskLevel string) mcpGovernanceDecision {
	switch {
	case locality == "remote":
		return requiredMCPDecision(ctx, "remote_mcp_config", []string{
			"Remote MCP configuration still requires an explicit approval gate.",
			"Owner-scoped settings installs auto-allow only local-first MCP configuration.",
		})
	case isOwnedGroupMCPConfig(ctx):
		return allowedMCPDecision(ctx, "user_owned_mcp_config", []string{
			"Owned MCP configuration from the MCP settings page is auto-allowed for the current root/owner user group.",
		})
	case riskLevel == "high":
		return requiredMCPDecision(ctx, "capability_risk", []string{
			"This MCP configuration crosses a higher-risk boundary and still needs an explicit approval gate.",
		})
	default:
		return allowedMCPDecision(ctx, "curated_mcp_config", []string{
			"Curated local-first MCP configuration is within the default policy boundary.",
		})
	}
}

func allowedMCPDecision(ctx mcpGovernanceContext, reason string, reasons []string) mcpGovernanceDecision {
	return mcpGovernanceDecision{
		Decision:         "allow",
		ApprovalRequired: false,
		ApprovalMode:     "auto_allowed",
		ApprovalReason:   reason,
		SourceSurface:    ctx.SourceSurface,
		ConfigScope:      ctx.ConfigScope,
		GroupID:          ctx.GroupID,
		Reasons:          reasons,
	}
}

func requiredMCPDecision(ctx mcpGovernanceContext, reason string, reasons []string) mcpGovernanceDecision {
	decision := allowedMCPDecision(ctx, reason, reasons)
	decision.Decision = "require_approval"
	decision.ApprovalRequired = true
	decision.ApprovalMode = "required"
	return decision
}

package capabilities

import "strings"

func completeManifestState(m *Manifest) {
	if m.CapabilityID == "" {
		m.CapabilityID = m.ID
	}
	if m.ManifestVersion == "" {
		m.ManifestVersion = firstNonEmpty(m.Version, ManifestVersion)
	}
	if m.Version == "" {
		m.Version = m.ManifestVersion
	}
	if m.Purpose == "" {
		m.Purpose = m.Description
	}
	if m.Description == "" {
		m.Description = m.Purpose
	}
	if len(m.AllowedRoles) == 0 {
		m.AllowedRoles = append([]string(nil), m.DefaultAllowedRoles...)
	}
	if len(m.DefaultAllowedRoles) == 0 {
		m.DefaultAllowedRoles = append([]string(nil), m.AllowedRoles...)
	}
	m.ApprovalPosture = approvalPosture(m.ApprovalRequired)
	m.Health = firstNonEmpty(m.Health, healthForStatus(m.Status))
	m.LastProbeStatus = firstNonEmpty(m.LastProbeStatus, m.Status, "unknown")
	m.InputSchemaRef = firstNonEmpty(m.InputSchemaRef, inputSchemaRefForKind(m.Kind, m.ID))
	m.OutputSchemaRef = firstNonEmpty(m.OutputSchemaRef, outputSchemaRefForKind(m.Kind, m.ID))
	m.FailurePosture = firstNonEmpty(m.FailurePosture, failurePostureForHealth(m.Health))
	m.RecoveryPosture = firstNonEmpty(m.RecoveryPosture, recoveryPostureForKind(m.Kind, m.Health))
	m.AuditPolicy = firstNonEmpty(m.AuditPolicy, auditPolicy(m.AuditRequired))
	m.SecretRefPolicy = firstNonEmpty(m.SecretRefPolicy, secretRefPolicy(m))
	m.Owner = firstNonEmpty(m.Owner, ownerForKind(m.Kind, m.Source))
	if m.LastProbeAt.IsZero() {
		m.LastProbeAt = m.DerivedAt
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = m.DerivedAt
	}
}

func approvalPosture(required bool) string {
	if required {
		return "required"
	}
	return "not_required"
}

func auditPolicy(required bool) string {
	if required {
		return "required"
	}
	return "not_required"
}

func healthForStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "available", "enabled", "installed", "allowlisted", "active", "online":
		return "healthy"
	case "disabled":
		return "disabled"
	case "blocked", "degraded", "offline", "unavailable":
		return "degraded"
	default:
		return "unknown"
	}
}

func failurePostureForHealth(health string) string {
	switch health {
	case "healthy":
		return "surface_runtime_error_and_audit"
	case "disabled":
		return "surface_disabled_state"
	case "degraded":
		return "surface_degraded_state"
	default:
		return "surface_unknown_state"
	}
}

func recoveryPostureForKind(kind, health string) string {
	if health == "healthy" {
		return "retry_with_audit"
	}
	switch kind {
	case "mcp_server", "mcp_tool", "mcp_library_entry":
		return "reapply_or_reconnect_mcp"
	case "search_capability":
		return "correct_search_configuration_and_refresh"
	case "host_command":
		return "review_host_allowlist_and_refresh"
	default:
		return "refresh_capability_manifest"
	}
}

func inputSchemaRefForKind(kind, id string) string {
	switch kind {
	case "mcp_tool":
		return "mcp.tool." + id + ".input_schema"
	case "search_capability":
		return "SearchQuery"
	case "host_command":
		return "HostActionRequest"
	case "exchange_capability":
		return "ExchangeCapabilityInput"
	case "mcp_server", "mcp_library_entry":
		return "MCPServerConfig"
	default:
		return "ToolArgs"
	}
}

func outputSchemaRefForKind(kind, id string) string {
	switch kind {
	case "mcp_tool":
		return "mcp.tool." + id + ".output_schema"
	case "search_capability":
		return "ResearchSummary"
	case "host_command":
		return "HostActionResult"
	case "exchange_capability":
		return "ExchangeItem"
	case "mcp_server", "mcp_library_entry":
		return "MCPToolResult"
	default:
		return "ToolResult"
	}
}

func secretRefPolicy(m *Manifest) string {
	if m == nil {
		return "no_raw_secrets"
	}
	if _, ok := m.Metadata["required_env"]; ok {
		return "declared_env_refs_only"
	}
	if _, ok := m.Metadata["allowlist_env_field"]; ok {
		return "declared_env_refs_only"
	}
	return "no_raw_secrets"
}

func ownerForKind(kind, source string) string {
	switch kind {
	case "mcp_server", "mcp_tool", "mcp_library_entry":
		return "runtime_capability_mcp"
	case "host_command":
		return "runtime_capability_host"
	case "search_capability":
		return "runtime_capability_search"
	default:
		if source != "" {
			return "runtime_capability_" + source
		}
		return "runtime_capability"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

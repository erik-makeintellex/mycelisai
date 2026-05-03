package server

import (
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/mycelis/core/internal/mcp"
)

type mcpGovernanceContext struct {
	SourceSurface string `json:"source_surface,omitempty"`
	ConfigScope   string `json:"config_scope,omitempty"`
	GroupID       string `json:"group_id,omitempty"`
	OwnerUserID   string `json:"owner_user_id,omitempty"`
	ActorRole     string `json:"actor_role,omitempty"`
}

type mcpGovernanceDecision struct {
	Decision         string   `json:"decision"`
	ApprovalRequired bool     `json:"approval_required"`
	ApprovalMode     string   `json:"approval_mode,omitempty"`
	ApprovalReason   string   `json:"approval_reason,omitempty"`
	SourceSurface    string   `json:"source_surface,omitempty"`
	ConfigScope      string   `json:"config_scope,omitempty"`
	GroupID          string   `json:"group_id,omitempty"`
	Reasons          []string `json:"reasons,omitempty"`
}

func normalizeMCPGovernanceContext(r *http.Request, incoming mcpGovernanceContext) mcpGovernanceContext {
	ctx := mcpGovernanceContext{
		SourceSurface: strings.TrimSpace(incoming.SourceSurface),
		ConfigScope:   strings.TrimSpace(incoming.ConfigScope),
		GroupID:       strings.TrimSpace(incoming.GroupID),
		OwnerUserID:   strings.TrimSpace(incoming.OwnerUserID),
		ActorRole:     normalizeGovernanceRole(incoming.ActorRole),
	}

	if identity := IdentityFromContext(r.Context()); identity != nil {
		if ctx.OwnerUserID == "" {
			ctx.OwnerUserID = strings.TrimSpace(identity.UserID)
		}
		if ctx.ActorRole == "" {
			ctx.ActorRole = normalizeGovernanceRole(identity.Role)
		}
	}

	if ctx.SourceSurface == "" {
		ctx.SourceSurface = "mcp_settings_page"
	}
	if ctx.ConfigScope == "" {
		ctx.ConfigScope = "user_group"
	}
	if ctx.GroupID == "" {
		ctx.GroupID = defaultOwnedMCPGroupID(ctx.OwnerUserID)
	}

	return ctx
}

func defaultOwnedMCPGroupID(userID string) string {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return ""
	}
	return "user:" + trimmed
}

func isOwnedGroupMCPConfig(ctx mcpGovernanceContext) bool {
	if ctx.SourceSurface != "mcp_settings_page" || ctx.ConfigScope != "user_group" {
		return false
	}
	if ctx.ActorRole != "owner" {
		return false
	}
	if ctx.OwnerUserID == "" || ctx.GroupID == "" {
		return false
	}
	return ctx.GroupID == defaultOwnedMCPGroupID(ctx.OwnerUserID)
}

func mcpLibraryLocality(entry *mcp.LibraryEntry) string {
	if entry == nil {
		return "unknown"
	}

	switch mcpLibraryDeploymentBoundary(entry) {
	case "external_saas", "remote_mcp":
		return "remote"
	case "local_loopback", "local_first":
		return "local"
	}

	tags := normalizeMCPEntryTags(entry.Tags)
	if slices.Contains(tags, "remote") {
		return "remote"
	}
	if entry.Transport == "sse" {
		if endpoint := strings.TrimSpace(entry.URL); endpoint != "" {
			if isLoopbackURL(endpoint) {
				return "local"
			}
			return "remote"
		}
	}
	return "local"
}

func mcpLibraryDeploymentBoundary(entry *mcp.LibraryEntry) string {
	if entry == nil {
		return "unknown"
	}

	boundary := strings.ToLower(strings.TrimSpace(entry.DeploymentBoundary))
	if boundary != "" {
		return boundary
	}

	tags := normalizeMCPEntryTags(entry.Tags)
	if slices.Contains(tags, "remote") {
		return "remote_mcp"
	}
	if entry.Transport == "sse" {
		if endpoint := strings.TrimSpace(entry.URL); endpoint != "" {
			if isLoopbackURL(endpoint) {
				return "local_loopback"
			}
			return "remote_mcp"
		}
	}
	return "local_first"
}

func mcpLibraryCredentialBoundary(entry *mcp.LibraryEntry) string {
	if entry == nil {
		return "unknown"
	}
	if entry.HasRequiredSecretEnvVar() {
		return "secret_required"
	}
	if len(entry.DeclaredEnvKeys()) > 0 {
		return "operator_supplied"
	}
	return "none"
}

func mcpLibraryBundleVersionPosture(entry *mcp.LibraryEntry) string {
	if entry == nil {
		return "unknown"
	}
	version := strings.ToLower(strings.TrimSpace(entry.Version))
	if version == "" || version == "latest" {
		return "floating"
	}
	for _, pkg := range entry.Packages {
		pkgVersion := strings.ToLower(strings.TrimSpace(pkg.Version))
		if pkgVersion == "" || pkgVersion == "latest" {
			return "floating"
		}
	}
	return "pinned"
}

func mcpLibraryRiskLevel(entry *mcp.LibraryEntry) string {
	if entry == nil {
		return "medium"
	}

	switch mcpLibraryDeploymentBoundary(entry) {
	case "external_saas", "remote_mcp":
		return "high"
	}

	locality := mcpLibraryLocality(entry)
	if locality == "remote" {
		return "high"
	}

	tags := normalizeMCPEntryTags(entry.Tags)
	if slices.Contains(tags, "browser") || slices.Contains(tags, "web") || slices.Contains(tags, "messaging") {
		return "medium"
	}
	return "low"
}

func normalizeMCPEntryTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(tag))
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func isLoopbackURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}

func buildMCPLibraryInspectionReport(entry *mcp.LibraryEntry, ctx mcpGovernanceContext) map[string]any {
	locality := mcpLibraryLocality(entry)
	riskLevel := mcpLibraryRiskLevel(entry)
	deploymentBoundary := mcpLibraryDeploymentBoundary(entry)
	credentialBoundary := mcpLibraryCredentialBoundary(entry)
	decision := buildMCPLibraryGovernanceDecision(entry, ctx)

	return map[string]any{
		"service_name":           strings.TrimSpace(entry.Name),
		"source":                 "curated_library",
		"risk_level":             riskLevel,
		"required_scopes":        []string{"mcp:write"},
		"network_locality":       locality,
		"deployment_boundary":    deploymentBoundary,
		"credential_boundary":    credentialBoundary,
		"secrets_declared":       sortedMCPLibraryEnvKeys(entry),
		"bundle_install_path":    "curated_library_only",
		"bundle_version_posture": mcpLibraryBundleVersionPosture(entry),
		"decision":               decision.Decision,
		"reasons":                decision.Reasons,
		"governance":             decision,
		"governance_context":     ctx,
	}
}

func buildOwnedMCPConfigDecision(r *http.Request) mcpGovernanceDecision {
	ctx := normalizeMCPGovernanceContext(r, mcpGovernanceContext{})
	return buildMCPConfigGovernanceDecision(ctx, "local", "low")
}

func sortedMCPEnvKeys(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for key := range env {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		keys = append(keys, trimmed)
	}
	slices.Sort(keys)
	return keys
}

func sortedMCPLibraryEnvKeys(entry *mcp.LibraryEntry) []string {
	if entry == nil {
		return nil
	}
	keys := entry.DeclaredEnvKeys()
	slices.Sort(keys)
	return keys
}

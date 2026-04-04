package server

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/pkg/protocol"
)

type userGovernanceProfile struct {
	Role                 string
	CostSensitivity      string
	ReviewStrictness     string
	AutomationTolerance  string
	EscalationPreference string
}

type approvalThresholds struct {
	MaxCost         float64
	MaxRisk         string
	AllowExternal   bool
	RequireEscalate bool
}

func defaultUserGovernanceProfile(identityRole string) userGovernanceProfile {
	role := normalizeGovernanceRole(identityRole)
	if role == "" || role == "admin" {
		role = "owner"
	}
	return userGovernanceProfile{
		Role:                 role,
		CostSensitivity:      "balanced",
		ReviewStrictness:     "standard",
		AutomationTolerance:  "balanced",
		EscalationPreference: "ask",
	}
}

func normalizeGovernanceRole(raw string) string {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" || normalized == "<nil>" {
		return ""
	}
	switch normalized {
	case "owner", "admin":
		return "owner"
	case "operator", "operations":
		return "operator"
	case "review", "reviewer", "qa":
		return "reviewer"
	default:
		return normalized
	}
}

func normalizeCostSensitivity(value any) string {
	switch strings.TrimSpace(strings.ToLower(fmt.Sprint(value))) {
	case "low":
		return "low"
	case "high":
		return "high"
	default:
		return "balanced"
	}
}

func normalizeReviewStrictness(value any) string {
	switch strings.TrimSpace(strings.ToLower(fmt.Sprint(value))) {
	case "light":
		return "light"
	case "strict":
		return "strict"
	default:
		return "standard"
	}
}

func normalizeAutomationTolerance(value any) string {
	switch strings.TrimSpace(strings.ToLower(fmt.Sprint(value))) {
	case "cautious":
		return "cautious"
	case "aggressive":
		return "aggressive"
	default:
		return "balanced"
	}
}

func normalizeEscalationPreference(value any) string {
	switch strings.TrimSpace(strings.ToLower(fmt.Sprint(value))) {
	case "notify":
		return "notify"
	case "halt":
		return "halt"
	default:
		return "ask"
	}
}

func userGovernanceProfileFromSettings(settings map[string]any, identityRole string) userGovernanceProfile {
	profile := defaultUserGovernanceProfile(identityRole)
	if settings == nil {
		return profile
	}

	if role := normalizeGovernanceRole(fmt.Sprint(settings["role"])); role != "" {
		profile.Role = role
	}
	profile.CostSensitivity = normalizeCostSensitivity(settings["cost_sensitivity"])
	profile.ReviewStrictness = normalizeReviewStrictness(settings["review_strictness"])
	profile.AutomationTolerance = normalizeAutomationTolerance(settings["automation_tolerance"])
	profile.EscalationPreference = normalizeEscalationPreference(settings["escalation_preference"])
	return profile
}

func (p userGovernanceProfile) snapshot() *protocol.GovernanceProfileSnapshot {
	return &protocol.GovernanceProfileSnapshot{
		Role:                 p.Role,
		CostSensitivity:      p.CostSensitivity,
		ReviewStrictness:     p.ReviewStrictness,
		AutomationTolerance:  p.AutomationTolerance,
		EscalationPreference: p.EscalationPreference,
	}
}

func userGovernanceProfileFromRequest(r *http.Request) userGovernanceProfile {
	identityRole := ""
	if identity := IdentityFromContext(r.Context()); identity != nil {
		identityRole = identity.Role
	}
	return userGovernanceProfileFromSettings(loadUserSettings(), identityRole)
}

func auditUserLabelFromRequest(r *http.Request) string {
	if identity := IdentityFromContext(r.Context()); identity != nil {
		if username := strings.TrimSpace(identity.Username); username != "" {
			return username
		}
		if userID := strings.TrimSpace(identity.UserID); userID != "" {
			return userID
		}
	}
	return "local-user"
}

func governanceProfileDirective(profile userGovernanceProfile) string {
	return strings.TrimSpace(fmt.Sprintf(
		"[USER GOVERNANCE PROFILE]\nRole: %s\nCost sensitivity: %s\nReview strictness: %s\nAutomation tolerance: %s\nEscalation preference: %s\nUse this profile when planning actions, choosing execution paths, and deciding whether approval is required.\n",
		profile.Role,
		profile.CostSensitivity,
		profile.ReviewStrictness,
		profile.AutomationTolerance,
		profile.EscalationPreference,
	))
}

func applyGovernanceProfileToLatestMessage(messages []chatRequestMessage, profile userGovernanceProfile) []chatRequestMessage {
	idx := latestUserMessageIndex(messages)
	if idx < 0 {
		return messages
	}

	normalized := make([]chatRequestMessage, len(messages))
	copy(normalized, messages)
	latest := strings.TrimSpace(normalized[idx].Content)
	normalized[idx].Content = governanceProfileDirective(profile) + "\nOriginal request:\n" + latest
	return normalized
}

func approvalThresholdsForProfile(profile userGovernanceProfile) approvalThresholds {
	// These are bounded release defaults for the single-user free-node posture.
	// They shape approval behavior without claiming full enterprise policy or
	// delegated-approval coverage.
	thresholds := approvalThresholds{
		MaxCost:         1.0,
		MaxRisk:         "medium",
		AllowExternal:   false,
		RequireEscalate: false,
	}

	switch profile.CostSensitivity {
	case "low":
		thresholds.MaxCost = 5.0
	case "high":
		thresholds.MaxCost = 0.25
	}

	switch profile.ReviewStrictness {
	case "light":
		thresholds.MaxRisk = "high"
	case "strict":
		thresholds.MaxRisk = "low"
	}

	switch profile.AutomationTolerance {
	case "cautious":
		thresholds.AllowExternal = false
		if thresholds.MaxRisk == "medium" {
			thresholds.MaxRisk = "low"
		}
	case "aggressive":
		thresholds.AllowExternal = true
		if thresholds.MaxRisk == "medium" {
			thresholds.MaxRisk = "high"
		}
	}

	if profile.EscalationPreference == "halt" {
		thresholds.RequireEscalate = true
	}

	return thresholds
}

func approvalRank(risk string) int {
	switch strings.TrimSpace(strings.ToLower(risk)) {
	case "high", "high-risk":
		return 3
	case "medium", "medium-risk":
		return 2
	default:
		return 1
	}
}

func maxRisk(left, right string) string {
	if approvalRank(left) >= approvalRank(right) {
		return left
	}
	return right
}

func capabilityForPlannedTool(name string) string {
	switch strings.TrimSpace(name) {
	case "write_file":
		return "file_output"
	case "generate_blueprint":
		return "planning"
	case "load_deployment_context", "remember", "summarize_conversation":
		return "learning"
	case "delegate":
		return "review"
	case "publish_signal", "broadcast":
		return "tool_execution"
	default:
		return ""
	}
}

func capabilityRiskForTool(name string, arguments map[string]any) string {
	switch strings.TrimSpace(name) {
	case "publish_signal", "broadcast":
		return "high"
	case "load_deployment_context":
		if strings.TrimSpace(fmt.Sprint(arguments["knowledge_class"])) == "company_knowledge" {
			return "high"
		}
		return "medium"
	case "write_file", "delegate", "remember", "summarize_conversation":
		return "medium"
	default:
		return "low"
	}
}

func estimateActionCost(name string, arguments map[string]any) float64 {
	switch strings.TrimSpace(name) {
	case "publish_signal", "broadcast":
		return 1.5
	case "write_file":
		if content, ok := arguments["content"].(string); ok && len(content) > 500 {
			return 0.75
		}
		return 0.35
	case "load_deployment_context":
		if content, ok := arguments["content"].(string); ok && len(content) > 2000 {
			return 0.8
		}
		return 0.45
	case "delegate":
		return 0.6
	case "generate_blueprint":
		return 0.2
	default:
		return 0.1
	}
}

func externalDataUseForTool(name string, arguments map[string]any) bool {
	switch strings.TrimSpace(name) {
	case "publish_signal", "broadcast":
		return true
	case "load_deployment_context":
		switch strings.TrimSpace(fmt.Sprint(arguments["source_kind"])) {
		case "web_research":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func buildApprovalPolicy(profile userGovernanceProfile, planned []protocol.PlannedToolCall, fallbackTools []string) *protocol.ApprovalPolicy {
	if len(planned) == 0 && len(fallbackTools) == 0 {
		return nil
	}

	thresholds := approvalThresholdsForProfile(profile)
	capabilityIDs := []string{}
	risk := "low"
	estimatedCost := 0.0
	externalDataUse := false

	if len(planned) == 0 {
		for _, tool := range fallbackTools {
			planned = append(planned, protocol.PlannedToolCall{Name: tool})
		}
	}

	for _, item := range planned {
		if capabilityID := capabilityForPlannedTool(item.Name); capabilityID != "" && !slices.Contains(capabilityIDs, capabilityID) {
			capabilityIDs = append(capabilityIDs, capabilityID)
			if capability, ok := exchange.CapabilityByID(capabilityID); ok {
				risk = maxRisk(risk, strings.TrimSuffix(capability.RiskClass, "-risk"))
				externalDataUse = externalDataUse || capability.Source == "mcp" || capability.Source == "api" || capability.Source == "node"
			}
		}
		risk = maxRisk(risk, capabilityRiskForTool(item.Name, item.Arguments))
		estimatedCost += estimateActionCost(item.Name, item.Arguments)
		externalDataUse = externalDataUse || externalDataUseForTool(item.Name, item.Arguments)
	}

	policy := &protocol.ApprovalPolicy{
		ApprovalRequired:     false,
		ApprovalMode:         "auto_allowed",
		CapabilityRisk:       risk,
		CapabilityIDs:        capabilityIDs,
		ExternalDataUse:      externalDataUse,
		EstimatedCost:        estimatedCost,
		RequiredApproverRole: profile.Role,
		ApprovalSteps:        []string{"operator_review", "future_role_gate"},
		GovernanceProfile:    profile.snapshot(),
	}

	switch {
	case approvalRank(risk) >= 3:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "capability_risk"
	case externalDataUse && !thresholds.AllowExternal:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "external_data_use"
	case estimatedCost > thresholds.MaxCost:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "cost"
	case approvalRank(risk) > approvalRank(thresholds.MaxRisk):
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "capability_risk"
	case thresholds.RequireEscalate && approvalRank(risk) >= 2:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "escalation_preference"
	case approvalRank(risk) == 2:
		policy.ApprovalMode = "optional"
		policy.ApprovalReason = "capability_risk"
	default:
		policy.ApprovalReason = "auto_approve"
	}

	return policy
}

func policyDecisionForApproval(policy *protocol.ApprovalPolicy) string {
	if policy == nil {
		return "allow"
	}
	if policy.ApprovalRequired {
		return "require_approval"
	}
	return "allow"
}

func approvalStatusValue(policy *protocol.ApprovalPolicy) string {
	if policy == nil {
		return "allow"
	}
	if policy.ApprovalRequired {
		return "approval_required"
	}
	return policy.ApprovalMode
}

func approvalReasonValue(policy *protocol.ApprovalPolicy) string {
	if policy == nil {
		return ""
	}
	return policy.ApprovalReason
}

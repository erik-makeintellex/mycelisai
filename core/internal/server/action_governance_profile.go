package server

import (
	"fmt"
	"net/http"
	"strings"

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

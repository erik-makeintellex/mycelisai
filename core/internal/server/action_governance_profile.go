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
	// FailStrict is set when the settings file exists but can't be read or
	// parsed (AUTH-C1b): every tool action then needs approval.
	FailStrict bool
}

// approvalReasonSettingsUnavailable: the approval policy failed strict
// because the saved settings could not be read (AUTH-C1b).
const approvalReasonSettingsUnavailable = "settings_unavailable"

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

// userGovernanceProfileFromSettings builds the approval profile. Role (and so
// RequiredApproverRole) comes only from the verified identity; the settings
// file's "role" is never read here (AUTH-C1 A2). The other keys are
// organization policy written only by root admin + governance:write.
func userGovernanceProfileFromSettings(settings map[string]any, identityRole string) userGovernanceProfile {
	profile := defaultUserGovernanceProfile(identityRole)
	if settings == nil {
		return profile
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

// userGovernanceProfileFromRequest reads the settings once; an unreadable or
// corrupt file yields the strictest values and FailStrict (AUTH-C1b). The role
// still comes only from the identity.
func userGovernanceProfileFromRequest(r *http.Request) userGovernanceProfile {
	identityRole := ""
	if identity := IdentityFromContext(r.Context()); identity != nil {
		identityRole = identity.Role
	}
	settings, err := loadPersistedUserSettingsWithStatus()
	profile := userGovernanceProfileFromSettings(settings, identityRole)
	profile.FailStrict = err != nil
	return profile
}

// applyStrictGovernanceSettings sets each approval-policy key to its strictest
// value: high cost sensitivity, strict review, cautious automation, halt.
func applyStrictGovernanceSettings(settings map[string]any) {
	settings["cost_sensitivity"] = "high"
	settings["review_strictness"] = "strict"
	settings["automation_tolerance"] = "cautious"
	settings["escalation_preference"] = "halt"
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

func auditActorIDFromRequest(r *http.Request) string {
	if identity := IdentityFromContext(r.Context()); identity != nil {
		if userID := strings.TrimSpace(identity.UserID); userID != "" {
			return userID
		}
		if username := strings.TrimSpace(identity.Username); username != "" {
			return username
		}
	}
	return "local-user"
}

func actorIdentitySnapshotFromRequest(r *http.Request) map[string]any {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		return nil
	}
	snapshot := map[string]any{
		"user_id":        strings.TrimSpace(identity.UserID),
		"user_label":     auditUserLabelFromRequest(r),
		"role":           strings.TrimSpace(identity.Role),
		"effective_role": strings.TrimSpace(identity.EffectiveRole),
		"principal_type": strings.TrimSpace(identity.PrincipalType),
		"auth_source":    strings.TrimSpace(identity.AuthSource),
	}
	for key, value := range snapshot {
		if strings.TrimSpace(fmt.Sprint(value)) == "" {
			delete(snapshot, key)
		}
	}
	if len(snapshot) == 0 {
		return nil
	}
	return snapshot
}

func attachActorIdentity(ctx map[string]any, r *http.Request) map[string]any {
	if ctx == nil {
		ctx = map[string]any{}
	}
	if actorIdentity := actorIdentitySnapshotFromRequest(r); len(actorIdentity) > 0 {
		ctx["actor_identity"] = actorIdentity
	}
	return ctx
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
	normalized[idx].Content = governanceProfileDirective(profile) + protocol.ChatOriginalRequestMarker + "\n" + latest
	return normalized
}

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// governanceSettingKeys feed the approval policy of every user (the settings
// file is global) or claim an approval role, so changing them is organization
// policy: root admin + governance:write, audited first (AUTH-C1 A2). Sorted, so
// refused_keys is stable.
var governanceSettingKeys = []string{"automation_tolerance", "cost_sensitivity", "escalation_preference", "review_strictness", "role"}

const maxUserSettingsBodyBytes = 64 << 10

// putUserSettings: identity, bounded decode, then personal keys save for any
// signed-in user while a change to a governance key needs root admin +
// governance:write. Without it the governance keys are refused (403) and the
// personal keys in the same body still save. With it: audit (requested), write,
// audit (result); an audit failure writes nothing.
func (s *AdminServer) putUserSettings(w http.ResponseWriter, r *http.Request) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	var input map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxUserSettingsBodyBytes)).Decode(&input); err != nil {
		respondAPIError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	current := loadPersistedUserSettings()
	next := mergeUserSettings(input)
	changed := changedGovernanceSettings(current, next)
	if len(changed) > 0 && (identity.Role != "admin" || !hasScope(identity, scopeGovernanceWrite)) {
		refuseGovernanceSettings(w, r, input, current, changed)
		return
	}
	if len(changed) > 0 {
		previous, proposed := map[string]any{}, map[string]any{}
		for _, key := range changed {
			previous[key], proposed[key] = current[key], next[key]
		}
		auditCtx := func(status string) map[string]any {
			return map[string]any{"action": "governance_settings_update", "changed_keys": changed,
				"previous": previous, "new": proposed, "result_status": status}
		}
		if s.auditGovernance(r, "user-settings", "Organization approval settings update requested", auditCtx("requested")) == "" {
			respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable,
				"The audit record could not be written, so no setting was changed.", nil)
			return
		}
		if err := saveUserSettings(next); err != nil {
			s.auditGovernance(r, "user-settings", "Organization approval settings update failed", auditCtx("failed"))
			respondAPIError(w, "failed to persist user settings", http.StatusInternalServerError)
			return
		}
		s.auditGovernance(r, "user-settings", "Organization approval settings update applied", auditCtx("applied"))
		respondJSON(w, next)
		return
	}
	if err := saveUserSettings(next); err != nil {
		respondAPIError(w, "failed to persist user settings", http.StatusInternalServerError)
		return
	}
	respondJSON(w, next)
}

// changedGovernanceSettings lists the governance keys whose normalized value
// in next differs from current. Echoing an unchanged value is not a change.
func changedGovernanceSettings(current, next map[string]any) []string {
	var changed []string
	for _, key := range governanceSettingKeys {
		if fmt.Sprint(current[key]) != fmt.Sprint(next[key]) {
			changed = append(changed, key)
		}
	}
	return changed
}

// refuseGovernanceSettings saves only the non-governance keys of input (when
// they change anything) and writes the 403 settings_policy_forbidden blocker.
func refuseGovernanceSettings(w http.ResponseWriter, r *http.Request, input, current map[string]any, changed []string) {
	personal := make(map[string]any, len(input))
	for key, value := range input {
		if !slices.Contains(governanceSettingKeys, key) {
			personal[key] = value
		}
	}
	next := mergeUserSettings(personal)
	saved := false
	if !reflect.DeepEqual(persistedUserSettings(next), current) {
		if err := saveUserSettings(next); err != nil {
			respondAPIError(w, "failed to persist user settings", http.StatusInternalServerError)
			return
		}
		saved = true
	}
	respondBlockerText(w, r, http.StatusForbidden, codeSettingsPolicyForbidden, settingsPolicyForbiddenCopy, "",
		map[string]string{"required_scope": scopeGovernanceWrite, "refused_keys": strings.Join(changed, ","),
			"preferences_saved": strconv.FormatBool(saved)})
}

func defaultPersistedUserSettings() map[string]any {
	return map[string]any{
		"theme":                 "aero-light",
		"matrix_view":           "grid",
		"assistant_name":        defaultAssistantName,
		"role":                  "owner",
		"cost_sensitivity":      "balanced",
		"review_strictness":     "standard",
		"automation_tolerance":  "balanced",
		"escalation_preference": "ask",
	}
}

func normalizeAssistantName(v any) string {
	name, ok := v.(string)
	if !ok {
		return defaultAssistantName
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return defaultAssistantName
	}
	runes := []rune(name)
	if len(runes) > 48 {
		name = string(runes[:48])
	}
	return name
}

func userSettingsPath() string {
	if p := strings.TrimSpace(os.Getenv("MYCELIS_USER_SETTINGS_PATH")); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".mycelis", "user-settings.json")
}

func loadUserSettings() map[string]any {
	return ResolveDeploymentContract().ApplyUserSettings(loadPersistedUserSettings())
}

func mergeUserSettings(input map[string]any) map[string]any {
	settings := loadPersistedUserSettings()
	mergeStringSetting(settings, input, "theme")
	mergeStringSetting(settings, input, "matrix_view")
	if _, hasAssistantName := input["assistant_name"]; hasAssistantName {
		settings["assistant_name"] = normalizeAssistantName(input["assistant_name"])
	}
	if _, hasRole := input["role"]; hasRole {
		if normalized := normalizeGovernanceRole(fmt.Sprint(input["role"])); normalized != "" {
			settings["role"] = normalized
		}
	}
	mergeGovernanceSettings(settings, input)
	return ResolveDeploymentContract().ApplyUserSettings(settings)
}

func loadPersistedUserSettings() map[string]any {
	settings := defaultPersistedUserSettings()
	path := userSettingsPath()
	if path == "" {
		return settings
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return settings
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return settings
	}

	mergeStringSetting(settings, raw, "theme")
	mergeStringSetting(settings, raw, "matrix_view")
	settings["assistant_name"] = normalizeAssistantName(raw["assistant_name"])
	if role, ok := raw["role"]; ok {
		if normalized := normalizeGovernanceRole(fmt.Sprint(role)); normalized != "" {
			settings["role"] = normalized
		}
	}
	settings["cost_sensitivity"] = normalizeCostSensitivity(raw["cost_sensitivity"])
	settings["review_strictness"] = normalizeReviewStrictness(raw["review_strictness"])
	settings["automation_tolerance"] = normalizeAutomationTolerance(raw["automation_tolerance"])
	settings["escalation_preference"] = normalizeEscalationPreference(raw["escalation_preference"])
	return settings
}

func mergeStringSetting(settings, input map[string]any, key string) {
	if value, ok := input[key].(string); ok && strings.TrimSpace(value) != "" {
		settings[key] = strings.TrimSpace(value)
	}
}

func mergeGovernanceSettings(settings, input map[string]any) {
	if _, ok := input["cost_sensitivity"]; ok {
		settings["cost_sensitivity"] = normalizeCostSensitivity(input["cost_sensitivity"])
	}
	if _, ok := input["review_strictness"]; ok {
		settings["review_strictness"] = normalizeReviewStrictness(input["review_strictness"])
	}
	if _, ok := input["automation_tolerance"]; ok {
		settings["automation_tolerance"] = normalizeAutomationTolerance(input["automation_tolerance"])
	}
	if _, ok := input["escalation_preference"]; ok {
		settings["escalation_preference"] = normalizeEscalationPreference(input["escalation_preference"])
	}
}

func saveUserSettings(settings map[string]any) error {
	path := userSettingsPath()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(persistedUserSettings(settings), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func persistedUserSettings(settings map[string]any) map[string]any {
	persisted := make(map[string]any, len(settings))
	for key, value := range settings {
		switch key {
		case "access_management_tier", "product_edition", "identity_mode", "shared_agent_specificity_owner":
			continue
		default:
			persisted[key] = value
		}
	}
	return persisted
}

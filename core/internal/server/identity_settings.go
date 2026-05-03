package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

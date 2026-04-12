package cognitive

import (
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

var providerOverrideFields = []string{
	"API_KEY_ENV",
	"DATA_BOUNDARY",
	"TOKEN_BUDGET_PROFILE",
	"MAX_OUTPUT_TOKENS",
	"USAGE_POLICY",
	"MODEL_ID",
	"ENDPOINT",
	"ENABLED",
	"API_KEY",
	"LOCATION",
	"TYPE",
}

func applyEnvOverrides(config *BrainConfig) {
	if config.Providers == nil {
		config.Providers = make(map[string]ProviderConfig)
	}
	if config.Profiles == nil {
		config.Profiles = make(map[string]string)
	}

	sort.Slice(providerOverrideFields, func(i, j int) bool {
		return len(providerOverrideFields[i]) > len(providerOverrideFields[j])
	})

	for _, env := range os.Environ() {
		key, value, ok := strings.Cut(env, "=")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if rawProviderField, ok := strings.CutPrefix(key, "MYCELIS_PROVIDER_"); ok {
			applyProviderEnvOverride(config, rawProviderField, value)
			continue
		}

		if rawProfileField, ok := strings.CutPrefix(key, "MYCELIS_PROFILE_"); ok {
			applyProfileEnvOverride(config, rawProfileField, value)
			continue
		}

		switch key {
		case "MYCELIS_MEDIA_ENDPOINT":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			config.Media.Endpoint = value
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_ENDPOINT=%s", value)
		case "MYCELIS_MEDIA_MODEL_ID":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			config.Media.ModelID = value
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_MODEL_ID=%s", value)
		case "MYCELIS_MEDIA_PROVIDER_ID":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			config.Media.Provider.ProviderID = value
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_PROVIDER_ID=%s", value)
		case "MYCELIS_MEDIA_TYPE":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			config.Media.Provider.Type = strings.ToLower(value)
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_TYPE=%s", value)
		case "MYCELIS_MEDIA_LOCATION":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			config.Media.Provider.Location = strings.ToLower(value)
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_LOCATION=%s", value)
		case "MYCELIS_MEDIA_DATA_BOUNDARY":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			config.Media.Provider.DataBoundary = strings.ToLower(value)
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_DATA_BOUNDARY=%s", value)
		case "MYCELIS_MEDIA_USAGE_POLICY":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			config.Media.Provider.UsagePolicy = strings.ToLower(value)
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_USAGE_POLICY=%s", value)
		case "MYCELIS_MEDIA_ENABLED":
			if config.Media == nil {
				config.Media = &MediaConfig{}
			}
			enabled, err := strconv.ParseBool(value)
			if err != nil {
				log.Printf("WARN: ignoring invalid bool for MYCELIS_MEDIA_ENABLED: %q", value)
				continue
			}
			config.Media.Provider.Enabled = &enabled
			log.Printf("DEBUG: Applied media env override MYCELIS_MEDIA_ENABLED=%s", value)
		}
	}
}

func applyProviderEnvOverride(config *BrainConfig, rawField string, value string) {
	for _, field := range providerOverrideFields {
		suffix := "_" + field
		if !strings.HasSuffix(rawField, suffix) {
			continue
		}

		rawID := strings.TrimSuffix(rawField, suffix)
		if rawID == "" {
			return
		}

		providerID := resolveProviderKey(rawID, config.Providers)
		provider := config.Providers[providerID]

		switch field {
		case "TYPE":
			provider.Type = strings.ToLower(value)
		case "ENDPOINT":
			provider.Endpoint = value
		case "MODEL_ID":
			provider.ModelID = value
		case "API_KEY":
			provider.AuthKey = value
		case "API_KEY_ENV":
			provider.AuthKeyEnv = value
		case "LOCATION":
			provider.Location = value
		case "DATA_BOUNDARY":
			provider.DataBoundary = value
		case "TOKEN_BUDGET_PROFILE":
			provider.TokenBudgetProfile = strings.ToLower(value)
		case "MAX_OUTPUT_TOKENS":
			maxTokens, err := strconv.Atoi(value)
			if err != nil || maxTokens <= 0 {
				log.Printf("WARN: ignoring invalid int for %s: %q", "MYCELIS_PROVIDER_"+rawField, value)
				return
			}
			provider.MaxOutputTokens = maxTokens
		case "USAGE_POLICY":
			provider.UsagePolicy = value
		case "ENABLED":
			enabled, err := strconv.ParseBool(value)
			if err != nil {
				log.Printf("WARN: ignoring invalid bool for %s: %q", "MYCELIS_PROVIDER_"+rawField, value)
				return
			}
			provider.Enabled = enabled
		}

		config.Providers[providerID] = NormalizeProviderTokenDefaults(provider)
		log.Printf("DEBUG: Applied provider env override MYCELIS_PROVIDER_%s for provider %s", rawField, providerID)
		return
	}
}

func applyProfileEnvOverride(config *BrainConfig, rawField string, value string) {
	if !strings.HasSuffix(rawField, "_PROVIDER") {
		return
	}

	rawProfile := strings.TrimSuffix(rawField, "_PROVIDER")
	if rawProfile == "" {
		return
	}

	profileID := resolveProfileKey(rawProfile, config.Profiles)
	providerID := resolveProviderReference(value, config.Providers)
	config.Profiles[profileID] = providerID
	log.Printf("DEBUG: Applied profile env override MYCELIS_PROFILE_%s_PROVIDER=%s", rawProfile, providerID)
}

func resolveProviderReference(value string, providers map[string]ProviderConfig) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if _, ok := providers[trimmed]; ok {
		return trimmed
	}
	return resolveProviderKey(trimmed, providers)
}

func resolveProviderKey(raw string, providers map[string]ProviderConfig) string {
	rawNorm := envToken(raw)
	for existing := range providers {
		if envToken(existing) == rawNorm {
			return existing
		}
	}
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), "_", "-"))
}

func resolveProfileKey(raw string, profiles map[string]string) string {
	rawNorm := envToken(raw)
	for existing := range profiles {
		if envToken(existing) == rawNorm {
			return existing
		}
	}
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), "_", "-"))
}

func envToken(value string) string {
	normalized := strings.TrimSpace(strings.ToUpper(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")
	return normalized
}

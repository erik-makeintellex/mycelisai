package cognitive

import (
	"net/url"
	"strings"
)

const (
	DefaultMediaProviderID           = "media"
	DefaultMediaProviderType         = "openai_compatible"
	MediaProviderTypeOpenAICompat    = "openai_compatible"
	MediaProviderTypeOpenAIHosted    = "openai_hosted"
	MediaProviderTypeStabilityCompat = "stability_compatible"
	MediaProviderTypeReplicateCompat = "replicate_compatible"
	DefaultMediaLocalLocation        = "local"
	DefaultMediaRemoteLocation       = "remote"
	DefaultMediaLocalBoundary        = "local_only"
	DefaultMediaHostedBoundary       = "leaves_org"
	DefaultMediaLocalUsagePolicy     = "local_first"
	DefaultMediaHostUsagePolicy      = "require_approval"
)

func NormalizeMediaConfig(cfg MediaConfig) MediaConfig {
	cfg.Provider = cfg.Provider.normalized(cfg.Endpoint, cfg.ModelID)
	if cfg.Endpoint == "" {
		cfg.Endpoint = cfg.Provider.Endpoint
	}
	if cfg.ModelID == "" {
		cfg.ModelID = cfg.Provider.ModelID
	}
	return cfg
}

func (cfg MediaConfig) EffectiveProvider() MediaProviderConfig {
	return NormalizeMediaConfig(cfg).Provider
}

func (cfg MediaConfig) IsConfigured() bool {
	normalized := NormalizeMediaConfig(cfg).Provider
	return strings.TrimSpace(normalized.Endpoint) != "" && strings.TrimSpace(normalized.ModelID) != ""
}

func (cfg MediaConfig) IsReadyForGeneration() bool {
	normalized := NormalizeMediaConfig(cfg).Provider
	return normalized.IsEnabled() && strings.TrimSpace(normalized.Endpoint) != "" && strings.TrimSpace(normalized.ModelID) != ""
}

func (cfg MediaProviderConfig) normalized(legacyEndpoint, legacyModelID string) MediaProviderConfig {
	if strings.TrimSpace(cfg.ProviderID) == "" {
		cfg.ProviderID = DefaultMediaProviderID
	}
	if strings.TrimSpace(cfg.Type) == "" {
		cfg.Type = DefaultMediaProviderType
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		cfg.Endpoint = strings.TrimSpace(legacyEndpoint)
	}
	if strings.TrimSpace(cfg.ModelID) == "" {
		cfg.ModelID = strings.TrimSpace(legacyModelID)
	}
	if strings.TrimSpace(cfg.Location) == "" {
		cfg.Location = inferMediaLocation(cfg.Endpoint, cfg.Type)
	}
	if strings.TrimSpace(cfg.DataBoundary) == "" {
		if cfg.Location == DefaultMediaRemoteLocation {
			cfg.DataBoundary = DefaultMediaHostedBoundary
		} else {
			cfg.DataBoundary = DefaultMediaLocalBoundary
		}
	}
	if strings.TrimSpace(cfg.UsagePolicy) == "" {
		if cfg.Location == DefaultMediaRemoteLocation {
			cfg.UsagePolicy = DefaultMediaHostUsagePolicy
		} else {
			cfg.UsagePolicy = DefaultMediaLocalUsagePolicy
		}
	}
	if cfg.Enabled == nil {
		enabled := strings.TrimSpace(cfg.Endpoint) != "" && strings.TrimSpace(cfg.ModelID) != ""
		cfg.Enabled = &enabled
	}
	return cfg
}

func (cfg MediaProviderConfig) IsEnabled() bool {
	return cfg.Enabled != nil && *cfg.Enabled
}

func (cfg MediaProviderConfig) GenerationEndpointURL() string {
	endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	if endpoint == "" {
		return ""
	}
	if strings.HasSuffix(endpoint, "/images/generations") || strings.HasSuffix(endpoint, "/v1/predictions") {
		return endpoint
	}
	return endpoint + "/images/generations"
}

func (cfg MediaProviderConfig) RequiresApprovalByPolicy() bool {
	normalized := cfg.normalized("", "")
	return strings.TrimSpace(normalized.UsagePolicy) == DefaultMediaHostUsagePolicy ||
		strings.TrimSpace(normalized.DataBoundary) == DefaultMediaHostedBoundary
}

func inferMediaLocation(endpoint, providerType string) string {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		switch strings.TrimSpace(providerType) {
		case MediaProviderTypeOpenAIHosted, MediaProviderTypeStabilityCompat, MediaProviderTypeReplicateCompat, "hosted_api":
			return DefaultMediaRemoteLocation
		default:
			return DefaultMediaLocalLocation
		}
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return DefaultMediaLocalLocation
	}

	host := strings.ToLower(parsed.Hostname())
	switch host {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0", "host.docker.internal":
		return DefaultMediaLocalLocation
	default:
		return DefaultMediaRemoteLocation
	}
}

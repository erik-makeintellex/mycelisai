package searchcap

import (
	"encoding/json"
	"net/url"
	"strings"
)

func disabledResponse(resp Response) Response {
	resp.Provider = ProviderDisabled
	resp.Status = "blocked"
	resp.Blocker = disabledBlocker()
	return resp
}

func disabledBlocker() *Blocker {
	return &Blocker{Code: "search_provider_disabled", Message: "Mycelis Search is disabled.", NextAction: "Set MYCELIS_SEARCH_PROVIDER=local_sources for governed local-source search, searxng, or local_api for a self-hosted HTTP search provider."}
}

func normalizeProvider(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case ProviderLocalSources, ProviderSearXNG, ProviderLocalAPI, ProviderBrave:
		return normalized
	case "self_hosted":
		return ProviderLocalAPI
	default:
		return ProviderDisabled
	}
}

func normalizeSourceScope(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "web", "all":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "local_sources"
	}
}

func limitFor(requested, fallback int) int {
	if requested > 0 && requested <= 50 {
		return requested
	}
	if fallback > 0 {
		return fallback
	}
	return 8
}

func stringMapValue(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func floatMapValue(m map[string]any, key string) float64 {
	if m == nil {
		return 0
	}
	switch value := m[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		parsed, _ := value.Float64()
		return parsed
	default:
		return 0
	}
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isAbsoluteHTTPURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

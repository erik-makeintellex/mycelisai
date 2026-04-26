package searchcap

import "strings"

func disabledResponse(resp Response) Response {
	resp.Provider = ProviderDisabled
	resp.Status = "blocked"
	resp.Blocker = disabledBlocker()
	return resp
}

func disabledBlocker() *Blocker {
	return &Blocker{Code: "search_provider_disabled", Message: "Mycelis Search is disabled.", NextAction: "Set MYCELIS_SEARCH_PROVIDER=local_sources for governed local-source search or searxng for self-hosted web search."}
}

func normalizeProvider(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ProviderLocalSources, ProviderSearXNG, ProviderBrave:
		return strings.ToLower(strings.TrimSpace(raw))
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

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

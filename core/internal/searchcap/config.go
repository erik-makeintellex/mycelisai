package searchcap

import (
	"os"
	"strconv"
	"strings"
)

func ConfigFromEnv() Config {
	maxResults := 8
	if raw := strings.TrimSpace(os.Getenv("MYCELIS_SEARCH_MAX_RESULTS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 50 {
			maxResults = parsed
		}
	}
	provider := normalizeProvider(os.Getenv("MYCELIS_SEARCH_PROVIDER"))
	if strings.TrimSpace(os.Getenv("MYCELIS_SEARCH_PROVIDER")) == "" {
		provider = ProviderLocalSources
	}
	return Config{
		Provider:         provider,
		SearXNGEndpoint:  strings.TrimRight(strings.TrimSpace(os.Getenv("MYCELIS_SEARXNG_ENDPOINT")), "/"),
		LocalAPIEndpoint: strings.TrimSpace(os.Getenv("MYCELIS_SEARCH_LOCAL_API_ENDPOINT")),
		MaxResults:       maxResults,
		OnlineAllowed:    parseBoolDefault(os.Getenv("MYCELIS_SEARCH_ONLINE_ALLOWED"), true),
		OnlineAllowedSet: true,
		ApprovalMode:     normalizeApprovalMode(os.Getenv("MYCELIS_SEARCH_APPROVAL_MODE")),
		DisclosureMode:   normalizeDisclosureMode(os.Getenv("MYCELIS_SEARCH_DISCLOSURE_MODE")),
	}
}

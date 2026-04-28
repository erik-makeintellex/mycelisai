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
	return Config{
		Provider:         normalizeProvider(os.Getenv("MYCELIS_SEARCH_PROVIDER")),
		SearXNGEndpoint:  strings.TrimRight(strings.TrimSpace(os.Getenv("MYCELIS_SEARXNG_ENDPOINT")), "/"),
		LocalAPIEndpoint: strings.TrimSpace(os.Getenv("MYCELIS_SEARCH_LOCAL_API_ENDPOINT")),
		MaxResults:       maxResults,
	}
}

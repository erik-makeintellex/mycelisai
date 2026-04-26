package searchcap

import "strings"

func (s *Service) Status() Status {
	if s == nil {
		return Status{
			Provider:              ProviderDisabled,
			SomaToolName:          "web_search",
			DirectSomaInteraction: true,
			MaxResults:            8,
			Blocker:               disabledBlocker(),
			NextActions:           []string{"Set MYCELIS_SEARCH_PROVIDER=local_sources or searxng."},
		}
	}
	status := Status{
		Provider:              s.Provider(),
		SomaToolName:          "web_search",
		DirectSomaInteraction: true,
		MaxResults:            limitFor(0, s.cfg.MaxResults),
	}
	switch s.cfg.Provider {
	case ProviderLocalSources:
		status.Enabled = true
		status.Configured = s.embedder != nil && s.mem != nil
		status.SupportsLocalSources = true
		status.NextActions = []string{"Ask Soma to search shared sources or uploaded organizational context."}
		if !status.Configured {
			status.Blocker = &Blocker{Code: "local_sources_unavailable", Message: "Local-source search needs the cognitive embedding engine and memory service.", NextAction: "Start Core with memory and an embedding-capable AI engine."}
			status.NextActions = []string{status.Blocker.NextAction}
		}
	case ProviderSearXNG:
		status.Enabled = true
		status.Configured = strings.TrimSpace(s.cfg.SearXNGEndpoint) != ""
		status.SupportsPublicWeb = true
		status.SearXNGEndpointConfigured = status.Configured
		status.NextActions = []string{"Ask Soma to search the public web through the self-hosted SearXNG provider."}
		if !status.Configured {
			status.Blocker = &Blocker{Code: "missing_searxng_endpoint", Message: "SearXNG search is selected but MYCELIS_SEARXNG_ENDPOINT is not configured.", NextAction: "Set MYCELIS_SEARXNG_ENDPOINT to the self-hosted SearXNG base URL."}
			status.NextActions = []string{status.Blocker.NextAction}
		}
	case ProviderBrave:
		status.Enabled = true
		status.Configured = false
		status.SupportsPublicWeb = true
		status.RequiresHostedAPIToken = true
		status.Blocker = &Blocker{Code: "brave_mcp_required", Message: "Brave search is exposed through the governed MCP path, not the Mycelis Search API yet.", NextAction: "Install and configure the curated brave-search MCP server with BRAVE_API_KEY, or use local_sources/searxng."}
		status.NextActions = []string{status.Blocker.NextAction}
	default:
		status.Blocker = disabledBlocker()
		status.NextActions = []string{status.Blocker.NextAction}
	}
	return status
}

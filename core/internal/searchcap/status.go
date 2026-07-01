package searchcap

import "strings"

func (s *Service) Status() Status {
	if s == nil {
		return Status{
			Provider:              ProviderDisabled,
			SomaToolName:          "web_search",
			DirectSomaInteraction: true,
			MaxResults:            8,
			OnlineAllowed:         true,
			ApprovalMode:          "notify",
			DisclosureMode:        "notice_and_interpretation",
			Blocker:               disabledBlocker(),
			NextActions:           []string{"Set MYCELIS_SEARCH_PROVIDER=local_sources, searxng, or local_api."},
		}
	}
	status := Status{
		Provider:              s.Provider(),
		SomaToolName:          "web_search",
		DirectSomaInteraction: true,
		MaxResults:            limitFor(0, s.cfg.MaxResults),
		OnlineAllowed:         s.cfg.OnlineAllowed,
		ApprovalMode:          s.cfg.ApprovalMode,
		DisclosureMode:        s.cfg.DisclosureMode,
	}
	switch s.cfg.Provider {
	case ProviderLocalSources:
		status.Enabled = true
		status.Configured = s.mem != nil
		status.SupportsLocalSources = true
		status.NextActions = []string{"Ask Soma to search shared sources or uploaded organizational context."}
		status.Sources = []Source{searchSource("local_sources", "Local Mycelis context", "local_sources", "", "retained Mycelis context", "none", "live", "governed", "bounded_internal", availability(status.Configured))}
		if !status.Configured {
			status.Blocker = &Blocker{Code: "local_sources_unavailable", Message: "Local-source search needs the memory service.", NextAction: "Start Core with memory enabled."}
			status.NextActions = []string{status.Blocker.NextAction}
			status.Sources[0].Recovery = status.Blocker.NextAction
		}
	case ProviderSearXNG:
		status.Enabled = true
		status.Configured = strings.TrimSpace(s.cfg.SearXNGEndpoint) != ""
		status.SupportsPublicWeb = true
		status.SearXNGEndpointConfigured = status.Configured
		status.NextActions = []string{"Ask Soma to search the public web through the self-hosted SearXNG provider."}
		status.Sources = []Source{searchSource("searxng", "Self-hosted public web", "public_web", s.cfg.SearXNGEndpoint, "self-hosted SearXNG endpoint", "none", "live", "public", "bounded_external", availability(status.Configured))}
		if !status.Configured {
			status.Blocker = &Blocker{Code: "missing_searxng_endpoint", Message: "SearXNG search is selected but MYCELIS_SEARXNG_ENDPOINT is not configured.", NextAction: "Set MYCELIS_SEARXNG_ENDPOINT to the self-hosted SearXNG base URL."}
			status.NextActions = []string{status.Blocker.NextAction}
			status.Sources[0].Recovery = status.Blocker.NextAction
		}
	case ProviderLocalAPI:
		status.Enabled = true
		status.Configured = isAbsoluteHTTPURL(s.cfg.LocalAPIEndpoint)
		status.SupportsPublicWeb = true
		status.NextActions = []string{"Ask Soma to search through the configured self-hosted HTTP search provider."}
		status.Sources = []Source{searchSource("local_api", "Operator-owned search API", "client_or_public_api", s.cfg.LocalAPIEndpoint, "configured local API endpoint", "service_managed", "live", "configured", "bounded_external", availability(status.Configured))}
		if !status.Configured {
			status.Blocker = &Blocker{Code: "missing_local_api_endpoint", Message: "Local API search is selected but MYCELIS_SEARCH_LOCAL_API_ENDPOINT is not configured.", NextAction: "Set MYCELIS_SEARCH_LOCAL_API_ENDPOINT to the self-hosted HTTP search endpoint."}
			status.NextActions = []string{status.Blocker.NextAction}
			status.Sources[0].Recovery = status.Blocker.NextAction
		}
	case ProviderBrave:
		status.Enabled = true
		status.Configured = false
		status.SupportsPublicWeb = true
		status.RequiresHostedAPIToken = true
		status.Blocker = &Blocker{Code: "brave_mcp_required", Message: "Brave search is exposed through the governed MCP path, not the Mycelis Search API yet.", NextAction: "Install and configure the curated brave-search MCP server with BRAVE_API_KEY, or use local_sources/searxng."}
		status.NextActions = []string{status.Blocker.NextAction}
		status.Sources = []Source{searchSource("brave-search", "Brave Search MCP", "public_web_mcp", "", "Brave Search API through MCP", "api_token", "live", "public", "bounded_external", "needs_attention")}
		status.Sources[0].SecretRef = "BRAVE_API_KEY"
		status.Sources[0].Recovery = status.Blocker.NextAction
	default:
		status.Blocker = disabledBlocker()
		status.NextActions = []string{status.Blocker.NextAction}
	}
	status.Sources = s.appendRegisteredSources(status.Sources)
	return status
}

func searchSource(id, name, sourceType, endpoint, boundary, auth, mode, sensitivity, trust, status string) Source {
	return Source{
		ID:               id,
		Name:             name,
		Provider:         id,
		SourceType:       sourceType,
		Endpoint:         endpoint,
		BaseURL:          endpoint,
		ScopeKind:        "all",
		Boundary:         boundary,
		AuthScheme:       auth,
		Mode:             mode,
		SensitivityClass: sensitivity,
		TrustClass:       trust,
		Status:           status,
	}
}

func availability(configured bool) string {
	if configured {
		return "available"
	}
	return "needs_attention"
}

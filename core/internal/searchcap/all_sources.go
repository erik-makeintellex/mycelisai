package searchcap

import (
	"context"
	"fmt"
)

func (s *Service) searchAllSources(ctx context.Context, req Request, resp Response) (Response, error) {
	localReq := req
	localReq.SourceScope = "local_sources"
	localResp, localErr := s.searchLocalSources(ctx, localReq, cloneResponse(resp, ProviderLocalSources, "local_sources"))

	webReq := req
	webReq.SourceScope = "web"
	webResp, webErr := s.searchConfiguredWeb(ctx, webReq, cloneResponse(resp, s.cfg.Provider, "web"))

	localCovered := localErr == nil && localResp.Status == "ok"
	webCovered := webErr == nil && webResp.Status == "ok"
	if localCovered || webCovered {
		resp.Provider = s.cfg.Provider
		resp.Status = "ok"
		resp.Results = mergeSourceResults(localResp.Results, webResp.Results, limitFor(req.MaxResults, s.cfg.MaxResults))
		resp.Count = len(resp.Results)
		resp.Metadata["source_scope"] = "all"
		resp.Metadata["local_source_count"] = len(localResp.Results)
		resp.Metadata["web_source_count"] = len(webResp.Results)
		copyMetadataValue(resp.Metadata, localResp.Metadata, "semantic_fallback")
		copyMetadataValue(resp.Metadata, localResp.Metadata, "semantic_fallback_reason")
		copyMetadataValue(resp.Metadata, localResp.Metadata, "mounted_folder_files_scanned")
		if localCovered && webCovered {
			resp.Metadata["source_coverage"] = "local_sources_and_web"
			return resp, nil
		}
		if webCovered {
			resp.Metadata["source_coverage"] = "web_only"
			resp.Metadata["partial_source_scope"] = "web_only"
			resp.Metadata["missing_source_scope"] = "local_sources"
			resp.Metadata["scope_warning"] = partialScopeWarning("Approved local data and mounted sources were not available for this mixed search.", localResp.Blocker, localErr)
			return resp, nil
		}
		resp.Metadata["source_coverage"] = "local_sources_only"
		resp.Metadata["partial_source_scope"] = "local_sources_only"
		resp.Metadata["missing_source_scope"] = "web"
		resp.Metadata["scope_warning"] = partialScopeWarning("Public web search was not available for this mixed search.", webResp.Blocker, webErr)
		return resp, nil
	}
	if webErr != nil {
		return resp, webErr
	}
	if localErr != nil {
		return resp, localErr
	}
	if webResp.Blocker != nil {
		resp.Status = "blocked"
		resp.Blocker = webResp.Blocker
		resp.Metadata["source_scope"] = "all"
		resp.Metadata["blocked_source_scope"] = "web"
		return resp, nil
	}
	if localResp.Blocker != nil {
		resp.Status = "blocked"
		resp.Blocker = localResp.Blocker
		resp.Metadata["source_scope"] = "all"
		resp.Metadata["blocked_source_scope"] = "local_sources"
		return resp, nil
	}
	resp.Status = "blocked"
	resp.Blocker = &Blocker{Code: "search_sources_unavailable", Message: "No configured local or public web search source was available.", NextAction: "Configure public web search or add approved local/mounted sources, then retry."}
	return resp, nil
}

func (s *Service) searchConfiguredWeb(ctx context.Context, req Request, resp Response) (Response, error) {
	if isPublicWebProvider(s.cfg.Provider) && !s.cfg.OnlineAllowed {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "online_search_not_allowed", Message: "Online search is disabled by config.", NextAction: "Set MYCELIS_SEARCH_ONLINE_ALLOWED=true to allow configured web_search without confirmation."}
		return resp, nil
	}
	switch s.cfg.Provider {
	case ProviderBuiltinWeb:
		return s.searchBuiltinWeb(ctx, req, resp)
	case ProviderSearXNG:
		return s.searchSearXNG(ctx, req, resp)
	case ProviderLocalAPI:
		return s.searchLocalAPI(ctx, req, resp)
	case ProviderBrave:
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "brave_mcp_required", Message: "Brave search is exposed through the governed MCP path, not the Mycelis Search API yet.", NextAction: "Install and configure the curated brave-search MCP server with BRAVE_API_KEY, or use local_sources/searxng."}
		return resp, nil
	default:
		return disabledResponse(resp), nil
	}
}

func cloneResponse(resp Response, provider, scope string) Response {
	metadata := map[string]any{}
	for key, value := range resp.Metadata {
		metadata[key] = value
	}
	metadata["source_scope"] = scope
	return Response{
		Query:    resp.Query,
		Provider: provider,
		Status:   resp.Status,
		Results:  []Result{},
		Metadata: metadata,
	}
}

func mergeSourceResults(localResults, webResults []Result, max int) []Result {
	if max <= 0 {
		max = len(localResults) + len(webResults)
	}
	merged := make([]Result, 0, minInt(max, len(localResults)+len(webResults)))
	for i := 0; len(merged) < max && (i < len(localResults) || i < len(webResults)); i++ {
		if i < len(localResults) {
			merged = append(merged, localResults[i])
			if len(merged) >= max {
				break
			}
		}
		if i < len(webResults) {
			merged = append(merged, webResults[i])
		}
	}
	return merged
}

func copyMetadataValue(dst, src map[string]any, key string) {
	if value, ok := src[key]; ok {
		dst[key] = value
	}
}

func partialScopeWarning(prefix string, blocker *Blocker, err error) string {
	if err != nil {
		return fmt.Sprintf("%s %v", prefix, err)
	}
	if blocker != nil && blocker.Message != "" {
		return prefix + " " + blocker.Message
	}
	return prefix
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

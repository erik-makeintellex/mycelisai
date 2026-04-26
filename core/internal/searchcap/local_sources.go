package searchcap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mycelis/core/internal/memory"
)

func (s *Service) searchLocalSources(ctx context.Context, req Request, resp Response) (Response, error) {
	scope := normalizeSourceScope(req.SourceScope)
	if scope == "web" {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "web_provider_not_configured", Message: "Public web search is not configured. Local-source search is available only for governed Mycelis context.", NextAction: "Set MYCELIS_SEARCH_PROVIDER=searxng for self-hosted web search or configure the brave-search MCP server."}
		return resp, nil
	}
	if s.embedder == nil || s.mem == nil {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "local_sources_unavailable", Message: "Local-source search needs the cognitive embedding engine and memory service.", NextAction: "Start Core with memory and an embedding-capable AI engine."}
		return resp, nil
	}
	vec, err := s.embedder.Embed(ctx, req.Query, "")
	if err != nil {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "embedding_unavailable", Message: "Local-source search could not embed the query.", NextAction: "Configure an embedding provider and retry."}
		return resp, nil
	}
	results, err := s.mem.SemanticSearchWithOptions(ctx, vec, memory.SemanticSearchOptions{
		Limit:               limitFor(req.MaxResults, s.cfg.MaxResults),
		TenantID:            "default",
		TeamID:              strings.TrimSpace(req.TeamID),
		AgentID:             strings.TrimSpace(req.AgentID),
		RunID:               strings.TrimSpace(req.RunID),
		Visibility:          strings.ToLower(strings.TrimSpace(req.Visibility)),
		Types:               req.Types,
		AllowGlobal:         true,
		AllowLegacyUnscoped: req.TeamID == "" && req.AgentID == "",
	})
	if err != nil {
		return resp, fmt.Errorf("local-source search failed: %w", err)
	}
	now := time.Now().UTC()
	for _, hit := range results {
		resp.Results = append(resp.Results, resultFromVector(hit, now))
	}
	resp.Count = len(resp.Results)
	return resp, nil
}

func resultFromVector(hit memory.VectorResult, retrievedAt time.Time) Result {
	meta := hit.Metadata
	title := firstString(stringMapValue(meta, "artifact_title"), stringMapValue(meta, "title"), hit.ID)
	return Result{
		Title:            title,
		LocalSourceID:    hit.ID,
		Snippet:          hit.Content,
		SourceKind:       "local_source",
		TrustClass:       firstString(stringMapValue(meta, "trust_class"), "user_provided"),
		SensitivityClass: stringMapValue(meta, "sensitivity_class"),
		RetrievedAt:      retrievedAt,
		Score:            hit.Score,
		ProviderMetadata: meta,
	}
}

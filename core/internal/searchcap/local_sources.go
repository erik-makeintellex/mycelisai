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
		resp.Blocker = &Blocker{Code: "web_provider_not_configured", Message: "Public web search is not configured. Local-source search is available only for governed Mycelis context.", NextAction: "Configure SearXNG or a local_api web provider for public research, or ask Soma to search local/shared sources only."}
		return resp, nil
	}
	if scope == "all" {
		resp.Metadata["partial_source_scope"] = "local_sources_only"
		resp.Metadata["missing_source_scope"] = "web"
		resp.Metadata["scope_warning"] = "Public web search is not configured; returning only governed local-source results."
	}
	mounts := s.availableMountedSources(req)
	if s.mem == nil && len(mounts) == 0 {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "local_sources_unavailable", Message: "Local-source search needs the memory service.", NextAction: "Start Core with memory enabled."}
		return resp, nil
	}
	if s.mem != nil {
		opts := memory.SemanticSearchOptions{
			Limit:               limitFor(req.MaxResults, s.cfg.MaxResults),
			TenantID:            "default",
			TeamID:              strings.TrimSpace(req.TeamID),
			AgentID:             strings.TrimSpace(req.AgentID),
			RunID:               strings.TrimSpace(req.RunID),
			Visibility:          strings.ToLower(strings.TrimSpace(req.Visibility)),
			Types:               req.Types,
			AllowGlobal:         true,
			AllowLegacyUnscoped: req.TeamID == "" && req.AgentID == "",
		}
		var vec []float64
		var err error
		if s.embedder != nil {
			vec, err = s.embedder.Embed(ctx, req.Query, "")
		} else {
			err = fmt.Errorf("embedding engine not configured")
		}
		var results []memory.VectorResult
		if err == nil {
			results, err = s.mem.SemanticSearchWithOptions(ctx, vec, opts)
		} else {
			resp.Metadata["semantic_fallback"] = "text_search"
			resp.Metadata["semantic_fallback_reason"] = "embedding_unavailable"
			results, err = s.mem.TextSearchWithOptions(ctx, req.Query, opts)
		}
		if err != nil {
			return resp, fmt.Errorf("local-source search failed: %w", err)
		}
		now := time.Now().UTC()
		for _, hit := range results {
			resp.Results = append(resp.Results, resultFromVector(hit, now))
		}
	}
	for _, mount := range mounts {
		mountedResp, err := s.searchMountedFolder(ctx, req, resp, mount)
		if err != nil {
			return resp, err
		}
		resp.Results = mountedResp.Results
	}
	resp.Count = len(resp.Results)
	return resp, nil
}

func (s *Service) availableMountedSources(req Request) []Source {
	sources := []Source{}
	for _, source := range s.ListSources() {
		if sourceProvider(source) != ProviderMountedFolder {
			continue
		}
		if sourceSelectionBlocker(source, req) != nil {
			continue
		}
		sources = append(sources, source)
	}
	return sources
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

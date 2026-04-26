package searchcap

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/memory"
)

type Service struct {
	cfg      Config
	embedder Embedder
	mem      *memory.Service
	client   *http.Client
}

func NewService(cfg Config, embedder Embedder, mem *memory.Service) *Service {
	if cfg.Provider == "" {
		cfg.Provider = ProviderDisabled
	}
	if cfg.MaxResults <= 0 {
		cfg.MaxResults = 8
	}
	return &Service{
		cfg:      cfg,
		embedder: embedder,
		mem:      mem,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{Proxy: nil},
		},
	}
}

func (s *Service) Provider() string {
	if s == nil {
		return ProviderDisabled
	}
	return s.cfg.Provider
}

func (s *Service) Search(ctx context.Context, req Request) (Response, error) {
	query := strings.TrimSpace(req.Query)
	resp := Response{
		Query:    query,
		Provider: s.Provider(),
		Status:   "ok",
		Results:  []Result{},
		Metadata: map[string]any{"source_scope": normalizeSourceScope(req.SourceScope)},
	}
	if query == "" {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "missing_query", Message: "Search requires a query.", NextAction: "Provide a non-empty search query."}
		return resp, nil
	}
	if s == nil {
		return disabledResponse(resp), nil
	}

	switch s.cfg.Provider {
	case ProviderLocalSources:
		return s.searchLocalSources(ctx, req, resp)
	case ProviderSearXNG:
		return s.searchSearXNG(ctx, req, resp)
	case ProviderBrave:
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "brave_mcp_required", Message: "Brave search is exposed through the governed MCP path, not the Mycelis Search API yet.", NextAction: "Install and configure the curated brave-search MCP server with BRAVE_API_KEY, or use local_sources/searxng."}
		return resp, nil
	default:
		return disabledResponse(resp), nil
	}
}

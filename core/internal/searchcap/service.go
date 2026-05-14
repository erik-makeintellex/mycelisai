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
	cfg.Provider = normalizeProvider(cfg.Provider)
	if cfg.MaxResults <= 0 {
		cfg.MaxResults = 8
	}
	if !cfg.OnlineAllowedSet {
		cfg.OnlineAllowed = true
	}
	cfg.ApprovalMode = normalizeApprovalMode(cfg.ApprovalMode)
	cfg.DisclosureMode = normalizeDisclosureMode(cfg.DisclosureMode)
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
	onlineAllowed := true
	approvalMode := "notify"
	disclosureMode := "notice_and_interpretation"
	if s != nil {
		onlineAllowed = s.cfg.OnlineAllowed
		approvalMode = s.cfg.ApprovalMode
		disclosureMode = s.cfg.DisclosureMode
	}
	resp := Response{
		Query:    query,
		Provider: s.Provider(),
		Status:   "ok",
		Results:  []Result{},
		Metadata: map[string]any{
			"source_scope":    normalizeSourceScope(req.SourceScope),
			"online_allowed":  onlineAllowed,
			"approval_mode":   approvalMode,
			"disclosure_mode": disclosureMode,
			"interpretation":  "external_results_are_leads",
			"confirmation":    "not_required",
			"operator_notice": "search_path_disclosed",
		},
	}
	if query == "" {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "missing_query", Message: "Search requires a query.", NextAction: "Provide a non-empty search query."}
		return resp, nil
	}
	if s == nil {
		return disabledResponse(resp), nil
	}
	if isPublicWebProvider(s.cfg.Provider) && normalizeSourceScope(req.SourceScope) != "local_sources" && !s.cfg.OnlineAllowed {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "online_search_not_allowed", Message: "Online search is disabled by config.", NextAction: "Set MYCELIS_SEARCH_ONLINE_ALLOWED=true to allow configured web_search without confirmation."}
		return resp, nil
	}

	switch s.cfg.Provider {
	case ProviderLocalSources:
		return s.searchLocalSources(ctx, req, resp)
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

func isPublicWebProvider(provider string) bool {
	return provider == ProviderSearXNG || provider == ProviderLocalAPI || provider == ProviderBrave
}

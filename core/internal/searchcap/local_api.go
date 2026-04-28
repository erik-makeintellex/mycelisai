package searchcap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (s *Service) searchLocalAPI(ctx context.Context, req Request, resp Response) (Response, error) {
	if s.cfg.LocalAPIEndpoint == "" {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "missing_local_api_endpoint", Message: "Local API search is selected but MYCELIS_SEARCH_LOCAL_API_ENDPOINT is not configured.", NextAction: "Set MYCELIS_SEARCH_LOCAL_API_ENDPOINT to the self-hosted HTTP search endpoint."}
		return resp, nil
	}
	endpoint, err := url.Parse(s.cfg.LocalAPIEndpoint)
	if err != nil {
		return resp, fmt.Errorf("invalid local API search endpoint: %w", err)
	}
	if !isAbsoluteHTTPURL(s.cfg.LocalAPIEndpoint) {
		return resp, fmt.Errorf("invalid local API search endpoint: endpoint must be an absolute http(s) URL")
	}
	max := limitFor(req.MaxResults, s.cfg.MaxResults)
	q := endpoint.Query()
	q.Set("q", req.Query)
	q.Set("query", req.Query)
	q.Set("max_results", strconv.Itoa(max))
	if req.TimeRange != "" {
		q.Set("time_range", req.TimeRange)
	}
	if req.SourceScope != "" {
		q.Set("source_scope", req.SourceScope)
	}
	setCSV(q, "allowed_domains", req.AllowedDomains)
	setCSV(q, "blocked_domains", req.BlockedDomains)
	if req.Visibility != "" {
		q.Set("visibility", req.Visibility)
	}
	setCSV(q, "types", req.Types)
	endpoint.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return resp, err
	}
	setHeader(httpReq, "X-Mycelis-Team-ID", req.TeamID)
	setHeader(httpReq, "X-Mycelis-Agent-ID", req.AgentID)
	setHeader(httpReq, "X-Mycelis-Run-ID", req.RunID)
	res, err := s.client.Do(httpReq)
	if err != nil {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "local_api_unreachable", Message: "The configured local API search endpoint could not be reached.", NextAction: "Start the self-hosted search service and verify MYCELIS_SEARCH_LOCAL_API_ENDPOINT."}
		return resp, nil
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "local_api_error", Message: fmt.Sprintf("Local API search returned HTTP %d.", res.StatusCode), NextAction: "Check the self-hosted search service health and query permissions."}
		return resp, nil
	}

	var payload struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return resp, fmt.Errorf("decode local API search response: %w", err)
	}
	now := time.Now().UTC()
	for _, raw := range payload.Results {
		if len(resp.Results) >= max {
			break
		}
		result := Result{
			Title:            firstString(stringMapValue(raw, "title"), stringMapValue(raw, "url")),
			URL:              stringMapValue(raw, "url"),
			Snippet:          firstString(stringMapValue(raw, "snippet"), stringMapValue(raw, "content")),
			SourceKind:       firstString(stringMapValue(raw, "source_kind"), "local_api"),
			TrustClass:       firstString(stringMapValue(raw, "trust_class"), "bounded_external"),
			SensitivityClass: firstString(stringMapValue(raw, "sensitivity_class"), "public"),
			RetrievedAt:      now,
			Score:            floatMapValue(raw, "score"),
			ProviderMetadata: safeLocalAPIMetadata(raw),
		}
		resp.Results = append(resp.Results, result)
	}
	resp.Count = len(resp.Results)
	return resp, nil
}

func setCSV(q url.Values, key string, values []string) {
	if len(values) > 0 {
		q.Set(key, strings.Join(values, ","))
	}
}

func setHeader(req *http.Request, key, value string) {
	if strings.TrimSpace(value) != "" {
		req.Header.Set(key, value)
	}
}

func safeLocalAPIMetadata(raw map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"provider", "engine", "source", "source_id", "rank"} {
		if value, ok := raw[key]; ok {
			out[key] = value
		}
	}
	return out
}

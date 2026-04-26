package searchcap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func (s *Service) searchSearXNG(ctx context.Context, req Request, resp Response) (Response, error) {
	if s.cfg.SearXNGEndpoint == "" {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "missing_searxng_endpoint", Message: "SearXNG search is selected but MYCELIS_SEARXNG_ENDPOINT is not configured.", NextAction: "Set MYCELIS_SEARXNG_ENDPOINT to the self-hosted SearXNG base URL."}
		return resp, nil
	}
	endpoint, err := url.Parse(s.cfg.SearXNGEndpoint + "/search")
	if err != nil {
		return resp, fmt.Errorf("invalid SearXNG endpoint: %w", err)
	}
	q := endpoint.Query()
	q.Set("q", req.Query)
	q.Set("format", "json")
	if req.TimeRange != "" {
		q.Set("time_range", req.TimeRange)
	}
	endpoint.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return resp, err
	}
	res, err := s.client.Do(httpReq)
	if err != nil {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "searxng_unreachable", Message: "The configured SearXNG endpoint could not be reached.", NextAction: "Start the self-hosted SearXNG service and verify MYCELIS_SEARXNG_ENDPOINT."}
		return resp, nil
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusForbidden {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "searxng_json_disabled", Message: "SearXNG returned 403; JSON output is likely disabled in settings.yml.", NextAction: "Enable json in the SearXNG search formats and retry."}
		return resp, nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "searxng_error", Message: fmt.Sprintf("SearXNG returned HTTP %d.", res.StatusCode), NextAction: "Check SearXNG service health and query permissions."}
		return resp, nil
	}
	var payload struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return resp, fmt.Errorf("decode SearXNG response: %w", err)
	}
	max := limitFor(req.MaxResults, s.cfg.MaxResults)
	now := time.Now().UTC()
	for _, raw := range payload.Results {
		if len(resp.Results) >= max {
			break
		}
		result := Result{
			Title:            stringMapValue(raw, "title"),
			URL:              stringMapValue(raw, "url"),
			Snippet:          firstString(stringMapValue(raw, "content"), stringMapValue(raw, "snippet")),
			SourceKind:       "searxng",
			TrustClass:       "bounded_external",
			SensitivityClass: "public",
			RetrievedAt:      now,
			ProviderMetadata: raw,
		}
		if result.Title == "" {
			result.Title = result.URL
		}
		resp.Results = append(resp.Results, result)
	}
	resp.Count = len(resp.Results)
	return resp, nil
}

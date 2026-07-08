package searchcap

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const builtinWebEndpoint = "https://html.duckduckgo.com/html/"

var (
	duckResultPattern  = regexp.MustCompile(`(?is)<a[^>]+class=["'][^"']*result__a[^"']*["'][^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	duckSnippetPattern = regexp.MustCompile(`(?is)<a[^>]+class=["'][^"']*result__snippet[^"']*["'][^>]*>(.*?)</a>`)
	htmlTagPattern     = regexp.MustCompile(`(?is)<[^>]+>`)
)

func (s *Service) searchBuiltinWeb(ctx context.Context, req Request, resp Response) (Response, error) {
	endpoint, err := url.Parse(builtinWebEndpoint)
	if err != nil {
		return resp, fmt.Errorf("invalid built-in web endpoint: %w", err)
	}
	q := endpoint.Query()
	q.Set("q", req.Query)
	endpoint.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return resp, err
	}
	httpReq.Header.Set("User-Agent", "Mycelis/8.3 (+https://mycelis.local)")

	res, err := s.client.Do(httpReq)
	if err != nil {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "builtin_web_unreachable", Message: "Built-in web search could not reach the token-free web search endpoint.", NextAction: "Check network access, or configure SearXNG/local_api for operator-owned web search."}
		return resp, nil
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "builtin_web_error", Message: fmt.Sprintf("Built-in web search returned HTTP %d.", res.StatusCode), NextAction: "Retry later, or configure SearXNG/local_api for operator-owned web search."}
		return resp, nil
	}

	raw := limitedReadString(res.Body, 1_500_000)
	matches := duckResultPattern.FindAllStringSubmatch(raw, -1)
	snippets := duckSnippetPattern.FindAllStringSubmatch(raw, -1)
	max := limitFor(req.MaxResults, s.cfg.MaxResults)
	now := time.Now().UTC()
	for i, match := range matches {
		if len(resp.Results) >= max {
			break
		}
		resultURL := normalizeDuckURL(match[1])
		title := cleanHTMLText(match[2])
		if title == "" && resultURL == "" {
			continue
		}
		snippet := ""
		if i < len(snippets) {
			snippet = cleanHTMLText(snippets[i][1])
		}
		resp.Results = append(resp.Results, Result{
			Title:            firstString(title, resultURL),
			URL:              resultURL,
			Snippet:          snippet,
			SourceKind:       ProviderBuiltinWeb,
			TrustClass:       "bounded_external",
			SensitivityClass: "public",
			RetrievedAt:      now,
			ProviderMetadata: map[string]any{"provider": "duckduckgo_html"},
		})
	}
	resp.Count = len(resp.Results)
	return resp, nil
}

func limitedReadString(body interface{ Read([]byte) (int, error) }, maxBytes int64) string {
	buf := make([]byte, 0, 32_768)
	tmp := make([]byte, 16_384)
	var total int64
	for {
		n, err := body.Read(tmp)
		if n > 0 {
			if total+int64(n) > maxBytes {
				n = int(maxBytes - total)
			}
			buf = append(buf, tmp[:n]...)
			total += int64(n)
			if total >= maxBytes {
				break
			}
		}
		if err != nil {
			break
		}
	}
	return string(buf)
}

func cleanHTMLText(raw string) string {
	text := htmlTagPattern.ReplaceAllString(raw, " ")
	text = html.UnescapeString(text)
	return strings.Join(strings.Fields(text), " ")
}

func normalizeDuckURL(raw string) string {
	clean := html.UnescapeString(strings.TrimSpace(raw))
	if clean == "" {
		return ""
	}
	parsed, err := url.Parse(clean)
	if err == nil {
		if target := parsed.Query().Get("uddg"); target != "" {
			if decoded, err := url.QueryUnescape(target); err == nil {
				return decoded
			}
			return target
		}
	}
	return clean
}

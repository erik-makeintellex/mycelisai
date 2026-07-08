package server

import "strings"

type directSearchRequest struct {
	Query       string
	SourceScope string
}

func hasConcreteSearchRequest(lower string) bool {
	_, ok := concreteSearchQuery(lower)
	return ok
}

func directSearchQuery(text string) (string, bool) {
	request, ok := directSearchRequestFromText(text)
	return request.Query, ok
}

func directSearchRequestFromText(text string) (directSearchRequest, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(strings.Join(strings.Fields(trimmed), " "))
	if trimmed == "" || lower == "" {
		return directSearchRequest{}, false
	}
	sourceScope := inferDirectSearchSourceScope(trimmed)
	if query, ok := quotedSearchQuery(trimmed); ok {
		return directSearchRequest{Query: query, SourceScope: sourceScope}, true
	}
	if query, ok := concreteSearchQuery(trimmed); ok {
		return directSearchRequest{Query: query, SourceScope: sourceScope}, true
	}
	explicit := []string{
		"web_search(", "search the web", "web search", "search on ", "search for ", "search about ",
		"public web", "web research",
		"internet research", "online research", "look up", "lookup", "find current", "find recent",
	}
	freshness := []string{"latest", "today", "recent", "news", "real-time", "up to date"}
	if requestContainsAny(lower, explicit) || requestContainsAny(lower, freshness) {
		return directSearchRequest{Query: trimmed, SourceScope: sourceScope}, true
	}
	return directSearchRequest{}, false
}

func shouldHandleDirectSearch(text string) (string, bool) {
	request, ok := shouldHandleDirectSearchRequest(text)
	return request.Query, ok
}

func shouldHandleDirectSearchRequest(text string) (directSearchRequest, bool) {
	if len(inferMutationToolsFromText(text)) > 0 {
		return directSearchRequest{}, false
	}
	return directSearchRequestFromText(text)
}

func quotedSearchQuery(text string) (string, bool) {
	lower := strings.ToLower(text)
	queryIndex := strings.Index(lower, "query=")
	if queryIndex < 0 {
		return "", false
	}
	rest := strings.TrimSpace(text[queryIndex+len("query="):])
	if rest == "" {
		return "", false
	}
	quote := rest[0]
	if quote != '"' && quote != '\'' {
		return "", false
	}
	end := strings.IndexRune(rest[1:], rune(quote))
	if end < 0 {
		return "", false
	}
	query := strings.TrimSpace(rest[1 : 1+end])
	return query, query != ""
}

func concreteSearchQuery(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	patterns := []string{
		"search the public web for ",
		"search the public web on ",
		"search the web for ",
		"search the web on ",
		"search online for ",
		"search online on ",
		"search local sources for ",
		"search local sources on ",
		"search shared sources for ",
		"search shared sources on ",
		"search retained context for ",
		"search retained context on ",
		"search for ",
		"search on ",
		"search about ",
		"look up ",
		"lookup ",
		"find current ",
		"find recent ",
	}
	for _, pattern := range patterns {
		if idx := concretePatternIndex(lower, pattern); idx >= 0 {
			query := strings.TrimSpace(trimmed[idx+len(pattern):])
			query = strings.Trim(query, " .?!")
			if query != "" && !isGenericSearchTarget(query) {
				return query, true
			}
		}
	}
	return "", false
}

func concretePatternIndex(lower, pattern string) int {
	searchFrom := 0
	for {
		idx := strings.Index(lower[searchFrom:], pattern)
		if idx < 0 {
			return -1
		}
		idx += searchFrom
		if idx == 0 || !isASCIILetterOrDigit(lower[idx-1]) {
			return idx
		}
		searchFrom = idx + 1
	}
}

func isASCIILetterOrDigit(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
}

func isGenericSearchTarget(query string) bool {
	lower := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(query)), " "))
	switch lower {
	case "", "it", "this", "that", "the web", "web", "internet", "online", "shared sources", "local sources":
		return true
	default:
		return false
	}
}

func inferDirectSearchSourceScope(query string) string {
	lower := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(query)), " "))
	if lower == "" {
		return "web"
	}
	localTerms := []string{
		"local source", "local sources", "local data", "local files", "retained", "mycelis context",
		"shared source", "shared sources", "uploaded", "mount", "mounted", "workspace",
		"company docs", "internal docs", "customer docs", "intranet", "approved data",
	}
	webTerms := []string{
		"public web", "internet", "online", "web research", "search the web", "browse the web",
	}
	allTerms := []string{
		"all sources", "both local and web", "both web and local", "local and web",
		"web and local", "shared sources and web", "web and shared sources",
		"internal and public", "public and internal", "compare local", "compare internal",
	}
	hasLocal := requestContainsAny(lower, localTerms)
	hasWeb := requestContainsAny(lower, webTerms)
	if requestContainsAny(lower, allTerms) || (hasLocal && hasWeb) {
		return "all"
	}
	if hasLocal {
		return "local_sources"
	}
	if hasWeb {
		return "web"
	}
	return "web"
}

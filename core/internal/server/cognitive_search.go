package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/searchcap"
	"github.com/mycelis/core/pkg/protocol"
)

func isSearchCapabilityQuestion(text string) bool {
	lower := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
	if lower == "" {
		return false
	}
	searchTerms := []string{"web request", "web requests", "web search", "make requests", "browse", "internet", "brave", "searxng", "shared sources"}
	capabilityTerms := []string{"can you", "are you able", "able to", "do you have", "current", "status", "instantiate", "own api", "tokens", "token"}
	return (hasExactWord(lower, "search") || requestContainsAny(lower, searchTerms)) && requestContainsAny(lower, capabilityTerms)
}

func hasExactWord(text, word string) bool {
	for _, field := range strings.FieldsFunc(text, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_'
	}) {
		if field == word {
			return true
		}
	}
	return false
}

func directSearchQuery(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(strings.Join(strings.Fields(trimmed), " "))
	if trimmed == "" || lower == "" {
		return "", false
	}
	if query, ok := quotedSearchQuery(trimmed); ok {
		return query, true
	}
	explicit := []string{"web_search(", "search the web", "web search", "look up", "lookup", "find current", "find recent"}
	freshness := []string{"latest", "today", "recent", "news", "real-time", "up to date"}
	if requestContainsAny(lower, explicit) || requestContainsAny(lower, freshness) {
		return trimmed, true
	}
	return "", false
}

func shouldHandleDirectSearch(text string) (string, bool) {
	if len(inferMutationToolsFromText(text)) > 0 {
		return "", false
	}
	return directSearchQuery(text)
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

func (s *AdminServer) buildSearchCapabilityAnswer() string {
	status := s.searchCapabilityStatus()
	lines := []string{"Current Mycelis search capability:"}
	provider := strings.TrimSpace(status.Provider)
	if provider == "" {
		provider = "disabled"
	}
	availability := "enabled"
	if !status.Enabled {
		availability = "disabled"
	} else if !status.Configured {
		availability = "selected but not fully configured"
	}
	lines = append(lines, fmt.Sprintf("- Provider: %s (%s).", provider, availability))
	if status.SupportsLocalSources {
		lines = append(lines, "- Local shared-source search is available through Soma's web_search tool.")
	}
	if status.SupportsPublicWeb {
		lines = append(lines, "- Public web search is available when the selected provider is configured.")
	}
	if status.DirectSomaInteraction {
		lines = append(lines, fmt.Sprintf("- Soma direct interaction: ask Soma to use %s for governed search requests.", status.SomaToolName))
	}
	if !status.RequiresHostedAPIToken {
		lines = append(lines, "- Hosted Brave tokens are not required for the Mycelis-owned path; use local_sources or self-hosted SearXNG.")
	} else {
		lines = append(lines, "- Brave still requires the curated brave-search MCP server and BRAVE_API_KEY.")
	}
	if status.Blocker != nil {
		lines = append(lines, fmt.Sprintf("- Current blocker: %s", status.Blocker.Message))
		if strings.TrimSpace(status.Blocker.NextAction) != "" {
			lines = append(lines, fmt.Sprintf("- Next action: %s", status.Blocker.NextAction))
		}
	} else if len(status.NextActions) > 0 {
		lines = append(lines, fmt.Sprintf("- Next action: %s", status.NextActions[0]))
	}
	return strings.Join(lines, "\n")
}

func (s *AdminServer) respondSearchCapabilitySummary(w http.ResponseWriter, r *http.Request) {
	s.respondSearchChatPayload(w, r, "Search capability summary", "Search capability summary", s.buildSearchCapabilityAnswer(), nil, protocol.ExecutionStatusCompleted, "", nil)
}

func (s *AdminServer) respondDirectSearchAnswer(w http.ResponseWriter, r *http.Request, query string) {
	searchSvc := s.Search
	if searchSvc == nil {
		searchSvc = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	}
	sourceScope := "web"
	if searchSvc.Provider() == searchcap.ProviderLocalSources {
		sourceScope = "local_sources"
	}
	resp, err := searchSvc.Search(r.Context(), searchcap.Request{
		Query:       query,
		SourceScope: sourceScope,
		MaxResults:  5,
		Visibility:  "visible_to_soma",
	})
	status := protocol.ExecutionStatusCompleted
	blocker := ""
	var degradation *protocol.ExecutionDegradation
	if err != nil {
		status = protocol.ExecutionStatusBlocked
		blocker = err.Error()
		degradation = searchDegradation("search_execution_error", blocker, "Retry after the selected search provider or runtime dependency is reachable.")
	} else if resp.Status != "ok" {
		status = protocol.ExecutionStatusBlocked
		if resp.Blocker != nil {
			blocker = resp.Blocker.Message
			degradation = searchBlockerDegradation(resp.Blocker)
		} else {
			blocker = resp.Status
			degradation = searchDegradation("search_blocked", blocker, "Retry after search capability configuration is corrected.")
		}
	}
	s.respondSearchChatPayload(w, r, "Direct web search", query, buildDirectSearchAnswer(resp, err), []string{"web_search"}, status, blocker, degradation)
}

func buildDirectSearchAnswer(resp searchcap.Response, err error) string {
	notice := directSearchNotice(resp)
	if err != nil {
		return strings.Join([]string{notice, fmt.Sprintf("Blocked: search failed before results were available: %v", err)}, "\n")
	}
	if resp.Blocker != nil {
		lines := []string{
			notice,
			"Blocked: web_search unavailable.",
			fmt.Sprintf("- Blocker: %s", resp.Blocker.Message),
		}
		if strings.TrimSpace(resp.Blocker.NextAction) != "" {
			lines = append(lines, fmt.Sprintf("- Next action: %s", resp.Blocker.NextAction))
		}
		return strings.Join(lines, "\n")
	}
	if len(resp.Results) == 0 {
		return strings.Join([]string{notice, fmt.Sprintf("No results: %q via %s.", resp.Query, resp.Provider)}, "\n")
	}
	lines := []string{notice, fmt.Sprintf("Results for %q:", resp.Query)}
	for i, result := range resp.Results {
		title := strings.TrimSpace(result.Title)
		if title == "" {
			title = "Untitled result"
		}
		line := fmt.Sprintf("%d. %s", i+1, title)
		if strings.TrimSpace(result.URL) != "" {
			line += fmt.Sprintf(" - %s", result.URL)
		}
		if strings.TrimSpace(result.Snippet) != "" {
			line += fmt.Sprintf("; %s", terseSearchSnippet(result.Snippet))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func terseSearchSnippet(snippet string) string {
	compact := strings.Join(strings.Fields(strings.TrimSpace(snippet)), " ")
	const maxLen = 140
	if len(compact) <= maxLen {
		return compact
	}
	return strings.TrimSpace(compact[:maxLen]) + "..."
}

func directSearchNotice(resp searchcap.Response) string {
	provider := strings.TrimSpace(resp.Provider)
	if provider == "" {
		provider = "configured provider"
	}
	mode := "no confirmation"
	if value, ok := resp.Metadata["approval_mode"].(string); ok && strings.TrimSpace(value) == "require_confirmation" {
		mode = "confirmation required"
	}
	if provider == searchcap.ProviderLocalSources {
		return fmt.Sprintf("Notice: web_search via %s; %s; governed local-source results come from retained Mycelis context, not the public web.", provider, mode)
	}
	return fmt.Sprintf("Notice: web_search via %s; %s; external results are leads, verify before relying.", provider, mode)
}

func (s *AdminServer) respondSearchChatPayload(w http.ResponseWriter, r *http.Request, summary, originalIntent, text string, tools []string, resultStatus protocol.ExecutionStatus, blocker string, degradation *protocol.ExecutionDegradation) {
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer, "admin", summary,
		map[string]any{
			"actor":         "Soma",
			"user":          auditUserLabelFromRequest(r),
			"ask_class":     string(protocol.AskClassDirectAnswer),
			"action":        "answer_delivered",
			"result_status": string(resultStatus),
			"source_kind":   "system",
		},
	)
	chatPayload := protocol.ChatResponsePayload{
		Text:      text,
		ToolsUsed: tools,
		AskClass:  protocol.AskClassDirectAnswer,
		Provenance: &protocol.AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		},
		ExecutionSummary: buildSearchExecutionSummary(originalIntent, text, auditEventID, tools, resultStatus, blocker, degradation),
	}
	payloadBytes, _ := json.Marshal(chatPayload)
	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: "admin",
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: protocol.TemplateChatToAnswer,
		Mode:       protocol.ModeAnswer,
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

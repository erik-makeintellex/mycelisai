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
	searchTerms := []string{"search", "web request", "web requests", "web search", "make requests", "browse", "internet", "brave", "searxng", "shared sources"}
	capabilityTerms := []string{"can you", "are you able", "able to", "do you have", "current", "status", "instantiate", "own api", "tokens", "token"}
	return requestContainsAny(lower, searchTerms) && requestContainsAny(lower, capabilityTerms)
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
	s.respondSearchChatPayload(w, r, "Search capability summary", s.buildSearchCapabilityAnswer(), nil, "completed")
}

func (s *AdminServer) respondDirectSearchAnswer(w http.ResponseWriter, r *http.Request, query string) {
	searchSvc := s.Search
	if searchSvc == nil {
		searchSvc = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	}
	resp, err := searchSvc.Search(r.Context(), searchcap.Request{
		Query:       query,
		SourceScope: "web",
		MaxResults:  5,
		Visibility:  "visible_to_soma",
	})
	status := "completed"
	if err != nil || resp.Status != "ok" {
		status = "blocked"
	}
	s.respondSearchChatPayload(w, r, "Direct web search", buildDirectSearchAnswer(resp, err), []string{"web_search"}, status)
}

func buildDirectSearchAnswer(resp searchcap.Response, err error) string {
	if err != nil {
		return fmt.Sprintf("I tried to use web_search, but search failed before results were available: %v", err)
	}
	if resp.Blocker != nil {
		lines := []string{
			"I tried to use web_search, but search is blocked right now.",
			fmt.Sprintf("- Blocker: %s", resp.Blocker.Message),
		}
		if strings.TrimSpace(resp.Blocker.NextAction) != "" {
			lines = append(lines, fmt.Sprintf("- Next action: %s", resp.Blocker.NextAction))
		}
		return strings.Join(lines, "\n")
	}
	if len(resp.Results) == 0 {
		return fmt.Sprintf("I used web_search for %q, but no results were returned by %s.", resp.Query, resp.Provider)
	}
	lines := []string{fmt.Sprintf("I used web_search through %s for %q. Current results:", resp.Provider, resp.Query)}
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
			line += fmt.Sprintf("\n   %s", strings.TrimSpace(result.Snippet))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (s *AdminServer) respondSearchChatPayload(w http.ResponseWriter, r *http.Request, summary, text string, tools []string, resultStatus string) {
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer, "admin", summary,
		map[string]any{
			"actor":         "Soma",
			"user":          auditUserLabelFromRequest(r),
			"ask_class":     string(protocol.AskClassDirectAnswer),
			"action":        "answer_delivered",
			"result_status": resultStatus,
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

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
	if hasConcreteSearchRequest(lower) {
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

func (s *AdminServer) buildSearchCapabilityAnswer() string {
	status := s.searchCapabilityStatus()
	lines := []string{"Soma Search is available for governed research and workspace lookup."}
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
	lines = append(lines, fmt.Sprintf("- Current search mode: %s.", searchModeLabel(provider, availability)))
	if status.SupportsLocalSources {
		lines = append(lines, "- Approved local data and mounted sources can be searched when they are configured.")
	}
	if status.SupportsPublicWeb {
		lines = append(lines, "- Public web research is available through the workspace search provider.")
	} else {
		lines = append(lines, "- Public web research is not enabled for this workspace yet.")
	}
	if status.DirectSomaInteraction {
		lines = append(lines, "- Ask Soma naturally, for example: \"research this\" or \"search our mounted files for this.\"")
	}
	if !status.RequiresHostedAPIToken {
		lines = append(lines, "- Brave tokens are not required for built-in search, approved local data, mounted sources, SearXNG, or local API search.")
	} else {
		lines = append(lines, "- Brave search requires a configured Brave capability and secret reference.")
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

func searchModeLabel(provider, availability string) string {
	switch provider {
	case searchcap.ProviderBuiltinWeb:
		return "built-in public web search (" + availability + ")"
	case searchcap.ProviderLocalSources:
		return "approved local data and mounted sources (" + availability + ")"
	case searchcap.ProviderSearXNG:
		return "SearXNG public web search (" + availability + ")"
	case searchcap.ProviderLocalAPI:
		return "operator-owned search API (" + availability + ")"
	case searchcap.ProviderBrave:
		return "Brave capability search (" + availability + ")"
	case searchcap.ProviderDisabled:
		return "disabled"
	default:
		return provider + " (" + availability + ")"
	}
}

func (s *AdminServer) respondSearchCapabilitySummary(w http.ResponseWriter, r *http.Request) {
	s.respondSearchChatPayload(w, r, "Search capability summary", "Search capability summary", s.buildSearchCapabilityAnswer(), nil, protocol.ExecutionStatusCompleted, "", nil)
}

func (s *AdminServer) respondDirectSearchAnswer(w http.ResponseWriter, r *http.Request, request directSearchRequest) {
	searchSvc := s.Search
	if searchSvc == nil {
		searchSvc = searchcap.NewService(searchcap.Config{Provider: searchcap.ProviderDisabled}, nil, nil)
	}
	query := strings.TrimSpace(request.Query)
	sourceScope := strings.TrimSpace(request.SourceScope)
	if sourceScope == "" {
		sourceScope = inferDirectSearchSourceScope(query)
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
	} else if directSearchMissingWebScope(resp) {
		status = protocol.ExecutionStatusBlocked
		blocker = "Public web search is not configured; only local-source coverage was available for this mixed search."
		degradation = searchDegradation(
			"partial_web_provider_not_configured",
			blocker,
			"Configure public web search, or rely only on the disclosed local-source result boundary.",
		)
	}
	s.respondSearchChatPayload(w, r, "Direct web search", query, buildDirectSearchAnswer(resp, err), []string{"web_search"}, status, blocker, degradation)
}

func buildDirectSearchAnswer(resp searchcap.Response, err error) string {
	notice := directSearchNotice(resp)
	warning, _ := resp.Metadata["scope_warning"].(string)
	warning = strings.TrimSpace(warning)
	if err != nil {
		return strings.Join([]string{notice, fmt.Sprintf("Blocked: search failed before results were available: %v", err)}, "\n")
	}
	if resp.Blocker != nil {
		blockedLine := "Blocked: Soma Search is unavailable."
		if resp.Blocker.Code == "web_provider_not_configured" {
			blockedLine = "Blocked: public web research is not enabled for this workspace."
		}
		lines := []string{
			notice,
			blockedLine,
			fmt.Sprintf("- Blocker: %s", resp.Blocker.Message),
		}
		if strings.TrimSpace(resp.Blocker.NextAction) != "" {
			lines = append(lines, fmt.Sprintf("- Next action: %s", resp.Blocker.NextAction))
		}
		return strings.Join(lines, "\n")
	}
	if len(resp.Results) == 0 {
		lines := []string{notice}
		if warning != "" {
			lines = append(lines, "Coverage warning: "+warning)
		}
		lines = append(lines, fmt.Sprintf("No matching results found for %q in %s.", resp.Query, searchResultBoundaryLabel(resp)))
		return strings.Join(lines, "\n")
	}
	lines := []string{notice, fmt.Sprintf("Results for %q:", resp.Query)}
	if warning != "" {
		lines = append(lines, "Coverage warning: "+warning)
	}
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
	sourceScope := ""
	if value, ok := resp.Metadata["source_scope"].(string); ok {
		sourceScope = strings.TrimSpace(value)
	}
	if provider == searchcap.ProviderLocalSources {
		if sourceScope == "web" {
			return "Notice: public web research was requested, but Soma Search is currently limited to approved local data and mounted sources."
		}
		if sourceScope == "all" {
			return "Notice: Soma Search checked approved local data and mounted sources; public web coverage is not enabled for this workspace yet."
		}
		return "Notice: Soma Search checked approved local data and mounted sources."
	}
	if sourceScope == "all" {
		coverage, _ := resp.Metadata["source_coverage"].(string)
		if coverage == "local_sources_and_web" {
			return "Notice: Soma Search checked approved local data, mounted sources, and the public web."
		}
		if coverage == "web_only" {
			return "Notice: Soma Search checked the public web; approved local data and mounted-source coverage was unavailable for this mixed search."
		}
		if coverage == "local_sources_only" {
			return "Notice: Soma Search checked approved local data and mounted sources; public web coverage was unavailable for this mixed search."
		}
		return "Notice: Soma Search was asked to check local and web sources; verify the source boundary before relying on the result."
	}
	return "Notice: Soma Search checked the public web. External results are leads; verify before relying."
}

func searchResultBoundaryLabel(resp searchcap.Response) string {
	provider := strings.TrimSpace(resp.Provider)
	sourceScope := ""
	if value, ok := resp.Metadata["source_scope"].(string); ok {
		sourceScope = strings.TrimSpace(value)
	}
	if provider == searchcap.ProviderLocalSources || sourceScope == "local_sources" {
		return "approved local data and mounted sources"
	}
	if sourceScope == "all" {
		return "the configured local/web source boundary"
	}
	return "public web search"
}

func directSearchMissingWebScope(resp searchcap.Response) bool {
	missing, _ := resp.Metadata["missing_source_scope"].(string)
	partial, _ := resp.Metadata["partial_source_scope"].(string)
	return strings.TrimSpace(missing) == "web" && strings.TrimSpace(partial) != ""
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

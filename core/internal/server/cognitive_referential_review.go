package server

import (
	"context"
	"fmt"
	"strings"
)

type somaReferentialReview struct {
	LatestRequest       string
	EffectiveRequest    string
	PriorRequest        string
	MatchedHistory      []string
	RegisteredServices  []string
	InstalledTools      []string
	RelevantLibrary     []string
	InferredAction      string
	TemplateID          string
	ThemeIDs            []string
	Concepts            []string
	ProtectionReason    string
	ConfirmationPrompt  string
	NeedsConfirmation   bool
	Confirmed           bool
	RepeatRequested     bool
	MutationTools       []string
	ConfigurationAdvice []string
}

func (r somaReferentialReview) active() bool {
	return r.InferredAction != "" || r.NeedsConfirmation || r.Confirmed
}

func (r somaReferentialReview) contextBlock() string {
	if !r.active() {
		return ""
	}
	var lines []string
	lines = append(lines, "[REFERENTIAL REVIEW]")
	lines = append(lines, "Use this before answering: review the latest user request, related conversation turns, command/action metadata, and registered MCP/tool posture.")
	if r.InferredAction != "" {
		lines = append(lines, "- Inferred action: "+r.InferredAction+".")
	}
	if r.TemplateID != "" {
		lines = append(lines, "- Interaction template: "+r.TemplateID+".")
	}
	if len(r.ThemeIDs) > 0 {
		lines = append(lines, "- Matched phrase themes: "+strings.Join(r.ThemeIDs, ", ")+".")
	}
	if len(r.Concepts) > 0 {
		lines = append(lines, "- Additional concepts: "+strings.Join(r.Concepts, ", ")+".")
	}
	if r.ProtectionReason != "" {
		lines = append(lines, "- Protection reason: "+r.ProtectionReason+".")
	}
	if r.Confirmed && r.PriorRequest != "" {
		lines = append(lines, "- User confirmation applies to prior request: "+r.PriorRequest)
	}
	if len(r.MatchedHistory) > 0 {
		lines = append(lines, "- Related prior turns: "+strings.Join(r.MatchedHistory, " | "))
	}
	if len(r.RegisteredServices) > 0 {
		lines = append(lines, "- Registered MCP/services: "+strings.Join(r.RegisteredServices, ", ")+".")
	}
	if len(r.InstalledTools) > 0 {
		lines = append(lines, "- Available MCP tools: "+strings.Join(r.InstalledTools, ", ")+".")
	}
	if len(r.RelevantLibrary) > 0 {
		lines = append(lines, "- Relevant installable MCP/library entries: "+strings.Join(r.RelevantLibrary, ", ")+".")
	}
	if len(r.ConfigurationAdvice) > 0 {
		lines = append(lines, "- Configuration guidance: "+strings.Join(r.ConfigurationAdvice, " "))
	}
	lines = append(lines, "Rules: confirm before creating teams, installing/enabling MCP servers, assigning tools, or changing capability bindings. If confirmed, proceed through the governed proposal/execution path. If MCP is missing, name the needed server/env var and guide the user to Settings -> Connected Tools without exposing secrets.")
	return strings.Join(lines, "\n")
}

func (s *AdminServer) buildSomaReferentialReview(ctx context.Context, messages []chatRequestMessage) somaReferentialReview {
	latest := latestUserMessageContent(messages)
	review := somaReferentialReview{LatestRequest: latest, EffectiveRequest: latest}
	if strings.TrimSpace(latest) == "" {
		return review
	}

	action, tools := inferSomaReferentialAction(latest)
	review.InferredAction = action
	review.MutationTools = tools
	match := matchSomaInteractionTemplate(latest)
	if match.Template.ID != "" {
		review.TemplateID = match.Template.ID
	}
	for _, theme := range match.Themes {
		review.ThemeIDs = append(review.ThemeIDs, theme.ID)
	}
	review.Concepts = match.Concepts
	review.ProtectionReason = match.ProtectionReason
	review.ConfirmationPrompt = match.ConfirmationPrompt
	review.MutationTools = uniqueOrderedTools(append(review.MutationTools, match.MutationTools...))
	review.RepeatRequested = hasRepeatCue(latest)
	review.MatchedHistory = matchRelatedChatHistory(messages, latest, 3)
	review.RegisteredServices, review.InstalledTools = s.summarizeRegisteredMCP(ctx)
	review.RelevantLibrary = s.summarizeRelevantMCPLibrary(latest, 5)
	review.ConfigurationAdvice = buildSomaMCPConfigurationAdvice(latest, review.RegisteredServices, review.RelevantLibrary)

	if isConfirmationTurn(latest) {
		if prior, priorAction, priorTools := priorActionRequest(messages); prior != "" {
			review.Confirmed = true
			review.PriorRequest = prior
			review.InferredAction = priorAction
			review.MutationTools = priorTools
			review.EffectiveRequest = "Confirmed action from prior user request:\n" + prior
			priorMatch := matchSomaInteractionTemplate(prior)
			if priorMatch.Template.ID != "" {
				review.TemplateID = priorMatch.Template.ID
				review.ProtectionReason = priorMatch.ProtectionReason
				review.ConfirmationPrompt = priorMatch.ConfirmationPrompt
				review.Concepts = priorMatch.Concepts
				review.MutationTools = uniqueOrderedTools(append(review.MutationTools, priorMatch.MutationTools...))
			}
		}
		return review
	}

	review.NeedsConfirmation = actionRequiresConfirmation(action) || match.Protected
	return review
}

func inferSomaReferentialAction(text string) (string, []string) {
	if match := matchSomaInteractionTemplate(text); match.Template.ActionSummary != "" {
		return match.Template.ActionSummary, match.MutationTools
	} else if interactionMatchHasTheme(match, "mcp_enablement") {
		return "confirm MCP/tool enablement or binding guidance for Soma and team agents", match.MutationTools
	}
	lower := normalizeIntentText(text)
	if lower == "" {
		return "", nil
	}
	teamCue := requestContainsAny(lower, []string{"team", "teams", "specialist", "members", "council", "manifest", "lane", "lanes"})
	createCue := requestContainsAny(lower, []string{"create", "build", "launch", "instantiate", "manifest", "put together", "orchestrate", "assign"})
	mcpCue := requestContainsAny(lower, []string{"mcp", "mcps", "tool", "tools", "web search", "web_search", "github", "fetch", "browser", "host data", "shared-sources", "services registered"})

	switch {
	case teamCue && createCue && mcpCue:
		return "confirm compact team manifestation with specialist roles, retained outputs, and target MCP/tool bindings", []string{"generate_blueprint", "delegate"}
	case teamCue && createCue:
		return "confirm compact team manifestation with council-informed specialist roles and retained outputs", []string{"generate_blueprint", "delegate"}
	case mcpCue:
		return "review current MCP/tool posture and recommend next connected tools", nil
	default:
		return "", nil
	}
}

func interactionMatchHasTheme(match somaInteractionMatch, themeID string) bool {
	for _, theme := range match.Themes {
		if theme.ID == themeID {
			return true
		}
	}
	return false
}

func priorActionRequest(messages []chatRequestMessage) (string, string, []string) {
	for i := len(messages) - 2; i >= 0; i-- {
		if !strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			continue
		}
		candidate := stripRouteAndContext(messages[i].Content)
		action, tools := inferSomaReferentialAction(candidate)
		if actionRequiresConfirmation(action) || matchSomaInteractionTemplate(candidate).Protected {
			return candidate, action, tools
		}
	}
	return "", "", nil
}

func actionRequiresConfirmation(action string) bool {
	return strings.Contains(action, "confirm compact team") ||
		strings.Contains(action, "confirm MCP") ||
		strings.Contains(action, "confirm protected") ||
		strings.Contains(action, "confirm governed") ||
		strings.Contains(action, "confirm reusable")
}

func isConfirmationTurn(text string) bool {
	lower := normalizeIntentText(text)
	if lower == "" {
		return false
	}
	confirmations := []string{"yes", "confirm", "confirmed", "proceed", "do it", "execute", "go ahead", "continue with that", "one time"}
	for _, cue := range confirmations {
		if lower == cue || strings.HasPrefix(lower, cue+" ") {
			return true
		}
	}
	return false
}

func hasRepeatCue(text string) bool {
	return requestContainsAny(normalizeIntentText(text), []string{"repetitively", "every time", "from now on", "always", "ongoing"})
}

func matchRelatedChatHistory(messages []chatRequestMessage, latest string, limit int) []string {
	latestTokens := intentTokenSet(latest)
	var matches []string
	for i := len(messages) - 2; i >= 0 && len(matches) < limit; i-- {
		content := stripRouteAndContext(messages[i].Content)
		if content == "" || strings.EqualFold(content, latest) {
			continue
		}
		if sharedIntentTokenCount(latestTokens, intentTokenSet(content)) < 2 {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(messages[i].Role))
		if role == "" {
			role = "turn"
		}
		matches = append(matches, fmt.Sprintf("%s: %s", role, truncateText(content, 120)))
	}
	return matches
}

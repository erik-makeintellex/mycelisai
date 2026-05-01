package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

func stripRouteAndContext(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var kept []string
	skip := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "[WORKSPACE CONTEXT]", "[REFERENTIAL REVIEW]", governedMutationRoutePrefix, directAnswerRoutePrefix, directAnswerRetryRoutePrefix:
			skip = true
			continue
		case "Original request:":
			skip = false
			continue
		}
		if skip && strings.HasPrefix(trimmed, "-") {
			continue
		}
		if trimmed != "" {
			kept = append(kept, trimmed)
		}
	}
	return strings.TrimSpace(strings.Join(kept, " "))
}

func intentTokenSet(text string) map[string]bool {
	stop := map[string]bool{"the": true, "and": true, "for": true, "with": true, "that": true, "this": true, "what": true, "from": true, "into": true, "have": true, "should": true}
	out := map[string]bool{}
	for _, token := range strings.Fields(normalizeIntentText(text)) {
		token = strings.Trim(token, ".,:;!?()[]{}\"'`")
		if len(token) < 4 || stop[token] {
			continue
		}
		out[token] = true
	}
	return out
}

func sharedIntentTokenCount(a, b map[string]bool) int {
	count := 0
	for token := range a {
		if b[token] {
			count++
		}
	}
	return count
}

func normalizeIntentText(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
}

func truncateText(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	return strings.TrimSpace(text[:limit]) + "..."
}

func truncateStrings(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	out := append([]string{}, items[:limit]...)
	out = append(out, fmt.Sprintf("%d more", len(items)-limit))
	return out
}

func applyConfirmedReferentialAction(messages []chatRequestMessage, review somaReferentialReview) []chatRequestMessage {
	idx := latestUserMessageIndex(messages)
	if idx < 0 || !review.Confirmed || strings.TrimSpace(review.EffectiveRequest) == "" {
		return messages
	}
	normalized := make([]chatRequestMessage, len(messages))
	copy(normalized, messages)
	normalized[idx].Content = review.EffectiveRequest
	return normalized
}

func prependReferentialReviewContext(messages []chatRequestMessage, review somaReferentialReview) []chatRequestMessage {
	context := review.contextBlock()
	if strings.TrimSpace(context) == "" {
		return messages
	}
	return append([]chatRequestMessage{{Role: "user", Content: context}}, messages...)
}

func (s *AdminServer) respondReferentialConfirmation(w http.ResponseWriter, r *http.Request, review somaReferentialReview) {
	text := buildReferentialConfirmationText(review)
	auditEventID, _ := s.createAuditEvent(protocol.TemplateChatToAnswer, "admin", "Soma referential action confirmation", map[string]any{
		"actor":           "Soma",
		"user":            auditUserLabelFromRequest(r),
		"ask_class":       string(protocol.AskClassDirectAnswer),
		"action":          "confirmation_requested",
		"result_status":   "blocked",
		"source_kind":     "system",
		"capability_used": strings.Join(review.MutationTools, ","),
	})
	chatPayload := protocol.ChatResponsePayload{
		Text:      text,
		ToolsUsed: review.MutationTools,
		AskClass:  protocol.AskClassDirectAnswer,
		Provenance: &protocol.AnswerProvenance{
			ResolvedIntent:  "confirmation_required",
			PermissionCheck: "pass",
			PolicyDecision:  "confirm",
			AuditEventID:    auditEventID,
		},
	}
	payloadBytes, _ := json.Marshal(chatPayload)
	envelope := protocol.CTSEnvelope{
		Meta:       protocol.CTSMeta{SourceNode: "admin", Timestamp: time.Now()},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: protocol.TemplateChatToAnswer,
		Mode:       protocol.ModeAnswer,
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

func buildReferentialConfirmationText(review somaReferentialReview) string {
	lines := []string{"I infer that you want me to " + review.InferredAction + "."}
	if review.TemplateID != "" {
		lines = append(lines, "", "Matched interaction template: "+review.TemplateID+".")
	}
	if len(review.ThemeIDs) > 0 {
		lines = append(lines, "Matched phrase themes: "+strings.Join(review.ThemeIDs, ", ")+".")
	}
	if len(review.Concepts) > 0 {
		lines = append(lines, "Additional concepts: "+strings.Join(review.Concepts, ", ")+".")
	}
	if review.ProtectionReason != "" {
		lines = append(lines, "Protection reason: "+review.ProtectionReason+".")
	}
	if len(review.MatchedHistory) > 0 {
		lines = append(lines, "", "Relevant context I matched:")
		for _, item := range review.MatchedHistory {
			lines = append(lines, "- "+item)
		}
	}
	if len(review.RegisteredServices) > 0 {
		lines = append(lines, "", "Current registered MCP/services: "+strings.Join(review.RegisteredServices, ", ")+".")
	}
	if len(review.RelevantLibrary) > 0 {
		lines = append(lines, "Likely target MCP/library entries: "+strings.Join(review.RelevantLibrary, ", ")+".")
	}
	if len(review.ConfigurationAdvice) > 0 {
		lines = append(lines, strings.Join(review.ConfigurationAdvice, " "))
	}
	confirm := strings.TrimSpace(review.ConfirmationPrompt)
	if confirm == "" {
		confirm = "Please confirm whether I should proceed once, or tell me to make this the standing behavior for this workflow."
	}
	lines = append(lines, "", confirm)
	return strings.Join(lines, "\n")
}

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

func applyRetryResolution(messages []chatRequestMessage) ([]chatRequestMessage, string) {
	latest := latestUserMessageContent(messages)
	retriedText, isRetried := resolveRetryLatestUserText(messages)
	if !isRetried {
		return messages, latest
	}
	if idx := latestUserMessageIndex(messages); idx >= 0 {
		messages[idx].Content = retriedText
	}
	return messages, retriedText
}

func normalizeIntentText(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
}

func isRetryTurn(text string) bool {
	lower := strings.TrimRight(normalizeIntentText(text), "?.! ")
	switch lower {
	case "try again", "retry", "try that again", "retry that", "please try again", "run it again", "try once more",
		"show me list again", "show me the list again", "show list again", "show the list again",
		"show me results again", "show me the results again", "show results again",
		"format that as a list", "format as a list", "repeat that", "show that again":
		return true
	}
	if strings.HasPrefix(lower, "show me ") && strings.HasSuffix(lower, " again") {
		return true
	}
	return false
}

func resolveRetryLatestUserText(messages []chatRequestMessage) (string, bool) {
	latest := latestUserMessageContent(messages)
	if !isRetryTurn(latest) {
		return latest, false
	}
	for i := len(messages) - 2; i >= 0; i-- {
		if !strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			continue
		}
		candidate := stripRouteAndContext(messages[i].Content)
		if candidate == "" || isRetryTurn(candidate) || strings.HasPrefix(candidate, "[WORKSPACE CONTEXT]") || strings.HasPrefix(candidate, "[REFERENTIAL REVIEW]") {
			continue
		}
		return candidate, true
	}
	return latest, false
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
	var lines []string
	if review.InferredAction != "" {
		lines = append(lines, "I can help you "+review.InferredAction+".")
	}
	if len(review.ConfigurationAdvice) > 0 {
		lines = append(lines, strings.Join(review.ConfigurationAdvice, " "))
	}
	confirm := strings.TrimSpace(review.ConfirmationPrompt)
	if confirm == "" {
		confirm = "Reply \"yes\" to build this now, or tell me what you'd like to change."
	}
	lines = append(lines, confirm)
	return strings.Join(lines, "\n\n")
}

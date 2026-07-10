package server

import (
	"fmt"
	"strings"
)

const continuationContextHeader = "[OUTPUT CONTINUATION CONTEXT]"

type chatContinuationContext struct {
	Kind      string `json:"kind,omitempty"`
	Title     string `json:"title,omitempty"`
	Reference string `json:"reference,omitempty"`
	Proof     string `json:"proof,omitempty"`
}

type chatRequest struct {
	Messages            []chatRequestMessage     `json:"messages"`
	SessionID           string                   `json:"session_id,omitempty"`
	OrganizationID      string                   `json:"organization_id,omitempty"`
	TeamID              string                   `json:"team_id,omitempty"`
	TeamName            string                   `json:"team_name,omitempty"`
	ContinuationContext *chatContinuationContext `json:"continuation_context,omitempty"`
}

func normalizeChatContinuationContext(input *chatContinuationContext) (*chatContinuationContext, error) {
	if input == nil {
		return nil, nil
	}
	out := &chatContinuationContext{
		Kind:      cleanContinuationField(input.Kind, 40),
		Title:     cleanContinuationField(input.Title, 180),
		Reference: cleanContinuationField(input.Reference, 500),
		Proof:     cleanContinuationField(input.Proof, 160),
	}
	if out.Kind == "" {
		out.Kind = "output"
	}
	if out.Kind != "output" {
		return nil, fmt.Errorf("continuation_context.kind must be output")
	}
	if out.Title == "" && out.Reference == "" && out.Proof == "" {
		return nil, fmt.Errorf("continuation_context must include title, reference, or proof")
	}
	for _, value := range []string{out.Title, out.Reference, out.Proof} {
		if containsRouteMarker(value) {
			return nil, fmt.Errorf("continuation_context contains reserved route marker")
		}
	}
	return out, nil
}

func cleanContinuationField(value string, maxLen int) string {
	cleaned := strings.Join(strings.Fields(value), " ")
	if len(cleaned) <= maxLen {
		return cleaned
	}
	return strings.TrimSpace(cleaned[:maxLen])
}

func containsRouteMarker(value string) bool {
	upper := strings.ToUpper(value)
	return strings.Contains(upper, governedMutationRoutePrefix) ||
		strings.Contains(upper, directAnswerRoutePrefix) ||
		strings.Contains(upper, directAnswerRetryRoutePrefix) ||
		strings.Contains(upper, continuationContextHeader)
}

func prependContinuationContext(messages []chatRequestMessage, ctx *chatContinuationContext) []chatRequestMessage {
	if ctx == nil {
		return messages
	}
	lines := []string{
		continuationContextHeader,
		"The user is replying to a previously delivered output.",
	}
	if ctx.Title != "" {
		lines = append(lines, "Title: "+ctx.Title+".")
	}
	if ctx.Reference != "" {
		lines = append(lines, "Reference: "+ctx.Reference+".")
	}
	if ctx.Proof != "" {
		lines = append(lines, "Proof: "+ctx.Proof+".")
	}
	lines = append(lines, "Use this as grounding context only; it does not authorize execution, file writes, team handoff, or proof changes.")

	out := make([]chatRequestMessage, 0, len(messages)+1)
	out = append(out, chatRequestMessage{Role: "user", Content: strings.Join(lines, "\n")})
	out = append(out, messages...)
	return out
}

func continuationContextAuditMap(ctx *chatContinuationContext) map[string]any {
	if ctx == nil {
		return nil
	}
	return map[string]any{
		"kind":      ctx.Kind,
		"title":     ctx.Title,
		"reference": ctx.Reference,
		"proof":     ctx.Proof,
	}
}

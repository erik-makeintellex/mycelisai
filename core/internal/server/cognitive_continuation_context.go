package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

const continuationContextHeader = "[OUTPUT CONTINUATION CONTEXT]"

type chatContinuationContext struct {
	Kind               string               `json:"kind,omitempty"`
	Title              string               `json:"title,omitempty"`
	Reference          string               `json:"reference,omitempty"`
	Proof              string               `json:"proof,omitempty"`
	Intent             string               `json:"intent,omitempty"`
	TeamID             string               `json:"team_id,omitempty"`
	RunID              string               `json:"run_id,omitempty"`
	WorkItemID         string               `json:"work_item_id,omitempty"`
	OutputID           string               `json:"output_id,omitempty"`
	SourceDigest       string               `json:"source_digest,omitempty"`
	SourceVersion      string               `json:"source_version,omitempty"`
	RevisionTarget     string               `json:"revision_target,omitempty"`
	SourceWorkIntent   *protocol.WorkIntent `json:"-"`
	OwnershipValidated bool                 `json:"-"`
}

type chatActiveWorkContext struct {
	Type       string `json:"type,omitempty"`
	ID         string `json:"id,omitempty"`
	RunID      string `json:"run_id,omitempty"`
	TeamID     string `json:"team_id,omitempty"`
	WorkItemID string `json:"work_item_id,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
	SteeringID string `json:"steering_id,omitempty"`
}

type chatRequest struct {
	Messages            []chatRequestMessage     `json:"messages"`
	SessionID           string                   `json:"session_id,omitempty"`
	OrganizationID      string                   `json:"organization_id,omitempty"`
	TeamID              string                   `json:"team_id,omitempty"`
	TeamName            string                   `json:"team_name,omitempty"`
	ContinuationContext *chatContinuationContext `json:"continuation_context,omitempty"`
	ActiveWorkContext   *chatActiveWorkContext   `json:"active_work_context,omitempty"`
}

func normalizeChatActiveWorkContext(input *chatActiveWorkContext) (*chatActiveWorkContext, error) {
	if input == nil {
		return nil, nil
	}
	out := &chatActiveWorkContext{
		Type:       cleanContinuationField(input.Type, 40),
		ID:         cleanContinuationField(input.ID, 80),
		RunID:      cleanContinuationField(input.RunID, 80),
		TeamID:     cleanContinuationField(input.TeamID, 120),
		WorkItemID: cleanContinuationField(input.WorkItemID, 80),
		ProjectID:  cleanContinuationField(input.ProjectID, 80),
		SteeringID: cleanContinuationField(input.SteeringID, 80),
	}
	if out.Type == "" {
		out.Type = "team_work"
	}
	if out.Type != "team_work" || out.TeamID == "" || out.WorkItemID == "" || out.RunID == "" {
		return nil, fmt.Errorf("active_work_context requires type team_work, team_id, work_item_id, and run_id")
	}
	for label, value := range map[string]string{
		"run_id": out.RunID, "work_item_id": out.WorkItemID, "steering_id": out.SteeringID,
	} {
		if value != "" {
			if err := validateOptionalUUID(label, value); err != nil {
				return nil, err
			}
		}
	}
	out.ID = firstNonEmptyString(out.ID, out.WorkItemID)
	return out, nil
}

func normalizeChatContinuationContext(input *chatContinuationContext) (*chatContinuationContext, error) {
	if input == nil {
		return nil, nil
	}
	out := &chatContinuationContext{
		Kind:          cleanContinuationField(input.Kind, 40),
		Title:         cleanContinuationField(input.Title, 180),
		Reference:     cleanContinuationField(input.Reference, 500),
		Proof:         cleanContinuationField(input.Proof, 160),
		TeamID:        cleanContinuationField(input.TeamID, 120),
		RunID:         cleanContinuationField(input.RunID, 80),
		WorkItemID:    cleanContinuationField(input.WorkItemID, 80),
		OutputID:      cleanContinuationField(input.OutputID, 160),
		SourceDigest:  cleanContinuationField(input.SourceDigest, 256),
		SourceVersion: cleanContinuationField(input.SourceVersion, 80),
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
	for label, value := range map[string]string{"run_id": out.RunID, "work_item_id": out.WorkItemID} {
		if value != "" {
			if err := validateOptionalUUID(label, value); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func applyContinuationIntent(ctx *chatContinuationContext, latestRequest string) *chatContinuationContext {
	if ctx == nil {
		return nil
	}
	out := *ctx
	out.Intent = inferContinuationIntent(latestRequest)
	return &out
}

func chatContinuationIntent(ctx *chatContinuationContext) *protocol.ChatContinuationIntent {
	if ctx == nil {
		return nil
	}
	kind := protocol.ContinuationIntentKind(firstNonEmptyString(ctx.Intent, "follow_up"))
	requiresProposal := kind == protocol.ContinuationIntentUpdate ||
		kind == protocol.ContinuationIntentFork ||
		kind == protocol.ContinuationIntentRoute
	return &protocol.ChatContinuationIntent{
		Kind:        kind,
		ContextKind: ctx.Kind,
		TargetTitle: ctx.Title,
		Reference:   ctx.Reference,
		Proof:       ctx.Proof,
		TeamID:      ctx.TeamID, RunID: ctx.RunID, WorkItemID: ctx.WorkItemID,
		OutputID: ctx.OutputID, SourceDigest: ctx.SourceDigest,
		SourceVersion: ctx.SourceVersion, RevisionTarget: ctx.RevisionTarget,
		RequiresProposal: requiresProposal,
		Reason:           continuationIntentReason(kind),
	}
}

func continuationIntentMutationTools(intent *protocol.ChatContinuationIntent) []string {
	if intent == nil || !intent.RequiresProposal {
		return nil
	}
	switch intent.Kind {
	case protocol.ContinuationIntentUpdate, protocol.ContinuationIntentFork:
		if intent.Reference != "" || intent.Proof != "" {
			return []string{"write_file"}
		}
	case protocol.ContinuationIntentRoute:
		return []string{"delegate"}
	}
	return nil
}

func continuationIntentReason(kind protocol.ContinuationIntentKind) string {
	switch kind {
	case protocol.ContinuationIntentUpdate:
		return "The user is asking to change the referenced output."
	case protocol.ContinuationIntentFork:
		return "The user is asking for an alternate version of the referenced output."
	case protocol.ContinuationIntentRoute:
		return "The user is asking to route the referenced output into follow-up work."
	case protocol.ContinuationIntentInspect:
		return "The user is asking to inspect or verify the referenced output."
	default:
		return "The user is continuing conversation from the referenced output."
	}
}

func inferContinuationIntent(text string) string {
	lower := normalizeIntentText(text)
	switch {
	case requestContainsAny(lower, []string{"inspect", "review", "check", "verify", "validate", "what should i look"}):
		return "inspect"
	case requestContainsAny(lower, []string{"alternate", "alternative", "variant", "version", "fork", "different take"}):
		return "fork"
	case requestContainsAny(lower, []string{"send to", "route", "handoff", "hand off", "ask another team", "marketing team", "support team"}):
		return "route"
	case requestContainsAny(lower, []string{"update", "revise", "change", "improve", "fix", "modify", "add ", "remove "}):
		return "update"
	default:
		return "follow_up"
	}
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
	if ctx.Intent != "" {
		lines = append(lines, "Continuation intent: "+ctx.Intent+".")
	}
	if ctx.OwnershipValidated {
		lines = append(lines,
			"Owning team: "+ctx.TeamID+".",
			"Source run: "+ctx.RunID+".",
			"Source work item: "+ctx.WorkItemID+".",
			"Source output: "+ctx.OutputID+".",
			"Immutable source digest: "+ctx.SourceDigest+".",
			"Distinct revision target: "+ctx.RevisionTarget+".",
		)
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
		"intent":    ctx.Intent,
		"team_id":   ctx.TeamID, "run_id": ctx.RunID, "work_item_id": ctx.WorkItemID,
		"output_id": ctx.OutputID, "source_digest": ctx.SourceDigest,
		"source_version": ctx.SourceVersion, "revision_target": ctx.RevisionTarget,
	}
}

func continuationIntentAuditMap(intent *protocol.ChatContinuationIntent) map[string]any {
	if intent == nil {
		return nil
	}
	return map[string]any{
		"kind":              intent.Kind,
		"context_kind":      intent.ContextKind,
		"target_title":      intent.TargetTitle,
		"reference":         intent.Reference,
		"proof":             intent.Proof,
		"requires_proposal": intent.RequiresProposal,
		"reason":            intent.Reason,
		"team_id":           intent.TeamID, "run_id": intent.RunID, "work_item_id": intent.WorkItemID,
		"output_id": intent.OutputID, "source_digest": intent.SourceDigest,
		"source_version": intent.SourceVersion, "revision_target": intent.RevisionTarget,
	}
}

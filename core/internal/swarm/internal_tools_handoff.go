package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// HandoffSourceChannel names the Core lane that notifies a receiving team.
const HandoffSourceChannel = "core.team_handoff"

const (
	maxHandoffArtifacts = 10
	maxHandoffNoteChars = 1000
)

// HandoffRequest is built from the caller's authoritative invocation
// identity plus the validated tool arguments. Source team, agent and run
// never come from model arguments.
type HandoffRequest struct {
	SourceTeamID   string
	SourceAgentID  string
	RunID          string
	TargetTeamID   string
	TargetRole     string
	ArtifactIDs    []string
	Note           string
	ExpectedAction string
}

// HandoffReceipt reports what Core recorded. Status is "queued" until the
// receiving team accepts; it never claims delivery.
type HandoffReceipt struct {
	HandoffID  string `json:"handoff_id"`
	WorkItemID string `json:"work_item_id"`
	Status     string `json:"status"`
	Replayed   bool   `json:"replayed,omitempty"`
}

// HandoffReadRequest asks for one handed-off input on behalf of the reader.
type HandoffReadRequest struct {
	HandoffID     string
	ArtifactID    string
	ReaderTeamID  string
	ReaderAgentID string
}

// HandoffInput is bounded artifact content returned to the receiving lead.
type HandoffInput struct {
	HandoffID   string `json:"handoff_id"`
	ArtifactID  string `json:"artifact_id"`
	Title       string `json:"title"`
	ContentType string `json:"content_type,omitempty"`
	Content     string `json:"content"`
	Truncated   bool   `json:"truncated,omitempty"`
}

// HandoffRecorder is the Core seam that persists and scopes handoffs. The
// server implements it and wires it with SetHandoffRecorder.
type HandoffRecorder interface {
	RecordHandoff(ctx context.Context, req HandoffRequest) (HandoffReceipt, error)
	ReadHandoffInput(ctx context.Context, req HandoffReadRequest) (HandoffInput, error)
}

// HandoffBlockedError is the normalized blocker a handoff tool returns.
type HandoffBlockedError struct {
	Code   string
	Detail string
}

func (e *HandoffBlockedError) Error() string { return e.Code + ": " + e.Detail }

// HandoffBlock builds a normalized handoff blocker.
func HandoffBlock(code, detail string) error { return &HandoffBlockedError{Code: code, Detail: detail} }

// HandoffBlockCode returns the blocker code carried by err, if any.
func HandoffBlockCode(err error) string {
	var blocked *HandoffBlockedError
	if errors.As(err, &blocked) {
		return blocked.Code
	}
	return ""
}

// SetHandoffRecorder wires Core's handoff recorder after construction.
func (r *InternalToolRegistry) SetHandoffRecorder(recorder HandoffRecorder) {
	r.handoffs = recorder
}

// handoffLeadAuthority: only the team's designated lead may hand off or read
// handed-off inputs (owner decision H1 Q2). Soma proposes handoffs instead
// of making them, so the admin-core team never qualifies.
func handoffLeadAuthority(inv ToolInvocationContext) bool {
	agentID, teamID := strings.TrimSpace(inv.AgentID), strings.TrimSpace(inv.TeamID)
	if agentID == "" || teamID == "" || isSomaTeam(teamID) {
		return false
	}
	return isDesignatedTeamLeadRole(inv.AgentRole)
}

// isDesignatedTeamLeadRole: only a member whose team config role names it the
// team lead ("lead", "team_lead", "team lead"). Specialist roles such as
// coder, creative, sentry or "story lead" never qualify, even though
// isLeadAgent treats some of them as leads for prompt context.
func isDesignatedTeamLeadRole(role string) bool {
	switch strings.Join(strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(strings.ToLower(role))), " ") {
	case "lead", "team lead":
		return true
	default:
		return false
	}
}

func isSomaTeam(teamID string) bool {
	return strings.EqualFold(strings.TrimSpace(teamID), "admin-core")
}

func (r *InternalToolRegistry) handoffCaller(ctx context.Context) (ToolInvocationContext, error) {
	inv, ok := ToolInvocationContextFromContext(ctx)
	if !ok || !handoffLeadAuthority(inv) {
		return inv, HandoffBlock("handoff_not_team_lead", "Only a team lead can hand off work or read handed-off inputs; ask your team lead.")
	}
	if inv.PlanningOnly {
		return inv, HandoffBlock("handoff_needs_approval", "Handoffs do not run while a proposal is being planned.")
	}
	if r.handoffs == nil {
		return inv, HandoffBlock("handoff_unavailable", "Core's handoff recorder is not connected; nothing was handed off.")
	}
	return inv, nil
}

func (r *InternalToolRegistry) handleHandOff(ctx context.Context, args map[string]any) (string, error) {
	inv, err := r.handoffCaller(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(inv.RunID) == "" {
		return "", HandoffBlock("handoff_needs_approval", "This work has no approved run, so the handoff needs approval. Ask Soma to propose it.")
	}
	req, err := handoffRequestFromArgs(inv, args)
	if err != nil {
		return "", err
	}
	receipt, err := r.handoffs.RecordHandoff(ctx, req)
	if err != nil {
		return "", err
	}
	return mustJSON(receipt), nil
}

func handoffRequestFromArgs(inv ToolInvocationContext, args map[string]any) (HandoffRequest, error) {
	invalid := func(detail string) error { return HandoffBlock("handoff_invalid_request", detail) }
	req := HandoffRequest{
		SourceTeamID: strings.TrimSpace(inv.TeamID), SourceAgentID: strings.TrimSpace(inv.AgentID), RunID: strings.TrimSpace(inv.RunID),
		TargetTeamID: strings.TrimSpace(stringValue(args["target_team_id"])), TargetRole: strings.TrimSpace(stringValue(args["target_role"])),
		Note: strings.TrimSpace(stringValue(args["note"])), ExpectedAction: strings.ToLower(strings.TrimSpace(stringValue(args["expected_action"]))),
	}
	if req.TargetTeamID == "" || strings.EqualFold(req.TargetTeamID, req.SourceTeamID) {
		return req, invalid("target_team_id must name another team")
	}
	if req.Note == "" || len([]rune(req.Note)) > maxHandoffNoteChars {
		return req, invalid(fmt.Sprintf("note is required and must be at most %d characters", maxHandoffNoteChars))
	}
	if req.ExpectedAction == "" {
		req.ExpectedAction = "use"
	}
	switch req.ExpectedAction {
	case "review", "use", "continue":
	default:
		return req, invalid("expected_action must be review, use, or continue")
	}
	ids := normalizeStringSlice(stringSlice(args["artifact_ids"]))
	if len(ids) == 0 || len(ids) > maxHandoffArtifacts {
		return req, invalid(fmt.Sprintf("artifact_ids must list 1 to %d stored artifacts", maxHandoffArtifacts))
	}
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return req, invalid("artifact_ids must be stored artifact ids")
		}
	}
	req.ArtifactIDs = ids
	return req, nil
}

func (r *InternalToolRegistry) handleReadHandoffInput(ctx context.Context, args map[string]any) (string, error) {
	inv, err := r.handoffCaller(ctx)
	if err != nil {
		return "", err
	}
	handoffID, artifactID := strings.TrimSpace(stringValue(args["handoff_id"])), strings.TrimSpace(stringValue(args["artifact_id"]))
	if handoffID == "" || artifactID == "" {
		return "", HandoffBlock("handoff_invalid_request", "read_handoff_input requires handoff_id and artifact_id")
	}
	input, err := r.handoffs.ReadHandoffInput(ctx, HandoffReadRequest{
		HandoffID: handoffID, ArtifactID: artifactID, ReaderTeamID: strings.TrimSpace(inv.TeamID), ReaderAgentID: strings.TrimSpace(inv.AgentID),
	})
	if err != nil {
		return "", err
	}
	return mustJSON(input), nil
}

// HandoffInputRef names one handed-off artifact.
type HandoffInputRef struct {
	ArtifactID string `json:"artifact_id"`
	Title      string `json:"title,omitempty"`
}

// HandoffCommand is the outbox payload Core dispatches to the receiving team.
type HandoffCommand struct {
	HandoffID      string            `json:"handoff_id"`
	WorkItemID     string            `json:"work_item_id"`
	TargetTeamID   string            `json:"target_team_id"`
	SourceTeamID   string            `json:"source_team_id"`
	RunID          string            `json:"run_id"`
	Note           string            `json:"note"`
	ExpectedAction string            `json:"expected_action"`
	IdempotencyKey string            `json:"idempotency_key"`
	Inputs         []HandoffInputRef `json:"inputs"`
}

// HandoffCommandEnvelope builds the correlated TeamAsk for the receiving
// team's command topic. Its context carries work_item_id, run_id and the
// idempotency key so the team's acceptance and results project onto the
// handoff work item instead of being dropped as uncorrelated.
func HandoffCommandEnvelope(cmd HandoffCommand) ([]byte, error) {
	refs := make([]map[string]any, 0, len(cmd.Inputs))
	reads := make([]string, 0, len(cmd.Inputs))
	for _, input := range cmd.Inputs {
		refs = append(refs, map[string]any{"artifact_id": input.ArtifactID, "title": input.Title})
		reads = append(reads, fmt.Sprintf("read_handoff_input handoff_id=%s artifact_id=%s (%s)", cmd.HandoffID, input.ArtifactID, input.Title))
	}
	ask := protocol.TeamAsk{
		AskKind: protocol.TeamAskKindCoordination,
		Goal:    fmt.Sprintf("Handoff from team %s (%s): %s", cmd.SourceTeamID, cmd.ExpectedAction, cmd.Note),
		Message: "Read the handed-off inputs before relying on them: " + strings.Join(reads, "; "),
		EvidenceRequired: []string{
			"read_handoff_input for each input the team relies on",
			"retained output that cites the handed-off inputs",
		},
		Context: map[string]any{
			"work_item_id": cmd.WorkItemID, "team_id": cmd.TargetTeamID, "run_id": cmd.RunID,
			"idempotency_key": cmd.IdempotencyKey, "handoff_id": cmd.HandoffID, "source_team_id": cmd.SourceTeamID,
			"expected_action": cmd.ExpectedAction, "input_refs": refs, "read_tool": "read_handoff_input",
		},
	}
	raw, err := json.Marshal(ask.Normalize())
	if err != nil {
		return nil, err
	}
	return protocol.WrapSignalPayloadWithMeta(protocol.SourceKindSystem, HandoffSourceChannel, protocol.PayloadKindCommand, cmd.RunID, cmd.TargetTeamID, "", raw)
}

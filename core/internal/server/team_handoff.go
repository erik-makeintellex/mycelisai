package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	teamHandoffChannel       = "organization.team.handoffs"
	teamHandoffReadLimit     = 32 << 10
	teamHandoffVerbQueued    = "handoff_queued"
	teamHandoffVerbSent      = "handoff_sent"
	teamHandoffVerbFailed    = "handoff_dispatch_failed"
	teamHandoffVerbInputRead = "handoff_input_read"
)

// teamHandoffNamespace makes handoff and work item ids deterministic from the
// idempotency key, so a repeated hand_off returns the same handoff.
var teamHandoffNamespace = uuid.MustParse("6f1c7b1e-4a53-4f0e-9d0a-2b8f5c3e7a41")

// teamHandoffRecorder implements swarm.HandoffRecorder: it checks provenance,
// the approved-run boundary and sensitivity, then records the handoff, the
// receiving team's work item and the dispatch intent in one transaction.
type teamHandoffRecorder struct {
	server    *AdminServer
	store     teamHandoffStore
	teamKnown func(context.Context, string) (bool, error)
}

func newTeamHandoffRecorder(s *AdminServer, store teamHandoffStore, teamKnown func(context.Context, string) (bool, error)) *teamHandoffRecorder {
	return &teamHandoffRecorder{server: s, store: store, teamKnown: teamKnown}
}

// TeamHandoffRecorder returns the Core handoff recorder wired into the
// internal tool registry at startup.
func (s *AdminServer) TeamHandoffRecorder() swarm.HandoffRecorder {
	return newTeamHandoffRecorder(s, &sqlTeamHandoffStore{server: s}, s.teamKnownForWork)
}

func teamHandoffKey(req swarm.HandoffRequest) string {
	ids := append([]string(nil), req.ArtifactIDs...)
	sort.Strings(ids)
	sum := sha256.Sum256([]byte(strings.Join([]string{req.RunID, req.SourceTeamID, req.TargetTeamID, strings.Join(ids, ",")}, "|")))
	return hex.EncodeToString(sum[:])
}

func (h *teamHandoffRecorder) RecordHandoff(ctx context.Context, req swarm.HandoffRequest) (swarm.HandoffReceipt, error) {
	key := teamHandoffKey(req)
	handoffID := uuid.NewSHA1(teamHandoffNamespace, []byte(key)).String()
	if existing, err := h.store.GetHandoff(ctx, handoffID); err != nil {
		return swarm.HandoffReceipt{}, fmt.Errorf("look up handoff: %w", err)
	} else if existing != nil {
		return h.replayedReceipt(ctx, *existing), nil
	}
	known, err := h.teamKnown(ctx, req.TargetTeamID)
	if err != nil {
		return swarm.HandoffReceipt{}, fmt.Errorf("check target team: %w", err)
	}
	if !known {
		return swarm.HandoffReceipt{}, swarm.HandoffBlock("handoff_target_unknown", "Team "+req.TargetTeamID+" does not exist.")
	}
	inputs, err := h.attributedInputs(ctx, req)
	if err != nil {
		return swarm.HandoffReceipt{}, err
	}
	target, err := h.store.RunWorkItemForTeam(ctx, req.RunID, req.TargetTeamID)
	if err != nil {
		return swarm.HandoffReceipt{}, fmt.Errorf("check approved run: %w", err)
	}
	if target == nil || strings.TrimSpace(target.IntentProofID) == "" {
		return swarm.HandoffReceipt{}, swarm.HandoffBlock("handoff_needs_approval", "Team "+req.TargetTeamID+" is not part of this approved run. Ask Soma to propose the handoff.")
	}
	record := teamHandoffRecord{
		HandoffID: handoffID, IdempotencyKey: key, RunID: req.RunID, SourceTeamID: req.SourceTeamID, SourceAgentID: req.SourceAgentID,
		TargetTeamID: req.TargetTeamID, TargetRole: req.TargetRole, Note: req.Note, ExpectedAction: req.ExpectedAction,
		WorkItemID: uuid.NewSHA1(teamHandoffNamespace, []byte(key+":work")).String(), Inputs: inputs,
	}
	item, event, interaction := teamHandoffWorkRows(record, target.IntentProofID, target.ContractID)
	commit := teamHandoffCommit{Record: record, WorkItem: item, Event: event, Interaction: interaction, Outbox: teamHandoffOutboxItem(record, item)}
	if err := h.store.CommitHandoff(ctx, commit); err != nil {
		if existing, lookupErr := h.store.GetHandoff(ctx, handoffID); lookupErr == nil && existing != nil {
			return h.replayedReceipt(ctx, *existing), nil
		}
		return swarm.HandoffReceipt{}, fmt.Errorf("record handoff: %w", err)
	}
	return swarm.HandoffReceipt{HandoffID: handoffID, WorkItemID: record.WorkItemID, Status: "queued"}, nil
}

// attributedInputs admits only artifacts Core attributed to the caller's team
// and run; restricted content always needs approval (owner decision H1 Q4).
func (h *teamHandoffRecorder) attributedInputs(ctx context.Context, req swarm.HandoffRequest) ([]swarm.HandoffInputRef, error) {
	artifacts, err := h.store.LoadArtifacts(ctx, req.ArtifactIDs)
	if err != nil {
		return nil, fmt.Errorf("load handoff artifacts: %w", err)
	}
	inputs := make([]swarm.HandoffInputRef, 0, len(req.ArtifactIDs))
	for _, id := range req.ArtifactIDs {
		artifact, ok := artifacts[id]
		meta := artifact.Metadata
		if !ok || stringField(meta, "provenance") != "attributed" || stringField(meta, "provenance_team_id") != req.SourceTeamID || stringField(meta, "provenance_run_id") != req.RunID {
			return nil, swarm.HandoffBlock("handoff_artifact_not_attributed", "Artifact "+id+" was not stored by your team in this run.")
		}
		// Core's stored classification only; the model's sensitivity_class
		// label is never trusted, and a missing class counts as restricted.
		if !swarm.ArtifactSensitivityAllowsHandoff(stringField(meta, "provenance_sensitivity")) {
			return nil, swarm.HandoffBlock("handoff_needs_approval", "Artifact "+id+" is restricted or unclassified. Ask Soma to propose the handoff for approval.")
		}
		inputs = append(inputs, swarm.HandoffInputRef{ArtifactID: id, Title: artifact.Title})
	}
	return inputs, nil
}

func (h *teamHandoffRecorder) replayedReceipt(ctx context.Context, record teamHandoffRecord) swarm.HandoffReceipt {
	status := "queued"
	if item, err := h.store.WorkItem(ctx, record.TargetTeamID, record.WorkItemID); err == nil {
		interactions, _ := h.store.Interactions(ctx, record.TargetTeamID, record.WorkItemID)
		status = teamHandoffStatus(item, interactions)
	}
	return swarm.HandoffReceipt{HandoffID: record.HandoffID, WorkItemID: record.WorkItemID, Status: status, Replayed: true}
}

func (h *teamHandoffRecorder) ReadHandoffInput(ctx context.Context, req swarm.HandoffReadRequest) (swarm.HandoffInput, error) {
	notInScope := swarm.HandoffBlock("handoff_input_not_in_scope", "That input was not handed to your team.")
	if _, err := uuid.Parse(req.HandoffID); err != nil {
		return swarm.HandoffInput{}, notInScope
	}
	record, err := h.store.GetHandoff(ctx, req.HandoffID)
	if err != nil {
		return swarm.HandoffInput{}, fmt.Errorf("look up handoff: %w", err)
	}
	if record == nil || !strings.EqualFold(record.TargetTeamID, req.ReaderTeamID) || !record.hasInput(req.ArtifactID) {
		return swarm.HandoffInput{}, notInScope
	}
	artifacts, err := h.store.LoadArtifacts(ctx, []string{req.ArtifactID})
	if err != nil {
		return swarm.HandoffInput{}, fmt.Errorf("load handoff input: %w", err)
	}
	artifact, ok := artifacts[req.ArtifactID]
	if !ok {
		return swarm.HandoffInput{}, swarm.HandoffBlock("handoff_input_missing", "The handed-off artifact is no longer stored.")
	}
	content, truncated := boundedHandoffContent(artifact.Content)
	interaction := protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(), TeamID: record.TargetTeamID, WorkItemID: record.WorkItemID, RunID: record.RunID,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: swarm.HandoffSourceChannel, ActorRef: req.ReaderAgentID,
		Verb: teamHandoffVerbInputRead, Summary: "Read handoff input: " + artifact.Title, PayloadKind: string(protocol.PayloadKindStatus),
		PayloadRef: record.HandoffID, Payload: map[string]any{"handoff_id": record.HandoffID, "artifact_id": req.ArtifactID, "truncated": truncated}, Version: "v1",
	})
	if err := h.store.RecordInteraction(ctx, interaction); err != nil {
		log.Printf("team handoff: read of %s not recorded: %v", record.HandoffID, err)
		return swarm.HandoffInput{}, fmt.Errorf("record handoff read: %w", err)
	}
	return swarm.HandoffInput{
		HandoffID: record.HandoffID, ArtifactID: req.ArtifactID, Title: artifact.Title, ContentType: artifact.ContentType,
		Content: content, Truncated: truncated,
	}, nil
}

func (r teamHandoffRecord) hasInput(artifactID string) bool {
	for _, input := range r.Inputs {
		if input.ArtifactID == artifactID {
			return true
		}
	}
	return false
}

func boundedHandoffContent(content string) (string, bool) {
	if len(content) <= teamHandoffReadLimit {
		return content, false
	}
	cut := teamHandoffReadLimit
	for cut > 0 && !utf8.RuneStart(content[cut]) {
		cut--
	}
	return content[:cut], true
}

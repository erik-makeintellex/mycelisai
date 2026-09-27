package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	teamHandoffDispatchKind = "team_handoff"
	teamHandoffMaxAttempts  = 3
)

// teamHandoffWorkRows builds the receiving team's queued work item, its
// queued status event and the handoff_queued interaction (handoff_sent is
// recorded only after the dispatcher publishes the notice). The work item joins
// the same run and approved intent as the target team's existing work.
func teamHandoffWorkRows(record teamHandoffRecord, intentProofID, contractID string) (protocol.TeamWorkItem, protocol.TeamStatusEvent, protocol.TeamInteraction) {
	headline := "Handoff from " + record.SourceTeamID
	item := protocol.NormalizeTeamWorkItem(protocol.TeamWorkItem{
		WorkItemID: record.WorkItemID, TeamID: record.TargetTeamID, RunID: record.RunID,
		IntentProofID: intentProofID, ContractID: contractID, Objective: record.Note, Owner: record.SourceAgentID,
		ExecutionShape: protocol.TeamExecutionShapeDelegatedWork, State: protocol.TeamWorkStateQueued,
		ExpectedOutputs: []string{"Team output that uses the handed-off inputs"},
		ExpectedProof:   []string{"handoff_input_read for each input relied on", "Team result with retained output refs"},
		TargetRef:       &protocol.TargetRef{Type: "team_handoff", ID: record.HandoffID, RunID: record.RunID, TeamID: record.TargetTeamID, WorkItemID: record.WorkItemID, Label: headline},
	})
	event := protocol.TeamStatusEvent{
		EventID: uuid.NewString(), TeamID: item.TeamID, WorkItemID: item.WorkItemID, RunID: item.RunID,
		IntentProofID: item.IntentProofID, ContractID: item.ContractID, State: protocol.TeamWorkStateQueued,
		Headline: headline, Details: fmt.Sprintf("Team %s handed %d input(s) to this team: %s", record.SourceTeamID, len(record.Inputs), record.Note),
		ConfidencePosture: "pending_team_acceptance", NextAction: "Wait for the team to accept and read the handed-off inputs.",
		ExpectedOutputs: item.ExpectedOutputs, ExpectedProof: item.ExpectedProof,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: swarm.HandoffSourceChannel, PayloadKind: string(protocol.PayloadKindStatus), Version: "v1",
	}
	interaction := protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(), TeamID: item.TeamID, WorkItemID: item.WorkItemID, RunID: item.RunID,
		IntentProofID: item.IntentProofID, ContractID: item.ContractID,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: swarm.HandoffSourceChannel, ActorRef: record.SourceAgentID,
		Verb: teamHandoffVerbQueued, Summary: headline + " (queued): " + record.Note, PayloadKind: string(protocol.PayloadKindCommand),
		PayloadRef: record.HandoffID, Payload: map[string]any{"handoff_id": record.HandoffID, "source_team_id": record.SourceTeamID, "artifact_refs": record.Inputs, "expected_action": record.ExpectedAction},
		Version: "v1",
	})
	return item, event, interaction
}

func teamHandoffOutboxItem(record teamHandoffRecord, item protocol.TeamWorkItem) dispatchoutbox.Item {
	key := "team-handoff:" + record.HandoffID
	payload, _ := json.Marshal(swarm.HandoffCommand{
		HandoffID: record.HandoffID, WorkItemID: item.WorkItemID, TargetTeamID: item.TeamID, SourceTeamID: record.SourceTeamID,
		RunID: record.RunID, Note: record.Note, ExpectedAction: record.ExpectedAction, IdempotencyKey: key, Inputs: record.Inputs,
	})
	return dispatchoutbox.Item{
		ID: uuid.NewString(), IdempotencyKey: key, DispatchKind: teamHandoffDispatchKind, RunID: record.RunID,
		IntentProofID: item.IntentProofID, ContractID: item.ContractID, TeamID: item.TeamID, WorkItemID: item.WorkItemID,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: swarm.HandoffSourceChannel, PayloadKind: string(protocol.PayloadKindCommand),
		Payload: payload, Recovery: json.RawMessage(`{"action":"retry_team_handoff","operator_required":false}`),
	}
}

// dispatchClaimedTeamHandoff notifies the receiving team on its command topic.
func (s *AdminServer) dispatchClaimedTeamHandoff(ctx context.Context, item *dispatchoutbox.Item) error {
	return s.dispatchTeamHandoff(ctx, item, &sqlTeamHandoffStore{server: s})
}

// dispatchTeamHandoff records handoff_sent only after a successful publish.
// NATS down leaves the handoff queued for retry; the final failed attempt
// records handoff_dispatch_failed, which shows as needs_attention.
func (s *AdminServer) dispatchTeamHandoff(ctx context.Context, item *dispatchoutbox.Item, store teamHandoffStore) error {
	var command swarm.HandoffCommand
	if err := json.Unmarshal(item.Payload, &command); err != nil {
		_ = s.DispatchOutbox.MarkFailed(ctx, item.ID, err)
		return fmt.Errorf("decode team handoff %s: %w", item.ID, err)
	}
	if s.NC == nil || !s.NC.IsConnected() {
		return s.retryOrFailTeamHandoff(ctx, item, command, store, errors.New("team bus unavailable"))
	}
	envelope, err := swarm.HandoffCommandEnvelope(command)
	if err != nil {
		return s.retryOrFailTeamHandoff(ctx, item, command, store, err)
	}
	if err := s.NC.Publish(fmt.Sprintf(protocol.TopicTeamInternalCommand, item.TeamID), envelope); err != nil {
		return s.retryOrFailTeamHandoff(ctx, item, command, store, err)
	}
	if err := s.NC.FlushTimeout(2 * time.Second); err != nil {
		return s.retryOrFailTeamHandoff(ctx, item, command, store, err)
	}
	// Until handoff_sent is recorded the outbox row stays claimable; a
	// redelivery is deduplicated by the team's command idempotency key.
	if err := store.RecordInteraction(ctx, teamHandoffDispatchInteraction(item, command, teamHandoffVerbSent, "Team notified of the handoff", nil)); err != nil {
		return s.retryOrFailTeamHandoff(ctx, item, command, store, fmt.Errorf("record handoff_sent: %w", err))
	}
	return s.DispatchOutbox.MarkCompleted(ctx, item.ID)
}

func teamHandoffDispatchInteraction(item *dispatchoutbox.Item, command swarm.HandoffCommand, verb, summary string, cause error) protocol.TeamInteraction {
	payload := map[string]any{"handoff_id": command.HandoffID, "attempt": item.AttemptCount, "subject": fmt.Sprintf(protocol.TopicTeamInternalCommand, item.TeamID)}
	if cause != nil {
		payload["error"] = cause.Error()
	}
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(), TeamID: item.TeamID, WorkItemID: item.WorkItemID, RunID: item.RunID,
		IntentProofID: item.IntentProofID, ContractID: item.ContractID,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: swarm.HandoffSourceChannel, ActorRef: "core.team_handoff",
		Verb: verb, Summary: summary, PayloadKind: string(protocol.PayloadKindStatus), PayloadRef: command.HandoffID, Payload: payload, Version: "v1",
	})
}

func (s *AdminServer) retryOrFailTeamHandoff(ctx context.Context, item *dispatchoutbox.Item, command swarm.HandoffCommand, store teamHandoffStore, cause error) error {
	if item.AttemptCount < teamHandoffMaxAttempts {
		return s.DispatchOutbox.MarkRetry(ctx, item.ID, cause, time.Duration(item.AttemptCount)*time.Second)
	}
	if err := store.RecordInteraction(ctx, teamHandoffDispatchInteraction(item, command, teamHandoffVerbFailed, "The team could not be notified of the handoff", cause)); err != nil {
		log.Printf("team handoff: dispatch failure for %s not recorded: %v", command.HandoffID, err)
	}
	return s.DispatchOutbox.MarkFailed(ctx, item.ID, cause)
}

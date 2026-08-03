package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type teamWorkSignalProjection struct {
	server *AdminServer
}

// StartTeamWorkSignalProjection projects correlated team signals into Active Work.
func StartTeamWorkSignalProjection(ctx context.Context, s *AdminServer) error {
	if s == nil || s.getDB() == nil {
		return fmt.Errorf("team work signal projection requires database")
	}
	if s.NC == nil || !s.NC.IsConnected() {
		return fmt.Errorf("team work signal projection requires connected NATS")
	}
	projection := &teamWorkSignalProjection{server: s}
	subs := make([]*nats.Subscription, 0, 2)
	for _, subject := range []string{protocol.TopicTeamSignalStatusWild, protocol.TopicTeamSignalResultWild} {
		sub, err := s.NC.Subscribe(subject, projection.handleNATSMessage)
		if err != nil {
			for _, existing := range subs {
				_ = existing.Unsubscribe()
			}
			return err
		}
		subs = append(subs, sub)
	}
	go func() {
		<-ctx.Done()
		for _, sub := range subs {
			_ = sub.Unsubscribe()
		}
	}()
	return nil
}

func (p *teamWorkSignalProjection) handleNATSMessage(msg *nats.Msg) {
	if msg == nil {
		return
	}
	if err := p.project(context.Background(), msg.Subject, msg.Data); err != nil {
		log.Printf("team work signal projection: %v", err)
	}
}

func (p *teamWorkSignalProjection) project(ctx context.Context, subject string, data []byte) error {
	env, payload, ok := parseTeamWorkSignalEnvelope(data)
	if !ok {
		return fmt.Errorf("ignored malformed signal envelope on %s", subject)
	}
	teamID := firstNonEmptyString(env.Meta.TeamID, teamIDFromSignalSubject(subject))
	workItemID := signalWorkItemID(payload)
	if strings.TrimSpace(teamID) == "" || strings.TrimSpace(workItemID) == "" {
		log.Printf("team work signal projection: ignored uncorrelated signal on %s (team_id=%q work_item_id=%q)", subject, teamID, workItemID)
		return nil
	}
	item, err := p.server.getTeamWorkItemDB(ctx, teamID, workItemID)
	if err != nil {
		return fmt.Errorf("ignored signal on %s: work item %s/%s not found: %w", subject, teamID, workItemID, err)
	}
	if item.State == protocol.TeamWorkStateArchived {
		log.Printf("team work signal projection: ignored signal on %s for archived work item %s/%s", subject, teamID, workItemID)
		return nil
	}
	payloadKind := env.Meta.PayloadKind
	if payloadKind == "" {
		payloadKind = payloadKindFromSignalSubject(subject)
	}
	incomingOutputRefs := projectedSignalOutputRefs(item, env, payload)
	candidateOutputRefs := mergeTeamOutputRefs(item.OutputRefs, incomingOutputRefs)
	projectedState := projectedSignalState(item, payloadKind, payload, incomingOutputRefs)
	if projectedSignalWouldRegress(item.State, projectedState) {
		log.Printf("team work signal projection: ignored stale %s signal for terminal work item %s/%s", projectedState, teamID, workItemID)
		return nil
	}
	item.State = projectedState
	var validationDispatch *teamWorkValidationDispatchPayload
	validationIssue := ""
	if payloadKind == protocol.PayloadKindResult && item.State == protocol.TeamWorkStateOutputReady {
		prepared, prepareErr := prepareTeamWorkValidation(item, candidateOutputRefs)
		if prepareErr != nil {
			item.State = protocol.TeamWorkStateDegraded
			validationIssue = "validation_plan_incomplete"
		} else if prepared != nil {
			item.State = protocol.TeamWorkStateReviewing
			validationDispatch = prepared
		}
	}
	item.NeedsOperator = item.State == protocol.TeamWorkStateNeedsOperator || item.State == protocol.TeamWorkStateDegraded
	if item.State == protocol.TeamWorkStateDegraded {
		item.DegradationState = firstNonEmptyString(validationIssue, stringField(payload, "degradation_state"), stringField(payload, "degradation"), item.DegradationState)
		if item.DegradationState == "" {
			item.DegradationState = deliverableResultOutputIssue(item, payloadKind, incomingOutputRefs)
		}
		item.RecoveryOptions = projectedRecoveryOptionsForItem(item, payload)
	} else if item.State != protocol.TeamWorkStateDegraded {
		item.DegradationState = ""
		item.RecoveryOptions = nil
	}
	db := p.server.getDB()
	if db == nil {
		return fmt.Errorf("team work signal projection requires database")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin projected team work transaction: %w", err)
	}
	defer tx.Rollback()
	receiptID, accepted, err := claimProjectedTeamSignal(ctx, tx, item, subject, payloadKind, payload, data)
	if err != nil {
		return fmt.Errorf("accept projected team signal: %w", err)
	}
	if !accepted {
		return nil
	}
	finalResult, err := p.isFinalLinkedTeamResult(ctx, tx, item, payloadKind)
	if err != nil {
		return err
	}
	proofArtifactID, err := p.recordAsyncCompletionProof(ctx, tx, item, payloadKind, incomingOutputRefs, finalResult)
	if err != nil {
		return err
	}
	incomingOutputRefs = stampTeamOutputRefsWithProof(incomingOutputRefs, proofArtifactID)
	item.OutputRefs = mergeTeamOutputRefs(item.OutputRefs, incomingOutputRefs)
	item.ProofRefs = mergeTeamSignalStrings(item.ProofRefs, proofRefsFromTeamOutputRefs(incomingOutputRefs))
	item.AuditRefs = mergeTeamSignalStrings(item.AuditRefs, auditRefsFromTeamOutputRefs(incomingOutputRefs))
	event := projectedSignalStatusEvent(item, env, subject, payloadKind, payload, incomingOutputRefs)
	event.EventID = receiptID
	if err := p.server.insertTeamStatusEventExec(ctx, tx, &event); err != nil {
		return fmt.Errorf("insert projected team status event: %w", err)
	}
	if err := p.server.updateTeamWorkItemLastEventExec(ctx, tx, &item, event); err != nil {
		return fmt.Errorf("update projected team work item: %w", err)
	}
	interaction := projectedSignalInteraction(item, env, subject, payloadKind, payload)
	if err := p.server.insertTeamInteractionExec(ctx, tx, &interaction); err != nil {
		return fmt.Errorf("insert projected team interaction: %w", err)
	}
	validationKey := ""
	if validationDispatch != nil {
		validationKey, err = p.server.stageTeamWorkValidationTx(ctx, tx, item, *validationDispatch)
		if err != nil {
			return fmt.Errorf("stage team output validation: %w", err)
		}
	}
	if finalResult {
		if err := p.server.markRunCompletedTx(tx, item.RunID, item.IntentProofID); err != nil {
			return fmt.Errorf("complete linked execution run: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit projected team work transaction: %w", err)
	}
	if validationKey != "" {
		if err := p.server.DispatchOutbox.Activate(ctx, validationKey); err != nil {
			log.Printf("team work signal projection: validation dispatch %s remains staged for recovery: %v", validationKey, err)
		}
	}
	if payloadKind == protocol.PayloadKindResult {
		p.server.broadcastTeamWorkResultThreadEvent(item, event)
	}
	return nil
}

func projectedSignalWouldRegress(current, incoming protocol.TeamWorkState) bool {
	if current == protocol.TeamWorkStateOutputReady {
		return true
	}
	if current == protocol.TeamWorkStateReviewing {
		return incoming != protocol.TeamWorkStateOutputReady && incoming != protocol.TeamWorkStateDegraded
	}
	if current != protocol.TeamWorkStateDegraded && current != protocol.TeamWorkStateNeedsOperator {
		return false
	}
	return incoming != protocol.TeamWorkStateOutputReady && incoming != protocol.TeamWorkStateDegraded && incoming != protocol.TeamWorkStateNeedsOperator
}

func parseTeamWorkSignalEnvelope(data []byte) (protocol.SignalEnvelope, map[string]any, bool) {
	var env protocol.SignalEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return env, nil, false
	}
	payload := map[string]any{}
	if len(env.Payload) > 0 && string(env.Payload) != "null" {
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			payload = map[string]any{"raw": string(env.Payload)}
		}
	}
	if strings.TrimSpace(env.Text) != "" {
		payload["text"] = env.Text
	}
	return env, payload, true
}

func projectedSignalState(item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind, payload map[string]any, outputRefs []protocol.TeamOutputRef) protocol.TeamWorkState {
	if state := protocol.TeamWorkState(stringField(payload, "state")); protocol.IsTeamExecutionState(state) {
		if state == protocol.TeamWorkStateOutputReady && deliverableResultMissingOutputs(item, payloadKind, outputRefs) {
			return protocol.TeamWorkStateDegraded
		}
		return state
	}
	if payloadKind == protocol.PayloadKindResult {
		if deliverableResultMissingOutputs(item, payloadKind, outputRefs) {
			return protocol.TeamWorkStateDegraded
		}
		return protocol.TeamWorkStateOutputReady
	}
	return item.State
}

func projectedSignalStatusEvent(item protocol.TeamWorkItem, env protocol.SignalEnvelope, subject string, payloadKind protocol.SignalPayloadKind, payload map[string]any, outputRefs []protocol.TeamOutputRef) protocol.TeamStatusEvent {
	return protocol.TeamStatusEvent{
		TeamID:            item.TeamID,
		WorkItemID:        item.WorkItemID,
		RunID:             firstNonEmptyString(env.Meta.RunID, item.RunID),
		IntentProofID:     item.IntentProofID,
		ContractID:        item.ContractID,
		ProofID:           item.ProofID,
		State:             item.State,
		Headline:          projectedHeadlineForItem(item, payloadKind, payload),
		Details:           projectedDetailsForItem(item, payload),
		ConfidencePosture: stringField(payload, "confidence_posture"),
		BlockedBy:         stringSliceField(payload, "blocked_by"),
		NextAction:        projectedNextActionForItem(item, payload),
		ExpectedOutputs:   item.ExpectedOutputs,
		ExpectedProof:     item.ExpectedProof,
		ExecutionMode:     item.ExecutionMode,
		WorkIntent:        item.WorkIntent,
		OutputRefs:        outputRefs,
		SourceKind:        string(env.Meta.SourceKind),
		SourceChannel:     firstNonEmptyString(env.Meta.SourceChannel, subject),
		PayloadKind:       string(payloadKind),
		Version:           "v1",
	}
}

func projectedSignalInteraction(item protocol.TeamWorkItem, env protocol.SignalEnvelope, subject string, payloadKind protocol.SignalPayloadKind, payload map[string]any) protocol.TeamInteraction {
	verb := "status"
	if item.State == protocol.TeamWorkStateDegraded {
		verb = "degraded"
	} else if item.State == protocol.TeamWorkStateReviewing {
		verb = "reviewing"
	} else if payloadKind == protocol.PayloadKindResult {
		verb = "output_ready"
	}
	interactionPayload := projectedInteractionPayload(item, payload)
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(),
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		RunID:         firstNonEmptyString(env.Meta.RunID, item.RunID),
		IntentProofID: item.IntentProofID,
		ContractID:    item.ContractID,
		ProofID:       item.ProofID,
		SourceKind:    string(env.Meta.SourceKind),
		SourceChannel: firstNonEmptyString(env.Meta.SourceChannel, subject),
		ActorRef:      firstNonEmptyString(env.Meta.AgentID, item.TeamID),
		Verb:          verb,
		Summary:       projectedSummary(payloadKind, payload),
		PayloadKind:   string(payloadKind),
		Payload:       interactionPayload,
		Version:       "v1",
	})
}

func projectedInteractionPayload(item protocol.TeamWorkItem, payload map[string]any) map[string]any {
	copied := map[string]any{}
	for key, value := range payload {
		copied[key] = value
	}
	if len(item.ExpectedOutputs) > 0 {
		copied["expected_outputs"] = item.ExpectedOutputs
	}
	if len(item.ExpectedProof) > 0 {
		copied["expected_proof"] = item.ExpectedProof
	}
	return copied
}

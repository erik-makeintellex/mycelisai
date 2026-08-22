package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/outputvalidation"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) finalizeTeamWorkValidation(
	ctx context.Context,
	outboxItem *dispatchoutbox.Item,
	payload teamWorkValidationDispatchPayload,
	report outputvalidation.Report,
	passed bool,
	degradation string,
) error {
	current, err := s.getTeamWorkItemDB(ctx, outboxItem.TeamID, outboxItem.WorkItemID)
	if err != nil {
		return err
	}
	currentDigest, err := teamWorkOutputDigest(current.OutputRefs)
	if err != nil {
		passed = false
		degradation = "runtime_validation_stale"
	} else if currentDigest != payload.ContentDigest {
		return nil // A newer retained candidate owns its own digest-keyed validation job.
	}
	validationRef := path.Join(payload.EvidenceRef, "validation-report.json")
	if err := retainServerValidationReport(payload.EvidenceRef, report); err != nil {
		return err
	}

	tx, err := s.getDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	item, err := lockTeamWorkItemTx(ctx, tx, outboxItem.TeamID, outboxItem.WorkItemID)
	if err != nil {
		return err
	}
	lockedDigest, err := teamWorkOutputDigest(item.OutputRefs)
	if err != nil || lockedDigest != payload.ContentDigest {
		return tx.Commit() // The locked record now belongs to a newer candidate.
	}
	if item.State == protocol.TeamWorkStateOutputReady || item.State == protocol.TeamWorkStateDegraded {
		return tx.Commit()
	}
	if item.State != protocol.TeamWorkStateReviewing {
		return fmt.Errorf("work item %s is %s, not reviewing", item.WorkItemID, item.State)
	}
	item.OutputRefs = stampTeamOutputRefsWithValidation(item.OutputRefs, validationRef)
	item.NeedsOperator = !passed
	item.DegradationState = degradation
	item.RecoveryOptions = nil
	if passed {
		item.State = protocol.TeamWorkStateOutputReady
	} else {
		item.State = protocol.TeamWorkStateDegraded
		item.RecoveryOptions = []string{"Ask Soma to have the same team repair the failed primary workflow, then return a new retained candidate for validation."}
	}

	finalResult, err := (&teamWorkSignalProjection{server: s}).isFinalLinkedTeamResult(ctx, tx, item, protocol.PayloadKindResult)
	if err != nil {
		return err
	}
	proofID := ""
	if passed {
		proofID, err = (&teamWorkSignalProjection{server: s}).recordRuntimeCompletionProof(ctx, tx, item, item.OutputRefs, finalResult, completionValidationEvidence{
			ValidationRef: validationRef, ContentDigest: payload.ContentDigest,
			EvidenceRefs: validationEvidenceRefs(report, validationRef),
		})
		if err != nil {
			return err
		}
		item.OutputRefs = stampTeamOutputRefsWithProof(item.OutputRefs, proofID)
		item.ProofRefs = mergeTeamSignalStrings(item.ProofRefs, proofRefsFromTeamOutputRefs(item.OutputRefs))
	}
	event := runtimeValidationStatusEvent(item, report, passed)
	if err := s.insertTeamStatusEventExec(ctx, tx, &event); err != nil {
		return err
	}
	if err := s.updateTeamWorkItemLastEventExec(ctx, tx, &item, event); err != nil {
		return err
	}
	interaction := runtimeValidationInteraction(item, report, passed, validationRef)
	if err := s.insertTeamInteractionExec(ctx, tx, &interaction); err != nil {
		return err
	}
	if finalResult {
		if err := s.markRunCompletedTx(tx, item.RunID, item.IntentProofID); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.broadcastTeamWorkResultThreadEvent(item, event)
	return nil
}

func lockTeamWorkItemTx(ctx context.Context, tx *sql.Tx, teamID, workItemID string) (protocol.TeamWorkItem, error) {
	return scanTeamWorkItem(tx.QueryRowContext(ctx, `
		SELECT id::text, team_id, COALESCE(run_id::text,''), COALESCE(intent_proof_id::text,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), objective, scope, owner,
		       execution_shape, execution_mode, work_intent, expected_outputs, expected_proof, capability_requirements,
		       governance_posture, state, COALESCE(last_event, 'null'::jsonb), needs_operator,
		       degradation_state, recovery_options, output_refs, proof_refs, audit_refs,
		       created_at, updated_at, version
		FROM team_work_items
		WHERE tenant_id='default' AND team_id=$1 AND id=$2
		FOR UPDATE`, strings.TrimSpace(teamID), workItemID))
}

func retainServerValidationReport(evidenceRef string, report outputvalidation.Report) error {
	target, _, err := resolveWorkspacePath(evidenceRef, false)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pathToFile(target, "validation-report.json"), append(raw, '\n'), 0o644)
}

func pathToFile(folder, name string) string {
	return strings.TrimRight(folder, "\\/") + string(os.PathSeparator) + name
}

func stampTeamOutputRefsWithValidation(refs []protocol.TeamOutputRef, validationRef string) []protocol.TeamOutputRef {
	stamped := append([]protocol.TeamOutputRef(nil), refs...)
	for index := range stamped {
		stamped[index].ValidationRef = validationRef
	}
	return stamped
}

func runtimeValidationStatusEvent(item protocol.TeamWorkItem, report outputvalidation.Report, passed bool) protocol.TeamStatusEvent {
	headline := "Deliverable verified"
	details := "Soma checked the approved primary workflow and retained the validation evidence."
	nextAction := "Open the deliverable or continue refining it with Soma."
	if !passed {
		headline = "Deliverable needs repair"
		details = validationDiagnosticSummary(report)
		nextAction = firstString(item.RecoveryOptions)
	}
	return protocol.TeamStatusEvent{
		EventID: uuid.NewString(), TeamID: item.TeamID, WorkItemID: item.WorkItemID,
		RunID: item.RunID, IntentProofID: item.IntentProofID, ContractID: item.ContractID,
		State: item.State, Headline: headline, Details: details, NextAction: nextAction,
		ExecutionMode: item.ExecutionMode, WorkIntent: item.WorkIntent, OutputRefs: item.OutputRefs,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: "team-work.runtime-validation",
		PayloadKind: string(protocol.PayloadKindResult), Timestamp: time.Now().UTC(), Version: "v1",
	}
}

func runtimeValidationInteraction(item protocol.TeamWorkItem, report outputvalidation.Report, passed bool, validationRef string) protocol.TeamInteraction {
	verb := "degraded"
	if passed {
		verb = "output_ready"
	}
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		InteractionID: uuid.NewString(), TeamID: item.TeamID, WorkItemID: item.WorkItemID,
		RunID: item.RunID, IntentProofID: item.IntentProofID, ContractID: item.ContractID,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: "team-work.runtime-validation",
		ActorRef: "soma", Verb: verb, Summary: validationDiagnosticSummary(report),
		PayloadKind: string(protocol.PayloadKindResult), PayloadRef: validationRef,
		Payload: map[string]any{"validation_ref": validationRef, "validation_status": report.Status}, Version: "v1",
	})
}

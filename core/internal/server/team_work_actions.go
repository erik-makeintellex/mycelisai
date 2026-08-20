package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

type teamWorkActionRequest struct {
	Action         protocol.TeamWorkAction `json:"action"`
	Summary        string                  `json:"summary,omitempty"`
	ActorRef       string                  `json:"actor_ref,omitempty"`
	SourceKind     string                  `json:"source_kind,omitempty"`
	SourceChannel  string                  `json:"source_channel,omitempty"`
	PayloadKind    string                  `json:"payload_kind,omitempty"`
	Payload        map[string]any          `json:"payload,omitempty"`
	AuditRefs      []string                `json:"audit_refs,omitempty"`
	IdempotencyKey string                  `json:"idempotency_key,omitempty"`
	Result         string                  `json:"result,omitempty"`
	EvidenceRefs   []string                `json:"evidence_refs,omitempty"`
	ExpectedRunID  string                  `json:"-"`
	RecordedAt     time.Time               `json:"-"`
}

var errTeamWorkActionRejected = errors.New("team work action rejected")

// HandleTeamWorkAction applies a durable operator control to an existing work item.
// POST /api/v1/teams/{id}/work/{workItemId}/actions
func (s *AdminServer) HandleTeamWorkAction(w http.ResponseWriter, r *http.Request) {
	teamID, workItemID := teamWorkPathIDs(r)
	if teamID == "" || workItemID == "" {
		respondAPIError(w, "team_id and work_item_id are required", http.StatusBadRequest)
		return
	}
	if err := validateOptionalUUID("work_item_id", workItemID); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req teamWorkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	action := protocol.NormalizeTeamWorkAction(req.Action)
	if action == "" {
		respondAPIError(w, "action is required", http.StatusBadRequest)
		return
	}
	if action == protocol.TeamWorkActionVerifyExternalOutcome {
		if req.Result == "" && req.Payload != nil {
			req.Result, _ = req.Payload["result"].(string)
		}
		if len(req.EvidenceRefs) == 0 && req.Payload != nil {
			req.EvidenceRefs = stringSliceFromAny(req.Payload["evidence_refs"])
		}
		req.Result = protocol.NormalizeWorkExternalOutcomeResult(req.Result)
		if req.Result == "" {
			respondAPIError(w, "result must be committed, not_committed, or still_unknown", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Summary) == "" {
			respondAPIError(w, "summary is required for verify_external_outcome", http.StatusBadRequest)
			return
		}
		req.EvidenceRefs = mergeStrings(nil, req.EvidenceRefs)
		req.ActorRef = auditUserLabelFromRequest(r)
		req.RecordedAt = time.Now().UTC()
		req.SourceKind = string(protocol.SourceKindWorkspaceUI)
		req.SourceChannel = "teams.external_outcome_verification"
		req.PayloadKind = "external_outcome_verification"
		req.AuditRefs = nil
	}
	if requiresActionDetail(action) && strings.TrimSpace(req.Summary) == "" && len(req.Payload) == 0 {
		respondAPIError(w, "summary or payload is required for "+string(action), http.StatusBadRequest)
		return
	}

	item, steeringKey, err := s.applyTeamWorkAction(r.Context(), teamID, workItemID, req, action)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondAPIError(w, "team work item not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, errTeamWorkActionRejected) {
			message := strings.TrimPrefix(err.Error(), errTeamWorkActionRejected.Error()+": ")
			respondAPIError(w, message, http.StatusBadRequest)
			return
		}
		respondAPIError(w, "Failed to update team work: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	if steeringKey != "" {
		if err := s.DispatchOutbox.Activate(r.Context(), steeringKey); err != nil {
			log.Printf("team steering %s remains staged for dispatch recovery: %v", steeringKey, err)
		}
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(item))
}

func (s *AdminServer) applyTeamWorkAction(
	ctx context.Context,
	teamID string,
	workItemID string,
	req teamWorkActionRequest,
	action protocol.TeamWorkAction,
) (protocol.TeamWorkItem, string, error) {
	db := s.getDB()
	if db == nil {
		return protocol.TeamWorkItem{}, "", errors.New("database not available")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return protocol.TeamWorkItem{}, "", err
	}
	defer tx.Rollback()

	item, err := s.getTeamWorkItemForUpdateTx(ctx, tx, teamID, workItemID)
	if err != nil {
		return protocol.TeamWorkItem{}, "", err
	}
	if expected := strings.TrimSpace(req.ExpectedRunID); expected != "" && item.RunID != expected {
		return protocol.TeamWorkItem{}, "", fmt.Errorf("%w: active work no longer matches this run", errTeamWorkActionRejected)
	}
	targetState, err := protocol.ApplyTeamWorkAction(item, action)
	if err != nil {
		return protocol.TeamWorkItem{}, "", fmt.Errorf("%w: %v", errTeamWorkActionRejected, err)
	}
	if action == protocol.TeamWorkActionVerifyExternalOutcome {
		item, err = protocol.ApplyExternalOutcomeVerification(item, protocol.WorkExternalOutcomeVerification{
			Result: req.Result, ActorRef: req.ActorRef, Summary: req.Summary,
			EvidenceRefs: req.EvidenceRefs, RecordedAt: req.RecordedAt,
		})
		if err != nil {
			return protocol.TeamWorkItem{}, "", fmt.Errorf("%w: %v", errTeamWorkActionRejected, err)
		}
	} else {
		item.State = targetState
	}
	applyTeamWorkActionPosture(&item)
	if err := protocol.ValidateTeamWorkItem(item); err != nil {
		return protocol.TeamWorkItem{}, "", fmt.Errorf("%w: %v", errTeamWorkActionRejected, err)
	}
	event := teamWorkActionStatusEvent(item, req, action)
	if err := s.insertTeamStatusEventExec(ctx, tx, &event); err != nil {
		return protocol.TeamWorkItem{}, "", err
	}
	if err := s.updateTeamWorkItemLastEventExec(ctx, tx, &item, event); err != nil {
		return protocol.TeamWorkItem{}, "", err
	}
	item.LastEvent = &event
	interaction := teamWorkActionInteraction(item, req, action)
	if err := s.insertTeamInteractionExec(ctx, tx, &interaction); err != nil {
		return protocol.TeamWorkItem{}, "", err
	}
	steeringKey := ""
	if action == protocol.TeamWorkActionSteer && item.RunID != "" && item.IntentProofID != "" {
		steeringKey, err = s.stageTeamWorkSteeringTx(ctx, tx, item, req)
		if err != nil {
			return protocol.TeamWorkItem{}, "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return protocol.TeamWorkItem{}, "", err
	}
	if !event.Timestamp.IsZero() {
		item.UpdatedAt = event.Timestamp
	}
	return item, steeringKey, nil
}

func applyTeamWorkActionPosture(item *protocol.TeamWorkItem) {
	switch item.State {
	case protocol.TeamWorkStateQueued, protocol.TeamWorkStateRunning:
		item.NeedsOperator = false
		item.DegradationState = ""
	case protocol.TeamWorkStatePaused, protocol.TeamWorkStateArchived:
		item.NeedsOperator = false
	}
}

func requiresActionDetail(action protocol.TeamWorkAction) bool {
	return action == protocol.TeamWorkActionSteer || action == protocol.TeamWorkActionRecover || action == protocol.TeamWorkActionVerifyExternalOutcome
}

func teamWorkActionStatusEvent(item protocol.TeamWorkItem, req teamWorkActionRequest, action protocol.TeamWorkAction) protocol.TeamStatusEvent {
	headline, details, nextAction := teamWorkActionCopy(action)
	confidence := "operator_recorded"
	blockedBy := []string(nil)
	if action == protocol.TeamWorkActionVerifyExternalOutcome {
		headline, details, nextAction = teamWorkExternalOutcomeCopy(req.Result)
		confidence = "operator_attested"
		switch req.Result {
		case protocol.WorkExternalOutcomeNotCommitted:
			blockedBy = []string{"new_governed_soma_proposal_required"}
		case protocol.WorkExternalOutcomeStillUnknown:
			blockedBy = []string{protocol.TeamWorkDegradationExternalMutationUnknown}
		}
	}
	if summary := strings.TrimSpace(req.Summary); summary != "" {
		details = summary
	}
	return protocol.TeamStatusEvent{
		TeamID:            item.TeamID,
		WorkItemID:        item.WorkItemID,
		RunID:             item.RunID,
		IntentProofID:     item.IntentProofID,
		ContractID:        item.ContractID,
		ProofID:           item.ProofID,
		State:             item.State,
		Headline:          headline,
		Details:           details,
		ConfidencePosture: confidence,
		BlockedBy:         blockedBy,
		NextAction:        nextAction,
		ExpectedOutputs:   item.ExpectedOutputs,
		ExpectedProof:     item.ExpectedProof,
		ExecutionMode:     item.ExecutionMode,
		WorkIntent:        item.WorkIntent,
		OutputRefs:        item.OutputRefs,
		SourceKind:        defaultString(req.SourceKind, "workspace_ui"),
		SourceChannel:     defaultString(req.SourceChannel, "teams.active_work"),
		PayloadKind:       defaultString(req.PayloadKind, teamWorkActionPayloadKind(action)),
		AuditRefs:         mergeStrings(item.AuditRefs, req.AuditRefs),
		Version:           "v1",
	}
}

func teamWorkActionInteraction(item protocol.TeamWorkItem, req teamWorkActionRequest, action protocol.TeamWorkAction) protocol.TeamInteraction {
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		headline, _, _ := teamWorkActionCopy(action)
		summary = headline
	}
	payload := req.Payload
	if action == protocol.TeamWorkActionVerifyExternalOutcome {
		payload = map[string]any{
			"result": req.Result, "evidence_refs": req.EvidenceRefs, "recorded_at": req.RecordedAt,
		}
	}
	return protocol.NormalizeTeamInteraction(protocol.TeamInteraction{
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		RunID:         item.RunID,
		IntentProofID: item.IntentProofID,
		ContractID:    item.ContractID,
		ProofID:       item.ProofID,
		SourceKind:    defaultString(req.SourceKind, "workspace_ui"),
		SourceChannel: defaultString(req.SourceChannel, "teams.active_work"),
		ActorRef:      defaultString(req.ActorRef, "operator"),
		Verb:          string(action),
		Summary:       summary,
		PayloadKind:   defaultString(req.PayloadKind, teamWorkActionPayloadKind(action)),
		Payload:       payload,
		AuditRefs:     mergeStrings(item.AuditRefs, req.AuditRefs),
		Version:       "v1",
	})
}

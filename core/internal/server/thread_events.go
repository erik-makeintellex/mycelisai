package server

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) broadcastConfirmActionThreadEvent(runID, proofID, contractID string, refs []confirmActionTeamWorkRef) {
	if s.Stream == nil || runID == "" {
		return
	}
	teamID, workItemID, outputs := firstThreadEventTeamWorkRef(refs)
	event := protocol.ThreadEventEnvelope{
		Type:       "thread_event",
		EventType:  protocol.EventTeamWorkStatus,
		ThreadID:   firstNonEmptyString(workItemID, runID),
		ThreadKind: firstNonEmptyString(threadKindForWorkItem(workItemID), "run"),
		EventID:    uuid.NewString(),
		Version:    "v1",
		Meta: protocol.SignalMeta{
			Timestamp:     time.Now().UTC(),
			SourceKind:    protocol.SourceKindWebAPI,
			SourceChannel: "api.intent.confirm-action",
			PayloadKind:   protocol.PayloadKindThreadEvent,
			RunID:         runID,
			TeamID:        teamID,
		},
		Payload: protocol.ThreadEventPayload{
			Kind:            protocol.ThreadEventExecutionStarted,
			Label:           "Execution started",
			Detail:          "Soma accepted the approved work and saved the run receipt for proof.",
			Tone:            "info",
			Status:          "running",
			Href:            "/runs/" + runID,
			HrefLabel:       "Open run receipt",
			TargetReference: "run:" + runID,
			WorkItemID:      workItemID,
			IntentProofID:   proofID,
			ContractID:      contractID,
			ProofID:         proofID,
			OutputRefs:      outputs,
		},
	}
	raw, err := json.Marshal(event)
	if err != nil {
		return
	}
	s.Stream.Broadcast(string(raw))
}

func firstThreadEventTeamWorkRef(refs []confirmActionTeamWorkRef) (string, string, []protocol.TeamOutputRef) {
	outputs := []protocol.TeamOutputRef{}
	var fallbackTeamID, fallbackWorkItemID string
	for _, ref := range refs {
		if fallbackTeamID == "" && (ref.TeamID != "" || ref.WorkItemID != "") {
			fallbackTeamID = ref.TeamID
			fallbackWorkItemID = ref.WorkItemID
		}
		if len(ref.OutputRefs) > 0 {
			if fallbackTeamID == "" {
				fallbackTeamID = ref.TeamID
				fallbackWorkItemID = ref.WorkItemID
			}
			outputs = append(outputs, ref.OutputRefs...)
		}
	}
	return fallbackTeamID, fallbackWorkItemID, outputs
}

func threadKindForWorkItem(workItemID string) string {
	if workItemID == "" {
		return ""
	}
	return "team_work"
}

package server

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
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
			Label:           "Work started",
			Detail:          "Soma handed this to the work bus. It is running, not complete, and you can keep talking here.",
			Tone:            "info",
			Status:          "running",
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

func (s *AdminServer) broadcastTeamWorkResultThreadEvent(item protocol.TeamWorkItem, status protocol.TeamStatusEvent) {
	if s.Stream == nil {
		return
	}
	event := teamWorkResultThreadEvent(item, status)
	raw, err := json.Marshal(event)
	if err != nil {
		return
	}
	s.Stream.Broadcast(string(raw))
}

func teamWorkResultThreadEvent(item protocol.TeamWorkItem, status protocol.TeamStatusEvent) protocol.ThreadEventEnvelope {
	isReady := item.State == protocol.TeamWorkStateOutputReady
	kind := protocol.ThreadEventAttentionNeeded
	label := "Work needs attention"
	tone := "warning"
	state := "degraded"
	detail := firstNonEmptyString(status.Details, status.Headline, "The team could not return a usable deliverable.")
	href, hrefLabel, target := "", "", ""
	if item.State == protocol.TeamWorkStateReviewing {
		kind = protocol.ThreadEventExecutionUpdate
		label = "Checking deliverable"
		tone = "info"
		state = "reviewing"
		detail = firstNonEmptyString(status.Details, status.Headline, "The team returned files. Soma is checking the approved primary workflow before calling the work complete.")
	} else if isReady {
		kind = protocol.ThreadEventResultReady
		label = "Work complete"
		tone = "success"
		state = "completed"
		detail = completedTeamWorkDetail(status, item.OutputRefs)
		href, hrefLabel, target = firstTeamOutputOpenTarget(item.OutputRefs)
	}
	return protocol.ThreadEventEnvelope{
		Type:       "thread_event",
		EventType:  protocol.EventTeamWorkStatus,
		ThreadID:   firstNonEmptyString(item.WorkItemID, item.RunID),
		ThreadKind: firstNonEmptyString(threadKindForWorkItem(item.WorkItemID), "run"),
		EventID:    uuid.NewString(),
		Version:    "v1",
		Meta: protocol.SignalMeta{
			Timestamp:     time.Now().UTC(),
			SourceKind:    protocol.SourceKindSystem,
			SourceChannel: "team-work.result-projection",
			PayloadKind:   protocol.PayloadKindThreadEvent,
			RunID:         item.RunID,
			TeamID:        item.TeamID,
		},
		Payload: protocol.ThreadEventPayload{
			Kind:            kind,
			Label:           label,
			Detail:          detail,
			Tone:            tone,
			Status:          state,
			Href:            href,
			HrefLabel:       hrefLabel,
			TargetReference: target,
			WorkItemID:      item.WorkItemID,
			IntentProofID:   item.IntentProofID,
			ContractID:      item.ContractID,
			ProofID:         completionProofRef(item),
			OutputRefs:      item.OutputRefs,
		},
	}
}

func completionProofRef(item protocol.TeamWorkItem) string {
	for _, ref := range item.OutputRefs {
		if proofID := firstNonEmptyString(ref.ProofID, ref.ProofRef); proofID != "" {
			return proofID
		}
	}
	return firstNonEmptyString(firstTeamSignalString(item.ProofRefs), item.ProofID)
}

func completedTeamWorkDetail(status protocol.TeamStatusEvent, outputs []protocol.TeamOutputRef) string {
	base := firstNonEmptyString(status.Details, status.Headline, "The team returned its deliverable.")
	if summary := completedOutputSummary(outputs); summary != "" {
		return base + " " + summary
	}
	return base
}

func completedOutputSummary(outputs []protocol.TeamOutputRef) string {
	count := len(outputs)
	if count == 0 {
		return ""
	}
	label := firstNonEmptyString(outputs[0].Label, outputs[0].OutputID, "deliverable")
	if count == 1 {
		return fmt.Sprintf("Ready to open: %s.", label)
	}
	return fmt.Sprintf("Ready to review: %s plus %s.", label, moreDeliverablesLabel(count-1))
}

func moreDeliverablesLabel(count int) string {
	if count == 1 {
		return "1 more deliverable"
	}
	return fmt.Sprintf("%d more deliverables", count)
}

func firstTeamOutputOpenTarget(outputs []protocol.TeamOutputRef) (string, string, string) {
	for _, ref := range outputs {
		if !strings.EqualFold(strings.TrimSpace(ref.Kind), "project_package") || strings.TrimSpace(ref.Entrypoint) == "" {
			continue
		}
		openPath := path.Join(strings.ReplaceAll(ref.StorageRef, "\\", "/"), strings.ReplaceAll(ref.Entrypoint, "\\", "/"))
		return workspaceFileOutputHref(openPath), "Open app", openPath
	}
	for _, ref := range outputs {
		openPath := firstNonEmptyString(ref.StorageRef, ref.Entrypoint)
		if openPath != "" {
			return workspaceFileOutputHref(openPath), "Open output", openPath
		}
	}
	return "", "", ""
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

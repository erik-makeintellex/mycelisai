package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func groupBriefWorkflowLogEntry(group CollaborationGroup) groupWorkflowLogEntry {
	return groupWorkflowLogEntry{
		ID:           "group:" + group.ID + ":brief",
		Kind:         "group_brief",
		GroupID:      group.ID,
		Title:        group.Name,
		Summary:      group.GoalStatement,
		State:        group.Status,
		TrustPosture: "trusted_group_record",
		AuditRefs:    compactWorkflowLogStrings([]string{group.CreatedAuditEventID}),
		Timestamp:    group.CreatedAt,
		SourceKind:   string(protocol.SourceKindWorkspaceUI),
		PayloadKind:  "group_brief",
	}
}

func groupLifecycleWorkflowLogEntry(group CollaborationGroup, item groupLifecycleItem, timestamp time.Time) groupWorkflowLogEntry {
	return groupWorkflowLogEntry{
		ID:             "group:" + group.ID + ":lifecycle",
		Kind:           "lifecycle",
		GroupID:        group.ID,
		Title:          "Group lifecycle recommendation",
		Summary:        item.Reason,
		State:          item.Recommendation,
		TrustPosture:   "computed_from_group_outputs_and_team_work",
		Timestamp:      timestamp,
		SourceKind:     string(protocol.SourceKindSystem),
		PayloadKind:    "group_lifecycle",
		CorrelationRef: item.Kind,
	}
}

func (s *AdminServer) groupBroadcastWorkflowLogEntries(group CollaborationGroup) []groupWorkflowLogEntry {
	if s.GroupBus == nil {
		return nil
	}
	snapshot := s.GroupBus.Snapshot(s.NC != nil && s.NC.IsConnected())
	if snapshot.LastGroupID != group.ID || strings.TrimSpace(snapshot.LastMessage) == "" {
		return nil
	}
	timestamp := group.UpdatedAt
	if parsed, err := time.Parse(time.RFC3339Nano, snapshot.LastPublishedAt); err == nil {
		timestamp = parsed.UTC()
	}
	return []groupWorkflowLogEntry{{
		ID:            "group:" + group.ID + ":last-broadcast",
		Kind:          "broadcast",
		GroupID:       group.ID,
		Title:         "Latest group message",
		Summary:       snapshot.LastMessage,
		State:         snapshot.Status,
		TrustPosture:  "transient_bus_monitor",
		Timestamp:     timestamp,
		SourceKind:    string(protocol.SourceKindWorkspaceUI),
		SourceChannel: fmt.Sprintf(protocol.TopicGroupCollabFmt, group.ID),
		PayloadKind:   "group_broadcast",
	}}
}

func teamWorkWorkflowLogEntry(group CollaborationGroup, item protocol.TeamWorkItem, includeAudit bool) groupWorkflowLogEntry {
	entry := groupWorkflowLogEntry{
		ID:            "team:" + item.TeamID + ":work:" + item.WorkItemID,
		Kind:          "team_work",
		GroupID:       group.ID,
		TeamID:        item.TeamID,
		WorkItemID:    item.WorkItemID,
		RunID:         item.RunID,
		Title:         item.Objective,
		State:         string(item.State),
		TrustPosture:  teamWorkTrustPosture(item),
		OutputRefs:    item.OutputRefs,
		ProofRefs:     compactWorkflowLogStrings(append([]string{item.ProofID}, item.ProofRefs...)),
		Recovery:      item.RecoveryOptions,
		NeedsOperator: item.NeedsOperator,
		Timestamp:     item.UpdatedAt,
		SourceKind:    string(protocol.SourceKindSystem),
		PayloadKind:   "team_work",
	}
	if item.LastEvent != nil {
		entry.Summary = item.LastEvent.Headline
		entry.SourceKind = item.LastEvent.SourceKind
		entry.SourceChannel = item.LastEvent.SourceChannel
		entry.PayloadKind = item.LastEvent.PayloadKind
		if strings.TrimSpace(item.LastEvent.ConfidencePosture) != "" {
			entry.TrustPosture = item.LastEvent.ConfidencePosture
		}
	}
	if includeAudit {
		entry.AuditRefs = compactWorkflowLogStrings(item.AuditRefs)
	}
	return entry
}

func teamWorkOutputWorkflowLogEntries(group CollaborationGroup, item protocol.TeamWorkItem, includeAudit bool) []groupWorkflowLogEntry {
	entries := make([]groupWorkflowLogEntry, 0, len(item.OutputRefs))
	for _, ref := range item.OutputRefs {
		entry := groupWorkflowLogEntry{
			ID:           firstNonEmptyString(ref.OutputID, "team:"+item.TeamID+":work:"+item.WorkItemID+":output:"+ref.StorageRef),
			Kind:         "team_output_ref",
			GroupID:      group.ID,
			TeamID:       firstNonEmptyString(ref.TeamID, item.TeamID),
			WorkItemID:   firstNonEmptyString(ref.WorkItemID, item.WorkItemID),
			RunID:        firstNonEmptyString(ref.RunID, item.RunID),
			Title:        firstNonEmptyString(ref.Label, ref.StorageRef, ref.OutputID, "Team output"),
			Summary:      ref.Kind,
			State:        string(item.State),
			TrustPosture: "linked_from_team_work",
			StorageRef:   ref.StorageRef,
			Entrypoint:   ref.Entrypoint,
			ProofRefs:    compactWorkflowLogStrings([]string{ref.ProofRef, ref.ProofID, ref.ValidationRef}),
			Timestamp:    firstNonZeroTime(ref.CreatedAt, item.UpdatedAt),
			SourceKind:   string(protocol.SourceKindInternalTool),
			PayloadKind:  "team_output_ref",
		}
		if includeAudit {
			entry.AuditRefs = compactWorkflowLogStrings(ref.AuditRefs)
		}
		entries = append(entries, entry)
	}
	return entries
}

func teamWorkProofWorkflowLogEntries(group CollaborationGroup, item protocol.TeamWorkItem, includeAudit bool) []groupWorkflowLogEntry {
	proofRefs := compactWorkflowLogStrings(append([]string{item.IntentProofID, item.ContractID, item.ProofID}, item.ProofRefs...))
	if len(proofRefs) == 0 && (!includeAudit || len(item.AuditRefs) == 0) {
		return nil
	}
	entry := groupWorkflowLogEntry{
		ID:           "team:" + item.TeamID + ":work:" + item.WorkItemID + ":proof",
		Kind:         "proof_cue",
		GroupID:      group.ID,
		TeamID:       item.TeamID,
		WorkItemID:   item.WorkItemID,
		RunID:        item.RunID,
		Title:        "Proof and audit cues",
		Summary:      "Refs retained for reconstructing this team work item.",
		State:        string(item.State),
		TrustPosture: "refs_only",
		ProofRefs:    proofRefs,
		Timestamp:    item.UpdatedAt,
		SourceKind:   string(protocol.SourceKindSystem),
		PayloadKind:  "proof_cue",
	}
	if includeAudit {
		entry.AuditRefs = compactWorkflowLogStrings(item.AuditRefs)
	}
	return []groupWorkflowLogEntry{entry}
}

func artifactWorkflowLogEntry(group CollaborationGroup, artifact artifacts.Artifact, includeAudit bool) groupWorkflowLogEntry {
	entry := groupWorkflowLogEntry{
		ID:           "artifact:" + artifact.ID.String(),
		Kind:         "retained_artifact",
		GroupID:      group.ID,
		TeamID:       artifactTeamID(artifact),
		ArtifactID:   artifact.ID.String(),
		Title:        artifact.Title,
		Summary:      string(artifact.ArtifactType),
		State:        artifact.Status,
		TrustPosture: "retained_output",
		StorageRef:   firstNonEmptyString(artifact.FilePath, artifact.ID.String()),
		Timestamp:    artifact.CreatedAt,
		SourceKind:   string(protocol.SourceKindInternalTool),
		PayloadKind:  "artifact",
	}
	if includeAudit {
		entry.AuditRefs = artifactAuditRefs(artifact.Metadata)
	}
	return entry
}

func degradedWorkflowLogEntry(group CollaborationGroup, item groupWorkflowLogDegrade) groupWorkflowLogEntry {
	return groupWorkflowLogEntry{
		ID:            "group:" + group.ID + ":degraded:" + item.Kind,
		Kind:          "degraded",
		GroupID:       group.ID,
		Title:         item.Kind,
		Summary:       item.Message,
		State:         "degraded",
		TrustPosture:  "partial_timeline",
		Recovery:      compactWorkflowLogStrings([]string{item.NextAction}),
		NeedsOperator: true,
		Timestamp:     item.Timestamp,
		SourceKind:    string(protocol.SourceKindSystem),
		PayloadKind:   "degradation",
	}
}

func teamWorkTrustPosture(item protocol.TeamWorkItem) string {
	if item.NeedsOperator || item.State == protocol.TeamWorkStateDegraded {
		return "needs_operator_attention"
	}
	if item.State == protocol.TeamWorkStateOutputReady {
		return "output_ready"
	}
	if len(item.ProofRefs) > 0 || strings.TrimSpace(item.ProofID) != "" {
		return "proof_linked"
	}
	return "work_recorded"
}

func artifactTeamID(artifact artifacts.Artifact) string {
	if artifact.TeamID != nil && *artifact.TeamID != uuid.Nil {
		return artifact.TeamID.String()
	}
	return strings.TrimSpace(artifact.AgentID)
}

func artifactAuditRefs(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil
	}
	return compactWorkflowLogStrings(workflowLogAnyStringSlice(metadata["audit_refs"]))
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}
	return time.Now().UTC()
}

func compactWorkflowLogStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func workflowLogAnyStringSlice(raw any) []string {
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		return []string{v}
	default:
		return nil
	}
}

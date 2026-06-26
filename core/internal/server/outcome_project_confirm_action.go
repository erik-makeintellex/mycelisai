package server

import (
	"context"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) ensureOutcomeOwnershipForConfirmedAction(ctx context.Context, link confirmedActionTeamWorkLink, refs []confirmActionTeamWorkRef) (*protocol.OutcomeProject, error) {
	if s.getDB() == nil || len(refs) == 0 {
		return nil, nil
	}
	workRefs := make([]string, 0, len(refs))
	outputRefs := []protocol.TeamOutputRef{}
	teamIDs := []string{}
	status := protocol.OutcomeProjectStatusActive
	for _, ref := range refs {
		if strings.TrimSpace(ref.WorkItemID) != "" {
			workRefs = append(workRefs, ref.WorkItemID)
		}
		if strings.TrimSpace(ref.TeamID) != "" && !containsToolName(teamIDs, ref.TeamID) {
			teamIDs = append(teamIDs, ref.TeamID)
		}
		outputRefs = append(outputRefs, ref.OutputRefs...)
		if ref.State == protocol.TeamWorkStateOutputReady {
			status = protocol.OutcomeProjectStatusOutputReady
		}
		if ref.State == protocol.TeamWorkStateDegraded || ref.State == protocol.TeamWorkStateNeedsOperator {
			status = protocol.OutcomeProjectStatusNeedsAttention
		}
	}
	if len(workRefs) == 0 && len(outputRefs) == 0 {
		return nil, nil
	}
	project := protocol.NormalizeOutcomeProject(protocol.OutcomeProject{
		OutcomeID:       firstNonEmptyString(link.RunID, link.ProofID),
		Title:           outcomeProjectTitle(link, refs, outputRefs),
		Purpose:         outcomeProjectPurpose(link),
		ExecutionMode:   "project",
		WorkspaceFolder: outcomeProjectWorkspaceFolder(outputRefs, teamIDs),
		Status:          status,
		RunID:           link.RunID,
		IntentProofID:   link.ProofID,
		ContractID:      link.ContractID,
		ProofID:         link.ProofArtifactID,
		WorkItemRefs:    workRefs,
		OutputRefs:      outputRefs,
		ProofRefs:       outcomeProofRefs(link),
		RecoveryRefs:    outcomeRecoveryRefs(refs),
		RetentionPolicy: "retained",
	})
	if err := s.insertOutcomeProjectDB(ctx, &project); err != nil {
		return nil, err
	}
	for i, teamID := range teamIDs {
		entry := protocol.NormalizeTeamRegistryEntry(protocol.TeamRegistryEntry{
			ProjectID:        project.ProjectID,
			Role:             teamRegistryRole(i),
			TeamID:           teamID,
			AssignmentReason: "Assigned by confirmed Soma execution to produce or review retained outcome work.",
			Temporary:        true,
			Status:           "active",
		})
		if err := s.insertTeamRegistryEntryDB(ctx, &entry); err != nil {
			return nil, err
		}
		project.TeamRegistryRefs = append(project.TeamRegistryRefs, entry.RegistryID)
	}
	return &project, nil
}

func outcomeProjectTitle(link confirmedActionTeamWorkLink, refs []confirmActionTeamWorkRef, outputRefs []protocol.TeamOutputRef) string {
	for _, output := range outputRefs {
		if label := strings.TrimSpace(output.Label); label != "" {
			return label + " outcome"
		}
	}
	for _, ref := range refs {
		if teamID := strings.TrimSpace(ref.TeamID); teamID != "" {
			return teamID + " outcome workspace"
		}
	}
	return "Soma outcome workspace " + firstNonEmptyString(link.RunID, link.ProofID)
}

func outcomeProjectPurpose(link confirmedActionTeamWorkLink) string {
	if link.Scope == nil {
		return "Retain confirmed Soma work, team ownership, outputs, and proof for revisit."
	}
	if len(link.Scope.AffectedResources) > 0 {
		return "Retain confirmed Soma work across " + strings.Join(link.Scope.AffectedResources, ", ") + "."
	}
	return "Retain confirmed Soma work, team ownership, outputs, and proof for revisit."
}

func outcomeProjectWorkspaceFolder(outputRefs []protocol.TeamOutputRef, teamIDs []string) string {
	for _, output := range outputRefs {
		if strings.TrimSpace(output.Kind) == "team" {
			continue
		}
		if folder := parentWorkspacePath(output.StorageRef); strings.TrimSpace(folder) != "" {
			return folder
		}
		if folder := parentWorkspacePath(output.Entrypoint); strings.TrimSpace(folder) != "" {
			return folder
		}
	}
	if len(teamIDs) > 0 {
		return "groups/" + teamIDs[0]
	}
	return ""
}

func outcomeProofRefs(link confirmedActionTeamWorkLink) []string {
	refs := []string{}
	for _, ref := range []string{link.ProofArtifactID, link.ContractID, link.ProofID, link.AuditID} {
		if trimmed := strings.TrimSpace(ref); trimmed != "" && !containsToolName(refs, trimmed) {
			refs = append(refs, trimmed)
		}
	}
	return refs
}

func outcomeRecoveryRefs(refs []confirmActionTeamWorkRef) []string {
	recoveryRefs := []string{}
	for _, ref := range refs {
		if ref.State == protocol.TeamWorkStateDegraded || ref.State == protocol.TeamWorkStateNeedsOperator {
			recoveryRefs = append(recoveryRefs, ref.WorkItemID)
		}
	}
	return recoveryRefs
}

func teamRegistryRole(index int) string {
	if index == 0 {
		return "lead"
	}
	return "specialist"
}

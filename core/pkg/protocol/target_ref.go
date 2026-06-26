package protocol

import "strings"

// TargetRef identifies the product object an operator-facing alert, event, or
// receipt should open. It intentionally carries identity, not UI routes.
type TargetRef struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	RunID      string `json:"run_id,omitempty"`
	TeamID     string `json:"team_id,omitempty"`
	WorkItemID string `json:"work_item_id,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
	OutputID   string `json:"output_id,omitempty"`
	Label      string `json:"label,omitempty"`
}

func NormalizeTargetRef(raw *TargetRef) *TargetRef {
	if raw == nil {
		return nil
	}
	item := *raw
	item.Type = strings.TrimSpace(item.Type)
	item.ID = strings.TrimSpace(item.ID)
	item.RunID = strings.TrimSpace(item.RunID)
	item.TeamID = strings.TrimSpace(item.TeamID)
	item.WorkItemID = strings.TrimSpace(item.WorkItemID)
	item.ProjectID = strings.TrimSpace(item.ProjectID)
	item.OutputID = strings.TrimSpace(item.OutputID)
	item.Label = strings.TrimSpace(item.Label)
	if item.Type == "" || item.ID == "" {
		return nil
	}
	return &item
}

func TargetRefForTeamWork(item TeamWorkItem) *TargetRef {
	workItemID := strings.TrimSpace(item.WorkItemID)
	runID := strings.TrimSpace(item.RunID)
	teamID := strings.TrimSpace(item.TeamID)
	if runID != "" {
		return &TargetRef{
			Type:       "run",
			ID:         runID,
			RunID:      runID,
			TeamID:     teamID,
			WorkItemID: workItemID,
			Label:      "Run receipt",
		}
	}
	if workItemID == "" {
		return nil
	}
	targetType := "work"
	label := "Work item"
	if item.NeedsOperator || len(item.RecoveryOptions) > 0 || item.State == TeamWorkStateDegraded || item.State == TeamWorkStateNeedsOperator {
		targetType = "recovery"
		label = "Recovery item"
	}
	return &TargetRef{
		Type:       targetType,
		ID:         workItemID,
		TeamID:     teamID,
		WorkItemID: workItemID,
		Label:      label,
	}
}

func TargetRefForTeamStatusEvent(item TeamStatusEvent) *TargetRef {
	workItemID := strings.TrimSpace(item.WorkItemID)
	runID := strings.TrimSpace(item.RunID)
	teamID := strings.TrimSpace(item.TeamID)
	if runID != "" {
		return &TargetRef{
			Type:       "run",
			ID:         runID,
			RunID:      runID,
			TeamID:     teamID,
			WorkItemID: workItemID,
			Label:      "Run receipt",
		}
	}
	if workItemID == "" {
		return nil
	}
	targetType := "work"
	label := "Work item"
	if item.State == TeamWorkStateDegraded || item.State == TeamWorkStateNeedsOperator {
		targetType = "recovery"
		label = "Recovery item"
	}
	return &TargetRef{
		Type:       targetType,
		ID:         workItemID,
		TeamID:     teamID,
		WorkItemID: workItemID,
		Label:      label,
	}
}

func TargetRefForOutcomeProject(item OutcomeProject) *TargetRef {
	projectID := strings.TrimSpace(item.ProjectID)
	if projectID == "" {
		return nil
	}
	return &TargetRef{
		Type:      "outcome_project",
		ID:        projectID,
		RunID:     strings.TrimSpace(item.RunID),
		ProjectID: projectID,
		Label:     "Outcome project",
	}
}

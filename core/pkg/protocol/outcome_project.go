package protocol

import (
	"fmt"
	"strings"
	"time"
)

type OutcomeProjectStatus string

const (
	OutcomeProjectStatusActive         OutcomeProjectStatus = "active"
	OutcomeProjectStatusNeedsAttention OutcomeProjectStatus = "needs_attention"
	OutcomeProjectStatusOutputReady    OutcomeProjectStatus = "output_ready"
	OutcomeProjectStatusArchived       OutcomeProjectStatus = "archived"
)

// OutcomeProject is the durable user-facing ownership record for complex work.
type OutcomeProject struct {
	ProjectID        string               `json:"project_id"`
	OutcomeID        string               `json:"outcome_id"`
	Title            string               `json:"title"`
	Purpose          string               `json:"purpose,omitempty"`
	ExecutionMode    string               `json:"execution_mode"`
	WorkspaceFolder  string               `json:"workspace_folder,omitempty"`
	Status           OutcomeProjectStatus `json:"status"`
	RunID            string               `json:"run_id,omitempty"`
	IntentProofID    string               `json:"intent_proof_id,omitempty"`
	ContractID       string               `json:"contract_id,omitempty"`
	ProofID          string               `json:"proof_id,omitempty"`
	WorkItemRefs     []string             `json:"work_item_refs,omitempty"`
	OutputRefs       []TeamOutputRef      `json:"output_refs,omitempty"`
	ProofRefs        []string             `json:"proof_refs,omitempty"`
	RecoveryRefs     []string             `json:"recovery_refs,omitempty"`
	RetentionPolicy  string               `json:"retention_policy"`
	TeamRegistryRefs []string             `json:"team_registry_refs,omitempty"`
	TargetRef        *TargetRef           `json:"target_ref,omitempty"`
	CreatedAt        time.Time            `json:"created_at,omitempty"`
	UpdatedAt        time.Time            `json:"updated_at,omitempty"`
	Version          string               `json:"version"`
}

type TeamRegistryEntry struct {
	RegistryID       string     `json:"registry_id"`
	ProjectID        string     `json:"project_id"`
	GroupID          string     `json:"group_id,omitempty"`
	Role             string     `json:"role"`
	TeamID           string     `json:"team_id,omitempty"`
	AgentID          string     `json:"agent_id,omitempty"`
	AssignmentReason string     `json:"assignment_reason,omitempty"`
	Temporary        bool       `json:"temporary"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at,omitempty"`
	Version          string     `json:"version"`
}

func NormalizeOutcomeProject(raw OutcomeProject) OutcomeProject {
	item := raw
	item.ProjectID = strings.TrimSpace(item.ProjectID)
	item.OutcomeID = strings.TrimSpace(item.OutcomeID)
	item.Title = strings.TrimSpace(item.Title)
	item.Purpose = strings.TrimSpace(item.Purpose)
	item.ExecutionMode = strings.TrimSpace(item.ExecutionMode)
	if item.ExecutionMode == "" {
		item.ExecutionMode = "project"
	}
	item.WorkspaceFolder = strings.TrimSpace(item.WorkspaceFolder)
	if item.Status == "" {
		item.Status = OutcomeProjectStatusActive
	}
	item.RunID = strings.TrimSpace(item.RunID)
	item.IntentProofID = strings.TrimSpace(item.IntentProofID)
	item.ContractID = strings.TrimSpace(item.ContractID)
	item.ProofID = strings.TrimSpace(item.ProofID)
	item.WorkItemRefs = compactStrings(item.WorkItemRefs)
	item.ProofRefs = compactStrings(item.ProofRefs)
	item.RecoveryRefs = compactStrings(item.RecoveryRefs)
	item.TeamRegistryRefs = compactStrings(item.TeamRegistryRefs)
	item.RetentionPolicy = strings.TrimSpace(item.RetentionPolicy)
	item.TargetRef = NormalizeTargetRef(item.TargetRef)
	if item.RetentionPolicy == "" {
		item.RetentionPolicy = "retained"
	}
	if item.TargetRef == nil {
		item.TargetRef = TargetRefForOutcomeProject(item)
	}
	if item.Version == "" {
		item.Version = "v1"
	}
	return item
}

func ValidateOutcomeProject(item OutcomeProject) error {
	if strings.TrimSpace(item.OutcomeID) == "" {
		return fmt.Errorf("outcome_id is required")
	}
	if strings.TrimSpace(item.Title) == "" {
		return fmt.Errorf("title is required")
	}
	switch item.Status {
	case OutcomeProjectStatusActive, OutcomeProjectStatusNeedsAttention, OutcomeProjectStatusOutputReady, OutcomeProjectStatusArchived:
		return nil
	default:
		return fmt.Errorf("invalid outcome project status")
	}
}

func NormalizeTeamRegistryEntry(raw TeamRegistryEntry) TeamRegistryEntry {
	item := raw
	item.RegistryID = strings.TrimSpace(item.RegistryID)
	item.ProjectID = strings.TrimSpace(item.ProjectID)
	item.GroupID = strings.TrimSpace(item.GroupID)
	item.Role = strings.TrimSpace(item.Role)
	item.TeamID = strings.TrimSpace(item.TeamID)
	item.AgentID = strings.TrimSpace(item.AgentID)
	item.AssignmentReason = strings.TrimSpace(item.AssignmentReason)
	item.Status = strings.TrimSpace(item.Status)
	if item.Status == "" {
		item.Status = "active"
	}
	if item.Version == "" {
		item.Version = "v1"
	}
	return item
}

func ValidateTeamRegistryEntry(item TeamRegistryEntry) error {
	if strings.TrimSpace(item.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(item.Role) == "" {
		return fmt.Errorf("role is required")
	}
	if strings.TrimSpace(item.TeamID) == "" && strings.TrimSpace(item.AgentID) == "" {
		return fmt.Errorf("team_id or agent_id is required")
	}
	return nil
}

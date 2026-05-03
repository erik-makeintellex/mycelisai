package server

type TeamLeadExecutionMode string

const (
	TeamLeadExecutionModeGuidedReview             TeamLeadExecutionMode = "guided_review"
	TeamLeadExecutionModeNativeTeam               TeamLeadExecutionMode = "native_team"
	TeamLeadExecutionModeExternalWorkflowContract TeamLeadExecutionMode = "external_workflow_contract"
	TeamLeadExecutionModeContinuityResume         TeamLeadExecutionMode = "continuity_resume"
)

type TeamLeadExecutionContract struct {
	ExecutionMode              TeamLeadExecutionMode         `json:"execution_mode"`
	OwnerLabel                 string                        `json:"owner_label"`
	Summary                    string                        `json:"summary"`
	ContinuityLabel            string                        `json:"continuity_label,omitempty"`
	ContinuitySummary          string                        `json:"continuity_summary,omitempty"`
	ResumeCheckpoint           string                        `json:"resume_checkpoint,omitempty"`
	TeamName                   string                        `json:"team_name,omitempty"`
	ExternalTarget             string                        `json:"external_target,omitempty"`
	CoordinationModel          string                        `json:"coordination_model,omitempty"`
	RecommendedTeamShape       string                        `json:"recommended_team_shape,omitempty"`
	RecommendedTeamCount       int                           `json:"recommended_team_count,omitempty"`
	RecommendedTeamMemberLimit int                           `json:"recommended_team_member_limit,omitempty"`
	TargetOutputs              []string                      `json:"target_outputs"`
	Workstreams                []TeamLeadExecutionWorkstream `json:"workstreams,omitempty"`
	WorkflowGroup              *TeamLeadWorkflowGroupDraft   `json:"workflow_group,omitempty"`
}

type TeamLeadExecutionWorkstream struct {
	Label         string   `json:"label"`
	OwnerLabel    string   `json:"owner_label"`
	Status        string   `json:"status,omitempty"`
	Summary       string   `json:"summary"`
	NextStep      string   `json:"next_step"`
	TargetOutputs []string `json:"target_outputs,omitempty"`
}

type TeamLeadWorkflowGroupDraft struct {
	GroupID                string   `json:"group_id,omitempty"`
	Name                   string   `json:"name"`
	GoalStatement          string   `json:"goal_statement"`
	WorkMode               string   `json:"work_mode"`
	CoordinatorProfile     string   `json:"coordinator_profile"`
	AllowedCapabilities    []string `json:"allowed_capabilities,omitempty"`
	RecommendedMemberLimit int      `json:"recommended_member_limit,omitempty"`
	ExpiryHours            int      `json:"expiry_hours,omitempty"`
	Summary                string   `json:"summary"`
}

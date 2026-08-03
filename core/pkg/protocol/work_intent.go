package protocol

// WorkIntent describes how Soma expects approved work to behave.
// It is descriptive metadata for UI/proof/recovery, not execution authority.
type WorkIntent struct {
	Kind            string                 `json:"kind,omitempty"`
	Objective       string                 `json:"objective,omitempty"`
	Cadence         string                 `json:"cadence,omitempty"`
	ScheduleSummary string                 `json:"schedule_summary,omitempty"`
	RuntimePosture  string                 `json:"runtime_posture,omitempty"`
	TargetTeamID    string                 `json:"target_team_id,omitempty"`
	BusScope        string                 `json:"bus_scope,omitempty"`
	NATSSubjects    []string               `json:"nats_subjects,omitempty"`
	ServiceRefs     []string               `json:"service_refs,omitempty"`
	ProjectRef      string                 `json:"project_ref,omitempty"`
	OutputContract  *WorkOutputContract    `json:"output_contract,omitempty"`
	Lifecycle       *WorkLifecycleContract `json:"lifecycle,omitempty"`
}

// WorkOutputContract names the expected deliverable shape for approved work.
type WorkOutputContract struct {
	Shape              string                `json:"shape,omitempty"`
	PrimaryDeliverable string                `json:"primary_deliverable,omitempty"`
	Retention          string                `json:"retention,omitempty"`
	LaunchHint         string                `json:"launch_hint,omitempty"`
	Validation         []string              `json:"validation,omitempty"`
	OutputValidation   *OutputValidationPlan `json:"output_validation,omitempty"`
}

// WorkLifecycleContract describes operator controls for an approved work mode.
// It does not grant authority or bypass the governed work-action path.
type WorkLifecycleContract struct {
	StopAction     string `json:"stop_action,omitempty"`
	RetryAction    string `json:"retry_action,omitempty"`
	RecoveryAction string `json:"recovery_action,omitempty"`
	ControlSummary string `json:"control_summary,omitempty"`
}

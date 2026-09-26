package protocol

import "time"

// WorkRunningTitleMaxRunes bounds the operator-facing title of a running item.
const WorkRunningTitleMaxRunes = 160

// WorkRunningItemDetails carries raw runtime vocabulary for Inspect only.
// Primary UI copy must use OutcomeHealth, never these fields.
type WorkRunningItemDetails struct {
	State            TeamWorkState      `json:"state"`
	ExecutionShape   TeamExecutionShape `json:"execution_shape,omitempty"`
	DegradationState string             `json:"degradation_state,omitempty"`
}

// WorkRunningItem is one nonterminal durable team-work item. It deliberately
// omits prompts, work_intent bodies, last_event payloads, outputs and proofs.
type WorkRunningItem struct {
	WorkItemID    string                 `json:"work_item_id"`
	TeamID        string                 `json:"team_id"`
	RunID         string                 `json:"run_id,omitempty"`
	ProjectID     string                 `json:"project_id,omitempty"`
	Title         string                 `json:"title"`
	OutcomeHealth OutcomeHealthState     `json:"outcome_health"`
	UpdatedAt     time.Time              `json:"updated_at"`
	NeedsOperator bool                   `json:"needs_operator"`
	Details       WorkRunningItemDetails `json:"details"`
}

// WorkRunningOutcome groups running items under their owning OutcomeProject.
type WorkRunningOutcome struct {
	ProjectID     string             `json:"project_id"`
	OutcomeID     string             `json:"outcome_id"`
	Title         string             `json:"title"`
	OutcomeHealth OutcomeHealthState `json:"outcome_health"`
	Items         []WorkRunningItem  `json:"items"`
}

// WorkRunningTeamGroup holds running items that no OutcomeProject references.
type WorkRunningTeamGroup struct {
	TeamID string            `json:"team_id"`
	Items  []WorkRunningItem `json:"items"`
}

// WorkRunningSummary counts items by Outcome Health only.
type WorkRunningSummary struct {
	Running  int `json:"running"`
	Waiting  int `json:"waiting"`
	Blocked  int `json:"blocked"`
	Degraded int `json:"degraded"`
}

// WorkRunningView is the read-only cross-team "what's running" projection.
type WorkRunningView struct {
	GeneratedAt time.Time              `json:"generated_at"`
	Truncated   bool                   `json:"truncated"`
	Summary     WorkRunningSummary     `json:"summary"`
	Outcomes    []WorkRunningOutcome   `json:"outcomes"`
	Unassigned  []WorkRunningTeamGroup `json:"unassigned"`
}

// WorkRunningFailure is the data body of a 503: it names the failed source and
// a recovery hint, and never carries item data.
type WorkRunningFailure struct {
	Source       string `json:"source"`
	RecoveryHint string `json:"recovery_hint"`
}

// IsWorkRunningVisible reports whether a team-work item may appear in the
// running view. Terminal work (archived, output_ready) and anything whose
// Outcome Health is Completed or Archived never appears.
func IsWorkRunningVisible(item TeamWorkItem) bool {
	switch item.State {
	case TeamWorkStateArchived, TeamWorkStateOutputReady:
		return false
	}
	switch OutcomeHealthForTeamWork(item) {
	case OutcomeHealthCompleted, OutcomeHealthArchived:
		return false
	}
	return true
}

package protocol

// ContinuationIntentKind describes what the user appears to want from a
// previous delivered output. It does not grant execution authority.
type ContinuationIntentKind string

const (
	ContinuationIntentFollowUp ContinuationIntentKind = "follow_up"
	ContinuationIntentUpdate   ContinuationIntentKind = "update"
	ContinuationIntentFork     ContinuationIntentKind = "fork"
	ContinuationIntentRoute    ContinuationIntentKind = "route"
	ContinuationIntentInspect  ContinuationIntentKind = "inspect"
)

type ChatContinuationIntent struct {
	Kind             ContinuationIntentKind `json:"kind"`
	ContextKind      string                 `json:"context_kind,omitempty"`
	TargetTitle      string                 `json:"target_title,omitempty"`
	Reference        string                 `json:"reference,omitempty"`
	Proof            string                 `json:"proof,omitempty"`
	TeamID           string                 `json:"team_id,omitempty"`
	RunID            string                 `json:"run_id,omitempty"`
	WorkItemID       string                 `json:"work_item_id,omitempty"`
	OutputID         string                 `json:"output_id,omitempty"`
	SourceDigest     string                 `json:"source_digest,omitempty"`
	SourceVersion    string                 `json:"source_version,omitempty"`
	RevisionTarget   string                 `json:"revision_target,omitempty"`
	RequiresProposal bool                   `json:"requires_proposal"`
	Reason           string                 `json:"reason,omitempty"`
}

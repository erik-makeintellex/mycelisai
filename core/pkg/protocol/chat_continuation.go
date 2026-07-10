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
	RequiresProposal bool                   `json:"requires_proposal"`
	Reason           string                 `json:"reason,omitempty"`
}

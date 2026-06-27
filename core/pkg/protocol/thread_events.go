package protocol

type ThreadEventKind string

const PayloadKindThreadEvent SignalPayloadKind = "thread_event"

const (
	ThreadEventExecutionStarted ThreadEventKind = "execution_started"
	ThreadEventExecutionUpdate  ThreadEventKind = "execution_update"
	ThreadEventResultReady      ThreadEventKind = "result_ready"
	ThreadEventAttentionNeeded  ThreadEventKind = "attention_required"
)

type ThreadEventPayload struct {
	Kind            ThreadEventKind `json:"kind"`
	Label           string          `json:"label"`
	Detail          string          `json:"detail,omitempty"`
	Tone            string          `json:"tone"`
	Status          string          `json:"status,omitempty"`
	Href            string          `json:"href,omitempty"`
	HrefLabel       string          `json:"href_label,omitempty"`
	TargetReference string          `json:"target_reference,omitempty"`
	WorkItemID      string          `json:"work_item_id,omitempty"`
	IntentProofID   string          `json:"intent_proof_id,omitempty"`
	ContractID      string          `json:"contract_id,omitempty"`
	ProofID         string          `json:"proof_id,omitempty"`
	OutputRefs      []TeamOutputRef `json:"output_refs,omitempty"`
}

type ThreadEventEnvelope struct {
	Type       string             `json:"type"`
	EventType  EventType          `json:"event_type"`
	ThreadID   string             `json:"thread_id"`
	ThreadKind string             `json:"thread_kind"`
	EventID    string             `json:"event_id,omitempty"`
	Version    string             `json:"version"`
	Meta       SignalMeta         `json:"meta"`
	Payload    ThreadEventPayload `json:"payload"`
}

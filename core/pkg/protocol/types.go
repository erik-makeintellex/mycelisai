package protocol

import "time"

// EnvelopeType defines the category of the message.
type EnvelopeType string

const (
	TypeThought    EnvelopeType = "thought"
	TypeMetric     EnvelopeType = "metric"
	TypeArtifact   EnvelopeType = "artifact"
	TypeGovernance EnvelopeType = "governance"
)

// Envelope is the standard container for all UI signals.
type Envelope struct {
	Type      EnvelopeType `json:"type"`
	Source    string       `json:"source"`
	Timestamp time.Time    `json:"timestamp"`
	Content   interface{}  `json:"content"`
}

// ThoughtContent represents LLM reasoning.
type ThoughtContent struct {
	Summary string `json:"summary"`
	Detail  string `json:"detail"`
	Model   string `json:"model"`
}

// MetricContent represents system telemetry.
type MetricContent struct {
	Label  string      `json:"label"`
	Value  interface{} `json:"value"` // number or string
	Unit   string      `json:"unit"`
	Status string      `json:"status"` // nominal | warning | critical
}

// ArtifactContent represents a deliverable file.
type ArtifactContent struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Title    string `json:"title"`
	Preview  string `json:"preview"`
	URI      string `json:"uri"`
}

// GovernanceContent represents a human-in-the-loop request.
type GovernanceContent struct {
	RequestID   string `json:"request_id"`
	AgentID     string `json:"agent_id"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Status      string `json:"status"` // pending | approved | denied
}

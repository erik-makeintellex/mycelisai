package protocol

import (
	"encoding/json"
	"fmt"
	"time"
)

// SignalType classifies a CTS signal flowing through the nervous system.
type SignalType string

const (
	SignalTelemetry      SignalType = "telemetry"
	SignalTaskComplete   SignalType = "task_complete"
	SignalTaskFailed     SignalType = "task_failed"
	SignalError          SignalType = "error"
	SignalHeartbeat      SignalType = "heartbeat"
	SignalGovernanceHalt SignalType = "governance_halt"
	SignalSensorData     SignalType = "sensor_data"
	SignalChatResponse   SignalType = "chat_response"
)

// Trust Economy — default trust scores by node category.
// Sensory/hardware nodes are fully trusted (they report facts).
// Cognitive/LLM nodes must earn trust (they hallucinate).
const (
	TrustScoreSensory   = 1.0 // Hardware/ingress — fully trusted
	TrustScoreCognitive = 0.5 // LLMs — trust must be earned
	TrustScoreActuation = 0.8 // Execution nodes — moderate trust
	TrustScoreLedger    = 1.0 // Memory/storage — fully trusted
)

// CTSMeta contains provenance metadata for every CTS signal.
// Every envelope MUST carry a non-empty SourceNode and a non-zero Timestamp.
type CTSMeta struct {
	SourceNode string    `json:"source_node"`
	Timestamp  time.Time `json:"timestamp"`
	TraceID    string    `json:"trace_id,omitempty"`
}

// CTSEnvelope is the mandatory wrapper for all sensor/agent outputs.
// Any message on swarm.team.*.telemetry MUST conform to this schema.
// TrustScore (0.0–1.0) drives the Governance Valve: envelopes above the
// AutoExecuteThreshold bypass human approval; below it, they halt for Zone D review.
type CTSEnvelope struct {
	Meta       CTSMeta         `json:"meta"`
	SignalType SignalType      `json:"signal_type"`
	TrustScore float64         `json:"trust_score,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

// HasTrustScore returns true if the envelope carries an explicit trust rating.
// Zero-value (0.0) is treated as "unscored" — the Governance Valve should
// apply the node-category default rather than blocking.
func (e *CTSEnvelope) HasTrustScore() bool {
	return e.TrustScore > 0
}

// Validate enforces the Zero-Trust schema contract.
func (e *CTSEnvelope) Validate() error {
	if e.Meta.SourceNode == "" {
		return fmt.Errorf("cts: meta.source_node is required")
	}
	if e.Meta.Timestamp.IsZero() {
		return fmt.Errorf("cts: meta.timestamp is required")
	}
	if e.SignalType == "" {
		return fmt.Errorf("cts: signal_type is required")
	}
	if len(e.Payload) == 0 {
		return fmt.Errorf("cts: payload is required")
	}
	return nil
}

// ChatResponsePayload is the CTS Payload for council chat responses.
// Any endpoint returning LLM-generated content wraps it in this struct
// inside a CTSEnvelope.
type ChatResponsePayload struct {
	Text          string   `json:"text"`
	Consultations []string `json:"consultations,omitempty"`
	ToolsUsed     []string `json:"tools_used,omitempty"`
}

// APIResponse is the standard response wrapper for all Mycelis API endpoints.
// New endpoints SHOULD return this envelope. Existing endpoints will be
// migrated incrementally.
type APIResponse struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// NewAPISuccess creates a successful response envelope.
func NewAPISuccess(data interface{}) APIResponse {
	return APIResponse{OK: true, Data: data}
}

// NewAPIError creates an error response envelope.
func NewAPIError(msg string) APIResponse {
	return APIResponse{OK: false, Error: msg}
}

// ValidateTelemetryMessage is NATS middleware that rejects any message on
// swarm.team.*.telemetry that fails CTS schema validation.
// It returns the parsed envelope on success, or an error explaining the rejection.
func ValidateTelemetryMessage(data []byte) (*CTSEnvelope, error) {
	var env CTSEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("cts: malformed envelope: %w", err)
	}
	if err := env.Validate(); err != nil {
		return nil, err
	}
	return &env, nil
}

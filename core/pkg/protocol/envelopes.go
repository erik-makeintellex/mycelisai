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
	SignalChatResponse      SignalType = "chat_response"
	SignalBlueprintProposal SignalType = "blueprint_proposal" // CE-1: proposal with confirm token
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
	// CE-1: Orchestration template metadata (backward-compatible, omitempty)
	TemplateID TemplateID    `json:"template_id,omitempty"`
	Mode       ExecutionMode `json:"mode,omitempty"`
	// V7 Event Spine: links this CTS signal to the persistent audit record in mission_events.
	// Set by events.Store.publishCTS. Consumers use this ID to fetch full event detail.
	MissionEventID string `json:"mission_event_id,omitempty"`
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

// BrainProvenance carries metadata about which provider/model executed a request.
// Phase 19: Every LLM response must be traceable to its source brain.
type BrainProvenance struct {
	ProviderID   string `json:"provider_id"`              // "ollama", "claude"
	ProviderName string `json:"provider_name,omitempty"`  // display name
	ModelID      string `json:"model_id"`                 // "qwen2.5-coder:7b"
	Location     string `json:"location"`                 // "local" | "remote"
	DataBoundary string `json:"data_boundary"`            // "local_only" | "leaves_org"
	TokensUsed   int    `json:"tokens_used,omitempty"`
}

// ChatProposal carries proposal metadata for chat-based mutation actions.
// Phase 19-B: When an agent uses mutation tools, the response is tagged as a proposal.
type ChatProposal struct {
	Intent        string   `json:"intent"`
	Tools         []string `json:"tools"`
	RiskLevel     string   `json:"risk_level"`      // "low" | "medium" | "high"
	ConfirmToken  string   `json:"confirm_token"`
	IntentProofID string   `json:"intent_proof_id"`
}

// ConsultationEntry records a single council member consultation made by an agent
// during its ReAct loop. Member is the agent ID (e.g. "council-architect"),
// Summary is the first 300 chars of the member's response.
type ConsultationEntry struct {
	Member  string `json:"member"`
	Summary string `json:"summary"`
}

// ChatResponsePayload is the CTS Payload for council chat responses.
// Any endpoint returning LLM-generated content wraps it in this struct
// inside a CTSEnvelope.
type ChatResponsePayload struct {
	Text          string              `json:"text"`
	Consultations []ConsultationEntry `json:"consultations,omitempty"`
	ToolsUsed     []string            `json:"tools_used,omitempty"`
	Artifacts     []ChatArtifactRef   `json:"artifacts,omitempty"`
	// CE-1: Answer provenance (audit linkage for Chat-to-Answer template)
	Provenance    *AnswerProvenance   `json:"provenance,omitempty"`
	// Phase 19: Brain provenance (which provider/model executed this response)
	Brain         *BrainProvenance    `json:"brain,omitempty"`
	// Phase 19-B: Chat proposal (when agent uses mutation tools)
	Proposal      *ChatProposal       `json:"proposal,omitempty"`
}

// ChatArtifactRef is an inline artifact reference embedded in a chat response.
// For small content (code snippets, chart specs, short documents) the Content
// field carries the data directly. For large/binary content, ID references an
// artifact in the artifacts table that can be fetched separately.
type ChatArtifactRef struct {
	ID          string `json:"id,omitempty"`           // artifact table ID (for stored artifacts)
	Type        string `json:"type"`                   // code | document | image | audio | data | chart | file
	Title       string `json:"title"`
	ContentType string `json:"content_type,omitempty"` // MIME type
	Content     string `json:"content,omitempty"`      // inline content (text, JSON, base64 for images)
	URL         string `json:"url,omitempty"`          // external URL (for links, images)
}

// DelegationHint carries optional scoring metadata for task delegation.
// V1: logged for observability only. Future: somatic modulation.
type DelegationHint struct {
	Confidence float64 `json:"confidence,omitempty"` // 0.0-1.0
	Urgency    string  `json:"urgency,omitempty"`    // low, medium, high, critical
	Complexity int     `json:"complexity,omitempty"` // 1-5
	Risk       string  `json:"risk,omitempty"`       // low, medium, high
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

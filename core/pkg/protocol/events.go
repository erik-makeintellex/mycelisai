package protocol

import (
	"context"
	"time"
)

// EventType classifies a mission event in the persistent audit trail.
// All 17 event types are defined here as constants to prevent typos.
type EventType string

const (
	// Mission lifecycle
	EventMissionStarted   EventType = "mission.started"
	EventMissionCompleted EventType = "mission.completed"
	EventMissionFailed    EventType = "mission.failed"
	EventMissionCancelled EventType = "mission.cancelled"

	// Team lifecycle
	EventTeamSpawned EventType = "team.spawned"
	EventTeamStopped EventType = "team.stopped"

	// Agent lifecycle
	EventAgentStarted EventType = "agent.started"
	EventAgentStopped EventType = "agent.stopped"

	// Tool execution (ReAct loop)
	EventToolInvoked   EventType = "tool.invoked"
	EventToolCompleted EventType = "tool.completed"
	EventToolFailed    EventType = "tool.failed"

	// Artifact lifecycle
	EventArtifactCreated EventType = "artifact.created"

	// Memory (MCP memory server)
	EventMemoryStored   EventType = "memory.stored"
	EventMemoryRecalled EventType = "memory.recalled"

	// Automation (Team B/C — registered here for completeness)
	EventTriggerFired   EventType = "trigger.fired"
	EventTriggerSkipped EventType = "trigger.skipped"
	EventSchedulerTick  EventType = "scheduler.tick"
)

// EventSeverity classifies the operational significance of an event.
type EventSeverity string

const (
	SeverityDebug EventSeverity = "debug"
	SeverityInfo  EventSeverity = "info"
	SeverityWarn  EventSeverity = "warn"
	SeverityError EventSeverity = "error"
)

// MissionEventEnvelope is the authoritative audit record for a mission execution event.
// Every action in a mission run MUST produce one of these records (V7 Event Spine rule).
// This is the PERSISTENT record. The CTS envelope is the real-time transport signal.
// CTS payload includes mission_event_id to link signals back to this record.
type MissionEventEnvelope struct {
	ID           string                 `json:"id"`
	RunID        string                 `json:"run_id"`
	TenantID     string                 `json:"tenant_id"`
	EventType    EventType              `json:"event_type"`
	Severity     EventSeverity          `json:"severity"`
	SourceAgent  string                 `json:"source_agent,omitempty"`
	SourceTeam   string                 `json:"source_team,omitempty"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
	AuditEventID string                 `json:"audit_event_id,omitempty"`
	EmittedAt    time.Time              `json:"emitted_at"`
}

// EventEmitter allows optional event emission without creating import cycles.
// The interface lives in protocol because swarm and server both import protocol.
// Implemented by internal/events.Store; nil receiver = silent mode (no panic).
type EventEmitter interface {
	Emit(ctx context.Context, runID string, eventType EventType, severity EventSeverity,
		sourceAgent, sourceTeam string, payload map[string]interface{}) (string, error)
}

// RunsManager allows optional run tracking without creating import cycles.
// The interface lives in protocol because swarm imports protocol.
// Implemented by internal/runs.Manager; nil receiver = no run tracking.
type RunsManager interface {
	CreateRun(ctx context.Context, missionID string) (string, error)
	UpdateRunStatus(ctx context.Context, runID string, status string) error
}

// ConversationTurnData holds all fields for a single conversation turn.
// Used by ConversationLogger.LogTurn to persist full-fidelity agent transcripts.
type ConversationTurnData struct {
	RunID          string                 `json:"run_id,omitempty"`     // nullable: standing-team chats
	SessionID      string                 `json:"session_id"`
	TenantID       string                 `json:"tenant_id"`
	AgentID        string                 `json:"agent_id"`
	TeamID         string                 `json:"team_id,omitempty"`
	TurnIndex      int                    `json:"turn_index"`
	Role           string                 `json:"role"`                 // system|user|assistant|tool_call|tool_result|interjection
	Content        string                 `json:"content"`
	ProviderID     string                 `json:"provider_id,omitempty"`
	ModelUsed      string                 `json:"model_used,omitempty"`
	ToolName       string                 `json:"tool_name,omitempty"`
	ToolArgs       map[string]interface{} `json:"tool_args,omitempty"`
	ParentTurnID   string                 `json:"parent_turn_id,omitempty"` // links tool_result → tool_call
	ConsultationOf string                 `json:"consultation_of,omitempty"`
}

// ConversationLogger allows optional conversation turn persistence.
// Analogous to EventEmitter — agents receive it via SetConversationLogger().
// nil = silent mode (no panic, no recording).
// Implemented by internal/conversations.Store.
type ConversationLogger interface {
	LogTurn(ctx context.Context, turn ConversationTurnData) (string, error)
}

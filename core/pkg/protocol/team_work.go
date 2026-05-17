package protocol

import (
	"fmt"
	"strings"
	"time"
)

type TeamWorkState string

const (
	TeamWorkStateNew           TeamWorkState = "new"
	TeamWorkStateBriefed       TeamWorkState = "briefed"
	TeamWorkStateQueued        TeamWorkState = "queued"
	TeamWorkStateRunning       TeamWorkState = "running"
	TeamWorkStateNeedsOperator TeamWorkState = "needs_operator"
	TeamWorkStateReviewing     TeamWorkState = "reviewing"
	TeamWorkStateOutputReady   TeamWorkState = "output_ready"
	TeamWorkStateDegraded      TeamWorkState = "degraded"
	TeamWorkStatePaused        TeamWorkState = "paused"
	TeamWorkStateArchived      TeamWorkState = "archived"
)

type TeamExecutionShape string

const (
	TeamExecutionShapeCreateTeam    TeamExecutionShape = "create_team"
	TeamExecutionShapeDelegatedWork TeamExecutionShape = "delegated_work"
	TeamExecutionShapeDeliverable   TeamExecutionShape = "deliverable"
)

// TeamOutputRef links team work to retained product objects.
type TeamOutputRef struct {
	OutputID      string    `json:"output_id"`
	TeamID        string    `json:"team_id"`
	WorkItemID    string    `json:"work_item_id"`
	RunID         string    `json:"run_id,omitempty"`
	Kind          string    `json:"kind"`
	Label         string    `json:"label"`
	StorageRef    string    `json:"storage_ref,omitempty"`
	Entrypoint    string    `json:"entrypoint,omitempty"`
	ValidationRef string    `json:"validation_ref,omitempty"`
	ProofRef      string    `json:"proof_ref,omitempty"`
	ContractID    string    `json:"contract_id,omitempty"`
	ProofID       string    `json:"proof_id,omitempty"`
	AuditRefs     []string  `json:"audit_refs,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

// TeamStatusEvent is the normalized operator-readable projection of team communication.
type TeamStatusEvent struct {
	EventID           string        `json:"event_id"`
	TeamID            string        `json:"team_id"`
	WorkItemID        string        `json:"work_item_id"`
	RunID             string        `json:"run_id,omitempty"`
	IntentProofID     string        `json:"intent_proof_id,omitempty"`
	ContractID        string        `json:"contract_id,omitempty"`
	ProofID           string        `json:"proof_id,omitempty"`
	State             TeamWorkState `json:"state"`
	Headline          string        `json:"headline"`
	Details           string        `json:"details,omitempty"`
	ConfidencePosture string        `json:"confidence_posture,omitempty"`
	BlockedBy         []string      `json:"blocked_by,omitempty"`
	NextAction        string        `json:"next_action,omitempty"`
	SourceKind        string        `json:"source_kind,omitempty"`
	SourceChannel     string        `json:"source_channel,omitempty"`
	PayloadKind       string        `json:"payload_kind,omitempty"`
	AuditRefs         []string      `json:"audit_refs,omitempty"`
	Timestamp         time.Time     `json:"timestamp,omitempty"`
	Version           string        `json:"version"`
}

// TeamInteraction is the durable record of a Soma, Council, operator, or team-lead exchange.
type TeamInteraction struct {
	InteractionID string         `json:"interaction_id"`
	TeamID        string         `json:"team_id"`
	WorkItemID    string         `json:"work_item_id"`
	RunID         string         `json:"run_id,omitempty"`
	IntentProofID string         `json:"intent_proof_id,omitempty"`
	ContractID    string         `json:"contract_id,omitempty"`
	ProofID       string         `json:"proof_id,omitempty"`
	SourceKind    string         `json:"source_kind"`
	SourceChannel string         `json:"source_channel"`
	ActorRef      string         `json:"actor_ref"`
	Verb          string         `json:"verb"`
	Summary       string         `json:"summary"`
	PayloadKind   string         `json:"payload_kind"`
	PayloadRef    string         `json:"payload_ref,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	ApprovalRef   string         `json:"approval_ref,omitempty"`
	AuditRefs     []string       `json:"audit_refs,omitempty"`
	Timestamp     time.Time      `json:"timestamp,omitempty"`
	Version       string         `json:"version"`
}

// TeamWorkItem is the unit of active team execution.
type TeamWorkItem struct {
	WorkItemID             string             `json:"work_item_id"`
	TeamID                 string             `json:"team_id"`
	RunID                  string             `json:"run_id,omitempty"`
	IntentProofID          string             `json:"intent_proof_id,omitempty"`
	ContractID             string             `json:"contract_id,omitempty"`
	ProofID                string             `json:"proof_id,omitempty"`
	Objective              string             `json:"objective"`
	Scope                  []string           `json:"scope,omitempty"`
	Owner                  string             `json:"owner,omitempty"`
	ExecutionShape         TeamExecutionShape `json:"execution_shape"`
	ExpectedOutputs        []string           `json:"expected_outputs,omitempty"`
	ExpectedProof          []string           `json:"expected_proof,omitempty"`
	CapabilityRequirements []string           `json:"capability_requirements,omitempty"`
	GovernancePosture      ApprovalPosture    `json:"governance_posture,omitempty"`
	State                  TeamWorkState      `json:"state"`
	LastEvent              *TeamStatusEvent   `json:"last_event,omitempty"`
	NeedsOperator          bool               `json:"needs_operator"`
	DegradationState       string             `json:"degradation_state,omitempty"`
	RecoveryOptions        []string           `json:"recovery_options,omitempty"`
	OutputRefs             []TeamOutputRef    `json:"output_refs,omitempty"`
	ProofRefs              []string           `json:"proof_refs,omitempty"`
	AuditRefs              []string           `json:"audit_refs,omitempty"`
	CreatedAt              time.Time          `json:"created_at,omitempty"`
	UpdatedAt              time.Time          `json:"updated_at,omitempty"`
	Version                string             `json:"version"`
}

func NormalizeTeamWorkItem(raw TeamWorkItem) TeamWorkItem {
	item := raw
	item.WorkItemID = strings.TrimSpace(item.WorkItemID)
	item.TeamID = strings.TrimSpace(item.TeamID)
	item.RunID = strings.TrimSpace(item.RunID)
	item.IntentProofID = strings.TrimSpace(item.IntentProofID)
	item.ContractID = strings.TrimSpace(item.ContractID)
	item.ProofID = strings.TrimSpace(item.ProofID)
	item.Objective = strings.TrimSpace(item.Objective)
	item.Owner = strings.TrimSpace(item.Owner)
	item.ExecutionShape = normalizeTeamExecutionShape(item.ExecutionShape)
	item.Scope = compactStrings(item.Scope)
	item.ExpectedOutputs = compactStrings(item.ExpectedOutputs)
	item.ExpectedProof = compactStrings(item.ExpectedProof)
	item.CapabilityRequirements = compactStrings(item.CapabilityRequirements)
	item.GovernancePosture = normalizeApprovalPosture(item.GovernancePosture)
	item.DegradationState = strings.TrimSpace(item.DegradationState)
	item.RecoveryOptions = compactStrings(item.RecoveryOptions)
	item.ProofRefs = compactStrings(item.ProofRefs)
	item.AuditRefs = compactStrings(item.AuditRefs)
	if item.State == "" {
		item.State = defaultTeamWorkState(item.ExecutionShape)
	}
	if item.Version == "" {
		item.Version = "v1"
	}
	return item
}

func ValidateTeamWorkItem(item TeamWorkItem) error {
	if strings.TrimSpace(item.TeamID) == "" {
		return fmt.Errorf("team_id is required")
	}
	if strings.TrimSpace(item.Objective) == "" {
		return fmt.Errorf("objective is required")
	}
	switch item.ExecutionShape {
	case TeamExecutionShapeCreateTeam:
		if !IsTeamCreationState(item.State) {
			return fmt.Errorf("create_team work must be new or briefed")
		}
	case TeamExecutionShapeDelegatedWork, TeamExecutionShapeDeliverable:
		if !IsTeamExecutionState(item.State) {
			return fmt.Errorf("delegated or deliverable work must be queued, running, output_ready, or degraded")
		}
	default:
		return fmt.Errorf("invalid execution_shape")
	}
	return nil
}

func NormalizeTeamInteraction(raw TeamInteraction) TeamInteraction {
	item := raw
	item.InteractionID = strings.TrimSpace(item.InteractionID)
	item.TeamID = strings.TrimSpace(item.TeamID)
	item.WorkItemID = strings.TrimSpace(item.WorkItemID)
	item.RunID = strings.TrimSpace(item.RunID)
	item.IntentProofID = strings.TrimSpace(item.IntentProofID)
	item.ContractID = strings.TrimSpace(item.ContractID)
	item.ProofID = strings.TrimSpace(item.ProofID)
	item.SourceKind = strings.TrimSpace(item.SourceKind)
	item.SourceChannel = strings.TrimSpace(item.SourceChannel)
	item.ActorRef = strings.TrimSpace(item.ActorRef)
	item.Verb = strings.TrimSpace(item.Verb)
	item.Summary = strings.TrimSpace(item.Summary)
	item.PayloadKind = strings.TrimSpace(item.PayloadKind)
	item.PayloadRef = strings.TrimSpace(item.PayloadRef)
	item.ApprovalRef = strings.TrimSpace(item.ApprovalRef)
	item.AuditRefs = compactStrings(item.AuditRefs)
	if item.Version == "" {
		item.Version = "v1"
	}
	return item
}

func ValidateTeamInteraction(item TeamInteraction) error {
	if strings.TrimSpace(item.TeamID) == "" {
		return fmt.Errorf("team_id is required")
	}
	if strings.TrimSpace(item.WorkItemID) == "" {
		return fmt.Errorf("work_item_id is required")
	}
	if strings.TrimSpace(item.Verb) == "" {
		return fmt.Errorf("verb is required")
	}
	if strings.TrimSpace(item.Summary) == "" {
		return fmt.Errorf("summary is required")
	}
	if strings.TrimSpace(item.SourceKind) == "" {
		return fmt.Errorf("source_kind is required")
	}
	if strings.TrimSpace(item.SourceChannel) == "" {
		return fmt.Errorf("source_channel is required")
	}
	if strings.TrimSpace(item.PayloadKind) == "" {
		return fmt.Errorf("payload_kind is required")
	}
	return nil
}

func IsTeamCreationState(state TeamWorkState) bool {
	return state == TeamWorkStateNew || state == TeamWorkStateBriefed
}

func IsTeamExecutionState(state TeamWorkState) bool {
	return state == TeamWorkStateQueued ||
		state == TeamWorkStateRunning ||
		state == TeamWorkStateNeedsOperator ||
		state == TeamWorkStateReviewing ||
		state == TeamWorkStateOutputReady ||
		state == TeamWorkStateDegraded ||
		state == TeamWorkStatePaused ||
		state == TeamWorkStateArchived
}

func normalizeTeamExecutionShape(shape TeamExecutionShape) TeamExecutionShape {
	switch shape {
	case TeamExecutionShapeCreateTeam, TeamExecutionShapeDelegatedWork, TeamExecutionShapeDeliverable:
		return shape
	case "":
		return TeamExecutionShapeDelegatedWork
	default:
		return shape
	}
}

func defaultTeamWorkState(shape TeamExecutionShape) TeamWorkState {
	if shape == TeamExecutionShapeCreateTeam {
		return TeamWorkStateNew
	}
	return TeamWorkStateQueued
}

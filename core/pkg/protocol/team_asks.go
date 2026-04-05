package protocol

import "strings"

type TeamAskKind string

const (
	TeamAskKindResearch       TeamAskKind = "research"
	TeamAskKindImplementation TeamAskKind = "implementation"
	TeamAskKindValidation     TeamAskKind = "validation"
	TeamAskKindCoordination   TeamAskKind = "coordination"
	TeamAskKindReview         TeamAskKind = "review"
)

type TeamLaneRole string

const (
	TeamLaneRoleCoordinator TeamLaneRole = "coordinator"
	TeamLaneRoleResearcher  TeamLaneRole = "researcher"
	TeamLaneRoleImplementer TeamLaneRole = "implementer"
	TeamLaneRoleValidator   TeamLaneRole = "validator"
	TeamLaneRoleReviewer    TeamLaneRole = "reviewer"
)

// TeamAsk is the canonical structured ask contract for team command-lane
// delegation. The first slice keeps it narrow so delegate_task can emit one
// stable payload shape before broader team-instantiation logic depends on it.
type TeamAsk struct {
	SchemaVersion        string          `json:"schema_version,omitempty"`
	AskKind              TeamAskKind     `json:"ask_kind,omitempty"`
	LaneRole             TeamLaneRole    `json:"lane_role,omitempty"`
	Goal                 string          `json:"goal,omitempty"`
	Operation            string          `json:"operation,omitempty"`
	Intent               string          `json:"intent,omitempty"`
	Message              string          `json:"message,omitempty"`
	OwnedScope           []string        `json:"owned_scope,omitempty"`
	Constraints          []string        `json:"constraints,omitempty"`
	RequiredCapabilities []string        `json:"required_capabilities,omitempty"`
	ApprovalPosture      ApprovalPosture `json:"approval_posture,omitempty"`
	ExitCriteria         []string        `json:"exit_criteria,omitempty"`
	EvidenceRequired     []string        `json:"evidence_required,omitempty"`
	Context              map[string]any  `json:"context,omitempty"`
}

func NormalizeTeamAsk(raw TeamAsk) TeamAsk {
	normalized := TeamAsk{
		SchemaVersion:        "v1",
		AskKind:              normalizeTeamAskKind(raw.AskKind),
		LaneRole:             normalizeTeamLaneRole(raw.LaneRole),
		Goal:                 strings.TrimSpace(raw.Goal),
		Operation:            strings.TrimSpace(raw.Operation),
		Intent:               strings.TrimSpace(raw.Intent),
		Message:              strings.TrimSpace(raw.Message),
		OwnedScope:           compactStrings(raw.OwnedScope),
		Constraints:          compactStrings(raw.Constraints),
		RequiredCapabilities: compactStrings(raw.RequiredCapabilities),
		ApprovalPosture:      normalizeApprovalPosture(raw.ApprovalPosture),
		ExitCriteria:         compactStrings(raw.ExitCriteria),
		EvidenceRequired:     compactStrings(raw.EvidenceRequired),
		Context:              raw.Context,
	}
	if normalized.Goal == "" {
		switch {
		case normalized.Intent != "":
			normalized.Goal = normalized.Intent
		case normalized.Message != "":
			normalized.Goal = normalized.Message
		case normalized.Operation != "":
			normalized.Goal = normalized.Operation
		}
	}
	if normalized.LaneRole == "" {
		normalized.LaneRole = defaultTeamLaneRole(normalized.AskKind)
	}
	return normalized
}

// TeamAskFromMap upgrades an ad hoc object payload into the canonical TeamAsk
// contract so runtime, audit, and future UI surfaces can interpret the same
// delegated ask shape consistently.
func TeamAskFromMap(raw map[string]any) TeamAsk {
	ask := TeamAsk{
		SchemaVersion:        firstNonEmptyMapString(raw, "schema_version", "version"),
		AskKind:              TeamAskKind(firstNonEmptyMapString(raw, "ask_kind", "kind")),
		LaneRole:             TeamLaneRole(firstNonEmptyMapString(raw, "lane_role", "role", "lane")),
		Goal:                 firstNonEmptyValue(firstNonEmptyMapString(raw, "goal"), firstNonEmptyMapString(raw, "intent"), firstNonEmptyMapString(raw, "message"), firstNonEmptyMapString(raw, "task")),
		Operation:            firstNonEmptyMapString(raw, "operation"),
		Intent:               firstNonEmptyMapString(raw, "intent"),
		Message:              firstNonEmptyMapString(raw, "message"),
		OwnedScope:           anyStringSlice(raw["owned_scope"]),
		Constraints:          anyStringSlice(raw["constraints"]),
		RequiredCapabilities: anyStringSlice(raw["required_capabilities"]),
		ApprovalPosture:      ApprovalPosture(firstNonEmptyMapString(raw, "approval_posture")),
		ExitCriteria:         anyStringSlice(raw["exit_criteria"]),
		EvidenceRequired:     anyStringSlice(raw["evidence_required"]),
	}
	if ctxRaw, ok := raw["context"]; ok {
		if ctxMap, ok := ctxRaw.(map[string]any); ok {
			ask.Context = ctxMap
		}
	}
	return ask.Normalize()
}

func SummarizeTeamAsk(raw TeamAsk) string {
	ask := raw.Normalize()
	if ask.IsZero() {
		return ""
	}

	role := strings.TrimSpace(string(ask.LaneRole))
	kind := strings.TrimSpace(string(ask.AskKind))
	goal := strings.TrimSpace(ask.Goal)

	switch {
	case role != "" && goal != "":
		return titleCase(role) + " ask: " + goal
	case kind != "" && goal != "":
		return titleCase(kind) + " ask: " + goal
	case goal != "":
		return goal
	case role != "" && kind != "":
		return titleCase(role) + " handling a " + kind + " ask"
	case kind != "":
		return titleCase(kind) + " team ask"
	default:
		return ""
	}
}

func (a TeamAsk) HasContent() bool {
	return a.Goal != "" || a.Operation != "" || a.Intent != "" || a.Message != ""
}

func (a TeamAsk) Normalize() TeamAsk {
	return NormalizeTeamAsk(a)
}

func (a TeamAsk) IsZero() bool {
	return !a.HasContent() &&
		len(a.OwnedScope) == 0 &&
		len(a.Constraints) == 0 &&
		len(a.RequiredCapabilities) == 0 &&
		len(a.ExitCriteria) == 0 &&
		len(a.EvidenceRequired) == 0 &&
		a.Context == nil
}

func normalizeTeamAskKind(kind TeamAskKind) TeamAskKind {
	switch kind {
	case TeamAskKindResearch, TeamAskKindImplementation, TeamAskKindValidation, TeamAskKindCoordination, TeamAskKindReview:
		return kind
	default:
		return TeamAskKindImplementation
	}
}

func normalizeTeamLaneRole(role TeamLaneRole) TeamLaneRole {
	switch role {
	case TeamLaneRoleCoordinator, TeamLaneRoleResearcher, TeamLaneRoleImplementer, TeamLaneRoleValidator, TeamLaneRoleReviewer:
		return role
	default:
		return ""
	}
}

func defaultTeamLaneRole(kind TeamAskKind) TeamLaneRole {
	switch kind {
	case TeamAskKindResearch:
		return TeamLaneRoleResearcher
	case TeamAskKindValidation:
		return TeamLaneRoleValidator
	case TeamAskKindCoordination:
		return TeamLaneRoleCoordinator
	case TeamAskKindReview:
		return TeamLaneRoleReviewer
	default:
		return TeamLaneRoleImplementer
	}
}

func normalizeApprovalPosture(posture ApprovalPosture) ApprovalPosture {
	switch posture {
	case ApprovalPostureOptional, ApprovalPostureRequired:
		return posture
	default:
		return ApprovalPostureAutoAllowed
	}
}

func compactStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstNonEmptyMapString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		if raw == nil {
			return ""
		}
		if value, ok := raw[key]; ok {
			if text, ok := value.(string); ok {
				if trimmed := strings.TrimSpace(text); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return ""
}

func anyStringSlice(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			if trimmed := strings.TrimSpace(text); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstNonEmptyValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func titleCase(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

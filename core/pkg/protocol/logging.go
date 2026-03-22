package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
)

const OperationalLogSchemaVersion = "v1"

// LogReviewScope defines how broadly a log entry should be interpreted.
type LogReviewScope string

const (
	LogReviewScopeServiceLocal  LogReviewScope = "service_local"
	LogReviewScopeTeamLocal     LogReviewScope = "team_local"
	LogReviewScopeCentralReview LogReviewScope = "central_review"
	LogReviewScopeAudit         LogReviewScope = "audit"
)

// LogAudience identifies which agent/operator layer should consume a log entry.
type LogAudience string

const (
	LogAudienceSoma       LogAudience = "soma"
	LogAudienceMeta       LogAudience = "meta_agentry"
	LogAudienceTeamLead   LogAudience = "team_lead"
	LogAudienceTeam       LogAudience = "team"
	LogAudienceGovernance LogAudience = "governance"
)

// OperationalLogContext is the canonical review-oriented schema stored inside
// log_entries.context. It keeps service-local output readable while making the
// same record suitable for centralized review by Soma, meta-agentry, and team
// leads.
type OperationalLogContext struct {
	SchemaVersion     string            `json:"schema_version"`
	ReviewScope       LogReviewScope    `json:"review_scope"`
	Audiences         []LogAudience     `json:"audiences,omitempty"`
	Service           string            `json:"service,omitempty"`
	Component         string            `json:"component,omitempty"`
	Summary           string            `json:"summary,omitempty"`
	Detail            string            `json:"detail,omitempty"`
	WhyItMatters      string            `json:"why_it_matters,omitempty"`
	SuggestedAction   string            `json:"suggested_action,omitempty"`
	SourceKind        SignalSourceKind  `json:"source_kind,omitempty"`
	SourceChannel     string            `json:"source_channel,omitempty"`
	PayloadKind       SignalPayloadKind `json:"payload_kind,omitempty"`
	RunID             string            `json:"run_id,omitempty"`
	TeamID            string            `json:"team_id,omitempty"`
	AgentID           string            `json:"agent_id,omitempty"`
	TraceID           string            `json:"trace_id,omitempty"`
	MissionEventID    string            `json:"mission_event_id,omitempty"`
	Status            string            `json:"status,omitempty"`
	CentralizedReview bool              `json:"centralized_review"`
	ReviewChannels    []string          `json:"review_channels,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
}

func (c OperationalLogContext) Normalize() OperationalLogContext {
	out := c
	out.SchemaVersion = OperationalLogSchemaVersion
	out.Service = strings.TrimSpace(out.Service)
	out.Component = strings.TrimSpace(out.Component)
	out.Summary = strings.TrimSpace(out.Summary)
	out.Detail = strings.TrimSpace(out.Detail)
	out.WhyItMatters = strings.TrimSpace(out.WhyItMatters)
	out.SuggestedAction = strings.TrimSpace(out.SuggestedAction)
	out.SourceChannel = strings.TrimSpace(out.SourceChannel)
	out.RunID = strings.TrimSpace(out.RunID)
	out.TeamID = strings.TrimSpace(out.TeamID)
	out.AgentID = strings.TrimSpace(out.AgentID)
	out.TraceID = strings.TrimSpace(out.TraceID)
	out.MissionEventID = strings.TrimSpace(out.MissionEventID)
	out.Status = strings.TrimSpace(strings.ToLower(out.Status))

	switch out.ReviewScope {
	case LogReviewScopeServiceLocal, LogReviewScopeTeamLocal, LogReviewScopeCentralReview, LogReviewScopeAudit:
	default:
		switch {
		case out.MissionEventID != "":
			out.ReviewScope = LogReviewScopeAudit
		case out.TeamID != "" || out.AgentID != "" || out.SourceChannel != "" || out.RunID != "":
			out.ReviewScope = LogReviewScopeCentralReview
		default:
			out.ReviewScope = LogReviewScopeServiceLocal
		}
	}

	if len(out.Audiences) == 0 {
		switch out.ReviewScope {
		case LogReviewScopeServiceLocal:
			out.Audiences = []LogAudience{LogAudienceTeam}
		case LogReviewScopeTeamLocal:
			out.Audiences = []LogAudience{LogAudienceTeamLead, LogAudienceTeam}
		case LogReviewScopeAudit:
			out.Audiences = []LogAudience{LogAudienceSoma, LogAudienceMeta, LogAudienceTeamLead, LogAudienceGovernance}
		default:
			out.Audiences = []LogAudience{LogAudienceSoma, LogAudienceMeta, LogAudienceTeamLead}
		}
	}
	out.Audiences = dedupeAudiences(out.Audiences)

	if !out.CentralizedReview {
		out.CentralizedReview = out.ReviewScope != LogReviewScopeServiceLocal
	}

	reviewChannels := append([]string{}, out.ReviewChannels...)
	reviewChannels = append(reviewChannels, "memory.stream")
	if out.SourceChannel != "" {
		reviewChannels = append(reviewChannels, out.SourceChannel)
	}
	if out.RunID != "" {
		reviewChannels = append(reviewChannels, fmt.Sprintf(TopicMissionEventsFmt, out.RunID))
	}
	out.ReviewChannels = dedupeStrings(reviewChannels)
	out.Tags = dedupeStrings(out.Tags)

	return out
}

func (c OperationalLogContext) ToMap() map[string]any {
	normalized := c.Normalize()
	data, err := json.Marshal(normalized)
	if err != nil {
		return map[string]any{
			"schema_version":     OperationalLogSchemaVersion,
			"review_scope":       string(normalized.ReviewScope),
			"centralized_review": normalized.CentralizedReview,
			"summary":            normalized.Summary,
		}
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return map[string]any{
			"schema_version":     OperationalLogSchemaVersion,
			"review_scope":       string(normalized.ReviewScope),
			"centralized_review": normalized.CentralizedReview,
			"summary":            normalized.Summary,
		}
	}
	return out
}

func ParseOperationalLogContext(raw map[string]any) OperationalLogContext {
	if len(raw) == 0 {
		return OperationalLogContext{}.Normalize()
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return OperationalLogContext{}.Normalize()
	}
	var ctx OperationalLogContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return OperationalLogContext{}.Normalize()
	}
	return ctx.Normalize()
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func dedupeAudiences(values []LogAudience) []LogAudience {
	seen := make(map[LogAudience]struct{}, len(values))
	out := make([]LogAudience, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

package protocol

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

type ConversationTemplateScope string
type ConversationTemplateCreatorKind string
type ConversationTemplateStatus string

const (
	ConversationTemplateScopeSoma           ConversationTemplateScope = "soma"
	ConversationTemplateScopeCouncil        ConversationTemplateScope = "council"
	ConversationTemplateScopeTeam           ConversationTemplateScope = "team"
	ConversationTemplateScopeTemporaryGroup ConversationTemplateScope = "temporary_group"
)

const (
	ConversationTemplateCreatorUser    ConversationTemplateCreatorKind = "user"
	ConversationTemplateCreatorSoma    ConversationTemplateCreatorKind = "soma"
	ConversationTemplateCreatorCouncil ConversationTemplateCreatorKind = "council"
	ConversationTemplateCreatorSystem  ConversationTemplateCreatorKind = "system"
)

const (
	ConversationTemplateStatusActive   ConversationTemplateStatus = "active"
	ConversationTemplateStatusDraft    ConversationTemplateStatus = "draft"
	ConversationTemplateStatusArchived ConversationTemplateStatus = "archived"
)

type ConversationTemplate struct {
	ID                   string                          `json:"id"`
	TenantID             string                          `json:"tenant_id"`
	Name                 string                          `json:"name"`
	Description          string                          `json:"description,omitempty"`
	Scope                ConversationTemplateScope       `json:"scope"`
	CreatedBy            string                          `json:"created_by"`
	CreatorKind          ConversationTemplateCreatorKind `json:"creator_kind"`
	Status               ConversationTemplateStatus      `json:"status"`
	TemplateBody         string                          `json:"template_body"`
	Variables            map[string]any                  `json:"variables,omitempty"`
	OutputContract       map[string]any                  `json:"output_contract,omitempty"`
	RecommendedTeamShape map[string]any                  `json:"recommended_team_shape,omitempty"`
	ModelRoutingHint     map[string]any                  `json:"model_routing_hint,omitempty"`
	GovernanceTags       []string                        `json:"governance_tags,omitempty"`
	CreatedAt            time.Time                       `json:"created_at,omitempty"`
	UpdatedAt            time.Time                       `json:"updated_at,omitempty"`
	LastUsedAt           *time.Time                      `json:"last_used_at,omitempty"`
}

type ConversationTemplateWorkflowGroupDraft struct {
	Name                string   `json:"name,omitempty"`
	GoalStatement       string   `json:"goal_statement,omitempty"`
	WorkMode            string   `json:"work_mode,omitempty"`
	CoordinatorProfile  string   `json:"coordinator_profile,omitempty"`
	AllowedCapabilities []string `json:"allowed_capabilities,omitempty"`
	ExpiryHours         int      `json:"expiry_hours,omitempty"`
	Summary             string   `json:"summary,omitempty"`
}

type ConversationTemplateInstantiation struct {
	TemplateID       string                                  `json:"template_id"`
	Scope            ConversationTemplateScope               `json:"scope"`
	RenderedPrompt   string                                  `json:"rendered_prompt"`
	AskContract      *AskContract                            `json:"ask_contract,omitempty"`
	TeamAsk          *TeamAsk                                `json:"team_ask,omitempty"`
	WorkflowGroup    *ConversationTemplateWorkflowGroupDraft `json:"workflow_group,omitempty"`
	OutputContract   map[string]any                          `json:"output_contract,omitempty"`
	ModelRoutingHint map[string]any                          `json:"model_routing_hint,omitempty"`
	GovernanceTags   []string                                `json:"governance_tags,omitempty"`
}

func NormalizeConversationTemplate(raw ConversationTemplate) ConversationTemplate {
	normalized := raw
	normalized.TenantID = firstNonEmptyValue(raw.TenantID, "default")
	normalized.Name = strings.TrimSpace(raw.Name)
	normalized.Description = strings.TrimSpace(raw.Description)
	normalized.Scope = NormalizeConversationTemplateScope(raw.Scope)
	normalized.CreatedBy = firstNonEmptyValue(raw.CreatedBy, "admin")
	normalized.CreatorKind = NormalizeConversationTemplateCreatorKind(raw.CreatorKind)
	normalized.Status = NormalizeConversationTemplateStatus(raw.Status)
	normalized.TemplateBody = strings.TrimSpace(raw.TemplateBody)
	if normalized.Variables == nil {
		normalized.Variables = map[string]any{}
	}
	if normalized.OutputContract == nil {
		normalized.OutputContract = map[string]any{}
	}
	if normalized.RecommendedTeamShape == nil {
		normalized.RecommendedTeamShape = map[string]any{}
	}
	if normalized.ModelRoutingHint == nil {
		normalized.ModelRoutingHint = map[string]any{}
	}
	normalized.GovernanceTags = compactStrings(raw.GovernanceTags)
	if normalized.GovernanceTags == nil {
		normalized.GovernanceTags = []string{}
	}
	return normalized
}

func NormalizeConversationTemplateScope(scope ConversationTemplateScope) ConversationTemplateScope {
	switch scope {
	case ConversationTemplateScopeSoma, ConversationTemplateScopeCouncil, ConversationTemplateScopeTeam, ConversationTemplateScopeTemporaryGroup:
		return scope
	default:
		return ConversationTemplateScopeSoma
	}
}

func NormalizeConversationTemplateCreatorKind(kind ConversationTemplateCreatorKind) ConversationTemplateCreatorKind {
	switch kind {
	case ConversationTemplateCreatorUser, ConversationTemplateCreatorSoma, ConversationTemplateCreatorCouncil, ConversationTemplateCreatorSystem:
		return kind
	default:
		return ConversationTemplateCreatorUser
	}
}

func NormalizeConversationTemplateStatus(status ConversationTemplateStatus) ConversationTemplateStatus {
	switch status {
	case ConversationTemplateStatusActive, ConversationTemplateStatusDraft, ConversationTemplateStatusArchived:
		return status
	default:
		return ConversationTemplateStatusActive
	}
}

func InstantiateConversationTemplate(raw ConversationTemplate, suppliedVariables map[string]any) ConversationTemplateInstantiation {
	tpl := NormalizeConversationTemplate(raw)
	vars := make(map[string]any, len(tpl.Variables)+len(suppliedVariables))
	for key, value := range tpl.Variables {
		vars[key] = value
	}
	for key, value := range suppliedVariables {
		vars[key] = value
	}
	rendered := renderConversationTemplateBody(tpl.TemplateBody, vars)
	out := ConversationTemplateInstantiation{
		TemplateID:       tpl.ID,
		Scope:            tpl.Scope,
		RenderedPrompt:   rendered,
		OutputContract:   tpl.OutputContract,
		ModelRoutingHint: tpl.ModelRoutingHint,
		GovernanceTags:   tpl.GovernanceTags,
	}
	switch tpl.Scope {
	case ConversationTemplateScopeTeam, ConversationTemplateScopeTemporaryGroup:
		ask := NormalizeTeamAsk(TeamAsk{
			AskKind:         TeamAskKindCoordination,
			LaneRole:        TeamLaneRoleCoordinator,
			Goal:            rendered,
			Message:         rendered,
			ApprovalPosture: ApprovalPostureAutoAllowed,
			Context: map[string]any{
				"conversation_template_id": tpl.ID,
				"conversation_template":    tpl.Name,
			},
		})
		out.TeamAsk = &ask
		if tpl.Scope == ConversationTemplateScopeTemporaryGroup {
			out.WorkflowGroup = workflowGroupDraftFromTemplate(tpl, rendered)
		}
	default:
		contract, ok := AskContractForClass(AskClassDirectAnswer)
		if ok {
			out.AskContract = &contract
		}
	}
	return out
}

var conversationTemplateVarPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

func renderConversationTemplateBody(body string, vars map[string]any) string {
	return conversationTemplateVarPattern.ReplaceAllStringFunc(body, func(match string) string {
		parts := conversationTemplateVarPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		value, ok := vars[parts[1]]
		if !ok {
			return match
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				return match
			}
			return typed
		default:
			encoded, err := json.Marshal(typed)
			if err != nil || string(encoded) == "null" {
				return match
			}
			return string(encoded)
		}
	})
}

func workflowGroupDraftFromTemplate(tpl ConversationTemplate, rendered string) *ConversationTemplateWorkflowGroupDraft {
	name := stringFromMap(tpl.RecommendedTeamShape, "name")
	if name == "" {
		name = tpl.Name + " temporary workflow"
	}
	coordinator := stringFromMap(tpl.RecommendedTeamShape, "coordinator_profile")
	if coordinator == "" {
		coordinator = tpl.Name + " lead"
	}
	workMode := stringFromMap(tpl.RecommendedTeamShape, "work_mode")
	if workMode == "" {
		workMode = "propose_only"
	}
	return &ConversationTemplateWorkflowGroupDraft{
		Name:                name,
		GoalStatement:       rendered,
		WorkMode:            workMode,
		CoordinatorProfile:  coordinator,
		AllowedCapabilities: stringSliceFromMap(tpl.RecommendedTeamShape, "allowed_capabilities"),
		ExpiryHours:         intFromMap(tpl.RecommendedTeamShape, "expiry_hours", 72),
		Summary:             "Instantiates a temporary workflow from the registered conversation template without executing it automatically.",
	}
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if value, ok := values[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func stringSliceFromMap(values map[string]any, key string) []string {
	if values == nil {
		return nil
	}
	if items, ok := values[key].([]string); ok {
		return compactStrings(items)
	}
	if items, ok := values[key].([]any); ok {
		out := make([]string, 0, len(items))
		for _, item := range items {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return compactStrings(out)
	}
	return nil
}

func intFromMap(values map[string]any, key string, fallback int) int {
	if values == nil {
		return fallback
	}
	switch value := values[key].(type) {
	case int:
		if value > 0 {
			return value
		}
	case float64:
		if value > 0 {
			return int(value)
		}
	}
	return fallback
}

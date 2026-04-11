package swarm

import (
	"context"
	"fmt"

	"github.com/mycelis/core/internal/conversationtemplates"
	"github.com/mycelis/core/pkg/protocol"
)

func (r *InternalToolRegistry) handleStoreConversationTemplate(ctx context.Context, args map[string]any) (string, error) {
	if r.db == nil {
		return "", fmt.Errorf("conversation template store not available")
	}
	name := stringValue(args["name"])
	scope := protocol.ConversationTemplateScope(stringValue(args["scope"]))
	templateBody := stringValue(args["template_body"])
	if name == "" || scope == "" || templateBody == "" {
		return "", fmt.Errorf("store_conversation_template requires 'name', 'scope', and 'template_body'")
	}
	memoryScope := resolveMemoryScope(ctx, args)
	tpl := protocol.ConversationTemplate{
		Name:                 name,
		Description:          stringValue(args["description"]),
		Scope:                scope,
		CreatedBy:            memoryScope.AgentID,
		CreatorKind:          protocol.ConversationTemplateCreatorSoma,
		Status:               protocol.ConversationTemplateStatusActive,
		TemplateBody:         templateBody,
		Variables:            mapValue(args["variables"]),
		OutputContract:       mapValue(args["output_contract"]),
		RecommendedTeamShape: mapValue(args["recommended_team_shape"]),
		ModelRoutingHint:     mapValue(args["model_routing_hint"]),
		GovernanceTags:       stringSlice(args["governance_tags"]),
	}
	created, err := conversationtemplates.NewStore(r.db).Create(ctx, tpl)
	if err != nil {
		return "", fmt.Errorf("store conversation template failed: %w", err)
	}
	return mustJSON(map[string]any{"message": "Conversation template stored.", "template": created}), nil
}

func (r *InternalToolRegistry) handleInstantiateConversationTemplate(ctx context.Context, args map[string]any) (string, error) {
	if r.db == nil {
		return "", fmt.Errorf("conversation template store not available")
	}
	templateID := firstNonEmptyString(stringValue(args["template_id"]), stringValue(args["id"]))
	if templateID == "" {
		return "", fmt.Errorf("instantiate_conversation_template requires 'template_id'")
	}
	instantiation, err := conversationtemplates.NewStore(r.db).Instantiate(ctx, templateID, mapValue(args["variables"]))
	if err != nil {
		return "", fmt.Errorf("instantiate conversation template failed: %w", err)
	}
	return mustJSON(map[string]any{"message": "Conversation template instantiated without execution.", "instantiation": instantiation}), nil
}

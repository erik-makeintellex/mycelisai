package protocol

import "testing"

func TestInstantiateConversationTemplate_RendersDirectSomaAsk(t *testing.T) {
	tpl := ConversationTemplate{
		ID:           "template-1",
		Name:         "Positioning",
		Scope:        ConversationTemplateScopeSoma,
		TemplateBody: "Draft {{output}} for {{company}}.",
		Variables:    map[string]any{"output": "a launch brief"},
	}

	instantiation := InstantiateConversationTemplate(tpl, map[string]any{"company": "Mycelis"})

	if instantiation.RenderedPrompt != "Draft a launch brief for Mycelis." {
		t.Fatalf("rendered prompt = %q", instantiation.RenderedPrompt)
	}
	if instantiation.AskContract == nil || instantiation.AskContract.AskClass != AskClassDirectAnswer {
		t.Fatalf("expected direct answer ask contract, got %#v", instantiation.AskContract)
	}
	if instantiation.TeamAsk != nil {
		t.Fatalf("direct Soma template should not create team ask: %#v", instantiation.TeamAsk)
	}
}

func TestInstantiateConversationTemplate_TemporaryGroupReturnsDraftWithoutExecution(t *testing.T) {
	tpl := ConversationTemplate{
		ID:           "template-2",
		Name:         "Marketing package",
		Scope:        ConversationTemplateScopeTemporaryGroup,
		TemplateBody: "Create a launch package for {{product}}.",
		RecommendedTeamShape: map[string]any{
			"name":                 "Marketing Launch Team temporary workflow",
			"coordinator_profile":  "Marketing Launch Team lead",
			"allowed_capabilities": []any{"content.plan", "artifact.review"},
		},
	}

	instantiation := InstantiateConversationTemplate(tpl, map[string]any{"product": "Mycelis"})

	if instantiation.TeamAsk == nil || instantiation.TeamAsk.Goal != "Create a launch package for Mycelis." {
		t.Fatalf("expected rendered team ask, got %#v", instantiation.TeamAsk)
	}
	if instantiation.WorkflowGroup == nil {
		t.Fatal("expected workflow group draft")
	}
	if instantiation.WorkflowGroup.Name != "Marketing Launch Team temporary workflow" {
		t.Fatalf("workflow group name = %q", instantiation.WorkflowGroup.Name)
	}
	if instantiation.WorkflowGroup.WorkMode != "propose_only" {
		t.Fatalf("workflow group work mode = %q", instantiation.WorkflowGroup.WorkMode)
	}
	if len(instantiation.WorkflowGroup.AllowedCapabilities) != 2 {
		t.Fatalf("allowed capabilities = %#v", instantiation.WorkflowGroup.AllowedCapabilities)
	}
}

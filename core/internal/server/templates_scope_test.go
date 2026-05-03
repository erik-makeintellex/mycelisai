package server

import (
	"encoding/json"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildExecutionAuditDetailsForTool_DelegateTaskIncludesStructuredSummary(t *testing.T) {
	details := buildExecutionAuditDetailsForTool(protocol.PlannedToolCall{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_id": "research-team",
			"ask": map[string]any{
				"ask_kind":  "research",
				"lane_role": "researcher",
				"goal":      "Map the provider authentication drift.",
			},
		},
	}, "delegate_task")

	if details["tool"] != "delegate_task" {
		t.Fatalf("tool = %v", details["tool"])
	}
	if details["team_id"] != "research-team" {
		t.Fatalf("team_id = %v", details["team_id"])
	}
	if details["ask_kind"] != "research" {
		t.Fatalf("ask_kind = %v", details["ask_kind"])
	}
	if details["operator_summary"] != "Researcher ask: Map the provider authentication drift." {
		t.Fatalf("operator_summary = %v", details["operator_summary"])
	}
}

// CE-1: Scope Validation Builder

func TestBuildScopeFromBlueprint(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: []protocol.BlueprintTeam{
			{
				Name: "research-team",
				Agents: []protocol.AgentManifest{
					{ID: "researcher", Tools: []string{"search_memory", "recall"}},
					{ID: "writer", Tools: []string{"write_file"}},
				},
			},
			{
				Name: "analysis-team",
				Agents: []protocol.AgentManifest{
					{ID: "analyst", Tools: []string{"search_memory"}},
				},
			},
		},
	}

	scope := buildScopeFromBlueprint(bp)
	if scope == nil {
		t.Fatal("Expected scope, got nil")
	}
	if len(scope.Tools) != 3 {
		t.Errorf("Expected 3 unique tools, got %d: %v", len(scope.Tools), scope.Tools)
	}
	if scope.RiskLevel != "low" {
		t.Errorf("Expected risk 'low' for 3 agents, got %q", scope.RiskLevel)
	}
}

func TestBuildScopeFromBlueprint_HighRisk(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: make([]protocol.BlueprintTeam, 5),
	}
	for i := range bp.Teams {
		bp.Teams[i].Agents = []protocol.AgentManifest{
			{ID: "a1"}, {ID: "a2"}, {ID: "a3"},
		}
	}

	scope := buildScopeFromBlueprint(bp)
	if scope.RiskLevel != "high" {
		t.Errorf("Expected risk 'high' for 15 agents / 5 teams, got %q", scope.RiskLevel)
	}
}

// CE-1: Template Protocol Types

func TestTemplateRegistryContainsBothTemplates(t *testing.T) {
	if _, ok := protocol.TemplateRegistry[protocol.TemplateChatToAnswer]; !ok {
		t.Error("Template registry missing chat-to-answer")
	}
	if _, ok := protocol.TemplateRegistry[protocol.TemplateChatToProposal]; !ok {
		t.Error("Template registry missing chat-to-proposal")
	}
}

func TestChatToAnswerTemplateSpec(t *testing.T) {
	spec := protocol.TemplateRegistry[protocol.TemplateChatToAnswer]
	if spec.RequiresConfirm {
		t.Error("Chat-to-Answer should not require confirmation")
	}
	if spec.MutatesState {
		t.Error("Chat-to-Answer should not mutate state")
	}
	if spec.Mode != protocol.ModeAnswer {
		t.Errorf("Expected mode 'answer', got %q", spec.Mode)
	}
}

func TestChatToProposalTemplateSpec(t *testing.T) {
	spec := protocol.TemplateRegistry[protocol.TemplateChatToProposal]
	if !spec.RequiresConfirm {
		t.Error("Chat-to-Proposal must require confirmation")
	}
	if !spec.MutatesState {
		t.Error("Chat-to-Proposal must mutate state")
	}
	if spec.Mode != protocol.ModeProposal {
		t.Errorf("Expected mode 'proposal', got %q", spec.Mode)
	}
}

// CE-1: CommitRequest serialization

func TestCommitRequest_ConfirmTokenField(t *testing.T) {
	body := `{"intent":"Build scraper","confirm_token":"abc-123","teams":[]}`
	var req protocol.CommitRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if req.ConfirmToken != "abc-123" {
		t.Errorf("Expected confirm_token 'abc-123', got %q", req.ConfirmToken)
	}
	if req.Intent != "Build scraper" {
		t.Errorf("Expected intent 'Build scraper', got %q", req.Intent)
	}
}

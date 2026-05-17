package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestInferAdapterKindFromTool(t *testing.T) {
	tests := []struct {
		tool string
		want string
	}{
		{tool: "mcp_fetch", want: "mcp"},
		{tool: "call_api", want: "openapi"},
		{tool: "local_command", want: "host"},
		{tool: "delegate", want: "internal"},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			got := inferAdapterKindFromTool(tt.tool)
			if got != tt.want {
				t.Fatalf("inferAdapterKindFromTool(%q) = %q, want %q", tt.tool, got, tt.want)
			}
		})
	}
}

func TestBuildMutationChatProposal(t *testing.T) {
	proposal := buildMutationChatProposal(
		[]string{"delegate", "mcp_fetch", "delegate"},
		"proof-123",
		"token-123",
		"admin-core",
		[]string{"admin"},
		&protocol.ApprovalPolicy{
			ApprovalRequired: true,
			ApprovalReason:   "capability_risk",
			ApprovalMode:     "required",
			CapabilityRisk:   "medium",
		},
		&protocol.GovernanceProfileSnapshot{Role: "owner"},
		proposalDisplayContract{
			OperatorSummary:   "Hand the requested work to the right team.",
			ExpectedResult:    "The approved task will be routed to the selected team with execution proof.",
			AffectedResources: []string{"governed state"},
		},
	)

	if proposal == nil {
		t.Fatal("expected non-nil proposal")
	}
	if proposal.Intent != "chat-action" {
		t.Fatalf("intent = %q, want chat-action", proposal.Intent)
	}
	if proposal.IntentProofID != "proof-123" {
		t.Fatalf("intent_proof_id = %q, want proof-123", proposal.IntentProofID)
	}
	if proposal.OperatorSummary != "Hand the requested work to the right team." {
		t.Fatalf("operator_summary = %q", proposal.OperatorSummary)
	}
	if proposal.ExpectedResult != "The approved task will be routed to the selected team with execution proof." {
		t.Fatalf("expected_result = %q", proposal.ExpectedResult)
	}
	if len(proposal.AffectedResources) != 1 || proposal.AffectedResources[0] != "governed state" {
		t.Fatalf("affected_resources = %#v", proposal.AffectedResources)
	}
	if proposal.ConfirmToken != "token-123" {
		t.Fatalf("confirm_token = %q, want token-123", proposal.ConfirmToken)
	}
	// duplicate "delegate" should be removed while preserving order
	if len(proposal.Tools) != 2 || proposal.Tools[0] != "delegate" || proposal.Tools[1] != "mcp_fetch" {
		t.Fatalf("tools = %#v, want [delegate mcp_fetch]", proposal.Tools)
	}
	if proposal.BusScope != "current_team" {
		t.Fatalf("bus_scope = %q, want current_team", proposal.BusScope)
	}
	if len(proposal.NATSSubjects) == 0 || proposal.NATSSubjects[0] != "swarm.team.admin-core.internal.command" {
		t.Fatalf("nats_subjects = %#v", proposal.NATSSubjects)
	}
	if len(proposal.TeamExpressions) != 2 {
		t.Fatalf("team_expressions length = %d, want 2", len(proposal.TeamExpressions))
	}
	if proposal.Approval == nil || !proposal.Approval.ApprovalRequired {
		t.Fatalf("expected approval metadata, got %+v", proposal.Approval)
	}
	if proposal.GovernanceProfile == nil || proposal.GovernanceProfile.Role != "owner" {
		t.Fatalf("expected governance profile owner, got %+v", proposal.GovernanceProfile)
	}

	first := proposal.TeamExpressions[0]
	if first.TeamID != "admin-core" {
		t.Fatalf("team_id = %q, want admin-core", first.TeamID)
	}
	if len(first.RolePlan) != 1 || first.RolePlan[0] != "admin" {
		t.Fatalf("role_plan = %#v, want [admin]", first.RolePlan)
	}
	if len(first.ModuleBindings) != 1 {
		t.Fatalf("module_bindings length = %d, want 1", len(first.ModuleBindings))
	}
	if first.ModuleBindings[0].ModuleID != "delegate" {
		t.Fatalf("module_id = %q, want delegate", first.ModuleBindings[0].ModuleID)
	}
	if first.ModuleBindings[0].AdapterKind != "internal" {
		t.Fatalf("adapter_kind = %q, want internal", first.ModuleBindings[0].AdapterKind)
	}

	second := proposal.TeamExpressions[1]
	if second.ModuleBindings[0].AdapterKind != "mcp" {
		t.Fatalf("adapter_kind = %q, want mcp", second.ModuleBindings[0].AdapterKind)
	}
}

func TestBuildProposalDisplayContractUsesDelegateTaskFallback(t *testing.T) {
	display := buildProposalDisplayContract(nil, "", []string{"delegate_task"})

	if display.OperatorSummary != "Hand the requested work to the right team." {
		t.Fatalf("operator_summary = %q", display.OperatorSummary)
	}
	if display.ExpectedResult != "The approved task will be routed to the selected team with execution proof." {
		t.Fatalf("expected_result = %q", display.ExpectedResult)
	}
	if len(display.AffectedResources) != 1 || display.AffectedResources[0] != "governed state" {
		t.Fatalf("affected_resources = %#v, want [governed state]", display.AffectedResources)
	}
}

func TestBuildProposalDisplayContractDefaultsCreateTeamToTeamBus(t *testing.T) {
	display := buildProposalDisplayContract([]protocol.PlannedToolCall{
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id": "research-team",
				"name":    "Research Team",
			},
		},
	}, "", []string{"create_team"})

	if display.OperatorSummary != "Create Research Team as a governed runtime team." {
		t.Fatalf("operator summary = %q", display.OperatorSummary)
	}
	if display.BusScope != "current_team" {
		t.Fatalf("bus_scope = %q, want current_team", display.BusScope)
	}
	want := []string{
		"swarm.team.research-team.internal.command",
		"swarm.team.research-team.signal.status",
		"swarm.team.research-team.signal.result",
	}
	for i, subject := range want {
		if len(display.NATSSubjects) <= i || display.NATSSubjects[i] != subject {
			t.Fatalf("nats_subjects = %#v, want prefix %#v", display.NATSSubjects, want)
		}
	}
	if len(display.NATSSubjects) != len(want) {
		t.Fatalf("nats_subjects = %#v, want exactly %#v", display.NATSSubjects, want)
	}
	proposal := buildMutationChatProposal(
		[]string{"create_team"},
		"proof-123",
		"token-123",
		"admin-core",
		nil,
		nil,
		nil,
		display,
	)
	if len(proposal.NATSSubjects) != len(want) {
		t.Fatalf("proposal nats_subjects = %#v, want exactly %#v", proposal.NATSSubjects, want)
	}
	for _, subject := range proposal.NATSSubjects {
		if strings.Contains(subject, "admin-core") {
			t.Fatalf("proposal nats_subjects leaked fallback team: %#v", proposal.NATSSubjects)
		}
	}
}

func TestBuildProposalDisplayContractExplainsTeamDeliverable(t *testing.T) {
	display := buildProposalDisplayContract([]protocol.PlannedToolCall{
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id": "game-team",
				"name":    "Game Team",
			},
		},
		{
			Name: "write_file",
			Arguments: map[string]any{
				"path": "workspace/generated/game-team-first-game/index.html",
			},
		},
	}, "", []string{"create_team", "write_file"})

	if display.OperatorSummary != "Create Game Team and start its first retained deliverable." {
		t.Fatalf("operator summary = %q", display.OperatorSummary)
	}
	if !strings.Contains(display.ExpectedResult, "workspace/generated/game-team-first-game/index.html") {
		t.Fatalf("expected_result = %q, want retained output path", display.ExpectedResult)
	}
	if len(display.AffectedResources) != 2 {
		t.Fatalf("affected_resources = %#v, want team and file path", display.AffectedResources)
	}
}

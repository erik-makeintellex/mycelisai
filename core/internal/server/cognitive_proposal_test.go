package server

import (
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
	if proposal.ConfirmToken != "token-123" {
		t.Fatalf("confirm_token = %q, want token-123", proposal.ConfirmToken)
	}
	// duplicate "delegate" should be removed while preserving order
	if len(proposal.Tools) != 2 || proposal.Tools[0] != "delegate" || proposal.Tools[1] != "mcp_fetch" {
		t.Fatalf("tools = %#v, want [delegate mcp_fetch]", proposal.Tools)
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

package protocol

import "testing"

func TestAskContractForClass_DirectAnswer(t *testing.T) {
	contract, ok := AskContractForClass(AskClassDirectAnswer)
	if !ok {
		t.Fatal("expected direct_answer contract")
	}
	if contract.TemplateID != TemplateChatToAnswer {
		t.Fatalf("template = %q, want %q", contract.TemplateID, TemplateChatToAnswer)
	}
	if contract.DefaultExecutionMode != ModeAnswer {
		t.Fatalf("mode = %q, want %q", contract.DefaultExecutionMode, ModeAnswer)
	}
	if contract.RequiresConfirmation {
		t.Fatal("direct answer should not require confirmation")
	}
	if contract.DefaultAgentTarget != "soma" {
		t.Fatalf("default agent target = %q, want soma", contract.DefaultAgentTarget)
	}
}

func TestAskContractForClass_GovernedMutation(t *testing.T) {
	contract, ok := AskContractForClass(AskClassGovernedMutation)
	if !ok {
		t.Fatal("expected governed_mutation contract")
	}
	if contract.TemplateID != TemplateChatToProposal {
		t.Fatalf("template = %q, want %q", contract.TemplateID, TemplateChatToProposal)
	}
	if contract.DefaultExecutionMode != ModeProposal {
		t.Fatalf("mode = %q, want %q", contract.DefaultExecutionMode, ModeProposal)
	}
	if !contract.RequiresConfirmation {
		t.Fatal("governed mutation should require confirmation")
	}
	if contract.ApprovalPosture != ApprovalPostureRequired {
		t.Fatalf("approval posture = %q, want %q", contract.ApprovalPosture, ApprovalPostureRequired)
	}
}

func TestAskContractForClass_Unknown(t *testing.T) {
	if _, ok := AskContractForClass(AskClass("unknown")); ok {
		t.Fatal("expected unknown ask class lookup to fail")
	}
}

package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestParsePlannedToolCall_NormalizesWriteFileAliases(t *testing.T) {
	call, ok := parsePlannedToolCall(`{"tool_call":{"name":"write_file","arguments":{"file_path":"workspace/logs/alias_test.py","body":"print('hello world')"}}}`)
	if !ok {
		t.Fatal("expected planned tool call to parse")
	}
	if call.Arguments["path"] != "workspace/logs/alias_test.py" {
		t.Fatalf("path = %#v", call.Arguments["path"])
	}
	if call.Arguments["content"] != "print('hello world')" {
		t.Fatalf("content = %#v", call.Arguments["content"])
	}
}

func TestResolveChatAskContract(t *testing.T) {
	direct := resolveChatAskContract("soma", false, chatAgentResult{})
	if direct.AskClass != protocol.AskClassDirectAnswer {
		t.Fatalf("direct ask class = %q, want %q", direct.AskClass, protocol.AskClassDirectAnswer)
	}
	if direct.TemplateID != protocol.TemplateChatToAnswer {
		t.Fatalf("direct template = %q, want %q", direct.TemplateID, protocol.TemplateChatToAnswer)
	}
	if direct.DefaultExecutionMode != protocol.ModeAnswer {
		t.Fatalf("direct mode = %q, want %q", direct.DefaultExecutionMode, protocol.ModeAnswer)
	}

	artifact := resolveChatAskContract("soma", false, chatAgentResult{
		Artifacts: []protocol.ChatArtifactRef{{Type: "document", Title: "Brief"}},
	})
	if artifact.AskClass != protocol.AskClassGovernedArtifact {
		t.Fatalf("artifact ask class = %q, want %q", artifact.AskClass, protocol.AskClassGovernedArtifact)
	}
	if artifact.DefaultExecutionMode != protocol.ModeAnswer {
		t.Fatalf("artifact mode = %q, want %q", artifact.DefaultExecutionMode, protocol.ModeAnswer)
	}

	specialist := resolveChatAskContract("specialist", false, chatAgentResult{})
	if specialist.AskClass != protocol.AskClassSpecialist {
		t.Fatalf("specialist ask class = %q, want %q", specialist.AskClass, protocol.AskClassSpecialist)
	}
	if specialist.DefaultAgentTarget != "specialist" {
		t.Fatalf("specialist default target = %q, want specialist", specialist.DefaultAgentTarget)
	}

	consulted := resolveChatAskContract("soma", false, chatAgentResult{
		Consultations: []protocol.ConsultationEntry{{Member: "council-architect", Summary: "Reviewed the plan."}},
	})
	if consulted.AskClass != protocol.AskClassSpecialist {
		t.Fatalf("consulted ask class = %q, want %q", consulted.AskClass, protocol.AskClassSpecialist)
	}

	mutation := resolveChatAskContract("soma", true, chatAgentResult{})
	if mutation.AskClass != protocol.AskClassGovernedMutation {
		t.Fatalf("mutation ask class = %q, want %q", mutation.AskClass, protocol.AskClassGovernedMutation)
	}
	if mutation.TemplateID != protocol.TemplateChatToProposal {
		t.Fatalf("mutation template = %q, want %q", mutation.TemplateID, protocol.TemplateChatToProposal)
	}
	if mutation.DefaultExecutionMode != protocol.ModeProposal {
		t.Fatalf("mutation mode = %q, want %q", mutation.DefaultExecutionMode, protocol.ModeProposal)
	}
	if !mutation.RequiresConfirmation {
		t.Fatal("mutation contract should require confirmation")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

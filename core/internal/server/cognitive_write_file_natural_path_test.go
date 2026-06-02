package server

import (
	"fmt"
	"strings"
	"testing"
)

func TestDeterministicGovernedMutationResult_BuildsWriteFileProposalFromNaturalAtPath(t *testing.T) {
	request := "Create a markdown file at generated/workbench-review/operator-note.md containing \"# Workbench Review\\n\\n- Open outputs near Soma.\""
	result, ok := deterministicGovernedMutationResult(request, []string{"write_file"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}

	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	if len(calls) != 1 || calls[0].Name != "write_file" {
		t.Fatalf("planned calls = %#v, want one write_file call", calls)
	}
	if calls[0].Arguments["path"] != "generated/workbench-review/operator-note.md" {
		t.Fatalf("path = %#v, want natural at-path target", calls[0].Arguments["path"])
	}
	if !strings.Contains(fmt.Sprint(calls[0].Arguments["content"]), "Open outputs near Soma") {
		t.Fatalf("content = %#v, want quoted markdown content", calls[0].Arguments["content"])
	}
}

func TestDeterministicGovernedMutationResult_BuildsWriteFileProposalFromDescribedContent(t *testing.T) {
	request := "Create a file named business-owner-welcome.txt with a plain-language welcome note that explains where my generated outputs will appear."
	result, ok := deterministicGovernedMutationResult(request, []string{"write_file"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}

	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	if len(calls) != 1 || calls[0].Name != "write_file" {
		t.Fatalf("planned calls = %#v, want one write_file call", calls)
	}
	if calls[0].Arguments["path"] != "business-owner-welcome.txt" {
		t.Fatalf("path = %#v, want named file target", calls[0].Arguments["path"])
	}
	content := fmt.Sprint(calls[0].Arguments["content"])
	if !strings.Contains(content, "Welcome to Mycelis") || !strings.Contains(content, "open the file or open the containing folder") {
		t.Fatalf("content = %#v, want synthesized business-owner output guidance", content)
	}
}

func TestDeterministicGovernedMutationResult_BuildsWriteFileProposalFromBusinessOwnerBrowserPrompt(t *testing.T) {
	request := strings.Join([]string{
		"Create a markdown file at generated/business-owner-flow-1780423128313/owner-note.md.",
		"Put exactly \"# Business Owner Flow\n\nThe approval path must return output or one clear recovery action.\"",
		"Return retained output and proof.",
	}, " ")
	result, ok := deterministicGovernedMutationResult(request, []string{"write_file"})
	if !ok {
		t.Fatal("expected deterministic governed mutation result")
	}

	calls := buildPlannedToolCalls(result, request, result.ToolsUsed)
	if len(calls) != 1 || calls[0].Name != "write_file" {
		t.Fatalf("planned calls = %#v, want one write_file call", calls)
	}
	if calls[0].Arguments["path"] != "generated/business-owner-flow-1780423128313/owner-note.md" {
		t.Fatalf("path = %#v, want browser prompt file target", calls[0].Arguments["path"])
	}
	content := fmt.Sprint(calls[0].Arguments["content"])
	if !strings.Contains(content, "# Business Owner Flow") || !strings.Contains(content, "one clear recovery action") {
		t.Fatalf("content = %#v, want exact requested markdown body", content)
	}
}

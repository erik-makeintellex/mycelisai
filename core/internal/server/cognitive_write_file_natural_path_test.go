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

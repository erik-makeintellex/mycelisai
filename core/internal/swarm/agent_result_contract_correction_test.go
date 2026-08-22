package swarm

import (
	"strings"
	"testing"
)

func TestResultContractCorrectionFocusesOneMissingFileBeforeReadback(t *testing.T) {
	issues := []string{
		"missing successful write for README.md",
		"missing successful write for PROOF.md",
		"missing successful structural readback of a written output",
	}
	focused := focusedResultContractCorrectionIssues(issues)
	if len(focused) != 1 || focused[0] != issues[0] {
		t.Fatalf("focused issues = %v, want first required write", focused)
	}
	prompt := resultContractCorrectionPrompt(&teamResultRequirement{}, issues)
	if !strings.Contains(prompt, "README.md") || strings.Contains(prompt, "PROOF.md") || strings.Contains(prompt, "structural readback") {
		t.Fatalf("correction prompt did not isolate the next write: %q", prompt)
	}
}

func TestResultContractCorrectionCombinesEntrypointRepairsBeforeReadback(t *testing.T) {
	issues := []string{
		"missing successful structural readback of a written output",
		"entrypoint readback does not expose visible control instructions",
		"entrypoint readback is missing approved validation target [data-mycelis-validation-surface]",
	}
	focused := focusedResultContractCorrectionIssues(issues)
	joined := strings.Join(focused, ";")
	if len(focused) != 2 || !strings.Contains(joined, "visible control") || !strings.Contains(joined, "validation target") {
		t.Fatalf("focused issues = %v, want both entrypoint repairs", focused)
	}
}

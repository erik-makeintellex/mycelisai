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
	prompt := resultContractCorrectionPrompt(&teamResultRequirement{}, issues, nil, nil)
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

func TestResultContractCorrectionSeparatesMarkupFromGameplayRepair(t *testing.T) {
	issues := []string{
		"entrypoint readback is missing approved validation target [data-mycelis-primary-action]",
		"game canvas has no inspectable 2d render implementation",
		"movement controls do not mutate game state consumed by the canvas render loop",
	}
	markup := focusedResultContractCorrectionIssues(issues)
	if len(markup) != 1 || !strings.Contains(markup[0], "validation target") {
		t.Fatalf("focused issues = %v, want markup repair first", markup)
	}
	gameplay := focusedResultContractCorrectionIssues(issues[1:])
	if len(gameplay) != 2 {
		t.Fatalf("gameplay issues = %v, want bounded gameplay repair group", gameplay)
	}
	requirement := &teamResultRequirement{
		Kind: "project_package", TeamID: "game-team",
		FilesRequired:      []string{"index.html", "game.js", "styles.css"},
		AcceptanceCriteria: []string{"playable controls move the player and change the visible game surface"},
	}
	prompt := resultContractCorrectionPrompt(requirement, issues[1:], nil, nil)
	if !strings.Contains(prompt, "groups/game-team/generated/package/game.js") ||
		!strings.Contains(prompt, "Repair the complete state model") {
		t.Fatalf("gameplay correction did not target game.js: %s", prompt)
	}
}

func TestResultContractCorrectionNamesCanonicalPackageEntrypoint(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind:          "project_package",
		TeamID:        "application-delivery-team-a1b2c",
		FilesRequired: []string{"index.html", "README.md", "PROOF.md", "project-package.json"},
	}
	prompt := resultContractCorrectionPrompt(requirement, []string{"missing successful write for index.html"}, nil, nil)

	for _, want := range []string{
		"groups/application-delivery-team-a1b2c/generated/package/index.html",
		"package_kind=project_package",
		"package_files=[index.html, README.md, PROOF.md, project-package.json]",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("correction prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestFunctionalGameExecutionPromptUsesSplitPackageAndRealState(t *testing.T) {
	requirement := &teamResultRequirement{
		Kind:          "project_package",
		FilesRequired: []string{"index.html", "game.js", "styles.css", "README.md", "PROOF.md", "project-package.json"},
		AcceptanceCriteria: []string{
			"playable controls move the player and change the visible game surface",
			"attack changes enemy, hazard, or score state",
		},
	}
	prompt := resultContractExecutionPrompt(requirement)
	for _, want := range []string{"small accessible shell", "styles.css and game.js", "canvas render loop", "one missing file per tool call"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
	correction := resultContractCorrectionPrompt(requirement, []string{"missing successful write for index.html"}, nil, nil)
	for _, want := range []string{"game.js", "styles.css", "project-package.json"} {
		if !strings.Contains(correction, want) {
			t.Fatalf("correction missing required package file %q: %s", want, correction)
		}
	}
}

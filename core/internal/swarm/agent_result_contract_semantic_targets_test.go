package swarm

import (
	"strings"
	"testing"
)

func TestSemanticAcceptanceEntrypointIssuesRequiresGameHooks(t *testing.T) {
	criteria := []string{
		"attack changes enemy, hazard, or score state",
		"hazard contact changes health state",
	}
	content := `<p id="score">0</p><button data-mycelis-validation-action="attack">Attack</button>`

	issues := strings.Join(semanticAcceptanceEntrypointIssues(criteria, content), "; ")
	if !strings.Contains(issues, `data-mycelis-validation-action="hazard"`) || !strings.Contains(issues, "#health") {
		t.Fatalf("issues = %q, want missing hazard and health targets", issues)
	}
	if strings.Contains(issues, `missing semantic validation target [data-mycelis-validation-action="attack"]`) || strings.Contains(issues, "missing semantic validation target #score") {
		t.Fatalf("issues = %q, present target hooks should not be reported missing", issues)
	}
	if !strings.Contains(issues, "attack action does not mutate governed") {
		t.Fatalf("issues = %q, target-only attack facade should fail functional validation", issues)
	}
}

func TestSemanticAcceptanceEntrypointIssuesAcceptsRevisionHooks(t *testing.T) {
	criteria := []string{"requested revision changes its declared visible revision state"}
	content := `<button data-mycelis-validation-action="revision">Try revision</button><p data-mycelis-revision-state>Ready</p>`
	if issues := semanticAcceptanceEntrypointIssues(criteria, content); len(issues) != 0 {
		t.Fatalf("issues = %#v, want complete revision hooks", issues)
	}
}

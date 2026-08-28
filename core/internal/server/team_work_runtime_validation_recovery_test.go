package server

import (
	"context"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/outputvalidation"
	"github.com/mycelis/core/pkg/protocol"
)

func TestGeneratedPackageFailureCopyUsesStableOperatorSafeLanguage(t *testing.T) {
	classes := []string{
		"runtime_validation_failed", "runtime_validation_unavailable", "runtime_validation_deadline",
		"runtime_validation_stale", "result_contract_unsatisfied", semanticAcceptanceUnverified,
	}
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			copy := normalizedGeneratedPackageFailure(class)
			combined := copy.Headline + " " + copy.Summary + " " + copy.Recovery
			if strings.Contains(combined, "\n") || len(combined) > 420 {
				t.Fatalf("copy is not concise single-line language: %q", combined)
			}
			for _, forbidden := range []string{"http://", "https://", "data-mycelis", "querySelector", "/home/", "playwright", "ollama", "tool_call", "\x1b"} {
				if strings.Contains(strings.ToLower(combined), strings.ToLower(forbidden)) {
					t.Fatalf("copy exposed forbidden detail %q: %q", forbidden, combined)
				}
			}
			if !strings.Contains(strings.ToLower(copy.Recovery), "same team") {
				t.Fatalf("recovery does not preserve team ownership: %q", copy.Recovery)
			}
		})
	}
}

func TestRuntimeValidationStatusAndInteractionDoNotExposeRawDiagnostics(t *testing.T) {
	raw := "\x1b[31mTimeout at http://127.0.0.1:8081/private using [data-mycelis-primary-action] /home/user/output provider=ollama\x1b[0m"
	report := outputvalidation.Report{Status: outputvalidation.StatusFailed, Diagnostics: []outputvalidation.Diagnostic{{Code: "raw", Message: raw}}}
	item := protocol.TeamWorkItem{
		TeamID: "game-team", WorkItemID: "work-1", State: protocol.TeamWorkStateDegraded,
		DegradationState: "runtime_validation_failed",
		RecoveryOptions:  []string{normalizedGeneratedPackageFailure("runtime_validation_failed").Recovery},
	}
	event := runtimeValidationStatusEvent(item, report, false)
	interaction := runtimeValidationInteraction(item, report, false, "groups/game-team/proof/report.json")
	for label, value := range map[string]string{"event": event.Details + " " + event.NextAction, "interaction": interaction.Summary} {
		if strings.Contains(value, "127.0.0.1") || strings.Contains(value, "data-mycelis") || strings.Contains(value, "/home/") || strings.Contains(value, "ollama") || strings.Contains(value, "\x1b") {
			t.Fatalf("%s exposed raw diagnostic: %q", label, value)
		}
	}
}

func TestRuntimeValidationUnavailableClassSeparatesDeadline(t *testing.T) {
	if got := runtimeValidationUnavailableClass(context.DeadlineExceeded); got != "runtime_validation_deadline" {
		t.Fatalf("deadline class = %q", got)
	}
	if got := runtimeValidationUnavailableClass(context.Canceled); got != "runtime_validation_unavailable" {
		t.Fatalf("unavailable class = %q", got)
	}
}

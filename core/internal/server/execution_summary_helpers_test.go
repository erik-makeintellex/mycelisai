package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/pkg/protocol"
)

func TestConfirmActionResponseDataIncludesTeamWorkRefs(t *testing.T) {
	refs := []confirmActionTeamWorkRef{
		{
			WorkItemID:      "work-1",
			TeamID:          "qa-team",
			State:           protocol.TeamWorkStateQueued,
			RunID:           "run-1",
			ExpectedOutputs: []string{"Playable browser app"},
		},
	}

	scope := &protocol.ScopeValidation{
		WorkIntent: &protocol.WorkIntent{
			Kind: "project",
			OutputContract: &protocol.WorkOutputContract{
				Shape:              "app_package",
				PrimaryDeliverable: "Playable browser app",
			},
		},
		ExecutionMode: "confirm_then_execute",
	}
	data := confirmActionResponseData("proof-1", "contract-1", "artifact-1", "run-1", "audit-1", scope, nil, refs, nil)

	if data["confirmed"] != true || data["verified"] != true {
		t.Fatalf("response flags = confirmed:%v verified:%v", data["confirmed"], data["verified"])
	}
	if data["run_status"] != runs.StatusCompleted {
		t.Fatalf("run_status = %v, want %s", data["run_status"], runs.StatusCompleted)
	}
	got, ok := data["team_work_refs"].([]confirmActionTeamWorkRef)
	if !ok {
		t.Fatalf("team_work_refs = %T, want []confirmActionTeamWorkRef", data["team_work_refs"])
	}
	if len(got) != 1 || got[0].WorkItemID != "work-1" || got[0].TeamID != "qa-team" || got[0].State != protocol.TeamWorkStateQueued || got[0].RunID != "run-1" {
		t.Fatalf("team_work_refs = %#v", got)
	}
	if len(got[0].ExpectedOutputs) != 1 || got[0].ExpectedOutputs[0] != "Playable browser app" {
		t.Fatalf("team_work_refs expected_outputs = %#v", got[0].ExpectedOutputs)
	}
	summary, ok := data["execution_summary"].(*protocol.ExecutionSummary)
	if !ok {
		t.Fatalf("execution_summary = %T, want *protocol.ExecutionSummary", data["execution_summary"])
	}
	if summary.WorkIntent == nil || summary.WorkIntent.OutputContract == nil {
		t.Fatalf("execution_summary missing work_intent: %+v", summary.WorkIntent)
	}
	if summary.WorkIntent.OutputContract.PrimaryDeliverable != "Playable browser app" {
		t.Fatalf("primary deliverable = %q", summary.WorkIntent.OutputContract.PrimaryDeliverable)
	}
}

func TestConfirmActionResponseDataKeepsDelegatedExecutionRunning(t *testing.T) {
	data := confirmActionResponseDataForStatus(
		"proof-1", "contract-1", "", "run-1", "audit-1", runs.StatusRunning,
		&protocol.ScopeValidation{},
		[]plannedToolExecutionResult{{Name: "delegate_task"}},
		[]confirmActionTeamWorkRef{{WorkItemID: "work-1", TeamID: "qa-team", State: protocol.TeamWorkStateQueued}},
		nil,
	)
	if data["verified"] != false || data["execution_state"] != "running" || data["run_status"] != runs.StatusRunning {
		t.Fatalf("running response = %#v", data)
	}
	summary, ok := data["execution_summary"].(*protocol.ExecutionSummary)
	if !ok {
		t.Fatalf("execution_summary = %T", data["execution_summary"])
	}
	if summary.Execution.Status != protocol.ExecutionStatusRunning {
		t.Fatalf("execution status = %q", summary.Execution.Status)
	}
	if summary.Proof.Verified == nil || *summary.Proof.Verified {
		t.Fatalf("pending proof must remain unverified: %#v", summary.Proof)
	}
	if summary.Proof.ProofID != "" {
		t.Fatalf("pending execution must not claim a completion proof: %#v", summary.Proof)
	}
}

func TestConfirmedActionHasPendingTeamWork(t *testing.T) {
	if confirmedActionHasPendingTeamWork([]plannedToolExecutionResult{{Name: "write_file"}}) {
		t.Fatal("synchronous file output should not remain pending")
	}
	if !confirmedActionHasPendingTeamWork([]plannedToolExecutionResult{{Name: "delegate_task"}}) {
		t.Fatal("delegated work must remain pending")
	}
}

func TestConfirmActionSummaryExplainsOutcomeTemplateLifecycle(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		summary  string
		nextStep string
	}{
		{
			name:     "stored",
			tool:     "store_config_document",
			summary:  "saved and remains inactive",
			nextStep: "Ask Soma to use this Outcome Template.",
		},
		{
			name:     "active",
			tool:     "activate_config_document",
			summary:  "active and ready to shape new work",
			nextStep: "Tell Soma what outcome you want to create with this template.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := buildConfirmActionExecutionSummary(
				"intent-proof", "contract", "proof", "run", "audit",
				&protocol.ScopeValidation{},
				[]plannedToolExecutionResult{{Name: tc.tool}},
			)
			if !strings.Contains(summary.Execution.Summary, tc.summary) {
				t.Fatalf("execution summary = %q, want %q", summary.Execution.Summary, tc.summary)
			}
			if summary.NextStep == nil || summary.NextStep.Label != tc.nextStep {
				t.Fatalf("next step = %#v, want %q", summary.NextStep, tc.nextStep)
			}
		})
	}
}

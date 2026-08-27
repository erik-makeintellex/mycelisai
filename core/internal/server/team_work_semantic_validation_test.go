package server

import (
	"testing"

	"github.com/mycelis/core/internal/outputvalidation"
	"github.com/mycelis/core/pkg/protocol"
)

func TestSemanticCriteriaRequireExplicitEvidenceForEveryCriterion(t *testing.T) {
	item := protocol.TeamWorkItem{WorkIntent: &protocol.WorkIntent{OutputContract: &protocol.WorkOutputContract{
		SemanticValidationRequired: true,
		AcceptanceCriteria:         []string{"Player can move and jump", "Win and restart are testable"},
	}}}
	report := outputvalidation.Report{CriterionEvidence: []outputvalidation.CriterionEvidence{{
		Criterion: "Player can move and jump", Passed: true, EvidenceRefs: []string{"proof/movement.png"},
	}}}
	if semanticCriteriaSatisfied(item, report) {
		t.Fatal("partial criterion evidence must not satisfy semantic acceptance")
	}
	report.CriterionEvidence = append(report.CriterionEvidence, outputvalidation.CriterionEvidence{
		Criterion: " Win and restart are testable ", Passed: true, EvidenceRefs: []string{"proof/win.json"},
	})
	if !semanticCriteriaSatisfied(item, report) {
		t.Fatal("explicit passing evidence for every criterion should satisfy semantic acceptance")
	}
}

func TestSemanticCriteriaFailClosedWithoutCriteriaOrEvidenceRefs(t *testing.T) {
	item := protocol.TeamWorkItem{WorkIntent: &protocol.WorkIntent{OutputContract: &protocol.WorkOutputContract{
		SemanticValidationRequired: true,
	}}}
	if semanticCriteriaSatisfied(item, outputvalidation.Report{}) {
		t.Fatal("semantic posture without criteria must fail closed")
	}
	item.WorkIntent.OutputContract.AcceptanceCriteria = []string{"Primary workflow works"}
	report := outputvalidation.Report{CriterionEvidence: []outputvalidation.CriterionEvidence{{
		Criterion: "Primary workflow works", Passed: true,
	}}}
	if semanticCriteriaSatisfied(item, report) {
		t.Fatal("criterion assertion without retained evidence must fail closed")
	}
}

func TestSemanticCompletionGateRejectsGenericPassingReport(t *testing.T) {
	item := protocol.TeamWorkItem{WorkIntent: &protocol.WorkIntent{OutputContract: &protocol.WorkOutputContract{
		SemanticValidationRequired: true,
		AcceptanceCriteria:         []string{"Win and restart are testable"},
	}}}
	passed, degradation := applySemanticCompletionGate(item, outputvalidation.Report{Status: outputvalidation.StatusPassed}, true, "")
	if passed || degradation != semanticAcceptanceUnverified {
		t.Fatalf("gate = (%v, %q), want false and semantic degradation", passed, degradation)
	}
	report := outputvalidation.Report{CriterionEvidence: []outputvalidation.CriterionEvidence{{
		Criterion: "Win and restart are testable", Passed: true, EvidenceRefs: []string{"proof/win.json"},
	}}}
	passed, degradation = applySemanticCompletionGate(item, report, true, "")
	if !passed || degradation != "" {
		t.Fatalf("evidenced gate = (%v, %q), want pass", passed, degradation)
	}
}

func TestConfirmedPackageCannotSkipSemanticValidation(t *testing.T) {
	item := protocol.TeamWorkItem{WorkIntent: &protocol.WorkIntent{OutputContract: &protocol.WorkOutputContract{
		SemanticValidationRequired: true,
		AcceptanceCriteria:         []string{"Win and restart are testable"},
	}}}
	if issue := confirmedProjectPackageOutputIssue(item); issue != semanticAcceptanceUnverified {
		t.Fatalf("issue = %q, want %q", issue, semanticAcceptanceUnverified)
	}
}

func TestProposalPersistsSemanticAcceptancePosture(t *testing.T) {
	contract := inferProposalOutputContract("Build an original gothic browser action-platformer", nil, "", proposalDisplayContract{})
	if !contract.SemanticValidationRequired || len(contract.AcceptanceCriteria) == 0 {
		t.Fatalf("semantic contract = %#v, want durable acceptance criteria", contract)
	}
	normalized := protocol.NormalizeWorkIntent(&protocol.WorkIntent{Kind: "project", OutputContract: contract})
	if normalized.OutputContract == nil || !normalized.OutputContract.SemanticValidationRequired || len(normalized.OutputContract.AcceptanceCriteria) == 0 {
		t.Fatalf("normalized contract lost semantic posture: %#v", normalized.OutputContract)
	}
}

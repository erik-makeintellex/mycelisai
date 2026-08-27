package server

import (
	"strings"

	"github.com/mycelis/core/internal/outputvalidation"
	"github.com/mycelis/core/pkg/protocol"
)

const semanticAcceptanceUnverified = "semantic_acceptance_unverified"

func semanticValidationRequired(item protocol.TeamWorkItem) bool {
	return item.WorkIntent != nil && item.WorkIntent.OutputContract != nil &&
		item.WorkIntent.OutputContract.SemanticValidationRequired
}

func semanticCriteriaSatisfied(item protocol.TeamWorkItem, report outputvalidation.Report) bool {
	if !semanticValidationRequired(item) {
		return true
	}
	criteria := item.WorkIntent.OutputContract.AcceptanceCriteria
	if len(criteria) == 0 {
		return false
	}
	evidence := make(map[string]bool, len(report.CriterionEvidence))
	for _, proof := range report.CriterionEvidence {
		criterion := normalizeSemanticCriterion(proof.Criterion)
		if criterion != "" && proof.Passed && len(normalizeStringSlice(proof.EvidenceRefs)) > 0 {
			evidence[criterion] = true
		}
	}
	for _, criterion := range criteria {
		if !evidence[normalizeSemanticCriterion(criterion)] {
			return false
		}
	}
	return true
}

func applySemanticCompletionGate(item protocol.TeamWorkItem, report outputvalidation.Report, passed bool, degradation string) (bool, string) {
	if passed && !semanticCriteriaSatisfied(item, report) {
		return false, semanticAcceptanceUnverified
	}
	return passed, degradation
}

func normalizeSemanticCriterion(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

package server

import (
	"context"
	"errors"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

type generatedPackageFailureCopy struct {
	Headline string
	Summary  string
	Recovery string
}

func normalizedGeneratedPackageFailure(class string) generatedPackageFailureCopy {
	switch strings.TrimSpace(class) {
	case "runtime_validation_unavailable":
		return generatedPackageFailureCopy{
			Headline: "Deliverable check unavailable",
			Summary:  "The runtime check could not complete for this retained candidate.",
			Recovery: "Ask Soma to retry validation for the same team's retained candidate.",
		}
	case "runtime_validation_deadline":
		return generatedPackageFailureCopy{
			Headline: "Deliverable check timed out",
			Summary:  "The runtime check did not finish within its bounded validation window.",
			Recovery: "Ask Soma to have the same team verify the primary workflow, then return a fresh candidate for validation.",
		}
	case "runtime_validation_stale":
		return generatedPackageFailureCopy{
			Headline: "Deliverable changed during validation",
			Summary:  "A newer retained candidate replaced the version being checked.",
			Recovery: "Ask Soma to validate the latest candidate from the same team.",
		}
	case "result_contract_unsatisfied":
		return generatedPackageFailureCopy{
			Headline: "Deliverable contract incomplete",
			Summary:  "The team did not complete the approved retained-package and evidence contract.",
			Recovery: "Ask Soma to have the same team complete the package and required evidence, then return a new candidate.",
		}
	case semanticAcceptanceUnverified:
		return generatedPackageFailureCopy{
			Headline: "Deliverable needs acceptance proof",
			Summary:  "The retained candidate does not yet have passing evidence for every approved acceptance criterion.",
			Recovery: "Ask Soma to have the same team repair and recheck the unmet acceptance criteria, then return a new candidate.",
		}
	default:
		return generatedPackageFailureCopy{
			Headline: "Deliverable needs repair",
			Summary:  "The retained candidate did not pass the approved runtime workflow check.",
			Recovery: "Ask Soma to have the same team repair the primary workflow, then return a new retained candidate for validation.",
		}
	}
}

func runtimeValidationUnavailableClass(cause error) string {
	if errors.Is(cause, context.DeadlineExceeded) {
		return "runtime_validation_deadline"
	}
	lower := strings.ToLower(strings.TrimSpace(cause.Error()))
	if strings.Contains(lower, "deadline") || strings.Contains(lower, "timed out") || strings.Contains(lower, "timeout") {
		return "runtime_validation_deadline"
	}
	return "runtime_validation_unavailable"
}

func runtimeValidationFailureCopy(item protocol.TeamWorkItem) generatedPackageFailureCopy {
	return normalizedGeneratedPackageFailure(item.DegradationState)
}

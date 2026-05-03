package server

import (
	"slices"
	"strings"

	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/pkg/protocol"
)

func approvalThresholdsForProfile(profile userGovernanceProfile) approvalThresholds {
	thresholds := approvalThresholds{
		MaxCost:         1.0,
		MaxRisk:         "medium",
		AllowExternal:   false,
		RequireEscalate: false,
	}

	switch profile.CostSensitivity {
	case "low":
		thresholds.MaxCost = 5.0
	case "high":
		thresholds.MaxCost = 0.25
	}

	switch profile.ReviewStrictness {
	case "light":
		thresholds.MaxRisk = "high"
	case "strict":
		thresholds.MaxRisk = "low"
	}

	switch profile.AutomationTolerance {
	case "cautious":
		thresholds.AllowExternal = false
		if thresholds.MaxRisk == "medium" {
			thresholds.MaxRisk = "low"
		}
	case "aggressive":
		thresholds.AllowExternal = true
		if thresholds.MaxRisk == "medium" {
			thresholds.MaxRisk = "high"
		}
	}

	if profile.EscalationPreference == "halt" {
		thresholds.RequireEscalate = true
	}

	return thresholds
}

func buildApprovalPolicy(profile userGovernanceProfile, planned []protocol.PlannedToolCall, fallbackTools []string) *protocol.ApprovalPolicy {
	if len(planned) == 0 && len(fallbackTools) == 0 {
		return nil
	}

	thresholds := approvalThresholdsForProfile(profile)
	capabilityIDs := []string{}
	risk := "low"
	estimatedCost := 0.0
	externalDataUse := false

	if len(planned) == 0 {
		for _, tool := range fallbackTools {
			planned = append(planned, protocol.PlannedToolCall{Name: tool})
		}
	}

	for _, item := range planned {
		if capabilityID := capabilityForPlannedTool(item.Name); capabilityID != "" && !slices.Contains(capabilityIDs, capabilityID) {
			capabilityIDs = append(capabilityIDs, capabilityID)
			if capability, ok := exchange.CapabilityByID(capabilityID); ok {
				risk = maxRisk(risk, strings.TrimSuffix(capability.RiskClass, "-risk"))
				externalDataUse = externalDataUse || capability.Source == "mcp" || capability.Source == "api" || capability.Source == "node"
			}
		}
		risk = maxRisk(risk, capabilityRiskForTool(item.Name, item.Arguments))
		estimatedCost += estimateActionCost(item.Name, item.Arguments)
		externalDataUse = externalDataUse || externalDataUseForTool(item.Name, item.Arguments)
	}

	policy := &protocol.ApprovalPolicy{
		ApprovalRequired:     false,
		ApprovalMode:         "auto_allowed",
		CapabilityRisk:       risk,
		CapabilityIDs:        capabilityIDs,
		ExternalDataUse:      externalDataUse,
		EstimatedCost:        estimatedCost,
		RequiredApproverRole: profile.Role,
		ApprovalSteps:        []string{"operator_review", "future_role_gate"},
		GovernanceProfile:    profile.snapshot(),
	}

	switch {
	case approvalRank(risk) >= 3:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "capability_risk"
	case externalDataUse && !thresholds.AllowExternal:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "external_data_use"
	case estimatedCost > thresholds.MaxCost:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "cost"
	case approvalRank(risk) > approvalRank(thresholds.MaxRisk):
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "capability_risk"
	case thresholds.RequireEscalate && approvalRank(risk) >= 2:
		policy.ApprovalRequired = true
		policy.ApprovalMode = "required"
		policy.ApprovalReason = "escalation_preference"
	case approvalRank(risk) == 2:
		policy.ApprovalMode = "optional"
		policy.ApprovalReason = "capability_risk"
	default:
		policy.ApprovalReason = "auto_approve"
	}

	return policy
}

func policyDecisionForApproval(policy *protocol.ApprovalPolicy) string {
	if policy == nil {
		return "allow"
	}
	if policy.ApprovalRequired {
		return "require_approval"
	}
	return "allow"
}

func approvalStatusValue(policy *protocol.ApprovalPolicy) string {
	if policy == nil {
		return "allow"
	}
	if policy.ApprovalRequired {
		return "approval_required"
	}
	return policy.ApprovalMode
}

func approvalReasonValue(policy *protocol.ApprovalPolicy) string {
	if policy == nil {
		return ""
	}
	return policy.ApprovalReason
}

package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

type proposalDisplayContract struct {
	OperatorSummary   string
	ExpectedResult    string
	AffectedResources []string
}

func firstStringArgument(arguments map[string]any, key string) string {
	if arguments == nil {
		return ""
	}
	raw, ok := arguments[key]
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func formatProposalResource(resource string) string {
	trimmed := strings.TrimSpace(resource)
	if trimmed == "" {
		return ""
	}
	if trimmed == "state" {
		return "governed state"
	}
	return trimmed
}

func buildProposalDisplayContract(planned []protocol.PlannedToolCall, latestRequest string, mutTools []string) proposalDisplayContract {
	display := proposalDisplayContract{
		OperatorSummary: "Carry out the requested governed action.",
		ExpectedResult:  "Soma will perform the approved action and return durable execution proof.",
	}

	for _, resource := range affectedResourcesForPlannedCalls(planned) {
		if formatted := formatProposalResource(resource); formatted != "" {
			display.AffectedResources = append(display.AffectedResources, formatted)
		}
	}

	if len(planned) > 0 {
		switch strings.TrimSpace(planned[0].Name) {
		case "write_file":
			path := firstStringArgument(planned[0].Arguments, "path")
			if path != "" {
				display.OperatorSummary = fmt.Sprintf("Create %q in the workspace.", path)
				display.ExpectedResult = fmt.Sprintf("One new workspace file will be created at %q after approval.", path)
				if len(display.AffectedResources) == 0 {
					display.AffectedResources = []string{path}
				}
				return display
			}
			display.OperatorSummary = "Create a new workspace file."
			display.ExpectedResult = "One new workspace file will be created after approval."
			return display
		case "publish_signal":
			subject := firstStringArgument(planned[0].Arguments, "subject")
			if subject != "" {
				display.OperatorSummary = fmt.Sprintf("Publish a governed signal to %q.", subject)
				display.ExpectedResult = fmt.Sprintf("A signal will be sent on %q after approval.", subject)
				if len(display.AffectedResources) == 0 {
					display.AffectedResources = []string{subject}
				}
				return display
			}
			display.OperatorSummary = "Publish a governed signal."
			display.ExpectedResult = "A governed signal will be sent after approval."
			return display
		case "generate_blueprint":
			display.OperatorSummary = "Prepare a reusable blueprint from this request."
			display.ExpectedResult = "A governed blueprint draft will be created for review."
			return display
		case "delegate", "delegate_task":
			display.OperatorSummary = "Hand the requested work to the right team."
			display.ExpectedResult = "The approved task will be routed to the selected team with execution proof."
			return display
		case "broadcast":
			display.OperatorSummary = "Broadcast the requested update to connected teams."
			display.ExpectedResult = "The approved broadcast will be sent and logged with execution proof."
			return display
		case "promote_deployment_context":
			title := firstStringArgument(planned[0].Arguments, "title")
			if title != "" {
				display.OperatorSummary = fmt.Sprintf("Promote %q into approved company knowledge.", title)
				display.ExpectedResult = fmt.Sprintf("%q will be stored as approved company knowledge with lineage back to the original customer context entry.", title)
				return display
			}
			sourceArtifactID := firstStringArgument(planned[0].Arguments, "source_artifact_id")
			if sourceArtifactID != "" {
				display.OperatorSummary = "Promote an existing customer context entry into approved company knowledge."
				display.ExpectedResult = fmt.Sprintf("The customer context entry %q will be converted into a new company knowledge record with preserved lineage.", sourceArtifactID)
				return display
			}
			display.OperatorSummary = "Promote an existing customer context entry into approved company knowledge."
			display.ExpectedResult = "A new approved company knowledge record will be created with lineage back to the original customer context entry."
			return display
		}
	}

	if len(mutTools) > 0 {
		switch strings.TrimSpace(mutTools[0]) {
		case "write_file":
			display.OperatorSummary = "Create a new workspace file."
			display.ExpectedResult = "One new workspace file will be created after approval."
		case "publish_signal":
			display.OperatorSummary = "Publish a governed signal."
			display.ExpectedResult = "A governed signal will be sent after approval."
		case "generate_blueprint":
			display.OperatorSummary = "Prepare a reusable blueprint from this request."
			display.ExpectedResult = "A governed blueprint draft will be created for review."
		case "delegate", "delegate_task":
			display.OperatorSummary = "Hand the requested work to the right team."
			display.ExpectedResult = "The approved task will be routed to the selected team with execution proof."
		case "broadcast":
			display.OperatorSummary = "Broadcast the requested update to connected teams."
			display.ExpectedResult = "The approved broadcast will be sent and logged with execution proof."
		case "promote_deployment_context":
			display.OperatorSummary = "Promote an existing customer context entry into approved company knowledge."
			display.ExpectedResult = "A new approved company knowledge record will be created with lineage back to the original customer context entry."
		}
	}

	if strings.TrimSpace(latestRequest) != "" && display.OperatorSummary == "Carry out the requested governed action." {
		display.ExpectedResult = "Soma will carry out the approved request and return durable execution proof."
	}

	return display
}

func buildMutationChatProposal(mutTools []string, proofID, confirmToken, teamID string, rolePlan []string, approval *protocol.ApprovalPolicy, profile *protocol.GovernanceProfileSnapshot, display proposalDisplayContract) *protocol.ChatProposal {
	deduped := uniqueOrderedTools(mutTools)
	return &protocol.ChatProposal{
		Intent:            "chat-action",
		OperatorSummary:   display.OperatorSummary,
		ExpectedResult:    display.ExpectedResult,
		AffectedResources: display.AffectedResources,
		Tools:             deduped,
		RiskLevel:         chatToolRisk(deduped),
		ConfirmToken:      confirmToken,
		IntentProofID:     proofID,
		TeamExpressions:   buildTeamExpressionsFromTools(deduped, teamID, rolePlan),
		Approval:          approval,
		GovernanceProfile: profile,
	}
}

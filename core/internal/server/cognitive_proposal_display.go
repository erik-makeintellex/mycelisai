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
	BusScope          string
	NATSSubjects      []string
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
	return buildProposalDisplayContractForTeam(planned, latestRequest, mutTools, "admin-core")
}

func buildProposalDisplayContractForTeam(planned []protocol.PlannedToolCall, latestRequest string, mutTools []string, fallbackTeamID string) proposalDisplayContract {
	display := proposalDisplayContract{
		OperatorSummary: "Carry out the requested governed action.",
		ExpectedResult:  "Soma will perform the approved action and return durable execution proof.",
	}
	display.BusScope, display.NATSSubjects = proposalBusWiring(planned, mutTools, resolveFocusedSomaTeamID(fallbackTeamID))

	for _, resource := range affectedResourcesForPlannedCalls(planned) {
		if formatted := formatProposalResource(resource); formatted != "" {
			display.AffectedResources = append(display.AffectedResources, formatted)
		}
	}

	if len(planned) > 0 {
		if strings.TrimSpace(planned[0].Name) == "create_team" {
			if paths := plannedWriteFilePaths(planned); len(paths) > 0 {
				teamID := firstStringArgument(planned[0].Arguments, "team_id")
				name := firstStringArgument(planned[0].Arguments, "name")
				label := firstNonEmptyString(name, teamID, "the requested team")
				display.OperatorSummary = fmt.Sprintf("Create %s and start its first retained deliverable.", label)
				display.ExpectedResult = fmt.Sprintf("%s will be created, then Soma will produce reviewable outputs at %s with run proof.", label, quotedProposalPaths(paths))
				return display
			}
		}
		if toolRef := strings.TrimSpace(planned[0].ToolRef); toolRef != "" {
			path := firstStringArgument(planned[0].Arguments, "path")
			if path != "" {
				display.OperatorSummary = fmt.Sprintf("Use %s on %q through a governed MCP capability.", toolRef, path)
				display.ExpectedResult = fmt.Sprintf("The MCP capability will run after approval and retain proof for %q.", path)
				if len(display.AffectedResources) == 0 {
					display.AffectedResources = []string{path}
				}
				return display
			}
			display.OperatorSummary = fmt.Sprintf("Use %s through a governed MCP capability.", toolRef)
			display.ExpectedResult = "The MCP capability will run after approval and return durable execution proof."
			return display
		}
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
		case "create_team":
			teamID := firstStringArgument(planned[0].Arguments, "team_id")
			name := firstStringArgument(planned[0].Arguments, "name")
			label := firstNonEmptyString(name, teamID, "the requested team")
			display.OperatorSummary = fmt.Sprintf("Create %s as a governed runtime team.", label)
			display.ExpectedResult = fmt.Sprintf("%s will be created, wired to team NATS subjects, logged to the run, and mirrored in Groups.", label)
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
		case "create_team":
			display.OperatorSummary = "Create the requested governed runtime team."
			display.ExpectedResult = "The requested team will be created, wired to team NATS subjects, logged to the run, and mirrored in Groups."
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

func plannedWriteFilePaths(planned []protocol.PlannedToolCall) []string {
	paths := []string{}
	for _, call := range planned {
		if strings.TrimSpace(call.Name) != "write_file" {
			continue
		}
		if path := firstStringArgument(call.Arguments, "path"); path != "" {
			paths = append(paths, path)
		}
	}
	return paths
}

func quotedProposalPaths(paths []string) string {
	quoted := make([]string, 0, len(paths))
	for _, path := range paths {
		quoted = append(quoted, fmt.Sprintf("%q", path))
	}
	return strings.Join(quoted, " and ")
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
		BusScope:          firstNonEmptyString(display.BusScope, "current_team"),
		NATSSubjects:      defaultProposalNATSSubjects(display.NATSSubjects, teamID),
		Approval:          approval,
		GovernanceProfile: profile,
	}
}

func proposalBusWiring(planned []protocol.PlannedToolCall, mutTools []string, fallbackTeamID string) (string, []string) {
	tools := append([]string{}, mutTools...)
	for _, call := range planned {
		tools = append(tools, firstNonEmptyString(call.ToolRef, call.Name))
		if isTeamBusTool(call.Name) {
			teamID := firstStringArgument(call.Arguments, "team_id")
			if teamID == "" {
				teamID = firstStringArgument(call.Arguments, "id")
			}
			if teamID == "" {
				teamID = firstStringArgument(call.Arguments, "team_name")
			}
			return "current_team", defaultProposalNATSSubjects(nil, firstNonEmptyString(teamID, fallbackTeamID))
		}
	}
	for _, tool := range tools {
		if isTeamBusTool(tool) {
			return "current_team", defaultProposalNATSSubjects(nil, fallbackTeamID)
		}
		if strings.TrimSpace(tool) == "broadcast" {
			return "global", []string{protocol.TopicGlobalBroadcast}
		}
		if strings.TrimSpace(tool) == "publish_signal" && len(planned) > 0 {
			if subject := firstStringArgument(planned[0].Arguments, "subject"); subject != "" {
				return "global", []string{subject}
			}
		}
	}
	return "current_team", defaultProposalNATSSubjects(nil, fallbackTeamID)
}

func isTeamBusTool(tool string) bool {
	switch strings.TrimSpace(tool) {
	case "create_team", "delegate", "delegate_task":
		return true
	default:
		return false
	}
}

func defaultProposalNATSSubjects(subjects []string, teamID string) []string {
	cleaned := make([]string, 0, len(subjects)+3)
	seen := map[string]struct{}{}
	for _, subject := range subjects {
		if trimmed := strings.TrimSpace(subject); trimmed != "" {
			if _, ok := seen[trimmed]; !ok {
				seen[trimmed] = struct{}{}
				cleaned = append(cleaned, trimmed)
			}
		}
	}
	if len(cleaned) > 0 {
		return cleaned
	}
	team := strings.TrimSpace(teamID)
	if team == "" {
		team = "admin-core"
	}
	for _, subject := range []string{
		fmt.Sprintf(protocol.TopicTeamInternalCommand, team),
		fmt.Sprintf(protocol.TopicTeamSignalStatus, team),
		fmt.Sprintf(protocol.TopicTeamSignalResult, team),
	} {
		if _, ok := seen[subject]; !ok {
			seen[subject] = struct{}{}
			cleaned = append(cleaned, subject)
		}
	}
	return cleaned
}

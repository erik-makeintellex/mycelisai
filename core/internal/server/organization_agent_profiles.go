package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func summarizeAgentTypeProfiles(team *swarm.TeamManifest) []OrganizationAgentTypeProfileSummary {
	if team == nil || len(team.Members) == 0 {
		return nil
	}

	profiles := make([]OrganizationAgentTypeProfileSummary, 0, len(team.Members))
	seen := make(map[string]struct{}, len(team.Members))
	for index, member := range team.Members {
		profile := buildAgentTypeProfileSummary(member, index)
		if _, exists := seen[profile.ID]; exists {
			continue
		}
		seen[profile.ID] = struct{}{}
		profiles = append(profiles, profile)
	}
	return profiles
}

func buildAgentTypeProfileSummary(member protocol.AgentManifest, fallbackIndex int) OrganizationAgentTypeProfileSummary {
	id, name, helpsWith := agentTypeProfileIdentity(member, fallbackIndex)
	return OrganizationAgentTypeProfileSummary{
		ID:                               id,
		Name:                             name,
		HelpsWith:                        helpsWith,
		AIEngineBindingProfileID:         inferAgentTypeAIEngineBinding(member),
		ResponseContractBindingProfileID: inferAgentTypeResponseBinding(member),
		OutputTypeID:                     string(inferAgentTypeOutputType(member)),
		OutputTypeLabel:                  outputTypeLabel(string(inferAgentTypeOutputType(member))),
	}
}

func agentTypeProfileIdentity(member protocol.AgentManifest, fallbackIndex int) (string, string, string) {
	role := strings.TrimSpace(strings.ToLower(member.Role))
	switch role {
	case "lead", "planner":
		return "planner", "Planner", "Turns organization goals into practical next steps, delivery sequencing, and clear priorities."
	case "research", "researcher":
		return "research-specialist", "Research Specialist", "Builds the background, options, and supporting context the Team Lead needs before decisions move forward."
	case "review", "reviewer", "qa", "quality":
		return "reviewer", "Reviewer", "Checks work for quality, risk, and readiness before the Team Lead advances the next move."
	case "builder", "implementer", "delivery":
		return "delivery-specialist", "Delivery Specialist", "Carries the work from plan into execution and keeps the main delivery lane moving."
	case "operations", "operator", "coordinator":
		return "operations-specialist", "Operations Specialist", "Keeps follow-through organized, reduces friction, and supports steady execution across the Department."
	case "support", "assistant", "guide":
		return "support-specialist", "Support Specialist", "Helps the Team Lead keep operator requests clear, coordinated, and easy to act on."
	}

	baseName := strings.TrimSpace(member.Role)
	if baseName == "" {
		baseName = strings.TrimSpace(member.ID)
	}
	baseName = humanizeMode(baseName)
	if baseName == "Guided" {
		baseName = fmt.Sprintf("Specialist %d", fallbackIndex+1)
	}
	return slugifyDepartmentID(baseName, fallbackIndex), baseName, "Supports the Department with a focused specialist role when the Team Lead needs more targeted help."
}

func inferAgentTypeAIEngineBinding(member protocol.AgentManifest) string {
	if strings.TrimSpace(member.Model) == "" && strings.TrimSpace(member.Provider) == "" {
		return ""
	}

	switch strings.TrimSpace(strings.ToLower(member.Role)) {
	case "lead", "planner", "review", "reviewer", "qa", "quality":
		return string(OrganizationAIEngineProfileHighReasoning)
	case "research", "researcher":
		return string(OrganizationAIEngineProfileDeepPlanning)
	case "builder", "implementer", "delivery", "operations", "operator", "coordinator":
		return string(OrganizationAIEngineProfileFastLightweight)
	default:
		return string(OrganizationAIEngineProfileBalanced)
	}
}

func inferAgentTypeResponseBinding(member protocol.AgentManifest) string {
	if strings.TrimSpace(member.SystemPrompt) == "" {
		return ""
	}

	switch strings.TrimSpace(strings.ToLower(member.Role)) {
	case "lead", "planner", "review", "reviewer", "qa", "quality":
		return string(ResponseContractProfileStructuredAnalytical)
	case "builder", "implementer", "delivery", "operations", "operator", "coordinator":
		return string(ResponseContractProfileConciseDirect)
	case "support", "assistant", "guide":
		return string(ResponseContractProfileWarmSupportive)
	default:
		return ""
	}
}

package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func normalizeChatWorkspaceName(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func resolveChatWorkspaceTeamLabel(teamID, teamName string, home OrganizationHomePayload, manifests []*swarm.TeamManifest) string {
	if normalized := normalizeChatWorkspaceName(teamName); normalized != "" {
		return normalized
	}
	targetID := strings.TrimSpace(teamID)
	if targetID != "" {
		for _, department := range home.Departments {
			if strings.TrimSpace(department.ID) == targetID {
				if name := normalizeChatWorkspaceName(department.Name); name != "" {
					return name
				}
			}
		}
		for _, manifest := range manifests {
			if manifest != nil && strings.TrimSpace(manifest.ID) == targetID {
				if name := normalizeChatWorkspaceName(manifest.Name); name != "" {
					return name
				}
			}
		}
	}
	return ""
}

func summarizeDepartmentNames(departments []OrganizationDepartmentSummary) string {
	names := make([]string, 0, len(departments))
	for _, department := range departments {
		if name := normalizeChatWorkspaceName(department.Name); name != "" {
			names = append(names, name)
		}
	}
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		return names[0] + ", " + names[1] + fmt.Sprintf(", and %d more", len(names)-2)
	}
}

func (s *AdminServer) buildChatWorkspaceContext(organizationID, teamID, teamName string) string {
	organizationID = strings.TrimSpace(organizationID)
	teamID = strings.TrimSpace(teamID)
	teamName = normalizeChatWorkspaceName(teamName)
	if organizationID == "" && teamID == "" && teamName == "" {
		return ""
	}

	home, hasOrganization := OrganizationHomePayload{}, false
	if organizationID != "" {
		home, hasOrganization = s.organizationStore().Get(organizationID)
	}

	manifests := []*swarm.TeamManifest(nil)
	if s.Soma != nil {
		manifests = s.Soma.ListTeams()
	}

	currentTeamLabel := resolveChatWorkspaceTeamLabel(teamID, teamName, home, manifests)
	var lines []string
	lines = append(lines, "[WORKSPACE CONTEXT]")
	if hasOrganization {
		lines = append(lines, fmt.Sprintf("Organization: %s.", normalizeChatWorkspaceName(home.Name)))
		if purpose := normalizeChatWorkspaceName(home.Purpose); purpose != "" {
			lines = append(lines, fmt.Sprintf("Organization purpose: %s.", purpose))
		}
		if departmentSummary := summarizeDepartmentNames(home.Departments); departmentSummary != "" {
			lines = append(lines, fmt.Sprintf("Visible departments/teams in this organization: %s.", departmentSummary))
		}
	}
	if currentTeamLabel != "" {
		if teamID != "" {
			lines = append(lines, fmt.Sprintf("Current team focus: %s (id: %s).", currentTeamLabel, teamID))
		} else {
			lines = append(lines, fmt.Sprintf("Current team focus: %s.", currentTeamLabel))
		}
	} else if teamID != "" {
		lines = append(lines, fmt.Sprintf("Current team focus id: %s.", teamID))
	}
	lines = append(lines, "Treat phrases like 'the team', 'this team', or 'do you see it' as referring to the current team focus above when one is present. Use this workspace context before falling back to a broader runtime roster check.")
	return strings.Join(lines, "\n")
}

func prependChatWorkspaceContext(messages []chatRequestMessage, context string) []chatRequestMessage {
	trimmed := strings.TrimSpace(context)
	if trimmed == "" {
		return messages
	}
	out := make([]chatRequestMessage, 0, len(messages)+1)
	out = append(out, chatRequestMessage{
		Role:    "user",
		Content: trimmed,
	})
	out = append(out, messages...)
	return out
}

func recentSessionChatMessages(turns []conversations.ConversationTurn, limit int) []chatRequestMessage {
	if limit <= 0 || len(turns) == 0 {
		return nil
	}
	start := len(turns) - limit
	if start < 0 {
		start = 0
	}
	out := make([]chatRequestMessage, 0, len(turns)-start)
	for _, turn := range turns[start:] {
		role := strings.ToLower(strings.TrimSpace(turn.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		content := strings.TrimSpace(turn.Content)
		if content == "" {
			continue
		}
		out = append(out, chatRequestMessage{Role: role, Content: content})
	}
	return out
}

func mergePersistedSessionMessages(current []chatRequestMessage, prior []conversations.ConversationTurn) []chatRequestMessage {
	if len(current) != 1 || len(prior) == 0 {
		return current
	}
	recent := recentSessionChatMessages(prior, 20)
	if len(recent) == 0 {
		return current
	}
	out := make([]chatRequestMessage, 0, len(recent)+len(current))
	out = append(out, recent...)
	out = append(out, current...)
	return out
}

func validateOptionalChatSessionID(sessionID string) (string, bool) {
	trimmed := strings.TrimSpace(sessionID)
	if trimmed == "" {
		return "", true
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}

func logSomaConversationTurn(ctx context.Context, store protocol.ConversationLogger, sessionID string, index int, role string, content string, agentResult chatAgentResult) {
	if store == nil || sessionID == "" || strings.TrimSpace(content) == "" {
		return
	}
	if _, err := store.LogTurn(ctx, protocol.ConversationTurnData{
		SessionID:  sessionID,
		TenantID:   "default",
		AgentID:    "admin",
		TeamID:     "admin-core",
		TurnIndex:  index,
		Role:       role,
		Content:    content,
		ProviderID: agentResult.ProviderID,
		ModelUsed:  agentResult.ModelUsed,
	}); err != nil {
		log.Printf("[chat] conversation turn persistence failed: %v", err)
	}
}

func resolveChatAskClass(defaultAgentTarget string, isMutation bool, agentResult chatAgentResult) protocol.AskClass {
	if isMutation {
		return protocol.AskClassGovernedMutation
	}
	if len(agentResult.Artifacts) > 0 {
		return protocol.AskClassGovernedArtifact
	}
	if defaultAgentTarget == "specialist" || len(agentResult.Consultations) > 0 {
		return protocol.AskClassSpecialist
	}
	return protocol.AskClassDirectAnswer
}

func resolveChatAskContract(defaultAgentTarget string, isMutation bool, agentResult chatAgentResult) protocol.AskContract {
	class := resolveChatAskClass(defaultAgentTarget, isMutation, agentResult)
	contract, ok := protocol.AskContractForClass(class)
	if !ok {
		// Keep the direct/mutation fallback safe if the registry is incomplete.
		if isMutation {
			return protocol.AskContract{
				AskClass:             protocol.AskClassGovernedMutation,
				DefaultAgentTarget:   "soma",
				DefaultExecutionMode: protocol.ModeProposal,
				TemplateID:           protocol.TemplateChatToProposal,
			}
		}
		return protocol.AskContract{
			AskClass:             protocol.AskClassDirectAnswer,
			DefaultAgentTarget:   defaultAgentTarget,
			DefaultExecutionMode: protocol.ModeAnswer,
			TemplateID:           protocol.TemplateChatToAnswer,
		}
	}
	return contract
}

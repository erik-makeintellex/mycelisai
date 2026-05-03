package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

func summarizeChatTeamManifest(manifest *swarm.TeamManifest) string {
	if manifest == nil {
		return ""
	}
	name := normalizeChatWorkspaceName(manifest.Name)
	if name == "" {
		name = strings.TrimSpace(manifest.ID)
	}
	if name == "" {
		return ""
	}
	memberCount := len(manifest.Members)
	if memberCount <= 0 {
		return name
	}
	label := "members"
	if memberCount == 1 {
		label = "member"
	}
	return fmt.Sprintf("%s (%d %s)", name, memberCount, label)
}

func truncateList(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	out := append([]string{}, items[:limit]...)
	out = append(out, fmt.Sprintf("%d more", len(items)-limit))
	return out
}

func joinHumanList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

func (s *AdminServer) buildRuntimeStateAnswer(organizationID, teamID, teamName string) string {
	var lines []string
	lines = append(lines, "Current Mycelis runtime state:")

	runtimeStatus := []string{"frontend route available"}
	if s.NC != nil {
		runtimeStatus = append(runtimeStatus, "NATS connected")
	} else {
		runtimeStatus = append(runtimeStatus, "NATS unavailable")
	}

	availability := s.chatExecutionAvailability()
	if availability.Available {
		runtimeStatus = append(runtimeStatus, "chat engine available")
	} else if summary := strings.TrimSpace(availability.Summary); summary != "" {
		runtimeStatus = append(runtimeStatus, summary)
	}
	lines = append(lines, "- Runtime: "+joinHumanList(runtimeStatus)+".")

	organizations := s.organizationStore().List()
	if len(organizations) > 0 {
		orgNames := make([]string, 0, len(organizations))
		for _, org := range organizations {
			if name := normalizeChatWorkspaceName(org.Name); name != "" {
				orgNames = append(orgNames, name)
			}
		}
		sort.Strings(orgNames)
		if len(orgNames) > 0 {
			lines = append(lines, fmt.Sprintf("- Organizations loaded: %s.", joinHumanList(truncateList(orgNames, 4))))
		}
	}

	manifests := []*swarm.TeamManifest(nil)
	if s.Soma != nil {
		manifests = s.Soma.ListTeams()
	}
	if len(manifests) > 0 {
		teamSummaries := make([]string, 0, len(manifests))
		for _, manifest := range manifests {
			if summary := summarizeChatTeamManifest(manifest); summary != "" {
				teamSummaries = append(teamSummaries, summary)
			}
		}
		sort.Strings(teamSummaries)
		if len(teamSummaries) > 0 {
			lines = append(lines, fmt.Sprintf("- Active teams: %s.", joinHumanList(truncateList(teamSummaries, 6))))
		}
	} else {
		lines = append(lines, "- Active teams: Soma runtime is not currently exposing a team roster.")
	}

	home, hasOrganization := OrganizationHomePayload{}, false
	if trimmedOrgID := strings.TrimSpace(organizationID); trimmedOrgID != "" {
		home, hasOrganization = s.organizationStore().Get(trimmedOrgID)
	}
	if hasOrganization {
		orgLabel := normalizeChatWorkspaceName(home.Name)
		if orgLabel == "" {
			orgLabel = strings.TrimSpace(home.ID)
		}
		if orgLabel != "" {
			lines = append(lines, fmt.Sprintf("- Current organization focus: %s.", orgLabel))
		}
	}

	currentTeamLabel := resolveChatWorkspaceTeamLabel(teamID, teamName, home, manifests)
	if currentTeamLabel != "" {
		lines = append(lines, fmt.Sprintf("- Current team focus: %s.", currentTeamLabel))
	}

	lines = append(lines, "If you want, I can also list the team leads, group outputs, or the organization-specific setup next.")
	return strings.Join(lines, "\n")
}

func (s *AdminServer) respondRuntimeStateSummary(w http.ResponseWriter, r *http.Request, organizationID, teamID, teamName string) {
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer, "admin",
		"Runtime state summary",
		map[string]any{
			"actor":         "Soma",
			"user":          auditUserLabelFromRequest(r),
			"ask_class":     string(protocol.AskClassDirectAnswer),
			"action":        "answer_delivered",
			"result_status": "completed",
			"source_kind":   "system",
		},
	)

	chatPayload := protocol.ChatResponsePayload{
		Text:     s.buildRuntimeStateAnswer(organizationID, teamID, teamName),
		AskClass: protocol.AskClassDirectAnswer,
		Provenance: &protocol.AnswerProvenance{
			ResolvedIntent:  "answer",
			PermissionCheck: "pass",
			PolicyDecision:  "allow",
			AuditEventID:    auditEventID,
		},
	}
	payloadBytes, _ := json.Marshal(chatPayload)
	envelope := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: "admin",
			Timestamp:  time.Now(),
		},
		SignalType: protocol.SignalChatResponse,
		TrustScore: protocol.TrustScoreCognitive,
		Payload:    payloadBytes,
		TemplateID: protocol.TemplateChatToAnswer,
		Mode:       protocol.ModeAnswer,
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(envelope))
}

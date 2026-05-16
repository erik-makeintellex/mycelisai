package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) persistConfirmedActionVisibility(ctx context.Context, runID, auditID, auditUser string, scope *protocol.ScopeValidation, results []plannedToolExecutionResult) error {
	var errs []error
	if err := s.logConfirmedActionConversation(ctx, runID, auditUser, results); err != nil {
		errs = append(errs, err)
	}
	if err := s.ensureGroupsForCreatedTeams(ctx, auditID, auditUser, scope); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (s *AdminServer) logConfirmedActionConversation(ctx context.Context, runID, auditUser string, results []plannedToolExecutionResult) error {
	if s.Conversations == nil || strings.TrimSpace(runID) == "" {
		return nil
	}
	sessionID := runID
	_, err := s.Conversations.LogTurn(ctx, protocol.ConversationTurnData{
		RunID:     runID,
		SessionID: sessionID,
		TenantID:  "default",
		AgentID:   "Soma",
		TeamID:    "admin-core",
		TurnIndex: 0,
		Role:      "assistant",
		Content:   "Confirmed proposal execution started.",
	})
	if err != nil {
		return err
	}

	turnIndex := 1
	for _, result := range results {
		toolName := strings.TrimSpace(result.Name)
		if toolName == "" {
			toolName = "tool"
		}
		teamID := confirmedActionTeamID(result.Arguments)
		callID, err := s.Conversations.LogTurn(ctx, protocol.ConversationTurnData{
			RunID:     runID,
			SessionID: sessionID,
			TenantID:  "default",
			AgentID:   "Soma",
			TeamID:    firstNonEmptyString(teamID, "admin-core"),
			TurnIndex: turnIndex,
			Role:      "tool_call",
			Content:   "call " + toolName,
			ToolName:  toolName,
			ToolArgs:  result.Arguments,
		})
		if err != nil {
			return err
		}
		turnIndex++
		if _, err := s.Conversations.LogTurn(ctx, protocol.ConversationTurnData{
			RunID:        runID,
			SessionID:    sessionID,
			TenantID:     "default",
			AgentID:      "Soma",
			TeamID:       firstNonEmptyString(teamID, "admin-core"),
			TurnIndex:    turnIndex,
			Role:         "tool_result",
			Content:      firstNonEmptyString(result.Output, toolName+" completed."),
			ToolName:     toolName,
			ParentTurnID: callID,
		}); err != nil {
			return err
		}
		turnIndex++
	}
	return nil
}

func (s *AdminServer) ensureGroupsForCreatedTeams(ctx context.Context, auditID, auditUser string, scope *protocol.ScopeValidation) error {
	if s.getDB() == nil || scope == nil {
		return nil
	}
	for _, planned := range scope.PlannedToolCalls {
		if strings.TrimSpace(planned.Name) != "create_team" {
			continue
		}
		if err := s.ensureGroupForCreatedTeam(ctx, auditID, auditUser, planned.Arguments); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminServer) ensureGroupForCreatedTeam(ctx context.Context, auditID, auditUser string, args map[string]any) error {
	merged := mergedTeamArgs(args)
	teamID := confirmedActionTeamID(merged)
	if teamID == "" {
		return nil
	}
	name := firstNonEmptyString(merged["name"], teamID)
	if existing, err := s.getGroupByNameDB(ctx, name); err != nil {
		return err
	} else if existing != nil {
		return nil
	}

	workMode := firstNonEmptyString(merged["work_mode"], "propose_only")
	if _, ok := validGroupWorkModes[workMode]; !ok {
		workMode = "propose_only"
	}
	role := firstNonEmptyString(merged["role"], "team")
	allowed := confirmedActionStringSlice(merged["allowed_capabilities"])
	if len(allowed) == 0 {
		allowed = confirmedActionStringSlice(merged["capabilities"])
	}
	if len(allowed) == 0 {
		allowed = confirmedActionStringSlice(merged["tools"])
	}
	if len(allowed) == 0 {
		allowed = []string{"team.coordinate", "artifact.review", "broadcast"}
	}

	group := CollaborationGroup{
		ID:                  uuid.NewString(),
		TenantID:            "default",
		Name:                name,
		GoalStatement:       firstNonEmptyString(merged["goal"], merged["task"], merged["description"], "Runtime team "+name+" created through confirmed Soma proposal."),
		WorkMode:            workMode,
		AllowedCapabilities: allowed,
		MemberUserIDs:       []string{},
		TeamIDs:             []string{teamID},
		CoordinatorProfile:  firstNonEmptyString(merged["coordinator_profile"], role+" lead"),
		ApprovalPolicyRef:   firstNonEmptyString(merged["approval_policy_ref"], "confirmed-chat-proposal"),
		Status:              groupStatusActive,
		CreatedBy:           firstNonEmptyString(auditUser, "system"),
		CreatedAuditEventID: auditID,
		UpdatedAuditEventID: auditID,
	}
	return s.insertGroupDB(ctx, &group)
}

func mergedTeamArgs(args map[string]any) map[string]any {
	merged := map[string]any{}
	for k, v := range args {
		merged[k] = v
	}
	if manifest, ok := args["manifest"].(map[string]any); ok {
		for k, v := range manifest {
			if _, exists := merged[k]; !exists {
				merged[k] = v
			}
		}
	}
	return merged
}

func confirmedActionTeamID(args map[string]any) string {
	merged := mergedTeamArgs(args)
	return firstNonEmptyString(merged["team_id"], merged["id"], merged["team_name"])
}

func confirmedActionStringSlice(raw any) []string {
	switch typed := raw.(type) {
	case []string:
		return normalizeStringSlice(typed)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, firstNonEmptyString(item))
		}
		return normalizeStringSlice(items)
	case string:
		if strings.TrimSpace(typed) == "" {
			return []string{}
		}
		var decoded []string
		if strings.HasPrefix(strings.TrimSpace(typed), "[") && json.Unmarshal([]byte(typed), &decoded) == nil {
			return normalizeStringSlice(decoded)
		}
		return normalizeStringSlice(strings.Split(typed, ","))
	default:
		return []string{}
	}
}

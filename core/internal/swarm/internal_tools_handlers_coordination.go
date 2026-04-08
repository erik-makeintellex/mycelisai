package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
)

func (r *InternalToolRegistry) handleConsultCouncil(ctx context.Context, args map[string]any) (string, error) {
	member := normalizeCouncilMember(pickFirstString(args, "member", "agent", "target"))
	question := pickFirstString(args, "question", "query", "prompt", "message")
	if member == "" || question == "" {
		return "", fmt.Errorf("consult_council requires 'member' and 'question'")
	}
	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot consult council")
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	msg, err := r.nc.RequestWithContext(reqCtx, fmt.Sprintf(protocol.TopicCouncilRequestFmt, member), []byte(question))
	if err != nil {
		return "", fmt.Errorf("council member %s did not respond: %w", member, err)
	}
	var result struct {
		Text      string                     `json:"text"`
		Artifacts []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &result); err == nil && result.Text != "" {
		if len(result.Artifacts) > 0 {
			data, _ := json.Marshal(map[string]any{"message": result.Text, "artifacts": result.Artifacts})
			return string(data), nil
		}
		return result.Text, nil
	}
	return string(msg.Data), nil
}

func (r *InternalToolRegistry) handleDelegateTask(ctx context.Context, args map[string]any) (string, error) {
	teamID, ask, err := normalizeDelegateTaskArgs(args)
	if err != nil {
		return "", err
	}
	if ask.IsZero() {
		return "", fmt.Errorf("delegate_task requires 'team_id' and 'task'")
	}
	if teamID == "" {
		teamID, err = r.resolveDelegationTeam(ask)
		if err != nil {
			return "", err
		}
	}
	if hint, ok := args["hint"].(map[string]any); ok {
		log.Printf("DelegationHint [%s]: confidence=%.2f urgency=%v complexity=%v risk=%v", teamID, hint["confidence"], hint["urgency"], hint["complexity"], hint["risk"])
	}
	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot delegate task")
	}

	task, err := json.Marshal(ask)
	if err != nil {
		return "", fmt.Errorf("failed to marshal delegated task payload: %w", err)
	}
	payload, err := r.wrapGovernedSignalPayload(ctx, "internal_tool.delegate_task", teamID, protocol.PayloadKindCommand, task)
	if err != nil {
		return "", fmt.Errorf("failed to wrap delegated task payload: %w", err)
	}
	if err := r.nc.Publish(fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID), payload); err != nil {
		return "", fmt.Errorf("failed to publish task to team %s: %w", teamID, err)
	}
	return fmt.Sprintf("Task delegated to team %s.", teamID), nil
}

func (r *InternalToolRegistry) resolveDelegationTeam(ask protocol.TeamAsk) (string, error) {
	if r.somaRef == nil {
		return "", fmt.Errorf("delegate_task requires 'team_id' unless Soma can resolve a routed team")
	}

	normalized := ask.Normalize()
	wantKind := string(normalized.AskKind)
	wantLane := string(normalized.LaneRole)
	if wantKind == "" || wantLane == "" {
		return "", fmt.Errorf("delegate_task requires explicit routing metadata when 'team_id' is omitted")
	}

	matches := make([]string, 0, 2)
	for _, manifest := range r.somaRef.ListTeams() {
		if manifest == nil || len(manifest.AskRouting) == 0 {
			continue
		}
		if strings.TrimSpace(manifest.AskRouting[wantKind]) == wantLane {
			matches = append(matches, manifest.ID)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", fmt.Errorf("delegate_task could not resolve a team for ask_kind=%s lane_role=%s", wantKind, wantLane)
	default:
		return "", fmt.Errorf("delegate_task routing is ambiguous for ask_kind=%s lane_role=%s", wantKind, wantLane)
	}
}

func (r *InternalToolRegistry) handleCreateTeam(_ context.Context, args map[string]any) (string, error) {
	if r.somaRef == nil {
		return "", fmt.Errorf("Soma not available — cannot create team")
	}
	manifest := buildRuntimeTeamManifest(args)
	if manifest == nil {
		return "", fmt.Errorf("create_team requires 'team_id'")
	}
	for _, m := range r.somaRef.ListTeams() {
		if m != nil && m.ID == manifest.ID {
			out, _ := json.Marshal(map[string]any{"status": "already_exists", "team_id": manifest.ID})
			return string(out), nil
		}
	}
	if err := r.somaRef.SpawnTeam(manifest); err != nil {
		return "", fmt.Errorf("create_team failed: %w", err)
	}
	out, _ := json.Marshal(map[string]any{"status": "created", "team_id": manifest.ID, "name": manifest.Name})
	return string(out), nil
}

func normalizeDelegateTaskArgs(args map[string]any) (teamID string, ask protocol.TeamAsk, err error) {
	teamID = pickFirstString(args, "team_id", "teamId", "target_team")
	if teamID == "" {
		if teamMap, ok := args["team"].(map[string]any); ok {
			teamID = pickFirstString(teamMap, "id", "team_id", "name")
		} else {
			teamID = stringValue(args["team"])
		}
	}
	if askRaw, ok := args["ask"].(map[string]any); ok {
		ask = protocol.TeamAskFromMap(askRaw)
		return teamID, ask.Normalize(), nil
	}
	switch t := args["task"].(type) {
	case string:
		ask = protocol.TeamAsk{Goal: strings.TrimSpace(t)}
	case map[string]any, []any:
		if taskMap, ok := t.(map[string]any); ok {
			ask = protocol.TeamAskFromMap(taskMap)
		} else {
			ask = protocol.TeamAsk{Goal: string(mustJSON(t))}
		}
	}
	if !ask.IsZero() {
		return teamID, ask.Normalize(), nil
	}

	ask = protocol.TeamAsk{
		AskKind:  protocol.TeamAskKind(stringValue(args["ask_kind"])),
		LaneRole: protocol.TeamLaneRole(stringValue(args["lane_role"])),
		Goal: firstNonEmptyString(
			stringValue(args["goal"]),
			stringValue(args["intent"]),
			stringValue(args["message"]),
			stringValue(args["operation"]),
		),
		OwnedScope:           stringSlice(args["owned_scope"]),
		Constraints:          stringSlice(args["constraints"]),
		RequiredCapabilities: stringSlice(args["required_capabilities"]),
		ApprovalPosture:      protocol.ApprovalPosture(stringValue(args["approval_posture"])),
		ExitCriteria:         stringSlice(args["exit_criteria"]),
		EvidenceRequired:     stringSlice(args["evidence_required"]),
	}
	if ctxRaw, ok := args["context"]; ok {
		if ctxMap, ok := ctxRaw.(map[string]any); ok {
			ask.Context = ctxMap
		}
	}
	return teamID, ask.Normalize(), nil
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (r *InternalToolRegistry) handleSearchMemory(ctx context.Context, args map[string]any) (string, error) {
	query := stringValue(args["query"])
	if query == "" {
		return "", fmt.Errorf("search_memory requires 'query'")
	}
	if r.brain == nil || r.mem == nil {
		return "Memory search unavailable — cognitive engine or memory service offline.", nil
	}
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	scope := resolveMemoryScope(ctx, args)
	searchTypes := stringSlice(args["types"])
	if singleType := stringValue(args["type"]); singleType != "" {
		searchTypes = append(searchTypes, singleType)
	}
	vec, err := r.brain.Embed(ctx, query, "")
	if err != nil {
		return "Embedding failed — no embed provider available.", nil
	}
	results, err := r.mem.SemanticSearchWithOptions(ctx, vec, memory.SemanticSearchOptions{
		Limit:               limit,
		TenantID:            scope.TenantID,
		TeamID:              scope.TeamID,
		AgentID:             scope.AgentID,
		RunID:               scope.RunID,
		Visibility:          normalizedVisibility(stringValue(args["visibility"])),
		Types:               dedupeStringValues(searchTypes),
		AllowGlobal:         true,
		AllowLegacyUnscoped: scope.TeamID == "" && scope.AgentID == "",
	})
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err), nil
	}
	return mustJSON(results), nil
}

func (r *InternalToolRegistry) handleListTeams(_ context.Context, _ map[string]any) (string, error) {
	if r.somaRef == nil {
		return "Soma not available — cannot list teams.", nil
	}
	type teamSummary struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Members int    `json:"members"`
	}
	manifests := r.somaRef.ListTeams()
	summaries := make([]teamSummary, 0, len(manifests))
	for _, m := range manifests {
		summaries = append(summaries, teamSummary{ID: m.ID, Name: m.Name, Type: string(m.Type), Members: len(m.Members)})
	}
	return mustJSON(summaries), nil
}

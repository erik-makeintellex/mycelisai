package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	const consultTimeout = 30 * time.Second
	reqCtx, cancel := context.WithTimeout(ctx, consultTimeout)
	defer cancel()
	// MEM-LANES-2: the council member runs under this turn's user, so its
	// memory reads and writes stay that user's instead of becoming no-user rows.
	msg, err := requestWithRecallTurn(reqCtx, r.nc, recallAccessFromContext(ctx), fmt.Sprintf(protocol.TopicCouncilRequestFmt, member), []byte(question), consultTimeout)
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
	ask = delegateAskAuthority(ctx, ask) // F16c: only Core's confirmed dispatch keeps claim keys
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
	// Core NATS publish is fire-and-forget: it proves the command reached the
	// bus client, not that the team received or accepted it.
	return fmt.Sprintf("Task queued for team %s. Delivery and acceptance are not yet confirmed; the team's receipt and status report them.", teamID), nil
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

func (r *InternalToolRegistry) handleCreateTeam(ctx context.Context, args map[string]any) (string, error) {
	if r.somaRef == nil {
		return "", fmt.Errorf("Soma not available — cannot create team")
	}
	candidate := buildRuntimeTeamManifest(args)
	if candidate == nil {
		return "", fmt.Errorf("create_team requires 'team_id'")
	}
	if IsReservedTeamID(candidate.ID) { // refused whether or not the Core team is loaded
		return "", fmt.Errorf("create_team %q: %w", candidate.ID, ErrReservedTeamID)
	}
	for _, m := range r.somaRef.ListTeams() {
		if m != nil && m.ID == candidate.ID {
			workspaceFolder, err := ensureRuntimeTeamWorkspace(candidate.ID)
			if err != nil {
				return "", fmt.Errorf("repair team workspace: %w", err)
			}
			out, _ := json.Marshal(map[string]any{"status": "already_exists", "team_id": candidate.ID, "workspace_folder": workspaceFolder})
			return string(out), nil
		}
	}
	if err := r.hydrateCreateTeamProfiles(ctx, args); err != nil {
		return "", err
	}
	manifest := buildRuntimeTeamManifest(args)
	if err := childToolsWithinCaller(ctx, manifest); err != nil {
		return "", err
	}
	if err := r.somaRef.SpawnTeamContext(ctx, manifest); err != nil {
		return "", fmt.Errorf("create_team failed: %w", err)
	}
	workspaceFolder, err := ensureRuntimeTeamWorkspace(manifest.ID)
	if err != nil {
		r.somaRef.StopTeam(manifest.ID)
		return "", fmt.Errorf("create team workspace: %w", err)
	}
	out, _ := json.Marshal(map[string]any{"status": "created", "team_id": manifest.ID, "name": manifest.Name, "workspace_folder": workspaceFolder})
	return string(out), nil
}

func ensureRuntimeTeamWorkspace(teamID string) (string, error) {
	trimmed := strings.TrimSpace(teamID)
	if trimmed == "" || trimmed == "." || trimmed == ".." || strings.ContainsAny(trimmed, `/\\`) {
		return "", fmt.Errorf("invalid team_id for workspace isolation")
	}
	relativeRoot := "groups/" + trimmed
	root, err := validateToolPath(relativeRoot)
	if err != nil {
		return "", err
	}
	for _, folder := range []string{"planning", "source", "generated"} {
		if err := os.MkdirAll(filepath.Join(root, folder), 0o755); err != nil {
			return "", fmt.Errorf("create %s folder: %w", folder, err)
		}
	}
	return relativeRoot, nil
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
		return teamID, askWithDelegateContext(args, ask), nil
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
		return teamID, askWithDelegateContext(args, ask), nil
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
	if r.mem == nil {
		return "", fmt.Errorf("search_memory unavailable: memory service offline")
	}
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	reader, err := recallAccessFromContext(ctx).governedReader() // SRU: the requesting user's read scope
	if err != nil {
		return "", fmt.Errorf("search_memory unavailable: %w", err)
	}
	scope := resolveMemoryScope(ctx, args)
	searchTypes := stringSlice(args["types"])
	if singleType := stringValue(args["type"]); singleType != "" {
		searchTypes = append(searchTypes, singleType)
	}
	// Semantic when an embedding engine works; PostgreSQL keyword ranking otherwise.
	results, mode, err := r.mem.RecallGoverned(ctx, r.brain, query, memory.SemanticSearchOptions{
		Limit:               limit,
		TenantID:            scope.TenantID,
		TeamID:              scope.TeamID,
		AgentID:             scope.AgentID,
		RunID:               scope.RunID,
		Visibility:          normalizedVisibility(stringValue(args["visibility"])),
		Types:               dedupeStringValues(searchTypes),
		AllowGlobal:         true,
		AllowLegacyUnscoped: scope.TeamID == "" && scope.AgentID == "",
		GoalSets:            goalSetArg(args["goal_set"]),
		Reader:              reader,
	})
	if err != nil {
		return "", fmt.Errorf("search_memory failed: %w", err)
	}
	return mustJSON(map[string]any{"retrieval_mode": mode, "results": results}), nil
}

func (r *InternalToolRegistry) handleListTeams(_ context.Context, _ map[string]any) (string, error) {
	if r.somaRef == nil {
		return "", fmt.Errorf("list_teams unavailable: Soma is not available")
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

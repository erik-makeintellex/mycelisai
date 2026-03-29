package swarm

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
)

func (r *InternalToolRegistry) handleListMissions(ctx context.Context, _ map[string]any) (string, error) {
	if r.db == nil {
		return "Database not available — cannot list missions.", nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.directive, COALESCE(m.status, 'active'),
		       COUNT(DISTINCT t.id), COUNT(DISTINCT sm.id)
		FROM missions m
		LEFT JOIN teams t ON t.mission_id = m.id
		LEFT JOIN service_manifests sm ON sm.team_id = t.id
		GROUP BY m.id, m.directive, m.status
		ORDER BY m.created_at DESC LIMIT 20
	`)
	if err != nil {
		return fmt.Sprintf("Query failed: %v", err), nil
	}
	defer rows.Close()
	type missionRow struct {
		ID     string `json:"id"`
		Intent string `json:"intent"`
		Status string `json:"status"`
		Teams  int    `json:"teams"`
		Agents int    `json:"agents"`
	}
	var missions []missionRow
	for rows.Next() {
		var m missionRow
		if err := rows.Scan(&m.ID, &m.Intent, &m.Status, &m.Teams, &m.Agents); err == nil {
			missions = append(missions, m)
		}
	}
	if missions == nil {
		missions = []missionRow{}
	}
	return mustJSON(missions), nil
}

func (r *InternalToolRegistry) handleGetSystemStatus(_ context.Context, _ map[string]any) (string, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	tokenRate := 0.0
	if r.brain != nil {
		tokenRate = r.brain.TokenRate()
	}
	return mustJSON(map[string]any{
		"goroutines":     runtime.NumGoroutine(),
		"heap_alloc_mb":  float64(memStats.HeapAlloc) / 1024 / 1024,
		"sys_mem_mb":     float64(memStats.Sys) / 1024 / 1024,
		"llm_tokens_sec": tokenRate,
		"timestamp":      time.Now().Format(time.RFC3339),
	}), nil
}

func (r *InternalToolRegistry) handleListAvailableTools(ctx context.Context, _ map[string]any) (string, error) {
	descs := r.ListDescriptions()
	if r.db != nil {
		queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		rows, err := r.db.QueryContext(queryCtx, `
			SELECT t.name, COALESCE(t.description, ''), s.name
			FROM mcp_tools t
			JOIN mcp_servers s ON s.id = t.server_id
			ORDER BY s.name, t.name
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var toolName, toolDesc, serverName string
				if err := rows.Scan(&toolName, &toolDesc, &serverName); err == nil {
					if toolDesc == "" {
						toolDesc = fmt.Sprintf("MCP tool via %s", serverName)
					} else {
						toolDesc = fmt.Sprintf("%s (MCP via %s)", toolDesc, serverName)
					}
					descs[toolName] = toolDesc
				}
			}
		}
	}
	return mustJSON(descs), nil
}

func (r *InternalToolRegistry) handleGenerateBlueprint(ctx context.Context, args map[string]any) (string, error) {
	intent := stringValue(args["intent"])
	if intent == "" {
		return "", fmt.Errorf("generate_blueprint requires 'intent'")
	}
	if r.architect == nil {
		return "", fmt.Errorf("Meta-Architect not available — cognitive engine offline")
	}
	bp, err := r.architect.GenerateBlueprint(ctx, intent)
	if err != nil {
		return "", fmt.Errorf("blueprint generation failed: %w", err)
	}
	return mustJSON(bp), nil
}

func (r *InternalToolRegistry) handleResearchForBlueprint(ctx context.Context, args map[string]any) (string, error) {
	intent := stringValue(args["intent"])
	if intent == "" {
		return "", fmt.Errorf("research_for_blueprint requires 'intent'")
	}
	var sb strings.Builder
	sb.WriteString("# Blueprint Research Report\n\n")
	memories, memoriesErr := r.handleRecall(ctx, map[string]any{"query": intent, "limit": float64(3)})
	appendResearchSection(&sb, "Past Mission Context", memories, memoriesErr, "No relevant past missions found.")
	catalogue, catalogueErr := r.handleListCatalogue(ctx, nil)
	appendResearchSection(&sb, "Available Agent Templates", catalogue, catalogueErr, "Agent catalogue not available.")
	tools, toolsErr := r.handleListAvailableTools(ctx, nil)
	appendResearchSection(&sb, "Tool Inventory", tools, toolsErr, "Tool listing not available.")
	if missions, err := r.handleListMissions(ctx, nil); err == nil && missions != "" {
		sb.WriteString("## Active Missions\n" + missions + "\n\n")
	}
	if recipes, err := r.handleRecallInceptionRecipes(ctx, map[string]any{"query": intent, "limit": float64(3)}); err == nil && recipes != "" && recipes != "[]" {
		sb.WriteString("## Inception Recipes (Proven Patterns)\n" + recipes + "\n\n")
	}
	sb.WriteString("Use this research to inform the blueprint you generate with generate_blueprint.\n")
	return sb.String(), nil
}

func appendResearchSection(sb *strings.Builder, heading string, content string, err error, fallback string) {
	if err == nil && content != "" {
		sb.WriteString("## " + heading + "\n" + content + "\n\n")
		return
	}
	sb.WriteString("## " + heading + "\n" + fallback + "\n\n")
}

func (r *InternalToolRegistry) handleListCatalogue(ctx context.Context, _ map[string]any) (string, error) {
	if r.catalogue == nil {
		return "Agent catalogue not available.", nil
	}
	agents, err := r.catalogue.List(ctx)
	if err != nil {
		return fmt.Sprintf("Catalogue query failed: %v", err), nil
	}
	return mustJSON(agents), nil
}

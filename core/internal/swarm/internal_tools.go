package swarm

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// InternalTool describes a built-in tool available to agents without an
// external MCP server. The Handler function is invoked during the ReAct loop.
type InternalTool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     func(ctx context.Context, args map[string]any) (string, error)
}

// InternalToolRegistry holds all built-in tools and the dependencies they need.
type InternalToolRegistry struct {
	tools     map[string]*InternalTool
	nc        *nats.Conn
	brain     *cognitive.Router
	mem       *memory.Service
	architect *cognitive.MetaArchitect
	catalogue *catalogue.Service
	db        *sql.DB
	// somaRef is set after Soma is constructed — allows list_teams etc.
	somaRef *Soma
}

// InternalToolDeps bundles all optional dependencies for the internal tools.
type InternalToolDeps struct {
	NC        *nats.Conn
	Brain     *cognitive.Router
	Mem       *memory.Service
	Architect *cognitive.MetaArchitect
	Catalogue *catalogue.Service
	DB        *sql.DB
}

// NewInternalToolRegistry creates and populates the built-in tool set.
func NewInternalToolRegistry(deps InternalToolDeps) *InternalToolRegistry {
	r := &InternalToolRegistry{
		tools:     make(map[string]*InternalTool),
		nc:        deps.NC,
		brain:     deps.Brain,
		mem:       deps.Mem,
		architect: deps.Architect,
		catalogue: deps.Catalogue,
		db:        deps.DB,
	}
	r.registerAll()
	return r
}

// SetSoma wires the Soma reference after construction (avoids circular init).
func (r *InternalToolRegistry) SetSoma(s *Soma) {
	r.somaRef = s
}

// Get returns a tool by name, or nil if not found.
func (r *InternalToolRegistry) Get(name string) *InternalTool {
	return r.tools[name]
}

// Has returns true if the named tool is an internal tool.
func (r *InternalToolRegistry) Has(name string) bool {
	_, ok := r.tools[name]
	return ok
}

// ListNames returns all registered internal tool names.
func (r *InternalToolRegistry) ListNames() []string {
	names := make([]string, 0, len(r.tools))
	for k := range r.tools {
		names = append(names, k)
	}
	return names
}

// ListDescriptions returns tool name→description pairs for prompt injection.
func (r *InternalToolRegistry) ListDescriptions() map[string]string {
	m := make(map[string]string, len(r.tools))
	for k, v := range r.tools {
		m[k] = v.Description
	}
	return m
}

// registerAll registers every built-in tool.
func (r *InternalToolRegistry) registerAll() {
	r.tools["consult_council"] = &InternalTool{
		Name:        "consult_council",
		Description: "Send a question to a specific council member (architect, coder, creative, sentry) and get their response. Use for specialist expertise.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"member":   map[string]any{"type": "string", "description": "Council member ID: council-architect, council-coder, council-creative, or council-sentry"},
				"question": map[string]any{"type": "string", "description": "The question or task for the council member"},
			},
			"required": []string{"member", "question"},
		},
		Handler: r.handleConsultCouncil,
	}

	r.tools["delegate_task"] = &InternalTool{
		Name:        "delegate_task",
		Description: "Publish a task to a specific team's trigger topic for processing.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"team_id": map[string]any{"type": "string", "description": "The target team ID"},
				"task":    map[string]any{"type": "string", "description": "The task description to send"},
			},
			"required": []string{"team_id", "task"},
		},
		Handler: r.handleDelegateTask,
	}

	r.tools["search_memory"] = &InternalTool{
		Name:        "search_memory",
		Description: "Semantic vector search over SitReps (situation reports) from past operations.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "The search query text"},
				"limit": map[string]any{"type": "integer", "description": "Max results (default 5)"},
			},
			"required": []string{"query"},
		},
		Handler: r.handleSearchMemory,
	}

	r.tools["list_teams"] = &InternalTool{
		Name:        "list_teams",
		Description: "Returns the active team roster with member counts.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleListTeams,
	}

	r.tools["list_missions"] = &InternalTool{
		Name:        "list_missions",
		Description: "Returns active missions with team/agent counts.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleListMissions,
	}

	r.tools["get_system_status"] = &InternalTool{
		Name:        "get_system_status",
		Description: "Returns system telemetry: goroutines, heap memory, LLM tokens/sec.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleGetSystemStatus,
	}

	r.tools["list_available_tools"] = &InternalTool{
		Name:        "list_available_tools",
		Description: "Returns all available tools (internal + MCP) with descriptions.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleListAvailableTools,
	}

	r.tools["generate_blueprint"] = &InternalTool{
		Name:        "generate_blueprint",
		Description: "Invoke the Meta-Architect to generate a MissionBlueprint from natural language intent.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"intent": map[string]any{"type": "string", "description": "The natural language intent to decompose"},
			},
			"required": []string{"intent"},
		},
		Handler: r.handleGenerateBlueprint,
	}

	r.tools["list_catalogue"] = &InternalTool{
		Name:        "list_catalogue",
		Description: "Returns agent catalogue templates (reusable agent blueprints).",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler:     r.handleListCatalogue,
	}

	r.tools["remember"] = &InternalTool{
		Name:        "remember",
		Description: "Store a learned fact, user preference, or goal into persistent memory. Stored both as structured RDBMS record and as a vector embedding for semantic recall. Use this whenever you learn something important about the user, their goals, or the project.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"category": map[string]any{"type": "string", "description": "Category: user_preference, goal, fact, decision, lesson_learned"},
				"content":  map[string]any{"type": "string", "description": "The information to remember"},
				"context":  map[string]any{"type": "string", "description": "Optional context about when/why this was learned"},
			},
			"required": []string{"category", "content"},
		},
		Handler: r.handleRemember,
	}

	r.tools["recall"] = &InternalTool{
		Name:        "recall",
		Description: "Recall stored memories by semantic similarity. Searches both vector embeddings and structured records. Use this to retrieve past learnings, user preferences, goals, and decisions.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":    map[string]any{"type": "string", "description": "What to recall — natural language query"},
				"category": map[string]any{"type": "string", "description": "Optional filter: user_preference, goal, fact, decision, lesson_learned"},
				"limit":    map[string]any{"type": "integer", "description": "Max results (default 5)"},
			},
			"required": []string{"query"},
		},
		Handler: r.handleRecall,
	}

	r.tools["store_artifact"] = &InternalTool{
		Name:        "store_artifact",
		Description: `Persist an agent output as a typed artifact. For type="chart", content MUST be a JSON chart spec: {"version":"1", "chart_type":"bar|line|area|dot|geo|table", "title":"...", "data":[{...}], "x":"col", "y":"col"}. Max 2000 data rows.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"type":     map[string]any{"type": "string", "description": "Artifact type: code, document, image, data, file, chart"},
				"title":    map[string]any{"type": "string", "description": "Human-readable title"},
				"content":  map[string]any{"type": "string", "description": "The artifact content. For chart type: JSON chart spec with version, chart_type, data[], x, y fields."},
				"metadata": map[string]any{"type": "object", "description": "Optional key-value metadata"},
			},
			"required": []string{"type", "title", "content"},
		},
		Handler: r.handleStoreArtifact,
	}

	r.tools["publish_signal"] = &InternalTool{
		Name:        "publish_signal",
		Description: "Publish a message to any NATS topic in the swarm. Use for broadcasting signals, triggering teams, or inter-agent communication.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"subject": map[string]any{"type": "string", "description": "NATS subject (e.g. swarm.team.council-core.internal.trigger, swarm.global.broadcast)"},
				"message": map[string]any{"type": "string", "description": "The message payload to publish"},
			},
			"required": []string{"subject", "message"},
		},
		Handler: r.handlePublishSignal,
	}

	r.tools["broadcast"] = &InternalTool{
		Name:        "broadcast",
		Description: "Send a message to ALL active teams in the swarm. Every team's agents will process it and respond. Use when the user says 'broadcast', 'tell everyone', or 'announce to all teams'.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{"type": "string", "description": "The message to broadcast to all teams"},
			},
			"required": []string{"message"},
		},
		Handler: r.handleBroadcast,
	}

	r.tools["read_signals"] = &InternalTool{
		Name:        "read_signals",
		Description: "Subscribe to a NATS topic pattern and collect messages for a brief window. Use to sense what's happening on a bus channel.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"subject":     map[string]any{"type": "string", "description": "NATS subject or wildcard (e.g. swarm.team.*.signal.status, swarm.global.>)"},
				"duration_ms": map[string]any{"type": "integer", "description": "How long to listen in milliseconds (default 3000, max 10000)"},
				"max_msgs":    map[string]any{"type": "integer", "description": "Max messages to collect (default 20)"},
			},
			"required": []string{"subject"},
		},
		Handler: r.handleReadSignals,
	}

	r.tools["read_file"] = &InternalTool{
		Name:        "read_file",
		Description: "Read the contents of a file from the local filesystem. Use for inspecting configs, logs, artifacts, or any file the organism can access.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "Absolute or relative file path to read"},
			},
			"required": []string{"path"},
		},
		Handler: r.handleReadFile,
	}

	r.tools["write_file"] = &InternalTool{
		Name:        "write_file",
		Description: "Write content to a file on the local filesystem. Creates parent directories if needed. Use for generating configs, code, reports, or any file output.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]any{"type": "string", "description": "File path to write to"},
				"content": map[string]any{"type": "string", "description": "The file content to write"},
			},
			"required": []string{"path", "content"},
		},
		Handler: r.handleWriteFile,
	}

	r.tools["generate_image"] = &InternalTool{
		Name:        "generate_image",
		Description: "Generate an image from a text prompt using the local Diffusers media engine (Stable Diffusion XL). Returns generation status and metadata. Use for illustrations, concept art, diagrams, or any visual content.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{"type": "string", "description": "Text description of the image to generate"},
				"size":   map[string]any{"type": "string", "description": "Image dimensions (default: 1024x1024). Options: 512x512, 1024x1024, 1024x768"},
			},
			"required": []string{"prompt"},
		},
		Handler: r.handleGenerateImage,
	}

	r.tools["research_for_blueprint"] = &InternalTool{
		Name:        "research_for_blueprint",
		Description: "Conduct multi-source research before generating a mission blueprint. Queries past missions (recall), agent catalogue (templates), and tool inventory (internal + MCP). Returns an enriched context block to feed into generate_blueprint.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"intent": map[string]any{"type": "string", "description": "The user's natural language intent to research"},
			},
			"required": []string{"intent"},
		},
		Handler: r.handleResearchForBlueprint,
	}

	r.tools["summarize_conversation"] = &InternalTool{
		Name:        "summarize_conversation",
		Description: "Summarize recent conversation into persistent memory. Extracts key topics, user preferences, personality notes, and data references. The summary is embedded into pgvector for future semantic recall. Call this when you want to checkpoint important conversation context.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"messages": map[string]any{"type": "string", "description": "The conversation text to summarize (last 15-20 messages)"},
			},
			"required": []string{"messages"},
		},
		Handler: r.handleSummarizeConversation,
	}
}

// ── Tool Handlers ──────────────────────────────────────────────

func (r *InternalToolRegistry) handleConsultCouncil(ctx context.Context, args map[string]any) (string, error) {
	member, _ := args["member"].(string)
	question, _ := args["question"].(string)
	if member == "" || question == "" {
		return "", fmt.Errorf("consult_council requires 'member' and 'question'")
	}

	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot consult council")
	}

	subject := fmt.Sprintf(protocol.TopicCouncilRequestFmt, member)
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	msg, err := r.nc.RequestWithContext(reqCtx, subject, []byte(question))
	if err != nil {
		return "", fmt.Errorf("council member %s did not respond: %w", member, err)
	}

	// Agent returns structured JSON (ProcessResult). Extract just the text.
	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(msg.Data, &result); err == nil && result.Text != "" {
		return result.Text, nil
	}
	return string(msg.Data), nil
}

func (r *InternalToolRegistry) handleDelegateTask(ctx context.Context, args map[string]any) (string, error) {
	teamID, _ := args["team_id"].(string)
	task, _ := args["task"].(string)
	if teamID == "" || task == "" {
		return "", fmt.Errorf("delegate_task requires 'team_id' and 'task'")
	}

	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot delegate task")
	}

	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID)
	if err := r.nc.Publish(subject, []byte(task)); err != nil {
		return "", fmt.Errorf("failed to publish task to team %s: %w", teamID, err)
	}
	r.nc.Flush()

	return fmt.Sprintf("Task delegated to team %s.", teamID), nil
}

func (r *InternalToolRegistry) handleSearchMemory(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
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

	vec, err := r.brain.Embed(ctx, query, "")
	if err != nil {
		return "Embedding failed — no embed provider available.", nil
	}

	results, err := r.mem.SemanticSearch(ctx, vec, limit)
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err), nil
	}

	data, _ := json.Marshal(results)
	return string(data), nil
}

func (r *InternalToolRegistry) handleListTeams(_ context.Context, _ map[string]any) (string, error) {
	if r.somaRef == nil {
		return "Soma not available — cannot list teams.", nil
	}

	manifests := r.somaRef.ListTeams()
	type teamSummary struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Members int    `json:"members"`
	}

	summaries := make([]teamSummary, 0, len(manifests))
	for _, m := range manifests {
		summaries = append(summaries, teamSummary{
			ID:      m.ID,
			Name:    m.Name,
			Type:    string(m.Type),
			Members: len(m.Members),
		})
	}

	data, _ := json.Marshal(summaries)
	return string(data), nil
}

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
		if err := rows.Scan(&m.ID, &m.Intent, &m.Status, &m.Teams, &m.Agents); err != nil {
			continue
		}
		missions = append(missions, m)
	}

	if missions == nil {
		missions = []missionRow{}
	}

	data, _ := json.Marshal(missions)
	return string(data), nil
}

func (r *InternalToolRegistry) handleGetSystemStatus(_ context.Context, _ map[string]any) (string, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var tokenRate float64
	if r.brain != nil {
		tokenRate = r.brain.TokenRate()
	}

	snap := map[string]any{
		"goroutines":    runtime.NumGoroutine(),
		"heap_alloc_mb": float64(memStats.HeapAlloc) / 1024 / 1024,
		"sys_mem_mb":    float64(memStats.Sys) / 1024 / 1024,
		"llm_tokens_sec": tokenRate,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	data, _ := json.Marshal(snap)
	return string(data), nil
}

func (r *InternalToolRegistry) handleListAvailableTools(ctx context.Context, _ map[string]any) (string, error) {
	// Start with internal (built-in) tools
	descs := r.ListDescriptions()

	// Merge MCP tools from the database
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

	data, _ := json.Marshal(descs)
	return string(data), nil
}

func (r *InternalToolRegistry) handleGenerateBlueprint(ctx context.Context, args map[string]any) (string, error) {
	intent, _ := args["intent"].(string)
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

	data, _ := json.Marshal(bp)
	return string(data), nil
}

func (r *InternalToolRegistry) handleResearchForBlueprint(ctx context.Context, args map[string]any) (string, error) {
	intent, _ := args["intent"].(string)
	if intent == "" {
		return "", fmt.Errorf("research_for_blueprint requires 'intent'")
	}

	var research strings.Builder
	research.WriteString("# Blueprint Research Report\n\n")

	// 1. Recall past missions with similar intents
	memories, err := r.handleRecall(ctx, map[string]any{"query": intent, "limit": float64(3)})
	if err == nil && memories != "" {
		research.WriteString("## Past Mission Context\n" + memories + "\n\n")
	} else {
		research.WriteString("## Past Mission Context\nNo relevant past missions found.\n\n")
	}

	// 2. Search catalogue for reusable agent templates
	catalogue, err := r.handleListCatalogue(ctx, nil)
	if err == nil && catalogue != "" {
		research.WriteString("## Available Agent Templates\n" + catalogue + "\n\n")
	} else {
		research.WriteString("## Available Agent Templates\nAgent catalogue not available.\n\n")
	}

	// 3. List installed + available tools
	tools, err := r.handleListAvailableTools(ctx, nil)
	if err == nil && tools != "" {
		research.WriteString("## Tool Inventory\n" + tools + "\n\n")
	} else {
		research.WriteString("## Tool Inventory\nTool listing not available.\n\n")
	}

	// 4. List active missions for context
	missions, err := r.handleListMissions(ctx, nil)
	if err == nil && missions != "" {
		research.WriteString("## Active Missions\n" + missions + "\n\n")
	}

	research.WriteString("Use this research to inform the blueprint you generate with generate_blueprint.\n")
	return research.String(), nil
}

func (r *InternalToolRegistry) handleListCatalogue(ctx context.Context, _ map[string]any) (string, error) {
	if r.catalogue == nil {
		return "Agent catalogue not available.", nil
	}

	agents, err := r.catalogue.List(ctx)
	if err != nil {
		return fmt.Sprintf("Catalogue query failed: %v", err), nil
	}

	data, _ := json.Marshal(agents)
	return string(data), nil
}

func (r *InternalToolRegistry) handleStoreArtifact(ctx context.Context, args map[string]any) (string, error) {
	artType, _ := args["type"].(string)
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)

	if artType == "" || title == "" || content == "" {
		return "", fmt.Errorf("store_artifact requires 'type', 'title', and 'content'")
	}

	if r.db == nil {
		return "", fmt.Errorf("database not available — cannot store artifact")
	}

	contentType := "text/plain"

	// Chart validation: must be valid JSON with required fields
	if artType == "chart" {
		var spec map[string]any
		if err := json.Unmarshal([]byte(content), &spec); err != nil {
			return "", fmt.Errorf("chart content must be valid JSON: %w", err)
		}
		if _, ok := spec["chart_type"]; !ok {
			return "", fmt.Errorf("chart spec missing required 'chart_type' field")
		}
		if dataArr, ok := spec["data"].([]any); !ok {
			return "", fmt.Errorf("chart spec missing required 'data' array")
		} else if len(dataArr) > 2000 {
			return "", fmt.Errorf("chart data exceeds 2000-row limit (%d rows); summarize or aggregate first", len(dataArr))
		}
		contentType = "application/vnd.mycelis.chart+json"
	}

	metaJSON := "{}"
	if m, ok := args["metadata"]; ok {
		if b, err := json.Marshal(m); err == nil {
			metaJSON = string(b)
		}
	}

	var artifactID string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO artifacts (agent_id, artifact_type, title, content_type, content, metadata, status)
		VALUES ('internal', $1, $2, $3, $4, $5, 'pending')
		RETURNING id
	`, artType, title, contentType, content, metaJSON).Scan(&artifactID)
	if err != nil {
		log.Printf("store_artifact: %v", err)
		return fmt.Sprintf("Failed to store artifact: %v", err), nil
	}

	artifactRef := map[string]any{
		"id": artifactID, "type": artType, "title": title, "content_type": contentType,
	}
	// Include inline content for text-based types (not images/audio — too large for LLM context)
	if artType != "image" && artType != "audio" && len(content) < 500_000 {
		artifactRef["content"] = content
	}

	result := map[string]any{
		"message":  fmt.Sprintf("Artifact '%s' stored (type: %s, id: %s).", title, artType, artifactID),
		"artifact": artifactRef,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (r *InternalToolRegistry) handleRemember(ctx context.Context, args map[string]any) (string, error) {
	category, _ := args["category"].(string)
	content, _ := args["content"].(string)
	memContext, _ := args["context"].(string)

	if category == "" || content == "" {
		return "", fmt.Errorf("remember requires 'category' and 'content'")
	}

	if r.db == nil {
		return "", fmt.Errorf("database not available — cannot persist memory")
	}

	// 1. Store structured record in RDBMS (agent_memories table)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO agent_memories (category, content, context, created_at)
		VALUES ($1, $2, $3, NOW())
	`, category, content, memContext)
	if err != nil {
		log.Printf("remember: RDBMS insert failed: %v", err)
		return fmt.Sprintf("Failed to store memory: %v", err), nil
	}

	// 2. Embed and store vector for semantic recall (best-effort)
	if r.brain != nil && r.mem != nil {
		embeddingText := fmt.Sprintf("[%s] %s", category, content)
		if memContext != "" {
			embeddingText += " (context: " + memContext + ")"
		}
		vec, err := r.brain.Embed(ctx, embeddingText, "")
		if err == nil {
			meta := map[string]any{"category": category, "source": "agent_memory"}
			if insertErr := r.mem.StoreVector(ctx, embeddingText, vec, meta); insertErr != nil {
				log.Printf("remember: vector insert failed: %v", insertErr)
			}
		} else {
			log.Printf("remember: embedding failed (non-fatal): %v", err)
		}
	}

	return fmt.Sprintf("Remembered [%s]: %s", category, content), nil
}

func (r *InternalToolRegistry) handleRecall(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("recall requires 'query'")
	}

	category, _ := args["category"].(string)
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	type memoryResult struct {
		Category  string  `json:"category"`
		Content   string  `json:"content"`
		Context   string  `json:"context,omitempty"`
		CreatedAt string  `json:"created_at"`
		Score     float64 `json:"score,omitempty"`
		Source    string  `json:"source"` // "rdbms" or "vector"
	}

	var results []memoryResult

	// 1. RDBMS keyword search (structured records)
	if r.db != nil {
		var rdbmsQuery string
		var queryArgs []any
		if category != "" {
			rdbmsQuery = `
				SELECT category, content, COALESCE(context, ''), created_at
				FROM agent_memories
				WHERE category = $1 AND content ILIKE '%' || $2 || '%'
				ORDER BY created_at DESC LIMIT $3
			`
			queryArgs = []any{category, query, limit}
		} else {
			rdbmsQuery = `
				SELECT category, content, COALESCE(context, ''), created_at
				FROM agent_memories
				WHERE content ILIKE '%' || $1 || '%'
				ORDER BY created_at DESC LIMIT $2
			`
			queryArgs = []any{query, limit}
		}

		rows, err := r.db.QueryContext(ctx, rdbmsQuery, queryArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var m memoryResult
				var createdAt time.Time
				if err := rows.Scan(&m.Category, &m.Content, &m.Context, &createdAt); err != nil {
					continue
				}
				m.CreatedAt = createdAt.Format(time.RFC3339)
				m.Source = "rdbms"
				results = append(results, m)
			}
		}
	}

	// 2. Semantic vector search (if embedding available)
	if r.brain != nil && r.mem != nil {
		vec, err := r.brain.Embed(ctx, query, "")
		if err == nil {
			vecResults, err := r.mem.SemanticSearch(ctx, vec, limit)
			if err == nil {
				for _, vr := range vecResults {
					results = append(results, memoryResult{
						Content: vr.Content,
						Score:   vr.Score,
						Source:  "vector",
					})
				}
			}
		}
	}

	if results == nil {
		results = []memoryResult{}
	}

	data, _ := json.Marshal(results)
	return string(data), nil
}

func (r *InternalToolRegistry) handlePublishSignal(_ context.Context, args map[string]any) (string, error) {
	subject, _ := args["subject"].(string)
	message, _ := args["message"].(string)
	if subject == "" || message == "" {
		return "", fmt.Errorf("publish_signal requires 'subject' and 'message'")
	}

	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot publish signal")
	}

	if err := r.nc.Publish(subject, []byte(message)); err != nil {
		return "", fmt.Errorf("failed to publish to %s: %w", subject, err)
	}
	r.nc.Flush()

	return fmt.Sprintf("Signal published to %s (%d bytes).", subject, len(message)), nil
}

func (r *InternalToolRegistry) handleBroadcast(_ context.Context, args map[string]any) (string, error) {
	message, _ := args["message"].(string)
	if message == "" {
		return "", fmt.Errorf("broadcast requires 'message'")
	}
	if r.nc == nil {
		return "", fmt.Errorf("NATS not available")
	}
	if r.somaRef == nil {
		return "", fmt.Errorf("Soma not available — cannot enumerate teams")
	}

	teams := r.somaRef.ListTeams()
	if len(teams) == 0 {
		return "No active teams to broadcast to.", nil
	}

	timeout := 60 * time.Second
	type reply struct {
		teamID  string
		content string
		err     error
	}
	ch := make(chan reply, len(teams))

	for _, t := range teams {
		go func(teamID string) {
			subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID)
			msg, err := r.nc.Request(subject, []byte(message), timeout)
			if err != nil {
				ch <- reply{teamID: teamID, err: err}
				return
			}
			ch <- reply{teamID: teamID, content: string(msg.Data)}
		}(t.ID)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Broadcast sent to %d team(s):\n\n", len(teams)))
	for range teams {
		r := <-ch
		if r.err != nil {
			sb.WriteString(fmt.Sprintf("- **%s**: _%v_\n", r.teamID, r.err))
		} else {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", r.teamID, r.content))
		}
	}
	return sb.String(), nil
}

func (r *InternalToolRegistry) handleReadSignals(ctx context.Context, args map[string]any) (string, error) {
	subject, _ := args["subject"].(string)
	if subject == "" {
		return "", fmt.Errorf("read_signals requires 'subject'")
	}

	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot read signals")
	}

	durationMs := 3000
	if d, ok := args["duration_ms"].(float64); ok && d > 0 {
		durationMs = int(d)
	}
	if durationMs > 10000 {
		durationMs = 10000
	}

	maxMsgs := 20
	if m, ok := args["max_msgs"].(float64); ok && m > 0 {
		maxMsgs = int(m)
	}

	type signalMsg struct {
		Subject string `json:"subject"`
		Data    string `json:"data"`
	}

	var collected []signalMsg
	sub, err := r.nc.Subscribe(subject, func(msg *nats.Msg) {
		if len(collected) < maxMsgs {
			collected = append(collected, signalMsg{
				Subject: msg.Subject,
				Data:    string(msg.Data),
			})
		}
	})
	if err != nil {
		return "", fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}
	defer sub.Unsubscribe()

	select {
	case <-time.After(time.Duration(durationMs) * time.Millisecond):
	case <-ctx.Done():
	}

	if collected == nil {
		collected = []signalMsg{}
	}

	result := map[string]any{
		"subject":   subject,
		"duration":  fmt.Sprintf("%dms", durationMs),
		"collected": len(collected),
		"messages":  collected,
	}

	data, _ := json.Marshal(result)
	return string(data), nil
}

func (r *InternalToolRegistry) handleReadFile(_ context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("read_file requires 'path'")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Truncate large files to prevent LLM context overflow
	content := string(data)
	if len(content) > 32000 {
		content = content[:32000] + "\n... [truncated at 32KB]"
	}

	return content, nil
}

func (r *InternalToolRegistry) handleWriteFile(_ context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	if path == "" || content == "" {
		return "", fmt.Errorf("write_file requires 'path' and 'content'")
	}

	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}

	return fmt.Sprintf("File written: %s (%d bytes).", path, len(content)), nil
}

func (r *InternalToolRegistry) handleGenerateImage(ctx context.Context, args map[string]any) (string, error) {
	prompt, _ := args["prompt"].(string)
	if prompt == "" {
		return "", fmt.Errorf("generate_image requires 'prompt'")
	}

	size := "1024x1024"
	if s, ok := args["size"].(string); ok && s != "" {
		size = s
	}

	// Get media endpoint from cognitive config
	if r.brain == nil || r.brain.Config == nil || r.brain.Config.Media == nil {
		return "", fmt.Errorf("media engine not configured — set media.endpoint in cognitive.yaml")
	}

	endpoint := r.brain.Config.Media.Endpoint + "/images/generations"

	reqBody := map[string]any{
		"prompt":          prompt,
		"n":               1,
		"size":            size,
		"response_format": "b64_json",
		"model":           r.brain.Config.Media.ModelID,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create image request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("media engine unreachable: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Media engine error (HTTP %d): %s", resp.StatusCode, string(respBody)), nil
	}

	// Parse response to extract base64 image data
	var imgResp struct {
		Created int `json:"created"`
		Data    []struct {
			B64JSON       string `json:"b64_json"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &imgResp); err != nil {
		return "Image generated but failed to parse response metadata.", nil
	}

	b64Content := ""
	if len(imgResp.Data) > 0 {
		b64Content = imgResp.Data[0].B64JSON
	}

	titleTrunc := prompt
	if len(titleTrunc) > 80 {
		titleTrunc = titleTrunc[:80]
	}
	title := fmt.Sprintf("Generated: %s", titleTrunc)

	// Store the image as an artifact if DB is available
	var artifactID string
	if r.db != nil {
		err := r.db.QueryRowContext(ctx, `
			INSERT INTO artifacts (agent_id, artifact_type, title, content_type, content, metadata, status)
			VALUES ('internal', 'image', $1, 'image/png', $2, $3, 'completed')
			RETURNING id
		`, title, b64Content, string(respBody)).Scan(&artifactID)
		if err != nil {
			log.Printf("generate_image: failed to store artifact: %v", err)
		}
	}

	// Return structured result: message for LLM, artifact for HTTP pipeline
	result := map[string]any{
		"message": fmt.Sprintf("Image generated for: \"%s\" (size: %s). Stored as artifact.", prompt, size),
		"artifact": map[string]any{
			"id":           artifactID,
			"type":         "image",
			"title":        title,
			"content_type": "image/png",
			"content":      b64Content,
		},
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

// ── Conversation Memory ─────────────────────────────────────────

func (r *InternalToolRegistry) handleSummarizeConversation(ctx context.Context, args map[string]any) (string, error) {
	messagesText, _ := args["messages"].(string)
	if messagesText == "" {
		return "", fmt.Errorf("summarize_conversation requires 'messages'")
	}

	if r.brain == nil {
		return "", fmt.Errorf("cognitive engine offline — cannot summarize")
	}
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — cannot store summary")
	}

	return r.summarizeAndStore(ctx, "admin", messagesText, 0)
}

// AutoSummarize compresses a chat history window into a persistent summary.
// Called as a background goroutine from the agent's ReAct pipeline — non-blocking, non-fatal.
func (r *InternalToolRegistry) AutoSummarize(ctx context.Context, agentID string, history []cognitive.ChatMessage) {
	if r.brain == nil || r.mem == nil {
		return
	}

	// Flatten last messages into a text block
	var sb strings.Builder
	for _, m := range history {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, m.Content))
	}

	summaryID, err := r.summarizeAndStore(ctx, agentID, sb.String(), len(history))
	if err != nil {
		log.Printf("AutoSummarize [%s]: failed: %v", agentID, err)
		return
	}
	log.Printf("AutoSummarize [%s]: stored summary %s (%d messages)", agentID, summaryID, len(history))
}

// summarizeAndStore sends conversation text to the LLM for compression, parses the
// structured output, and stores it in the conversation_summaries table + pgvector.
func (r *InternalToolRegistry) summarizeAndStore(ctx context.Context, agentID, messagesText string, msgCount int) (string, error) {
	compressionPrompt := `Summarize this conversation in 2-3 sentences. Then extract structured metadata.

Respond with ONLY this JSON (no markdown fences):
{
  "summary": "2-3 sentence summary of the conversation",
  "key_topics": ["topic1", "topic2"],
  "user_preferences": {"preference_key": "preference_value"},
  "personality_notes": "how the user wants to be addressed or treated",
  "data_references": [{"type": "file/url/artifact", "ref": "the reference"}]
}

Conversation:
` + messagesText

	req := cognitive.InferRequest{
		Profile: "chat",
		Messages: []cognitive.ChatMessage{
			{Role: "system", Content: "You are a conversation summarizer. Output only valid JSON."},
			{Role: "user", Content: compressionPrompt},
		},
	}

	resp, err := r.brain.InferWithContract(ctx, req)
	if err != nil {
		return "", fmt.Errorf("LLM compression failed: %w", err)
	}

	// Parse structured output from LLM
	var parsed struct {
		Summary          string         `json:"summary"`
		KeyTopics        []string       `json:"key_topics"`
		UserPreferences  map[string]any `json:"user_preferences"`
		PersonalityNotes string         `json:"personality_notes"`
		DataReferences   []any          `json:"data_references"`
	}

	// Strip markdown fences if present
	text := strings.TrimSpace(resp.Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		// Fallback: store the raw text as the summary
		log.Printf("summarizeAndStore: LLM output not valid JSON, using raw text: %v", err)
		parsed.Summary = resp.Text
	}

	if parsed.Summary == "" {
		parsed.Summary = resp.Text
	}
	if parsed.KeyTopics == nil {
		parsed.KeyTopics = []string{}
	}
	if parsed.UserPreferences == nil {
		parsed.UserPreferences = map[string]any{}
	}
	if parsed.DataReferences == nil {
		parsed.DataReferences = []any{}
	}

	embedFunc := memory.EmbedFunc(r.brain.Embed)
	summaryID, err := r.mem.StoreConversationSummary(ctx, embedFunc, agentID, parsed.Summary,
		parsed.KeyTopics, parsed.UserPreferences, parsed.PersonalityNotes, parsed.DataReferences, msgCount)
	if err != nil {
		return "", fmt.Errorf("store summary failed: %w", err)
	}

	return summaryID, nil
}

// ── Runtime Context Builder ─────────────────────────────────────

// BuildContext generates a live system state block for injection into an agent's
// system prompt. Gives agents awareness of active teams, NATS topology, cognitive
// config, and installed MCP servers before they process any message.
func (r *InternalToolRegistry) BuildContext(agentID, teamID string, teamInputs, teamDeliveries []string, currentInput string) string {
	var sb strings.Builder
	sb.WriteString("\n\n## Runtime Context (Live System State)\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339)))

	// 1. Active Teams
	r.writeTeamRoster(&sb)

	// 2. Agent Identity & NATS Topology
	r.writeAgentTopology(&sb, agentID, teamID, teamInputs, teamDeliveries)

	// 3. Cognitive Engine
	r.writeCognitiveStatus(&sb)

	// 4. Installed MCP Servers
	r.writeMCPServers(&sb)

	// 5. Recalled Conversation Context (pgvector semantic recall)
	r.writeRecalledMemory(&sb, agentID, currentInput)

	// 6. Interaction Protocol
	sb.WriteString("### Interaction Protocol\n")
	sb.WriteString("**Pre-response** (before answering):\n")
	sb.WriteString("1. Check if past context is relevant → `recall` or `search_memory`\n")
	sb.WriteString("2. Check if specialist knowledge is needed → `consult_council`\n")
	sb.WriteString("3. Check if actionable work should be delegated → `delegate_task`\n")
	sb.WriteString("4. Check if MCP tools can fulfill the request directly\n")
	sb.WriteString("5. Check if data would benefit from visualization → `store_artifact` with type=chart\n\n")
	sb.WriteString("**Post-response** (after completing a task):\n")
	sb.WriteString("1. Store important learnings or decisions → `remember`\n")
	sb.WriteString("2. Store significant outputs → `store_artifact`\n")
	sb.WriteString("3. Report actions taken and outcomes clearly to the user\n")

	return sb.String()
}

func (r *InternalToolRegistry) writeTeamRoster(sb *strings.Builder) {
	sb.WriteString("### Active Teams\n")
	if r.somaRef == nil {
		sb.WriteString("- (Soma offline — team roster unavailable)\n\n")
		return
	}

	manifests := r.somaRef.ListTeams()
	if len(manifests) == 0 {
		sb.WriteString("- No active teams\n\n")
		return
	}

	for _, m := range manifests {
		desc := m.Description
		if desc == "" {
			desc = fmt.Sprintf("%d agent(s)", len(m.Members))
		}
		sb.WriteString(fmt.Sprintf("- **%s** (`%s`, %s): %s\n",
			m.Name, m.ID, m.Type, desc))
		for _, a := range m.Members {
			tools := ""
			if len(a.Tools) > 0 {
				tools = " [" + strings.Join(a.Tools, ", ") + "]"
			}
			sb.WriteString(fmt.Sprintf("  - `%s` (%s)%s\n", a.ID, a.Role, tools))
		}
	}
	sb.WriteString("\n")
}

func (r *InternalToolRegistry) writeAgentTopology(sb *strings.Builder, agentID, teamID string, inputs, deliveries []string) {
	sb.WriteString("### Your Identity & NATS Topology\n")
	sb.WriteString(fmt.Sprintf("- **Agent ID**: `%s`\n", agentID))
	sb.WriteString(fmt.Sprintf("- **Team ID**: `%s`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Team trigger bus**: `swarm.team.%s.internal.trigger`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Team respond bus**: `swarm.team.%s.internal.respond`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Direct address**: `swarm.council.%s.request`\n", agentID))

	if len(inputs) > 0 {
		sb.WriteString(fmt.Sprintf("- **Team inputs**: %s\n", strings.Join(inputs, ", ")))
	}
	if len(deliveries) > 0 {
		sb.WriteString(fmt.Sprintf("- **Team deliveries**: %s\n", strings.Join(deliveries, ", ")))
	}

	sb.WriteString("- **Global broadcast**: `swarm.global.broadcast`\n")
	sb.WriteString("- **Heartbeat**: `swarm.global.heartbeat` (5s, protobuf)\n")
	sb.WriteString("\n")
}

func (r *InternalToolRegistry) writeCognitiveStatus(sb *strings.Builder) {
	sb.WriteString("### Cognitive Engine\n")
	if r.brain == nil || r.brain.Config == nil {
		sb.WriteString("- (Cognitive engine offline)\n\n")
		return
	}

	cfg := r.brain.Config
	for id, prov := range cfg.Providers {
		if prov.Endpoint == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s** (%s): model=`%s`, endpoint=`%s`\n",
			id, prov.Type, prov.ModelID, prov.Endpoint))
	}

	sb.WriteString("- **Profile routing**: ")
	parts := make([]string, 0, len(cfg.Profiles))
	for profile, provID := range cfg.Profiles {
		parts = append(parts, fmt.Sprintf("%s→%s", profile, provID))
	}
	sb.WriteString(strings.Join(parts, ", "))
	sb.WriteString("\n\n")
}

func (r *InternalToolRegistry) writeMCPServers(sb *strings.Builder) {
	sb.WriteString("### Installed MCP Servers & Tools\n")
	if r.db == nil {
		sb.WriteString("- (Database offline — MCP registry unavailable)\n\n")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Servers
	srvRows, err := r.db.QueryContext(ctx, `SELECT id, name, transport, status FROM mcp_servers ORDER BY name`)
	if err != nil {
		sb.WriteString(fmt.Sprintf("- (query failed: %v)\n\n", err))
		return
	}
	defer srvRows.Close()

	type srvInfo struct {
		id, name, transport, status string
	}
	var servers []srvInfo
	for srvRows.Next() {
		var s srvInfo
		if err := srvRows.Scan(&s.id, &s.name, &s.transport, &s.status); err == nil {
			servers = append(servers, s)
		}
	}

	if len(servers) == 0 {
		sb.WriteString("- No MCP servers installed. Install via `/settings` → MCP Tools tab.\n\n")
		return
	}

	// Tools by server (with descriptions for prompt injection)
	type mcpTool struct {
		name, desc string
	}
	toolRows, err := r.db.QueryContext(ctx, `SELECT server_id, name, COALESCE(description, '') FROM mcp_tools ORDER BY name`)
	toolMap := make(map[string][]mcpTool)
	if err == nil {
		defer toolRows.Close()
		for toolRows.Next() {
			var serverID, toolName, toolDesc string
			if err := toolRows.Scan(&serverID, &toolName, &toolDesc); err == nil {
				toolMap[serverID] = append(toolMap[serverID], mcpTool{name: toolName, desc: toolDesc})
			}
		}
	}

	hasMCPTools := false
	for _, s := range servers {
		statusLabel := s.status
		if statusLabel == "connected" {
			statusLabel = "online"
		}
		tools := toolMap[s.id]
		if len(tools) > 0 {
			hasMCPTools = true
			toolNames := make([]string, len(tools))
			for i, t := range tools {
				toolNames[i] = t.name
			}
			sb.WriteString(fmt.Sprintf("- **%s** (%s, %s): tools=[%s]\n",
				s.name, s.transport, statusLabel, strings.Join(toolNames, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf("- **%s** (%s, %s): no tools discovered\n",
				s.name, s.transport, statusLabel))
		}
	}

	if hasMCPTools {
		sb.WriteString("\n**MCP tools are callable.** Use the tool name directly in a tool_call:\n")
		sb.WriteString("```json\n{\"tool_call\": {\"name\": \"<mcp_tool_name>\", \"arguments\": {...}}}\n```\n")

		// List each MCP tool with its description for the LLM
		sb.WriteString("\n**MCP Tool Reference:**\n")
		for _, s := range servers {
			for _, t := range toolMap[s.id] {
				desc := t.desc
				if desc == "" {
					desc = "(no description)"
				}
				sb.WriteString(fmt.Sprintf("- **%s**: %s (via %s)\n", t.name, desc, s.name))
			}
		}
	}
	sb.WriteString("\n")
}

func (r *InternalToolRegistry) writeRecalledMemory(sb *strings.Builder, agentID, currentInput string) {
	if r.brain == nil || r.mem == nil || currentInput == "" {
		return
	}

	// Embed the current input (truncated) for semantic search
	query := currentInput
	if len(query) > 200 {
		query = query[:200]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vec, err := r.brain.Embed(ctx, query, "")
	if err != nil {
		return // Silent — embedding not available
	}

	summaries, err := r.mem.RecallConversations(ctx, vec, agentID, 3)
	if err != nil || len(summaries) == 0 {
		return
	}

	sb.WriteString("### Previous Context (from past conversations)\n")
	for _, s := range summaries {
		age := time.Since(s.CreatedAt)
		var ageStr string
		switch {
		case age < time.Hour:
			ageStr = fmt.Sprintf("%d min ago", int(age.Minutes()))
		case age < 24*time.Hour:
			ageStr = fmt.Sprintf("%d hours ago", int(age.Hours()))
		default:
			ageStr = fmt.Sprintf("%d days ago", int(age.Hours()/24))
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s", ageStr, s.Summary))
		if len(s.KeyTopics) > 0 {
			sb.WriteString(fmt.Sprintf(" (topics: %s)", strings.Join(s.KeyTopics, ", ")))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

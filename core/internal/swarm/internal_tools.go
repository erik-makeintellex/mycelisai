package swarm

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
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
	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/internal/hostcmd"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Phase 0 security: workspace sandbox for file tools
const maxWriteSize = 1 << 20 // 1 MB

func validateToolPath(rawPath string) (string, error) {
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("invalid workspace path: %w", err)
	}

	var absTarget string
	if filepath.IsAbs(rawPath) {
		absTarget = filepath.Clean(rawPath)
	} else {
		absTarget = filepath.Clean(filepath.Join(absWorkspace, rawPath))
	}

	rel, err := filepath.Rel(absWorkspace, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %q escapes workspace boundary", rawPath)
	}

	// Symlink escape check (best-effort for existing paths)
	checkPath := absTarget
	if _, statErr := os.Lstat(absTarget); os.IsNotExist(statErr) {
		checkPath = filepath.Dir(absTarget)
	}
	if realPath, evalErr := filepath.EvalSymlinks(checkPath); evalErr == nil {
		relReal, _ := filepath.Rel(absWorkspace, realPath)
		if strings.HasPrefix(relReal, "..") {
			return "", fmt.Errorf("symlink at %q resolves outside workspace", rawPath)
		}
	}
	return absTarget, nil
}

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
	inception *inception.Store
	comms     *comms.Gateway
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
	Inception *inception.Store
	Comms     *comms.Gateway
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
		inception: deps.Inception,
		comms:     deps.Comms,
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
		Description: "Publish a task to a specific team's command topic for processing.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"team_id": map[string]any{"type": "string", "description": "The target team ID"},
				"task":    map[string]any{"type": "string", "description": "The task description to send"},
				"hint": map[string]any{
					"type":        "object",
					"description": "Optional scoring hints for delegation priority",
					"properties": map[string]any{
						"confidence": map[string]any{"type": "number", "description": "0.0-1.0 confidence this is the right team"},
						"urgency":    map[string]any{"type": "string", "enum": []string{"low", "medium", "high", "critical"}},
						"complexity": map[string]any{"type": "integer", "description": "1-5 complexity rating"},
						"risk":       map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
					},
				},
			},
			"required": []string{"team_id", "task"},
		},
		Handler: r.handleDelegateTask,
	}

	r.tools["create_team"] = &InternalTool{
		Name:        "create_team",
		Description: "Create and start a new team at runtime with a minimal manifest.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"team_id":       map[string]any{"type": "string", "description": "Unique team ID"},
				"name":          map[string]any{"type": "string", "description": "Display name (optional)"},
				"type":          map[string]any{"type": "string", "enum": []string{"action", "expression"}, "description": "Team type (default action)"},
				"role":          map[string]any{"type": "string", "description": "Primary agent role (default worker)"},
				"agent_id":      map[string]any{"type": "string", "description": "Optional first agent ID"},
				"system_prompt": map[string]any{"type": "string", "description": "Optional first agent system prompt"},
				"tools":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional first agent tools"},
				"inputs":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional team input subjects"},
				"deliveries":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional team delivery subjects"},
			},
			"required": []string{"team_id"},
		},
		Handler: r.handleCreateTeam,
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

	r.tools["temp_memory_write"] = &InternalTool{
		Name:        "temp_memory_write",
		Description: "Persist a temporary working-memory checkpoint (restart-safe) for lead-agent continuity. Use channels like lead.shared, lead.<agent_id>, or interaction.contract.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"channel":        map[string]any{"type": "string", "description": "Channel key, e.g. lead.shared or lead.council-architect"},
				"content":        map[string]any{"type": "string", "description": "Checkpoint content"},
				"owner_agent_id": map[string]any{"type": "string", "description": "Optional owner identity"},
				"ttl_minutes":    map[string]any{"type": "integer", "description": "Optional expiration in minutes (<=0 means no expiry)"},
				"metadata":       map[string]any{"type": "object", "description": "Optional structured metadata"},
			},
			"required": []string{"channel", "content"},
		},
		Handler: r.handleTempMemoryWrite,
	}

	r.tools["temp_memory_read"] = &InternalTool{
		Name:        "temp_memory_read",
		Description: "Read recent temporary working-memory checkpoints for a channel.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"channel": map[string]any{"type": "string", "description": "Channel key"},
				"limit":   map[string]any{"type": "integer", "description": "Max entries (default 10)"},
			},
			"required": []string{"channel"},
		},
		Handler: r.handleTempMemoryRead,
	}

	r.tools["temp_memory_clear"] = &InternalTool{
		Name:        "temp_memory_clear",
		Description: "Clear a temporary working-memory channel.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"channel": map[string]any{"type": "string", "description": "Channel key"},
			},
			"required": []string{"channel"},
		},
		Handler: r.handleTempMemoryClear,
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
		Description: "Publish a message to a NATS topic in the swarm. Supports private reference mode for channel-safe file/message handoff plus latest-checkpoint persistence for relaunch continuity.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"subject":      map[string]any{"type": "string", "description": "NATS subject (e.g. swarm.team.council-core.internal.command, swarm.global.broadcast)"},
				"message":      map[string]any{"type": "string", "description": "Message payload to publish. In private reference mode this is stored privately and only a reference is published."},
				"channel_key":  map[string]any{"type": "string", "description": "Optional private checkpoint channel key. Defaults to signal.latest.<subject>."},
				"privacy_mode": map[string]any{"type": "string", "description": "full (default) or reference. reference publishes only channel/file references while preserving full payload privately."},
				"private":      map[string]any{"type": "boolean", "description": "Alias for privacy_mode=reference."},
				"file_path":    map[string]any{"type": "string", "description": "Optional workspace file path to attach as a private file reference in channel payloads."},
			},
			"required": []string{},
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
				"urgency": map[string]any{"type": "string", "description": "Optional urgency level", "enum": []string{"low", "medium", "high", "critical"}},
			},
			"required": []string{"message"},
		},
		Handler: r.handleBroadcast,
	}

	r.tools["send_external_message"] = &InternalTool{
		Name:        "send_external_message",
		Description: "Send a message through an external communication provider (whatsapp, telegram, slack, webhook). Use for operator/user notifications and out-of-band alerts.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"provider":  map[string]any{"type": "string", "description": "Provider name: whatsapp, telegram, slack, webhook"},
				"recipient": map[string]any{"type": "string", "description": "Recipient handle/ID/number/channel"},
				"message":   map[string]any{"type": "string", "description": "Message body"},
				"metadata":  map[string]any{"type": "object", "description": "Optional metadata map"},
			},
			"required": []string{"provider", "message"},
		},
		Handler: r.handleSendExternalMessage,
	}

	r.tools["read_signals"] = &InternalTool{
		Name:        "read_signals",
		Description: "Subscribe to NATS and collect messages for a brief window, or read the latest persisted channel checkpoint (latest_only) for relaunch-safe continuity.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"subject":     map[string]any{"type": "string", "description": "NATS subject or wildcard (e.g. swarm.team.*.signal.status, swarm.global.>)"},
				"channel_key": map[string]any{"type": "string", "description": "Optional checkpoint channel key. Defaults to signal.latest.<subject>."},
				"latest_only": map[string]any{"type": "boolean", "description": "When true, returns the latest persisted channel checkpoint instead of live subscription."},
				"duration_ms": map[string]any{"type": "integer", "description": "How long to listen in milliseconds (default 3000, max 10000)"},
				"max_msgs":    map[string]any{"type": "integer", "description": "Max messages to collect (default 20)"},
			},
			"required": []string{"subject"},
		},
		Handler: r.handleReadSignals,
	}

	r.tools["read_file"] = &InternalTool{
		Name:        "read_file",
		Description: "Read the contents of a file within the workspace sandbox (MYCELIS_WORKSPACE). Relative paths resolve from the workspace root. Absolute paths must be inside the workspace boundary.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "File path relative to workspace root, or absolute path within workspace"},
			},
			"required": []string{"path"},
		},
		Handler: r.handleReadFile,
	}

	r.tools["write_file"] = &InternalTool{
		Name:        "write_file",
		Description: "Write content to a file within the workspace sandbox (MYCELIS_WORKSPACE). Creates parent directories if needed. Max 1MB per write. Paths must resolve inside the workspace boundary.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]any{"type": "string", "description": "File path relative to workspace root, or absolute path within workspace"},
				"content": map[string]any{"type": "string", "description": "The file content to write"},
			},
			"required": []string{"path", "content"},
		},
		Handler: r.handleWriteFile,
	}

	r.tools["local_command"] = &InternalTool{
		Name:        "local_command",
		Description: "Execute one local command from the MYCELIS_LOCAL_COMMAND_ALLOWLIST without shell interpolation. Uses bounded timeout and argument validation.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command":    map[string]any{"type": "string", "description": "Allowlisted command name"},
				"args":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Command arguments"},
				"timeout_ms": map[string]any{"type": "integer", "description": "Optional timeout in milliseconds (clamped)"},
			},
			"required": []string{"command"},
		},
		Handler: r.handleLocalCommand,
	}

	r.tools["generate_image"] = &InternalTool{
		Name:        "generate_image",
		Description: "Generate an image from a text prompt using the local Diffusers media engine (Stable Diffusion XL). Generated images are cache-first and expire in 60 minutes unless explicitly saved with save_cached_image.",
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

	r.tools["save_cached_image"] = &InternalTool{
		Name:        "save_cached_image",
		Description: "Persist a cached generated image into the workspace saved-media folder (or a specified subfolder). Use when user asks to keep an image beyond cache TTL.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"artifact_id": map[string]any{"type": "string", "description": "Optional image artifact ID. If omitted, saves the latest unsaved cached image."},
				"folder":      map[string]any{"type": "string", "description": "Optional workspace-relative folder (default: saved-media)."},
				"filename":    map[string]any{"type": "string", "description": "Optional target filename (extension inferred if omitted)."},
			},
		},
		Handler: r.handleSaveCachedImage,
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

	r.tools["store_inception_recipe"] = &InternalTool{
		Name:        "store_inception_recipe",
		Description: "Store a structured inception recipe — a pattern for how to ask for or create a specific thing. Use after successfully completing a complex task to capture the approach for future reuse. The recipe is dual-persisted: RDBMS (structured query) + pgvector (semantic recall).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"category":       map[string]any{"type": "string", "description": "Recipe category: blueprint, analysis, team_setup, deployment, research, configuration, migration, etc."},
				"title":          map[string]any{"type": "string", "description": "Short descriptive title: 'How to create a microservices blueprint'"},
				"intent_pattern": map[string]any{"type": "string", "description": "The structured way to phrase this request. Be specific about parameters, constraints, and expected format."},
				"parameters":     map[string]any{"type": "object", "description": "Key parameters as name→description pairs. E.g. {\"service_count\": \"number of services to deploy\"}"},
				"example_prompt": map[string]any{"type": "string", "description": "A concrete example prompt that successfully produced the desired outcome"},
				"outcome_shape":  map[string]any{"type": "string", "description": "What the successful outcome looks like (format, structure, key sections)"},
				"tags":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Searchable tags for filtering"},
			},
			"required": []string{"category", "title", "intent_pattern"},
		},
		Handler: r.handleStoreInceptionRecipe,
	}

	r.tools["recall_inception_recipes"] = &InternalTool{
		Name:        "recall_inception_recipes",
		Description: "Recall inception recipes — structured patterns for how to ask for or create specific things. Search by keyword or category to find proven approaches from past successful tasks. Use before starting complex tasks to leverage existing knowledge.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":    map[string]any{"type": "string", "description": "Search query (keyword or semantic)"},
				"category": map[string]any{"type": "string", "description": "Optional category filter: blueprint, analysis, team_setup, etc."},
				"limit":    map[string]any{"type": "integer", "description": "Max results (default 5)"},
			},
			"required": []string{"query"},
		},
		Handler: r.handleRecallInceptionRecipes,
	}
}

// ── Tool Handlers ──────────────────────────────────────────────

func (r *InternalToolRegistry) handleConsultCouncil(ctx context.Context, args map[string]any) (string, error) {
	member, _ := args["member"].(string)
	if strings.TrimSpace(member) == "" {
		if v, _ := args["agent"].(string); strings.TrimSpace(v) != "" {
			member = v
		} else if v, _ := args["target"].(string); strings.TrimSpace(v) != "" {
			member = v
		}
	}
	member = normalizeCouncilMember(member)

	question, _ := args["question"].(string)
	if strings.TrimSpace(question) == "" {
		if v, _ := args["query"].(string); strings.TrimSpace(v) != "" {
			question = v
		} else if v, _ := args["prompt"].(string); strings.TrimSpace(v) != "" {
			question = v
		} else if v, _ := args["message"].(string); strings.TrimSpace(v) != "" {
			question = v
		}
	}
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

	// Agent returns structured JSON (ProcessResult). Preserve artifacts so
	// Soma can return rich outputs from consulted specialists through the
	// primary conversation channel instead of dropping them to plain text.
	var result struct {
		Text      string                     `json:"text"`
		Artifacts []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
	}
	if err := json.Unmarshal(msg.Data, &result); err == nil && result.Text != "" {
		if len(result.Artifacts) > 0 {
			payload := map[string]any{
				"message":   result.Text,
				"artifacts": result.Artifacts,
			}
			data, _ := json.Marshal(payload)
			return string(data), nil
		}
		return result.Text, nil
	}
	return string(msg.Data), nil
}

func (r *InternalToolRegistry) handleDelegateTask(ctx context.Context, args map[string]any) (string, error) {
	teamID, task := normalizeDelegateTaskArgs(args)
	if teamID == "" || task == "" {
		return "", fmt.Errorf("delegate_task requires 'team_id' and 'task'")
	}

	// Log delegation hint if provided (observability only — V1)
	if hintRaw, ok := args["hint"]; ok {
		if hint, ok := hintRaw.(map[string]any); ok {
			log.Printf("DelegationHint [%s]: confidence=%.2f urgency=%v complexity=%v risk=%v",
				teamID,
				hint["confidence"],
				hint["urgency"],
				hint["complexity"],
				hint["risk"])
		}
	}

	if r.nc == nil {
		return "", fmt.Errorf("NATS not available — cannot delegate task")
	}

	subject := fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID)
	payload, err := r.wrapGovernedSignalPayload(
		ctx,
		"internal_tool.delegate_task",
		teamID,
		protocol.PayloadKindCommand,
		[]byte(task),
	)
	if err != nil {
		return "", fmt.Errorf("failed to wrap delegated task payload: %w", err)
	}
	if err := r.nc.Publish(subject, payload); err != nil {
		return "", fmt.Errorf("failed to publish task to team %s: %w", teamID, err)
	}
	r.nc.Flush()

	return fmt.Sprintf("Task delegated to team %s.", teamID), nil
}

func (r *InternalToolRegistry) handleCreateTeam(_ context.Context, args map[string]any) (string, error) {
	if r.somaRef == nil {
		return "", fmt.Errorf("Soma not available — cannot create team")
	}

	stringFrom := func(v any) string {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
		return ""
	}
	normalizeTeamID := func(raw string) string {
		s := strings.ToLower(strings.TrimSpace(raw))
		s = strings.ReplaceAll(s, " ", "-")
		s = strings.ReplaceAll(s, "_", "-")
		return s
	}
	toStrings := func(v any) []string {
		switch t := v.(type) {
		case []string:
			return t
		case []any:
			out := make([]string, 0, len(t))
			for _, item := range t {
				if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
					out = append(out, strings.TrimSpace(s))
				}
			}
			return out
		default:
			return nil
		}
	}

	manifestMap, _ := args["manifest"].(map[string]any)
	merged := map[string]any{}
	for k, v := range args {
		merged[k] = v
	}
	for k, v := range manifestMap {
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
	}
	// Nested agents[0] compatibility
	if agentsRaw, ok := merged["agents"].([]any); ok && len(agentsRaw) > 0 {
		if first, ok := agentsRaw[0].(map[string]any); ok {
			if _, exists := merged["agent_id"]; !exists {
				if v := stringFrom(first["agent_id"]); v != "" {
					merged["agent_id"] = v
				} else if v := stringFrom(first["id"]); v != "" {
					merged["agent_id"] = v
				}
			}
			if _, exists := merged["role"]; !exists {
				if v := stringFrom(first["role"]); v != "" {
					merged["role"] = v
				}
			}
			if _, exists := merged["tools"]; !exists {
				if v, ok := first["tools"]; ok {
					merged["tools"] = v
				}
			}
			if _, exists := merged["system_prompt"]; !exists {
				if v := stringFrom(first["system_prompt"]); v != "" {
					merged["system_prompt"] = v
				}
			}
		}
	}

	teamID := normalizeTeamID(stringFrom(merged["team_id"]))
	if teamID == "" {
		teamID = normalizeTeamID(stringFrom(merged["id"]))
	}
	if teamID == "" {
		teamID = normalizeTeamID(stringFrom(merged["team_name"]))
	}
	if teamID == "" {
		return "", fmt.Errorf("create_team requires 'team_id'")
	}

	for _, m := range r.somaRef.ListTeams() {
		if m != nil && m.ID == teamID {
			out, _ := json.Marshal(map[string]any{
				"status":  "already_exists",
				"team_id": teamID,
			})
			return string(out), nil
		}
	}

	name := stringFrom(merged["name"])
	if name == "" {
		name = teamID
	}

	teamType := TeamType(stringFrom(merged["type"]))
	if teamType == "" {
		teamType = TeamTypeAction
	}
	if teamType != TeamTypeAction && teamType != TeamTypeExpression {
		teamType = TeamTypeAction
	}

	role := stringFrom(merged["role"])
	if role == "" {
		role = "worker"
	}
	agentID := normalizeTeamID(stringFrom(merged["agent_id"]))
	if agentID == "" {
		agentID = teamID + "-agent"
	}
	systemPrompt := stringFrom(merged["system_prompt"])
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf("You are %s in team %s. Execute assigned tasks and report outcomes.", role, teamID)
	}

	inputs := toStrings(merged["inputs"])
	if len(inputs) == 0 {
		inputs = []string{protocol.TopicGlobalBroadcast}
	}
	deliveries := toStrings(merged["deliveries"])
	if len(deliveries) == 0 {
		if teamType == TeamTypeExpression {
			deliveries = []string{fmt.Sprintf(protocol.TopicTeamSignalStatus, teamID)}
		} else {
			deliveries = []string{fmt.Sprintf(protocol.TopicTeamSignalResult, teamID)}
		}
	}
	tools := toStrings(merged["tools"])

	manifest := &TeamManifest{
		ID:          teamID,
		Name:        name,
		Type:        teamType,
		Description: "Runtime-created team",
		Members: []protocol.AgentManifest{
			{
				ID:            agentID,
				Role:          role,
				SystemPrompt:  systemPrompt,
				Tools:         tools,
				MaxIterations: 6,
			},
		},
		Inputs:     inputs,
		Deliveries: deliveries,
	}

	if err := r.somaRef.SpawnTeam(manifest); err != nil {
		return "", fmt.Errorf("create_team failed: %w", err)
	}

	out, _ := json.Marshal(map[string]any{
		"status":  "created",
		"team_id": teamID,
		"name":    name,
	})
	return string(out), nil
}

func normalizeDelegateTaskArgs(args map[string]any) (teamID string, task string) {
	stringFrom := func(v any) string {
		switch s := v.(type) {
		case string:
			return strings.TrimSpace(s)
		default:
			return ""
		}
	}
	serialize := func(v any) string {
		if v == nil {
			return ""
		}
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}

	teamID = stringFrom(args["team_id"])
	if teamID == "" {
		teamID = stringFrom(args["teamId"])
	}
	if teamID == "" {
		if teamMap, ok := args["team"].(map[string]any); ok {
			teamID = stringFrom(teamMap["id"])
			if teamID == "" {
				teamID = stringFrom(teamMap["team_id"])
			}
			if teamID == "" {
				teamID = stringFrom(teamMap["name"])
			}
		} else {
			teamID = stringFrom(args["team"])
		}
	}
	if teamID == "" {
		teamID = stringFrom(args["target_team"])
	}

	switch t := args["task"].(type) {
	case string:
		task = strings.TrimSpace(t)
	case map[string]any:
		task = serialize(t)
	case []any:
		task = serialize(t)
	}

	if task == "" {
		payload := map[string]any{}
		if op := stringFrom(args["operation"]); op != "" {
			payload["operation"] = op
		}
		if intent := stringFrom(args["intent"]); intent != "" {
			payload["intent"] = intent
		}
		if msg := stringFrom(args["message"]); msg != "" {
			payload["message"] = msg
		}
		if ctxRaw, ok := args["context"]; ok {
			payload["context"] = ctxRaw
		}
		if len(payload) > 0 {
			task = serialize(payload)
		}
	}

	return teamID, task
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
		"goroutines":     runtime.NumGoroutine(),
		"heap_alloc_mb":  float64(memStats.HeapAlloc) / 1024 / 1024,
		"sys_mem_mb":     float64(memStats.Sys) / 1024 / 1024,
		"llm_tokens_sec": tokenRate,
		"timestamp":      time.Now().Format(time.RFC3339),
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

	// 5. Recall inception recipes — proven patterns for similar tasks
	recipes, err := r.handleRecallInceptionRecipes(ctx, map[string]any{"query": intent, "limit": float64(3)})
	if err == nil && recipes != "" && recipes != "[]" {
		research.WriteString("## Inception Recipes (Proven Patterns)\n" + recipes + "\n\n")
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

func (r *InternalToolRegistry) handleBroadcast(_ context.Context, args map[string]any) (string, error) {
	message, _ := args["message"].(string)
	if message == "" {
		return "", fmt.Errorf("broadcast requires 'message'")
	}

	// Log urgency hint if provided (observability only — V1)
	if urgency, ok := args["urgency"].(string); ok && urgency != "" {
		log.Printf("Broadcast urgency: %s", urgency)
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

func (r *InternalToolRegistry) handleSendExternalMessage(ctx context.Context, args map[string]any) (string, error) {
	if r.comms == nil {
		return "", fmt.Errorf("communications gateway unavailable")
	}

	provider, _ := args["provider"].(string)
	recipient, _ := args["recipient"].(string)
	message, _ := args["message"].(string)
	if strings.TrimSpace(provider) == "" || strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("send_external_message requires provider and message")
	}

	metadata, _ := args["metadata"].(map[string]any)
	res, err := r.comms.Send(ctx, comms.SendRequest{
		Provider:  provider,
		Recipient: recipient,
		Message:   message,
		Metadata:  metadata,
	})
	if err != nil {
		return "", fmt.Errorf("external send failed: %w", err)
	}

	out := map[string]any{
		"message":  fmt.Sprintf("external message sent via %s", res.Provider),
		"provider": res.Provider,
		"status":   res.Status,
		"result":   res,
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func (r *InternalToolRegistry) handleReadFile(_ context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("read_file requires 'path'")
	}

	safePath, err := validateToolPath(path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(safePath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", safePath, err)
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

	if len(content) > maxWriteSize {
		return "", fmt.Errorf("content size %d exceeds maximum write size of %d bytes", len(content), maxWriteSize)
	}

	safePath, err := validateToolPath(path)
	if err != nil {
		return "", err
	}

	// Create parent directories
	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(safePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", safePath, err)
	}

	return fmt.Sprintf("File written: %s (%d bytes).", safePath, len(content)), nil
}

func (r *InternalToolRegistry) handleLocalCommand(ctx context.Context, args map[string]any) (string, error) {
	command, _ := args["command"].(string)
	if strings.TrimSpace(command) == "" {
		return "", fmt.Errorf("local_command requires 'command'")
	}

	var cmdArgs []string
	if raw, ok := args["args"].([]any); ok {
		cmdArgs = make([]string, 0, len(raw))
		for _, item := range raw {
			v, ok := item.(string)
			if !ok {
				continue
			}
			cmdArgs = append(cmdArgs, v)
		}
	}

	timeoutMS := 5000
	if v, ok := args["timeout_ms"].(float64); ok {
		timeoutMS = int(v)
	}

	if len(cmdArgs) == 0 && strings.ContainsAny(command, " \t\r\n'\"`|&;<>") {
		return "", fmt.Errorf(
			"local_command requires a bare allowlisted command name in 'command' and separate 'args'; shell snippets are not allowed. For creating text output, answer directly or use write_file with 'path' and 'content'",
		)
	}

	result, err := hostcmd.Execute(ctx, command, cmdArgs, time.Duration(timeoutMS)*time.Millisecond)
	if err != nil {
		return "", err
	}
	payload, _ := json.Marshal(result)
	return string(payload), nil
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
	expiresAt := time.Now().UTC().Add(60 * time.Minute).Format(time.RFC3339)

	meta := map[string]any{
		"cache_policy": "ephemeral",
		"saved":        false,
		"ttl_minutes":  60,
		"expires_at":   expiresAt,
		"prompt":       prompt,
		"size":         size,
	}
	if len(imgResp.Data) > 0 && imgResp.Data[0].RevisedPrompt != "" {
		meta["revised_prompt"] = imgResp.Data[0].RevisedPrompt
	}
	metaJSON, _ := json.Marshal(meta)

	// Store the image as an artifact if DB is available
	var artifactID string
	if r.db != nil {
		err := r.db.QueryRowContext(ctx, `
			INSERT INTO artifacts (agent_id, artifact_type, title, content_type, content, metadata, status)
			VALUES ('internal', 'image', $1, 'image/png', $2, $3, 'completed')
			RETURNING id
		`, title, b64Content, metaJSON).Scan(&artifactID)
		if err != nil {
			log.Printf("generate_image: failed to store artifact: %v", err)
		}
	}

	// Return structured result: message for LLM, artifact for HTTP pipeline
	result := map[string]any{
		"message": fmt.Sprintf("Image generated for: \"%s\" (size: %s). Cached for 60 minutes unless saved.", prompt, size),
		"artifact": map[string]any{
			"id":           artifactID,
			"type":         "image",
			"title":        title,
			"content_type": "image/png",
			"content":      b64Content,
			"cached":       true,
			"expires_at":   expiresAt,
		},
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (r *InternalToolRegistry) handleSaveCachedImage(ctx context.Context, args map[string]any) (string, error) {
	if r.db == nil {
		return "", fmt.Errorf("database not available — cannot save cached image")
	}

	artifactID, _ := args["artifact_id"].(string)
	folder, _ := args["folder"].(string)
	filename, _ := args["filename"].(string)

	var (
		id          string
		title       string
		contentType string
		contentB64  string
	)

	if strings.TrimSpace(artifactID) != "" {
		err := r.db.QueryRowContext(ctx, `
			SELECT id::text, title, content_type, content
			FROM artifacts
			WHERE id = $1::uuid
			  AND artifact_type = 'image'
		`, artifactID).Scan(&id, &title, &contentType, &contentB64)
		if err != nil {
			return "", fmt.Errorf("cached image %q not found", artifactID)
		}
	} else {
		err := r.db.QueryRowContext(ctx, `
			SELECT id::text, title, content_type, content
			FROM artifacts
			WHERE artifact_type = 'image'
			  AND COALESCE(metadata->>'cache_policy', '') = 'ephemeral'
			  AND COALESCE(metadata->>'saved', 'false') <> 'true'
			ORDER BY created_at DESC
			LIMIT 1
		`).Scan(&id, &title, &contentType, &contentB64)
		if err != nil {
			return "", fmt.Errorf("no unsaved cached image found")
		}
	}

	if strings.TrimSpace(contentB64) == "" {
		return "", fmt.Errorf("cached image has no content")
	}

	data, err := base64.StdEncoding.DecodeString(contentB64)
	if err != nil {
		return "", fmt.Errorf("decode cached image: %w", err)
	}

	if strings.TrimSpace(folder) == "" {
		folder = "saved-media"
	}
	if strings.TrimSpace(filename) == "" {
		filename = sanitizeImageFilename(title)
		if filename == "" {
			filename = fmt.Sprintf("image-%s", id[:8])
		}
	}
	if filepath.Ext(filename) == "" {
		filename += imageFileExt(contentType)
	}

	targetPath, err := validateToolPath(filepath.ToSlash(filepath.Join(folder, filename)))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", fmt.Errorf("create target directory: %w", err)
	}
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write target file: %w", err)
	}

	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	absWorkspace, _ := filepath.Abs(workspace)
	rel, relErr := filepath.Rel(absWorkspace, targetPath)
	if relErr != nil {
		rel = targetPath
	}
	rel = filepath.ToSlash(rel)

	_, err = r.db.ExecContext(ctx, `
		UPDATE artifacts
		SET file_path = $1,
		    file_size_bytes = $2,
		    metadata = COALESCE(metadata, '{}'::jsonb) ||
		               jsonb_build_object('saved', true, 'saved_path', $1, 'saved_at', NOW())
		WHERE id = $3::uuid
	`, rel, int64(len(data)), id)
	if err != nil {
		return "", fmt.Errorf("update image save metadata: %w", err)
	}

	result := map[string]any{
		"message": fmt.Sprintf("Saved cached image to %s", rel),
		"artifact": map[string]any{
			"id":         id,
			"type":       "file",
			"title":      filename,
			"saved_path": rel,
		},
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}

func sanitizeImageFilename(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-.")
	if len(out) > 80 {
		out = out[:80]
	}
	return out
}

func imageFileExt(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

// ── Conversation Memory ─────────────────────────────────────────

func (r *InternalToolRegistry) handleTempMemoryWrite(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}

	channel, _ := args["channel"].(string)
	content, _ := args["content"].(string)
	owner, _ := args["owner_agent_id"].(string)
	metadata, _ := args["metadata"].(map[string]any)
	ttl := 0
	if ttlRaw, ok := args["ttl_minutes"].(float64); ok {
		ttl = int(ttlRaw)
	}

	id, err := r.mem.PutTempMemory(ctx, "default", channel, owner, content, metadata, ttl)
	if err != nil {
		return "", err
	}

	result := map[string]any{
		"message": fmt.Sprintf("temp memory stored in channel %q", channel),
		"id":      id,
		"channel": channel,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (r *InternalToolRegistry) handleTempMemoryRead(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	channel, _ := args["channel"].(string)
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	entries, err := r.mem.GetTempMemory(ctx, "default", channel, limit)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(entries)
	return string(data), nil
}

func (r *InternalToolRegistry) handleTempMemoryClear(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	channel, _ := args["channel"].(string)
	deleted, err := r.mem.ClearTempMemory(ctx, "default", channel)
	if err != nil {
		return "", err
	}
	result := map[string]any{
		"message": fmt.Sprintf("temp memory channel %q cleared", channel),
		"deleted": deleted,
		"channel": channel,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

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

// ── Inception Recipe Handlers ────────────────────────────────────

func (r *InternalToolRegistry) handleStoreInceptionRecipe(ctx context.Context, args map[string]any) (string, error) {
	category, _ := args["category"].(string)
	title, _ := args["title"].(string)
	intentPattern, _ := args["intent_pattern"].(string)

	if category == "" || title == "" || intentPattern == "" {
		return "", fmt.Errorf("store_inception_recipe requires 'category', 'title', and 'intent_pattern'")
	}

	if r.inception == nil {
		return "", fmt.Errorf("inception store not available")
	}

	recipe := inception.Recipe{
		Category:      category,
		Title:         title,
		IntentPattern: intentPattern,
		AgentID:       "admin",
	}

	if params, ok := args["parameters"].(map[string]any); ok {
		recipe.Parameters = params
	}
	if example, ok := args["example_prompt"].(string); ok {
		recipe.ExamplePrompt = example
	}
	if outcome, ok := args["outcome_shape"].(string); ok {
		recipe.OutcomeShape = outcome
	}
	if tags, ok := args["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				recipe.Tags = append(recipe.Tags, s)
			}
		}
	}

	id, err := r.inception.CreateRecipe(ctx, recipe)
	if err != nil {
		return "", fmt.Errorf("store inception recipe failed: %w", err)
	}

	// Best-effort: embed into pgvector for semantic recall
	if r.brain != nil && r.mem != nil {
		embeddingText := fmt.Sprintf("[inception:%s] %s — %s", category, title, intentPattern)
		vec, err := r.brain.Embed(ctx, embeddingText, "")
		if err == nil {
			meta := map[string]any{
				"category":  category,
				"source":    "inception_recipe",
				"recipe_id": id,
			}
			if insertErr := r.mem.StoreVector(ctx, embeddingText, vec, meta); insertErr != nil {
				log.Printf("store_inception_recipe: vector insert failed (non-fatal): %v", insertErr)
			}
		} else {
			log.Printf("store_inception_recipe: embedding failed (non-fatal): %v", err)
		}
	}

	result := map[string]any{
		"message":   fmt.Sprintf("Inception recipe stored: '%s' (category: %s, id: %s)", title, category, id),
		"recipe_id": id,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (r *InternalToolRegistry) handleRecallInceptionRecipes(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("recall_inception_recipes requires 'query'")
	}

	category, _ := args["category"].(string)
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	type recipeResult struct {
		ID            string         `json:"id"`
		Category      string         `json:"category"`
		Title         string         `json:"title"`
		IntentPattern string         `json:"intent_pattern"`
		Parameters    map[string]any `json:"parameters,omitempty"`
		ExamplePrompt string         `json:"example_prompt,omitempty"`
		OutcomeShape  string         `json:"outcome_shape,omitempty"`
		QualityScore  float64        `json:"quality_score"`
		UsageCount    int            `json:"usage_count"`
		Source        string         `json:"source"` // "rdbms" or "vector"
	}

	var results []recipeResult

	// 1. RDBMS keyword/title search
	if r.inception != nil {
		var recipes []inception.Recipe
		var err error
		if category != "" {
			recipes, err = r.inception.ListRecipes(ctx, category, "", limit)
		} else {
			recipes, err = r.inception.SearchByTitle(ctx, query, limit)
		}
		if err == nil {
			for _, rec := range recipes {
				results = append(results, recipeResult{
					ID: rec.ID, Category: rec.Category, Title: rec.Title,
					IntentPattern: rec.IntentPattern, Parameters: rec.Parameters,
					ExamplePrompt: rec.ExamplePrompt, OutcomeShape: rec.OutcomeShape,
					QualityScore: rec.QualityScore, UsageCount: rec.UsageCount,
					Source: "rdbms",
				})
				// Bump usage count asynchronously
				go func(id string) { _ = r.inception.IncrementUsage(context.Background(), id) }(rec.ID)
			}
		}
	}

	// 2. Semantic vector search for inception recipes (complements keyword search)
	if r.brain != nil && r.mem != nil {
		vec, err := r.brain.Embed(ctx, fmt.Sprintf("[inception] %s", query), "")
		if err == nil {
			vecResults, err := r.mem.SemanticSearch(ctx, vec, limit)
			if err == nil {
				for _, vr := range vecResults {
					// Only include inception recipe vectors
					if src, ok := vr.Metadata["source"].(string); ok && src == "inception_recipe" {
						results = append(results, recipeResult{
							ID:       fmt.Sprintf("%v", vr.Metadata["recipe_id"]),
							Category: fmt.Sprintf("%v", vr.Metadata["category"]),
							Title:    vr.Content,
							Source:   "vector",
						})
					}
				}
			}
		}
	}

	if results == nil {
		results = []recipeResult{}
	}

	data, _ := json.Marshal(results)
	return string(data), nil
}

// ── Runtime Context Builder ─────────────────────────────────────

// BuildContext generates a live system state block for injection into an agent's
// system prompt. Gives agents awareness of active teams, NATS topology, cognitive
// config, and installed MCP servers before they process any message.
func (r *InternalToolRegistry) BuildContext(agentID, teamID, role string, teamInputs, teamDeliveries []string, currentInput string) string {
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

	// 6. Lead-agent temp memory channels (restart-safe working context)
	r.writeLeadTempMemory(&sb, agentID, teamID, role)

	// 7. Interaction Protocol
	sb.WriteString("### Interaction Protocol\n")
	sb.WriteString("**Pre-response** (before answering):\n")
	sb.WriteString("1. Check if past context is relevant → `recall` or `search_memory`\n")
	sb.WriteString("2. Check if specialist knowledge is needed → `consult_council`\n")
	sb.WriteString("3. Check if actionable work should be delegated → `delegate_task`\n")
	sb.WriteString("4. For software/dev tasks, prefer quick ephemeral code execution and bounded validation (`local_command`) before introducing new MCP dependencies\n")
	sb.WriteString("5. For web access tasks (search/site retrieval), default to coder-owned ephemeral web code first; use adaptive engine/query strategy\n")
	sb.WriteString("6. Check if installed MCP tools can fulfill remaining external integration requirements or provide easier execution\n")
	sb.WriteString("7. MCP Translation Procedure: map user intent -> operation/target/constraints/output, pick the narrowest installed MCP tool, then execute with minimal valid arguments\n")
	sb.WriteString("8. Check if data would benefit from visualization → `store_artifact` with type=chart\n\n")
	sb.WriteString("9. For explicit image requests, call `generate_image`; if user asks to keep the image, call `save_cached_image`\n\n")
	sb.WriteString("**Post-response** (after completing a task):\n")
	sb.WriteString("1. Store important learnings or decisions → `remember`\n")
	sb.WriteString("2. Store significant outputs → `store_artifact`\n")
	sb.WriteString("3. Distill successful complex approaches → `store_inception_recipe` (for future reuse)\n")
	sb.WriteString("4. Report actions taken and outcomes clearly to the user\n")
	sb.WriteString("5. For lead-agent workflows, checkpoint state via `temp_memory_write` and reload with `temp_memory_read`\n")
	sb.WriteString("6. Generated images are ephemeral cache by default (60m). Persist only on user request via `save_cached_image`\n")

	return sb.String()
}

func (r *InternalToolRegistry) isLeadAgent(agentID, teamID, role string) bool {
	id := strings.ToLower(agentID)
	tid := strings.ToLower(teamID)
	rl := strings.ToLower(role)

	if tid == "admin-core" || tid == "council-core" {
		return true
	}
	if id == "admin" || strings.HasPrefix(id, "council-") {
		return true
	}
	if strings.Contains(rl, "lead") || rl == "architect" || rl == "coder" || rl == "creative" || rl == "sentry" {
		return true
	}
	return false
}

func (r *InternalToolRegistry) writeLeadTempMemory(sb *strings.Builder, agentID, teamID, role string) {
	if !r.isLeadAgent(agentID, teamID, role) {
		return
	}

	sb.WriteString("### Persistent Temp Memory Channels (Restart-Safe)\n")
	sb.WriteString("Use these checkpoints to continue work consistently across provider/service restarts.\n")

	if r.mem != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		channels := []string{
			"interaction.contract",
			"lead.shared",
			fmt.Sprintf("lead.%s", agentID),
		}

		type channelDump struct {
			Channel string
			Entries []memory.TempMemoryEntry
		}
		dumps := []channelDump{}
		for _, ch := range channels {
			entries, err := r.mem.GetTempMemory(ctx, "default", ch, 3)
			if err != nil || len(entries) == 0 {
				continue
			}
			dumps = append(dumps, channelDump{Channel: ch, Entries: entries})
		}

		for _, dump := range dumps {
			sb.WriteString(fmt.Sprintf("- **%s**\n", dump.Channel))
			for _, e := range dump.Entries {
				content := strings.TrimSpace(e.Content)
				if len(content) > 220 {
					content = content[:220] + "..."
				}
				sb.WriteString(fmt.Sprintf("  - [%s] %s\n", e.OwnerAgentID, content))
			}
		}
	} else {
		sb.WriteString("- Memory backend unavailable; rely on current contract and checkpoint once memory is restored.\n")
	}

	sb.WriteString("Stability rules: preserve user interaction style, output shape, and action sequencing unless user explicitly changes intent.\n\n")
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
	sb.WriteString(fmt.Sprintf("- **Team command bus**: `swarm.team.%s.internal.command`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Team status bus**: `swarm.team.%s.signal.status`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Team result bus**: `swarm.team.%s.signal.result`\n", teamID))
	sb.WriteString(fmt.Sprintf("- **Legacy internal worker buses**: `swarm.team.%s.internal.trigger`, `swarm.team.%s.internal.response`\n", teamID, teamID))
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
		sb.WriteString("\n**MCP Translation Procedure (Use Current Inventory):**\n")
		sb.WriteString("1. Parse user request into operation + target + constraints + expected output.\n")
		sb.WriteString("2. Match against installed MCP tools listed below (source of truth).\n")
		sb.WriteString("3. Choose the narrowest matching tool and execute with minimal valid args.\n")
		sb.WriteString("4. If no match exists, report missing MCP dependency and required credentials.\n")

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

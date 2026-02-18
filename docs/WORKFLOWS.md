# Mycelis User Workflows

> **Authority**: This document defines the precise user-facing workflows for the Mycelis platform.
> Every UI component, API call, and state transition described here is the specification
> that frontend implementation must follow exactly.

---

## Table of Contents

1. [Workflow 1: Agent Catalogue — Browse & Configure Available Agents](#workflow-1-agent-catalogue)
2. [Workflow 2: MCP Tool Registry — Install & Manage External Tools](#workflow-2-mcp-tool-registry)
3. [Workflow 3: Team Builder — Compose Teams from Agents + Tools](#workflow-3-team-builder)
4. [Workflow 4: Mission Authoring — Negotiate & Commit a Mission](#workflow-4-mission-authoring)
5. [Workflow 5: Mission Monitoring — Live Execution Oversight](#workflow-5-mission-monitoring)
6. [Workflow 6: Governance — Human-in-the-Loop Trust Decisions](#workflow-6-governance)
7. [Workflow 7: Direct Chat — Conversational Interaction with Agents](#workflow-7-direct-chat)
8. [Workflow 8: Settings — Model Matrix & System Configuration](#workflow-8-settings)
9. [Workflow 9: Team Actuation & Output Viewer — Read-Only Mission Inspection](#workflow-9-team-actuation--output-viewer)
10. [Workflow 10: Three-Tier Memory — Fast Access, Learned Layer, Long-Term Recall](#workflow-10-three-tier-memory)

---

## Workflow 1: Agent Catalogue

**Purpose**: The user browses, creates, and configures reusable agent definitions. These are *templates* — not running instances. They define what an agent *is* (role, model, prompt, tools) before it is assigned to a team.

**Route**: `/catalogue`

**Backend Dependencies**:
- `GET /api/v1/catalogue/agents` — list all agent definitions (new endpoint)
- `POST /api/v1/catalogue/agents` — create a new agent definition (new endpoint)
- `PUT /api/v1/catalogue/agents/{id}` — update an agent definition (new endpoint)
- `DELETE /api/v1/catalogue/agents/{id}` — remove an agent definition (new endpoint)
- `GET /api/v1/mcp/tools` — list available MCP tools for tool binding (existing)
- `GET /api/v1/cognitive/matrix` — list available model profiles (existing)

**Database**: New table `agent_catalogue` (migration required)

### Step-by-Step User Flow

#### 1.1 — Browse the Catalogue

The user lands on `/catalogue`. The page shows a grid of agent cards, each displaying:

| Field | Source | Description |
|-------|--------|-------------|
| **Name** | `agent.name` | Human-readable label (e.g., "Code Reviewer") |
| **Role** | `agent.role` | Functional classification (e.g., `cognitive`, `sensory`, `actuation`) |
| **Model** | `agent.model` | Assigned LLM profile (e.g., `qwen2.5-coder:7b`) |
| **Tools** | `agent.tools[]` | Bound MCP tool names (e.g., `["read_file", "write_file"]`) |
| **Category Icon** | derived from role | Purple (cognitive), Cyan (sensory), Green (actuation), Muted (ledger) |
| **Status Badge** | computed | "Ready" if model + tools are all resolvable, "Incomplete" otherwise |

Cards are filterable by:
- **Category**: cognitive / sensory / actuation / ledger (chip toggle)
- **Text search**: name, role, or tool name substring match

The grid uses `cortex-surface` card backgrounds, `cortex-border` separators, matching the existing Vuexy Dark protocol.

**API Call on mount**: `GET /api/v1/catalogue/agents` → populates Zustand `catalogueAgents[]`

#### 1.2 — Create a New Agent Definition

The user clicks "+ New Agent" button in the catalogue header. A slide-out drawer opens from the right (same pattern as `BlueprintDrawer`). The form contains:

| Field | Input Type | Validation | Notes |
|-------|-----------|------------|-------|
| **Name** | text input | required, unique | Human label |
| **ID** | auto-generated | `kebab-case(name)` | Shown read-only, editable |
| **Role** | dropdown | required | `cognitive`, `sensory`, `actuation`, `ledger` |
| **System Prompt** | textarea (monospace) | optional | Multi-line, 4000 char max |
| **Model Profile** | dropdown | optional | Populated from `GET /api/v1/cognitive/matrix` |
| **Tools** | multi-select chips | optional | Populated from `GET /api/v1/mcp/tools`, shows tool name + server |
| **Inputs** | tag input | optional | NATS topic patterns this agent listens to |
| **Outputs** | tag input | optional | NATS topic patterns this agent publishes to |
| **Verification Strategy** | dropdown | optional | `none`, `semantic`, `empirical` |
| **Verification Rubric** | textarea | shown if strategy != none | Criteria lines |

**Submit**: `POST /api/v1/catalogue/agents` with the full definition → returns created agent with `id` → drawer closes → card appears in grid.

#### 1.3 — Edit an Existing Agent

The user clicks an agent card. The same drawer opens, pre-populated with the agent's current configuration. All fields are editable. The user modifies fields and clicks "Save".

**Submit**: `PUT /api/v1/catalogue/agents/{id}` → returns updated agent → card refreshes.

**Constraint**: If this agent is currently used in an active mission, a warning banner appears: "This agent is active in N mission(s). Changes apply to future deployments only."

#### 1.4 — Delete an Agent Definition

The user clicks the trash icon on an agent card (or in the edit drawer). A confirmation modal appears: "Delete agent '{name}'? This removes the template only — running instances are unaffected."

**Submit**: `DELETE /api/v1/catalogue/agents/{id}` → card removed from grid.

**Constraint**: Blocked if agent is referenced in a pending (non-committed) blueprint. Error: "Agent is referenced in draft blueprint '{name}'. Remove it from the blueprint first."

#### 1.5 — Preview Agent Capabilities

The user clicks "Test" on an agent card. A compact inline panel expands below the card showing:
- A single-turn inference test: the user types a prompt, the system calls `POST /api/v1/cognitive/infer` with the agent's model profile and system prompt prepended → response displayed in a monospace bubble.
- If the agent has tools bound, the response may contain `tool_call` JSON — rendered as a structured block showing tool name + arguments (read-only, not executed).

This is a dry-run. No NATS messages, no team context, no side effects.

---

## Workflow 2: MCP Tool Registry

**Purpose**: The user installs, manages, and tests external MCP tool servers. These provide tools (filesystem, GitHub, databases, Slack, etc.) that can be bound to agents.

**Route**: `/settings/tools` (sub-route of Settings)

**Backend Dependencies**:
- `POST /api/v1/mcp/install` — install a new MCP server (existing)
- `GET /api/v1/mcp/servers` — list servers with tools (existing)
- `DELETE /api/v1/mcp/servers/{id}` — remove a server (existing)
- `POST /api/v1/mcp/servers/{id}/tools/{tool}/call` — invoke a tool (existing)
- `GET /api/v1/mcp/tools` — flat tool list (existing)

### Step-by-Step User Flow

#### 2.1 — View Installed Servers

The page shows a vertical list of installed MCP servers, each as an expandable card:

| Field | Description |
|-------|-------------|
| **Name** | Server name (e.g., "filesystem") |
| **Transport** | `stdio` or `sse` badge |
| **Status** | `connected` (green dot), `error` (red dot), `installed` (yellow dot) |
| **Tool Count** | Number of discovered tools |
| **Error** | If status=error, the error message in red monospace |

Expanding a server card reveals its tools list — each tool showing: name, description (truncated), input schema preview (collapsed JSON).

**API Call on mount**: `GET /api/v1/mcp/servers` → populates Zustand `mcpServers[]`

#### 2.2 — Install a New Server

The user clicks "+ Install Server". A modal appears with:

| Field | Input Type | Shown When | Validation |
|-------|-----------|------------|------------|
| **Name** | text | always | required, unique |
| **Transport** | radio: `stdio` / `sse` | always | required |
| **Command** | text | transport=stdio | required for stdio |
| **Arguments** | tag input | transport=stdio | optional, array of strings |
| **Environment** | key-value editor | transport=stdio | optional |
| **URL** | text (url) | transport=sse | required for sse |
| **Headers** | key-value editor | transport=sse | optional |

The transport toggle dynamically shows/hides the relevant fields.

**Submit**: `POST /api/v1/mcp/install` → server installs + connects + discovers tools → modal closes → server card appears in list with its tools.

**Loading state**: The submit button shows a spinner with "Installing... Connecting... Discovering tools..." stages (SSE would be ideal here, but for now a single POST with loading state suffices — the connect+discover happens server-side).

#### 2.3 — Test a Tool

The user expands a server card, finds a tool, and clicks "Test". An inline panel opens below the tool showing:
- The tool's `input_schema` rendered as a dynamic form (JSON Schema → form fields)
- A "Run" button
- A response panel (initially empty)

The user fills in arguments and clicks "Run".

**Submit**: `POST /api/v1/mcp/servers/{serverId}/tools/{toolName}/call` with `{ arguments: {...} }` → response rendered in a monospace block.

This is a direct invocation. No agent context, no trust scoring. The user is testing the tool directly.

#### 2.4 — Remove a Server

The user clicks the trash icon on a server card. Confirmation: "Remove server '{name}'? This disconnects the live client and removes all cached tools."

**Submit**: `DELETE /api/v1/mcp/servers/{id}` → server removed from list.

**Side effect**: Any agents in the catalogue that reference tools from this server will show a "Tool unavailable" warning badge on their cards.

---

## Workflow 3: Team Builder

**Purpose**: The user composes a team by selecting agents from the catalogue, assigning them roles within the team, defining the team's I/O contracts (which NATS topics it listens to and delivers on), and configuring inter-agent data flow.

**Route**: `/wiring` (existing route, enhanced)

**Backend Dependencies**:
- `GET /api/v1/catalogue/agents` — list available agent definitions (new)
- `POST /api/v1/intent/negotiate` — LLM-assisted blueprint generation (existing)
- `POST /api/v1/intent/commit` — persist + activate (existing)
- `GET /api/v1/mcp/tools` — available tools for binding (existing)

### Step-by-Step User Flow

#### 3.1 — Enter the Wiring Page

The user navigates to `/wiring`. The existing Workspace layout loads:
- **Left panel** (360px): ArchitectChat — conversational intent negotiation
- **Right panel** (flex): CircuitBoard — ReactFlow canvas showing the team topology

The user has two paths to build a team:

**Path A: Conversational (existing)**
The user types a mission intent in the ArchitectChat (e.g., "I need a team to monitor my GitHub repos and summarize PRs"). The MetaArchitect LLM generates a `MissionBlueprint` and the CircuitBoard renders it as ghost-draft nodes.

**Path B: Manual Assembly (new)**
The user drags agents from a catalogue sidebar onto the CircuitBoard canvas, wires them together, and defines team metadata manually.

#### 3.2 — Path B: Manual Team Assembly

A new collapsible left sidebar (inside the CircuitBoard panel) shows the Agent Catalogue as a compact list. Each entry is draggable.

**Drag-and-drop flow**:
1. User drags an agent card from the sidebar onto the canvas
2. A new `agentNode` appears at the drop position with the agent's role, icon, and model profile
3. The node is in `ghost-draft` state (dashed border, dim)
4. User can drag multiple agents onto the canvas

**Wiring flow**:
1. User connects agent output handles to other agent input handles by drawing edges
2. Each edge represents a data flow (NATS topic subscription)
3. The edge type is `dataWire` (existing component)

**Team grouping**:
1. User selects multiple agents (shift-click or lasso)
2. Right-click → "Create Team" → a team group node wraps the selected agents
3. A small form appears to name the team and assign a role

**Team I/O**:
1. User clicks a team group node → a panel appears showing:
   - **Team Name**: editable text
   - **Team Role**: editable text
   - **Inputs**: tag input for NATS topic patterns
   - **Deliveries**: tag input for output topic patterns

#### 3.3 — Blueprint Preview

Whether via Path A (conversational) or Path B (manual), the user now has a blueprint represented as ghost-draft nodes on the CircuitBoard. A "Preview Blueprint" button in the canvas toolbar opens a slide-out showing the full JSON blueprint:

```json
{
  "mission_id": "auto-generated",
  "intent": "user-provided or LLM-generated",
  "teams": [
    {
      "name": "...",
      "role": "...",
      "agents": [
        { "id": "...", "role": "...", "model": "...", "tools": [...], ... }
      ]
    }
  ],
  "constraints": [...]
}
```

The user can edit this JSON directly (Monaco editor with syntax highlighting) or continue using the visual canvas.

#### 3.4 — Commit (Activate)

The user clicks "Activate Mission" in the ArchitectChat or the canvas toolbar. This triggers the existing `instantiateMission()` flow:

1. `POST /api/v1/intent/commit` with the blueprint
2. Backend persists to `missions`, `teams`, `service_manifests` tables
3. Backend activates teams in Soma (spawns agents, starts NATS subscriptions)
4. Ghost-draft nodes solidify (solid borders, green status dots)
5. Mission status transitions from `draft` → `active`

The user sees a confirmation message in the ArchitectChat: "Mission {id} instantiated. N teams, M agents now ACTIVE."

---

## Workflow 4: Mission Authoring

**Purpose**: Full end-to-end mission lifecycle from intent to activation. This is the primary "happy path" combining workflows 1-3.

**Route**: Starts at `/` (Mission Control), flows to `/wiring`, returns to `/`

### Step-by-Step User Flow

#### 4.1 — Express Intent

The user is on Mission Control (`/`). They click "NEW MISSION" → navigated to `/wiring`.

In the ArchitectChat, the user types their intent in natural language:

> "Build a research team that monitors Hacker News for AI papers, summarizes them, and posts a daily digest to our Slack channel."

#### 4.2 — Negotiate with Meta-Architect

The MetaArchitect LLM processes the intent and returns a `MissionBlueprint`:

```
Blueprint mission-hacker-news-digest generated.
2 teams, 5 agents.
1 constraint applied.
```

The CircuitBoard renders:
- **Team 1: "HN Monitor"** (sensory)
  - `hn-scraper` (sensor, polls HN API)
  - `paper-filter` (cognitive, filters AI-relevant posts)
- **Team 2: "Digest Composer"** (cognitive+actuation)
  - `summarizer` (cognitive, summarizes papers)
  - `digest-writer` (cognitive, composes daily digest)
  - `slack-poster` (actuation, posts to Slack via MCP tool)

The user sees agent nodes in ghost-draft state with dashed cyan borders.

#### 4.3 — Refine the Blueprint

The user can:
- **Chat refinement**: "Add a sentiment analysis step before the summary" → MetaArchitect regenerates with an additional agent
- **Visual editing**: Drag a new agent from the catalogue sidebar, re-wire edges
- **Agent tuning**: Click an agent node → edit its system prompt, model, or tool bindings in a side panel
- **Tool binding**: Click `slack-poster` → in the side panel, bind the `post_message` MCP tool from the installed Slack server

#### 4.4 — Review Constraints

The blueprint may include constraints (budget limits, trust thresholds, time windows). The user reviews them in the ArchitectChat or the blueprint preview panel. They can:
- Add constraints: "Only run during business hours"
- Remove constraints: Click the X on a constraint chip

#### 4.5 — Commit & Activate

The user clicks "Activate Mission". The commit flow (Workflow 3.4) executes. The user is shown a summary and optionally navigated back to Mission Control.

#### 4.6 — Post-Activation State

On Mission Control (`/`):
- The ActiveMissionsBar shows the new mission with team/agent counts
- The PriorityStream begins showing signals from the new teams
- The ActivityStream (SSE) shows agent heartbeats, cognitive events, tool calls
- The SensorLibrary shows any new sensor agents

---

## Workflow 5: Mission Monitoring

**Purpose**: The user observes running missions in real-time, inspects agent activity, reviews outputs, and intervenes when needed.

**Route**: `/` (Mission Control) and `/wiring` (drill-down)

**Backend Dependencies**:
- `GET /api/v1/missions` — list missions (existing)
- `GET /api/v1/stream` — SSE live feed (existing)
- `GET /api/v1/telemetry/compute` — system metrics (existing)
- `GET /api/v1/sensors` — sensor status (existing)
- `GET /api/v1/memory/sitreps` — situation reports (existing)
- `GET /api/v1/memory/search` — semantic memory search (existing)

### Step-by-Step User Flow

#### 5.1 — Mission Control Dashboard

The user lands on `/`. The layout (existing `MissionControl.tsx`) shows:

**Top**: TelemetryRow — four live sparkline cards (goroutines, heap MB, system MB, tokens/sec) polling `/api/v1/telemetry/compute` every 5 seconds.

**Below**: ActiveMissionsBar — horizontal scroll of active mission chips showing name, team count, agent count, status color.

**Main Grid** (3 columns):
- **Left**: SensorLibrary — grouped sensor subscriptions (toggle on/off)
- **Center Top**: PriorityStream — filtered high-priority signals (governance halts, errors, artifacts)
- **Center Bottom**: ManifestationPanel — pending team proposals
- **Right**: ActivityStream — full SSE feed (all events)

#### 5.2 — Inspect a Mission

The user clicks a mission chip in the ActiveMissionsBar. This navigates to `/wiring` with that mission's blueprint loaded on the CircuitBoard. Agent nodes show live status:

| Status | Visual |
|--------|--------|
| `online` | Green left-border accent, pulsing activity ring |
| `thinking` | Purple left-border, thought bubble with truncated last thought |
| `error` | Red left-border, error badge |
| `offline` | Dim, no accent |

The NatsWaterfall (bottom panel) filters to show only signals from this mission's teams.

#### 5.3 — Drill into a Team (Squad Room)

The user double-clicks a team group node → fractal navigation activates → `SquadRoom` component loads (existing). This shows:
- **Internal debate feed**: messages between agents on the team's internal bus
- **Proof-of-work artifacts**: verification results, rubric scores
- **Tool call log**: MCP tool invocations with arguments and results

Back button returns to the CircuitBoard overview.

#### 5.4 — Review SitReps

The user clicks "SitReps" in the Mission Control sidebar (or a dedicated tab). A panel shows Archivist-generated situation reports — periodic summaries of team activity, compressed by the cognitive engine. Each sitrep is a card with:
- Timestamp
- Team scope
- Summary text
- "Search Related" link → triggers semantic search in `context_vectors`

#### 5.5 — Search Memory

The user types a query in a search bar (top of SitReps panel or a global search): "What did the summarizer agent do with the last batch of papers?"

**Submit**: `GET /api/v1/memory/search?q=...` → returns semantically similar sitreps and log fragments → displayed as ranked cards.

---

## Workflow 6: Governance

**Purpose**: The Trust Economy requires human approval for low-trust agent outputs. The user reviews halted envelopes, inspects proof-of-work, and approves or rejects.

**Route**: `/wiring` (Zone D overlay) and `/` (PriorityStream)

**Backend Dependencies**:
- SSE `governance_halt` signals (existing)
- `POST /admin/approvals/{id}` with `{ action: "APPROVE" | "DENY" }` (existing)
- `PUT /api/v1/trust/threshold` — adjust autonomy threshold (existing)

### Step-by-Step User Flow

#### 6.1 — Trust Threshold Configuration

In the ArchitectChat (`/wiring`), the TrustSlider component shows a 0.0–1.0 range slider. The user adjusts it:
- **High (0.9)**: Almost all agent outputs auto-execute — minimal human review
- **Low (0.3)**: Most outputs halt for review — maximum human control
- **Default (0.7)**: Balanced — cognitive agents halt, sensory/actuation pass

**On change**: `PUT /api/v1/trust/threshold` syncs to backend → Overseer Governance Valve updates.

#### 6.2 — Governance Halt Notification

When an agent produces output with `TrustScore < AutoExecuteThreshold`:
1. The Overseer Governance Valve intercepts the CTSEnvelope
2. A `governance_halt` SSE event broadcasts to all connected clients
3. The frontend intercepts this in the Zustand SSE handler
4. The DeliverablesTray (bottom of CircuitBoard) gains a new pulsing green entry
5. The PriorityStream on Mission Control shows the halt

#### 6.3 — Review an Envelope

The user clicks a halted envelope in the DeliverablesTray. The GovernanceModal opens (Zone D overlay) showing:

**Left column**: The output content
- Rendered according to `content_type`: markdown (rendered), JSON (syntax-highlighted), text (monospace), image (rendered)
- Source agent name and team

**Right column**: Proof-of-work
- Verification method (`semantic` or `empirical`)
- Verification logs (command output or LLM rubric evaluation)
- Rubric score
- Pass/fail badge

**Bottom**: Two action buttons
- **APPROVE** (green): Releases the envelope for execution
- **REJECT** (red): Opens a text input for rejection reason, then discards the envelope

#### 6.4 — Approve or Reject

**Approve**: `POST /admin/approvals/{id}` with `{ action: "APPROVE" }` → envelope re-published to NATS → continues through the system.

**Reject**: `POST /admin/approvals/{id}` with `{ action: "DENY" }` → envelope discarded → agent notified (optional).

The envelope is removed from the DeliverablesTray. The GovernanceModal closes.

---

## Workflow 7: Direct Chat

**Purpose**: The user has a direct conversational interaction with the system. This is the primary human↔motherbrain interface. Today it's text-only; the API is designed to grow to audio and visual.

**Route**: `/wiring` (ArchitectChat, existing) and `/chat` (new dedicated route)

**Backend Dependencies**:
- `POST /api/v1/chat` — general chat endpoint (existing but basic)
- `POST /api/v1/intent/negotiate` — intent-specific chat (existing)
- `POST /api/v1/cognitive/infer` — raw inference (existing)
- `GET /api/v1/stream` — SSE for async responses (existing)

### Step-by-Step User Flow

#### 7.1 — Open Chat Interface

**Option A**: The ArchitectChat in `/wiring` — contextual to the current mission. Messages are scoped to intent negotiation and blueprint refinement.

**Option B** (new): A dedicated `/chat` route with a full-screen chat interface. This is the general-purpose motherbrain conversation. No mission context unless explicitly loaded.

The chat UI shows:
- Message history (scrollable, auto-scroll on new messages)
- User messages (right-aligned, `cortex-bg` background)
- System messages (left-aligned, `cortex-info` accent, Bot icon)
- Typing indicator (bouncing dots) when the system is processing
- Input bar at the bottom (text input + send button)

#### 7.2 — Send a Message

The user types a message and presses Enter or clicks Send.

**Frontend flow**:
1. Message appended to `chatHistory[]` in Zustand as `{ role: 'user', content: text }`
2. `isDrafting` set to `true` → typing indicator shown
3. `POST /api/v1/chat` with `{ content: text, context: { mission_id?, agent_id? } }`
4. Backend routes to the cognitive engine → response returned
5. Response appended to `chatHistory[]` as `{ role: 'architect', content: responseText }`
6. `isDrafting` set to `false`

#### 7.3 — Response Rendering

Today, responses are plain text rendered in a `<div>` bubble. The API is designed to eventually return typed responses:

```json
{
  "role": "architect",
  "content": "Here is the analysis...",
  "content_type": "text",
  "widgets": []
}
```

The `content_type` field determines how the message body is rendered:
- `text` → plain text in a monospace bubble (current behavior)
- `markdown` → rendered markdown (bold, code blocks, lists, tables)
- `code` → syntax-highlighted code block with language detection
- `json` → collapsible syntax-highlighted JSON tree
- `image` → inline image render (base64 or URL)
- `audio` → audio player widget (future)
- `chart` → data visualization widget (future)

The `widgets` array carries structured data for rich rendering (future — see section below).

#### 7.4 — Widget Protocol (Future Growth Path)

Each widget in the `widgets` array conforms to:

```typescript
interface Widget {
  type: 'code' | 'table' | 'chart' | 'image' | 'audio' | 'file' | 'action';
  title?: string;
  data: any;         // type-specific payload
  interactive?: boolean;  // can the user interact with this widget?
}
```

Example widget types:
- `code`: `{ language: "go", source: "...", filename?: "..." }` → Monaco editor view
- `table`: `{ columns: [...], rows: [...] }` → data table
- `chart`: `{ chartType: "line", series: [...] }` → chart component
- `image`: `{ url: "...", alt: "..." }` → image display
- `audio`: `{ url: "...", format: "mp3", duration_ms: 12340 }` → audio player
- `file`: `{ name: "...", content: "...", mime: "..." }` → file download/preview
- `action`: `{ label: "...", endpoint: "...", method: "POST" }` → clickable action button

The `UniversalRenderer` component (existing, currently basic) will be the rendering engine for these widgets.

#### 7.5 — Interaction Channels (Growth Path)

The chat API is designed to support multiple input/output channels:

| Channel | Input Method | Output Method | Status |
|---------|-------------|---------------|--------|
| **Text** | Keyboard typing | Rendered text + widgets | Active (current) |
| **Voice** | Microphone → speech-to-text | Text response (+ future TTS) | Planned next |
| **Audio** | Audio file upload | Audio analysis + text response | Future |
| **Visual** | Image/screenshot upload | Image analysis + text response | Future |

The API envelope supports this via a `channel` field:

```json
{
  "content": "...",
  "channel": "text",
  "media": null
}
```

For voice input, the frontend captures audio via `MediaRecorder` API, transcribes locally or sends to a speech-to-text endpoint, then sends the transcript as a text message with `channel: "voice"`. The backend treats it identically to text input. Voice output (TTS) is a future rendering concern — the backend response is the same.

---

## Workflow 8: Settings

**Purpose**: System-wide configuration — model matrix, MCP tools, trust economy, sensor management.

**Route**: `/settings` (existing) with sub-routes

**Sub-routes**:
- `/settings` — General settings (existing)
- `/settings/brain` — Model matrix configuration (existing)
- `/settings/tools` — MCP Tool Registry (Workflow 2, new)
- `/settings/trust` — Trust Economy configuration (new dedicated page)

### Step-by-Step User Flow

#### 8.1 — Model Matrix (`/settings/brain`)

Existing page showing `ModelHealth` component. The user sees:
- List of configured LLM providers (from `llm_providers` table + `cognitive.yaml`)
- Health probe status for each (online/offline)
- Profile → provider mapping (architect → ollama, coder → ollama, etc.)

The user can:
- Edit provider endpoints
- Change profile assignments
- Trigger re-discovery probe

#### 8.2 — MCP Tools (`/settings/tools`)

Full MCP Tool Registry interface (Workflow 2). Manages external tool servers and their discovered tools.

#### 8.3 — Trust Configuration (`/settings/trust`)

Dedicated page for the Trust Economy:
- **Global Threshold**: slider (same as TrustSlider but full-page context with explanation)
- **Per-Category Defaults**: table showing trust scores by node category (sensory=1.0, cognitive=0.5, etc.) — read-only for now, editable in future
- **Governance Log**: recent approval/rejection history from the Overseer

---

## Cross-Cutting Concerns

### State Management (Zustand)

All workflow state lives in `useCortexStore`. New state slices needed:

```typescript
// Agent Catalogue
catalogueAgents: CatalogueAgent[];
isFetchingCatalogue: boolean;
selectedCatalogueAgent: CatalogueAgent | null;

// MCP Registry
mcpServers: MCPServerWithTools[];
isFetchingMCPServers: boolean;

// Direct Chat (dedicated route)
directChatHistory: ChatMessage[];
isDirectChatDrafting: boolean;
```

### SSE Signal Dispatch

All live data flows through the existing SSE stream at `/api/v1/stream`. The Zustand `initializeStream()` handler already intercepts and classifies signals. New signal types to support:

| Signal Type | Source | Action |
|-------------|--------|--------|
| `agent.heartbeat` | Agent heartbeat loop | Update node status in CircuitBoard |
| `cognitive` | Agent inference | Show thought bubble on node |
| `tool_call` | Agent MCP tool use | Show tool invocation in NatsWaterfall |
| `tool_result` | MCP pool response | Show result in NatsWaterfall + SquadRoom |
| `artifact` | Agent output | Push to DeliverablesTray |
| `governance_halt` | Overseer | Push to DeliverablesTray + PriorityStream |
| `sensor_data` | SensorAgent | Update SensorLibrary feed |
| `mcp_connected` | MCP Pool | Update server status in Settings |
| `mcp_disconnected` | MCP Pool | Update server status in Settings |

### Navigation Map

```
/                     → Mission Control (dashboard)
/wiring               → Workspace (ArchitectChat + CircuitBoard + NatsWaterfall)
/chat                 → Direct Chat (new)
/catalogue            → Agent Catalogue (new)
/settings             → General Settings
/settings/brain       → Model Matrix
/settings/tools       → MCP Tool Registry (new)
/settings/trust       → Trust Configuration (new)
/telemetry            → Telemetry deep-dive
/approvals            → Approval queue
/registry             → Connector marketplace
/marketplace          → Template marketplace
/missions/{id}/teams  → Team Actuation Viewer (new)
/memory               → Memory Explorer (new)
```

### API Endpoints Summary

**Existing** (no backend changes needed):
- `POST /api/v1/intent/negotiate` — blueprint generation
- `POST /api/v1/intent/commit` — mission activation
- `POST /api/v1/intent/seed/symbiotic` — seed sensor team
- `GET /api/v1/missions` — list missions
- `GET /api/v1/stream` — SSE live feed
- `GET /api/v1/telemetry/compute` — system metrics
- `GET /api/v1/cognitive/matrix` — model config
- `POST /api/v1/cognitive/infer` — raw inference
- `POST /api/v1/chat` — general chat
- `GET/PUT /api/v1/trust/threshold` — trust threshold
- `GET /api/v1/memory/search` — semantic search
- `GET /api/v1/memory/sitreps` — situation reports
- `GET /api/v1/sensors` — sensor status
- `GET/POST /api/v1/proposals` — team proposals
- `POST /api/v1/mcp/install` — install MCP server
- `GET /api/v1/mcp/servers` — list MCP servers
- `DELETE /api/v1/mcp/servers/{id}` — remove MCP server
- `POST /api/v1/mcp/servers/{id}/tools/{tool}/call` — tool invocation
- `GET /api/v1/mcp/tools` — flat tool list

**New** (backend changes required):
- `GET /api/v1/catalogue/agents` — list agent definitions
- `POST /api/v1/catalogue/agents` — create agent definition
- `PUT /api/v1/catalogue/agents/{id}` — update agent definition
- `DELETE /api/v1/catalogue/agents/{id}` — remove agent definition
- `GET /api/v1/artifacts` — list artifacts (filterable by mission_id, team_id, agent_id)
- `GET /api/v1/artifacts/{id}` — get single artifact detail
- `POST /api/v1/artifacts` — store new artifact
- `PUT /api/v1/artifacts/{id}/status` — update artifact governance status

### Database Changes Required

**New table**: `agent_catalogue`

```sql
CREATE TABLE IF NOT EXISTS agent_catalogue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,                          -- cognitive, sensory, actuation, ledger
    system_prompt TEXT,
    model TEXT,                                  -- LLM profile name
    tools JSONB DEFAULT '[]',                    -- MCP tool name bindings
    inputs JSONB DEFAULT '[]',                   -- NATS topic patterns
    outputs JSONB DEFAULT '[]',                  -- NATS topic patterns
    verification_strategy TEXT,                  -- none, semantic, empirical
    verification_rubric JSONB DEFAULT '[]',      -- rubric lines
    validation_command TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_agent_catalogue_role ON agent_catalogue(role);
```

---

## Workflow 9: Team Actuation & Output Viewer

**Purpose**: The user inspects active missions from Mission Control, drills into a specific mission's teams, observes real-time agent actuation, and reviews artifacts (code, documents, images, data) produced by agents. This is a **read-only** workflow — no edits, no commands, pure observation and retrieval.

**Route**: `/missions/{id}/teams` (linked from Mission Control)

**Backend Dependencies**:
- `GET /api/v1/missions` — list missions (existing)
- `GET /api/v1/stream` — SSE live feed for agent heartbeats and signals (existing)
- `GET /api/v1/artifacts?mission_id={id}` — list artifacts for a mission (new)
- `GET /api/v1/artifacts?team_id={id}` — list artifacts for a team (new)
- `GET /api/v1/artifacts?agent_id={id}` — list artifacts for an agent (new)
- `GET /api/v1/artifacts/{id}` — get single artifact detail with content (new)
- `PUT /api/v1/artifacts/{id}/status` — approve/reject artifact (new — governance action)
- `GET /api/v1/memory/sitreps` — team-scoped situation reports (existing)

**Database**: `artifacts` table + `agent_state` table (migration 018)

### Step-by-Step User Flow

#### 9.1 — Mission Control Entry Point

The user is on `/` (Mission Control). The dashboard shows a **Teams Summary** card in the right column:

| Column | Content |
|--------|---------|
| **Active Teams** | Count of teams with at least one agent online |
| **Total Agents** | Count of agents across all active missions |
| **Recent Outputs** | Count of artifacts produced in last hour |

Below the summary, a **Recent Missions** list shows:

| Field | Source | Description |
|-------|--------|-------------|
| **Mission Name** | `mission.name` | From committed blueprint |
| **Status** | `mission.status` | `active` / `completed` / `failed` |
| **Team Count** | computed | Number of teams in this mission |
| **Last Activity** | SSE heartbeat age | Time since last agent signal |
| **View** | link | Arrow icon → navigates to `/missions/{id}/teams` |

**Empty State**: "No active missions. Commit a blueprint from the Wiring page to get started."

**Data Flow**:
```
Page Load → GET /api/v1/missions → filter status=active → render list
SSE stream → agent.heartbeat events → update "Last Activity" per team
```

#### 9.2 — Team Actuation View

Clicking "View" navigates to `/missions/{id}/teams`. This page shows a **two-panel layout**:

**Left Panel (40%): Team Roster**

A vertical list of teams in this mission, each rendered as a card:

| Field | Source | Description |
|-------|--------|-------------|
| **Team Name** | `team.name` | From blueprint manifest |
| **Status Dot** | SSE heartbeat | Green (active), Yellow (stale >15s), Red (offline >60s) |
| **Agent Count** | `team.agents.length` | "3 agents" |
| **Role Icons** | `agent.role` | Cognitive (purple), Sensory (cyan), Actuation (green), Ledger (muted) |
| **Output Count** | artifact query | Badge showing artifact count for this team |

Clicking a team card selects it and updates the right panel.

**Right Panel (60%): Agent Activity Feed**

When a team is selected, this panel shows:

**Agent Row** — horizontal cards for each agent in the selected team:

| Field | Source | Description |
|-------|--------|-------------|
| **Agent ID** | `agent.id` | Short ID display |
| **Role** | `agent.role` | Icon + label |
| **Status** | SSE heartbeat | Online/Thinking/Offline |
| **Current Action** | Last SSE signal | "Processing...", "Tool: read_file", "Idle" |
| **Trust Score** | Last CTSEnvelope | Numeric 0.0-1.0 with color gradient |

**Activity Timeline** — below the agent row, a reverse-chronological feed:

| Entry Type | Visual | Source |
|------------|--------|--------|
| **Cognitive** | Purple left-border, brain icon | SSE `cognitive` signals |
| **Tool Call** | Blue left-border, wrench icon | SSE `tool_call` signals |
| **Tool Result** | Green/red left-border, check/x icon | SSE `tool_result` signals |
| **Output** | Gold left-border, file icon | Artifact stored event |
| **Governance Halt** | Red left-border, shield icon | SSE `governance_halt` signal |

Each timeline entry shows: timestamp, agent source, message summary (truncated to 200 chars), expandable detail.

**Data Flow**:
```
Team Select → GET /api/v1/artifacts?team_id={id} → populate output count
SSE stream → filter by team agent IDs → populate Activity Timeline
Heartbeat timeout → mark agent row status as stale/offline
```

#### 9.3 — Artifact Viewer

Clicking an output entry (gold-bordered) or the "Outputs" tab opens the **Artifact Viewer** within the right panel:

**Artifact List** — filterable grid:

| Field | Source | Description |
|-------|--------|-------------|
| **Title** | `artifact.title` | Human-readable name |
| **Type** | `artifact.artifact_type` | Badge: code/document/image/audio/data/file/chart |
| **Agent** | `artifact.agent_id` | Which agent produced it |
| **Trust** | `artifact.trust_score` | Score from producing envelope |
| **Status** | `artifact.status` | pending/approved/rejected/archived |
| **Created** | `artifact.created_at` | Relative timestamp |

**Filter Controls**:
- Type chips: All / Code / Document / Image / Data / File
- Status chips: All / Pending / Approved / Rejected
- Agent dropdown: All / specific agent

Clicking an artifact opens a **detail view** (slide-over or modal):

**For `code` artifacts**:
- Syntax-highlighted code block (language detected from content_type)
- Copy button
- Line numbers

**For `document` artifacts**:
- Rendered markdown content
- Scroll container

**For `image` artifacts**:
- Image preview (loaded from `/data/artifacts/{file_path}`)
- Metadata: dimensions, file size

**For `data` artifacts** (JSON/CSV):
- Formatted JSON viewer (collapsible tree)
- Or tabular CSV view

**For `file` artifacts** (generic binary):
- File icon, name, size
- Download link

**Governance Actions** (on each artifact detail):
- "Approve" button → `PUT /api/v1/artifacts/{id}/status` with `{"status":"approved"}`
- "Reject" button → same endpoint with `{"status":"rejected"}`
- Only shown when `artifact.status == "pending"`

**Data Flow**:
```
Artifact List → GET /api/v1/artifacts?team_id={id} → render grid
Click artifact → GET /api/v1/artifacts/{id} → render detail view
Approve/Reject → PUT /api/v1/artifacts/{id}/status → update badge
```

#### 9.4 — Mission-Level Summary

A "Summary" tab at the top of the mission page shows aggregate metrics:

| Metric | Source | Display |
|--------|--------|---------|
| **Total Artifacts** | artifact count | Numeric |
| **Pending Review** | status=pending count | Badge (amber if > 0) |
| **Cognitive Cycles** | SSE cognitive event count | Numeric |
| **Tool Invocations** | SSE tool_call count | Numeric |
| **Governance Halts** | SSE governance_halt count | Badge (red if > 0) |
| **Mission Duration** | `activated_at` → now | HH:MM:SS |
| **Latest SitRep** | `GET /api/v1/memory/sitreps` | Expandable card with summary text |

---

## Workflow 10: Three-Tier Memory

**Purpose**: The user explores the platform's memory architecture across three speed tiers — hot operational memory (real-time), warm structured storage (queryable), and cold vector memory (semantic recall). The cold tier enables **memory-driven actuation**: past experience encoded as vectors can be retrieved by agents during mission execution to inform decisions, trigger learned behaviors, and avoid repeated mistakes.

**Route**: `/memory`

**Backend Dependencies**:
- `GET /api/v1/stream` — SSE live event buffer (hot tier, existing)
- `GET /api/v1/memory/sitreps` — compressed situation reports (warm tier, existing)
- `GET /api/v1/memory/search?q={query}` — semantic vector search (cold tier, existing)
- `GET /api/v1/artifacts` — agent outputs with metadata (warm tier, existing)
- `GET /api/v1/telemetry/compute` — system performance metrics (existing)

**No new backend changes required** — this workflow exposes existing memory layers through a unified UI.

### Architecture Overview

```
┌───────────────────────────────────────────────────────────┐
│                    THREE-TIER MEMORY                       │
├─────────────┬──────────────────┬──────────────────────────┤
│   HOT       │   WARM           │   COLD                   │
│   (Live)    │   (Structured)   │   (Semantic)             │
│             │                  │                          │
│ eventBuffer │ log_entries      │ context_vectors          │
│ 20-event    │ sitreps          │ pgvector 768-dim         │
│ threshold   │ artifacts        │ nomic-embed-text         │
│ flush       │ agent_state      │ cosine distance          │
│             │ agent_catalogue  │                          │
│ ↓           │ ↓                │ ↓                        │
│ SSE Stream  │ PostgreSQL       │ pgvector + embeddings    │
│ NatsWater-  │ REST queries     │ Semantic search          │
│ fall panel  │ Paginated lists  │ RAG retrieval            │
│             │                  │ Actuation triggers       │
└─────────────┴──────────────────┴──────────────────────────┘
```

### Step-by-Step User Flow

#### 10.1 — Memory Explorer Landing

The user navigates to `/memory`. The page shows a **three-column layout** with one column per memory tier:

**Column 1 — Hot Memory (Live Feed)**

A real-time stream of the last N events from the SSE connection. This is a compact version of the NatsWaterfall panel, scoped to memory-relevant signals:

| Field | Source | Description |
|-------|--------|-------------|
| **Timestamp** | SSE event | HH:MM:SS.ms format |
| **Source** | `source_agent_id` | Agent that produced the signal |
| **Type** | signal classification | cognitive / tool_call / tool_result / heartbeat |
| **Preview** | message body | First 80 chars of content |

- Auto-scrolls (pause on hover)
- Event count badge: "247 events buffered"
- Flush indicator: shows when the 20-event threshold triggers a flush to warm storage
- Color coding: matches NatsWaterfall direction classification (input=cyan, output=green, internal=muted)

**Column 2 — Warm Memory (Structured)**

Tabbed view with three sub-panels:

**Tab: SitReps**
- Reverse-chronological list of compressed situation reports
- Each card: timestamp, team scope, summary text (first 200 chars), expand to read full
- Source: `GET /api/v1/memory/sitreps`

**Tab: Artifacts**
- Recent agent outputs across all missions
- Each card: title, type badge, agent source, trust score, status
- Filterable by type (code/document/image/data)
- Source: `GET /api/v1/artifacts?limit=50`

**Tab: Agent State**
- Key-value pairs persisted by agents during execution
- Grouped by agent_id
- Shows: key, value (truncated), mission scope, last updated, TTL remaining (if set)
- Source: Direct DB read (future API — display placeholder for now)

**Column 3 — Cold Memory (Semantic)**

The semantic search and vector recall panel:

**Search Bar** — prominent at top:
- Text input with brain icon
- "Search long-term memory..." placeholder
- Debounced 500ms → `GET /api/v1/memory/search?q={query}`

**Results** — ranked by cosine similarity:

| Field | Source | Description |
|-------|--------|-------------|
| **Relevance** | cosine distance | Percentage bar (0-100%) |
| **Content** | sitrep summary | The embedded text chunk |
| **Source** | team/timestamp | Which team, when compressed |
| **Vector ID** | `context_vectors.id` | For reference/debug |

**Memory Insights** — below results, static cards showing:
- Total vectors stored (count)
- Embedding model: `nomic-embed-text`
- Dimension: 768
- Last embedding: timestamp of most recent vector

#### 10.2 — Memory-Driven Actuation (Conceptual)

This section explains to the user how cold memory drives agent behavior. It's displayed as an **info panel** (collapsible) at the top of the Cold Memory column:

**How Memory Drives Action:**

> When an agent receives a mission trigger, the cognitive router can retrieve relevant
> vectors from long-term memory before inference. This means:
>
> - **Pattern Recognition**: If a similar task was executed before, the agent sees the
>   previous SitRep summary and adjusts its approach.
> - **Error Avoidance**: Failed tool calls or rejected artifacts are embedded — agents
>   can learn from past governance rejections.
> - **Context Enrichment**: Sensor data summaries from previous missions enrich the
>   agent's context window without re-polling external sources.

The retrieval path:

```
Agent trigger → Router.InferWithContract()
                  ↓
              Router checks agent.Tools for "memory_search"
                  ↓
              If bound: SemanticSearch(agent's last message, top_k=5)
                  ↓
              Prepend retrieved context to system prompt
                  ↓
              LLM inference with enriched context
```

This is **not a user action** — it happens automatically when agents have `memory_search` in their tool bindings. The user configures this in the Agent Catalogue (Workflow 1) by adding `memory_search` to the agent's tool list.

#### 10.3 — Cross-Tier Navigation

Each memory tier links to the others:

| From | To | Action |
|------|------|--------|
| Hot event | Warm SitRep | "View SitRep" link on cognitive events |
| Warm SitRep | Cold search | "Find Related" button triggers vector search with SitRep text |
| Warm Artifact | Hot stream | "View Timeline" link filters hot feed to artifact's agent |
| Cold result | Warm SitRep | Click result card navigates to the source SitRep |
| Cold result | Mission view | "View Mission" link navigates to `/missions/{id}/teams` |

**Data Flow Summary**:
```
SSE Stream (hot) ──→ eventBuffer flush ──→ log_entries (warm)
                                              ↓
                                    Archivist compress (5min)
                                              ↓
                                        sitreps (warm)
                                              ↓
                                    auto-embed (nomic-embed-text)
                                              ↓
                                     context_vectors (cold)
                                              ↓
                                    SemanticSearch() ←── Agent RAG query
                                              ↓
                                    Enriched context → LLM inference → Actuation
```

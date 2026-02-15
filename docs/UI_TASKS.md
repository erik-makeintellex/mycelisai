# UI Implementation Task Set

> Generated from [WORKFLOWS.md](WORKFLOWS.md). Each task is atomic, testable, and
> references the exact workflow step it implements.

---

## Prerequisites (Backend)

These backend tasks MUST be completed before the corresponding UI tasks can be functionally tested.

### B1: Agent Catalogue Migration
- **File**: `core/migrations/017_agent_catalogue.up.sql`
- **Deliverable**: `agent_catalogue` table with id, name, role, system_prompt, model, tools (JSONB), inputs (JSONB), outputs (JSONB), verification fields, timestamps
- **Verify**: `inv db.migrate` succeeds, `inv db.status` shows table

### B2: Agent Catalogue Service
- **File**: `core/internal/catalogue/service.go`
- **Deliverable**: `Service` struct with `List()`, `Get()`, `Create()`, `Update()`, `Delete()` methods
- **Pattern**: Follow `internal/mcp/service.go` (same DB CRUD pattern)
- **Verify**: Unit test with mock DB

### B3: Agent Catalogue HTTP Handlers
- **File**: `core/internal/server/catalogue.go`
- **Deliverable**: 4 handlers: `handleListCatalogue`, `handleCreateCatalogue`, `handleUpdateCatalogue`, `handleDeleteCatalogue`
- **Wiring**: Register routes in `admin.go` RegisterRoutes, add `Catalogue *catalogue.Service` field to AdminServer
- **Verify**: `curl` smoke tests against running core

### B4: Wire Catalogue in main.go
- **File**: `core/cmd/server/main.go`
- **Deliverable**: Initialize `catalogue.NewService(sharedDB)` and pass to `NewAdminServer()`
- **Verify**: `inv core.build` passes

### B5: Artifacts Migration + Service

- **File**: `core/migrations/018_artifacts.up.sql`, `core/internal/artifacts/service.go`
- **Deliverable**: `artifacts` table (id, mission_id FK, team_id FK, agent_id, artifact_type, title, content_type, content, file_path, file_size_bytes, metadata JSONB, trust_score, status, created_at) + `agent_state` table (agent_id+key PK, value, mission_id FK, expires_at)
- **Service methods**: `Store()`, `ListByMission()`, `ListByTeam()`, `ListByAgent()`, `ListRecent()`, `Get()`, `UpdateStatus()`
- **Verify**: `inv db.migrate` succeeds

### B6: Artifacts HTTP Handlers + Wiring

- **File**: `core/internal/server/artifacts.go`, `core/internal/server/admin.go`, `core/cmd/server/main.go`
- **Deliverable**: 4 handlers (`handleListArtifacts`, `handleGetArtifact`, `handleStoreArtifact`, `handleUpdateArtifactStatus`), `Artifacts *artifacts.Service` field on AdminServer, wired in main.go
- **Routes**: `GET /api/v1/artifacts`, `GET /api/v1/artifacts/{id}`, `POST /api/v1/artifacts`, `PUT /api/v1/artifacts/{id}/status`
- **Verify**: `inv core.build` passes

### B7: K8s Data Volume for Agent Outputs

- **Files**: `charts/mycelis-core/templates/data-pvc.yaml`, `charts/mycelis-core/templates/deployment.yaml`, `charts/mycelis-core/values.yaml`
- **Deliverable**: 2Gi PVC mounted at `/data/artifacts`, DATA_DIR env var, readOnlyRootFilesystem constraint respected
- **Verify**: `helm template` renders correctly

---

## Track 1: Agent Catalogue Page (Workflow 1)

### T1.1: Catalogue Page Shell
- **Workflow Step**: 1.1
- **Route**: `/catalogue`
- **Files**: `app/catalogue/page.tsx`, `components/catalogue/CatalogueGrid.tsx`
- **Deliverable**: New page at `/catalogue` with Vuexy Dark header ("AGENT CATALOGUE"), empty grid placeholder
- **Dependencies**: Navigation link in Header.tsx
- **Acceptance**: Page renders, nav link works, dark theme correct

### T1.2: Zustand Catalogue Slice
- **Workflow Step**: 1.1
- **File**: `store/useCortexStore.ts` (extend)
- **Deliverable**: New state slice:
  ```typescript
  catalogueAgents: CatalogueAgent[];
  isFetchingCatalogue: boolean;
  selectedCatalogueAgent: CatalogueAgent | null;
  fetchCatalogue: () => Promise<void>;
  createCatalogueAgent: (agent: CreateAgentInput) => Promise<void>;
  updateCatalogueAgent: (id: string, agent: UpdateAgentInput) => Promise<void>;
  deleteCatalogueAgent: (id: string) => Promise<void>;
  ```
- **Type**: `CatalogueAgent` mirrors backend `agent_catalogue` row
- **Dependencies**: B3 (API endpoints)
- **Acceptance**: Store compiles, types exported

### T1.3: Agent Card Component
- **Workflow Step**: 1.1
- **File**: `components/catalogue/AgentCard.tsx`
- **Deliverable**: Card component showing: name, role category icon (cognitive=purple/sensory=cyan/actuation=green/ledger=muted), model badge, tool count chip, status badge (Ready/Incomplete)
- **Style**: `cortex-surface` bg, `cortex-border` border, `rounded-xl shadow-lg`, matches AgentNode accent colors
- **Dependencies**: T1.2
- **Acceptance**: Card renders with mock data, icons correct

### T1.4: Catalogue Grid with Filters
- **Workflow Step**: 1.1
- **File**: `components/catalogue/CatalogueGrid.tsx`
- **Deliverable**: Grid of AgentCards, filter bar with:
  - Category chip toggles (cognitive/sensory/actuation/ledger)
  - Text search input (filters name, role, tool names)
  - "+ New Agent" button in header
- **Dependencies**: T1.3, T1.2
- **Acceptance**: Grid populates from Zustand, filters work client-side, empty state message

### T1.5: Agent Editor Drawer
- **Workflow Step**: 1.2, 1.3
- **File**: `components/catalogue/AgentEditorDrawer.tsx`
- **Deliverable**: Right slide-out drawer (same pattern as `BlueprintDrawer.tsx`) with form fields:
  - Name (text, required)
  - ID (auto-kebab, shown read-only)
  - Role (dropdown: cognitive/sensory/actuation/ledger)
  - System Prompt (textarea, monospace, 4000 char limit)
  - Model Profile (dropdown, populated from `GET /api/v1/cognitive/matrix`)
  - Tools (multi-select chips, populated from `GET /api/v1/mcp/tools`)
  - Inputs (tag input, NATS topic patterns)
  - Outputs (tag input, NATS topic patterns)
  - Verification Strategy (dropdown: none/semantic/empirical)
  - Verification Rubric (textarea, shown if strategy != none)
  - Save / Cancel buttons
- **Mode**: Create (empty form) or Edit (pre-populated from selected agent)
- **Dependencies**: T1.2, existing API endpoints
- **Acceptance**: Create flow calls POST, edit flow calls PUT, drawer opens/closes, validation works

### T1.6: Agent Delete with Confirmation
- **Workflow Step**: 1.4
- **File**: `components/catalogue/AgentCard.tsx` (add delete button) or inline in drawer
- **Deliverable**: Trash icon on card, confirmation modal: "Delete agent '{name}'?"
- **Dependencies**: T1.2, T1.3
- **Acceptance**: DELETE call fires, card removed from grid, modal renders

### T1.7: Agent Capability Preview
- **Workflow Step**: 1.5
- **File**: `components/catalogue/AgentPreview.tsx`
- **Deliverable**: Expandable panel below agent card with:
  - Single-line prompt input
  - "Test" button
  - Response display (monospace bubble)
  - Calls `POST /api/v1/cognitive/infer` with agent's model profile + system prompt prepended
- **Dependencies**: T1.3, existing `/api/v1/cognitive/infer`
- **Acceptance**: Inference call succeeds, response displays, loading state shows

---

## Track 2: MCP Tool Registry Page (Workflow 2)

### T2.1: Tools Settings Page
- **Workflow Step**: 2.1
- **Route**: `/settings/tools`
- **Files**: `app/settings/tools/page.tsx`, `components/settings/MCPServerList.tsx`
- **Deliverable**: New sub-route under settings with server list view
- **Dependencies**: Navigation tab in settings layout
- **Acceptance**: Page renders at `/settings/tools`, dark theme

### T2.2: Zustand MCP Slice
- **Workflow Step**: 2.1
- **File**: `store/useCortexStore.ts` (extend)
- **Deliverable**: New state slice:
  ```typescript
  mcpServers: MCPServerWithTools[];
  isFetchingMCPServers: boolean;
  fetchMCPServers: () => Promise<void>;
  installMCPServer: (config: MCPInstallInput) => Promise<void>;
  deleteMCPServer: (id: string) => Promise<void>;
  callMCPTool: (serverId: string, toolName: string, args: Record<string, any>) => Promise<any>;
  ```
- **Dependencies**: Existing MCP API endpoints
- **Acceptance**: Store compiles, fetch populates from backend

### T2.3: MCP Server Card (Expandable)
- **Workflow Step**: 2.1
- **File**: `components/settings/MCPServerCard.tsx`
- **Deliverable**: Expandable card showing:
  - Name, transport badge (stdio/sse), status dot (green/red/yellow), tool count
  - Expand reveals: tool list with name, description, input schema preview
  - Delete button (trash icon)
- **Style**: Vuexy Dark, monospace for tool names
- **Dependencies**: T2.2
- **Acceptance**: Card renders, expands/collapses, status colors correct

### T2.4: Install Server Modal
- **Workflow Step**: 2.2
- **File**: `components/settings/MCPInstallModal.tsx`
- **Deliverable**: Modal with:
  - Name (text, required)
  - Transport (radio: stdio/sse)
  - Conditional fields: command+args+env (stdio) OR url+headers (sse)
  - "Install" button with loading spinner
- **Dependencies**: T2.2
- **Acceptance**: Install flow calls POST, modal closes, server appears in list

### T2.5: Tool Test Panel
- **Workflow Step**: 2.3
- **File**: `components/settings/ToolTestPanel.tsx`
- **Deliverable**: Inline panel below a tool in the expanded server card:
  - Dynamic form generated from tool's `input_schema` (JSON Schema → form fields)
  - "Run" button
  - Response display (monospace, collapsible)
- **Dependencies**: T2.3, existing tool call endpoint
- **Acceptance**: Tool call succeeds, response renders, error states handled

---

## Track 3: Team Builder Enhancement (Workflow 3)

### T3.1: Catalogue Sidebar in CircuitBoard
- **Workflow Step**: 3.2
- **File**: `components/workspace/CatalogueSidebar.tsx`
- **Deliverable**: Collapsible left sidebar inside the CircuitBoard panel:
  - Compact agent list from catalogue (name + role icon)
  - Each entry is draggable (HTML5 drag API or react-dnd)
  - Toggle button to show/hide
- **Dependencies**: T1.2 (catalogue state)
- **Acceptance**: Sidebar renders inside CircuitBoard, agents listed, toggle works

### T3.2: Drag-and-Drop Agent Placement
- **Workflow Step**: 3.2
- **File**: `components/workspace/CircuitBoard.tsx` (modify)
- **Deliverable**: Handle drop events on the ReactFlow canvas:
  - On drop: create a new `agentNode` at drop coordinates
  - Node data populated from dropped catalogue agent (role, model, tools)
  - Node is in `ghost-draft` state
- **Dependencies**: T3.1
- **Acceptance**: Drag agent from sidebar, drop on canvas, node appears at correct position

### T3.3: Agent Node Side Panel
- **Workflow Step**: 3.3 (refine)
- **File**: `components/workspace/AgentNodePanel.tsx`
- **Deliverable**: Side panel that opens when clicking an agent node:
  - Shows: name, role, system prompt (editable), model (dropdown), tools (multi-select)
  - Changes update the node data in Zustand
  - "Remove from Canvas" button
- **Dependencies**: Existing AgentNode component
- **Acceptance**: Panel opens on node click, edits persist, removal works

### T3.4: Team Group Creation
- **Workflow Step**: 3.2
- **File**: `components/workspace/CircuitBoard.tsx` (modify)
- **Deliverable**: Multi-select agents (shift+click or lasso) → right-click context menu → "Create Team":
  - Wraps selected nodes in a team group node
  - Opens mini form: team name, team role
  - Group node styled as existing team groups
- **Dependencies**: T3.2
- **Acceptance**: Group creation works, visual wrapping correct, metadata editable

### T3.5: Blueprint JSON Preview
- **Workflow Step**: 3.3
- **File**: `components/workspace/BlueprintPreview.tsx`
- **Deliverable**: Slide-out panel (from canvas toolbar button "Preview") showing:
  - Full blueprint JSON with syntax highlighting
  - Read-only by default, "Edit JSON" toggle makes it editable (textarea with monospace)
  - "Apply" button to sync JSON edits back to canvas
- **Dependencies**: Existing blueprint state in Zustand
- **Acceptance**: JSON matches canvas state, edits apply, syntax highlighting renders

---

## Track 4: Mission Monitoring Enhancement (Workflow 5)

### T4.1: Mission Chip Navigation
- **Workflow Step**: 5.2
- **File**: `components/dashboard/ActiveMissionsBar.tsx` (modify)
- **Deliverable**: Clicking a mission chip navigates to `/wiring?mission={id}`:
  - Loads that mission's blueprint into CircuitBoard
  - Filters NatsWaterfall to mission's teams
- **Dependencies**: Router integration, mission data in Zustand
- **Acceptance**: Click navigates, blueprint loads, filter applies

### T4.2: Live Node Status from SSE
- **Workflow Step**: 5.2
- **File**: `store/useCortexStore.ts` (modify `dispatchSignalToNodes`)
- **Deliverable**: Enhanced SSE signal dispatch:
  - `agent.heartbeat` → update node status to `online`, reset stale timer
  - `cognitive` → set `isThinking: true`, show thought bubble
  - `tool_call` → show tool name on node briefly
  - `tool_result` → clear thinking state
  - Stale detection: if no heartbeat for 15s, mark node `offline`
- **Dependencies**: Existing SSE infrastructure
- **Acceptance**: Nodes visually respond to live signals

### T4.3: SitRep Viewer Panel
- **Workflow Step**: 5.4
- **File**: `components/dashboard/SitRepViewer.tsx`
- **Deliverable**: Panel (accessible from Mission Control sidebar or tab) showing:
  - List of SitRep cards (timestamp, team scope, summary text)
  - "Search Related" link on each card → triggers memory search
  - Populated from `GET /api/v1/memory/sitreps`
- **Dependencies**: Existing sitrep API
- **Acceptance**: Cards render, data fetched, search link works

### T4.4: Memory Search Bar
- **Workflow Step**: 5.5
- **File**: `components/dashboard/MemorySearch.tsx`
- **Deliverable**: Search bar at top of SitRep panel:
  - Text input with search icon
  - Debounced (300ms) query to `GET /api/v1/memory/search?q=...`
  - Results displayed as ranked cards with relevance score
- **Dependencies**: Existing search API, T4.3
- **Acceptance**: Search returns results, relevance shown, empty state handled

---

## Track 5: Direct Chat (Workflow 7)

### T5.1: Dedicated Chat Page
- **Workflow Step**: 7.1 (Option B)
- **Route**: `/chat`
- **Files**: `app/chat/page.tsx`, `components/chat/DirectChat.tsx`
- **Deliverable**: Full-screen chat interface:
  - Message history (scrollable, auto-scroll)
  - User messages (right-aligned, `cortex-bg`)
  - System messages (left-aligned, Bot icon, `cortex-info` accent)
  - Input bar (text + send button)
  - Typing indicator (bouncing dots)
- **Style**: Vuexy Dark, same bubble pattern as ArchitectChat but full-width
- **Dependencies**: Navigation link, Zustand direct chat slice
- **Acceptance**: Page renders, messages display, send works

### T5.2: Zustand Direct Chat Slice
- **Workflow Step**: 7.2
- **File**: `store/useCortexStore.ts` (extend)
- **Deliverable**:
  ```typescript
  directChatHistory: ChatMessage[];
  isDirectChatDrafting: boolean;
  sendDirectChat: (text: string) => Promise<void>;
  ```
  - `sendDirectChat` calls `POST /api/v1/chat` with `{ messages: [...] }`
- **Dependencies**: Existing `/api/v1/chat` endpoint
- **Acceptance**: Messages round-trip to backend, response appended

### T5.3: Markdown Message Rendering
- **Workflow Step**: 7.3
- **File**: `components/chat/MessageBubble.tsx`
- **Deliverable**: Enhanced message bubble that renders content based on type:
  - Plain text → monospace text
  - Markdown → rendered markdown (using `react-markdown` or similar)
  - Code blocks → syntax highlighted (using `highlight.js` or `prism`)
- **Dependencies**: T5.1
- **Acceptance**: Markdown renders, code blocks highlighted, links clickable

---

## Track 6: Settings Enhancements (Workflow 8)

### T6.1: Settings Tab Navigation
- **Workflow Step**: 8.x
- **File**: `app/settings/layout.tsx` (create or modify)
- **Deliverable**: Tab bar under settings header with:
  - General | Brain | Tools | Trust
  - Each tab routes to sub-page
- **Dependencies**: Existing settings pages
- **Acceptance**: Tabs render, navigation works, active state shown

### T6.2: Trust Configuration Page
- **Workflow Step**: 8.3
- **Route**: `/settings/trust`
- **Files**: `app/settings/trust/page.tsx`, `components/settings/TrustConfig.tsx`
- **Deliverable**: Full trust economy configuration:
  - Global threshold slider (same as TrustSlider but with explanation text)
  - Per-category defaults table (read-only: sensory=1.0, cognitive=0.5, etc.)
  - Recent governance log (approvals/rejections)
- **Dependencies**: Existing trust API
- **Acceptance**: Slider syncs, table displays, log populates

---

## Track 7: Navigation & Layout Updates

### T7.1: Header Navigation Update
- **Workflow Step**: Cross-cutting
- **File**: `components/layout/Header.tsx` (modify)
- **Deliverable**: Add navigation links:
  - Mission Control (/) | Wiring (/wiring) | Chat (/chat) | Catalogue (/catalogue) | Settings (/settings)
  - Active state indicator (underline or color change)
- **Dependencies**: New routes exist
- **Acceptance**: All links navigate correctly, active state works

### T7.2: Navigation Sidebar (Optional)
- **Workflow Step**: Cross-cutting
- **File**: `components/layout/Sidebar.tsx` (create)
- **Deliverable**: Collapsible icon sidebar (Zone A style) with routes:
  - Dashboard icon → /
  - Wiring icon → /wiring
  - Chat icon → /chat
  - Catalogue icon → /catalogue
  - Settings icon → /settings
  - Telemetry icon → /telemetry
- **Dependencies**: T7.1
- **Acceptance**: Sidebar renders, icons correct, collapse works

---

## Track 8: Team Actuation Viewer (Workflow 9)

### T8.1: Mission Control Teams Summary Card

- **Workflow Step**: 9.1
- **File**: `components/dashboard/TeamsSummaryCard.tsx`
- **Deliverable**: Card in Mission Control right column showing:
  - Active Teams count (teams with online agents)
  - Total Agents count (across all active missions)
  - Recent Outputs count (artifacts in last hour)
- **Style**: `cortex-surface` bg, `cortex-border` border, `rounded-xl shadow-lg`
- **Dependencies**: Zustand missions slice, SSE heartbeat data
- **Acceptance**: Card renders with live counts, updates as SSE events arrive

### T8.2: Recent Missions List

- **Workflow Step**: 9.1
- **File**: `components/dashboard/RecentMissionsList.tsx`
- **Deliverable**: Vertical list below TeamsSummaryCard showing active missions:
  - Mission name, status badge (active/completed/failed), team count, last activity timestamp
  - "View" arrow icon → navigates to `/missions/{id}/teams`
  - Empty state: "No active missions. Commit a blueprint from the Wiring page to get started."
- **Dependencies**: `GET /api/v1/missions`, SSE heartbeat for last activity
- **Acceptance**: List populates, navigation works, empty state renders

### T8.3: Team Actuation Page Shell

- **Workflow Step**: 9.2
- **Route**: `/missions/[id]/teams`
- **Files**: `app/missions/[id]/teams/page.tsx`, `components/missions/TeamActuationView.tsx`
- **Deliverable**: Two-panel layout:
  - Left (40%): Team Roster panel
  - Right (60%): Agent Activity Feed panel
  - Breadcrumb: Mission Control > Mission Name > Teams
- **Style**: Vuexy Dark, `next/dynamic({ ssr: false })` for SSE-dependent components
- **Dependencies**: T8.2 navigation link, mission ID from URL params
- **Acceptance**: Page renders at dynamic route, panels visible, breadcrumb works

### T8.4: Team Roster Panel

- **Workflow Step**: 9.2
- **File**: `components/missions/TeamRoster.tsx`
- **Deliverable**: Vertical card list for each team in the mission:
  - Team name, status dot (green/yellow/red based on heartbeat age), agent count
  - Role icons per agent (cognitive=purple, sensory=cyan, actuation=green, ledger=muted)
  - Output count badge (from artifacts query)
  - Click to select → updates right panel
- **Dependencies**: T8.3, SSE heartbeat stream, `GET /api/v1/artifacts?team_id={id}`
- **Acceptance**: Cards render, selection highlights, status dots update in real-time

### T8.5: Agent Activity Feed

- **Workflow Step**: 9.2
- **File**: `components/missions/AgentActivityFeed.tsx`
- **Deliverable**: When a team is selected:
  - **Agent Row**: Horizontal cards per agent (ID, role icon, status, current action, trust score)
  - **Activity Timeline**: Reverse-chronological feed with color-coded left borders:
    - Cognitive (purple), Tool Call (blue), Tool Result (green/red), Output (gold), Governance Halt (red)
  - Each entry: timestamp, agent source, message preview (200 char truncate), expand on click
- **Dependencies**: T8.4 (team selection), SSE stream filtered by team agent IDs
- **Acceptance**: Agent rows update live, timeline populates from SSE, entries expandable

### T8.6: Zustand Artifacts Slice

- **Workflow Step**: 9.3
- **File**: `store/useCortexStore.ts` (extend)
- **Deliverable**: New state slice:
  ```typescript
  artifacts: Artifact[];
  isFetchingArtifacts: boolean;
  selectedArtifact: Artifact | null;
  fetchArtifacts: (filters?: ArtifactFilters) => Promise<void>;
  getArtifact: (id: string) => Promise<void>;
  updateArtifactStatus: (id: string, status: string) => Promise<void>;
  ```
- **Type**: `Artifact` mirrors backend `artifacts` row, `ArtifactFilters` has optional mission_id, team_id, agent_id, limit
- **Dependencies**: B6 (artifacts API endpoints)
- **Acceptance**: Store compiles, fetch populates from backend, filters work

### T8.7: Artifact Viewer

- **Workflow Step**: 9.3
- **File**: `components/missions/ArtifactViewer.tsx`
- **Deliverable**: Filterable grid of artifact cards + detail slide-over:
  - **Grid**: title, type badge, agent source, trust score, status badge, created timestamp
  - **Filters**: type chips (All/Code/Document/Image/Data/File), status chips, agent dropdown
  - **Detail view** (slide-over on click):
    - Code: syntax-highlighted block with copy button
    - Document: rendered markdown
    - Image: preview with metadata (dimensions, file size)
    - Data (JSON): collapsible tree viewer
    - File: icon + name + download link
  - **Governance actions**: Approve/Reject buttons (only when status=pending)
- **Dependencies**: T8.6, existing syntax highlighting library
- **Acceptance**: Grid renders, filters work, detail views render per type, governance actions call API

### T8.8: Mission Summary Tab

- **Workflow Step**: 9.4
- **File**: `components/missions/MissionSummaryTab.tsx`
- **Deliverable**: Aggregate metrics tab at top of mission page:
  - Total Artifacts, Pending Review (amber badge), Cognitive Cycles, Tool Invocations
  - Governance Halts (red badge), Mission Duration (HH:MM:SS from activated_at)
  - Latest SitRep (expandable card from `GET /api/v1/memory/sitreps`)
- **Dependencies**: T8.3, T8.6, existing sitreps API
- **Acceptance**: Metrics compute correctly, duration ticks live, SitRep expands

---

## Track 9: Memory Explorer (Workflow 10)

### T9.1: Memory Explorer Page Shell

- **Workflow Step**: 10.1
- **Route**: `/memory`
- **Files**: `app/memory/page.tsx`, `components/memory/MemoryExplorer.tsx`
- **Deliverable**: Three-column responsive layout:
  - Left (33%): Hot Memory (Live Feed)
  - Center (33%): Warm Memory (Structured)
  - Right (33%): Cold Memory (Semantic)
  - Header: "THREE-TIER MEMORY" with tier labels
- **Style**: Vuexy Dark, `next/dynamic({ ssr: false })`, responsive grid (minmax like MissionControl)
- **Dependencies**: Navigation link in header (T7.1)
- **Acceptance**: Page renders at `/memory`, three columns visible, responsive

### T9.2: Hot Memory Panel

- **Workflow Step**: 10.1 (Column 1)
- **File**: `components/memory/HotMemoryPanel.tsx`
- **Deliverable**: Real-time event stream (compact NatsWaterfall variant):
  - Event rows: timestamp (HH:MM:SS.ms), source agent, type badge, preview (80 chars)
  - Auto-scroll (pause on hover)
  - Event count badge: "247 events buffered"
  - Flush indicator: visual pulse when 20-event threshold triggers warm flush
  - Color coding: input=cyan, output=green, internal=muted (matches NatsWaterfall)
- **Dependencies**: SSE stream from Zustand `streamLogs[]`
- **Acceptance**: Events stream in real-time, auto-scroll works, pause on hover, colors correct

### T9.3: Warm Memory Panel — SitReps Tab

- **Workflow Step**: 10.1 (Column 2, Tab 1)
- **File**: `components/memory/WarmMemoryPanel.tsx`
- **Deliverable**: Tabbed container (SitReps | Artifacts | Agent State) with SitReps tab:
  - Reverse-chronological list of SitRep cards
  - Each card: timestamp, team scope, summary text (200 char truncate), expand to full
  - "Find Related" button on each card → triggers cold tier search with SitRep text
  - Source: `GET /api/v1/memory/sitreps`
- **Dependencies**: Existing sitreps API
- **Acceptance**: Tab renders, cards populate, expand works, "Find Related" triggers search

### T9.4: Warm Memory Panel — Artifacts Tab

- **Workflow Step**: 10.1 (Column 2, Tab 2)
- **File**: `components/memory/WarmMemoryPanel.tsx` (add tab)
- **Deliverable**: Artifacts tab showing recent outputs across all missions:
  - Card per artifact: title, type badge, agent source, trust score, status
  - Filterable by type chips (code/document/image/data)
  - Source: `GET /api/v1/artifacts?limit=50`
- **Dependencies**: T8.6 (artifacts Zustand slice)
- **Acceptance**: Tab renders, artifacts populate, type filter works

### T9.5: Warm Memory Panel — Agent State Tab

- **Workflow Step**: 10.1 (Column 2, Tab 3)
- **File**: `components/memory/WarmMemoryPanel.tsx` (add tab)
- **Deliverable**: Agent state key-value display:
  - Grouped by agent_id (collapsible groups)
  - Each row: key, value (truncated), mission scope, last updated, TTL remaining
  - Placeholder state: "Agent state API coming soon" (future endpoint)
- **Dependencies**: None (placeholder for future API)
- **Acceptance**: Tab renders, placeholder displays cleanly

### T9.6: Cold Memory Panel

- **Workflow Step**: 10.1 (Column 3)
- **File**: `components/memory/ColdMemoryPanel.tsx`
- **Deliverable**: Semantic search + results panel:
  - **Search bar**: text input with brain icon, "Search long-term memory..." placeholder
  - Debounced 500ms → `GET /api/v1/memory/search?q={query}`
  - **Results**: ranked cards with relevance percentage bar, content preview, source team/timestamp
  - **Memory Insights**: static info cards (total vectors, embedding model, dimension, last embedding)
  - **Info panel** (collapsible): "How Memory Drives Action" explanation text
- **Dependencies**: Existing `/api/v1/memory/search` API
- **Acceptance**: Search returns results, relevance bars render, insights display

### T9.7: Cross-Tier Navigation Links

- **Workflow Step**: 10.3
- **File**: All memory panel components (modify)
- **Deliverable**: Inter-tier navigation:
  - Hot event → "View SitRep" link (scrolls to warm SitReps tab)
  - Warm SitRep → "Find Related" (populates cold search)
  - Warm Artifact → "View Timeline" (filters hot feed to artifact's agent)
  - Cold result → "View Source" (navigates to warm SitRep)
  - Cold result → "View Mission" (navigates to `/missions/{id}/teams`)
- **Dependencies**: T9.2, T9.3, T9.6
- **Acceptance**: All cross-links navigate correctly, state passes between panels

---

## Track 10: Mission Control Enhancement (Workflow 9)

### T10.1: Mission Control Layout Update

- **Workflow Step**: 9.1
- **File**: `components/dashboard/MissionControl.tsx` (modify)
- **Deliverable**: Update right column to include TeamsSummaryCard and RecentMissionsList:
  - Existing: TelemetryRow, PriorityStream
  - Add: TeamsSummaryCard (below TelemetryRow), RecentMissionsList (below summary)
  - Responsive grid maintains 3-column layout with minmax
- **Dependencies**: T8.1, T8.2
- **Acceptance**: New cards appear in right column, layout doesn't collapse

### T10.2: Navigation Update for New Routes

- **Workflow Step**: Cross-cutting
- **File**: `components/layout/Header.tsx` (modify)
- **Deliverable**: Add navigation links for new routes:
  - Existing: Mission Control (/) | Wiring (/wiring)
  - Add: Memory (/memory) — brain icon
  - Mission detail pages linked from Mission Control (no nav entry needed)
- **Dependencies**: T9.1 (memory page exists)
- **Acceptance**: Navigation link works, active state correct

---

## Implementation Order

**Phase 1 — Foundation** (Backend + Core UI Shell)
1. B1 → B2 → B3 → B4 (backend catalogue API)
2. T7.1 (navigation update)
3. T1.1 → T1.2 (catalogue page shell + store)

**Phase 2 — Agent Catalogue** (CRUD operations)
4. T1.3 → T1.4 (cards + grid with filters)
5. T1.5 (editor drawer)
6. T1.6 (delete)
7. T1.7 (preview/test)

**Phase 3 — MCP Tool Registry**
8. T2.1 → T2.2 (page + store)
9. T2.3 (server cards)
10. T2.4 (install modal)
11. T2.5 (tool test panel)

**Phase 4 — Team Builder Enhancement**
12. T3.1 → T3.2 (catalogue sidebar + drag-drop)
13. T3.3 (agent node panel)
14. T3.4 (team group creation)
15. T3.5 (blueprint preview)

**Phase 5 — Direct Chat**
16. T5.1 → T5.2 (chat page + store)
17. T5.3 (markdown rendering)

**Phase 6 — Mission Monitoring**
18. T4.1 (mission navigation)
19. T4.2 (live node status)
20. T4.3 → T4.4 (sitrep viewer + search)

**Phase 7 — Settings & Polish**
21. T6.1 (settings tabs)
22. T6.2 (trust config page)
23. T2.1 integration into settings tabs
24. T7.2 (optional sidebar)

**Phase 8 — Team Actuation Viewer** (Workflow 9)
25. B5 → B6 → B7 (backend artifacts + K8s volume)
26. T8.1 → T8.2 (Mission Control summary + missions list)
27. T10.1 (Mission Control layout update)
28. T8.3 (team actuation page shell)
29. T8.4 → T8.5 (team roster + agent activity feed)
30. T8.6 → T8.7 (artifacts Zustand + artifact viewer)
31. T8.8 (mission summary tab)

**Phase 9 — Memory Explorer** (Workflow 10)
32. T9.1 (memory explorer page shell)
33. T9.2 (hot memory panel)
34. T9.3 → T9.4 → T9.5 (warm memory tabs)
35. T9.6 (cold memory panel)
36. T9.7 (cross-tier navigation links)
37. T10.2 (navigation update for new routes)

---

## Verification Checklist

After each track, verify:
- [ ] `inv interface.build` passes (no type errors)
- [ ] `inv interface.check` passes (pages load without errors)
- [ ] All new components use Vuexy Dark tokens (`cortex-*`)
- [ ] No `bg-white`, `zinc-*`, or `slate-*` classes in new components
- [ ] SSR-unsafe components use `next/dynamic({ ssr: false })`
- [ ] All API calls handle error states (loading, error, empty)
- [ ] Backend: `inv core.build` passes after API changes
- [ ] Backend: `inv core.test` passes
- [ ] Dynamic routes use `next/dynamic({ ssr: false })` where SSE/ReactFlow is used
- [ ] Cross-tier links in Memory Explorer navigate correctly
- [ ] Artifact viewer renders all content types (code, document, image, data, file)

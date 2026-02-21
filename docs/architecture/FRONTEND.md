# Mycelis Cortex — Frontend Specification

> **Load this doc when:** Working on Next.js code, React components, Zustand store, or visual design.
>
> **Related:** [Overview](OVERVIEW.md) | [Backend](BACKEND.md) | [Operations](OPERATIONS.md)

---

## I. Technology Stack

| Component | Version | Notes |
|-----------|---------|-------|
| Next.js | 16.1.6 | Turbopack, App Router |
| React | 19.2.3 | `"use client"` required for hooks/state |
| Tailwind CSS | v4 | `@import "tailwindcss"`, `@theme` directive in globals.css |
| ReactFlow | 11.11.4 | Package is `reactflow`, NOT `@xyflow/react` |
| Zustand | 5.0.11 | Single atomic store: `useCortexStore` |
| Observable Plot | 0.6.17 | Charting (bar, line, area, dot, waffle, tree) |
| D3 | 7.9.0 | Data visualization primitives |
| Leaflet / react-leaflet | 1.9.4 / 5.0.0 | Geo mapping |
| @tanstack/react-table | 8.21.3 | Data tables |
| @ai-sdk/react / ai | 3.0.70 / 6.0.68 | Vercel AI SDK |
| axios | 1.13.4 | HTTP client |
| lucide-react | 0.563.0 | Icon library |
| xterm | 5.3.0 | Terminal emulator |
| TypeScript | 5 | |
| Vitest | 4.0.18 | Unit tests |
| Playwright | 1.58.2 | E2E tests |

---

## II. The 4-Zone Shell

**File:** `components/shell/ShellLayout.tsx`

```
┌──────┬─────────────────────────────────────────┐
│      │                                         │
│ Zone │  Zone B (Workspace / Page Content)       │
│  A   │  ┌──────────┬──────────────────────┐    │
│      │  │ Chat     │ Canvas / Content     │    │
│ Rail │  │ (360px)  │ (flex-1)             │    │
│      │  │          │                      │    │
│      │  ├──────────┴──────────────────────┤    │
│      │  │ Zone C: Spectrum (collapsible)  │    │
│      │  │ NatsWaterfall telemetry feed    │    │
│      │  └─────────────────────────────────┘    │
└──────┴─────────────────────────────────────────┘
                 Zone D (Governance Overlay, z-50)
```

### Zone Components

| Zone | Component | Purpose |
|------|-----------|---------|
| **A** | `ZoneA_Rail.tsx` | Vertical navigation sidebar, route links, status indicators |
| **B** | `ZoneB_Workspace.tsx` | Main canvas: `grid-cols-[360px_1fr]`, `gap-6 p-6`, `rounded-xl shadow-lg` cards |
| **C** | `ZoneC_Stream.tsx` / `NatsWaterfall` | Bottom panel: scrolling SSE telemetry, collapsible |
| **D** | `ZoneD_Decision.tsx` + `GovernanceModal.tsx` | z-50 overlay for governance approvals, trust alerts |

### Layout Implementation

```tsx
<div className="flex h-screen w-screen overflow-hidden bg-cortex-bg">
  <ZoneA />
  <div className="flex-1 flex flex-col bg-cortex-bg">
    <ZoneB>{children}</ZoneB>
  </div>
  <ZoneD />
</div>
```

### Layout Standards

| Property | Value |
|----------|-------|
| Card wrappers | `rounded-xl shadow-lg` |
| Grid spacing | `gap-6 p-6` |
| Workspace grid | `grid-cols-[360px_1fr]` |
| ReactFlow background | `#09090b` (cortex-bg) |
| Z-index governance | z-50 |

---

## III. Route Map (15 Pages)

### Route Groups

Routes use Next.js App Router layout groups:
- `(marketing)` — Public pages
- `(app)` — Authenticated pages (wrapped in ShellLayout)

### All Routes

| Route | Page File | Component | Purpose |
|-------|-----------|-----------|---------|
| `/` | `(marketing)/page.tsx` | Landing | Product overview, links to `/dashboard` |
| `/dashboard` | `(app)/dashboard/page.tsx` | MissionControl | Council Chat (member selector) + OperationsBoard + Telemetry + Sensors + CognitiveStatus |
| `/wiring` | `(app)/wiring/page.tsx` | Workspace | ArchitectChat + CircuitBoard (ReactFlow) + ToolsPalette + NatsWaterfall |
| `/architect` | `(app)/architect/page.tsx` | Redirect | Redirects to `/wiring` |
| `/teams` | `(app)/teams/page.tsx` | TeamsPage | Standing + mission teams, agent roster, delivery targets |
| `/catalogue` | `(app)/catalogue/page.tsx` | CataloguePage | Agent blueprint CRUD |
| `/memory` | `(app)/memory/page.tsx` | MemoryExplorer | Hot/Warm/Cold three-tier browser |
| `/approvals` | `(app)/approvals/page.tsx` | DecisionCard | Governance queue, policy config, proposals (3 tabs) |
| `/missions/[id]/teams` | `(app)/missions/[id]/teams/page.tsx` | TeamActuationView | Live team drill-down for active missions |
| `/settings` | `(app)/settings/page.tsx` | Settings | 4 tabs: Profile, Teams, Cognitive Matrix, MCP Tools |
| `/settings/brain` | `(app)/settings/brain/page.tsx` | CognitiveMatrix | Provider routing + ProviderConfigModal |
| `/settings/tools` | `(app)/settings/tools/page.tsx` | MCPToolRegistry | MCP management + curated library |
| `/matrix` | `(app)/matrix/page.tsx` | MatrixGrid | Cognitive Matrix data table |
| `/marketplace` | `(app)/marketplace/page.tsx` | Skills Market | Connector registry |
| `/telemetry` | `(app)/telemetry/page.tsx` | System Status | Infrastructure monitoring |

---

## IV. Component Inventory (93 files, 20 folders)

### Shell (6 files)
| Component | Purpose |
|-----------|---------|
| `ShellLayout.tsx` | Root layout: 4-zone structure |
| `ZoneA_Rail.tsx` | Navigation sidebar |
| `ZoneB_Workspace.tsx` | Main canvas area |
| `ZoneC_Stream.tsx` | Bottom telemetry panel |
| `ZoneD_Decision.tsx` | Governance overlay container |
| `GovernanceModal.tsx` | Approval/trust modal |

### Workspace (9 files)
| Component | Purpose |
|-----------|---------|
| `Workspace.tsx` | Workspace page container |
| `ArchitectChat.tsx` | Intent input + chat with Meta-Architect |
| `CircuitBoard.tsx` | ReactFlow DAG visualizer (wiring view) |
| `SquadRoom.tsx` | Fractal drill-down into team internals |
| `BlueprintDrawer.tsx` | Blueprint library browser (right drawer) |
| `DeliverablesTray.tsx` | Pending artifacts for human review |
| `ToolsPalette.tsx` | Internal + MCP tool browser (collapsible drawer) |
| `TrustSlider.tsx` | Trust threshold control (0.0–1.0) |
| `CognitiveStatusPanel.tsx` | LLM provider health display |

### Dashboard (15 files)
| Component | Purpose |
|-----------|---------|
| `MissionControl.tsx` | Dashboard layout (3-column responsive grid) |
| `MissionControlChat.tsx` | Council chat with member selector dropdown |
| `OperationsBoard.tsx` | Priority alerts, standing workloads, missions |
| `TelemetryRow.tsx` | Runtime metrics row (goroutines, memory, token rate) |
| `SensorLibrary.tsx` | Grouped sensor subscriptions (library pattern) |
| `SensoryPeriphery.tsx` | Sensor data display |
| `SignalContext.tsx` | Signal context enrichment |
| `ManifestationPanel.tsx` | Team manifestation proposals |
| `ActivityStream.tsx` | Live activity feed |
| `ActivityFeed.tsx` | Activity feed variant |
| `AgentPanel.tsx` | Agent status panel |
| `CommandDeck.tsx` | Command interface |
| `TeamList.tsx` | Team listing |
| `TeamRoster.tsx` | Team member roster |
| `MetricCard.tsx` | Single metric display card |

### Wiring (4 files)
| Component | Purpose |
|-----------|---------|
| `AgentNode.tsx` | ReactFlow custom node — category iconography, status glow, pencil hover |
| `CircuitBoard.tsx` | ReactFlow canvas variant for wiring page |
| `DataWire.tsx` | Custom edge component for data flow visualization |
| `WiringAgentEditor.tsx` | Right-side drawer for editing agent manifest (role, prompt, model, tools, I/O) |

### Teams (3 files)
| Component | Purpose |
|-----------|---------|
| `TeamsPage.tsx` | Team management page layout |
| `TeamCard.tsx` | Individual team card |
| `TeamDetailDrawer.tsx` | Team detail side drawer |

### Charts (3 files)
| Component | Purpose |
|-----------|---------|
| `ChartRenderer.tsx` | Observable Plot (bar, line, area, dot, waffle, tree) — renders from `MycelisChartSpec` |
| `MapRenderer.tsx` | Leaflet geo maps |
| `DataTable.tsx` | @tanstack/react-table data grid |

### Missions (5 files)
| Component | Purpose |
|-----------|---------|
| `MissionSummaryTab.tsx` | Mission overview |
| `AgentActivityFeed.tsx` | Per-agent activity stream |
| `ArtifactViewer.tsx` | Inline artifact rendering (text, code, charts) |
| `TeamActuationView.tsx` | Live team drill-down |
| `TeamRoster.tsx` | Mission team member list |

### Memory (4 files)
| Component | Purpose |
|-----------|---------|
| `MemoryExplorer.tsx` | Three-tier memory browser container |
| `HotMemoryPanel.tsx` | Active/recent memories |
| `WarmMemoryPanel.tsx` | SitReps and compressed context |
| `ColdMemoryPanel.tsx` | Archived/cold storage |

### Stream (2 files)
| Component | Purpose |
|-----------|---------|
| `NatsWaterfall.tsx` | Scrolling SSE telemetry feed with signal classification (input/output/internal) and filter tabs |
| `SignalDetailDrawer.tsx` | Signal detail side drawer |

### Settings (5 files)
| Component | Purpose |
|-----------|---------|
| `MCPToolRegistry.tsx` | MCP server management page |
| `MCPLibraryBrowser.tsx` | Curated MCP library browser |
| `MCPServerCard.tsx` | Individual server card |
| `MCPInstallModal.tsx` | Server installation dialog |
| `ModelHealth.tsx` | LLM provider health display |

### Catalogue (3 files)
| Component | Purpose |
|-----------|---------|
| `CataloguePage.tsx` | Agent template CRUD page |
| `AgentCard.tsx` | Individual agent template card |
| `AgentEditorDrawer.tsx` | Agent template editor drawer |

### Shared (5 files)
| Component | Purpose |
|-----------|---------|
| `UniversalRenderer.tsx` | Dynamic content renderer |
| `ArtifactCard.tsx` | Artifact display card |
| `ApprovalCard.tsx` | Approval request card |
| `ThoughtCard.tsx` | Agent thought display |
| `MetricPill.tsx` | Compact metric badge |

### Other Components

| Folder | Components |
|--------|------------|
| `command/` | CommandBar, SessionFeed, StatusBar |
| `matrix/` | MatrixGrid, columns, data-table |
| `approvals/` | DecisionCard |
| `genesis/` | GenesisTerminal |
| `hud/` | LogStream |
| `layout/` | Header |
| `operator/` | Console |
| `registry/` | NetworkMap |
| `swarm/` | TeamBuilder |
| `ui/` | table (shadcn-style) |
| Root | ApprovalDeck, LogStream, SystemStatus |

---

## V. Zustand Store (`useCortexStore`)

**File:** `store/useCortexStore.ts`

**Rule:** Zero `useState` for application state. All shared state in this single atomic store.

### State Fields (~60 fields)

#### Core Blueprint & Graph
```typescript
chatHistory: ChatMessage[]           // Chat with user + architect
nodes: Node[]                        // ReactFlow nodes
edges: Edge[]                        // ReactFlow edges
isDrafting: boolean                  // Negotiation in progress
isCommitting: boolean                // Mission instantiation in progress
error: string | null                 // Last operation error
blueprint: MissionBlueprint | null   // Current blueprint
missionStatus: MissionStatus         // 'idle' | 'draft' | 'active'
activeMissionId: string | null       // Current active mission
```

#### SSE Stream
```typescript
streamLogs: StreamSignal[]           // SSE signals (capped at 100)
isStreamConnected: boolean           // EventSource status
```

#### Navigation & UI
```typescript
activeSquadRoomId: string | null     // Fractal drill-down target
isBlueprintDrawerOpen: boolean
isToolsPaletteOpen: boolean
```

#### Governance & Deliverables
```typescript
pendingArtifacts: CTSEnvelope[]      // Pending human-review artifacts
selectedArtifact: CTSEnvelope | null
```

#### Trust Economy
```typescript
trustThreshold: number               // 0.0–1.0, default 0.7
isSyncingThreshold: boolean
```

#### Blueprint Library
```typescript
savedBlueprints: MissionBlueprint[]
```

#### Sensory Periphery
```typescript
sensorFeeds: SensorNode[]
isFetchingSensors: boolean
subscribedSensorGroups: string[]
```

#### Team Proposals
```typescript
teamProposals: TeamProposal[]
isFetchingProposals: boolean
```

#### Agent Catalogue
```typescript
catalogueAgents: CatalogueAgent[]
isFetchingCatalogue: boolean
selectedCatalogueAgent: CatalogueAgent | null
```

#### Artifacts
```typescript
artifacts: Artifact[]
isFetchingArtifacts: boolean
selectedArtifactDetail: Artifact | null
```

#### MCP Servers
```typescript
mcpServers: MCPServerWithTools[]
isFetchingMCPServers: boolean
mcpTools: MCPTool[]
mcpLibrary: MCPLibraryCategory[]
isFetchingMCPLibrary: boolean
```

#### Mission Control Chat
```typescript
missionChat: ChatMessage[]           // Council chat history
isMissionChatting: boolean
missionChatError: string | null
councilTarget: string                // Active member ID (default 'admin')
councilMembers: CouncilMember[]
```

#### Broadcast
```typescript
isBroadcasting: boolean
lastBroadcastResult: { teams_hit: number } | null
```

#### Team Explorer
```typescript
teamRoster: TeamDetail[]
isFetchingTeamRoster: boolean
```

#### Governance
```typescript
policyConfig: PolicyConfig | null
pendingApprovals: PendingApproval[]
isFetchingPolicy: boolean
isFetchingApprovals: boolean
cognitiveStatus: CognitiveStatus | null
```

#### Team Management
```typescript
teamsDetail: TeamDetailEntry[]
isFetchingTeamsDetail: boolean
selectedTeamId: string | null
isTeamDrawerOpen: boolean
teamsFilter: TeamsFilter             // 'all' | 'standing' | 'mission'
```

#### Wiring Edit/Delete
```typescript
selectedAgentNodeId: string | null
isAgentEditorOpen: boolean
```

#### Signal Detail
```typescript
selectedSignalDetail: SignalDetail | null
```

#### ReactFlow Handlers
```typescript
onNodesChange: OnNodesChange
onEdgesChange: OnEdgesChange
```

### Actions (~40 methods)

#### Intent & Blueprint
| Action | API | Purpose |
|--------|-----|---------|
| `submitIntent(text)` | POST /intent/negotiate | Generate blueprint |
| `instantiateMission()` | POST /intent/commit | Persist + activate |

#### SSE Stream
| Action | API | Purpose |
|--------|-----|---------|
| `initializeStream()` | EventSource /stream | Connect SSE |
| `disconnectStream()` | — | Close EventSource |

#### Navigation
| Action | API | Purpose |
|--------|-----|---------|
| `enterSquadRoom(teamId)` | — | Fractal drill-down |
| `exitSquadRoom()` | — | Return to full DAG |
| `toggleBlueprintDrawer()` | — | Toggle drawer |
| `toggleToolsPalette()` | — | Toggle palette |

#### Deliverables
| Action | API | Purpose |
|--------|-----|---------|
| `selectArtifact(artifact)` | — | Select for detail |
| `approveArtifact(id)` | — | Remove from pending |
| `rejectArtifact(id, reason)` | — | Remove from pending |
| `selectSignalDetail(detail)` | — | Signal detail view |

#### Trust Economy
| Action | API | Purpose |
|--------|-----|---------|
| `setTrustThreshold(value)` | PUT /trust/threshold | Sync locally + backend |
| `fetchTrustThreshold()` | GET /trust/threshold | Load threshold |

#### Blueprint Library
| Action | API | Purpose |
|--------|-----|---------|
| `saveBlueprint(bp)` | — | Save to library |
| `loadBlueprint(bp)` | — | Load to canvas |

#### Sensory
| Action | API | Purpose |
|--------|-----|---------|
| `fetchSensors()` | GET /sensors | Load sensor feeds |
| `toggleSensorGroup(group)` | — | Toggle subscription |

#### Proposals
| Action | API | Purpose |
|--------|-----|---------|
| `fetchProposals()` | GET /proposals | Load proposals |
| `approveProposal(id)` | POST /proposals/{id}/approve | Approve |
| `rejectProposal(id)` | POST /proposals/{id}/reject | Reject |

#### Missions
| Action | API | Purpose |
|--------|-----|---------|
| `fetchMissions()` | GET /missions | Load mission list |

#### Agent Catalogue
| Action | API | Purpose |
|--------|-----|---------|
| `fetchCatalogue()` | GET /catalogue/agents | Load templates |
| `createCatalogueAgent(agent)` | POST /catalogue/agents | Create |
| `updateCatalogueAgent(id, agent)` | PUT /catalogue/agents/{id} | Update |
| `deleteCatalogueAgent(id)` | DELETE /catalogue/agents/{id} | Delete |
| `selectCatalogueAgent(agent)` | — | Select for edit |

#### Artifacts
| Action | API | Purpose |
|--------|-----|---------|
| `fetchArtifacts(filters)` | GET /artifacts | Load (filterable) |
| `getArtifactDetail(id)` | GET /artifacts/{id} | Detail |
| `updateArtifactStatus(id, status)` | PUT /artifacts/{id}/status | Status change |

#### MCP
| Action | API | Purpose |
|--------|-----|---------|
| `fetchMCPServers()` | GET /mcp/servers | Load servers |
| `installMCPServer(config)` | POST /mcp/install | Install |
| `deleteMCPServer(id)` | DELETE /mcp/servers/{id} | Uninstall |
| `fetchMCPTools()` | GET /mcp/tools | Load tools |
| `fetchMCPLibrary()` | GET /mcp/library | Load library |
| `installFromLibrary(name, env)` | POST /mcp/library/install | Install from library |

#### Council Chat
| Action | API | Purpose |
|--------|-----|---------|
| `setCouncilTarget(id)` | — | Switch active member |
| `fetchCouncilMembers()` | GET /council/members | Load roster |
| `sendMissionChat(message)` | POST /council/{target}/chat | Query council |
| `clearMissionChat()` | — | Clear chat |

#### Broadcast
| Action | API | Purpose |
|--------|-----|---------|
| `broadcastToSwarm(message)` | POST /swarm/broadcast | Send to all teams |

#### Team Management
| Action | API | Purpose |
|--------|-----|---------|
| `fetchTeamsDetail()` | GET /teams/detail | Load roster |
| `selectTeam(teamId)` | — | Select in drawer |
| `setTeamsFilter(filter)` | — | Filter teams |

#### Wiring Edit/Delete
| Action | API | Purpose |
|--------|-----|---------|
| `selectAgentNode(nodeId)` | — | Select for edit |
| `updateAgentInDraft(teamIdx, agentIdx, updates)` | — (local) | Modify blueprint, regenerate graph |
| `deleteAgentFromDraft(teamIdx, agentIdx)` | — (local) | Remove agent |
| `discardDraft()` | — (local) | Clear canvas |
| `updateAgentInMission(name, manifest)` | PUT /missions/{id}/agents/{name} | Update live agent |
| `deleteAgentFromMission(name)` | DELETE /missions/{id}/agents/{name} | Remove live agent |
| `deleteMission(missionId)` | DELETE /missions/{id} | Deactivate + delete |

#### Governance
| Action | API | Purpose |
|--------|-----|---------|
| `fetchPolicy()` | GET /governance/policy | Load rules |
| `updatePolicy(config)` | PUT /governance/policy | Update rules |
| `fetchPendingApprovals()` | GET /governance/pending | Load queue |
| `resolveApproval(id, approved)` | POST /governance/resolve/{id} | Approve/deny |
| `fetchCognitiveStatus()` | GET /cognitive/status | Provider health |

---

## VI. Key TypeScript Types

### ChatMessage
```typescript
interface ChatMessage {
  role: 'user' | 'architect' | 'admin' | 'council'
  content: string
  consultations?: ChatConsultation[]
  tools_used?: string[]
  source_node?: string
  trust_score?: number
  timestamp?: string
}
```

### StreamSignal
```typescript
interface StreamSignal {
  type: string
  source?: string
  level?: string
  message?: string
  timestamp?: string
  payload?: any
  topic?: string
}
```

### CTSEnvelope (Frontend)
```typescript
interface CTSEnvelope {
  id: string
  source: string
  signal: 'artifact' | 'governance_halt'
  timestamp: string
  trust_score?: number
  payload: {
    content: string
    content_type: 'markdown' | 'json' | 'text' | 'image'
    title?: string
  }
  proof?: {
    method: 'semantic' | 'empirical'
    logs: string
    rubric_score: string
    pass: boolean
  }
}
```

### MissionBlueprint
```typescript
interface MissionBlueprint {
  mission_id: string
  intent: string
  teams: BlueprintTeam[]
  constraints?: Constraint[]
}
```

### AgentManifest
```typescript
interface AgentManifest {
  id: string
  role: string
  system_prompt?: string
  model?: string
  inputs?: string[]
  outputs?: string[]
  tools?: string[]
}
```

### Artifact
```typescript
interface Artifact {
  id: string
  mission_id?: string
  team_id?: string
  agent_id: string
  trace_id?: string
  artifact_type: 'code' | 'document' | 'image' | 'audio' | 'data' | 'file' | 'chart'
  title: string
  content_type: string
  content?: string
  file_path?: string
  file_size_bytes?: number
  metadata: Record<string, any>
  trust_score?: number
  status: 'pending' | 'approved' | 'rejected' | 'archived'
  created_at: string
}
```

### MycelisChartSpec
```typescript
interface MycelisChartSpec {
  chart_type: 'bar' | 'line' | 'area' | 'dot' | 'waffle' | 'tree' | 'geo' | 'table'
  data: any[]
  x?: string
  y?: string
  color?: string
  size?: string
  sort?: { field: string; order: 'ascending' | 'descending' }
  geo?: { lat: string; lng: string; label?: string }
  width?: number
  height?: number
}
```

### Protocol Envelope Types (`lib/types/protocol.ts`)
```typescript
type EnvelopeType = 'thought' | 'metric' | 'artifact' | 'governance'

interface Envelope<T> { type: EnvelopeType; payload: T; timestamp: string }
interface ThoughtContent { summary: string; detail: string; model: string }
interface MetricContent { label: string; value: number; unit: string; status: string }
interface ArtifactContent { id: string; mime: string; title: string; preview: string; uri: string }
interface GovernanceContent { request: string; agent: string; description: string; action: string; status: string }
```

---

## VII. Visual Design System — Midnight Cortex

### Cortex Token Palette (Active Theme)

All new components use `cortex-*` tokens. **Never** use `bg-white`, `bg-zinc-*`, or `bg-slate-*`.

| Token | Color | Hex | Usage |
|-------|-------|-----|-------|
| `cortex-bg` | Zinc-950 | `#09090b` | Canvas background, ReactFlow bg |
| `cortex-surface` | Zinc-900 | `#18181b` | Card/panel backgrounds |
| `cortex-primary` | Cyan-500 | `#06b6d4` | Primary actions, cognitive nodes, active states |
| `cortex-primary-glow` | Cyan (alpha) | `rgba(6,182,212,0.39)` | Glow effects |
| `cortex-success` | Emerald-500 | `#10b981` | Success states, actuation nodes |
| `cortex-warning` | Amber-500 | `#f59e0b` | Warning states |
| `cortex-danger` | Red-500 | `#ef4444` | Error, critical alerts |
| `cortex-info` | Blue-500 | `#3b82f6` | Info indicators, sensory nodes |
| `cortex-text-main` | Zinc-300 | `#d4d4d8` | Primary text |
| `cortex-text-muted` | Zinc-500 | `#71717a` | Secondary text, labels |
| `cortex-border` | Zinc-800 | `#27272a` | Borders, dividers |

### CSS Utility Classes

| Class | Effect |
|-------|--------|
| `.cortex-card` | Border, shadow, hover transition |
| `.cortex-panel` | Sidebar/header panel |
| `.glass` | Glass morphism (backdrop blur + border) |
| `.glow-stable` | Green glow (rgb(40,199,111)) — online agents |
| `.glow-active` | Primary glow — active states |
| `.glow-critical` | Red glow (rgb(234,84,85)) — errors |
| `.glow-primary` | Primary glow variant |
| `.nav-pill-active` | Active nav pill (cyan bg + shadow) |
| `.ghost-draft` | Pulsing dashed border — draft nodes |
| `.activity-ring` | Orbiting glow — thinking agents |
| `.animate-pulse-subtle` | Deliverables tray pulse |

### Animations

| Name | Duration | Purpose |
|------|----------|---------|
| `pulse-border` | 2s | Border color pulse for draft nodes |
| `ring-pulse` | 1.5s | Orbiting glow for agent activity |
| `pulse-subtle` | 3s | Box-shadow pulse for deliverables |

### Agent Node Iconography

| Category | Color | Hex | Icon | Roles |
|----------|-------|-----|------|-------|
| Cognitive | Cyan | `#06b6d4` | Cpu | architect, coder, creative, chat |
| Sensory | Blue | `#3b82f6` | Eye | observer, ingress |
| Actuation | Green | `#10b981` | Zap | sentry, executor |
| Ledger | Muted | `#71717a` | Database | archivist, memory |

### Status Indicators

| Status | Glow Color | Hex |
|--------|------------|-----|
| online | Green | `#28C76F` |
| busy | Amber | `#f59e0b` |
| error | Red | `#ef4444` |
| offline | Gray | muted |

### Legacy Vuexy Variables (Deprecated — DO NOT USE in new code)

Still in `globals.css` for unmigrated pages (`/telemetry`, `/marketplace`, `/approvals`):

```css
--background: 37 41 60;       /* #25293C */
--foreground: 207 211 236;    /* #CFD3EC */
--surface: 47 51 73;          /* #2F3349 */
--surface-hover: 55 60 85;    /* #373C55 */
--border: 67 73 104;          /* #434968 */
--border-active: 115 103 240; /* #7367F0 */
```

---

## VIII. Architectural Patterns

### Hydration Safety
- Workspace pages: `next/dynamic({ ssr: false })` — ReactFlow and Zustand cannot safely SSR
- ReactFlow canvas is **ALWAYS** mounted (never conditionally unmount)
- Empty state = absolute-positioned overlay on top of mounted ReactFlow

### Blueprint → Graph Conversion
- `blueprintToGraph()` converts MissionBlueprint to ReactFlow nodes/edges
- Teams = group nodes, agents = child nodes within groups
- Agent outputs → input topics = edge sources → targets
- Ghost-draft mode: 50% opacity + dashed cyan border + `ghost-draft` animation
- On commit: nodes solidify (100% opacity, solid border)

### Fractal Navigation
- Double-click team group node → `enterSquadRoom(teamId)` → SquadRoom replaces CircuitBoard
- Back button → `exitSquadRoom()` → return to full DAG
- `activeSquadRoomId` in Zustand controls which view is active

### Signal Flow
- SSE: `/api/v1/stream` → EventSource → Zustand `streamLogs[]` (capped at 100)
- Signal dispatch: `artifact` → `pendingArtifacts[]`; `governance_halt` → GovernanceModal; default → NatsWaterfall
- NatsWaterfall: signal direction classification (input/output/internal), filter tabs

### Draft vs Active Mode (Wiring)
- **Draft mode:** Local Zustand mutations + `blueprintToGraph()` re-run. No API calls.
- **Active mode:** Backend API calls (`PUT`/`DELETE /missions/{id}/agents/{name}`) + local state sync

### API Response Pattern
- Most calls return raw JSON or arrays
- Some use `APIResponse<T>`: `{ ok, data, error }`
- Council chat returns `CTSChatEnvelope` with metadata + trust score

### Chart Rendering
- `ChartRenderer` accepts `MycelisChartSpec` and dispatches to:
  - Observable Plot: bar, line, area, dot, waffle, tree
  - Leaflet: geo (lat/lng markers)
  - DataTable: tabular data
- Charts inherit cortex palette for text, lines, domains
- `ArtifactViewer` renders charts inline from agent artifacts

---

## IX. Library Files

### `lib/utils.ts`
```typescript
cn(...inputs: ClassValue[]) → string  // clsx + tailwind-merge
```

### `lib/signalNormalize.ts`
```typescript
streamSignalToDetail(signal: StreamSignal) → SignalDetail
logEntryToDetail(entry: LogEntry) → SignalDetail
```

### `lib/types/protocol.ts`
Protocol envelope types (EnvelopeType, Envelope<T>, ThoughtContent, MetricContent, ArtifactContent, GovernanceContent)

### `lib/mock.ts`
Mock data for development

---

## X. File Map (Quick Reference)

| Area | Key Files |
|------|-----------|
| Store | `store/useCortexStore.ts` |
| Shell | `components/shell/{ShellLayout,ZoneA_Rail,ZoneB_Workspace,ZoneC_Stream,ZoneD_Decision,GovernanceModal}.tsx` |
| Workspace | `components/workspace/{Workspace,ArchitectChat,CircuitBoard,SquadRoom,BlueprintDrawer,DeliverablesTray,ToolsPalette,TrustSlider,CognitiveStatusPanel}.tsx` |
| Dashboard | `components/dashboard/{MissionControl,MissionControlChat,OperationsBoard,TelemetryRow,SensorLibrary,SignalContext,ManifestationPanel}.tsx` |
| Wiring | `components/wiring/{AgentNode,CircuitBoard,DataWire,WiringAgentEditor}.tsx` |
| Charts | `components/charts/{ChartRenderer,MapRenderer,DataTable}.tsx` |
| Teams | `components/teams/{TeamsPage,TeamCard,TeamDetailDrawer}.tsx` |
| Stream | `components/stream/{NatsWaterfall,SignalDetailDrawer}.tsx` |
| Settings | `components/settings/{MCPToolRegistry,MCPLibraryBrowser,MCPServerCard,MCPInstallModal,ModelHealth}.tsx` |
| Catalogue | `components/catalogue/{CataloguePage,AgentCard,AgentEditorDrawer}.tsx` |
| Missions | `components/missions/{MissionSummaryTab,AgentActivityFeed,ArtifactViewer,TeamActuationView,TeamRoster}.tsx` |
| Memory | `components/memory/{MemoryExplorer,HotMemoryPanel,WarmMemoryPanel,ColdMemoryPanel}.tsx` |
| Shared | `components/shared/{UniversalRenderer,ArtifactCard,ApprovalCard,ThoughtCard,MetricPill}.tsx` |
| Lib | `lib/{utils,signalNormalize}.ts`, `lib/types/protocol.ts` |
| Styling | `app/globals.css` |

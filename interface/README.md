# Mycelis Cortex UI

The conscious face of the Mycelis Swarm OS. Built with Next.js 16, React 19, Tailwind v4, ReactFlow 11, and Zustand 5.

## Stack

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| Framework | Next.js 16.1.6 (Turbopack) | App Router, SSR, API routes |
| State | Zustand 5.0.11 | Atomic store (`useCortexStore`) — single source of truth |
| Graph | ReactFlow 11.11.4 | DAG visualization (agent nodes, data wires) |
| Styling | Tailwind CSS v4 | Utility-first, dark theme, custom animations |
| Icons | Lucide React | Consistent icon set |

## Architecture

### The Shell Layout

```
+------+-------------------------------------------+
| Zone | Zone B (Workspace)                        |
|  A   | ArchitectChat | CircuitBoard / SquadRoom  |
| Rail |               |                           |
|      |---------------|---------------------------|
|      | Spectrum (NatsWaterfall — collapsible)     |
+------+-------------------------------------------+
              Zone D (Governance Overlay)
```

- **Zone A** (`ZoneA_Rail.tsx`): Navigation rail + system vitals
- **Zone B** (`ZoneB_Workspace.tsx`): Page content area
- **Spectrum** (`NatsWaterfall.tsx`): Collapsible bottom panel — real-time SSE signal waterfall
- **Zone D** (`GovernanceModal.tsx`): Human-in-the-loop governance overlay — z-50 backdrop-blur modal with two-column layout (output + proof-of-work), APPROVE & DISPATCH / REJECT & REWORK controls
- **Deliverables Tray** (`DeliverablesTray.tsx`): Bottom-docked panel showing pending `CTSEnvelope` artifacts — auto-hides when empty, pulsing green glow signals pending human action

### Fractal Navigation

Double-click a team group node in CircuitBoard to drill into the **SquadRoom** — a fractal sub-view showing the team's internal debate feed and proof-of-work artifacts. Press Back to return to the circuit board.

State managed via `activeSquadRoomId` in Zustand:
- `enterSquadRoom(teamId)` — swaps CircuitBoard for SquadRoom
- `exitSquadRoom()` — returns to CircuitBoard

### Human-in-the-Loop Governance

When an agent emits a `signal: artifact` via SSE, the store intercepts it and constructs a `CTSEnvelope` into `pendingArtifacts[]`. The **DeliverablesTray** renders these as clickable cards. Clicking opens the **GovernanceModal** for review.

State managed via Zustand:

- `pendingArtifacts: CTSEnvelope[]` — auto-populated from SSE artifact signals
- `selectedArtifact: CTSEnvelope | null` — triggers GovernanceModal when set
- `approveArtifact(id)` — removes from pending, logs decision
- `rejectArtifact(id, reason)` — removes from pending, logs decision with reason

### State Flow (Unidirectional)

```
User Intent
    |
    v
ArchitectChat --> submitIntent() --> POST /api/v1/intent/negotiate
    |
    v
Zustand Store (useCortexStore)
    |
    +-- chatHistory[] --> ArchitectChat (read-only)
    +-- nodes[], edges[] --> CircuitBoard (ReactFlow)
    +-- activeSquadRoomId --> SquadRoom (fractal drill-down)
    +-- blueprint, missionStatus --> CircuitBoard metadata bar
    +-- streamLogs[] --> NatsWaterfall (bottom panel)
    +-- pendingArtifacts[] --> DeliverablesTray (bottom-docked cards)
    +-- selectedArtifact --> GovernanceModal (z-50 overlay)
    +-- isThinking per node --> AgentNode activity ring
```

### SSE Stream Pipeline

```
Backend /api/v1/stream (EventSource)
    |
    v
initializeStream() --> Zustand streamLogs[] (capped at 100)
    |
    +-- NatsWaterfall renders signal waterfall
    +-- dispatchSignalToNodes() matches signal.source to node ID
        |
        +-- thought/cognitive --> isThinking=true, cyan ring
        +-- artifact/output --> isThinking=false, CTSEnvelope → pendingArtifacts[]
        +-- error --> status='error'
```

## Key Files

| File | Purpose |
| :--- | :--- |
| `store/useCortexStore.ts` | Zustand store: all state, actions, SSE consumer |
| `components/workspace/Workspace.tsx` | Split-pane: ArchitectChat + CircuitBoard/SquadRoom + bottom panel |
| `components/workspace/ArchitectChat.tsx` | Chat UI + intent submission |
| `components/workspace/CircuitBoard.tsx` | ReactFlow canvas + Instantiate button + fractal drill-down |
| `components/workspace/SquadRoom.tsx` | Fractal sub-view: team debate feed + proof-of-work viewer |
| `components/wiring/AgentNode.tsx` | Custom node: status, role, thought bubble, activity ring |
| `components/wiring/DataWire.tsx` | Custom animated edge |
| `components/stream/NatsWaterfall.tsx` | Collapsible bottom panel — real-time signal spectrum |
| `components/workspace/DeliverablesTray.tsx` | Bottom-docked pending artifact cards (CTSEnvelope[]) |
| `components/shell/GovernanceModal.tsx` | Zone D — approval overlay: output + proof columns, approve/reject |
| `components/shell/ShellLayout.tsx` | Shell layout: ZoneA + ZoneB + ZoneD |
| `app/globals.css` | Vuexy Dark palette (`cortex-*` tokens), animations, ReactFlow overrides |

## Development

Use the `inv` (invoke) task runner from the project root:

```bash
inv interface.dev       # Start dev server (Turbopack)
inv interface.build     # Production build
inv interface.lint      # ESLint check
inv interface.test      # Run Vitest unit tests
inv interface.check     # Smoke-test running server (fetches key pages, checks for errors)
inv interface.stop      # Kill dev server by port
inv interface.clean     # Clear .next build cache
inv interface.restart   # Full restart: stop → clean → build → dev → check
```

Open [http://localhost:3000](http://localhost:3000) to see the Cortex UI.

### Pages

| Route | Component |
| :--- | :--- |
| `/architect` | Workspace (ArchitectChat + CircuitBoard) |
| `/dashboard` | System dashboard |
| `/wiring` | Workspace (Neural Wiring — same split-pane as Architect) |
| `/telemetry` | Telemetry view |
| `/settings` | Configuration |

## Design Tokens

Cortex palette defined via `@theme` in `app/globals.css`:

| Token | Hex | Usage |
| :--- | :--- | :--- |
| `cortex-bg` | `#25293C` | Page background, input fields |
| `cortex-surface` | `#2F3349` | Cards, panels, sidebar |
| `cortex-primary` | `#7367F0` | Active nav pills, focus rings, send button |
| `cortex-success` | `#28C76F` | Online indicators, approve button, heartbeat |
| `cortex-warning` | `#FF9F43` | Constraints, governance badges |
| `cortex-danger` | `#EA5455` | Errors, reject button |
| `cortex-info` | `#00CFE8` | Architect bot avatar, drafting indicator |
| `cortex-text-main` | `#CFD3EC` | Primary text |
| `cortex-text-muted` | `#7983BB` | Secondary text, placeholders |
| `cortex-border` | `#434968` | Borders, dividers |

Utilities: `.glow-stable`, `.glow-active`, `.glow-critical`, `.glow-primary`, `.glass`, `.ghost-draft`, `.activity-ring`, `.animate-pulse-subtle`

## Conventions

- **No useState for API/graph state** — all in Zustand store
- **Dumb components** — read slices from store, never manage their own async state
- **Ghost draft pattern** — nodes appear at 50% opacity with dashed cyan border, solidify on commit
- **Activity ring** — pulsing cyan border on agent nodes when `isThinking === true`

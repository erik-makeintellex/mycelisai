# Mycelis Interface
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

The frontend for the current Soma-first Mycelis operator product. Built with Next.js 16, React 19, Tailwind v4, ReactFlow 11, and Zustand 5.

## TOC

- [Stack](#stack)
- [Architecture](#architecture)
- [Key Files](#key-files)
- [Development](#development)
- [Pages](#pages)
- [Design Tokens](#design-tokens)
- [Conventions](#conventions)

## Stack

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| Framework | Next.js 16.1.6 (Turbopack) | App Router, SSR, API routes |
| State | Zustand 5.0.11 | Atomic store (`useCortexStore`) — single source of truth |
| Graph | ReactFlow 11.11.4 | DAG visualization (agent nodes, data wires) |
| Styling | Tailwind CSS v4 | Utility-first aero-light default theme, midnight alternate, custom animations |
| Icons | Lucide React | Consistent icon set |

## Architecture

### Product Surfaces

The current frontend is organized around three primary operator surfaces:

- **Dashboard** (`/dashboard`): Central Soma home plus AI Organization creation/re-entry
- **Organization workspace** (`/organizations/[id]`): Soma-led governed interaction inside a chosen AI Organization
- **Automations advanced workspace** (`/automations?tab=wiring`): wiring, graph editing, launch flow, and deeper execution controls

### Shell Layout

The app shell is implemented through `ShellLayout.tsx` and keeps the global product frame stable while routes swap underneath it.

- **Zone A** (`ZoneA_Rail.tsx`): workflow-first navigation, advanced toggle, and current-organization return path
- **Zone B** (`ZoneB_Workspace.tsx`): route content host
- **Global degraded/status surfaces**: `DegradedModeBanner.tsx` and `StatusDrawer.tsx`
- **Governed advanced overlays**: `GovernanceModal.tsx` plus route-local stream/detail surfaces in advanced workspaces

Route-local telemetry and signal inspection now live in advanced workspace surfaces such as `NatsWaterfall.tsx` and `SignalDetailDrawer.tsx`, not in the global shell.

### Soma-First Flow

The default operator path is no longer a generic architect console. The current delivery model is:

1. **Central Soma home** (`CentralSomaHome.tsx`) teaches one persistent Soma across governed contexts.
2. **AI Organization entry** (`CreateOrganizationEntry.tsx`) opens or creates the scoped working context.
3. **Organization workspace** (`OrganizationContextShell.tsx`) keeps the operator in one Soma-led panel while details, support views, and advanced execution remain reachable without taking over the default experience.

### Advanced Wiring Surface

The older graph-centric workflow still exists as an advanced surface rather than the primary product story.

- `Workspace.tsx` hosts the advanced wiring environment
- `ArchitectChat.tsx`, `CircuitBoard.tsx`, and `SquadRoom.tsx` still power mission drafting and graph inspection
- `LaunchCrewModal.tsx` bridges default Soma requests into governed launch/execution flows when crew creation is needed

### State Flow

`useCortexStore` is composed from bounded slices under `interface/store/`. The current hot paths are:

- chat and proposal handling for Soma-led interaction
- graph and draft state for advanced wiring
- governance, artifact, and runtime status handling
- organization settings, AI engine, and response-style surfaces
- stream/service health feeding degraded and diagnostics surfaces

## Key Files

| File | Purpose |
| :--- | :--- |
| `components/dashboard/CentralSomaHome.tsx` | Dashboard entry framing for one persistent Soma |
| `components/organizations/CreateOrganizationEntry.tsx` | AI Organization creation and recent-organization return path |
| `components/organizations/OrganizationContextShell.tsx` | Main Soma-led organization workspace |
| `components/dashboard/MissionControlChat.tsx` | Soma-first governed chat, proposal, and execution surface |
| `components/organizations/TeamLeadInteractionPanel.tsx` | Guided first-run and organization-support actions |
| `components/system/SystemQuickChecks.tsx` | Recovery-oriented system checks surfaced in the current product |
| `components/workspace/LaunchCrewModal.tsx` | Governed launch/crew execution modal |
| `components/workspace/Workspace.tsx` | Advanced wiring workspace host |
| `components/workspace/ArchitectChat.tsx` | Advanced intent negotiation chat for wiring mode |
| `components/workspace/CircuitBoard.tsx` | Advanced graph canvas and mission draft editor |
| `components/workspace/SquadRoom.tsx` | Team drill-down for advanced graph workflow |
| `components/wiring/AgentNode.tsx` | Custom graph node for wiring/editor surfaces |
| `components/wiring/DataWire.tsx` | Custom graph edge for wiring/editor surfaces |
| `components/stream/NatsWaterfall.tsx` | Route-local signal waterfall used in advanced surfaces |
| `components/shell/ShellLayout.tsx` | Shell layout: ZoneA + ZoneB + ZoneD |
| `app/globals.css` | Mycelis Interface theme tokens (`cortex-*` compatibility names), aero-light and midnight palettes, animations, ReactFlow overrides |

## Development

Use the `inv` (invoke) task runner from the project root:

```bash
uv run inv interface.dev       # Start dev server (Turbopack)
uv run inv interface.build     # Production build
uv run inv interface.lint      # ESLint check
uv run inv interface.test      # Run Vitest unit tests with sequential file execution
uv run inv interface.check     # Smoke-test running server (fetches key pages, checks for errors)
uv run inv interface.stop      # Kill dev server by port
uv run inv interface.clean     # Clear .next build cache
uv run inv interface.restart   # Full restart: stop → clean → build → dev → check
```

Open [http://localhost:3000](http://localhost:3000) to see the Mycelis Interface.

### Pages

| Route | Component |
| :--- | :--- |
| `/dashboard` | Central Soma home + AI Organization entry flow |
| `/organizations/[id]` | Soma-primary AI Organization workspace |
| `/automations` | Automation hub with approvals, trigger rules, teams, and advanced wiring |
| `/docs` | In-app docs browser |
| `/settings` | Preferences, mission profiles, people access, and advanced setup |
| `/resources` | Advanced support hub for connected tools, workspace files, AI engines, and role library |
| `/memory` | Advanced memory explorer |
| `/system` | Advanced diagnostics and recovery checks |

## Design Tokens

Mycelis Interface theme tokens are defined via `@theme` and runtime `data-theme` overrides in `app/globals.css`. The default product theme is `aero-light`; `midnight-cortex` remains an alternate. The `cortex-*` token names are retained as compatibility names across existing components:

| Token | Hex | Usage |
| :--- | :--- | :--- |
| `cortex-bg` | `#111821` | Page background, input fields |
| `cortex-surface` | `#1a2430` | Cards, panels, sidebar |
| `cortex-primary` | `#1ab4cf` | Active nav pills, focus rings, send button |
| `cortex-success` | `#23b26f` | Online indicators, approve button, heartbeat |
| `cortex-warning` | `#e6a63d` | Constraints, governance badges |
| `cortex-danger` | `#df5d5d` | Errors, reject button |
| `cortex-info` | `#4f8ef7` | Architect bot avatar, drafting indicator |
| `cortex-text-main` | `#e8eff7` | Primary text |
| `cortex-text-muted` | `#9fb0c2` | Secondary text, placeholders |
| `cortex-border` | `#324353` | Borders, dividers |

Utilities: `.glow-stable`, `.glow-active`, `.glow-critical`, `.glow-primary`, `.glass`, `.ghost-draft`, `.activity-ring`, `.animate-pulse-subtle`

## Conventions

- **Shared app state lives in the composed Cortex store** — keep cross-route behavior in bounded Zustand slices
- **Default presentation stays Soma-first** — advanced execution detail is reachable, but should not dominate the default operator flow
- **Advanced graph tooling remains available** — keep wiring/editor depth without making it the default product identity
- **Route docs and task contracts must stay in sync** — when product surfaces move, update the relevant repo docs in the same slice

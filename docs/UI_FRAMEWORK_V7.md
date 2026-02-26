# Mycelis V7 UI Framework
## Canonical Contract for Workflow-First, Parallel UI Delivery

Version: `2.0`  
Status: `Authoritative`  
Last Updated: `2026-02-26`  
Scope: `interface/*` and all docs that define UI behavior

This document is the single UI framework source of truth.  
If another document conflicts with this file, this file wins for UI implementation rules.

Planning companion:
- `docs/UI_ELEMENTS_PLANNING_V7.md` (research-backed element planning, Soma interaction patterns, and element management workflow)

---

## Table of Contents

1. [Purpose](#purpose)
2. [How Teams Use This](#how-teams-use-this)
3. [Core Product UX Rules](#core-product-ux-rules)
4. [UI Taxonomy](#ui-taxonomy)
5. [Canonical Input and Output Model](#canonical-input-and-output-model)
6. [Canonical State Model and Transitions](#canonical-state-model-and-transitions)
7. [Global Operational State Contract](#global-operational-state-contract)
8. [Status Semantics and Tokens](#status-semantics-and-tokens)
9. [Surface Composition Standard](#surface-composition-standard)
10. [Failure and Degraded UX Contract](#failure-and-degraded-ux-contract)
11. [Component Instantiation Templates](#component-instantiation-templates)
12. [Data Access and Side-Effect Standard](#data-access-and-side-effect-standard)
13. [Accessibility and Interaction Baseline](#accessibility-and-interaction-baseline)
14. [Observability and Telemetry Contract](#observability-and-telemetry-contract)
15. [Testing Matrix by Core Layer](#testing-matrix-by-core-layer)
16. [Parallel Delivery and Handoff Rules](#parallel-delivery-and-handoff-rules)
17. [PR Gate Checklist](#pr-gate-checklist)
18. [Release Gate Checklist](#release-gate-checklist)
19. [Deviation Policy](#deviation-policy)
20. [Appendix A - Surface Spec Template](#appendix-a---surface-spec-template)
21. [Appendix B - Component Spec Template](#appendix-b---component-spec-template)
22. [Appendix C - Failure Card Template](#appendix-c---failure-card-template)

---

## Purpose

Mycelis UI is an operational control surface, not a chat wrapper.

Every screen must answer in under 2 seconds:
1. What can I do here?
2. What is system state now?
3. What is my next best step?

This framework enforces a trusted, continuous structure as UI scope expands.

---

## How Teams Use This

1. Classify work as `Primitive`, `Composite`, or `Surface`.
2. Define input/output contract before coding.
3. Implement required state model (`loading/ready/empty/degraded/error`).
4. Apply failure contract and operational status semantics.
5. Complete unit/integration/E2E/reliability/performance tests for lane scope.
6. Attach evidence to lane document before merge.

Parallel lane references:
- `docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md`
- `docs/ui-delivery/LANE_A_GLOBAL_OPS.md`
- `docs/ui-delivery/LANE_B_WORKSPACE_RELIABILITY.md`
- `docs/ui-delivery/LANE_C_WORKFLOW_SURFACES.md`
- `docs/ui-delivery/LANE_D_SYSTEM_OBSERVABILITY.md`
- `docs/ui-delivery/LANE_Q_QA_REGRESSION.md`

---

## Core Product UX Rules

1. No dead ends.
2. No unexplained failures.
3. No raw backend errors as final user copy.
4. Every failure includes at least 2 recovery actions.
5. Every non-ready state includes a visible next step.
6. Operational state is always visible from any page.

---

## UI Taxonomy

### Primitive
- Example: badge, icon-state, button, input row, label.
- Behavior: pure presentation, no fetch logic.

### Composite
- Example: drawer, banner, card, tab panel, list row.
- Behavior: receives data, renders states, emits actions.

### Surface
- Example: route-level page.
- Behavior: owns orchestration, fetch, and view-model mapping.

Rule:
- Fetch in Surface.
- Normalize once.
- Pass normalized models downward.

---

## Canonical Input and Output Model

Each Surface and Composite must define:
- Inputs
- Transformations
- Outputs
- Emitted events

```ts
type UIInput = {
  routeParams?: Record<string, string>;
  query?: Record<string, string>;
  storeSnapshot?: unknown;
  sseEvent?: unknown;
  apiPayload?: unknown;
};

type UIOutput = {
  renderState: "loading" | "ready" | "empty" | "degraded" | "error";
  blocks: string[]; // e.g. ["status", "primary_action", "diagnostics"]
  actions: Array<{ id: string; label: string; intent: "primary" | "secondary" }>;
  telemetryEvents: string[];
};
```

---

## Canonical State Model and Transitions

All Surface and Composite elements must implement this contract:

```ts
type UIState<T> = {
  status: "loading" | "ready" | "empty" | "degraded" | "error";
  data?: T;
  reason?: string;
  nextActions: Array<{ id: string; label: string }>;
  diagnostics?: Record<string, unknown>;
};
```

Allowed transition rules:
- `loading -> ready | empty | degraded | error`
- `ready -> degraded | error | loading`
- `degraded -> ready | error | loading`
- `error -> loading | degraded`
- `empty -> loading | degraded | ready`

Disallowed:
- Any implicit state without declared `status`
- Collapsing `degraded` into `empty`

---

## Global Operational State Contract

```ts
type UiOperationalState = {
  nats: "healthy" | "degraded" | "offline";
  sse: "healthy" | "degraded" | "offline";
  db: "healthy" | "degraded" | "offline";
  council: Record<string, "healthy" | "degraded" | "offline">;
  llmProvider: { name: string; mode: "local" | "remote" };
  governance: "passive" | "approval_required" | "halted";
  activeMissions: number;
  degraded: boolean;
  lastUpdatedAt: string;
};
```

Required consumers:
- `StatusDrawer`
- `DegradedModeBanner`
- Workspace status strip
- System quick checks

---

## Status Semantics and Tokens

Canonical semantics:
- Green = healthy
- Yellow = degraded
- Red = failure
- Gray = offline
- Blue = informational

Required shared token names:

```css
:root {
  --status-healthy: #22c55e;
  --status-degraded: #eab308;
  --status-failure: #ef4444;
  --status-offline: #6b7280;
  --status-info: #38bdf8;
}
```

No page-level overrides for semantic meaning.

---

## Surface Composition Standard

Every Surface must include:
1. Purpose line (what this surface is for)
2. Current state line (what is healthy/degraded)
3. Primary action (single best next action)
4. Secondary actions (1-3 options)
5. Diagnostics path (status/runs/logs)

Minimum layout contract:
- Top: status and purpose
- Middle: main workflow content
- Bottom or side: diagnostics and recovery entry points

---

## Failure and Degraded UX Contract

Every failure/degraded block must include:
1. What happened
2. Likely cause
3. Impact
4. Next actions (2+)
5. Copy diagnostics action

Canonical action set:
- Retry
- Open Status Drawer
- Continue in degraded mode (if safe)
- Switch to Soma/default path

Forbidden:
- Showing only `500` or raw error strings
- Redirecting user away without inline recovery option

---

## Component Instantiation Templates

### Required metadata for any new component

```ts
type ComponentContract = {
  name: string; // XxxCard/XxxDrawer/XxxBanner/XxxPanel/XxxTab
  type: "Primitive" | "Composite" | "Surface";
  dataSource: "props" | "store" | "props+store";
  states: Array<"loading" | "ready" | "empty" | "degraded" | "error">;
  actions: string[];
  diagnosticsPath?: string;
  testIds: string[];
};
```

Folder standards:
- `interface/components/shared/*`
- `interface/components/dashboard/*`
- `interface/components/automations/*`
- `interface/components/system/*`

---

## Data Access and Side-Effect Standard

Default route behavior:
1. Read store snapshot.
2. Start fetch/poll in effect.
3. Normalize payload into view model.
4. Render by state contract.
5. Expose inline retries.

Polling defaults:
- Critical health: `5-10s`
- Operational telemetry: `10-15s`
- Reference data: `30-60s`

SSE rules:
- On disconnect: set degraded state immediately.
- Keep last-good timestamp visible.
- Provide retry/reconnect actions inline.

---

## Accessibility and Interaction Baseline

Required:
- Keyboard access for all interactive controls
- Focus-visible styles
- `aria-label` for icon-only actions
- Escape behavior for overlays
- Live regions for high-priority failure/degraded updates
- No focus traps in drawer/banner/focus mode toggles

---

## Observability and Telemetry Contract

Each Surface and Composite must emit UI telemetry for:
- `ui_render_state_changed`
- `ui_action_clicked`
- `ui_error_presented`
- `ui_recovery_attempted`
- `ui_recovery_succeeded|failed`

Every telemetry event should include:
- `surface`
- `component`
- `status`
- `run_id` (if present)
- `timestamp`

---

## Testing Matrix by Core Layer

| Core Layer | Unit | Integration | E2E | Reliability | Performance |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Intent | Input mapping, status labels | chat/council response mapping | message -> actionability | timeout/unreachable fallback | first interaction latency |
| Orchestration | provider/gov badges | proposal/approval mapping | propose -> approve -> execute | provider/gov degraded handling | metadata non-blocking render |
| Event Spine | stream-state mapping | services status wiring | banner + drawer lifecycle | reconnect and jitter behavior | burst event resilience |
| Execution | capability cards and permissions | MCP/toolset status mapping | capability discovery <=2 clicks | denied/offline fallback | large catalog responsiveness |
| Observability | filters, tabs, axis semantics | runs/events/chain binding | run trace <=2 clicks | partial payload tolerance | timeline smooth at 1000+ |

Mandatory for delivery:
- New UI behavior must have tests across all relevant classes for its layer.

---

## Parallel Delivery and Handoff Rules

Execution model:
1. Lanes run in parallel.
2. Gates merge in sequence (`Gate A -> Gate B -> Gate C -> RC`).
3. Lane Q validates each gate before promotion.

Handoff contract between lanes:
- Shared components: version and usage notes
- Store changes: state keys + selectors + migration notes
- API changes: endpoint/schema impact
- Tests: what was added and how to run
- Evidence: screenshots/video/log references

No lane merge without completed handoff note.

---

## PR Gate Checklist

- [ ] Component/surface declares all 5 states
- [ ] Failure/degraded UX contract implemented
- [ ] Status semantics follow canonical tokens
- [ ] Accessibility checks pass
- [ ] Unit + integration + E2E + reliability coverage added where applicable
- [ ] Diagnostics path visible
- [ ] Docs updated (framework/lane/state as needed)

---

## Release Gate Checklist

- [ ] Gate A suites pass
- [ ] Gate B suites pass
- [ ] Gate C suites pass
- [ ] Cross-layer regression passes
- [ ] No non-actionable failure copy
- [ ] Performance targets met for critical surfaces
- [ ] Evidence attached in lane docs

---

## Deviation Policy

If a change cannot follow this framework:
1. Record deviation in a decision log.
2. State rationale and risk.
3. Add compensating tests.
4. Define reconciliation plan and target date.

Undocumented deviation is not allowed.

---

## Appendix A - Surface Spec Template

```md
Surface: <name>
Owner Lane: <A|B|C|D|Q>
Purpose: <one sentence>
Primary Action: <single best action>
Secondary Actions:
- <action 1>
- <action 2>
Diagnostics Path: <status/runs/logs>

Inputs:
- route/query:
- store selectors:
- api endpoints:
- stream events:

ViewModel:
- shape:
- mapper:

States:
- loading:
- ready:
- empty:
- degraded:
- error:

Telemetry:
- events emitted:

Tests:
- unit:
- integration:
- e2e:
- reliability:
- performance:
```

---

## Appendix B - Component Spec Template

```md
Component: <name>
Type: Primitive|Composite
Surface(s): <where used>
Inputs:
- props:
- store selectors (if any):
Outputs:
- rendered blocks:
- emitted events:
States:
- loading:
- ready:
- empty:
- degraded:
- error:
Actions:
- onRetry:
- onOpenStatus:
- onFallback:
Diagnostics:
- copied fields:
Test IDs:
- <id>
Accessibility:
- keyboard:
- aria labels:
- focus behavior:
```

---

## Appendix C - Failure Card Template

```md
What happened:
Likely cause:
Impact:
Next actions:
- Retry
- Open Status Drawer
- Continue degraded mode
- Route via Soma/default path
Diagnostics:
- run_id:
- subsystem:
- timestamp:
```

# V7 Step 01 — Workflow-First Information Architecture

> **Status:** Implemented
> **Date:** 2026-02-21
> **Scope:** Frontend navigation restructure only — no backend changes

## Navigation Structure

| # | Nav Label        | Route          | Visibility       | Tabs |
|---|------------------|----------------|------------------|------|
| 1 | Mission Control  | `/dashboard`   | Always           | — (single pane) |
| 2 | Automations      | `/automations` | Always           | Active Automations, Draft Blueprints, Trigger Rules, Approvals, Teams, Neural Wiring* |
| 3 | Resources        | `/resources`   | Always           | Brains, MCP Tools, Workspace Explorer, Capabilities |
| 4 | Memory           | `/memory`      | Always           | — (single pane) |
| 5 | System           | `/system`      | Advanced only    | Event Health, NATS Status, Database, Cognitive Matrix, Debug |
|   | Settings         | `/settings`    | Always (footer)  | — |

\* Neural Wiring tab requires `advancedMode === true`.

## Route Mapping (Legacy → V7)

| Legacy Route       | V7 Destination                       | Method            |
|--------------------|--------------------------------------|-------------------|
| `/wiring`          | `/automations?tab=wiring`            | Server redirect   |
| `/architect`       | `/automations?tab=wiring`            | Server redirect   |
| `/teams`           | `/automations?tab=teams`             | Server redirect   |
| `/catalogue`       | `/resources?tab=catalogue`           | Server redirect   |
| `/marketplace`     | `/resources?tab=catalogue`           | Server redirect   |
| `/approvals`       | `/automations?tab=approvals`         | Server redirect   |
| `/telemetry`       | `/system?tab=health`                 | Server redirect   |
| `/matrix`          | `/system?tab=matrix`                 | Server redirect   |

## Route Tree

```
app/(app)/
├── dashboard/page.tsx        → Mission Control (unchanged)
├── automations/page.tsx      → Tabbed: active | drafts | triggers | approvals | teams | wiring*
├── resources/page.tsx        → Tabbed: brains | tools | workspace | catalogue
├── memory/page.tsx           → Memory (unchanged)
├── system/page.tsx           → Tabbed: health | nats | database | matrix | debug
├── settings/page.tsx         → Settings (unchanged)
├── wiring/page.tsx           → redirect('/automations?tab=wiring')
├── architect/page.tsx        → redirect('/automations?tab=wiring')
├── teams/page.tsx            → redirect('/automations?tab=teams')
├── catalogue/page.tsx        → redirect('/resources?tab=catalogue')
├── marketplace/page.tsx      → redirect('/resources?tab=catalogue')
├── approvals/page.tsx        → redirect('/automations?tab=approvals')
├── telemetry/page.tsx        → redirect('/system?tab=health')
└── matrix/page.tsx           → redirect('/system?tab=matrix')
```

## Key Components

| Component | Location | Role |
|-----------|----------|------|
| `ZoneA_Rail` | `components/shell/ZoneA_Rail.tsx` | Sidebar navigation (5 items + advanced toggle + settings) |
| `DegradedState` | `components/shared/DegradedState.tsx` | Placeholder for unimplemented tabs (title, reason, unavailable/available lists) |
| `ApprovalsTab` | `components/automations/ApprovalsTab.tsx` | Full approvals content (Queue, Policy CRUD, Proposals) |

## Patterns

### Tab Deep-Linking

Each tabbed page reads `?tab=` from URL search params via `useSearchParams()`. This requires a `<Suspense>` boundary per Next.js 16 rules.

```tsx
export default function Page() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <PageContent />
        </Suspense>
    );
}
```

### Advanced Mode Gating

- `advancedMode` boolean lives in `useCortexStore`, persisted to `localStorage('mycelis-advanced-mode')`.
- ZoneA_Rail: System nav item conditionally rendered.
- Automations page: Neural Wiring tab conditionally rendered; URL `?tab=wiring` falls back to `active` when advanced off.

### Legacy Redirects

Server components using `redirect()` from `next/navigation`. No `"use client"` directive — runs at server level.

```tsx
import { redirect } from 'next/navigation';
export default function LegacyRedirect() {
    redirect('/target?tab=id');
}
```

## Definition of Done

- [x] 5 workflow-first nav items in ZoneA_Rail
- [x] System tab gated behind `advancedMode`
- [x] Neural Wiring tab gated behind `advancedMode`
- [x] 8 legacy routes redirect to correct V7 parent + tab
- [x] Tab deep-linking via `?tab=` search params on all 3 tabbed pages
- [x] Every tab renders content or DegradedState (no dead panels)
- [x] PolicyTab CRUD migrated from standalone `/approvals` into `ApprovalsTab`
- [x] Zero `bg-white` or light-mode classes
- [x] Midnight Cortex theme throughout

## Test Coverage

### Unit Tests (Vitest)

| Test File | Coverage |
|-----------|----------|
| [`ZoneA_Rail.test.tsx`](../../interface/__tests__/shell/ZoneA_Rail.test.tsx) | V7 nav entries, active/inactive styling, advanced toggle, no old labels, no bg-white |
| [`AutomationsPage.test.tsx`](../../interface/__tests__/pages/AutomationsPage.test.tsx) | All tabs render, Neural Wiring gated, deep-link via `?tab=`, default tab |
| [`ResourcesPage.test.tsx`](../../interface/__tests__/pages/ResourcesPage.test.tsx) | All tabs render, deep-link via `?tab=`, default tab |
| [`SystemPage.test.tsx`](../../interface/__tests__/pages/SystemPage.test.tsx) | All tabs render, Advanced badge, deep-link via `?tab=`, default tab |
| [`WiringPage.test.tsx`](../../interface/__tests__/pages/WiringPage.test.tsx) | Redirect to `/automations?tab=wiring` |
| [`TeamsPage.test.tsx`](../../interface/__tests__/pages/TeamsPage.test.tsx) | Redirect to `/automations?tab=teams` |
| [`CataloguePage.test.tsx`](../../interface/__tests__/pages/CataloguePage.test.tsx) | Redirect to `/resources?tab=catalogue` |
| [`ApprovalsPage.test.tsx`](../../interface/__tests__/pages/ApprovalsPage.test.tsx) | Redirect to `/automations?tab=approvals` |
| [`TelemetryPage.test.tsx`](../../interface/__tests__/pages/TelemetryPage.test.tsx) | Redirect to `/system?tab=health` |
| [`MarketplacePage.test.tsx`](../../interface/__tests__/pages/MarketplacePage.test.tsx) | Redirect to `/resources?tab=catalogue` |
| [`MatrixPage.test.tsx`](../../interface/__tests__/pages/MatrixPage.test.tsx) | Redirect to `/system?tab=matrix` |
| [`ArchitectPage.test.tsx`](../../interface/__tests__/pages/ArchitectPage.test.tsx) | Redirect to `/automations?tab=wiring` |
| [`DegradedState.test.tsx`](../../interface/__tests__/shared/DegradedState.test.tsx) | Renders title/reason, unavailable/available lists, action text, theme compliance |
| [`PrimaryRoutes.test.tsx`](../../interface/__tests__/pages/PrimaryRoutes.test.tsx) | Console error guards for all primary routes |

### E2E Tests (Playwright)

| Spec | Coverage |
|------|----------|
| [`navigation.spec.ts`](../../interface/e2e/specs/navigation.spec.ts) | V7 nav entries, route highlighting, advanced toggle, nav order, legacy redirects (6 URLs), no bg-white |

### Manual Verification

| Doc | Purpose |
|-----|---------|
| [`v7-step-01-ui.md`](../verification/v7-step-01-ui.md) | Step-by-step manual verification script for QA |

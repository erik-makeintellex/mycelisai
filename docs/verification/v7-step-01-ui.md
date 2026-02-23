# V7 Step 01 — Manual UI Verification Script

> **Prerequisite:** `inv interface.dev` running on `localhost:3000`
> **Time estimate:** 10–15 minutes
> **Reference:** [`docs/product/ia-v7-step-01.md`](../product/ia-v7-step-01.md)

---

## 1. Navigation Rail (ZoneA)

### 1.1 Default State (Advanced Off)

- [ ] Visit `http://localhost:3000/dashboard`
- [ ] Verify ZoneA sidebar shows exactly 4 nav items: **Mission Control**, **Automations**, **Resources**, **Memory**
- [ ] Verify **Settings** link in footer section
- [ ] Verify **Advanced: Off** toggle button in footer
- [ ] Verify **System** nav item is NOT visible
- [ ] Verify **no old labels** visible: Neural Wiring, Team Management, Agent Catalogue, Skills Market, Governance, System Status, Cognitive Matrix

### 1.2 Active Route Highlighting

- [ ] On `/dashboard`, verify Mission Control link has `bg-cortex-primary` (cyan background)
- [ ] Verify all other nav items have muted text (no cyan background)
- [ ] Click **Automations** — verify it gets cyan background, Mission Control loses it
- [ ] Click **Resources** — verify highlight follows
- [ ] Click **Memory** — verify highlight follows

### 1.3 Advanced Mode Toggle

- [ ] Click **Advanced: Off** toggle
- [ ] Verify text changes to **Advanced: On**
- [ ] Verify **System** nav item appears between Memory and Settings
- [ ] Click **System** — verify it navigates to `/system` and gets highlighted
- [ ] Click **Advanced: On** toggle — verify System disappears from nav
- [ ] Verify you are still on `/system` page (no redirect on toggle-off)

### 1.4 Theme Compliance

- [ ] Inspect ZoneA sidebar background — should be `bg-cortex-surface` (#18181b)
- [ ] Verify NO `bg-white`, `bg-slate-*`, or `bg-zinc-*` classes in the rail
- [ ] Verify text colors use `cortex-text-main` and `cortex-text-muted`

---

## 2. Automations Page (`/automations`)

### 2.1 Tab Rendering

- [ ] Visit `http://localhost:3000/automations`
- [ ] Verify page title: **Automations**
- [ ] Verify subtitle: "Scheduled missions, trigger rules, drafts, and approvals"
- [ ] Verify 5 tabs visible (Advanced Off): **Active Automations**, **Draft Blueprints**, **Trigger Rules**, **Approvals**, **Teams**
- [ ] Verify **Neural Wiring** tab is NOT visible

### 2.2 Default Tab Content

- [ ] Verify "Active Automations" tab is active (cyan underline)
- [ ] Verify DegradedState panel shows with title "Scheduled Missions"
- [ ] Verify "Unavailable" list: Cron scheduling, Recurring mission execution
- [ ] Verify "Still Working" list: Manual mission activation, One-time mission execution

### 2.3 Tab Navigation

- [ ] Click **Draft Blueprints** — verify DegradedState with title "Draft Blueprints"
- [ ] Click **Trigger Rules** — verify DegradedState with title "Trigger Rules"
- [ ] Click **Approvals** — verify approvals content loads (Queue/Policy/Proposals sub-tabs)
- [ ] Click **Teams** — verify teams content loads

### 2.4 Neural Wiring (Advanced Mode)

- [ ] Toggle Advanced Mode ON (via ZoneA footer)
- [ ] Visit `/automations` — verify **Neural Wiring** tab appears (6th tab)
- [ ] Click **Neural Wiring** — verify Workspace/ReactFlow canvas loads
- [ ] Toggle Advanced Mode OFF — verify Neural Wiring tab disappears, tab falls back to Active Automations

### 2.5 Deep-Linking

- [ ] Visit `http://localhost:3000/automations?tab=approvals` — verify Approvals tab is active
- [ ] Visit `http://localhost:3000/automations?tab=teams` — verify Teams tab is active
- [ ] Visit `http://localhost:3000/automations?tab=invalid` — verify defaults to Active Automations

---

## 3. Resources Page (`/resources`)

### 3.1 Tab Rendering

- [ ] Visit `http://localhost:3000/resources`
- [ ] Verify page title: **Resources**
- [ ] Verify 4 tabs: **Brains**, **MCP Tools**, **Workspace Explorer**, **Capabilities**

### 3.2 Default Tab Content

- [ ] Verify "Brains" tab is active (cyan underline)
- [ ] Verify BrainsPage component loads (brain/provider management)

### 3.3 Tab Navigation

- [ ] Click **MCP Tools** — verify MCP tool registry loads
- [ ] Click **Workspace Explorer** — verify DegradedState with title "Workspace Explorer"
- [ ] Click **Capabilities** — verify catalogue page loads

### 3.4 Deep-Linking

- [ ] Visit `http://localhost:3000/resources?tab=catalogue` — verify Capabilities tab is active
- [ ] Visit `http://localhost:3000/resources?tab=tools` — verify MCP Tools tab is active

---

## 4. System Page (`/system`) — Advanced Only

### 4.1 Access

- [ ] Toggle Advanced Mode ON
- [ ] Click **System** in nav or visit `http://localhost:3000/system`
- [ ] Verify **Advanced** badge (amber/warning style) in header

### 4.2 Tab Rendering

- [ ] Verify 5 tabs: **Event Health**, **NATS Status**, **Database**, **Cognitive Matrix**, **Debug**

### 4.3 Default Tab Content

- [ ] Verify "Event Health" tab is active
- [ ] Verify LIVE/OFFLINE status indicator
- [ ] Verify 4 metric cards: Goroutines, Heap Alloc, Sys Memory, Token Rate (values or "...")

### 4.4 Tab Navigation

- [ ] Click **NATS Status** — verify NATS JetStream connection status panel
- [ ] Click **Database** — verify PostgreSQL + pgvector status panel
- [ ] Click **Cognitive Matrix** — verify MatrixGrid renders
- [ ] Click **Debug** — verify debug console with build info

### 4.5 Deep-Linking

- [ ] Visit `http://localhost:3000/system?tab=matrix` — verify Cognitive Matrix tab is active
- [ ] Visit `http://localhost:3000/system?tab=debug` — verify Debug tab is active

---

## 5. Legacy Route Redirects

For each legacy URL, verify browser navigates to the correct V7 destination:

| # | Visit URL | Expected Final URL | Expected Active Tab |
|---|-----------|-------------------|---------------------|
| 1 | `/wiring` | `/automations?tab=wiring` | Neural Wiring (if Advanced on) or Active Automations |
| 2 | `/architect` | `/automations?tab=wiring` | Same as above |
| 3 | `/teams` | `/automations?tab=teams` | Teams |
| 4 | `/catalogue` | `/resources?tab=catalogue` | Capabilities |
| 5 | `/marketplace` | `/resources?tab=catalogue` | Capabilities |
| 6 | `/approvals` | `/automations?tab=approvals` | Approvals |
| 7 | `/telemetry` | `/system?tab=health` | Event Health |
| 8 | `/matrix` | `/system?tab=matrix` | Cognitive Matrix |

---

## 6. Console Error Check

- [ ] Open browser DevTools → Console
- [ ] Visit each primary route: `/dashboard`, `/automations`, `/resources`, `/memory`, `/system`
- [ ] Verify **zero console errors** on each page
- [ ] Warnings about missing backend APIs are acceptable (Core not running)
- [ ] No React hydration errors
- [ ] No unhandled promise rejections

---

## 7. Theme Compliance (Global)

- [ ] Inspect all visited pages — verify zero `bg-white` classes
- [ ] Verify all backgrounds use `cortex-bg` (#09090b) or `cortex-surface` (#18181b)
- [ ] Verify text uses `cortex-text-main` (#d4d4d8) and `cortex-text-muted` (#71717a)
- [ ] Verify borders use `cortex-border` (#27272a)
- [ ] Verify primary accent is `cortex-primary` (#06b6d4 cyan)

---

## Result

| Section | Pass/Fail | Notes |
|---------|-----------|-------|
| 1. Navigation Rail | | |
| 2. Automations Page | | |
| 3. Resources Page | | |
| 4. System Page | | |
| 5. Legacy Redirects | | |
| 6. Console Errors | | |
| 7. Theme Compliance | | |

**Verified by:** _______________
**Date:** _______________

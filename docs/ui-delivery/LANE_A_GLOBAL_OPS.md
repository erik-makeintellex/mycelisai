# Lane A - Global Ops UX

Priority: `P0`  
Status: `In Progress`  
Primary goal: reliable global operational truth across all routes.
Owner team: `Team Circuit`

## Scope

1. `StatusDrawer` live system state and diagnostics actions
2. `DegradedModeBanner` trigger and auto-clear behavior
3. Shared global status semantics (healthy/degraded/failure/offline/info)

## Files

- `interface/components/dashboard/StatusDrawer.tsx`
- `interface/components/dashboard/DegradedModeBanner.tsx`
- `interface/components/shell/ShellLayout.tsx`
- `interface/components/dashboard/ModeRibbon.tsx`
- `interface/store/useCortexStore.ts`

## Deliverables

1. Drawer reachable from any route
2. Banner appears on critical outages and clears on recovery
3. Retry/Open Drawer/Switch actions available from degraded states
4. Consistent status badges and color semantics

## Dependencies

1. `/api/v1/services/status` stable response mapping
2. SSE/state hooks for live refresh

## Acceptance Tests

1. Unit:
   - drawer open/close and reachability rendering
   - banner visibility mapping by subsystem state
2. Integration:
   - services payload transforms into UI state contract
3. E2E:
   - outage triggers banner; recovery auto-clears without reload
4. Reliability:
   - NATS/SSE disconnected states show actionable next steps

## Exit Criteria

1. No global status dead-end states
2. Banner and drawer remain in sync under reconnect churn
3. P0 Gate A tests pass with evidence attached

## Sprint 0 Focus

1. unify health-state mapping across banner, drawer, and system checks
2. finalize reconnect action behavior for NATS and SSE degraded states
3. publish stable health selectors for dependent lanes

## Evidence

- Test output:
  - `cd interface; npx vitest run __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/shell/ShellLayout.test.tsx --reporter=dot`
  - Result: pass (`10` tests)
- Video/screenshot:
- Notes:
  - Added status drawer accessibility affordances (`role="dialog"`, close aria-label).

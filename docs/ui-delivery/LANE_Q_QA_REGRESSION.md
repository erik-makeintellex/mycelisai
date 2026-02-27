# Lane Q - QA Reliability and Regression

Priority: `Cross-cutting`  
Status: `In Progress`  
Primary goal: enforce release gates across all UI lanes with evidence-backed reliability.
Owner team: `Team Sentinel`

## Scope

1. Cross-lane regression matrix and gate enforcement
2. Playwright flows for failure recovery and degraded lifecycle
3. Performance and reliability checks tied to architecture layers

## Inputs

- [UI Framework V7](../UI_FRAMEWORK_V7.md)
- [UI Parallel Delivery Board](./PARALLEL_DELIVERY_BOARD.md)
- [mycelis-architecture-v7.md](../../mycelis-architecture-v7.md)

## Required Suites

1. P0:
   - council failure -> reroute via Soma in one click
   - degraded banner appears and clears automatically
   - status drawer reflects failing subsystem
2. P1:
   - automations hub actionable when scheduler disabled
   - teams diagnostic path in <= 2 clicks
3. P2:
   - system quick checks status + timestamp correctness
   - runs/timeline responsiveness with high event counts

## Gate Enforcement

1. Gate A:
   - Lane A + B critical tests green
2. Gate B:
   - Lane C workflow and reliability tests green
3. Gate C:
   - Lane D observability + performance checks green
4. Gate RC:
   - full cross-layer suite green

## Exit Criteria

1. Every gate includes test output and visual evidence
2. No known non-actionable failure copy remains
3. Regression checklist signed before merge window closes

## Sprint 0 Focus

1. formalize Gate A and Gate B evidence format for all lanes
2. add reliability cases for readiness gate partial data and bus reconnect churn
3. enforce cross-lane handoff checks for state-shape drift

## Evidence

- Unit/Integration:
  - `cd interface; npx vitest run __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/pages/AutomationsPage.test.tsx __tests__/shell/ShellLayout.test.tsx __tests__/pages/SystemPage.test.tsx __tests__/teams/TeamsPage.test.tsx --reporter=dot`
  - Result: pass (`46` tests)
- E2E:
  - Gate A Playwright suite scaffolded: `interface/e2e/specs/v7-operational-ux.spec.ts` (6 tests)
  - Parse/list check completed:
    - `cd interface; npx playwright test e2e/specs/v7-operational-ux.spec.ts --list`
  - Execution attempt status:
    - `cd interface; npx playwright test e2e/specs/v7-operational-ux.spec.ts --project=chromium`
    - Blocked: Playwright browser executable missing (`npx playwright install` required)
- Performance:
  - Pending
- Reliability:
  - Covered in component-level fallback and actionability tests above
- Signoff:
  - Gate A baseline test hardening complete; proceed to broader gate suite

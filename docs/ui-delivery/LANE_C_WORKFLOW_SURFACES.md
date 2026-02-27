# Lane C - Workflow Surfaces UX

Priority: `P1`  
Status: `In Progress`  
Primary goal: make automations, teams, and resources operationally actionable.
Owner team: `Team Helios + Team Argus`

## Scope

1. `AutomationHub` default actionable layout
2. Teams cards with diagnostic quick actions and health signals
3. Resources visibility for execution capabilities and governance posture

## Files

- `interface/components/automations/AutomationHub.tsx`
- `interface/components/automations/TeamInstantiationWizard.tsx`
- `interface/components/automations/CapabilityReadinessGateCard.tsx`
- `interface/components/automations/RouteTemplatePicker.tsx`
- `interface/app/(app)/automations/page.tsx`
- `interface/components/teams/TeamCard.tsx`
- `interface/components/teams/TeamsPage.tsx`
- `interface/app/(app)/resources/page.tsx`
- `interface/lib/workflowContracts.ts`
- `interface/__tests__/automations/RouteTemplatePicker.test.tsx`

## Deliverables

1. Automations page never presents dead scheduler-only state
2. Team cards show online count, heartbeat, health, and quick actions
3. Resources page communicates capability access and constraints

## Dependencies

1. Lane A shared degraded/status primitives
2. MCP/resources backend data availability for capability display

## Acceptance Tests

1. Unit:
   - hub cards and CTA rendering by state
   - team card status badges and action handlers
2. Integration:
   - automations/resources payloads map into UI framework state contract
3. E2E:
   - user can start first automation chain from landing hub
   - unreachable agent diagnosable from Teams in <= 2 clicks
4. Reliability:
   - degraded data sources still show next-action path

## Exit Criteria

1. No workflow surface is empty without next actions
2. Teams and resources maintain diagnostics path under degraded states
3. Gate B tests pass with evidence attached

## Sprint 0 Focus

1. scaffold `TeamInstantiationWizard` and `CapabilityReadinessGateCard`
2. define shared readiness and channel metadata contracts for surfaces
3. add team-card quick pivots to latest run and logs

## Evidence

- Test output:
  - `cd interface; npx vitest run __tests__/automations/TeamInstantiationWizard.test.tsx __tests__/automations/RouteTemplatePicker.test.tsx __tests__/pages/AutomationsPage.test.tsx __tests__/teams/TeamsPage.test.tsx`
  - Result: pass (`13` tests)
- Video/screenshot:
- Notes:
  - Sprint 0 scaffolds added for guided instantiation, readiness gate contracts, and NATS route exposure mode controls.
  - Launch and propose-only actions now persist mission profiles and activate selected profile routes.

# Lane C - Workflow Surfaces UX

Priority: `P1`  
Status: `Planned`  
Primary goal: make automations, teams, and resources operationally actionable.

## Scope

1. `AutomationHub` default actionable layout
2. Teams cards with diagnostic quick actions and health signals
3. Resources visibility for execution capabilities and governance posture

## Files

- `interface/components/automations/AutomationHub.tsx`
- `interface/app/(app)/automations/page.tsx`
- `interface/components/teams/TeamCard.tsx`
- `interface/components/teams/TeamsPage.tsx`
- `interface/app/(app)/resources/page.tsx`

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

## Evidence

- Test output:
- Video/screenshot:
- Notes:


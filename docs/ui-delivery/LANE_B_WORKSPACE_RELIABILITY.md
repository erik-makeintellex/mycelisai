# Lane B - Workspace Reliability UX

Priority: `P0`  
Status: `In Progress`  
Primary goal: remove non-actionable failure paths from workspace interactions.

## Scope

1. `CouncilCallErrorCard` structured error contract
2. Inline retry/reroute behavior without retyping
3. Focus mode workflow continuity (`F` shortcut)

## Files

- `interface/components/dashboard/CouncilCallErrorCard.tsx`
- `interface/components/dashboard/MissionControlChat.tsx`
- `interface/components/dashboard/FocusModeToggle.tsx`
- `interface/components/dashboard/MissionControl.tsx`
- `interface/store/useCortexStore.ts`

## Deliverables

1. Error card includes what happened, likely cause, impact, next actions
2. Retry preserves message context
3. Switch-to-Soma reroutes inline
4. Focus mode collapse/expand persists per session

## Dependencies

1. Lane A status actions available globally
2. Workspace request/response state hooks in store

## Acceptance Tests

1. Unit:
   - error card content and action callbacks
   - focus mode toggle and keyboard shortcut behavior
2. Integration:
   - council failures map to structured card states
3. E2E:
   - council failure reroute via Soma in one click
4. Reliability:
   - timeout/unreachable/500 variants all actionable

## Exit Criteria

1. No raw 500 strings visible in workspace failure UX
2. User can recover from council failure in <= 2 actions
3. P0 Gate A tests pass with evidence attached

## Evidence

- Test output:
  - `cd interface; npx vitest run __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/CouncilCallErrorCard.test.tsx --reporter=dot`
  - Result: pass (`23` tests)
- Video/screenshot:
- Notes:
  - MissionControl error-state assertions updated to structured error-card behavior.

# UI Target And Transaction Contract V7
> Navigation: [Project README](../../README.md) | [V8 UI/API Contract](../architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) | [Testing](../TESTING.md)

Status: V7 migration input with active testing vocabulary.

This file no longer owns the full UI PRD. Use [V8 UI/API and Operator Experience Contract](../architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) for current screen/API behavior and [Testing](../TESTING.md) for validation policy.

## Retained Transaction Rules

Every user-visible action should end in one of these terminal UI states:
- `answer`
- `proposal`
- `execution_result`
- `blocker`
- `empty`
- `loading`
- `error`

Invalid outcomes:
- silent no-op
- raw backend exception
- unbounded spinner
- mutation without visible approval/proof
- output that cannot be reviewed or retained

## Backend/API -> UI Target Plan

When backend/API behavior changes, include the plan from [Testing](../TESTING.md#backendapi---ui-target-plan):
- backend/API change
- impacted UI surfaces
- expected terminal state
- failure/recovery expectation
- evidence commands

## Current UI Path Families

Current proof should map to the V8 surfaces:
- AI Organization entry and Soma workspace
- direct Soma answer
- team/group creation and retained output review
- governed proposal cancel/execute
- resources and Connected Tools
- memory and continuity
- runs/activity/audit
- settings and advanced boundaries

## Required Test Layers

Use the narrowest layer that proves the risk:
- backend handler/service tests for payload/runtime behavior
- store/component tests for UI state transitions
- Playwright for operator-visible workflows
- live-backend Playwright for proxy/Core/runtime contracts
- remote user testing for delivered topology proof

## Migration Rule

Do not expand this V7 file with new route matrices. Promote active requirements into V8 UI docs, [Testing](../TESTING.md), and `.state/V8_DEV_STATE.md`.

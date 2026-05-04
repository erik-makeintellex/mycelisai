# Intent To Manifestation And Team Interaction V7
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8 UI/API Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)

Status: historical V7 migration input.

This file preserves the V7 design signal for intent-to-manifestation and team interaction. Current authority lives in V8 UI/API, runtime, bootstrap, and workflow docs.

## Retained V7 Signal

Mycelis should turn operator intent into visible, governed work:
- user intent enters through Soma
- Soma may answer directly or propose structured work
- broad work should become compact, reviewable lanes rather than one giant roster
- created teams/groups need interaction, status, result, and retained-output review surfaces
- channel metadata and UI state must stay aligned
- manifestation must remain auditable

## Current Translation

| V7 idea | Current owner |
| --- | --- |
| Intent stack | V8 UI/API contract and Soma workspace |
| Team expression | groups/teams workflow and V8 workflow variants |
| Channel inspector | Resources, groups, activity, and run timeline surfaces |
| Metadata envelope | backend CTS/API/event contracts |
| GUI manifestation | V8 UI Team Full Test Set |

## MVP Target State

For current slices, the minimum target remains:
- direct Soma answers for non-mutating prompts
- proposal state for protected/mutating prompts
- compact team or temporary workflow lane when work needs multiple roles
- visible status/result/audit trail
- retained outputs that survive refresh

## Migration Rule

Promote active requirements into:
- [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
- [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md)
- [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md)
- [Testing](../TESTING.md)

Do not add new canonical runtime subjects or UI route matrices here.

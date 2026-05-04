# V8 UI Team Browser Workflows
> Navigation: [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md)

Status: ACTIVE workflow detail.

## Critical Headed Chromium Matrix

Run sequentially:

```bash
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-ui-testing-agentry.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.setup.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.template-creation.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.empty-start.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.continuity.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.workspace-persistence.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.native-vs-external.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.ask-class.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.recovery.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/teams.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/team-creation.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/settings.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/mcp-connected-tools.spec.ts
```

## Workflow Coverage

The matrix certifies:
- Central Soma home
- primary navigation
- template organization creation
- empty-start organization creation
- organization re-entry and continuity
- guided start and runtime summary
- direct Soma answer
- inline content vs artifacts
- native team instantiation
- external workflow-contract instantiation
- media/output readiness
- governed mutation proposal, cancel, and execute
- activity/audit visibility
- settings persistence
- permissions and AI Engine inspectability
- advanced mode boundaries
- disruption and recovery

## Route Families

Required route families:
- `/dashboard`
- `/organizations/[id]`
- `/groups`
- `/teams`
- `/teams/create`
- `/resources`
- `/memory`
- `/system`
- `/settings`
- `/runs`, `/runs/[id]`, `/runs/[id]/chain`

## Evidence

Capture screenshots or traces for:
- entry/re-entry
- direct answer
- proposal/cancel
- proposal/execute
- retained output
- activity/timeline
- recovery blocker

# V8 New-User Acceptance Matrix
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Testing](../TESTING.md)

This matrix is the browser-facing acceptance contract for first-run setup,
local-source deployment, login, Resources, MCP, workspace roots, team work, demo
outputs, and recovery. A delivery slice touching those surfaces must leave these
gates provable from a fresh operator session.

| Gate | Operator action | Required proof |
| --- | --- | --- |
| Login boundary | Open `/`, land on `/login`, sign in with local owner or enabled SSO | Login explains local owner vs enterprise SSO; invalid credentials and domain mismatch fail closed with actionable copy; `/dashboard` shows the signed-in Soma operating environment with Access, Identity, and Scope. |
| Dashboard orientation | Review `/dashboard` before asking anything | Soma expression input is primary; Trust does not imply prior work; Active Work is attention-first and capped, with full backlog linked to `/teams`; Output and Context explain where retained results and setup evidence will appear. |
| Setup checklist | Review setup surfaces before a demo | Operator can identify provider/search posture, Connected Tools/MCP posture, workspace root, artifact/output root, and deployment trust route without reading logs. |
| Provider readiness | Open Settings/Resources/System as needed | Configured local or hosted provider posture is visible; missing provider or search configuration becomes setup truth or degraded guidance, not a hidden chat failure. |
| MCP readiness | Open `Resources -> Connected Tools`, choose **Add MCP Server** when needed | Curated library path is visible; required env vars are named without secret values; installed server/tool list and recent MCP activity confirm what Soma and teams can use. |
| Workspace/output roots | Open `Resources -> Output Files` and `System -> Deployments` when file output is expected | `MYCELIS_WORKSPACE` and `MYCELIS_ARTIFACT_ROOT` are identifiable; filesystem MCP writes and project packages stay inside the governed workspace boundary, and **Open folder** exposes the local generated-content folder. |
| Canonical demo | Ask Soma for a concrete retained demo output, such as the first-demo project package | Browser proof shows proposal/approval when mutation is required, run/proof linkage, retained output/package metadata, open/reload behavior, and reviewable output after refresh. |
| Team active/degraded proof | Use `/teams` or Active Work to ask/respond/recover for one bounded team work item | Durable `TeamWorkItem` reaches `output_ready` with readable output or `degraded` with timeout/offline/unreadable proof, recovery option, status events/interactions, and no raw bus/topic dump in the default row. |

Suggested focused browser proof after local source changes:

```bash
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/homepage.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/mcp-connected-tools.spec.ts
uv run inv interface.e2e --headed --project=chromium --workers=1 --spec=e2e/specs/desktop-mobile-compression.spec.ts
$env:PLAYWRIGHT_ACTIVE_WORK_API_LIVE='1'; uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/active-work-api.spec.ts
```

Use `active-work-ask-live.spec.ts` instead of the broader API proof when the
question is specifically whether a responsive local runtime team can return a
readable `output_ready` reply through the GUI. Set
`PLAYWRIGHT_TEAM_WORK_GUI_LIVE=1` and the target team id before that stricter
proof, and record skipped live gates as environment skips rather than accepted
new-user delivery proof.

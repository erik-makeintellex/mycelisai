# V8.3 MVP UI Runtime Detail Checklist
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-06-16
> Purpose: Keep detailed route and component guidance out of the compact delivery plan while preserving the visible-UX gate.

## Usability North Star

Every major route should answer three questions in the first viewport:

1. What can I do here?
2. What currently needs my attention?
3. Where is the output or proof?

Rules:

- one primary job per page
- one primary action per row or card
- details open on selection, not by extending the whole page
- long prompts, logs, outputs, and lists scroll inside bounded panels
- default labels describe operator outcomes, not implementation objects
- visual density increases only after Advanced or Inspect
- generated content has an obvious open path and folder path
- failure copy names the safe next action
- first-run views start with starters/templates, not empty admin tables

## Route Audit

| Surface | Required Direction |
| --- | --- |
| Dashboard / Soma | Use Soma Work Inbox as the attention model; keep chat, active work, latest output, and recovery tied to the selected context. |
| Review Work / Teams | Use inbox/list-detail with one row action and selected detail for output, trust, recovery, conversation, and Inspect. |
| Groups | Keep the master/detail rail and selected-group tabs; standardize output cards with Soma and Resources. |
| Resources | Promote Capabilities as answer cards for available/degraded/missing state, risk, approval, outputs, last use, and repair; raw MCP stays detail. |
| Runs | Default to Receipt, then Timeline, Output, Proof, Inspect; list views group Active, Needs Review, Completed, Failed, Archived. |
| Activity | Treat as advanced operations review; ordinary output/proof paths route to Work Inbox or Run Receipt. |
| System | Lead with capability impact and recovery state; service commands stay secondary. |
| Docs | Keep user docs first; architecture remains searchable but not the default learning path. |
| Settings | Keep personal preferences separate from deployment-owned access, engines, and tool setup. |
| Automations | Show schedules as governed future work with approval/review status and last output/proof. |

## Simplification Patterns

| Problem | Preferred Pattern | Avoid |
| --- | --- | --- |
| Many records | bounded list/detail with filters collapsed | page-length record stacks |
| Many details for one record | tabs inside selected detail | cards stacked below cards |
| Technical proof | receipt first, timeline second, raw payload last | JSON/event payload in default card |
| Many next actions | primary action plus More menu | five equal buttons |
| Missing capability | availability card with repair action | raw server error or env dump |
| Generated content | package card with preview/open/folder/proof | raw HTML/code as main output |
| Advanced workflow logic | Inspect map or stepper | default node graph |
| Long-running work | inbox state plus progressive receipt | blocking spinner or hidden background task |
| Search/current info | source-boundary trust line | generic model limitation boilerplate |
| Sensitive/private action | confirmation with scope/risk/output | hidden mutation or ambiguous consent |

## Starter Families

Complex work should start from Soma and compact starters instead of blank complex forms:

- Create output: report, browser app/game, media pack, code package, data review
- Review output: inspect proof, summarize quality, compare versions, prepare handoff
- Research: use local sources, use web/search, cite evidence, store output
- Build team: smallest useful team, specialist-output team, temporary group lane
- Recover: repair search, reconnect filesystem, retry failed media, resume queued work
- Deploy: check local readiness, run release proof, inspect deployment roots

Each starter should name expected output, approval requirement, likely capability use, result location, and recovery path if a capability is missing.

## Target State Components

| Pattern | Component Direction |
| --- | --- |
| Soma Work Inbox | `ActiveWorkLane`, `ReviewQueueSummary`, `WorkReviewInbox`, and `SomaWorkspaceFrame` show Needs Review, Running, Outputs, Failed/Recovery, and Archived through a bounded inbox. |
| Run Receipt | `ExecutionSummaryCard`, `ExecutionSummaryCompactCard`, `ExpressionFrame`, and run pages show outcome, status, output, trust, proof, recovery, and open/download/inspect actions. |
| Output Package Card | `OutputWorkbench`, `OutputWorkbenchProjectPackage`, `OutputAccessActions`, `GroupOutputsPanel`, and `WorkspaceExplorer` expose preview/open/folder/Resources/README/PROOF/download/copy-path/proof. |
| Bounded Routes | Groups, Resources, Runs, Run Detail, and System use menu/detail or list/detail panes with inner scroll. |
| Capability Catalog | Resources answers what Soma can use, what needs repair, and what can be requested before raw MCP topology. |
| Recovery Queue | Recovery copy states Failed, Still trusted, Not trusted, and Safe next, with retry/repair/continue/archive actions. |
| Advanced Run Map | Run map/timeline remains Inspect-only; start with a readable stepper before a graph. |

## Visible Slice Review Gate

Before accepting a visible slice:

- A new user can explain the page job in five seconds.
- Primary action is visible without page scrolling at desktop and short laptop heights.
- There is one obvious next action for the current state.
- Generated content opens directly and exposes storage/folder access.
- Failure says what remains trusted and what to do next.
- Raw logs, bus subjects, MCP internals, and JSON stay behind Inspect.
- The page uses bounded internal scroll instead of growing the route.
- Keyboard focus reaches every tab, action, and input.
- Mobile preserves the same primary action.

# V8.3 MVP UI Runtime Delivery Plan
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-06-14
> Purpose: Turn the V8.3 operational embodiment PRD and current agent-platform market patterns into executable UI/runtime delivery slices.

## Source Review

This plan does not replace the [V8.3 Operational Embodiment PRD](V8_3_OPERATIONAL_EMBODIMENT_PRD.md). It details how to deliver the PRD's operator-visible threshold:

```text
I tell Soma what I want.
Soma safely directs execution.
I can see active work, outputs, proof, and recovery.
I can trust what happened later.
```

The current external market pattern now validates this direction. LangSmith/LangGraph emphasizes agent deployment around assistants, threads, runs, streaming, human review, MCP/A2A, and production failure diagnosis. CrewAI presents production-ready crews, flows, triggers, deployment, team management, RBAC, and live-run monitoring. Dify, Flowise, Gumloop, and n8n make workflows legible through visual builders, execution logs, integrations, templates, and bounded app/workspace concepts. OpenAI Agents SDK confirms that tight primitives plus tracing, guardrails, handoffs, sessions, MCP, and workspace execution are enough for real delivery without exposing every abstraction to the user.

Mycelis should not copy visual node builders as the default experience. The Mycelis advantage is Soma-first governed execution with progressive operational disclosure. The visual graph belongs behind Inspect, while the default surface should be an operator inbox, receipt, output package, and recovery path.

The [V8.3 Soma User Experience Contract](V8_3_SOMA_USER_EXPERIENCE_CONTRACT.md) is the operator-facing acceptance lens for this plan: users work through Soma, deliverables matter more than responses, proof/recovery must be visible, and advanced Inspect is optional.

## Delivery Thesis

The release-candidate product should feel less like a topology console and more like an operating inbox for governed work.

Default operator shape:

```text
Intent -> Soma -> Work Inbox -> Run Receipt -> Output Package -> Proof / Recovery
```

Advanced operator shape:

```text
Inspect -> Timeline / Run Map -> Capability / Event / Proof Detail -> Deployment Roots
```

The runtime can keep deep concepts. The UI should translate them into a few stable objects:

| Object | Operator Meaning | Default Action |
| --- | --- | --- |
| Work Item | Something Soma is doing or needs reviewed | Review, open output, recover, clear |
| Run Receipt | What happened and whether to trust it | Open output, view proof, retry |
| Output Package | Durable result worth revisiting | Preview, Open file, Open folder, Open in Resources, download |
| Capability | What Soma can use safely | Check availability, enable, inspect risk |
| Recovery Item | A failed or degraded thing that can be handled | Retry, repair dependency, archive |
| Run Map | Technical execution trace | Advanced inspect only |

## Current Product Review

Implemented foundations:

- `ExpressionFrame` and execution-summary components already model direct answers, proposals, results, blockers, recovery, trust, and compact receipts.
- `SomaOperatingSurface`, `SomaCurrentWorkLane`, and `MissionControlChat` already support a Soma-first workbench.
- `OutputWorkbench`, `OutputWorkbenchProjectPackage`, and `OutputAccessActions` already support retained files, project packages, folder access, and proof details.
- `ActiveWorkLane`, `WorkTruthSummary`, `ReviewDecisionGuide`, and team-work action handlers already support active/review/degraded team work.
- `/groups` now uses a bounded master-detail workspace with selected-group tabs and openable saved outputs.
- `/resources` already has Output Files, Capabilities, Exchange, Deployment Context, AI Engines, and Role Library concepts.
- `/runs/[id]` already exposes timeline events and conversation tabs.

Remaining coherence gaps:

- Review work is still experienced as dense row/card detail instead of a clean operator inbox with selected-item detail.
- Result and output cards are similar but not yet one shared receipt/package pattern across Dashboard, Teams, Groups, Resources, and Runs.
- Run Timeline is audit-readable, but not yet a concise run receipt plus advanced visual/timeline inspection.
- Resources now presents Capabilities first, but full capability families and repair actions still need broader runtime coverage.
- Recovery is partially actionable but not yet a unified queue with failed/still-trusted/not-trusted/retry/operator-action language.
- Visual workflow/run map does not exist as an advanced Inspect view.
- Current proof is strong in focused paths, but release proof needs one standardized user journey after the UI language converges.

## UI Pattern Gate

Detailed route audits, simplification patterns, starter families, and target-state component notes live in [V8.3 MVP UI Runtime Detail Checklist](V8_3_MVP_UI_RUNTIME_DETAIL_CHECKLIST.md). Keep this plan as the execution map and use the checklist when a slice changes visible workflow.

Visible slices must still satisfy the same gate:

- first viewport answers what the page does, what needs attention, and where output/proof lives
- generated content opens directly and exposes folder/storage access
- failures name what remains trusted, what is not trusted, and the safe next action
- raw logs, bus subjects, MCP internals, JSON, and topology stay behind Inspect
- long lists, logs, prompts, and outputs scroll inside bounded panels instead of growing the route
- desktop, short laptop, mobile, keyboard focus, and active browser-window proof all pass

## Delivery Lane Plan

### Lane A - Output And Runtime Proof Foundation

Owner focus: P0.1 through P0.3.

Deliverables:

1. Make the output package card the first shared MVP object across Soma, Groups, and Resources.
2. Absorb shared language helpers and test fixtures as support for output package states, actions, and proof copy.
3. Make service health/runtime proof visible before deeper review and catalog work.
4. Define the headed browser proof path for the output package and health/runtime surfaces.

Proof:

- unit coverage for file, media, document, and `project_package`
- runtime proof evidence for service health states
- headed browser proof for output open/folder/resources paths

### Lane B - Operator Attention And Capability Readiness

Owner focus: P0.4 and P0.5.

Deliverables:

1. Convert review work into an inbox/list-detail pattern after output and proof surfaces are stable.
2. Standardize row language around `needs review`, `running`, `output ready`, `needs recovery`, and `archived`.
3. Reframe Resources around capability availability, risk, approval, output types, and recovery.
4. Keep raw MCP/tool details behind selected capability detail or Inspect.

Proof:

- focused Vitest for active/review and capability states
- headed browser proof for Dashboard review work and Resources capability states
- docs update for `docs/user/resources.md` when capability meaning changes

### Lane C - Receipts And Recovery

Owner focus: P0.6 and P0.7.

Deliverables:

1. Add `/runs/[id]` receipt tab before timeline.
2. Move raw event detail into Timeline/Inspect.
3. Create unified recovery queue language and action model.
4. Keep advanced run-map work behind Inspect after receipts and recovery are stable.

Proof:

- unit coverage for run receipt status and recovery copy
- headed browser proof for failed/degraded run detail
- docs update for `docs/user/run-timeline.md` when run meaning changes

### Lane D - Release And Documentation Alignment

Owner focus: P0.8 and P0.9.

Deliverables:

1. Maintain a single Trusted Outcome Journey proof script:
   `Ask -> Understand -> Approve -> Execute -> Deliver -> Trust -> Recover -> Revisit`.
2. Keep the first proof deterministic and browser-visible before adding live-provider variability.
3. Keep headless and headed variants aligned with the authoritative P0 order.
4. Complete documentation alignment only after the MVP proof evidence is current.
5. Reject delivery if the Soma input is unreachable, if page scroll hides primary actions, or if any UI-visible success lacks durable API/runtime proof.

Proof:

- fresh-state GUI proof from Soma ask through retained output, trust proof, recovery, and revisit
- `uv run inv interface.e2e --server-mode=external --project=chromium --workers=1`
- headed live-backend proof for changed visible paths
- `uv run inv quality.max-lines --limit 300`
- docs link/workflow tests

## Implementation Order

### P0.1 - Output Package Standard

Status: IN_REVIEW

Unify output package cards in Soma and Groups, because generated outputs are the core MVP proof object.

Exit:

- shared package path/action helper lives in `interface/lib/outputPackageModel.ts`
- shared language and test fixture work supports output states, actions, receipt copy, and recovery copy
- Soma and Groups use the same package open/folder/resources labels
- generated HTML/app package opens rendered output
- folder/resources actions are visible
- raw source is secondary

### P0.2 - Service Health / Runtime Proof

Status: IN_REVIEW

Make service readiness and runtime proof visible before review, catalog, and recovery expansion.

Exit:

- `/system` and related runtime surfaces show what is healthy, degraded, or unavailable
- proof copy names affected operator paths
- repair guidance stays secondary to capability impact

### P0.3 - Headed Browser Proof

Status: COMPLETE

Prove the visible output and runtime paths in a headed browser before expanding the review inbox.

Exit:

- headed proof covers output open/folder/resources behavior
- headed proof covers health/runtime proof states
- viewport checks include 1366x720, 1280x640, and 1280x560

### P0.4 - Review Inbox

Status: IN_REVIEW

Make Work Review understandable after output and proof paths are stable.

Exit:

- active/review work uses inbox/list-detail pattern
- one obvious action per row
- recovery and output-ready states are visually distinct

### P0.5 - Capability Catalog

Status: IN_REVIEW

Make `Resources -> Capabilities` the place to answer what Soma can use.

Exit:

- capability availability, risk, approval, output types, and recovery are visible
- MCP raw structure remains Inspect detail
- capability manifest output-schema labels do not expose raw tool refs as outputs

### P0.6 - Run Receipt Standard

Status: IN_REVIEW

Make runs readable as receipts before timeline/debugging.

Exit:

- `/runs/[id]` opens to receipt
- Timeline remains available
- default receipt shows outcome, status, output, trust, proof, recovery, and open/download/inspect actions

### P0.7 - Recovery Queue

Status: IN_REVIEW

Make degraded work actionable after the run receipt contract is stable.

Exit:

- failed/degraded paths show still-trusted/not-trusted/safe-next
- recovery actions cover retry, repair, continue without capability, and archive
- advanced run-map work remains behind Inspect

### P0.8 - Trusted Outcome Journey Proof

Status: ACTIVE
Run and record one stitched operator journey from clean committed state. This is not a subsystem sweep. It proves that a non-technical user can ask Soma for meaningful work, approve governed execution, open the delivered output, trust proof, recover from a controlled degraded state, and revisit the result later without understanding agents, MCP, runs, topology, or infrastructure.

Execution architecture:

1. Keep the deterministic browser proof at `interface/e2e/specs/trusted-outcome-journey.spec.ts`.
2. Keep reusable mocked runtime proof in `interface/e2e/support/trusted-outcome-journey.ts`.
3. Keep the live source-stack smoke at `interface/e2e/specs/trusted-outcome-journey-live.spec.ts`.
4. Start from Soma, then use Resources, Groups, and Runs only as proof and revisit destinations.
5. Pair every visible success with durable API/runtime proof: confirmation, team work, proof artifact, execution contract, run events, group outputs, workspace view, and folder reveal.
6. Treat service readiness as a gate through `lifecycle.status`, `lifecycle.health`, and `/api/v1/services/status`.

Current proof:

`trusted-outcome-journey.spec.ts` covers Ask, Understand, Approve, Execute, Deliver, Trust, Recover, and Revisit with deterministic run-event and recovery proof. `trusted-outcome-journey-live.spec.ts` proves the source-stack ask -> approval -> retained package -> rendered output -> proof artifacts/contracts -> Resources/Groups/Runs re-entry path, while exposing that live confirm-action run events may still be empty for this route.

Primary proof commands: deterministic `uv run inv interface.e2e --server-mode=dev --project=chromium --workers=1 --spec=e2e/specs/trusted-outcome-journey.spec.ts`; live `$env:PLAYWRIGHT_LIVE_BACKEND='1'; uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --workers=1 --spec=e2e/specs/trusted-outcome-journey-live.spec.ts`.

Exit:
- source gates pass
- headed GUI proof passes
- release preflight passes
- journey proof packet is filled with pass/fail/blocker status for each journey step
- live proof records exact environment skips instead of treating skipped gates as green
- release handoff records the accepted proof result before promotion

### P0.9 - Documentation Alignment
Status: NEXT

Align user, testing, operations, state, and in-app docs after the MVP proof result is known.

Exit:

- changed behavior has owning docs updated
- reviewed-but-unchanged docs are named in slice close-out
- `.state/V8_DEV_STATE.md` records the accepted release proof status

## Acceptance Standard
Each slice must report:

- operator problem solved
- files changed
- runtime/API contract touched or reviewed
- docs changed
- tests run
- headed/browser proof when visible UI changed
- remaining trust/recovery gaps

Do not accept a slice that makes the product more conceptually dense, increases whole-page scrolling, exposes raw topology by default, hides outputs, or leaves failed execution without an actionable next step.

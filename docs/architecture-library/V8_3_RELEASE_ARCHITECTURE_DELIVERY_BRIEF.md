# Mycelis V8.3 Release Architecture Delivery Brief
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-06-19
> Purpose: Provide a compact, shareable architecture delivery document for the current MVP-to-release path.
> Printable HTML: [V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.html](V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.html)

## Executive Summary

Mycelis V8.3 is converging from architecture expansion into operational embodiment.

The product target is:

```text
I tell Soma what I want.
Soma safely directs execution.
I can see active work, outputs, proof, and recovery.
I can trust what happened later.
```

Mycelis should now be described as a **Soma-centered governed cognitive operating environment**. The architecture remains deep internally, but the default user experience must collapse around visible execution, durable outputs, proof, recovery, and operator trust.

The operator-facing acceptance lens is defined in the [V8.3 Soma User Experience Contract](V8_3_SOMA_USER_EXPERIENCE_CONTRACT.md): users work through Soma, deliverables matter more than responses, proof/recovery must be visible, and advanced Inspect is optional.

The product-manifestation review is defined in the [V8.3 Product Manifestation Architecture Review](V8_3_PRODUCT_MANIFESTATION_REVIEW.md): every subsystem must justify user value, visibility, MVP classification, and adoption risk before it expands the visible product surface.

## Product Architecture Shape

The canonical operator path is:

```text
Intent
-> Soma understanding
-> governed proposal or direct answer
-> approved run when mutation/execution is required
-> durable output package
-> run receipt
-> proof, recovery, or follow-up
```

The default UI should not expose topology first. Advanced runtime detail belongs behind `Inspect`.

Autonomy remains future-facing and control-first. The governing boundary is [V8.3 Autonomy Control Architecture](V8_3_AUTONOMY_CONTROL_ARCHITECTURE.md): autonomous work may supply intent, but it must not bypass Soma, ExecutionContract, policy, governed runs, capabilities, events, output/proof, review, or recovery. V8.3 should only add autonomy foundations when they strengthen the current MVP's observability, interruptibility, proof, recovery, capability boundaries, and budget/policy readiness.

## Current Release Architecture Objects

| Object | Operator Meaning | Release Requirement |
| --- | --- | --- |
| Soma | Singular operating surface and continuity layer | Ask, review, approve, recover, and re-enter work from one trusted surface |
| Work Item | Something Soma or a team is doing, waiting on, or finished | Visible state, next action, output/proof links, and recovery path |
| Run Receipt | User-readable record of what happened | Outcome, status, output, trust, proof, recovery, inspect |
| Output Package | Durable generated result | Open file/app, open folder, open in Resources, proof, download/copy path where applicable |
| Capability | Governed tool/provider/integration Soma can use | Availability, risk, approval posture, output types, repair path |
| Recovery Item | Failed or degraded work that needs action | Failed, still trusted, not trusted, safe next |
| Advanced Run Map | Technical reconstruction of execution | Stepper/timeline/graph behind Inspect only |
| Autonomy Control Boundary | Guardrails for future automated intent sources | No silent mutation, no hidden learning, no self-granted permissions, and no unbounded execution |

## Trusted Outcome Journey

All remaining P0 work is evaluated through the Trusted Outcome Journey, not subsystem completion:

```text
Ask
-> Understand
-> Approve
-> Execute
-> Deliver
-> Trust
-> Recover
-> Revisit
```

Subsystems matter only insofar as they strengthen the journey. Output Packages strengthen Deliver. Run Receipts strengthen Trust. Recovery Queue strengthens Recover. Review Inbox strengthens Understand and Approve. Capability Catalog strengthens Trust. Resources and Groups convergence strengthens Revisit.

The MVP is complete when a non-technical user can complete the full journey without understanding agents, MCP, workflows, runs, topology, or infrastructure.

## MVP Delivery Spine

The release-candidate MVP should prove one excellent journey before broadening:

```text
User asks Soma to create or review meaningful work
-> Soma evaluates intent
-> proposal appears when governed execution is required
-> operator approves
-> execution creates a run/work item
-> output package appears
-> proof/audit/recovery is visible
-> user revisits and trusts the result later
```

This is the delivery benchmark for the alpha demonstration and release confidence gate.

## Delivery Ownership Model

### Lane A - Soma Inbox And Review

Owns Dashboard, review work, current work, and operator attention.

Deliver:

- Work Inbox tabs: Needs Review, Running, Outputs, Failed / Recovery, Archived
- list-detail review pattern
- one primary action per row
- selected detail with outcome, output, trust, recovery, conversation/events, and Inspect

### Lane B - Output Package And Receipt

Owns durable outputs and run receipts.

Deliver:

- one shared output package card
- Soma and Groups output card convergence
- rendered generated HTML/app opening
- folder/resources/proof access
- source/code as secondary detail, not the default output experience

### Lane C - Resources Capability Catalog

Owns tools, MCP, search, media, filesystem, AI engines, and runtime capability posture.

Deliver:

- capability-first Resources view
- available/degraded/missing/needs approval states
- risk and approval posture
- repair paths for missing capabilities
- raw MCP server detail behind Inspect

### Lane D - Runs, Proof, And Recovery

Owns run receipts, recovery queue, and advanced inspect.

Deliver:

- `/runs/[id]` receipt-first default
- timeline/proof/raw payload behind tabs
- unified recovery queue language
- first advanced run-map stepper from existing events

### Lane E - QA And Release Proof

Owns validation and release confidence.

Deliver:

- one MVP proof script:

```text
login
-> Soma ask
-> proposal
-> approve
-> running/active work
-> output package
-> proof
-> Resources re-entry
-> degraded sample recovery
```

- headless and headed variants
- service readiness and docs/test synchronization
- rejection of changes that hide Soma input, output access, or recovery actions

## Current P0 Delivery Train

The P0 order is now release-convergence order. Do not reorder it unless a blocker demands it.

| Priority | Slice | Journey Step | Status | Exit Criteria |
| --- | --- | --- | --- | --- |
| P0.1 | Output package standard | Deliver, Revisit | IN_REVIEW | Soma, Groups, and Resources use one retained deliverable package pattern with preview/open/download/proof/inspect/revisit; headless and headed live Resources/Groups re-entry proof is green |
| P0.2 | Service health and runtime proof | Execute, Trust | IN_REVIEW | Infrastructure, migrations, Core, Interface, and health reporting are clean and reachable |
| P0.3 | Headed browser proof | Ask through Revisit | COMPLETE | A real browser proves ask -> approve -> deliverable -> open -> Resources/Groups re-entry |
| P0.4 | Review inbox | Understand, Approve | IN_REVIEW | Dashboard and `/teams?view=work` now use focused review-inbox/list-detail patterns with summary, selected detail, one primary row action, recovery/clear/output actions, and headed proof green |
| P0.5 | Capability catalog | Trust | IN_REVIEW | Resources now opens as Capabilities for tool/search readiness, grouped can-use/needs-repair/can-request state, and MCP topology behind Inspect |
| P0.6 | Run receipt standard | Trust | NEXT | Runs open to outcome, status, trust, proof, recovery, and outputs before raw logs |
| P0.7 | Recovery queue | Recover | NEXT | Failure states show what failed, what remains trusted, what can retry, and what needs attention |
| P0.8 | Full MVP proof | Full journey | NEXT | Login -> Soma request -> proposal -> approval -> running work -> deliverable -> proof -> re-entry -> recovery is green |
| P0.9 | Documentation alignment | All journey steps | NEXT | State/docs explain proof and reality, not aspiration |

## P0.1 Current Implementation

The output-package standardization slice is the active release-convergence priority. The earlier shared language/test fixture work is supporting groundwork inside this slice, not a separate release priority.

Implemented:

- shared delivery-runtime language helper:
  - `interface/lib/deliveryRuntimeLanguage.ts`
- shared test fixture builders:
  - `interface/__tests__/support/deliveryRuntimeFixtures.ts`
- low-risk callers now use shared language for:
  - output folder action states
  - local media dependency recovery copy
  - compact review/failure copy
  - Active Work state labels
- shared output-package path/action helper:
  - `interface/lib/outputPackageModel.ts`
- Soma and Groups now share package open/folder/resources labels
- Groups project packages expose `Open in Resources` beside file and folder actions
- generated HTML outputs now use the same `Open file` language instead of a separate raw-output path
- mocked browser package proof now follows Soma output -> Review output -> Open in Resources -> Workspace Explorer folder re-entry

Proof:

- `npm test -- outputPackageModel.test.ts OutputWorkbench.test.tsx GroupManagementPanel.project-package.test.tsx deliveryRuntimeLanguage.test.ts`
- `npm test -- outputPackageModel.test.ts deliveryRuntimeLanguage.test.ts OutputWorkbench.test.tsx GroupManagementPanel.project-package.test.tsx ResourcesPage.test.tsx WorkspaceExplorer.test.tsx`
- `uv run inv interface.typecheck`
- `uv run inv interface.e2e --project=chromium --workers=1 --server-mode=external --spec=e2e/specs/first-demo-success.spec.ts`
- `uv run inv interface.e2e --project=chromium --workers=1 --server-mode=external --spec=e2e/specs/ui-finalization-browser-package-retry.spec.ts`
- `uv run inv lifecycle.health`

Remaining before acceptance:

- keep headed browser proof opening a generated browser app/game from Soma and Groups green during release promotion
- keep live provider-backed package proof with Resources re-entry green during release promotion
- final docs/state close-out after browser proof

P0.1 acceptance must satisfy the Soma UX contract: the operator should receive a retained deliverable package, not just a chat response or raw source dump.

## P0.2 And P0.3 Proof

P0.2 brings the services up cleanly and proves runtime readiness. P0.3 proves the output package in a real headed browser. These are separate because a healthy stack is necessary but not sufficient; the user behavior must also work. Current source proof has passed both the headless and headed live package path; release promotion should keep this gate green rather than replacing it with mocked evidence.

## P0.5 Current Implementation

Resources now labels the tool/search surface as `Capabilities` while keeping `tab=tools` deep links stable. The default view groups capability state into can-use, needs-repair, and can-request/add posture; MCP server cards and binding details remain available only through Inspect. Manifest normalization maps `output_schema_ref` into operator-readable output labels instead of exposing raw `tool_refs` as outputs.

Proof:

- `npm test -- MCPToolRegistry.test.tsx ResourcesPage.test.tsx useCortexStore.resource-registry.test.ts SettingsPage.test.tsx WorkspaceExplorer.test.tsx`
- `npx tsc --noEmit`
- `uv run inv interface.e2e --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/mcp-connected-tools.spec.ts`
- `uv run inv interface.e2e --headed --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/mcp-connected-tools.spec.ts`

## UX Architecture Rules

Default routes must answer three questions in the first viewport:

1. What can I do here?
2. What needs my attention?
3. Where is the output or proof?

Complex surfaces should use:

- bounded list-detail layouts
- tabs inside selected detail
- receipts before logs
- output packages before raw tool output
- one primary action per row/card
- raw payloads, MCP internals, bus subjects, and topology behind Inspect

Avoid:

- page-length stacked control panels
- raw HTML/code as the default generated-output view
- equal-weight button clusters
- exposing architecture terminology before user value
- failure text that does not name what remains trusted and what happens next

## Capability And MCP Posture

Capabilities should answer:

```text
What can Soma use?
Is it available?
What risk/approval applies?
What output can it produce?
What should the operator do if it is degraded?
```

MCP servers, providers, and custom tools remain important, but they are not the default operator concept. They are implementation detail inside a capability catalog.

## Recovery And Confidence Posture

Every degraded state should expose:

- what failed
- what remains trusted
- what proof is invalid
- what can continue safely
- what requires retry
- what requires operator attention
- what uncertainty now exists

Confidence provenance should be prepared through schemas and receipts, but not overbuilt before the MVP path works.

## Release Acceptance Standard

A slice is not accepted unless it reports:

- operator problem solved
- files changed
- runtime/API contract touched or reviewed
- docs changed
- tests run
- headed/browser proof when visible UI changed
- remaining trust/recovery gaps

The decisive release question is:

```text
Can the user complete the Trusted Outcome Journey and trust the result later?
```

Not:

```text
Did we finish the subsystem?
```

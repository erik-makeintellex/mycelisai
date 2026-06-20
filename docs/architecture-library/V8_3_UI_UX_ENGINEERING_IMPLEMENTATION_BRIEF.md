# V8.3 UI/UX Engineering Implementation Brief
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-06-20
> Purpose: Layer the threaded Soma workspace, rich action cards, action library, outcome vault, and zero-friction configuration mandate onto the current Trusted Outcome Journey release plan.

## Source Of Authority

This brief extends the V8.3 release architecture. It does not replace:

- [V8.3 Release Architecture Delivery Brief](V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.md)
- [V8.3 Operational Embodiment PRD](V8_3_OPERATIONAL_EMBODIMENT_PRD.md)
- [V8.3 Soma User Experience Contract](V8_3_SOMA_USER_EXPERIENCE_CONTRACT.md)
- [V8.3 MVP UI Runtime Delivery Plan](V8_3_MVP_UI_RUNTIME_DELIVERY_PLAN.md)

All implementation is still judged through the Trusted Outcome Journey:

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

The new UI mandate changes the product expression: the default surface should feel like a threaded Soma workspace with actionable cards, not a terminal, log console, or infrastructure dashboard.

## Product Translation

The phrase "abandon dashboard" means:

```text
Dashboard becomes the Soma threaded workspace.
```

It does not mean deleting the current proof path. Current release proof already depends on Dashboard, Soma, Resources, Groups, and Runs as coordinated re-entry points. The next UI work should compress those surfaces around one conversation-led outcome lane, with advanced detail behind Inspect.

Default operator shape:

```text
I ask Soma.
Soma shows understanding, approval, work state, outputs, trust, and recovery as cards.
I open or revisit the outcome from the same thread or its vault entry.
```

Advanced operator shape:

```text
Inspect opens runtime detail, traces, MCP/tool bindings, events, raw payloads, and deployment state.
```

## Epic 1 - Conversational Engine

Goal: make `Talk to Soma` the persistent command surface and outcome thread.

Current foundation:

- `interface/components/soma/ExpressionFrame.tsx`
- `interface/components/soma/ExecutionSummaryCard.tsx`
- `interface/components/soma/ExecutionSummaryReceipt.tsx`
- `interface/components/soma/OutputWorkbench.tsx`
- `interface/components/workspace/MissionControlChat.tsx`
- `interface/store/useCortexStore.ts`

Required direction:

1. Render standard user text, direct answers, proposals, execution started, output ready, proof, recovery, and blocked states as typed UI frames.
2. Treat structured runtime payloads as component selection input. Text is fallback content, not the primary UI contract for meaningful work.
3. Add an interception layer that converts transport, provider, MCP, filesystem, timeout, and validation failures into an `OperationalAlertFrame`.
4. Keep raw stack traces, raw JSON, NATS subjects, and MCP schema detail behind Inspect.
5. Preserve instant user feedback: after click or submit, the thread shows `Execution started` before background work completes.

Implementation notes:

- First component target: normalize existing blocked/degraded chat states into a shared operational alert frame.
- First state target: add a typed thread-event adapter that maps API/NATS payload kind to frame kind.
- First test target: prove raw backend errors are suppressed in the thread while recovery choices remain visible.

## Epic 2 - Action Library And Button Studio

Goal: make repeated work start quickly without forcing the operator to rewrite prompts.

Current foundation:

- Soma starter suggestions and workflow prompts
- `SomaActionShelf` as the first pinned-action dashboard layer
- proposal/confirm execution flow
- persisted organization/team work contexts
- Resources capability state

Required direction:

1. Add a top action shelf above the Soma thread with the user's top 3 to 5 pinned actions.
2. Clicking a pinned action renders an immediate `Execution started` card and dispatches a governed backend action asynchronously.
3. Add a Button Studio wizard for creating saved actions from an intent, desired output format, required checklist, and approval posture.
4. Add an Action Library panel for searching saved actions and toggling pinned state.
5. Saved actions must compile to structured payloads, not prompt-only text.

Implementation notes:

- First component target: `SomaActionShelf` using deterministic local starter actions before persistence exists.
- First API target: saved action list/create/update endpoints with pinned ordering.
- First runtime target: saved action execution emits a normal ExecutionContract and run receipt.
- First test target: clicking a shelf action immediately changes the thread before backend completion.

## Epic 3 - Outcomes And Vault Panel

Goal: make outputs durable, associated with outcomes, and easy to revisit.

Current foundation:

- `SomaOutcomeVaultPanel` as the first persistent dashboard vault rail
- `OutputWorkbench`
- `OutputWorkbenchDigest`
- `OutputAccessActions`
- Resources Output Files
- Groups selected-output and workflow-log tabs
- Run receipts and proof artifacts

Required direction:

1. Right-side work/vault panels should survive browser refreshes and chat scroll.
2. Active work should show clean background operation indicators instead of logs by default.
3. Finalized assets should be referenced by the outcome folder and visible in the Vault/Outputs view.
4. User-facing final outputs and team-intermediate outputs must remain separated.
5. Team-intermediate outputs may be included through an explicit control such as `Include team source files`.

Implementation notes:

- First component target: turn the current Work panel output tab into an `Outcome Vault` language layer.
- First API target: ensure output refs carry `outcome_id`, `group_id`, `source_role`, and user-facing versus source/intermediate classification.
- First test target: from a Soma output card, revisit the same file through Resources and Groups without exposing source files by default.

## Epic 4 - Zero-Friction System Configuration

Goal: make setup feel like permissions and capabilities, not topology.

Current foundation:

- Resources capability catalog
- MCP layered configuration for `all`, `group`, and `host` scopes
- Settings access and admin-gated views
- System status and deployment trust surfaces

Required direction:

1. Default settings must use capability permission language such as `Connect file storage`, `Connect search`, `Connect accounting software`, or `Enable media generation`.
2. Terms such as node topology, vector indexing, MCP server auth, raw tool schema, and NATS subject should stay behind Inspect or admin docs.
3. Users should define outcome environments such as `Marketing Department`; Soma should allocate the smallest useful specialist/team structure behind the scenes.
4. The UI should summarize assigned support as an operational outcome context, not ask users to design agent topology.

Implementation notes:

- First component target: Settings and Resources cards that show capability, availability, risk, approval, and repair path.
- First API target: expose environment definitions separately from team implementation detail.
- First test target: non-advanced settings route contains no topology/MCP-auth vocabulary in default copy.

## Middleware And Event Binding

The desired runtime shape is a WebSocket bridge from event bus to frontend state. The current UI may continue through API polling or existing streams while the bridge is implemented, but component contracts should be written as if events arrive as typed payloads.

Target event envelope:

```json
{
  "event_id": "evt_...",
  "payload_kind": "output_ready",
  "source_kind": "web_api",
  "source_channel": "soma_thread",
  "outcome_id": "out_...",
  "run_id": "run_...",
  "group_id": "grp_...",
  "trust_state": "verified",
  "recovery_state": "none",
  "timestamp": "2026-06-20T00:00:00Z"
}
```

Minimum payload kinds:

| Payload Kind | UI Frame |
| --- | --- |
| `user_message` | User text bubble |
| `understanding` | Soma understanding card |
| `proposal_ready` | Approval card |
| `execution_started` | Active operation card |
| `output_ready` | Output package card |
| `proof_ready` | Run receipt/proof card |
| `recovery_required` | Recovery card |
| `operational_alert` | Operational alert card |

The UI must not wait for final backend success to acknowledge user intent. It should acknowledge click/submit immediately, then reconcile with final proof, output, or recovery state.

## Design System Target

Target aesthetic: quiet, high-trust, business-grade, and low-noise. The working name is "stealth wealth", but the implementation requirement is clarity, not decoration.

Target tokens:

| Token | Target |
| --- | --- |
| Base background | `#F4F5F7` |
| Cards and panels | `#FFFFFF` |
| Action/user input | Blue |
| Verified/running smoothly | Green |
| Blocked/needs executive decision | Orange or red |
| Code/monospace | Inspect only |

Migration rule: do not flip the whole product theme in one unproofed pass. Apply the palette first to the threaded workspace and new components behind focused visual proof, then migrate adjacent pages.

## P0 Integration Order

| Order | Slice | Journey Step | Why Now |
| --- | --- | --- | --- |
| 1 | Operational alert frame | Recover, Trust | Stops raw errors and makes degraded work actionable in Soma |
| 2 | Soma action shelf | Ask, Execute | Gives instant-start repeated work without prompt friction |
| 3 | Typed thread-event adapter | Understand, Deliver, Trust | Lets structured runtime payloads mount the right card |
| 4 | Outcome Vault language layer | Deliver, Revisit | Makes output ownership clearer without another page |
| 5 | Saved action persistence | Ask, Execute, Revisit | Turns quick actions into durable product objects |
| 6 | Capability-permission settings cards | Trust, Recover | Converts setup from topology to user permissions |
| 7 | Event/WebSocket bridge | Execute, Trust | Provides sub-second reconciliation for active work |
| 8 | Light design-token migration | All | Aligns final visual language once behavior is stable |

## Acceptance Gates

Each UI slice must prove:

- Soma input remains reachable.
- Primary action is visible in the first viewport.
- Whole-page scrolling does not hide the active outcome.
- Typed cards explain outcome, status, proof, and recovery without infrastructure vocabulary.
- Raw errors and raw payloads are hidden by default.
- A user can revisit final outputs through Soma, Resources, Groups, or Runs as appropriate.
- Headless and headed browser proof cover the changed interaction.
- User docs, architecture docs, docs manifest, and state are updated when meaning changes.

Do not accept a UI slice that makes the user learn agents, MCP, NATS, workflows, runs, topology, or infrastructure before they can complete the Trusted Outcome Journey.

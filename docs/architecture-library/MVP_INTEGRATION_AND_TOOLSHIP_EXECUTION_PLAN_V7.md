# MVP Integration And Toolship Execution Plan V7

> Status: Canonical execution plan
> Last Updated: 2026-03-10
> Purpose: Drive MVP from internal-test posture to externally usable product by hardening AI interaction, internal toolship, and service-connection contracts.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [MVP Release Strike Team Plan V7](MVP_RELEASE_STRIKE_TEAM_PLAN_V7.md)

## Status Markers

Use only:
- `REQUIRED`
- `NEXT`
- `ACTIVE`
- `IN_REVIEW`
- `COMPLETE`
- `BLOCKED`

## MVP Integration Scope

This plan defines the minimum integration posture required to support real users beyond a test team.

MVP-critical integration surfaces:
1. AI interaction contract (`Soma` direct-first; council on explicit planning/architecture/delivery paths)
2. Internal toolship for real work (`read_file`, `write_file`, `local_command`, `delegate_task`, governed signaling)
3. Service connections (`NATS`, `Postgres`, MCP/OpenAPI/host adapters) with stable health, timeout, and fallback behavior
4. Operator-visible lineage and governance (proposal/confirm/run linkage, signal metadata, failure recovery)

## Product Contract For MVP

Every user intent must terminate in exactly one:
- `answer`
- `proposal`
- `execution_result`
- `blocker`

For any mutating intent:
1. proposal details are explicit
2. confirmation path is explicit
3. execution side effects are auditable
4. recovery actions are explicit on failure

## Integration Architecture Rules

1. All governed product signals must carry metadata:
   - `run_id` (when execution-linked)
   - `team_id` (when team-scoped)
   - `agent_id` (when agent-scoped)
   - `source_kind`
   - `source_channel`
   - `payload_kind`
   - `timestamp`
2. Internal tool calls are treated as first-class execution operations, not hidden helper behavior.
3. Service adapters must normalize errors into operator-usable blockers.
4. High-volume telemetry channels are not reused as operator status/result channels.
5. Runtime publishes must prefer canonical subject constants and canonical subject families.

## Execution Lanes

1. `prime-architect` (`ACTIVE`): architecture contract governance, acceptance gate ownership
2. `prime-development` (`ACTIVE`): core runtime + toolship contract implementation
3. `agui-design-architect` (`ACTIVE`): UI interaction clarity + failure/recovery UX
4. `council-sentry` (`ACTIVE`): test-depth ownership and release-gate verification
5. `admin-core` (`ACTIVE`): lifecycle/startup reliability and operator runbook readiness

## Phase Plan

## Phase A - Tool Invocation Contract Hardening

Status:
- `ACTIVE`

Goal:
- ensure internal toolship and team command/result paths emit canonical metadata envelopes.

Primary delivery:
1. internal tool publish/delegation paths wrap canonical signal metadata
2. command payload normalization keeps agent UX readable while preserving governance metadata
3. tests cover metadata presence and routing compatibility

Proof:
- `cd core && go test ./internal/swarm ./pkg/protocol -count=1`
- `uv run inv logging.check-topics`

## Phase B - File Toolship Reliability And Guardrails

Status:
- `NEXT`

Goal:
- make file toolship safe and production-traceable for non-test users.

Primary delivery:
1. workspace sandbox + path normalization + size guardrails remain enforced
2. mutation actions generate governed proposal/confirm paths when risk threshold requires it
3. actionable operator blocker messages for filesystem/service unavailability

Proof:
- focused `internal/swarm` file-tool tests
- Workspace UI tests for proposal/blocker rendering on file mutations

## Phase C - Service Connection Governance

Status:
- `NEXT`

Goal:
- stabilize service connectors and normalize connector failures.

Primary delivery:
1. connector health classification (`online|degraded|offline`) is consistent across runtime and UI
2. timeout/retry/circuit-breaker behavior is explicit and test-proven
3. MCP and third-party adapter failures map to one shared blocker model

Proof:
- `uv run inv lifecycle.health`
- `uv run inv core.test`
- focused Playwright degraded/recovery proof

## Phase D - External-User MVP Gate

Status:
- `REQUIRED`

Goal:
- move from test-team confidence to broader-user confidence.

Primary delivery:
1. startup/restart reliability from clean machine
2. core journey pass-rate and blocker quality thresholds
3. operator docs and in-app docs are aligned with actual behavior

Proof:
- `uv run inv ci.baseline`
- `uv run inv interface.e2e` (Chromium/Firefox gating; WebKit tracked with stabilization plan)
- pilot usage metrics and blocker-rate review

## Immediate Execution Slice (Started 2026-03-10)

Status:
- `ACTIVE`

Slice objective:
- implement governed metadata wrapping for internal toolship command publishing without regressing agent execution usability.

Scoped files:
- `core/internal/swarm/internal_tools.go`
- `core/internal/swarm/internal_tools_signals.go`
- `core/internal/swarm/signal_channel_checkpoint.go`
- `core/internal/swarm/team.go`
- `core/internal/swarm/internal_tools_signal_test.go`
- `core/internal/swarm/agent_signal_bridge_test.go`
- `core/internal/swarm/team_test.go`
- `core/pkg/protocol/envelopes.go` (if metadata wrapper extension is required)

Acceptance criteria:
1. `delegate_task` and canonical `publish_signal` command/result/status flows emit metadata envelopes
2. MCP tool execution emits canonical team-bus status/result envelopes (no direct-only execution path)
3. team trigger forwarding keeps agent input coherent (no forced JSON-only prompt regression)
4. private channel/file relay path supports referential channel payloads (`privacy_mode=reference`) without exposing full private payloads on public channels
5. relaunch path can recover latest channel output using `read_signals` `latest_only=true` + `channel_key`
6. tests prove metadata + compatibility behavior together

## Deep Testing Matrix For This Plan

Mandatory:
1. `cd core && go test ./internal/swarm ./pkg/protocol -count=1`
2. `uv run inv interface.test`
3. `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v7-operational-ux.spec.ts`
4. `uv run inv ci.baseline`

When service-connection logic changes:
1. `uv run inv lifecycle.down`
2. `uv run inv lifecycle.up --frontend`
3. `uv run inv lifecycle.health`
4. `uv run inv db.status`

## MVP Exit Criteria (Beyond Test Team)

MVP is `IN_REVIEW` only when all are true:
1. startup and health checks are reproducible on fresh local environment
2. AI/toolship core journeys are stable with governed outcomes (`answer|proposal|execution_result|blocker`)
3. toolship actions are auditable via run/team/agent metadata
4. docs, runbooks, and in-app docs match real command contracts and failure behavior
5. gate evidence is current in `V7_DEV_STATE.md`

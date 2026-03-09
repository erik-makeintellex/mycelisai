# Next Execution Slices V7

> Status: Canonical working queue
> Last Updated: 2026-03-09
> Purpose: Translate the modular architecture library into the next concrete delivery slices with explicit development and testing references.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)

## Feature Status Markers

Use only:
- `REQUIRED`
- `NEXT`
- `ACTIVE`
- `IN_REVIEW`
- `COMPLETE`
- `BLOCKED`

## Commitment Logic For This Queue

Each slice below should be committed only when all of the following are true:
- the operator-visible outcome is explicit
- the backend/runtime effect is explicit
- the required proof commands have been run
- the canonical docs for that slice are updated in the same change
- the rollback scope is narrow and documented

This queue is ordered by product evocation first:
1. operator-facing execution clarity
2. operator-facing error and recovery clarity
3. internal coordination that unlocks governed execution
4. structural cleanup behind behavior contracts
5. manifest-lifecycle unification for later phases

## Slice 1: Launch Crew And Workflow Onboarding

Status:
- `COMPLETE`

Objective:
- make Launch Crew and onboarding execution-facing instead of planning-facing.

Scoped outcome:
- every onboarding path must end in `answer`, `proposal`, `execution_result`, or `blocker`

Scoped files:
- `interface/components/workspace/LaunchCrewModal.tsx`
- `interface/components/dashboard/MissionControl.tsx`
- `interface/components/dashboard/MissionControlChat.tsx`
- `interface/store/useCortexStore.ts`
- relevant route/API integration code

Development docs:
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Workflow Composer Delivery V7](../architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)

Testing docs:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Testing](../TESTING.md)

Required proof:
- component tests for terminal-state rendering
- integration tests for frontend-triggered backend transaction
- product-flow tests proving onboarding reaches a real outcome instead of a planning-only return
- live-backend browser proof for at least one onboarding or Workspace-adjacent path when the flow depends on the real UI proxy and Core API

Acceptance criteria:
- onboarding no longer ends in planning-only language
- backend/API effect is visible and test-proven
- failure states are actionable blockers

Rollback:
- revert UI/state changes while preserving already-enforced Workspace chat contract

Completion evidence:
- terminal-state component and store coverage now includes proposal, answer, blocker, execution result with `run_id`, and execution result without `run_id`
- browser coverage proves proposal outcome, blocker/recovery outcome, and live-backend confirm-action execution through the real UI proxy
- onboarding no longer relies on planning-only modal states

## Slice 2: P1 Logging, Error Handling, And Execution Feedback

Status:
- `ACTIVE`

Objective:
- make operator-visible logs, blockers, retries, and degraded-state feedback consistent across the highest-risk execution paths.

Scoped outcome:
- Workspace, council, lifecycle, and bootstrap failures must explain what happened and what to do next instead of collapsing into generic noise.

Scoped files:
- `docs/logging.md`
- `core/internal/swarm/agent.go`
- `core/internal/server/comms.go`
- `core/internal/mcp/service.go`
- `ops/lifecycle.py`
- `interface/components/dashboard/CouncilCallErrorCard.tsx`
- `interface/components/dashboard/DegradedModeBanner.tsx`
- `interface/components/dashboard/MissionControlChat.tsx`
- `interface/store/useCortexStore.ts`
- focused tests around the changed paths

Development docs:
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Operations](../architecture/OPERATIONS.md)

Testing docs:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Testing](../TESTING.md)

Required proof:
- focused UI tests for blocker/recovery behavior
- focused runtime/task tests for startup, teardown, and degraded messaging
- logging/schema or topic checks where runtime event shape changes

Acceptance criteria:
- operator-facing failure states are actionable and specific
- no generic planning-only or opaque 500-style outcomes remain in the scoped paths
- docs and tests describe the same failure/recovery contract

Rollback:
- revert only the scoped feedback/logging slice while preserving previously proven execution contracts

## Slice 3: Prime-Development Reply Reliability

Status:
- `NEXT`

Objective:
- complete central-architect team coordination by making `prime-development` reply reliably over the canonical team NATS lanes.

Why this matters:
- the bus contract is already standardized
- `prime-architect` and `agui-design-architect` reply
- `prime-development` is the remaining coordination gap
- this matters after the operator-facing onboarding path because it supports governed multi-team execution instead of replacing it

Scoped files:
- `core/config/teams/prime-development.yaml`
- `ops/misc.py`
- `tests/test_misc_tasks.py`
- runtime/team prompt or context files as needed

Development docs:
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)
- [Soma Team + Channel Architecture V7](../architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md)

Testing docs:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Testing](../TESTING.md)

Required proof:
- task test coverage for canonical publish/collect path
- live `uv run inv team.architecture-sync` evidence showing all three standing teams reply cleanly

Acceptance criteria:
- `prime-development` replies within the sync window
- replies are plain-text and operator-readable
- no regression in architect or AGUI replies

Rollback:
- revert prompt/runtime-team changes and preserve canonical NATS lane handling

## Slice 4: P1 Hot-Path Cleanup

Status:
- `REQUIRED`

Objective:
- reduce hot-path complexity under the no-regression max-lines contract.

Scoped files:
- `core/internal/swarm/agent.go`
- `ops/lifecycle.py`
- `core/internal/swarm/internal_tools.go`
- any directly extracted helpers

Development docs:
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [Operations](../architecture/OPERATIONS.md)

Testing docs:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Testing](../TESTING.md)

Required proof:
- focused unit/integration tests around the extracted behaviors
- `uv run inv quality.max-lines --limit 350`
- no regression in lifecycle/chat/runtime behaviors covered by the changed files

Acceptance criteria:
- one or more hot-path files become smaller or stop growing under legacy caps
- operator-visible behavior is preserved
- logging/error clarity does not regress

Rollback:
- revert the extraction slice only; do not revert unrelated behavior fixes

## Slice 5: Manifest Pipeline Preparation

Status:
- `REQUIRED`

Objective:
- prepare `P2` so workflows and teams move through one manifest lifecycle with support for recurring and always-on plans.

Scoped outcome:
- explicit lifecycle for `draft -> validated -> proposed -> approved -> activated -> paused/completed/cancelled`
- explicit mode support for `one_shot`, `scheduled`, `persistent_active`, `event_driven`

Scoped files:
- manifest model and validation paths
- workflow/team activation handlers
- composer-facing backend contracts

Development docs:
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [Workflow Composer Delivery V7](../architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)

Testing docs:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Testing](../TESTING.md)

Required proof:
- validation failure coverage
- activation/approval coverage
- recurring-plan persistence behavior tests

Acceptance criteria:
- invalid manifests stop before activation
- recurring modes are explicit in model and tests
- UI-facing contracts can consume the same lifecycle model

Rollback:
- keep current runtime flows and back out only unfinished manifest-lifecycle changes

## Slice 6: Soma-First Team Expression And Module Binding

Status:
- `ACTIVE`

Objective:
- unify intent execution across internal tools, MCP tools, and third-party APIs using one module-binding contract in the operator flow.

Scoped outcome:
- Soma responses for executable intent include structured Team Expressions with explicit module bindings
- manifestation flow remains terminal-state compliant (`answer`/`proposal`/`execution_result`/`blocker`)

Scoped files:
- `docs/architecture-library/INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md`
- `docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md`
- `interface/components/dashboard/MissionControlChat.tsx`
- `interface/store/useCortexStore.ts`
- backend contracts for module-binding payload normalization

Development docs:
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Universal Action Interface V7](../architecture/UNIVERSAL_ACTION_INTERFACE_V7.md)

Testing docs:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Testing](../TESTING.md)

Required proof:
- component and store tests for Team Expression rendering/editing state
- integration tests proving module binding payload normalization across adapter kinds
- product-flow proof showing one intent path through proposal and run-linked result

Acceptance criteria:
- operator can inspect module bindings before manifest confirmation
- module adapter details do not leak into primary terminal state messaging
- execution metadata remains run-linked and auditable

Rollback:
- revert Team Expression/module-binding UX and payload additions while preserving existing proposal flow

Checkpoint (2026-03-09):
- runtime proposal payload now carries `team_expressions` with `module_bindings` in both `/api/v1/chat` and `/api/v1/council/{member}/chat` mutation paths
- UI proposal surfaces now show expression/binding counts and binding adapter context before confirmation
- store normalization now derives `teams`, `agents`, and `tools` from structured `team_expressions` when present
- focused proof executed:
  - `cd core; go test ./internal/server -run "TestInferAdapterKindFromTool|TestBuildMutationChatProposal" -count=1` -> pass
  - `cd interface; npx vitest run __tests__/store/useCortexStore.test.ts __tests__/dashboard/ProposedActionBlock.test.tsx --reporter=dot` -> pass
  - `cd interface; npx tsc --noEmit` -> pass
- remaining for `IN_REVIEW` transition:
  - attach explicit product-flow proof of proposal -> confirmation -> run-linked outcome tied to Team Expression payload
  - keep Slice 7 blocked until scheduler + chain prerequisites are accepted

## Slice 7: Created Team Workspace And Channel Inspector

Status:
- `BLOCKED`

Objective:
- give operators first-class interaction with created teams across canonical channels without requiring raw infrastructure tooling.

Blocked by:
- scheduler and recurring-plan visibility completion
- chain-view/operator lineage clarity completion

Scoped outcome:
- Team Workspace tabs (`Overview`, `Communications`, `Members`, `Manifest`, `Controls`) available for created teams
- unified communications timeline filtered by `run_id`, `team_id`, `agent_id`, `source_kind`, `payload_kind`

Scoped files:
- team workspace route/components
- backend aggregation endpoints for team communication inspection
- relevant signal normalization/store slices
- docs and in-app docs registration

Development docs:
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [Soma Team + Channel Architecture V7](../architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md)
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)

Testing docs:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Testing](../TESTING.md)

Required proof:
- integration tests for created-team command -> status/result round-trip
- UI tests for communication inspector filters and operator actions
- product-flow tests for interjection/reroute/pause-resume controls

Acceptance criteria:
- created-team execution is inspectable and steerable from product UI
- operator status channels remain distinct from high-volume telemetry
- all rendered communication entries carry required metadata fields

Rollback:
- revert team workspace surfaces while preserving existing run timeline and conversation views

## Immediate Working Order

1. Slice 1
2. Slice 2
3. Slice 3
4. Slice 4
5. Slice 5
6. Slice 6
7. Slice 7

That order now follows the architecture library more directly:
- operator-facing execution clarity first
- operator-facing failure/recovery clarity second
- internal coordination third
- structural cleanup fourth
- manifest-lifecycle unification fifth
- module-binding unification sixth
- created-team interaction seventh

# Soma Workspace QA Recovery Strike Plan V7

> Status: Historical recovery record
> Outcome: `COMPLETE`
> Closed: 2026-03-26 via `bfb3b80`, `212d5d2`, and `94b531c`

## 1. Purpose

This document turns the March 25, 2026 Soma workspace browser QA failures into a coordinated delivery plan with explicit agent ownership, file boundaries, state-model changes, and acceptance targets.

The goal is not only to patch individual bugs, but to restore a trustworthy Soma-primary operator path across:

- direct Soma answers
- governed mutation proposal generation
- proposal cancellation
- trust after successful execution
- state persistence and re-entry after successful action

## 2. Current Release Status

- Runtime routing guardrails: `COMPLETE`
- Direct Soma answer fidelity: `COMPLETE`
- Governed mutation proposal path: `COMPLETE`
- Proposal lifecycle truthfulness: `COMPLETE`
- Successful execution trust semantics: `COMPLETE`
- Organization-scoped continuity and re-entry: `COMPLETE`
- Browser QA regression proof for Soma workspace: `COMPLETE`

## 3. Observed Failure Summary

The March 25, 2026 live QA pass established the following product failures before the recovery work landed:

1. Soma chat initially booted into a `provider_disabled` state because the startup provider path still pinned Soma/admin to `local-ollama-dev` while the clean repo disabled that provider.
2. After a non-persistent runtime override restored chat transport, `/api/v1/chat` still returned `mode=answer` with empty `payload.text` for both read-only and mutation requests.
3. Mutation requests therefore never reached `proposal`, which blocked cancellation and downstream trust testing.
4. Frontend proposal and continuity behavior has additional trust gaps even after backend recovery:
   - proposal cancellation is not a real lifecycle transition
   - `execution_result` is claimed too early
   - chat continuity is global instead of organization-scoped
   - the causal strip overstates productive work after ambiguous or failed outcomes

## 4. Strike Team Composition

### 4.1 Orchestrator

- Role: `workspace_recovery_orchestrator`
- Responsibility:
  - preserve the Soma-first product contract
  - sequence the slices below in dependency order
  - enforce terminal-state truthfulness across runtime, backend, frontend, and QA
  - approve only changes that keep UI claims aligned with durable backend evidence

### 4.2 Delivery Agents

#### Agent A: Runtime Guard

- Status: `NEXT`
- Ownership:
  - `core/config/templates/v8-migration-standing-team-bridge.yaml`
  - `core/cmd/server/bootstrap_startup.go`
  - `core/internal/cognitive/router.go`
  - `core/internal/cognitive/availability.go`
  - `core/internal/swarm/provider_policy_resolve.go`
  - `core/internal/swarm/provider_policy_helpers.go`
  - `core/internal/swarm/team.go`
- Mission:
  - prevent Soma/admin/council defaults from booting on a disabled explicit provider when a healthy fallback exists
  - validate startup provider policy against runtime availability before Soma is declared ready
  - keep DB/YAML/env overlays from silently trapping chat on a dead provider

#### Agent B: Chat Contract

- Status: `NEXT`
- Ownership:
  - `core/internal/cognitive/openai.go`
  - `core/internal/swarm/agent.go`
  - `core/internal/swarm/agent_parsing.go`
  - `core/internal/server/cognitive.go`
- Mission:
  - normalize provider responses so empty or structured action outputs are not silently downgraded
  - stop `/api/v1/chat` from emitting invalid `answer` terminal states
  - preserve mutation intent strongly enough to terminate mutation requests in `proposal` or `blocker`, never empty `answer`

#### Agent C: Governance Lifecycle UI

- Status: `NEXT`
- Ownership:
  - `interface/store/useCortexStore.ts`
  - `interface/components/dashboard/MissionControlChat.tsx`
  - `interface/components/dashboard/ProposedActionBlock.tsx`
- Mission:
  - make proposal cancellation a real lifecycle transition
  - stop treating a bare confirmation as a completed execution result
  - render governance controls from proposal lifecycle state, not from a stale embedded proposal blob alone

#### Agent D: Continuity and Re-entry

- Status: `NEXT`
- Ownership:
  - `interface/store/cortexStoreUtils.ts`
  - `interface/store/useCortexStore.ts`
  - `interface/components/organizations/OrganizationContextShell.tsx`
- Mission:
  - scope workspace chat continuity by organization
  - derive the causal strip from verified terminal states
  - ensure re-entry reflects what actually happened, not what the UI hoped happened

#### Agent E: QA Gate

- Status: `NEXT`
- Ownership:
  - `interface/e2e/specs/workspace-live-backend.spec.ts`
  - new or updated Soma workspace live-backend specs
  - `tests/ui/browser_qa_plan_workspace_chat.md`
  - `docs/TESTING.md` if acceptance coverage language changes
- Mission:
  - encode the five blocked browser journeys as release-gating proof
  - fail when Soma emits empty `answer`
  - fail when mutation does not reach `proposal`
  - fail when UI claims `execution_result` without concrete execution proof
  - fail when continuity leaks across organizations or misreports failure as productive action

## 5. Delivery Order

### Slice 1: Runtime Routing Guardrails

- Owner: Agent A
- Dependency: none
- Required outputs:
  - Soma/admin default route cannot bind to a disabled provider when a healthy fallback exists
  - startup and runtime availability surfaces agree on Soma readiness

### Slice 2: Backend Terminal-State Integrity

- Owner: Agent B
- Dependency: Slice 1
- Required outputs:
  - `answer` requires readable text or artifacts
  - mutation intent reaches `proposal` or `blocker`
  - empty provider output cannot masquerade as success

### Slice 3: Proposal and Execution Trust Semantics

- Owner: Agent C
- Dependency: Slice 2
- Required outputs:
  - cancellation changes proposal lifecycle state visibly
  - confirmation does not equal execution unless the backend returns proof
  - no success-like UI without run or artifact evidence

### Slice 4: Organization-Scoped Continuity

- Owner: Agent D
- Dependency: Slice 3
- Required outputs:
  - chat persistence keyed by organization
  - causal strip mirrors verified outcomes
  - re-entry after success shows the same trustworthy state the operator left

### Slice 5: Browser Gate Proof

- Owner: Agent E
- Dependency: Slices 1-4
- Required outputs:
  - automated and manual QA proof for all five critical journeys
  - release recommendation flips only after live-backed Soma workspace proof passes

## 6. Required State Changes

The current product contract needs explicit state-model tightening.

### 6.1 Runtime Availability State

Required change:

- make startup provider-policy resolution and execution availability converge on one truth source

Target state:

- `chat_available`
- `chat_blocked_provider_disabled`
- `chat_blocked_provider_unreachable`
- `chat_blocked_no_model`

Rule:

- the org home must not say `Soma ready` unless the chat path for Soma/admin is actually executable

### 6.2 Chat Terminal State

Required change:

- forbid invalid `answer` states with empty payloads

Allowed terminal states:

- `answer`
- `proposal`
- `execution_result`
- `blocker`

Rule:

- `answer` requires non-empty readable text or artifacts
- `proposal` requires confirm metadata and clear operator-review copy
- `execution_result` requires durable execution proof
- `blocker` is mandatory when the system cannot produce a truthful answer or proposal

### 6.3 Proposal Lifecycle State

Required change:

- move from implicit proposal truth to explicit lifecycle truth

Target lifecycle:

- `active`
- `cancelled`
- `confirmed_pending_execution`
- `executed`
- `failed`

Rule:

- UI controls must render from lifecycle state, not just from raw historical proposal presence

### 6.4 Continuity Scope

Required change:

- move persistence from one global workspace chat key to organization-scoped continuity

Target relation:

- `organization_id -> chat thread`
- `organization_id -> last verified causal outcome`

Rule:

- re-entering one AI Organization must never inherit another organization's proposal, execution, or fallback state

## 7. Required Relational Changes

### 7.1 Provider Policy to Runtime Availability

Current issue:

- provider policy can bypass healthy profile fallback

Required relation:

- explicit provider resolution must still validate against runtime availability before becoming authoritative for Soma/admin defaults

### 7.2 Backend Envelope to Frontend Claim

Current issue:

- backend can emit `answer` with no readable output, and frontend can emit success-like execution state with no proof

Required relation:

- frontend claims must be derivable from backend evidence, not optimistic local inference

### 7.3 Proposal Record to Chat Message

Current issue:

- proposal visibility is tied to stale chat content while actionable state lives elsewhere

Required relation:

- each rendered proposal must have a lifecycle-bearing identity that both the store and the UI can agree on

### 7.4 Causal Strip to Verified Outcome

Current issue:

- the causal strip currently treats any non-empty conversation content as productive guidance

Required relation:

- the strip must map from the last verified terminal state and evidence object, not from generic message presence

## 8. Agent-to-Agent Expression Contract

Each delivery agent must express outputs to the next agent using the same recovery language.

### Agent A -> Agent B

- provide the resolved runtime rule for Soma/admin provider selection
- identify what explicit provider values may still appear in `InferRequest`
- document fallback behavior when a provider is disabled

### Agent B -> Agent C

- provide the exact backend envelope rules for `answer`, `proposal`, `execution_result`, and `blocker`
- define the minimum proof object required for UI `execution_result`
- define how mutation intent is surfaced when execution has not yet happened

### Agent C -> Agent D

- provide the proposal lifecycle state model
- provide the UI-visible meaning of `cancelled`, `confirmed_pending_execution`, `executed`, and `failed`

### Agent D -> Agent E

- provide the continuity contract:
  - what persists
  - when it persists
  - how it scopes by organization
  - what the causal strip may claim on re-entry

## 9. Acceptance Targets

The strike team is not done until all of the following are true:

1. A simple Soma question returns a readable direct answer and never falls back to empty `answer`.
2. A mutation request in Soma chat terminates in `proposal`, not empty `answer`.
3. Cancelling a proposal updates visible state to `cancelled` and removes executable affordances.
4. Confirming a proposal does not show `execution_result` until the backend returns concrete execution evidence.
5. Re-entering the same AI Organization after a successful action restores the same verified success state.
6. Re-entering a different AI Organization does not inherit another org's chat/proposal/execution history.
7. The org home readiness copy matches actual Soma chat availability.

## 10. Release Rule

Release remains `BLOCKED` until all five QA blocker journeys pass in a live-backed browser run on the clean repo without process-local overrides that mask a broken default startup path.

## 11. Resolution Summary

The recovery work captured by this strike plan is now integrated and validated:

1. Soma runtime/provider routing now converges on executable availability truth.
2. Empty or structured provider outputs no longer masquerade as successful `answer` states.
3. Mutation requests terminate in governed proposal flow instead of bypassing into ambiguous answer-mode output.
4. Proposal cancellation, confirmation, and verified execution now render truthful lifecycle state in the UI.
5. Organization-scoped continuity and browser QA coverage now reflect the repaired contract.

This document remains as a historical strike-plan reference rather than an active release blocker.

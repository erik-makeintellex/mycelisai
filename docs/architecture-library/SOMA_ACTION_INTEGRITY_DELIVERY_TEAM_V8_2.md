# Soma Action Integrity Delivery Team V8.2

> Status: Historical delivery record
> Outcome: `COMPLETE`
> Closed: 2026-03-26 via `212d5d2` and `94b531c`

## Purpose

This document instantiates the focused delivery team for the March 25, 2026 release-blocking governance failures in the Soma action path.

The team goal is to restore operator trust so:

- mutation requests always route into governed proposal mode
- no mutation-capable tool executes before confirmation
- confirmed execution always produces durable proof (`run_id`, verified state, retrievable record)

## Release Status

- Governed mutation integrity: `COMPLETE`
- Mixed-thread mutation routing: `COMPLETE`
- Durable execution proof: `COMPLETE`
- Truthful UI lifecycle states: `COMPLETE`
- Independent QA revalidation: `COMPLETE`

## Shared Invariants

### Invariant 1

Proposal phase is planning-only.
No mutation side effects are allowed.

### Invariant 2

Mutation requests must enter governed proposal mode every time, regardless of prior chat context.

### Invariant 3

Confirmation must produce durable execution proof.

### Invariant 4

UI must never imply verified execution when proof does not exist.

## Live Failure Baseline

The March 25, 2026 live/browser proof established these release blockers before the repair work landed:

1. Mixed-thread mutation downgrade still exists: a normal answer followed by a file-creation ask can fall back to answer mode instead of proposal mode.
2. Fresh mutation proposal generation is not side-effect free: a request can return `mode=proposal` while still executing `write_file` before confirm.
3. Confirmation returns `confirmed: true` with `run_id: null`, leaving the UI in an honest but incomplete `confirmed_pending_execution` state.

## Delivery Team

### Architecture Lead

- Status: `ACTIVE`
- Owner: orchestrator
- Responsibilities:
  - hold the invariants above
  - assign bounded slices and handoff contracts
  - reject any integration that weakens governance
  - approve final integration only after QA proof
- Primary outputs:
  - execution plan
  - role assignments
  - integration checklist
  - final acceptance decision

### Runtime Governance Engineer

- Status: `ACTIVE`
- Responsibility:
  - own the proposal-vs-execution boundary
  - make proposal generation side-effect free
- File ownership:
  - `core/internal/swarm/**`
  - directly related runtime/tool execution tests under `core/internal/swarm/**`
- Must prove:
  - no file writes before confirmation
  - no external writes before confirmation
  - no hidden side effects before confirmation
  - proposal payloads contain intended actions only

### Intent / Routing Engineer

- Status: `ACTIVE`
- Responsibility:
  - own mutation classification and routing stability
- File ownership:
  - `core/internal/server/cognitive.go`
  - `core/internal/cognitive/openai.go`
  - related tests under `core/internal/server/**` and `core/internal/cognitive/**`
- Must prove:
  - mutation detection is stable across thread history
  - mutating asks trigger proposal mode even after prior answer-mode chat
  - phrasing variants remain safe

### Execution Proof Engineer

- Status: `ACTIVE`
- Responsibility:
  - own post-confirm execution truth and durable proof
- File ownership:
  - `core/internal/server/templates.go`
  - directly related run/proof/event surfaces
  - related tests in the same slice
- Must prove:
  - non-null `run_id`
  - verified execution state
  - retrievable execution record
  - reload-safe persistence contract for the UI

### UI / Product Reliability Engineer

- Status: `ACTIVE`
- Responsibility:
  - own truthful proposal / confirm / execution operator states
- File ownership:
  - `interface/store/**`
  - `interface/components/**`
  - `interface/__tests__/**`
  - targeted specs under `interface/e2e/specs/**`
- Must prove:
  - no misleading success state
  - proposal / cancel / confirm flows are understandable
  - reload preserves truthful lifecycle state
  - browser coverage exists for proposal, cancel, confirmed-awaiting-proof, executed, and failed

### QA / Verification Engineer

- Status: `REQUIRED`
- Responsibility:
  - independently verify the restored contract after integration
- Must prove:
  - the previous blockers are actually closed
  - evidence is reproducible
  - no hidden gaps remain

## Collaboration Pattern

### Step 1

Architecture Lead defines bounded ownership and exact handoff contracts.

### Step 2

Parallel work:

- Runtime Governance Engineer removes pre-confirm mutation
- Intent / Routing Engineer fixes mixed-thread mutation routing
- Execution Proof Engineer restores durable proof
- UI / Product Reliability Engineer aligns operator-state truth

### Step 3

Architecture Lead integrates the slices and checks for contract contradiction.

### Step 4

QA / Verification Engineer reruns the blocked live scenarios independently.

## Required Scenarios To Close

### Scenario A

Fresh org, normal direct question, direct answer works.

### Scenario B

Normal direct answer followed by file-creation ask still produces proposal mode.

### Scenario C

Fresh file-creation ask produces proposal mode and the target file does not exist before confirm.

### Scenario D

Cancel proposal, no side effect, truthful state persists.

### Scenario E

Confirm proposal, real `run_id`, verified execution state, file exists only after confirm, state persists across reload.

## Integration Checklist

- mixed-thread mutation downgrade is closed
- proposal generation is side-effect free
- confirm returns durable proof
- UI state names and copy match runtime truth
- reload/re-entry reflects the same verified lifecycle state
- no fix weakens governance to make the UI look green

## Validation Gate

Required before completion:

- `cd core && go test ./... -count=1 -p 1`
- `uv run inv interface.test`
- `uv run inv interface.typecheck`
- `uv run inv interface.e2e`
- rerun the blocked browser scenarios:
  - mixed-chat mutation downgrade
  - pre-confirm file creation
  - null `run_id` after confirm
- `uv run pytest tests/test_docs_links.py -q`

## Final Deliverable

Architecture Lead returns:

1. team members / roles evoked
2. work completed by each role
3. integration summary
4. validation results
5. whether blockers are fully closed
6. release impact:
   - still blocked
   - ready with minor fixes
   - blocker resolved

## Resolution Summary

The blockers captured here are now resolved in the committed branch state:

- mutation requests route into governed proposal mode consistently
- proposal planning is side-effect free for mutation-capable tools
- confirmation returns durable proof and truthful lifecycle state
- UI and browser coverage now reflect `active`, `cancelled`, `confirmed_pending_execution`, `executed`, and `failed` honestly

This document remains as a historical delivery/team record rather than live blocker truth.

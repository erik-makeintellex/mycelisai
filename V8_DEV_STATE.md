# Mycelis V8 - Development State

> Updated: 2026-03-12
> Canonical state file for active V8 grading and delivery tracking
> References: `README.md`, `mycelis_vnext_implementation_tracker.md`, `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `docs/architecture-library/TASK_001_VNEXT_ARCHITECTURE_ALIGNMENT_ARTIFACT.md`, `V7_DEV_STATE.md` (legacy migration input)

---

## Purpose

`V8_DEV_STATE.md` is now the authoritative global state file for active V8 grading.

Use this file to track:
- current V8 delivery status
- accepted marker transitions
- gate evidence
- blockers and dependencies
- immediate next actions

Use `V7_DEV_STATE.md` only as a migration input and historical checkpoint source while V7 planning artifacts are being migrated.

## Feature Status Legend

- `REQUIRED`: must exist for target delivery or gate pass
- `NEXT`: highest-priority upcoming slice
- `ACTIVE`: currently in development
- `IN_REVIEW`: implemented and awaiting validation/review
- `COMPLETE`: delivered and accepted
- `BLOCKED`: cannot advance until a dependency or defect is resolved

## V8 Delivery Program Snapshot

```text
V8-0  Migration baseline and architecture alignment                 [ACTIVE]
V8-1  Canonical contract definition                                [COMPLETE]
V8-2  Config and bootstrap model planning                          [NEXT]
V8-3  Backend primitive refactor                                   [REQUIRED]
V8-4  Frontend/operator refactor                                   [REQUIRED]
V8-5  Documentation and naming migration                           [ACTIVE]
V8-6  Verification and release hardening                           [REQUIRED]
```

## Current V8 Grading Baseline

### 1. Architecture alignment and migration inventory

Status:
1. `IN_REVIEW` Task 001 alignment artifact now exists and is linked from the VNext tracker.
2. `COMPLETE` current fixed-organism assumptions vs V8 target model have been mapped.
3. `COMPLETE` backend/frontend/docs hotspots have been identified for the first migration wave.

Primary references:
- `mycelis_vnext_implementation_tracker.md`
- `docs/architecture-library/TASK_001_VNEXT_ARCHITECTURE_ALIGNMENT_ARTIFACT.md`
- `docs/architecture-library/TASK_001_STEP_1_ALIGNMENT_AUDIT_NOTE.md`

Key grading conclusion:
- Mycelis already has the execution substrate needed for V8, but still exposes one hardcoded Soma/Council organizational expression as if it were the platform itself.

### 2. Provider and service architecture

Status:
1. `ACTIVE` V8 explicitly requires support for local, hosted, and hybrid model-provider infrastructure.
2. `COMPLETE` provider abstraction is confirmed as an existing core subsystem, not a plugin afterthought.
3. `NEXT` move provider-policy grading from global/default posture to inception-, council-, and role-aware posture.

Required V8 provider support:
- local/open-source: Ollama, LM Studio, vLLM, custom OpenAI-compatible endpoints
- hosted/commercial: OpenAI, Anthropic, Gemini, other OpenAI-compatible providers

Required routing posture:
- role-level routing
- council-level routing
- inception-level policy boundaries
- local-first and hybrid deployment support
- explicit data-boundary and approval posture

### 3. State and documentation migration

Status:
1. `ACTIVE` canonical grading is moving from `V7_DEV_STATE.md` to `V8_DEV_STATE.md`.
2. `ACTIVE` active non-archive docs are being updated to reference the V8 state file.
3. `REQUIRED` V7 architecture-library documents remain authoritative migration inputs until V8 replacements exist.

State-file rules:
1. all new status transitions, blocker updates, and gate evidence belong in `V8_DEV_STATE.md`
2. `V7_DEV_STATE.md` remains readable for historical migration evidence, but should not be the primary grading target for new work
3. active docs should describe V7 documents as migration inputs and V8 state as the live delivery scoreboard

### 4. Runtime and UI migration posture

Status:
1. `COMPLETE` canonical V8 contracts for `Inception`, `Soma Kernel`, `Central Council`, provider-policy scope, and identity/continuity state now exist.
2. `NEXT` define the config/bootstrap model that introduces those contracts through configuration, templates, bootstrap resolution, and precedence rules.
3. `REQUIRED` replace standing-team bootstrap assumptions with configurable organization-resolution contracts.
4. `REQUIRED` keep Workspace simple and Soma-first while making the operator model kernel-aware instead of fixed-identity-bound.
5. `REQUIRED` continue the full UI retheme and density-reduction effort under V8 delivery targets, not legacy V7 framing.

### 5. Testing and gate posture

Status:
1. `REQUIRED` every V8 migration slice must still carry backend, UI, and docs proof in the same delivery window.
2. `ACTIVE` architecture and state-file changes must be validated by doc-surface checks and in-app docs alignment where applicable.
3. `REQUIRED` runtime/API changes must include explicit UI review/test targets and focused live-flow evidence when user-visible behavior is touched.

### 6. Implementation-slice cleanup and convergence rule

Status:
1. `REQUIRED` every implementation slice that edits runtime, config, UI, or supporting logic must include a touched-file cleanup/convergence pass.
2. `REQUIRED` chunk completion for code-editing slices must include a focused review-team sweep covering contract alignment, dead-code scan, integration continuity, and test alignment.
3. `REQUIRED` deferred cleanup that cannot be completed safely in the same slice must be reported explicitly in the chunk result.

Implementation rule:
1. review each changed file for stale fixed-Soma/fixed-Council assumptions, dead branches, outdated comments, duplicated helpers, obsolete config paths, and naming drift
2. remove safe-to-remove stale code in the touched files as part of the same slice
3. keep the review scope to touched files and directly adjacent files needed for correctness
4. commit cleanup changes in the same chunk when they are part of the same logical convergence work

## Active V8 Queue

```text
Task 001  V8 alignment artifact and migration inventory             [IN_REVIEW]
Task 002  Inception / kernel / council contract definition          [COMPLETE]
Task 003  Provider-policy scope contract                            [COMPLETE]
Task 004  Config and bootstrap model planning                       [NEXT]
Task 005  Standing-team bootstrap de-hardcoding plan                [REQUIRED]
Task 006  Workspace/UI kernel-aware refactor plan                   [REQUIRED]
Task 007  V8 docs/state migration and grading discipline            [ACTIVE]
```

## Current Checkpoint (2026-03-12)

Delivery updates in this checkpoint:
1. `COMPLETE` reviewed the new root `README.md` and adopted V8 as the active development/grading target.
2. `COMPLETE` established `V8_DEV_STATE.md` as the canonical state file for new work.
3. `ACTIVE` migrated active non-archive documentation references from `V7_DEV_STATE.md` to `V8_DEV_STATE.md` where those docs define current execution discipline.
4. `COMPLETE` preserved V7 architecture-library documents as migration inputs rather than rewriting them into premature V8 specifications.

Evidence:
1. README directive review completed against `README.md`
2. V8 tracker alignment confirmed against `mycelis_vnext_implementation_tracker.md`
3. Task 001 artifact confirmed at `docs/architecture-library/TASK_001_VNEXT_ARCHITECTURE_ALIGNMENT_ARTIFACT.md`
4. active documentation references updated to point at `V8_DEV_STATE.md`

### 6. V8 contract shell introduction

Status:
1. `COMPLETE` `docs/architecture-library/V8_RUNTIME_CONTRACTS.md` now exists as the new V8 runtime contract shell.
2. `COMPLETE` the initial V8 runtime contract set is now fully defined.

### 7. Inception contract definition

Status:
1. `COMPLETE` the Inception contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes the top-level organization contract for V8.

### 8. Soma Kernel contract definition

Status:
1. `COMPLETE` the Soma Kernel contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes the configurable coordination/runtime layer for an Inception.

### 9. Central Council contract definition

Status:
1. `COMPLETE` the Central Council contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes configurable advisory composition rather than a fixed built-in council pattern.

### 10. Provider Policy contract definition

Status:
1. `COMPLETE` the Provider Policy contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` team-level and agent-level configuration scope are now captured explicitly as first-class contract needs.

### 11. Identity and Continuity State contract definition

Status:
1. `COMPLETE` the Identity and Continuity State contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` the initial V8 runtime contract set is now complete.
3. `NEXT` move to config/bootstrap model planning before backend refactor work begins.

### 12. V8 config/bootstrap planning shell introduction

Status:
1. `COMPLETE` `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` now exists as the planning shell for V8 config/bootstrap behavior.
2. `NEXT` define the `Configuration sources` section as the next bootstrap planning slice.

## Immediate Next Actions

1. `NEXT` define configuration sources in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` so V8 contracts can enter the system through explicit bootstrap inputs.
2. `NEXT` update the next-execution and governance guidance so delivery slices are expressed as V8 migration slices rather than only V7 holdovers.
3. `REQUIRED` apply the touched-file cleanup/convergence rule and review-team sweep to the first backend/runtime refactor chunk and all later code-editing slices.
4. `REQUIRED` validate doc-surface integrity after the state-file migration (`docs links`, `docs manifest`, and in-app docs visibility).
5. `REQUIRED` keep all new implementation/testing checkpoints in `V8_DEV_STATE.md` going forward.

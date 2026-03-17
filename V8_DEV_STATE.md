# Mycelis V8 - Development State

> Updated: 2026-03-17
> Canonical state file for active V8 grading and delivery tracking
> References: `README.md`, `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, `V7_DEV_STATE.md` (legacy migration input)

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
V8-2  Config and bootstrap model planning                          [COMPLETE]
V8-3  Backend primitive refactor                                   [REQUIRED]
V8-4  Frontend/operator refactor                                   [REQUIRED]
V8-5  Documentation and naming migration                           [ACTIVE]
V8-6  Verification and release hardening                           [REQUIRED]
```

## Current V8 Grading Baseline

### 1. Architecture alignment and migration inventory

Status:
1. `ACTIVE` V8 alignment is now carried by the README directive, the V8 runtime contracts, and the V8 bootstrap-planning surface instead of a separate tracker artifact.
2. `COMPLETE` current fixed-organism assumptions vs V8 target model have been mapped at the planning level.
3. `COMPLETE` backend/frontend/docs hotspots have been identified for the first migration wave.
4. `ACTIVE` carry those alignment conclusions into explicit V7-bootstrap migration rules and de-hardcoding work now that the canonical plan exists.

Primary references:
- `README.md`
- `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
- `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`

Key grading conclusion:
- Mycelis already has the execution substrate needed for V8, but still exposes one hardcoded Soma/Council organizational expression as if it were the platform itself.

### 2. Provider and service architecture

Status:
1. `ACTIVE` V8 explicitly requires support for local, hosted, and hybrid model-provider infrastructure.
2. `COMPLETE` provider abstraction is confirmed as an existing core subsystem, not a plugin afterthought.
3. `ACTIVE` instantiated-organization provider policy now drives runtime routing with inherited organization, kernel/council-role, team, and agent scope resolution.
4. `NEXT` extend provider-policy grading beyond the standing-team bridge so future instantiated organizations can declare richer approval/data-boundary posture without fallback-era assumptions.

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
3. `ACTIVE` top documentation entrypoints now link the detailed framework-memory surfaces needed for restart/onboarding (`README.md`, `docs/README.md`, `docs/archive/README.md`).
4. `COMPLETE` clearly stale dated architecture-review addenda were removed from `docs/architecture/`.
5. `REQUIRED` V7 architecture-library documents remain authoritative migration inputs until V8 replacements exist.

State-file rules:
1. all new status transitions, blocker updates, and gate evidence belong in `V8_DEV_STATE.md`
2. `V7_DEV_STATE.md` remains readable for historical migration evidence, but should not be the primary grading target for new work
3. active docs should describe V7 documents as migration inputs and V8 state as the live delivery scoreboard

### 4. Runtime and UI migration posture

Status:
1. `COMPLETE` canonical V8 contracts for `Inception`, `Soma Kernel`, `Central Council`, provider-policy scope, and identity/continuity state now exist.
2. `COMPLETE` the config/bootstrap model now documents configuration sources, template entry points, scope inheritance, precedence, and V7-to-V8 migration rules.
3. `REQUIRED` replace standing-team bootstrap assumptions with configurable organization-resolution contracts.
4. `REQUIRED` keep Workspace simple and Soma-first while making the operator model kernel-aware instead of fixed-identity-bound.
5. `REQUIRED` continue the full UI retheme and density-reduction effort under V8 delivery targets, not legacy V7 framing.

### 5. Testing and gate posture

Status:
1. `REQUIRED` every V8 migration slice must still carry backend, UI, and docs proof in the same delivery window.
2. `ACTIVE` architecture and state-file changes must be validated by doc-surface checks and in-app docs alignment where applicable.
3. `ACTIVE` clean-run discipline is now part of the delivery contract for runtime and integration-style checks: stop prior services, verify ports/processes are clear, sweep stray compiled Go binaries from prior runs, start only the minimum required stack, run the check, and tear the stack back down unless explicitly needed.
4. `REQUIRED` runtime/API changes must include explicit UI review/test targets and focused live-flow evidence when user-visible behavior is touched.

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
Task 001  V8 alignment and migration inventory                      [IN_REVIEW]
Task 002  Inception / kernel / council contract definition          [COMPLETE]
Task 003  Provider-policy scope contract                            [COMPLETE]
Task 004  Config and bootstrap model planning                       [COMPLETE]
Task 005  Standing-team bootstrap de-hardcoding plan                [ACTIVE]
Task 006  Workspace/UI kernel-aware refactor plan                   [REQUIRED]
Task 007  V8 docs/state migration and grading discipline            [ACTIVE]
Task 008  Planning-integration validation pass                      [COMPLETE]
Task 009  Next-execution/governance guidance migration              [NEXT]
```

## Current Checkpoint (2026-03-16)

Delivery updates in this checkpoint:
1. `COMPLETE` reviewed the new root `README.md` and adopted V8 as the active development/grading target.
2. `COMPLETE` established `V8_DEV_STATE.md` as the canonical state file for new work.
3. `ACTIVE` migrated active non-archive documentation references from `V7_DEV_STATE.md` to `V8_DEV_STATE.md` where those docs define current execution discipline.
4. `COMPLETE` preserved V7 architecture-library documents as migration inputs rather than rewriting them into premature V8 specifications.
5. `COMPLETE` documented how V7 bootstrap sources, implicit behaviors, fixed Soma/Council posture, and runtime-state coupling roll forward into the explicit V8 configuration/bootstrap model.
6. `COMPLETE` validated cross-doc planning consistency so README, the architecture-library index, runtime contracts, bootstrap model, docs manifest, and tests all reference the canonical V7->V8 migration contract.
7. `ACTIVE` landed the next Task 005 implementation cut: startup now instantiates runtime organization truth directly from self-contained bundle data in `core/config/templates/*.yaml` and fails closed unless a valid bootstrap bundle is available.
8. `COMPLETE` promoted clean-run testing discipline into the active testing/operations contract so runtime and integration checks must not stack on unknown local processes.
9. `COMPLETE` extended the clean-run contract so compiled Go services from prior `go build`, `go run`, or manual binary launches are explicitly detected and terminated before the next runtime or integration check.
10. `BLOCKED` the latest lifecycle/doc hardening commits are local-only until GitHub SSH-agent/key access is restored; branch publication is currently blocked by `Permission denied (publickey)` during `git push`.
11. `COMPLETE` refreshed top documentation entrypoints so restart/onboarding flow now points directly at the detailed framework-memory surfaces (`README.md`, `docs/README.md`, `docs/archive/README.md`).
12. `COMPLETE` removed clearly stale dated review docs from `docs/architecture/` to reduce documentation clutter without disturbing active migration inputs.
13. `COMPLETE` doc-surface validation is green again after the documentation cleanup pass.
14. `COMPLETE` wired provider-policy inheritance from the instantiated runtime organization into live Soma routing so organization defaults, kernel/council role defaults, team defaults, and agent overrides now resolve through one policy-bounded path.
15. `COMPLETE` retired the `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` runtime compatibility path so provider routing now resolves only from the instantiated organization policy and never from startup env maps.
16. `COMPLETE` retired the remaining no-bundle bootstrap fallback: normal startup now has one truth path only, `template -> instantiation -> runtime organization`, while mirrored `config/teams/*.yaml` packaging remains compatibility input rather than startup truth.
17. `COMPLETE` operator startup requirements are now explicit: Core requires at least one valid bootstrap bundle under `config/templates/`, and `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must be set whenever more than one bundle is present.

Evidence:
1. README directive review completed against `README.md`
2. runtime-contract alignment confirmed against `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
3. bootstrap-planning alignment confirmed against `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
4. active documentation references updated to point at `V8_DEV_STATE.md`
5. V7-to-V8 bootstrap migration narrative committed in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
6. planning-integration validation sweep recorded in this chunk plus `tests/test_docs_links.py`
7. template-bundle loader, runtime-organization instantiation, fail-closed startup selection logic, and the `v8-migration-standing-team-bridge` bundle landed in `core/internal/bootstrap/template_bundle.go`, `core/internal/bootstrap/template_bundle_test.go`, `core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup.go`, and mirrored chart config packaging
8. clean-run testing discipline is now documented in `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md`
9. lifecycle cleanup/status now sweep and report stray compiled Go services in `ops/lifecycle.py` with focused regression coverage in `tests/test_lifecycle_tasks.py`
10. compiled-service inspection now fails closed: `lifecycle.status` reports unknown inspection state and `lifecycle.down` blocks runtime/integration testing when local process inspection cannot verify cleanup
11. local branch currently carries `e64e249` and `01a4aca` for Chunk 4.7a, but remote publication is blocked until SSH-agent/key state is repaired
12. documentation cleanup refreshed `README.md`, `docs/README.md`, and `docs/archive/README.md`, and removed `docs/architecture/AGI_ARCHITECTURE_REVIEW_2026-03-06.md` plus `docs/architecture/STANDARDIZATION_REVIEW_2026-03-06.md`
13. validation: `uv run pytest tests/test_docs_links.py -q` -> `21 passed`
14. instantiated-organization provider policy now resolves through `core/internal/swarm/provider_policy.go`, is carried by `core/internal/bootstrap/template_bundle.go`, is applied during startup in `core/cmd/server/main.go`, and is exercised by focused bootstrap/swarm coverage
15. the standing-team bridge bundle now declares a conservative provider-policy default in both `core/config/templates/v8-migration-standing-team-bridge.yaml` and `charts/mycelis-core/config/templates/v8-migration-standing-team-bridge.yaml` so local and charted startup follow the same instantiated-organization routing path as tests
16. startup now fails closed when the bootstrap bundle set is missing or invalid, when `MYCELIS_BOOTSTRAP_TEMPLATE_ID` requests a bundle that is absent, and when multiple bundles exist without an explicit selection; runtime provider routing ignores legacy env-map inputs and startup truth now remains bundle-only in code/tests/state (`core/internal/bootstrap/template_bundle.go`, `core/cmd/server/bootstrap_startup.go`, `core/cmd/server/main.go`, `core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup_test.go`)
18. Windows `lifecycle.down` CIM timeout/inspection failure remains separate tooling debt; it is not part of the retired bootstrap fallback surface and still tracks under the lifecycle hardening work until resolved

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
3. `COMPLETE` the program has already moved into the config/bootstrap planning phase before backend refactor work begins.

### 12. V8 config/bootstrap planning shell introduction

Status:
1. `COMPLETE` `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` now exists as the planning shell for V8 config/bootstrap behavior.
2. `COMPLETE` the shell now has initial structure for user-facing and bootstrap-model planning slices.

### 13. User Concept Layer definition

Status:
1. `COMPLETE` the `User Concept Layer` is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` architecture-to-user concept mapping is now documented for the V8 UI mental model.
3. `COMPLETE` the beginner-friendly AI-organization mental model is now part of the bootstrap planning surface.

### 14. Bootstrap resolution flow definition

Status:
1. `COMPLETE` the `Bootstrap resolution flow` is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` team-level defaults, agent-level overrides, and beginner-versus-advanced bootstrap entry modes are now explicit in the V8 planning model.
3. `COMPLETE` the staged organization-resolution path is now documented for later implementation work.

### 15. Scope inheritance definition

Status:
1. `COMPLETE` the `Scope inheritance` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` the inheritance chain from `Inception -> Soma Kernel -> Central Council roles -> Team defaults -> Agent overrides` is now documented.
3. `COMPLETE` team and agent configuration are now explicitly treated as first-class scopes in bootstrap planning.

### 16. Precedence rules definition

Status:
1. `COMPLETE` the `Precedence rules` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` source precedence, scope precedence, policy-before-override handling, and conflict resolution are now part of the V8 bootstrap planning model.
3. `COMPLETE` the planning model now has enough structure to define how organizations are created before bootstrap activation.

### 17. Template and instantiation entry points definition

Status:
1. `COMPLETE` the `Template and instantiation entry points` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` V8 now distinguishes template-based creation, empty/manual creation, operator/API creation, and config-file/bootstrap creation as canonical organization-entry modes.
3. `COMPLETE` templates are now documented as reusable organization blueprints that feed bootstrap resolution with default kernel, council, team, and policy shape.
4. `COMPLETE` the `Migration from V7 bootstrap assumptions` section now explains how fixed V7 startup assumptions collapse into explicit V8 configuration sources, inheritance, and precedence rules.

### 18. Standing-team bootstrap de-hardcoding plan

Status:
1. `ACTIVE` Chunk 4.2 captured the standing-team de-hardcoding plan inside `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `ACTIVE` runtime/config references to fixed `prime-*` teams, hard-coded council rosters, kernel defaults, and provider wiring have documented migration paths into template definitions, bootstrap resolution, scoped inheritance, and provider-policy configuration.
3. `ACTIVE` Helm/runtime concerns (env injection for `MYCELIS_API_KEY`, 8080/8081 port normalization, config-file mounts, container storage) are now described alongside the bootstrap plan so infrastructure slices can align with runtime refactors.
4. `COMPLETE` current-state scan now records that the no-bundle startup dependency is gone: `core/internal/bootstrap/template_bundle.go` now fails closed when bundles are missing, env-map provider overrides are retired, and mirrored Helm/config packaging of `config/teams/*.yaml` no longer participates in startup truth.
5. `ACTIVE` the bridge now owns runtime startup truth through native bundle content: `core/internal/bootstrap/template_bundle.go` loads `core/config/templates/*.yaml` bundles, instantiates a runtime organization directly from bundle-defined org/team/agent data, and fails closed when no valid bootstrap bundle is available.

Evidence:
1. Plan section `## Standing-team bootstrap de-hardcoding plan` (2026-03-13) now explicitly calls out the active files and assumptions discovered in Chunk 4.2 (Helm values/env/ports, ops bridge tasks, bootstrap services, template APIs, team manifests, runtime storage expectations).
2. README review confirmed no additional guidance changes were required after adding the plan (`README.md`, 2026-03-13).
3. `v8-migration-standing-team-bridge` now validates as a self-contained V8 bundle through the bootstrap loader path without requiring standing-team manifest refs (`core/config/templates/v8-migration-standing-team-bridge.yaml`, `core/internal/bootstrap/template_bundle_test.go`).
4. startup selection regression coverage now explicitly proves bundle-instantiated runtime organizations, missing-bundle failure, invalid-bundle failure, missing-requested-bundle failure, and registry continuity with bundle presence (`core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup_test.go`, `core/internal/swarm/registry_test.go`).

Next steps:
1. Promote generated per-organization bootstrap bundles so startup stays bundle-only without depending on the fixed standing-team bridge asset forever.
2. Replace bootstrap seeding logic with template-instantiation + scope-aware inheritance.
3. Promote provider-policy scopes and Helm env/port/mount/storage alignment into actionable runtime slices with tests.

### 19. Cluster/runtime bootstrap contract alignment

Status:
1. `COMPLETE` Helm deployment now provisions the configuration bundle (`cognitive.yaml`, `policy.yaml`, standing-team YAMLs) via a ConfigMap volume so Pods read deterministic bootstrap inputs instead of whatever was baked in the image.
2. `COMPLETE` `MYCELIS_API_KEY` is required and injected through a Kubernetes Secret created by the chart (or supplied via `coreAuth.existingSecret`); `ops.k8s.deploy` now refuses to proceed when the key is missing.
3. `COMPLETE` Core HTTP port contract is unified on `8080` (`core/cmd/server` default, Helm `PORT` env, Service/bridge forwarding, ops defaults).
4. `COMPLETE` Charts and ops scripts document/mount the writable storage contract (`/data` PVC for artifacts + `$MYCELIS_WORKSPACE` under `/data/workspace`), and the ConfigMap keeps read-only bootstrap files under `/app/config`.

Evidence:
1. Helm templates: `charts/mycelis-core/templates/deployment.yaml`, `configmap-config.yaml`, `_helpers.tpl`, `core-auth-secret.yaml`, updated `values.yaml`, plus new `config/` assets.
2. Ops automation: `ops/k8s.py` enforces API key injection and 8080 port-forward; `ops/config.py` defaults shift to 8080; `ops/interface.py` startup guidance updated.
3. Runtime entrypoint: `core/cmd/server/main.go` now defaults to `PORT=8080`.
4. Validation: `uv run pytest tests/test_docs_links.py -q`.

Next steps:
1. Wire provider-policy scopes and template-instantiation flow into the runtime without depending on standing-team tables.
2. Extend Helm/ops surface so template bundles are generated from the new template serialization path instead of the baked defaults once Task 005 code work lands.

### 20. Testing/QA alignment review

Status:
1. `COMPLETE` browser QA test plan for Workspace chat now references the V8 contract (inline Soma chat, terminal states, direct-first routing) instead of the outdated V7 framing.
2. `COMPLETE` manual testing expectations now list the canonical outcomes (`answer`, `proposal`, `execution_result`, `blocker`) and reinforce the V8 inline-governance posture.

Evidence:
1. `tests/ui/browser_qa_plan_workspace_chat.md` updated on 2026-03-13 to remove V7 references and describe the V8 workspace, happy paths, and edge cases.

Next steps:
1. Keep automated Playwright suites aligned with the same inline-chat expectations once Workspace refactors land.

## Immediate Next Actions

1. `COMPLETE` run the planning-integration validation pass so README, the architecture-library index, docs manifests, and doc-tests all confirm the new V7-to-V8 bootstrap migration contract.
2. `NEXT` update the next-execution and governance guidance so delivery slices are expressed as V8 migration slices rather than only V7 holdovers.
3. `REQUIRED` apply the touched-file cleanup/convergence rule and review-team sweep to the first backend/runtime refactor chunk and all later code-editing slices.
4. `REQUIRED` validate doc-surface integrity after the state-file migration (`docs links`, `docs manifest`, and in-app docs visibility).
5. `REQUIRED` keep all new implementation/testing checkpoints in `V8_DEV_STATE.md` going forward.
6. `BLOCKED` restore SSH-agent/key access and push the latest local lifecycle/doc/state commits to the remote branch.
7. `NEXT` continue the documentation authority cleanup so active entrypoints stay lean while compatibility docs and archive material remain intentionally separated.
8. `NEXT` promote generated per-organization bootstrap bundles so startup remains bundle-only without relying on the fixed standing-team bridge asset long term.

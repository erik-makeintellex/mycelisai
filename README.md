# Mycelis - V8 Cognitive Infrastructure

Mycelis V8 is the active development target for the platform.

V8 extends the V7 operational foundation into an inception-driven system that can instantiate configurable AI organizations, preserve governed execution lineage, and support local, hosted, and hybrid cognitive infrastructure.

The V7 architecture-library remains the current authoritative planning surface until V8 replacement documents are written. Treat V7 architecture docs as migration inputs and `V8_DEV_STATE.md` as the live grading/state scoreboard for new work.

## README TOC

- [Fresh Agent Start Here](#fresh-agent-start-here)
- [Detailed Framework Memory](#detailed-framework-memory)
- [V8 Directive](#v8-directive)
- [Versioning Update](#versioning-update)
- [Feature Status Standard](#feature-status-standard)
- [Required Review Targets](#required-review-targets)
- [Command Contract](#command-contract)
- [Playwright Contract](#playwright-contract)
- [Development Workflow](#development-workflow)
- [Documentation Responsibilities](#documentation-responsibilities)
- [Status](#status)

## Fresh Agent Start Here

Review these in order before touching code or planning state:

1. [AGENTS.md](AGENTS.md)
2. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
3. [Target Deliverable V7](docs/architecture-library/TARGET_DELIVERABLE_V7.md)
4. [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
5. [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md)
6. [UI Target And Transaction Contract V7](docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
7. [Operations](docs/architecture/OPERATIONS.md)
8. [Testing](docs/TESTING.md)
9. [V7 Development State](V7_DEV_STATE.md)
10. [V8 Development State](V8_DEV_STATE.md)
11. [Docs Manifest](interface/lib/docsManifest.ts)

Bootstrap reminder:
- treat `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` as the canonical V7->V8 migration and bootstrap contract, not just another planning note
- always translate V7 YAML, runtime config, DB seeding, and operator wizard flows through that model before they touch a live organization
- `Template ≠ instantiated organization`, so only instantiated orgs enter bootstrap resolution while templates stay reusable blueprints
- Task 005 bridge layer: `core/config/templates/*.yaml` now instantiates the startup runtime organization through the bundle path, and normal startup fails closed unless a valid bootstrap bundle is present; if more than one bundle is mounted, `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must select one explicitly

Fresh-agent review rule:
- V7 docs define the current authoritative architecture/planning contract until migrated.
- `V7_DEV_STATE.md` is legacy migration history.
- `V8_DEV_STATE.md` is the active grading target for new work.

## Detailed Framework Memory

Use these as the top detailed references when you need the deeper framework contract rather than just the quick-start path.

1. [V8 Config and Bootstrap Model](docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md)
   - canonical memory for template vs instantiated organization, bootstrap resolution, inheritance, precedence, and V7-to-V8 bootstrap translation
2. [V8 Runtime Contracts](docs/architecture-library/V8_RUNTIME_CONTRACTS.md)
   - canonical memory for Inception, Soma Kernel, Central Council, Provider Policy, and Identity / Continuity State
3. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
   - canonical map of which detailed planning doc owns which part of the framework
4. [System Architecture V7](docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md)
   - detailed runtime, storage, NATS, deployment, and service-boundary memory until V8 replacements land
5. [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
   - detailed workflow, run, manifest, recurring-plan, and activation memory
6. [Delivery Governance And Testing V7](docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
   - detailed acceptance, gate, and proof requirements for implementation slices
7. [Team Execution And Global State Protocol V7](docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md)
   - detailed state-file, coordination, and execution-discipline memory for multi-slice work
8. [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md)
   - detailed operator experience, simplification, and anti-complexity memory for the UI layer
9. [UI Target And Transaction Contract V7](docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
   - detailed UI transaction/state expectations for operator-visible behavior

Rule:
- when framework behavior, bootstrap posture, organization shape, or operator model is unclear, load the owning detailed doc above before making assumptions
- keep README as the entrypoint, but treat the documents in this section as the deeper memory surface for framework specifics
- current MVP UI release posture is Team Lead-first by default; `Resources`, `Memory`, and `System` are advanced support routes rather than default operator entrypoints

## V8 Directive

Mycelis V8 introduces inception-driven AI organizations.

Canonical runtime shape:

```text
Inception
  -> Soma Kernel
  -> Central Council
  -> Specialist Teams / Agents
  -> Runs
  -> Events
  -> Memory
  -> Reflection
```

Initial V8 contract set now covers:
- Inception
- Soma Kernel
- Central Council
- Provider Policy
- Identity and Continuity State

The active V8 bootstrap-planning surface now also defines how organizations enter the system through templates, manual creation, operator/API creation, and config-file/bootstrap paths.

For default user-facing experience, these concepts should translate into a simpler team model:

```text
AI Organization
  -> Team Lead
  -> Advisors
  -> Departments
  -> Specialists
```

The continuity layer is intended to persist structured self-related state across runs while being informed by memory and reflection.

V8 implementation philosophy:
1. preserve existing runtime infrastructure where it already supports the target model
2. replace fixed Soma/Council assumptions with configurable cognition layers
3. keep local-first deployment as a core capability
4. preserve deterministic execution and auditable event trails
5. support hybrid model-provider infrastructures

## Versioning Update

All references to V7 in the repository should now be interpreted as legacy architecture planning documents supporting the transition to V8.

Migration inputs that remain authoritative until replaced:
- `docs/architecture-library/TARGET_DELIVERABLE_V7.md`
- `docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md`
- `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md`
- `docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md`
- `docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md`
- `V7_DEV_STATE.md` (legacy migration state history)
- `V8_DEV_STATE.md` (active grading state)

Versioning rule:
- use V7 documents as migration inputs
- use `V8_DEV_STATE.md` as the live delivery scoreboard
- migrate naming, runtime assumptions, and documentation incrementally so planning continuity is not broken
- treat every V7 bootstrap asset (YAML, runtime config, DB seeding flows, operator wizards) as input material that must be translated into the explicit V8 configuration/bootstrap model before it shapes a running organization
- remember that templates remain reusable blueprints (`Template ≠ instantiated organization`); only instantiated organizations enter runtime bootstrap resolution
- the compatibility contract for that translation lives in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`

## Feature Status Standard

Use these canonical status markers in planning and state docs:
- `REQUIRED`
- `NEXT`
- `ACTIVE`
- `IN_REVIEW`
- `COMPLETE`
- `BLOCKED`

Do not replace them with ad hoc synonyms when a canonical marker already fits.

## Required Review Targets

Agents implementing V8 must review these areas first:
- `docs/architecture-library/`
- `docs/architecture/`
- `docs/TESTING.md`
- `docs/logging.md`
- `V7_DEV_STATE.md`
- `V8_DEV_STATE.md`

Particular attention belongs on:
- execution slices
- team execution protocol
- delivery governance rules
- UI operator experience contracts
- runtime orchestration assumptions
- provider routing and hybrid deployment posture
- When working in the execution/gov docs (`docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md`, `docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md`, `docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md`), treat all V7 content as migration input: translate assets through `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, keep `Template ≠ instantiated organization`, and record slice state in `V8_DEV_STATE.md`.

## Command Contract

Required command references for active V8 work:
- `uv run inv ci.entrypoint-check`
- `uv run inv cache.status`
- `uv run inv cache.clean`
- `uv run inv lifecycle.memory-restart`
- `uv run inv team.architecture-sync`

Use `uv run inv ...` for execution.
Use `uvx --from invoke inv -l` only as a compatibility probe.
Do not use bare `uvx inv ...`.

Provider/runtime workflow reminders:
- review architecture and state docs before implementation slices
- attach tests and evidence in the same delivery window
- keep state-file updates current with gate results and blocker changes
- keep repo-managed caches under `workspace/tool-cache` and use `cache.apply-user-policy` when a Windows user profile needs heavy tool caches moved off `C:`
- check `uv run inv cache.status` before large build/test/browser runs when disk headroom is tight; the main repo-local growth surfaces are `workspace/tool-cache`, `interface/.next`, Playwright browser binaries, and other generated test artifacts
- use `uv run inv cache.clean` as the first repo-safe reclaim path when builds or tests start failing under disk pressure instead of manually deleting random working files
- expect Invoke-managed Interface build/test/browser tasks to sweep repo-local Next/Vitest/Playwright worker residue after each run so old `node.exe` workers do not accumulate between sessions
- expect Invoke-managed Interface and CI tasks to execute from the `interface/` working directory through the same `npm`/`node` entrypoints on Windows and Linux rather than relying on shell-specific `cd ... &&` wrappers
- project-owned config backstops now keep direct local commands aligned too: root `.npmrc` anchors npm/npx cache in `workspace/tool-cache`, pytest stores cache metadata in `workspace/tool-cache/pytest`, and task-managed Interface runs disable Next telemetry while routing Playwright browser binaries through the managed cache root

Suggested development build configuration by platform:
- Windows: keep repo work on a spacious non-system drive when possible, use `uv run inv cache.apply-user-policy` so uv/pip/npm/go/Playwright stop drifting back onto `C:`, and treat Docker Desktop / WSL storage separately from repo-managed cache cleanup
- Linux/macOS: keep `MYCELIS_PROJECT_CACHE_ROOT` on a volume with headroom if the default workspace disk is small, and export user-level cache roots only when you need tool caches off your default home volume; the repo task path already keeps build/test/browser churn inside `workspace/tool-cache`
- All platforms: prefer `uv run inv ...` over raw tool commands for repeated build/test cycles, because the task path applies the managed cache roots, disables low-value telemetry writes, and sweeps leftover Interface workers that can hold build outputs open

## Playwright Contract

`uv run inv interface.e2e` owns the local Next.js server lifecycle for browser test runs, routes Playwright browsers through the managed project cache, and leaves no repo-local UI workers behind when it exits. Playwright owns the Next.js server lifecycle for the default browser gate.

Browser matrix baseline:
- `chromium firefox webkit`
- `mobile-chromium` when route/mobile smoke is part of the gate

Documentation rule:
- root and testing docs must not imply that default browser validation depends on a manually pre-started server
- default release-candidate browser coverage is MVP-aligned; legacy V7 or raw-endpoint-only specs should stay outside the default gate unless a slice explicitly revives them
- live-backend browser checks are still required when proxy/runtime contracts change

## Development Workflow

Agents implementing V8 should follow this process:
1. clean runtime environment and running services when the slice requires a deterministic baseline
   - `uv run inv lifecycle.down` now treats repo-local Interface worker residue as part of the shutdown contract, not just bound ports
   - when repeated build/test cycles have been running for a while, clear stale runtime residue before assuming the issue is just disk: leaked Interface workers and long-lived local services can keep caches hot and hold build outputs open
2. review the implementation tracker
3. review V7 architecture-library documentation as migration input
4. identify migration targets and required contract updates
5. implement incremental runtime or documentation updates
6. verify with tests and execution gates
   - if the machine is low on free space, prefer the repo task path in this order: `uv run inv lifecycle.down`, `uv run inv cache.status`, then `uv run inv cache.clean`
   - if you are setting up a new development machine, treat cache placement as part of the build config, not an afterthought: Windows should stamp user-level cache vars early, and Linux/macOS should point project/user cache roots at the volume you actually want repeated builds and browser runs to consume
7. update `V8_DEV_STATE.md` with current status and evidence

## Documentation Responsibilities

Every implementation slice must update:
- `README.md`
- `V8_DEV_STATE.md`
- architecture-library planning documents when target, execution, UI, or delivery rules change
- documentation manifest when a canonical doc should be visible in the in-app docs page

The architecture-library remains the authoritative planning surface until the V8 library replaces it.

## Status

Mycelis is currently transitioning from V7 operational architecture to V8 inception-driven cognitive infrastructure.

The V7 system provides the runtime foundation.
V8 extends that foundation to support configurable AI organizations, provider-policy scoping, kernel-aware execution architecture, and structured identity/continuity state across runs.

# Mycelis - V8 Cognitive Infrastructure

Mycelis V8 is the active development target for the platform.

V8 extends the V7 operational foundation into an inception-driven system that can instantiate configurable AI organizations, preserve governed execution lineage, and support local, hosted, and hybrid cognitive infrastructure.

The V7 architecture-library remains the current authoritative planning surface until V8 replacement documents are written. Treat V7 architecture docs as migration inputs and `V8_DEV_STATE.md` as the live grading/state scoreboard for new work.

## README TOC

- [Fresh Agent Start Here](#fresh-agent-start-here)
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
- Task 005 bridge layer: `core/config/templates/*.yaml` now instantiates the startup runtime organization through the bundle path, while direct `config/teams` scanning remains a temporary compatibility fallback only when no bundle is configured

Fresh-agent review rule:
- V7 docs define the current authoritative architecture/planning contract until migrated.
- `V7_DEV_STATE.md` is legacy migration history.
- `V8_DEV_STATE.md` is the active grading target for new work.

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
- `uv run inv lifecycle.memory-restart`
- `uv run inv team.architecture-sync`

Use `uv run inv ...` for execution.
Use `uvx --from invoke inv -l` only as a compatibility probe.
Do not use bare `uvx inv ...`.

Provider/runtime workflow reminders:
- review architecture and state docs before implementation slices
- attach tests and evidence in the same delivery window
- keep state-file updates current with gate results and blocker changes

## Playwright Contract

Playwright owns the Next.js server lifecycle for browser test runs.

Browser matrix baseline:
- `chromium firefox webkit`
- `mobile-chromium` when route/mobile smoke is part of the gate

Documentation rule:
- root and testing docs must not imply that Playwright depends on manually pre-started servers
- live-backend browser checks are still required when proxy/runtime contracts change

## Development Workflow

Agents implementing V8 should follow this process:
1. clean runtime environment and running services when the slice requires a deterministic baseline
2. review the implementation tracker
3. review V7 architecture-library documentation as migration input
4. identify migration targets and required contract updates
5. implement incremental runtime or documentation updates
6. verify with tests and execution gates
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

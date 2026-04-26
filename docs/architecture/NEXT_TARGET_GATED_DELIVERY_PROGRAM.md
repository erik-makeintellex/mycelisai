# Next Target Gated Delivery Program

> Status: `P1 ACTIVE`
> Last Updated: 2026-03-07
> Execution Rule: advance only after the current phase records a gate pass with evidence.

## Phase Register

| Phase | Status | Theme |
| --- | --- | --- |
| `P0` | COMPLETE | Operational foundation and gate discipline |
| `P1` | ACTIVE | Logging, error handling, and source-tree cleanup under the `<=300` LOC no-regression policy |
| `P2` | BLOCKED | Meta-agent-owned manifest pipeline |
| `P3` | BLOCKED | Workflow-composer onboarding and execution-facing UI |
| `P4` | BLOCKED | Release hardening + final regression caps |

## P0 Action Card

- Objective: make the logging/quality gate surface executable from the project environment, eliminate the raw `dotenv` traceback on lifecycle tasks, and align the operator command contract before any downstream phase starts.
- Scoped files:
  - `ops/db.py`
  - `ops/lifecycle.py`
  - `README.md`
  - `docs/TESTING.md`
  - `docs/architecture/OPERATIONS.md`
  - `docs/LOCAL_DEV_WORKFLOW.md`
  - `docs/architecture/NEXT_TARGET_GATED_DELIVERY_PROGRAM.md`
  - `tests/test_lifecycle_tasks.py`
  - `tests/test_db_tasks.py`
- Branch name: `phase/p0-invoke-gate-normalization`
- Required `uv run inv` commands:
  - Target contract after gate pass: `uv run inv logging.check-schema`
  - Target contract after gate pass: `uv run inv logging.check-topics`
  - Target contract after gate pass: `uv run inv quality.max-lines --limit 300`
  - Target contract after gate pass: `uv run inv lifecycle.memory-restart --build --frontend`
  - Compatibility probe only: `uvx --from invoke inv -l`
  - Current execution note: this workspace verified bare `uvx inv -l` fails with `Package inv does not provide any executables`; operator evidence must be captured with `uv run inv ...`
- Tests:
  - `uv run pytest tests/test_db_tasks.py tests/test_lifecycle_tasks.py tests/test_logging_tasks.py tests/test_quality_tasks.py -q`
  - `uv run inv -l`
  - Optional destructive validation after local stack readiness: `uv run inv lifecycle.memory-restart --build --frontend`
- Acceptance criteria:
  - lifecycle/db invoke tasks exit with a clear remediation message when `python-dotenv` is missing from the active invoke environment
  - top-level operator docs no longer imply that bare `uvx inv ...` is valid for project tasks, and keep `uvx --from invoke inv -l` as probe-only
  - logging and max-lines gate commands are documented as P0 prerequisites for all later phases
  - no new P1-P4 implementation work merges before P0 evidence is attached
- Rollback plan:
  - revert the P0 doc and operator-doc contract updates
  - revert the task-layer env-loading guidance in `ops/db.py` and `ops/lifecycle.py`
  - rerun `uv run pytest tests/test_db_tasks.py tests/test_lifecycle_tasks.py -q`
- PR evidence block:
  - Branch: `phase/p0-invoke-gate-normalization`
  - Date: `2026-03-06`
  - Commands:
    - `uv run inv -l`
    - `uv run pytest tests/test_db_tasks.py tests/test_lifecycle_tasks.py tests/test_logging_tasks.py tests/test_quality_tasks.py -q`
  - Required output summary:
    - `uv run inv -l` lists the `lifecycle.memory-restart`, `logging.check-schema`, `logging.check-topics`, and `quality.max-lines` tasks
    - pytest suite passes with no failures
  - Additional evidence:
    - `uv run inv ci.entrypoint-check`
    - `uv run inv lifecycle.memory-restart --frontend`
  - Required output summary:
    - `uv run inv ci.entrypoint-check` confirms `uv run inv` and `uvx --from invoke inv` behavior while bare `uvx inv` remains unsupported
    - `uv run inv lifecycle.memory-restart --frontend` completes with clean forward migrations, healthy endpoints, and memory probes at HTTP 200
  - Gate result: `COMPLETE`

## P1 Action Card

- Objective: standardize operator-facing logging and error handling in the highest-risk execution paths while reducing oversized source files toward the global `300` LOC policy under the temporary no-regression caps tracked in `ops/quality_legacy_caps.txt`.
- Scoped files:
  - `docs/logging.md`
  - `core/internal/swarm/agent.go`
  - `core/internal/swarm/internal_tools.go`
  - `core/internal/swarm/soma.go`
  - `core/internal/server/comms.go`
  - `core/internal/mcp/service.go`
  - `interface/components/dashboard/CouncilCallErrorCard.tsx`
  - `interface/components/dashboard/DegradedModeBanner.tsx`
  - `interface/components/dashboard/MissionControlChat.tsx`
  - `interface/store/useCortexStore.ts`
  - `ops/quality.py`
  - `ops/quality_legacy_caps.txt`
  - `ops/lifecycle.py`
- Branch name: `phase/p1-hotpath-loc-decomposition`
- Required `uv run inv` commands:
  - `uv run inv logging.check-schema`
  - `uv run inv logging.check-topics`
  - `uv run inv quality.max-lines --limit 300`
  - `uv run inv ci.baseline`
- Tests:
  - `uv run inv logging.check-schema`
  - `uv run inv logging.check-topics`
  - `uv run inv quality.max-lines --limit 300`
  - `uv run inv ci.baseline`
  - targeted package and UI tests for each scoped error-handling or decomposition area
- Acceptance criteria:
  - scoped operator-facing failures use actionable blocker or recovery language instead of opaque generic failures
  - logging and signal metadata remain compliant with the documented standard where touched
  - no scoped file grows beyond its current legacy cap
  - at least one legacy cap is reduced or removed
  - extracted modules keep behavior stable under existing tests
- Rollback plan:
  - revert the scoped logging/error/decomposition commits per file family
  - restore prior values in `ops/quality_legacy_caps.txt`
  - rerun baseline and targeted tests
- PR evidence block:
  - Branch: `phase/p1-hotpath-loc-decomposition`
  - Date: `2026-03-07`
  - Commands:
    - `uv run inv logging.check-schema`
    - `uv run inv logging.check-topics`
    - `uv run inv quality.max-lines --limit 300`
    - `uv run inv ci.baseline`
  - Gate result: `ACTIVE`

## P2 Action Card

- Objective: make the meta-agent path the authoritative producer and validator of manifests from intent through activation.
- Scoped files:
  - `core/internal/provisioning/engine.go`
  - `core/internal/server/provision.go`
  - `core/internal/server/proposals.go`
  - `core/internal/swarm/internal_tools.go`
  - `core/pkg/protocol/manifest.go`
  - `core/pkg/protocol/workflow_composer.go`
  - `core/tests/provisioning_test.go`
  - `docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md`
- Branch name: `phase/p2-meta-agent-manifest-pipeline`
- Required `uv run inv` commands:
  - `uv run inv core.test`
  - `uv run inv logging.check-schema`
  - `uv run inv logging.check-topics`
- Tests:
  - `uv run inv core.test`
  - targeted provisioning/server integration tests
  - manifest validation contract tests
- Acceptance criteria:
  - intent to manifest generation uses one canonical validation path
  - invalid manifests fail before activation with actionable diagnostics
  - manifest activation remains linked to run/event lineage
- Rollback plan:
  - revert manifest pipeline changes
  - restore prior provisioning/activation path
  - rerun core tests
- PR evidence block:
  - Branch: `phase/p2-meta-agent-manifest-pipeline`
  - Date: `TBD after P1 pass`
  - Commands:
    - `uv run inv core.test`
    - targeted manifest contract suite
  - Gate result: `BLOCKED`

## P3 Action Card

- Objective: onboard operators into the workflow-composer UI with clear single-agent versus manifested-team paths, policy-aware validation, and execution-facing terminal states.
- Scoped files:
  - `interface/lib/types/workflowComposer.ts`
  - `interface/app/(app)/docs/page.tsx`
  - `interface/lib/docsManifest.ts`
  - workflow-composer route/components under `interface/app` and `interface/components`
  - `docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md`
- Branch name: `phase/p3-workflow-composer-onboarding`
- Required `uv run inv` commands:
  - `uv run inv interface.test`
  - `uv run inv interface.build`
  - `uv run inv logging.check-schema`
- Tests:
  - `uv run inv interface.test`
  - `uv run inv interface.build`
  - targeted route/component tests for onboarding flows
- Acceptance criteria:
  - onboarding exposes direct versus manifested execution choices
  - docs/browser manifest includes the authoritative composer guidance
  - composer entry path preserves policy and lineage vocabulary from P0/P2
- Rollback plan:
  - revert onboarding UI/doc manifest changes
  - rerun interface test/build gates
- PR evidence block:
  - Branch: `phase/p3-workflow-composer-onboarding`
  - Date: `TBD after P2 pass`
  - Commands:
    - `uv run inv interface.test`
    - `uv run inv interface.build`
  - Gate result: `BLOCKED`

## P4 Action Card

- Objective: close the program with release-hardening gates, evidence automation, and no-regression enforcement across logging, manifests, cleanup, and UI onboarding.
- Scoped files:
  - `ops/ci.py`
  - `README.md`
  - `V8_DEV_STATE.md`
  - `docs/TESTING.md`
  - `docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md`
- Branch name: `phase/p4-release-hardening`
- Required `uv run inv` commands:
  - `uv run inv ci.baseline`
  - `uv run inv ci.release-preflight --e2e --strict-toolchain`
  - `uv run inv lifecycle.health`
- Tests:
  - `uv run inv ci.baseline`
  - `uv run inv ci.release-preflight --strict-toolchain`
  - release-note and docs-manifest checks
- Acceptance criteria:
  - all prior phase gates are referenced in one release evidence bundle
  - release preflight enforces clean-tree, docs/state sync, and reproducible validation commands
  - unresolved critical defects block promotion
- Rollback plan:
  - revert release-hardening changes
  - restore prior CI/doc gating if required
  - rerun baseline
- PR evidence block:
  - Branch: `phase/p4-release-hardening`
  - Date: `TBD after P3 pass`
  - Commands:
    - `uv run inv ci.baseline`
    - `uv run inv ci.release-preflight --strict-toolchain`
  - Gate result: `BLOCKED`

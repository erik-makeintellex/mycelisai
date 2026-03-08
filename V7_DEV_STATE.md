# Mycelis V7 — Development State

> **Updated:** 2026-03-07
> **References:** `mycelis-architecture-v7.md` (PRD index), `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md` (canonical planning library), `docs/V7_IMPLEMENTATION_PLAN.md` (Blueprint)

---

## Progress Summary

### Feature Status Legend

- `REQUIRED`: must exist for target delivery or gate pass
- `NEXT`: highest-priority upcoming slice
- `ACTIVE`: currently in development
- `IN_REVIEW`: implemented and awaiting validation/review
- `COMPLETE`: delivered and accepted
- `BLOCKED`: cannot advance until a dependency or defect is resolved

### Delivery Program Snapshot

```
P0  Operational foundation and gate discipline                       [COMPLETE]
P1  Logging, error handling, and hot-path cleanup                    [ACTIVE]
P2  Meta-agent-owned manifest pipeline                               [REQUIRED]
P3  Workflow-composer onboarding and execution-facing UI             [REQUIRED]
P4  Release hardening and promotion gates                            [REQUIRED]
```

### Active Queue In Canonical Order

```
Slice 1  Launch Crew and workflow onboarding execution contract      [ACTIVE]
Slice 2  P1 logging, error handling, and execution feedback          [ACTIVE]
Slice 3  Prime-development reply reliability                         [NEXT]
Slice 4  P1 hot-path cleanup under max-lines gates                   [REQUIRED]
Slice 5  P2 manifest pipeline preparation                            [REQUIRED]
```

### Delivered Foundation Context

- `COMPLETE`: Workspace, run/timeline, docs browser, provider/profile/runtime management, event spine, trigger engine, and the core operational foundation are in place.
- `COMPLETE`: `P0` gate passed live and the canonical task/docs/runtime contracts are now normalized around `uv run inv ...`.
- `ACTIVE`: current work should be interpreted through the `P0 -> P4` delivery model first, not legacy Team A/B/C or Phase 19 naming.

---

## Current Checkpoint (2026-03-07)

Latest integration checkpoint:
- Base commit: `0b07fb5`
- Branch: `feature/enterprise-multihost-soma-routing`
- Runtime note: local verification used `go1.25.6 windows/amd64` while docs still target Go `1.26`
- Delivered:
  - Launch Crew onboarding execution contract is now the active product slice:
    - state/readme alignment keeps Launch Crew/workflow onboarding as the first execution-facing queue item
    - the next implementation change must prove Launch Crew terminates in `answer`, `proposal`, `execution_result`, or `blocker`
    - proof must include component tests, store/API transaction tests, and browser coverage for both onboarding proposal and blocker outcomes
  - Canonical planning/state cleanup completed:
    - the active queue now follows the architecture-library ordering of operator-facing execution clarity -> operator-facing failure/recovery clarity -> internal coordination -> structural cleanup -> manifest unification
    - the gated delivery program now aligns `P0 -> P4` themes with the target-deliverable library instead of mixing older narrower phase labels
    - this state document now leads with the delivery-program snapshot instead of legacy phase framing
  - DB-backed collaboration groups (`migration 034`, `groups.go`)
  - scoped permissions (`groups:read|write|broadcast`)
  - high-impact confirm-token gate on group mutation
  - group bus monitor API and system-status integration
  - Teams page group-management panel + status drawer signal
  - targeted integration tests for lifecycle, denial paths, fanout, monitor
  - Soma/Council engagement-path protocol codified (internal vs MCP vs external API vs code-loop vs team instantiation)
  - Standing team prompt contracts updated (`core/config/teams/admin.yaml`, `core/config/teams/council.yaml`)
  - New architecture authority doc for deterministic execution pathing (`docs/architecture/SOMA_COUNCIL_ENGAGEMENT_PROTOCOL_V7.md`)
  - In-app docs browser manifest updated to include engagement protocol doc
  - Local command V0 started: host actions/status APIs + allowlisted no-shell execution module (`core/internal/hostcmd`, `/api/v1/host/actions`)
  - Soma internal tool added for bounded local command execution (`local_command`)
  - Soma/Council prompt contracts updated for symbiote partner posture + code-first ephemeral execution preference + plan/develop/verify/deliver team formation
  - Delegate-task execution hardening: tolerant argument normalization for object/alias payloads to prevent schema-only failure loops
  - MCP translation behavior hardened: Soma/Council now required to map user intent to currently installed MCP tool inventory and execute concrete tool calls (or emit explicit missing dependency requirements)
  - Coder-first web access contract codified: search/site retrieval defaults to ephemeral code execution with adaptive query strategy; MCP used when easier/required
  - Root-admin scope expansion codified: Soma/Council must handle full-platform configuration execution requests (providers/profiles, governance, MCP/toolsets, users/groups, runtime settings), not only new-team flows
  - Cognitive startup probing hardened:
    - startup calibration now scopes connectivity checks to default `ollama` plus profile-routed providers
    - avoids startup connection attempts to unrelated declared backends
    - default `cognitive.yaml` now keeps `vllm` and `lmstudio` disabled until explicitly configured
  - Architecture authority updated to include startup connectivity policy:
    - `mycelis-architecture-v7.md` provider-routing section now defines startup probe scope and opt-in model for additional backends
  - Soma execution guardrail added: non-executing "Step 1 / we need to delegate" responses now trigger a policy-correction re-inference pass to force real tool execution or a concrete blocker
  - Soma execution guardrail hardened:
    - detects "Example Input / this will route / I'll consult" instructional narration as non-executing output
    - parses legacy `{"operation":"...","arguments":...}` payloads as executable tool calls to prevent instruction-only loops
    - auto-fills `consult_council.question` from the latest user request when omitted by model output
    - normalizes council member aliases (`Architect` → `council-architect`) and signal aliases (`topic_pattern` → `subject`)
    - extracts missing NATS subjects from user input for `read_signals` execution fallback
    - recovers malformed/incomplete `tool_call` JSON via loose parser fallback
  - ReAct failure handling hardened:
    - tool execution/lookup failures now feed back into re-inference instead of immediately terminating the turn
    - prevents premature user-facing failure when a recoverable follow-up tool call is available
  - Council-first execution policy enforced in runtime for high-impact action paths:
    - before `create_team` / `delegate_task` / `local_command`, Soma now runs a council preflight consult
    - preflight feedback is injected back into the same turn so execution parameters can be refined before action
  - Dynamic team instantiation path added:
    - new internal `create_team` tool for runtime team creation via Soma
    - admin team tool inventory now includes `create_team`
    - adapter path maps malformed `delegate_task` team-creation payloads into `create_team` calls
    - `create_team` accepts nested `manifest` payload shape for compatibility with local model outputs
    - added swarm policy tests for operation-payload fallback and stronger non-executing response detection
  - User-facing docs expanded in the in-app `/docs` browser for runtime recovery and execution behavior:
    - `docs/user/system-status-recovery.md` (new)
    - updated `docs/user/core-concepts.md`, `docs/user/automations.md`, `docs/user/soma-chat.md`
  - In-app docs manifest updated to expose the new user guide (`system-status-recovery`)
  - Assistant display-name support delivered (`assistant_name`) with persisted backend settings and UI propagation across Workspace/status/error/runs surfaces
  - Settings Profile now exposes orchestrator rename flow (`Settings -> Profile -> Assistant Name`)
  - Fresh local deployment reset sequence documented for operators (`lifecycle.down -> k8s.reset -> lifecycle.up --build --frontend -> lifecycle.health`)
  - Workspace cleanup pass: chat-first default split ratio (68/32) and telemetry row moved behind Advanced Mode to reduce operational clutter in normal mode
  - Image lifecycle hardening: `generate_image` now cache-first (60m TTL), periodic backend cleanup of expired unsaved images, and explicit save flow (`save_cached_image` tool + `POST /api/v1/artifacts/{id}/save` API) to persist into `workspace/saved-media`
  - User/operator surfaces expanded: `Settings -> Users & Groups` now includes actionable user-management elements and embedded collaboration-group management UI (no longer user-table stub only)
  - Cluster bring-up contract hardened in ops/runtime:
    - new `uv run inv k8s.up` canonical sequence (`init -> deploy -> wait`)
    - new `uv run inv k8s.wait` readiness gates (`PostgreSQL -> NATS -> Core API`)
    - `k8s.recover` corrected to restart actual chart resources (`mycelis-core`, `mycelis-core-nats`, `mycelis-core-postgresql`)
    - Helm core deployment now has startup/readiness/liveness probes on `/healthz` for rollout-health accuracy
  - Operator docs aligned to canonical cluster sequencing:
    - `README.md`
    - `docs/LOCAL_DEV_WORKFLOW.md`
    - `docs/architecture/OPERATIONS.md`
  - Logging authority refresh completed:
    - `docs/logging.md` rewritten as implementation-aligned standard (`mission_events` + `log_entries` surfaces, taxonomy, onboarding checklist, anti-patterns)
  - Agent operating manual now enforces logging first-read checklist:
    - `CLAUDE.md` updated with mandatory logging contract section
  - PRD execution manifest updated for immediate delivery sequencing:
    - `docs/architecture-library/TARGET_DELIVERABLE_V7.md` and `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md` now carry the authoritative `P0 -> P4` sequence and slice ordering
  - Workflow composer architecture plan aligned with logging prerequisite gate:
    - `docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md`
  - Invoke quality/logging gate tasks implemented and wired into baseline:
    - `uv run inv logging.check-schema`
    - `uv run inv logging.check-topics`
    - `uv run inv quality.max-lines --limit 350`
    - `uv run inv ci.baseline` now runs these gates before core/interface validation
  - Hot-path topic literals normalized to protocol constants:
    - added `protocol.TopicGlobalInputFmt`
    - replaced inline subject literals in `core/internal/swarm/internal_tools.go`, `core/internal/swarm/soma.go`, and `core/internal/server/comms.go`
  - Memory recovery workflow hardened:
    - new lifecycle task `uv run inv lifecycle.memory-restart` added (`down -> db.reset -> up -> health -> memory endpoint probes`)
    - memory endpoint probes now part of deterministic reset readiness (`/api/v1/memory/stream`, `/api/v1/memory/sitreps?limit=1`)
    - README + testing docs updated with fresh memory restart runbook and expected outcomes
    - lifecycle task unit tests added for sequence and failure handling (`tests/test_lifecycle_tasks.py`)
  - Invoke/runtime contract normalized for local operator workflows:
    - canonical task runner docs now use `uv run inv ...` / `.\.venv\Scripts\inv.exe ...`
    - `ops/db.py` and `ops/lifecycle.py` now fail with actionable guidance when launched from an invoke environment missing `python-dotenv`
    - gated execution program recorded at `docs/architecture/NEXT_TARGET_GATED_DELIVERY_PROGRAM.md` with `P0` active and `P1 -> P4` locked behind gate pass
  - Memory restart database path corrected:
    - `lifecycle.memory-restart` now restores the PostgreSQL bridge before `db.reset`
    - database tasks fail fast when `127.0.0.1:5432` is unreachable instead of printing false-success reset/migration messages
    - targeted DB/lifecycle unit coverage expanded to include bridge-unavailable and connection-guidance cases (`tests/test_db_tasks.py`, `tests/test_lifecycle_tasks.py`)
  - Team coordination and manifest delivery hardened:
    - standing `prime-architect`, `prime-development`, and `agui-design-architect` are now manifest-backed under `core/config/teams/`
    - `uv run inv team.architecture-sync` provides a Python-native NATS coordination path for central architect sync requests
    - central architect sync now uses canonical team lanes (`swarm.team.{team_id}.internal.command` -> `signal.status` / `signal.result`) instead of ad hoc direct-agent request subjects
    - live sync currently receives canonical status replies from `prime-architect` and `agui-design-architect`; `prime-development` remains the open reply-discipline gap
  - Runner contract and task verification standardized:
    - `uv run inv ...` is the supported task path for real execution
    - `uvx --from invoke inv -l` is retained only as a compatibility probe
    - bare `uvx inv ...` is treated as unsupported and checked as a negative control by `uv run inv ci.entrypoint-check`
  - NATS signal standardization codified:
    - new authority doc `docs/architecture/NATS_SIGNAL_STANDARD_V7.md` defines source kinds, subject families, payload classes, and product-vs-dev channel boundaries
    - channel architecture doc now points to the signal standard for `internal.command`, `signal.status`, `signal.result`, telemetry, and source normalization
    - repo `AGENTS.md` now enforces Python-first management scripting, the runner contract, and the infrastructure-development channel exclusion rule
  - NATS signal standardization enforced in runtime:
    - protocol constants now include `TopicTeamSignalResult` for operator-facing bounded outputs
    - `delegate_task` now publishes to team `internal.command` subjects instead of legacy direct trigger subjects
    - team `signal.status` and `signal.result` deliveries are wrapped with standardized source metadata before publish
    - runtime-created action teams default to `signal.result`; expression teams default to `signal.status`
  - Lifecycle restart reliability hardened further:
    - `lifecycle.up` now requires Core `/healthz` readiness after the port opens instead of printing a false-success ready state
    - Core startup now fails fast when the HTTP surface never becomes healthy, which makes restart and memory-restart behavior deterministic for operators
    - lifecycle unit coverage now includes the unhealthy-Core-after-port-open failure path
  - `P0` gate is now passed live:
    - `uv run inv lifecycle.memory-restart --frontend` completed successfully on 2026-03-06 after a full DB reset, stack bring-up, health sweep, and memory endpoint verification
    - `docs/architecture/NEXT_TARGET_GATED_DELIVERY_PROGRAM.md` now promotes `P1` to active
  - Database reset/migration path standardized for deterministic recovery:
    - `db.migrate` now applies only `001_init_memory.sql` plus `*.up.sql`; `.down.sql` files are excluded from the forward path
    - `psql` execution now uses `ON_ERROR_STOP=1`, so reset-time SQL failures stop the workflow instead of printing false-success output
    - `001_init_memory.sql` now enables `pg_trgm` in addition to `vector`, which resolves the prior trigram-index failures in migrations `019` and `031`
  - Core startup diagnostics and MCP bootstrap reliability hardened:
    - lifecycle background Core launches now write to `workspace/logs/core-startup.log`
    - MCP default bootstrap now uses bounded per-server connect timeouts, preventing a hung default MCP server from blocking `HTTP Server listening on :8081`
  - Persistent manifestation storage standardized for Kubernetes:
    - filesystem MCP bootstrap now resolves from `MYCELIS_WORKSPACE` instead of incorrectly reusing `DATA_DIR`
    - Helm now mounts the data PVC at `/data` and sets `MYCELIS_WORKSPACE=/data/workspace` plus `DATA_DIR=/data/artifacts`
    - Core prepares workspace and artifact directories on startup so manifested material lands on mounted storage
  - Soma/Council chat failure path hardened:
    - conversation turn persistence now casts `run_id` and `session_id` to UUID correctly for standing-team chat turns
    - direct draft requests such as simple letters/messages now prefer in-chat text responses over unnecessary tool execution
    - admin prompt contract now explicitly forbids tool use for plain drafting unless the user requested file/system/delegation actions
  - Lifecycle teardown hardened for local testing:
    - shutdown cleanup helpers now use bounded subprocess timeouts so hung `taskkill` / `kubectl` cleanup cannot stall `uv run inv lifecycle.down`
    - `lifecycle.down` now waits for Core and Frontend ports to close before reporting completion
  - UI signal normalization aligned to the standard:
    - shared frontend normalization now accepts both legacy SSE signals and standardized signal envelopes
    - signal detail surfaces now expose source kind, payload kind, source channel, and run/team/agent metadata when present
    - dashboard signal context and main store stream ingestion now normalize at the boundary instead of assuming raw payload shape
  - UI delivery targeting tightened:
    - new authority doc `docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md` defines required terminal UI states, backend transaction expectations, and failure/recovery rules for execution-facing UI
    - frontend and testing docs now require UI work to prove user-visible outcome plus backend API/NATS/runtime side effect together
    - docs regression coverage now validates both top-level README links and in-app docs manifest paths
  - First execution-facing UI contract slice enforced in code:
    - Workspace / Soma chat now uses `/api/v1/chat`; direct specialist targeting uses `/api/v1/council/{member}/chat`
    - blocked Workspace chat requests now set explicit `blocker` mode and render a Soma-specific recovery card instead of generic council wording
    - focused chat tests now prove route selection and blocker-state behavior for Workspace vs direct council interactions
  - README navigation contract standardized:
    - `README.md` now exposes a structured `README TOC` near the top for development-agent navigation
    - `AGENTS.md` now requires README TOC maintenance whenever major README sections change
    - docs regression coverage now verifies README TOC anchors resolve to live headings
  - Canonical planning documentation refactored:
    - monolithic `mycelis-architecture-v7.md` is now a compatibility PRD index instead of the detailed authority
    - new modular planning library lives in `docs/architecture-library/`
    - canonical planning is now split across target-deliverable, system-architecture, execution-manifest, UI-operator-experience, and delivery-governance docs
    - README and in-app docs navigation now point future work toward the modular library instead of expanding one giant PRD file
    - `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md` now translates the modular library into the current implementation queue with scoped files plus development/testing references
  - Invoke docs/task contract refreshed:
    - canonical operator docs (`README.md`, `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, `ops/README.md`) now align to the live `uv run inv -l` surface for auth, lifecycle, CI, logging, quality, and team-sync tasks
    - user recovery docs and workflow-composer delivery planning now use `uv run inv ...` for real task execution instead of stale executable bare `uvx inv ...` examples
    - docs regression coverage now treats executable bare `uvx inv ...` lines in canonical docs as drift unless they are explicitly documented as unsupported negative controls
  - Test readiness for latest changes refreshed:
    - task/gate unit suites pass (`test_ci_tasks`, `test_auth_tasks`, `test_logging_tasks`, `test_quality_tasks`, `test_lifecycle_tasks`)
    - focused runner/task suite passes with `uv run inv ci.entrypoint-check` plus pytest task coverage
    - focused Go protocol/swarm suite passes for the signal-runtime contract
    - focused frontend signal-normalization suite and typecheck pass
    - focused lifecycle/runner/logging pytest suite passes after readiness hardening
  - UI test harness hardened for reproducible local and CI delivery proof:
    - Playwright now owns the Next.js Interface server lifecycle through `interface/playwright.config.ts`
    - `uv run inv interface.e2e` and `uv run inv test.e2e` now support focused `--project` and `--spec` execution and clear stale listeners on `:3000` before and after each run
    - `uv run inv interface.e2e --live-backend ...` now loads auth proxy env and enables real Core-backed UI specs instead of route-fulfilled-only browser coverage
    - new Playwright spec `interface/e2e/specs/workspace-live-backend.spec.ts` proves live `/api/v1/services/status` and `/api/v1/council/members` traffic through the Workspace UI proxy
    - accessibility coverage now depends on a committed `@axe-core/playwright` dev dependency instead of silently skipping when missing
    - Playwright project coverage now includes `chromium`, `firefox`, `webkit`, and `mobile-chromium`
    - `.github/workflows/e2e-ci.yaml` now installs the full browser matrix and lets Playwright manage the UI server while Core remains a separately started backend dependency

Verification evidence (latest targeted slice):
- `cd core && go test ./internal/server -run "TestHandle(CreateAndListGroups_HappyPath_DB|CreateGroup_Unauthorized|CreateGroup_ScopeDenied|CreateGroup_HighImpact_RequiresApproval|CreateGroup_InvalidWorkMode|UpdateGroup_NotFound|GroupBroadcast_FanoutParallel_DB|GroupMonitor_ReturnsSnapshot)" -count=1`
- `cd core && go test ./internal/server -run "TestHandleServicesStatus_AllOffline" -count=1`
- `cd interface && npm run build`
- `cd core && go test ./internal/swarm -count=1`
- `cd core && go test ./internal/server -count=1`
- `cd interface && npx vitest run --reporter=dot`
- `cd core && go test ./internal/server -run "TestHandleMe|TestHandleUpdateSettings" -count=1`
- `cd interface && npx vitest run __tests__/lib/labels.test.ts __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx --reporter=dot`
- `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`
- `cd interface && npx tsc --noEmit`
- `uv run inv interface.e2e --project=chromium --spec=e2e/specs/accessibility.spec.ts`
- `uv run inv test.e2e --project=mobile-chromium --spec=e2e/specs/mobile.spec.ts`
- `uv run inv interface.e2e --live-backend --project=chromium --spec=e2e/specs/workspace-live-backend.spec.ts`
- `python -m py_compile ops/k8s.py`
- `cd core && go test ./internal/cognitive -count=1`
- `python -m py_compile ops/lifecycle.py`
- `$env:PYTHONPATH='.'; uv run pytest tests/test_ci_tasks.py tests/test_auth_tasks.py tests/test_logging_tasks.py tests/test_quality_tasks.py tests/test_lifecycle_tasks.py -q`
- `.\.venv\Scripts\inv.exe ci.baseline`
- `$env:PYTHONPATH='.'; uv run pytest tests/test_db_tasks.py tests/test_lifecycle_tasks.py tests/test_logging_tasks.py tests/test_quality_tasks.py -q`
- `uv run inv ci.entrypoint-check`
- `uv run inv -l`
- `uv run inv db.status`
- `uv run inv lifecycle.up --frontend`
- `uv run inv lifecycle.health`
- `cd core && go test ./internal/swarm ./pkg/protocol -count=1`
- `cd core && go test ./internal/mcp ./internal/swarm ./pkg/protocol -count=1`
- `cd interface && npx vitest run __tests__/dashboard/SignalContext.test.tsx __tests__/lib/signalNormalize.test.ts --reporter=dot`
- `cd interface && npx tsc --noEmit`
- `cd interface && npx vitest run __tests__/dashboard/MissionControlChat.test.tsx __tests__/lib/labels.test.ts --reporter=dot`
- `$env:PYTHONPATH='.'; uv run pytest tests/test_misc_tasks.py -q`
- `uv run inv team.architecture-sync`
- `$env:PYTHONPATH='.'; uv run pytest tests/test_lifecycle_tasks.py tests/test_ci_tasks.py tests/test_logging_tasks.py -q`
- `$env:PYTHONPATH='.'; uv run pytest tests/test_db_tasks.py tests/test_lifecycle_tasks.py tests/test_ci_tasks.py tests/test_logging_tasks.py -q`
- `uv run inv lifecycle.memory-restart --frontend`

Verification evidence (latest full sweep — 2026-03-03):
- `cd core && go test ./... -count=1` -> pass
- `cd interface && npm run build` -> pass
- `cd interface && npx vitest run --reporter=dot` -> pass (`55` files, `322` tests)
- `cd interface && npx playwright test --reporter=dot` -> pass (`51` passed, `4` skipped)
- worktree state during sweep: dirty (`59` entries -> `50` modified, `9` untracked)

---

## Next Parallel Sprint (A/B/C/Q)

Execution queue:
- `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md`

Parallel team tracks:
1. Team A (Core/Persistence): propagate `group_id` across runs/events/timeline lineage.
2. Team B (Governance/Auth): enforce capability bounds on dispatch and emit denial audit records.
3. Team C (Runtime/UI): wire approval continuity and group-aware operator views (timeline/filtering).
4. Team Q (QA): integration + E2E gate suite for lifecycle, denial, lineage, and monitor integrity.

### Sprint G2 (Engaged Now)

Parallel team tracks:
1. Team A (Core/Persistence): persist execution-path metadata (`path_selected`, `execution_owner`, `delivery_mode`) for run/event diagnostics.
2. Team B (Governance/Auth): enforce delivery-over-tutorial behavior and coder-first web-access routing defaults with structured denial reasons.
3. Team C (Runtime/UI): surface execution-path outcomes (direct/code-first/MCP/team) and MCP translation outcomes in Workspace/timeline surfaces.
4. Team Q (QA): add regression/E2E gates for the schema-only failure mode, coder-first web execution, and MCP translation fallback correctness.

---

## Documentation Discipline Gate (Mandatory)

Every merged slice must update:
1. `README.md` (operator/runtime behavior changes)
2. `V7_DEV_STATE.md` (checkpoint + evidence + next actions)
3. `docs/V7_IMPLEMENTATION_PLAN.md` (roadmap/dependency changes)
4. `interface/lib/docsManifest.ts` (for every new authoritative doc)

No promotion without this gate.

---

## What Is Done

### UI Delivery Governance (Canonicalized)

Parallel UI lane boards and superseded execution playbooks have been purged from the canonical surface.

Current authoritative UI delivery references:

| Deliverable | File |
|------------|------|
| Canonical operator journeys and anti-information-swarm rules | `docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md` |
| Required UI terminal states and backend transaction rules | `docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md` |
| Current implementation queue and acceptance logic | `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md` |
| Delivery proof, testing structure, and evidence rules | `docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md` |
| Frontend implementation reference | `docs/architecture/FRONTEND.md` |
| Implementation roadmap/state sync | `docs/V7_IMPLEMENTATION_PLAN.md` |

Execution policy:
- Lanes run in parallel.
- Merges are gate-sequenced (`P0 -> P1 -> P2 -> RC`).
- Lane Q evidence is required at every gate.
- Team ownership model: Circuit (Lane A), Atlas (Lane B), Helios + Argus (Lane C), Argus (Lane D), Sentinel (Lane Q).

Gate A baseline progress:
- Added/updated UI tests for `StatusDrawer`, `DegradedModeBanner`, `CouncilCallErrorCard`, `MissionControlChat` error UX, `ShellLayout` status action wiring, and Automations landing expectations.
- Verified targeted gate suite: `50` tests passing across dashboard/pages/shell/teams/automations subsets.
- Added Playwright Gate A operational UX suite scaffold:
  - `interface/e2e/specs/v7-operational-ux.spec.ts` (6 scenarios: degraded banner lifecycle, status drawer access, council reroute via Soma, automations actionable hub, system quick checks, focus mode toggle).
  - Verified green locally in current environment (`cd interface && npx playwright test --reporter=dot` -> `51` passed, `4` skipped).
- Framework hardening completed:
  - canonical UI guidance now lives in the architecture library and transaction contract docs rather than a separate planning layer.
- Recovery UX hardening completed:
  - `DegradedModeBanner` Retry now re-checks services, refreshes council/missions, and re-initializes SSE when disconnected.
  - Workspace council direct-target indicator now synchronizes with global `councilTarget` (including one-click fallback to Soma from banner/error card).
  - `StatusDrawer` council failure highlighting now resolves from recent failed council message source instead of only current target selection.

### Workflow Onboarding + Bus Planning (Canonicalized)

Execution-grade planning for workflow onboarding and bus exposure is now tracked through:
- `docs/architecture-library/TARGET_DELIVERABLE_V7.md`
- `docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md`
- `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md`
- `docs/architecture/NATS_SIGNAL_STANDARD_V7.md`
- parallel Sprint 0 and Sprint 1 work packages across Atlas/Helios/Circuit/Argus/Sentinel

Planning source:
- `docs/architecture-library/TARGET_DELIVERABLE_V7.md`
- `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md`

### Sprint 0 UI Implementation (Lane C Started)

Implemented Sprint 0 scaffolds for guided team instantiation and readiness gating.

| Deliverable | File |
|------------|------|
| Team instantiation contract models (`TeamProfileTemplate`, `ReadinessSnapshot`, I/O envelope) | `interface/lib/workflowContracts.ts` |
| Capability readiness gate UI | `interface/components/automations/CapabilityReadinessGateCard.tsx` |
| Guided wizard scaffold (Objective -> Profile -> Readiness -> Launch) | `interface/components/automations/TeamInstantiationWizard.tsx` |
| NATS route exposure picker (Basic/Guided/Expert + rollback) | `interface/components/automations/RouteTemplatePicker.tsx` |
| Wizard integration into Automations hub | `interface/components/automations/AutomationHub.tsx` |
| Wizard tests | `interface/__tests__/automations/TeamInstantiationWizard.test.tsx` |
| Route template tests | `interface/__tests__/automations/RouteTemplatePicker.test.tsx` |

Verification run:
- `cd interface && npx vitest run __tests__/automations/TeamInstantiationWizard.test.tsx __tests__/automations/RouteTemplatePicker.test.tsx __tests__/pages/AutomationsPage.test.tsx __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/shell/ShellLayout.test.tsx __tests__/teams/TeamsPage.test.tsx __tests__/pages/SystemPage.test.tsx` -> pass (50 tests)
- `cd interface && npm run build` -> pass
- `cd core && go test ./internal/mcp/ -count=1` -> pass
- `cd core && go test ./internal/server/ -run TestHandleMCP -count=1` -> pass
- `cd core && go test ./internal/swarm/ -run TestScoped -count=1` -> pass

### Resource API Standardization + Workspace Explorer (Current)

| Deliverable | File |
|------------|------|
| Shared API contract helpers (`extractApiData`, `extractApiError`, MCP result formatter) | `interface/lib/apiContracts.ts` |
| Store-level standardized parsing for MCP + services status endpoints | `interface/store/useCortexStore.ts` |
| Global services-status polling moved into shell/store path | `interface/components/shell/ShellLayout.tsx` |
| Degraded banner/status drawer/quick checks now consume shared store status | `interface/components/dashboard/DegradedModeBanner.tsx`, `interface/components/dashboard/StatusDrawer.tsx`, `interface/components/system/SystemQuickChecks.tsx` |
| Services tab now uses centralized status contract | `interface/app/(app)/system/page.tsx` |
| Workspace Explorer activated in Resources tab (filesystem MCP tool calls) | `interface/components/resources/WorkspaceExplorer.tsx`, `interface/app/(app)/resources/page.tsx` |
| Focus mode NATS status now sourced from global services status state | `interface/components/dashboard/MissionControl.tsx` |

Verification run:
- `cd interface && npm run build` -> pass
- `cd interface && npx vitest run __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/shell/ShellLayout.test.tsx __tests__/pages/ResourcesPage.test.tsx __tests__/pages/SystemPage.test.tsx` -> pass (18 tests)
- `cd interface && npx vitest run __tests__/store/useCortexStore.test.ts` -> pass (25 tests)

### Soma Extension-of-Self Architecture (Canonicalized)

Detailed architecture and delivery planning is now locked for extension-of-self growth beyond baseline MCP operations.

| Deliverable | File |
|------------|------|
| V7 architecture update with extension-of-self modes, local Ollama contract, and parallel tracks | `docs/architecture-library/TARGET_DELIVERABLE_V7.md`, `docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md` |
| Soma symbiote architecture expanded with explicit local Ollama communication contract | `docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md` |
| Inception "harsh-truth" guardrails codified (heartbeat budgets, single-use-case ratchet, staged self-update, sandbox-first skill injection) | `docs/architecture-library/TARGET_DELIVERABLE_V7.md`, `docs/architecture/SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md` |
| Integrated implementation-plan section for parallel package execution | `docs/V7_IMPLEMENTATION_PLAN.md` (Section 9) |

### Phase 19 Foundation

| Capability | Notes |
|-----------|-------|
| Agent + Provider lifecycle | Soma, Axon, Swarm fully wired |
| Conversation memory (pgvector) | Migrations 001-022 applied |
| Intent proof pipeline | Confirm token flow + CE-1 templates |
| Mutation detection in chat | `mutationTools` set in cognitive.go |
| Brains toggle persistence | SaveConfig() persists to cognitive.yaml |
| Auth middleware | `interface/.env.local`, header injected on all proxied requests |

---

### V7 Step 01 — Navigation (Team D)

Collapsed 12+ architecture-surface routes into 5 workflow-first panels.

| Deliverable | File |
|------------|------|
| ZoneA_Rail (5 items + Docs + Advanced toggle) | `interface/components/shell/ZoneA_Rail.tsx` |
| Automations page (6 tabs + deep-link + advanced gate) | `interface/app/(app)/automations/page.tsx` |
| Resources page (4 tabs + deep-link) | `interface/app/(app)/resources/page.tsx` |
| System page (5 tabs + advanced gate) | `interface/app/(app)/system/page.tsx` |
| DegradedState shared component | `interface/components/shared/DegradedState.tsx` |
| PolicyTab CRUD → ApprovalsTab | `interface/components/automations/ApprovalsTab.tsx` |
| 8 legacy routes → server redirects | `/wiring` `/architect` `/teams` `/catalogue` `/marketplace` `/approvals` `/telemetry` `/matrix` |
| 56 unit tests passing | `interface/__tests__/pages/`, `__tests__/shared/`, `__tests__/shell/` |

---

### V7 Team A — Event Spine

Persistent audit record for every mission action.

| Deliverable | File |
|------------|------|
| `mission_runs` table (023) | `core/migrations/023_mission_runs.up.sql` |
| `mission_events` table (024) | `core/migrations/024_mission_events.up.sql` |
| MissionEventEnvelope + EventType constants | `core/pkg/protocol/events.go` |
| events.Store (Emit, GetRunTimeline) | `core/internal/events/store.go` |
| runs.Manager (CreateRun, UpdateRunStatus, ListRecentRuns) | `core/internal/runs/manager.go` |
| GET /api/v1/runs/{id}/events | `core/internal/server/runs.go` |
| GET /api/v1/runs/{id}/chain | `core/internal/server/runs.go` |
| GET /api/v1/runs (global list) | `core/internal/server/runs.go` |
| TypeScript types (MissionRun, MissionEvent) | `interface/types/events.ts` |

---

### V7 Soma Workflow E2E

Complete loop from chat → consultation trace → proposal → confirm → run_id → timeline.

**Backend changes:**

| Deliverable | File |
|------------|------|
| ConsultationEntry type in ChatResponsePayload | `core/pkg/protocol/envelopes.go` |
| ReAct loop captures consult_council into ProcessResult.Consultations | `core/internal/swarm/agent.go` |
| agentResult.Consultations wired into chatPayload | `core/internal/server/cognitive.go` |
| HandleConfirmAction returns run_id in response | `core/internal/server/templates.go` |

**Frontend — Chat UI:**

| Deliverable | File |
|------------|------|
| Soma-locked header (no dropdown) | `interface/components/dashboard/MissionControlChat.tsx` |
| DirectCouncilButton (⚡ Direct popover) | `interface/components/dashboard/MissionControlChat.tsx` |
| DelegationTrace council cards below response | `interface/components/dashboard/MissionControlChat.tsx` |
| SomaActivityIndicator (live tool.invoked labels) | `interface/components/dashboard/MissionControlChat.tsx` |
| System message bubble → /runs/{id} pill | `interface/components/dashboard/MissionControlChat.tsx` |
| LaunchCrewModal always targets Soma, clears stale proposals | `interface/components/workspace/LaunchCrewModal.tsx` |

**Frontend — Runs UI:**

| Deliverable | File |
|------------|------|
| RunTimeline.tsx (auto-poll 5s, stops on terminal events) | `interface/components/runs/RunTimeline.tsx` |
| EventCard.tsx (colored dots, expandable payload) | `interface/components/runs/EventCard.tsx` |
| /runs/[id] page | `interface/app/(app)/runs/[id]/page.tsx` |
| /runs list page | `interface/app/(app)/runs/page.tsx` |

**Frontend — Store:**

| Deliverable | Notes |
|------------|-------|
| activeRunId, runTimeline, recentRuns state | Zustand slices in useCortexStore.ts |
| confirmProposal injects system message with run_id | Replaces old stub |
| fetchRunTimeline, fetchRecentRuns actions | Poll-ready |

**Frontend — OpsOverview:**

| Deliverable | File |
|------------|------|
| OpsWidget Registry (registerOpsWidget / getOpsWidgets) | `interface/lib/opsWidgetRegistry.ts` |
| RecentRunsSection widget (order 60, fullWidth) | `interface/components/dashboard/OpsOverview.tsx` |
| OpsOverview renders all 6 widgets from registry | `interface/components/dashboard/OpsOverview.tsx` |

---

### In-App Docs Browser

Fully functional. `/docs` page with sidebar, search, and rendered markdown.

| Deliverable | File |
|------------|------|
| GET /docs-api (manifest) | `interface/app/docs-api/route.ts` |
| GET /docs-api/[slug] (content, path-validated) | `interface/app/docs-api/[slug]/route.ts` |
| docsManifest.ts (31 entries, 7 sections) | `interface/lib/docsManifest.ts` |
| /docs page (sidebar + react-markdown, deep-link) | `interface/app/(app)/docs/page.tsx` |
| Internal .md link resolution (stays in-app) | `interface/app/(app)/docs/page.tsx` |
| Docs nav item in main rail (below Memory) | `interface/components/shell/ZoneA_Rail.tsx` |

**User guide docs (docs/user/):**

| Doc | Covers |
|-----|--------|
| core-concepts.md | Soma, Council, Mission, Run, Brain, Event, Trust, NATS, MCP |
| soma-chat.md | Message → delegation trace → proposal → confirm → run link |
| meta-agent-blueprint.md | Architect as meta-agent, blueprint structure, activation pipeline |
| run-timeline.md | Event types, colors, common patterns |
| automations.md | Triggers, schedules, approvals, teams, policy |
| resources.md | Brains, Cognitive Matrix, MCP tools, workspace, catalogue |
| memory.md | Semantic search, SitReps, artifacts, hot/warm/cold tiers |
| governance-trust.md | Trust scores, halts, approval flow, policy config |

---

### Provider CRUD, Mission Profiles & Reactive Subscriptions

Full hot-reload provider management, named workflow profiles with role→provider routing, context snapshot/restore, reactive NATS subscriptions, and service health dashboard.

**Migrations:**

| Migration | Table | File |
|-----------|-------|------|
| 028 | `context_snapshots` | `core/migrations/028_context_snapshots.up.sql` |
| 029 | `mission_profiles` | `core/migrations/029_mission_profiles.up.sql` |

**Backend:**

| Deliverable | File |
|------------|------|
| `AddProvider` / `UpdateProvider` / `RemoveProvider` with `RWMutex` | `core/internal/cognitive/router.go` |
| `POST /api/v1/brains` (add), `PUT /api/v1/brains/{id}` (update), `DELETE /api/v1/brains/{id}` (delete), `POST /api/v1/brains/{id}/probe` | `core/internal/server/brains.go` |
| Context snapshot CRUD — `POST /api/v1/context/snapshot`, `GET /api/v1/context/snapshots`, `GET /api/v1/context/snapshots/{id}` | `core/internal/server/context.go` |
| Mission profile CRUD + activate — `GET/POST/PUT/DELETE /api/v1/mission-profiles`, `POST /api/v1/mission-profiles/{id}/activate` | `core/internal/server/profiles.go` |
| Reactive NATS subscription engine — `Subscribe`, `Unsubscribe`, `ReactivateFromDB`, `Connected`, `ActiveSubscriptionCount` | `core/internal/reactive/engine.go` |
| `GET /api/v1/services/status` — NATS, PostgreSQL, Cognitive, Reactive health aggregation | `core/internal/server/services.go` |
| `MaxReconnects(-1)` unlimited NATS reconnects + ping interval | `core/internal/transport/nats/client.go` |
| DB startup retry loop (45×2s), NATS startup retry (45×2s), connection pool config, reactive reactivation on boot | `core/cmd/server/main.go` |
| Longer port-forward wait (30s), Core API wait (120s), WARN instead of FATAL for infra slow-start | `ops/lifecycle.py` |

**Frontend:**

| Deliverable | File |
|------------|------|
| Provider Add/Edit/Delete/Probe UI with type presets (Ollama, vLLM, LM Studio, OpenAI, Anthropic, Google, Custom) | `interface/components/settings/BrainsPage.tsx` |
| ContextSwitchModal — Cache & Transfer / Start Fresh / Load Snapshot strategies | `interface/components/settings/ContextSwitchModal.tsx` |
| MissionProfilesPage — role→provider table, NATS subscriptions editor, context strategy, auto-start | `interface/components/settings/MissionProfilesPage.tsx` |
| Profiles tab in Settings | `interface/app/(app)/settings/page.tsx` |
| Services tab in System — live polling, service cards, restart command reference with copy | `interface/app/(app)/system/page.tsx` |
| Mission profile + context snapshot types, state, and all async actions | `interface/store/useCortexStore.ts` |

---

### V7 Team B — Trigger Engine

Declarative IF/THEN rules evaluated on CTS event ingest. Four guards: cooldown, recursion depth, concurrency, condition (reserved). Default mode `propose` — auto-execute requires explicit policy.

**Migrations:**

| Migration | Table | File |
|-----------|-------|------|
| 025 | `trigger_rules` | `core/migrations/025_trigger_rules.up.sql` |
| 026 | `trigger_executions` | `core/migrations/026_trigger_executions.up.sql` |

**Backend:**

| Deliverable | File |
|------------|------|
| TriggerRule + TriggerExecution types, in-memory cache, CRUD, LogExecution, ActiveCount | `core/internal/triggers/store.go` |
| Engine — CTS subscription, 4-guard evaluateRule, fireTrigger (child run), proposeTrigger | `core/internal/triggers/engine.go` |
| 6 HTTP handlers — List, Create, Update, Delete, Toggle, History | `core/internal/server/triggers.go` |
| AdminServer wiring — Triggers + TriggerEngine fields, 6 routes | `core/internal/server/admin.go` |
| main.go — trigger store + engine init, graceful shutdown | `core/cmd/server/main.go` |

**Frontend:**

| Deliverable | File |
|------------|------|
| TriggerRulesTab — full CRUD UI, RuleCard, CreateRuleForm, guard badges, mode warnings | `interface/components/automations/TriggerRulesTab.tsx` |
| Trigger types + state + 5 async actions (fetch, create, update, delete, toggle) | `interface/store/useCortexStore.ts` |
| Automations → Triggers tab now renders live TriggerRulesTab (was DegradedState) | `interface/app/(app)/automations/page.tsx` |

**Bug fixes (pre-existing, discovered during build verification):**

| Fix | File |
|-----|------|
| Added `"use client"` + `use(params)` for Next.js 15+ async params | `interface/app/(app)/runs/[id]/page.tsx` |
| Wrapped `useSearchParams()` in Suspense boundary | `interface/app/(app)/docs/page.tsx` |

---

### MCP Test Hardening (Service + Handlers)

Comprehensive MCP coverage added across service, adapter, and HTTP handler layers.

| Deliverable | File |
|------------|------|
| Library tests (YAML load, lookup, config conversion) | `core/internal/mcp/library_test.go` |
| Service tests (Install/List/Get/Delete/UpdateStatus/CacheTools/ListTools/ListAllTools/Find*) | `core/internal/mcp/service_test.go` |
| Tool set service tests (CRUD, FindByName, ResolveRefs, nil DB guards) | `core/internal/mcp/toolsets_test.go` |
| Executor adapter tests (`FindToolByName`, `CallTool`, text formatting edge cases) | `core/internal/mcp/executor_test.go` |
| MCP handler DB-backed happy paths (`/servers`, `/tools`, `/library/install`) + raw install forbidden route | `core/internal/server/mcp_test.go` |
| Tool set handler update-path tests (happy, bad UUID, missing name, nil service) | `core/internal/server/mcp_toolsets_test.go` |
| Update not-found HTTP semantics for tool set update | `core/internal/server/mcp_toolsets.go` (`404` when tool set not found) |

---

### Swarm Parallel Activation

Blueprint activation now fans out team startup in parallel where safe, with idempotent race-safe insertion into active runtime state.

| Deliverable | File |
|------------|------|
| Bounded parallel team startup in `ActivateBlueprint` | `core/internal/swarm/activation.go` |
| Idempotent duplicate-skip under concurrent activation | `core/internal/swarm/activation.go` |
| New tests: idempotent repeat activation + concurrent activation no-dup | `core/internal/swarm/activation_test.go` |

---

## What Is Pending

### V7 Team C — Scheduler (NEXT)

**Depends on:** runs.Manager, events.Store.

| File to Create | Purpose |
|---------------|---------|
| `core/migrations/027_scheduled_missions.up.sql` | scheduled_missions table with next_run_at index |
| `core/internal/scheduler/scheduler.go` | Goroutine ticker, Suspend/Resume, checkDue() |
| `core/internal/server/schedules.go` | GET/POST/PUT/DELETE /api/v1/schedules + pause/resume |

**Done when:** Schedule creates a run on tick, enforces max_active_runs, suspends on NATS disconnect.

---

### Causal Chain UI (after B+C)

| File to Create | Purpose |
|---------------|---------|
| `interface/components/runs/ViewChain.tsx` | Parent → event → trigger → child run traversal |
| `interface/components/runs/RunChainNode.tsx` | Recursive tree node |
| `interface/app/(app)/runs/[id]/chain/page.tsx` | /runs/{id}/chain route |

Backend handler `GET /api/v1/runs/{id}/chain` is already live.

---

### MCP Baseline (Parallel — Independent of B/C)

Per `docs/V7_MCP_BASELINE.md`.

| Server | Status |
|--------|--------|
| `filesystem` MCP | BOOTSTRAP DEFAULT (auto-install/connect path present) |
| `fetch` MCP | BOOTSTRAP DEFAULT (auto-install/connect path present) |
| `memory` MCP | CURATED LIBRARY INSTALL (available, not bootstrap default) |
| `artifact-renderer` MCP | PLANNED (not bootstrap default yet) |

Resources → Workspace Explorer tab shows DegradedState until implemented.

---

### Soma Extension-of-Self Runtime (Parallel with Team C, gated by contracts)

| Track | Status |
|------|--------|
| Decision frame contracts (`direct|manifest_team|propose|scheduled_repeat`) | IN PROGRESS (P0 contract types + API bundle scaffolded) |
| Local Ollama readiness and fallback status contract | IN PROGRESS (`GET /api/v1/services/status` now emits `ollama` status row) |
| Universal action adapter parity (MCP + non-MCP pilot) | PLANNED |
| Repeat-promotion policy (`one-off` -> `scheduled`) | PLANNED |
| Host/hardware governed actuation scaffolds | PLANNED |

P0 backend slice delivered on 2026-03-01:
- Added frozen protocol contract bundle in `core/pkg/protocol/inception_contracts.go`:
  - `SomaDecisionFrame`
  - `HeartbeatBudget`
  - `UniversalInvokeResult`
  - allowed decision paths and team lifetime enums.
- Added API endpoint:
  - `GET /api/v1/inception/contracts` (served by `HandleInceptionContracts`).
- Registered route in `core/internal/server/admin.go`.
- Extended services health contract:
  - `GET /api/v1/services/status` now includes explicit `ollama` service readiness semantics.
- Test coverage added/updated:
  - `core/internal/server/inception_test.go` (`TestHandleInceptionContracts_HappyPath`)
  - `core/internal/server/provider_crud_test.go` service status suite extended for `ollama`.
- Verification:
  - `cd core && go test ./pkg/protocol -count=1` -> pass
  - `cd core && go test ./internal/server -count=1` -> pass
  - `cd core && go test ./... -count=1` -> pass

---

## Current Navigation Structure

```
ZoneA_Rail
├── [logo] → /
├── Workspace     (Home)      → /dashboard
├── Automations   (Workflow)  → /automations
├── Resources     (FolderCog) → /resources
├── Memory        (Brain)     → /memory
├── Docs          (BookOpen)  → /docs
├── [System       (Activity)  → /system  ← Advanced Mode only]
└── Footer
    ├── Advanced toggle (Eye/EyeOff)
    └── Settings  (Settings)  → /settings
```

---

## Current Tab Map

| Page | Tabs | Notes |
|------|------|-------|
| `/automations` | Active · Drafts · Triggers · Approvals · Teams · Wiring* | *Wiring = Advanced Mode only |
| `/resources` | Brains · Tools · Workspace · Catalogue | |
| `/system` | Health · NATS · Database · Matrix · Debug · Services | Advanced Mode gated |

**Tabs still showing DegradedState (real data pending):**

| Page | Tab | Blocked by |
|------|-----|------------|
| Automations | Active Automations | Team C (Scheduler) |
| System | Event Health | Team A live data wiring |

---

## Current Build State

```
next build:         PASSES (all routes)
vitest:             PASSES (`55` files, `322` tests)
playwright:         PASSES (`51` passed, `4` skipped)
Go build:           go build ./... PASSES (includes triggers package)
Go tests:           `go test ./... -count=1` PASSES
Go runtime:         `go1.25.6 windows/amd64` (local sweep); docs lock target remains `1.26`
                    Migrations 023-029 must be applied for full test coverage
Go test packages:   internal/server (157), internal/events (16), internal/runs (19), internal/triggers (new), others (~80)
MCP verification:   `go test ./internal/mcp/ -count=1` PASSES
                    `go test ./internal/server/ -run TestHandleMCP -count=1` PASSES
                    `go test ./internal/swarm/ -run TestScoped -count=1` PASSES
TypeScript check:   `cd interface && npx tsc --noEmit` currently fails on pre-existing test typing gaps:
                    `TelemetryRow.test.tsx` missing `afterEach`, and typed-mock issues in
                    `MemoryPage.test.tsx`, `PrimaryRoutes.test.tsx`, `ResourcesPage.test.tsx`
```

---

## Architecture Debt (Known Gaps)

| Gap | Location | Priority |
|-----|---------|---------|
| 2 pre-existing DashboardPage test failures | `__tests__/pages/DashboardPage.test.tsx` | Low (pre-V7) |
| 14 pre-existing test file transform errors | Various `__tests__/{workspace,dashboard,teams,...}` | Low (pre-V7) |
| Causal Chain UI | ViewChain.tsx + /runs/[id]/chain | After Team C |
| Automations → Active Automations: DegradedState | `app/(app)/automations/page.tsx` | Blocked by Team C |

---

## Next Potential Steps

1. Finish Team C scheduler runtime (critical path):
   - Complete recurring schedule execution, NATS-offline suspension, and pause/resume API behavior.
2. Lock extension-of-self contract freeze (Sprint 0):
   - Freeze decision frame DTOs, universal action envelope, and local Ollama readiness fields.
3. Deliver first extension-of-self vertical slice (Sprint 1):
   - Ship direct-vs-team decision path + Workspace decision trace + run-linked evidence.
4. Raise test depth for execution channels:
   - Add adapter contract tests (MCP + one non-MCP pilot) and replay/idempotency side-effect protections.
5. Expand governed capability surfaces:
   - Add repeat-promotion path (`one-off` to scheduled) and prepare host/hardware proposal flow scaffolds with approval gates.

## Decision Log (Locked)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Trigger default mode | `propose` | Human-in-the-loop first posture |
| run_id requirement | Mandatory on all events | A mission is a definition; a run is an execution |
| Event persist order | DB first, then CTS publish | NATS offline = degrade, not data loss |
| Scheduler resolution | 1-minute | Acceptable for V7; sub-minute is Phase 20 |
| Multi-tenant schema | `tenant_id = 'default'` everywhere | Schema-ready, operationally single-tenant |
| Advanced mode gate | System nav + Neural Wiring tab | Reduces cognitive overload for normal users |
| Legacy route strategy | Server-side `redirect()` | Runs at server level, no client state needed |
| Tab deep-linking | `?tab=` URL search params + `<Suspense>` | Next.js 16 requirement for `useSearchParams()` |
| Docs API prefix | `/docs-api/` not `/api/docs/` | `/api/*` → Go proxy rewrite would intercept |
| Docs params | `await params` in route handler | Next.js 15+ async params requirement |
| Nav item order | Docs below Memory in main nav | Workflow items stay together; docs is reference |


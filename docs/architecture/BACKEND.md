# Mycelis Backend Implementation Contract
> Navigation: [Project README](../../README.md) | [Canonical PRD](../architecture-library/MYCELIS_CANONICAL_PRD.md) | [Frontend](FRONTEND.md) | [Operations](OPERATIONS.md)

This file is the scoped backend implementation contract. Product, UX, and release authority remains in the Canonical PRD.

## I. Package Structure

### Entry Points (`cmd/`)

Core service entrypoints live under `core/cmd/**`. The primary server owns HTTP startup, dependency wiring, graceful shutdown, and runtime config loading.

### Public API (`pkg/`)

Public packages expose reusable contracts and helpers. Keep product runtime behavior in Go-owned backend modules, not Python task code or UI-only state.

### Private Implementation (`internal/` - packages)

Private packages own API handlers, persistence, cognitive routing, governance, MCP integration, NATS orchestration, memory, worker execution backends, and service glue. `internal/workers` defines the normalized worker-library interface, the central default backend, and the framework-neutral `framework_runs` client boundary. Framework-specific code stays outside Core behind that Runs API and cannot acquire Outcome authority.

### Go Dependencies (Direct)

Dependency changes should be intentional, reflected in `go.mod`/`go.sum`, and validated through `uv run inv core.test` plus the relevant build or runtime proof.

## II. Swarm Orchestration

### Soma (Executive Cell) - `swarm/soma.go`

Soma is the operator-facing orchestrator. It maps intent to answers, proposals, team activity, and retained outputs while respecting governance.

### Axon (Messenger) - `swarm/axon.go`

Axon routes signals through canonical NATS subjects and event envelopes.

### Agent (LLM Reasoning Node) - `swarm/agent.go`

Agents execute role-scoped reasoning and tool loops under provider, capability, memory, and policy constraints.

### SensorAgent (Poll-Based) - `swarm/sensor_agent.go`

Sensor agents ingest external or device-like signals. They must keep device/feed origin explicit before normalization into operator-facing channels.

### Team - `swarm/team.go`

Teams coordinate agents for scoped work. Default team shaping should stay compact and reviewable.

### Internal Tool Registry - `swarm/internal_tools.go`

Internal tools are governed runtime capabilities. Tool metadata must identify source, scope, and intended consumer. Pitfalls: a handler that cannot do its work returns `error`, never failure text with a nil error, so the tool loop emits `tool.failed` and feeds the model `Tool <name> failed: ...` (storage, memory, search, catalogue, mission, team-list and image-store failures all follow this, and a `web_search` whose provider returns `status=blocked` is a `*WebSearchBlockedError` carrying the blocker code and next action); `mcp.ToolExecutorAdapter` turns an MCP `isError` result into `*mcp.ToolResultError` the same way, with the server text capped at 2 KB and obvious credentials redacted before it reaches events, the Exchange or HTTP; `delegate_task` uses a core NATS publish with no ack, so it reports the task as queued, not delivered; `write_file` writes only the `project-package.json` manifest for a package and never synthesizes `README.md`/`PROOF.md`; there is no runtime package fallback, so inference failure is the `provider_inference_failed` blocker and a contract left short is `result_contract_unsatisfied`.

### Composite And Scoped Tool Executors - `swarm/tool_executor.go`, `swarm/tool_scope.go`

Composite execution must preserve bounded outputs, error normalization, and auditability. The composite is unscoped; every agent gets it only through `ScopedToolExecutor` (`Team.startLocked`, and `NewAgent` wraps any other executor), which enforces the declared `Tools` list on both `FindToolByName` and `CallTool`: names are trimmed then matched exactly and case-sensitively, `mcp:<server>/<tool>` and `mcp:<server>/*` grant MCP tools, a bare declared name may also match an installed MCP tool of that exact name, and `toolset:<name>` expands through the MCP tool-set registry (unresolved grants nothing). An empty or absent list grants nothing and skips the tool loop. The only base toolset is runtime-owned (`ToolInvocationContext.RuntimeOwned`): `consult_council` for council preflight and `read_file` for entrypoint readback; runtime-owned `write_file` still needs a declaration. Pitfall: council preflight runs only after the scope check passes, so an undeclared `local_command`/`create_team`/`delegate_task` never reaches the council. `CallTool` hands handlers the caller scope, and `create_team` refuses members whose tools fall outside it (`childToolsWithinCaller`). Echoed tool names are capped at 128 runes. Proposal-planning capture uses the same check. Runtime-context lines that cite a registered tool the agent did not declare are dropped (`withoutUndeclaredToolLines`), and `TestShippedAgentPromptToolsAreDeclared` requires every tool a shipped YAML prompt cites (and, for Soma's admin, the full lead protocol) to be declared. A denied call does not execute, feeds the model `Tool '<name>' is not permitted for this agent and was not executed. Use only your declared tools, or answer directly.` (internally `ToolNotPermittedError`: `tool "<name>" is not permitted for this agent`), and emits a `protocol.EventToolDenied` (`tool.denied`) mission event (`phase`: lookup, execute, or planning). Known gap (S7b): approved-plan execution (`cognitive_tool_plan.go`, `server/templates_execution.go`) and the operator MCP call route are not yet scoped to the originating agent's declared tools.

### Blueprint Activation - `swarm/activation.go`, `converter.go`, `seeds.go`

Blueprint activation turns approved intent into teams, agents, events, and persisted run state.

## III. Cognitive Layer

### Router - `cognitive/router.go`

The router resolves provider policy and profile routing into a concrete model call path. Deployment/env overrides configure endpoints and profiles; they do not replace instantiated organization truth.

### Adapters

Adapters normalize provider-specific request/response behavior for supported model backends.

### Optional Model Gateway

LiteLLM is an optional external model-transport gateway behind the existing OpenAI-compatible adapter; its Python SDK does not run inside Go Core. Core remains authoritative for provider eligibility, profile/team routing, local-versus-remote data boundaries, semantic and Outcome budgets, approval, audit, correlation, and proof. `model_gateway: true` marks this infrastructure boundary, requires an explicit `local_only` or `leaves_org` classification, and disables legacy cross-provider self-recovery after a gateway failure. A gateway may translate provider APIs, hold deployment-owned upstream credentials, enforce operational RPM/TPM or spend ceilings, and retry or load-balance only among deployments inside the same Core-approved boundary. Gateway aliases must never turn a `local_only` selection into remote inference. Gateway telemetry is evidence to reconcile with Core, not completion authority.

The disabled `litellm` provider entry is a conformance and configuration boundary, not a bundled service. Authoritative swarm run/team/agent scope becomes a keyed pseudonymous OpenAI `user` value for gateway calls; raw identifiers are not sent, and ordinary providers omit this field. The operator preflight checks a separately operated proxy's liveness, readiness, scoped-key authentication, and exact alias without sending a completion. Current inference responses retain the configured Mycelis provider ID and pass through the upstream-reported model, opaque response-body id, prompt tokens, completion tokens, and aggregate total when present. The response-body id is response provenance—not gateway request/trace identity or spend-log correlation—and the accounting fields do not expose cost or make Outcome proof. Before production enablement, deployment must still prove all required call-path correlation, secret isolation, prompt/completion and cost reconciliation, boundary-partitioned fallbacks, bounded retry behavior, redacted errors, and reviewed logging, callback, cache, persistence, and tenancy posture. Browser clients and framework workers never receive proxy administration or upstream provider secrets. See the [official LiteLLM documentation](https://docs.litellm.ai/) and [source repository](https://github.com/BerriAI/litellm).

### Discovery - `cognitive/discovery.go`

Discovery reports provider availability and health without leaking secrets.

### Architect - `cognitive/architect.go`

Architect behavior decomposes intent into structured plans, teams, and governed proposals.

## IV. Execution Pipelines

### Pipeline 1: Intent -> Blueprint -> Activation

User intent enters through API/UI, is normalized, may become a blueprint/proposal, and activates only after policy allows it or the operator approves it.

### Pipeline 2: Council Chat (Request-Reply)

Council/member chat uses request-reply routing and returns normalized API envelopes for UI consumption.

### Pipeline 3: Agent ReAct Loop

Agents build context, call a model, parse tool/final output, execute approved tools, and publish bounded results.

### Pipeline 4: Memory Archival & Compression

Run and conversation events can be summarized, embedded, and stored for continuity. Memory promotion must stay explicit and reviewable.

### Pipeline 5: Governance & Zero-Trust Actuation

Mutating or protected actions flow through policy checks, proposals, approvals, proof envelopes, and persistent mission events.

### Pipeline 6: SSE Real-Time Streaming

Runtime events are streamed to the UI through normalized, route-safe state. High-volume telemetry must not substitute for operator status/result channels. Work handoffs that should appear in the Soma conversation use typed `thread_event` envelopes with operator-safe copy, run/work/proof targets, and source metadata; raw bus envelopes remain inspect-only.

### Pipeline 7: Optional Framework Runs

`internal/workers.WorkerBackend` exposes create, event stream, read, stop, approval, capability, and health operations. `central` remains the default and the only proven backend for confirmed Mycelis execution. `framework_runs` is an HTTP client for `/v1/runs`, `/v1/runs/{id}`, `/v1/runs/{id}/events`, `/v1/runs/{id}/stop`, `/v1/runs/{id}/approvals/{approval_id}`, `/v1/capabilities`, and `/health`. It requires the exact framework-neutral `runs_api` status, event, output, and identity vocabulary, fails closed on legacy aliases or malformed lifecycle shapes, sends credentials only after resolving a managed secret reference, and deliberately does not implement `RunFinalizer`.

Core's current-schema authority tables bind one Mycelis run to `framework_runs`/`runs_api`, retain the exact create digest and typed correlation, claim unique event id/sequence receipts, advance a contiguous cursor with version CAS, and persist Core-owned approvals and control commands. Receipt claim, existing mission projection, and cursor advancement are atomic. The external-create outbox kind is deliberately held in non-claimable `awaiting_handler` state until slice C; it has no dispatcher or network path. The Python `framework_runs` package remains a bounded, process-isolated protocol oracle: its conformance driver and memory store are non-production, and its SSE endpoint is a finite retained snapshot. All completions carry `completion_authority=candidate`, `requires_core_validation=true`, `verified=false`, and `execution_authority=mycelis_core`. The optional `LangGraphDriver` still requires an injected compiled graph. Core continues to reject external selection until every P0.10 neutral gate passes.

`GovernedControlBackend` is the strict external control surface: Core supplies `command_id`, positive `expected_version`, and actor identity, and receives an exact durable receipt whose HTTP status distinguishes a new command from replay. The receipt-free `WorkerBackend.StopRun` and `SubmitApproval` methods remain only for the current central execution interface during slices A through D; `framework_runs` stop fails closed through that legacy method. Slice E owns migrating the final callers and removing the receipt-free compatibility methods after central and external control tests pass together.

The production implementation is the separate Go module at `services/framework-runs`. It starts only with PostgreSQL and a canonical bearer token of at least 32 bytes, verifies that its database contains exactly its six owned tables and required invariants, and exposes the frozen seven-route API. Its journal makes accepted create/control commits durable before executor effects, leases commands for restart recovery, fences stale claims, and persists immutable candidate manifests. Production deliberately injects no executor in slice B: health may report database readiness, capabilities report `production_ready=false`, and create/stop/approval fail before persistence. The in-memory repository is retained only as a deterministic controller/HTTP conformance test double through the neutral gate; it is never selected by `cmd/framework-runs`, and slice F must remove or relocate it if it no longer supplies unique test value.

### Durable external-run projection target

The target projection sequence keeps one authoritative Mycelis lifecycle:

1. Confirmation commits the Mycelis approval, WorkIntent, ExecutionContract, Outcome/team-work visibility, dispatch record, and authoritative `run_id` before any external call.
2. Core sends that `run_id` unchanged to the facade with a correlation envelope containing the intent proof, execution contract, required work-item id, optional team/Outcome identifiers, idempotency key, source metadata, and graph revision. The facade must treat an identical duplicate create as the same run and reject conflicting reuse.
3. Core durably records the backend binding and event position before relying on stream progress. Event ingestion validates the backend run id and correlation envelope, claims replay-safe receipts, and projects accepted/progress/approval/terminal evidence into the existing mission-event, run, team-work, interaction, and operator-event paths. Reconnect or restart resumes from durable state and reconciles with `GET /v1/runs/{id}`; it never creates a replacement run silently.
4. An `approval_needed` event pauses projection in a Mycelis-controlled state. Only a governed Mycelis decision may call the facade approval endpoint for the matching run and approval id. Framework-local self-approval, approval-id substitution, and generic unowned resume fail closed.
5. `completed` is never direct completion authority. Output URIs and metadata are normalized as candidates, confined to the approved scope, and passed through the same retained-file, digest/readback, semantic, browser/runtime, proof, and Outcome finalization gates used by central execution. Failed validation remains degraded/recoverable; it cannot become `output_ready` because the graph or facade said completed.
6. Stop, disconnect, malformed/stale events, correlation mismatch, duplicate terminal events, and Core restart preserve one durable truth. Core owns the operator-visible recovery action and never falls back to a second execution after external side effects may have started.

LangGraph's role is intentionally worker-local: execute an injected compiled graph, keep framework-specific state inside the worker boundary, emit normalized progress or approval requests, and return candidate outputs. It does not own Mycelis run identity, capability admission, data boundary, approval, artifacts, audit, or Outcome completion. LangGraph Agent Server is not adopted in this slice because a separately operated server would introduce another run API/control plane plus deployment, authentication, persistence, tenancy, and recovery decisions before the smaller `framework_runs` projection contract is proven. A future Agent Server adapter would require an explicit owner, migration/lifetime decision, protocol mapping, and the same replay, approval, isolation, secret, and result-validation certification; it is not a compatibility alias for `framework_runs`.

LiteLLM is orthogonal. It may be an approved OpenAI-compatible inference transport used by Core or by a worker-local graph, but it never becomes the worker execution backend or receives Mycelis approval/completion authority. `worker_runtime` selects execution transport; the cognitive provider registry selects model transport. Each keeps its own scoped secret references, data-boundary policy, correlation, and production proof.

Framework posture was rechecked on 2026-09-02. The implementation order preserves the central authority/finalization reference path, completes the framework-neutral durability/control/finalization gate, and then certifies the existing LangGraph worker-local driver. CrewAI follows only as an optional compatibility/import compiler; it is not a backend name, runtime dependency, or P0.10 requirement. AG2 and Microsoft Agent Framework remain separate later evaluations. None is installed or enabled by this slice, and none may replace Mycelis' Outcome, policy, approval, audit, projection, or validation boundaries. Source ownership, deployment topology, rollout, proof, and team handoffs are fixed by the canonical PRD's [P0.10 delivery referential](../architecture-library/MYCELIS_CANONICAL_PRD.md#p010-framework-execution-delivery-referential).

## V. Data Contracts & Protocols

### 1. The CTS Envelope (Cortex Telemetry Standard)

Product signals must include enough metadata to identify source, scope, payload kind, and intended consumer. Required governed metadata includes `run_id` when execution-linked, `team_id` when team-scoped, `agent_id` when agent-scoped, `source_kind`, `source_channel`, `payload_kind`, and `timestamp`.

### 2. The ChatResponsePayload

Chat responses normalize direct answers, proposals, execution results, blocker states, consultations, tools used, and trust/governance metadata for UI rendering.

### 3. The APIResponse Envelope

HTTP responses should use the standard `{ ok, data, error }` posture with stable errors and no raw backend noise in UI-facing payloads.

### 4. The MissionBlueprint

Blueprints describe decomposed work: teams, agents, constraints, resources, governance posture, and expected outputs.

### 5. The AgentManifest

Manifests define role identity, system prompt, model/profile, tools, inputs, outputs, and verification expectations.

### 6. The ProofEnvelope

Proof envelopes carry approval, execution, evidence, policy, and audit context for governed actions.

### 7. The SitRep Schema (Archivist Output)

SitReps are bounded summaries of run or memory-relevant activity for continuity and review.

### 8. API Graceful Degradation

Handlers should return normalized degraded/blocker states when dependencies are unavailable, not panic text or provider internals.

## VI. NATS Topic Architecture

Use canonical subject constants from Go protocol/topic definitions. Do not hardcode `swarm.*` literals in runtime code.

### Global Control Plane

Global subjects are for governed broadcast/control only.

### Team Internal (per team)

Use directed team input, status, result, and telemetry families with explicit `team_id`.

### Wildcards

Wildcard subscriptions must not blur operator status with high-volume telemetry.

### Council Request-Reply

Council calls use bounded request-reply subjects for specialist/member interaction.

### Sensor Data Ingress

Sensor/IoT input must identify device/feed origin and stay separate until normalized.

Registered external input sources are the durable ingress boundary for MCP
callbacks, webhooks, local APIs, service buses, UDP/sensor feeds, hardware
devices, schedulers, file-watchers, and host probes. Producers may publish only
to registered ingress subjects. Core validates source status, scope, auth
posture, and schema; persists either an append buffer, latest-state buffer, or
append-with-latest buffer; then dispatches bounded references to team
command/status/result lanes only when a binding exists. Agents and teams do not
subscribe to raw high-volume subjects. For real-time feeds, the buffer owns
backpressure, sampling, rollups, and dropped-count proof while teams receive
summaries, latest state, anomaly refs, or approved window reads.
The current backend contract persists these registrations in
`input_sources`, stores evidence in `input_source_events`,
`input_source_latest`, and `input_source_windows`, and exposes the guarded
management/read API through `/api/v1/input-sources` and
`/api/v1/input-sources/{id}/buffer`. The live registered-input projection
subscribes to the global input ingress lane, ignores unregistered subjects,
normalizes envelope/header/raw payload metadata, and persists matched messages
into the selected append/latest/window buffer. New service/device integrations
must build on this registry rather than adding raw NATS listeners to teams.

### Mission DAG

Mission events are run-linked and persistent when tied to mutating or auditable work.

### Agent Output

Agent outputs must be bounded, typed, scoped, and safe for downstream consumers.

## VII. Database Schema

### Tables And Migrations

SQL owns schema and migration contracts. Runtime tables cover identity, organizations, runs, events, memory, artifacts, governance, teams/groups, MCP/tool activity, and configuration state.

### Migration Index

The `ORGANIZATIONS_EXTENSION` marker block in `001_current_schema.sql` (after C2A, before the single `COMMIT;`) adds the durable `organizations` table; the later `CONFIRM_TOKEN_BINDING_EXTENSION` marker (after `ROLE_SEED_RETIREMENT_EXTENSION`, before the same `COMMIT;`) adds `confirm_tokens.purpose`/`binding_digest`/`minted_by` and the purpose `CHECK` constraint (A2b).

Use migration files as the source of exact DDL truth. When API behavior or payload meaning changes, review [API Reference](../API_REFERENCE.md) and the affected migration docs/tests.

## VIII. API Surface

### Identity & Users
User, local-admin, break-glass, and future enterprise auth endpoints.

### Chat & Council
Soma, council, and member chat/proposal routes.

### Mission Orchestration
Run, mission, proposal, approval, execution, and timeline routes.

### Cognitive Engine
Provider profile, health, discovery, and routing routes.

### Telemetry & Trust
Status, trust, event, and stream routes.

### Memory & RAG
Memory, semantic search, context, and continuity routes.

### Governance & Proposals
Policy, proposal, proof, and approval routes.

### Teams
Team, group, temporary workflow, and member routes.

### MCP Management
Connected Tools registry, library, install, activity, and health routes.

### Agent Catalogue
Agent/template catalogue and manifest routes.

### Artifacts
Generated output, file, media, and retained artifact routes.

### Provisioning & Registry
Bootstrap, templates, resource registry, and deployment-context routes.

### Health
Readiness, liveness, and dependency health routes.

## Area Contracts

Fixed-key blocks for development agents. Blocks cite the PRD and source; they never restate or override it.

### Area: Code Context
- PRD: §P0.7c L370 · Scoreboard: P0.7c native code context maps
- Owned paths: `core/internal/codecontext/**` (mycelis-core-authority) · `docker-compose.yml`, `.env.compose.example`, `ops/compose_env.py`, `ops/code-context/**` (mycelis-platform-ops) · Do-not-touch: `core/internal/server/code_context.go` routes (owned elsewhere)
- Seams: `codecontext.sensitiveName`, `codecontext.pathAllowed`, `codecontext.confine`, `codecontext.resolveRoot`, `Service.rootAllowed`, `compose_env.validate_code_context_host_root`
- Invariants: one `sensitiveName` exclusion (`.git`, `.env*`, keys/certs, and names containing credential/token/secret/kubeconfig/service-account, plus rc files) applies case-insensitively on walk, subpath, explain, and symlink targets; symlinks confine via `EvalSymlinks` to the resolved root; `RegisterSource` fails closed with zero configured roots; every blocked path returns one host-path-free `path_unavailable` message; `compose_env` rejects a `HOST_ROOT` whose `.git` is a directory (primary checkout) or that has a top-level `.env*`
- Authority: `/api/v1/code-context/*` -> root admin + `code_context:read`/`write`, default deny; negatives: anon 401, standard 403; swarm `code_context.*` tools are NOT admin-gated (open finding, tracked below)
- Proof: `go test ./internal/codecontext/...`; `uv run pytest tests/test_compose_code_context_contract.py -q`; `uv run inv quality.max-lines`
- Pitfalls: the `token` name fragment over-excludes ordinary files (e.g. `templates_tokens.go`); inline secrets inside otherwise-allowed files are not redacted; the confine-then-read step has a TOCTOU window on the `:ro` mount; swarm `code_context.*` tools remain non-admin-gated pending A2's route-scope matrix follow-up
- Verified: 2026-09-25 lead merge gate: go test -race ./internal/codecontext, compose code-context contract, docs links, max-lines

### Area: ConfigDocuments
- PRD: §Bounded Discovery And Outcome Templates L22-23, L100, L102, L169-170 · Scoreboard: P0.3a
- Owned paths: `core/config/documents/templates/*.yaml` (mycelis-ai-runtime) · `core/internal/configdocuments/{builtin_guard,builtin_seed}.go`, `core/cmd/server/startup_config_documents.go` (mycelis-core-authority, S3b lease) · Do-not-touch: `core/config/templates/**` (bundle-loader family), `core/pkg/protocol/**`
- Seams: `ParseDocument`, `ValidateConfigDocument`, `CompileDocument`, `Store.StoreRevision`/`StoreRevisionTx`, `activateRevisionTx`, `guardPublicStore`, `guardPublicActivation`, `Store.SeedBuiltInRevisions`, `LoadBuiltInSeedDirectory`, `seedBuiltInConfigDocuments`, `respondConfigDocumentError`
- Invariants: only `SeedBuiltInRevisions` (actor `system:bootstrap`) inserts or activates built-in scope or source; every public store (HTTP, Soma direct, Soma confirmed) returns `metadata.reserved_built_in_scope`/`_source` (400) and every public activate/rollback of a built-in revision returns `ErrBuiltInReserved` (403) after the row lock; public callers cannot use a `system:` actor; seeding is all-or-nothing, idempotent by (id, version, digest), touches only `(built_in, '')`, and never alters operator/workspace/organization rows; preview, dry-run and compile never write
- Authority: `POST /api/v1/config-documents` and `.../activate|rollback` -> `config_documents:write`, root admin only, default deny; negatives: anon 401, standard user 403; built-in rows 400/403 for every caller
- Proof: `uv run inv core.test --package=./internal/configdocuments`; `uv run inv core.test --package=./cmd/server --run='BuiltIn|Seed'`; lead: `uv run inv lifecycle.first-boot-proof --isolated --build` (4 built-in revisions + 4 activations, second boot adds 0 rows)
- Pitfalls: the bundle loader FATALs on any stray file under `core/config/templates/`; a DB error during seeding stops Core (fail closed), and with the DB unavailable at startup, files are still validated and seeding is skipped with a WARN; Core refuses to start if any built-in row was not created by `system:bootstrap` (run the read-only precheck before redeploying a retained stack); changing a seeded file without bumping `metadata.version` is fatal; removing a file leaves its last activation (no deactivate action); an architecture test fails if `SeedBuiltInRevisions` is referenced outside `configdocuments` and `cmd/server`
- Verified: 2026-09-25 lead merge gate: isolated first boot, TestSeedBuiltInRevisionsRealDB (real PG), security-qa GO, core.test, docs links, max-lines

### Area: Governance authority
- PRD: §Bounded Discovery And Outcome Templates L100, L102, L169; Core authority L155, L223 · Scoreboard: P0.3a, A2
- Owned paths: `core/internal/governance/{guard,guard_degraded,policy,policy_posture}.go`, `core/internal/router/router.go` (Gatekeeper branch), `core/internal/state/registry.go` (`RefreshKnown`), `core/internal/server/governance*.go` (incl. `governance_approver_tier.go`), `core/internal/server/action_governance_posture.go`, `cognitive_council.go`, the `invocation/approval.go` token mint, `core/config/policy.yaml` + `charts/mycelis-core/config/policy.yaml` (identical pair) (mycelis-core-authority) · Do-not-touch: `buildApprovalPolicy` thresholds, `auth.go`, `admin_routes.go`
- Seams: `governance.ValidatePolicyConfig`, `governance.NewDegradedGuard`, `Guard.Degraded`, `Guard.ReplacePolicy(cfg, apply(previous))`, `Guard.PostureRequiresApproval`, `applyPostureApprovalFloor` then `applyApproverTier` (chat handler and council), `approverTier`, `confirmTokenMint`, `Registry.RefreshKnown`, `handleUpdatePolicy`, `resolveGuardApproval`, `requireApprover`, `confirmerMayApprove` (in `prepareConfirmedAction` before any effect), `withRoleGateStep`, `loadGovernanceGuardFrom`
- Invariants: `ValidatePolicyConfig` (load and PUT) requires exact `ALLOW`/`DENY`/`REQUIRE_APPROVAL` actions and rejects empty or allow-only policies, and Intercept denies unknown actions; posture groups hold only posture targets, `REQUIRE_APPROVAL`, no condition, anchored intent; the floor and `applyApproverTier` are monotone; approval tier is judged at confirm time on the stored scope, for blueprints also on `buildScopeFromBlueprintFor` of the committed body (tier 2: posture/role gate, high/critical risk, cost > 5.0; applies to chat, council, and blueprint commit); every token records `purpose`, `binding_digest`, `minted_by` at mint, routes decide from the column (NULL refused, wrong purpose not consumed), tier 0/1 chat/blueprint tokens are confirmable only by `minted_by` or an approver, and commit must match the blueprint digest; tokens are single-winner (`RowsAffected`==1, else 409); heartbeats are exact (canonical subject + `agent.heartbeat` + source) and while degraded only `RefreshKnown` runs; `loadGovernanceGuard` never returns nil (load failure = degraded guard: Gatekeeper denies all but heartbeats, posture work requires approval, services row `governance: degraded`); policy PUT and approval decisions are audit-first (no audit id -> 503, no change), `previous_digest` is read under `applyMu`, and a PUT writes the file atomically before swapping memory
- Authority: governance routes need root admin + `governance:read` / `governance:write` / `approvals:decide`, and so does confirm-action for tier-2 approvals (`confirmerMayApprove`, blocker `approver_required`; audit `approval_authority`, `approval_tier`, `self_approved`; rows in `docs/API_REFERENCE.md`); negatives: anon 401, standard 403, admin without the scope 403, other principal on tier 0/1 403 `confirmer_not_proposer`
- Proof: `uv run inv core.test --package=./internal/governance --race` (and `router`, `state`, `invocation`); `uv run inv core.test --package=./internal/server --run='Governance|Approv|Tier|Token|ConfirmAction|Council|Commit|Posture|Blueprint|ServicesStatus' --race`; `uv run pytest tests/test_k8s_config_parity.py tests/test_db_confirm_token_binding.py -q`
- Pitfalls: the Gatekeeper is an observer: DENY stops Core's reaction, not NATS delivery; PUT persists to the in-container policy file, so a redeploy restores the shipped policy; posture binds to a referenced template id, so work naming no template gets capability tiers only; legacy NULL-purpose tokens are refused (re-propose); known UI gaps until the interface slice lands (after U1): the negotiate proposal card confirms through confirm-action and now gets 400 `token_wrong_purpose`, token kept (previously it silently reported success without saving a mission), and CircuitBoard "Instantiate" sends no confirm token, 403 since CE-1
- Verified: 2026-09-26 lead merge `cd8f1b9e`: security QA GO after two fix rounds (blueprint tiers, fail-closed tool risk), docs gate PASS, 116 real-PG schema tests, invocation PG suite, isolated first boot, server -race 1335, core.test

### Area: Work projections
- PRD: §Information Architecture L321-326, §Outcome Vault L136-137, §API And Event Contracts L277-284 · Scoreboard: Result-first Outcome UI
- Owned paths: `core/internal/server/work_running*.go`, `core/pkg/protocol/work_running.go` (mycelis-core-execution) · Do-not-touch: `core/internal/server/team_work_store.go`, `core/internal/server/teams_detail.go`, `core/internal/server/auth*.go`
- Seams: `team_work_store_scan.go:scanTeamWorkItem`, `outcome_projects_store.go:scanOutcomeProject`, `protocol.OutcomeHealthForTeamWork`, `protocol.OutcomeHealthForProject`, `protocol.AggregateOutcomeHealth`, `protocol.IsWorkRunningVisible`, `groups_auth.go:requireRootAdminScope`
- Invariants: read-only (no writes, events, NATS, or in-memory registry); durable `team_work_items` is the only work source; `archived`/`output_ready` never shown (SQL filter plus `IsWorkRunningVisible`); health only from `protocol.OutcomeHealth*`; one items query plus one `outcome_projects` query, item under the most recently updated non-archived referencing project, archived Outcomes never group work (SQL `status <> 'archived'` plus Go guard); any source failure is `503` naming the source, never a partial list
- Authority: `GET /api/v1/work/running` -> root admin + `groups:read`, default deny; negatives: anon 401, standard 403, admin without scope 403
- Proof: `uv run inv core.test --package=./internal/server --run='TestWorkRunning' --race`
- Pitfalls: `work_item_refs` is JSONB, so match with `?| $1::text[]` and `pq.Array`, not per-item lookups; L2 no GIN index on `work_item_refs`, so the link query scans tenant projects (fine at current scale; add a GIN index via mycelis-schema before it grows); summary counts are Outcome Health, not review counts, so review stays in the Work review panel; UI blocked on U1 — `WorkRunningPanel.tsx`, `useWorkRunning.ts`, and the `panel=running` branch land only after U1 merges to `dev`
- Verified: 2026-09-25 lead merge gate: TestWorkRunning -race, real-PG query probe (security-qa), docs links, max-lines

### Area: Organizations
- PRD: §Projects Teams And Capability Use L141, §Outcome Vault L136, §Clean Deployment And First-Boot Contract L20-25 · Scoreboard: Current-schema convergence
- Owned paths: `core/internal/server/organization_store*.go`, `organization_persistence_authz_test.go`, organization case of `qa_fixtures_purge.go` (core writer) · Do-not-touch: `core/internal/server/auth*.go`, `audit.go`, 001 G4/C2A block bytes
- Seams: `organization_store.go:OrganizationStore`, `OrganizationRepository`, `respondOrganizationStoreError`, `organization_store_postgres.go:postgresOrganizationRepository`, `qa_fixtures_claims.go:withQAFixtureScopeLock`, `qa_fixtures_purge.go:deleteQAFixtureDatabaseResource` map, `organization_handlers.go:organizationsWriteScope`
- Invariants: no in-memory fallback (nil DB -> every call `503`; memory repo exists only in `_test.go`); storage errors are `503`, never `404` or success; create is `INSERT` (duplicate id `409`), claim then insert inside the fixture fence; update is `SELECT ... FOR UPDATE` + write-back in one tx preserving `id`, `tenant_id`, `qa_fixture_scope_id`; every query filters `tenant_id='default'`; list order `name COLLATE "C", id`; purge deletes the row inside the purge tx
- Authority: `POST /api/v1/organizations` and every `PATCH /api/v1/organizations/{id}/...` -> root admin + `organizations:write` (interim until A2); reads authenticated only; negatives: anon 401, standard 403, admin without scope 403, no row written
- Proof: `uv run inv core.test --package=./internal/server --run='Test(Organization|QAFixture|ReviewLoop)' --race` with `MYCELIS_ORGANIZATION_STORE_TEST_DSN` on a disposable pg16 with 001 installed (absent DSN skips `TestOrganizationStoreRealDB_*`; skip is not proof); `uv run inv lifecycle.first-boot-proof --isolated --build`
- Pitfalls: `QAFixtureScopeID` is `json:"-"`, so it is a column, never document data; non-UUID ids are `404` (no cast error); purge deletes only where `qa_fixture_scope_id` = the purging scope (a claimed org existing outside it fails the purge as unowned; an absent row no-ops so purges resume); a profile PATCH to a missing department/agent type aborts under the lock with no write; open for A2 (pre-S1): reads do not hide QA-fixture orgs, and `POST /api/v1/internal/organizations/{id}/loops/{loopId}/trigger` has no scope gate; the retained stack needs owner-approved `compose.migrate` before Core reads the table
- Verified: 2026-09-25 lead merge gate: isolated first boot, 41 real-PG upgrade tests, 4 TestOrganizationStoreRealDB_* -race, security-qa GO, core.test, docs links, max-lines

### Area: Model routing
- PRD: §Projects Teams And Capability Use L155 (model gateways are transport, Core owns provider eligibility/routing) · Scoreboard: root model provider, per-profile route health
- Owned paths: `core/internal/cognitive/root_provider.go`, `profile_overrides.go`, `profile_route_health.go`, `profile_snapshot.go`, their tests, `server/cognitive_profile_overrides.go`, `server/cognitive_status_profiles.go`, `server/cognitive_status_*_test.go`, `server/routing_mutation_authority.go`, `brains.go`, `profiles.go`, `profile_activation.go`, their tests (mycelis-ai-runtime; S6c/S6d authority by mycelis-core-authority); `docker-compose.yml` root/profile env, `ops/cognitive_root.py`, `ops/cognitive_profiles.py`, `ops/compose_probe.py` cognitive hook (mycelis-platform-ops) · Do-not-touch: `env_overrides.go`, `server/auth.go`; `router.go`, `types.go`, `router_config.go` only for source/origin tracking
- Seams: `BrainConfig.RootProvider`/`ProfileSources`/`ProfileOverrideOrigins`/`OverlayError`, `Router.SetProfileOverrides`/`ClearProfileOverride`/`RestoreProfileBindings`/`ProfileRoutes`/`ProfileProviderSnapshot`/`AdapterSnapshot`/`ProviderSnapshot`/`StoreProviderConfig`/`ConfigSnapshot`, `recomputeProfileBinding`, `ProfileRouteHealth`, `validateRootProvider`, `persistableProfiles`, `NewRouter`, `saveConfigMu`, `server:requireRoutingWriter`/`lockRoutingWrite`/`auditRoutingMutation`/`rejectIfProviderBound`/`cognitiveFullView`, status `profiles`/`profile_route_health`, `ops/cognitive_profiles.py`
- Invariants: precedence is env override > DB `role.<name>` / runtime override > `RootProvider` > shipped `cognitive.yaml` defaults; empty env and blank `role.*` values are unset; a broken override fails that profile closed per request (never root fallback, never a startup refusal); a reset equals a restart; overrides are never persisted to YAML; `Router`'s own methods and the override/status/snapshot paths (`SetProfileOverrides`, `ClearProfileOverride`, `ProfileRoutes`, `ProfileOverrideState`, `ProfileProviderSnapshot`, `AdapterSnapshot`) hold `Router.mu` for every profile-map or adapter-map read or write they perform; `SaveConfig` is serialized by `saveConfigMu`, taken before `Router.mu`; `UpdateProvider` re-checks the provider still exists with the same credentials under the lock before writing; every routing mutation (brains, profile PUT/DELETE, provider PUT, mission-profile activation) takes the shared routing-write mutex via `lockRoutingWrite` before the DB tx, and takes `Router.mu` only inside the locked accessors, never the reverse (S6e); an unconfigured or disabled root fails `NewRouter` closed
- Authority: `PUT /cognitive/profiles`, `DELETE /cognitive/profiles/{profile}/override`, and `PUT /cognitive/providers/{id}` require root admin + `cognitive:write` (default-deny, audited before mutation, 503 without DB or audit); `GET /cognitive/status`, `GET /brains`, and the `ollama` row of `GET /services/status` are open to any caller but narrow their body/detail to an operational summary (no endpoints, model/provider ids, or config detail) unless the caller is root admin with `cognitive:read`/`cognitive:write` (S6e); `GET /cognitive/config` requires root admin + `cognitive:read`/`cognitive:write` (401 anon, 403 otherwise, no body); every mutating `/api/v1/brains/*` route (toggle, policy, add, update, delete, probe) and every `/api/v1/mission-profiles` route except GET require root admin + `cognitive:write`, default-deny, audited first as `cognitive_provider_*`/`mission_profile_*` (source `cognitive-routing-mutation`); disabling, deleting, or updating to non-executable (disabled, or blank/whitespace `model_id`) a provider an execution profile still resolves to -> `409 provider_bound` (unbound providers may still go non-executable; provider PUT has no `enabled` field, so only the blank-model case applies there); mission-profile activation is all-or-nothing (bad pair -> the `PUT /cognitive/profiles` codes, malformed `role_providers` -> `400`); routing applies before `is_active` commits (S6e): apply failure -> `500 mission_profile_routing_rejected`, nothing written; commit failure -> `500 mission_profile_activation_commit_failed`, routing reverted byte-for-byte via `RestoreProfileBindings` plus a failure audit
- Proof: `uv run inv core.test --package=./internal/cognitive --race`; `uv run inv core.test --package=./internal/server --run='TestCognitiveProfile|TestCognitiveStatus|Brain|MissionProfile|Provider|RoutingMutation' --race`; `uv run inv core.test --package=./internal/swarm --race`; `uv run pytest tests/test_cognitive_root.py tests/test_cognitive_tasks.py tests/test_cognitive_profiles.py -q`; live: `uv run inv compose.health`, `uv run inv cognitive.status`
- Pitfalls: a DB row under an env override is hidden but WARNs, and DELETE removes it while env still pins; PUT on an env-pinned profile returns 409; mission-profile activation is runtime-only and lost on restart; the old post-commit `mission_profile_routing_not_applied` code is dead (S6e made that path unreachable) and is never returned by the activation handler; `Config.Media` is read unlocked because it is populated once at startup and never mutated afterward; some tests (for example `tests/test_compose_identity_contract.py`) still `import yaml` and fail without PyYAML
- Verified: 2026-09-26 lead merge `20534a78`: security QA GO after fix round (services endpoint leak, dead code, lock helper), docs gate PASS, server/cognitive/swarm -race, core.test

### Area: Auth / web session
- PRD: §Clean Deployment And First-Boot Contract L20-25 · Scoreboard: A1 access hardening
- Owned paths: `core/internal/server/auth*.go`, `interface/lib/webAuth.ts`, `interface/proxy.ts`, `interface/app/api/auth/**`, `ops/auth.py` (core writer, A1) · Do-not-touch: `groups_auth.go` scope helpers (owned elsewhere)
- Seams: `identityForToken`, `ValidateWebIdentityForwardSecret`, `signedForwardedWebIdentityFromRequest`, `requireRootAdminScope`, `webAuth.verifyLocalPassword`, `webAuth.requestOriginAllowed`, `proxy.ts` `INBOUND_AUTHORITY_HEADERS` strip, `ops/auth.py:_ensure_web_secrets`/`_auth_posture_warnings`
- Invariants: API key/password compares are `subtle.ConstantTimeCompare`; Core accepts the API key only via `Authorization: Bearer`, no `?token=`; `MYCELIS_WEB_SESSION_SECRET`/`MYCELIS_WEB_IDENTITY_FORWARD_SECRET` are each one required variable, never falling back to `MYCELIS_API_KEY`; local sign-in verifies `MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256` or `_PASSWORD`, never the API key; the proxy strips client-supplied identity/auth headers before forwarding; `/api/v1/audit` requires root admin + `audit:read`
- Authority: `/api/v1/audit` -> root admin + `audit:read`, default deny; negatives: anon 401, standard 403; `/auth/local` and `/auth/logout` reject cross-site POST with 403
- Proof: `uv run inv core.test --package=./internal/server --run='Auth|Audit'`; `uv run --no-sync pytest tests/test_auth_tasks.py tests/test_compose_identity_contract.py -q`; `cd interface && npx vitest run __tests__/auth __tests__/lib/webAuth.test.ts __tests__/lib/proxyAuth.test.ts`
- Pitfalls: a retained stack redeployed onto this hardening has no local-admin hash yet, so local login fails until an operator sets `MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256`/`_PASSWORD` (see auth-modes.md Operator Upgrade Note); upgrading in place invalidates sessions signed under the old fallback secret, so everyone re-authenticates once; `tests/test_compose_identity_contract.py` needs PyYAML (`import yaml`) and fails in a fresh worktree venv without it
- Verified: 2026-09-25 lead merge gate: core.test, auth/compose pytest, typecheck, auth vitest 61, security-qa

## IX. Governance & Policy Engine

Deploy-owned identity posture is backend-owned: deploy-owned People & Access posture surfaced read-only, and settings PUT ignores/preserves those deploy-owned fields instead of persisting them.

### Guard - `governance/guard.go`

The guard decides whether an action is allowed, blocked, or requires proposal/approval. It must be deterministic and auditable, and it fails closed: with no loaded policy it runs degraded (see Area: Governance authority).

### Default Rules (`core/config/policy.yaml`)

Default policy should be safe for self-hosted operation and explicit about mutating actions, external services, private data, and tool use.

## X. MCP Integration

### Architecture (`internal/mcp/`)

MCP integration covers server registry, library entries, installation, activation, activity, health, and governed tool calls.

### Transport: stdio or SSE

Supported transports should be explicit and observable. Curated stdio servers must run inside the configured output/workspace boundary.

### Curated Library (`core/config/mcp-library.yaml`)

Library changes require docs/tests when they affect operator workflow, capability posture, or task behavior.

## XI. Startup & Shutdown

### Startup Sequence

Startup resolves config, policy, bootstrap bundles, DB, NATS, providers, MCP posture, and HTTP services. Normal startup should fail closed when required bootstrap truth is missing.

### Graceful Shutdown

Shutdown should stop HTTP, streams, NATS consumers, background workers, and local service resources cleanly. Use `uv run inv lifecycle.down` or the matching runtime task for operator control.

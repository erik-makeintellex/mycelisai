# Mycelis Agent Rules

These rules apply to all agents in this repository. The repository-specific contracts below remain in force.

Priority: correctness → safety / authority → accepted architecture → tests / evidence → minimal change → token efficiency.

## Read First And Scope

- Search first; read narrow ranges. Inspect this file, the current task, relevant canonical PRD and owning architecture sections, files to edit, and nearby tests. Consult recovery/state when relevant. Do not scan the whole repository without reason.
- Reuse existing Core contracts, registries, APIs, event spine, memory, execution, governance, tests, and task runners. Never invent parallel architecture. If an architecture conflict appears, stop and report the exact conflict.
- Change only the requested scope. No opportunistic refactoring, unrelated formatting, broad renames, unnecessary dependencies, or new services when an existing seam works.
- Inspect implementation first. Prefer small functions, stable contracts, explicit state, deterministic behavior, and fail-closed authority/security. Avoid duplicate registries, queues, approval systems, memory authority, hidden fallback, and magic provider behavior.

## Area AGENTS.md Index

Feature folders carry their own `AGENTS.md` with owned paths, the roster role that writes there, contracts/invariants, exact targeted gates, and gotchas for that area. Root rules always win on conflict; an area file only adds area-specific detail. Any writer or reviewer must read the area `AGENTS.md` of every folder it touches, in addition to this file, before editing.

| Area | File | Covers |
| --- | --- | --- |
| Core (Go) | [`core/AGENTS.md`](core/AGENTS.md) | Backend runtime, orchestration, identity, execution; splits by roster role |
| Core HTTP/auth | [`core/internal/server/AGENTS.md`](core/internal/server/AGENTS.md) | Routes, blocker envelope, auth gate patterns |
| Core cognitive | [`core/internal/cognitive/AGENTS.md`](core/internal/cognitive/AGENTS.md) | Provider routing, the token-budget choke point |
| Core swarm | [`core/internal/swarm/AGENTS.md`](core/internal/swarm/AGENTS.md) | Agent dispatch, tool scope, NATS subjects |
| Core migrations | [`core/migrations/AGENTS.md`](core/migrations/AGENTS.md) | Single-writer schema baseline, isolated first boot |
| Framework Runs service | [`services/framework-runs/AGENTS.md`](services/framework-runs/AGENTS.md) | Separate Go module, own gate, B2 Compose-only posture |
| Interface (TS) | [`interface/AGENTS.md`](interface/AGENTS.md) | UI, BFF routes, Vitest; the U1 lane boundary |
| Interface E2E | [`interface/e2e/AGENTS.md`](interface/e2e/AGENTS.md) | Playwright lease, `.env` sign-in, fixture purge |
| Ops/platform | [`ops/AGENTS.md`](ops/AGENTS.md) | Task runner, Compose, CI (task names are a moving target while "Task runner tightening" is active) |
| Docs | [`docs/AGENTS.md`](docs/AGENTS.md) | Docs ownership, link/layout gates |
| Charts (Helm) | [`charts/AGENTS.md`](charts/AGENTS.md) | Helm packaging, cross-linked with deploy |
| Deploy (Compose) | [`deploy/AGENTS.md`](deploy/AGENTS.md) | Compose overlays, cross-linked with charts |
| Tests | [`tests/AGENTS.md`](tests/AGENTS.md) | Python test/docs-contract ownership by area |
| Proto | [`proto/AGENTS.md`](proto/AGENTS.md) | Protobuf sources and generated consumers |
| SDK | [`sdk/AGENTS.md`](sdk/AGENTS.md) | Python relay client |

## Authority And Effects

- Preserve Human/Soma → BFF → Core → authority → execution. The browser does not orchestrate internal services. Core owns authorization; the BFF is a transport/session boundary, not a second authority.
- NATS is transport, not authority. Models provide cognition, not authority. MCP/tool discovery, memory, provider keys, and prompts do not grant permission or establish agent identity.
- Tool success is not Outcome success. Token EOF is not completion. An artifact candidate is not a trusted artifact.
- Before an external effect, require the accepted contract: identity, current authority, immutable grant, capability binding, reservation when needed, durable invocation, ownership, and audit.
- Never blindly retry an unknown external effect. Transport redelivery does not prove that an effect is safe to replay.
- Enforce important security rules below prompts. Ask: "If the model ignores the prompt, what stops it?" If the answer is nothing, implementation is incomplete.

## APIs, Events, Models, And Memory

- Frontend code uses stable Mycelis APIs without exposing internal topology. Keep query, command, approval, control, and admin semantics distinct when their authority differs.
- Browsers use Mycelis event projection, never raw NATS. Distinguish durable state from transient token data; preserve scope during reconnect/replay, make gaps explicit, and never infer success from stream close.
- Brain semantics belong to Mycelis; provider/model routing is operational. Avoid unnecessary vendor/model coupling. Capabilities require evidence; unknown required capabilities fail closed. Embeddings follow generation privacy and egress discipline.
- Memory is advisory context, never permission. Preserve provenance, scope, correction, deletion, and lifecycle; do not create another authoritative memory store.

## Worktree Safety

- **Clean Git is a mandatory delivery gate.** Start each slice from a clean worktree and index. Before handoff, starting another slice, merging, or promotion, commit the reviewed work and verify `git status --porcelain=v1 --untracked-files=all` is empty. Expected edits may remain uncommitted only while the current slice is actively being worked.
- Inspect every registered worktree at close-out. Report its branch, HEAD, cleanliness, and any outstanding owner; a clean current checkout does not mean the repository setup is clean. Record local/upstream divergence without treating a local commit as a push or integration proof.
- Never accumulate unrelated slices in one dirty branch or claim completion with unexplained staged, unstaged, or untracked files. Preserve inherited work, establish its owner, and isolate/review it before proceeding. If interrupted or blocked, report the exact remaining work and owner; a recovery checkpoint preserves work but does not certify it for integration.
- Do not manufacture cleanliness by deleting unknown files, hiding changes with ignore rules or index flags, or stashing them without an explicit recovery/handoff record. Keep secrets and generated proof artifacts out of commits; commit only inspected, scoped source changes with their applicable docs and evidence.
- Before branch or worktree mutation, inspect `git status`, the current branch, worktrees, and diff. One integration owner controls Git topology; agents must not mutate it concurrently.
- Never destroy unknown work, wholesale-merge stale WIP, or delete untracked files without inspection. Destructive reset/deletion requires explicit task approval and recovery.

## Evidence And Completion

- Run the smallest owned tests first, then required gates. New behavior needs positive and negative tests, plus regression tests when relevant. Authority work requires adversarial negative proof. Do not write fake test-only implementations.
- Completion needs current evidence: executed tests/builds, browser or DB proof, traces, artifacts, exact diffs, or reproducible commands. Old results are not current proof; fixture proof is not live proof; documentation proof is not runtime proof. State what was not verified.
- Completion includes applicable code, tests, negative cases, migration/recovery, docs, evidence, and independent QA. An implementer does not self-certify final GO. Truth takes precedence over confidence.
- Delegate only independent work, with one writer per overlapping code area. Use independent QA when required; keep subagent output concise. Do not spawn agents for trivial work.

## Communication And Decisions

- Be terse. Avoid repeating the task, obvious commands, reasoning narration, or long summaries. Use tables only when they save tokens. Default progress is at most eight lines.
- Send useful progress for blockers, architecture conflicts, risky decisions, major phase completion, and final results. Use `STATUS: working|blocked|done`, then relevant `CHANGED`, `TEST`, `RISK`, and `NEXT` fields; omit empty fields.
- Final responses default to `STATUS: GO | CONDITIONAL | BLOCKED`, with terse `CHANGED`, `TESTS`, real `RISKS`, and one-line `NEXT` as applicable. Target at most 120 words, or 50 for simple tasks, unless the task asks for more.
- Resolve normal engineering choices without asking the human. Ask only for architecture conflicts, destructive actions, unknown valuable work, credential/production activation, required scope changes, or accepted-invariant changes.
- Update the canonical owner when a contract changes. Planning notes are not product truth; do not create permanent duplicate authority docs. Retire superseded proposals when accepted text moves into canonical documentation.

## Development Model Routing

Use the smallest tier that passes the task's proof gate; optimize accepted results, not token price alone. These developer-agent choices do not configure Mycelis runtime providers. Pick the tier first, then the harness model.

| Tier | Work | Codex | Claude | Local model (Ollama `127.0.0.1:11434`) |
| --- | --- | --- | --- | --- |
| T0 inventory | Read-only search, file/route inventory, log triage, evidence formatting | `gpt-6-luna` / high | Haiku | Allowed: summarizing supplied text |
| T1 routine | Specified coding, tests, UI copy, docs sync, automation | `gpt-6-sol` / medium | Sonnet | Only small scoped edits with a passing test gate |
| T2 hard | Concurrency, persistence, schema, security fixes, difficult coding | `gpt-6-sol` / high | Opus | Not allowed |
| T3 authority | Contract freeze, architecture conflicts, independent security QA, final GO | `gpt-6-astra` / low→high | Opus | Not allowed |

- The local model (currently Windows Ollama qwen3:14b; WSL vLLM is stopped because it locked WSL up) keeps data on the host and suits confidential or offline work. It never owns authority, schema, security review, or final GO, and its output needs a stronger-tier review before merge. Record the served model id; the provider's `/v1/models` is the source of truth.
- A newly available model enters a tier only after it passes that tier's proof gate on a bounded slice. A model name is not evidence.
- One agent is the default; spawn only authorized, independent work. Escalate a tier after a concrete failed proof or unresolved ambiguity, not for routine command execution.
- Record model, tier, token usage, latency, and retries in close-out when the harness exposes them. Do not invent savings or confuse the lead session's model with spawn settings.
- Harness rosters (for example local `.claude/agents/` or Codex agent config) must follow this table and the rules in this file; they are not authority. Codex effort guidance follows [official Codex documentation](https://learn.chatgpt.com/docs/agent-configuration/subagents). All assignments are repository operating choices, not capability guarantees.

## Context Execution

Context is the scarcest resource on every harness. The lead keeps conclusions; files keep detail.

- **Budget by window.** Local 16k: give one contract packet (at most ~4k tokens) and one owned file range at a time, with no repository sweeps. Mid tier: the packet plus owned files. Top tier: cross-file reasoning, but still range reads.
- **Packets, not history.** Freeze contracts to a file (a session scratch path, or the owning plan once accepted). Pass implementers the packet path, owned paths, invariants, and proof gate. Never pass the thread history.
- **Read narrowly.** Run `rg -n` first, then ranged reads. Read the PRD and scoreboard by section. Read a whole large file only when editing it.
- **Delegate sweeps.** Broad searches run in a T0/T1 subagent. It returns at most 300 words with `file:line` refs; full output goes to a file. Spot-check the claims the lead will act on before relying on them.
- **Evidence to files.** Logs and reports go to ignored runtime or `/tmp` paths. The thread carries commands, pass/fail/skip counts, and artifact paths, never raw logs.
- **Reuse over respawn.** Continue a live agent that already holds the relevant context. Close agents at handoff. Do not poll background work: run it in the background and act on completion.

## Repository Standards

This repository is Go-first for product/runtime work and Python-first for management automation.

## Language Ownership

- Go owns core runtime, orchestration, APIs, NATS integrations, and persistence-facing backend logic.
- TypeScript owns the interface, in-app docs browser, and operator-facing workflow surfaces.
- Python owns app management tasks, operator automation, CI task orchestration, and repo-local test harnesses.
- SQL owns schema and migration contracts.
- PowerShell is allowed only as a thin host wrapper when the local platform requires it. App-tied management logic must not live in PowerShell scripts.

## Task Runner Contract

- Use `uv run inv ...` for real task execution.
- Use `uvx --from invoke inv -l` only as a compatibility probe.
- Do not use bare `uvx inv ...`.
- Use `uv run inv lifecycle.first-boot-proof` for clean deployment/startup proof when a slice changes startup, migrations, persistence, generated workspace roots, bootstrap state, or deployment assumptions. Do not substitute ad hoc DB/file/NATS cleanup commands unless the task itself is being repaired.
- Keep the public Invoke surface at or below 95 registered tasks. New tasks must provide distinct operator value and should replace or consolidate an existing entry when possible.
- Do not register convenience aliases for an existing task. Documentation and automation must call the canonical owning namespace directly.
- When invoke task behavior or task names change, update `README.md`, `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, `ops/README.md`, and any affected in-app docs surface in `interface/lib/docsManifest.ts` in the same slice.

## README Navigation Contract

- Keep a structured `## README TOC` near the top of `README.md`.
- When adding, removing, or renaming major README sections, update the TOC links in the same change.
- Treat the README TOC as the stable navigation contract that future development agents should use before scanning the full file.

## Feature Branch And Merge Quality Contract

- `main` is the production-promotion branch. `dev` is the shared integration branch. Product/runtime feature work must start from a clean, updated `dev` on an intentionally named `feature/*` branch unless the user explicitly asks for a different branch shape.
- Keep each branch scoped to one reviewable slice. If work expands, split follow-on work into a new branch instead of letting one branch become a mixed backlog.
- Keep one active feature branch per delivery goal; use scoped commits for its intermediate checkpoints, not additional checkpoint branches. Merge to `dev` only after proof, rerun affected integration gates, then remove the merged local feature branch and its owned temporary worktree.
- Before engaging teams or implementing a substantial next slice, review current branch state, the active scoreboard, canonical PRD alignment, and likely proof gates. Write down the execution shape before spawning or redirecting agents.
- Before spawning new sub-agents for any work, review existing open agentry for reuse or closure. Reuse relevant active agents when their context matches the slice; close completed, stale, duplicate, or no-longer-relevant agents before adding more background work.
- Spawn narrowly scoped sub-agents without inherited long-thread context unless that history is essential. Close agents after handoff so persisted development sessions do not grow without bound.
- A feature branch reaches integration quality only after code, docs/state, focused tests, typecheck/build gates, and any required live GUI proof pass together. Commit that proven state before merging it into `dev`.
- After every feature merge, test the resulting `dev` state again. Run the affected integration suites, service health, and live GUI journeys needed to detect cross-slice regressions; feature-branch proof is not a substitute for post-merge integration proof.
- Promote `dev` to `main` only from a clean, committed integration checkpoint after the required broader release preflight, deployment/runtime proof, and user-facing browser certification pass. Rerun the release smoke and health checks after the promotion.
- Before every merge or promotion, review `git status --short --branch`, `git diff --check`, branch divergence, untracked files, temporary proof artifacts, and affected docs. Resolve or record every item.
- After a feature is merged and its `dev` proof passes, delete the merged local feature branch. Explicitly review remote branches before deletion. Keep unmerged/archive branches only with a named purpose.
- If urgent work must happen directly on `main`, the close-out must still follow the same branch-quality checklist before commit, push, or handoff.

## Team Orchestration And Messaging Contract

- The lead agent is the messaging avatar for team execution. It coordinates intent, decisions, dependencies, and proof across sub-agents and Mycelis teams instead of letting background work drift into disconnected threads.
- When the local NATS-backed Mycelis stack is intentionally running and relevant to the slice, prefer using the product's bus-facing workflows for team coordination proof, status, and handoff checks. If the bus is unavailable or unnecessary, record that explicitly and keep coordination in the lead thread.
- Team communication should mirror the product architecture: concise intent, assigned ownership, expected output, proof gate, status updates, blockers, and handoff notes. Avoid spawning parallel teams without a clear owner, bounded deliverable, and cleanup path.
- **Delivery target teams (owner rule, 2026-09-28): at most 3 agents per delivery target.** A delivery target is one named outcome, for example "MCP authority" or "e2e suite green". At most 3 agents may be live for one target at a time; the lead is not counted. Separate targets may run in parallel only when their owned files are disjoint, and the total must stay within the account's rate limits.
- Before spawning for a target, declare its team: at most 3 roles from the harness roster (for example writer, second writer or interface writer, then independent QA), each with its model tier. Record the team in the scoreboard's "Delivery Targets And Teams" table (`.state/V8_DEV_STATE.md`). Changing a role means updating that row first.
- Continuing a target reuses its declared team. Resume a live agent that holds the context; otherwise respawn the same role with a fresh packet. When a target closes, mark its row closed and list the agents actually used, so the next continuation starts from the record, not from memory.
- Close-out must include what teams or agents were reused, spawned, messaged, closed, or intentionally skipped. List them per delivery target.

## Canonical Docs Location

- Keep user-shared root-level architecture entrypoints under `architecture/`.
- Put new canonical planning, target-delivery, UI-target, execution-model, and delivery-governance docs under `docs/architecture-library/`.
- Treat `docs/architecture-library/MYCELIS_CANONICAL_PRD.md` as the single PRD, product, UX, runtime, MVP, and release-gate authority.
- Do not restore old versioned V7, V8.2, or split V8.3 architecture files; promote current truth into the canonical PRD instead.
- If a canonical doc is meant to be readable in the in-app `/docs` page, add or update its entry in `interface/lib/docsManifest.ts` in the same change.

## State Location

- Keep mutable delivery state under `.state/`, with `.state/V8_DEV_STATE.md` as the active scoreboard.
- Historical migration evidence lives in Git history, not retained state docs.
- `.state/` is ignored for new local/session artifacts, but tracked state files already under `.state/` remain part of the repository contract.
- Do not add transient run logs, browser reports, kubeconfigs, temporary plans, or local service snapshots to root.

## Documentation Synchronization Contract

- Every implementation slice that changes product behavior, runtime behavior, operator workflow, API contract, governance posture, or canonical terminology must include a documentation review in the same slice.
- Whenever a confirmed task, feature, workflow, or spectrum of work changes, expands, or replaces prior behavior, perform an obsolescence review in the same slice across commands, code, configuration, tests, docs, routes, fixtures, and generated scaffolding. Remove items that no longer serve the confirmed path, update items that still apply, and record the canonical replacement in the owning docs or state file.
- Do not retain obsolete compatibility aliases, parallel implementations, archived doctrine, or stale tests by default. Keep one only when an explicit compatibility requirement names its owner, supported lifetime, and removal gate.
- User-facing Soma proposal, completion, and deliverable copy must foreground the actual Outcome result target. Internal handoff/planning files such as team evocation briefs belong behind Details, proof, or Inspect when a delegated result contract exists.
- Completed team-work events should say what named deliverable is ready and carry one primary open action before secondary proof, folder, or technical actions.
- Generated app/package requests must preserve the requested package folder, entrypoint, title, and validation contract through proposal, team handoff, runtime proof, and recovery. If the worker produces no trusted file evidence, runtime-owned recovery may create a clearly labeled validated fallback package only inside the approved group package target; it must not claim semantic final delivery from planning-only output or expose loose general-bucket scripts as the user deliverable.
- Interactive package validation markers must be attached to the actual surface being validated. A canvas or visual workflow may not place `data-mycelis-validation-surface` on a static instruction label while changing another element; the marked surface must change or the worker must repair before completion.
- Update the owning docs in the same change whenever meaning changed, not later as cleanup.
- At minimum review `README.md`, `.state/V8_DEV_STATE.md`, the owning canonical/user/ops docs for the touched surface, and any affected in-app docs entry in `interface/lib/docsManifest.ts`.
- When API behavior or payload meaning changes, review `docs/API_REFERENCE.md` in the same slice.
- When testing or task-running behavior changes, review `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md` in the same slice.
- Slice close-out should explicitly report which docs changed and which touched docs were reviewed but left unchanged.

## Native Code Context Map Standard

- Mycelis may use local code-structure maps as a native governed source/capability for repository understanding, impact review, implementation planning, and proof grounding. This is not support for an external graph service and must not create a new primary product surface.
- Code context maps are source aids, not authority. Verify relevant source files before editing or asserting behavior, and use exact file/path refs in findings and proof.
- Prefer deterministic local extraction for structure. LLMs may interpret or summarize the map, but they must not be required to construct parser facts such as files, symbols, imports, references, or extracted edges.
- Keep extracted facts separate from inferred relationships. Any inferred edge, ownership, or impact claim must be labeled as inferred and remain behind Inspect or proof details unless the user asks for depth.
- Generated graph/index/cache artifacts are runtime or workspace artifacts. Do not commit them unless an explicit fixture or migration test names why the file belongs in source control.
- When broad code changes are planned and a native code context map exists, consult it for impact before editing. If it is unavailable or stale, proceed with `rg`, source reads, and tests, and record the missing map only when it affects delivery confidence.

## Runtime Config And Proof Boundary

- `.env` is the repo-local secret store across runtime paths. Use secret references in committed config and never store raw secrets in UI, logs, state files, or architecture docs.
- `.env.compose` is for Compose topology and non-secret runtime shape; secret-like values from `.env` are authoritative over stale Compose values.
- Mycelis development is platform-aware but platform-agnostic. Do not assume Windows, WSL, Linux, Docker Desktop, Rancher Desktop, or Kubernetes is the active runtime just because a prior session used it. Discover the current host, checkout location, configured Docker owner, service endpoints, and proof lane before starting or judging services.
- Prefer the configured service targets over host folklore. Read process env plus `.env`, `.env.compose`, task defaults, Compose/Helm values, and the active proof command before choosing addresses, ports, storage roots, or provider endpoints.
- Do not treat `localhost`, `127.0.0.1`, `0.0.0.0`, `host.docker.internal`, a Windows LAN IP, an in-cluster service name, or a port-forward as interchangeable. Each name is valid only from a particular network namespace. Prove reachability from the process that will use it.
- `0.0.0.0` is a bind/listen address, not a client/probe target. Service probes and browser/API clients should use the configured reachable host such as `127.0.0.1`, a published host port, a service DNS name, or an operator-facing URL.
- Windows remains a valid editing, git, browser, and local-service surface when configured; WSL/Linux remains a valid development and proof surface when configured. The current environment, not historical habit, decides where install/build/test/Compose/browser proof runs.
- When working from WSL or Linux, keep performance-sensitive checkouts, virtualenvs, Node modules, Go caches, Playwright browsers, generated outputs, and tool caches on the native Linux filesystem unless the task explicitly proves a mounted Windows path. Avoid `/mnt/*` for hot build/test paths by default.
- When working from Windows, keep Windows-only cleanup and host-service actions scoped to repo-owned artifacts or explicitly approved user-profile caches. Do not delete Docker volumes, WSL distros, Rancher/Desktop state, or shared package caches as a substitute for configured cleanup tasks.

## Configured Service Target Standard

- Before starting, stopping, testing, or declaring a service healthy, identify the active target set:
  - Core API bind/probe: `PORT`, `MYCELIS_API_HOST`, `MYCELIS_API_PORT`, or `MYCELIS_API_BASE_URL`.
  - Interface bind/probe: `MYCELIS_INTERFACE_BIND_HOST`, `MYCELIS_INTERFACE_HOST`, `MYCELIS_INTERFACE_PORT`, `INTERFACE_PORT`, `PLAYWRIGHT_PORT`, and the browser URL actually opened.
  - PostgreSQL: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, `DB_SSLMODE`, plus Compose/K8s port mappings when applicable.
  - NATS: `NATS_URL`, `MYCELIS_NATS_SERVICE_ID`, and any monitor/published port.
  - Compose: `MYCELIS_DOCKER_HOST`, `MYCELIS_WSL_DISTRO`, `MYCELIS_COMPOSE_POSTGRES_PORT`, `MYCELIS_COMPOSE_NATS_PORT`, `MYCELIS_COMPOSE_NATS_MONITOR_PORT`, `MYCELIS_COMPOSE_CORE_PORT`, `MYCELIS_COMPOSE_INTERFACE_PORT`, `MYCELIS_COMPOSE_OLLAMA_HOST`, `MYCELIS_OUTPUT_BLOCK_MODE`, and `MYCELIS_OUTPUT_HOST_PATH`.
  - Kubernetes/Helm: active namespace, values file, ingress/operator URL, port-forward targets, storage class/PVC roots, and `MYCELIS_K8S_*` provider endpoints.
  - AI providers: provider-specific `MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT`, `MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID`, enabled flags, media endpoint variables, and Compose/K8s adapter variables. Do not rely on legacy `OLLAMA_HOST` as the Mycelis provider-routing contract.
  - Storage roots: `MYCELIS_WORKSPACE`, `MYCELIS_ARTIFACT_ROOT`, `DATA_DIR`, `MYCELIS_CONFIG_ROOT`, Compose output mounts, and any Playwright backend workspace probe override.
- Record the active target set in close-out when it influenced proof. Use addresses without secrets: include hostnames, ports, lane, and provider posture, but never API keys, session secrets, OAuth secrets, tokens, or raw credentials.
- Service proof must exercise the same route the user or runtime will use:
  - Browser proof uses the delivered UI address, not a convenient alternate port.
  - Interface proxy proof uses its configured Core target.
  - Core proof uses its configured database, NATS, workspace, artifact roots, and provider endpoints.
  - Docker/Compose proof validates from inside the relevant container when the dependency is consumed inside the container.
  - WSL-to-Windows or container-to-host AI proof must run from WSL and from a container when Core will run in Docker.
  - Kubernetes proof validates the cluster service/ingress/port-forward that the release lane names, not local source services.
- Treat missing target discovery as a blocker for live proof, not a reason to guess. If the configured target is absent, malformed, loopback-only from the wrong namespace, or unreachable, report `BLOCKED` with the exact probe command and failing endpoint.
- Do not silently fall back from one lane to another. A passing mocked browser test, source-mode route, or alternate local service cannot certify a Compose, WSL, Kubernetes, remote, or production-facing lane.
- Start services through their owning task or platform lane:
  - Source development: `uv run inv compose.infra-up`, `uv run inv db.migrate`, `uv run inv lifecycle.up --frontend`, and `uv run inv lifecycle.health` when the repo is configured for local source Core/Interface.
  - Full Compose: `uv run inv compose.up`, `uv run inv compose.health`, and storage health tasks when validating packaged single-host runtime.
  - Kubernetes: `uv run inv k8s.deploy`, `uv run inv k8s.wait`, `uv run inv k8s.bridge` or the configured ingress path when validating clustered runtime.
  - Existing external services: prove health through configured probes and do not restart or replace them unless the task explicitly owns that service.
- Stop services only inside the active lane's ownership boundary. `lifecycle.down` stops repo-owned local app services and preserves reusable data-plane dependencies unless `--include-data-plane` is intentionally requested. Do not stop Ollama, WSL, Docker, Rancher Desktop, Kubernetes, or shared brokers unless the user explicitly asks or the owning task documents that behavior.
- When Docker is uninstalled, unavailable, or intentionally moved to another host, do not keep probing the stale Docker endpoint. Switch to the configured Docker owner or mark Docker-backed live proof `BLOCKED` until the configured owner is healthy.
- When services are already running, first determine whether they are repo-owned and match the configured target. A port occupied by another project is not proof; it is a conflict to resolve or record.
- For AI endpoints, prove the exact API shape:
  - Ollama native health uses `/api/tags`.
  - OpenAI-compatible provider routing typically uses a `/v1` base URL.
  - Compose maps `MYCELIS_COMPOSE_OLLAMA_HOST` into provider-specific `/v1` endpoints for Core containers.
  - Media providers such as Forge/ComfyUI use their own configured readiness/API paths.
- Keep setup and proof failure language precise:
  - Incomplete `.venv` blocks repo task execution.
  - Docker/data-plane failure blocks live backend proof.
  - Core or Interface down usually explains browser `500`/`503` network failures.
  - AI endpoint refusal blocks live AI-backed proof, but not non-AI unit/type/mocked browser gates.
  - A provider/model unavailable state is not equivalent to a UI regression unless the UI fails to show the expected normalized blocker/recovery path.

## Feature Status Standard

- Use these canonical status markers in planning and state docs: `REQUIRED`, `NEXT`, `ACTIVE`, `IN_REVIEW`, `COMPLETE`, `BLOCKED`.
- Preferred meanings:
  - `REQUIRED`: must exist for target delivery or gate pass, but not started/ready yet
  - `NEXT`: highest-priority upcoming implementation slice
  - `ACTIVE`: currently being worked
  - `IN_REVIEW`: implemented and awaiting validation/review/gate decision
  - `COMPLETE`: accepted and delivered
  - `BLOCKED`: cannot advance until a named dependency or defect is resolved
- Avoid inventing synonymous markers like "in progress", "done-ish", or "pending review" when one of the canonical markers fits.

## NATS Signal Standard

- Use canonical subject constants from Go protocol/topic definitions for product subjects. Do not hardcode `swarm.*` literals in runtime code.
- Every bus payload that represents product behavior must declare enough metadata to identify source, scope, and intended consumer.

Required metadata for governed product signals:
- `run_id` when the signal is execution-linked
- `team_id` when team-scoped
- `agent_id` when agent-scoped
- `source_kind`
- `source_channel`
- `payload_kind`
- `timestamp`

Canonical `source_kind` values:
- `workspace_ui`
- `web_api`
- `automation_trigger`
- `scheduler`
- `sensor`
- `iot`
- `internal_tool`
- `mcp`
- `system`

Preferred subject families:
- `swarm.team.{team_id}.internal.command` for directed team input
- `swarm.team.{team_id}.signal.status` for concise operator-readable status
- `swarm.team.{team_id}.signal.result` for bounded execution outcomes
- `swarm.team.{team_id}.telemetry` for high-volume machine telemetry
- `swarm.council.{agent_id}.request` for request-reply specialist calls
- `swarm.mission.events.{run_id}` for run-linked fanout
- `swarm.global.broadcast` for governed fanout

Channel rules:
- Web/API results must normalize to the standard API envelope before UI consumption.
- IoT and sensor payloads must identify device/feed origin and stay separated from operator-facing result channels until normalized.
- High-volume telemetry must not be reused as operator status or workflow-result channels.
- Mutating actions must emit persistent mission events in addition to transient bus signals.

## Infrastructure Development Channel Boundary

- Infrastructure-development or experimentation subjects are local-only and must not be committed as canonical orchestration channels.
- Do not add development-only infrastructure subjects to shared architecture docs, protocol constants, standing manifests, or operator UI flows unless they are intentionally promoted through architecture review.
- If temporary infrastructure-dev subjects are needed for local work, keep them out of the authoritative channel taxonomy and out of persisted workflow orchestration.

## Logging and Error Handling

- Go runtime logs should be structured and component-identified.
- Python task output should be operator-readable, fail fast on broken prerequisites, and avoid false-success messaging.
- UI surfaces should show normalized error states, not raw backend noise.

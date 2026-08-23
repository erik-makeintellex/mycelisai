# Repository Standards

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
- When the local NATS-backed Mycelis stack is intentionally running and relevant to the slice, prefer using the product's bus-facing workflows for team coordination proof, status, and handoff checks. If the bus is unavailable or unnecessary, record that explicitly and keep coordination in the Codex thread.
- Team communication should mirror the product architecture: concise intent, assigned ownership, expected output, proof gate, status updates, blockers, and handoff notes. Avoid spawning parallel teams without a clear owner, bounded deliverable, and cleanup path.
- Close-out must include what teams or agents were reused, spawned, messaged, closed, or intentionally skipped.

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
- Windows is the source-edit and git surface. WSL is the release-proof environment for install, build, tests, Compose, and live GUI validation.

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

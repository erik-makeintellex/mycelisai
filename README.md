# Mycelis

Mycelis is an AI Organization platform for creating, operating, and evolving governed multi-role systems through a Soma-primary operator experience.

This README is the primary development-swarm inception document. It defines the top-level architecture truth split for active work:
- final production architecture target
- current release target
- current implementation state

Canonical ownership reminder:
- `README.md` = development-swarm inception and layered truth
- `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` = current release architecture
- `v8-2.md` = canonical full actuation / production target
- `V8_DEV_STATE.md` = actual implementation truth
- `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md` = UX/operator truth

## README TOC

- [Fresh Agent Start Here](#fresh-agent-start-here)
- [What Mycelis Is](#what-mycelis-is)
- [Final Production Architecture (V8.2)](#final-production-architecture-v82)
- [Current Release Target (V8.1)](#current-release-target-v81)
- [Current Implementation State](#current-implementation-state)
- [Default And Advanced Surfaces](#default-and-advanced-surfaces)
- [Architecture Terms To Operator Terms](#architecture-terms-to-operator-terms)
- [Detailed Framework Memory](#detailed-framework-memory)
- [Feature Status Standard](#feature-status-standard)
- [Required Review Targets](#required-review-targets)
- [Command Contract](#command-contract)
- [Development Contract](#development-contract)
- [Playwright Contract](#playwright-contract)
- [Development Workflow](#development-workflow)
- [Documentation Responsibilities](#documentation-responsibilities)
- [Status](#status)

## Fresh Agent Start Here

Review these in order before touching code or planning state:

1. [AGENTS.md](AGENTS.md)
2. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
3. [README](README.md)
4. [V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md)
5. [V8.2 Production Architecture Target](v8-2.md)
6. [V8 Runtime Contracts](docs/architecture-library/V8_RUNTIME_CONTRACTS.md)
7. [V8 Config and Bootstrap Model](docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md)
8. [V8 UI/API and Operator Experience Contract](docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
9. [Target Deliverable V7](docs/architecture-library/TARGET_DELIVERABLE_V7.md)
10. [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
11. [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md)
12. [UI Target And Transaction Contract V7](docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
13. [Operations](docs/architecture/OPERATIONS.md)
14. [Testing](docs/TESTING.md)
15. [V7 Development State](V7_DEV_STATE.md)
16. [V8 Development State](V8_DEV_STATE.md)
17. [Docs Manifest](interface/lib/docsManifest.ts)

Fresh-agent review rule:
- README is the primary architecture inception document for active development.
- [V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) is the current release target.
- [V8.2 Production Architecture Target](v8-2.md) is the canonical full production target and full actuation architecture.
- [V8 Development State](V8_DEV_STATE.md) is the live implementation scoreboard.
- V7 documents remain migration inputs until replaced, but they do not override the V8 bootstrap and release truth.

Bootstrap reminder:
- treat `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` as the canonical V7->V8 migration and bootstrap contract, not just another planning note
- always translate V7 YAML, runtime config, DB seeding, and operator wizard flows through that model before they touch a live organization
- `Template ≠ instantiated organization`, so only instantiated orgs enter bootstrap resolution while templates stay reusable blueprints
- startup truth is bundle-driven and fail-closed: normal startup fails closed unless a valid bootstrap bundle is present, and `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must select a bundle whenever more than one is mounted

## What Mycelis Is

Mycelis is a governed AI Organization system.

In operator-facing language, the product lets someone:
- create an AI Organization
- work through Soma instead of a raw agent swarm
- inspect advisors, departments, automations, learning signals, and settings in human-readable terms
- keep execution, automation, and future learning behavior bounded by policy and inheritance

In architecture terms, Mycelis is built around:
- instantiated organizations as runtime truth
- a Soma coordination layer that can engage Team Leads, advisors, departments, and specialists
- advisory and specialist layers beneath that orchestrator
- auditable execution, automation, memory, and continuity contracts

## Final Production Architecture (V8.2)

The final production architecture target is [V8.2 Production Architecture Target](v8-2.md), the full actuation architecture for Mycelis.

V8.2 is the distributed end-state we are building toward:
- distributed execution across a control plane and execution nodes
- an active learning system that can evaluate, promote, and reuse reviewed learning safely
- a capability system that governs what execution surfaces agents may use
- editable automations, policy-bounded actuation, and stronger continuous operation

V8.2 summary:
- distributed execution means the AI Organization can coordinate work across more than one host or environment
- the learning system turns reviewed outcomes into governed memory, reusable procedures, and safer organization improvement
- the capability system keeps action surfaces allowlisted, scoped, auditable, and policy-checked

Explicit distinction:
- V8.1 is the current release target
- V8.2 is the full production target and full actuation architecture

V8.2 is not the current MVP release surface. It is the final production architecture truth the rest of the docs must converge toward.

## Current Release Target (V8.1)

The current release target is [V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md).

V8.1 defines the MVP release we are aligning implementation to now: a Soma-primary AI Organization system with bounded automation visibility, learning visibility, inheritance contracts, and safe organization structure surfaces.

Included in the V8.1 release target:
- AI Organization creation and Soma-primary workspace flow
- guided first-run Soma actions that help a new operator choose a visible next step quickly
- organization, department, advisor, and role-type visibility in operator language
- bundle-driven startup truth with policy-bounded inheritance
- bounded AI Engine and Response Style controls
- read-only Automations visibility
- read-only learning visibility
- intentional empty states, retry guidance, and partial-failure-safe workspace panels
- Loop Profiles, Runtime Capabilities, semantic continuity, and Procedure / Skill Sets defined as architecture truth even where implementation is still partial

Excluded from the V8.1 release target:
- distributed multi-host execution
- editable automations
- broad live actuation
- unrestricted capability controls
- autonomous memory mutation or silent self-rewrite
- advanced raw architecture/configuration panels in the default operator flow

Release rule:
- if a surface belongs to V8.2 but not the V8.1 MVP, it should remain out of the default release surface until explicitly promoted

## Current Implementation State

Actual implementation state lives in [V8_DEV_STATE.md](V8_DEV_STATE.md).

Use that file for:
- completed slices and accepted evidence
- active work
- next slices
- blockers and validation status

Do not duplicate the full live checklist in this README. Keep the implementation truth in the state file and update it in the same slice as any architecture, release-target, or UI-surface change.

Current operator experience summary:
- a new operator lands in AI Organization setup, not a blank assistant thread
- Soma always presents guided starting actions instead of a dead-end blank state
- Team Leads remain visible as the operational leaders Soma works through
- Recent Activity, Automations, Learning, Advisors, and Departments keep readable empty, loading, and failure states without collapsing the workspace

## Default And Advanced Surfaces

Mycelis intentionally supports two separate UX/control layers.

Default Operator Surface:
- Create AI Organization
- Soma-primary workspace
- intent-driven interaction
- Advisors, Departments, Automations, Recent Activity, and Learning & Context
- AI Engine Settings and Response Style as guided, bounded controls

Advanced Architecture / Runtime Surface:
- separate and non-default for operators who understand the system deeply
- organization defaults and inheritance visibility
- department overrides and specialist role bindings
- automation definitions, capability posture, and response-style inheritance
- bundle/config source truth, deployment/env influence, and later runtime availability or distributed execution posture

source-of-truth layers remain separate:
- guided UI settings for bounded operator-visible changes
- bundle/file configuration for reproducible organization defaults and automation truth
- deployment/env overrides for environment-specific provider/media/runtime wiring
- runtime state for the live resolved organization and service posture
- README, V8.1, V8.2, and `V8_DEV_STATE.md` for architecture, release, and implementation truth

Contract rule:
- the default UX must stay simple and intent-first
- the advanced architecture/runtime surface must stay separate, make inheritance legible, and make config origin legible
- the advanced layer must not replace bundle/file/env/runtime truth or collapse the Soma-primary MVP into a config dashboard

Implementation note:
- V8.1 currently ships the default operator surface plus bounded guided controls and inspect-only detail where explicitly called out in `V8_DEV_STATE.md`
- the advanced architecture/runtime surface is now defined as a contract, but it is not fully implemented yet

## Architecture Terms To Operator Terms

Use these translations consistently:

| Architecture term | User-facing term |
| --- | --- |
| Inception | AI Organization |
| Soma Kernel | Soma |
| Team Leads | Team Leads |
| Central Council | Advisors |
| Specialist Teams | Departments |
| Agent Instances / Agent Types | Specialists / Roles |
| Provider Policy / Routing | AI Engine Settings |
| Response Contract | Response Style |
| Identity / Continuity State | Learning & Context |
| Loop Profiles | Automations |
| Learning Loops / reviewed learning | What the Organization is Learning |

Translation rule:
- architecture docs may use the precise runtime terms
- default UI, README summary language, and operator-facing copy should prefer the user-facing terms unless a lower-level contract requires the architecture wording

## Detailed Framework Memory

Use these as the top detailed references when you need the deeper framework contract rather than just the quick-start path.

1. [V8 Runtime Contracts](docs/architecture-library/V8_RUNTIME_CONTRACTS.md)
   - canonical runtime layer contract for organization, Soma, Team Leads, advisors, provider-policy scope, and continuity
2. [V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md)
   - canonical V8.1 release architecture for loops, learning, semantic continuity, capabilities, and the current Soma-primary release posture
3. [V8.2 Production Architecture Target](v8-2.md)
   - final production architecture target for distributed execution, active learning, capability-backed execution, and editable automations
4. [V8 Config and Bootstrap Model](docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md)
   - canonical memory for template vs instantiated organization, bootstrap resolution, inheritance, precedence, and V7-to-V8 bootstrap translation
5. [V8 UI/API and Operator Experience Contract](docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
   - canonical operator workflow contract for AI Organization creation, Soma-primary workspace behavior, visibility boundaries, and screen-to-API mapping
6. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
   - canonical map of which detailed planning doc owns which part of the framework
7. [System Architecture V7](docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md)
   - detailed runtime, storage, NATS, deployment, and service-boundary memory until V8 replacements land
8. [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
   - detailed workflow, run, manifest, recurring-plan, and activation memory
9. [Delivery Governance And Testing V7](docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
   - detailed acceptance, gate, and proof requirements for implementation slices
10. [Team Execution And Global State Protocol V7](docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md)
   - detailed state-file, coordination, and execution-discipline memory for multi-slice work
11. [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md)
   - detailed operator experience, simplification, and anti-complexity memory for the UI layer
12. [UI Target And Transaction Contract V7](docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
   - detailed UI transaction/state expectations for operator-visible behavior

Rule:
- when framework behavior, bootstrap posture, organization shape, or operator model is unclear, load the owning detailed doc above before making assumptions
- keep README as the entrypoint and inception summary, but treat the documents in this section as the deeper memory surface for framework specifics
- current MVP UI release posture is Soma-primary by default; `Resources`, `Memory`, and `System` are advanced support routes rather than default operator entrypoints

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
- release-target alignment between README, V8.1, V8.2, and `V8_DEV_STATE.md`
- execution slices
- team execution protocol
- delivery governance rules
- centralized review logging: team-local output stays on canonical team lanes, but Soma, meta-agentry, and team leads reason over the mirrored `log_entries` review stream plus mission events
- UI operator experience contracts
- runtime orchestration assumptions
- provider routing and hybrid deployment posture
- when working in the execution/gov docs (`docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md`, `docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md`, `docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md`), treat all V7 content as migration input: translate assets through `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, keep `Template ≠ instantiated organization`, and record slice state in `V8_DEV_STATE.md`

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
- deployment automation may override provider/model/profile/media config through env vars using `MYCELIS_PROVIDER_<PROVIDER_ID>_*`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, and `MYCELIS_MEDIA_*`; use that path instead of the retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` env maps
- env overrides are deployment-time infrastructure wiring, not runtime organization behavior: they define provider instances, profile defaults, and environment-specific endpoints or model ids
- env overrides must not become a shadow runtime architecture: team, role, and agent routing truth still comes from `Bundle -> Instantiated Organization -> Inheritance -> Routing`
- keep repo-managed caches under `workspace/tool-cache` and use `cache.apply-user-policy` when a Windows user profile needs heavy tool caches moved off `C:`
- check `uv run inv cache.status` before large build/test/browser runs when disk headroom is tight; the main repo-local growth surfaces are `workspace/tool-cache`, `interface/.next`, Playwright browser binaries, and other generated test artifacts
- use `uv run inv cache.clean` as the first repo-safe reclaim path when builds or tests start failing under disk pressure instead of manually deleting random working files
- local lifecycle tasks target the bridged Core API on `localhost:8081` by default unless `MYCELIS_API_PORT` overrides it
- Interface tasking now separates the bind host from the local probe host: by default the UI binds on `[::]:3000` for dual-stack LAN reachability, while local checks and browser tooling target `127.0.0.1:3000` unless `MYCELIS_INTERFACE_BIND_HOST` / `MYCELIS_INTERFACE_HOST` override that split
- expect Invoke-managed Interface build/test/browser tasks to sweep repo-local Next/Vitest/Playwright worker residue after each run so old `node.exe` workers do not accumulate between sessions
- expect Invoke-managed Interface and CI tasks to execute from the `interface/` working directory through the same `npm`/`node` entrypoints on Windows and Linux rather than relying on shell-specific `cd ... &&` wrappers
- project-owned config backstops now keep direct local commands aligned too: root `.npmrc` anchors npm/npx cache in `workspace/tool-cache`, pytest stores cache metadata in `workspace/tool-cache/pytest`, and task-managed Interface runs disable Next telemetry while routing Playwright browser binaries through the managed cache root

Suggested development build configuration by platform:
- Windows: keep repo work on a spacious non-system drive when possible, use `uv run inv cache.apply-user-policy` so uv/pip/npm/go/Playwright stop drifting back onto `C:`, and treat Docker Desktop / WSL storage separately from repo-managed cache cleanup
- Linux/macOS: keep `MYCELIS_PROJECT_CACHE_ROOT` on a volume with headroom if the default workspace disk is small, and export user-level cache roots only when you need tool caches off your default home volume; the repo task path already keeps build/test/browser churn inside `workspace/tool-cache`
- All platforms: prefer `uv run inv ...` over raw tool commands for repeated build/test cycles, because the task path applies the managed cache roots, disables low-value telemetry writes, and sweeps leftover Interface workers that can hold build outputs open

Deployment guidance by host architecture:
- Windows x86_64: supported as the main local development and operator host; prefer repo-managed caches on a non-system drive and treat Docker Desktop / WSL storage as a separate disk budget
- Linux x86_64: preferred for longer-running single-host or containerized deployments when you want the cleanest service-host posture for Core, Postgres, NATS, and chart-driven deployment
- Linux arm64: use for lighter edge/control-host roles or remote-provider-connected deployments; do not assume local heavyweight model serving on small ARM hosts unless you have verified headroom
- Mixed-architecture deployments: keep runtime truth bundle-driven, use env overrides only for deployment-time provider/profile/media wiring, and point smaller hosts at remote Ollama or hosted providers instead of reintroducing host-local routing hacks
- Build rule: binaries, images, and provider endpoints must be selected for the target host architecture; do not assume a Windows dev build or amd64 image is the correct artifact for an arm64 deployment

## Development Contract

A slice is not complete unless:
- tests pass
- documentation is updated where meaning changed
- architecture alignment is verified across the layered truth surfaces

Deployment/runtime boundary:
- README is the primary architecture inception document for active work
- `v8-2.md` is the canonical full architecture target
- V8.1 is the current release target
- `V8_DEV_STATE.md` is the source of actual implementation truth
- deployment env overrides configure infrastructure and profile defaults, but they do not replace bundle-defined runtime organization truth

Development contract:
- `README.md` is the primary architecture inception document
- `v8-2.md` is the canonical full architecture
- `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` is the current release target
- `V8_DEV_STATE.md` is the source of actual implementation truth
- all slices must update these surfaces when implementation, release posture, or target meaning changes

Completion rule:
- code and tests alone do not finish a slice
- if implementation changes meaning, the slice must also update the owning docs, verify architecture alignment, and record the resulting state
- no slice should complete with silent divergence between implementation, V8.1 release scope, V8.2 target scope, and `V8_DEV_STATE.md`
- end-of-slice reporting must explicitly state which tests ran, which docs changed, and which scoped docs were reviewed but left unchanged

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
2. review the layered architecture truth in README, the owning architecture doc, and `V8_DEV_STATE.md`
3. review V7 architecture-library documentation as migration input when a V8 replacement has not fully landed yet
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
- docs tests when the contract they enforce changes

Synchronization rule:
- README is the primary architecture inception doc for active development
- V8.1 is the current release target
- V8.2 is the final production target
- `V8_DEV_STATE.md` is the actual implementation scoreboard
- slices that change architecture, release posture, operator wording, or documentation authority must keep README, the owning docs, `docsManifest.ts`, and `tests/test_docs_links.py` synchronized in the same change
- slice close-out should explicitly report tests run, docs updated, and docs reviewed unchanged for the touched scope

The architecture-library remains the authoritative detailed planning surface until the V8 library replaces the remaining V7 migration inputs.

## Status

Mycelis is currently shipping toward a V8.1 Soma-primary MVP while aligning its long-range architecture to the V8.2 distributed, learning, capability-governed production target.

The V7 system still provides important migration input and substrate memory.
The V8.1 release defines what belongs in the current MVP.
The V8.2 PRD defines where the product is ultimately headed.
`V8_DEV_STATE.md` records what is actually complete, active, next, or blocked right now.

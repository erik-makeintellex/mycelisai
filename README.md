# Mycelis

Mycelis is an AI Organization platform for creating, operating, and evolving governed multi-role systems through a Soma-primary operator experience.

This README is the primary development-swarm inception document. It defines the top-level architecture truth split for active work:
- active V8.2/B2+ delivery target
- V8.1 foundation and compatibility baseline
- current implementation state

Canonical ownership reminder:
- `README.md` = development-swarm inception and layered truth
- `architecture/v8-2.md` = canonical V8.2 full actuation / production target and current B2+ delivery frame
- `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` = V8.1 foundation and compatibility baseline
- `.state/V8_DEV_STATE.md` = actual implementation truth
- `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md` = UX/operator truth

## README TOC

- [Fresh Agent Start Here](#fresh-agent-start-here)
- [User Guidance](#user-guidance)
- [Agent Guidance](#agent-guidance)
- [What Mycelis Is](#what-mycelis-is)
- [Active Delivery Target (V8.2 B2+)](#active-delivery-target-v82-b2)
- [V8.1 Foundation And Compatibility Baseline](#v81-foundation-and-compatibility-baseline)
- [Current Implementation State](#current-implementation-state)
- [Compact Team Orchestration](#compact-team-orchestration)
- [Default And Advanced Surfaces](#default-and-advanced-surfaces)
- [Architecture Terms To Operator Terms](#architecture-terms-to-operator-terms)
- [Detailed Framework Memory](#detailed-framework-memory)
- [Feature Status Standard](#feature-status-standard)
- [Required Review Targets](#required-review-targets)
- [Command Contract](#command-contract)
- [Development Contract](#development-contract)
- [Playwright Contract](#playwright-contract)
- [Testing Gate](#testing-gate)
- [Fastest Start](#fastest-start)
- [Cross-Platform Setup](#cross-platform-setup)
- [Development Workflow](#development-workflow)
- [Licensing & Editions](#licensing-editions)
- [Binary Releases](#binary-releases)
- [Documentation Responsibilities](#documentation-responsibilities)
- [Status](#status)

## Fresh Agent Start Here

Review these in order before touching code or planning state:

1. [AGENTS.md](AGENTS.md)
2. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
3. [README](README.md)
4. [V8.2 Production Architecture Target](architecture/v8-2.md)
5. [V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md)
6. [V8 Runtime Contracts](docs/architecture-library/V8_RUNTIME_CONTRACTS.md)
7. [V8 Config and Bootstrap Model](docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md)
8. [V8 UI/API and Operator Experience Contract](docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
9. [Target Deliverable V7](docs/architecture-library/TARGET_DELIVERABLE_V7.md)
10. [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
11. [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md)
12. [UI Target And Transaction Contract V7](docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
13. [Operations](docs/architecture/OPERATIONS.md)
14. [Testing](docs/TESTING.md)
15. [Cognitive Architecture](docs/COGNITIVE_ARCHITECTURE.md)
16. [V8 UI Testing Agentry Product Contract](docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md)
17. [V8 UI Team Full Test Set](docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md)
18. [V8 MVP Media, Team Output, And Template Registry](docs/architecture-library/V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md)
19. [V8.2 User Management And Enterprise Auth Module](docs/architecture-library/V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md)
20. [V8 Development State](.state/V8_DEV_STATE.md)
21. [V7 Development State (Historical)](.state/V7_DEV_STATE.md)
22. [Docs Manifest](interface/lib/docsManifest.ts)

Fresh-agent review rule:
- README is the primary architecture inception document for active development.
- [V8.2 Production Architecture Target](architecture/v8-2.md) is the canonical full production target, full actuation architecture, and current B2+ delivery frame.
- [V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) is the foundation and compatibility baseline for the Soma-primary operator surface.
- [V8 Development State](.state/V8_DEV_STATE.md) is the live implementation scoreboard.
- V7 documents remain migration inputs until replaced, but they do not override the V8 bootstrap and release truth.

Windows dev + WSL proof rule:
- treat the Windows repo as the active edit, review, and git-push surface for day-to-day development work
- treat the `mother-brain` WSL checkout backed by `D:\\wsl-distro` as the authoritative deployment-mimic proof checkout for install, build, test, runtime bring-up, and release-style validation
- use `.env.compose` plus the Compose task path first in that WSL proof checkout, not the Windows Kind/lifecycle flow, unless the slice is explicitly Windows-native or Kubernetes-specific
- refresh the WSL proof checkout from git after a Windows-side commit/push instead of copying source or generated artifacts across the boundary
- do not share one long-lived generated environment across Windows and WSL; recreate `.venv`, `interface/node_modules`, and `interface/.next` in the WSL proof checkout before trusting results
- use the dedicated WSL task lane when you want the guarded handoff/proof flow from Windows: `uv run inv wsl.status`, `uv run inv wsl.refresh`, `uv run inv wsl.validate`, and `uv run inv wsl.cycle`
- `uv run inv wsl.refresh` runs WSL git fetch noninteractively, tries a repo-local Git Credential Manager helper repair for GitHub HTTPS remotes when Git for Windows is visible from WSL, and otherwise fails before reset/clean with SSH/HTTPS auth guidance; its source cleanup preserves generated `workspace/tool-cache`, `workspace/logs`, and `workspace/docker-compose` roots so permission-owned runtime/cache mounts do not block the git handoff; keep the handoff git-backed rather than copying source across the host boundary
- `uv run inv wsl.validate` now bootstraps `.env.compose` from `.env.compose.example` when the clean WSL proof checkout has no local compose env yet, ensures the configured Compose output-block host path exists, loads that Compose env into the managed Interface proxy/browser proof path, then runs Compose-safe release-preflight, Compose health/storage proof, focused live-backend browser workflows, and the Windows-side GUI probe in one guarded pass; when invoked with `--lane=service` or `--lane=release`, it uses the runtime preflight first because service/browser proof is owned by the Compose sequence in this task

Bootstrap reminder:
- treat `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` as the canonical V7->V8 migration and bootstrap contract, not just another planning note
- always translate V7 YAML, runtime config, DB seeding, and operator wizard flows through that model before they touch a live organization
- `Template ≠ instantiated organization`, so only instantiated orgs enter bootstrap resolution while templates stay reusable blueprints
- startup truth is bundle-driven and fail-closed: normal startup fails closed unless a valid bootstrap bundle is present, and `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must select a bundle whenever more than one is mounted
- the deployed Core image resolves runtime config from `/core/config`, and the Helm chart mounts the config volume there so bootstrap bundles, cognitive defaults, homepage templates, and policy files line up with the container workdir

Tasking note:
- `uv run inv db.migrate` is a forward-bootstrap task for schemas that are not yet compatible with the current runtime; if the `cortex` schema already has the required runtime tables and columns, it skips replay and points operators to `uv run inv db.reset` for a clean rebuild
- `uv run inv compose.up` and `uv run inv compose.migrate` follow the same compatibility posture for the supported home-runtime stack: they replay canonical forward migrations only when the compose `cortex` schema is not yet compatible with the current runtime, and otherwise keep the existing schema in place until you intentionally use `uv run inv compose.down --volumes` for a fresh rebuild
- `uv run inv compose.infra-up` starts only the Compose data plane (`postgres` + `nats`), leaves Core/Interface down, and prints owner-facing connection settings for same-project containers, host-native tools, and separate Compose app projects; `uv run inv compose.infra-health` probes that data plane without checking Core/UI
- `uv run inv compose.storage-health` is the post-migration long-term storage gate for Compose PostgreSQL: it checks pgvector plus the durable semantic/context, memory, conversation, artifact, managed-exchange, group, and template tables Mycelis uses for long-horizon recall and retained outputs
- `uv run inv compose.up` now emits deterministic numbered stages with expectations and recovery guidance, and accepts `--wait-timeout=<seconds>` when a slower host or first rebuild needs longer readiness windows
- Compose output storage is explicitly configurable: `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted` with `MYCELIS_OUTPUT_HOST_PATH=<host-directory>` mounts a local or Pinokio/media-hosted output block into Core as `/data`, while Kubernetes/chart usage should keep output generation on the cluster-managed PVC path. The Invoke compose task resolves host paths with Python `pathlib` across Windows, Linux, and macOS before startup.
- Compose and Helm expose the same governed search knobs: `MYCELIS_SEARCH_PROVIDER=disabled|local_sources|searxng|local_api|brave`, `MYCELIS_SEARXNG_ENDPOINT=<self-hosted endpoint>`, `MYCELIS_SEARCH_LOCAL_API_ENDPOINT=<self-hosted HTTP endpoint>`, and `MYCELIS_SEARCH_MAX_RESULTS=<n>`. `local_sources`, self-hosted `searxng`, and self-hosted `local_api` do not require Brave tokens; they only configure the runtime path and still require any chosen endpoint to be reachable.
- on Windows hosts that no longer have a native `docker` binary, the Compose task path now falls back to Docker inside WSL when available, translates the compose file/env file paths into WSL form, and rewrites `MYCELIS_OUTPUT_HOST_PATH` for the container runtime; use `MYCELIS_WSL_DISTRO` when the default WSL distro is not the Docker host
- on that same Windows + WSL Docker path, `compose.up` and `compose.health` now auto-start a WSL-host relay for the AI endpoint when needed, including when the operator runs the tasks directly inside the Docker-owning WSL distro, so Core can still use a Windows-hosted Ollama service through `host.docker.internal` even when bridge containers cannot reach the Windows LAN IP directly
- `uv run inv k8s.deploy` now accepts `MYCELIS_K8S_TEXT_ENDPOINT` and `MYCELIS_K8S_MEDIA_ENDPOINT` so the Helm deploy path can target an explicit external AI service such as a Windows-hosted Ollama box without editing chart source; use a reachable host/IP like `http://192.168.x.x:11434/v1`, not `localhost`
- `uv run inv k8s.deploy` and `uv run inv k8s.up` also accept `MYCELIS_K8S_VALUES_FILE` so operators can apply promoted Helm preset files such as `charts/mycelis-core/values-k3d.yaml`, `charts/mycelis-core/values-enterprise.yaml`, or `charts/mycelis-core/values-enterprise-windows-ai.yaml` without patching the chart
- the `charts/mycelis-core/values-enterprise-windows-ai.yaml` preset now fails closed unless `MYCELIS_K8S_TEXT_ENDPOINT` is set to a real Windows-hosted AI endpoint; do not treat the placeholder value as deployable
- `uv run inv k8s.init` / `k8s.up` / `k8s.deploy` now prefer `k3d` as the local Kubernetes backend when it is available, while keeping `MYCELIS_K8S_BACKEND=kind` as the explicit fallback for older local workflows
- `uv run inv ci.release-preflight --runtime-posture` now adds a tighter runtime gate before baseline proof: 12 GiB disk headroom, explicit AI-endpoint discovery from process env plus `.env.compose` / `.env`, fail-closed behavior when no supported non-loopback endpoint contract is configured, and WSL-host probe mirroring for `host.docker.internal` so the Windows-local Ollama contract can still be verified before the Compose relay starts
- `uv run inv quality.max-lines` now enforces the source-tree `300` LOC no-regression contract by default; existing oversized files are tracked in `ops/quality_legacy_caps.txt`, stale caps fail the gate, and cap values should only move downward as modules are split.
- live Playwright proof that asserts filesystem side effects may need `MYCELIS_BACKEND_WORKSPACE_ROOT` (or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT`) when the browser tests run from a different worktree than the live Core backend; use the backend's actual workspace root, for example `core/workspace` for a repo-local Core process or `workspace/docker-compose/data/workspace` for the supported compose stack
- deployer-editable homepage framing lives in `core/config/homepage.yaml` or `MYCELIS_HOMEPAGE_CONFIG_PATH`; Core serves only sanitized copy/link fields through `/api/v1/homepage`, and Helm mounts the same template through the chart ConfigMap for self-hosted and enterprise branded portals.
- workspace-relative file/tool requests may use `workspace/...` as a readable alias for the configured workspace root; runtime normalization strips that leading alias instead of nesting a second `workspace` directory under `MYCELIS_WORKSPACE`
- the supported home-runtime Docker Compose path uses `.env.compose`, not `.env`; use `MYCELIS_COMPOSE_OLLAMA_HOST` there so host-level `OLLAMA_HOST` settings cannot leak into the container runtime, point it at a host-reachable endpoint such as `http://host.docker.internal:11434`, and let Compose map that value into the provider-specific runtime overrides inside Core
- the supported Core container image includes Node/npm/npx for curated stdio MCP servers; default MCP auto-bootstrap still stays disabled in Compose, but a manual `filesystem` library install must be able to launch and bind to the configured `/data/workspace` output block
- when Docker runs inside WSL and the AI engine is on the same Windows host, keep `MYCELIS_COMPOSE_OLLAMA_HOST` pointed at the intended Windows service address; the task layer can relay that through the WSL host so bridge containers do not need direct access to the Windows LAN IP

## User Guidance

For someone using Mycelis rather than changing the codebase, start here:

1. [Docs Navigation](docs/README.md)
2. [Deployment Method Selection](docs/user/deployment-methods.md)
3. [Core Concepts](docs/user/core-concepts.md)
4. [Using Soma Chat](docs/user/soma-chat.md)
5. [Teams](docs/user/teams.md)
6. [Governance & Trust](docs/user/governance-trust.md)
7. [Automations](docs/user/automations.md)
8. [Resources](docs/user/resources.md)
9. [Memory](docs/user/memory.md)
10. [Settings And Access](docs/user/settings-access.md)
11. [System Status & Recovery](docs/user/system-status-recovery.md)
12. [Run Timeline](docs/user/run-timeline.md)

Operator guidance rule:
- user guidance should explain how to get value from Soma, organizations, approvals, artifacts, and settings without assuming architecture knowledge
- user guidance should include concrete Soma request wording for web search, temporary-team creation, team communication, host-data use, and MCP structure review instead of abstract capability labels only
- the in-app `/docs` surface should expose the same canonical operator-facing documents through `interface/lib/docsManifest.ts`

## Agent Guidance

For someone implementing, reviewing, or testing Mycelis, use this authority order:

1. [AGENTS.md](AGENTS.md)
2. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
3. [V8 Development State](.state/V8_DEV_STATE.md)
4. [Testing Guide](docs/TESTING.md)
5. [Operations](docs/architecture/OPERATIONS.md)
6. [Docs Navigation](docs/README.md)

Agent guidance rule:
- agent/developer guidance should point contributors to current V8 authority and live implementation truth before historical V7 material
- user guidance, agent guidance, and testing guidance should stay cross-linked but distinct so the product docs do not read like implementation notes and the implementation docs do not read like operator help

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

## Active Delivery Target (V8.2 B2+)

The active delivery target is [V8.2 Production Architecture Target](architecture/v8-2.md), the full actuation architecture for Mycelis. The repo is now operating in a V8.2/B2+ delivery frame: work may advance V8.2 modules directly, provided each module keeps a named boundary, proof lane, and promotion rule.

V8.2 is the distributed end-state and active architecture frame we are building through:
- distributed execution across a control plane and execution nodes
- an active learning system that can evaluate, promote, and reuse reviewed learning safely
- a capability system that governs what execution surfaces agents may use
- editable automations, policy-bounded actuation, and stronger continuous operation

V8.2 summary:
- distributed execution means the AI Organization can coordinate work across more than one host or environment
- the learning system turns reviewed outcomes into governed memory, reusable procedures, and safer organization improvement
- the capability system keeps action surfaces allowlisted, scoped, auditable, and policy-checked
- user-level governance profiles, approval thresholds, and audit trails make enterprise-style traceability available even in the single-user free-node posture
- multi-user enterprise identity grows from that foundation: federated users, local break-glass admins, and one shared organization-owned Soma persona must coexist without fragmenting memory, RAG, privacy, or audit

Explicit distinction:
- V8.2 B2+ is the active delivery target and full production architecture frame
- V8.1 remains the foundation and compatibility baseline for the current Soma-primary operator surface

V8.2 B2+ does not mean every full-production feature is implemented or promoted into the default UI. It means active work should speak from the V8.2 architecture, keep B2+ modules explicit, and preserve the simple Soma-first operator flow until a capability is intentionally promoted.

## V8.1 Foundation And Compatibility Baseline

[V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) is the foundation and compatibility baseline for the Soma-primary operator experience, not the place to add new B2+ target scope.

V8.1 defines the operator baseline we must not regress: a Soma-primary AI Organization system with bounded automation visibility, memory and continuity visibility, inheritance contracts, and safe organization structure surfaces.

V8.2/B2+ work should keep flowing, but it must stay modular: each V8.2-aligned slice should advance one named boundary, keep default operator surfaces simple unless explicitly promoted, and preserve the current proof lane.

Included in the V8.1 compatibility baseline:
- AI Organization creation and Soma-primary workspace flow
- a primary Soma conversation surface for discussing plans, samples, and delivery intent
- startup and runtime checks that keep Soma bound to an available AI Engine or surface explicit setup guidance before generic interaction failures
- conversation outputs from Soma may include playable imagery, audio, video, and rich artifacts generated directly by Soma or returned from consulted specialists
- a managed exchange foundation for governed channels, threads, schemas, and normalized outputs beyond single-model request/response
- foundational managed exchange security for channel visibility, thread participation, artifact sensitivity, capability risk classes, and trust-classified external outputs
- guided first-run Soma actions that help a new operator choose a visible next step quickly
- organization, department, advisor, and role-type visibility in operator language
- bundle-driven startup truth with policy-bounded inheritance
- bounded AI Engine and Response Style controls
- read-only Automations visibility
- read-only retained-knowledge visibility
- governed deployment-context intake for user-private content, customer-provided material, approved company-authored guidance, admin-shaped Soma context, and reflection/synthesis memory, kept separate from ordinary Soma memory
- intentional empty states, retry guidance, and partial-failure-safe workspace panels
- Loop Profiles, Runtime Capabilities, semantic continuity, and Procedure / Skill Sets defined as architecture truth even where implementation is still partial

B2+ modules that now belong to the active V8.2 frame but still require explicit proof/promotion:
- distributed multi-host execution
- editable automations
- broad live actuation
- governed capability controls beyond current approved paths
- reviewed memory promotion and learning behavior beyond bounded visibility
- advanced architecture/runtime panels outside the default operator flow
- full enterprise multi-user IAM, federated SAML/OIDC/SSO, optional lifecycle sync, and delegated enterprise admin/recovery flows; the target module is [V8.2 User Management And Enterprise Auth Module](docs/architecture-library/V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md)

Promotion rule:
- if a surface belongs to V8.2/B2+ but is not ready for the default operator experience, it should remain behind an explicit module boundary or advanced surface until promoted
- B2+ work must name its boundary: runtime/deployment, memory/learning, team/workflow, capability/MCP, advanced UI, user/auth, or governance/trust

## Current Implementation State

Actual implementation state lives in [.state/V8_DEV_STATE.md](.state/V8_DEV_STATE.md).

Use that file for:
- completed slices and accepted evidence
- active work
- next slices
- blockers and validation status

Do not duplicate the full live checklist in this README. Keep the implementation truth in the state file and update it in the same slice as any architecture, release-target, or UI-surface change.

Current operator experience summary:
- a new operator lands on "What do you want Soma to do?", and Soma remains the default workspace route with simple Plan, Research, Create, Review, and Configure tools intent cards
- the default AI Organization workspace now includes one primary Soma interaction surface with mode switching inside the same panel instead of parallel front doors
- the default Soma conversation hides raw broadcast and direct-council routing controls until Advanced mode is intentionally opened
- that same Soma conversation surface is also the canonical operator output lane for playable media, briefs, charts, code, scheduled/continuous task proposals, and other rich artifacts, even when specialist or council paths generated them on Soma's behalf
- team design now lives as a guided Soma mode inside that same workspace so the operator can stay in one continuous interaction context
- team design defaults to small focused teams; broad asks should split into compact lanes orchestrated by Soma with Council help over NATS and managed exchange, including current-team or multi-team agentry wiring when configured
- advanced Resources support now includes an inspect-only Managed Exchange surface for channels, active threads, and recent normalized outputs
- advanced Resources now also includes Deployment Context intake so operators can upload/paste private user records, customer context, approved company knowledge, and admin-shaped Soma guidance into governed pgvector stores with explicit trust, sensitivity, visibility, and target-goal posture
- Soma always presents guided starting actions instead of a dead-end blank state
- Groups, Activity/Runs, Resources/MCP, Memory, System, audit-style review, and deeper settings stay advanced/admin support surfaces rather than first-level concepts for new users
- Team Leads remain visible as the operational leaders Soma works through
- a visible `Soma just did this` strip now ties the last action to engaged teams, generated outputs, and updated support panels
- Departments now show visible specialist-role summaries instead of only counts in the default workspace
- People & Access now also exposes a reviewable deploy-owned access model so investors and operators can see the layered product story clearly: self-hosted release, self-hosted enterprise, or hosted admin control plane, plus the intended identity mode and who controls shared Soma output specificity
- the Soma workspace keeps the in-progress request draft and the last guided outcome visible when the operator leaves and returns to the same AI Organization
- the organization-wide AI Engine and Response Style chosen during setup shape Soma's initial working posture, while the assistant name remains operator-configurable
- AI Engine providers now carry a safe default token budget profile and max output budget so local and hosted agentry usage stays bounded by configuration instead of hidden hardcoded limits
- Recent Activity, Automations, Memory & Continuity, Advisors, and Departments keep readable empty, loading, and failure states without collapsing the workspace
- the default theme now uses a calmer light neutral surface with lower visual harshness while keeping advanced/runtime surfaces visually consistent

## Default And Advanced Surfaces

Mycelis intentionally supports two separate UX/control layers.

Default Operator Surface:
- Create AI Organization
- Soma-primary workspace
- intent-driven interaction
- progress/outcome feedback: what Soma understood, what Soma is doing, what changed, and where output was stored
- Advisors, Departments, Automations, Recent Activity, and Memory & Continuity
- AI Engine Settings and Response Style as guided, bounded controls

Advanced Architecture / Runtime Surface:
- separate and non-default for operators who understand the system deeply
- Groups, Activity/Runs, Resources/MCP, Memory, System, auth, and audit-style administration
- organization defaults and inheritance visibility
- department overrides and specialist role bindings
- automation definitions, capability posture, and response-style inheritance
- bundle/config source truth, deployment/env influence, and later runtime availability or distributed execution posture

source-of-truth layers remain separate:
- guided UI settings for bounded operator-visible changes
- bundle/file configuration for reproducible organization defaults and automation truth
- deployment/env overrides for environment-specific provider/media/runtime wiring
- runtime state for the live resolved organization and service posture
- README, V8.2, the V8.1 baseline, and `.state/V8_DEV_STATE.md` for architecture, delivery, compatibility, and implementation truth

Managed exchange note:
- channels, typed schemas, structured fields, threads, and normalized artifacts are now part of runtime truth for advanced orchestration
- this exchange layer is how Soma, Team Leads, specialist roles, automations, and MCP-backed systems share governed outputs without falling back to raw message blobs
- managed exchange is permissioned: channels, threads, and artifacts now carry reader/writer/reviewer, visibility, sensitivity, capability, and trust metadata
- normalization into managed exchange does not imply unrestricted trust; external and MCP-backed outputs remain governed by trust class, capability policy, and review requirements
- the default operator flow still stays Soma-first; inspect-only exchange visibility belongs to Advanced mode rather than the primary workspace

Managed exchange security note:
- connected tools and external services are governed rather than implicitly trusted
- the free-node release now includes foundational security boundaries for exchange visibility, capability risk, user-level approval posture, audit records, and external trust classification
- env overrides are deployment-time infrastructure wiring, not runtime organization behavior
- do not replace bundle-defined runtime organization truth

Contract rule:
- the default UX must stay simple and intent-first
- the advanced architecture/runtime surface must stay separate, make inheritance legible, and make config origin legible
- the advanced layer must not replace bundle/file/env/runtime truth or collapse the Soma-primary operator flow into a config dashboard

## Compact Team Orchestration

Team creation should stay compact by default:
- a normal launch should start with 3 precise roles: Team Lead, Architect Prime, and focused builder/developer
- a single team should stay at 5 members or fewer, with 4th/5th roles justified by a distinct output need
- broad asks should become several small teams or lane bundles rather than one oversized roster
- Soma remains the root orchestrator that can split work, coordinate lanes, and pull Council in when specialist review helps
- NATS and managed exchange are the communication and observability fabric for that coordination story

The user-facing rule is simple: if the work is broad, the product should split it cleanly instead of scaling a single roster until it becomes hard to understand or test.

Implementation note:
- the current B2+ frame still preserves the V8.1 Soma-primary default operator surface plus bounded guided controls and inspect-only detail where explicitly called out in `.state/V8_DEV_STATE.md`
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
| Identity / Continuity State | Memory & Continuity |
| Loop Profiles | Automations |
| Learning Loops / reviewed learning | What the Organization Is Retaining |

Translation rule:
- architecture docs may use the precise runtime terms
- default UI, README summary language, and operator-facing copy should prefer the user-facing terms unless a lower-level contract requires the architecture wording

## Detailed Framework Memory

Use these as the top detailed references when you need the deeper framework contract rather than just the quick-start path.

1. [V8 Runtime Contracts](docs/architecture-library/V8_RUNTIME_CONTRACTS.md)
   - canonical runtime layer contract for organization, Soma, Team Leads, advisors, provider-policy scope, and continuity
2. [V8.1 Living Organization Architecture](docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md)
   - canonical V8.1 baseline architecture for loops, learning, semantic continuity, capabilities, and the Soma-primary compatibility posture
3. [V8.2 Production Architecture Target](architecture/v8-2.md)
   - final production architecture target for distributed execution, active learning, capability-backed execution, and editable automations
4. [V8 Config and Bootstrap Model](docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md)
   - canonical memory for template vs instantiated organization, bootstrap resolution, inheritance, precedence, and V7-to-V8 bootstrap translation
5. [V8 UI/API and Operator Experience Contract](docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
   - canonical operator workflow contract for AI Organization creation, Soma-primary workspace behavior, visibility boundaries, and screen-to-API mapping
6. [V8 Memory Layer And Reflection Delivery Contract](docs/architecture-library/V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md)
   - canonical memory-layer contract for `SOMA_MEMORY`, `AGENT_MEMORY`, `PROJECT_MEMORY`, `REFLECTION_MEMORY`, candidate-first reflection, and promotion guardrails
7. [V8 Trusted Memory Arbitration And Team Vector Contract](docs/architecture-library/V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md)
   - trusted memory control plane for Soma personal continuity, team-shared vector memory, governed doctrine, evidence anchors, and arbitration order
8. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
   - canonical map of which detailed planning doc owns which part of the framework
9. [System Architecture V7](docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md)
   - detailed runtime, storage, NATS, deployment, and service-boundary memory until V8 replacements land
10. [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
   - detailed workflow, run, manifest, recurring-plan, and activation memory
11. [Delivery Governance And Testing V7](docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
   - detailed acceptance, gate, and proof requirements for implementation slices
12. [Team Execution And Global State Protocol V7](docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md)
   - detailed state-file, coordination, and execution-discipline memory for multi-slice work
13. [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md)
   - detailed operator experience, simplification, and anti-complexity memory for the UI layer
14. [UI Target And Transaction Contract V7](docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
   - detailed UI transaction/state expectations for operator-visible behavior

Rule:
- when framework behavior, bootstrap posture, organization shape, or operator model is unclear, load the owning detailed doc above before making assumptions
- keep README as the entrypoint and inception summary, but treat the documents in this section as the deeper memory surface for framework specifics
- current B2+ UI delivery posture remains Soma-primary by default; `Resources`, `Memory`, and `System` are advanced support routes rather than default operator entrypoints

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
- `.state/V8_DEV_STATE.md`
- `.state/V7_DEV_STATE.md` when historical migration evidence is needed

Particular attention belongs on:
- delivery-target alignment between README, V8.2, the V8.1 baseline, and `.state/V8_DEV_STATE.md`
- trusted memory alignment between memory-layer docs, governance docs, user memory/resources docs, and the in-app docs manifest
- execution slices
- team execution protocol
- delivery governance rules
- centralized review logging: team-local output stays on canonical team lanes, but Soma, meta-agentry, and team leads reason over the mirrored `log_entries` review stream plus mission events
- UI operator experience contracts
- runtime orchestration assumptions
- provider routing and hybrid deployment posture
- when working in the execution/gov docs (`docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md`, `docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md`), treat all V7 content as migration input: translate assets through `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, keep `Template ≠ instantiated organization`, and record slice state in `.state/V8_DEV_STATE.md`

## Command Contract

Required command references for active V8 work:
- `uv run inv install`
- `uv run inv auth.dev-key`
- `uv run inv auth.break-glass-key`
- `uv run inv auth.posture`
- `uv run inv ci.entrypoint-check`
- `uv run inv ci.service-check`
- `uv run inv cache.status`
- `uv run inv cache.guard`
- `uv run inv cache.clean`
- `uv run inv lifecycle.memory-restart`
- `uv run inv compose.infra-up`
- `uv run inv compose.infra-health`
- `uv run inv compose.storage-health`
- `uv run inv compose.up`
- `uv run inv compose.health`
- `uv run inv team.architecture-sync`

Use `uv run inv ...` for execution.
Use `uvx --from invoke inv -l` only as a compatibility probe.
Do not use bare `uvx inv ...`.
- CI validation should reuse the same invoke task surfaces for interface build/test/browser execution after the workflow-native dependency/bootstrap steps complete, but push-triggered GitHub pipeline runs are intentionally paused until initial release readiness so local proof and PR-time validation remain the active gate

Provider/runtime workflow reminders:
- review architecture and state docs before implementation slices
- attach tests and evidence in the same delivery window
- keep state-file updates current with gate results and blocker changes
- use `uv run inv core.compile` when you need the repo-local Core binary refreshed without building a Docker image; reserve `uv run inv core.build` for immutable image work
- `uv run inv install` now installs the supported default Core + Interface stack only; use `uv run inv install --optional-engines` or `uv run inv cognitive.install` when you explicitly need local vLLM/Diffusers helpers
- use `uv run inv interface.typecheck` for managed TypeScript validation instead of ad hoc `npx tsc --noEmit` runs
- deployment automation may override provider/model/profile/media config through env vars using `MYCELIS_PROVIDER_<PROVIDER_ID>_*`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, and `MYCELIS_MEDIA_*`; use that path instead of the retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` env maps
- env overrides are deployment-time infrastructure wiring, not runtime organization behavior: they define provider instances, profile defaults, and environment-specific endpoints or model ids
- env overrides must not become a shadow runtime architecture: team, role, and agent routing truth still comes from `Bundle -> Instantiated Organization -> Inheritance -> Routing`
- token budgets are configuration-owned rather than hardcoded: use `token_budget_profile` and `max_output_tokens` in `core/config/cognitive.yaml`, override them at deploy time with `MYCELIS_PROVIDER_<PROVIDER_ID>_TOKEN_BUDGET_PROFILE` and `MYCELIS_PROVIDER_<PROVIDER_ID>_MAX_OUTPUT_TOKENS`, or manage them from `/settings` -> `AI Engines` in Advanced mode
- safe output-budget presets follow common operational ranges: `conservative=512`, `standard=1024`, `extended=2048`, `deep=4096`; default local providers should stay on `standard` unless the role really needs longer output, while remote or heavier reasoning providers can opt into `extended`
- keep repo-managed caches under `workspace/tool-cache` and use `cache.apply-user-policy` when a Windows user profile needs heavy tool caches moved off `C:`
- check `uv run inv cache.status` before large build/test/browser runs when disk headroom is tight; the main repo-local growth surfaces are `workspace/tool-cache`, `interface/.next`, Playwright browser binaries, and other generated test artifacts
- use `uv run inv cache.guard` when you want a fail-fast preflight on repo/cache disk headroom before repeated builds, Playwright runs, or CI-style local validation
- use `uv run inv cache.clean` as the first repo-safe reclaim path when builds or tests start failing under disk pressure instead of manually deleting random working files
- local lifecycle tasks target the bridged Core API on `localhost:8081` by default unless `MYCELIS_API_PORT` overrides it
- the home-runtime compose path keeps secrets in `.env` and host/container topology in `.env.compose`, including the same host port contract (`3000`, `8081`, `5432`, `4222`) so browser/operator proof does not need a separate port story
- Interface tasking now separates the bind host from the local probe host: by default the UI binds on `[::]:3000` for dual-stack LAN reachability, while local checks and browser tooling target `127.0.0.1:3000` unless `MYCELIS_INTERFACE_BIND_HOST` / `MYCELIS_INTERFACE_HOST` override that split
- when the stack is hosted inside WSL on a Windows machine, treat the Windows browser as the primary self-hosted operator path: first prove `http://localhost:3000` from Windows on that same machine, then use the host/IP path for second-machine or LAN-user proof
- expect Invoke-managed Interface build/test/browser tasks to sweep repo-local Next/Vitest/Playwright worker residue after each run so old `node.exe` workers do not accumulate between sessions; `uv run inv interface.build` now retries once after a stale repo-local Next build lock before failing the task
- expect the Interface Vitest gate to run with file parallelism disabled; the suite contains several full-page jsdom workflows, and the deterministic release gate is more valuable than parallel wall-clock speed
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
- `architecture/v8-2.md` is the canonical full architecture target
- V8.2 B2+ is the active delivery target
- V8.1 is the foundation and compatibility baseline
- `.state/V8_DEV_STATE.md` is the source of actual implementation truth
- deployment env overrides configure infrastructure and profile defaults, but they do not replace bundle-defined runtime organization truth

Development contract:
- `README.md` is the primary architecture inception document
- `architecture/v8-2.md` is the canonical full architecture
- `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` is the V8.1 foundation and compatibility baseline
- `.state/V8_DEV_STATE.md` is the source of actual implementation truth
- all slices must update these surfaces when implementation, release posture, or target meaning changes

Completion rule:
- code and tests alone do not finish a slice
- if implementation changes meaning, the slice must also update the owning docs, verify architecture alignment, and record the resulting state
- no slice should complete with silent divergence between implementation, the V8.2/B2+ delivery target, the V8.1 compatibility baseline, and `.state/V8_DEV_STATE.md`
- end-of-slice reporting must explicitly state which tests ran, which docs changed, and which scoped docs were reviewed but left unchanged
- branches are not ready to merge until the relevant validation is rerun against the final code state for that branch

## Playwright Contract

`uv run inv interface.e2e` owns the local Next.js server lifecycle for browser test runs, routes Playwright browsers through the managed project cache, and leaves no repo-local UI workers behind when it exits. The managed task now defaults to the managed `dev` server and `--workers=1` for stable mocked browser proof. Use `--server-mode=start` when you need the built production Interface server path for stricter or live-backend proof; start-mode runs still refresh the production bundle first, retry once after a stale repo-local Next build lock, a stale `.next/standalone` cleanup lock, incomplete built-server packaging, or transient missing `.next/types` output, and must successfully hold the managed server before the browser run continues. On Windows, the managed Playwright server binds to `127.0.0.1` so the browser, readiness probe, and Next listener use the same loopback family instead of drifting between IPv4 and IPv6. Managed `dev` runs also clear an orphaned `interface/.next/dev/lock` only when no repo-local Next worker remains, so stale local lock files do not block browser proof.

`uv run inv test.e2e` is the root alias for that same managed browser contract and forwards the same `--workers` and `--server-mode` controls when a caller prefers the generic `test.*` surface.

`uv run inv ci.baseline` uses a reduced Playwright worker count (`--workers=1`) and the built production Interface server path so merge-readiness browser proof stays repeatable without stretching the gate into an impractical wall-clock run on local Windows hosts. `uv run inv ci.service-check --live-backend` stays serial (`--workers=1`) because it proves the current live governed Soma browser contract, restores the local bridge/core stack before the browser proof when needed, and reuses an already-initialized `cortex` schema instead of replaying non-idempotent migrations on every run.

Browser matrix baseline:
- `chromium firefox webkit`
- `mobile-chromium` when route/mobile smoke is part of the gate

Documentation rule:
- root and testing docs must not imply that default browser validation depends on a manually pre-started server
- `uv run inv ci.baseline` now includes Playwright by default; use `--no-e2e` only for intentionally narrower local debugging
- default release-candidate browser coverage is MVP-aligned; legacy V7 or raw-endpoint-only specs should stay outside the default gate unless a slice explicitly revives them
- live-backend browser checks are still required when proxy/runtime contracts change
- live service issues belong in the release story too: use `uv run inv ci.service-check` for running-stack verification and `uv run inv ci.release-preflight --lane=release` when a branch changes service/runtime contracts; the narrower flags remain available for custom proof

## Testing Gate

The canonical testing guidance lives in [Testing](docs/TESTING.md), with browser proof boundaries captured in [V8 UI Testing Agentry Product Contract](docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) and [V8 UI Team Full Test Set](docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md).

Use those when you need one release-style pass that ties together:
- repo baseline proof
- stable browser workflow proof
- live service/browser proof
- compose-aware runtime proof when the home-runtime stack is part of the release story
- documentation and state synchronization after the gate finishes

## Fastest Start

For a new user who wants the quickest supported path to a running service:

Need help choosing the right runtime first? Start with [Deployment Method Selection](docs/user/deployment-methods.md).

- Supported user runtime lanes:
  - Windows Docker Desktop Compose: rapid local development and same-machine proof; use the Windows browser on that machine and start with `http://localhost:3000`
  - Windows + WSL Docker Compose: rapid local development and WSL proof; use the Windows browser as the first operator proof path and start with `http://localhost:3000`
  - Kubernetes / Helm clustered deployment: target release lane for self-hosted and enterprise environments; operators should open the UI through the same ingress/hostname/IP they will really use remotely

- Windows Docker Desktop:
  1. `copy .env.compose.example .env.compose`
  2. `uv run inv auth.posture --compose`
  3. `uv run inv install`
  4. optional data-plane proof: `uv run inv compose.infra-up --wait-timeout=180`
  5. optional long-term storage proof after migration: `uv run inv compose.migrate`, `uv run inv compose.storage-health`
  6. `uv run inv compose.up --build --wait-timeout=240`
  7. `uv run inv compose.health`
  8. open `http://localhost:3000` from the Windows browser
- WSL2/Linux/macOS:
  1. `cp .env.compose.example .env.compose`
  2. `uv run inv auth.posture --compose`
  3. `uv run inv install`
  4. optional data-plane proof: `uv run inv compose.infra-up --wait-timeout=180`
  5. optional long-term storage proof after migration: `uv run inv compose.migrate`, `uv run inv compose.storage-health`
  6. `uv run inv compose.up --build --wait-timeout=240`
  7. `uv run inv compose.health`
  8. open `http://localhost:3000` on the same machine, or `http://<host-ip>:3000` from another client
- Kubernetes / Helm clustered deployment:
  1. `cp .env.example .env`
  2. set `MYCELIS_API_KEY` in `.env`, or point the chart at an existing Kubernetes Secret
  3. set `MYCELIS_K8S_VALUES_FILE=charts/mycelis-core/values-enterprise.yaml` or another promoted values file
  4. set `MYCELIS_K8S_TEXT_ENDPOINT=http://<reachable-ai-host>:11434/v1` when the selected values require an external text engine
  5. `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`
  6. `uv run inv k8s.deploy --verify-package --values-file=charts/mycelis-core/values-enterprise.yaml --release-label=enterprise`
  7. deploy the verified chart/values into the target cluster through the platform-owned Helm or GitOps path, then prove the real ingress/hostname/IP from the operator machine
- Windows native source fallback:
  1. `copy .env.example .env`
  2. `uv run inv auth.dev-key`
  3. `uv run inv auth.break-glass-key`
  4. `uv run inv auth.posture`
  5. `uv run inv install`
  6. `uv run inv k8s.up` (`k3d` preferred when installed; use `MYCELIS_K8S_BACKEND=kind` for the legacy fallback)
  7. `uv run inv lifecycle.up --frontend`
  8. `uv run inv lifecycle.health`
  9. open `http://localhost:3000`

Active-code rule for Windows hosts:
- use the Windows repo as the day-to-day editing and push surface
- refresh the WSL `mother-brain` proof checkout from git after each committed Windows-side slice instead of copying files or generated artifacts between hosts
- use `uv run inv wsl.status` to confirm branch/commit drift and `uv run inv wsl.refresh` when you want the guarded git-only proof-checkout reset from Windows
- if `wsl.refresh` cannot authenticate from the WSL checkout yet, follow its SSH/HTTPS remediation guidance and keep the boundary git-backed rather than copying files between Windows and WSL
- use `uv run inv wsl.validate` after refresh when you want the WSL proof checkout to self-seed `.env.compose` from the tracked example if needed, create the configured Compose output-block host path, load the same Compose auth/proxy env into the managed Interface proof path, run Compose-safe release-preflight plus Compose health/storage proof, and then certify the live Soma, team-creation, groups, and workspace browser flows before probing `http://localhost:3000` from Windows; `--lane=service` and `--lane=release` still run the WSL Compose/browser gates, but their preflight step is intentionally `--lane=runtime`
- run the full build/test/runtime proof from WSL, then use the Windows-side browser as the first operator proof path against that WSL-hosted stack
- keep the Windows-native source path only for explicit local-Kubernetes/source validation or host-specific debugging

What the user still needs on the host:
- Docker
- Node.js
- Go
- `uv`
- PostgreSQL tooling (`psql`)
- an available AI provider endpoint such as local Ollama

Bootstrap note:
- `uv run inv install` now provisions the supported Core + Interface dependency set plus the managed Playwright Chromium binary needed for the repo’s supported browser-proof lane.

## Cross-Platform Setup

Recommended host posture:
- Windows repo: canonical editing, review, and git-push surface for active contributors
- WSL `mother-brain` checkout backed by `D:\\wsl-distro`: canonical rapid proof lane for build, test, Compose runtime, and pre-cluster validation
- Windows Docker Desktop: rapid local Compose runtime and Windows-browser operator lane on one machine
- Windows native source mode: best for explicit local-Kubernetes or host-specific source validation; use Ollama locally or point at remote providers
- Linux server or platform hosts: use Kubernetes/Helm for target deployment, then prove the UI through the same ingress/host/IP operators will use remotely
- Linux GPU hosts: optional `cognitive.*` helpers are appropriate there when you intentionally want local vLLM/Diffusers

Deployment guidance by target environment:
- Docker Compose: rapid local development, demo, and same-machine proof runtime; it is not the target clustered deployment contract
- `k3d`: preferred local Kubernetes validation lane for Helm and cluster behavior
- Kubernetes / Helm: target self-hosted and enterprise deployment contract using standard Kubernetes resources, promoted values, real registry, Secret, Ingress, NetworkPolicy, PVC, probe, and workload-security settings
- packaged binary or node-attached service: fit for small Linux nodes or edge/control-host roles that should point at a remote AI service

Promoted Kubernetes preset files:
- `charts/mycelis-core/values-k3d.yaml`: local `k3d` validation
- `charts/mycelis-core/values-enterprise.yaml`: enterprise-shaped self-hosted cluster
- `charts/mycelis-core/values-enterprise-windows-ai.yaml`: enterprise-shaped self-hosted cluster with an explicit Windows-hosted AI endpoint

Kubernetes standards gate:
- run `uv run inv k8s.standards` for the static open-standard chart contract
- run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml` before release packaging or cluster handoff
- the gate checks for standard Helm/Kubernetes surfaces: Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy, probes, non-root security context, image pull/digest posture, and cluster-managed output storage

Recommended easiest setup path by host:
- Windows Docker Desktop:
  1. `copy .env.compose.example .env.compose`
  2. `uv run inv auth.posture --compose`
  3. `uv run inv install`
  4. optional data-plane proof: `uv run inv compose.infra-up --wait-timeout=180`
  5. optional long-term storage proof after migration: `uv run inv compose.migrate`, `uv run inv compose.storage-health`
  6. `uv run inv compose.up --build --wait-timeout=240`
  7. `uv run inv compose.health`
  8. open `http://localhost:3000` from the Windows browser
- WSL2/Linux/macOS:
  1. `cp .env.compose.example .env.compose`
  2. `uv run inv auth.posture --compose`
  3. `uv run inv install`
  4. optional data-plane proof: `uv run inv compose.infra-up --wait-timeout=180`
  5. optional long-term storage proof after migration: `uv run inv compose.migrate`, `uv run inv compose.storage-health`
  6. `uv run inv compose.up --build --wait-timeout=240`
  7. `uv run inv compose.health`
  8. open `http://localhost:3000` on the same machine, or `http://<host-ip>:3000` from another client
- Kubernetes / Helm clustered deployment:
  1. `cp .env.example .env`
  2. configure `.env` as the local secret store for task-owned secrets, or use chart `existingSecret` values for cluster-owned credentials
  3. `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`
  4. `uv run inv k8s.deploy --verify-package --values-file=charts/mycelis-core/values-enterprise.yaml --release-label=enterprise`
  5. deploy the verified chart into the target cluster through the platform-owned Helm/GitOps path
  6. prove the real ingress/hostname/IP from the operator machine
- Windows native source validation:
  1. `copy .env.example .env`
  2. `uv run inv install`
  3. `uv run inv k8s.up` (`k3d` preferred when installed; use `MYCELIS_K8S_BACKEND=kind` for the legacy fallback)
  4. `uv run inv lifecycle.up --frontend`
  5. `uv run inv lifecycle.health`

Cross-host artifact rule:
- do not bounce the same working copy back and forth between Windows Python/Node artifacts and WSL/Linux/macOS artifacts
- safest posture is one Windows dev repo plus one clean WSL proof checkout that is refreshed from git, not by copying files
- if you switch host environment, recreate repo-local generated surfaces such as `.venv`, `interface/node_modules`, and `interface/.next`

Local engine rule:
- Windows and macOS should treat Ollama or a remote OpenAI-compatible provider as the normal local-engine story
- repo-local `cognitive.*` helpers are intended for supported Linux GPU hosts, not as the default path on Windows or macOS

## Development Workflow

Agents implementing V8 should follow this process:
1. clean runtime environment and running services when the slice requires a deterministic baseline
   - `uv run inv lifecycle.down` now treats repo-local Interface worker residue as part of the shutdown contract, not just bound ports
   - when the validation target is the supported home single-host stack, use `uv run inv compose.down --volumes` instead of Kind teardown so the compose runtime starts from a truly clean database/bus state
   - when repeated build/test cycles have been running for a while, clear stale runtime residue before assuming the issue is just disk: leaked Interface workers and long-lived local services can keep caches hot and hold build outputs open
2. review the layered architecture truth in README, the owning architecture doc, and `.state/V8_DEV_STATE.md`
3. review V7 architecture-library documentation as migration input when a V8 replacement has not fully landed yet
4. identify migration targets and required contract updates
5. implement incremental runtime or documentation updates in the Windows dev repo
6. commit and push the Windows-side slice before authoritative proof, then refresh the WSL proof checkout from git
   - expected handoff shape: `git push` from Windows -> `git fetch --prune`, `git checkout`, `git reset --hard`, and `git clean -fdx` in the WSL proof checkout while preserving generated `workspace/tool-cache`, `workspace/logs`, and `workspace/docker-compose` roots
   - keep that destructive reset/clean behavior scoped to the dedicated WSL proof checkout, not the active Windows dev repo
7. run authoritative build, API, UI, and runtime proof from the WSL `mother-brain` checkout
   - use the WSL proof checkout for install, backend tests, interface tests/build, Compose bring-up, browser automation, and release-style gates
   - when the runtime is hosted in WSL on the same Windows machine, use the Windows browser at `http://localhost:3000` as the required operator-facing access path
8. verify with tests and execution gates
- if the machine is low on free space, prefer the repo task path in this order: `uv run inv lifecycle.down`, `uv run inv cache.status`, then `uv run inv cache.clean`
- heavy repo-managed build/test paths now run a disk-headroom preflight automatically; if it fails, reclaim space before retrying instead of pushing through partial builds
   - if you are setting up a new development machine, treat cache placement as part of the build config, not an afterthought: Windows should stamp user-level cache vars early, and Linux/macOS should point project/user cache roots at the volume you actually want repeated builds and browser runs to consume
9. update `.state/V8_DEV_STATE.md` with current status and evidence

## Licensing & Editions

The canonical product-layer and licensing posture now lives in [docs/licensing.md](docs/licensing.md).

Use that document when you need the precise edition story for:
- self-hosted release
- self-hosted enterprise
- hosted admin control plane
- modular user-management and identity packaging
- shared Soma governance boundaries that must stay consistent across paid variants

Rule:
- edition or licensing documentation should describe packaging, control-plane layering, and governance boundaries
- it should not silently redefine the technical behavior owned by governance, identity, or runtime contracts

## Binary Releases

There is one canonical Core binary-release path now:

- repo-local binary only:
  - `uv run inv core.compile`
- Docker image build:
  - `uv run inv core.build`
- versioned binary archive:
  - `uv run inv core.package`
  - emits the archive plus a sidecar manifest and checksum in `dist/`
  - example cross-target packaging:
    - `uv run inv core.package --target-os=windows --target-arch=amd64 --version-tag=v0.1.0`

Release-packaging verification:
- enterprise/self-hosted Helm verification package:
  - `uv run inv k8s.deploy --verify-package --values-file=charts/mycelis-core/values-enterprise.yaml --release-label=enterprise`
  - writes rendered/package artifacts, manifest, and checksums under `dist/helm/`

Release automation:
- `.github/workflows/release.yaml`
- manual packaging workflow for enterprise/self-hosted Helm verification artifacts under `dist/helm/`
- `.github/workflows/release-binaries.yaml`
- runs on `v*` tag pushes
- also supports manual `workflow_dispatch`
- publishes versioned binary archives plus their manifest/checksum sidecars from `dist/` as GitHub release assets

Initial release handoff rule:
- before tagging or handing off a second-machine checkout, run `uv run inv ci.release-preflight --lane=release`
- keep current delivery blockers explicit in `.state/V8_DEV_STATE.md`; the latest state board is the authority for whether an issue is blocking initial user testing or release lock
- use [Testing](docs/TESTING.md) and [Remote User Testing](docs/REMOTE_USER_TESTING.md) as the operator-facing proof sequence for the handoff machine

## Documentation Responsibilities

Every implementation slice must update:
- `README.md`
- `.state/V8_DEV_STATE.md`
- architecture-library authority documents when target, execution, UI, or delivery rules change
- user-facing docs when operator meaning, workflow wording, memory/governance behavior, or product posture changes
- `docs/API_REFERENCE.md` when API behavior, payload meaning, or endpoint contract changes
- `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md` when tasking, validation, or operational behavior changes
- documentation manifest when a canonical doc should be visible in the in-app docs page
- docs tests when the contract they enforce changes

Synchronization rule:
- README is the primary architecture inception doc for active development
- V8.2 B2+ is the active delivery target and full production architecture frame
- V8.1 is the foundation and compatibility baseline for the Soma-primary operator surface
- `.state/V8_DEV_STATE.md` is the actual implementation scoreboard
- every implementation slice must include a docs review for the touched surface, even when the result is "reviewed, no content change required"
- slices that change architecture, release posture, operator wording, API meaning, or documentation authority must keep README, the owning docs, `docsManifest.ts`, and `tests/test_docs_links.py` synchronized in the same change
- slice close-out should explicitly report tests run, docs updated, and docs reviewed unchanged for the touched scope

The architecture-library remains the authoritative detailed documentation surface until the V8 library replaces the remaining V7 migration inputs.

## Status

Mycelis is now operating in a V8.2/B2+ delivery frame while preserving the V8.1 Soma-primary operator baseline.

The V7 system still provides important migration input and substrate memory.
The V8.1 architecture defines the compatibility baseline and default operator posture.
The V8.2 PRD defines the active full architecture and B2+ delivery frame.
`.state/V8_DEV_STATE.md` records what is actually complete, active, next, or blocked right now.

# Mycelis
Mycelis is a Soma-centered threaded workspace for shaping, executing, operating, and revisiting trusted AI outcomes through governed execution. This README is the development-swarm inception document. It points to current authority, defines the command and documentation contracts, and avoids duplicating the deeper architecture specs.
Canonical ownership:
- `README.md`: inception, navigation, and repo-wide working rules.
- `docs/architecture-library/MYCELIS_CANONICAL_PRD.md`: single product, architecture, UX, runtime, MVP, and release-gate authority.
- `.state/V8_DEV_STATE.md`: live implementation scoreboard; read the active snapshot and immediate next actions before historical notes.

## README TOC
- [Fresh Agent Start Here](#fresh-agent-start-here)
- [User Guidance](#user-guidance)
- [Agent Guidance](#agent-guidance)
- [What Mycelis Is](#what-mycelis-is)
- [Active Delivery Target (V8.3 Embodiment)](#active-delivery-target-v83-embodiment)
- [Compatibility Baseline](#compatibility-baseline)
- [Current Implementation State](#current-implementation-state)
- [Default And Advanced Surfaces](#default-and-advanced-surfaces)
- [Architecture Terms To Operator Terms](#architecture-terms-to-operator-terms)
- [Feature Status Standard](#feature-status-standard)
- [Review Targets](#review-targets)
- [Command Contract](#command-contract)
- [Development Contract](#development-contract)
- [Playwright Contract](#playwright-contract)
- [Testing Gate](#testing-gate)
- [Fastest Start](#fastest-start)
- [Cross-Platform Setup](#cross-platform-setup)
- [Development Workflow](#development-workflow)
- [Licensing And Releases](#licensing-and-releases)
- [Documentation Responsibilities](#documentation-responsibilities)
- [Status](#status)
## Fresh Agent Start Here
Review these before planning or editing:
1. [AGENTS.md](AGENTS.md)
2. [Mycelis Canonical PRD](docs/architecture-library/MYCELIS_CANONICAL_PRD.md)
3. [V8 Development State](.state/V8_DEV_STATE.md)
4. [Operations](docs/architecture/OPERATIONS.md)
5. [Testing](docs/TESTING.md)
6. [User Acceptance Testing](docs/REMOTE_USER_TESTING.md)
7. [Docs Manifest](interface/lib/docsManifest.ts)

Fresh-agent rules:
- The canonical PRD owns current release-candidate embodiment. Do not create a product-architecture archive or parallel doctrine inside the active documentation tree.
- `.state/V8_DEV_STATE.md` is the implementation truth for what is actually complete; use its active snapshot first and treat dated boards as evidence unless reactivated.
- Keep user-facing docs and engineering docs cross-linked but distinct.

## User Guidance
For product use rather than implementation, start with [Docs Navigation](docs/README.md), then the user docs under `docs/user/`. The in-app `/docs` surface should expose the same canonical operator-facing documents through `interface/lib/docsManifest.ts`. When a docs-only slice requires a manifest change outside ownership, report it rather than editing interface files.

## Agent Guidance
For implementation, review in this order:

1. [AGENTS.md](AGENTS.md)
2. [Mycelis Canonical PRD](docs/architecture-library/MYCELIS_CANONICAL_PRD.md)
3. [V8 Development State](.state/V8_DEV_STATE.md)
4. [Testing](docs/TESTING.md)
5. [Operations](docs/architecture/OPERATIONS.md)

Before changing runtime, API, operator workflow, governance, testing, or task behavior, identify the owning doc and update or explicitly review it in the same slice.

## What Mycelis Is
In operator language, Mycelis lets someone ask Soma for an outcome, see what happened, open the durable result, recover when trust is broken, and return later knowing what is active, delivered, incomplete, or needs attention.

In architecture language, Mycelis is built around user-owned Workspaces and Outcomes, a Soma operational identity layer, governed execution, memory/continuity contracts, durable deliverables, recoverable runs, and auditable automation.

## Active Delivery Target (V8.3 Embodiment)

The active delivery target is [Mycelis Canonical PRD](docs/architecture-library/MYCELIS_CANONICAL_PRD.md): make the architecture operationally trustworthy through natural Soma conversation, bounded clarification, compact governance, async execution, durable outputs, proof, recovery, capability settings, and fresh-user GUI validation.

The active UI expression target is the human-first threaded Soma workspace defined in the canonical PRD: users talk with Soma before launching work, approve through a compact conversational pause, keep steering while asynchronous work runs, receive a concise completion summary with one primary output action, and open substantial deliverables in a dedicated surface. Work lists, capability setup, proof, recovery, and raw infrastructure remain contextual or behind Inspect.

Delivery rule:
- advance V8.3 slices only with a named boundary, proof lane, promotion rule, and documentation review
- use the staged delivery path `feature/* -> dev -> main`: prove the feature first, prove the integrated `dev` state after merge, and promote only a clean committed release candidate to `main`
- prefer operational embodiment over new doctrine: the canonical MVP workflow is natural Soma conversation -> shaped outcome -> approval when needed -> owned work -> deliverables -> proof/recovery -> revisit
- keep complex requests conversational: Soma gathers only the minimum sufficient brief, asks at most the few material questions needed for safe delivery, and records safe defaults when the operator asks it to proceed
- extend the existing spine through reusable Outcome Templates and declarative configuration; templates may shape WorkIntent but must not become another permanent user-facing object or parallel execution model
- do not create new split doctrine documents for current release-candidate scope

## Compatibility Baseline

The compatibility baseline is now inside the canonical PRD. Older versioned architecture docs were deleted from the active tree so current work does not split across historical doctrine. Actual implementation state lives in [.state/V8_DEV_STATE.md](.state/V8_DEV_STATE.md).

## Current Implementation State

Use `.state/V8_DEV_STATE.md` for the active scoreboard. Its active snapshot and immediate next actions are the current execution truth; older dated boards remain historical evidence only through Git history unless explicitly copied into the active snapshot.

Current configuration delivery boundary:
- `NEXT` P0.3a introduces the shared `ConfigDocument` validation, preview, activation, version, and rollback path plus the first Outcome Template compiler. Soma-authored configuration and direct YAML/JSON must pass through the same governed pipeline.
- `REQUIRED` P0.7b extends that proven path to Worker Profiles, team templates, reusable asks/actions, capability and MCP connections, search/data sources, registered inputs, governed NATS actions, and scheduled/service definitions.
- These are delivery targets, not claims of completed behavior. The current implementation and accepted proof remain authoritative in [.state/V8_DEV_STATE.md](.state/V8_DEV_STATE.md).

Status changes in planning/state docs must use the canonical markers: `REQUIRED`, `NEXT`, `ACTIVE`, `IN_REVIEW`, `COMPLETE`, `BLOCKED`.

## Default And Advanced Surfaces

Default surfaces should read as product workflows, not raw system internals:
- The root URL is no longer a public marketing page; every edition enters through `/login` and then the authenticated Soma workspace.
- Soma is the primary counterpart.
- The authenticated Dashboard is the primary Soma Workspace; compatibility organization routes must not introduce a competing product hierarchy.
- Intent suggestions live inside Soma, not as competing panels or separate front doors; they should frame outcome, output shape, proof, and next action rather than raw prompts.
- Meaningful actions must show a causal summary: understood intent, coordination, outputs, state changes, and next step.
- Simple requests stay simple. Soma independently infers answer depth and execution intent, while complex work uses a bounded minimum-sufficient brief covering the Outcome, audience, essential behavior, quality bar, delivery form, constraints, and acceptance evidence.
- Workspaces are governed user contexts; Outcomes hold deliverables, active work, proof, recovery, history, and continuity.
- Teams and groups are visible when they help the operator review or steer work.
- Advanced controls expose runtime depth, MCP/resources, deep memory, groups, runs, settings, auth, and docs without polluting first-run or default use; long topology surfaces should use focused menu/detail or list/detail panes rather than primary-page sprawl.

Use the [Mycelis Canonical PRD](docs/architecture-library/MYCELIS_CANONICAL_PRD.md) for screen/API expectations and browser-proof standards.

Default Operator Surface:
- the default UX must stay simple and intent-first while making Workspace and Outcome ownership obvious
- Soma, Outcomes, Work, deliverables, recovery, and focused Resources access are the normal operator concepts; Groups, teams, runs, capabilities, and transport stay contextual or behind Inspect

Advanced Architecture / Runtime Surface:
- the advanced architecture/runtime surface is now defined as a contract, but it is not fully implemented yet
- Admin tools expose Activity/Runs, deep Memory, System, Settings, and explicit Inspect detail; they must stay separate, make inheritance legible, and make config origin legible

source-of-truth layers remain separate:
- guided UI settings
- file-authoritative declarative configuration authored through Soma or direct YAML/JSON
- deployment/env overrides
- runtime state
- state and architecture docs

Planned configuration authoring follows one path: `parse -> validate -> resolve scope and references -> dry-run -> preview effects -> approve when required -> atomic write -> compile/register -> activate -> proof/rollback`. User, organization, and Workspace scope must remain explicit; committed configuration stores secret references rather than raw secrets. Configuration adapters must compile into the existing WorkIntent, team, capability, source, NATS, scheduling, and service contracts instead of bypassing their governance.

managed exchange foundation: channels, threads, schemas, and normalized outputs remain the governed substrate. managed exchange is permissioned; normalization into managed exchange does not imply unrestricted trust. The free-node release now includes foundational security boundaries.
capability manifest foundation: MCP tools, custom connectors, local scripts, external APIs, generated artifacts, and future plugins/modules must register as governed capabilities before Soma, teams, groups, or automations use them. Meaningful executions attach to runs; meaningful outputs normalize into exchange, artifacts, audit, or learning candidates instead of remaining raw tool side effects.
directed execution foundation: default UX must show the outcome need, Soma understanding, owned work, deliverables, proof, recovery, and next step. Runs, teams, capabilities, and deployment trust are supporting surfaces, not default user vocabulary.
finalization GUI posture: live Soma governance, team execution, first-demo project-package proof, proof opening/reload, Groups output, and degraded retry are green or in review. Cold-start Soma must not imply prior work, and runtime teams must not be presented as production delivery collaborators until bounded role-specific asks return within timeout with visible output/proof refs or actionable degradation.

## Architecture Terms To Operator Terms

| Architecture term | Operator term |
| --- | --- |
| Workspace and Outcome contracts | Workspace and Outcomes |
| Soma Kernel | Soma |
| WorkIntent and execution contract | Proposed work and approval details |
| Provider policy | AI Engines |
| Identity and continuity state | Memory and continuity |
| Mission and run events | Activity and run receipt |
| Capability and tool policy | Resources and capabilities |

## Feature Status Standard

Use only these markers in planning and state docs:
- `REQUIRED`: must exist for delivery or gate pass, but not started or ready
- `NEXT`: highest-priority upcoming slice
- `ACTIVE`: currently being worked
- `IN_REVIEW`: implemented and awaiting validation/review/gate decision
- `COMPLETE`: accepted and delivered
- `BLOCKED`: cannot advance until a named dependency or defect is resolved

## Review Targets

At minimum review these when the touched surface changes:
- `README.md`
- `.state/V8_DEV_STATE.md`
- the owning canonical/user/ops docs
- `docs/API_REFERENCE.md` for API behavior or payload meaning
- `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md` for testing/task-running behavior
- `interface/lib/docsManifest.ts` when an in-app docs entry is added, removed, or moved

## Command Contract

Use `uv run inv ...` for real task execution. Managed Interface dependency bootstrap uses `npm ci` so proof checkouts remain lockfile-clean.

```bash
uvx --from invoke inv -l
```

Do not use bare `uvx inv ...`.

The root install path also provisions Reticulum as part of the default app substrate. `uv run inv install` installs the `rns` Python package into the managed project environment and verifies the Reticulum utility path through `uvx --from rns rnstatus --help`, so Mycelis can immediately import Reticulum and operators can reach the standard Reticulum CLI tools from the same setup lane.

Common commands:

```bash
uv run inv install
uv run inv lifecycle.up --frontend
uv run inv lifecycle.status
uv run inv lifecycle.health
uv run inv lifecycle.down
uv run inv lifecycle.down --include-data-plane
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
uv run inv ci.baseline
uv run inv api.delivery-proof
```

Compose launch and readiness use the same configurable host ports from `.env.compose`: `MYCELIS_COMPOSE_POSTGRES_PORT`, `MYCELIS_COMPOSE_NATS_PORT`, `MYCELIS_COMPOSE_CORE_PORT`, and `MYCELIS_COMPOSE_INTERFACE_PORT`. The repository default publishes PostgreSQL on `15432`; local Core uses `DB_HOST=127.0.0.1` and `DB_PORT=15432`, while Compose Core uses the container address `postgres:5432`.

Development persistence has one contract: Docker Compose runs `pgvector/pgvector:pg16` as the sole PostgreSQL server, and relational data plus vector data share its `postgres-data` volume. A host `psql` binary is a client for that containerized server only. Running a native host PostgreSQL server for Mycelis development is unsupported.

NATS is a reusable transport host, not a Mycelis-owned workflow object. Set `NATS_URL` to the broker this Core process should use and give each Mycelis deployment a stable `MYCELIS_NATS_SERVICE_ID`; Core names only its own runtime and observer clients and drains only those clients on shutdown. Other services may share the broker by publishing to concrete channels registered through `/api/v1/input-sources`; duplicate channel claims and wildcard ingress are rejected, and high-rate traffic is buffered before teams consume it.

`uv run inv install` includes Reticulum bootstrap: it syncs the locked `rns` package through `uv`, then warms/verifies `uvx --from rns rnstatus --help` before continuing with Go, Interface, and Playwright setup.

Cleanup note: `uv run inv clean.generated` removes repo-local generated artifacts but skips the active Python runtime directory when the task is running from that environment. If you intentionally need to remove `.venv`, do it from an external shell after leaving the environment.

Task boundary: repo Invoke tasks manage Mycelis tools, app services, data-plane dependencies, and proof lanes. Default development uses `compose.infra-up` for Dockerized PostgreSQL/NATS and runs Core/Interface locally; there is no supported native-host PostgreSQL fallback. Host runtimes such as WSL distros, Rancher Desktop itself, Docker Desktop itself, and OS-level VM resets are operator/platform responsibilities outside the task registry; use repo tasks to probe, validate, and run Mycelis on those tools, not to manage the whole host environment.

Service-control boundary: `lifecycle.down` stops local Core, Interface, repo-owned helpers, and Kubernetes port-forwards while retaining reusable data-plane services. In Compose development, use `lifecycle.down --include-data-plane` to stop Core/Interface plus PostgreSQL/NATS in one command without deleting volumes; `compose.down` remains the lower-level container-only equivalent. Ollama, Rancher Desktop, Docker Desktop, WSL distributions, and external/shared NATS hosts are not silently stopped.

`lifecycle.status` is the quick local snapshot and now confirms Core through `/healthz` plus Ollama through `/api/tags` across loopback fallbacks; use `lifecycle.health` for deeper endpoint proof, `uv run inv api.delivery-proof` for API self-use, and `uv run inv ci.entrypoint-check` for runner matrix proof. The deeper health gate gives `/api/v1/cognitive/status` enough time to return bounded provider evidence instead of timing out at the client edge.

## Development Contract

A slice is not complete unless:
- tests pass
- documentation is updated where meaning changed
- architecture alignment is verified across the layered truth surfaces

`README.md` is the repo navigation document. `docs/architecture-library/MYCELIS_CANONICAL_PRD.md` is the canonical product and architecture authority. `.state/V8_DEV_STATE.md` is the source of actual implementation truth. all slices must update these surfaces when implementation, release posture, or target meaning changes.

end-of-slice reporting must explicitly state which tests ran, which docs changed, and which scoped docs were reviewed but left unchanged.

every implementation slice must include a docs review for the touched surface, even when the result is "reviewed, no content change required". Review `docs/API_REFERENCE.md` when API behavior, payload meaning, or endpoint contract changes.

Go owns runtime, orchestration, API, NATS, and backend persistence-facing logic. TypeScript owns the interface and in-app docs browser. Python owns repo management, operator automation, CI orchestration, and local test harnesses. SQL owns migrations, and `db.migrate` compatibility gates include capability, proof, trust, team-work lifecycle and recovery deadlines, collaboration-group workspace, OutcomeProject, TeamRegistry, search-source, Worker Profile, dispatch-outbox, operator SSE event-ledger, and team signal-receipt schema. Known late contracts use targeted migration replay instead of silently accepting a broadly compatible but incomplete database. `uv run inv db.clear-runtime-context` is the guarded source-mode reset for stale Soma/team runtime context before fresh UX proof. PowerShell may only be a thin host wrapper when the platform requires it.

Keep secrets in `.env` or deployment secret backends. Use `.env.compose` for Compose topology and non-secret runtime shape. Runtime config and UI surfaces should carry env-var or `SecretRef` references, not raw secret values; see the settings and capability configuration section of the [Mycelis Canonical PRD](docs/architecture-library/MYCELIS_CANONICAL_PRD.md).

Env override contract: `MYCELIS_PROVIDER_<PROVIDER_ID>_*`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, `MYCELIS_MEDIA_*`, and `MYCELIS_MEDIA_GATEWAY_*` are supported deployment-time knobs. Core can call a Pinokio-hosted Forge/AUTOMATIC1111 API directly with `MYCELIS_MEDIA_TYPE=forge`; Forge must be launched with API mode enabled. The optional local media gateway remains the adapter for ComfyUI or OpenAI-compatible media endpoints and blocks public upstreams by default unless `MYCELIS_MEDIA_GATEWAY_ALLOW_PUBLIC_UPSTREAM=1` is intentionally set. The retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` must not return. `Bundle -> Instantiated Organization -> Inheritance -> Routing` remains the runtime truth. Env overrides are deployment-time infrastructure wiring, not runtime organization behavior and do not replace bundle-defined runtime organization truth.

Deployment guidance by host architecture: Windows x86_64, Linux x86_64, Linux arm64, and Mixed-architecture deployments are supported through the lane-specific guidance in local dev and operations docs. The deployed Core image resolves runtime config from `/core/config`.

Supported user access lanes: source-mode local development with Dockerized PostgreSQL/NATS first, then Windows Rancher Desktop or Docker Desktop Compose, Windows + WSL Docker Compose, Rancher Desktop K3s, guarded WSL Compose, and Kubernetes / Helm clustered deployment when container proof is intentionally requested. Run/build/test Core and Interface locally before containerizing app services; open `http://localhost:3000` from the Windows browser for same-machine proof, and for clustered proof, prove the real ingress/hostname/IP from the operator machine. Rancher Desktop K3s is the preferred Windows local commercial-release parity lane once local source proof is acceptable.

Deployment target contract: Kubernetes / Helm targets self-hosted and enterprise deployment using standard Kubernetes resources; Docker Compose remains rapid local development, demo, and same-machine proof runtime, not the clustered deployment contract. Run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml` and cover Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy. Local Windows K3s proof uses `MYCELIS_K8S_BACKEND=rancher` against Rancher Desktop.

AI endpoint contract: use a reachable host/IP like `http://192.168.x.x:11434/v1`, not `localhost`; for Compose point it at a host-reachable endpoint such as `http://host.docker.internal:11434`; WSL proof may auto-start a WSL-host relay for the AI endpoint when needed. K8s deployments can set `MYCELIS_K8S_TEXT_ENDPOINT` plus `MYCELIS_K8S_TEXT_MODEL_ID`; the Helm chart projects provider endpoint/model env vars and opens explicit AI egress ports only when configured.

Kubernetes values contract: prefer Rancher Desktop K3s on Windows and `k3d` on WSL/Linux as the local Kubernetes backends; prefer `k3d` as the local Kubernetes backend when it is available on WSL/Linux; set `MYCELIS_K8S_BACKEND=kind` only as fallback. `MYCELIS_K8S_VALUES_FILE` may select `charts/mycelis-core/values-k3d.yaml`, `charts/mycelis-core/values-enterprise.yaml`, or `charts/mycelis-core/values-enterprise-windows-ai.yaml`.

Runtime packaging contract: the supported Core container image includes Node/npm/npx for curated stdio MCP servers, and manual `filesystem` library install must be able to launch and bind to the configured `/data/workspace` output block. `MYCELIS_WORKSPACE` is the governed workspace for generated files, project packages, browser games, filesystem MCP writes, and group-owned output folders under `groups/...`; generated project packages should retain support files such as `README.md`, `PROOF.md`, and `project-package.json`. `MYCELIS_ARTIFACT_ROOT` is the file-backed artifact/cache root; `DATA_DIR` is a legacy alias that should stay aligned until removed. Compose maps these to `/data/workspace` and `/data/artifacts`, and System -> Deployments reports both roots for operator proof.

Release proof contract: start with local source gates (`core.test`, `interface.test`, `interface.typecheck`, `interface.build`, focused Playwright) against Dockerized PostgreSQL/NATS when live data or bus proof is required. Only after those pass, containerize Core/Interface for full Compose/K8s proof. `uv run inv ci.release-preflight --lane=release` is the full local release gate and always runs repository lint before baseline, service, or deployment-facing proof; its managed baseline uses the repo-provisioned Chromium project with one worker and retained JSON/JUnit evidence. A lint failure stops the lane before later work can create false confidence. CI-owned Go commands select the Core module explicitly with `go -C` instead of trusting mutable shell-directory state, so Windows and Linux prove the same package tree, and failed hidden commands emit console-safe diagnostics on Windows as well as Linux. Use guarded WSL tasks when WSL Compose deployment-mimic proof matters, and Rancher Desktop K3s with `MYCELIS_K8S_BACKEND=rancher` when the release slice needs local Kubernetes parity proof. Hosted release jobs remain manual; `Full Release Candidate` chains source gates, authenticated browser proof, optional hosted source API proof, Helm packaging, optional images, and binary packaging. Hosted GitHub lanes use Node 24-capable action majors, Node.js 24 for Interface proof and container builds, and checksum-verified pinned Helm 3 setup instead of Node-20 Helm actions.

## Playwright Contract

Invoke-managed Playwright owns the local Next.js server lifecycle. Run `uv run inv interface.e2e ...` sequentially for a workspace and port, and keep managed `interface.build`, `interface.test`, and managed browser proof out of parallel runs because they own the same Next/Vitest worker surface. The Playwright config owns reporters for both readable console output and retained `interface/test-results/playwright-results.json` plus `.xml` evidence; Invoke must not override that reporter set. Merge-readiness proof uses the built production Interface server path, covers chromium firefox webkit where relevant, and `uv run inv ci.baseline` includes Playwright by default. Use `uv run inv interface.e2e --server-mode=start --project=chromium --workers=1` for broad local UI certification without live provider effects; use `uv run inv ci.service-check`, headed Chromium, live-backend proof, and the guarded `wsl.validate --lane=release --headed-browser` path when the delivered operator path, proxy, runtime, retained outputs, or governance behavior changes. Native Core live proof infers the host-visible backend workspace from the loaded `.env`/process `MYCELIS_WORKSPACE`: absolute roots are used directly, while repo-local `./workspace` maps to `core/workspace`; K8s/PVC-backed browser proof that asserts backend-written files should set `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s` and the matching namespace/selector/workspace root so the proof checks the pod workspace instead of host-only Compose paths. `uv run inv interface.check` retries transient Windows socket-reuse failures after heavy browser runs before reporting a route failure.

## Testing Gate

Canonical testing guidance lives in [Testing](docs/TESTING.md). Browser proof depth and product acceptance live in the [Mycelis Canonical PRD](docs/architecture-library/MYCELIS_CANONICAL_PRD.md). `.state/V8_DEV_STATE.md` remains the detailed delivery scoreboard and current proof index.

End-of-slice reporting should name evidence commands run, docs changed, touched docs reviewed unchanged, and any UI visual-expression review for surfaces the slice touched.

## Fastest Start

```bash
uv run inv install
uv run inv auth.dev-key
uv run inv db.migrate
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```

For the supported home-runtime stack:
```bash
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
```

Bootstrap reminder: normal startup fails closed unless a valid bootstrap bundle is present, and `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must choose a bundle when more than one is mounted.

## Cross-Platform Setup

Windows is the source-edit and git surface. WSL is the guarded Compose proof checkout for install, build, tests, Compose, and live GUI validation. Rancher Desktop K3s is the Windows local Kubernetes proof lane for Helm/commercial-release parity.

Compose projects the same `MYCELIS_WEB_SESSION_SECRET` and `MYCELIS_WEB_IDENTITY_FORWARD_SECRET` references into Core and Interface. When either value is omitted, both containers use the repo-local `MYCELIS_API_KEY` fallback; deployment-specific secret values belong in `.env` and must remain identical across both services.

Use the guarded WSL handoff lane when release-style proof matters:

```bash
uv run inv wsl.status
uv run inv wsl.refresh
uv run inv wsl.validate --lane=release
```

Refresh WSL from git; do not copy source trees across the host boundary. WSL tasks are proof-checkout tasks (`status`, `refresh`, `validate`, `cycle`), not host lifecycle controls; shut down or repair the WSL/Rancher/Desktop runtime with platform tools when the host itself is unhealthy. Use `uv run inv wsl.validate --lane=release --headed-browser` for visible live-window Playwright proof on the same WSL-hosted Compose UI.

Guarded WSL proof commands are `uv run inv wsl.status`, `uv run inv wsl.refresh`, `uv run inv wsl.validate`, and `uv run inv wsl.cycle`.

## Development Workflow

Use [Local Development Workflow](docs/LOCAL_DEV_WORKFLOW.md) for first-time setup, deployment selection, ports, health checks, cognitive engine setup, and troubleshooting.

Use [Operations](docs/architecture/OPERATIONS.md) for task ownership, lifecycle, Compose, Kubernetes, CI, and release-lane sequencing.

## Licensing And Releases

Licensing guidance lives in [Licensing](docs/licensing.md). Binary release and packaging commands live in [Local Development Workflow](docs/LOCAL_DEV_WORKFLOW.md#binary-release-process) and [Operations](docs/architecture/OPERATIONS.md).

Release licensing separates the local self-hosted node from the hosted admin control plane and from full enterprise multi-user IAM, federated SAML/OIDC/SSO, optional lifecycle sync, and delegated enterprise admin/recovery flows. The current Interface always requires a signed web session; free/self-hosted nodes can use local owner login, while enterprise deployments can enable Google Workspace OIDC with internal Mycelis admin/standard roles.

## Documentation Responsibilities

Every implementation slice that changes product behavior, runtime behavior, operator workflow, API contract, governance posture, canonical terminology, task execution, or validation must include documentation review in the same slice. When adding, removing, or renaming major README sections, update this README TOC in the same change.

Keep documentation split into two layers:
- **User help layer**: `docs/user/*` and the in-app `/docs` manifest should explain how a person uses Soma, groups, resources, outputs, proof, settings, and recovery. Lead with the task, the expected result, and the next action. Keep implementation contracts, raw topology, and historical doctrine out of the default path.
- **Architecture and repo layer**: `README.md`, `docs/README.md`, `.state/V8_DEV_STATE.md`, and `docs/architecture-library/MYCELIS_CANONICAL_PRD.md` should define current delivery truth, proof gates, and engineering contracts. Update this layer only when the product meaning, workflow contract, or release target changes.

After each subjective UI step, pair the code/test change with the matching docs cleanup: update the affected user guide if the operator experience changed, update the README/architecture layer if the product contract changed, update `interface/lib/docsManifest.ts` if the in-app help entry moved, and record the active state/proof in `.state/V8_DEV_STATE.md`.

## Status

The repo is in active V8.3 operational embodiment with one canonical PRD. Treat this README as a compact navigation contract; do not re-expand it into a duplicate architecture monolith.

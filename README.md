# Mycelis

Mycelis is a governed cognitive operating environment for creating, operating, and evolving AI Organizations through a Soma-primary operator experience.

This README is the development-swarm inception document. It points to current authority, defines the command and documentation contracts, and avoids duplicating the deeper architecture specs.

Canonical ownership:
- `README.md`: inception, navigation, and repo-wide working rules
- `architecture/v8-2.md`, `docs/architecture-library/V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md`, and `docs/architecture-library/V8_2_FINALIZATION_CONCRETIZATION_CONTRACT.md`: active V8.2/B2+ production target plus embodiment rule, exact first demo slice, runtime object schemas, degradation posture, and product-category framing
- `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md` and `docs/architecture-library/V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md`: operator UX, screen/API truth, and directed-execution alignment
- `docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md`: capability manifest, run/output, MCP/custom integration, and proof truth
- `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`: V7-to-V8 bootstrap migration truth
- `.state/V8_DEV_STATE.md`: live implementation scoreboard; read the active snapshot and immediate next actions before dated historical boards
## README TOC

- [Fresh Agent Start Here](#fresh-agent-start-here)
- [User Guidance](#user-guidance)
- [Agent Guidance](#agent-guidance)
- [What Mycelis Is](#what-mycelis-is)
- [Active Delivery Target (V8.2 B2+)](#active-delivery-target-v82-b2)
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
2. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
3. [V8 Development State](.state/V8_DEV_STATE.md)
4. [V8.2 Production Architecture Target](architecture/v8-2.md)
5. [V8.2 Operational Embodiment Directive](docs/architecture-library/V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md)
6. [V8 Runtime Contracts](docs/architecture-library/V8_RUNTIME_CONTRACTS.md)
7. [V8 Config and Bootstrap Model](docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md)
8. [V8 UI/API and Operator Experience Contract](docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
9. [Operations](docs/architecture/OPERATIONS.md)
10. [Testing](docs/TESTING.md)
11. [Remote User Testing](docs/REMOTE_USER_TESTING.md)
12. [V8.2 Current State And Finalization PRD](docs/architecture-library/V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md)
13. [V8.2 Finalization Delivery Plan](docs/architecture-library/V8_2_FINALIZATION_DELIVERY_PLAN.md), [V8.2 Finalization Concretization Contract](docs/architecture-library/V8_2_FINALIZATION_CONCRETIZATION_CONTRACT.md), and [V8.2 Soma Team Interaction Contract](docs/architecture-library/V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md)
14. [V8 Capability Manifest And Runtime Integration Standard](docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md)
15. [V8 UI Testing Agentry Product Contract](docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md)
16. [Docs Manifest](interface/lib/docsManifest.ts)

Fresh-agent rules:
- V8.2 docs are the active authority. Historical V7 state remains migration evidence only, not active product authority.
- `.state/V8_DEV_STATE.md` is the implementation truth for what is actually complete; use its active snapshot first and treat dated boards as evidence unless reactivated.
- Keep user-facing docs and engineering docs cross-linked but distinct.

## User Guidance

For product use rather than implementation, start with [Docs Navigation](docs/README.md), then the user docs under `docs/user/`. The in-app `/docs` surface should expose the same canonical operator-facing documents through `interface/lib/docsManifest.ts`. When a docs-only slice requires a manifest change outside ownership, report it rather than editing interface files.

## Agent Guidance

For implementation, review in this order:

1. [AGENTS.md](AGENTS.md)
2. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
3. [V8 Development State](.state/V8_DEV_STATE.md)
4. [Testing](docs/TESTING.md)
5. [Operations](docs/architecture/OPERATIONS.md)

Before changing runtime, API, operator workflow, governance, testing, or task behavior, identify the owning doc and update or explicitly review it in the same slice.

## What Mycelis Is

In operator language, Mycelis lets someone ask Soma to create or review meaningful work, see what happened, inspect the durable result, and recover when trust is broken.

In architecture language, Mycelis is built around instantiated organizations as runtime truth, a Soma operational identity layer, governed execution, memory/continuity contracts, durable outputs, recoverable runs, and auditable automation.

## Active Delivery Target (V8.2 B2+)

The active target is [V8.2 Production Architecture Target](architecture/v8-2.md), the full actuation architecture. V8.2 B2+ is the active delivery target.

Delivery rule:
- advance V8.2 modules only with a named boundary, proof lane, promotion rule, and documentation review
- prefer operational embodiment over new doctrine: the canonical MVP workflow is intent -> Soma -> governed approval when needed -> execution -> run -> durable output -> proof/recovery
- do not grow V8.1 docs with new B2+ target scope

## Compatibility Baseline

The Soma-primary compatibility baseline is now held in current V8.2 docs instead of an older versioned architecture file. The baseline remains: AI Organizations as governed work contexts, Soma as the primary operating surface, bounded teams/capabilities, response contracts, continuity, and inspectable governance. Actual implementation state lives in [.state/V8_DEV_STATE.md](.state/V8_DEV_STATE.md).

## Current Implementation State

Use `.state/V8_DEV_STATE.md` for the active scoreboard. Its active snapshot and immediate next actions are the current execution truth; older dated boards remain historical evidence unless explicitly referenced by the active snapshot. Use `.state/V7_DEV_STATE.md` only as historical migration evidence.

Status changes in planning/state docs must use the canonical markers: `REQUIRED`, `NEXT`, `ACTIVE`, `IN_REVIEW`, `COMPLETE`, `BLOCKED`.

## Default And Advanced Surfaces

Default surfaces should read as product workflows, not raw system internals:
- The root URL is no longer a public marketing page; every edition enters through `/login` and then the authenticated Soma workspace.
- Soma is the primary counterpart.
- Dashboard and organization workspaces share the same Soma operating surface.
- Intent suggestions live inside Soma, not as competing panels or separate front doors; they should frame outcome, output shape, proof, and next action rather than raw prompts.
- Meaningful actions must show a causal summary: understood intent, coordination, outputs, state changes, and next step.
- AI Organizations are governed work contexts.
- Teams and groups are visible when they help the operator review or steer work.
- Advanced controls expose runtime depth, MCP/resources, deep memory, groups, runs, settings, auth, and docs without polluting first-run or default use; long topology surfaces should use focused menu/detail or list/detail panes rather than primary-page sprawl.

Use [V8 UI/API and Operator Experience Contract](docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) for screen/API expectations and [V8 UI Team Full Test Set](docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md) for browser proof.

Default Operator Surface:
- the default UX must stay simple and intent-first
- Soma, AI Organizations, teams/groups, activity, memory, resources, and settings are the operator-visible product surfaces

Advanced Architecture / Runtime Surface:
- the advanced architecture/runtime surface is now defined as a contract, but it is not fully implemented yet
- the advanced architecture/runtime surface must stay separate, make inheritance legible, and make config origin legible

source-of-truth layers remain separate:
- guided UI settings
- bundle/file configuration
- deployment/env overrides
- runtime state
- state and architecture docs

managed exchange foundation: channels, threads, schemas, and normalized outputs remain the governed substrate. managed exchange is permissioned; normalization into managed exchange does not imply unrestricted trust. The free-node release now includes foundational security boundaries.
capability manifest foundation: MCP tools, custom connectors, local scripts, external APIs, generated artifacts, and future plugins/modules must register as governed capabilities before Soma, teams, groups, or automations use them. Meaningful executions attach to runs; meaningful outputs normalize into exchange, artifacts, audit, or learning candidates instead of remaining raw tool side effects.
directed execution foundation: default UX must show intent, Soma understanding, execution shape, capability/team use, outputs, proof, and next step. Runs, outputs, governance, capabilities, and deployment trust are product surfaces, not hidden diagnostics.
finalization GUI posture: live Soma governance, team execution, and playable-output proof are green, but the exact first demo slice still needs package-level proof for README/validation metadata, visible proof-link opening, and degraded retry. Cold-start Soma must not imply prior work, and active team work needs durable work-item state before runtime teams are presented as production collaborators.

## Architecture Terms To Operator Terms

| Architecture term | Operator term |
| --- | --- |
| Inception | AI Organization |
| Soma Kernel | Soma |
| Central Council | Advisors / governance support |
| Provider Policy | AI Engine Settings |
| Identity and Continuity State | Memory & Continuity |
| Mission / run events | Activity / Run timeline |
| Capability and tool policy | Resources / Connected Tools |

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

Common commands:

```bash
uv run inv install
uv run inv native-infra.install-nats
uv run inv native-infra.up
uv run inv lifecycle.up --frontend
uv run inv lifecycle.status
uv run inv lifecycle.health
uv run inv lifecycle.down
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
uv run inv ci.baseline
uv run inv api.delivery-proof
uv run inv lifecycle.memory-restart && uv run inv team.architecture-sync && uv run inv quality.max-lines --limit 300
```

Task boundary: repo Invoke tasks manage Mycelis tools, app services, data-plane dependencies, and proof lanes. `native-infra.*` owns the Windows/source-mode PostgreSQL + NATS dependency path. Host runtimes such as WSL distros, Rancher Desktop itself, Docker Desktop itself, and OS-level VM resets are operator/platform responsibilities outside the task registry; use repo tasks to probe, validate, and run Mycelis on those tools, not to manage the whole host environment.

`lifecycle.status` is the quick local snapshot and now confirms Core through `/healthz` plus Ollama through `/api/tags` across loopback fallbacks; use `lifecycle.health` for deeper endpoint proof, `uv run inv api.delivery-proof` for API self-use, and `uv run inv ci.entrypoint-check` for runner matrix proof.

## Development Contract

A slice is not complete unless:
- tests pass
- documentation is updated where meaning changed
- architecture alignment is verified across the layered truth surfaces

`README.md` is the primary architecture inception document. `architecture/v8-2.md` is the canonical full architecture. `.state/V8_DEV_STATE.md` is the source of actual implementation truth. all slices must update these surfaces when implementation, release posture, or target meaning changes.

end-of-slice reporting must explicitly state which tests ran, which docs changed, and which scoped docs were reviewed but left unchanged.

every implementation slice must include a docs review for the touched surface, even when the result is "reviewed, no content change required". Review `docs/API_REFERENCE.md` when API behavior, payload meaning, or endpoint contract changes.

Go owns runtime, orchestration, API, NATS, and backend persistence-facing logic. TypeScript owns the interface and in-app docs browser. Python owns repo management, operator automation, CI orchestration, and local test harnesses. SQL owns migrations, and `db.migrate` compatibility gates include the V8.2 capability, proof, and team-work tables. PowerShell may only be a thin host wrapper when the platform requires it.

Keep secrets in `.env`. Use `.env.compose` for Compose topology and non-secret runtime shape.

Env override contract: `MYCELIS_PROVIDER_<PROVIDER_ID>_*`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, and `MYCELIS_MEDIA_*` are supported deployment-time knobs. The retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` must not return. `Bundle -> Instantiated Organization -> Inheritance -> Routing` remains the runtime truth. env overrides are deployment-time infrastructure wiring, not runtime organization behavior and do not replace bundle-defined runtime organization truth.

Deployment guidance by host architecture: Windows x86_64, Linux x86_64, Linux arm64, and Mixed-architecture deployments are supported through the lane-specific guidance in local dev and operations docs. The deployed Core image resolves runtime config from `/core/config`.

Supported user access lanes: source-mode local development with native PostgreSQL/NATS first, then Windows/Rancher Desktop Compose, Windows Docker Desktop Compose, Windows + WSL Docker Compose, Rancher Desktop K3s, WSL Compose, and Kubernetes / Helm clustered deployment when container proof is intentionally requested. Run/build/test Core and Interface locally before containerizing app services; open `http://localhost:3000` from the Windows browser for same-machine proof, and for clustered proof, prove the real ingress/hostname/IP from the operator machine. Rancher Desktop K3s is the preferred Windows local commercial-release parity lane once local source proof is acceptable.

Deployment target contract: Kubernetes / Helm targets self-hosted and enterprise deployment using standard Kubernetes resources; Docker Compose remains rapid local development, demo, and same-machine proof runtime. Kubernetes / Helm: target self-hosted and enterprise deployment contract using standard Kubernetes resources. Docker Compose: rapid local development, demo, and same-machine proof runtime; it is not the target clustered deployment contract. Run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml` and cover Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy. Local Windows K3s proof uses `MYCELIS_K8S_BACKEND=rancher` against Rancher Desktop.

AI endpoint contract: use a reachable host/IP like `http://192.168.x.x:11434/v1`, not `localhost`; for Compose point it at a host-reachable endpoint such as `http://host.docker.internal:11434`; WSL proof may auto-start a WSL-host relay for the AI endpoint when needed. K8s deployments can set `MYCELIS_K8S_TEXT_ENDPOINT` plus `MYCELIS_K8S_TEXT_MODEL_ID`; the Helm chart projects provider endpoint/model env vars and opens explicit AI egress ports only when configured.

Kubernetes values contract: prefer Rancher Desktop K3s on Windows and `k3d` on WSL/Linux as the local Kubernetes backends; prefer `k3d` as the local Kubernetes backend when it is available on WSL/Linux; set `MYCELIS_K8S_BACKEND=kind` only as fallback. `MYCELIS_K8S_VALUES_FILE` may select `charts/mycelis-core/values-k3d.yaml`, `charts/mycelis-core/values-enterprise.yaml`, or `charts/mycelis-core/values-enterprise-windows-ai.yaml`.

Runtime packaging contract: the supported Core container image includes Node/npm/npx for curated stdio MCP servers, and manual `filesystem` library install must be able to launch and bind to the configured `/data/workspace` output block.

Release proof contract: start with local source gates (`core.test`, `interface.test`, `interface.typecheck`, `interface.build`, focused Playwright) against native PostgreSQL/NATS when live data or bus proof is required. Only after those pass, containerize Core/Interface for Compose/K8s proof. Use `uv run inv ci.release-preflight --lane=release` for the full local release gate, guarded WSL tasks when WSL Compose deployment-mimic proof matters, and Rancher Desktop K3s with `MYCELIS_K8S_BACKEND=rancher` when the release slice needs local Kubernetes parity proof. Hosted release jobs remain manual; `Full Release Candidate` chains source gates, authenticated browser proof, optional hosted source API proof, Helm packaging, optional images, and binary packaging.

## Playwright Contract

Invoke-managed Playwright owns the local Next.js server lifecycle. Run `uv run inv interface.e2e ...` sequentially for a workspace and port. Merge-readiness proof uses the built production Interface server path, covers chromium firefox webkit where relevant, and `uv run inv ci.baseline` now includes Playwright by default. Use `uv run inv ci.service-check`, headed Chromium, live-backend proof, and the guarded `wsl.validate --lane=release --headed-browser` path when the delivered operator path, proxy, runtime, retained outputs, or governance behavior changes. Native Core live proof infers the host-visible backend workspace as `core/workspace` from `MYCELIS_WORKSPACE=./workspace`; K8s/PVC-backed browser proof that asserts backend-written files should set `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s` and the matching namespace/selector/workspace root so the proof checks the pod workspace instead of host-only Compose paths. `uv run inv interface.check` retries transient Windows socket-reuse failures after heavy browser runs before reporting a route failure.

## Testing Gate

Canonical testing guidance lives in [Testing](docs/TESTING.md). Browser proof depth lives in [V8 UI Testing Agentry Product Contract](docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) and [V8 UI Team Full Test Set](docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md). Current RC proof and operator handoff evidence lives in [Release Handoff](docs/RELEASE_HANDOFF.md); `.state/V8_DEV_STATE.md` remains the detailed delivery scoreboard.

End-of-slice reporting should name evidence commands run, docs changed, and touched docs reviewed unchanged.

## Fastest Start

```bash
uv run inv install
uv run inv auth.dev-key
uv run inv native-infra.install-nats
uv run inv native-infra.up
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

## Status

The repo is in active V8.2/B2+ delivery with a protected V8.1 compatibility baseline. Treat this README as a compact navigation contract; do not re-expand it into a duplicate architecture monolith.

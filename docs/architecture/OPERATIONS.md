# Mycelis Operations Contract
> Navigation: [Project README](../../README.md) | [Canonical PRD](../architecture-library/MYCELIS_CANONICAL_PRD.md) | [Backend](BACKEND.md) | [Frontend](FRONTEND.md) | [Testing](../TESTING.md)
This manual owns task and runtime operations. It links to [Local Development Workflow](../LOCAL_DEV_WORKFLOW.md) for setup details and [Testing](../TESTING.md) for evidence gates.

Implementation slices that change runtime, tasking, validation, API meaning, or operator behavior must review and update the owning docs in the same change rather than leaving docs drift for later cleanup.

Current proof posture: workflows are manual-only, source-mode Core/Interface run against Dockerized PostgreSQL/NATS first, and containerized Core/Interface or Kubernetes app services are brought up only for explicit deployment proof.

## TOC

- [I. Prerequisites](#i-prerequisites)
- [II. Task Automation](#ii-task-automation)
- [III. Development Workflow](#iii-development-workflow)
- [IV. Configuration System](#iv-configuration-system)
- [V. Testing Strategy](#v-testing-strategy)
- [VI. CI/CD](#vi-cicd)
## I. Prerequisites

Operational lanes use:
- `uv` and Invoke for repo task execution
- Docker/Compose for the supported home-runtime stack
- Rancher Desktop K3s on Windows, or k3d/Kind plus kubectl/Helm for local Kubernetes
- Go, Node.js, and Python for source-mode development
- PostgreSQL, NATS, and a reachable AI endpoint for live runtime proof

Use `uv run inv ...` for real tasks. Use `uvx --from invoke inv -l` only as a compatibility probe. The root `install` task runs `uv sync --all-packages --dev`, verifies Reticulum import access with `uv run python`, and warms/verifies `uvx --from rns rnstatus --help` before Go, Interface, Playwright, and optional cognitive setup. Cleanup tasks skip the active Python runtime directory when invoked from that environment, then report the skip so Windows does not fail while deleting the running `.venv`.

## II. Task Automation

Task modules live under `ops/*.py` and are registered through `tasks.py`. App-tied management logic belongs in Python; `uv run inv api.delivery-proof` exercises the live Mycelis API as a source-mode delivery lane, while `uv run inv ci.entrypoint-check` proves runner registration.

The public task surface is capped at 95 registered commands. Each command must own distinct operator or proof behavior; aliases such as a second browser-test entrypoint or a second combined unit-test entrypoint are intentionally excluded. Use the task's owning namespace directly, and consolidate an existing task before increasing the budget.

Task ownership boundary: Invoke tasks manage repo tools, Mycelis app services, data-plane dependencies, and proof lanes; WSL/Rancher/Docker host lifecycle and VM repair stay outside repo tasking.

Compose infrastructure is the default source-mode data plane: `compose.infra-up` and `compose.infra-health` manage only Dockerized PostgreSQL and NATS, while `lifecycle.up --frontend` runs Core and Interface locally from source. Neither task builds application images. Use `MYCELIS_DEV_INFRA_MODE=compose` for development and `MYCELIS_DEV_INFRA_MODE=k8s` only for explicit clustered bridge proof; native host PostgreSQL is unsupported.

NATS may be shared by Mycelis and separately developed services. `NATS_URL` selects the single broker host used by a Core process, while `MYCELIS_NATS_SERVICE_ID` gives that deployment stable, distinguishable runtime and observer client names. Routine Core and `lifecycle.down` shutdown drains Mycelis client connections only; it does not stop, purge, or claim shared broker storage. External producers enter Mycelis only through concrete subjects registered in `/api/v1/input-sources`; wildcard and duplicate source-channel claims are rejected. Use a separately configured bridge/Core deployment for another NATS host instead of silently spanning hosts.

### Master Registry

List tasks:

```bash
uv run inv -l
```

### Core Tasks (`ops/core.py`)

```bash
uv run inv core.test
uv run inv core.compile
uv run inv core.run
uv run inv core.stop
uv run inv core.restart
uv run inv core.package
uv run inv core.smoke
```

### Interface Tasks (`ops/interface.py`, `ops/interface_runtime.py`)

```bash
uv run inv interface.install
uv run inv interface.test
uv run inv interface.typecheck
uv run inv interface.build
uv run inv interface.e2e
uv run inv interface.check
uv run inv interface.stop
```

### Database Tasks (`ops/db.py`)

```bash
uv run inv db.create
uv run inv db.migrate
uv run inv db.reset
uv run inv db.status
```

`db.migrate` owns fresh installation, not development-history replay. It applies the single `001_current_schema.sql` contract atomically only when the public schema is empty, makes a true no-op when the complete compatibility contract already passes, and fails before opening the installer when any nonempty schema is partial or incompatible. Back up retained data and choose an explicitly supported upgrade path or use `db.reset` for an intentional disposable rebuild; do not run removed historical migrations against an installed platform. A successful install reruns the complete compatibility contract before reporting success. Clean-deployment proof then starts the candidate and verifies first boot from code/config/secrets, not historical rows: Groups, Runs, Outcomes, deliverables, conversations, vectors, generated files, and user-created teams must remain empty until the first operator ask, while idempotent bootstrap support such as exchange registries, capability manifests, configured MCP defaults, nodes, and built-in runtime identities may be recreated and must not duplicate across restart. Use `db.clear-runtime-context` before fresh Soma/team GUI proof when stale conversations, team work, run/proof handshakes, or temp memory would influence the experience; it dry-runs by default and requires `--yes` to delete rows. Owner-scoped retained-fixture cleanup is separately exposed through the opt-in test API, rejects pre-existing team/runtime/workspace state, commits actual created-team and workspace ownership independently of the outer execution transaction, fences creation and purge with the same PostgreSQL advisory lock, resumes interrupted purges, releases active resource leases after successful purge, and never deletes shared NATS storage.

### Auth Tasks (`ops/auth.py`)

```bash
uv run inv auth.dev-key
uv run inv auth.break-glass-key
uv run inv auth.posture
```

### Cache Tasks (`ops/cache.py`)

```bash
uv run inv cache.status
uv run inv cache.guard
uv run inv cache.clean
```

`cache.guard` checks the repository, managed-cache, user-profile/system, and locally visible Docker-storage volumes because one can exhaust independently of the others. The default adaptive policy reserves 5% of the cache filesystem (bounded to 8-64 GiB), caps aggregate managed caches at 25% of currently available space above that reserve (maximum 64 GiB), and caps Playwright at 25% of that managed budget (maximum 12 GiB). Override those GiB decisions with `MYCELIS_CACHE_MIN_FREE_GB`, `MYCELIS_CACHE_MAX_GB`, and `MYCELIS_PLAYWRIGHT_CACHE_MAX_GB`. Heavy tasks fail closed with current usage and recovery guidance. `cache.clean` remains limited to repo-owned caches; it does not delete Docker volumes or unrelated user files.

For broader source-checkout hygiene, run `uv run inv clean.disk-status` before `uv run inv clean.generated`. The latter removes only its explicit dependency/build/report targets, `core/workspace/tool-cache`, `interface/workspace/tool-cache`, and discovered source-tree `__pycache__` directories. It preserves `.env`, `.env.compose`, runtime logs, whole runtime workspaces, Compose data, Docker volumes, and the Python environment executing the task.

### Lifecycle Tasks (`ops/lifecycle.py`)

```bash
uv run inv lifecycle.up --frontend
uv run inv lifecycle.status
uv run inv lifecycle.health
uv run inv lifecycle.restart --frontend
uv run inv lifecycle.down
uv run inv lifecycle.down --include-data-plane
uv run inv lifecycle.first-boot-proof
```

`lifecycle.status` is the fast local snapshot. It reports process/port state and confirms Core through `/healthz` plus Ollama through `/api/tags` over loopback fallbacks. Use `lifecycle.health` for the deeper endpoint proof gate before claiming service readiness; its cognitive-status probe uses a longer client timeout than the endpoint's bounded provider probes so failures return as evidence instead of socket timeouts.
`lifecycle.up` defaults to the Compose dependency lane and starts local Core/Interface only after Dockerized PostgreSQL and NATS are reachable. It does not run full `compose.up`, build application images, enable Kubernetes, or repair Docker/Rancher/WSL. `lifecycle.down` stops local app processes and leaves dependency containers running for reuse. On Linux, Next's rewritten `next-server` title is insufficient ownership evidence by itself: the listener is stoppable only when `/proc/<pid>/cwd` resolves to this repository's exact Interface directory, with `ss` used when `lsof` cannot identify the listener. An unrelated working directory remains untouched. `lifecycle.down --include-data-plane` also runs the non-destructive Compose shutdown for PostgreSQL/NATS; named volumes remain intact. It does not stop Ollama, shared external brokers, Docker/Rancher Desktop, WSL, or a Kubernetes cluster.

### Compose Tasks (`ops/compose.py`)

```bash
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
uv run inv compose.warm-cognitive
uv run inv compose.migrate
uv run inv compose.storage-health
uv run inv compose.infra-up --wait-timeout=180
uv run inv compose.infra-health
uv run inv compose.down
```

Compose is the supported single-host runtime lane. `.env.compose` owns container topology; `.env` remains the secret source.
Full bring-up resolves PostgreSQL, NATS, Core, and Interface readiness from the configured `MYCELIS_COMPOSE_*_PORT` host bindings. PostgreSQL is published on `15432` by default for local Core and host clients; Compose Core always uses `postgres:5432`. A host `psql` binary is client-only and must not be treated as a second development server.
The WSL release proof health-gates each live browser spec with `compose.health` because the runner executes specs through separate WSL shell invocations.

### Kubernetes Tasks (`ops/k8s.py`)

```bash
uv run inv k8s.up
uv run inv k8s.status
uv run inv k8s.wait
uv run inv k8s.deploy
uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml
```

Prefer Rancher Desktop K3s on Windows and `k3d` on WSL/Linux; use `MYCELIS_K8S_BACKEND=kind` only for the explicit legacy path. `MYCELIS_K8S_BACKEND=rancher` targets the existing Rancher Desktop cluster and does not create or reset it.
Do not use `k8s.reset` as Rancher Desktop, Docker Desktop, WSL, or VM repair. That task is valid only for supported repo-owned local Kubernetes backends; Rancher host repair belongs to platform tooling before rerunning `k8s.status`, `k8s.wait`, and the relevant proof lane.

### Cognitive Tasks (`ops/cognitive.py`)

```bash
uv run inv cognitive.install
uv run inv cognitive.up
uv run inv cognitive.media-gateway
uv run inv cognitive.status
uv run inv cognitive.status --litellm --litellm-endpoint=https://gateway.example.com/v1 --litellm-api-key-env=LITELLM_PROXY_API_KEY --litellm-model=mycelis-default
uv run inv cognitive.stop
```

These are optional local GPU/helper tasks, not the default path for every host; `cognitive.media-gateway` is the Windows-friendly Pinokio Forge/AUTOMATIC1111 or ComfyUI lane for local/private media generation.

### Test Tasks (`ops/test.py`) And CI Tasks (`ops/ci.py`)

```bash
uv run inv test.coverage
uv run inv interface.e2e
uv run inv ci.test
uv run inv ci.baseline
uv run inv ci.service-check
uv run inv ci.release-preflight --lane=release
uv run inv api.delivery-proof
uv run inv team.architecture-sync
```

`team.architecture-sync` is a bounded coordination check, not an autonomous implementation run. Its standing-team prompts follow the current Workspace/Outcome hierarchy and Ask-to-Recover journey, with validated execution-to-deliverable handoff as the active release gate.

Use `--live-backend` for browser proof that must hit a real Core backend. Service certification permits one serial retry after a ten-second backoff for the idempotent lifecycle bring-up following managed browser cleanup; persistent startup failure remains a hard gate and does not fall through to browser proof.

### Logging & Quality Gates (`ops/logging.py`, `ops/quality.py`)

Playwright is a repository-scoped exclusive runtime. The Interface E2E task holds a PID/session lease under `workspace/runtime/instance-locks` for its full lifecycle, including port selection and cleanup. Never bypass that lease by starting a second repo-local Playwright run: wait for the recorded owner, and stop only processes whose repository and port ownership have been verified.

```bash
uv run inv logging.check-schema
uv run inv logging.check-topics
uv run inv quality.max-lines --limit 385
```

## III. Development Workflow

Use `feature/* -> dev -> main` as the delivery ladder. Feature branches are reviewable implementation slices and must pass focused code, docs, build, and visible-workflow proof before merging. The merged `dev` checkpoint must then pass affected integration and live-service proof. Promote `dev` to `main` only from a clean commit after release preflight and required deployment proof; confirm health again after promotion and remove merged feature branches.

Compose auth boundary: Core and Interface must receive the same `MYCELIS_WEB_SESSION_SECRET` and `MYCELIS_WEB_IDENTITY_FORWARD_SECRET` references. Repo-local Compose falls back to `MYCELIS_API_KEY` for either omitted value; production deployments should provide distinct matching secrets through `.env`, never committed Compose values.

Choose the runtime lane first:
- Compose for supported home-runtime proof
- Rancher Desktop K3s or k3d/Helm for local cluster proof
- source lifecycle for implementation
- WSL proof checkout for guarded Compose release-style validation from Windows

Windows edit -> git push -> WSL refresh -> WSL validate is the guarded WSL Compose proof path:

```bash
uv run inv wsl.status
uv run inv wsl.refresh
uv run inv wsl.validate --lane=release
uv run inv wsl.validate --lane=release --headed-browser
```

The root and Interface install tasks use `npm ci` for Interface dependencies; release proof expects
dependency bootstrap to preserve a clean lockfile and fail before validation if it cannot.

### Deployment Guidance Across Host Architectures

- Windows x86_64: edit on Windows; prove through Compose, WSL Compose, or Kubernetes as the slice requires.
- Linux x86_64: use Compose for home-runtime and k3d/Helm for cluster proof.
- Linux arm64: prefer Compose or binary/control-node lanes with explicit remote AI.
- Mixed-architecture deployment rule: keep AI, database, NATS, Core, and Interface addresses explicit.

## IV. Configuration System

Configuration sources:
- `.env`: secrets and host-local runtime values
- `.env.compose`: Compose topology
- `cognitive/config/engine.yaml`: cognitive-engine shape and framework-neutral worker-runtime defaults
- `core/config/cognitive.yaml`: provider profiles/routing
- `core/config/homepage.yaml`: deployer branding/portal copy retained for authenticated entry surfaces
- `core/config/policy.yaml`: governance
- `core/config/templates/*.yaml`: bootstrap bundles/templates
- `core/config/teams/*.yaml`: standing team and legacy migration inputs
- Helm values files: cluster deployment shape

Provider/media env overrides are deployment-time infrastructure configuration only. Use `MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID`, `MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, `MYCELIS_MEDIA_MODEL_ID`, and `MYCELIS_MEDIA_GATEWAY_*`; the retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` must not return. `Bundle -> Instantiated Organization -> Inheritance -> Routing` remains authoritative, so do not treat env overrides as runtime organization behavior.

The Helm chart mounts this runtime config tree at `/core/config`. The source and chart copies of `engine.yaml` are exact and contain secret references only. For the optional local vLLM launcher, `text.api_key_secret_ref` names `env:MYCELIS_TEXT_ENGINE_API_KEY`; `cognitive.llm` and `cognitive.up` resolve it from the shell first and then the repo-local `.env`, fail if it is unavailable, and do not supply a hardcoded fallback. Core's `vllm` provider resolves that same variable through `api_key_env`.

LiteLLM configuration belongs in the cognitive provider registry, not `worker_runtime`. The shipped source and Helm provider records are disabled, set `model_gateway: true`, and use the existing `openai_compatible` contract with `api_key_env: LITELLM_PROXY_API_KEY`; the root project does not install the Python SDK, start a proxy, create LiteLLM persistence, or expose a proxy UI. An operator-managed proxy must be treated as a separate authenticated service. Keep its upstream credentials and administration key outside Core, inject only a scoped client key into Core, and partition aliases/fallbacks so local-only selection cannot reach a remote provider. Do not enable prompt/response logging, callbacks, caching, virtual-key persistence, or distributed rate limiting until their retention, redaction, tenant isolation, database/Redis topology, backup, and recovery contracts are approved. Core remains authoritative for profiles, data boundary, approval, Outcome budgets, correlation, audit, and proof; LiteLLM operational spend/rate enforcement is an additional ceiling, not the product policy source of truth. Official references: [LiteLLM documentation](https://docs.litellm.ai/) and [repository](https://github.com/BerriAI/litellm).

External model-gateway preflight reuses `cognitive.status` so the public Invoke surface stays at its 95-task cap. Supply an explicit `/v1` endpoint, `LITELLM_PROXY_API_KEY`, and expected alias as shown above. Public endpoints must use HTTPS; private, loopback, cluster-local, and service-DNS endpoints may use HTTP. Redirects are rejected so the scoped key remains on the exact operator-approved origin. The command prints only the normalized target origin, secret-reference name, and pass/fail phases; it never prints the credential or response body and never sends a completion. Run it from the same source, Compose, or Kubernetes network location that Core will use—host loopback does not prove pod/container reachability. A pass is preflight evidence only, not production certification. Before enabling the provider, separately prove boundary-partitioned aliases/fallbacks, all required call-path correlation, usage/cost reconciliation, retry/rate-limit behavior, and any approved logging, persistence, or Redis recovery posture.

Framework-worker configuration lives only under `worker_runtime` in `cognitive/config/engine.yaml`. Core checks `MYCELIS_ENGINE_CONFIG_PATH` first, then packaged `config/engine.yaml` and source `cognitive/config/engine.yaml` or `../cognitive/config/engine.yaml` paths. Source, Core image, and Helm carry the same file. Unknown fields fail closed. The committed defaults preserve central behavior:

```yaml
worker_runtime:
  backend: central
  base_url: ""
  api_key_secret_ref: ""
  capabilities_endpoint: /v1/capabilities
  health_endpoint: /health
  preferred_protocol: runs_api
  approval_mode: mycelis_control_plane
  event_stream_mode: sse
  timeout_policy: {connect_ms: 5000, run_ms: 900000, stream_ms: 900000}
```

`framework_runs` is the only external backend name. It requires an absolute HTTP(S) `base_url` without embedded credentials, query, or fragment; absolute health/capability paths; `runs_api`; SSE; central approval; and non-negative timeouts. The stream timeout remains reserved for the replay/restart supervisor in delivery slices C/D and must not be presented as active enforcement before that owner lands. `api_key_secret_ref` accepts `env:NAME` through the current deployment resolver; `secret://...` is valid only when a deployment supplies its own resolver. Never put a raw token in the YAML or URL. The committed backend stays `central`; there is no automatic external-to-central fallback. Tool access remains owned by Core's existing governed capability and execution paths rather than an unenforced worker-runtime configuration field. Do not select `framework_runs` until the exact candidate has proven durable run binding/event replay, restart reconciliation, governed approvals, idempotent stop, candidate-output validation, and fail-closed isolation. Configuration or health alone is not enablement.

The bundled `framework_runs` Python service is protocol-conformance tooling, not a production worker service. It binds loopback-only at `127.0.0.1:8091`; `MYCELIS_FRAMEWORK_RUNS_PORT` may change the local proof port, but the host is intentionally not configurable. It reports `framework runs facade ready` from its health endpoint, uses bounded process-local memory, performs no authentication of inbound requests, and refuses to start unless the operator explicitly opts into the conformance driver:

```bash
MYCELIS_FRAMEWORK_RUNS_ALLOW_CONFORMANCE=1 uv run python -m framework_runs
```

Use it only on an isolated local test boundary. Core owns the run id: the facade must accept it unchanged, return the same record for an identical create, and reject conflicting reuse. The correlation envelope carries the intent proof, execution contract, required work-item id, optional team/Outcome ids, idempotency key, source metadata, and graph revision. Completion remains candidate-only.

LangGraph remains an optional injected worker-local driver and is not installed by the root project. LangGraph Agent Server is not deployed, configured, or preflighted here: doing so would add another operator service, authentication boundary, persistence/recovery owner, tenancy posture, and run control plane before the smaller framework-neutral projection is certified. Do not point `base_url` at an Agent Server and assume protocol compatibility. The fixed delivery order is preserve the central authority/finalization reference path -> neutral durability/control/finalization -> existing LangGraph driver certification -> optional CrewAI compatibility/import work -> later AG2 or Microsoft Agent Framework evaluation. CrewAI and AG2 are not installed, backend names, operator-selectable choices, or P0.10 dependencies. Every later adapter requires an explicit owner and separate conformance, migration, security, recovery, and removal/lifetime decision. Microsoft AutoGen is not supported as a new dependency. Follow the canonical PRD's [P0.10 delivery referential](../architecture-library/MYCELIS_CANONICAL_PRD.md#p010-framework-execution-delivery-referential) for source ownership and deployment gates.

LiteLLM is configured separately in the cognitive provider registry. An adapter-backed worker may call a policy-approved LiteLLM model alias using a scoped secret, but this does not make LiteLLM a worker backend or give either component Mycelis approval, audit, retained-output, or completion authority. Certify model transport and worker execution independently.

Target posture for slice B of the neutral durability gate: the production Runs service is a separately deployed, internal-only service—never an embedded Core runtime, Core sidecar, public endpoint, or repurposed conformance process. Its Compose and Kubernetes posture is disabled by default, privately authenticated, independently persistent, and denied direct NATS, Core database, Core workspace, Docker socket, browser, and operator-secret access. Rollout and rollback must preserve accepted-run bindings and evidence; do not redirect an accepted external run to `central`.

Runtime bootstrap follows the runtime architecture and settings sections of the [Mycelis Canonical PRD](../architecture-library/MYCELIS_CANONICAL_PRD.md): template inputs become instantiated organizations through inheritance, precedence, and policy checks.

Startup should select a valid bundle, instantiate the runtime organization from it, fail closed if no valid bundle is available, and require `MYCELIS_BOOTSTRAP_TEMPLATE_ID` when multiple bundles are mounted. `lifecycle.up` ensures the `cortex` database exists before Core starts.

## V. Testing Strategy

Use [Testing](../TESTING.md) for gate details. Operational summary:
- backend: `uv run inv core.test`
- frontend: `uv run inv interface.test` and `interface.typecheck`
- browser: `uv run inv interface.e2e`
- live stack: `lifecycle.health` or `compose.health`
- release: `ci.baseline` or `wsl.validate --lane=release`
- visible live-window proof: `wsl.validate --lane=release --headed-browser` or focused `interface.e2e --headed --live-backend`

Runtime checks must start clean, verify readiness, run proof once services are healthy, and shut down unless a follow-on check needs them.

Clean first-boot release proof must exercise the normal data-plane path and the normal app startup path:

```bash
uv run inv lifecycle.first-boot-proof
```

The task stops local app services, keeps Dockerized PostgreSQL/NATS volumes available, resets the app database, clears generated workspace output roots while preserving local mounts, starts Core/Interface, runs health checks, verifies empty user product state, restarts Core/Interface, and verifies bootstrap row counts remain stable. The next gate is one Soma-created Outcome from ask to approved execution, isolated deliverable, validation, direct open/reply/recover actions, and scoped cleanup.

## VI. CI/CD

Local CI tasks:
- `ci.test`: Go + Interface tests
- `ci.baseline`: branch-readiness gate
- `ci.service-check`: currently running local stack
- `ci.release-preflight`: lane-aware release gate that runs `ci.lint` first and stops before later stages on lint failure

GitHub CI proves repo health without hosted agentry. Live service/browser proof is local, WSL, Compose, or target-cluster evidence. Hosted workflow maintenance stays on Node 24-capable action majors, runs Interface CI with Node.js 24, uses checksum-verified pinned Helm 3 instead of `azure/setup-helm@v4`, and requires self-hosted runners that support Node 24 actions.
Invoke manages the Next.js server lifecycle for browser proof. Merge gates stop repo-owned Interface processes and managed Playwright listeners, verify service ownership before relying on occupied ports, and clear `.next` before production builds so Windows file locks cannot invalidate the candidate, then use the built production Interface server path. Windows process inventory uses one bounded repo-scoped Node/cmd query, and Windows tree termination has an extended bounded timeout so loaded hosts can finish descendant enumeration before any parent-only fallback. The default baseline targets the managed Chromium binary installed by the supported bootstrap path with one worker; cross-engine matrices require explicit browser provisioning. CI also keeps a Chromium authenticated-front-door smoke. Playwright reporter ownership remains in `interface/playwright.config.ts`, which emits readable output and retained JSON/JUnit evidence under `interface/test-results`; task wrappers must not replace that reporter set. CI Go gates address the Core module explicitly with `go -C` rather than relying on shell `cd` state, and release baseline Core tests run with `-p 1` so local runtime/browser probes are not destabilized by package-level contention. A failed captured Core command prints console-safe diagnostic output even when the Windows host code page cannot represent a test character. Run managed Interface build/test/browser proof serially for a workspace and port because those gates own shared Next/Vitest workers and server ports. Use `uv run inv ci.service-check` for the currently running stack. `ci.release-preflight` supports `--lane=baseline|runtime|service|release`; every lane runs lint first, and `--lane=release` is the recommended full runtime/operator gate. Its runtime posture rejects loopback AI endpoints except for a WSL source endpoint that exactly matches the scheme and port of the explicit `host.docker.internal` Compose contract used for Windows mirrored networking. Guarded WSL proof-checkout tasks are `wsl.status`, `wsl.refresh`, `wsl.validate`, `wsl.cycle`; add `--headed-browser` to `wsl.validate` or `wsl.cycle` when focused live-backend Playwright proof must open visible browser windows.
`uv run inv interface.check` includes a small retry loop for transient Windows socket-reuse errors after heavy browser proof, but persistent route failures still fail the task.

## VII. Deployment Architecture

### Kubernetes (Self-Hosted / Helm)

This chart is the target clustered deployment contract for self-hosted and enterprise Kubernetes. Use Helm with explicit values, secrets, persistent storage, ingress, and reachable AI endpoints. Use `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml` for chart posture. local Kubernetes now prefers `k3d` when it is installed on WSL/Linux, prefers Rancher Desktop K3s on Windows, and falls back with `MYCELIS_K8S_BACKEND=kind`.

Open-standard resources include Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy. `MYCELIS_K8S_VALUES_FILE` may select `values-k3d.yaml`, `values-enterprise.yaml`, or `values-enterprise-windows-ai.yaml`. `MYCELIS_K8S_TEXT_ENDPOINT` and `MYCELIS_K8S_TEXT_MODEL_ID` project provider endpoint/model overrides into Core, and configured text endpoints enable explicit AI egress in the chart NetworkPolicy.

### Docker

Docker Compose is rapid local development/proof only and must not become the production deployment standard. Use Compose for single-host self-hosted runtime. Keep `MYCELIS_COMPOSE_OLLAMA_HOST` container-reachable and avoid `localhost` unless it is meaningful inside the container path. use explicit reachable AI endpoints for deployed text or media engines instead of localhost assumptions; the Helm chart applies `MYCELIS_K8S_TEXT_ENDPOINT` through provider-specific env overrides.

use a reachable Windows IP or hostname such as `http://192.168.x.x:11434/v1`, not `localhost`; Compose can auto-start a WSL-host relay for `MYCELIS_COMPOSE_OLLAMA_HOST`, that relay is restartable, and `compose.warm-cognitive` proves the configured text model can complete a tiny chat before live browser proof starts.

the compose Core image includes Node/npm/npx so manual curated stdio MCP installs can launch from the shipped container; manual `filesystem` installs from the curated library are runtime-normalized to the configured `MYCELIS_WORKSPACE` root.

Local stdio MCP startup allows a bounded 60-second connection window so an intentionally cleaned npm cache can repopulate. Connection failures persist their error state through a separate short database deadline, preventing an expired launch context from leaving a stale `connected` status.

### Persistent Storage Contract

Docker Compose `pgvector/pgvector:pg16` is the sole development PostgreSQL server. Its `postgres-data` volume holds both relational data and pgvector-backed memory/context; these are not separate databases or host stores. Generated files, browser game packages, filesystem MCP writes, and retained project outputs land under `MYCELIS_WORKSPACE`. File-backed artifacts and cached media land under `MYCELIS_ARTIFACT_ROOT`; `DATA_DIR` remains a legacy fallback and should stay aligned with `MYCELIS_ARTIFACT_ROOT` until it is removed. Declarative YAML/JSON configuration lives beneath `MYCELIS_CONFIG_ROOT`, which defaults to `${MYCELIS_WORKSPACE}/config`; file inputs outside that resolved root fail closed. Compose maps workspace and artifact paths into `/data/workspace` and `/data/artifacts`; Kubernetes uses the chart output block/PVC mounted at `/data`. Local source Core live-browser proof infers the host-visible workspace from the loaded `.env`/process `MYCELIS_WORKSPACE`, using absolute roots directly and mapping repo-local `./workspace` to `core/workspace`; K8s live-browser proof that checks backend-written files should use `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s` so the assertion targets the Core pod workspace/PVC rather than host-only paths.

`uv run inv compose.migrate` uses the same complete schema and storage compatibility contract as source mode. It installs `001_current_schema.sql` atomically into an empty public schema, makes a row-preserving no-op for a fully compatible retained volume, and rejects any nonempty partial or incompatible schema before installer execution. Back up retained data, then use an explicitly supported upgrade path or `uv run inv compose.down --volumes` only when an intentional disposable reset is appropriate. Pair migration with `uv run inv compose.storage-health` before treating the data plane as ready.

Live retained-template proof claims each created ConfigDocument inside the existing confirmation transaction. Cleanup purges only the exact record owned by the fixture scope and restores the activation that preceded the test. Do not delete configuration rows by name prefix or clear activation history as general test cleanup.

### Startup Sequence

Core startup depends on resolved config, database, NATS, provider posture, policy, and bootstrap bundle availability. Normal startup should fail closed when required bootstrap inputs are missing.

### Graceful Shutdown

Use task-owned teardown:

```bash
uv run inv lifecycle.down
uv run inv compose.down
```

### Readiness Probe

Use task health checks rather than raw port checks when claiming product readiness.

## VIII. Environment Gotchas

- Windows `OLLAMA_HOST=0.0.0.0` is a listen address, not a container-reachable endpoint.
- Docker-in-WSL may require `MYCELIS_WSL_DISTRO`.
- Compose may relay a Windows-hosted AI endpoint through WSL for bridge containers.
- Do not share generated environments across Windows and WSL.
- Browser proof that checks backend-written files infers the active native Core workspace from `.env`/process `MYCELIS_WORKSPACE`; Compose or split-checkout proof may still set `MYCELIS_BACKEND_WORKSPACE_ROOT`.
- K8s/PVC browser proof should use `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s`, `PLAYWRIGHT_K8S_NAMESPACE`, `PLAYWRIGHT_K8S_CORE_SELECTOR`, and `PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT`.

## IX. Monitoring & Observability

Use:
- `lifecycle.status` / `lifecycle.health`
- `compose.status` / `compose.health`
- `k8s.status`
- structured Core logs
- NATS and PostgreSQL health probes
- activity/run timeline in the UI
Operator-facing errors should be normalized and human-readable; raw backend noise belongs in logs, not the default UI.

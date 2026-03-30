# V8 Home Docker Compose Runtime

> Status: ACTIVE
> Last Updated: 2026-03-29
> Purpose: Define the supported single-host Docker Compose runtime for home and small-team operators who want Mycelis without Kind/Kubernetes.

## Goal

Provide a supported local runtime path that keeps:
- Soma-first product behavior
- governed execution and approvals
- pgvector-backed PostgreSQL
- NATS messaging
- operator-readable health and logs

without requiring a Kind/Kubernetes cluster.

## Supported Scope

The Docker Compose runtime is for:
- home-lab operators
- local product demos
- partner/funder walkthroughs
- single-host technical evaluation

It is not the replacement for the chart/Kubernetes posture when:
- cluster semantics are the thing being tested
- distributed deployment behavior matters
- rollout/readiness/PVC/network-policy behavior is the target

## Runtime Topology

The compose stack owns:
- `postgres` for pgvector-backed persistence
- `nats` for messaging and NATS monitoring
- `core` for the Go runtime
- `interface` for the Next.js product surface

Default cognitive posture:
- compose assumes a host-available Ollama endpoint unless the operator deliberately changes `OLLAMA_HOST`
- the default example points Core at `http://host.docker.internal:11434`

## Env Contract

Compose uses `.env.compose`, not `.env`, so the home runtime can keep container-host assumptions separate from Kind/bridge assumptions.

Required:
- `MYCELIS_API_KEY`

Important defaults:
- `DB_HOST=postgres`
- `DB_PORT=5432`
- `DB_USER=mycelis`
- `DB_PASSWORD=password`
- `DB_NAME=cortex`
- `NATS_URL=nats://nats:4222`
- `MYCELIS_BOOTSTRAP_TEMPLATE_ID=v8-migration-standing-team-bridge`
- `MYCELIS_COMPOSE_OLLAMA_HOST=http://host.docker.internal:11434`
- `MYCELIS_DISABLE_DEFAULT_MCP_BOOTSTRAP=true`
- `DATA_DIR=/data/artifacts`
- `MYCELIS_WORKSPACE=/data/workspace`

Compose guardrail:
- `MYCELIS_COMPOSE_OLLAMA_HOST` in `.env.compose` must be container-reachable. `localhost`, `127.0.0.1`, and `0.0.0.0` are invalid for the home-runtime Core container.
- The compose-specific variable name is intentional so host-machine `OLLAMA_HOST` bind settings do not override container runtime inference routing.

Host port defaults stay aligned with the normal local operator path:
- frontend `3000`
- core `8081`
- postgres `5432`
- nats `4222`
- nats monitor `8222`

## Storage Contract

Compose persistence is split intentionally:
- named volume for PostgreSQL data
- named volume for NATS JetStream data
- bind-mounted repo path at `workspace/docker-compose/data` for Core `/data`

That keeps:
- database and bus state durable across restarts
- generated artifacts and workspace files inspectable from the repo workspace

## Monitoring And Logging Contract

Compose cannot reproduce the full chart/Kubernetes monitoring posture, but it must still stay operator-readable.

Required compose monitoring surfaces:
- `uv run inv compose.status`
- `uv run inv compose.health`
- `uv run inv compose.logs`

Minimum standards:
- every compose service uses bounded json-file logging rotation
- PostgreSQL has a container healthcheck
- Core has a `/healthz` healthcheck
- Interface has an HTTP root healthcheck
- NATS exposes monitoring on `:8222` and compose health checks probe it from the host task layer
- compose health must fail when the text cognitive engine is offline, even if the API is still responding

Compose MCP note:
- the slim home-runtime Core image does not bundle `npm`/`npx`, so compose disables default npm-backed MCP auto-install by default
- this does not remove manual MCP capability; it only keeps startup logs honest for the home-runtime image

## Managed Task Contract

Supported task path:
- `uv run inv compose.up`
- `uv run inv compose.down`
- `uv run inv compose.migrate`
- `uv run inv compose.status`
- `uv run inv compose.health`
- `uv run inv compose.logs`

`compose.up` is the canonical bring-up path because it:
1. starts PostgreSQL and NATS first
2. waits for local ports to bind
3. applies canonical migrations through the PostgreSQL container
4. starts Core and Interface
5. verifies Core `/healthz` and the Interface root

## Release/Testing Guidance

For the compose runtime, clean proof should use this order:
1. `uv run inv compose.down --volumes`
2. `uv run inv compose.up --build`
3. `uv run inv compose.status`
4. `uv run inv compose.health`
5. focused browser/product proof against the compose-hosted stack

Classification reminder:
- compose startup failure is `environment`
- compose task contract drift is `test` or `docs`
- visible broken product behavior after compose health is green is `product`

## Non-Negotiable Rule

The compose runtime is a supported operator path, not a lesser unofficial shortcut.

That means:
- docs must stay synchronized with it
- task automation must stay synchronized with it
- health/logging expectations must stay explicit
- release review may use it for single-host product proof when cluster behavior is not the thing under test

# V8 Home Docker Compose Runtime
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: ACTIVE
> Last Updated: 2026-04-12
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
- `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted`
- `MYCELIS_OUTPUT_HOST_PATH=./workspace/docker-compose/data`
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
- bind-mounted output block path at `MYCELIS_OUTPUT_HOST_PATH` for Core `/data`

That keeps:
- database and bus state durable across restarts
- generated artifacts and workspace files inspectable from the configured host output block

Output block modes:
- `local_hosted`: the operator provides a real host directory in `MYCELIS_OUTPUT_HOST_PATH`. This is the Docker Compose and Pinokio/local-media handoff path.
- `cluster_generated`: the chart/Kubernetes path owns `/data` through the cluster-managed PVC. Compose keeps its repo-managed default path when this mode is used outside a cluster.

The Invoke compose task validates the host path with Python `pathlib` before Docker starts. It expands `~` and environment variables, accepts platform-native path syntax, rejects file paths, and fails early when `local_hosted` points at a missing directory instead of letting Docker silently create or mount the wrong output block.

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
- `uv run inv compose.infra-up`
- `uv run inv compose.infra-health`
- `uv run inv compose.storage-health`
- `uv run inv compose.up`
- `uv run inv compose.down`
- `uv run inv compose.migrate`
- `uv run inv compose.status`
- `uv run inv compose.health`
- `uv run inv compose.logs`

`compose.infra-up` is the data-plane-only path because it:
1. starts PostgreSQL and NATS only
2. leaves Core and Interface down
3. waits for PostgreSQL and NATS readiness
4. prints same-project, host-native, and separate-Compose-project connection settings
5. keeps migrations explicit through `--migrate` or `compose.migrate`

`compose.infra-health` is the matching data-plane-only probe because it:
1. checks the PostgreSQL host port
2. checks PostgreSQL query readiness with the configured DB user/name
3. checks the NATS host port
4. checks the NATS monitor endpoint
5. does not check Core or Interface

`compose.storage-health` is the post-migration long-term storage probe because it:
1. checks the `vector` extension for pgvector-backed semantic recall
2. checks `context_vectors` for durable RAG/deployment context storage
3. checks `agent_memories` for reviewed Soma/agent memory
4. checks `conversation_turns` for conversation continuity
5. checks `artifacts` for retained outputs
6. checks `temp_memory_channels` for restart-safe temporary continuity
7. checks `collaboration_groups` for temporary/standing group persistence
8. checks managed exchange tables for channel/thread/item retention
9. checks `conversation_templates` for reusable ask/template recall

When the base runtime schema is already compatible, `compose.migrate` skips unsafe full replay of older migrations but still applies known missing late storage migrations before the storage-health gate is allowed to pass.

`compose.up` is the canonical full-stack bring-up path because it:
1. starts PostgreSQL and NATS first
2. waits for local ports to bind
3. applies canonical migrations through the PostgreSQL container
4. starts Core and Interface
5. verifies Core `/healthz` and the Interface root

## Release/Testing Guidance

For the compose runtime, clean proof should use this order:
1. `uv run inv compose.down --volumes`
2. `uv run inv compose.infra-up --wait-timeout=180`
3. `uv run inv compose.infra-health`
4. `uv run inv compose.migrate`
5. `uv run inv compose.storage-health`
6. `uv run inv compose.status`
7. `uv run inv compose.up --build`
8. `uv run inv compose.health`
9. focused browser/product proof against the compose-hosted stack

Use [V8 Compose Personal Owner Deployment Test Plan](V8_COMPOSE_PERSONAL_OWNER_DEPLOYMENT_TEST_PLAN.md) for the thorough PRD-style validation path that starts with the shared data plane and proceeds into near-enterprise owner workflow proof.

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

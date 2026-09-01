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

Internal tools are governed runtime capabilities. Tool metadata must identify source, scope, and intended consumer.

### Composite Tool Executor - `swarm/tool_executor.go`

Composite execution must preserve bounded outputs, error normalization, and auditability.

### Blueprint Activation - `swarm/activation.go`, `converter.go`, `seeds.go`

Blueprint activation turns approved intent into teams, agents, events, and persisted run state.

## III. Cognitive Layer

### Router - `cognitive/router.go`

The router resolves provider policy and profile routing into a concrete model call path. Deployment/env overrides configure endpoints and profiles; they do not replace instantiated organization truth.

### Adapters

Adapters normalize provider-specific request/response behavior for supported model backends.

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

`internal/workers.WorkerBackend` normalizes create, event stream, read, stop, approval, capability, and health operations. `central` remains the default and the only backend enabled for confirmed Mycelis execution. `framework_runs` is an HTTP client for `/v1/runs`, `/v1/runs/{id}`, `/v1/runs/{id}/events`, `/v1/runs/{id}/stop`, `/v1/runs/{id}/approvals/{approval_id}`, `/v1/capabilities`, and `/health`. It normalizes common external status/event/output shapes, requires the durable `runs_api` protocol, sends credentials only after resolving a managed secret reference, and deliberately does not implement `RunFinalizer`.

The Python `framework_runs` package is a bounded, process-isolated protocol facade. Its built-in conformance driver and in-memory store are non-production; the store defaults to 256 runs and 256 events per run, prunes only the oldest terminal run, and reports capacity failure while every slot is active. Its SSE endpoint is a finite snapshot of retained events rather than a production live-event bus. All returned completions carry `completion_authority=candidate`, `requires_core_validation=true`, `verified=false`, and `execution_authority=mycelis_core`. The optional `LangGraphDriver` accepts an injected compiled graph only when the optional dependency is installed. Core's execution selector rejects `framework_runs` until durable event projection is wired, so setting the external backend cannot silently bypass central policy, audit, retained-output validation, or completion proof.

Framework posture was rechecked against official sources on 2026-09-01. [LangGraph](https://docs.langchain.com/oss/python/langgraph/overview) is the first adapter candidate because its documented low-level runtime centers durable execution, streaming, persistence, and human-in-the-loop control. [AG2 1.0](https://github.com/ag2ai/ag2) is distinct from AG2 Classic and is not a drop-in upgrade, so each would require an explicit adapter and conformance proof. [CrewAI](https://docs.crewai.com/) combines crews with persistent, resumable flows and remains another adapter candidate rather than an embedded authority. [Microsoft AutoGen](https://github.com/microsoft/autogen) is in maintenance mode and directs new projects to [Microsoft Agent Framework](https://learn.microsoft.com/en-us/agent-framework/overview/), which is a separate future evaluation target. None is installed or enabled by this slice, and none may replace Mycelis' Outcome, policy, approval, audit, projection, or validation boundaries.

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

## IX. Governance & Policy Engine

Deploy-owned identity posture is backend-owned: deploy-owned People & Access posture surfaced read-only, and settings PUT ignores/preserves those deploy-owned fields instead of persisting them.

### Guard - `governance/guard.go`

The guard decides whether an action is allowed, blocked, or requires proposal/approval. It must be deterministic and auditable.

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

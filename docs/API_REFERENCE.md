# API Reference

> Back to [README](../README.md) | See also: [Swarm Operations](SWARM_OPERATIONS.md) | [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

## Endpoints

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| **Council Chat** | | |
| `/api/v1/council/{member}/chat` | POST | Chat with any council member via NATS request-reply. Returns `APIResponse<CTSEnvelope>` with trust score + provenance |
| `/api/v1/council/members` | GET | List all addressable council members from standing teams (admin-core, council-core) |
| **Chat & Cognitive** | | |
| `/api/v1/chat` | POST | Legacy: chat routed through Admin agent via NATS request-reply (backward compat — prefer `/council/{member}/chat`) |
| `/api/v1/cognitive/infer` | POST | Direct cognitive inference (profile-routed) |
| `/api/v1/cognitive/config` | GET | Read cognitive router configuration (providers, profiles, media) |
| `/api/v1/cognitive/matrix` | GET | Alias for cognitive config (matrix view) |
| `/api/v1/cognitive/status` | GET | Live health probe of text (vLLM/Ollama) + media (Diffusers) engines |
| `/api/v1/cognitive/profiles` | PUT | Update profile→provider routing (persists to cognitive.yaml) |
| `/api/v1/cognitive/providers/{id}` | PUT | Configure provider (endpoint, model_id, api_key, api_key_env) |
| **Intent & Missions** | | |
| `/api/v1/intent/negotiate` | POST | Blueprint generation from natural language intent |
| `/api/v1/intent/commit` | POST | Instantiate mission from blueprint |
| `/api/v1/intent/seed/symbiotic` | POST | Seed Gmail+Weather mission (no LLM required) |
| `/api/v1/missions` | GET | List missions with team/agent counts |
| `/api/v1/missions/{id}` | GET | Full mission detail with teams and agent manifests |
| `/api/v1/missions/{id}` | DELETE | Delete mission (cascade to teams/agents, deactivates Soma runtime) |
| `/api/v1/missions/{id}/agents/{name}` | PUT | Update agent manifest within an active mission |
| `/api/v1/missions/{id}/agents/{name}` | DELETE | Remove agent from an active mission |
| **Swarm & Teams** | | |
| `/api/v1/teams` | GET | List active teams |
| `/api/v1/teams/detail` | GET | Team detail with agent rosters, delivery targets, and status |
| `/api/swarm/teams` | POST | Create team via Soma |
| `/api/swarm/command` | POST | Send command to specific team |
| `/api/v1/swarm/broadcast` | POST | Fan out directive to ALL active teams |
| `/agents` | GET | List active agents with heartbeat status |
| **Telemetry & Trust** | | |
| `/api/v1/stream` | GET (SSE) | Real-time NATS signal stream |
| `/api/v1/telemetry/compute` | GET | Goroutines, heap, system memory, LLM tokens/sec |
| `/api/v1/trust/threshold` | GET/PUT | Read/write autonomy threshold |
| `/api/v1/sensors` | GET | Sensor library (static + dynamic) |
| **Memory & Search** | | |
| `/api/v1/memory/search` | GET | Semantic vector search over SitReps |
| `/api/v1/memory/sitreps` | GET | Recent SitReps (filterable by team) |
| `/api/v1/memory/stream` | GET | Live memory/log stream (polling) |
| `/api/v1/memory/sitrep` | POST | Trigger SitRep generation |
| **Governance & Proposals** | | |
| `/api/v1/proposals` | GET/POST | Team manifestation proposals |
| `/api/v1/proposals/{id}/approve` | POST | Approve proposal |
| `/api/v1/proposals/{id}/reject` | POST | Reject proposal |
| `/admin/approvals` | GET | List pending governance approvals |
| `/admin/approvals/{id}` | POST | Approve/reject governance action |
| **Agent Catalogue** | | |
| `/api/v1/catalogue/agents` | GET/POST | List/create agent blueprints |
| `/api/v1/catalogue/agents/{id}` | PUT/DELETE | Update/delete agent blueprint |
| **Artifacts** | | |
| `/api/v1/artifacts` | GET/POST | List/store artifacts (filterable) |
| `/api/v1/artifacts/{id}` | GET | Artifact detail |
| `/api/v1/artifacts/{id}/status` | PUT | Update artifact status |
| **MCP Ingress** | | |
| `/api/v1/mcp/install` | POST | Raw MCP install endpoint — **disabled by Phase 0 security** (`403`), use library install |
| `/api/v1/mcp/servers` | GET | List installed MCP servers |
| `/api/v1/mcp/servers/{id}` | DELETE | Remove MCP server |
| `/api/v1/mcp/tools` | GET | List all MCP tools across servers |
| `/api/v1/mcp/servers/{id}/tools/{tool}/call` | POST | Invoke a specific MCP tool |
| `/api/v1/mcp/library` | GET | Browse curated MCP server library (categorized) |
| `/api/v1/mcp/library/install` | POST | One-click install from library by name |
| `/api/v1/mcp/toolsets` | GET | List MCP tool sets (`tenant_id='default'`) |
| `/api/v1/mcp/toolsets` | POST | Create MCP tool set (name, description, tool_refs) |
| `/api/v1/mcp/toolsets/{id}` | PUT | Update MCP tool set by ID (`404` if not found) |
| `/api/v1/mcp/toolsets/{id}` | DELETE | Delete MCP tool set by ID |
| **Governance Policy** | | |
| `/api/v1/governance/policy` | GET/PUT | Read/update governance policy rules |
| `/api/v1/governance/pending` | GET | List pending governance approvals |
| `/api/v1/governance/resolve/{id}` | POST | Approve/reject a pending governance action |
| **Provisioning & Registry** | | |
| `/api/v1/provision/draft` | POST/GET | Draft provisioning plan |
| `/api/v1/provision/deploy` | POST | Deploy provisioned resources |
| `/api/v1/registry/templates` | GET/POST | List/register connector templates |
| `/api/v1/teams/{id}/connectors` | POST | Install connector on team |
| `/api/v1/teams/{id}/wiring` | GET | Get team wiring graph |
| **Identity** | | |
| `/api/v1/user/me` | GET | Current user identity |
| `/api/v1/user/settings` | PUT | Update user settings |
| `/healthz` | GET | Health check |
| **Brains (Provider CRUD)** | | |
| `/api/v1/brains` | GET | List all providers with health status, location, data boundary |
| `/api/v1/brains` | POST | Add a new provider — hot-injects into running router, immediate probe |
| `/api/v1/brains/{id}` | PUT | Update provider config — empty api_key keeps existing |
| `/api/v1/brains/{id}` | DELETE | Remove provider — rejected if last remaining |
| `/api/v1/brains/{id}/toggle` | PUT | Enable/disable provider — persists to cognitive.yaml |
| `/api/v1/brains/{id}/policy` | PUT | Update usage_policy + roles_allowed — persists to cognitive.yaml |
| `/api/v1/brains/{id}/probe` | POST | Live health check — returns `{"alive":bool,"latency_ms":int}` |
| **Mission Profiles** | | |
| `/api/v1/mission-profiles` | GET | List all profiles (role_providers, subscriptions, active flag) |
| `/api/v1/mission-profiles` | POST | Create profile — name, role_providers, subscriptions, context_strategy, auto_start |
| `/api/v1/mission-profiles/{id}` | PUT | Update profile config |
| `/api/v1/mission-profiles/{id}` | DELETE | Delete profile — unsubscribes reactive engine first |
| `/api/v1/mission-profiles/{id}/activate` | POST | Activate profile — applies role routing, registers NATS subscriptions, marks active in DB |
| **Context Snapshots** | | |
| `/api/v1/context/snapshot` | POST | Save a context snapshot (messages, run_state, role_providers, source_profile) |
| `/api/v1/context/snapshots` | GET | List 20 most recent snapshots (no messages payload) |
| `/api/v1/context/snapshots/{id}` | GET | Full snapshot including messages array |
| **Service Health** | | |
| `/api/v1/services/status` | GET | Aggregate health — NATS, PostgreSQL (with latency), Cognitive (N/M enabled), Reactive (subscription count) |
| **Mission Runs & Events (V7)** | | |
| `/api/v1/runs` | GET | List recent runs across all missions — status, timing, trigger source |
| `/api/v1/runs/{id}/events` | GET | Full event timeline for a run (MissionEventEnvelope records) |
| `/api/v1/runs/{id}/chain` | GET | Causal chain — parent run → event → trigger → child run traversal |
| **Intent (CE-1)** | | |
| `/api/v1/intent/confirm-action` | POST | Consume confirm token, execute mutation, return run_id |
| `/api/v1/intent/proof/{id}` | GET | Retrieve intent proof bundle by ID |
| `/api/v1/templates` | GET | List CE-1 orchestration templates |

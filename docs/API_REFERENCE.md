# API Reference
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

> Back to [README](../README.md) | See also: [Swarm Operations](SWARM_OPERATIONS.md) | [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

## API TOC

- [Endpoints](#endpoints)
- [Provider Auth Notes](#provider-auth-notes)

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
| `/api/v1/memory/search` | GET | Semantic vector search over durable memory, with optional team/agent/type scope filters across Soma-personal, team-shared, and governed memory lanes |
| `/api/v1/search/status` | GET | Current Mycelis Search provider posture for UI/Soma capability answers, including provider, configured/enabled flags, direct `web_search` support, token requirements, and blocker/next-action copy |
| `/api/v1/search` | POST | Governed Mycelis Search API over `local_sources`, optional self-hosted `searxng`, optional hosted search/MCP bridge posture, or structured disabled-provider blockers |
| `/api/v1/memory/deployment-context` | GET/POST | List or load governed user/private content, deployment knowledge, admin-owned Soma context, and reflection/synthesis observations into separate `user_private_context`, `customer_context`, `company_knowledge`, `soma_operating_context`, and `reflection_synthesis` pgvector stores; POST supports text content plus metadata such as `content_domain` and `target_goal_sets`. This is governed source/doctrine intake, not implicit team-shared `AGENT_MEMORY`. |
| `/api/v1/memory/sitreps` | GET | Recent SitReps (filterable by team) |
| `/api/v1/memory/stream` | GET | Live memory/log stream (polling) |
| `/api/v1/memory/sitrep` | POST | Trigger SitRep generation |
| **Managed Exchange** | | |
| `/api/v1/exchange/fields` | GET | List exchange field definitions, including learning-candidate fields such as `classification`, `memory_layer`, `confidence`, `review_required`, `promotion_target`, and `evidence_refs` |
| `/api/v1/exchange/schemas` | GET | List exchange schemas, including `LearningCandidate` for classified reflection/learning candidates before memory promotion |
| `/api/v1/exchange/channels` | GET | List governed exchange channels such as `organization.learning.candidates` |
| `/api/v1/exchange/items` | GET/POST | List or publish structured exchange items; `LearningCandidate` items are the candidate-first boundary before reflection, team, or governed durable-memory promotion |
| **Governance & Proposals** | | |
| `/api/v1/proposals` | GET/POST | Team manifestation proposals |
| `/api/v1/proposals/{id}/approve` | POST | Approve proposal |
| `/api/v1/proposals/{id}/reject` | POST | Reject proposal |
| `/admin/approvals` | GET | List pending governance approvals |
| `/admin/approvals/{id}` | POST | Approve/reject governance action |
| **Agent Catalogue** | | |
| `/api/v1/catalogue/agents` | GET/POST | List/create agent blueprints |
| `/api/v1/catalogue/agents/{id}` | PUT/DELETE | Update/delete agent blueprint |
| **Template Marketplace (Planned V7.x)** | | |
| `/api/v1/template-market/sources` | GET/POST | List/register marketplace sources (clawhub/private hubs) |
| `/api/v1/template-market/sources/{source_id}` | PATCH | Update source credentials/status |
| `/api/v1/template-market/sources/{source_id}/probe` | POST | Probe source connectivity/auth |
| `/api/v1/template-market/templates` | GET | Discover templates across configured sources |
| `/api/v1/template-market/templates/{template_id}` | GET | Template package detail (requirements/pricing/license) |
| `/api/v1/template-market/templates/{template_id}/purchase-intent` | POST | Create governed purchase proposal |
| `/api/v1/template-market/purchases/{purchase_id}/confirm` | POST | Confirm purchase after approval |
| `/api/v1/template-market/templates/{template_id}/install` | POST | Install template into tenant catalog |
| `/api/v1/template-market/installs` | GET | List installed marketplace templates |
| `/api/v1/template-market/installs/{install_id}/upgrade` | POST | Upgrade installed template version |
| `/api/v1/templates/custom` | GET/POST | List/create tenant custom templates |
| `/api/v1/templates/custom/{template_id}` | GET/PUT/DELETE | Read/update/archive custom template |
| `/api/v1/templates/custom/{template_id}/publish` | POST | Publish immutable custom template version |
| `/api/v1/templates/custom/{template_id}/fork` | POST | Fork marketplace/builtin/custom template |
| **Artifacts** | | |
| `/api/v1/artifacts` | GET/POST | List/store artifacts (filterable) |
| `/api/v1/artifacts/{id}` | GET | Artifact detail |
| `/api/v1/artifacts/{id}/download` | GET | Download a stored or inline artifact as a file attachment for chat/operator review |
| `/api/v1/artifacts/{id}/status` | PUT | Update artifact status |
| `/api/v1/artifacts/{id}/save` | POST | Persist cached image artifact to workspace folder (`saved-media` default) |
| **MCP Ingress** | | |
| `/api/v1/mcp/install` | POST | Raw MCP install endpoint — **disabled by Phase 0 security** (`403`), use library install |
| `/api/v1/mcp/servers` | GET | List installed MCP servers |
| `/api/v1/mcp/servers/{id}` | DELETE | Remove MCP server |
| `/api/v1/mcp/tools` | GET | List all MCP tools across servers |
| `/api/v1/mcp/activity` | GET | List recent persisted MCP activity from Managed Exchange, including server/tool/state visibility for operator review |
| `/api/v1/mcp/servers/{id}/tools/{tool}/call` | POST | Invoke a specific MCP tool |
| `/api/v1/mcp/library` | GET | Browse curated MCP server library (categorized), including server.json-aligned metadata such as version, package transport, repository/homepage links when known, and typed environment-variable declarations |
| `/api/v1/mcp/library/inspect` | POST | Policy inspection preview for a library candidate (`allow|require_approval|deny`) before install. MCP settings installs may send `governance_context` so owner-scoped current-group config can auto-allow without a second approval loop |
| `/api/v1/mcp/library/install` | POST | Apply/install from library by name. Allowed installs are idempotent by server name: a repeated install updates/reconnects the existing server instead of failing on duplicate registry state. Curated `filesystem` installs are runtime-normalized to the configured workspace root before persistence/launch. Returns `202` with inspection details when the candidate still requires approval |
| `/api/v1/mcp/library/apply` | POST | One-call inspect/apply path for curated MCP candidates. Allowed installs are idempotent by server name, curated `filesystem` installs use the deployment workspace root, and success returns `status=installed` with server/tools/governance; boundary cases return `status=requires_approval` with inspection details |
| `/api/v1/mcp/toolsets` | GET | List MCP tool sets (`tenant_id='default'`) |
| `/api/v1/mcp/toolsets` | POST | Create MCP tool set (name, description, tool_refs, optional `governance_context`) |
| `/api/v1/mcp/toolsets/{id}` | PUT | Update MCP tool set by ID (`404` if not found, optional `governance_context`) |
| `/api/v1/mcp/toolsets/{id}` | DELETE | Delete MCP tool set by ID (response includes normalized governance posture) |
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
| `/api/v1/user/me` | GET | Current user identity, including normalized principal metadata (`principal_type`, `auth_source`, `effective_role`, `break_glass`) plus the deploy-owned People & Access contract surfaced read-only through `settings` (`access_management_tier`, `product_edition`, `identity_mode`, `shared_agent_specificity_owner`) |
| `/api/v1/user/settings` | GET/PUT | Read or update persisted user preferences such as assistant name/theme; GET overlays the deploy-owned People & Access contract (`access_management_tier`, `product_edition`, `identity_mode`, `shared_agent_specificity_owner`), while PUT ignores/preserves those deploy-owned fields instead of persisting them |
| `/api/v1/groups` | GET/POST | List/create root-admin collaboration groups (DB-backed, tenant scoped) |
| `/api/v1/groups/{id}` | PUT | Update root-admin collaboration group |
| `/api/v1/groups/{id}/broadcast` | POST | Publish group coordination message to group + team NATS channels |
| `/api/v1/groups/monitor` | GET | Live group-bus monitor snapshot (published count, last group, last error) |
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
| `/api/v1/services/status` | GET | Aggregate health — NATS, PostgreSQL (with latency), Cognitive, Reactive, Comms, Group Bus monitor |

Memory/governance note:
- `AGENT_MEMORY` is the canonical team-shared execution lane
- governed context intake is separate from ordinary Soma memory and separate from team-shared memory
- reflection and synthesized learning should enter through `LearningCandidate` before durable promotion
| **Host Actions (Local Command V0)** | | |
| `/api/v1/host/status` | GET | Local host actuation status (OS/arch/workspace + effective local command allowlist) |
| `/api/v1/host/actions` | GET | List host actions available for invocation (currently `local-command`) |
| `/api/v1/host/actions/{id}/invoke` | POST | Invoke host action by ID. V0 supports allowlisted no-shell local commands only |
| **Mission Runs & Events (V7)** | | |
| `/api/v1/runs` | GET | List recent runs across all missions — status, timing, trigger source |
| `/api/v1/runs/{id}/events` | GET | Full event timeline for a run (MissionEventEnvelope records) |
| `/api/v1/runs/{id}/chain` | GET | Causal chain — parent run → event → trigger → child run traversal |
| **Intent (CE-1)** | | |
| `/api/v1/intent/confirm-action` | POST | Consume confirm token, execute mutation, return run_id |
| `/api/v1/intent/proof/{id}` | GET | Retrieve intent proof bundle by ID |
| `/api/v1/templates` | GET | List CE-1 orchestration templates or V8 AI Organization starters when `view=organization-starters` |
| `/api/v1/conversation-templates` | GET/POST | List/create DB-backed reusable Soma/Council/team ask templates |
| `/api/v1/conversation-templates/{id}` | GET/PATCH | Read/update a reusable conversation template |
| `/api/v1/conversation-templates/{id}/instantiate` | POST | Render a template with variables and return a non-executing ask package, team ask, or temporary-group draft |
| **AI Organizations (V8)** | | |
| `/api/v1/organizations` | GET | List created AI Organization summaries for the entry flow |
| `/api/v1/organizations` | POST | Create an AI Organization from template or empty start |
| `/api/v1/organizations/{id}/home` | GET | Load the minimal AI Organization context shell |
| `/api/v1/organizations/{id}/output-model-routing` | GET | Read admin-configurable output-model routing for the organization, including detected output-type bindings, locally installed models, and recommended self-hosted starting points |
| `/api/v1/organizations/{id}/output-model-routing` | PATCH | Update the organization default model or detected output-type model bindings for team and specialist output delivery |

## Provider Auth Notes

Provider inventory and auth contract:

| Provider type | Typical provider IDs | Auth expectation | Config contract |
| :--- | :--- | :--- | :--- |
| `openai_compatible` | `ollama`, `vllm`, `lmstudio`, custom local gateways | `Authorization: Bearer <api_key>` when the upstream checks keys. Local tools such as Ollama can ignore the placeholder key while still requiring the client field. | `endpoint`, `model_id`, optional `api_key` or `api_key_env` |
| `openai` | `production_gpt4` | `Authorization: Bearer $OPENAI_API_KEY` | `endpoint=https://api.openai.com/v1`, `api_key_env=OPENAI_API_KEY` |
| `anthropic` | `production_claude` | `x-api-key: $ANTHROPIC_API_KEY` plus `anthropic-version` | `api_key_env=ANTHROPIC_API_KEY`, optional custom endpoint |
| `google` | `production_gemini` | `x-goog-api-key: $GEMINI_API_KEY` | `api_key_env=GEMINI_API_KEY`, endpoint defaults to the Gemini `models` REST root |

Implementation notes:
- use `/api/v1/brains` and `/api/v1/cognitive/providers/{id}` to manage the provider inventory exposed in the product
- provider secrets are write-only in API payloads; `api_key` and `api_key_env` never come back in provider reads
- for local-model switching and profile routing, see [Local Dev Workflow](LOCAL_DEV_WORKFLOW.md) and [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

# API Reference
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

> Back to [README](../README.md) | See also: [Swarm Operations](SWARM_OPERATIONS.md) | [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

## API TOC

- [Endpoints](#endpoints)
- [Deployments](#deployments)
- [View Config](#view-config)
- [Clone And Apply Config](#clone-and-apply-config)
- [Provider Auth Notes](#provider-auth-notes)

## Endpoints

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| **Council Chat** | | |
| `/api/v1/council/{member}/chat` | POST | Chat with any council member via NATS request-reply. Returns `APIResponse<CTSEnvelope>` with trust score + provenance |
| `/api/v1/council/members` | GET | List all addressable council members from standing teams (admin-core, council-core) |
| **Chat & Cognitive** | | |
| `/api/v1/chat` | POST | Soma/Admin chat. Runtime-state and search-capability questions answer directly; freshness-oriented search prompts call configured Mycelis `web_search` before the NATS Admin-agent path. |
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
| `/api/v1/teams/detail` | GET | Legacy team detail (raw JSON, matches missions by name, declared model only). Use `/api/v1/deployments` for new work |
| `/api/v1/deployments` | GET | Deployments aggregate across standing teams, mission teams, groups, and planned departments; filters `kind`, `organization_id`, `state`. See [Deployments](#deployments) |
| `/api/v1/deployments/{kind}/{id}` | GET | One deployment record with members expanded; `404` when unknown, `503` when its source is unavailable |
| `/api/v1/config/{teams\|groups\|organizations\|templates}/{id}` | GET | Read-only, redacted config document (`?format=yaml\|json`) that `config_ref.path` resolves. See [View Config](#view-config) |
| `/api/v1/config/templates/{id}/clone` | POST | Governed clone of a template into an operator ("Custom") template; `202` preview with `confirm_token`, retry to store. See [Clone And Apply Config](#clone-and-apply-config) |
| `/api/v1/config/{groups\|templates}/{id}/apply` | POST | Governed apply of a YAML/JSON document; always a dry-run `202` preview first, then apply with the hash-bound `confirm_token`. See [Clone And Apply Config](#clone-and-apply-config) |
| `/api/swarm/teams` | POST | Create team via Soma |
| `/api/swarm/command` | POST | Send command to specific team |
| `/api/v1/swarm/broadcast` | POST | Fan out directive to ALL active teams |
| `/agents` | GET | List active agents with heartbeat status |
| **Telemetry & Trust** | | |
| `/api/v1/stream` | GET (SSE) | Real-time NATS signal stream |
| `/api/v1/telemetry/compute` | GET | Goroutines, heap, system memory, LLM tokens/sec |
| `/api/v1/trust/threshold` | GET/PUT | Read/write autonomy threshold |
| `/api/v1/sensors` | GET | Sensor library (static + dynamic) |
| `/api/v1/homepage` | GET | Return the sanitized deployer-editable homepage template from `core/config/homepage.yaml` or `MYCELIS_HOMEPAGE_CONFIG_PATH`, falling back to Soma orchestration defaults when missing or invalid |
| **Memory & Search** | | |
| `/api/v1/memory/search` | GET | Semantic vector search over durable memory with optional `team_id`/`agent_id`/`type`/`layer_filter` and retrieval-time arbitration (`arbitration`, `contradiction_halts`, `GOVERNED_DOCTRINE` > `PROJECT_MEMORY` > `AGENT_MEMORY` > `SOMA_MEMORY` > `REFLECTION_MEMORY`) |
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
| `/api/v1/artifacts/{id}/save` | POST | Persist cached image artifact to workspace folder (`saved-media` default) with verified SHA-256 disk readback and `OutputProofEnvelope` metadata |
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
| `/api/v1/user/me` | GET | Current user identity, including normalized principal metadata (`principal_type`, `auth_source`, `effective_role`, `break_glass`; supports `MYCELIS_API_KEY`, `MYCELIS_BREAK_GLASS_API_KEY`, `MYCELIS_OPERATOR_API_KEY`, `MYCELIS_VIEWER_API_KEY`) plus the deploy-owned People & Access contract surfaced read-only through `settings` (`access_management_tier`, `product_edition`, `identity_mode`, `shared_agent_specificity_owner`) |
| `/api/v1/user/settings` | GET/PUT | Read or update persisted user preferences (`viewer` is read-only `403` on PUT; `operator` updates general preferences; `admin`/`owner` required for access-posture keys); GET overlays the deploy-owned People & Access contract (`access_management_tier`, `product_edition`, `identity_mode`, `shared_agent_specificity_owner`), while PUT ignores/preserves those deploy-owned fields instead of persisting them |
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

Conversation-template instantiation is non-executing by default. Protected Soma interaction templates are a runtime classification layer used before action execution; they map phrase themes such as private service, private data, MCP/tool enablement, team manifestation, and recurring behavior to confirmation/proposal requirements. A protected phrase can review and explain directly, but must not launch teams, bind tools, use credentials/private data, or store recurring behavior without explicit confirmation and the relevant governed proposal path.
| **AI Organizations (V8)** | | |
| `/api/v1/organizations` | GET | List created AI Organization summaries for the entry flow |
| `/api/v1/organizations` | POST | Create an AI Organization from template or empty start |
| `/api/v1/organizations/{id}/home` | GET | Load the minimal AI Organization context shell |
| `/api/v1/organizations/{id}/output-model-routing` | GET | Read admin-configurable output-model routing for the organization, including detected output-type bindings, locally installed models, and recommended self-hosted starting points |
| `/api/v1/organizations/{id}/output-model-routing` | PATCH | Update the organization default model or detected output-type model bindings for team and specialist output delivery |

## Deployments

`GET /api/v1/deployments?kind=&organization_id=&state=` returns one normalized envelope, `{"ok": true, "data": {"deployments": [...], "sources": [...], "generated_at": "..."}}`. Core builds the list server-side from four sources:

| `kind` | Source | `id` | `scope.kind` |
| :--- | :--- | :--- | :--- |
| `standing_team` | Live Soma teams loaded from the boot bundle (mission runtime teams excluded) | team manifest id | `global` |
| `mission_team` | `teams` rows joined to `missions`, keyed by `teams.id`; live state is matched inside the same mission only | `teams.id` (UUID) | `mission` |
| `group` | `collaboration_groups`; returned only to root admins with `groups:read`, otherwise the `groups` source is `restricted` | group id | `global` |
| `planned_department` | Departments of persisted organizations; nothing runs until the operator approves a launch | `<organization id>:<department id>` | `organization` |

Record fields: `id`, `kind`, `name`, `purpose`, `what_runs`, `scope {kind, id, name}`, `origin` (`bundle`, `mission:<id>`, `operator`, `template:<id>`), `runtime_id`, `lifecycle_state` (`running`, `idle`, `planned`, `paused`, `archived`, `expired`, `error`), `since`, `expires_at`, `last_activity`, `members[]`, `tools[]` (`{name, source: internal|mcp|capability}`), `work_mode`, `approval_policy_ref`, `approval_posture`, `outputs_summary`, and `config_ref {kind, id, path}` (the `/api/v1/config/...` path that View config resolves).

Each member has `id`, `role`, `profile`, `provider`, `model`, `provider_origin` (`manifest`, `profile`, `fallback`, `declared`, `planned`), `provider_status`, `status`, `last_heartbeat_at`, and `tools`. For teams, `provider` and `model` come from the cognitive router's execution resolution (profile = manifest `model` or `chat`, plus any explicit manifest provider), not from the declared `model` field. Only provider and model ids are returned. Endpoints, API keys, and key environment names never are.

`outputs_summary` counts retained artifacts (`retained`, `by_status`, `last_output_at`). Standing teams count artifacts by member agent id, mission teams by team id, and groups sum their teams. `proof_status` is `not_recorded` until outputs carry proof envelopes. `outputs_summary` is omitted when counts are unknown and for planned departments.

Filters: `kind` and `state` must be known values (`400` otherwise); `kind` also skips the other sources. `organization_id` keeps only records whose `scope` is that organization (planned departments today).

Degradation: every source reports `{source, status, detail, count}` in `sources`, where `source` is `standing_teams`, `mission_teams`, `groups`, `organizations`, or `outputs` and `status` is `ok`, `degraded`, `unavailable`, or `restricted`. A failing or missing source never fails the response. With no database, standing teams and planned departments still return, and `mission_teams`, `groups`, and `outputs` report `unavailable`. Raw database errors are logged, not returned.

`GET /api/v1/deployments/{kind}/{id}` returns `{"ok": true, "data": {"deployment": {...}, "sources": [...]}}`. It returns `400` for an unknown kind, `404` when the record does not exist, and `503` when the record's source is unavailable or restricted. `/api/v1/teams/detail` stays for compatibility and is legacy.

## View Config

`GET /api/v1/config/{kind}/{id}?format=yaml|json` is the read-only View config surface (RC-004F). `{kind}` is `teams`, `groups`, `organizations`, or `templates`; it is the `config_ref.path` on every Deployments record. `format` defaults to `yaml`.

```json
{"ok": true, "data": {
  "kind": "template", "id": "starter-secretive", "format": "yaml",
  "document": "id: starter-secretive
provider_policy:
  openai:
    api_key: ${env:OPENAI_API_KEY}
    api_key_env: OPENAI_API_KEY
",
  "source": "file", "source_path": "templates/starters/secretive.yaml",
  "redactions": ["provider_policy.openai.api_key"],
  "generated_at": "2026-09-25T17:58:09Z"}}
```

| Kind | Resolves | `source` | Document |
| :--- | :--- | :--- | :--- |
| `teams` | A live Soma team by runtime id (standing teams, or a live mission team by `runtime_id`), else a mission team by `teams.id` | `runtime`, or `database` for a mission team that is not live | `kind: TeamManifest` with `metadata` (id, origin, mission, runtime id), `spec` (the manifest), and `members[].resolved {profile, provider, model, provider_origin, provider_status}` from the cognitive router |
| `groups` | `collaboration_groups` by id | `database` | `kind: CollaborationGroup` with `spec` in the shape group create/update accepts |
| `organizations` | The organization home | `database` (`runtime` for the in-memory store) | `kind: Organization`; `spec` is the home payload, including `template_id`, `delivery_defaults`, AI engine, and response contract settings |
| `templates` | A top-level bundle, then `starters/`, matched by bundle `id` | `file`, with `source_path` relative to the templates directory | The bundle file as loaded, every field kept |

`document` is text in the requested format; `redactions` lists the dotted paths (`a.b[0].c`) of every value that was hidden. Documents are built from declared fields and ids only: team members show provider and model ids, never provider blocks, endpoints, or keys.

Redaction is server-side and always runs before rendering:

- Keys are matched case- and separator-insensitively at any depth (`apiKey`, `API-KEY`, and `OPENAI_API_KEY` are the same). A key is secret when any segment is `password`, `passwd`, `secret(s)`, `credential(s)`, `authorization`, `bearer`, `apikey`, or `privatekey`; when its last segment is `key` or `token`; or when it ends in `api_key`, `private_key`, `client_secret`, `password`, `token`, and similar. `max_tokens`, `token_budget_profile`, and other look-alikes stay visible.
- A secret value becomes `${env:NAME}` when a sibling `<key>_env` names a valid environment variable, otherwise `<redacted>`. Numbers, maps, and lists under a secret key are hidden too. Booleans (`secret: true`), nulls, and empty strings carry no secret and stay as they are.
- References stay visible: `*_env`, `*_env_var`, and `*_ref` keys show the variable or reference name. A reference whose value is not a plain name or looks like a credential is redacted.
- Values under any key are scanned for credential shapes: `Bearer`/`Basic` tokens, known key prefixes (`sk-`, `ghp_`, `github_pat_`, `xoxb-`, `AKIA`, `AIza`, `hf_`, and others), JWTs, PEM private keys, URL passwords (`postgres://user:<redacted>@host`), sensitive URL query parameters, `NAME=value` assignments of secret names, and the argument after a sensitive CLI flag (`--api-key <redacted>`).

Access follows the underlying resource: `groups` requires the root admin role with `groups:read` (`403` otherwise), as `/api/v1/groups` does; `teams`, `organizations`, and `templates` need an authenticated operator. The interface keeps View config behind Advanced mode. Errors use the normalized envelope: `400` for an unknown kind or format, `401` without an identity, `404` for an unknown id, `503` when the runtime or database behind the kind is unavailable. Raw database errors are logged, not returned.

## Clone And Apply Config

Clone and apply are the governed write half of View config (RC-004G). Both use the governed-confirmation pattern that group mutations use: the first call never mutates; it validates, records an audit entry, persists an intent proof, and returns `202` with a single-use confirm token (15 minute TTL). The retry with `confirm_token` consumes the token inside the same transaction as the write, so a token is spent only when the change commits.

The token is bound to the document: the proof's `resolved_intent` is `<action>.<kind>.<id>@sha256:<hash>`, where the hash is SHA-256 over the canonical JSON of the parsed document (sorted keys, so the same document in YAML or JSON has the same hash). Confirming with a different document returns `409` and leaves the token unspent; request a new preview.

### Clone

`POST /api/v1/config/templates/{id}/clone` with `{"name": "Acme Studio", "new_id"?: "custom-acme-studio", "overrides"?: {...}}`. `{id}` is a top-level bundle, a file starter, or another operator template. The clone copies the source, deep-merges `overrides` (maps merge, lists and scalars replace; `overrides` cannot change `id`), and sets `id`, `name`, and `starter.label` to the clone's identity. `new_id` defaults to `custom-<slug of name>` and must match `^[a-z0-9][a-z0-9-]{1,62}$`. The clone is validated with the same bundle and starter rules as starter files, including `approval_policy_ref` resolution against `policy.yaml`, before any token is issued.

```json
{"ok": true, "data": {
  "requires_approval": true,
  "confirm_token": {"token": "7d7c...", "intent_proof_id": "3f1e...", "expires_at": "2026-09-25T18:30:00Z"},
  "intent_proof": {"id": "3f1e...", "resolved_intent": "template.clone.template.custom-acme-studio@sha256:9b1c...", "status": "pending"},
  "audit_event_id": "c0a8...",
  "preview": {"kind": "template", "id": "custom-acme-studio", "source_template_id": "starter-client-delivery-studio",
    "document_sha256": "9b1c...", "format": "yaml", "document": "id: custom-acme-studio\n...", "redactions": [],
    "valid": true, "validation_errors": []}}}
```

Retrying the same body with `"confirm_token": "<token>"` stores the clone in `operator_templates` and returns `201` with `template {id, name, source: "operator", source_template_id, document_sha256}`, `intent_proof_id`, `audit_event_id`, `run_id`, and `mission_event_id`. Clones appear in `GET /api/v1/templates?view=organization-starters` with `source: "operator"` (labelled "Custom") and `source_template_id`, under the same availability rules as file starters; an operator template whose id shadows a file template is listed as unavailable. `GET /api/v1/config/templates/{id}` shows a clone with `source: "operator"`.

### Apply

`POST /api/v1/config/{kind}/{id}/apply` with `{"document": "<yaml or json text>", "format": "yaml|json", "confirm_token"?: "<token>"}` (1 MiB limit).

| Kind | Applies | Document |
| :--- | :--- | :--- |
| `templates` | Operator templates only | The bundle, as View config shows it; `id` must equal the path id (added when missing); validated with the starter rules |
| `groups` | Collaboration groups | `kind: CollaborationGroup` as View config shows it; only `spec` is applied, `metadata` is informational except `metadata.id`; unknown `spec` fields are rejected |
| `teams`, `organizations` | Never (`409`) | Standing-team manifests change only through bundle files and restart; organizations use their settings endpoints |

Without `confirm_token` the endpoint is a dry run. A valid document returns `202` with `requires_approval`, `intent_proof`, `confirm_token`, `audit_event_id`, `blocked: false`, and `preview {kind, id, document_sha256, diff {changed, added, removed}, valid, validation_errors}`. The diff lists dotted paths only, never values. An invalid document returns `202` with `blocked: true`, the errors, and no token. With a valid token the endpoint applies exactly the previewed document and returns `200` with `applied: true`, the diff, `intent_proof_id`, `audit_event_id`, `run_id`, `mission_event_id`, and the reloaded `result`.

Write-back guard: a document is rejected with `400` and `data.write_back_violations` when any value contains the redaction marker `<redacted>`, when a secret field holds an `${env:...}` placeholder, or when it carries a literal secret (the View config redaction rules). Secrets belong in `*_env` or `*_ref` references.

Events: every preview writes an audit entry (`template.clone.previewed` or `config.apply.previewed`). Every confirmed clone or apply writes, in its transaction, an audit entry, a governance run keyed by the intent proof, a `template.cloned` or `config.applied` mission event (payload: kind, id, `document_sha256`, `intent_proof_id`, actor, `source_kind: web_api`), and `mission.completed`, then publishes a best-effort signal on the run's mission-events subject.

Access: the root admin role with `config:write`; groups also need `groups:write`. The interface keeps clone and apply behind Advanced mode. Errors: `400` bad body, format, document, write-back violation, invalid, expired, or used token, or a confirm of an invalid document; `401` without identity; `403` without the role or scopes; `404` unknown template, source, or group; `409` view-only kind, read-only file template, taken `new_id`, or a token issued for another document; `503` without a database, or when migration 041 is missing (the message names `uv run inv compose.migrate`).

## Provider Auth Notes

Provider inventory and auth contract:

| Provider type | Typical provider IDs | Auth expectation | Config contract |
| :--- | :--- | :--- | :--- |
| `openai_compatible` | `ollama`, `vllm`, `lmstudio`, custom local gateways | `Authorization: Bearer <api_key>` when the upstream checks keys. Local tools such as Ollama can ignore the placeholder key while still requiring the client field. | `endpoint`, `model_id`, optional `api_key` or `api_key_env` |
| `openai` | `production_gpt4` | `Authorization: Bearer $OPENAI_API_KEY` | `endpoint=https://api.openai.com/v1`, `api_key_env=OPENAI_API_KEY` |
| `anthropic` | `production_claude` | `x-api-key: $ANTHROPIC_API_KEY` plus `anthropic-version` | `api_key_env=ANTHROPIC_API_KEY`, optional custom endpoint |
| `google` | `production_gemini` | `x-goog-api-key: $GEMINI_API_KEY` | `api_key_env=GEMINI_API_KEY`, endpoint defaults to the Gemini `models` REST root |
| `media` (`openai_compatible`, `openai_hosted`, `stability_compatible`, `replicate_compatible`) | `media` | Optional `Authorization: Bearer $<auth_key_env>` | Normalized generation endpoint (`/images/generations` or `/v1/predictions`); remote/leaves-org boundaries enforce `require_approval` |

Implementation notes:
- use `/api/v1/brains` and `/api/v1/cognitive/providers/{id}` to manage the provider inventory exposed in the product
- provider secrets are write-only in API payloads; `api_key` and `api_key_env` never come back in provider reads
- for local-model switching and profile routing, see [Local Dev Workflow](LOCAL_DEV_WORKFLOW.md) and [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

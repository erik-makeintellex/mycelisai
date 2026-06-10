# API Reference
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

> Back to [README](../README.md) | See also: [Architecture Overview](architecture/OVERVIEW.md) | [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

## API TOC

- [Endpoints](#endpoints)
- [Directed Execution Payloads](#directed-execution-payloads)
- [Provider Auth Notes](#provider-auth-notes)

## Endpoints

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| **Interface Auth** | | |
| `/auth/local` | POST | Interface-local owner login. Accepts form fields `username` and `password`, verifies against deployment env, and writes a signed `mycelis_web_session` cookie. |
| `/auth/google/start` | GET | Starts Google Workspace OIDC login when `MYCELIS_AUTH_GOOGLE_*` env is configured. |
| `/auth/google/callback` | GET | Completes Google OIDC login, validates allowed Workspace domain, maps `MYCELIS_AUTH_ADMIN_EMAILS` to admin role, and writes a signed web session. |
| `/auth/session` | GET | Returns current Interface session posture: authenticated user, role, provider, and enabled login providers. |
| `/auth/logout` | POST | Clears the Interface web session and redirects to `/login`. |

Interface proxy routes sign the current web session into `X-Mycelis-Web-Identity` and `X-Mycelis-Web-Identity-Signature` when calling Core with the deployment API key. Core verifies the HMAC with `MYCELIS_WEB_IDENTITY_FORWARD_SECRET` or `MYCELIS_WEB_SESSION_SECRET` before using that principal for governance/audit context and `actor_identity` metadata. Invalid forwarded identity headers fail closed; missing headers retain the local API-key identity. When no browser session exists, document navigations to protected API or workspace-file URLs redirect to `/login?next=...`; programmatic fetches still receive structured `401` JSON with `{"ok":false,"error":"authentication_required"}` so components can handle the state without losing request context.
| **Council Chat** | | |
| `/api/v1/council/{member}/chat` | POST | Chat with any council member via NATS request-reply. Returns `APIResponse<CTSEnvelope>` with trust score + provenance |
| `/api/v1/council/members` | GET | List all addressable council members from standing teams (admin-core, council-core) |
| **Chat & Cognitive** | | |
| `/api/v1/chat` | POST | Soma/Admin chat. Runtime-state, Workspace V8 design-state, and search-capability questions answer directly; freshness-oriented search prompts call configured Mycelis `web_search` before the NATS Admin-agent path only when the latest request is not a governed mutation/team-creation/delegation prompt. Requests may include a focused `team_id`; blank/root Soma context resolves to `admin-core`, while focused-team Soma turns, proposal bus wiring, and team expressions preserve the selected team so team-scoped work does not leak back to root context. Proposal payloads include `bus_scope` and `nats_subjects`; explicit `create_team` proposals use only the requested team's command/status/result subjects, not a fallback admin-core bus. Explicit team requests can retain concrete output asks such as `write_file`, `generate_image`, and `save_cached_image` in the same proposal. Inferred generic `create_team` calls start lead-only with `initial_member_count=1`, a bounded `recommended_member_limit`, and explicit expansion guidance; explicit specialist-output asks may carry a bounded `agents` roster, media capability requirements, and first retained deliverable steps in the same ExecutionContract. Chat responses may include `execution_summary` so the UI can show intent, Soma understanding, execution shape, capability use, outputs, proof, audit/recovery state, and next step. |
| `/api/v1/cognitive/infer` | POST | Direct cognitive inference (profile-routed) |
| `/api/v1/cognitive/config` | GET | Read cognitive router configuration (providers, profiles, media) |
| `/api/v1/cognitive/matrix` | GET | Alias for cognitive config (matrix view) |
| `/api/v1/cognitive/status` | GET | Live health probe of enabled text engines and the configured local/private or hosted media provider; disabled text providers remain configurable but are not probed as health candidates |
| `/api/v1/cognitive/profiles` | PUT | Update profile→provider routing (persists to cognitive.yaml) |
| `/api/v1/cognitive/providers/{id}` | PUT | Configure provider (endpoint, model_id, api_key_env). Raw `api_key` values are rejected; use env/secret references. |
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
| `/api/v1/teams/{id}` | DELETE | Stop and remove one active runtime team. Use this for temporary-team cleanup so delivery tests stay bounded to 1-3 targeted teams instead of accumulating stale runtime teams. |
| `/api/v1/teams/detail` | GET | Team detail with agent rosters, delivery targets, and status |
| `/api/v1/teams/{id}/work` | GET/POST | Durable team-work items for Soma/Council/operator management. GET accepts `limit` and `include_archived=false` for active/review queues that should hide cleared work while preserving retained history. `create_team` items may only be `new` or `briefed`; delegated or deliverable work may be `queued`, `running`, `needs_operator`, `reviewing`, `output_ready`, `degraded`, `paused`, or `archived`. Confirmed `create_team` actions now persist a non-active team-shell item, while confirmed delegated work queues a real item and confirmed retained deliverables persist output-ready items with `run_id`, `intent_proof_id`, `contract_id`, `proof_id`, `audit_refs`, and retained `output_refs` where available. |
| `/api/v1/teams/{id}/work/ask` | POST | Submit one bounded runtime-team ask and persist the result as Active Work truth. Body accepts `message` or structured `ask`, optional `async`, `summary`, `actor_ref`, `timeout_seconds` capped at `60`, `expected_outputs`, `expected_proof`, `capability_requirements`, `governance_posture`, and bounded `payload`. The endpoint creates a durable delegated `TeamWorkItem`, records queued status and ask interaction, and returns the durable work state. With `async=true` it publishes and flushes a governed command envelope to `swarm.team.{id}.internal.command` containing `work_item_id`, team metadata, expected outputs/proof, and request context, then returns `202` with `accepted=true`, `dispatch_state=published`, and a durable running dispatch event/interaction with subject/channel and recovery-deadline hints. Team runtime correlation is FIFO for plain responses and preserves explicit `work_item_id` values when a response carries its own correlation, so overlapping direct asks do not collapse into a single latest-work slot. Correlated team status/result signals carry the same `work_item_id` back to Active Work so queued items can advance to `output_ready` or `degraded`. Result payloads may include `outputs[]` using the directed-execution output shape or normalized `output_refs[]`; retained paths are projected onto the original work item so Active Work can expose openable output/proof evidence. Direct ask dispatch, output-ready, and degraded result persistence writes status event, run-linked mission event where applicable, work-item update, and interaction in one transaction so proof/output refs do not commit without their matching interaction record. Durable `output_refs[].storage_ref` is a workspace file/folder path, not a viewer URL; project-package refs store the package folder and a relative `entrypoint` where possible, while browser clients derive open/reveal controls from those workspace refs. NATS offline, publish failure, or flush failure immediately records `degraded` with recovery options and returns `202`. Without `async`, the compatibility path waits for a bounded team reply and records either `output_ready` with a visible reply excerpt plus a retained `text_reply` output/proof ref, or `degraded` with `degradation_state=nats_offline|team_response_timeout|team_response_unreadable`. |
| `/api/v1/teams/{id}/work/{workItemId}/interactions` | GET/POST | Durable `TeamInteraction` records for a work item, including `source_kind`, `source_channel`, `actor_ref`, verb, summary, `payload_kind`, optional bounded payload/ref, approval ref, audit refs, and optional run/proof links. |
| `/api/v1/teams/{id}/work/{workItemId}/actions` | POST | Apply an audited operator control to a durable work item. Body uses `action=start_work|pause|resume|archive|steer|recover` with optional `summary`, `actor_ref`, `source_kind`, `source_channel`, `payload_kind`, `payload`, and `audit_refs`; `steer` and `recover` require either `summary` or `payload` so the audit trail records real guidance. The endpoint validates transitions, rejects `create_team` shell records, writes a `TeamStatusEvent`, writes a `TeamInteraction`, updates the work item state/last event, and returns the updated `TeamWorkItem`. Current transitions: `new|briefed|queued -> running` for `start_work`, `queued|running|needs_operator|reviewing|degraded -> paused`, `paused -> queued` for `resume`, non-archived delegated/deliverable work -> `archived`, no state change for `steer`, and `degraded|needs_operator -> queued` for `recover`. Browser UI labels `archive` as `Clear from review`; this is audited retention cleanup, not deletion. For `create_team` shell records, clients should route the operator through a Soma/team ask that creates a delegated or deliverable work item before showing lifecycle controls as executable. |
| `/api/v1/teams/{id}/work/{workItemId}/status-events` | GET | Durable `TeamStatusEvent` timeline for a work item. Returns operator-readable state history for queued/running/output-ready/degraded/control transitions; `limit` is bounded and capped at `100`. Run-linked status events also mirror into the persistent Event Spine as `mission_events.event_type=team_work.status` with normalized source, state, proof, blocked-by, and next-action metadata so team work can be reconstructed from the run timeline. |
| `/api/swarm/teams` | POST | Create team via Soma |
| `/api/swarm/command` | POST | Send command to specific team |
| `/api/v1/swarm/broadcast` | POST | Fan out directive to ALL active teams |
| `/agents` | GET | List active agents with heartbeat status |
| **Telemetry & Trust** | | |
| `/api/v1/stream` | GET (SSE) | Real-time NATS signal stream |
| `/api/v1/telemetry/compute` | GET | Goroutines, heap, system memory, LLM tokens/sec |
| `/api/v1/audit` | GET | Inspect normalized audit records. Confirmed governed actions include `actor_identity` when the request arrived through a signed Interface web session, so proof review can distinguish local API-key execution from local web or Google Workspace SSO execution. |
| `/api/v1/trust/threshold` | GET/PUT | Read/write autonomy threshold |
| `/api/v1/triggers` | GET/POST | List or create automation rules. Event rules use `trigger_kind=event` with `event_pattern`; schedule rules use `trigger_kind=schedule`, `event_pattern=scheduler.due`, `mode=propose`, `schedule_interval_seconds`, `next_run_at`, `proof_expectations`, and `recovery_behavior`. Scheduler ticks record proposed cadence outcomes, persist durable handoff refs, and advance next-run state only; they do not autonomously execute the target mission. |
| `/api/v1/triggers/{id}` | PUT/DELETE | Update or delete an automation rule. Schedule updates preserve the propose-only boundary and should keep proof/recovery copy operator-readable. |
| `/api/v1/triggers/{id}/toggle` | POST | Activate or pause an automation rule with body `{"is_active": true|false}`. |
| `/api/v1/triggers/{id}/history` | GET | Return recent rule execution records with `status=fired|skipped|proposed`; schedule-rule proposed rows may include `handoff_key`, `intent_proof_id`, `contract_id`, `proposal_status`, and `handoff_payload` with `autonomous_execution=false`. History never returns confirm tokens. |
| `/api/v1/triggers/{id}/history/{executionId}/approval` | POST | Transition one persisted schedule handoff from `proposal_status=awaiting_approval` to `approved`, `rejected`, or `cancelled`. The transition is limited to schedule handoff rows with `handoff_key`, no `run_id`, and `autonomous_execution=false`; it does not create a run, publish a team command, or consume a confirm token. Body accepts `{"status":"approved|rejected|cancelled"}` or equivalent `action` verbs such as `approve`, `reject`, or `cancel`. |
| `/api/v1/trust/execution-contracts` | GET | List durable `ExecutionContract` records for confirmed/proposed execution handshakes. Supports bounded `limit` plus `run_id`, `intent_proof_id`, and `status` filters; `limit` is capped at `100`. |
| `/api/v1/trust/execution-contracts/{id}` | GET | Read one durable `ExecutionContract` by UUID, including execution shape/status, validation source, evidence strength, proof quality, output/audit refs, degradation/recovery payloads, and latest proof artifact link. |
| `/api/v1/trust/proof-artifacts` | GET | List durable `ProofArtifact` records. Supports bounded `limit` plus `contract_id`, `run_id`, `intent_proof_id`, and `status` filters; `limit` is capped at `100`. |
| `/api/v1/trust/proof-artifacts/{id}` | GET | Read one durable `ProofArtifact` by UUID, including run/contract links, proof class, validation source, evidence strength, proof quality, output/audit refs, degradation/recovery payloads, and proof payload. |
| `/api/v1/sensors` | GET | Sensor library (static + dynamic) |
| `/api/v1/homepage` | GET | Return the sanitized deployer-editable branding/portal template from `core/config/homepage.yaml` or `MYCELIS_HOMEPAGE_CONFIG_PATH`, falling back to Soma orchestration defaults when missing or invalid. The root UI still requires login before any operator surface. |
| **Memory & Search** | | |
| `/api/v1/memory/search` | GET | Semantic vector search over durable memory, with optional team/agent/type scope filters across Soma-personal, team-shared, and governed memory lanes |
| `/api/v1/search/status` | GET | Current Mycelis Search provider posture for UI/Soma capability answers, including provider, configured/enabled flags, direct `web_search` support, token requirements, online-allowed/no-confirm disclosure posture, and blocker/next-action copy |
| `/api/v1/search` | POST | Governed Mycelis Search API. Native and Helm defaults use `local_sources` for retained Mycelis context without external tokens; Compose defaults to self-hosted `searxng` for public web search. Configured online search runs without a separate confirmation prompt when `MYCELIS_SEARCH_ONLINE_ALLOWED=true`, while responses disclose provider/path and treat external results as leads to verify. Response `data.metadata.semantic_fallback` can report bounded local text fallback such as `text_search` when semantic embeddings are unavailable. |
| **Runtime Capabilities** | | |
| `/api/v1/capabilities` | GET | Return the canonical runtime Capability Manifest snapshot as `APIResponse<Snapshot>`. When PostgreSQL is available, the response is read from durable `capability_manifests` rows refreshed from exchange capabilities, installed/available MCP, Mycelis Search status, internal tools, and host-command allowlist. Each manifest exposes `capability_id`, `manifest_version`, `health`, `last_probe_status`, `risk_class`, `approval_posture`, `allowed_roles`, `input_schema_ref`, `output_schema_ref`, `failure_posture`, `recovery_posture`, `audit_policy`, `secret_ref_policy`, `owner`, and `updated_at`; runtime team creation/delegation maps to the medium-risk `team_orchestration` capability. |
| `/api/v1/capabilities/{id}` | GET | Return one durable runtime capability manifest by ID (`404` when absent) with the same health/probe/risk/approval/schema/failure/recovery state fields as the list response. |
| `/api/v1/capabilities/refresh` | POST | Re-derive the current runtime Capability Manifest snapshot, upsert it into `capability_manifests`, remove stale manifest rows from the refreshed set, and return the refreshed `APIResponse<Snapshot>`. |
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
| `/api/v1/intent/confirm-action` | POST | Consume a chat proposal confirm token and replay the stored approved plan. Stored `planned_tool_calls` may carry `tool_ref` values such as `mcp:filesystem/read_text_file`; explicit MCP refs execute through the registered MCP executor instead of being shadowed by same-named internal tools, and retained outputs are returned as `mcp_tool_result` proof entries. |
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
| `/api/v1/artifacts/{id}/save` | POST | Persist cached image artifact to workspace folder (`saved-media` default); returned `file_path` can be used by the UI to open the mounted storage folder through workspace reveal |
| `/api/v1/workspace/files/view?path=...` | GET | Serve a bounded workspace file inline for retained chat outputs; paths are workspace-confined and HTML is sandboxed for generated game/code review |
| `/api/v1/workspace/files/reveal?path=...` | POST | Open the containing mounted workspace folder on the local Core host for a generated output or saved media artifact; `path=workspace` opens the governed workspace root; requires host invoke scope and keeps paths workspace-confined |
| **MCP Ingress** | | |
| `/api/v1/mcp/install` | POST | Raw MCP install endpoint — **disabled by Phase 0 security** (`403`), use library install |
| `/api/v1/mcp/servers` | GET | List installed MCP servers |
| `/api/v1/mcp/servers/{id}` | DELETE | Remove MCP server |
| `/api/v1/mcp/tools` | GET | List all MCP tools across servers |
| `/api/v1/mcp/activity` | GET | List recent persisted MCP activity from Managed Exchange, including server/tool/state visibility for operator review |
| `/api/v1/mcp/servers/{id}/tools/{tool}/call` | POST | Invoke a specific MCP tool. Canonical request body is `{"arguments": {...}}`; direct top-level argument objects such as `{"path":"workspace/file.md"}` are also accepted for operator scripts and compatibility. |
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
| `/api/v1/provision/draft` | POST | Draft a provisioning manifest; deployment still routes through governed Soma/team execution, not this draft endpoint |
| `/api/v1/registry/templates` | GET/POST | List/register connector templates |
| `/api/v1/teams/{id}/connectors` | POST | Install connector on team |
| `/api/v1/teams/{id}/wiring` | GET | Get team wiring graph |
| **Identity** | | |
| `/api/v1/user/me` | GET | Current user identity, including normalized principal metadata (`principal_type`, `auth_source`, `effective_role`, `break_glass`) plus the deploy-owned People & Access contract surfaced read-only through `settings` (`access_management_tier`, `product_edition`, `identity_mode`, `shared_agent_specificity_owner`) |
| `/api/v1/user/settings` | GET/PUT | Read or update persisted user preferences such as assistant name/theme; GET overlays the deploy-owned People & Access contract (`access_management_tier`, `product_edition`, `identity_mode`, `shared_agent_specificity_owner`), while PUT ignores/preserves those deploy-owned fields instead of persisting them |
| `/api/v1/groups` | GET/POST | List/create root-admin collaboration groups (DB-backed, tenant scoped). Group records include `workspace_folder`, a workspace-relative folder under `groups/` used for standing/user-defined/Soma-defined group outputs. If omitted on create, Core derives a stable folder from the first `team_id` or the group name plus ID and creates it under `MYCELIS_WORKSPACE`. |
| `/api/v1/groups/{id}` | PUT | Update root-admin collaboration group. Optional `workspace_folder` may move the group output lane to another workspace-confined `groups/...` path; omitted values preserve the existing folder. |
| `/api/v1/groups/{id}/broadcast` | POST | Publish group coordination message to group + team NATS channels |
| `/api/v1/groups/monitor` | GET | Live group-bus monitor snapshot (published count, last group, last error) |
| `/healthz` | GET | Health check |
| **Brains (Provider CRUD)** | | |
| `/api/v1/brains` | GET | List all providers with health status, location, data boundary |
| `/api/v1/brains` | POST | Add a new provider using env/secret references — hot-injects into running router, immediate probe. Raw `api_key` values are rejected. |
| `/api/v1/brains/{id}` | PUT | Update provider config using env/secret references. Raw `api_key` values are rejected. |
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
| `/api/v1/services/status` | GET | Aggregate health — NATS, PostgreSQL (with latency), Cognitive, Reactive, Scheduler, Comms, Group Bus monitor |
| `/api/v1/system/quick-checks/{id}` | GET | Focused system quick check; currently supports `scheduler` for Automation timing |
| `/api/v1/system/deployments/trust` | GET | Deployment trust snapshot for System -> Deployments: deployment/execution/workspace/artifact roots, current commit, image tag, chart version, deployment/proof lanes, endpoint and recovery posture, and runtime health summary. `workspace_root` reports `MYCELIS_BACKEND_WORKSPACE_ROOT` or `MYCELIS_WORKSPACE`; `artifact_root` reports `MYCELIS_ARTIFACT_ROOT`, `MYCELIS_ARTIFACTS_ROOT`, or legacy `DATA_DIR`. Unknown or unavailable values are returned as `unknown`; secrets are never exposed. |

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
| `/api/v1/intent/confirm-action` | POST | Consume confirm token, execute mutation, return `run_id` plus `execution_summary` proof for verified guided execution. Successful responses include `data.team_work_refs[]` when confirmed work creates durable team visibility, with `work_item_id`, `team_id`, `state`, `run_id`, and `output_refs` when retained output refs are already available. Failed approved execution responses also return `data.execution_summary` with failed run/proof/audit metadata and `audit_recovery.degradation` so the UI can show what failed, what remains trusted, and what must be retried. Confirmed tool calls are logged to `/api/v1/runs/{id}/conversation`; approved `create_team` calls mirror a collaboration-group record with a dedicated `workspace_folder` and a durable `create_team` work item in `new` state so team creation is visible without implying active work. Confirmed delegated tasks create queued durable work; confirmed retained deliverables, including generated media save steps, create output-ready work items with status events, interactions, run/contract/proof links, workspace-path output refs, and run-linked `team_work.status` mission events. Soma-owned team project packages default under `groups/{team_id}/generated/...`, while Soma-owned team media saves default under `groups/{team_id}/media`. Retained outputs identify created teams as `kind=team`, approved file writes as retained `kind=file` or `kind=code`, and saved media as retained artifact/media refs where available; workspace-readable file and saved-media outputs include an `href` to the sandboxed workspace viewer so browser clients can preview/open them directly while durable team output refs keep workspace paths/folders for local reveal and later focused-team lookup. |
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
| `/api/v1/organizations/{id}/workspace/actions` | POST | Return Team Lead guidance, execution contract, and optional `execution_summary` for operator review without attaching a `run_id` until real execution exists. Native-team and temporary-workflow contracts now include `initial_member_count`, `recommended_team_member_limit`/`recommended_member_limit`, `expansion_policy`, and `temporary_addition_guidance` so the UI can show lead-only creation plus deliberate operator/team-lead expansion rules. |
| `/api/v1/organizations/{id}/output-model-routing` | GET | Read admin-configurable output-model routing for the organization, including detected output-type bindings, locally installed models, and recommended self-hosted starting points |
| `/api/v1/organizations/{id}/output-model-routing` | PATCH | Update the organization default model or detected output-type model bindings for team and specialist output delivery |

## Directed Execution Payloads

`execution_summary` is the additive V8.2 directed-execution contract for Soma-facing runtime responses. It is optional for compatibility, but meaningful Soma actions should populate it as they move into the directed-execution model.

The object can include:
- `intent`: original and resolved request classification
- `understanding`: Soma's concise interpretation and assumptions
- `execution`: shape, status, and summary such as `direct_soma`, `guided_proposal`, `tool_assisted_work`, or `team_execution`
- `capability_use`: governed tools, teams, MCP capabilities, automations, or plugins used
- `capability_use.reason`: operator-facing provenance detail when available, such as the active `web_search` source boundary
- `outputs`: answer, proposal, artifact, media, tool result, retained team, retained file/code output references, or `project_package` deliverables. Browser clients may preview image/audio/video output URLs inline when the MIME type or file extension is recognizable, while durable file paths stay available through workspace reveal. Confirmed execution outputs may also include `proof_artifact_id`, `open_url`, and a per-output `proof` envelope with run/contract linkage, path-boundary status, readback status, checksum metadata, storage ref, and recovery hint.
- `outputs[].kind=project_package`: retained complex deliverable package for generated projects such as playable browser games. Package outputs may include `entrypoint`, `folder`, `files`, and `validation`; generated file packages also include support files such as `README.md`, `PROOF.md`, and `project-package.json` when requested or inferred. The UI should expose the entrypoint as the primary open action, the folder through workspace reveal, the package folder through Resources deep-linking, the file list as source context, and validation as proof context. Confirmed `write_file` and `store_artifact` executions can both produce this output shape; `store_artifact` should use artifact `type="project_package"` or metadata `package_kind/output_kind/kind="project_package"` with `entrypoint/package_entrypoint`, `folder/package_folder`, `files/package_files`, and `validation/validation_summary` metadata.
- `proof`: `run_id`, `audit_event_id`, `intent_proof_id`, and verification state
- `audit_recovery`: approval status, recovery state, blocker, retry posture, and optional `degradation`
- `audit_recovery.degradation`: operational degradation metadata covering `code`, `what_failed`, `trusted_state`, `invalidated_proof`, `safe_continuation`, and `requires_attention`
- `next_step`: suggested continuation or proof/review link

Current producers:
- `/api/v1/chat` direct answers and guided proposals
- `/api/v1/intent/confirm-action` verified proposal execution results, failed approved-execution recovery summaries, retained tool outputs, run-conversation turns, and durable team-work records for confirmed `create_team`, delegation, and retained deliverable tool calls
- `/api/v1/organizations/{id}/workspace/actions` Team Lead guidance responses with proof explicitly left unrun until execution exists
- `/api/v1/groups/{id}/broadcast` accepted group broadcasts with `audit_event_id` proof and no fabricated `run_id`

## Provider Auth Notes

Provider inventory and auth contract:

| Provider type | Typical provider IDs | Auth expectation | Config contract |
| :--- | :--- | :--- | :--- |
| `openai_compatible` | `ollama`, `vllm`, `lmstudio`, custom local gateways | `Authorization: Bearer <resolved secret>` when the upstream checks keys. Local tools such as Ollama can ignore the placeholder key while still requiring the client field. | `endpoint`, `model_id`, optional `api_key_env`; future `secret_ref` |
| `openai` | `production_gpt4` | `Authorization: Bearer $OPENAI_API_KEY` | `endpoint=https://api.openai.com/v1`, `api_key_env=OPENAI_API_KEY` |
| `anthropic` | `production_claude` | `x-api-key: $ANTHROPIC_API_KEY` plus `anthropic-version` | `api_key_env=ANTHROPIC_API_KEY`, optional custom endpoint |
| `google` | `production_gemini` | `x-goog-api-key: $GEMINI_API_KEY` | `api_key_env=GEMINI_API_KEY`, endpoint defaults to the Gemini `models` REST root |

Implementation notes:
- use `/api/v1/brains` and `/api/v1/cognitive/providers/{id}` to manage the provider inventory exposed in the product
- provider secrets resolve through env/secret references; raw `api_key` update payloads are rejected by provider-management APIs
- provider reads never return raw secret values; safe configuration responses may expose configured/readiness posture only
- the canonical secret boundary is [V8 Secret Storage And Credential Boundary](architecture-library/V8_SECRET_STORAGE_AND_CREDENTIAL_BOUNDARY.md)
- for local-model switching and profile routing, see [Local Dev Workflow](LOCAL_DEV_WORKFLOW.md) and [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

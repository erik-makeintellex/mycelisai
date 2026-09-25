# MYC-RC-004 Delivery Experience Plan
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) | [MYC-RC-003 Proof Envelope Delivery Plan](MYC_RC_003_PROOF_ENVELOPE_DELIVERY_PLAN.md)

> Status: `COMPLETE` (Windows proof lane) - accepted from committed state `bb6ea007`: full Chromium `93` passed / `14` skipped / `0` failed; Vitest `123` files; tsc clean; `go vet` + `go test ./... -count=1 -p 1` green; `pytest` `327` passed; line gate and `git diff --check` clean. Live walkthrough against a running Core waits on the Environment lane.
> Date: 2026-09-25
> Purpose: make strong delivery the zero-config default for simple operators, make delivery config readable and governable for advanced operators, and make every team deployment explainable from one page.

## Goal

`MYC-RC-004` turns delivery from hidden runtime shape into a visible, company-shaped product surface:

1. A simple operator picks a starter that matches their kind of company and gets strong delivery with zero config: an approval policy, retained outputs with proof, and a compact team shape.
2. An advanced operator can read the resolved config behind any template, team, or group, clone a template, and apply a config file through the same governed confirm path the product already uses.
3. One Deployments page answers "what is running, where, with which models and tools, under which approval policy, since when, and what did it produce".
4. The words organization, template, team, group, deployment, and knowledge each mean one thing.

## Solution Architecture

### Verified Current State (2026-09-25)

- One bundle exists: `core/config/templates/v8-migration-standing-team-bridge.yaml` (570 lines, legacy-capped). `TemplateLoader.LoadBundles` reads only the top level of `config/templates/` (no recursion); boot fails closed when more than one top-level bundle exists without `MYCELIS_BOOTSTRAP_TEMPLATE_ID`, which only `.env.compose` sets.
- `GET /api/v1/templates?view=organization-starters` (`templates_handlers.go` -> `organization_templates.go`) already feeds the create entry's template list from that same directory. At plan time, the "Product delivery organization" / "Operations organization" chips in `CreateOrganizationEntry.tsx` only prefilled name and purpose; RC-004C replaced them with data-driven starter cards that select a template.
- The bundle schema (`TemplateBundle`, `swarm.TeamManifest`, `protocol.AgentManifest`) has no work mode, approval policy, or output-retention field. `TeamManifest` carries `provider`, `members[].provider/model/tools`, `inputs`, `deliveries`, `schedule`.
- Organizations persist in the `organizations` table (migration 040, RC-004-0); `OrganizationStore` falls back to memory only when Core has no database. Creating one from a template copies a summary (departments, counts) plus starter delivery defaults and instantiates no runtime teams. At plan time the store was in-memory and organizations were lost on restart.
- `GET /api/v1/teams/detail` (`teams_detail.go`) already merges standing and mission teams (`type: standing|mission`), but it matches missions by team **name**, returns raw JSON instead of the API envelope, reports `agent.Model` rather than the resolved provider, and omits groups.
- Groups (`collaboration_groups`, migration 034) carry `work_mode` (`read_only|propose_only|execute_with_approval|execute_bounded|resume_continuity`), `allowed_capabilities`, a free-text `approval_policy_ref`, `status`, and `expiry`. High-impact group mutations already use the governed pattern in `groups_model.go`: first call returns `202` with `requires_approval`, `intent_proof`, `confirm_token`; the retry with `confirm_token` applies.
- `core/config/policy.yaml` defines named policy groups that target `team:<id>`; this is the natural target for `approval_policy_ref`.
- `/teams` is not in the rail; the rail's Advanced section has Groups, Activity, Resources, Memory, System. Resources shows a "Deployment Context" tab backed by `/api/v1/memory/deployment-context`, which is a knowledge store.
- `cognitive.yaml` mixes literal `api_key` values with `api_key_env` references; any config view must redact by key, not trust the file.

### Personas (Company Types)

| Company type | Simple path (zero config) | What Soma says after creation | Advanced path |
| --- | --- | --- | --- |
| Agency / services studio | Starter "Client delivery studio": one Delivery Lead plus Writer and Reviewer; `propose_only` by default; every client-facing output is retained with proof; approval before anything leaves the org | "Your studio is ready. I'll draft client work, a Reviewer checks it, and nothing is sent until you approve." | View the bundle; clone it to add a client-specific specialist or a brand-voice prompt; apply a group file per client engagement with an expiry |
| SaaS product team | Starter "Product delivery team": Product Lead, Builder, Verifier; `execute_with_approval`; code/package outputs retained with proof envelopes (RC-003); compact three-member shape | "Your product team is ready. I'll plan the work, the Builder proposes changes, and the Verifier attaches proof before you approve." | View resolved provider routing per member; clone to route Builder to a coding model; apply a group file that binds `mcp.filesystem` and a repo workspace |
| Operations-heavy business | Starter "Operations desk": Ops Lead, Runbook Operator, Monitor; `execute_bounded` for read-only checks, approval for anything that changes state; scheduled status summaries retained | "Your operations desk is ready. I'll watch what you connect, summarize daily, and ask before changing anything." | View schedules and bound tools; clone to add a schedule or a monitor feed; apply group files for on-call rotations with expiry |
| Regulated enterprise | Starter "Governed enterprise program": Program Lead, Analyst, Compliance Reviewer; `propose_only`; every mutation requires approval and every output needs a verified proof; local providers only | "Your governed program is ready. Everything I produce is a proposal with proof, and only local AI engines are used." | View the provider allow-list and approval policy; clone with stricter tool allow-lists; apply config only through the governed confirm path with audit events |

Rules shared by all four: the default team never exceeds three members plus Soma; outputs are retained artifacts with a proof envelope where RC-003 supports one and an explicit `not_recorded` otherwise; no starter needs a secret to run on the default local provider.

### Starter Template Catalog

Starters live in `core/config/templates/starters/` so boot selection stays single-bundle (the loader does not recurse) and host runs keep working without `MYCELIS_BOOTSTRAP_TEMPLATE_ID`. `organization_templates.go` loads the top level plus `starters/`; boot loads only the top level. Each file stays under the 300-line gate and is mirrored in `charts/mycelis-core/config/templates/starters/` and the chart configmap.

| File | Kernel / council | Provider policy | Teams (members) | Work mode | Approval policy ref | Delivery outputs |
| --- | --- | --- | --- | --- | --- | --- |
| `starter-client-delivery-studio.yaml` | `kernel.mode: guided`, `council.mode: compact` | `provider: ollama` | `studio-delivery` (delivery-lead, writer, reviewer) | `propose_only` | `delivery-approve-external` | client brief, draft deliverable, review notes |
| `starter-product-delivery-team.yaml` | `guided` / `compact` | `provider: ollama`, `role_providers.builder` overridable | `product-delivery` (product-lead, builder, verifier) | `execute_with_approval` | `delivery-approve-changes` | plan, change package, verification proof |
| `starter-operations-desk.yaml` | `guided` / `compact` | `provider: ollama` | `ops-desk` (ops-lead, runbook-operator, monitor) with a `schedule` for the daily summary | `execute_bounded` | `delivery-approve-state-change` | daily status, incident note, runbook change proposal |
| `starter-governed-enterprise-program.yaml` | `governed` / `review-board` | `allowed_providers: [ollama]`, no hosted fallback | `governed-program` (program-lead, analyst, compliance-reviewer) | `propose_only` | `delivery-approve-all` | decision memo, evidence pack, compliance review |

Schema additions (optional, so the bridge bundle still validates):

```yaml
starter:                      # beginner-facing labels only; never shapes runtime
  label: Product delivery team
  persona: saas_product_team  # studio | saas_product_team | operations | regulated_enterprise
  order: 20
  sample_name: Northstar Labs
  sample_purpose: Plan priorities, propose changes, and ship them with proof.
  soma_intro: "Your product team is ready. ..."
delivery_defaults:            # runtime-shaping; inherited by groups Soma creates for the org
  work_mode: execute_with_approval
  approval_policy_ref: delivery-approve-changes
  outputs: { retain: true, proof_required: true }
  max_team_members: 3
```

Validation: `work_mode` must be a known group work mode (move the set from `groups_model.go` into `core/pkg/protocol` so `bootstrap` and `server` share it); `approval_policy_ref` must name a group in `policy.yaml`; `max_team_members` must be 1..5 and no team in the bundle may exceed it. Four new `policy.yaml` groups (`delivery-approve-external`, `delivery-approve-changes`, `delivery-approve-state-change`, `delivery-approve-all`) express the postures with `REQUIRE_APPROVAL` rules.

Data-driven chips: `OrganizationTemplateSummary` gains `starter` (label, persona, order, sample name/purpose) and `delivery_summary` (work mode label, approval posture sentence, output proof sentence, team shape sentence). `CreateOrganizationEntry.tsx` renders chips from the starter response sorted by `order`; clicking a chip selects template mode for that template and prefills the sample name and purpose. The hardcoded chips are deleted. When starters fail to load, the existing "start empty" fallback stays.

Organization delivery defaults: `OrganizationHomePayload` stores `template_id` and `delivery_defaults`; groups Soma creates for that organization default to its `work_mode` and `approval_policy_ref`, and Soma's first reply uses `starter.soma_intro`.

### Deployments View Contract

New aggregate read API in Go (not client-side composition), because it must key records by id instead of name, normalize one envelope, resolve providers, and redact server-side:

- `GET /api/v1/deployments?kind=&organization_id=&state=` -> `APIResponse{data: DeploymentRecord[]}`
- `GET /api/v1/deployments/{kind}/{id}` -> one `DeploymentRecord` with members expanded

Owner files: `core/internal/server/deployments.go` (handler), `deployments_model.go` (types), `deployments_sources.go` (source adapters). Sources: `Soma.ListTeams()` and `state.GlobalRegistry` (standing teams and heartbeats, as in `teams_detail.go`); `teams`/`missions` rows by `teams.id` (as in `mission_detail.go`); `listGroupsDB` (`groups_store.go`); `groups_outputs.go` and `artifacts` for retained outputs and proof; `runs.Manager.ListRecentRuns` and mission events for last activity; `swarm.ProviderPolicy` resolution plus `cognitive.Router` for effective provider/model; `mcp` service for tool binding state; `OrganizationStore` for template departments not yet running. `/api/v1/teams/detail` stays for compatibility and is marked legacy.

`DeploymentRecord` fields:

| Field | Meaning |
| --- | --- |
| `kind` | `standing_team`, `mission_team`, `group`, `planned_department` |
| `id`, `name`, `purpose` | stable id, display name, one-line purpose |
| `what_runs` | plain sentence, for example "3 agents answering product requests" |
| `scope` | organization id/name, tenant, and bus scope (`swarm.team.{id}` family), never raw subjects in simple view |
| `origin` | `template:<id>`, `bundle:<id>`, `mission:<id>`, `operator`, or `api` |
| `members[]` | id, role, effective provider, effective model, provider origin (`team`, `agent`, `policy`, `default`), status, last heartbeat |
| `tools[]` | internal and MCP tools with binding state (`ready`, `missing_server`, `blocked_by_policy`) |
| `work_mode`, `approval_policy_ref` | groups: stored values; teams: resolved from organization `delivery_defaults` or `unset` |
| `state` | `running`, `idle`, `paused`, `expired`, `archived`, `offline`, `planned` |
| `since`, `expires_at`, `last_activity_at` | creation or boot time, group expiry, latest run/event time |
| `outputs[]` | latest retained outputs with `proof_status` (`verified`, `degraded`, `missing`, `blocked`, `not_recorded`) and `open_url` |
| `config_ref` | `{kind, id}` for "View config" |

Page: `/deployments` (route), with `/teams` redirecting to it. Simple columns: Name, Kind, What runs, State, Last activity, Outputs (count + worst proof status). Row detail is list-plus-detail (not a drawer of copy): members with model/provider, tools, work mode and approval policy in one sentence, outputs with proof chips, and "View config". Advanced mode adds origin, scope, provider origin, and the clone/apply actions. `TeamsIntroPanel` copy and the member-template panel move to Docs or behind Advanced. Load failures use `LoadFailureState`, never "nothing here yet".

### View Config, Clone, And Apply

Read-only (any operator, rendered in a monospace viewer with copy/download):

- `GET /api/v1/config/templates/{id}?format=yaml|json` - bundle as loaded, plus `x_origin: file:<path>|operator`
- `GET /api/v1/config/teams/{id}?format=` - the effective `TeamManifest` with resolved provider per member and `x_origin` per field group
- `GET /api/v1/config/groups/{id}?format=` - the group projected to `apiVersion: mycelis/v1alpha1`, `kind: CollaborationGroup`, `spec:` (the same shape apply accepts)
- `GET /api/v1/config/organizations/{id}?format=` - template id, delivery defaults, AI engine and response-contract settings

Redaction is server-side and key-based (`api_key`, `token`, `secret`, `password`, `authorization`, `*_key`): values become `${env:NAME}` when an `*_env` reference exists, otherwise `<redacted>`. A Go test asserts no known secret value from `.env` or `cognitive.yaml` literal keys appears in any response. System prompts are shown (not secrets) but collapsed.

Governed mutations reuse the group pattern, generalized into `requireGovernedConfirmation(w, intent, scope)` in `governed_confirm.go`:

- `POST /api/v1/config/templates/{id}/clone` body `{new_id, name, overrides}` -> `202 {requires_approval, intent_proof, confirm_token, preview}`; retry with `confirm_token` stores the clone.
- `POST /api/v1/config/{kind}/{id}/apply` body `{document, format, confirm_token?}` for `templates` (operator clones only) and `groups`; file templates, teams, and organizations return `409`. Without a token it always returns `202` with a diff/validation `preview`, intent proof, and a single-use 15-minute confirm token bound to the document's canonical SHA-256 (`blocked: true` and no token when invalid); the retry applies exactly that document in one transaction. Documents containing `<redacted>`, `${env:...}` in secret fields, or literal secrets are rejected with `write_back_violations`. As built in `9efb20de`; mutations require the root admin role with `config:write` (plus `groups:write` for groups).

Operator templates persist in a new `operator_templates` table (migration 041: id, name, source_template_id, document JSONB, created_by, audit ids) because the Helm config directory is read-only; the starter list merges file bundles and operator templates, with operator entries labelled "Custom". Apply emits `config.applied` / `template.cloned` audit and mission events. Standing-team manifests are view-only; they change only through bundle files and restart.

### Naming And Glossary

- Resources tab "Deployment Context" becomes "Knowledge" (label, loading/empty strings, `DeploymentContextPanel` headings). The API path `/api/v1/memory/deployment-context` and the Go package keep their names for compatibility.
- New user doc `docs/user/glossary.md` (in `docsManifest.ts`, linked from `core-concepts.md` and `teams.md`): **Organization** (an instantiated AI Organization you work in), **Template** (a reusable blueprint; editing it never rewrites existing organizations), **Team** (a named set of agents with a lead; standing teams load at boot, mission teams are created for one mission), **Group** (a governed work scope with a work mode, approval policy, and optional expiry that can include teams and people), **Deployment** (anything currently defined to run: a team or group, as listed on Deployments), **Knowledge** (documents and context Soma can cite; formerly "Deployment Context").

### Page-By-Page Review

| Page | Simple user sees / does | Advanced user sees / does | RC-004 change |
| --- | --- | --- | --- |
| `/` landing | what Mycelis does, one "Start with Soma" action | same | one line naming the four starters; no config |
| `/dashboard` Soma | asks Soma; proposals, approvals, retained outputs | same plus Advanced panels | Soma intro from `starter.soma_intro`; management workbench links "Deployments" |
| Organization setup (on `/dashboard`) | picks a company chip, sees a 3-line delivery summary, creates | "View template config" link | chips data-driven; delivery summary; hardcoded chips removed |
| `/organizations/[id]` | "What's running" (top 5 deployment rows) and outputs with proof | origin, delivery defaults, View config | compact deployments strip fed by `?organization_id=` |
| `/teams` -> `/deployments` | one list: what runs, state, last activity, outputs | origin, scope, provider origin, clone/apply | new page; `/teams` redirects; intro copy removed |
| `/groups` | groups with work mode and approval sentence | edit, broadcast, View config / apply group file | "View config"; link to its Deployments row |
| `/runs` | runs with outcome and outputs | events, chain, conversation | each run links its deployment record |
| `/resources` | not in simple nav | Brains, Tools, Knowledge, Catalogue | "Deployment Context" -> "Knowledge" |
| `/settings` | theme, identity, Advanced toggle | providers and tools | none beyond the Advanced toggle gating clone/apply |
| `/docs` | Glossary, Teams, Core Concepts | architecture docs | glossary added; "how delivery is configured" page for advanced users |

## Delivery Team

| Team | Model | Status | Owner Scope | Output |
| --- | --- | --- | --- | --- |
| Delivery Contract | Opus | `ACTIVE` | this plan, schema and API contracts, redaction and governance review | accepted plan, review sign-off per slice |
| Backend Delivery | Opus then Sonnet | `NEXT` | bundle schema, starters loader, deployments aggregate, config view/apply, migration 041 | Go handlers, tests, `docs/API_REFERENCE.md` |
| Interface Delivery | Sonnet | `NEXT` | chips, Deployments page, config viewer, apply dialog, Knowledge rename | components, Vitest, mocked Chromium specs |
| Docs / Mechanical | Haiku | `NEXT` | glossary, manifest, label renames, Helm mirror, index/state sync | docs and doc-test updates |

## Delivery Sequence

Every slice runs `uv run inv quality.max-lines --limit 300` and `uv run pytest tests/test_docs_links.py tests/test_documentation_layout_contract.py -q`, reviews `README.md`, `.state/V8_DEV_STATE.md`, `docs/API_REFERENCE.md`, and `interface/lib/docsManifest.ts`, and records evidence in state.

1. `COMPLETE` RC-004 design (Opus): this document.
1a. `COMPLETE` (`f4f73548`) RC-004-0 Persist organizations (Opus): `organizations` table (migration 040), Postgres repository with in-memory fallback, 503 naming `uv run inv compose.migrate` when the table is missing.
2. `COMPLETE` (`5d0fc034`) RC-004A Starter schema and loader (Opus). Files: `core/internal/bootstrap/template_bundle.go`, new `template_starter.go`, `core/pkg/protocol` work-mode constants, `core/internal/server/groups_model.go`, `organization_templates.go`, `organizations.go` (delivery defaults on the home payload). Proof: `uv run inv core.test`. Accept: starters load from `starters/` without affecting boot selection; invalid `work_mode`/`approval_policy_ref` fails validation; the bridge bundle still validates.
3. `COMPLETE` (`84fba32a`; regulated starter corrected to `execute_with_approval`) RC-004B Four starter bundles (Sonnet). Files: `core/config/templates/starters/*.yaml`, `core/config/policy.yaml`, `charts/mycelis-core/config/templates/starters/*`, `charts/mycelis-core/templates/configmap-config.yaml`. Proof: `uv run inv core.test` (a table test loads each starter and asserts <= 3 members plus Soma, a known approval policy, `outputs.proof_required: true`). Accept: `GET /api/v1/templates?view=organization-starters` returns four starters plus the bridge.
4. `COMPLETE` (`54dd19bb`; `organization-starters.spec.ts` landed in `289a9afa`, organization-home intro in RC-004C-2) RC-004C Data-driven chips and delivery summary (Sonnet). Files: `CreateOrganizationEntry.tsx`, `useCreateOrganizationEntryState.ts`, `createOrganizationEntryUi.tsx`, `interface/e2e/support/organization-entry.ts`, a new `e2e/specs/organization-starters.spec.ts`. Proof: `uv run inv interface.test`, `uv run inv interface.typecheck`, `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/organization-starters.spec.ts`. Accept: no hardcoded chip remains; chip click selects the template; starter failure still offers Start empty.
5. `COMPLETE` (`69c689e4`; list payload is `data.deployments` + `data.sources`, fields `lifecycle_state` / `last_activity` / `outputs_summary`) RC-004D Deployments aggregate API (Opus). Files: `core/internal/server/deployments.go`, `deployments_model.go`, `deployments_sources.go`, tests, `admin_routes.go`, `docs/API_REFERENCE.md`. Proof: `uv run inv core.test`. Accept: one envelope across all four kinds; mission teams keyed by id; effective provider per member; DB-down returns standing teams plus a `degraded` source note, not an error page.
6. `COMPLETE` (`3fe61883`; browser proof `deployments.spec.ts` in `289a9afa`) RC-004E Deployments page (Sonnet). Files: `interface/app/(app)/deployments/page.tsx`, `interface/components/deployments/*`, `interface/app/(app)/teams/page.tsx` (redirect only), `ZoneA_Rail.tsx`, organization "What's running" strip. Proof: `uv run inv interface.test`, `uv run inv interface.typecheck`, `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/deployments.spec.ts`. Accept: a mocked standing team, mission team, group, and planned department each render with state, members/models, work mode, and proof status; Core-down shows `LoadFailureState`.
7. `COMPLETE` (backend `563cd0e1`, UI `b5d79582`) RC-004F View config (Sonnet build, Opus redaction review). Files: `core/internal/server/config_view.go`, `config_redact.go`, tests; `interface/components/deployments/ConfigViewer.tsx`. Proof: `uv run inv core.test`, `uv run inv interface.test`, focused `deployments.spec.ts`. Accept: YAML and JSON for all four config kinds; redaction test proves no secret value leaks.
8. `COMPLETE` (backend `9efb20de`, UI `88af69c0`, starter-card entry point `bb6ea007`) RC-004G Clone and apply (Opus backend, Sonnet UI). Files: `governed_confirm.go`, `config_apply.go`, `core/migrations/041_operator_templates.{up,down}.sql`, `ApplyConfigDialog.tsx`. Proof: `uv run inv core.test`, `uv run inv interface.test`, `uv run inv interface.typecheck`, `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/config-apply.spec.ts`. Accept: dry-run preview always precedes apply; no mutation without a valid `confirm_token`; audit and mission events recorded; clones appear as "Custom" starters.
9. `COMPLETE` (`637adfd9`; remaining user-doc updates owned by the Docs sync and RC-004E) RC-004H Naming and glossary (Haiku). Files: `interface/app/(app)/resources/page.tsx`, `DeploymentContextPanel.tsx` strings, `docs/user/glossary.md`, `docs/user/core-concepts.md`, `docs/user/teams.md`, `docs/user/resources.md`, `interface/lib/docsManifest.ts`. Proof: `uv run inv interface.test`, doc tests. Accept: no user-facing "Deployment Context" string remains; glossary readable in `/docs`.
10. `COMPLETE` integrated proof (`bb6ea007`: full Chromium `93` passed / `14` skipped / `0` failed; Vitest `123` files; tsc clean; `go vet` + `go test ./... -count=1 -p 1` green; `pytest` `327` passed; line gate and `git diff --check` clean): `uv run inv interface.e2e --project=chromium --workers=1` from committed state, then the live walkthrough once the Environment lane clears.

Parallelism: A precedes B, C, and G; D precedes E and F; H is independent. RC-004E must not edit `TeamCreationPage.tsx`, the automations page, or `e2e/specs/settings.spec.ts`, which other lanes own.

## Acceptance

`MYC-RC-004` is accepted when:

- a simple operator creates any of the four starter organizations with zero config and sees its approval posture, output-proof rule, and team shape before creating it
- groups Soma creates for that organization inherit its work mode and approval policy
- the Deployments page lists standing teams, mission teams, groups, and planned departments in one list and explains each one from its detail pane
- "View config" returns readable YAML/JSON for templates, teams, groups, and organizations with secret references only
- clone and apply never mutate without a dry-run preview and a valid confirm token, and both emit audit events
- no user-facing surface says "Deployment Context"; the glossary defines organization, template, team, group, deployment, and knowledge
- focused specs and full Windows Chromium proof are green from committed state

## Owner Decisions (2026-09-25)

1. `DECIDED` Persist organizations first: new slice RC-004-0 replaces the in-memory `OrganizationStore` with a Postgres-backed store (migration 040) before starters ship. Operator templates move to migration 041.
2. `DECIDED` Picking a starter sets delivery defaults and shows departments as planned; Soma proposes launching teams and the operator approves. No automatic team instantiation.
3. `DECIDED` Deployments is in the main rail as a read-only view for every user; View config, clone, and apply stay behind Advanced mode.
4. `DECIDED` Graduated approval postures: agency approves external sends; SaaS approves deploy/write actions; operations runs bounded execution with approval on high-risk actions; regulated approves every mutation with full audit.

## Explicit Non-Goals

- Do not instantiate live runtime teams from a starter at organization creation in this plan; starters shape defaults and planned departments.
- Do not make standing-team manifests editable at runtime; they change through bundle files and restart.
- Do not add hosted providers or secrets to any starter.
- Do not rename the `deployment-context` API path or Go package.
- Do not revive the legacy `BlueprintDrawer` import/export.
- Do not require WSL proof while the Environment lane is `BLOCKED`.

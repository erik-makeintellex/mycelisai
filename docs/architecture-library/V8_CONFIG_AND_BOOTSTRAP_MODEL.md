# V8 Config and Bootstrap Model
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

Status: canonical V7-to-V8 bootstrap migration contract.

This index defines the bootstrap model and points durable detail to:
- [V8 Config Bootstrap Resolution](V8_CONFIG_BOOTSTRAP_RESOLUTION.md)
- [V8 Config Template Instantiation](V8_CONFIG_TEMPLATE_INSTANTIATION.md)
- [V8 Config Migration Dehardening](V8_CONFIG_MIGRATION_DEHARDENING.md)
- [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md)

## Purpose

Bootstrap turns source inputs into active organization/runtime shape. It is the required translation layer for V7 YAML, runtime config, DB seeding, operator wizard flows, templates, and deployment-specific settings.

Governing rule:

```text
Template != instantiated organization
```

Templates stay reusable blueprints. Instantiated organizations enter runtime resolution.

## Configuration Sources

Recognized source classes:
- templates and bootstrap bundles
- static configuration files
- database/runtime state
- operator input
- derived context
- deployment/env overrides for endpoint/profile wiring

Secrets stay in `.env` or external secret stores. `.env.compose` owns Compose topology and non-secret container shape.

## Bootstrap Resolution Flow

High-level stages:
1. Load source inputs.
2. Select or create the target organization/Inception.
3. Resolve Soma Kernel profile.
4. Resolve council/advisor composition.
5. Resolve team/department and agent/specialist defaults.
6. Resolve provider policy.
7. Resolve identity, memory, and continuity defaults.
8. Produce runtime-ready organization state.

See [V8 Config Bootstrap Resolution](V8_CONFIG_BOOTSTRAP_RESOLUTION.md).

## Scope Inheritance

Conceptual chain:

```text
Inception
  -> Soma Kernel
  -> Central Council roles
  -> Team / department defaults
  -> Agent / specialist overrides
```

Lower scopes may refine inherited settings only where higher policy allows.

## Precedence Rules

Policy is evaluated before overrides are accepted. Later sources or lower scopes do not automatically win if they violate governance, provider policy, or locked organization constraints.

Conflict handling should produce deterministic effective configuration and observable blockers for invalid input.

## Template and instantiation entry points

Entry paths:
- create from template
- create empty
- create from config/API
- clone and modify a template later

Templates define reusable shape. Instantiation binds identity, source metadata, initial policy, and runtime state. See [V8 Config Template Instantiation](V8_CONFIG_TEMPLATE_INSTANTIATION.md).

Template = reusable blueprint. Inception / AI Organization = actual instantiated organization. Templates are reusable organization blueprints, not just UI presets.

A template can include:
- organization type
- default Team Lead / Soma Kernel posture
- default Advisors / Council composition
- default Departments / Teams
- default Specialists / Agents
- default AI Engine Settings / provider policy
- default Memory & Personality settings
- optional beginner-facing labels and descriptions

By default, a template does not contain:
- live runtime state
- execution history
- per-run outcomes
- user-specific secrets

Required template kinds:
- starter templates for beginners
- domain templates such as Research, Engineering, and Marketing
- executive templates such as CTO, COO, and Product Lead
- personal / continuity templates
- empty / minimal template

### Template validation expectations

Templates should be treated as governed blueprint inputs, not free-form UI metadata blobs.
- declare enough structure to identify its organization type and intended operating posture
- organize defaults along the same conceptual scopes used by bootstrap resolution
- separate beginner-facing labels from runtime-shaping defaults
- be resolvable into an instantiated organization without inventing a second hidden template model

Instantiation paths:
1. create from template
2. create empty
3. create from config/API
4. clone and modify an existing template later

Template behavior:
1. the template supplies defaults
2. the instantiated organization becomes its own object after creation
3. later edits to the template do not silently rewrite existing organizations unless that behavior is explicitly designed and governed

A template can:
- define team defaults
- define agent defaults
- define optional advanced overrides

Beginner UI should mainly show:
- template name
- purpose
- a simple summary

Advanced panels may expose:
- internal structure
- routing posture
- scoped defaults

### Target-delivery implication

Templates must support target delivery rather than acting as decorative presets.
- templates are bootstrap inputs for real organizations that should be capable of participating in the governed execution platform
- instantiated organizations created from templates must still resolve into runtime behavior that can produce target product outcomes such as `answer`, `proposal`, `execution_result`, and `blocker`
- the template system should not become a second planning-only layer that is disconnected from execution, governance, or operator-visible delivery behavior

Templates must support target delivery rather than acting as decorative presets. Outcomes include `answer`, `proposal`, `execution_result`, and `blocker`.

## User Concept Layer

Beginner UI should present:
- AI Organization
- Soma
- Team Leads / advisors
- AI Engine Settings
- Memory & Continuity

Advanced views may expose config origin, inheritance, and overrides.

### Config exposure (View config)

Advanced operators read effective config through `GET /api/v1/config/{teams|groups|organizations|templates}/{id}` ([API Reference: View Config](../API_REFERENCE.md#view-config)). The surface is read-only and every document is redacted server-side before it is rendered:

- secret keys are hidden at any depth, matched case- and separator-insensitively: any key containing `password`, `passwd`, `secret`, `credential(s)`, `authorization`, `bearer`, `apikey`, or `privatekey`, and any key ending in `key` or `token` (`api_key`, `OPENAI_API_KEY`, `GITHUB_TOKEN`)
- a hidden value becomes `${env:NAME}` when a sibling `*_env` names its variable, otherwise `<redacted>`; the document lists every hidden path in `redactions`
- `*_env`, `*_env_var`, and `*_ref` references stay visible because they name where a secret lives, not the secret; a reference that holds a literal credential is hidden
- credential-shaped values (bearer tokens, known key prefixes, JWTs, PEM keys, URL passwords, `NAME=value` secret assignments) are hidden under any key
- team documents carry provider and model ids only; provider blocks from `cognitive.yaml`, which may hold literal `api_key` values, are never included

Standing-team manifests stay view-only; they change through bundle files and restart.

### Config clone and apply (RC-004G)

Advanced operators change config only through governed clone and apply ([API Reference: Clone And Apply Config](../API_REFERENCE.md#clone-and-apply-config)):

- `POST /api/v1/config/templates/{id}/clone` creates an operator template from a file bundle, a file starter, or another operator template. The Helm-mounted templates directory is read-only, so clones live in the `operator_templates` table (migration 041: `id`, `tenant_id`, `name`, `source_template_id`, `document` JSONB, `document_sha256`, `created_by`, created/updated audit event ids, `intent_proof_id`, timestamps). Clones are validated with the starter rules, including approval-policy resolution, and appear in the organization-starters catalog as `source: "operator"` ("Custom") with the same availability rules as file starters.
- `POST /api/v1/config/{groups|templates}/{id}/apply` applies a YAML or JSON document to a collaboration group or an operator template. File templates, standing-team manifests, and organizations are not applicable.
- Every clone and apply is a dry run first: the first call validates, diffs (paths only), and returns `202` with an intent proof and a single-use, expiring confirm token whose intent is bound to the SHA-256 of the canonical document. The retry consumes the token in the same transaction as the write; a different document gets `409`.
- A document holding `<redacted>`, an `${env:...}` placeholder in a secret field, or a literal secret is rejected, so a redacted View config can never be written back.
- Every confirmed mutation records an audit entry and `template.cloned` or `config.applied` mission events in the write transaction; previews record `template.clone.previewed` or `config.apply.previewed` audit entries.

## Migration from V7 bootstrap assumptions

V7 standing-team and YAML assumptions must be de-hardened into bundle-driven bootstrap. Normal startup should fail closed when required bootstrap truth is absent or ambiguous.

This is the canonical V7->V8 migration contract. V7 never exposed a single declarative bootstrap contract. Inputs included **YAML manifests and sidecar config files**, **Runtime configuration** (env vars, `.env`, CLI arguments), **Database state** produced by ad hoc migrations or interactive setup scripts, and **Operator flows** (UI wizards, CLI prompts) that mutated standing-team rows directly.

V7 assumptions included: standing team definitions were hydrated automatically at process start if the database was empty; Soma + Council roles were fixed to a single canonical lineup; runtime state (runs, manifests, NATS registrations) doubled as bootstrap inputs.

V8 replacement:
1. **Templates** provide reusable blueprints but remain separate from instantiated organizations (`Template ≠ instantiated organization`).
2. **Declarative configuration artifacts** (files, APIs, automation payloads) describe organization inputs
3. **Operator flows** submit explicit organization creation or update intents that feed the same instantiation pipeline
4. **Runtime/persistent state** supplies lineage and continuity only after an organization exists

V7 YAML and manifest assets remain valid migration inputs, but they must be translated into V8 template/config packages before use. Retired assumptions include Auto-hydrating standing-team rows at process start. and Treating last-run database state as the bootstrap plan. V8 keeps prior assets useful, but only after they conform to the explicit template + instantiation + inheritance + precedence pipeline.

See [V8 Config Migration Dehardening](V8_CONFIG_MIGRATION_DEHARDENING.md).

## Migration Note

Any slice touching startup, templates, runtime config, provider routing, team defaults, or operator creation flows must review this file and the relevant split doc.

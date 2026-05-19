# V8 Capability Manifest And Runtime Integration Standard
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-05-19
> Module Boundary: capability/MCP, runtime/deployment, governance/trust, advanced UI
> Purpose: Define the mandatory bridge between Soma actions, team/group execution, MCP tools, custom additions, outputs, runs, exchange normalization, artifacts, governance, and UI proof.

## Goal

Every execution path in Mycelis must be manageable, inspectable, governed, and reusable.

This applies to:

- Soma actions
- direct answers when they create retained proof
- team and group work
- automations and persistent work
- MCP tools
- custom connectors
- local scripts
- external APIs
- generated artifacts
- future plugins and modules

No execution path should exist as an unmanaged side effect.

## Canonical Concept: Capability Manifest

A `CapabilityManifest` is the product/runtime bridge between tools, governance, execution, outputs, and UI.

Every built-in MCP entry, external tool, local script, API connector, custom plugin, or future module must define a manifest before it can be assigned to Soma, Council, a team, a group, an automation, or a plugin surface.

Minimum manifest shape:

```yaml
capability:
  id: browser.search
  name: Web Search
  source: mcp
  category: research
  risk: medium
  approval: optional
  inputs:
    - query
  outputs:
    - ToolResult
    - ResearchSummary
  writes:
    - exchange.browser.research.results
    - artifacts.research
  allowed_roles:
    - soma
    - research_lead
    - reviewer
  audit: required
  health_check: true
```

Required manifest fields:

| Field | Contract |
| --- | --- |
| `id` | Stable capability identifier. |
| `name` | User-facing display name. |
| `source` | `builtin`, `mcp`, `openapi`, `local_script`, `external_api`, `python`, `a2a`, `plugin`, or another promoted adapter class. |
| `category` | Operator-readable purpose such as research, filesystem, media, deployment, review, memory, or automation. |
| `risk` | Capability risk class, at minimum `low`, `medium`, or `high`. |
| `approval` | `none`, `optional`, `required`, or `policy_resolved`. |
| `inputs` | Input schema references or compact field list. |
| `outputs` | Output schema references or output object classes. |
| `writes` | Exchange channels, artifact families, memory-candidate lanes, audit families, or deployment targets the capability may write to. |
| `allowed_roles` | Roles allowed to invoke or request the capability. |
| `audit` | Whether audit is required and which action family is recorded. |
| `health_check` | Whether runtime availability can be probed. |

Recommended manifest fields:

- `description`
- `version`
- `server_or_package`
- `config_refs`
- `secret_refs`
- `availability_status`
- `fallback_behavior`
- `retention_policy`
- `review_required`
- `uninstall_or_disable_behavior`
- `test_probe`

## Execution Contract

Every meaningful action must create or attach to a `Run`.

A `Run` must include:

- `run_id`
- initiating user or request
- Soma intent summary
- execution type
- capability or tool used
- status
- timestamps
- outputs
- errors or blockers
- audit records
- recovery or retry path

Run creation rules:

- Direct answer may stay as a chat record when it is non-mutating and not retained, but it must attach to a run or retained record when it produces durable proof, artifacts, reviewed output, learning candidates, or operator-visible execution evidence.
- Guided proposals must not execute before approval. Confirmation creates or resumes the execution run and records proof.
- Tool-assisted work must record capability use, normalized output, and failure/recovery state.
- Team/group work must attach team output, review state, retained artifacts, and audit records to the run.
- Automation and persistent work must create a run per cycle and expose next-run state.
- Custom plugin execution uses the same run contract as built-in capability execution.

## Output Contract

Every meaningful output must become a durable product object.

Output object types include:

- text answer
- plan
- review
- generated file
- media artifact
- tool result
- MCP result
- audit event
- learning candidate
- deployment proof
- external API result

Each output must include:

- `output_id`
- `run_id`
- type
- summary
- source capability
- creator role or team
- artifact path or payload reference
- `created_at`
- visibility
- retention policy
- review status
- downstream usability

Output normalization rules:

- Raw MCP, API, script, or plugin output is not the final product state.
- Tool results normalize into Managed Exchange first.
- Durable user-facing results also create an Artifact or retained output reference.
- Governance-sensitive outputs create or attach to audit evidence.
- Useful learning signals create `LearningCandidate` records before any durable memory promotion.

## MCP Integration Contract

Every MCP server/tool must register as a capability.

Each MCP capability must define:

- id
- display name
- tool/server source
- input schema
- output schema
- risk level
- allowed roles
- approval requirement
- audit requirement
- availability status
- fallback behavior
- user-facing description

MCP output must normalize into:

- Exchange item
- Artifact when durable
- Run evidence
- Audit event
- Learning candidate when useful

The UI should not treat MCP as only a list of servers. It should show what Soma can currently use, what each capability is for, whether it is available, what risk level it has, what outputs it can produce, and where those results appear.

## Custom Addition Contract

Any custom tool, connector, plugin, script, or module must register through the same capability system.

A custom addition must provide:

- manifest
- capability definitions
- input and output schemas
- permission requirements
- risk classification
- config and secret references
- health check
- test probe
- uninstall or disable behavior

Custom additions must not bypass:

- governance
- audit
- capability policy
- exchange normalization
- output persistence
- run proof
- UI availability and failure reporting

## Expected Execution Shapes

| Shape | Contract |
| --- | --- |
| Direct Answer | No tool, no mutation, lightweight run/chat record, optional retained summary. |
| Guided Proposal | Mutating or high-risk work, proposal generated, no execution before approval, confirmation creates run/proof. |
| Tool-Assisted Work | MCP/API/browser/file/local capability invoked, run records capability use, output normalizes into artifact/exchange. |
| Team Execution | Soma assigns work to team/group, team output becomes retained artifact or exchange item, review/audit attached. |
| Automation / Persistent Work | Scheduled or event-triggered, run created per cycle, outputs/blockers retained, next run visible. |
| Custom Plugin Execution | Plugin capability invoked, same governance/output/run contract applies. |

## UI Requirements

Default Soma UI must show:

- what was requested
- what Soma did
- tool, source, team, or capability used when relevant
- output produced
- proof/run link when retained or execution-backed
- next suggested action

Admin and Advanced UI must show:

- capability registry
- MCP servers and tools as capabilities
- custom additions and plugin capabilities
- input/output schemas
- run traces
- artifacts and retained outputs
- audit entries
- failures and retries
- health and availability state
- disable/uninstall path where applicable

## Integration Validation

Every MCP/custom capability must prove:

- registration works
- health check works
- unavailable state is clear
- input schema validation works
- output schema normalization works
- approval rules apply
- audit record is created
- artifact or output is retained when meaningful
- run proof exists
- failure is recoverable
- Soma can explain how the capability was used

## Release Standard

No integration is release-ready unless it is:

- registered
- permissioned
- auditable
- normalized through Exchange/output contracts
- visible in failure
- attached to run proof
- explainable by Soma
- covered by a test probe and at least one operator-visible validation path

## Orchestrated Delivery Teams

| Team | Status | Ownership | Next Deliverable |
| --- | --- | --- | --- |
| Architecture Lead | ACTIVE | Own `CapabilityManifest`, execution/output shape, and promotion rule across MCP/custom/local/API/plugin paths. | Keep this standard and the architecture index authoritative. |
| Runtime / Backend | IN_REVIEW | Typed manifest persistence/API exists; run attachment, output object mapping, and universal adapter enforcement continue. | Extend normalized capability invocation across every promoted execution path. |
| Capability / MCP | IN_REVIEW | MCP library/server/tool records map into visible capability posture while preserving MCP registry metadata. | Complete MCP-to-capability adapter enforcement and install/update/probe validation. |
| Exchange / Artifacts | NEXT | Ensure every tool/custom output lands in Managed Exchange and durable artifacts when meaningful. | Output normalization contract with retained object references. |
| Governance / Trust | NEXT | Resolve risk, role permission, approval, audit, and retention policy from manifest plus org/user policy. | Policy resolver test matrix for low/medium/high and local/remote/custom additions. |
| Interface / Operator UX | IN_REVIEW | Connected Tools now emphasizes what Soma can use and exposes capability posture. | Broaden run/output proof routing across all capability families. |
| Automation / Persistent Work | REQUIRED | Align recurring work with run-per-cycle and retained blocker/output semantics. | Automation run cycle proof with next-run visibility. |
| Validation | REQUIRED | Add registration, health, schema, approval, audit, output, run proof, and recovery tests for each promoted capability family. | Capability integration validation suite. |
| Docs / In-App Docs | ACTIVE | Keep state, architecture, user Resources, and docs manifest aligned. | In-app docs entry and user-facing Resources wording. |

## Current Adoption Posture

Existing managed exchange, capability-risk, MCP library, persisted capability state, audit, artifacts, and run-timeline work provides much of the foundation. Persistence/API and operator visibility are in review; the remaining promotion is universal enforcement: every tool/custom execution path must declare a manifest and every meaningful result must attach to run/output proof before it is considered product-complete.

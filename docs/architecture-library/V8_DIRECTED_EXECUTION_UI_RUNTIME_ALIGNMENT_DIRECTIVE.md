# V8 Directed Execution UI And Runtime Alignment Directive
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-05-10
> Module Boundary: advanced UI, runtime/deployment, capability/MCP, governance/trust, team/workflow
> Purpose: Align UI, runtime, and delivery discipline around Mycelis as a Soma-centered directed execution system.

## Canonical Product Truth

Mycelis is not chat plus tools. Mycelis is a Soma-centered directed execution system operating over governed capabilities, durable runs, normalized outputs, and inspectable proof.

Soma is:

- persistent orchestration interface
- planner
- runtime coordinator
- governance-aware execution layer

The intended product path is:

```text
Intent
  -> Soma understanding
  -> directed execution
  -> teams/tools/capabilities
  -> runs/artifacts/proof
  -> review/audit/recovery
```

The system should not feel like a dashboard, admin console, generic chat app, or MCP registry. It should feel like a governed AI operating environment.

## Runtime Principles

Every meaningful action is directed execution.

Required runtime posture:

- attach to a Run
- carry runtime identity
- carry capability context
- produce normalized outputs
- produce audit/proof metadata
- support recovery/review

Outputs are durable product objects. They may be answers, plans, reviews, artifacts, generated files, media results, MCP/tool results, audit events, learning candidates, deployment proofs, or external API results.

Capabilities are governed runtime objects. MCP, tools, APIs, plugins, scripts, and future modules must have manifest, schema, risk class, approval behavior, allowed roles, outputs, write destinations, audit expectations, and health/fallback status.

Runs are first-class UI concepts: execution objects, proof objects, operational history, and recovery surfaces.

Governance is part of runtime. Proposal -> confirm -> execute must stay visible, durable, reviewable, resumable, and trustworthy.

## Required Default Soma Surface

The primary Soma surface must show:

| Layer | What the user should understand |
| --- | --- |
| Intent | What the user asked for. |
| Soma understanding | What Soma interpreted. |
| Directed execution | Whether the shape is direct answer, guided proposal, tool-assisted work, team execution, automation, or plugin/custom capability. |
| Capability/team use | What was used and why. |
| Outputs | What was produced and where it lives. |
| Proof | Run ID, audit, verification, recovery, and blocker state when relevant. |
| Next step | Recommended continuation. |

Execution shape labels should be user-legible:

- Direct Soma
- Guided Proposal
- Tool-Assisted Work
- Team Execution
- Automation / Persistent Work
- Plugin / Custom Capability Execution

The UI may hide internal orchestration jargon, but it must not hide action, consequence, risk, proof, or recovery.

## Current-State Mapping

| Area | Current fit |
| --- | --- |
| Soma-first posture | The dashboard and organization workspaces already center Soma and direct answers. |
| Proposal governance | Proposal/cancel/confirm/audit behavior exists and is tested in the Soma governance lane. |
| Teams/groups | Compact team/group surfaces and retained-output review exist, including archived temporary-group review work. |
| Runs | `/runs` and run timelines exist, with status, chain, artifacts, retry/error concepts documented. |
| Outputs | Chat output cards, artifacts, media lanes, retained group outputs, and managed exchange outputs exist. |
| Capabilities | Connected Tools, MCP library, search capability, and capability-risk/audit foundations exist. |
| Directed-execution trust spine | Wave 1 execution summaries, proof links, tool-assisted work, and team/group retained-output proof are proven and in final review. |
| Deployment trust | The dedicated `mycelis-root` WSL proof lane is authoritative and green; Compose, deployment roots, artifact/log/cache directories, and proof environments exist operationally. |
| Docs/state | Capability Manifest and product-standard contracts now exist and are in-app discoverable. |

## Gap Analysis

| Gap | Why it matters |
| --- | --- |
| Default Soma surface does not yet consistently present intent -> understanding -> execution shape -> output -> proof -> next step as one causal package. | The product can still feel conversational instead of operational. |
| Runs are still partly secondary navigation rather than naturally linked proof after meaningful execution. | Users may miss the evidence trail that makes the system trustworthy. |
| Outputs are not yet consistently treated as durable objects across every path. | Tool/team/API results can feel transient or hard to revisit. |
| Connected Tools can still read as MCP/server inventory instead of capability availability for Soma. | Operators should understand what Soma can use, not raw integration plumbing. |
| Deployment/execution roots are not yet a first-class UI trust surface. | Self-hosting and release proof need visible checkout, deployment, execution, artifact, log, cache, and recovery context. |
| Governance metadata can be either too hidden or too technical depending on surface. | Trust requires clear proposal/proof language with advanced detail behind disclosure. |
| Admin/runtime structure can still leak into default UX. | The product should feel like directed work, not module operation. |

## Structural Recommendations

Prefer:

- unify execution summary around causal steps
- collapse duplicate activity/proof views into one run/output proof story
- elevate proof links near outputs
- demote raw infrastructure to advanced layers
- progressively reveal manifests, schemas, exchange internals, topology, and policy
- make capability use understandable as "what Soma used and why"

Avoid:

- adding more default panels
- creating parallel interaction paths for the same execution
- exposing raw MCP registry details as the primary story
- treating runs as debugging-only
- treating artifacts as downloads without provenance

## Ordered Execution Plan

1. Trust impact: add a reusable execution-summary model for Soma responses that carries intent, understanding, execution shape, capability/team use, outputs, proof, and next step.
2. Runtime clarity: ensure meaningful execution paths attach to Run and Output objects consistently across direct retained answers, proposals, tools, teams, automations, and plugins.
3. Release importance: make post-execution UI link outputs to run proof, audit/proposal state, and recovery where applicable.
4. Capability clarity: reshape Connected Tools around capabilities first, with MCP/server registry details behind advanced disclosure.
5. Run/Output hardening: tighten durable output objects, proof links, retained artifacts, and recovery metadata across meaningful execution paths.
6. Deployment trust: add System/Deployments trust surface showing checkout, deployment root, execution root, artifact/log/cache roots, current commit, proof status, and recovery action.
7. Governance clarity and GUI cleanup: normalize proposal/proof language, keep default Soma directed and outcome-shaped, and retire or collapse dashboard/admin-feeling panels that duplicate Soma-directed workflow state.

## Validation Standard

The product passes this directive only if:

- users understand Soma quickly
- actions visibly become directed executions
- runs/proof are understandable
- outputs are durable and traceable
- governance is trustworthy
- capabilities feel operational, not infrastructural
- failures and recovery are understandable
- the system feels alive, inspectable, and controlled

Final release standard:

```text
I tell Soma what I want.
Soma directs execution safely.
I can inspect the work, outputs, proof, and recovery path.
I trust the system.
```

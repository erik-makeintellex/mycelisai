# Mycelis V8.3 - Product Manifestation Architecture Review
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-06-14
> Purpose: Challenge the V8.3 architecture against product adoption, trusted deliverables, deployment reality, and operator comprehension without expanding doctrine.

## Review Rule

Every architecture concept must pass two questions:

| Question | Required Answer |
| --- | --- |
| Runtime | How does it work? |
| Product | How does the operator benefit? |

If a concept has a runtime answer but no clear user-value answer, it should be simplified, deferred, hidden behind `Inspect`, or removed from the MVP surface.

The controlling product test is:

```text
I tell Soma what I need.
Soma understands the goal.
Soma owns the outcome path.
I approve anything risky.
Useful deliverables appear.
Proof exists.
Recovery exists.
I can trust what happened later.
```

## Verdict

The architecture is aligned with the target product when it is judged by runtime capability. The remaining release risk is product manifestation: the same architecture must feel easier to trust, adopt, deploy, recover, and buy.

The MVP must therefore stay outcome-driven:

```text
Outcome Need
-> Soma Understanding
-> Proposal / Approval when needed
-> Execution
-> Deliverable
-> Proof
-> Recovery
-> Revisit Later
```

Runtime terms such as Event Spine, Execution Contract, capability manifest, MCP server, exchange schema, bus subject, team topology, provider routing, and deployment topology may remain internally important. They should not be the default product language.

## Subsystem Review

| Subsystem | Purpose | User Value | Runtime Value | Visibility | MVP | Risk If Removed |
| --- | --- | --- | --- | --- | --- | --- |
| Soma | Singular operating surface for intent, planning, execution, proof, and recovery. | The user can work through one counterpart instead of learning agents, tools, MCP, and topology. | Routes intent into governed plans, direct answers, work items, capabilities, and retained outputs. | Default | P0 | Product collapses into disconnected admin pages and tool consoles. |
| Outcome | The user-owned thing being completed, trusted, recovered, or revisited. | The user can see what is active, delivered, incomplete, trusted, and next. | Binds work items, output refs, proof refs, recovery refs, and state into a product object. | Default | P0 | Work remains scattered across chat, runs, groups, and resources. |
| Soma Understanding | Translate user language into goal, output shape, constraints, uncertainty, and approval posture. | The user sees what Soma understood before trusting or approving work. | Normalizes intent for contracts, teams, capabilities, and output selection. | Default | P0 | Users cannot tell whether Soma is doing the right work. |
| Governed Proposal / Approval | Prevent hidden mutation and risky execution. | The user approves important actions with a clear expected outcome. | Creates durable authority boundary before execution. | Default for risky work | P0 | Trust breaks through silent mutation or ambiguous consent. |
| Work Item / Work Inbox | Represent active, waiting, reviewable, failed, and completed work. | The user knows what needs attention and what happened next. | Projects run/team/capability state into operator-readable work state. | Default | P0 | Work disappears into chat history or logs. |
| Deliverable / Output Package | Treat outputs as retained results for outcomes. | The user can open, inspect, download, revisit, and trust generated work. | Normalizes files, apps, media, reports, datasets, and packages. | Default | P0 | Mycelis feels like chat instead of an operating environment. |
| Run Receipt | Summarize what happened, what was produced, trust, proof, and recovery. | The user can understand a run without reading timeline events. | Binds output refs, proof refs, status, recovery, and inspect links. | Default for runs | P0 | Runs remain engineering traces rather than trust objects. |
| Proof Artifact | Explain why a result is trustworthy. | The user can answer "why should I trust this?" without logs. | Persists evidence, validation state, capability refs, and degradation refs. | Secondary by default, expandable | P0 | Deliverables become unverifiable claims. |
| Recovery Action | Make failure actionable. | The user sees what failed, what remains trusted, and what to retry or repair. | Encodes retries, fallback, repair, resume, and target state. | Default when degraded | P0 | Degradation becomes dead-end error text. |
| Capability Catalog | Answer what Soma can use and what is degraded. | The user can understand search, files, media, code, and integrations by outcome. | Wraps MCP/tools/providers/scripts with risk, approval, health, output, and repair metadata. | Secondary, Resources | P0 | Operators must debug MCP/server plumbing to get value. |
| MCP / Tool Servers | Connect external and local capabilities. | Indirect: tools power deliverables, but raw server detail rarely helps normal users. | Provides concrete tool invocation surface and integration adapters. | Advanced / Inspect | P1 for raw detail, P0 through capability catalog | Hidden failure modes if not represented by capability health. |
| Teams / Groups | Scope work, output ownership, and collaboration lanes. | The user can delegate complex work while retaining reviewable outputs. | Provides bounded execution lanes, status events, and output refs. | Secondary | P0 for minimal team execution, P1 for rich management | Complex work has no durable owner or review lane. |
| Event Spine | Reconstruct execution, audit, proof, and recovery. | Indirect: enables trust and replay, but users should not see bus topology. | Durable execution truth for replay, audit, debugging, and proof lineage. | Hidden by default, Inspect for advanced | P0 internally | Proof/recovery cannot be reconstructed reliably. |
| Execution Contract | Define allowed capabilities, expected outputs, proof, risk, and recovery. | Indirect: the user benefits through safer proposals and predictable receipts. | Authority and shape for governed execution. | Hidden by default, summarized in proposal/receipt | P0 internally | Execution becomes ambiguous and hard to audit. |
| Managed Exchange / Artifacts | Normalize retained outputs, messages, files, and learning candidates. | The user can revisit and reuse work instead of losing it in chat. | Shared persistence and output normalization substrate. | Hidden by default, surfaced as outputs/history | P0 internally | Deliverables and context fragment across storage paths. |
| Memory / Continuity | Preserve useful context, decisions, and boundaries. | The user does not need to restate everything and can trust remembered context boundaries. | Stores scoped recall, user/company/customer/Soma context, and review posture. | Secondary | P1 for richer UX, P0 for safe boundaries | Soma loses continuity or overuses untrusted context. |
| Governance / Audit | Make authority, approvals, and accountability durable. | Enterprise and self-hosted users can trust who approved what and why. | Policy, approval, audit, and compliance substrate. | Default for approvals, secondary for audit | P0 | Buyers cannot trust or deploy the system safely. |
| Deployment / System Health | Explain runtime readiness and degraded services. | Operators can start, inspect health, and recover local/self-hosted runtime. | Maps processes, roots, endpoints, providers, and proof lanes. | Secondary, System | P0 for health, P1 for deep topology | Self-hosted adoption fails under opaque setup/runtime errors. |
| Resources / Workspace Files | Give direct access to generated content and storage. | The user can open output folders and inspect retained work. | File and workspace bridge for outputs, packages, and local capabilities. | Secondary | P0 | Generated content feels trapped or unverifiable. |
| Media Lane | Produce and retain private/local generated assets. | The user can generate media packages without public upstream dependency when configured. | Local gateway, provider adaptation, retained media proof. | Secondary when used | P1 for broad release, P0 for media workflows | Media work degrades or becomes privacy-dependent on external services. |
| Advanced Run Map | Reconstruct complex execution visually or step-by-step. | Advanced users can inspect how work happened without default clutter. | Projects event/proof/capability chain into a readable debug surface. | Advanced / Inspect | P1 | Default trust can still work, but deep debugging remains weaker. |
| Enterprise IAM / SSO | Support enterprise identity, access, and buying requirements. | Buyers can map Mycelis to organizational access and audit expectations. | Auth, federation, role, lifecycle, and admin posture. | Secondary/admin | P1/post-MVP depth, P0 basic auth | Enterprise adoption is blocked if baseline is absent. |
| Marketplace / Plugin Expansion | Future ecosystem for capabilities. | Future value, not required for first trusted workflow. | Extension and distribution substrate. | Hidden/post-MVP | Post-MVP | No MVP risk if capability catalog covers local tools. |

## MVP Alignment

| Architecture Component | MVP Classification | Product Decision |
| --- | --- | --- |
| Soma operating surface | Supports MVP directly | Default first surface. |
| Outcome object | Supports MVP directly | The user should see ownership, state, deliverables, proof, recovery, and revisit paths. |
| Soma understanding and proposal copy | Supports MVP directly | Must be concise and visible before approval. |
| Approval/governance | Supports MVP directly | Required for mutation and risky execution. |
| Work item and current work lane | Supports MVP directly | Must become inbox/list-detail rather than dense cards. |
| Output package | Supports MVP directly | Current P0.1 priority; must be openable and reusable. |
| Run receipt | Supports MVP directly | Must precede raw timeline/logs. |
| Proof artifact | Supports MVP directly | Summarize proof first, deep inspect second. |
| Recovery action | Supports MVP directly | Failed/degraded work must be actionable. |
| Capability catalog | Supports MVP directly | Resources must answer "what Soma can use." |
| Execution contract | Supports MVP indirectly | Hide by default; summarize in proposals/receipts. |
| Event Spine | Supports MVP indirectly | Hide topology; use it for proof/replay/recovery. |
| Capability manifests | Supports MVP indirectly | Surface as capability availability/risk/repair. |
| Teams/groups | Supports MVP directly for complex deliverables | Keep minimal by default; management depth is P1. |
| Memory/continuity | Supports MVP indirectly | Keep boundaries visible; avoid overbuilding personalization before output proof. |
| Deployment topology | Supports MVP indirectly | Surface health and impact, not topology diagrams. |
| Advanced visual run graph | Post-MVP unless needed for Inspect | Deliver a stepper before a graph. |
| Marketplace/plugin ecosystem | Post-MVP | Defer until P0 deliverable loop is trusted. |

## Visibility Rules

| Concept | Default Product Language | Visibility |
| --- | --- | --- |
| Execution Contract | What Soma will do and what approval means | Summarized; raw contract hidden |
| Event Spine | Activity, proof, and run history | Hidden; Inspect only |
| Capability Manifest | What Soma can use, risk, and repair | Capability catalog |
| MCP Server | Connected tool detail | Advanced inside Resources |
| Exchange Schema | Retained outputs, history, and reusable context | Hidden; surfaced as objects |
| Team Topology | Soma coordinated this work with a focused lane | Secondary |
| Deployment Topology | What is healthy, degraded, or needs repair | System summary, topology advanced |
| Provider Routing | AI engine used / local or external boundary | Secondary trust detail |
| Raw Logs / Payloads | Proof summary and safe next action | Inspect only |

## P0 Product Manifestation Actions

| Priority | Action | Product Reason | Exit Proof |
| --- | --- | --- | --- |
| P0.1 | Finish output package standard | Deliverables must support owned outcomes, not chat residue. | Soma, Groups, and Resources open the same retained package with preview/open/download/proof/inspect/revisit and no raw HTML default. |
| P0.2 | Prove service health and runtime readiness | The product cannot be trusted if startup is inconsistent. | Infrastructure, migrations, Core, Interface, and health reporting are clean and reachable. |
| P0.3 | Run headed browser proof for generated package | Architecture must prove real local execution and adoption path. | Visible browser proof: ask Soma, approve, execute, open generated app/file, open folder, re-enter through Resources/Groups. |
| P0.4 | Build review inbox/list-detail | Review work must become understandable. | Needs Review, Running, Outputs, Failed/Recovery, Archived have one obvious primary action. |
| P0.5 | Reframe Resources as capability catalog | Users need to know what Soma can use, not which servers exist. | Files/search/media/code/business/AI capabilities show available/degraded/missing/risk/repair. |
| P0.6 | Add receipt-first run pages | Proof must be user-facing trust, not engineering trace. | `/runs/[id]` opens to outcome/output/trust/recovery before timeline. |
| P0.7 | Standardize recovery queue language | Failure must become recoverable work. | Failed/degraded states show failed, still trusted, not trusted, safe next, operator action. |
| P0.8 | Record full MVP proof | Release confidence must be repeatable. | Clean-state source gates plus headed/headless proof of the canonical loop. |
| P0.9 | Align documentation to proof | State must reflect reality, not aspiration. | Docs and state explain current proof and next blockers after the MVP loop runs. |

## Adoption Review

### Self-Hosted Operator

The product succeeds when a self-hosted user can install, start, check health, understand degraded services, and trust local execution without infrastructure expertise.

P0 implication: System and Resources must translate service state into product impact:

```text
Search is unavailable, so Soma cannot verify current web sources.
Generated files and local memory still work.
Repair search in Resources, then retry the research task.
```

### Enterprise Reviewer

The product succeeds when an enterprise reviewer can understand governance, approval, audit, proof, and deployment posture without reading architecture docs.

P0 implication: governance proof must appear through proposals, receipts, retained outputs, and audit links, not as a separate doctrine requirement.

## Final Review Question

Every architecture or implementation slice should close by answering:

```text
Does this make Mycelis easier to trust,
easier to adopt,
easier to deploy,
easier to recover,
and better at owning trusted outcomes?
```

If the answer is not clearly yes, reconsider the slice.

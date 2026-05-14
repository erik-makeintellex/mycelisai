# V8.2 Operational Embodiment Directive
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8 MVP Governed Execution Mission Plan](V8_MVP_GOVERNED_EXECUTION_MISSION_PLAN.md) | [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md)

> Status: Canonical
> Last Updated: 2026-05-14
> Purpose: Shift V8.2 work from doctrine expansion into visible, repeatable, recoverable operational proof.

## TOC

- [Strategic Transition](#strategic-transition)
- [Core Product Truth](#core-product-truth)
- [Priority Order](#priority-order)
- [Visible Experience Rule](#visible-experience-rule)
- [Canonical MVP Workflow](#canonical-mvp-workflow)
- [Operational Degradation](#operational-degradation)
- [Confidence Provenance](#confidence-provenance)
- [Soma As Operational Identity](#soma-as-operational-identity)
- [Product Category](#product-category)
- [Team Focus](#team-focus)
- [Documentation Discipline](#documentation-discipline)
- [Acceptance Standard](#acceptance-standard)

## Strategic Transition

The architecture, runtime model, governance direction, and Soma-centered operating model are sufficiently defined for the next phase.

The next major risk is no longer insufficient architecture. The next major risk is continued doctrine expansion replacing operational embodiment.

V8.2 work must now prove itself through:
- visible execution
- durable outputs
- recoverable runs
- operator trust
- deployment reality
- repeatable workflows

## Core Product Truth

Mycelis is a Soma-centered governed cognitive operating environment.

The operator experience must collapse into:

```text
I tell Soma what I want.
Soma directs the work safely.
I can see what happened.
I can trust the result.
```

Runs, teams, memory, exchange, MCP, capabilities, deployment topology, audit, proof, and logs exist to support that experience. They are scoped operational contexts, not separate assistant identities.

## Priority Order

1. Operational embodiment: visible execution, durable outputs, inspectable proof, and recoverable runs.
2. Soma UX compression: reduce visible complexity while preserving capability.
3. Deployment trust: make runtime, roots, health, and proof understandable.
4. Recovery and degradation: treat failure as a normal operational state.
5. Confidence provenance: evolve proof toward why a result is trustworthy.

## Visible Experience Rule

Internal architecture may continue to deepen, but the visible operator experience must compress aggressively.

Default UX must collapse toward:

```text
Intent
-> Soma
-> Visible result
-> Proof / recovery if needed
```

Do not require operators to learn Event Spine, topology, capability manifests, governance layers, or execution contracts before they receive value.

Default surfaces should help the operator know:
- what they asked Soma to do
- what happened
- what result exists
- why the result can be trusted
- what to retry or review when trust is broken

## Canonical MVP Workflow

The single canonical MVP workflow is:

```text
User asks Soma to create or review something meaningful
-> Soma evaluates the request
-> Proposal generated if governed
-> User approves
-> Capability, tool, or team execution occurs
-> Run is created
-> Output or artifact is produced
-> Proof and audit are visible
-> User can revisit and trust the result
```

This workflow is the alpha demonstration, release confidence benchmark, and operational truth proof.

Everything else is secondary until this flow feels excellent, replayable, governed, visible, trustworthy, and recoverable.

## Operational Degradation

Provider failures, partial capability failures, malformed outputs, network instability, degraded execution, inconsistent model behavior, and hallucinated reasoning are normal operating conditions.

Runtime and UI must answer:
- what failed
- what remains trusted
- what proof is now invalid
- what can continue safely
- what must be retried
- what requires operator attention
- what uncertainty is now exposed

Operational degradation management is a first-class runtime concern, not a support-page afterthought.

Current runtime embodiment:
- `execution_summary` is the shared API/UI object for Soma-facing proof packages.
- The default Soma surface renders that object as an Operator trust package, not as architecture terminology.
- `audit_recovery.degradation` carries `code`, `what_failed`, `trusted_state`, `invalidated_proof`, `safe_continuation`, and `requires_attention`.
- Approved proposal execution failures must return failed run/proof/audit metadata instead of only a flat error.
- Search blockers must preserve provider blocker code and next action as degradation metadata.

## Confidence Provenance

The current proof layer shows providers, capabilities, approval state, outputs, and execution proof. The next trust layer is why Mycelis believes a result is trustworthy.

Schemas, runs, outputs, audit, and review surfaces should evolve toward:
- validation source
- independent verification
- cross-model agreement or disagreement
- execution confidence
- proof strength
- evidence quality
- review status

Full confidence provenance is not required immediately. New work should avoid shapes that make it hard to add later.

## Soma As Operational Identity

Soma should be singular and treated as persistent operational cognition identity, not merely an orchestrator, agent coordinator, or assistant.

Reflect this through:
- continuity
- trust mediation
- memory boundaries
- execution visibility
- organizational persistence

Do not over-emotionalize Soma. Make continuity and trust observable through product behavior.

## Product Category

Mycelis should be framed as a governed cognitive operating environment.

This category is defined by:
- governed execution
- inspectable agency
- authority boundaries
- runtime visibility
- durable outputs
- proof
- reconstructability
- operator trust

This language should shape product copy, demos, release strategy, deployment posture, and investor positioning.

## Team Focus

Active agentry should stay organized around five embodiment lanes:
- Soma Experience: one dominant operating surface, compressed complexity, stronger causality, unified dashboard/workspace behavior.
- Runtime and Capability: governed capabilities, normalized outputs, deployment/runtime trust, capability visibility.
- Governance and Trust: proposal/confirm/execute, proof visibility, degradation handling, audit/recovery, confidence provenance preparation.
- Deployment and Proof: deployment roots, execution roots, runtime health, proof visibility, self-hosted operator confidence.
- QA and Embodiment: first-run understanding, continuity, re-entry, recovery, visible execution, operator trust, replayable demos.

## Documentation Discipline

Canonical docs must stay compressed. Keep README, V8.2, runtime contracts, and UI contracts stable; keep `V8_DEV_STATE.md` operational; archive stale topology language and superseded doctrine aggressively.

Do not let documentation become another runtime system. Product experience must become simpler even while architecture deepens.

## Acceptance Standard

The key review question is:

```text
Can a user trust this system through visible execution and recovery?
```

Do not accept work merely because the architecture can describe it. Accept work when it strengthens the canonical MVP workflow, improves operational degradation handling, or makes proof more visible and repeatable.

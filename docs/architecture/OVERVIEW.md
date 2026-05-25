# Mycelis Architecture Overview
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Current implementation overview
> Last Updated: 2026-05-24
> Purpose: Summarize the active architecture without preserving superseded V7 doctrine as operational authority.

## TOC

- [Current Product Shape](#current-product-shape)
- [Layered Architecture](#layered-architecture)
- [Execution Model](#execution-model)
- [Durable Outputs And Trust](#durable-outputs-and-trust)
- [Supporting Specs](#supporting-specs)

## Current Product Shape

Mycelis is a Soma-centered governed cognitive operating environment. The product goal is not to expose every runtime subsystem by default; it is to let an operator ask Soma for meaningful work, watch governed execution, inspect durable outputs, and recover when trust breaks.

The active delivery target is V8.3 release-candidate embodiment on top of the V8.2/B2+ full architecture baseline:
- visible runs
- durable outputs
- inspectable proof
- recoverable failure states
- understandable deployment/runtime trust
- confidence provenance preparation

## Layered Architecture

| Layer | Owner | Current Role |
| --- | --- | --- |
| Interface | TypeScript / Next.js | Soma-primary workspace, teams, automations, resources, memory, system status, settings, docs, and run inspection |
| API and runtime | Go | execution contracts, governance, teams, NATS, persistence-facing behavior, services health, and normalized API envelopes |
| Persistence | SQL / Postgres | missions, teams, runs, outputs, events, approvals, memory, providers, settings, and audit data |
| Automation | Python / invoke | operator task runner, CI orchestration, local/runtime validation, release proof, and repo-local test harnesses |
| Deployment | Compose / Kubernetes | personal-owner Compose proof, Rancher/k3d Kubernetes lanes, and self-hosted deployment confidence |

## Execution Model

The canonical runtime shape is:

```text
Intent
 -> Soma understanding
   -> Execution Contract
     -> Governed Run
       -> Capability Invocation
         -> Output / Artifact
           -> Proof / Audit / Recovery
```

Every promoted capability should be visible as something Soma can use, not as hidden plumbing. Capabilities require manifests, risk classes, permissions, approvals where needed, audit, normalized outputs, and recovery behavior.

## Durable Outputs And Trust

Outputs are product objects, not transient chat text. Plans, reviews, files, media, generated artifacts, deployment proof, capability results, audit events, and retained learning should remain reviewable, reconstructable, attributable, and inspectable.

Governance is runtime infrastructure:
- proposal / confirm / execute must be durable and resumable
- mutating actions must not happen silently
- proof must make clear what happened and what remains trusted
- degradation states must explain what failed, what can continue, and what requires operator attention

## Supporting Specs

Use these documents for detailed authority:

- [Architecture Library Index](../architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
- [V8.2 Full Production Architecture](../../architecture/v8-2.md)
- [V8.3 Operational Embodiment PRD](../architecture-library/V8_3_OPERATIONAL_EMBODIMENT_PRD.md)
- [V8.3 Multi-Agentry Steering Doctrine](../architecture-library/V8_3_MULTI_AGENTRY_STEERING_DOCTRINE.md)
- [Backend](BACKEND.md)
- [Frontend](FRONTEND.md)
- [Operations](OPERATIONS.md)

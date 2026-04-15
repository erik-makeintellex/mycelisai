# V8 Self-Hosted Runtime Delivery Program
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical
> Last Updated: 2026-04-15
> Purpose: Define the compact delivery team, management contract, and acceptance gates for the deployable self-hosted runtime.

## TOC

- [Delivery Target](#delivery-target)
- [Focused Team](#focused-team)
- [Management Contract](#management-contract)
- [Coordination Rules](#coordination-rules)
- [Current Delivery Board](#current-delivery-board)
- [Acceptance Gates](#acceptance-gates)

## Delivery Target

The current delivery target is not "make local development convenient."

The current delivery target is:
- ship a deployable self-hosted Mycelis runtime on the supported single-host Docker Compose path
- keep the self-hosted Kubernetes and Helm path aligned as the scalable deployment contract
- treat external AI engines as first-class network services instead of assuming they run inside the same host or container boundary as Mycelis
- support the current real-world operator topology where the GPU-attached Windows host runs Ollama or another self-hosted model service, while Compose or Kubernetes workloads reach it through an explicit reachable host or IP

Canonical deployment posture:
- Linux-first Docker Compose is the primary single-host deployment/runtime story
- self-hosted Kubernetes is the scale-up and enterprise-aligned deployment story
- Windows-specific notes are host-connectivity guidance only; they are not the canonical runtime model
- `localhost`, `127.0.0.1`, and `0.0.0.0` are not valid AI-engine assumptions for containerized deployments

## Focused Team

Keep the active delivery team compact and role-specific.

| Role | Ownership | Success Condition |
| --- | --- | --- |
| Team Lead / Delivery Manager | scope, sequencing, blockers, acceptance, state updates | one active delivery target, explicit owners, no hidden blockers, clean handoffs |
| Platform Architect | deployment contract, topology rules, config/secrets/network/storage decisions | Compose and Kubernetes assumptions match real self-hosted deployment reality |
| Backend / Runtime Engineer | provider wiring, runtime config, health checks, external AI endpoint behavior | runtime treats external AI services as normal endpoints, not desktop-local shortcuts |
| Ops / Deployment Engineer | Compose, Helm, env contracts, volumes, recovery, operator runbooks | deployments can be brought up, checked, and repaired with documented task/ops flows |
| Validation / Release Engineer | browser proof, service-health proof, docs synchronization, acceptance evidence | every delivered slice has proof tied to the supported runtime contract |

Pull specialist support only when blocked:
- Security / Governance review when policy, approval, or trust posture changes
- Frontend / UX review when operator workflows or settings meaning changes

## Management Contract

- one delivery target at a time: `deployable self-hosted runtime with external AI host support`
- one board: `NEXT`, `ACTIVE`, `IN_REVIEW`, `COMPLETE`, `BLOCKED`
- one owner per slice
- every slice must declare:
  - owner
  - exact touched surfaces
  - proof required
  - docs that must change
  - acceptance statement
- no slice is complete until implementation, verification, and docs are all aligned
- no host-specific convenience path may overwrite the canonical Compose or Kubernetes deployment contract

Daily management rhythm:
1. confirm the single active target
2. review blockers and dependencies
3. confirm the next proof-producing slice
4. update `V8_DEV_STATE.md` when delivery truth changes

## Coordination Rules

- Start with runtime contract clarity before expanding feature work.
- Pull Compose delivery work ahead of broader Kubernetes work when the same assumption is shared by both lanes.
- Treat external AI engines as network dependencies owned by operator configuration, not as hidden side effects of a developer desktop.
- Keep docs and tests synchronized with the real runtime story in the same slice.
- Use compact ownership boundaries: architecture, runtime, deployment, validation.
- If a lane finds a deployability blocker outside its write scope, record the blocker and hand it to the owning lane instead of patching around it informally.

Required handoff order:
1. Platform Architect locks the contract.
2. Backend and Ops implement against that contract.
3. Validation proves the contract through service and browser evidence.
4. Team Lead accepts or requeues the slice.

## Current Delivery Board

| Team / Lane | Status | Current Target | Dependencies / Handoffs | Next Action |
| --- | --- | --- | --- | --- |
| Delivery Management | `ACTIVE` | Keep one active delivery target, explicit ownership, and acceptance tied to deployable runtime proof. | Receives status from every lane. | Hold the board to runtime delivery only; reject work that does not move Compose or self-hosted Kubernetes closer to deployable truth. |
| Runtime Contract + Architecture | `ACTIVE` | Lock the canonical deployment story to Linux-first Compose and self-hosted Kubernetes with external AI services reached by explicit host/IP or operator-supplied hostname. | Feeds Backend, Ops, Validation, and docs. | Audit remaining canonical docs for desktop-local or Windows-only runtime assumptions and correct them. |
| Backend / Runtime Integration | `ACTIVE` | Keep provider configuration, health checks, and runtime behavior aligned to external AI endpoints instead of local loopback assumptions. | Depends on Runtime Contract + Architecture. | Tighten remaining runtime/config surfaces so external Ollama or other self-hosted AI endpoints are first-class configuration. |
| Compose Deployment | `ACTIVE` | Keep the Compose path deployable from Linux or WSL-hosted Docker without Docker Desktop assumptions, with explicit external AI host configuration. | Depends on Runtime Contract + Architecture and Backend / Runtime Integration. | Continue proving `compose.up`, `compose.status`, and `compose.health` against the supported self-hosted topology. |
| Kubernetes / Helm Deployment | `NEXT` | Define the self-hosted Kubernetes contract for external AI services, secrets, bootstrap config, storage, and operator recovery. | Reuses the same runtime contract as Compose. | Convert the external-AI-host and retained-storage assumptions into explicit Helm and operator guidance. |
| Validation + Release Proof | `ACTIVE` | Make service-health, browser proof, and docs-link proof certify deployable runtime behavior instead of development-only convenience. | Depends on Compose and Backend lanes; later on Kubernetes. | Keep Compose proof current now, then add the Kubernetes acceptance checklist once its contract is documented. |
| Operator Handoff Docs | `ACTIVE` | Keep canonical docs and in-app docs aligned to the same deployment story. | Depends on every lane. | Update state, architecture docs, and docs manifest whenever deployment meaning changes. |

## Acceptance Gates

The runtime delivery lane is only complete when all of the following are true:

1. The Compose path is documented and validated as a real self-hosted deployment path.
2. The Kubernetes / Helm path has an explicit operator contract for the same runtime assumptions.
3. External AI providers are configured through explicit reachable endpoints, not desktop-local shortcuts.
4. Health checks and browser proof reflect deployable runtime behavior.
5. Canonical docs, in-app docs, and live state all tell the same story.

Current acceptance bias:
- ship the real deployment contract first
- keep the team compact
- require proof for every slice
- avoid dev-only assumptions becoming architecture truth

# V8 Enterprise Self-Hosted Kubernetes Delivery Plan
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical
> Last Updated: 2026-04-15
> Purpose: Define the delivery plan, local proof model, and compact team structure for turning the current self-hosted Kubernetes lane into an enterprise-compatible deployment contract.

## TOC

- [Why This Plan Exists](#why-this-plan-exists)
- [Deployment Modes](#deployment-modes)
- [Delivery Target](#delivery-target)
- [Local Enterprise-Simulation Stack](#local-enterprise-simulation-stack)
- [Focused Team](#focused-team)
- [Management Contract](#management-contract)
- [Delivery Board](#delivery-board)
- [Acceptance Gates](#acceptance-gates)
- [Immediate Execution Sequence](#immediate-execution-sequence)

## Why This Plan Exists

Mycelis now has a real supported Compose runtime and a partial Helm/Kubernetes path.

That is not yet enough for enterprise-compatible self-hosting.

Enterprise deployment readiness means:
- the same chart can be promoted across local validation, staging, and customer-managed clusters without patching templates
- configuration, secrets, networking, storage, and image delivery work through standard cluster contracts
- local Kubernetes validation proves enterprise-compatible assumptions instead of only proving a developer convenience path

This plan defines how to get there without turning the delivery team into a sprawling planning swarm.

## Deployment Modes

Use these deployment modes deliberately and keep their purposes distinct.

| Mode | Purpose | Canonical posture |
| --- | --- | --- |
| `compose_single_host` | single-owner or small-team self-hosted runtime | supported deployable runtime for home-lab, partner demos, and compact self-hosting |
| `k3d_local_k8s` | local Kubernetes validation lane inside WSL/Linux Docker | enterprise-simulation dev/test lane, not the production target |
| `enterprise_self_hosted_k8s` | customer-managed cluster deployment | Helm/GitOps/promoted-values contract for real enterprise clusters |
| `edge_binary_node` | lightweight node-attached runtime helper or specialist host | binary/image variant for local-only users or edge nodes such as Raspberry Pi-class systems |

Rules:
- `k3d_local_k8s` replaces Kind as the preferred local Kubernetes validation lane for WSL/Linux work.
- `k3d_local_k8s` exists to simulate enterprise cluster posture locally; it is not itself the enterprise deployment target.
- `enterprise_self_hosted_k8s` must not depend on Docker Desktop, Kind-only image loading, or localhost-only AI endpoints.
- `edge_binary_node` is a deployment variant, not a reason to weaken the cluster deployment contract.

## Delivery Target

The current delivery target is:
- keep Compose as the supported single-host self-hosted runtime
- establish `k3d` as the canonical local Kubernetes validation lane
- make the Helm chart enterprise-compatible through values, secrets, networking, storage, and image-delivery contracts
- prove that Mycelis can be promoted into a customer-managed Kubernetes cluster without forking templates or reintroducing desktop-local assumptions

The current non-targets are:
- treating Kind as the long-term local Kubernetes truth
- baking desktop-only shortcuts into the chart or runtime
- hardcoding customer-specific storage classes, ingress controllers, or secret managers
- claiming enterprise readiness based only on a single-user desktop success path

## Local Enterprise-Simulation Stack

Local proof for enterprise compatibility must use a stack that exercises real cluster contracts:

1. Docker inside WSL/Linux
2. `k3d`
3. `kubectl`
4. `helm`
5. local registry or `k3d image import`
6. manifest validation (`helm lint`, rendered-manifest/schema checks)
7. policy/security checks for chart posture
8. explicit external AI endpoint such as a Windows-hosted Ollama box reached by hostname or IP

Required local enterprise-simulation behaviors:
- explicit image tag or digest, not `latest`
- externalized secrets posture
- PVC-backed storage
- ingress or equivalent service exposure contract
- no required `hostPath` assumptions for normal cluster deployment
- no `localhost` AI provider assumptions in cluster mode
- chart renders cleanly for promoted values files

## Focused Team

Keep the delivery team compact and specialized.

| Role | Ownership | Primary surfaces | Success condition |
| --- | --- | --- | --- |
| Delivery Manager | scope, sequencing, blockers, state updates, gate decisions | `V8_DEV_STATE.md`, delivery docs, board cadence | one active target, explicit owners, no silent drift |
| Platform Architect | deployment contract, environment promotion model, topology rules | architecture docs, chart contract, values model | local `k3d` and enterprise cluster assumptions align |
| Chart / Runtime Engineer | Helm surfaces, config projection, storage/network/image hooks | `charts/mycelis-core/**` | chart supports enterprise-compatible overrides without template forks |
| Ops / Deployment Engineer | local cluster automation, image flow, local-k8s tasks, runbooks | `ops/k8s.py`, `docs/LOCAL_DEV_WORKFLOW.md`, tests | local `k3d` validation is deterministic and repeatable |
| Validation / Release Engineer | chart/render checks, browser/runtime proof, release evidence | `tests/**`, validation docs, CI tasks | enterprise-simulated local proof is automated and reviewable |

Pull specialist support only when blocked:
- Security review for secret-manager, network-policy, or policy-control changes
- Frontend review when operator-facing deployment or recovery flows change

## Management Contract

- one active target: `enterprise-compatible self-hosted Kubernetes delivery with k3d as the local validation lane`
- one board: `REQUIRED`, `NEXT`, `ACTIVE`, `IN_REVIEW`, `COMPLETE`, `BLOCKED`
- one owner per slice
- every slice must declare:
  - owner
  - exact touched surfaces
  - proof required
  - docs that must change
  - acceptance statement

Cadence:
1. confirm the active target
2. review blockers by owner
3. move one proof-producing slice per lane
4. update state/docs when delivery truth changes

PM rules:
- protect the boundary between `k3d_local_k8s` and `enterprise_self_hosted_k8s`
- do not accept “works in my desktop cluster” as enterprise evidence
- do not let one lane patch around another lane’s blocker without recording the contract change

## Delivery Board

| Lane | Status | Owner | Current target | Next proof |
| --- | --- | --- | --- | --- |
| Delivery Management | `ACTIVE` | Delivery Manager | keep one enterprise-Kubernetes target and compact team ownership | updated state + accepted slice board |
| Platform Contract | `ACTIVE` | Platform Architect | define `k3d` as local K8s lane and enterprise chart promotion rules | canonical plan + docs alignment |
| Chart Enterprise Readiness | `NEXT` | Chart / Runtime Engineer | add enterprise-friendly Helm surfaces for ingress, scheduling, image pull, and secrets posture | rendered chart diff + focused tests |
| Local K8s Ops | `NEXT` | Ops / Deployment Engineer | replace remaining Kind-biased local tasking with `k3d`-aware automation and runbooks | deterministic local `k3d` up/deploy/status proof |
| Validation + Release | `NEXT` | Validation / Release Engineer | make chart render/policy/runtime proof part of the repo gate | local enterprise-sim render/test gate |
| Enterprise Staging Contract | `REQUIRED` | Shared after preceding lanes | define promoted values files and customer-managed dependency posture | documented deploy contract for real clusters |

## Acceptance Gates

This delivery lane is only complete when all of the following are true:

1. `k3d` is the documented and supported local Kubernetes validation lane for WSL/Linux work.
2. The Helm chart supports enterprise-friendly override surfaces without template patching.
3. The runtime uses explicit reachable AI endpoints in cluster mode.
4. Local Kubernetes proof validates enterprise-compatible assumptions: storage, ingress/service exposure, image delivery, config/secret posture, and health checks.
5. The repo contains promoted chart values guidance for local, staging, and enterprise-managed cluster deployment.
6. Docs, tasks, and tests all describe the same Kubernetes story.

Evidence bias:
- local `k3d` proof is required
- render/validation proof is required
- Compose success alone is not sufficient
- customer-specific infrastructure is not baked into the generic chart

## Immediate Execution Sequence

Execute in this order:

1. `ACTIVE` publish this delivery plan and align state/docs to it.
2. `NEXT` add enterprise-ready chart surfaces:
   - ingress
   - service account
   - image pull secrets
   - node selector
   - tolerations
   - affinity
   - topology spread constraints
   - explicit image tag/digest posture
3. `NEXT` convert local Kubernetes tasking from Kind bias to `k3d` bias:
   - cluster create/delete/status contract
   - image import or local registry flow
   - kubeconfig guidance
4. `NEXT` add enterprise-simulated validation gates:
   - Helm lint/render
   - chart config tests
   - local `k3d` smoke deploy
5. `NEXT` publish promoted-values guidance:
   - local `k3d`
   - enterprise-managed services
   - external AI host
6. `IN_REVIEW` run the full local proof and update `V8_DEV_STATE.md` with the resulting delivery truth.

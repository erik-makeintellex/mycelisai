# Mycelis Active Development State

> Canonical implementation scoreboard. Product and architecture authority lives in [Mycelis Canonical PRD](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md). This file contains current delivery state only; completed history and superseded plans live in Git history.

## Active Snapshot

| Field | Current state |
| --- | --- |
| Updated | 2026-08-11 |
| Integration branch | `dev` |
| Active slice | P0.13 fresh-user desktop/mobile acceptance is `IN_REVIEW`. The feature branch passes the integrated authenticated desktop/compact matrix and adjacent headed UI regression set; merge and post-merge `dev` proof remain. |
| Production branch | `main` remains behind the accepted `dev` integration train pending release certification. |
| Runtime posture | Dockerized PostgreSQL and NATS plus local Core and Interface are healthy; Ollama is running. Full Compose app containers, Kubernetes, and WSL remain explicit release-proof lanes. |
| Delivery target | Integrate the proven P0.13 fresh-user device matrix, close ambiguous external-mutation recovery, then execute P0.12 release certification and intentional `dev -> main` promotion. |
| Main release risk | The canonical natural and trusted delivery journeys now produce and validate usable browser packages, but production promotion still requires ambiguous external-mutation recovery and deployment certification from a clean release candidate. |
| Canonical PRD alignment | Workspace owns Outcomes; Soma owns execution; WorkIntent is transitional; deliverables, proof, recovery, and continuity remain Outcome-owned user concepts. |

## Documentation Shortcuts

| Need | Open |
| --- | --- |
| Start using Mycelis | [User Docs Home](../docs/user/README.md) |
| Ask Soma, approve work, and continue a result | [Using Soma Chat](../docs/user/soma-chat.md) |
| Open outputs or configure capabilities and data access | [Outputs And Resources](../docs/user/resources.md) |
| Understand failures and recover work | [System Status And Recovery](../docs/user/system-status-recovery.md) |
| Execute human release acceptance | [User Acceptance Runbook](../docs/REMOTE_USER_TESTING.md) |
| Run automated or live browser proof | [Testing Guide](../docs/TESTING.md) |
| Operate the local and deployment stack | [Operations](../docs/architecture/OPERATIONS.md) |
| Inspect endpoint contracts | [API Reference](../docs/API_REFERENCE.md) |
| Review target product architecture and P0 gates | [Mycelis Canonical PRD](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md) |

## Delivery Map

| Lane | Status | Current evidence | Next gate |
| --- | --- | --- | --- |
| P0.9 full Trusted Outcome Journey | `COMPLETE` | Governed ask, approval, asynchronous NATS team execution, retained interactive package, validation, proof, and cross-surface revisit pass on the local source stack. | Preserve during release certification. |
| P0.11 documentation convergence | `COMPLETE` | The canonical PRD contains all product architecture. The duplicate architecture overview and obsolete user-facing blueprint guide are deleted; user concepts, run/recovery, acceptance, deployment, API/provider, implementation, README, state, and in-app Help guidance match current Workspace/Outcome delivery and the Docker data-plane development lane. Documentation contracts, Help component tests, typecheck, and a headed real-manifest PRD-to-acceptance journey pass. | Keep one product-architecture authority and update owned support docs with behavior changes. |
| P0.13 retained-state projection | `COMPLETE` | TeamWork has an explicit attention view; Dashboard requests operator attention only; Groups defaults to Current with separate Completed and Archived history. | Preserve in fresh-user proof. |
| P0.13 cross-device operator acceptance | `IN_REVIEW` | Dashboard, Groups, Resources, Memory, Docs, and Settings have focused desktop/mobile proof. Exact owner-scoped creation and purge pass visibly in Chromium; 32 fixture scopes are purged, no non-terminal scopes remain, and active claims are zero. Headed natural delivery creates a complete browser package, validates its declared interaction, opens it directly, and purges its resources. A post-merge run that crossed five minutes exposed orphan result correlation; the corrected feature and merged `dev` runs passed in 5.8 and 5.5 minutes. The headed Trusted Outcome Journey also passes Ask through Revisit with retained proof and cleanup. Conversational control is merged and post-merge proven. The integrated fresh-user matrix now starts from a stale work URL, signs into Soma, traverses Groups creation, Resources, Memory, Docs, Settings, and returns to Soma with zero page, console, hydration, `5xx`, or horizontal-overflow failures at desktop and compact widths. The adjacent 40-journey headed set passes after focused Groups history and Vault hydration reruns. | Merge to `dev` and repeat the affected integration proof. |
| P0.12 release hygiene and certification | `ACTIVE` | Clean Windows release preflight passes lint, Core, build, typecheck, Vitest, 121 Chromium journeys, source service health, and all four live governance scenarios. Configured Compose ports, shared identity-secret boundaries, and the Docker PostgreSQL/NATS plus local Core/Interface development topology are merged and re-proven on `dev` without app image builds. | Finish deterministic QA ownership and P0.13 acceptance, then certify full Compose/Kubernetes and Windows-browser release paths intentionally. |
| Production promotion | `REQUIRED` | `main` is intentionally not promoted from the current `dev` train. | Merge certified `dev` into `main`, rerun health/browser smoke, update state, and synchronize remotes. |

## Immediate Work Order

1. Merge the P0.13 fresh-user acceptance slice to `dev` and repeat its focused unit, type, health, and headed browser proof.
2. Close ambiguous external-mutation recovery before accepting retries after uncertain timeouts.
3. Run `uv run inv ci.release-preflight --lane=release` from a clean committed `dev` state.
4. Certify the supported full Compose/Kubernetes runtime and Windows-browser path; use guarded WSL proof only when that environment supplies distinct evidence.
5. Promote `dev` to `main` only after every required gate passes.

## Accepted Proof

### Runtime And Trust

- Approved work commits durable WorkIntent, ExecutionContract, Outcome/work ownership, and idempotent dispatch state before asynchronous handoff.
- Soma remains conversational while NATS-backed work progresses; start, progress, result, proof, and recovery are projected from correlated durable state.
- Final interactive deliverables require successful retained-file readback and contract-selected browser validation before verified completion.
- Focused live ownership proof opens an exactly scoped organization and proves it is unavailable after purge. Thirty-two scopes are terminally purged with zero active claims and no non-terminal scope. Failed attempts and successful delivery both purge only their exact Group, team-work, run, organization, and workspace resources without touching NATS.
- Headed natural delivery and Trusted Outcome journeys prove asynchronous team execution, complete retained browser packages, structural and interaction validation, direct opening, proof/revisit, and deterministic cleanup.
- Team command correlation outlives the durable recovery deadline; a 5.8-minute local-model package delivery remains attached to its work item instead of publishing an orphan result.
- PostgreSQL recovery deadlines and restart reconciliation expose overdue accepted work as degraded operator attention instead of silent loss.
- Direct run readback includes non-empty run-event evidence for completed governed execution.

### Operator Experience

- The authenticated default is the Soma threaded workspace with compact governance and Outcomes opened on demand.
- Proposal approval, cancellation, and revision stay in the Soma composer: bounded replies execute the matching lifecycle action, while qualified replies remain normal conversation. The default workspace has no action shelf, team/routing picker, mode selector, status cluster, or proposal button stack.
- Opening proposal Details scrolls the disclosed content into the bounded conversation viewport so inspection does not hide behind the composer.
- The composer remains reachable and expands with user input before bounded scrolling.
- Groups, Resources, Memory, Docs, and Settings use focused list/detail, tabs, or overlays rather than full-page content stacks.
- Authenticated desktop and compact-phone navigation reaches every primary operator surface through visible controls while retaining a reachable Soma composer and zero document-level horizontal overflow.
- Output actions open retained files/packages directly; runtime IDs and transport details remain behind Details or Inspect.
- Current, Completed, and Archived work are distinct; retained history does not inflate active review counts.

### Documentation And Quality

- `docs/architecture-library/MYCELIS_CANONICAL_PRD.md` is the single product architecture and PRD authority.
- Supporting Backend, Frontend, Operations, API, Testing, and user guides have bounded implementation or operator ownership.
- Superseded product doctrine is deleted, not archived in the active docs tree.
- The real in-app Help manifest opens the Canonical PRD and User Acceptance Runbook in headed Chromium and no longer exposes the deleted overview.
- Every behavior slice reviews README, this scoreboard, the owning docs, and the in-app Help manifest when applicable.

## Open Release Risks

| Risk | Status | Required resolution |
| --- | --- | --- |
| Deterministic QA ownership | `COMPLETE` | Migrations `058`-`060`, opt-in root-admin API, exact creation provenance, reentrant creation/purge fencing, current-run staged-ownership allowance, durable pre-existence rejection, independently committed created-team/workspace claims, resumable purge, atomic claim release, and NATS exclusion are covered by Core, Python, UI proxy, and post-merge headed Chromium proof. Thirty-one scopes are purged with zero active claims or non-terminal scopes. | Preserve in every retained-state live journey and release certification. |
| Generated package delivery and correlation | `COMPLETE` | Worker correction requests one missing contract item at a time, entrypoint guidance forbids positional event selectors, and completion still requires retained-file readback plus structural and browser interaction validation. The command correlation lease exceeds the durable work recovery deadline; its six-minute regression and headed natural journeys pass on the feature branch and merged `dev` in 5.8 and 5.5 minutes. | Preserve in the fresh-user device matrix and release certification. |
| Ambiguous external mutation | `REQUIRED` | Capability-level idempotency or explicit uncertain-result recovery must prevent false completion after timeouts. |
| Windows Core restart wrapper | `IN_REVIEW` | Distinguish a foreground wrapper timeout from service failure and retain useful diagnostics. |
| Deployment certification | `REQUIRED` | WSL, Compose, Kubernetes, and operator-facing browser paths must pass from the committed release candidate. |
| Production promotion | `REQUIRED` | Promote only after clean-tree release proof; do not mix new feature work into the candidate. |

## Release Proof Order

```text
clean feature proof
-> merge to dev
-> affected post-merge integration proof
-> clean release preflight
-> optional refreshed WSL validation when it supplies distinct evidence
-> Compose and Kubernetes deployment proof
-> headed browser certification
-> dev to main promotion
-> post-promotion health and browser smoke
```

## Documentation Rule

Do not add historical transcripts, architecture archives, temporary team plans, browser logs, or superseded doctrine here. Update this file when current status, accepted evidence, risks, or the immediate work order changes. Use Git history for prior checkpoints.

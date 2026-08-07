# Mycelis Active Development State

> Canonical implementation scoreboard. Product and architecture authority lives in [Mycelis Canonical PRD](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md). This file contains current delivery state only; completed history and superseded plans live in Git history.

## Active Snapshot

| Field | Current state |
| --- | --- |
| Updated | 2026-08-06 |
| Integration branch | `dev` |
| Active slice | P0.12 Compose port certification on `feature/promotion-compose-port-certification`; full bring-up is being aligned to configured host ports after independent WSL proof exposed hardcoded readiness probes. |
| Production branch | `main` remains behind the accepted `dev` integration train pending release certification. |
| Runtime posture | Native PostgreSQL, NATS, Core, Interface, and Ollama are the supported source-development stack. Current lifecycle health is green. |
| Delivery target | Finish P0.13 operator acceptance and deterministic QA-data ownership, then execute P0.12 release certification and intentional `dev -> main` promotion. |
| Main release risk | Fresh-user proof can still inherit retained QA fixtures. Cleanup must identify owned test records explicitly and must not infer ownership from names or clear shared NATS state. |
| Canonical PRD alignment | Workspace owns Outcomes; Soma owns execution; WorkIntent is transitional; deliverables, proof, recovery, and continuity remain Outcome-owned user concepts. |

## Delivery Map

| Lane | Status | Current evidence | Next gate |
| --- | --- | --- | --- |
| P0.9 full Trusted Outcome Journey | `COMPLETE` | Governed ask, approval, asynchronous NATS team execution, retained interactive package, validation, proof, and cross-surface revisit pass on the native source stack. | Preserve during release certification. |
| P0.11 documentation convergence | `COMPLETE` | The canonical PRD contains the worker execution contract and all product architecture. The redundant architecture index/source-map documents are removed, the state transcript contains current truth only, and post-merge docs/Core/Interface/headed desktop/mobile Help proof passes on `dev`. | Keep one product-architecture authority and update owned support docs with behavior changes. |
| P0.13 retained-state projection | `COMPLETE` | TeamWork has an explicit attention view; Dashboard requests operator attention only; Groups defaults to Current with separate Completed and Archived history. | Preserve in fresh-user proof. |
| P0.13 cross-device operator acceptance | `IN_REVIEW` | Dashboard, Groups, Resources, Memory, Docs, and Settings have focused desktop/mobile proof with no blocking overlap, hidden composer, page errors, console errors, or horizontal document overflow. | Add deterministic fixture ownership and rerun the integrated new-user matrix. |
| P0.12 release hygiene and certification | `ACTIVE` | Clean Windows release preflight passes lint, Core, build, typecheck, Vitest, 121 Chromium journeys, native service health, and all four live governance scenarios. Independent WSL runtime preflight also passes; Compose proof exposed that full `compose.up` readiness ignored configured host ports. | Integrate configured-port readiness, rerun WSL headed Compose validation, then complete Kubernetes proof and promotion decision. |
| Production promotion | `REQUIRED` | `main` is intentionally not promoted from the current `dev` train. | Merge certified `dev` into `main`, rerun health/browser smoke, update state, and synchronize remotes. |

## Immediate Work Order

1. Implement owner-scoped QA fixture lifecycle and purge across Groups, Outcomes, artifacts, team work, runs, and proof without deleting legitimate operator data.
2. Rerun P0.13 fresh-user desktop/mobile acceptance against clean owned fixtures.
3. Run `uv run inv ci.release-preflight --lane=release` from a clean committed `dev` state.
4. Validate the guarded `uv run inv wsl.validate --lane=release` path from the refreshed WSL proof checkout, then certify Compose and Kubernetes runtime paths.
5. Run broader headed browser certification from committed state and retain JSON/JUnit evidence.
6. Promote `dev` to `main` only after every required gate passes.

## Accepted Proof

### Runtime And Trust

- Approved work commits durable WorkIntent, ExecutionContract, Outcome/work ownership, and idempotent dispatch state before asynchronous handoff.
- Soma remains conversational while NATS-backed work progresses; start, progress, result, proof, and recovery are projected from correlated durable state.
- Final interactive deliverables require successful retained-file readback and contract-selected browser validation before verified completion.
- PostgreSQL recovery deadlines and restart reconciliation expose overdue accepted work as degraded operator attention instead of silent loss.
- Direct run readback includes non-empty run-event evidence for completed governed execution.

### Operator Experience

- The authenticated default is the Soma threaded workspace with compact governance and Outcomes opened on demand.
- The composer remains reachable and expands with user input before bounded scrolling.
- Groups, Resources, Memory, Docs, and Settings use focused list/detail, tabs, or overlays rather than full-page content stacks.
- Output actions open retained files/packages directly; runtime IDs and transport details remain behind Details or Inspect.
- Current, Completed, and Archived work are distinct; retained history does not inflate active review counts.

### Documentation And Quality

- `docs/architecture-library/MYCELIS_CANONICAL_PRD.md` is the single product architecture and PRD authority.
- Supporting Backend, Frontend, Operations, API, Testing, and user guides have bounded implementation or operator ownership.
- Superseded product doctrine is deleted, not archived in the active docs tree.
- Every behavior slice reviews README, this scoreboard, the owning docs, and the in-app Help manifest when applicable.

## Open Release Risks

| Risk | Status | Required resolution |
| --- | --- | --- |
| Deterministic QA ownership | `NEXT` | Tag fixtures at creation and provide owner-scoped cleanup with database and filesystem safeguards. |
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
-> refreshed WSL validation
-> Compose and Kubernetes deployment proof
-> headed browser certification
-> dev to main promotion
-> post-promotion health and browser smoke
```

## Documentation Rule

Do not add historical transcripts, architecture archives, temporary team plans, browser logs, or superseded doctrine here. Update this file when current status, accepted evidence, risks, or the immediate work order changes. Use Git history for prior checkpoints.

# Mycelis Active Development State

> Canonical implementation scoreboard. Product and architecture authority lives in [Mycelis Canonical PRD](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md). This file contains current delivery state only; completed history and superseded plans live in Git history.

## Active Snapshot

| Field | Current state |
| --- | --- |
| Updated | 2026-08-15 |
| Integration branch | `dev` |
| Active slice | Development data-plane and disk hygiene is `COMPLETE` on `dev`. P0.9 deterministic package delivery remains `ACTIVE`; P0.3a bounded discovery and Outcome Templates is `NEXT` as its clarification/configuration prerequisite. |
| Production branch | `main` remains behind the accepted `dev` integration train pending release certification. |
| Runtime posture | The development stack is currently stopped. Normal startup uses Dockerized PostgreSQL/pgvector and NATS as the reusable data plane with Core and Interface running locally from source; Ollama remains independently managed. Full Compose app containers, Kubernetes, and WSL remain explicit release-proof lanes. |
| Delivery target | Pass human-first Soma acceptance, accept the external-mutation recovery contract, then execute P0.12 release certification and intentional `dev -> main` promotion. |
| Main release risk | The headed delayed-work journey proves guidance reaches the active worker while Soma stays conversational, but the local `qwen3:8b` package worker still exhausts its bounded loop before required entrypoint readback and degrades honestly. Manual novice comprehension, external-mutation recovery acceptance, and exact-candidate deployment certification remain release gates; optional provider-specific media proof is not the critical path. |
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
| Development environment and persistence hygiene | `COMPLETE` | Docker pgvector is the sole development PostgreSQL server and stores relational plus vector state in one retained volume. Source Core resolves the configured published port, native host-server tasks are removed, cache guard checks both repository and system volumes, and shutdown now distinguishes retained dependencies from `--include-data-plane` without deleting volumes or claiming control of Ollama/host runtimes. Focused lifecycle/service-control, cache, DB, docs, Go MCP, UI, typecheck, storage-health, live-backend, and headed fresh-user proof cover this topology. | Preserve this topology without routine app-image rebuilds. |
| P0.3a bounded discovery and Outcome Templates | `NEXT` | The canonical PRD now defines a minimum sufficient brief, bounded clarification, reusable scoped Outcome Templates, version/digest snapshots, and one ConfigDocument pipeline shared by Soma and direct YAML/JSON. Existing WorkIntent, conversation-template, Worker Profile, capability, source, and governance contracts provide the integration seams. | Implement ConfigDocument validation/store/dry-run/activation, then compile the first Outcome Template into WorkIntent and prove conversational plus direct-file authoring. |
| P0.9 full Trusted Outcome Journey | `ACTIVE` | Governed ask through revisit remains proven. The human-first slice keeps the composer visible, prevents automatic drawers/runtime vocabulary/horizontal overflow, preserves a compact Work drawer, and collapses repeated machine updates. Retained outputs open in a dedicated canvas with exact Back-to-Soma context. Active run/team/work identity survives in the thread; explicit composer guidance is validated, retained atomically with mission evidence, queued idempotently, accepted through the team command channel, and translated to the active agent interjection channel instead of starting competing work. Focused Go/UI/typecheck/NATS tests pass. In headed Chromium, the acknowledgement returned within ten seconds and the active worker consumed the guidance during its existing tool loop. The broader app journey then degraded because the local model did not perform required structural readback; no false completion was shown. | Repair deterministic package readback completion, rerun the full headed journey, then run manual novice comprehension before release certification. |
| Media Outcome execution | `ACTIVE` | Direct Forge `/sdapi/v1/txt2img`, group-scoped media targeting, provider-specific recovery, broad Core tests, UI unit tests, typecheck/build, and headed Chromium preflight pass. With the running Forge UI returning `404` for `/sdapi/v1/options`, Soma now shows one image-generator setup alert and creates no proposal, approval, team, failed run, or internal planning deliverable. | Restart Forge with API mode (`--api`) enabled, then complete the headed natural Soma ask through generated image preview, saved group path, proof, and revisit. |
| P0.11 documentation convergence | `COMPLETE` | The canonical PRD contains all product architecture. The duplicate architecture overview and obsolete user-facing blueprint guide are deleted; user concepts, run/recovery, acceptance, deployment, API/provider, implementation, README, state, and in-app Help guidance match current Workspace/Outcome delivery and the Docker data-plane development lane. Documentation contracts, Help component tests, typecheck, and a headed real-manifest PRD-to-acceptance journey pass. | Keep one product-architecture authority and update owned support docs with behavior changes. |
| P0.13 cross-device operator acceptance | `COMPLETE` | Dashboard, Groups, Resources, Memory, Docs, and Settings have focused desktop/mobile proof. Exact owner-scoped creation and purge pass visibly in Chromium; 32 fixture scopes are purged, no non-terminal scopes remain, and active claims are zero. Headed natural delivery creates a complete browser package, validates its declared interaction, opens it directly, and purges its resources. A post-merge run that crossed five minutes exposed orphan result correlation; the corrected feature and merged `dev` runs passed in 5.8 and 5.5 minutes. The headed Trusted Outcome Journey also passes Ask through Revisit with retained proof and cleanup. Conversational control is merged and post-merge proven. The integrated fresh-user matrix starts from a stale work URL, signs into Soma, traverses Groups creation, Resources, Memory, Docs, Settings, and returns to Soma with zero page, console, hydration, `5xx`, or horizontal-overflow failures at desktop and compact widths. The adjacent 40-journey headed set passes after focused Groups history and Vault hydration reruns. The merged `dev` state repeats the four-journey headed fresh-user matrix and service health successfully. | Preserve through release certification. |
| P0.12 release hygiene and certification | `ACTIVE` | Clean Windows release preflight passes lint, Core, build, typecheck, Vitest, 121 Chromium journeys, source service health, and all four live governance scenarios. Configured Compose ports, shared identity-secret boundaries, and the Docker PostgreSQL/NATS plus local Core/Interface development topology are merged and re-proven on `dev` without app image builds. | Complete external-mutation operator acceptance, full Compose/Kubernetes certification, exact-candidate release preflight, and intentional promotion. |
| Shared NATS service boundary | `COMPLETE` | One Core process uses explicit `NATS_URL`; `MYCELIS_NATS_SERVICE_ID` names its runtime/observer clients on a shared broker. Registered external inputs own concrete `swarm.global.input.*` subjects, duplicate/wildcard claims fail closed, and existing source buffers protect teams from high-rate traffic. Live shared-broker registration/ingest isolation and post-merge Core/service health are proven. Routine app shutdown drains Mycelis clients and leaves shared broker state intact. | Preserve this boundary through P0.12 release certification and require an explicit bridge or separate Core deployment for another NATS host. |
| P0.7b declarative configuration authoring | `REQUIRED` | Worker Profiles, team manifests, Soma command manifests, MCP connections, Search Sources, registered inputs, and NATS buffering already have bounded runtime contracts, but users cannot yet manage every family through one file-authoritative Soma/direct pipeline. | After P0.3a and the package journey pass, extend the proven ConfigDocument path to profiles, actions, MCP/capabilities, sources, NATS channels, and service/schedule definitions with focused Resources editing and rollback. |
| Production promotion | `REQUIRED` | `main` is intentionally not promoted from the current `dev` train. | Merge certified `dev` into `main`, rerun health/browser smoke, update state, and synchronize remotes. |

## Immediate Work Order

1. Implement P0.3a slices A-B: ConfigDocument validation/store/digest/dry-run/atomic activation plus Outcome Template matching, bounded discovery, and Soma/direct-file authoring.
2. Make package-contract correction deterministic using the resolved brief so workers finish required writes, entrypoint readback, and validation within the bounded loop.
3. Rerun the full headed Ask -> Clarify -> Approve -> Execute -> Steer -> Deliver -> Open -> Revisit journey, including template reload/override/rollback and manual novice acceptance.
4. Extend the proven configuration path through P0.7b adapters, then review external-mutation recovery and execute P0.12 certification only if every release gate passes.
5. Exercise optional providers, including direct Forge media delivery, as capability-specific coverage without displacing the core release journey.

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
- Docker PostgreSQL/pgvector and NATS remain reusable development dependencies while Core and Interface run locally; lifecycle status resolves the configured PostgreSQL host port instead of assuming a native server.
- Repository cache cleanup cannot hide low space on the user/system volume, and it does not delete retained Docker volumes or unrelated user data.

### Operator Experience

- The authenticated default is the Soma threaded workspace with compact governance and Outcomes opened on demand.
- Proposal approval, cancellation, and revision stay in the Soma composer: bounded replies execute the matching lifecycle action, while qualified replies remain normal conversation. The default workspace has no action shelf, team/routing picker, mode selector, status cluster, or proposal button stack.
- Opening proposal Details scrolls the disclosed content into the bounded conversation viewport so inspection does not hide behind the composer.
- The composer remains reachable and expands with user input before bounded scrolling.
- Groups, Resources, Memory, Docs, and Settings use focused list/detail, tabs, or overlays rather than full-page content stacks.
- Authenticated desktop and compact-phone navigation reaches every primary operator surface through visible controls while retaining a reachable Soma composer and zero document-level horizontal overflow.
- Leaving Help while its manifest is still loading cannot pull the user back from their selected destination; the focused regression and post-merge headed navigation sweep pass.
- Output actions open retained files/packages directly; runtime IDs and transport details remain behind Details or Inspect.
- Current, Completed, and Archived work are distinct; retained history does not inflate active review counts.
- Default navigation speaks in user work (`Soma`, `Work`, `Resources`, `Help`), the first viewport contains one reachable composer and no automatic review dialog, and runtime vocabulary remains behind deliberate depth.
- Output review overlays Soma instead of narrowing it, uses the desktop application canvas or contained compact sheet, and presents one primary open action while paths, folders, replies, and proof remain under Details.

### Documentation And Quality

- `docs/architecture-library/MYCELIS_CANONICAL_PRD.md` is the single product architecture and PRD authority.
- Supporting Backend, Frontend, Operations, API, Testing, and user guides have bounded implementation or operator ownership.
- Superseded product doctrine is deleted, not archived in the active docs tree.
- The real in-app Help manifest opens the Canonical PRD and User Acceptance Runbook in headed Chromium and no longer exposes the deleted overview.
- Every behavior slice reviews README, this scoreboard, the owning docs, and the in-app Help manifest when applicable.

## Open Release Risks

| Risk | Status | Current evidence | Required resolution |
| --- | --- | --- | --- |
| Human-first default surface | `IN_REVIEW` | Untouched first-viewport, user-work navigation, one primary output action, dedicated retained-output canvas, exact Soma-context return contract, overlay stability, runtime-vocabulary exclusion, zero horizontal overflow, broad route navigation, and consolidated user-readable attention state pass focused and headed desktop/compact proof. | Prove delayed-work conversational steering and run manual novice comprehension. |
| Deterministic QA ownership | `COMPLETE` | Migrations `058`-`060`, opt-in root-admin API, exact creation provenance, reentrant creation/purge fencing, current-run staged-ownership allowance, durable pre-existence rejection, independently committed created-team/workspace claims, resumable purge, atomic claim release, and NATS exclusion are covered by Core, Python, UI proxy, and post-merge headed Chromium proof. Thirty-one scopes are purged with zero active claims or non-terminal scopes. | Preserve in every retained-state live journey and release certification. |
| Generated package delivery and correlation | `COMPLETE` | Worker correction requests one missing contract item at a time, entrypoint guidance forbids positional event selectors, and completion still requires retained-file readback plus structural and browser interaction validation. The command correlation lease exceeds the durable work recovery deadline; its six-minute regression and headed natural journeys pass on the feature branch and merged `dev` in 5.8 and 5.5 minutes. | Preserve in the fresh-user device matrix and release certification. |
| Ambiguous external mutation | `IN_REVIEW` | WorkIntent carries side-effect kind, retry safety, stable idempotency key, and observed side-effect state. Independent acceptance review found that the generic team-work action endpoint and UI could still offer recovery after an external outcome became unknown. Core now fails closed on start/pause/resume/recover while preserving steering/archive, and the UI directs verification through Soma without enabled retry. Focused Core, UI, typecheck, quality, service-health, and headed Chromium proof pass on the feature branch and merged `dev`. | Obtain operator acceptance of the verification-first language and controls, then preserve the boundary through release certification. |
| Windows Core restart wrapper | `IN_REVIEW` | Foreground wrapper timeout can still be confused with service failure. | Distinguish those states and retain useful diagnostics. |
| Deployment certification | `REQUIRED` | Source-mode and focused deployment proof pass. | WSL, Compose, Kubernetes, and operator-facing browser paths must pass from the committed release candidate. |
| Production promotion | `REQUIRED` | `dev` remains intentionally ahead of `main`. | Promote only after clean-tree release proof; do not mix new feature work into the candidate. |

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

# Mycelis Active Development State

> Canonical implementation scoreboard. Product and architecture authority lives in [Mycelis Canonical PRD](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md). This file contains current delivery state only; completed history and superseded plans live in Git history.

## Active Snapshot

| Field | Current state |
| --- | --- |
| Updated | 2026-08-22 |
| Integration branch | `dev` |
| Active slice | `feature/p0-7b-resources-worker-profile-authoring` adds a focused Resources Worker Profile authoring lane. Inline dry-run preview now uses the real ConfigDocument preview proxy, and save/activation continue through governed Soma handoff without starting agents, providers, sessions, or bus subscriptions. |
| Production branch | `main` contains the certified release candidate at `c1971894` and is clean locally, ahead of `origin/main` pending push. |
| Runtime posture | Dockerized PostgreSQL/pgvector and NATS are running as the reusable data plane with Core running locally from source; Interface is currently down outside browser proof. Ollama remains independently managed. Full Compose app containers, Kubernetes, and WSL remain explicit release-proof lanes. |
| Delivery target | Trusted Outcome Journey and exact release-candidate certification are complete, and local `dev -> main` promotion is complete. Current development has resumed on the deferred Worker Profile authoring lane as a separate feature branch. |
| Main release risk | Deterministic package generation and live opening now pass. Exact-candidate release and post-promotion browser certification gates are complete locally; current risk is keeping the new Worker Profile authoring slice bounded, tested, and out of `main` until feature and integration proof pass. |
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
| Development environment and persistence hygiene | `COMPLETE` | Docker pgvector is the sole development PostgreSQL server and stores relational plus vector state in one retained volume. Source Core resolves the configured published port, native host-server tasks are removed, cache guard checks both repository and system volumes, and shutdown distinguishes retained dependencies from `--include-data-plane` without deleting volumes or claiming control of Ollama/host runtimes. The Invoke surface is capped at 95 registered tasks after removing duplicate test aliases and the obsolete all-in-one deploy shortcut. Focused lifecycle/service-control, task-registration, docs, cache, DB, Go MCP, UI, typecheck, storage-health, live-backend, and headed fresh-user proof cover this topology. | Preserve this topology and the task budget without routine app-image rebuilds. |
| P0.3a bounded discovery and Outcome Templates | `COMPLETE` | A shared `mycelis.ai/v1` ConfigDocument contract provides strict YAML/JSON parsing, path-root guards, structured validation, canonical digest, immutable PostgreSQL revisions, kind-scoped activation history, explicit rollback, and dry-run. Outcome Templates compile a bounded minimum brief into WorkIntent with version/digest snapshot and template/operator/policy precedence. Config-backed Soma quotes select preview/store/activate behavior; direct answers cannot execute mutation tools. Focused Core, UI receipt, typecheck, migration, mocked headed, and authenticated visible live proof pass. The live journey saves, approves, activates, reloads, applies the exact record id/version/digest to work, then purges only the owned revision and restores prior activation state. Confirmation revalidates persisted request authority, rejects same-scope identity/digest substitution, serializes atomic receipt turns, and claims configuration ownership inside the existing transaction without self-deadlock. | Preserve this accepted contract while later ConfigDocument families extend it. |
| P0.9 full Trusted Outcome Journey | `IN_REVIEW` | Governed ask through revisit is proven. Contract correction resets only after successful evidence progress while retaining the hard 20-iteration bound. Project packages require current writes for required files and local dependencies, rendered validation markers, a usable interaction path across inline or referenced JavaScript, and a Core-owned exact readback of the latest entrypoint bytes before worker completion. Stale or divergent readback is rejected. The headed Chromium journey passes natural ask, approval, async NATS handoff, active-work steering, retained package generation, runtime browser validation, concise completion, safe embedded opening, visible Restart, observable Play behavior, proof, and fixture cleanup on both the feature branch and merged `dev`; the integrated run completed in 2.6 minutes. The dedicated output canvas also passes desktop and compact proof. | Run manual novice comprehension before release certification. |
| Media Outcome execution | `ACTIVE` | Direct Forge `/sdapi/v1/txt2img`, group-scoped media targeting, provider-specific recovery, broad Core tests, UI unit tests, typecheck/build, and headed Chromium preflight pass. With the running Forge UI returning `404` for `/sdapi/v1/options`, Soma now shows one image-generator setup alert and creates no proposal, approval, team, failed run, or internal planning deliverable. | Restart Forge with API mode (`--api`) enabled, then complete the headed natural Soma ask through generated image preview, saved group path, proof, and revisit. |
| P0.11 documentation convergence | `COMPLETE` | The canonical PRD contains all product architecture. The duplicate architecture overview and obsolete user-facing blueprint guide are deleted; user concepts, run/recovery, acceptance, deployment, API/provider, implementation, README, state, and in-app Help guidance match current Workspace/Outcome delivery and the Docker data-plane development lane. Documentation contracts, Help component tests, typecheck, and a headed real-manifest PRD-to-acceptance journey pass. | Keep one product-architecture authority and update owned support docs with behavior changes. |
| P0.13 cross-device operator acceptance | `COMPLETE` | Dashboard, Groups, Resources, Memory, Docs, and Settings have focused desktop/mobile proof. Exact owner-scoped creation and purge pass visibly in Chromium; 32 fixture scopes are purged, no non-terminal scopes remain, and active claims are zero. Headed natural delivery creates a complete browser package, validates its declared interaction, opens it directly, and purges its resources. A post-merge run that crossed five minutes exposed orphan result correlation; the corrected feature and merged `dev` runs passed in 5.8 and 5.5 minutes. The headed Trusted Outcome Journey also passes Ask through Revisit with retained proof and cleanup. Conversational control is merged and post-merge proven. The integrated fresh-user matrix starts from a stale work URL, signs into Soma, traverses Groups creation, Resources, Memory, Docs, Settings, and returns to Soma with zero page, console, hydration, `5xx`, or horizontal-overflow failures at desktop and compact widths. The adjacent 40-journey headed set passes after focused Groups history and Vault hydration reruns. The merged `dev` state repeats the four-journey headed fresh-user matrix and service health successfully. | Preserve through release certification. |
| P0.12 release hygiene and certification | `COMPLETE` | Prior Windows release proof passes lint, Core, build, typecheck, Vitest, 121 Chromium journeys, source service health, and all four live governance scenarios. The release-candidate remediation then closed stale recovery, retained Playwright lock, obsolete browser expectations, false-clean process inventory, review-rail Resources access, duplicate accessible output links, and parallel Core package validation risks. Exact `dev` release preflight passes clean-tree, runtime posture, lint, serial Core sweep, production Interface build/typecheck/Vitest, 132 mocked Chromium journeys with 36 live-only skips, Docker PostgreSQL/NATS/Core/Ollama service health, and all four live governed Soma scenarios. `dev` was promoted into `main` as `c1971894`; promoted `main` passed docs-link proof, service health, and headed desktop/compact Soma human-first browser smoke. | Push/sync remotes when requested, then keep this gate as the promotion standard for future release candidates. |
| Shared NATS service boundary | `COMPLETE` | One Core process uses explicit `NATS_URL`; `MYCELIS_NATS_SERVICE_ID` names its runtime/observer clients on a shared broker. Registered external inputs own concrete `swarm.global.input.*` subjects, duplicate/wildcard claims fail closed, and existing source buffers protect teams from high-rate traffic. Live shared-broker registration/ingest isolation and post-merge Core/service health are proven. Routine app shutdown drains Mycelis clients and leaves shared broker state intact. | Preserve this boundary for future release candidates and require an explicit bridge or separate Core deployment for another NATS host. |
| P0.7b declarative configuration authoring | `IN_REVIEW` | D1 is integrated with strict Worker Profile ConfigDocument validation and immutable preview snapshots. Worker Profiles are inert catalogue templates; activation means eligible for future selection and performs no agent, provider/API, session, backend connection, or NATS startup. D2a resolves only profiles selected by approved team creation from the activation index in operator, workspace, organization, then built-in order; only the trusted confirmed request boundary supplies scope. `create_team` is the sole profile-to-runtime boundary and pins the selected profile/provider/backend/capability lineage before starting member runtime resources. D2b adds an authoritative PostgreSQL runtime-team manifest with schema version and canonical digest, exact full-manifest Core restart restoration, idempotent identical save, conflicting replacement rejection, and durable stop/mission deactivation. Direct HTTP callers cannot inject profile refs or snapshots; only Soma's governed resolver establishes lineage. Focused Core tests and the headed live A/B/rollback/A2/navigation re-fetch/cleanup journey pass on rebuilt merged `dev`; a regression proves unselected team members perform zero catalogue resolution. Resources and Teams keep ready-made profiles inspect-only and creation/customization Soma-first. Current branch adds the focused Resources authoring journey: inline dry-run validation plus Soma-governed save/activate handoff without direct UI mutation. Focused Catalogue Vitest, interface typecheck, docs links, and headed Chromium catalogue proof pass. | Commit the feature branch, merge to `dev`, rerun affected integration proof, and clean the branch after successful integration. |
| Production promotion | `COMPLETE` | `main` contains promoted release candidate `c1971894` from certified `dev`. Promoted `main` passed docs-link proof, service health, and headed Soma human-first desktop/compact browser smoke. | Push/synchronize remotes when requested and use the same gate for the next candidate. |

## Immediate Work Order

1. Complete the Worker Profile authoring branch with inline ConfigDocument dry-run, governed Soma handoff, focused UI tests, docs, and state.
2. Prove the feature branch, merge to `dev`, rerun affected integration proof, then clean the branch.
3. Resume broader deployment certification and optional provider-specific media coverage only after this slice is clean.

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
- ConfigDocument revisions are immutable, activation identity includes kind and scope, and explicit rollback preserves actor/audit history while compiled work snapshots the selected template version and digest.
- A read-only Outcome Template preview is deterministic after Soma or the operator supplies a parseable document: the real validator runs, `preview_config_document` is reported, and no proposal, revision, activation, or output artifact is created. Direct-answer inference cannot expose or execute mutation-capable tools.
- A retained Outcome Template can be saved and activated through compact governed Soma turns, survives reload, and applies the exact stored record id, version, and digest to WorkIntent. Confirmation rejects wrong identities, forged digests, foreign request boundaries, and concurrent receipt reordering. QA cleanup owns only the scoped revision, restores any previous activation, and leaves no stale configuration or fixture claim.
- Runtime-created teams persist their complete effective manifest before acknowledgement. Digest validation restores exact member prompts, routing, model/provider, tools, context bindings, verification, schedule, and Worker Profile lineage after restart; current activation is not re-resolved. Identical save is idempotent, conflicting same-id replacement and direct-HTTP profile injection fail closed, and durable stop or mission deactivation prevents restart resurrection.
- Docker PostgreSQL/pgvector and NATS remain reusable development dependencies while app services run locally only when needed; lifecycle status resolves the configured PostgreSQL host port instead of assuming a native server.
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
| Generated package delivery and correlation | `IN_REVIEW` | Worker correction requests one missing contract item at a time and resets its correction allowance only after new successful evidence. Completion requires required-file and local-dependency writes, rendered validation markers, exact latest-entrypoint readback, and digest-bound browser interaction proof. The merged `dev` headed proof completes steering, package delivery, embedded opening, and visible state change in 2.6 minutes; prior command-correlation proof remains accepted. | Complete manual novice acceptance, then preserve the journey in release certification. |
| Ambiguous external mutation | `IN_REVIEW` | WorkIntent records a typed operator verification with server-attributed actor, timestamp, and event provenance plus operator observation and optional evidence references. `committed` closes without replay; `not_committed` requires a new governed Soma proposal; `still_unknown` remains degraded. Delayed results cannot overwrite the verified posture. Generic start, pause, resume, and recovery stay blocked while steering and archive remain available. Focused protocol/API persistence and projection tests, Core build, 16 UI tests, typecheck, docs links, line policy, service health, and 7/7 headed Teams journeys pass on both the feature branch and merged `dev`. | Obtain operator acceptance and preserve the boundary through exact-candidate release certification. |
| Windows Core restart wrapper | `IN_REVIEW` | Foreground wrapper timeout can still be confused with service failure. | Distinguish those states and retain useful diagnostics. |
| Deployment certification | `REQUIRED` | Source-mode and focused deployment proof pass. | WSL, Compose, Kubernetes, and operator-facing browser paths must pass from the committed release candidate. |
| Production promotion | `COMPLETE` | `main` and `dev` share promoted local tip `c1971894` after the certified release candidate was merged and post-promotion health/browser smoke passed. | Push/synchronize remotes when requested; do not mix new feature work into this candidate. |

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

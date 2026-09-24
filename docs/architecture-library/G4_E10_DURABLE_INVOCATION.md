# G4/E10 — Minimum Durable Invocation Acceptance Contract

Status: `COMPLETE` for the bounded counting contract. Frozen 2026-09-23; independent architecture QA `/root/authority_audit`: GO; independent implementation and delivered-runtime QA `/root/recovery_qa`: GO. Current delivery evidence is in the [scoreboard](../../.state/V8_DEV_STATE.md).
Owner: Core authority/execution. Product authority remains the [canonical PRD](MYCELIS_CANONICAL_PRD.md).
Baseline: clean `dev` `7b36e734b14ff70239c2947713fd96150f5a7e6a`, authenticated fetch 2026-09-23.

## Decision and bounded delivery

The operator's G4/E10 goal authorizes this missing acceptance packet and places this prerequisite before further effectful framework execution. No separate Platform Delivery Plan, accepted gateway packet, or E10 definition exists in this baseline; this document must not imply otherwise. Existing PRD unknown-effect semantics remain binding. P0.10 B2 remains the next framework-specific slice, not a substitute for effect ownership.

Deliver one non-secret counting capability through a real Core invocation API and PostgreSQL authority boundary. A model tool-call parser produces the same typed proposal; it cannot execute or supply trusted identity. This certifies that representative path, not all existing tools or provider-native tool calling. No broad UI, new queue, broker consumer, model engine, secret service, or approval system.

The invocation ledger is the durable receipt/ownership extension of existing execution contracts. It is not another dispatch queue: this slice executes synchronously after durable admission. Existing outbox remains the only dispatch queue. Future outbox consumers must reference invocation IDs and obey this contract.

## Identity and resource authority

- HTTP acting user comes only from Core-authenticated request context. Ignore no unknown authority fields: reject supplied subject/user/agent/parent/delegation claims in the proposal schema.
- Resolve the existing `users`, `accounts`, `groups`, `org_memberships`, `roles`, and `role_permissions` records in PostgreSQL. API-key possession, forwarded role, wildcard route scope, or provider key alone cannot admit an effect.
- Initial subject resolution accepts only an authenticated UUID `RequestIdentity.UserID` that exactly matches a provisioned `users.id`; its account comes from that row, never the request. No email, username, external-subject fallback or auto-registration. Non-UUID forwarded subjects are denied until separately accepted external identity mapping exists. Local admin (including the default zero UUID) still needs a real active user/account and matching membership. Group and role joins must match that account (an explicitly system-scoped role may have no account); cross-account or ambiguous membership fails closed.
- The initial resource is an existing identity group within its account; require an active, unexpired group membership and the explicit counting permission. Identity group and collaboration workspace are different existing concepts: do not invent equivalence. Workspace/run/contract references, if present, must be server-derived from the accepted execution contract; absent values remain absent.
- User/group/account are mandatory. Agent, parent-agent and delegated execution are unsupported in this slice and fail closed rather than accepting self-reported lineage. Their later admission must use existing runtime team manifests and Core-minted lineage.
- Current-authority revision hashes the locked account/user/group/membership/role/positive-permission evidence, status, UTC-normalized expiry and each row's PostgreSQL [`xmin`](https://www.postgresql.org/docs/16/ddl-system-columns.html) opaque version token during the one-hour grant lifetime. Compare tokens only for equality, never global ordering or identity. Any row update, including disable then restore, conservatively requires fresh acceptance; this does not rely on writers updating timestamps. Database restore/version changes may invalidate outstanding grants safely; historical receipts remain inspectable. Record the snapshot with the invocation. Memory, prompts and cached capability lists confer nothing.

## Accepted immutable grant

A grant is an immutable per-effect authorization snapshot linked to an existing accepted execution contract and intent proof, not a new approval workflow. Minting is a Core operation after existing confirmation, never a public client-supplied grant document. The proof's operator/resource/action/input must match; an unrelated confirmed contract cannot authorize counting.

Required evidence: Core UUID, revision/digest, existing contract/proof IDs, acting user/account/group, capability ID, pinned binding digest, exact normalized input digest, membership/authority digest at acceptance, one unit per invocation, total allowed units, validity deadline, and explicit absent delegation. Grant row content is immutable; revocation is separate mutable state. Replacement mints a new grant; old evidence survives. Changing authority, grant content, scope, expiry, input or binding requires fresh acceptance. An old grant ID/digest cannot silently select its replacement.

Acceptance write path: a bounded counting-proposal handler receives only group, typed counting arguments and bounded requested units. It resolves authenticated identity/current authority and the registry pin, then writes a server-owned `invocation_boundary` inside the existing `intent_proofs.scope_validation`: version, resolved user/account/group/membership, permission, capability, pin, normalized arguments/digest, authority digest, unit budget and expiry. Persist proof, existing execution contract and existing confirm token atomically. The model parser can provide the same proposal fields, never this boundary or a confirm token.

Existing `POST /api/v1/intent/confirm-action` recognizes this boundary before legacy tool dispatch. Its counting branch re-resolves the authenticated subject, locks authority rows then proof/contract/token, compares all boundary fields with current authority and pin, and requires pending proof, exact action, matching same-subject scope, valid token and both expiries. Consume the token using conditional update with exactly one affected row. Mark proof/contract accepted and insert one immutable grant in the same transaction, protected by a unique proof reference. Confirmation returns the grant reference; it does not execute. Unsupported or malformed boundaries fail closed and cannot fall through to legacy execution. Grant minting is not exposed as a separate approval endpoint.

The isolated fixture provisions existing identity and registry rows, then uses this real proposal → existing confirmation → invocation path; it may not fabricate accepted grant/proof rows as proof of approval. Duplicate confirmation cannot mint another grant. An unrelated accepted proof, changed JSON, wrong subject, expired token, double consumption and missing boundary are negative cases.

## Capability and binding

Reuse `capability_manifests`; no second registry. The counting capability has one concrete non-secret HTTP adapter, fixed request schema and bounded timeout. Its binding includes capability ID, adapter name/version, configured endpoint, method, schema version and unit cost. Persist a canonical binding snapshot/digest in the immutable grant and invocation.

The endpoint is operator-configured, never model/request supplied. Validate scheme/host and disable redirects; no arbitrary URL tool. The fixture must explicitly be enabled and is off by default. Current manifest must exist and be enabled. Admission and execution-start compare the current registry binding with the pin under a database lock; missing, disabled or changed binding fails closed. Registry refresh may not silently erase or retarget an in-flight snapshot. Historical snapshot survives registry replacement.

Projection owner is the existing `capabilities.Service.derive/Refresh`: when configured, it includes the counting manifest with `metadata.invocation_binding` containing endpoint, adapter/version, method, schema and unit cost. Its existing snapshot upsert/delete SQL contends on the same manifest row locked by admission/execution-start. Disabling configuration removes/disables the current manifest and prevents new execution-start. An old process configuration is not consulted during execution: the adapter uses only the validated immutable binding read from the invocation. A changed endpoint must first be committed in the registry and requires fresh acceptance; it cannot change an already executing call's destination. No standalone binding registry.

## Admission, budget and durable record

One transaction checks current authority, accepted grant/digest, capability/binding, exact input, deadline and limits; reserves one unit; inserts the invocation and initial audit record; then commits before any adapter call or credential resolution.

Lock order: existing account → user → group → membership → role/permission → proof/contract/token (when confirming) → grant/revocation/budget → capability manifest → invocation. Use consistent ordering for multiple rows. Existing SQL updates/deletes must contend on locked rows; do not rely on an advisory lock that existing writers ignore. Positive permission rows are locked. Missing permissions fail closed. A reservation is an atomic conditional increment under the grant row lock, with a unique invocation reference. No read-then-write budget checks. One admission consumes one unit, including uncertain effects; do not automatically refund.

Invocation fields: UUID; server-authenticated user/account/group; optional existing contract/run/Outcome references; immutable grant ID/digest/snapshot; current-authority digest/snapshot; capability ID and binding digest/snapshot; normalized input and digest; scoped idempotency key; reservation units/reference; owner token and monotonic fencing generation; attempt; state; created/updated/lease/execution/result times; sanitized result/effect observation; reconciliation evidence. Credential references only, never secret values.

Uniqueness is `(grant_id, idempotency_key)`. Same key and exact request returns the existing scoped receipt without a new reservation/effect; changed input, binding or subject returns conflict/denial. Receipt lookup still authenticates scope. Revocation forbids new admission, but does not erase readable historical evidence. Invocation ID is not a NATS transport event ID.

## States and ownership

| State | Meaning / allowed next step |
| --- | --- |
| `ready` | Admission and reservation atomically committed; also represents admitted/reserved. No effect attempted. |
| `claimed` | Owner token/generation and bounded lease acquired; still no effect permitted. |
| `executing` | Durable point of no automatic replay committed before network call. Current authority, expiry and binding rechecked. |
| `observed` | Adapter response durably recorded; response alone is not trusted Outcome completion. |
| `failed_before_effect` | Validation, revocation or pre-call failure proves no call; terminal receipt. |
| `failed_known_no_effect` | Adapter-specific evidence proves rejection without effect; terminal, no automatic new invocation. |
| `unknown_effect` | Call may have happened; no automated retry/reclaim execution. |
| `verified` | Counting evidence confirms the observed effect; no further execution. |
| `reconciled` | Authorized evidence records committed/not-committed; never reopens execution. |
| `cancelled` | Explicit cancellation/revocation before executing; reservation retained. |

Claim uses compare-and-swap and a new unpredictable owner token plus generation. Only the current unexpired owner can cross `claimed → executing`. An expired `claimed` lease may be reclaimed; stale owners cannot pass this transition. Recheck current authority in that transition transaction. PostgreSQL `clock_timestamp()` after locks governs membership/grant/token/lease expiry, including waits beyond deadlines; transaction-start time and host clocks are insufficient. Revocation racing execution serializes with this boundary: if revocation wins, no call; if execution-start wins, the effect is already authorized/in flight and cannot be promised cancelled.

No transaction spans the external call. After `executing` commits, any crash, timeout, response loss, unexpected adapter error, cancelled context, or expired lease without durable observation becomes `unknown_effect`, even if the process actually crashed before sending. Conservative uncertainty is required. Recovery cannot reclaim an executing invocation for execution. An unexpired owner may commit `executing → observed`; after expiry or transition to unknown, its late response is append-only audit evidence and leaves the state unknown until authorized reconciliation. Stale generations cannot update receipts. No late evidence resets execution or supersedes an existing reconciliation.

Restart opens the same ledger. Ready/unattempted claims can progress; executing rows are conservatively reconciled as unknown. Broker redelivery or repeated HTTP requests only consult this ledger. A process-local mutex is insufficient.

## Result, reconciliation and credentials

Counting fixture exposes increment and read-only count/effect evidence keyed by invocation ID. The fixture deliberately does not deduplicate increments, so duplicate calls are measurable. Response-loss fault increments then drops the response. The adapter never retries HTTP calls. Evidence records an effect count and invocation association, without secrets.

Unknown remains unknown until scoped authorized reconciliation supplies attributable evidence. An `observed` response that lacks independent counting proof uses the same reconciliation path. `committed` closes as reconciled; `not_committed` also closes without re-execution; `still_unknown` preserves unknown. A new attempt after verified non-effect requires a new accepted grant/proposal, not resetting the old receipt. Reconciliation checks identity/resource permission and owner, and records actor/time/evidence durably. It cannot refund budget or call the adapter.

Credential resolution hook, where used later, runs only after admission and execution-start authorization. This fixture has no credentials. Full credential/process isolation is a separate packet; do not certify it here. Provider identity never substitutes for Mycelis acting subject.

## Explicit bypass inventory and future gates

Only the counting capability is admitted by this packet. It must not be registered in a legacy executor as an alternate callable tool. Tests attempt direct/tool-name/model-field bypasses. The following existing paths are outside this certification; their owners must integrate this ledger before claiming G4/E10 coverage. This is not permission to call them through the counting adapter.

| Existing path / source | Disposition and owner / gate |
| --- | --- |
| New counting API and typed model proposal | In scope: identical Core admission, reservation, execution and receipt path. |
| Confirmed actions: `server/templates_execution.go`, `confirmed_action_dispatch.go` | Existing approval source reused for grant evidence; general tool-batch dispatch deferred to Core E10 adapter integration. Existing batch retry is not certified safe for external effects. |
| Direct MCP: `server/mcp.go`, `mcp/pool.go` | Deferred: MCP owner must route credentialed calls through ledger; discovery grants nothing. Counting binding unavailable here. |
| Host action API: `server/host_actions.go` | Deferred: host capability owner, same admission gate before broader tool certification. |
| Agent/internal/MCP: `swarm/agent_tool_execution.go`, `tool_executor.go` | Deferred: swarm owner must bind trusted agent lineage and each effect; empty allowlist is not authorization. |
| Internal send/write/command/media/config/team/NATS tools | Deferred by concrete effect to Core capability owners; existing local/config transactions retain their contracts, not E10 external-effect certification. |
| Runtime fallback: `agent_result_contract_runtime_fallback.go` | Deferred with swarm; cannot discover or call counting adapter directly. |
| Scheduler/trigger: `schedule_trigger_scheduler.go`, `triggers/engine.go` | Proposal/run creation is not effect permission. Automation owner must obtain accepted grant before ledger admission. |
| Framework/Runs executor and reclaim | External create remains non-dispatchable. P0.10 B2/C must carry invocation references; journal retry is not effect retry permission. |
| Outbox reclaim/redelivery: `dispatchoutbox/store.go` | Existing queue unchanged; future consumers must use the same ledger. This slice has no second queue or NATS effect consumer. |

## Persistence, recovery and proof gates

Preserve the single fresh-install schema contract. Add the bounded schema extension to the canonical baseline with explicit integrity coverage and an owned, additive retained-schema upgrade path; never rerun historical baseline against populated databases. Upgrade must be transactional, idempotent and preserve existing rows. Historical source-block hashes remain immutable. No destructive down migration: roll back runtime with retained tables/receipts and deny unsupported execution.

Persistence changes require the canonical isolated `uv run inv lifecycle.first-boot-proof`. Extend its isolation where needed; never reset the recovered Compose database, volumes, workspace or shared NATS. An isolated disposable fixture must prove empty boot, schema upgrade/idempotence and restart, then leave the user's services intact.

Required owned tests: real PostgreSQL, positive and denied authority, stale/revoked grants, forged subject, wrong group, contract/proof mismatch, expiry, binding drift, duplicate key conflict, concurrent budget race, duplicate workers, stale lease tokens, restart, crash before execution (zero), crash after executing (unknown), response loss (one count/no replay), late response, unknown reconciliation, direct bypass and typed model-proposal parsing. Run race-enabled Go tests. No SQL mocks as persistence proof; no model inference response as invocation authority proof.

Independent architecture QA checks the frozen contract before implementation. Independent implementation QA inspects source, migrations, bypass classification and measured counts before final GO. Record exact commands/results in the scoreboard; no inference from old recovery tests.

Embeddings: `DEFERRED: E03/E04 provider/data-boundary certification`.
Broader MCP/host/swarm/framework integration, provider-native tool-call support, credential isolation and E12/E07 intervention UI remain separate gates. No claim of universal effect protection until those paths are integrated.

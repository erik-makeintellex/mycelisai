# Gatekeeper Logic (Governance Guard)

## Overview
The Gatekeeper is the governance `Guard` (`core/internal/governance/guard.go`) run by the Core `Router` (`core/internal/router/router.go`). It evaluates bus envelopes against `core/config/policy.yaml` (see `docs/governance.md` and the Governance Policy rows in `docs/API_REFERENCE.md`).

The Router subscribes to `protocol.TopicSwarmWild` as an observer. NATS has already delivered every message to its direct subscribers, so the Gatekeeper decides only whether Core reacts to a message. It can neither hold nor release one.

## The Intercept Pipeline
`Guard.Intercept(msg) (proceed bool, action string)`:

| Policy action | Core behavior |
| --- | --- |
| `ALLOW` | Proceed: registry refresh for canonical heartbeats, then the `swarm.audit.trace` re-publish for the Archivist. |
| `DENY` | Do not proceed. Logged. |
| `REQUIRE_APPROVAL` | Do not proceed. The Router writes one `policy_approval_required_observed` audit record (team, source agent, intent, subject, message id; `result_status=not_acted_on`), or logs `audit_unavailable` when the audit store is down. |
| unknown | Fail closed as `DENY`. |

A nil or degraded Guard (no loaded policy) denies everything; the Router only refreshes already-registered agents from canonical heartbeats.

## No in-memory approval queue
REQUIRE_APPROVAL does not park anything (C2-RETIRE, 2026-09-29, owner decision: retire with no alias). The former in-memory buffer, its list and resolve routes and legacy admin alias, the approval-needed bus publish and the re-publish of "approved" messages to a reconstructed `*.output` subject were removed. Approving a message the bus had already delivered could only report success that did not happen.

Real approval stays on the durable path: Soma proposals, confirm-action with approver tiers (`approvals:decide` for tier 2), and proposal approve/reject.

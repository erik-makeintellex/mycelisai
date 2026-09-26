# Governance & Policy System
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

This document describes the current release governance model, not just the older router-guard pattern.

## Current Release Truth

Mycelis now governs actions through three linked layers:

1. User governance profile
   - role
   - cost sensitivity
   - review strictness
   - automation tolerance
   - escalation preference
2. Capability-aware approval policy
   - low-risk actions can be auto-allowed
   - medium-risk actions can stay optional
   - high-risk or explicitly bounded actions require approval
3. Audit lineage
   - proposal generated
   - proposal confirmed or cancelled
   - execution run created
   - capability used
   - artifact created
   - channel write recorded

This governance model also applies to durable context loading:
- user-uploaded private records, diary entries, finance notes, and other sensitive references belong in `user_private_context` with private/restricted defaults and explicit target goal sets
- customer-provided deployment material belongs in the separate `customer_context` pgvector lane
- approved company-authored guidance belongs in `company_knowledge`
- admin-authored shared Soma guidance belongs in `soma_operating_context`
- team-shared execution memory belongs in scoped `AGENT_MEMORY`; it is not implicit organization doctrine
- distilled lessons, inferred patterns, contradictions, user-trajectory shifts, and meta-observations start as classified Managed Exchange `LearningCandidate` items, then promote into `reflection_synthesis` only after confidence and review posture are explicit
- external research loaded into either lane must stay explicitly classified by source kind, trust class, and sensitivity posture

Current release boundary note:
- this is still a free-node governance foundation, not a full multi-user IAM system
- the target multi-user contract is one shared organization-owned Soma persona plus many governed human principals, not one unrelated Soma per end user
- the current People & Access surface now makes the edition/identity story reviewable without pretending the full enterprise control plane already exists: operators can inspect self-hosted release, self-hosted enterprise, or hosted admin control plane posture alongside identity mode and shared Soma output-specificity ownership
- future enterprise identity should support federation and local break-glass admins without bypassing governance or audit
- future shared-Soma governance must also distinguish admin-shaped organization-level Soma context from ordinary user interaction context
- future shared-Soma governance must reserve durable agent/output-specificity assignment to the root admin or explicitly delegated environment owner rather than letting ordinary user chats redefine shared behavior

Trusted memory rule:
- recalled memory is a candidate input, not automatic truth
- deterministic logs, turns, artifacts, and approved governed doctrine outrank lower-order recalled memory
- team-shared memory may guide execution inside team scope, but it must not silently override newer organization doctrine
- Soma personal continuity may shape local style and relationship memory, but it must not outrank governed policy or verified evidence

The normal operator-facing outcomes are:
- `answer`
- `proposal`
- `execution_result`
- `blocker`

Mutation-capable work should enter `proposal` mode before execution.

## Approval Model

Every governed action can carry:
- `approval_required`
- `approval_reason`
- `approval_mode`
- threshold context for cost, capability risk, and external data use

Who confirms a `required` approval (A2b, judged at confirm time on the stored server-authored proposal, so no client field can lower it):

| Tier | When | Who confirms |
|---|---|---|
| 0 | approval not required | the proposer |
| 1 (self-review) | external data use, escalation preference, medium capability risk, estimated cost up to 5.0 | the proposer (the principal that proposed it) or an approver |
| 2 (approver) | posture/policy raised, high or critical capability risk, or estimated cost above 5.0 | a root admin with `approvals:decide`; an admin may approve their own proposal, audited `self_approved=true` |

The proposal already shows "needs admin approval" for tier 2. Every confirm audit records `approval_authority` (`proposer` or `approvals:decide`), `approval_tier`, and `self_approved`. Council chat applies the same posture floor and tiers as Soma chat; template work the council cannot resolve is refused and nothing is minted. Posture binds to a referenced template id, so work that names no template gets capability tiers only.

Current release posture:
- capability risk drives the baseline expectation
- user governance profile shapes how strict the approval policy becomes
- confirm/cancel is explicit and inspectable
- approval and execution decisions are audit-linked
- `user_private_context` loading is treated as high-risk because it can contain sensitive personal or business records even when the operator wants Soma to use it
- `company_knowledge` loading is stricter than `customer_context` loading because it promotes durable company reference material
- `soma_operating_context` loading is stricter again because it can shape shared Soma identity, stance, and output specificity across users
- `reflection_synthesis` loading is also treated as high-risk because it can encode sensitive meta-observations about user trajectory, contradictions, and changing work patterns
- Managed Exchange learning-candidate capture is not the same as memory mutation; promotion from a candidate into durable memory remains the stricter governed step

When conflict is detected:
- factual or event-history conflicts should defer to deterministic evidence
- policy or shared-behavior conflicts should defer to governed organization memory
- lower-order recalled memory should halt mutation and raise review or proposal instead of silently winning

## Audit Model

The base audit system is inspect-only in the operator UI.

Operators should expect to see:
- recent actions
- execution status
- approvals
- capability usage

Raw backend logs are not the default operator surface. The default surface is the normalized Activity Log / Audit view.

## Legacy Router Guard Note

The lower-level governance guard still exists for message/policy enforcement, but it is no longer the whole product story by itself.

Release review should treat governance as:
- policy enforcement
- proposal/approval flow
- capability risk mapping
- audit visibility

not only as a raw allow/deny/intercept subsystem.

Governance authority (A2a):
- reading the policy or pending Guard approvals requires a root admin with `governance:read`; replacing the policy requires `governance:write`; approving or rejecting a Guard-parked message requires `approvals:decide` (an admin `*` covers all). Anonymous callers get 401 and other principals 403. See the Governance Policy rows in `docs/API_REFERENCE.md`
- policy replacement and approval decisions are audit-first (`governance_policy_update`, `governance_approval_resolved`); if the audit row cannot be written, nothing changes
- every policy action must be exactly `ALLOW`, `DENY`, or `REQUIRE_APPROVAL`, and an `ALLOW` default needs at least one restricting rule; an empty, typo'd, or allow-only policy is invalid (PUT returns 400)
- if `core/config/policy.yaml` is missing or invalid (including empty), Core still starts but governance is degraded and fails closed: the Gatekeeper denies everything except heartbeats, posture-shaped work requires approval, and `/api/v1/services/status` shows `governance: degraded`. Fix the file and restart Core, or PUT a valid policy as an admin to recover without a restart
- the Gatekeeper observes the NATS bus: a DENY stops Core from reacting (registry, audit trace, approval request), not delivery to direct subscribers
- a policy PUT persists to the policy file inside the running Core; a redeploy restores the shipped policy
- confirming a tier-2 proposal (see Approval Model) requires a root admin with `approvals:decide`; anyone else sees a "needs admin approval" blocker, nothing runs, and the proposal stays confirmable by an admin. Tier 0/1 chat and blueprint proposals are confirmable only by their proposer or an approver (A2b)
- confirm tokens are single-use and durably purpose-bound (A2b: `purpose`, `binding_digest`, `minted_by` recorded at mint): a token only works on the path it was issued for (chat proposal, blueprint commit, one group operation, or an invocation), a blueprint commit must send exactly the negotiated blueprint, and a wrong-path, edited, or other-principal attempt is rejected without using the token up. Tokens minted before A2b are refused; propose again
- while governance is degraded, only a canonical heartbeat from an already-registered agent refreshes the registry; unknown agents appear after the policy is restored. Policy PUT audits `previous_digest` under the replacement lock
- still open: an approver queue in the Review Inbox, and operator-role approvers

## Operator Guidance

- treat mutating and external work as governed by default
- use `Automations -> Approvals` for inspect-only approval and audit review
- use the live governed browser proof when release work changes proposal, confirm, or execution behavior

Related references:
- `docs/user/governance-trust.md`
- `docs/licensing.md`
- `docs/architecture-library/MYCELIS_CANONICAL_PRD.md`

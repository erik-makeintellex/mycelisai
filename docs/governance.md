# Governance & Policy System
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

> 🔙 **Navigation**: [Back to README](../README.md)

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
- distilled lessons, inferred patterns, contradictions, user-trajectory shifts, and meta-observations start as classified Managed Exchange `LearningCandidate` items, then promote into `reflection_synthesis` only after confidence and review posture are explicit
- external research loaded into either lane must stay explicitly classified by source kind, trust class, and sensitivity posture

Current release boundary note:
- this is still a free-node governance foundation, not a full multi-user IAM system
- the target multi-user contract is one shared organization-owned Soma persona plus many governed human principals, not one unrelated Soma per end user
- the current People & Access surface now makes the edition/identity story reviewable without pretending the full enterprise control plane already exists: operators can inspect self-hosted release, self-hosted enterprise, or hosted-control-plane posture alongside identity mode and shared Soma output-specificity ownership
- future enterprise identity should support federation and local break-glass admins without bypassing governance or audit
- future shared-Soma governance must also distinguish admin-shaped organization-level Soma context from ordinary user interaction context
- future shared-Soma governance must reserve durable agent/output-specificity assignment to the root admin or explicitly delegated environment owner rather than letting ordinary user chats redefine shared behavior

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

## Operator Guidance

- treat mutating and external work as governed by default
- use `Automations -> Approvals` for inspect-only approval and audit review
- use the live governed browser proof when release work changes proposal, confirm, or execution behavior

Related references:
- `docs/user/governance-trust.md`
- `docs/licensing.md`
- `docs/architecture-library/V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md`
- `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`
- `docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md`

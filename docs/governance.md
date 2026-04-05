# Governance & Policy System

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
- customer-provided deployment material belongs in the separate `customer_context` pgvector lane
- approved company-authored guidance belongs in `company_knowledge`
- external research loaded into either lane must stay explicitly classified by source kind, trust class, and sensitivity posture

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
- `company_knowledge` loading is stricter than `customer_context` loading because it promotes durable company reference material

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
- `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`
- `docs/architecture-library/V8_RELEASE_PLATFORM_REVIEW_SECURITY_MONITORING_DEBUG.md`

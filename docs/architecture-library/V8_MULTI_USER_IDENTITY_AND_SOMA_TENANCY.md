# V8 Multi-User Identity And Soma Tenancy
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-04-05
> Purpose: Define the target identity, user-management, and shared-Soma contract for multi-user Mycelis deployments.

## TOC

- [Why This Exists](#why-this-exists)
- [Core Position](#core-position)
- [Identity Layers](#identity-layers)
- [Authentication Modes](#authentication-modes)
- [Administrative Access](#administrative-access)
- [Soma Tenancy Model](#soma-tenancy-model)
- [Admin-Shaped Soma Context](#admin-shaped-soma-context)
- [Memory And RAG Access Model](#memory-and-rag-access-model)
- [Governance And Audit Requirements](#governance-and-audit-requirements)
- [Hosted vs Self-Hosted Layering](#hosted-vs-self-hosted-layering)
- [Release Posture](#release-posture)
- [Rollout Sequence](#rollout-sequence)

## Why This Exists

The free-node governance foundation is single-user first, but the target product must support multi-user organizations without losing the one-Soma operating model.

This document makes three things explicit:
- external enterprise identity must be supported through standard federation
- self-hosted environments must retain a local administrative recovery path
- multiple human users must not fragment Soma into separate uncontrolled personas

## Core Position

Mycelis should support multiple authenticated human users, but each organization still operates through one governed Soma persona.

That means:
- users are distinct principals with their own identity, access, approvals, and privacy boundaries
- the organization owns one configured Soma identity, including Soma's name and operating posture
- private user interactions with Soma stay private to the authorized audience, even though they are all interactions with the same organizational Soma
- shared durable context should remain governed and lineage-aware rather than silently forked per user

## Identity Layers

The target identity model has three layers:

1. External identity adapter
   - support standard enterprise federation such as SAML and OIDC/SSO
   - support lifecycle sync such as SCIM when a deployment needs third-party user provisioning
   - map external groups/claims into Mycelis roles, teams, and governance posture

2. Local Mycelis principal layer
   - every authenticated user resolves to a Mycelis principal with a stable internal id
   - Mycelis stores local role, governance profile, approval posture, and environment scope even when the identity source is external
   - local principals remain the audit and policy subject, not raw IdP claims

3. Soma tenancy layer
   - Soma is not a separate human account per user
   - Soma is an organization-owned operating persona bound to the environment or organization root
   - all Soma interactions inherit the configured Soma identity and policy posture for that environment

## Authentication Modes

Target supported modes:

- Local-only
  - self-hosted username/password or equivalent local credentials
  - intended for small, isolated, or lab-style deployments

- Federated enterprise
  - SAML and/or OIDC SSO
  - optional SCIM-based provisioning and deprovisioning
  - intended for enterprise-managed user lifecycle and centralized access policy

- Hybrid
  - enterprise SSO for normal users
  - local break-glass admin accounts for recovery when SSO is unavailable

Rules:
- federation should be the preferred path for multi-user enterprise operation
- self-hosted local auth must remain available where external identity infrastructure is not desired
- hybrid mode must be supported so administrators are not locked out by an external IdP outage

## Administrative Access

Self-hosted Mycelis must support local administrative users even when third-party identity is configured.

This local admin path exists to:
- recover from SSO/SAML outage or IdP misconfiguration
- rotate trust, policy, and provider settings
- manage environment-level bootstrap, audit, and recovery actions
- preserve operator control in self-hosted deployments

Administrative guidance:
- local break-glass admins should be minimal, explicit, and auditable
- break-glass use should be visible in audit and environment health review
- local admin credentials should not be required for normal day-to-day federated use

## Soma Tenancy Model

Soma is the root operational persona for a deployment or organization, not a separate assistant instance invented independently per end user.

Target behavior:
- the root admin or configured environment owner defines Soma's name and baseline operating identity
- other users talk to that same governed Soma persona
- those conversations can remain private or scope-limited, but they still inherit the organization's configured Soma identity
- approvals, audit, and team delegation should show both the human actor and Soma as separate roles in the chain

Important distinction:
- shared Soma persona does not mean shared private chat visibility
- it means the same governed organizational Soma is serving each authorized user through scoped privacy and policy controls

## Admin-Shaped Soma Context

The shared Soma persona is not value-neutral. It is shaped by the root admin or configured environment owner.

This admin-shaped layer defines:
- Soma's configured name
- baseline operating stance and organization tone
- the agent/output specificity Soma and engaged specialists should follow for organization-facing work
- approved company guidance Soma may treat as durable organizational reference
- deployment-specific constraints, policies, and context posture
- organization-level memory and retrieval boundaries

This creates a strict separation between:

1. Admin-shaped Soma context
   - durable
   - organization-owned
   - policy-bearing
   - able to influence how Soma behaves for all authorized users

2. Ordinary user interaction context
   - scoped to the user, thread, team, or allowed audience
   - useful for immediate work and continuity
   - not allowed to silently redefine Soma's baseline identity or global operating posture

Product rule:
- admin interaction may shape Soma's durable organization-level operating context
- only the root admin or explicitly delegated environment owner may assign or change durable agent/output specificity for shared organizational behavior
- ordinary user interaction should remain scoped unless explicitly reviewed and promoted

Output-specificity rule:
- per-user requests can ask for temporary output preferences inside the current scoped interaction
- those temporary preferences must not rewrite shared agent specificity, shared specialist output posture, or Soma's organization-level output contract
- durable output-specificity changes for shared Soma/team behavior must go through a governed root-admin path with audit and lineage

Implementation implication:
- Mycelis likely needs one explicit admin-owned foundational layer in addition to ordinary Soma memory, `customer_context`, and `company_knowledge`
- that layer should be treated as organization-operating context, not generic chat history

## Memory And RAG Access Model

Soma must have access to the memory and knowledge layers required to produce the correct product behavior for authorized users.

Those layers remain distinct:
- Soma memory
  - remembered facts, continuity, summaries, and operating memory
- Customer context
  - customer-provided deployment material loaded into governed `customer_context`
- Company knowledge
  - approved company-authored material loaded into governed `company_knowledge`
- Admin-shaped Soma context
  - durable foundational context approved by the root admin or configured owner and allowed to shape Soma's shared organization-level behavior

Multi-user implications:
- access to those stores must be policy-scoped by user, role, team, and sensitivity
- Soma may retrieve from the right store for the current interaction, but should not bypass visibility and sensitivity policy
- private user interactions may contribute to continuity or audit only according to explicit policy, not silent global sharing
- promotion into durable shared knowledge must remain approval-backed and lineage-preserving
- admin-shaped Soma context should be writable only through explicitly governed admin paths, not through ordinary user chat

## Governance And Audit Requirements

Multi-user identity must integrate with the existing governance model instead of bypassing it.

Required behavior:
- every action is attributable to both the human principal and the acting Soma/runtime role where applicable
- approvals can become role-based and eventually multi-step
- identity source, auth mode, and effective policy context should be reconstructable from audit
- federated users, local admins, and service principals should all normalize to the same governance and audit model

At minimum, audit lineage should preserve:
- human actor
- effective role
- Soma/runtime actor
- organization or environment scope
- capability used
- result status
- approval posture
- auth source

## Hosted vs Self-Hosted Layering

User management must be deployable as either:
- a self-hosted layer inside the Mycelis environment
- an external or hosted control layer offered as a paid feature

That means the identity layer should be modular:
- the runtime and governance core should depend on a stable principal/policy contract
- federation, provisioning, and advanced directory management should plug into that contract
- self-hosted deployments should still work without the paid hosted identity plane

Practical consequence:
- "paid user management" should be an addable control-plane/service layer, not a hard dependency for self-hosted operation

## Release Posture

Current release truth:
- the free-node release ships user-level governance foundations, not full enterprise IAM
- the current auth/runtime is still closer to single-user local operation than multi-user enterprise identity
- the current Settings -> People & Access surface now exposes the intended product layering for review: `product_edition`, `identity_mode`, and `shared_agent_specificity_owner` can be persisted as a visible contract even though the underlying enterprise adapters and hosted control-plane services are still future implementation work
- the runtime now also normalizes the authenticated caller into explicit principal metadata on `/api/v1/user/me`: `principal_type`, `auth_source`, `effective_role`, and `break_glass` distinguish ordinary local admin posture from the self-hosted hybrid/federated recovery path

Required next target:
- formalize principal, auth-source, and role mapping contracts
- add SAML/OIDC-capable identity adapters
- add local break-glass admin support for self-host
- bind shared Soma identity to organization/environment ownership instead of ad hoc per-user assistant state
- define an explicit admin-shaped Soma context layer instead of letting baseline behavior emerge from unclassified chat history
- enforce scoped retrieval rules across Soma memory, customer context, and company knowledge

## Rollout Sequence

The clean implementation path should be:

1. Principal contract
   - define local principal types such as `federated_user`, `local_admin`, `service_principal`, and `break_glass_admin`
   - normalize `auth_source`, `effective_role`, and organization/environment scope in runtime code

2. Soma ownership contract
   - bind Soma identity to environment or organization ownership
   - make Soma name and baseline posture configuration-owned, not per-user session-owned

3. Admin-shaped context layer
   - introduce an explicit governed store for admin-authored Soma-shaping context
   - keep it separate from ordinary Soma memory, `customer_context`, and `company_knowledge`
   - require audit, lineage, and governed promotion/update paths
   - make root-admin control of shared agent/output specificity explicit instead of allowing it to emerge from general chat

4. Scoped user interaction rules
   - ensure ordinary user chats remain private or audience-scoped by policy
   - prevent silent promotion of ordinary user interaction into shared Soma identity or company-wide knowledge

5. Retrieval and RAG policy
   - enforce retrieval rules per store based on principal, scope, trust, and sensitivity
   - allow Soma to use the correct store for the current task without flattening all stores into one global semantic pool

6. Enterprise identity adapters
   - add SAML/OIDC federation support
   - add optional SCIM lifecycle sync
   - preserve local break-glass admin operation for self-host

7. Audit and approval expansion
   - record both the human actor and Soma/runtime actor in every relevant action
   - support future role-based and multi-step approvals for shared-organization changes

Related references:
- [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md)
- [V8 Universal Soma And Context Model PRD](V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md)
- [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
- [Governance System](../governance.md)
- [Governance & Trust](../user/governance-trust.md)

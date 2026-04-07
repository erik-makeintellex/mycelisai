# Governance & Trust
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> How Mycelis keeps operators in control of actions that could change files, trigger execution, or use higher-risk capabilities.

## Core Principle

Mycelis does not treat important actions as blind auto-execution.

For the current release, Soma should either:
- answer directly
- present a governed proposal
- return an execution result after confirmation
- show a blocker when the system cannot proceed safely

The important boundary is not a generic trust score anymore. It is whether the requested action should remain answer-only or move into governed proposal/approval flow.

## What Drives Approval

Approval posture is shaped by:
- your governance profile
- capability risk
- external data use
- estimated cost

Current profile inputs include:
- role
- cost sensitivity
- review strictness
- automation tolerance
- escalation preference

That profile is read when Soma plans work, chooses an execution path, and decides whether approval is required.

This same model also applies to governed deployment knowledge:
- loading customer-provided deployment material into the separate context store is a governed action
- loading approved company-authored knowledge is stricter than loading customer context
- external/web research used as future context should stay explicitly classified and reviewable

## Capability Risk

Current release behavior:

| Risk | Expected posture |
|------|------------------|
| Low | auto-allowed |
| Medium | optional approval |
| High | approval required |

This keeps low-risk answer work lightweight while forcing higher-risk mutations and external actions through explicit review.

Examples:
- ordinary direct explanation -> usually stays `answer`
- load customer deployment brief into `customer_context` -> governed, medium-risk by default
- load approved company-authored rollout playbook into `company_knowledge` -> governed, higher-risk and more likely to require approval
- web-fed research promoted into durable context -> governed and shaped by external-data rules

## Reviewing Proposals

Navigate to **Automations -> Approvals** or inspect proposal cards in the Workspace.

A governed proposal should show:
- risk level
- approval posture
- approval reason when relevant
- capability/tool context

You can then:
- approve and execute
- cancel before execution

The system should preserve causality:
- proposal first
- execution only after confirmation when required

## Activity Log / Audit

Mycelis records the governance trail behind operator-visible actions.

The current inspect-only audit surface is the **Activity Log** in the Approvals area. It is intended to show:
- recent actions
- approvals
- execution status
- capability usage
- artifacts and channel activity

The default UI should not dump raw logs. It should show normalized, operator-readable audit events.

## Current Release Boundaries

What exists now:
- governed proposal/confirm/cancel flow
- capability-aware approval posture
- user-level governance profile
- base audit trail and inspect-only activity view
- a reviewable People & Access model that shows the layered product story for self-hosted release, self-hosted enterprise, and hosted control plane, plus identity posture and who controls shared Soma output specificity

What is still future work:
- full multi-user IAM with SAML/OIDC federation, optional lifecycle sync, and local break-glass admin recovery for self-hosted environments
- delegated approval chains
- richer enterprise policy administration
- one shared organization-owned Soma persona across many users with scoped privacy, audit, and memory/RAG access rules
- explicit admin-owned Soma-shaping context so root/admin guidance can shape durable organization behavior without letting ordinary user chat silently redefine Soma
- root-admin control over durable shared agent/output specificity so user-local preferences do not silently redefine organization-wide output behavior

Related references:
- `docs/governance.md`
- `docs/licensing.md`
- `docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`
- `docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md`

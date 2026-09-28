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
- the delivery posture, when work uses a posture Outcome Template: a posture can only add an approval requirement (shown with the reason `outcome_posture`), never lower one, and if the safety rules cannot load, posture-shaped work requires approval. A posture-raised approval, a high-risk capability, or an estimated cost above 5.0 needs admin approval, and only an admin with approval authority can confirm it (an admin may approve their own proposal; that is recorded). If you try to approve it yourself, Soma tells you it needs admin approval and, when the reason is known, names it. Other approvals (external data, medium risk, smaller cost) you confirm yourself.

Current profile inputs include:
- role
- cost sensitivity
- review strictness
- automation tolerance
- escalation preference

That profile is read when Soma plans work, chooses an execution path, and decides whether approval is required.

This same model also applies to governed deployment knowledge:
- loading customer-provided deployment material into the separate context store is a governed action
- loading approved company-authored knowledge is stricter than loading customer context; an agent can save company knowledge or Soma operating guidance only after you confirm the action
- when saved context shapes an answer, the reply's **Sources** line says which sources were **Used** and which were only **Consulted**
- team-shared execution memory belongs in scoped `AGENT_MEMORY`; loading a governed document does not silently make it team memory
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
- load private user work log/finance/record material into `user_private_context` -> governed, high-risk by default, private/restricted unless explicitly scoped otherwise
- load approved company-authored rollout playbook into `company_knowledge` -> governed, higher-risk and more likely to require approval
- promote a distilled pattern or contradiction into `reflection_synthesis` -> governed and review-shaped rather than a casual memory write
- web-fed research promoted into durable context -> governed and shaped by external-data rules

## Trusted Memory Posture

Recalled memory is not automatic truth.

The intended precedence is:
1. deterministic logs, turns, artifacts, and explicit execution evidence
2. governed organization memory such as `company_knowledge`, `soma_operating_context`, and approved `reflection_synthesis`
3. scoped team-shared `AGENT_MEMORY`
4. Soma personal `SOMA_MEMORY`

That means:
- team-shared memory can guide execution inside its scope
- Soma may read team memory without turning it into organization doctrine
- lower-order recalled memory should not silently override newer governed policy or verified evidence

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

Only you can confirm your own proposal, unless it needs admin approval, in which case an admin confirms it. A proposal from a council member, and a team plan you launch, follow the same rules as one from Soma: a large team plan (more than 10 agents) or one using broadcast or external (MCP) tools needs admin approval. If a proposal expired, or was made before the latest governance update, ask Soma to propose it again. A team plan must be launched exactly as proposed; to change it, negotiate again with Soma first.

Using a connected tool directly (Resources → Workspace, or the tools palette) follows the same line. Anyone signed in can read the workspace directly: list folders, read files, and search. Everything else, including creating or writing workspace files and every GitHub, Slack, fetch, database or paid-media tool, is for admins who can approve high-risk work. Anyone else sees "Only an admin can use this tool directly" and nothing runs; ask Soma to propose it and an admin approves the proposal. Each direct admin call is written to the activity log before it runs, and it is refused if the log is unavailable. Secrets in a call's inputs (tokens, keys, passwords, credentials in links) are hidden in the saved record.

Known gap: approving a negotiated team plan from its chat proposal card still fails, so do not use it for this. Plan the team with Soma, then open the Workspace canvas and use its **Launch teams** button — it sends the confirm token from that negotiation and launches the plan.

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
- a reviewable People & Access model that shows the layered product story for self-hosted release, self-hosted enterprise, and hosted admin control plane, plus identity posture and who controls shared Soma output specificity; that edition/auth posture is deploy-owned review state, not an ordinary user preference
- posture-raised approvals require an admin to confirm: a non-admin who tries sees a clear "needs admin approval" blocker, nothing runs, and the proposal stays valid for an admin to confirm later

What is still future work:
- full multi-user IAM with SAML/OIDC federation, optional lifecycle sync, delegated enterprise admin flows, and hosted management-plane layering
- delegated approval chains
- richer enterprise policy administration
- one shared organization-owned Soma persona across many users with scoped privacy, audit, and memory/RAG access rules
- explicit admin-owned Soma-shaping context so root/admin guidance can shape durable organization behavior without letting ordinary user chat silently redefine Soma
- root-admin control over durable shared agent/output specificity so user-local preferences do not silently redefine organization-wide output behavior

Recovery rule:
- local break-glass recovery remains part of the self-hosted posture even when future enterprise federation is enabled

Related references:
- `docs/governance.md`
- `docs/licensing.md`
- `docs/user/memory.md`
- `docs/architecture-library/MYCELIS_CANONICAL_PRD.md`

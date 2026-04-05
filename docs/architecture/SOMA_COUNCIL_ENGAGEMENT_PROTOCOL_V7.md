# Soma-Council Engagement Protocol V7
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-03-10`

This document defines the canonical execution-path selection contract for Soma, council members, and the meta-agent (Architect). It ensures consistent behavior when choosing between:

1. direct internal tools
2. MCP tools
3. external API integration
4. direct code development and execution
5. multi-agent mission instantiation

If runtime prompts, UI hints, or ad hoc operator habits conflict with this protocol, this protocol wins.

---

## Table of Contents

1. Purpose
2. Core Decision Model
3. External API Engagement Contract
4. MCP Selection and Escalation Contract
5. Direct Code-to-Execution Contract
6. Meta-Agent (Architect) Planning Contract
7. Council Role Responsibilities
8. Governance and Safety Boundaries
9. Test and Verification Gates
10. Operator-Visible Outcomes

---

## 1. Purpose

Soma must behave as a deterministic orchestrator, not an ad hoc assistant. For any user request, the system must converge on one explicit path and show why:

- `Path A`: direct internal tool execution
- `Path B`: direct MCP tool execution
- `Path C`: blueprint + team instantiation
- `Path D`: code implementation loop (Coder + Sentry)

Required outcomes:

1. No ambiguous execution path.
2. No hidden dependency assumptions (credentials/tools/models).
3. No “manual operator glue” for common flows.
4. Every path must remain governance-auditable.
5. No schema-only or tutorial-only responses when direct execution is feasible.

---

## 2. Core Decision Model

For every actionable request, Soma (or Architect when delegated) must evaluate in order:

1. `Capability Fit`: can an existing internal tool satisfy the request safely?
2. `MCP Fit`: is an installed MCP server/tool the best fit?
3. `Integration Gap`: does this require adding/configuring an MCP or API channel?
4. `Execution Complexity`: is this single-step or multi-step with dependencies?
5. `Governance Class`: read-only, propose-only, bounded execute, or high-impact.

Decision outputs:

- `execute_now`: direct tool action (A or B)
- `propose_with_requirements`: waiting on credentials/config approvals
- `instantiate_team`: mission blueprint path (C)
- `code_loop`: Coder implementation + Sentry verification (D)

### 2.1 Consultation trigger policy (token economy)

Default posture:
- keep conversation Soma-first and direct for routine interaction.

Council consultation is required only when at least one condition is true:
1. user explicitly asks to plan, architect, design architecture, or deliver an implementation plan
2. user explicitly requests specialist/council involvement
3. request is high-impact mutation and policy requires specialist review
4. request complexity exceeds direct execution confidence and needs decomposition

Consultation modes:
- `none`: no council call, Soma responds directly
- `targeted`: one specialist selected by need
- `full_council`: only for explicitly broad architectural/program decisions

Every response should declare consultation mode implicitly through payload metadata:
- no consultation metadata for `none`
- populated consultation trace when `targeted` or `full_council`

---

## 3. External API Engagement Contract

When user intent requires external API interaction (GitHub, Slack, WhatsApp, email, CRM, etc.):

1. Enumerate required API capabilities.
2. Check available tools (`list_available_tools`) for existing supported path.
3. Prefer installed MCP route when it exists and is policy-compliant.
4. If missing, produce explicit requirements:
   - required MCP service/template
   - credentials/env vars required
   - scopes/permissions needed
   - expected risk class and approval gate
5. Never assume credentials exist.
6. Never silently fall back to ungoverned direct network patterns.

Minimum external API requirement bundle:

- `service_name`
- `required_capabilities[]`
- `credentials[]`
- `scope_bounds`
- `data_boundary` (`local|org|remote`)
- `approval_required` (`true|false`)

---

## 4. MCP Selection and Escalation Contract

MCP is the default service integration layer unless explicitly overridden by policy.

Selection rules:

1. Use existing installed MCP tool first.
2. If not installed, use local-first MCP library install plan.
3. If no MCP fit exists, route to Universal Action Interface onboarding flow.
4. High-impact MCP additions require approval token path.

Escalation levels:

- `L0`: existing tool, read-only
- `L1`: existing tool, bounded mutation
- `L2`: new MCP service install/config
- `L3`: remote actuation pathway / privileged integration

`L2` and `L3` require explicit proposal + confirm flow.

MCP translation rule:

1. Parse user request into `operation`, `target`, `constraints`, and `expected_output`.
2. Use current installed MCP inventory as source of truth (`list_available_tools` / runtime MCP reference).
3. Select the narrowest matching MCP tool and build minimal valid arguments.
4. Execute; do not respond with schema-only guidance when execution is feasible.
5. If tooling is missing, emit explicit install/credential requirements and next action.

---

## 5. Direct Code-to-Execution Contract

For software development requests, execution must follow a repeatable closed loop:

1. `Scope`: objective + acceptance criteria + runtime constraints.
2. `Implement`: Coder modifies code and tests.
3. `Verify`: Sentry reviews security/risk/compliance.
4. `Validate`: run relevant test suites/build checks.
5. `Report`: summarize changes, evidence, residual risk, next action.

Rules:

- No “code dump only” for execution requests.
- No completion claim without test evidence or explicit test blocker.
- If external API dependencies exist, include requirement bundle before execution.
- Prefer quick, ephemeral code paths that can be validated immediately before introducing
  new MCP dependencies for software-only tasks.

Web-access execution rule:

1. Requests for search/site retrieval should default to development-specialist-owned ephemeral code execution.
2. Build adaptive search-engine strategy (Google/Bing/DDG based on availability and constraints).
3. Construct high-quality, intent-scoped queries (operators, site/domain scope, recency qualifiers).
4. Use onboarded MCP for web tasks only when it is clearly easier or required by environment/policy.

---

## 6. Meta-Agent (Architect) Planning Contract

Architect is responsible for choosing and justifying the path when requests are multi-step, cross-domain, or capability-uncertain.

Architect outputs must include:

1. selected path (`A|B|C|D`, or staged combination)
2. dependency list (MCP/API/model/profile/policy)
3. team topology (if instantiated)
4. I/O contracts and channel subjects
5. governance checkpoints and approval triggers
6. rollback and degraded-mode fallback strategy

Blueprints must avoid over-instantiation: prefer 2-4 agents per mission unless clear justification exists.

---

## 7. Council Role Responsibilities

- `Architect`: pathing, decomposition, dependency contract, blueprint quality.
- `Coder`: implementation quality, test readiness, deterministic outputs.
- `Creative`: UX/content/interaction design with governance-aware execution paths.
- `Sentry`: risk analysis, policy checks, approval boundaries, hard-stop escalation.

Council members must not bypass Soma governance paths for high-impact actions.

---

## 8. Governance and Safety Boundaries

Mandatory controls:

1. no credential assumptions
2. no unscoped privileged API usage
3. no unapproved high-impact mutation
4. no execution without path declaration
5. no opaque fallback to legacy endpoints

For denied or degraded operations, responses must provide:

- what failed
- likely cause
- safe next action
- diagnostics reference

---

## 9. Test and Verification Gates

Required coverage for this contract:

1. Prompt/manifest contract tests for Soma and council directives.
2. Integration tests for:
   - direct-first small-intent behavior (no forced council consult)
   - explicit plan/architect/deliver intents trigger targeted consultation path
   - MCP-first selection
   - missing dependency requirement emission
   - approval-required high-impact path
3. UI tests for operator visibility of path/requirements/degraded actions.
4. Regression checks for legacy endpoint avoidance.

---

## 10. Operator-Visible Outcomes

Operator should always be able to answer:

1. Which execution path was selected and why?
2. Which dependencies are required before execution?
3. Whether action is direct, delegated, or waiting for approval?
4. How to recover when channels degrade?

This is the minimum bar for trustworthy orchestration at V7.

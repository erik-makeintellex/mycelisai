# V7 Parallel Implementation Board (Unified)

Source of truth for prioritized, parallel execution aligned to:
- [Soma Extension-of-Self PRD](../product/SOMA_EXTENSION_OF_SELF_PRD_V7.md)
- [V7 Implementation Plan](../V7_IMPLEMENTATION_PLAN.md)
- [UI Framework V7](../UI_FRAMEWORK_V7.md)
- [mycelis-architecture-v7.md](../../mycelis-architecture-v7.md)
- [V7_DEV_STATE.md](../../V7_DEV_STATE.md)

---

## Execution State

Status: `ENGAGED`
Date: `2026-02-28`
Program Phase: `Inception Reset - Sprint 0 (Contract Freeze + Single Use Case)`

Implementation is now governed by:
1. Single-use-case ratchet (prove operator benefit before scope expansion)
2. Heartbeat autonomy budgets + kill-switch
3. Staging-only self-update path with rollback
4. Sandbox-first dynamic skill/runtime execution
5. No mission-critical/financial scopes until gates pass

---

## Program Priorities

## P0 - Safety and Contract Baseline (must complete first)
1. Decision frame contract freeze (`direct|manifest_team|propose|scheduled_repeat`)
2. Universal invoke envelope freeze (MCP + non-MCP compatible)
3. Local Ollama readiness/fallback status contract
4. Security controls: replay/idempotency/self-update staging path
5. One stable pipeline selected and instrumented with "worth-it" metrics

## P1 - First Vertical Slice (governed execution)
1. Direct-vs-team decision endpoint operational
2. Team instantiation wizard + lifecycle evidence path
3. Workspace decision trace + proposal/approval continuity
4. Degraded and recovery continuity for SSE/NATS/Ollama

## P2 - Repeat and Long-Horizon Reliability
1. Scheduler + repeat-promotion path (`one-off` -> scheduled)
2. Heartbeat worker with budget enforcement
3. Run timeline/chain evidence complete for long-horizon workflows

## P3 - Controlled Expansion
1. Non-MCP adapter parity pilot
2. Trusted skill import pilot (signed + sandboxed + scope-limited)
3. Host/hardware governed actuation scaffolds

---

## Lanes and Ownership (parallel)

1. Lane A: Decision + Memory Contracts (`Team Helios`)
   - Scope: decision frame types/APIs, learning signals, run-linked events
   - Priority: `P0/P1`
2. Lane B: Universal Action Runtime (`Team Forge`)
   - Scope: action registry, invoke pipeline, adapter contract, idempotency/replay
   - Priority: `P0/P3`
3. Lane C: Scheduler + Team Lifetime (`Team Atlas`)
   - Scope: schedule runtime, repeat promotion, lifetime policy transitions
   - Priority: `P1/P2`
4. Lane D: Operator UX + Readiness Surfaces (`Team Circuit + Team Argus`)
   - Scope: Workspace decision trace, readiness diagnostics, degraded recovery UX
   - Priority: `P1/P2`
5. Lane Q: QA, Security, and Release Gates (`Team Sentinel`)
   - Scope: cross-lane contract validation, harsh-truth controls, release evidence
   - Priority: `Cross-cutting`

Legacy UI lane documents remain as implementation references:
- [LANE_A_GLOBAL_OPS.md](./LANE_A_GLOBAL_OPS.md)
- [LANE_B_WORKSPACE_RELIABILITY.md](./LANE_B_WORKSPACE_RELIABILITY.md)
- [LANE_C_WORKFLOW_SURFACES.md](./LANE_C_WORKFLOW_SURFACES.md)
- [LANE_D_SYSTEM_OBSERVABILITY.md](./LANE_D_SYSTEM_OBSERVABILITY.md)
- [LANE_Q_QA_REGRESSION.md](./LANE_Q_QA_REGRESSION.md)

---

## Gate Model (promotion order)

1. Gate P0 - Contract + Safety Baseline
   - Lanes A/B/Q green
   - Required: schema freeze, security baseline, single-use-case chosen
2. Gate P1 - First Vertical Slice
   - Lanes A/C/D/Q green
   - Required: direct/team decision flow + governed UX continuity
3. Gate P2 - Repeat + Long-Horizon Reliability
   - Lanes C/D/Q green
   - Required: scheduler + heartbeat budget + recovery evidence
4. Gate P3 - Controlled Expansion
   - Lanes B/D/Q green
   - Required: non-MCP pilot + trusted skill import + sandbox controls
5. Gate RC - Release Candidate
   - Full cross-layer suite green
   - No unresolved high-severity security regressions

---

## Inception Harsh-Truth Controls (must be test-covered)

1. Integration overhead guard:
- every new pipeline must include operator-benefit KPI evidence

2. Self-update guard:
- no direct production self-update path
- all update intents route to staging workflow with rollback

3. Trojan/credential overreach guard:
- credential probing denied by default policy
- deterministic secret redaction in diagnostics/logs

4. Heartbeat guard:
- budgets for time/actions/spend/escalation
- budget exhaustion triggers safe pause + auditable event

5. Skill economy guard:
- external skill imports require signed manifest, scope review, sandbox profile
- medium/high-risk scopes blocked for unverified skills

---

## Dependency Matrix

| Lane | Depends On | Unblocks |
| :--- | :--- | :--- |
| A | Event spine + run store contracts | Decision trace and learning signal integrity |
| B | Decision contract + security policy primitives | Adapter expansion and governed actuation |
| C | Runs/events APIs + policy engine | Repeat/scheduled workflows and team-lifetime transitions |
| D | Lanes A/B/C contracts + status APIs | Operator-safe execution and recovery UX |
| Q | All lanes | Promotion confidence and release authority |

---

## Sprint 0 Work Queue (active)

1. Helios (Lane A)
- freeze decision frame DTOs + persistence contracts
- emit run-linked decision transition events

2. Forge (Lane B)
- freeze universal invoke envelope
- implement idempotency/replay validation middleware
- define staging-only self-update contract path

3. Atlas (Lane C)
- finish scheduler critical path and pause/resume reliability
- draft repeat-promotion policy contract

4. Circuit + Argus (Lane D)
- expose decision trace in Workspace
- expose local readiness/fallback in StatusDrawer/System
- add single-use-case KPI panel shell

5. Sentinel (Lane Q)
- finalize P0 pass criteria and evidence format
- add tests for self-update denial, credential-probe denial, and heartbeat budget enforcement

---

## Evidence Requirements (every lane PR)

Each lane handoff must include:
1. changed files and reason
2. API/schema diffs
3. tests added + pass output
4. known risks and mitigations
5. rollback plan
6. operator-visible behavior summary
7. gate label: `READY_FOR_INTEGRATION`, `BLOCKED_<reason>`, or `REQUIRES_POLICY_DECISION`

No merge without lane evidence.

---

## Definition of Engaged Delivery

Delivery is engaged when:
1. every lane has active checklist, owner, and evidence block
2. harsh-truth controls are implemented and test-covered
3. promotion gates have explicit pass/fail criteria and artifacts
4. single-use-case pipeline has measurable operator-benefit proof before autonomy expansion


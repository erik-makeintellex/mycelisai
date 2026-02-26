# UI Parallel Delivery Board (V7)

Source of truth for prioritized, parallel UI execution aligned to:
- [UI Framework V7](../UI_FRAMEWORK_V7.md)
- [mycelis-architecture-v7.md](../../mycelis-architecture-v7.md)
- [V7_DEV_STATE.md](../../V7_DEV_STATE.md)

---

## Execution State

Status: `ENGAGED`
Date: `2026-02-26`

Parallel delivery is now active with five lanes and gated merge flow.

---

## Prioritization

## P0 - Operator Trust and Recoverability (must complete first)
1. Global status truth: drawer + degraded banner hardening
2. Workspace failure recovery: structured council errors, reroute, retry
3. System quick checks confidence path

## P1 - Workflow Clarity and Throughput
1. Automations actionable hub and first-chain guidance
2. Teams diagnostics and quick-action pathways
3. Resources capability visibility and governance clarity

## P2 - Observability Depth and Scale
1. Runs discoverability and detail consistency
2. Timeline/chain performance hardening
3. Hardware channels scaffold UX

---

## Lanes (parallel)

1. Lane A: Global Ops UX
   - File: [LANE_A_GLOBAL_OPS.md](./LANE_A_GLOBAL_OPS.md)
   - Priority: `P0`
2. Lane B: Workspace Reliability UX
   - File: [LANE_B_WORKSPACE_RELIABILITY.md](./LANE_B_WORKSPACE_RELIABILITY.md)
   - Priority: `P0`
3. Lane C: Workflow Surfaces UX
   - File: [LANE_C_WORKFLOW_SURFACES.md](./LANE_C_WORKFLOW_SURFACES.md)
   - Priority: `P1`
4. Lane D: System + Observability UX
   - File: [LANE_D_SYSTEM_OBSERVABILITY.md](./LANE_D_SYSTEM_OBSERVABILITY.md)
   - Priority: `P2`
5. Lane Q: QA Reliability and Regression
   - File: [LANE_Q_QA_REGRESSION.md](./LANE_Q_QA_REGRESSION.md)
   - Priority: `Cross-cutting`

---

## Gate Model (merge order)

1. Gate A - P0 foundation:
   - Lane A + Lane B + Lane Q P0 suite green
2. Gate B - P1 workflow:
   - Lane C + Lane Q workflow suite green
3. Gate C - P2 observability:
   - Lane D + Lane Q observability/perf suite green
4. Gate RC - release candidate:
   - Full regression suite green

---

## Dependency Matrix

| Lane | Depends On | Unblocks |
| :--- | :--- | :--- |
| A | Services status endpoint + store status flags | Banner and drawer consistency across app |
| B | Lane A status actions + workspace chat hooks | Reliable council fallback UX |
| C | Lane A shared status primitives | Actionable automations/teams/resources surfaces |
| D | Runs/events APIs + Lane A status primitives | Deep observability and quick diagnostics |
| Q | All lanes | Release confidence |

---

## Definition of Engaged Delivery

Delivery is engaged when:
1. Every lane has active checklist, owner placeholder, and evidence block.
2. Every lane ties back to the framework state contract and failure template.
3. Every gate has explicit pass/fail criteria with test artifacts.

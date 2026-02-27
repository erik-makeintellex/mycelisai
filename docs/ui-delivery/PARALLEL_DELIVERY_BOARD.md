# UI Parallel Delivery Board (V7)

Source of truth for prioritized, parallel UI execution aligned to:
- [UI Framework V7](../UI_FRAMEWORK_V7.md)
- [UI Workflow Instantiation and Bus Plan](../product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md)
- [mycelis-architecture-v7.md](../../mycelis-architecture-v7.md)
- [V7_DEV_STATE.md](../../V7_DEV_STATE.md)

---

## Execution State

Status: `ENGAGED`
Date: `2026-02-27`
Phase: `Sprint 0 - Parallel kickoff`

Parallel delivery is active with five lanes, named team ownership, and gated merge flow.

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
4. Guided team instantiation flow and readiness gate

## P2 - Observability Depth and Scale
1. Runs discoverability and detail consistency
2. Timeline/chain performance hardening
3. Hardware channels scaffold UX

---

## Lanes and Team Ownership (parallel)

1. Lane A: Global Ops UX (`Team Circuit`)
   - File: [LANE_A_GLOBAL_OPS.md](./LANE_A_GLOBAL_OPS.md)
   - Priority: `P0`
2. Lane B: Workspace Reliability UX (`Team Atlas`)
   - File: [LANE_B_WORKSPACE_RELIABILITY.md](./LANE_B_WORKSPACE_RELIABILITY.md)
   - Priority: `P0`
3. Lane C: Workflow Surfaces UX (`Team Helios + Team Argus`)
   - File: [LANE_C_WORKFLOW_SURFACES.md](./LANE_C_WORKFLOW_SURFACES.md)
   - Priority: `P1`
4. Lane D: System + Observability UX (`Team Argus`)
   - File: [LANE_D_SYSTEM_OBSERVABILITY.md](./LANE_D_SYSTEM_OBSERVABILITY.md)
   - Priority: `P2`
5. Lane Q: QA Reliability and Regression (`Team Sentinel`)
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
| C | Lane A shared status primitives + readiness data model | Actionable automations/teams/resources + instantiation flow |
| D | Runs/events APIs + Lane A status primitives | Deep observability and quick diagnostics |
| Q | All lanes | Release confidence |

---

## Team Interface Contracts

1. Atlas -> Helios
   - needs `ReadinessSnapshot` and capability blockers
   - receives typed readiness adapter output only
2. Atlas -> Circuit
   - needs bus health tier and reconnect actions
   - receives `Basic|Guided|Expert` exposure state
3. Helios -> Argus
   - provides channel metadata (`channel`, `run_id`, `team_id`)
4. Circuit -> Sentinel
   - provides degraded/recovery scenarios and expected UI state transitions
5. Argus -> Atlas
   - provides run pivot endpoints and latest-run summary for team cards

Contract rule:
- no lane merges state shape changes without a handoff note in PR description.

---

## Sprint 0 Work Queue (active)

1. Circuit (Lane A)
   - finalize global health model parity across banner/drawer/system checks
   - enforce deterministic reconnect action behavior
2. Atlas (Lane B)
   - align workspace failure cards with reroute and retry continuity
   - preserve context for inline recovery actions
3. Helios + Argus (Lane C bootstrap)
   - define and scaffold `TeamInstantiationWizard` and `CapabilityReadinessGateCard`
   - integrate team-to-run quick pivot contract
4. Sentinel (Lane Q)
   - codify Gate A/B pass criteria and evidence format
   - add reliability scenarios for bus and readiness gate degradation

---

## Parallel Control Loop

Daily control cadence:
1. 15-minute lane sync: blockers, handoffs, schema changes
2. Midday integration check: store and route contract drift review
3. End-of-day evidence update: lane doc test output + state notes

Weekly gate review:
1. Validate gate suite status
2. Approve promotion or return lanes with defect list

---

## Definition of Engaged Delivery

Delivery is engaged when:
1. every lane has active checklist, owner, and evidence block
2. every lane ties back to framework state contract and failure template
3. every gate has explicit pass/fail criteria with test artifacts
4. handoff contracts are documented for cross-lane dependencies


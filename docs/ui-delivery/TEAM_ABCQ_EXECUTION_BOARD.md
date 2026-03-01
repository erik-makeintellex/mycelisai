# Team A/B/C/Q Parallel Execution Board (Sprint G1 + G2)

Status: `ACTIVE`  
Updated: `2026-03-01`  
Scope:
- Sprint G1: Root-admin collaboration groups hardening and integration into mission runtime
- Sprint G2: Delivery behavior hardening (symbiote execution, coder-first web access, MCP translation reliability)

---

## Objective

Execute the next 4-team slice in parallel without losing operational hygiene:
1. complete runtime lineage (`group_id` propagation)
2. enforce group capability policy at execution points
3. deliver operator UX continuity
4. prove behavior with integration + E2E evidence

For Sprint G2, execute in parallel:
1. enforce delivery-over-tutorial behavior for action requests
2. codify coder-first web/search execution with adaptive query strategy
3. harden MCP translation reliability from user intent -> concrete MCP calls
4. expose execution-path evidence in operator UI and tests

---

## Team Charters

### Team A — Core/Persistence
- Add `group_id` to mission execution lineage (`mission_runs`, `mission_events`, run APIs).
- Ensure `group_id` is persisted, queryable, and returned in timeline/chain payloads.
- Add migration(s), backfill/default handling, and compatibility path.
- Sprint G2:
  - Persist execution path metadata (`path_selected`, `execution_owner`, `translation_source`) in run/event payloads.
  - Add `delivery_mode` markers (`executed`, `blocked`, `requirements_pending`) for post-run diagnostics.

### Team B — Governance/Auth
- Enforce group capability bounds at dispatch/invoke points.
- Block out-of-scope actions and emit explicit denial audit records.
- Keep root-admin + scope checks fail-closed and test-covered.
- Sprint G2:
  - Add policy guard for guidance-only fallback when direct execution is feasible.
  - Enforce web-access routing defaults (coder-first ephemeral path; MCP exception only when easier/required).
  - Emit structured denial reasons for missing MCP/tool dependencies and credential gaps.

### Team C — Runtime/UI
- Add group lifecycle UX: create/update + high-impact approval continuity.
- Add group-aware filtering in relevant run/timeline/operator surfaces.
- Surface denial reasons and next actions without dead-end states.
- Sprint G2:
  - Show selected execution path in Workspace and run timeline (`direct`, `code-first`, `mcp`, `team`).
  - Add “Delivered vs Guidance” indicator for action requests.
  - Display MCP translation outcomes (executed tool, missing dependency bundle, next action).

### Team Q — QA
- Integration tests for A+B invariants and denial paths.
- Playwright E2E for group lifecycle + approval + broadcast + monitor.
- Gate promotion only when evidence bundle is complete.
- Sprint G2:
  - Add regression tests for tutorial-only failure mode (must execute or return concrete blocker).
  - Add E2E for coder-first web task execution and MCP translation fallback.
  - Validate no legacy endpoint fallbacks and no schema-only dead-end responses.

---

## Parallelization Plan

## Wave 0 (Contract Freeze, Day 0)
- Owner: A+B+C+Q
- Freeze contracts:
  - `group_id` runtime payload shape
  - capability-denial error envelope
  - UI state model for high-impact group mutation approval
- Output: one contract note in PR + test cases mapped.

## Wave 1 (Parallel Build, Days 1-2)
- Team A:
  - migrations + store/API propagation
  - unit/integration for lineage integrity
- Team B:
  - execution guard hooks + denial audit emission
  - unit/integration for allow/deny matrix
- Team C:
  - UI flow for group mutation and approval continuity
  - status/timeline surfaces for group context
- Team Q:
  - build test harnesses in parallel against frozen contracts
  - build Sprint G2 delivery-behavior harnesses in parallel

## Wave 2 (Integration, Day 3)
- Joint merge of A+B first (runtime contract complete).
- Team C rebases once on integrated API.
- Team Q runs full integration + E2E gate suite.
- Team Q runs both G1 and G2 gate suites.

## Wave 3 (Release Candidate, Day 4)
- All team evidence posted.
- Docs/state gate applied.
- Promote only if all acceptance criteria pass.

---

## Acceptance Criteria

### Team A
- `group_id` appears in run and event APIs.
- Backward compatibility maintained for older runs with null/empty group.
- Migration + tests pass.
- G2: execution path metadata is queryable in timeline/diagnostic payloads.

### Team B
- Out-of-scope capability invocation denied deterministically.
- Denials produce auditable records with actor, group, capability, reason.
- No bypass paths for root-admin scope enforcement.
- G2: action requests do not regress into schema/tutorial-only replies when execution path exists.

### Team C
- Operator can see and act on group context in UI without route dead-ends.
- High-impact mutation approval flow is explicit and recoverable.
- Degraded mode messages remain actionable.
- G2: operator can see path chosen and why, including MCP translation success/fallback details.

### Team Q
- Integration: group lifecycle + policy enforcement + lineage integrity.
- E2E: create group -> approve -> execute bounded flow -> inspect run/timeline.
- Regression: no legacy endpoint fallback, no hydration regressions in touched pages.
- G2: E2E proves execution delivery for action requests and web-search tasks under coder-first defaults.

---

## Required Evidence Bundle (Per Team PR)

1. changed files + rationale  
2. API/schema diff summary  
3. tests added/updated and pass output  
4. known risks + mitigations  
5. rollback plan  
6. operator-visible behavior summary  
7. status tag: `READY_FOR_INTEGRATION` | `BLOCKED_<reason>` | `REQUIRES_POLICY_DECISION`

---

## Documentation Gate (Mandatory)

Every merged slice must update:
1. `README.md` (operator-facing behavior/config changes)
2. `V7_DEV_STATE.md` (what changed, what is next, evidence commands)
3. `docs/V7_IMPLEMENTATION_PLAN.md` (if roadmap/dependency changed)
4. `interface/lib/docsManifest.ts` (if a new authoritative doc is added)

No merge when documentation gate is incomplete.

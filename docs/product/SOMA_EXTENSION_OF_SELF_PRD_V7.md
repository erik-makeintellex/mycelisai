# Soma Extension-of-Self PRD (V7)

Status: `Authoritative Planning`
Last Updated: `2026-03-01`
Scope: extension-of-self execution model, local Ollama-first cognition, universal action integration, and parallel delivery plan

---

## Table of Contents

1. Why This PRD Exists
2. Current State Baseline
3. Target Product Outcomes
4. Operator Workflows (End-to-End)
5. Architecture Scope and Contracts
6. Local Ollama Communication and Readiness
7. Parallel Delivery Program
8. Detailed Sprint Plan
9. Test and Validation Program
10. Risks and Controls
11. Handoff Package for Parallel Agent Teams
12. Definition of Done

---

## 1. Why This PRD Exists

Mycelis must evolve from a workflow UI into a governed extension-of-self runtime where Soma can:
- act directly when safe
- manifest short-lived or long-lived teams when needed
- retain and improve reasoning quality over time
- execute local host and hardware actions through controlled channels

This PRD turns that objective into executable delivery tracks with explicit ownership, dependencies, and quality gates.

---

## 2. Current State Baseline

As of `2026-02-28`:
- Workflow-first IA is live (`/dashboard`, `/automations`, `/resources`, `/memory`, `/system`).
- V7 Team A (event spine) and Team B (triggers) are complete.
- Scheduler (Team C) remains the primary backend gap.
- UI framework and lane-based parallel delivery are documented and active.
- MCP local-first architecture, universal action architecture, and Soma symbiote architecture are documented.
- Verified UI host runs v1 council APIs and automations baseline/wizard paths.

Primary constraints:
- No uncontrolled remote actuation.
- Governance approval remains mandatory for risky mutations.
- Backward compatibility must be maintained while introducing universal action pathways.

### 2.1 Inception Lessons Breakdown (Video-Derived)

Strategic value to adopt:
1. Agent-to-agent skill economy pattern:
- support controlled dynamic skill injection from trusted repositories/peers
- treat imported skills as governed capability units, not implicit code trust
2. Long-horizon autonomy ("heartbeat") pattern:
- allow multi-day workflows via bounded heartbeat loops
- require periodic progress emissions and continuation checks
3. On-prem private-cloud posture:
- local-first execution with sensitive data/model processing on controlled hardware
- cloud pathways remain optional and policy-scoped

Architectural friction to plan for:
1. Integration overhead and debugging drag:
- pipeline automation often fails on auth/session/key coordination
- measure "automation worth it" before expanding workflow scope
2. Self-update failure mode:
- autonomous self-update can brick runtime if directly applied
- enforce staging, validation, and rollback before production apply
3. Security overreach and trojan vectors:
- local probing/credential discovery must be denied by default
- deterministic redaction and bounded scanning scopes are mandatory

Deployment guidance (mandatory at inception):
1. Sandbox isolation by default for dynamic execution.
2. One stable use case first, then permission expansion by gates.
3. No financial or mission-critical scopes until reliability/security thresholds are proven.

### 2.2 Five-Pattern Assimilation (Execution Policy)

The following patterns are in-scope and must be implemented as governed capabilities:

1. Sub-agent orchestration (manager-worker):
- Soma remains manager/orchestrator.
- workers run with scoped model/context/permission profiles.
- delegation decisions and worker outcomes must be persisted per run.

2. Modular skill injection:
- marketplace/private-hub skills are capability packages, not trusted code.
- install/activation requires signature + scope + sandbox checks.
- all installs must support rollback and audit trace.

3. Proactive autonomy:
- heartbeat jobs can proactively execute within explicit budgets.
- proactive flows must always emit progress and next-step evidence.
- operator can pause/kill all proactive loops from one control plane.

4. Compaction-safe memory continuity:
- critical state must be checkpointed before context compaction/truncation.
- resume logic reads checkpoints + distilled memory, not transcript alone.
- checkpoint and resume integrity must be test-covered.

5. Self-improving correction loops:
- user corrections are stored as structured hot-memory corrections.
- promotion into durable recipes requires review and rollback guardrails.
- no silent behavior drift from unreviewed corrections.

---

## 3. Target Product Outcomes

### O1. Direct vs Team Execution Intelligence
Soma chooses between:
- direct action
- manifested team
- propose/defer

Based on typed decision inputs: complexity, risk, repeat intent, channel requirements, and governance mode.

### O2. Extension-of-Self Memory and Adaptation
Soma builds durable capability through:
- episodic records
- distilled recipe memory
- vector retrieval
- bounded adaptation with operator review
- manifest-profile deepening tied to user-invoked objective classes

### O3. Unified Action Plane
All execution channels (MCP, OpenAPI, Python manager, hardware adapters) use one canonical invoke contract with predictable status semantics.

### O4. Local-First Actualization
Ollama-backed local cognition is first-class and observable:
- startup/health checks
- latency/throughput reporting
- explicit fallback routing when local inference degrades

### O5. Low-Manual Operations
Operators manage teams/channels through guided surfaces and templates, not raw topic or ad hoc API work.

---

## 4. Operator Workflows (End-to-End)

### W1. Intent to Governed Action
1. User submits intent in Workspace.
2. Soma emits decision frame (`direct|manifest_team|propose`) with rationale.
3. If approval required, structured proposal card is shown inline.
4. On confirm, run starts and timeline/chain evidence is linked.

### W2. Manifest Team for Complex Task
1. User opens Automations wizard.
2. Chooses objective/profile/lifetime (`ephemeral|persistent|auto`).
3. Readiness checks validate required capabilities.
4. Team instantiated; run + team health visible immediately.

### W3. Repeat/Scheduled Upgrade
1. User flags “repeat this action.”
2. Scheduler rule is generated from existing run context.
3. Team lifetime is promoted when policy dictates persistent operation.
4. User monitors future executions in run history and quick checks.

### W4. Local Host/Hardware Action
1. User requests host/device action.
2. Soma routes via universal action + hardware/host adapter.
3. Governance gate evaluates risk and approval requirement.
4. Action result and diagnostics are recorded with replay-safe identifiers.

### W5. Root-admin Group Orchestration
1. Root admin asks Soma to define a collaboration group for a specific goal.
2. Soma proposes group contract (`members`, `goal`, `work_mode`, `capabilities`, `approval policy`).
3. Root admin confirms; group becomes active with scoped permissions.
4. Group actions/runs are tracked with `group_id` linkage and governed by policy.
5. Root admin can pause, revise, or sunset the group through Soma with full audit trail.

---

## 5. Architecture Scope and Contracts

### 5.1 Canonical Decision Contract
Required fields in decision frame:
- `request_id`
- `run_id`
- `path_selected`
- `team_lifetime`
- `risk_level`
- `confidence`
- `approval_required`
- `required_capabilities[]`
- `fallback_plan`

### 5.2 Universal Invoke Contract
Use shared envelope for all action providers:

```json
{
  "ok": true,
  "data": {
    "request_id": "req_123",
    "run_id": "run_123",
    "action_id": "host.process.restart",
    "provider_type": "mcp|openapi|python|hardware",
    "status": "success|degraded|failure",
    "idempotency_key": "idem_123",
    "output": {}
  },
  "error": ""
}
```

### 5.3 Team Lifecycle Contract
- `ephemeral`: short-lived, single objective
- `persistent`: long-lived, stateful process ownership
- `auto`: promoted/demoted by policy + repeat behavior + channel type

### 5.4 Memory Contract
Separate layers:
- episodic (`events`, `conversation_turns`)
- distilled (`inception_recipes`, reviewed patterns)
- retrieval (vector index)
- procedural (approved automation playbooks)

Permanent memory loop requirements:
1. capture run evidence and user outcomes to Postgres
2. distill reusable patterns to recipe/sitrep artifacts
3. embed retrieval artifacts into pgvector indexes
4. recall via semantic + objective-class retrieval
5. promote improvements only through review gates
6. rollback automatically/manual on drift regressions

No-silence preference rule:
- do not infer preferences from silence
- only explicit confirmation/correction can promote preference-level changes

Local continuity channels (simple 3-layer files):
- hot: `~/.mycelis/memory/hot.jsonl` (always loaded, confirmed rules only)
- context: `~/.mycelis/memory/context/{tenant_id}/{project_slug}.jsonl`
- archive: `~/.mycelis/memory/archive/{yyyy-mm}/{project_slug}.jsonl` (inactive by default)
- Postgres remains system-of-record; local files are continuity channels and must sync deterministically

Central agentry memory policy:
1. every user correction is logged with timestamp + context
2. correction repetition is tracked
3. on the 3rd repeat of equivalent correction, prompt promotion scope (`global|domain|project`)
4. weekly maintenance run must:
- remove duplicates
- keep confirmed rules concise
- move stale entries to archive
- emit a short "what changed" digest
5. when applying learned rules, include provenance (`file` + `line`) in response metadata
6. never persist sensitive data (passwords, tokens, financial/health data, private data about third parties)
7. `forget X` deletes matching memory everywhere and confirms completion
8. `forget everything` wipes memory subsystem and confirms clean reset
9. ask before high-impact workflow changes inferred from learned memory

### 5.5 Manifest Improvement Contract
Each active Soma profile must include a `manifest_improvement` policy that binds learning to user-invoked objectives.

Required fields:
- `enabled`
- `target_alignment`
- `signal_sources[]`
- `promotion_policy` (`review_required`, `min_confidence`, `min_sample_size`)
- `rollback_policy`
- `max_change_rate_per_week`
- `drift_guard`

Rules:
- improvements must map to objective classes that users actually invoke
- unaligned improvements are rejected by default
- profile promotions are auditable and reversible

### 5.6 Multi-User Collaboration Group Contract
Root-admin-defined groups are first-class execution scopes.

Required fields:
- `group_id`
- `tenant_id`
- `name`
- `goal_statement`
- `work_mode` (`read_only|propose_only|execute_with_approval|execute_bounded`)
- `allowed_capabilities[]`
- `member_user_ids[]`
- `coordinator_profile`
- `approval_policy_ref`
- `status` (`active|paused|archived`)

Rules:
- group-scope actions must remain within `goal_statement` + `allowed_capabilities`
- group mutations by Soma require root-admin governance confirmation when high-impact
- all group runs emit `group_id`-linked evidence for audit and replay

---

## 6. Local Ollama Communication and Readiness

### 6.1 Operational Requirement
Ollama must be treated as a first-class dependency for local autonomy, not a hidden fallback.

### 6.2 Readiness Checks (Required)
Expose and test:
1. `OLLAMA_HOST` reachability
2. selected model availability
3. inference p95 latency
4. token throughput
5. recent failure ratio

### 6.3 Runtime Behavior Rules
- If local model healthy: prefer local routes for allowed roles.
- If degraded: show explicit status and apply configured fallback profile.
- If unavailable: block risky auto-execution paths that rely on local reasoning quality unless policy allows fallback.

### 6.4 UI Exposure
System + StatusDrawer must show:
- active local model
- local/remote badge
- last inference check time
- fallback mode currently active (if any)

---

## 7. Parallel Delivery Program

### Lane A — Decision + Memory Contracts (Backend)
Owns:
- Soma decision frame structs/APIs
- memory signal capture and review scaffolds
- event bindings for decision transitions

### Lane B — Universal Action + Adapters (Backend)
Owns:
- action registry/invoke APIs
- adapter interfaces for MCP/OpenAPI/Python/hardware
- idempotency + replay protections

### Lane C — Scheduler + Team Lifetime Runtime (Backend)
Owns:
- scheduled mission runtime completion
- repeat intent promotion logic
- team lifetime transition policies

### Lane D — UX Surfaces (Frontend)
Owns:
- decision-frame visibility in Workspace
- team lifecycle/instantiation UX completion
- status/readiness exposure for Ollama and channels

### Lane Q — QA + Reliability (Cross-cutting)
Owns:
- test matrix enforcement
- degraded/recovery verification
- release gate evidence bundles

Parallelization rules:
- Lanes A/B/C/D run concurrently.
- Lane Q validates each lane increment and cross-lane integration gates.
- No merge to `main` without lane evidence attached.

---

## 8. Detailed Sprint Plan

### Sprint 0: Contract Freeze and Scaffolding
Deliver:
- decision contract schemas
- universal action envelope freeze
- Ollama readiness API contract
- test fixture baselines
- heartbeat budget schema (`time/actions/spend/escalation`)
- staging-only self-update workflow contract
- skill import trust policy contract (signature/scope/sandbox profile)
- manager-worker delegation schema (`manager_plan`, `worker_assignment`, `worker_result`, `resume_hint`)
- compaction-safe memory checkpoint schema (`checkpoint_id`, `task_state`, `next_intent`, `resume_source`)
- manifest improvement schema and objective-alignment scoring contract

Exit criteria:
- schemas published and reviewed
- contract tests passing

### Sprint 1: First Vertical Slice (Direct + Manifest)
Deliver:
- direct vs team decision endpoint wired
- wizard launches manifested team with typed lifetime
- decision trace visible in Workspace
- single-use-case production pipeline wired with "worth it" metrics
- first manager-worker route active with one persistent specialist worker profile
- worker handoff evidence visible in run timeline
- capture improvement signals on manifest runs (`acceptance`, `correction`, `fallback`, `outcome`)
- implement loop contract + local file sync module in core
- correction repetition tracker with 3-strike promotion scope prompt
- root-admin group create/update/pause API slice with governance checks

Exit criteria:
- end-to-end path works on local host with run evidence

### Sprint 2: Repeat + Scheduler + Adaptation Signals
Deliver:
- scheduler runtime and repeat promotion
- adaptation signal capture/review APIs
- initial policy-bounded recipe feedback loop
- heartbeat worker with bounded continuation logic and kill-switch integration
- proactive autonomy pilot ("surprise me" class job) with safe default scope
- correction log workflow from user correction to reviewed recipe candidate
- reviewed manifest-profile promotion flow with objective-alignment gating
- automated rollback on drift-threshold breach
- permanent memory loop pilot (capture -> distill -> vectorize -> retrieve -> promote -> rollback)
- no-silence promotion guard and drift rollback tests
- recall-quality gates (pgvector retrieval metrics) bound to promotion decisions
- weekly maintenance review worker + digest emission
- group-scoped run linkage (`group_id`) across events/timelines and policy enforcement

Exit criteria:
- repeat flow works with safeguards and observability

### Sprint 3: Action Plane Expansion + Hardware Path
Deliver:
- MCP + one non-MCP adapter parity
- hardware interface action scaffolds
- stronger recovery + replay/idempotency controls
- controlled skill-import pilot from trusted source with sandboxed activation
- marketplace/private-hub sync + rollback drills for installed skill/template capabilities

Exit criteria:
- governed host/hardware invocation tested and auditable

---

## 9. Test and Validation Program

### Core Layer Tests
1. Intent/Decision Layer
- path selection correctness
- approval_required behavior

2. Orchestration Layer
- team lifecycle transitions
- scheduler-triggered run behavior

3. Event Spine Layer
- decision/action events persisted and linked
- degraded and recovery event paths

4. Execution Layer
- adapter contract compliance
- idempotency/replay protection tests

5. Observability Layer
- run timeline/chain completeness
- diagnostics and next-step links present
6. Manifest Profile Layer
- objective-aligned improvement scoring correctness
- promotion/rollback guard behavior
7. Permanent Memory Loop Layer
- no-silence preference promotion enforcement
- Postgres/pgvector recall quality and drift rollback integrity
8. Central Agentry Governance Layer
- correction repetition threshold behavior and 3-strike prompt
- provenance metadata correctness (`file`, `line`) on learned rule application
- forget-X and forget-everything propagation + confirmation
- sensitive data exclusion in memory persistence path
9. Multi-User Group Governance Layer
- root-admin group contract validation and policy scoping
- group action/run audit integrity (`group_id` linkage)

### UI/E2E Mandatory Flows
1. direct action with approval path
2. team manifestation through wizard
3. repeat scheduling from existing run
4. degraded Ollama/NATS/SSE recovery without page reload
5. host/hardware action proposal and governed execution
6. heartbeat loop budget exhaustion triggers safe pause and audit event
7. self-update attempt is routed to staging workflow, never direct production mutation
8. unauthorized credential/probe behavior is denied and logged as security event
9. dynamic skill import fails closed when signature/scope/sandbox checks are not satisfied
10. single-use-case KPI panel shows operator-benefit metrics before autonomy expansion
11. manager-worker delegation trace appears for manifested and specialist routes
12. compaction checkpoint + resume preserves active objective without manual restitching
13. correction log entry can be reviewed, promoted, and rolled back through governed flow
14. manifest profile report shows objective-aligned improvements and rollback history
15. third repeated correction triggers scope-prompt (`global|domain|project`)
16. weekly maintenance digest reports dedupe/concise/archive changes
17. learned-rule application displays provenance (`file`, `line`)
18. forget commands complete and confirm (`forget X`, `forget everything`)
19. root-admin can define goal-scoped user group through Soma and execute within policy bounds
20. group actions include `group_id` in run timeline and governance evidence

### Release Gates
- Gate P0: contract + health + degraded recovery
- Gate P1: direct/manifest/repeat usability
- Gate P2: adapter expansion + host/hardware scaffolds
- Gate RC: full regression + evidence archive

---

## 10. Risks and Controls

### R1: Unbounded adaptation drift
Control: explicit review and rollback for profile/recipe promotions.

### R2: Cross-channel contract drift
Control: shared DTO package + schema tests; no ad hoc payloads.

### R3: Unsafe actuation escalation
Control: capability scopes, approval gates, audit chain, allowlist-only adapters.

### R4: Local model instability
Control: health checks, visible fallback state, policy-aware routing constraints.

### R5: Automation overhead exceeds operator benefit
Control: single-use-case rollout gate with explicit "worth it" metrics before permission expansion.

### R6: Long-horizon heartbeat loops run away
Control: bounded autonomy budgets (time, actions, spend, escalation) + global kill-switch.

### R7: Self-update bricks runtime
Control: staging-only update path, signed artifacts, health verification, and rollback checkpoints.

### R8: External skill injection introduces unsafe capability
Control: signed manifest verification, scope-limited sandbox activation, and governance approval before enablement.

### R9: Trojan-horse/credential overreach through local scans
Control: default deny on broad credential probing, deterministic secret redaction, and strict allowlist channel boundaries.

### Inception Deployment Discipline (Mandatory)

1. Start with one stable production use case only.
2. Keep financial/mission-critical permissions disabled until reliability thresholds are met.
3. Expand permissions only after two consecutive release gates pass with zero high-severity security regressions.

---

## 11. Handoff Package for Parallel Agent Teams

Each lane handoff must include:
1. changed files
2. API/schema diffs
3. tests added and pass output
4. known risks
5. rollback plan
6. operator-visible behavior change summary

Handoff readiness label:
- `READY_FOR_INTEGRATION`
- `BLOCKED_<reason>`
- `REQUIRES_POLICY_DECISION`

---

## 12. Definition of Done

This program is complete when:
1. Soma can reliably choose direct/team/propose with traceable rationale.
2. Team lifetime and repeat scheduling behave predictably.
3. Local Ollama health and fallback are visible and operational.
4. Universal action plane supports at least MCP + one non-MCP path in production shape.
5. Host/hardware actions are governed, auditable, and replay-safe.
6. All core-layer and E2E gate suites pass with evidence.

---

End of document.

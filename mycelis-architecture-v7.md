# Mycelis PRD — Orchestration Core, Event-Driven Mission Graph, and Non-Black-Box Observability
## Product Requirements Document (PRD) — Highly Detailed Implementation Doctrine

**Owner:** Principal Architect (Erik)  
**Audience:** Claude Dev Agentry (team spawner), Lead Engineer, Frontend, Backend, Security, QA  
**Primary References (Must Consult):**  
- `mycelis-architecture-v6.2.md` (Master State Authority)  
- `README.md` (Usage and general behavior)  
**Scope:** Product workflow and UI/UX aligned to Orchestration Templates, local-first providers, multi-user default, event-triggerable missions, and flight-recorder observability.  
**Non-Goal:** Building every “architecture surface” as a first-class user panel. We prioritize workflow usability.

---

# 0. Summary (Why This Exists)

Mycelis is a governed orchestration system (“Neural Organism”) where users express intent, Mycelis proposes structured plans, and any state mutation requires explicit confirmation plus a complete Intent Proof bundle. Missions are not isolated; they can emit events that trigger other missions. Observability is not optional: execution must never be a black box.

This PRD defines:
- core user workflows (conversation → proposal → confirm → execute → report),
- mission chaining via event triggers (end-of-mission and mid-stream),
- provider routing (local Ollama default, remote Claude optional),
- multi-user posture (Linux-like startup),
- and a unified monitoring model (run timeline / flight recorder + causal chains).

This document is specific enough that Claude dev agentry can spawn the right team, execute tasks, and prove completion.

---

# 1. Goals and Non-Goals

## 1.1 Goals
1. **Workflow-first product**: UI follows user workflows, not architecture surfaces.
2. **Two-phase mutation**: No state change without Proposal → Confirm Token → Intent Proof → Execute.
3. **Event-driven mission graph**: Missions emit structured events; triggers link missions.
4. **Non-black-box execution**: Every run has a visible timeline; every trigger is traceable.
5. **Local-first**: Default provider is local Ollama; remote providers are explicit, gated, and visible.
6. **Multi-user default**: Every install supports multiple users; single-user is “create one user”.
7. **Security-first**: Tool gating, MCP install approval, scope restrictions, audit logging.

## 1.2 Non-Goals (for this PRD cycle)
- Full enterprise RBAC matrix and SSO integrations (later)
- Full DAG editor as primary UX (advanced view only)
- Prometheus/Grafana-grade telemetry dashboards (later)
- Automatic policy changes or automatic tool installs
- “Magic autonomous execution” without confirm/approval

---

# 2. Personas and Roles

## 2.1 Personas
- **Owner/Admin**: Sets up providers, policies, tools, and enables remote.
- **Operator**: Runs missions, confirms proposals, manages automations within policy.
- **Viewer**: Observes, audits, reads outcomes. Cannot mutate state.

## 2.2 Multi-User Default (“Linux Startup”)
- First-run creates an Owner/Admin.
- System supports additional users and roles from day one.
- Single-user = only one account created.

---

# 3. User Workflows (End-to-End)

## 3.1 Workflow A — Conversation → Answer (Read-Only)
**User story:** “I ask questions and get trustworthy responses with provenance.”

### Steps
1. User opens **Mission Control**.
2. User sends message to Soma/council member.
3. System classifies intent as Answer (no mutation).
4. System responds with:
   - answer
   - provenance (brain/provider, consult chain)
   - trust/confidence
   - tools used (if any)
5. User can open **Inspector** for details.

### Success criteria
- Response clearly indicates: Mode = Answer, Brain used, Role used, Provenance available.
- Audit event recorded and linked to response.

---

## 3.2 Workflow B — Conversation → Proposal → Confirm → Execute (Mutation)
**User story:** “I ask the system to do something, it proposes safely, I confirm, it executes.”

### Steps
1. User sends actionable request.
2. System detects mutation risk (tool risk / operation risk).
3. System responds with **Proposed Action Block**:
   - resolved intent
   - operations list
   - tools involved
   - providers/roles planned
   - scope constraints
   - governance requirement status
4. User clicks **Confirm & Execute**.
5. System validates confirm token.
6. System creates **Intent Proof** bundle.
7. Execution starts and emits events.
8. System ends with **Execution Report**.
9. Artifacts and audit links are visible.

### Success criteria
- No execution occurs without confirm token validation.
- Intent Proof ID + Audit Event ID visible to user.
- Run timeline shows what happened and why.

---

## 3.3 Workflow C — Create Automation (Scheduled Mission)
**User story:** “I want recurring execution on a schedule with governance and visibility.”

### Steps
1. User asks for scheduled behavior (e.g., “Daily research and summarize”).
2. System generates proposal including schedule constraints and concurrency guard.
3. Confirm required.
4. Mission is created with schedule.
5. Runs emit events; outcomes visible in timelines and automations view.
6. User can pause/disable/edit (via proposal flow).

### Success criteria
- Schedules enforce safety: min interval, concurrency limits, recursion detection.
- Each run has a trace; failures and retries visible.

---

## 3.4 Workflow D — Mission Chaining (Event-Triggered Missions)
**User story:** “When research completes, generate variant media; when a signal appears, trigger a review.”

### Steps
1. Mission A emits events during execution and on completion.
2. Trigger rules evaluate events.
3. If matched, system:
   - creates proposal for downstream mission OR executes automatically if policy allows
4. Trigger firing is logged and visible:
   - parent mission run → event → trigger → child mission run
5. User can view chain as a causal graph/timeline.

### Success criteria
- Triggers are visible, inspectable, and auditable.
- No hidden cascades.
- Governance can require approval for trigger-based mutations.

---

# 4. Product Information Architecture (Workflow-First Navigation)

## 4.1 Primary Navigation (User-Facing)
1. **Mission Control** (primary): chat, proposals, confirms, execution reports, timelines
2. **Automations**: scheduled missions + triggers + drafts + approvals (workflow focus)
3. **Resources**: brains/providers + tool library + service targets + capabilities
4. **Memory**: what the organism knows, sources, recall events, scoped search
5. **System** (advanced): health, event health, debug (hidden behind “Advanced”)

## 4.2 Remove/De-emphasize as top-level panels (move under Automations/Resources/System)
- Neural Wiring (advanced view inside Automations)
- Cognitive Matrix (System or Resources)
- Skills Market (Resources)
- Agent Catalogue (Resources)
- Governance (integrated into Automations + Mission Control + System)

**Requirement:** Panels must not show meaningless error states. If dependencies offline, show a clear degraded message and what still works.

---

# 5. Orchestration Templates (First-Class Primitive)

## 5.1 Template Families (v1)
- **Chat-to-Answer** (read-only)
- **Chat-to-Proposal** (mutation gating)
- **Confirm-Action** (token validation + proof)
- **Execute-Action** (state mutation & event emission)
- **Schedule** (recurring execution)
- **Trigger-Rule** (event routing)

## 5.2 Two Mandatory v1 Templates
1. `chat_to_answer.v1`
2. `chat_to_proposal.v1`

**Rule:** If ambiguous, default to Proposal (least privilege).

---

# 6. Event System (Mission Graph + Observability Spine)

## 6.1 Principle
**Events are both:**
- orchestration triggers
- monitoring visibility

If we have events, we automatically have observability.

## 6.2 Canonical Event Envelope: `MissionEventEnvelope`
**Required fields**
- `event_id` (uuid)
- `timestamp`
- `tenant_id`
- `user_id`
- `mission_id`
- `run_id` (unique per execution)
- `event_type` (enum)
- `severity` (info/warn/error)
- `provider_id` (ollama/claude/etc)
- `model_used`
- `role`
- `mode` (answer/proposal/execution)
- `payload` (typed JSON)
- `artifact_refs[]` (ids/urls)
- `cause` (optional)
  - `parent_event_id`
  - `parent_run_id`
  - `trigger_rule_id`
- `audit_event_id` (link to authoritative audit record)
- `intent_proof_id` (if mutation path)

## 6.3 Minimum Event Types (v1)
- `mission.started`
- `mission.completed`
- `mission.failed`
- `mission.step.started`
- `mission.step.completed`
- `tool.invoked`
- `tool.completed`
- `tool.failed`
- `policy.evaluated`
- `approval.requested`
- `approval.granted`
- `approval.denied`
- `trigger.rule.evaluated`
- `trigger.fired`
- `artifact.created`
- `memory.recalled`
- `memory.stored`

## 6.4 Trigger Rules (Simple, Declarative)
A trigger rule is:

**IF** `event_type` matches AND optional predicates pass  
**THEN** initiate downstream mission (proposal or execution)

**Rule format (v1 fields)**
- `trigger_rule_id`
- `tenant_id`
- `enabled`
- `source_mission_id` (optional, can be any)
- `event_type`
- `predicate` (optional JSON logic, minimal)
- `action`:
  - `target_mission_blueprint_id` OR `target_mission_template`
  - `mode`: propose | execute
  - `requires_approval`: boolean override (or derived from policy)
- `cooldown_ms`
- `recursion_guard`:
  - max_depth
  - cycle_detect (on/off)
- `concurrency_guard`:
  - max_active_runs_for_target

**Important:** Start simple. Do not build a full rules engine in v1.

---

# 7. Observability UX (Non-Black-Box)

## 7.1 “Run Timeline” (Flight Recorder) — Required UI Component
Every mission run renders a vertical timeline:

- Policy decision events
- Approval events
- Provider routing decisions
- Tool invocations and results
- Trigger firings
- Artifact creation
- Completion/failure

Each item is expandable and links to:
- audit event
- related artifacts
- trigger rule
- child run (if fired)

## 7.2 “Causal Chain View” — Required
A button: **View Chain**
Shows:

Mission A Run → Event → Trigger → Mission B Run → Event → Trigger → Mission C Run

This can be a simple expandable list in v1.
Graph view is optional later.

## 7.3 “Event Health” (Replace empty telemetry dashboards)
System page must show:
- NATS connected/disconnected
- events/sec
- trigger queue length
- failed triggers count
- last error summary
- degraded mode notices (what works / what doesn’t)

**If offline:** show clear degraded capability statement.

---

# 8. Provider Routing (Local-First, Transparent)

## 8.1 Provider Policy Model
- Provider: id, location (local/remote), data_boundary, enabled, usage_policy, roles_allowed
- Default provider: local Ollama
- Remote provider: disabled by default; enable requires explicit modal confirmation

## 8.2 Routing Decision Record
Every chat/run records:
- provider chosen
- model used
- reason (policy/user/escalation)
- data boundary classification
- constraints applied

## 8.3 UI Requirements
- Mode ribbon always shows:
  - MODE / ROLE / BRAIN / GOV
- Per-message header shows:
  - provider + local/remote badge
  - role
  - confidence/trust
- Inspector reveals:
  - provider/model routing reasons
  - tool chain

---

# 9. Security & Governance Requirements (Workflow-Scoped)

## 9.1 No mutation without confirm + proof
- Confirm endpoint must reject invalid/replayed/mismatched tokens.
- Mutation tools must be gated by allowlist and scope.
- MCP install requires approval.
- Filesystem is sandboxed; no arbitrary paths.

## 9.2 Governance integration
If policy requires approval:
- Proposal is generated
- Execution blocked until approval granted
- Approval events emitted and appear in timeline

## 9.3 Degraded mode rules (offline components)
If NATS offline:
- No trigger firing
- No scheduled execution
- Chat-to-Answer may still operate if inference available
- UI must show offline and which functions are disabled

---

# 10. Functional Requirements (Detailed)

## 10.1 Mission Control (Primary Surface)
Must support:
- Conversation
- Proposal cards
- Confirm action
- Execution reports
- Run timelines per run
- Inspector per message
- Clear offline/degraded messaging

## 10.2 Automations
Must support:
- Active automations list (scheduled missions)
- Trigger rules list
- Draft blueprints list
- Approvals queue (if used)
- Per automation:
  - emits (event types)
  - triggers (rules)
  - recent runs + timeline access

## 10.3 Resources
Must support:
- Brains/providers management
- MCP tools/library management
- Service targets/capabilities
- Access policies per role (even if minimal v1)

## 10.4 Memory
Must support:
- tenant-scoped semantic search
- show sources ingested
- show recall events
- show store events
- link memory usage to run timelines

## 10.5 System (Advanced)
Must support:
- Event health
- Core status
- NATS status
- DB status
- “what is degraded” messages

---

# 11. UX Requirements (Strict)

## 11.1 No meaningless empty panels
If a panel cannot function due to offline core:
- must show:
  - why
  - what is unavailable
  - what remains usable
  - next step to restore

## 11.2 Reduce cognitive overload
- Workflows first; advanced views hidden behind toggles.
- Default path: Mission Control → Proposal → Confirm → Report.

## 11.3 Avoid architecture-surface navigation
Navigation must reflect user tasks, not internal architecture layers.

---

# 12. Implementation Plan (Team Spawn Guidance)

Claude dev agentry must spawn teams with zero overlap:

## Team A — Event Emission + Storage (Backend)
- Implement MissionEventEnvelope (schema)
- Emit minimum event set at key points
- Persist events and link to audit/intent proof
- Expose event query APIs:
  - get run timeline by run_id
  - get chain by intent_proof_id / parent_event_id
- Ensure degraded behavior when NATS offline

## Team B — Trigger Rules Engine (Backend)
- Implement trigger rules data model (simple)
- Evaluate rules on event ingest
- Enforce cooldown, recursion guard, concurrency guard
- On match:
  - propose downstream OR execute if policy allows
- Emit trigger events: evaluated, fired

## Team C — Run Timeline + Chain UI (Frontend)
- Build RunTimeline component
- Wire to event query APIs
- Build ViewChain UI (list-based v1)
- Integrate into Mission Control and Automations pages

## Team D — Information Architecture & Panel Consolidation (Frontend)
- Collapse navigation to workflow-first IA
- Move/merge panels under Automations/Resources/System
- Replace error pages with degraded-mode messaging
- Hide advanced panels behind “Advanced” toggle

## Team E — Acceptance Tests + Verification Harness (QA/Dev)
- Add integration tests:
  - proposal → confirm → execute emits events
  - trigger fired creates downstream run
  - degraded mode prevents triggers when NATS offline
  - remote provider visibility and gating
- Provide manual verification scripts + screenshots

---

# 13. Acceptance Criteria (Proven Completion)

## 13.1 Workflow Proof
- User can:
  - ask question → get answer with provenance
  - request action → receive proposal card
  - confirm → execution starts
  - see outcome report
  - view run timeline
  - see triggered downstream mission

## 13.2 Non-Black-Box Proof
- For any run:
  - timeline shows policy decisions, tool usage, trigger firings, artifacts, completion
  - inspector links to audit and intent proof
- Trigger firings show causal links:
  - parent run → event → trigger rule → child run

## 13.3 Safety Proof
- No mutation without confirm token validation
- Trigger loops prevented by guards
- NATS offline disables triggers and schedules safely with clear UI notice
- Remote provider use is explicit, gated, and visible

## 13.4 UX Proof
- Navigation is reduced to workflow-first panels
- No major panel shows blank errors without explanation
- Advanced internals are not front-and-center for normal users

---

# 14. Verification Steps (Manual Script)

1. Start system with NATS online.
2. Mission Control: ask a read-only question.
   - Verify provenance shown; inspect opens; audit ID exists.
3. Mission Control: request a mutation action.
   - Verify proposal card; confirm token present.
4. Confirm action.
   - Verify intent proof + audit IDs displayed.
5. Verify run timeline shows:
   - mission.started → tool.invoked → artifact.created → mission.completed
6. Create a trigger rule:
   - IF mission.completed THEN start downstream mission (propose or execute per policy)
7. Run mission again.
   - Verify trigger evaluated + fired events.
   - Verify downstream mission run created and chain view links them.
8. Stop NATS / simulate offline.
   - Verify triggers do not fire; UI shows degraded mode.

---

# 15. Open Decisions (Must Resolve Before Full Build)

1. Trigger execution mode default:
   - propose-only (safer)
   - or allow auto-execute for low-risk triggers under policy
2. Event storage volume strategy:
   - retention duration per tenant
   - summarization (later)
3. Artifact model:
   - minimum metadata fields
   - storage location pattern

---

# 16. Deliverable Format From Claude Dev Agentry

Return output as:

1. Architecture alignment notes (what docs were referenced)
2. Team spawn plan (A–E) with file ownership and boundaries
3. Incremental implementation tasks (small steps, verifiable)
4. APIs to add/change (routes, payloads)
5. UI components to add/change (file paths, state)
6. Tests + manual verification scripts
7. “Definition of Done” mapped to acceptance criteria

---

End of PRD.
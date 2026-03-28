# V8 Memory Continuity And RAG Strike Team Plan

> Status: ACTIVE
> Last Updated: 2026-03-28
> Purpose: Coordinate the cross-team review and delivery plan for team-scoped vector memory, temporary planning continuity, and trace-clean execution context.

## Mission

Mycelis must preserve a clean distinction between:

1. durable semantic memory that belongs in pgvector-backed recall,
2. temporary planning continuity that should survive restarts without polluting long-term memory,
3. trace and audit records that exist for causality and review, not as planning memory.

This strike plan exists to align runtime, API, UI, docs, and QA around that boundary before release promotion.

## Current Review Summary

What is already strong:

- PostgreSQL + pgvector support is still present and active through `context_vectors`.
- durable semantic memory already exists for SitReps, remembered facts, conversation summaries, and inception recipes.
- temporary restart-safe memory already exists through `temp_memory_channels`.
- conversation tracing already exists through `conversation_turns`.
- mission/audit causality already exists through `mission_events` and centralized review through `log_entries`.

What is currently misaligned:

- semantic search is still effectively global unless the caller manually constrains it outside the current handler path.
- team-scoped and temporary-team memory ownership is not enforced clearly enough in the retrieval contract.
- automatic conversation summarization currently risks promoting exploratory planning into durable vector memory.
- docs already describe stronger recall boundaries than the runtime consistently enforces today.

## Strike Team

### 1. Program Planner and Synchronization Lead

Primary role:
- own the full dependency map
- translate findings from every lane into concrete deliverables
- keep runtime, docs, UI, and QA in sync
- block merges when one lane changes behavior without matching validation or documentation

Required behaviors:
- maintain the canonical task breakdown
- track cross-lane impacts before code lands
- keep acceptance criteria explicit and testable
- maintain a release-facing change log for the whole slice

Primary deliverables:
- final scoped implementation plan
- cross-lane dependency checklist
- acceptance gate scoreboard
- merge order and release packaging recommendation

### 2. Runtime Memory and RAG Lead

Owns:
- `core/internal/memory/*`
- `core/migrations/*` when memory schema changes are needed
- durable pgvector retrieval and storage rules

Primary goals:
- ensure durable vector memory supports team-aware planning and recall
- ensure memory metadata is rich enough for scope-safe retrieval
- preserve backward compatibility for existing stored records where possible

Primary deliverables:
- scoped vector query contract
- durable memory metadata normalization
- tests for scoped semantic recall

### 3. Swarm Continuity and Agent Behavior Lead

Owns:
- `core/internal/swarm/*`
- internal tool behavior for `remember`, `recall`, `search_memory`, and temporary continuity

Primary goals:
- ensure Soma, Council, and team leads use durable memory only for reusable knowledge
- ensure exploratory planning and working context use temporary continuity lanes instead of polluting long-term recall
- ensure temporary teams can still benefit from scoped continuity

Primary deliverables:
- explicit memory-class behavior in internal tools
- agent-context guidance that reflects the new boundaries
- tests for scoped recall and temporary planning continuity

### 4. API and Operator Experience Lead

Owns:
- `core/internal/server/*` memory endpoints
- `interface/components/memory/*`
- any operator-facing workflow that depends on memory recall

Primary goals:
- make the memory API honest about scope and result class
- ensure UI search and operator surfaces reflect what is durable, scoped, and trace-only
- avoid turning the UI into a raw memory-debug console

Primary deliverables:
- query/filter API contract
- operator-visible scope semantics
- UI/API contract update where behavior changes

### 5. Trace Cleanliness and Governance Lead

Owns:
- audit/logging interaction review
- conversation trace vs memory promotion rules
- release commentary for what should and should not become durable knowledge

Primary goals:
- keep planning chatter, draft thinking, and exploratory user concept work out of general promoted vector memory unless explicitly promoted
- preserve audit and conversation causality without turning trace storage into semantic memory by default
- keep governance language aligned with memory behavior

Primary deliverables:
- memory-class policy notes
- code comments where behavior would otherwise look arbitrary
- matching docs for trace vs temporary vs durable usage

### 6. Verification and Release Lead

Owns:
- focused backend tests
- live browser/API proof where memory behavior is operator-visible
- final release gate summary

Primary goals:
- prove the runtime matches the plan from committed state
- prove scope-safe recall works
- prove exploratory planning does not accidentally become durable memory

Primary deliverables:
- validation matrix
- release pass/fail verdict
- residual-risk note for anything intentionally deferred

## Canonical Deliverables

The team is done only when all of the following are true:

1. durable vector memory can be scoped by team and related provenance where required.
2. temporary planning continuity is clearly available without durable promotion.
3. exploratory conversation and draft planning are not automatically promoted into the shared pgvector memory substrate.
4. tracing and audit surfaces remain intact and clearly separated from promoted memory.
5. Soma, Council, and team leads can still retrieve relevant continuity during planning.
6. docs, UI wording, and runtime behavior all describe the same memory model.

## Memory Class Model

### Durable Memory

Use for:
- reusable facts
- decisions
- lessons learned
- SitReps
- approved or intentionally promoted summaries
- recipes and reusable procedures

Storage expectation:
- structured record plus pgvector metadata where appropriate

Planning expectation:
- safe for future recall across runs within allowed scope

### Temporary Continuity

Use for:
- in-flight planning checkpoints
- working assumptions
- partial execution context
- restart-safe collaboration context
- draft thinking that is useful for current work but not yet worthy of promotion

Storage expectation:
- `temp_memory_channels`
- bounded TTL or explicit clear path
- no automatic durable vector promotion

Planning expectation:
- readable by the relevant lead or team during current work

### Trace and Audit

Use for:
- conversation history
- mission causality
- approvals
- tool execution evidence
- operational review

Storage expectation:
- `conversation_turns`
- `mission_events`
- `log_entries`

Planning expectation:
- supports replay, review, and causality
- does not become semantic continuity by default

## Execution Sequence

### Phase 1: Truth Mapping

Tasks:
- map all current durable memory writers
- map all current temporary memory writers
- map all current trace-only writers
- identify where automatic promotion currently occurs

Exit criteria:
- planner publishes a touched-surface inventory
- each surface is labeled `durable`, `temporary`, `trace`, or `mixed`

### Phase 2: Runtime Boundary Repair

Tasks:
- tighten vector search so team-aware and provenance-aware retrieval is possible
- normalize durable memory metadata for future recall filtering
- preserve compatibility with existing unscoped records where practical

Exit criteria:
- scoped semantic recall exists in runtime code
- tests prove filters work and do not silently widen scope

### Phase 3: Agent Continuity Repair

Tasks:
- change automatic conversation checkpointing so it favors temporary continuity over durable promotion
- keep explicit durable promotion available through deliberate tool use
- ensure team leads can still see relevant recent continuity in runtime context

Exit criteria:
- exploratory planning no longer auto-pollutes pgvector
- continuity remains available to active planning paths

### Phase 4: API and UI Contract Sync

Tasks:
- expose honest memory-query scope in API
- make UI memory surfaces reflect scoped semantic search cleanly
- update wording so operators understand durable vs temporary vs trace behavior

Exit criteria:
- API docs and UI behavior match runtime
- no user-facing copy implies that all chat is promoted into durable memory

### Phase 5: Release Proof

Tasks:
- run focused backend memory tests
- run docs-link and typecheck gates
- run browser/API proof where memory behavior is visible
- do one final planner review across code, docs, and tests

Exit criteria:
- committed-state validation is green
- planner marks all deliverables complete or explicitly deferred

## Non-Negotiable Rules

1. No new memory behavior lands without a named memory class.
2. No automatic durable promotion lands without an explicit product reason.
3. No team-scoped recall feature ships if the runtime still widens to global results without warning.
4. No doc may describe filtered or governed recall behavior that the runtime does not actually provide.
5. No release-ready claim is accepted until tests prove the new boundary from committed state.

## Required Test Proof

### Backend

- durable vector search with no filters still works for legacy records
- team-scoped vector search returns matching team content
- team-scoped vector search does not leak mismatched-team content
- temporary continuity storage and retrieval remain intact
- conversation checkpoint behavior uses temporary continuity where intended
- trace/audit persistence remains intact

### API

- memory search accepts scoped query parameters correctly
- invalid scope combinations fail clearly
- temp-memory endpoints continue to work

### Swarm

- `remember` stores scoped durable metadata when invoked from a team context
- `recall` defaults to safe scoped retrieval when planning from a team context
- automatic conversation checkpointing does not write to durable vector memory unless intentionally promoted

### Interface

- memory search UI still renders results correctly
- scope-aware calls do not break the current operator flow
- operator copy distinguishes durable memory from temporary continuity when exposed

## Planner Review Checklist

The Program Planner and Synchronization Lead must sign off on:

- touched file list is complete
- docs changed where meaning changed
- unchanged docs were reviewed for drift
- validation matrix reflects actual changed behavior
- no lane silently changed the product contract
- release note language matches implementation truth

## Target Output For This Slice

When this strike team is complete, Mycelis should support the following product truth:

- teams can rely on pgvector-backed knowledge during planning without accidentally querying unrelated memory first
- temporary teams can use continuity generated by Soma and Council without forcing everything into shared long-term memory
- exploratory user concept work stays useful for current planning while remaining cleanly separated from durable organizational memory
- tracing remains complete and reviewable without becoming a hidden long-term semantic memory channel

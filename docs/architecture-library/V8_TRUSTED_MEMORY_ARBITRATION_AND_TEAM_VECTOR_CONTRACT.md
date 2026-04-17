# V8 Trusted Memory Arbitration And Team Vector Contract
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-04-17
> Purpose: Define the trusted memory architecture for Soma personal continuity, team-shared vector memory, governed swarm memory, anchor verification, arbitration, and bounded growth.

## TOC

- [Core Principle](#core-principle)
- [Current Repo Reality](#current-repo-reality)
- [Target Memory Topology](#target-memory-topology)
- [Trusted Memory Record Model](#trusted-memory-record-model)
- [Anchor And Verification Contract](#anchor-and-verification-contract)
- [Retrieval And Arbitration](#retrieval-and-arbitration)
- [Growth And Retention](#growth-and-retention)
- [Team Delivery Architecture](#team-delivery-architecture)
- [Implementation Sequence](#implementation-sequence)
- [Proof Plan](#proof-plan)
- [Open Risks](#open-risks)

## Core Principle

Vector memory is an index, not a sovereign truth source.

The system may retrieve semantically relevant memories from `context_vectors`, but no retrieved memory should become execution-authoritative unless it can be:
- scoped
- classified
- anchored to deterministic evidence
- assigned an explicit trust posture
- compared against higher-order policy or newer evidence

Postgres event history remains the deterministic truth layer.

The vector layer is a recall surface for claims about that truth.

## Current Repo Reality

The repo already has the right bones:
- `context_vectors` is the shared pgvector substrate for durable semantic recall
- `agent_memories` stores structured scoped memory
- `conversation_summaries` stores summarized continuity
- `temp_memory_channels` provides restart-safe short-horizon continuity
- `LearningCandidate` already exists as the candidate-first exchange boundary for promoted learning

The current gap is not “we have no memory system.”

The real gap is that Soma personal memory, team-shared memory, and global governed memory are still mostly conventions over shared tables and metadata. The runtime can scope retrieval, but it does not yet anchor, arbitrate, supersede, prune, or halt on contradictions as a first-class control plane.

## Target Memory Topology

Keep the canonical four-layer model stable:
- `SOMA_MEMORY`: Soma's personal durable continuity and leadership heuristics
- `AGENT_MEMORY`: team-shared and specialist-shared execution memory
- `PROJECT_MEMORY`: governed source context such as `customer_context`, `company_knowledge`, and `user_private_context`
- `REFLECTION_MEMORY`: synthesized lessons, contradictions, trajectory shifts, and other promoted reflection artifacts

Map the desired operating model onto those existing layers:
- Soma local vector DB = `SOMA_MEMORY`
- team shared contextual vector DB = `AGENT_MEMORY` with explicit team scope and Soma read access
- global swarm culture / shared doctrine = governed promotion targets inside `PROJECT_MEMORY` and `REFLECTION_MEMORY`, especially `company_knowledge`, `soma_operating_context`, and `reflection_synthesis`

This avoids inventing a fifth top-level memory class while still supporting:
- private Soma continuity
- shared team execution context
- curated organization-wide doctrine and learned experience

Storage rule:
- keep `context_vectors` as the embedding substrate
- add a typed control plane above it for authority, verification, and retention behavior
- do not create one new vector table per lane unless performance data later proves it is necessary

## Trusted Memory Record Model

Add a first-class durable `memory_records` control-plane table.

Each durable memory claim should have one `memory_records` row, with optional linkage to:
- a `context_vectors` row
- an `agent_memories` row
- a `conversation_summaries` row
- an artifact or exchange item

Required record fields:
- `id`
- `tenant_id`
- `memory_layer`
- `knowledge_class`
- `owner_kind` (`soma`, `team`, `agent`, `system`)
- `owner_id`
- `team_id`
- `agent_id`
- `run_id`
- `visibility`
- `trust_class`
- `sensitivity_class`
- `continuity_key`
- `content_hash`
- `dedupe_key`
- `promotion_state` (`candidate`, `reviewed`, `promoted`, `superseded`, `expired`, `rejected`)
- `verification_status` (`unverified`, `anchored`, `verified`, `conflicted`, `rejected`)
- `confidence`
- `review_required`
- `supersedes_memory_id`
- `superseded_by_memory_id`
- `expires_at`
- `last_accessed_at`
- `access_count`
- `created_at`
- `updated_at`

Important modeling rule:
- `context_vectors` remains optimized for similarity search
- `memory_records` becomes the source of truth for scope, authority, verification, and lifecycle state

This gives maintainability without multiplying storage primitives.

## Anchor And Verification Contract

Every promoted durable memory must carry at least one immutable anchor back to deterministic evidence.

Add `memory_event_anchors` as a join model from `memory_records` to source evidence. Anchors may point at:
- `log_entries`
- `conversation_turns`
- exchange items
- artifacts
- mission events
- deployment context artifacts

Required anchor fields:
- `memory_record_id`
- `anchor_kind`
- `source_table`
- `source_id`
- `source_timestamp`
- `source_hash`
- `verification_method`
- `verified_at`
- `verified_by`

Verification rules:
- unanchored temporary working memory may exist, but it is never execution-authoritative
- anchored but unreviewed memory may be retrieved as context, but it must be marked as provisional
- verified memory may shape execution within its scope
- conflicted memory must not silently win retrieval; it must either defer to the higher-order source or trigger review

Hallucination rule:
- if a memory claims prior success, policy, or precedent and the runtime cannot verify at least one valid anchor, that memory is downgraded to `verification_status=unverified` and cannot be used as authoritative justification for mutation or policy-shaped action

## Retrieval And Arbitration

Retrieval must become a two-stage process:

1. semantic recall
- query `context_vectors` and structured memory sources for candidate hits

2. arbitration
- resolve candidate hits through `memory_records`, anchors, scope, trust, recency, and policy precedence

Introduce a `Memory Arbitration Service` that assigns final precedence in this order:

1. deterministic source truth
- event logs, mission events, conversation turns, explicit artifacts, persisted exchange items

2. governed organization memory
- `company_knowledge`
- `soma_operating_context`
- approved `reflection_synthesis`

3. team-shared execution memory
- `AGENT_MEMORY` scoped to the active team

4. Soma personal memory
- `SOMA_MEMORY`

5. temporary or candidate memory
- temp channels, unreviewed candidates, unverified continuity notes

Conflict rules:
- if lower-order memory conflicts with higher-order governed or deterministic truth, higher-order memory wins
- if team memory conflicts with current global doctrine, halt mutation and raise review or proposal
- if Soma personal memory conflicts with team-shared execution memory, team memory wins for team execution and Soma memory remains advisory
- if the conflict concerns style, relationship dynamics, or local working preference rather than policy or factual state, Soma personal memory may win inside its scope

Soma access rule:
- Soma may read team-shared memory for teams it leads or supervises
- Soma may not silently widen a team memory into organization-wide doctrine
- promotion to global doctrine must pass through governed promotion into `PROJECT_MEMORY` or `REFLECTION_MEMORY`

## Growth And Retention

The system must retain usefulness without becoming a landfill.

Required controls:
- budget per Soma lane and per team lane
- TTL for temporary continuity and unreviewed candidates
- semantic dedupe on `dedupe_key` and `content_hash`
- supersession rather than blind append for repeated SOP-like memories
- periodic compaction of low-access, stale, or superseded entries
- re-verification window for governed doctrine that depends on time-sensitive operating assumptions

Preferred defaults:
- temp memory: short TTL, aggressive cleanup
- team-shared memory: compact active set plus archive/supersession
- Soma personal memory: compact durable set with periodic review of stale heuristics
- governed doctrine: lower churn, but explicit supersession and change history

## Team Delivery Architecture

The execution model should stay compact but specialized.

| Lane | Ownership | Immediate target |
| --- | --- | --- |
| Memory Architecture Lead | taxonomy, storage contract, arbitration model | keep the four-layer model stable while making `AGENT_MEMORY` explicitly team-shared and defining the precedence ladder |
| Persistence Lead | schema and lifecycle | add `memory_records`, anchor linkage, supersession, and retention primitives without fragmenting the vector substrate |
| Runtime Lead | retrieval, write paths, and policy behavior | introduce arbitration, conflict halt behavior, and scoped promotion flow |
| Mother Brain Lead | synthesis and candidate translation | convert raw lessons into typed `LearningCandidate` items with anchors, confidence, and promotion intent |
| Governance Sentry | review posture and approval policy | keep policy-shaping promotion and high-risk memory mutation governed and reviewable |
| Operator UX Lead | visibility and debugging | expose memory layer, trust, anchor status, and supersession in advanced UI surfaces |
| Validation Lead | tests and red-team proof | prove the system halts on contradiction, rejects unanchored authority claims, and keeps memory growth bounded |

Coordination rule:
- Soma leads the compact program
- standing teams own their scoped memory lanes
- Mother Brain curates cross-team experience into governed doctrine
- no team gets a hidden sovereign memory surface that bypasses anchors, review posture, or arbitration

## Implementation Sequence

1. `REQUIRED`: document the trusted-memory topology and arbitration order across canonical and user-facing docs.
2. `NEXT`: add `memory_records` and `memory_event_anchors` schema plus read models that link existing durable memory rows to those records.
3. `NEXT`: add typed helpers for creating anchored `LearningCandidate` items so teams do not hand-build promotion payloads.
4. `NEXT`: route team-shared memory writes through `memory_records` so `AGENT_MEMORY` becomes a first-class team lane rather than metadata convention only.
5. `NEXT`: route Soma personal memory writes through the same control plane so `SOMA_MEMORY` becomes first-class and auditable.
6. `NEXT`: introduce retrieval-time arbitration and conflict halts before any policy-shaping or mutating action.
7. `NEXT`: add compaction, supersession, and budget enforcement jobs for durable memory.
8. `NEXT`: add advanced UI inspection for anchor state, verification posture, and supersession history.
9. `NEXT`: add browser and API proof for the end-to-end path from event -> candidate -> review -> promoted memory -> governed retrieval.

## Proof Plan

Required proof:
- unit tests for arbitration precedence and conflict outcomes
- unit tests for anchor verification and unverified-memory downgrade
- API tests for candidate publication with anchor metadata
- API tests for promotion into team-shared, Soma personal, and governed global lanes
- integration tests proving team-shared memory is visible to Soma but not silently promoted globally
- integration tests proving newer governed doctrine beats older team or Soma memory
- browser proof showing anchor state, review posture, and conflict halt messaging

Critical negative tests:
- raw transcript cannot become authoritative durable memory without anchor linkage
- unverified memory cannot justify a governed mutation
- team memory cannot overwrite `soma_operating_context`
- Soma personal memory cannot outrank governed doctrine for policy-shaped behavior
- superseded memory cannot remain the active answer when a verified newer replacement exists

## Open Risks

- The repo already stores many important scope fields inside vector metadata, so migration must avoid breaking current retrieval while introducing typed control-plane records.
- Candidate-first reflection exists in doctrine, but a dedicated Mother Brain runtime path does not yet exist.
- Time-sensitive doctrine will need re-verification rules or it will fossilize into bad advice.
- If compaction is too aggressive, teams will lose useful continuity; if compaction is too weak, retrieval quality will decay under memory bloat.

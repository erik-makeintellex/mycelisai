# V8 Memory Layer And Reflection Delivery Contract
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-04-17
> Purpose: Define the explicit memory layer taxonomy, promotion rules, team-shared memory posture, and reflection-first learning contract without weakening governance.

## TOC

- [Core Principle](#core-principle)
- [Explicit Memory Layers](#explicit-memory-layers)
- [Promotion Rules](#promotion-rules)
- [Reflection First, Not Recall First](#reflection-first-not-recall-first)
- [Managed Exchange Boundary](#managed-exchange-boundary)
- [Governance Guardrails](#governance-guardrails)
- [Trusted Retrieval And Arbitration](#trusted-retrieval-and-arbitration)
- [Team Delivery Architecture](#team-delivery-architecture)
- [Implementation Sequence](#implementation-sequence)
- [Test Plan](#test-plan)
- [Open Risks](#open-risks)

## Core Principle

Mycelis must not let memory become a casual transcript sink.

The system should first understand what matters, classify it, attach confidence, preserve evidence, and only then promote it into the right memory layer.

No path should go directly from interaction to durable memory without:
- classification
- confidence
- explicit review posture
- source or evidence linkage where available

## Explicit Memory Layers

The canonical layer names are:

| Layer | Purpose | Current backing | Write posture |
| --- | --- | --- | --- |
| `SOMA_MEMORY` | Soma-owned continuity, orchestrator facts, and reviewed Soma-level operating recall. | `remember` / `recall` / `search_memory`, plus the governed `soma_operating_context` sublane for root-admin shared Soma behavior. | Classified, scoped, and review-aware. Admin-shaped behavior must stay governed. |
| `AGENT_MEMORY` | Team-shared and specialist-shared continuity, decisions, role-specific observations, and execution lessons. | Scoped durable memory rows/vectors with `team_id`, `agent_id`, and visibility metadata. This is the canonical lane for team-shared vector memory. | Team/agent scoped by default. Wider promotion requires review. |
| `PROJECT_MEMORY` | Source context for the work: user-private records, customer docs, approved company knowledge, and deployment context. | Governed context vectors such as `user_private_context`, `customer_context`, and `company_knowledge`. | Operator or governed workflow intake only. It is RAG/source context, not agent identity. |
| `REFLECTION_MEMORY` | Distilled lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations about what is changing over time. | Managed Exchange `LearningCandidate` first; promotion target is `reflection_synthesis` only after classification, confidence, and review posture are explicit. | Candidate-first. Private/restricted by default. |

Layer separation rules:
- `SOMA_MEMORY` must not absorb project documents just because Soma saw them in chat.
- `AGENT_MEMORY` must remain a team/shared execution lane until explicitly promoted.
- `PROJECT_MEMORY` must not become company knowledge without governed promotion.
- `REFLECTION_MEMORY` must not store raw transcripts; it stores synthesized meaning with evidence references.
- `soma_operating_context` is a governed Soma sublane, not ordinary chat memory.

Interpretation rule:
- Soma personal continuity belongs in `SOMA_MEMORY`.
- team-shared contextual vector memory belongs in `AGENT_MEMORY`.
- governed organization doctrine belongs in `PROJECT_MEMORY` and `REFLECTION_MEMORY`, not in unreviewed team or chat memory.

## Promotion Rules

Every promotion candidate must carry:
- `classification`: what kind of memory the candidate represents
- `memory_layer`: one of `SOMA_MEMORY`, `AGENT_MEMORY`, `PROJECT_MEMORY`, or `REFLECTION_MEMORY`
- `confidence`: numeric confidence attached by the producing agent or review lane
- `review_required`: explicit boolean review posture
- `continuity_key`: stable key linking related work over time
- `tags`: operator-readable classification labels
- `evidence_refs`: source pointers when available

Confidence handling:
- `< 0.65`: keep in Managed Exchange as a low-confidence candidate; do not promote.
- `0.65 - 0.84`: candidate may be reviewed or refined; promotion requires review if sensitive, cross-user, policy-shaping, or organization-wide.
- `>= 0.85`: candidate may be eligible for promotion only if classification, scope, sensitivity, trust, and review posture are explicit.

Review handling:
- `review_required=true` for `REFLECTION_MEMORY` that describes user trajectory, contradictions, identity/behavior changes, cross-team patterns, or shared policy implications.
- `review_required=true` for promotion into `company_knowledge` or `soma_operating_context`.
- `review_required=false` is allowed only for low-risk scoped candidates that remain inside their original memory layer and scope.

## Reflection First, Not Recall First

The first goal of reflection is not "remember more."

The first goal is to detect meaning:
- what changed
- what repeated
- what contradicted earlier assumptions
- what became more important to the user
- what can improve future team execution

Reflection workflow:

```text
interaction / result / review
  -> reflection analysis
  -> classified LearningCandidate in Managed Exchange
  -> optional review or refinement
  -> promotion decision
  -> reflection_synthesis memory only when approved or policy-allowed
```

This keeps reflection anchored to understanding before recall.

## Managed Exchange Boundary

Memory candidates attach to Managed Exchange, not raw chat.

Current exchange anchor:
- channel: `organization.learning.candidates`
- schema: `LearningCandidate`
- capability: `learning`

Required `LearningCandidate` fields:
- `summary`
- `status`
- `classification`
- `memory_layer`
- `confidence`
- `review_required`
- `tags`
- `continuity_key`
- `created_at`

Recommended optional fields:
- `classification_reason`
- `promotion_target`
- `evidence_refs`
- `source_role`
- `source_team`
- `target_role`
- `target_team`
- `sensitivity_class`
- `trust_class`
- `allowed_consumers`

Rule:
- agents may publish candidates to exchange
- agents may not silently promote exchange candidates to durable memory
- promotion is a separate governed action with audit evidence

## Governance Guardrails

Do not loosen:
- capability constraints
- approval policy
- memory mutation rules
- trust and sensitivity labels
- scoped retrieval rules
- audit lineage

Candidate capture is allowed to be lower risk than promotion because it does not mutate memory.

Promotion remains higher risk because it changes what the system can reuse later.

## Trusted Retrieval And Arbitration

Scoped retrieval is necessary, but it is not sufficient.

The runtime should treat recalled memory as a candidate claim that still needs authority resolution.

Required precedence:
1. deterministic event and artifact truth
2. governed organization memory such as `company_knowledge`, `soma_operating_context`, and approved `reflection_synthesis`
3. active team-shared `AGENT_MEMORY`
4. Soma personal `SOMA_MEMORY`
5. temporary continuity and unreviewed candidates

Required conflict rules:
- lower-order memory must not silently override higher-order governed or deterministic truth
- team-shared memory that conflicts with newer organization doctrine should halt mutation and raise review or proposal
- Soma personal memory may shape style and relationship continuity inside scope, but it must not outrank governed policy or factual truth
- recalled memory should preserve evidence linkage and review posture wherever available

Design note:
- the detailed control-plane design for anchors, verification status, arbitration, supersession, and bounded growth lives in [V8 Trusted Memory Arbitration And Team Vector Contract](V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md)

## Team Delivery Architecture

The work should be delivered by small expert lanes, not one oversized team.

| Lane | Ownership | Delivery target |
| --- | --- | --- |
| Memory Architecture Lead | Layer taxonomy and promotion model | Keep `SOMA_MEMORY`, `AGENT_MEMORY`, `PROJECT_MEMORY`, and `REFLECTION_MEMORY` explicit in docs, prompts, API contracts, and tests, with `AGENT_MEMORY` clearly defined as the team-shared execution lane. |
| Exchange Integration Lead | Managed Exchange candidate flow | Ensure `LearningCandidate` carries classification, memory layer, confidence, review posture, evidence, and promotion target. |
| Governance Sentry | Guardrails and approvals | Keep candidate capture separate from mutation, preserve high-risk treatment for promotion, and prevent silent self-rewrite. |
| Runtime/Backend Lead | Tool and persistence contracts | Add promotion APIs/tools only after candidate review semantics are stable. Maintain scoped retrieval boundaries and retrieval-time arbitration. |
| Mother Brain Lead | Candidate synthesis and translation | Turn raw lessons, contradictions, and repeated execution patterns into classified `LearningCandidate` items with evidence and promotion intent. |
| UX/Operator Lead | Advanced review and memory visibility | Show learning candidates, review posture, and promotion targets in advanced surfaces without polluting the default Soma chat. |
| QA Lead | Verification | Build unit, API, and browser tests proving no interaction-to-memory shortcut exists. |

Coordination rule:
- Soma orchestrates these lanes through compact teams.
- Council helps review architecture and governance.
- NATS and Managed Exchange carry status, review, and output evidence.
- No single memory team should grow beyond the compact-team default unless the work is explicitly split into multiple coordinated lanes.

## Implementation Sequence

1. `COMPLETE`: introduce `reflection_synthesis` as a governed memory class with private/restricted defaults.
2. `ACTIVE`: tighten runtime prompts and exchange schema so reflection starts as a classified `LearningCandidate`.
3. `NEXT`: add typed backend helpers for publishing learning candidates so agents do not hand-build exchange payloads.
4. `NEXT`: introduce typed trusted-memory control-plane records and evidence anchors above the shared vector substrate.
5. `NEXT`: add a promotion command that consumes a reviewed exchange candidate and writes to the correct layer.
6. `NEXT`: add retrieval-time arbitration and contradiction-halt behavior across Soma, team, and governed memory.
7. `NEXT`: add advanced UI review for learning candidates, confidence, evidence, anchor state, and promotion target.
8. `NEXT`: add browser proof for a reflection candidate that is captured, reviewed, promoted, and then recalled from `reflection_synthesis`.

## Test Plan

Required tests:
- unit test that `LearningCandidate` requires `classification`, `memory_layer`, `confidence`, and `review_required`
- runtime context test that reflection instructions point to Managed Exchange first, not direct memory write
- governance test that promotion into `reflection_synthesis`, `company_knowledge`, or `soma_operating_context` remains high risk
- API/tool test for publishing a learning candidate with evidence and promotion target
- promotion test that rejects unclassified or low-confidence candidates
- browser test that shows candidate capture and promotion review without hiding review posture

Negative tests:
- interaction text alone cannot create `REFLECTION_MEMORY`
- raw transcript cannot be stored as reflection without synthesis classification
- candidate capture cannot change scoped retrieval results until promoted
- ordinary user chat cannot update `soma_operating_context`
- lower-order recalled memory cannot silently override newer governed doctrine

## Open Risks

- The current `load_deployment_context` path can still be used directly by an operator for explicit context intake; agent-driven reflection should use Managed Exchange candidates first.
- The current candidate schema is now explicit, but a typed helper will reduce malformed `publish_exchange_item` payloads.
- The advanced UI can inspect exchange activity, but it still needs a dedicated candidate-review experience before reflection promotion feels complete.
- The current runtime still scopes recall mostly through metadata; the trusted control plane and arbitration layer remain the next implementation slice.

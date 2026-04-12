# Memory
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Persistent semantic memory — what the system knows, what it has done, and what it can recall.

---

## Overview

The **Memory** page (`/memory`) is available in Advanced mode and provides a unified view of the system's three-tier memory architecture:

```
HOT   → Live signal stream (real-time events, last N minutes)
WARM  → Structured logs, SitReps, artifacts (recent history)
COLD  → Semantic vector store (long-term, searchable by meaning)
```

All three tiers are populated automatically as agents work. You can also intentionally load governed private records, customer/deployment knowledge, approved company guidance, admin-shaped Soma operating context, and reflection/synthesis observations through **Resources → Deployment Context** so Soma has durable goal-relevant context to reuse later without mixing it into ordinary remembered facts.

---

## Memory Classes

Mycelis now treats memory as several different classes with different purposes:

- **`SOMA_MEMORY`**: Soma-owned continuity and reviewed orchestrator facts. Admin-shaped Soma behavior belongs in the governed `soma_operating_context` sublane, not casual chat memory.
- **`AGENT_MEMORY`**: team/agent-scoped specialist continuity, decisions, and role-specific execution lessons.
- **`PROJECT_MEMORY`**: governed source context for work, including `user_private_context`, `customer_context`, and `company_knowledge`.
- **`REFLECTION_MEMORY`**: distilled lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations. Reflection starts as a Managed Exchange `LearningCandidate` before promotion into `reflection_synthesis`.
- **Durable semantic memory**: reusable facts, decisions, SitReps, recipes, and intentionally promoted summaries. This is the pgvector-backed recall substrate.
- **User-private context store**: user-uploaded or pasted records, diary notes, finance/legal/health references, and other sensitive material intentionally made available for specific target goal sets. This is private/restricted by default and is not company knowledge.
- **Customer context store**: operator- or customer-provided docs, notes, briefs, and research intentionally loaded into pgvector so Soma, Council, and teams can reason with deployment-specific requirements across future sessions.
- **Company knowledge store**: approved company-authored guidance or playbooks that Soma or teams are explicitly allowed to treat as durable organizational reference.
- **Reflection / synthesis memory**: distilled lessons, inferred patterns, contradictions, user-trajectory shifts, and meta-observations about what is changing over time. This is stored as `reflection_synthesis`, private/restricted by default, and should not be treated as raw transcript or customer content.
- **Temporary continuity**: restart-safe planning checkpoints and in-flight working context. This stays in temporary memory channels and does **not** automatically become long-term semantic memory.
- **Trace and audit**: conversation turns, mission events, and operational review logs used for causality, inspection, and governance. These are review surfaces, not default semantic memory.

Rule of thumb:

- if it should be reusable later by meaning as a learned fact, promote it into durable memory
- if it is private user-owned reference material for a specific personal/business goal, load it into the user-private context store and name the target goal sets
- if it is customer-provided or deployment-shaping reference material, load it into the customer context store
- if it is approved company-authored guidance, load it into the company knowledge store
- if customer context needs to become durable company reference, promote it through a governed approval path instead of silently reclassifying the original entry
- if it is a durable lesson, inferred pattern, contradiction, trajectory shift, or meta-observation that Soma should remember about how work is changing, first publish a classified Managed Exchange `LearningCandidate` with confidence and review posture, then promote it into `reflection_synthesis` only through the governed path
- if it is only useful for the current planning cycle, keep it in temporary continuity
- if it exists to explain what happened, treat it as trace or audit

---

## Semantic Search

The primary interface on the Memory page is the **semantic search bar**.

Type a natural-language query — not exact keywords, but the *meaning* of what you're looking for:

```
"Python file parsing functions we wrote last week"
"decisions made about the auth module"
"errors encountered in the CSV processor mission"
```

Results are ranked by **cosine similarity** to your query, not by keyword match. Relevant memories from different time periods surface together.

Each result card shows:
- **Content** — the stored text or artifact summary
- **Source** — which agent stored it and in which run
- **Score** — similarity confidence (0.0–1.0)
- **Timestamp** — when it was stored

Semantic search can also be scoped for teams and planning lanes through the API when a narrower recall boundary is required.

Governed deployment knowledge is stored under dedicated vector types:

- `customer_context` for operator- or customer-provided source material
- `company_knowledge` for approved company-authored guidance
- `soma_operating_context` for admin-owned guidance that shapes shared Soma posture and output specificity
- `user_private_context` for user-owned private records, diary entries, finance notes, and other sensitive references tied to explicit goal sets
- `reflection_synthesis` for lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations that Soma should retain as synthesis rather than transcript

That lets Soma, Council, and teams recall deployment knowledge independently from ordinary remembered facts when a stricter context boundary is needed.

---

## Storing a Memory

Agents store memories automatically during runs (via the `remember` tool). For larger operator-provided docs and reference material, use **Resources → Deployment Context** instead of treating them as small facts.

Use the Memory page when you want to query what is already retained, or when a smaller deliberate fact/preference/promotion is the right tool.

For general stored facts:

1. Click **+ Store**
2. Choose type: `fact`, `preference`, `goal`, or `observation`
3. Enter content
4. Click **Save**

Stored facts are immediately available to agents for RAG recall within their allowed memory scope.

General exploratory planning and routine conversation checkpoints are no longer promoted into semantic memory automatically. They stay in temporary continuity unless an agent deliberately promotes them.

Reflection is stricter than ordinary recall:

- raw interaction text should not go directly into `REFLECTION_MEMORY`
- reflection candidates must carry classification, confidence, and review posture in Managed Exchange first
- promotion into `reflection_synthesis` should preserve evidence and trust/sensitivity metadata

---

## SitReps (Situation Reports)

The **SitReps** tab shows compressed summaries of past mission activity. The system's Archivist process compresses raw log events into SitReps every 5 minutes.

Each SitRep covers:
- Active missions and their state transitions
- Key tool calls and artifacts produced
- Notable errors or governance events

SitReps are the "warm" tier — indexed for fast retrieval and embedded for semantic search.

---

## Artifacts

The **Artifacts** tab lists everything agents have created and stored:

| Column | Description |
|--------|-------------|
| **Title** | Artifact name or filename |
| **Type** | `code`, `document`, `data`, `image`, `report` |
| **Source** | Agent and run that created it |
| **Size** | File size |
| **Created** | Timestamp |

Click any artifact to preview it inline (markdown rendered, code with syntax highlighting, JSON formatted).

---

## Agent State

The **Agent State** tab shows the last known state of each active agent:
- Current status (`thinking`, `idle`, `tool_calling`, `offline`)
- Last message processed
- Tool call in progress (if any)
- Trust score for most recent output

---

## How Memory Flows

```
Agent calls remember() or store_artifact()
or operator loads Deployment Context from Resources
    ↓
Stored in PostgreSQL (log_entries + artifacts tables)
    ↓
Embedded via nomic-embed-text → context_vectors (pgvector)
    ↓
Available for semantic search plus governed customer/company/private/reflection context recall
    ↓
Archivist compresses raw logs into SitReps every 5 minutes
```

Every `tool.invoked`, `memory.stored`, and `artifact.created` event in a run's timeline corresponds to an entry in this store.

Important boundary:

- ordinary chat continuity and draft planning do **not** automatically become durable semantic memory
- they remain available through temporary continuity and trace surfaces until deliberately promoted
- governed deployment knowledge does **not** become ordinary Soma memory; it stays in the separate customer/company/private/reflection context store unless deliberately reclassified through an approved workflow

---

## Tips

- **Query broadly**: "authentication decisions" finds more than "auth_module_decision_2026-02-15"
- **Check artifacts before re-creating**: Agents and users often store work products here — search before asking Soma to write something from scratch
- **SitReps as quick history**: The SitReps tab is faster than reading full run timelines when you just need a summary of what happened in a session
- **Memory persists across sessions**: Everything stored here survives server restarts and is available in future sessions
- **Use Deployment Context for larger briefs**: architecture docs, MCP constraints, web research summaries, customer requirements, and approved company rollout policies belong in the dedicated governed-context intake lane so provenance and security posture stay explicit
- **Use reflection/synthesis for lessons, not transcripts**: capture the distilled change, contradiction, or pattern as a Managed Exchange learning candidate first, and keep the raw conversation in trace/continuity instead

Architecture reference:
- [V8 Memory Layer And Reflection Delivery Contract](../architecture-library/V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md)

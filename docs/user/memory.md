# Memory
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Workspace continuity — what the system may recall, where it came from, and whether it can be trusted.

---

## Overview

The **Memory** page (`/memory`) is available from Admin tools and provides a unified view of the system's three storage temperatures:

```
HOT   → Live signal stream (real-time events, last N minutes)
WARM  → Structured logs, SitReps, artifacts (recent history)
COLD  → Semantic vector store (long-term, searchable by meaning)
```

All three tiers are populated automatically as agents work. They are separate from the memory-layer model below, and the HOT/WARM/COLD temperature view is operational storage posture rather than the governing authority model.

The Memory page presents this as focused tabs so records do not compete for horizontal space:

- **Recent Work** opens the warm memory lanes. Use **Warm** for a combined recent view, **SitReps** for structured summaries, and **Artifacts** for retained records.
- **Search Memory** opens cold semantic recall. Use it when you want to search prior meaning rather than browse recent work.
- **Details** opens the selected search result or artifact for inspection, download, and provenance review.

The three views form one horizontal tab strip on compact screens and keep the active view in the URL. Browser Back returns to the previous Memory view, and a refresh preserves whether you were browsing recent work, searching, or inspecting details. Use Left/Right or Up/Down arrow keys while a view tab is focused; Home and End move to the first and last view.

You can also intentionally load governed private records, customer/deployment knowledge, approved company guidance, admin-shaped Soma operating context, and reflection/synthesis observations through **Resources → Deployment Context** so Soma has durable goal-relevant context to reuse later without mixing it into ordinary remembered facts.

---

## Memory Classes

Mycelis now treats memory as several different classes with different purposes:

- **`SOMA_MEMORY`**: Soma-owned continuity and reviewed orchestrator facts. Admin-shaped Soma behavior belongs in the governed `soma_operating_context` sublane, not casual chat memory.
- **`AGENT_MEMORY`**: team-shared and specialist-shared continuity, decisions, and execution lessons. This is the canonical lane for team-shared vector memory.
- **`PROJECT_MEMORY`**: governed source context for work, including `user_private_context`, `customer_context`, and `company_knowledge`.
- **`REFLECTION_MEMORY`**: distilled lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations. Reflection starts as a Managed Exchange `LearningCandidate` before promotion into `reflection_synthesis`.
- **Durable semantic memory**: reusable facts, decisions, SitReps, recipes, and intentionally promoted summaries. This is the pgvector-backed recall substrate.
- **User-private context store**: user-uploaded or pasted records, work log notes, finance/legal/health & safety references, and other sensitive material intentionally made available for specific target goal sets. This is private/restricted by default and is not company knowledge.
- **Customer context store**: operator- or customer-provided docs, notes, briefs, and research intentionally loaded into pgvector so Soma and governed teams can reason with deployment-specific requirements across future sessions.
- **Company knowledge store**: approved company-authored guidance or playbooks that Soma or teams are explicitly allowed to treat as durable organizational reference.
- **Reflection / synthesis memory**: distilled lessons, inferred patterns, contradictions, user-trajectory shifts, and meta-observations about what is changing over time. This is stored as `reflection_synthesis`, private/restricted by default, and should not be treated as raw transcript or customer content.
- **Temporary continuity**: restart-safe planning checkpoints and in-flight working context. This stays in temporary memory channels and does **not** automatically become long-term semantic memory.
- **Trace and audit**: conversation turns, execution events, and operational review logs used for causality, inspection, and governance. These are review surfaces, not default semantic memory.

Important distinction:
- `SOMA_MEMORY` is Soma's personal durable continuity
- `AGENT_MEMORY` is shared team execution memory
- governed organization doctrine belongs in `company_knowledge`, `soma_operating_context`, and approved `reflection_synthesis`, not in ordinary chat memory

Rule of thumb:

- if it should be reusable later by meaning as a learned fact, promote it into durable memory
- if it is private user-owned reference material for a specific personal/business goal, load it into the user-private context store and name the target goal sets
- if it is customer-provided or deployment-shaping reference material, load it into the customer context store
- if it is approved company-authored guidance, load it into the company knowledge store
- if customer context needs to become durable company reference, promote it through a governed approval path instead of silently reclassifying the original entry
- if it is a durable lesson, inferred pattern, contradiction, trajectory shift, or meta-observation that Soma should remember about how work is changing, first publish a classified Managed Exchange `LearningCandidate` with confidence and review posture, then promote it into `reflection_synthesis` only through the governed path
- if it is only useful for the current planning cycle, keep it in temporary continuity
- if it is a file handed to one team for one deliverable, keep it as source/support material on that Outcome rather than promoting it into durable memory
- if it exists to explain what happened, treat it as trace or audit

---

## Semantic Search

The primary recall interface on the Memory page is the **Search Memory** tab.

Type a natural-language query — not exact keywords, but the *meaning* of what you're looking for:

```
"Python file parsing functions we wrote last week"
"decisions made about the auth module"
"errors encountered while producing the CSV processor outcome"
```

With an embedding engine, results are ranked by **cosine similarity** to your query (semantic recall), and entries still waiting for an embedding are added by keyword rank (hybrid). Without one, recall still works: PostgreSQL full-text search ranks entries by the meaningful words in your query (common words such as "our" or "for" are ignored), so a question like "write a promo for our weekend special" finds a saved note about the weekend special.

Each result card shows:
- **Content** — the stored text or artifact summary
- **Source** — which agent stored it and in which run
- **Score** — similarity confidence (0.0–1.0)
- **Timestamp** — when it was stored

Semantic search can also be scoped for teams and planning lanes through the API when a narrower recall boundary is required.

If the embedding-capable AI engine is unavailable, the API returns keyword-ranked results with `retrieval_mode: keyword` and a `degraded.code` of `embedding_unavailable`, so the UI can say that semantic recall needs an embedding-capable engine. Mycelis checks the engine once and remembers a failure for five minutes, so a missing engine does not slow every chat turn.

Governed deployment knowledge is stored under dedicated vector types:

- `customer_context` for operator- or customer-provided source material
- `company_knowledge` for approved company-authored guidance
- `soma_operating_context` for admin-owned guidance that shapes shared Soma posture and output specificity
- `user_private_context` for user-owned private records, work log entries, finance notes, and other sensitive references tied to explicit goal sets (older `diary_entry`/`diary` values are stored as `worklog_entry`/`worklog`)
- `reflection_synthesis` for lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations that Soma should retain as synthesis rather than transcript

That lets Soma and governed teams recall deployment knowledge independently from ordinary remembered facts when a stricter context boundary is needed. When Soma's reply draws on this knowledge, the chat shows a **Sources** line naming each source as **Used** (the reply repeats its wording) or **Consulted**.

## Trusted Recall

Semantic search finds relevant memories by meaning. It does not decide final authority by itself.

When multiple memories disagree, the intended precedence is:
1. deterministic logs, turns, artifacts, and explicit run evidence
2. governed organization memory such as `company_knowledge`, `soma_operating_context`, and approved `reflection_synthesis`
3. team-shared `AGENT_MEMORY`
4. Soma personal `SOMA_MEMORY`
5. temporary continuity and unreviewed candidates

That means:
- team-shared memory can guide team execution
- Soma can read team memory without turning it into global doctrine
- personal continuity can shape style and relationship memory
- lower-order recalled memory should not silently override newer governed policy or verified evidence

---

## Storing Memory Or Governed Context

Agents store memories automatically during runs through governed runtime tools. For larger operator-provided docs and reference material, use **Resources -> Deployment Context** instead of treating them as small facts.

Use the Memory page when you want to query what is already retained. Use Deployment Context when you want new source material to influence future Soma reasoning with explicit provenance, sensitivity, and trust posture.

The user-facing distinction:

- Soma docs lookup is read-only and citable help/architecture reference
- Team/source files are handoff material for the current Outcome unless promoted
- Deployment Context is durable governed source material for future reasoning
- Memory is where already-retained facts, SitReps, artifacts, and continuity are inspected

Who sees a saved Deployment Context entry in the Memory list, entry actions, and memory search: everyone sees org-wide entries (company knowledge, Soma operating context, and anything saved with global visibility). There, private entries are visible only to the person who saved them, admins included. Team entries are visible to the person who saved them and to members of that team. Memory search follows the same rule. Only the saver can change their private or team entries; org-wide entries need an admin with `memory:write`. An entry you cannot see answers "This saved item no longer exists."

> **Known limitation (2026-09-28):** Soma chat does not yet apply this rule. When another user chats with Soma, Soma's recall can include your private or team entries in its answer and in its **Sources** line, and Soma's memory tools can return their text. Until this is fixed, do not save material in Deployment Context that other users of this deployment must not see.

Manual memory creation may appear through admin/runtime tooling where enabled, but it is not the default path for larger source documents. Stored facts or context are available to agents only within their allowed memory scope and trust boundary.

General exploratory planning and routine conversation checkpoints are no longer promoted into semantic memory automatically. They stay in temporary continuity unless an agent deliberately promotes them.

Reflection is stricter than ordinary recall:

- raw interaction text should not go directly into `REFLECTION_MEMORY`
- reflection candidates must carry classification, confidence, and review posture in Managed Exchange first
- promotion into `reflection_synthesis` should preserve evidence and trust/sensitivity metadata

---

## SitReps (Situation Reports)

The **SitReps** tab shows compressed summaries of past execution activity. The system's Archivist process compresses raw log events into SitReps every 5 minutes.

Each SitRep covers:
- Active work and its state transitions
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
- **Keep important plans out of chat-only state**: if a workflow should survive a full reboot, ask Soma to retain a plan summary, checklist, or output contract as an artifact and keep working state in temporary continuity
- **Use Deployment Context for larger briefs**: architecture docs, MCP constraints, web research summaries, customer requirements, and approved company rollout policies belong in the dedicated governed-context intake lane so provenance and security posture stay explicit
- **Use reflection/synthesis for lessons, not transcripts**: capture the distilled change, contradiction, or pattern as a Managed Exchange learning candidate first, and keep the raw conversation in trace/continuity instead

Related references:
- [Workflow Variants And Plan Memory](workflow-variants-and-plan-memory.md)
- [Governance & Trust](governance-trust.md)
- [Mycelis Canonical PRD](../architecture-library/MYCELIS_CANONICAL_PRD.md)

# Memory

> Persistent semantic memory — what the system knows, what it has done, and what it can recall.

---

## Overview

The **Memory** page (`/memory`) provides a unified view of the system's three-tier memory architecture:

```
HOT   → Live signal stream (real-time events, last N minutes)
WARM  → Structured logs, SitReps, artifacts (recent history)
COLD  → Semantic vector store (long-term, searchable by meaning)
```

All three tiers are populated automatically as agents work. You don't need to manage them manually — but you can query, review, and store information directly.

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

---

## Storing a Memory

Agents store memories automatically during runs (via the `remember` tool). You can also store directly from the Memory page:

1. Click **+ Store**
2. Choose type: `fact`, `preference`, `goal`, or `observation`
3. Enter content
4. Click **Save**

Stored facts are immediately available to agents for RAG recall.

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
    ↓
Stored in PostgreSQL (log_entries + artifacts tables)
    ↓
Embedded via nomic-embed-text → context_vectors (pgvector)
    ↓
Available for semantic search by any agent or user
    ↓
Archivist compresses raw logs into SitReps every 5 minutes
```

Every `tool.invoked`, `memory.stored`, and `artifact.created` event in a run's timeline corresponds to an entry in this store.

---

## Tips

- **Query broadly**: "authentication decisions" finds more than "auth_module_decision_2026-02-15"
- **Check artifacts before re-creating**: Agents and users often store work products here — search before asking Soma to write something from scratch
- **SitReps as quick history**: The SitReps tab is faster than reading full run timelines when you just need a summary of what happened in a session
- **Memory persists across sessions**: Everything stored here survives server restarts and is available in future sessions

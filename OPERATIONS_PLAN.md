# Mycelis — Operations Panel Plan
> Generated: 2026-02-22
> Status: PLANNING — awaiting execution

---

## Executive Summary

Four interlocking problems to solve, starting from base layers:

1. **Agentry interaction broken** — Diagnose + fix the council chat chain
2. **Data boundary** — Surface clearly in Resources → Brains
3. **Memory page rethink** — Too technical; redesign for non-technical users
4. **"Mission Control" rename** — Less aggressive name; full implementation map

---

## 1. Why Agentry Interaction Is Broken

### The Chat Chain (what must all be alive)

```
PostgreSQL (:5432)
  └─ NATS (embedded in Go binary)
       └─ Go Core Server (:8081)
            ├─ Soma (loads admin.yaml + council.yaml manifests)
            │    ├─ admin agent → subscribes to swarm.council.admin.request
            │    └─ council-architect/coder/creative/sentry agents → individual topics
            └─ Cognitive Provider (Ollama :11434 local OR remote API key)
                 └─ Next.js (:3000) → proxy /api/* → :8081
                      └─ sendMissionChat() → POST /api/v1/council/{member}/chat
```

**Every node must be green. One failure breaks the entire chain.**

### What HandleCouncilChat requires

```go
// cognitive.go:412
func (s *AdminServer) HandleCouncilChat(w http.ResponseWriter, r *http.Request) {
    memberID := r.PathValue("member")       // "admin" (default)
    teamID, _, ok := s.isCouncilMember(memberID)  // checks admin-core → "admin" member
    // if !ok → 404 "Unknown council member: admin"
    // if s.NC == nil → 503 "Swarm offline"
    msg, err := s.NC.RequestWithContext(ctx, "swarm.council.admin.request", payload, 60s)
    // if no subscriber → timeout error
}
```

The error IS displayed in the UI (missionChatError). If users see it but don't know what
it means, that's a UX gap we need to fill.

### Verification steps (run these first, in order)

```bash
inv lifecycle.status                          # Dashboard: all services + PIDs
curl localhost:8081/healthz                   # Backend responding?
curl localhost:8081/api/v1/council/members    # Returns admin + council?
curl localhost:8081/api/v1/cognitive/status   # Brain provider online?
```

### Code regression to fix (V7 Step 01 leftover)

**MissionControl.tsx:98** — "NEW MISSION" button still routes to old `/wiring` route:
```tsx
// BEFORE (broken intent — /wiring is now server-redirect to /automations?tab=wiring)
onClick={() => router.push("/wiring")}

// AFTER (V7-correct — wiring is inside Automations, Advanced Mode only)
// Remove the button entirely; mission creation happens via chat with Soma
// OR: replace with "Start a Mission..." that focuses the chat input
```

**MissionControl.tsx:85** — Header title is aggressive and will be renamed (see §4).

**Missing: Offline state** — When `fetchCouncilMembers` returns 503, the panel shows
an empty dropdown with no recovery guidance. Need a startup guide overlay.

### New: Offline/Degraded state for the panel

When `GET /api/v1/council/members` fails, instead of silent empty state, show:

```
┌─────────────────────────────────────┐
│  Soma is Offline                    │
│  The neural organism needs to be    │
│  started before you can direct it.  │
│                                     │
│  Run: inv lifecycle.up              │
│  Then refresh this page.            │
│                                     │
│  [Check Status]  [Learn More]       │
└─────────────────────────────────────┘
```

---

## 2. Data Boundary — What It Is and Where It Belongs

### Definition

`data_boundary: 'local_only' | 'leaves_org'` (on `BrainProvenance`)

| Value | Meaning |
|-------|---------|
| `local_only` | Inference on YOUR hardware (Ollama, local vLLM). Nothing leaves. |
| `leaves_org` | Inference via remote API (OpenAI, Anthropic, etc.). Your prompts + context go to third-party servers. |

### Current surfacing (chat messages only)

In `MissionControlChat.tsx`, it appears as:
- Tooltip on the brain badge: `Data: ${msg.brain.data_boundary}`
- Amber "External" badge when `location === 'remote'`

This is **reactive** — users only see it AFTER a message has been processed.

### What's missing in Resources → Brains

The `BrainsPage` component (`components/settings/BrainsPage.tsx`) shows brain
providers and their enable/disable toggle, but does NOT show data boundary.

Users need to know BEFORE enabling a brain whether their data leaves the org.

### Plan: Data Boundary indicators in Brains tab

Each brain card should show:
```
┌─────────────────────────────────────────────┐
│ 🧠 Ollama                    [Enabled ✓]   │
│ llama3.2:3b                                 │
│ ■■■ Local inference                         │
│ 🛡️ On-device — data stays on your machine  │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ 🌐 OpenAI                    [Disabled]    │
│ gpt-4o                                      │
│ ■■■ Remote API                              │
│ ⚠️  External — data leaves your org         │
└─────────────────────────────────────────────┘
```

Implementation: Read `data_boundary` from `cognitive.yaml` provider config and
surface it in `BrainsPage` per-provider. No new API needed — data is already in
`HandleCognitiveConfig` response.

---

## 3. Memory Page Rethink

### What users actually want from Memory

> "What does Mycelis remember? What have we worked on? Find something we discussed."

### What they're getting instead

| Panel | What it actually is | Who it's for |
|-------|---------------------|--------------|
| Hot | Raw SSE/NATS stream (real-time events) | Developers |
| Warm | Recent conversation history | Users ✓ |
| Cold | pgvector semantic search | Users ✓ (if discoverable) |

**The "Hot" panel is developer telemetry surfaced to end users. Wrong audience.**

"THREE-TIER MEMORY" is the internal architecture name, not a user concept.

### Proposed redesign

**Layout: Search-first, single column, progressive disclosure**

```
Memory
"Everything Mycelis has learned and remembered"

╔═══════════════════════════════════════════════╗
║  🔍  What do you want to find?               ║
║  [Ask: "What did we work on last week?"]  [→] ║
╚═══════════════════════════════════════════════╝

─── Recent Conversations ────────────────────────
[conversation card — topic, date, key outcomes]
[conversation card — topic, date, key outcomes]
[conversation card — topic, date, key outcomes]
[Load more]

─── Stored Knowledge ────────────────────────────
[fact card — what Soma remembered]
[fact card — what Soma remembered]

─── [Advanced Mode Only] Signal Monitor ─────────
[collapsible — the current Hot panel content]
[collapsed by default]
```

### What changes

- Remove "THREE-TIER MEMORY" title
- Remove the 3-column layout
- Move semantic search to top (primary action)
- "Warm" content becomes "Recent Conversations"
- "Cold" content appears in the search results
- "Hot" content moves to Advanced Mode, collapsed by default
- Page title stays "Memory" (already correct)

### Subcomponents to create/refactor

| Component | Action |
|-----------|--------|
| `MemoryExplorer.tsx` | Refactor layout from 3-col to single-column |
| `MemorySearch.tsx` | New: prominent search box at top |
| `RecentConversations.tsx` | Rename from WarmMemoryPanel, simplified UI |
| `MemoryFacts.tsx` | Rename from ColdMemoryPanel, result display |
| `SignalMonitor.tsx` | Rename from HotMemoryPanel, collapsible, Advanced Mode only |

---

## 4. Rename "Mission Control"

### Why it needs to change

"Mission Control" evokes:
- NASA launch command (high-stakes, momentary, military)
- Call centers ("Mission Control is standing by")
- Video games ("initiate mission")

What the panel IS:
- Persistent co-presence with Soma (your organism's executive mind)
- Ongoing work with long-lived and short-lived teams
- The natural entry point — where you direct, review, and exist with your agentry

### Name candidates

| Name | Vibe | Pros | Cons |
|------|------|------|------|
| **Operations** | Calm, professional | Clear, non-military, implies ongoing work | Still operational/military adjacent |
| **Studio** | Creative, collaborative | Warm, implies craftsmanship | Implies creative work only |
| **Workspace** | Familiar, neutral | Instantly understood | Generic, no Mycelis character |
| **Nexus** | Network hub | Connection-centric, Mycelis-aligned | Feels technical |
| **Hearth** | Warm, persistent | Evokes home base, permanence | Too cozy? |
| **Ground** | Foundational | "Home ground," stable, rooted | Too abstract |

### Recommendation: **"Operations"**

Rationale:
- Drops the word "Mission" entirely (removes the triggering word)
- "Operations" = ongoing, continuous, not one-time missions
- "Ops" shortens naturally for the nav label
- Implies "where things happen" without aggression
- Professional enough for enterprise use

**Nav label**: `Ops` (was `Mission Control` — fits in narrow rail)
**Page title**: `Operations` (was `MISSION CONTROL`)
**Sub-title**: `Direct Soma and your active teams`

Alternative if "Operations" still sounds too military: **"Workspace"** (safe fallback).

### Files to update (route stays `/dashboard`)

| File | Change |
|------|--------|
| `components/dashboard/MissionControl.tsx` | Header text, component name (refactor) |
| `components/dashboard/MissionControlChat.tsx` | Component name, placeholder text |
| `app/(app)/dashboard/page.tsx` | Loading text "Loading Mission Control..." |
| `store/useCortexStore.ts` | `CHAT_STORAGE_KEY` (needs migration), state comments |
| `components/shell/ZoneA_Rail.tsx` | Nav label (Mission Control → Ops) |
| `docs/product/ia-v7-step-01.md` | Documentation |
| `V7_DEV_STATE.md` | Reference update |
| `MEMORY.md` | Auto-memory reference |

**localStorage migration**: `mission-control-split` → `mycelis-ops-split`
**localStorage migration**: `mycelis-mission-chat` → `mycelis-ops-chat`
(preserve old keys as fallback on first load)

---

## 5. Full Map: What Operations Panel Needs to Work

### Layer 0: Infrastructure

```
Component           Port    Status check
────────────────────────────────────────
PostgreSQL          5432    inv lifecycle.status
NATS (Go-embedded)  4222    inv lifecycle.status
Go Core Server      8081    curl localhost:8081/healthz
Cognitive Provider  11434   curl localhost:8081/api/v1/cognitive/status
Next.js Dev         3000    curl localhost:3000/
```

### Layer 1: Soma Instantiation (Go boot sequence)

```
main.go:
  1. db, _ = sql.Open("postgres", dsn)          → PostgreSQL
  2. nc, _ = nats.Connect(natsURL)              → NATS
  3. brain = cognitive.NewRouter(cognitiveYAML)  → Provider adapters
  4. registry = swarm.NewRegistry(teamsDir)      → Reads YAML manifests
  5. soma = swarm.NewSoma(nc, guard, registry, brain, stream, mcpExec, internalTools)
  6. soma.Start():
       a. LoadManifests() → admin.yaml, council.yaml, genesis.yaml, telemetry.yaml
       b. For each manifest → NewTeam() → team.Start():
            For each member → NewAgent() → agent.Start():
              nc.Subscribe("swarm.council.{memberID}.request", handleDirectRequest)
       c. nc.Subscribe("swarm.global.input.>", handleGlobalInput)
       d. axon.Start()
```

**Failure modes:**
- NATS offline → soma.Start() returns error → no agents subscribe → all chat timeouts
- Bad YAML → manifests fail to load → isCouncilMember() returns false → 404
- Ollama offline → brain.Infer() fails → agent returns error → 502

### Layer 2: Council Member Validation

```
GET /api/v1/council/members
  → isCouncilMember() searches teams "admin-core" and "council-core"
  → admin-core.members = [{ id: "admin", role: "admin" }]  (admin.yaml)
  → council-core.members = [architect, coder, creative, sentry]  (council.yaml)
  → returns 5 members total

Frontend: fetchCouncilMembers() on mount
  → if OK: councilMembers = [{id: "admin", role: "admin"}, ...]
  → if fail: councilMembers = []  ← need offline state here
  → UI dropdown shows members (or "Soma — Executive Cortex" fallback)
```

### Layer 3: Chat Interaction

```
User types → handleSubmit() → sendMissionChat(message)
  → missionChat.push({ role: 'user', content: message })
  → isMissionChatting = true
  → fetch POST /api/v1/council/admin/chat
       body: { messages: last 20 chat messages }
     ↓
  HandleCouncilChat()
    → isCouncilMember("admin") → admin-core → "admin" ✓
    → nc.RequestWithContext("swarm.council.admin.request", payload, 60s)
      ↓
    Agent.handleDirectRequest()  [admin agent listening on NATS topic]
      → unmarshal messages array → get last user message
      → build system prompt from admin.yaml
      → brain.Infer(systemPrompt, messages) → Ollama/remote API
      → if tool_call in response → execute tool → re-infer
      → return ProcessResult{Text, ToolsUsed, Artifacts, ProviderID, ModelUsed}
      ↓ (back via NATS reply)
    HandleCouncilChat() builds CTSEnvelope:
      → BrainProvenance (provider, model, location, data_boundary)
      → Mode: answer | proposal (if mutation tools used)
      → TrustScore: 0.7 (TrustScoreCognitive)
      → respondAPIJSON(200, APIResponse{ok: true, data: envelope})
     ↓
  Frontend parses CTSChatEnvelope
    → ChatMessage with brain, mode, provenance, proposal (if any)
    → missionChat.push(chatMsg)
    → persistChat() → localStorage
    → if mode === "proposal": ProposedActionBlock shown
```

### Layer 4: Team Manifestation (spawning new teams)

```
Soma generates blueprint → mode = "proposal"
  ↓
ProposedActionBlock.tsx → user clicks "Confirm"
  → confirmProposal() → fetch POST /api/v1/intent/confirm-action { token }
    → HandleConfirmAction()
      → validateToken(token)
      → loadBlueprint(intentProof)
      → commitAndActivate(blueprint):
          for each team in blueprint:
            soma.SpawnTeam(TeamManifest{
              ID, Name, Type, Members, Inputs, Deliveries
            })
            → NewTeam() → team.Start()
            → for each member: NewAgent() → nc.Subscribe(...)
      → return activation result
     ↓
  Frontend: activeMissionId updated
  Teams tab in Automations shows new teams
  Wiring tab shows blueprint graph
```

### Layer 5: Broadcast (all teams at once)

```
User clicks broadcast mode → broadcastToSwarm(message)
  → fetch POST /api/v1/swarm/broadcast { content, source }
    → soma.HandleBroadcast()
      → snapshot all active team IDs
      → fan-out: nc.Request("swarm.team.{teamID}.trigger", payload, 60s) for each team
      → collect replies
      → return { status: "broadcast", teams_hit: N, replies: [...] }
  → lastBroadcastResult = { teams_hit: N }
```

---

## 6. Execution Order

### Step 0: Diagnose agentry failure FIRST (before any code changes)

```bash
inv lifecycle.status
inv lifecycle.health
```

If services are down: `inv lifecycle.up --build --frontend`

If services are up but chat fails: check browser devtools Network tab for
`/api/v1/council/members` and `/api/v1/council/admin/chat` errors.

### Step 1: Fix code regressions (small, isolated)

1. `MissionControl.tsx:98` — Remove or update "NEW MISSION" button
2. `MissionControl.tsx` — Add offline/degraded state when council members = 0
3. `BrainsPage.tsx` — Add data boundary indicators

### Step 2: Rename (coordinated rename across files)

1. Rename nav label in ZoneA_Rail.tsx
2. Update MissionControl.tsx header and loading text
3. Update dashboard/page.tsx
4. Migrate localStorage keys (with fallback)
5. Update docs and state files

### Step 3: Memory page redesign

1. Refactor MemoryExplorer.tsx layout (3-col → single column)
2. Elevate semantic search to top
3. Move HotMemoryPanel (Signal Monitor) behind Advanced Mode gate
4. Simplify component naming

### Step 4: V7 Team A (Event Spine) — proceeds as planned after above is stable

Migrations 023-024, events/store.go, runs/manager.go, runs.go handlers.

---

## 7. Design Principles for Operations Panel (Post-Rename)

### UX principles

1. **Soma is the gateway** — Users talk to Soma. Soma decides to consult council or delegate to teams. Users shouldn't need to know who does what internally.

2. **Progressive revelation** — Council member selector is visible but secondary. Most users should just talk to Soma. Advanced users can target specific council members.

3. **Persistent presence** — Chat history survives page reloads. Teams that are active show their status. The panel is a living co-working space, not a transaction screen.

4. **Proposal → Confirm flow** — Any state mutation shows a proposal card. Users can review before confirming. Trust score and data boundary are always visible.

5. **Offline gracefully** — When backend is down, the panel says so clearly with recovery instructions.

### Layout principles (current layout — preserve)

```
Operations
├── Header (12px): title + SIGNAL status + New Intent button
├── Mode Ribbon: current brain/mode/governance state
├── Telemetry Row: live metrics
└── Resizable vertical split:
    ├── TOP (55%): Soma chat + council routing
    │   ├── Council target selector (dropdown)
    │   ├── Broadcast mode toggle
    │   └── Chat log with message bubbles
    └── BOTTOM (45%): Active Teams overview
         └── OpsOverview: missions + team status
```

---

## Decisions Needed From Architect

1. **Name**: Operations (recommended), Studio, Workspace, or Nexus?
2. **Memory page**: Full redesign (§3 plan) or iterative (just hide Hot, elevate search)?
3. **NEW MISSION button**: Remove (use chat) or redirect to `/automations?tab=wiring`?
4. **Offline state**: Full startup guide overlay or simple error message improvement?

# PRD: Phase 12 — Persistent Agent Memory + System Hardening

**Version:** 1.0
**Date:** 2026-02-18
**Status:** DRAFT — Awaiting Architect Approval
**Prerequisites:** Phase 17 (Legacy Migration) COMPLETE

> This PRD combines the next feature phase (Persistent Agent Memory) with a critical
> system hardening track. The hardening clears production-blocking debt discovered in
> the full-stack audit conducted on 2026-02-18.

---

## I. Executive Summary

Mycelis V7.0 has 56 backend API routes, 14 frontend pages, 20 migrations, and ~85 React
components delivering a recursive swarm operating system. Phase 17 completed the
Midnight Cortex theme migration, eliminating all legacy Vuexy artifacts.

A comprehensive audit of backend, frontend, and documentation surfaces **4 critical
blockers**, **6 high-priority debts**, and **5 medium-priority improvements** that
must be resolved before — or alongside — the Phase 12 feature work.

This PRD defines two parallel tracks:

| Track | Codename | Objective |
|:------|:---------|:----------|
| **Track A** | Hardening | Clear production blockers, debug code, test gaps, API hygiene |
| **Track B** | Agent Memory | Cross-mission persistence, semantic recall, memory consolidation |

Both tracks share the Phase 12 migration window (Migrations 021–023).

---

## II. Current State Assessment

### Backend (Go 1.26)

| Metric | Value | Assessment |
|:-------|:------|:-----------|
| HTTP routes | 56 | Healthy |
| SQL migrations | 20 (only 1 has down file) | Risk: no rollback |
| Internal packages | 18 | Healthy |
| Packages with zero tests | 8 of 18 (44%) | **Critical gap** |
| TODOs in code | 5 actionable | Medium |
| Dead code | None detected | Clean |
| Graceful degradation | All services handle NATS/DB down | Excellent |

**Untested packages (zero unit tests):**
1. `internal/provisioning` — blueprint-to-manifest conversion
2. `internal/router` — core NATS message dispatch + governance intercept
3. `internal/registry` — connector template CRUD
4. `internal/mcp` — 7 HTTP endpoints, client pool, executor
5. `internal/signal` — SSE broadcasting
6. `internal/transport/nats` — NATS client wrapper
7. `internal/agentry` — agent runner
8. `internal/identity` — type definitions

### Frontend (Next.js 16 + React 19 + Zustand 5)

| Metric | Value | Assessment |
|:-------|:------|:-----------|
| Routes | 14 app + 1 marketing | Complete |
| Components | ~85 TSX files | Healthy |
| Zustand store | 65 state fields, 60+ actions, 1886 LOC | **Needs audit** |
| Unit tests | 30 files | 50% component coverage |
| E2E specs | 20 Playwright files | Good |
| `any` type usage | 121 occurrences | Acceptable (mostly tests) |
| `@ts-ignore` | 0 | Clean |
| Accessibility score | 4/10 | **Needs work** |

### Documentation

| Metric | Status |
|:-------|:-------|
| README.md accuracy | 96% — Phase 17 double-listed fix needed |
| API_REFERENCE.md | 100% accurate |
| BACKEND.md | 98% — tool inventory mismatch with SWARM_OPERATIONS.md |
| FRONTEND.md | 92% — component count stale |
| TESTING.md | 85% — E2E count stale (12 → 20) |
| COGNITIVE_ARCHITECTURE.md | 60% consistency — Ollama endpoint contradicts 4 other docs |

---

## III. Track A — System Hardening

### A1. Critical Blockers (MUST before any release)

#### A1.1 Remove Debug Code from Production

**CircuitBoard DEBUG button** — [components/wiring/CircuitBoard.tsx:113-125](interface/components/wiring/CircuitBoard.tsx#L113-L125)

A hardcoded red `DEBUG STATE` button renders in the top-left of the ReactFlow canvas.
Logs full blueprint state to console on click. Must be deleted.

```
Files: interface/components/wiring/CircuitBoard.tsx (lines 113-125)
Action: DELETE the entire <div> block containing the debug button
```

**Zustand DOM injection** — [store/useCortexStore.ts:1820-1832](interface/store/useCortexStore.ts#L1820-L1832)

`deleteAgentFromMission()` creates a visible `<div id="debug-result-*">` with inline
styles (fixed position, z-index 9999) injected into the document body. Production
liability — will overlay UI unexpectedly.

```
Files: interface/store/useCortexStore.ts (lines 1820-1832)
Action: DELETE the DOM injection block. Replace with proper toast/notification.
```

**Console debug logging** — [store/useCortexStore.ts:1733-1757](interface/store/useCortexStore.ts#L1733-L1757)

`deleteAgentFromDraft()` has 5 `console.log("[DEBUG]")` statements.

```
Files: interface/store/useCortexStore.ts (lines 1733, 1740, 1747, 1751, 1756)
Action: DELETE all console.log("[DEBUG]") statements
```

#### A1.2 Resolve Duplicate Components

| Duplicate | Location A | Location B | Action |
|:----------|:-----------|:-----------|:-------|
| CircuitBoard | `components/workspace/CircuitBoard.tsx` | `components/wiring/CircuitBoard.tsx` | DELETE workspace/ version — wiring/ is canonical (has Phase 9 editor) |
| LogStream | `components/hud/LogStream.tsx` | `components/stream/LogStream.tsx` | Audit usage, DELETE non-canonical |
| TeamRoster | `components/dashboard/TeamRoster.tsx` | `components/missions/TeamRoster.tsx` | Audit usage, DELETE non-canonical |

#### A1.3 Implement Governance Stubs

`approveArtifact()` and `rejectArtifact()` in useCortexStore only call `console.log()`.
No backend API integration.

```
Files: interface/store/useCortexStore.ts
Action: Wire to PUT /api/v1/artifacts/{id}/status with status='approved'|'rejected'
```

#### A1.4 Add Error Boundaries

Zero `error.tsx` or `loading.tsx` files exist in the entire app router.

```
Action: Add to app/(app)/layout:
  - error.tsx — catch render failures with cortex-themed error card
  - loading.tsx — skeleton screen with cortex pulse animation
```

---

### A2. High Priority (Should complete with Phase 12)

#### A2.1 Centralized API Client

48 unique endpoints are called via scattered `fetch()` across 20+ files. No retry
logic, no interceptors, no global error handling. One endpoint (`/agents`) is missing
the `/api/v1` prefix.

**Deliverables:**
```
NEW FILE: interface/lib/api.ts
  - API constant map (all 48 endpoints)
  - apiRequest<T>() wrapper with:
    - Automatic JSON parsing
    - Global error handler
    - Retry logic (1 retry on 500/503)
    - Request/response logging (dev only)
  - Type-safe response envelope: APIResponse<T>
```

**Migration:** Refactor all `fetch()` calls in useCortexStore.ts and page components
to use `apiRequest()`. This is a sweeping change across ~55 call sites but is
mechanical (search-and-replace pattern).

#### A2.2 Backend Test Coverage for Critical Paths

Target: Cover the 4 most critical untested packages.

| Package | Priority | Why | Test File |
|:--------|:---------|:----|:----------|
| `internal/router` | P0 | Core message dispatch, governance intercept | `router/router_test.go` |
| `internal/mcp` | P0 | 7 HTTP endpoints, client pool | `mcp/service_test.go` |
| `internal/signal` | P1 | SSE broadcasting — real-time | `signal/stream_test.go` |
| `internal/provisioning` | P1 | Blueprint-to-manifest | `provisioning/engine_test.go` |

Target: Raise package-level coverage from 36% to 58%+.

#### A2.3 Fix Hardcoded Values in Cognitive Router

| Value | Location | Fix |
|:------|:---------|:----|
| Temperature 0.7 | `cognitive/router.go:363` | Load from profile config YAML |
| Provider fallback | `cognitive/router.go:312` | Match provider tier (local > cloud) |

#### A2.4 Robust Archivist JSON Extraction

`memory/archivist.go:145` — LLM output parsing assumes well-formed JSON. Can fail
on multi-object output or commentary wrapping.

```
Action: Implement regex-based JSON extraction with fallback:
  1. Try direct JSON.Unmarshal
  2. Try regex extraction of first { ... } block
  3. Fall back to raw string storage with error flag
```

---

### A3. Medium Priority (Nice to have)

#### A3.1 Component Splitting

| Component | Lines | Split Into |
|:----------|:------|:-----------|
| NatsWaterfall.tsx | 400+ | PriorityAlerts, SignalRow, FilterTabs, WaterfallLog |
| OperationsBoard.tsx | 310 | PrioritySection, WorkloadsSection, MissionsSection |
| ApprovalsPage | 608 | ApprovalsTab, PolicyTab, ProposalsTab (extract sub-components) |

#### A3.2 Unused Zustand State Cleanup

| Field | Issue | Action |
|:------|:------|:-------|
| `subscribedSensorGroups` | Toggled but never filters feeds | Wire to filter logic OR remove |
| `lastBroadcastResult` | Set but never displayed | Display in toast OR remove |
| `savedBlueprints` | Local only, never persisted | Add persistence API OR document as local-only |

#### A3.3 Documentation Fixes

| Document | Fix |
|:---------|:----|
| OVERVIEW.md | Remove Phase 17 from upcoming phases (it's delivered) |
| TESTING.md | Update E2E count from 12 to 20, list all spec files |
| COGNITIVE_ARCHITECTURE.md | Fix Ollama endpoint to `192.168.50.156:11434` (LAN) |
| SWARM_OPERATIONS.md | Reconcile tool inventory with BACKEND.md |
| FRONTEND.md | Update component count |

---

## IV. Track B — Persistent Agent Memory (Phase 12 Feature)

### B1. Problem Statement

Agents currently have no memory across missions. When an agent spawns, it starts with
a blank slate — losing all expertise, observations, and context from prior runs. The
`remember`/`recall` tools exist but write to a global admin memory store
(`agent_memories` table), not agent-scoped persistent storage.

The `agent_state` table (Migration 018) was designed for per-agent state but has a
critical flaw: `mission_id` FK with `ON DELETE CASCADE` — when a mission is deleted,
all agent state is destroyed, violating the cross-mission persistence requirement.

### B2. Requirements

#### B2.1 Cross-Mission Agent State (Migration 021)

**Problem:** `agent_state.mission_id` FK with CASCADE deletes agent memory when missions end.

**Solution:**
```sql
-- Migration 021: Make mission_id nullable for cross-mission persistence
ALTER TABLE agent_state ALTER COLUMN mission_id DROP NOT NULL;
ALTER TABLE agent_state DROP CONSTRAINT IF EXISTS agent_state_mission_id_fkey;
ALTER TABLE agent_state ADD CONSTRAINT agent_state_mission_id_fkey
    FOREIGN KEY (mission_id) REFERENCES missions(id) ON DELETE SET NULL;

-- Add origin tracking
ALTER TABLE agent_state ADD COLUMN origin_mission_id UUID;
ALTER TABLE agent_state ADD COLUMN memory_type VARCHAR(20) DEFAULT 'ephemeral';
-- memory_type: 'ephemeral' (dies with mission), 'persistent' (survives), 'protected' (governance-locked)

CREATE INDEX idx_agent_state_memory_type ON agent_state(memory_type);
```

#### B2.2 Semantic Agent Memory (Migration 022)

**Problem:** No vector search on agent state. Agents can't semantically recall past
observations.

**Solution:**
```sql
-- Migration 022: Add vector column for semantic agent memory
ALTER TABLE agent_state ADD COLUMN context_vector vector(768);

CREATE INDEX idx_agent_state_vector ON agent_state
    USING ivfflat (context_vector vector_cosine_ops) WITH (lists = 50);

-- Add agent namespace for scoped queries
ALTER TABLE agent_memories ADD COLUMN agent_id TEXT;
ALTER TABLE agent_memories ADD COLUMN mission_id UUID;
CREATE INDEX idx_agent_memories_agent ON agent_memories(agent_id);
```

#### B2.3 Agent Memory CRUD (Go Backend)

**New file:** `internal/memory/agent_memory.go`

```go
type AgentMemoryService struct {
    db *pgxpool.Pool
}

// Core operations
func (s *AgentMemoryService) SetState(ctx, agentID, key, value string, memoryType string) error
func (s *AgentMemoryService) GetState(ctx, agentID, key string) (string, error)
func (s *AgentMemoryService) ListState(ctx, agentID string) ([]AgentStateEntry, error)
func (s *AgentMemoryService) SemanticRecall(ctx, agentID, query string, limit int) ([]AgentStateEntry, error)
func (s *AgentMemoryService) DeleteState(ctx, agentID, key string) error
func (s *AgentMemoryService) PromoteState(ctx, agentID, key string) error  // ephemeral → persistent
```

**API endpoints (4 new routes):**

| Method | Route | Handler |
|:-------|:------|:--------|
| GET | `/api/v1/agents/{id}/memory` | List agent's persistent state |
| POST | `/api/v1/agents/{id}/memory` | Store agent memory entry |
| GET | `/api/v1/agents/{id}/memory/search?q=` | Semantic search agent memory |
| DELETE | `/api/v1/agents/{id}/memory/{key}` | Delete memory entry |

#### B2.4 Tool Integration

Extend `remember` and `recall` internal tools to support agent-scoped mode:

```
remember:
  - When called BY an agent during mission: write to agent_state with agent_id + mission_id
  - Auto-embed via Router.Embed() for semantic recall
  - Default memory_type: 'ephemeral' (can be promoted)

recall:
  - When called BY an agent: query agent_state WHERE agent_id = caller
  - Support semantic search via context_vector
  - Return cross-mission results (not limited to current mission)
```

#### B2.5 Context Injection at Spawn

Extend `InternalToolRegistry.BuildContext()` to inject agent's prior memories:

```go
// In BuildContext(), after existing context assembly:
if agentMemSvc != nil {
    memories, _ := agentMemSvc.ListState(ctx, agentID)
    if len(memories) > 0 {
        contextBlock += "\n## Prior Memory (Cross-Mission)\n"
        for _, m := range memories {
            contextBlock += fmt.Sprintf("- [%s] %s: %s\n", m.MemoryType, m.Key, m.Value)
        }
    }
}
```

#### B2.6 Memory Consolidation Daemon

**New file:** `internal/memory/consolidation.go`

Background goroutine that periodically:
1. Scans `agent_state` for ephemeral entries older than 24h
2. Groups by agent_id
3. Sends to LLM for compression (like Archivist does for SitReps)
4. Stores compressed summary as a single `persistent` entry
5. Deletes consumed ephemeral entries
6. Runs every 6 hours (configurable)

```go
type ConsolidationDaemon struct {
    db     *pgxpool.Pool
    brain  *cognitive.Router
    ticker time.Duration  // default 6h
}

func (d *ConsolidationDaemon) StartLoop(ctx context.Context)
func (d *ConsolidationDaemon) ConsolidateAgent(ctx context.Context, agentID string) error
```

#### B2.7 Frontend: Memory Explorer Enhancement

Extend `/memory` route with agent-scoped view:

| Tab | Source | Status |
|:----|:-------|:-------|
| Hot (Working Memory) | Existing | Keep |
| Warm (SitReps) | Existing | Keep |
| Cold (Archive) | Existing | Keep |
| **Agent Memory** (NEW) | `GET /api/v1/agents/{id}/memory` | Agent timeline, search-by-agent |

**New component:** `AgentMemoryPanel.tsx`
- Agent selector dropdown (from active agent registry)
- Timeline of memory entries (key/value, origin mission, timestamp)
- Semantic search input
- Memory type badges (ephemeral/persistent/protected)
- Promote/delete actions

---

## V. Migration Window

All schema changes execute in the Phase 12 migration window:

| Migration | File | Purpose | Track |
|:----------|:-----|:--------|:------|
| 021 | `021_agent_state_cross_mission.up.sql` | Make mission_id nullable, add memory_type | B |
| 022 | `022_agent_memory_vectors.up.sql` | Add context_vector to agent_state, agent_id to agent_memories | B |
| 023 | `023_agent_state_cross_mission.down.sql` | Rollback for 021 (restore FK cascade) | B |

**Down migrations:** Phase 12 introduces proper down migrations for new schemas. Legacy
migrations (001-019) remain one-way — rolling them back would destroy the entire system.

---

## VI. Acceptance Criteria

### Track A — Hardening

- [ ] **A1.1** Zero debug code: No `DEBUG STATE` button, no DOM injection, no `[DEBUG]` console.log
- [ ] **A1.2** Zero duplicate components: Only 1 CircuitBoard, 1 LogStream, 1 TeamRoster
- [ ] **A1.3** Governance actions work: approveArtifact/rejectArtifact call backend API
- [ ] **A1.4** Error boundaries exist: error.tsx + loading.tsx in app layout
- [ ] **A2.1** API client: All fetch() calls route through `lib/api.ts`
- [ ] **A2.2** Backend tests: router, mcp, signal, provisioning packages have unit tests
- [ ] **A2.3** Temperature configurable via cognitive.yaml profile
- [ ] **A2.4** Archivist JSON parsing handles malformed LLM output gracefully
- [ ] **A3.3** Docs updated: E2E count, Phase 17 status, Ollama endpoint, tool inventory

### Track B — Agent Memory

- [ ] **B2.1** Migration 021 applied: agent_state.mission_id nullable, memory_type column
- [ ] **B2.2** Migration 022 applied: context_vector column, agent_memories.agent_id column
- [ ] **B2.3** Agent memory CRUD: 4 new API endpoints operational
- [ ] **B2.4** `remember`/`recall` tools write to agent-scoped state when called by agent
- [ ] **B2.5** BuildContext() injects prior memories for returning agents
- [ ] **B2.6** Consolidation daemon runs on schedule, compresses ephemeral → persistent
- [ ] **B2.7** `/memory` page has Agent Memory tab with search + timeline

### Verification

```bash
uvx inv core.test         # All Go tests pass (including new router/mcp/signal tests)
uvx inv interface.test    # All Vitest tests pass
uvx inv interface.build   # Production build succeeds
uvx inv interface.e2e     # All 20 Playwright specs pass
uvx inv db.migrate        # Migrations 021-022 apply cleanly
uvx inv interface.check   # Smoke test all pages (no errors)
```

---

## VII. Execution Order

### Sprint 1: Hardening (Track A — Critical + High)

| Step | Task | Estimate |
|:-----|:-----|:---------|
| 1 | A1.1 Remove all debug code (3 locations) | Small |
| 2 | A1.2 Delete duplicate components + fix imports | Small |
| 3 | A1.3 Wire governance stubs to backend API | Small |
| 4 | A1.4 Add error.tsx + loading.tsx | Small |
| 5 | A2.1 Create lib/api.ts + migrate 55 fetch() calls | Medium |
| 6 | A2.2 Write tests for router, mcp, signal, provisioning | Medium |
| 7 | A2.3 + A2.4 Fix cognitive temperature + archivist parsing | Small |

### Sprint 2: Agent Memory (Track B)

| Step | Task | Estimate |
|:-----|:-----|:---------|
| 8 | B2.1 + B2.2 Write and apply migrations 021-022 | Small |
| 9 | B2.3 Implement AgentMemoryService + 4 API handlers | Medium |
| 10 | B2.4 Extend remember/recall tools for agent scope | Medium |
| 11 | B2.5 Integrate BuildContext() with agent memory | Small |
| 12 | B2.6 Implement ConsolidationDaemon | Medium |
| 13 | B2.7 Build AgentMemoryPanel + wire to /memory page | Medium |

### Sprint 3: Polish (Track A Medium + Docs)

| Step | Task | Estimate |
|:-----|:-----|:---------|
| 14 | A3.1 Split large components (NatsWaterfall, OperationsBoard) | Medium |
| 15 | A3.2 Clean unused Zustand state | Small |
| 16 | A3.3 Fix all documentation discrepancies | Small |
| 17 | Full verification pass (all tests + smoke) | Small |

---

## VIII. Risk Register

| Risk | Impact | Mitigation |
|:-----|:-------|:-----------|
| API client refactor breaks fetch calls | HIGH | Mechanical change; run full E2E after each batch |
| Migration 021 breaks existing agent_state FK | MEDIUM | Migration is additive (nullable), not destructive |
| Consolidation daemon LLM costs | LOW | Same pattern as Archivist; rate-limited by 6h interval |
| Duplicate component deletion breaks imports | MEDIUM | Grep for all import paths before deletion |
| Agent memory volume growth | LOW | TTL expires_at + consolidation daemon prunes |

---

## IX. Out of Scope

The following are explicitly NOT part of Phase 12:

- **Phase 13:** Multi-Agent Collaboration (debate protocol, consensus detection)
- **Phase 14:** Hot-Reload Runtime (live goroutine replacement)
- **Phase 15:** Advanced Governance & RBAC (role enforcement, API keys)
- **Phase 18:** Streaming LLM (token-by-token relay)
- **Accessibility overhaul** (deferred; audit score 4/10 noted for future phase)
- **Down migrations for 001-019** (legacy; too risky to retrofit)
- **K8s deployment automation** (TODOs in provisioning/registry; Phase 20+)

---

## X. Appendix: Full Audit Data Sources

| Audit | Scope | Key Findings |
|:------|:------|:-------------|
| Go Backend | 18 packages, 56 routes, 20 migrations | 8 untested packages, 5 TODOs, clean code |
| Frontend | 85 components, 14 routes, 1886-line store | Debug code, duplicates, no API client |
| Documentation | 9 documents, cross-referenced | Phase 17 ambiguity, E2E count stale, Ollama conflict |
| Phase 12 Schema | Migrations 018-019 | mission_id FK CASCADE blocks cross-mission memory |

---

*PRD authored from full-stack audit data. All line numbers and file paths verified against codebase as of 2026-02-18.*

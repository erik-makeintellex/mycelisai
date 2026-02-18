# QA Test Plan: Council Chat API + Mission Control Restructure

> **Scope:** Standardized Council Chat API (backend + frontend), OperationsBoard, `/approvals` Team Proposals tab, and regression of existing features.
>
> **Prerequisites:** All tests require the full stack running unless noted otherwise.
> ```
> Terminal 1: uvx inv k8s.bridge
> Terminal 2: uvx inv core.run
> Terminal 3: uvx inv interface.dev
> ```

---

## 1. Backend API Contract Tests

Run these with `curl` or any HTTP client against `http://localhost:8081`.

### 1.1 GET /api/v1/council/members

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 1.1.1 | Returns all council members | `GET /api/v1/council/members` | `200` — JSON body: `{ "ok": true, "data": [...] }` where `data` is an array of 5 objects |
| 1.1.2 | Member shape is correct | Inspect each item in `data` | Each has `{ "id": string, "role": string, "team": string }` |
| 1.1.3 | Admin is included | Check `data` for `id: "admin"` | Present with `role: "admin"`, `team: "admin-core"` |
| 1.1.4 | All 4 council members present | Check for architect, coder, creative, sentry | `council-architect` (role: architect, team: council-core), `council-coder` (role: coder), `council-creative` (role: creative), `council-sentry` (role: sentry) |
| 1.1.5 | No mission agents leak through | Activate a mission, then re-query | Only admin-core and council-core members appear — no mission-spawned agents |

### 1.2 POST /api/v1/council/{member}/chat — Happy Path

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 1.2.1 | Chat with admin | POST to `/api/v1/council/admin/chat` with `{ "messages": [{"role":"user","content":"Hello"}] }` | `200` — `{ "ok": true, "data": { "meta": {...}, "signal_type": "chat_response", "trust_score": 0.5, "payload": {...} } }` |
| 1.2.2 | Response envelope structure | Inspect `data` field | `meta.source_node` = `"admin"`, `meta.timestamp` = ISO 8601, `signal_type` = `"chat_response"`, `trust_score` = `0.5` |
| 1.2.3 | Payload contains text | Inspect `data.payload` | `{ "text": "<non-empty string>" }` — the agent's actual response |
| 1.2.4 | Chat with architect | POST to `/api/v1/council/council-architect/chat` with same body | `200` — same envelope structure, `meta.source_node` = `"council-architect"` |
| 1.2.5 | Chat with coder | POST to `/api/v1/council/council-coder/chat` | `200` — `meta.source_node` = `"council-coder"` |
| 1.2.6 | Chat with creative | POST to `/api/v1/council/council-creative/chat` | `200` — `meta.source_node` = `"council-creative"` |
| 1.2.7 | Chat with sentry | POST to `/api/v1/council/council-sentry/chat` | `200` — `meta.source_node` = `"council-sentry"` |
| 1.2.8 | Multi-turn conversation | Send 3 messages in sequence, passing full history each time | Agent response shows awareness of prior turns (not just last message) |
| 1.2.9 | Trust score is cognitive default | Check `trust_score` on all responses | Always `0.5` (the `TrustScoreCognitive` constant) |

### 1.3 POST /api/v1/council/{member}/chat — Error Cases

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 1.3.1 | Unknown member | POST to `/api/v1/council/nonexistent/chat` | `404` — `{ "ok": false, "error": "Unknown council member: nonexistent" }` |
| 1.3.2 | Empty messages array | POST with `{ "messages": [] }` | `400` — `{ "ok": false, "error": "Empty conversation" }` |
| 1.3.3 | Missing messages field | POST with `{}` | `400` — `{ "ok": false, "error": "Empty conversation" }` |
| 1.3.4 | Malformed JSON | POST with `not-json` | `400` — `{ "ok": false, "error": "Bad JSON" }` |
| 1.3.5 | Missing member param | POST to `/api/v1/council//chat` | `404` (Go mux won't match the route) |
| 1.3.6 | Wrong HTTP method | GET to `/api/v1/council/admin/chat` | `405` Method Not Allowed (Go 1.26 mux enforces `POST`) |

### 1.4 Degraded Mode — NATS Offline

> Stop NATS or run core without NATS connection.

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 1.4.1 | Council chat without NATS | POST to `/api/v1/council/admin/chat` | `503` — `{ "ok": false, "error": "Swarm offline — council agents unavailable. Start the organism first." }` |
| 1.4.2 | Members list without Soma | Stop core entirely, query members | Connection refused (expected — no server) |
| 1.4.3 | Old /api/v1/chat without NATS | POST to `/api/v1/chat` | `503` — JSON error about swarm offline (legacy format, not APIResponse) |

### 1.5 Backward Compatibility

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 1.5.1 | Legacy chat endpoint works | POST to `/api/v1/chat` with `{ "messages": [{"role":"user","content":"ping"}] }` | `200` — plain text response (NOT APIResponse wrapped — legacy behavior preserved) |
| 1.5.2 | Broadcast endpoint works | POST to `/api/v1/swarm/broadcast` | `200` — existing broadcast behavior unchanged |
| 1.5.3 | SSE stream works | GET `/api/v1/stream` (EventSource) | SSE connection established, signals flow |

---

## 2. Frontend Integration Tests

Open `http://localhost:3000` in browser.

### 2.1 Mission Control — Council Selector

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 2.1.1 | Selector populated | Load Mission Control (`/`) | Header shows dropdown with council members (ADMIN, ARCHITECT, CODER, CREATIVE, SENTRY) — not just static "Admin" label |
| 2.1.2 | Fallback when backend offline | Stop backend, reload page | Dropdown shows single "Admin" option (fallback) — no crash |
| 2.1.3 | Switching target | Select "ARCHITECT" from dropdown | Dropdown value updates, input placeholder changes to "Ask the architect..." |
| 2.1.4 | Placeholder is dynamic | Cycle through all members | Placeholder text reflects the selected member's role each time |
| 2.1.5 | Selector hidden in broadcast | Toggle broadcast mode ON | Dropdown replaced by "Broadcast" label with megaphone icon |
| 2.1.6 | Selector returns on broadcast off | Toggle broadcast mode OFF | Dropdown reappears with last-selected member still active |

### 2.2 Mission Control — Council Chat Flow

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 2.2.1 | Send to admin | Select Admin, type "Hello", press Enter | User bubble appears right-aligned, then response bubble appears left-aligned with "ADMIN" source label and "T:0.5" trust badge |
| 2.2.2 | Send to architect | Select Architect, type "What's the system design?", Enter | Response shows "ARCHITECT" source label and "T:0.5" badge (yellow color) |
| 2.2.3 | Source label formatting | Send message to council-architect | Source label shows "ARCHITECT" (not "council-architect" — prefix stripped) |
| 2.2.4 | Trust badge color | Observe trust badge on any response | "T:0.5" shown in yellow/warning color (score >= 0.5 but < 0.8) |
| 2.2.5 | Tools pills render | If agent uses tools (e.g., admin uses `search_memory`) | Small purple pills appear below the message showing tool names |
| 2.2.6 | Multi-turn context | Send follow-up questions | Agent responses show awareness of prior messages in the conversation |
| 2.2.7 | Loading indicator | Send a message | Animated dots appear while waiting for response, disappear when response arrives |
| 2.2.8 | Clear chat | Click trash icon after sending messages | All messages cleared, empty state shown with updated placeholder |
| 2.2.9 | Scroll to bottom | Send enough messages to overflow | Chat auto-scrolls to newest message |

### 2.3 Mission Control — Error States

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 2.3.1 | Backend offline | Stop core, send a message | Error message appears in chat as a council-role bubble with source_node, error bar appears at top |
| 2.3.2 | NATS offline | Run core without NATS, send a message | Structured error: "Swarm offline — council agents unavailable..." rendered as chat bubble |
| 2.3.3 | Network error | Disconnect network, send a message | "Chat failed" error message in chat bubble |
| 2.3.4 | Error bar dismissal | Trigger an error, then send a successful message | Error bar content shows the error text, clears on next successful interaction |

### 2.4 Mission Control — Broadcast Mode (Regression)

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 2.4.1 | Toggle broadcast | Click megaphone icon | Header switches to "Broadcast" mode, warning banner appears, input border turns yellow |
| 2.4.2 | Send broadcast | Type "Status report" in broadcast mode, Enter | Message prefixed with `[BROADCAST]` in orange bubble, response from broadcast endpoint |
| 2.4.3 | `/all` shortcut | In normal mode, type `/all Status report`, Enter | Treated as broadcast (message sent via broadcastToSwarm, not sendMissionChat) |
| 2.4.4 | Broadcast mode styling | Observe broadcast mode | Yellow border on input, yellow send button, warning-colored loading dots |

---

## 3. Mission Control — OperationsBoard (Regression)

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 3.1 | Board renders | Load Mission Control | Center column shows OperationsBoard with 3 sections: Priority Alerts, Standing Workloads, Missions |
| 3.2 | Priority alerts from SSE | Trigger a governance_halt or error signal | Alert appears in the Priority Alerts section with correct icon and color |
| 3.3 | Alerts collapsible | Click section header when alerts exist | Section collapses/expands |
| 3.4 | Standing workloads listed | With core running (teams active) | Standing teams (admin-core, council-core) appear with status dots |
| 3.5 | Missions listed | With at least one active mission | Mission card shows intent, status badge (active/completed/failed), agent count |
| 3.6 | Mission click navigates | Click a mission row | Browser navigates to `/missions/{id}/teams` |
| 3.7 | Empty state | No missions, no alerts | Sections show graceful empty states (not broken/blank) |
| 3.8 | Polling refreshes | Wait 10-15 seconds | Data auto-refreshes (teamsDetail at 10s, missions at 15s) |

---

## 4. Approvals Page — Team Proposals Tab (Regression)

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 4.1 | Three tabs visible | Navigate to `/approvals` | Three tabs: "Approvals", "Policy", "Team Proposals" |
| 4.2 | Proposals tab loads | Click "Team Proposals" tab | ManifestationPanel renders (may be empty if no proposals) |
| 4.3 | Tab switching | Click between all three tabs | Content swaps correctly, no flash/crash |
| 4.4 | Approvals tab works | Click "Approvals" tab | Governance approval cards render (or empty state) |
| 4.5 | Policy tab works | Click "Policy" tab | Policy editor renders with current config |

---

## 5. Layout & Navigation Regression

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 5.1 | 3-column grid stable | Resize browser window | Left (CognitiveStatus + Sensors), Center (OperationsBoard), Right (Chat) columns maintain proportions with `minmax` CSS grid |
| 5.2 | Header functional | Check Mission Control header | "MISSION CONTROL" title, SIGNAL:LIVE/OFFLINE badge, "NEW MISSION" button (→ /wiring), Settings button (→ /settings) |
| 5.3 | Cognitive status panel | Check left column | CognitiveStatusPanel shows provider status (online/offline) |
| 5.4 | Sensor library | Check left column below cognitive | SensorLibrary renders with grouped subscriptions |
| 5.5 | Telemetry row | Check below header | TelemetryRow renders compute metrics |
| 5.6 | Wiring page loads | Navigate to `/wiring` | ArchitectChat + CircuitBoard + NatsWaterfall render without error |
| 5.7 | Teams page loads | Navigate to `/teams` | Team cards render with standing + mission teams |
| 5.8 | Settings page loads | Navigate to `/settings` | 4 tabs: Profile, Teams, Cognitive Matrix, MCP Tools |

---

## 6. Cross-Browser / Responsive

| # | Test | Steps | Expected Result |
|---|------|-------|-----------------|
| 6.1 | Chrome latest | Full test suite above | All pass |
| 6.2 | Firefox latest | Core council chat flow (2.1-2.3) | All pass |
| 6.3 | 1920x1080 | Full layout check | 3-column grid renders properly, no overflow |
| 6.4 | 1366x768 (laptop) | Full layout check | Grid adapts, columns may narrow but remain usable |
| 6.5 | Council selector on small screen | Resize to ~1200px width | Dropdown still accessible, not clipped |

---

## 7. API Response Schema Validation

All new council endpoints return `APIResponse` envelopes. Verify the schema contract:

### Success envelope
```json
{
    "ok": true,
    "data": { ... }
}
```
- `ok` is boolean `true`
- `data` contains the payload (CTSEnvelope for chat, array for members)
- `error` field is absent or empty

### Error envelope
```json
{
    "ok": false,
    "error": "Human-readable error message"
}
```
- `ok` is boolean `false`
- `error` is a non-empty string
- `data` field is absent or null

### CTSEnvelope within chat response `data`
```json
{
    "meta": {
        "source_node": "admin",
        "timestamp": "2026-02-16T12:00:00Z"
    },
    "signal_type": "chat_response",
    "trust_score": 0.5,
    "payload": "{\"text\":\"...\",\"consultations\":null,\"tools_used\":null}"
}
```
- `meta.source_node` matches the `{member}` from the URL
- `meta.timestamp` is a valid ISO 8601 datetime (recent, not zero)
- `signal_type` is always `"chat_response"`
- `trust_score` is always `0.5` (cognitive default)
- `payload` is a JSON-encoded string containing `ChatResponsePayload`
- `payload.text` is the agent's response text (non-empty on success)
- `payload.consultations` and `payload.tools_used` are null/empty for now (agents return plain text today — structured fields are a future phase)

---

## 8. Quick Smoke Commands

```bash
# 1. Members list
curl -s http://localhost:8081/api/v1/council/members | python -m json.tool

# 2. Chat with admin
curl -s -X POST http://localhost:8081/api/v1/council/admin/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"What teams are active?"}]}' \
  | python -m json.tool

# 3. Chat with architect
curl -s -X POST http://localhost:8081/api/v1/council/council-architect/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"Describe the system architecture."}]}' \
  | python -m json.tool

# 4. Unknown member (expect 404)
curl -s -X POST http://localhost:8081/api/v1/council/nobody/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"hello"}]}' \
  | python -m json.tool

# 5. Legacy endpoint still works
curl -s -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"ping"}]}'

# 6. Via frontend proxy (Next.js rewrite)
curl -s http://localhost:3000/api/v1/council/members | python -m json.tool
```

---

## Test Result Summary Template

| Section | Total | Pass | Fail | Skip | Notes |
|---------|-------|------|------|------|-------|
| 1. Backend API | 20 | | | | |
| 2. Frontend Chat | 18 | | | | |
| 3. OperationsBoard | 8 | | | | |
| 4. Approvals Tabs | 5 | | | | |
| 5. Layout Regression | 8 | | | | |
| 6. Cross-Browser | 5 | | | | |
| 7. Schema Validation | 3 | | | | |
| **Total** | **67** | | | | |

# Agent Acceptance Runbook (API + GUI)

> Navigation: [User Acceptance Runbook](REMOTE_USER_TESTING.md) | [API Reference](API_REFERENCE.md) | [Truthful Delivery](architecture-library/TRUTHFUL_DELIVERY_AND_NEXT_ARCHITECTURE.md)

This runbook is for an **automated user agent**: a browser-driving AI or a scripted tester. The agent walks Mycelis the way a user would, through the HTTP API **and** the GUI, then reports what really happened. The human-oriented walkthrough lives in [REMOTE_USER_TESTING.md](REMOTE_USER_TESTING.md); this document is the precise, repeatable version.

## Ground rules (read first)
1. **Judge outcomes, not labels.** A step passes only when the real result exists. Open the file and read its content, recall the saved fact, and confirm the stop actually happened. A "Completed"/"Verified" badge alone is never a pass. Any fake success is a **P0 finding**.
2. **Never print or paste secrets.** Read `MYCELIS_API_KEY` from `.env` into a shell variable. The local-admin password comes from the owner out of band. Never echo either value into logs, screenshots, or the report.
3. **At most 3 browser sessions at once**, whether tabs, Playwright workers, or agents.
4. **Do not change live configuration** except where a step says so. Revert anything you change in the same step. Never disable the provider that Soma currently uses.
5. **Test data is fictional and idempotent.** Use the "Juniper & Rye Bakery" profile below. Check that it exists before saving a second copy.
6. **Stop and report** if a step would need real credentials, payments, or external accounts.

## Setup
- The stack runs from `dev`: `uv run inv compose.up --build`, then `uv run inv compose.migrate` and `uv run inv compose.health`. It is healthy when every profile shows `vllm [root] OK`.
- API base: `http://127.0.0.1:8081`. GUI: `http://127.0.0.1:3000`.
- Set up the shell:

```bash
KEY=$(grep -E '^MYCELIS_API_KEY=' .env | cut -d= -f2-)
H=(-H "Authorization: Bearer $KEY" -H "Content-Type: application/json")
```

- GUI sign-in: `/login` → "Sign in as local admin" (username `admin`; the owner provides the password). A standard user signs in with Google, and only if the owner provides such an account.
- Viewports: desktop **1366×768** and phone **390×844**.
- **Automated baseline first:** run `uv run python ops/live_journey_probe.py` and paste its PASS/FAIL table into the report. The manual journeys below go deeper than the probe.

## Journeys
Each journey lists the **API** steps, then the **GUI** steps, then the **Pass** criteria. Record every HTTP status and the key response fields.

### J0. System health and honest status
- **API**:
  - `GET /api/v1/services/status`: expect `governance` = online (not degraded), `cognitive` online.
  - `GET /api/v1/cognitive/status`: `profile_route_health` = `ok`.
  - `GET /api/v1/sensors`: returns only real sources, or `{"count":0,"status":"none_configured"}`.
- **GUI**: open the Status drawer from the rail. Every status is in plain words, with no raw ids for a standard user.
- **Pass**:
  - Nothing claims "online" without a real source.
  - A non-admin sees no endpoints, model URLs, or env var names.

### J1. Ask Soma a question
- **API**: `POST /api/v1/chat` with `{"messages":[{"role":"user","content":"In one sentence, what can you do for me?"}]}`. Expect `data.payload.text` to be non-empty.
- **GUI**: on `/dashboard`, type into "Tell Soma what you want…" and press Enter.
- **Pass**:
  - A real answer arrives in about 30 s or less.
  - The answer contains no jargon ("neural organism", "swarm", "governed").

### J2. Get a deliverable made (the core journey)
- **API**:
  1. Send `POST /api/v1/chat` with "Create a markdown file named `aa-bakery-<timestamp>.md` listing three concrete ways a small bakery can use AI assistants."
  2. Expect `mode=proposal` and `proposal.confirm_token`. `draft_previews[]` must contain the file content, and must not be a template ("What you asked for").
  3. Send `POST /api/v1/intent/confirm-action` with `{"confirm_token":"…"}`. Expect `confirmed=true`.
  4. `GET /api/v1/workspace/files/view?path=<file>`: expect 200, `content-type: text/plain`, and content about bakeries.
- **GUI**:
  1. Make the same request.
  2. The proposal card shows **what will be written** (the preview) and a **Start** button.
  3. Click Start, then **Open result**.
  4. The viewer renders the text.
- **Pass**:
  - The preview equals the executed content.
  - The content is real and relevant.
  - It is "verified" only because the file was read back.
  - The viewer is not blank.
  - There is no stale "Building… running, not complete" card after completion.
  - There is no echoed "start" message.

### J3. Maintained memory: save, recall, cite
- **API**:
  1. `GET /api/v1/memory/deployment-context`. If "Rye Bakery" is absent, `POST` the profile:

```json
{"title":"Company profile: Juniper & Rye Bakery","source_label":"owner-provided company profile",
 "content":"Juniper & Rye Bakery is on East 6th Street in Austin. Weekend special: blueberry-lavender scones, 3 for $10. Always sign off with \"See you at the counter.\"",
 "knowledge_class":"company_knowledge","source_kind":"user_note","content_domain":"operations",
 "visibility":"global","sensitivity_class":"role_scoped","trust_class":"user_provided"}
```

  Expect 201 with an honest `embedding_status`.

  2. In a new conversation, send `POST /api/v1/chat` with "Write a two-line promo for our weekend special."
- **GUI**:
  1. Resources → **Deployment Context**: save the same note using the defaults, with the classification left under "More options".
  2. On failure, a visible error appears, never a green "Saved".
  3. Then ask Soma the promo question.
- **Pass**:
  - The answer uses the stored facts (blueberry-lavender, 3 for $10). These facts cannot be guessed.
  - `context_sources[]` lists the profile with `used=true`.
  - The sign-off "See you at the counter" is a soft check.
- **After M2 merges**:
  - Archive the note. It disappears from recall and citations, and Restore brings it back.
  - Delete asks for confirmation, then the note is gone.
  - Non-owners get a plain "managed by admins" blocker.

### J4. Approvals and governance
- **API** (admin):
  - `GET /api/v1/governance/policy`: expect 200.
  - A tier-2 request such as "Broadcast a status update to every team", or one that uses an MCP tool, yields a proposal with an approver requirement.
- **GUI**:
  - The proposal shows **"Needs admin approval" before Start**.
  - An admin can approve it; self-approval is allowed and audited.
- **Standard user, if available**:
  - Governance GETs return 403, rendered as an access state, never "All Clear" or an empty policy.
  - Confirming a tier-2 proposal gives "needs admin approval", and the proposal is kept.
- **Pass**:
  - No false success.
  - Tokens are kept on refusals.
  - Wording follows the plain vocabulary: no token, scope, or digest in headlines.

### J5. AI engines (read-only unless stated)
- **API**:
  - `GET /api/v1/cognitive/status`: as admin you get the full view; as a non-admin you see no endpoints or model ids.
  - `GET /api/v1/brains`: a non-admin sees only id, type, enabled, location, data_boundary, and status.
- **GUI**: Resources → **AI Engines**: a summary of which engine does the work, then the table under Advanced.
- **Optional write test** (revert immediately): toggling a provider that is **not** in use is allowed. Toggling the provider that is in use (`vllm`) must return **409 `provider_bound`** with a plain explanation and a "Use default engine" option. **Do not** try to force it.
- **Pass**:
  - Honest refusal; nothing changes on a 409.

### J6. Team handoffs (H1)
- **API** (admin): `GET /api/v1/exchange/handoffs` returns the handoff chain (created → handed → acknowledged) with real agent, team, and run provenance.
- **GUI**: Soma → ask the architect team to prepare a short brief and hand it to the development team. Watch Resources → Exchange "Team handoffs" for the chain.
- **Pass**:
  - Only team leads hand off.
  - A handoff never shows "handed" before it is dispatched.
  - An out-of-run handoff shows "needs approval".
  - A reader outside the scope is denied.
  - If you can't trigger a live handoff from chat, report it as a **gap**, not as a pass.

### J7. Web search and capabilities
- **API**:
  - `POST /api/v1/chat` with "Search the web for the current population of Austin, Texas and cite a source." The answer is real, with a URL, or honestly blocked.
  - The connector install endpoint returns **501 `connector_deployment_unavailable`**, never a fake "provisioning".
- **GUI**: Resources → **Capabilities**. The copy says what's ready; there are no fake "online" items.
- **Pass**:
  - A failed search is reported as failed, never as success.

### J8. Token budgets (after B1 merges)
- **API**: `GET` the effective budget and usage for a team/run (see [B1 contract](architecture-library/B1_TOKEN_BUDGETS_CONTRACT.md)). An admin lowers a test team's budget, runs a long request, and restores the budget afterwards.
- **GUI**: the usage summary ("used X of Y"). On exhaustion, a plain stop message appears with a next action.
- **Pass**:
  - The run stops with `token_budget_exhausted`.
  - The partial output is **not** shown as complete.
  - A warning appears at 80%.
  - A non-admin cannot change budgets (403).

### J9. Tool scope and isolation (after S7b merges)
- **API/GUI**:
  - Ask a team agent to read another group's files. It is denied with a plain message.
  - An approved plan cannot run a tool the agent never declared.
  - A non-owner cannot cancel someone else's proposal (403, proposal kept).
- **Pass**:
  - Denials are real, and nothing executes.

### J10. Find and reuse results
- **GUI**:
  - Return to Soma after a reload. The latest result is still available.
  - Check Resources → Deliverables and Work for your J2 file.
- **Pass/Gap**:
  - A dead end (for example "install Filesystem MCP to see files") or a result missing from Work is a **known gap**. Record it with the exact copy you saw.

### J11. Phone and accessibility sweep
- Repeat J1, J2, and J3 at 390×844 and check:
  - No horizontal scroll.
  - Primary actions are reachable.
  - Blocker titles are text, not color alone.
  - Buttons have accessible names (use the accessibility tree).

## Report format
Deliver one table and one findings list:

| Journey | Path | Result | Evidence |
|---|---|---|---|
| J2 | API | PASS / FAIL / GAP | HTTP codes, key fields, file excerpt |
| J2 | GUI | … | screenshot path, exact copy seen |

Findings go **most severe first**. Each finding carries:
- a severity: P0 = fake success or security; P1 = broken journey; P2 = confusing wording or layout
- exact repro steps
- expected vs actual
- the exact UI text or response body (secrets redacted)
- the suspected area: backend, interface, or U1-owned UI

Never mark a journey PASS without real-outcome evidence.

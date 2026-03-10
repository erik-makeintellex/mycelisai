# Browser QA Test Plan: Soma Workspace Chat

## Application
Mycelis Cortex V7 - Workspace UI (`/dashboard`)

## Context and Strategy
Existing Playwright tests (`workspace-live-backend`, `proposals`, `error-scenarios`, `v7-operational-ux`) heavily cover the "Launch Crew" modal proposal generation, system health fallback UI (Degraded Mode), and API hard-crashes. 

This manual Browser QA pass will focus strictly on the **inline Soma Chat experience**, filling the gaps in exploratory testing, UI edge cases, and V7 terminal state progression within the primary message feed.

---

## Part 1: Structured functional paths (Happy Path)

### Test 1.1: Direct Answer (Non-Mutating Intent)
- **Purpose:** Ensure Soma can answer a question without unnecessarily pulling in the council or triggering a mutation proposal.
- **Steps:**
  1. Open `/dashboard`.
  2. Type: "Summarize the current UI V7 design objectives."
  3. Click Send.
  4. Wait for the response.
- **Expected Result:**
  - `SomaActivityIndicator` shows streaming UI.
  - Final message bubble shows the answer.
  - State is `answer` (read-only).
  - No `proposal` card or Confirm/Cancel buttons appear.

### Test 1.2: Inline Governed Mutation (Proposal Generation)
- **Purpose:** Test that an explicit mutation request in the chat (not the Launch Crew modal) correctly halts at the `proposal` terminal state.
- **Steps:**
  1. Open `/dashboard`.
  2. Type: "Create a simple python file named `hello_world.py` in the workspace."
  3. Click Send.
- **Expected Result:**
  - Soma responds with markdown text explaining the action.
  - An inline `proposal` card appears below the message.
  - Risk Level is displayed.
  - Confirm and Cancel controls are visible and actionable.

### Test 1.3: Inline Proposal Cancellation
- **Purpose:** Verify the cancellation flow correctly resets system intent without mutation.
- **Steps:**
  1. Follow steps in Test 1.2 to generate a proposal.
  2. Click the "Cancel" button on the proposal card.
- **Expected Result:**
  - The proposal card is dismissed or marked cancelled.
  - No file actually gets written to the backend workspace.
  - Chat remains usable for the next intent.

---

## Part 2: Exploratory & UI Edge Cases

### Test 2.1: Mid-Stream Interruption (Page Refresh)
- **Purpose:** Ensure the UI can recover or gracefully handle a user refreshing the browser while a SSE request to Soma is actively streaming.
- **Steps:**
  1. Open `/dashboard`.
  2. Request a complex task: "Analyze the last 5 mission runs and give me a detailed table."
  3. While `SomaActivityIndicator` is active (e.g. "Consulting..."), hit F5 / refresh the page.
- **Expected Result:**
  - Workspace reloads successfully without fatal React hydration errors.
  - The state resets cleanly, or the disconnected request is gracefully ignored.

### Test 2.2: Input Sanitization & Empty States
- **Purpose:** Check how the chat input handles extreme boundary values.
- **Steps:**
  1. Open `/dashboard`.
  2. Try submitting empty text, or just spaces.
  3. Try submitting common XSS vectors: `<script>alert("test")</script>`
- **Expected Result:**
  - Empty submissions are prevented (button disabled or input ignored).
  - HTML tags are escaped and rendered as raw text in the message bubble, not executed as DOM elements.

### Test 2.3: Visual Breakage on Large Layouts
- **Purpose:** Verify that oversized markdown elements (like wide tables or extremely long code blocks) do not break the 68% bounded width of the chat pane.
- **Steps:**
  1. Ask Soma: "Generate a markdown table with 10 columns and 5 rows containing dummy data."
- **Expected Result:**
  - The table renders via GFM.
  - The container handles horizontal overflow via scrolling (`overflow-x-auto`), keeping the rest of the message layout intact without pushing the OpsOverview off-screen.

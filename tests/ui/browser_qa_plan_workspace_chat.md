# Browser QA Test Plan: Soma Workspace Chat

## Application
Mycelis Cortex V8 Workspace (`/dashboard`)

## Context and Strategy
Existing Playwright suites (`workspace-live-backend`, `proposals`, `error-scenarios`, `v7-operational-ux`) already cover Launch Crew proposal flows, degraded-state UI, and proxy regressions. V8 delivery now expects governed inline chat behavior backed by the same template → instantiation → inheritance → precedence contract, so this manual pass concentrates on the **inline Soma chat** happy paths and edge cases that the automated suites only sample.

The focus is validating the four canonical terminal states (`answer`, `proposal`, `execution_result`, `blocker`) and confirming Soma remains direct-first unless the operator explicitly triggers a council consultation.

---

## Release-Decision Rules

- If Soma is central in layout but not central in behavior, the product is not ready.
- If support panels are visible but do not help explain state, the product is not ready.
- If the operator cannot re-enter and resume with confidence, the product is not ready.

## Release Blockers (Failure Conditions)

The following conditions are release blockers:
- Soma is not obviously the primary entry within 10 seconds.
- The main conversation area is visible but does not feel actionable.
- Team mode feels like a second competing front door.
- Support panels update, but the user cannot explain why they changed.
- Returning to the workspace creates uncertainty about current org or current state.

---

## Test Execution Passes

### Pass 1: Direct Answer (Non-Mutating Intent)
- **Purpose:** Ensure Soma can answer a question without unnecessarily pulling in the council or triggering a mutation proposal.
- **Steps:**
  1. Open `/dashboard`.
  2. Type: "Summarize the current Workspace V8 design objectives."
  3. Click Send.
  4. Wait for the response.
- **Expected Result:**
  - `SomaActivityIndicator` shows streaming UI.
  - Final message bubble shows the answer.
  - State is `answer` (read-only).

### Pass 2: Default Path Isolation
- **Purpose:** Prove the default path stands on its own without requiring team mode.
- **Steps:**
  1. Open `/dashboard`.
  2. Conduct a series of 3-4 tasks (e.g., querying data, simple requests).
  3. **Strict Constraint:** The user stays *only* in the main Soma conversation and *never* enters team mode.
- **Expected Result:**
  - The user can successfully complete tasks and understand context without ever feeling forced to enter team mode.

### Pass 3: Inline Governed Mutation (Causality & Latency)
- **Purpose:** Test explicit mutation in chat and measure perceived causality.
- **Steps:**
  1. Open `/dashboard`.
  2. Type: "Create a simple python file named `hello_world.py` in the workspace."
  3. Click Send.
  4. **Record:** Report the exact elapsed time from action to visible UI change.
- **Expected Result:**
  - Soma responds with an inline `proposal` card.
  - The timeline separating user action and visible state change is tight enough to feel strictly causal, avoiding vague background motion.

### Pass 4: Inline Proposal Cancellation
- **Purpose:** Verify the cancellation flow correctly resets system intent without mutation.
- **Steps:**
  1. Generate a proposal as in Pass 3.
  2. Click the "Cancel" button.
- **Expected Result:**
  - Proposal is dismissed or marked cancelled. No back-end mutation occurs.

### Pass 5: Mid-Stream Interruption & Navigation
- **Purpose:** Ensure the UI recovers gracefully from navigation interruptions during active streams.
- **Steps:**
  1. Ask a complex task that initiates streaming.
  2. While it streams, hit F5 or refresh the page.
  3. Start a new complex task. While it streams, explicitly test the browser **Back**, then **Forward** buttons.
- **Expected Result:**
  - State resets cleanly. No fatal React hydration errors. The operator can re-enter and resume with confidence.

### Pass 6: Wording and Content Audit
- **Purpose:** Check for user-friendly terminology and identify leaking architecture concepts.
- **Steps:**
  1. Review all side panels, modal dialogs, and Soma responses encountered.
  2. **Require:** Note whether any panel uses wording that feels internal, architectural, or dev-facing rather than operator-centric.
- **Expected Result:**
  - No raw internal architecture terms are exposed in the standard UI surfaces.

### Pass 7: Visual Breakage on Large Layouts
- **Purpose:** Verify container constraints on oversized content.
- **Steps:**
  1. Ask Soma to generate a huge markdown table or long code block.
- **Expected Result:**
  - Container handles horizontal overflow without breaking layout.

### Pass 8: Failure Recovery & Trust
- **Purpose:** Assess whether the system regains operator trust after a failure.
- **Steps:**
  1. Deliberately trigger an error scenario or impossible request.
  2. Observe the failure state and attempt to retry or recover.
  3. **Require:** The tester must explicitly state whether *trust recovered* after the failure, not just whether a "retry" button existed.
- **Expected Result:**
  - The failure is explained clearly, recovery options are present, and the user feels confident continuing.

---

## Best Output Shape

The testing agent should return each finding in the following format to ensure the next fix pass is easy to triage:

- **Severity:** [e.g., Blocker, High, Medium, Low]
- **Surface:** [e.g., Main Soma Chat, Support Panel, Proposal Card]
- **User action:** [What the user did]
- **Expected:** [What should have happened]
- **Observed:** [What actually happened]
- **Why it matters:** [Impact on trust, causality, etc.]

---

## Recommended Final Prompt Add-on

*Append this text to the end of the testing agent prompt:*

```text
For every criticism, include one concrete piece of visible evidence.
For every positive claim, include one concrete reason it earned trust.
If something feels "off," explain whether it is:
- clarity
- hierarchy
- continuity
- causality
- navigation
- trust
```

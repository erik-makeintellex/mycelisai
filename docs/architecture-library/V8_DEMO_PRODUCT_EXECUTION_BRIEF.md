# V8 Demo Product Execution Brief

> Status: ACTIVE
> Last Updated: 2026-04-04
> Source Plan: `V8_DEMO_PRODUCT_STRIKE_TEAM_PLAN.md`
> Purpose: Actuate the demo-product strike team with concrete first deliverables for product framing, feature preservation, demo flow, and UI verification.

---

## Primary Product Statement

Mycelis is the governed AI Organization product where Soma helps operators plan, act, review, and retain what matters without turning the experience into a fragile agent console.

This brief exists to make that sentence true in the product experience, not just in architecture language.

---

## Team Engagement Status

- `ACTIVE` Product Narrative Team
- `ACTIVE` Default Experience Team
- `ACTIVE` Capability Preservation Team
- `ACTIVE` Memory, Continuity, and Trust Team
- `ACTIVE` Demo Scenario Team
- `ACTIVE` UI Testing Agentry Team
- `ACTIVE` Release and Ops Team

Engagement approval standard has been met:

- the default product story is defined
- advanced power preservation is explicitly owned
- the first shared deliverables are now concrete

---

## Current Platform Review

### What is already strong

- default navigation already keeps the main story focused on `AI Organization`, the current organization, and `Docs`
- advanced routes for `Resources`, `Memory`, and `System` already exist behind Advanced mode instead of crowding the default product flow
- the governed product spine already exists:
  - direct-answer Soma path
  - proposal-first mutation path
  - approval/cancel behavior
  - recent activity and continuity visibility
- organization creation already opens with product language instead of a blank assistant thread

### Current drift the teams must now resolve

- some repo-facing product language still uses `Learning` / `Learning & Context` where the actual product should read as `Memory & Continuity` or retained knowledge
- some workspace/internal component naming still reflects implementation-era concepts even when user-facing copy has been improved
- the README still carries older product-language examples that are no longer the best representation of the shipped experience
- the default product story is present, but it still needs one explicit route-by-route acceptance pass so a technical partner sees a product first and depth second

### Shared product truth for this lane

The default story must now read in this order:

1. create an AI Organization
2. work with Soma
3. see recent activity and approvals
4. confirm governed actions when needed
5. understand what the organization retained

Everything else remains important, but must stop competing with that story by default.

---

## Deliverable 1: Default vs Advanced Surface Inventory

## Default product surfaces

These are the surfaces that must read as obvious product first.

### 1. Landing

- Route: `/`
- Purpose: explain what Mycelis is in minutes
- Must communicate:
  - AI Organization
  - Soma-first operation
  - governed execution
  - visible reviews and continuity
- Must avoid:
  - swarm-console framing
  - architecture jargon
  - raw infrastructure vocabulary

### 2. AI Organization Entry

- Route: `/dashboard`
- Current implementation: `CreateOrganizationEntry`
- Purpose: start with a product action, not a blank assistant thread
- Must communicate:
  - create an AI Organization
  - choose a starting point
  - enter the organization home
- Must avoid:
  - advanced setup overload
  - operator confusion about what happens next

### 3. AI Organization Home / Soma Workspace

- Route: `/organizations/[id]`
- Current implementation: `OrganizationContextShell`
- Purpose: prove the main product in one screen
- Must communicate:
  - Soma is the working counterpart
  - the organization structure is visible
  - recent activity is visible
  - approvals and continuity are understandable
- Must avoid:
  - parallel front doors
  - technical telemetry as the main story
  - generic chatbot feeling

### 4. Automations

- Route: `/automations`
- Default-safe tabs:
  - `Active Automations`
  - `Trigger Rules`
  - `Approvals`
- Purpose: reinforce trust and continuity around ongoing work
- Must communicate:
  - what is reviewing work
  - what needs approval
  - how ongoing checks support the organization

### 5. Settings

- Route: `/settings`
- Default-safe tabs:
  - `Profile`
  - `Mission Profiles`
  - `People & Access`
- Purpose: support operator identity and bounded setup without collapsing into platform complexity

### 6. Docs

- Route: `/docs`
- Purpose: support explanation and evaluation
- Use in demo only as reinforcement, not as the primary proof surface

## Advanced retained-power surfaces

These must remain available, but they must not compete with the default product story.

### 1. Resources

- Route: `/resources`
- Reason retained:
  - connected tools
  - exchange
  - workspace files
  - AI engine setup
  - role library
- Rule:
  - advanced-gated
  - not part of first three-minute understanding

### 2. Memory

- Route: `/memory`
- Reason retained:
  - deeper memory inspection
  - semantic search
  - retained work exploration
- Rule:
  - advanced-gated
  - should support deeper proof, not default comprehension

### 3. System

- Route: `/system`
- Reason retained:
  - runtime health
  - storage checks
  - bus/service diagnostics
- Rule:
  - advanced-gated
  - used for reliability and recovery proof, not primary product framing

### 4. Advanced tabs inside default surfaces

- Automations:
  - `Shared Teams`
  - `Workflow Builder`
- Settings:
  - `AI Engines`
  - `Connected Tools`

### 5. Advanced chat capabilities

- Direct specialist routing
- Broadcast mode
- Managed exchange detail
- deep configuration/admin behavior

These remain important platform capabilities, but they should stay out of the default product story unless the operator intentionally reveals them.

---

## Deliverable 2: Feature Preservation Map

## Features that must not be reduced

- governed proposal and approval flow
- audit and activity visibility
- direct-answer Soma path
- specialist/council consultation
- AI engine profile control
- response-style control
- memory and continuity behavior
- pgvector-backed reusable recall
- tools, MCP, workspace file, and advanced resource access
- securable web/external research through governed connected-tool posture
- system diagnostics and advanced operator recovery

## Preservation rule by team

### Product Narrative Team

- may simplify wording
- may not delete capability meaning

### Default Experience Team

- may hide advanced surfaces from default view
- may not orphan advanced routes or break discoverability

### Capability Preservation Team

- must verify every simplification still leaves a reachable retained home

### Memory, Continuity, and Trust Team

- must ensure durable memory, temporary continuity, and trace remain conceptually distinct

## Explicit preservation decisions

- `Resources`, `Memory`, and `System` remain advanced, not removed
- direct council and broadcast remain advanced, not removed
- settings depth remains available through tabs and redirects, not removed
- docs remain the place for deep technical explanation, not removed
- advanced architecture/runtime truth stays documented, not surfaced as first-run clutter

---

## Deliverable 3: Golden-Path Partner Demo

## Goal

Show a highly technical partner or funder a product in under ten minutes that feels:

- understandable
- controllable
- differentiated
- technically credible

## Demo promise

The demo must prove:

1. Mycelis starts as an AI Organization, not a chatbot.
2. Soma gives a real answer immediately.
3. Risky work becomes a governed proposal instead of hidden action.
4. Execution results stay visible.
5. The organization carries continuity forward instead of acting statelessly.

## Canonical demo route

### Step 1: Landing

- Open `/`
- Narrate:
  - “This is an AI Organization product, not a generic agent shell.”
- Show:
  - `Create AI Organization`
  - product framing around governed execution and continuity

### Step 2: Create an organization

- Open `/dashboard`
- Create:
  - Name: `Northstar Labs`
  - Purpose: `Operate a governed AI product-delivery organization for planning, review, and execution.`
- Preferred starting point:
  - starter template

### Step 3: Show the organization home

- Open `/organizations/{id}`
- Narrate:
  - “Soma is the working counterpart, and the organization structure stays visible.”
- Show:
  - Soma
  - structure cards
  - recent activity
  - memory & continuity

### Step 4: Run a direct value prompt

- Ask Soma:
  - `Review this AI Organization and recommend the first operating priority.`
- Expected outcome:
  - direct answer
  - no unnecessary governance interruption

### Step 5: Show governed execution

- Ask Soma:
  - `Create a kickoff brief in the workspace called northstar_kickoff.md summarizing the first operating priority and next steps.`
- Expected outcome:
  - proposal card
  - visible approval posture
  - clear reason or risk framing

### Step 6: Approve and execute

- Approve the proposal
- Expected outcome:
  - execution result
  - created artifact/file proof
  - visible continuity in the same workspace

### Step 7: Show continuity

- Use one of:
  - recent activity panel
  - retained pattern / memory & continuity panel
  - return-to-organization continuity after refresh/re-entry
- Narrate:
  - “The system keeps continuity without forcing every conversation into long-term memory.”

### Step 8: Optional advanced reveal

- Only if useful for the audience:
  - open `Memory` in Advanced mode
  - show deeper retained knowledge or recall
  - or open `Resources` / `System` to show retained power

This step is optional because the primary goal is product legibility, not maximum technical density.

## Demo anti-patterns

- do not start in advanced mode
- do not open with tools, wiring, or system diagnostics
- do not lead with architecture docs
- do not demo five different agent-routing modes
- do not make approval feel like friction
- do not bury the actual output under explanation

---

## Deliverable 4: UI Testing Agentry Checklist

## Critical workflow checks

### A. Landing comprehension

- user can identify what Mycelis is from the landing page
- CTA hierarchy is obvious
- no default copy reads like swarm-console jargon

### B. AI Organization entry

- creating an AI Organization feels like the first obvious action
- partial API failure still leaves a legible path forward
- advanced setup is hidden without feeling missing

### C. Organization home

- Soma is visibly primary
- structure remains visible
- activity and continuity panels are readable
- no default copy suggests a generic assistant

### D. Direct-answer path

- non-mutating planning/review prompt returns direct answer
- user does not get shoved into unnecessary proposal flow

### E. Governed mutation path

- mutating request returns proposal
- approval language is intelligible
- confirm/cancel both behave clearly

### F. Execution proof

- result appears in the same workflow
- user can understand what happened without raw logs

### G. Continuity proof

- same-organization re-entry preserves useful continuity
- continuity language is distinct from long-term memory language

### H. Advanced power retention

- `Resources`, `Memory`, and `System` remain reachable in Advanced mode
- advanced tabs in `Automations` and `Settings` remain reachable
- direct specialist and advanced routing remain available when intentionally invoked

---

## Immediate Team Assignments

## Product Narrative Team

- review landing headline, CTA hierarchy, and first-screen copy against this brief
- identify remaining wording that still reads like infrastructure or internal architecture
- produce a concrete wording-drift list for:
  - `README.md`
  - landing page
  - organization creation
  - organization home
  - docs/help surfaces

## Default Experience Team

- review `/dashboard`, `/organizations/[id]`, `/automations`, and `/settings` against the default-surface rules
- flag anything still too complex for a partner demo
- produce a route-first acceptance table showing:
  - primary user action
  - expected visible outcome
  - what remains intentionally hidden until Advanced mode

## Capability Preservation Team

- produce a route-and-feature retention table proving advanced power remains reachable
- confirm no simplification work removes feature access
- verify retained homes for:
  - `Resources`
  - `Memory`
  - `System`
  - advanced `Automations` tabs
  - advanced `Settings` tabs
  - specialist/council and deep runtime capability surfaces

## Memory, Continuity, and Trust Team

- review every default mention of memory, continuity, approval, and retained knowledge for consistency
- identify every remaining place where:
  - `Learning` should become `Memory & Continuity`
  - durable memory and temporary continuity are still conceptually blurred
  - approval language feels technical instead of trustworthy

## Demo Scenario Team

- dry-run the golden path
- refine the prompt wording for the most reliable result
- produce one fallback path if the governed mutation example is noisy
- capture a demo operator sheet with:
  - opening narration
  - canonical prompts
  - expected visible states
  - fallback branch

## UI Testing Agentry Team

- convert the checklist above into a browser evidence pass
- report failures in terms of product trust, not just test mismatch
- treat these as the first demo-critical workflow lanes:
  - landing comprehension
  - AI Organization creation
  - direct-answer Soma value
  - governed proposal and approval
  - continuity visibility after refresh/re-entry

## Release and Ops Team

- prove the demo flow from a clean startup path
- document any environment prerequisites required for the partner demo
- keep the demo environment checklist focused on:
  - fresh startup
  - service health proof
  - recovery path if one surface degrades

---

## Cross-Team Dependency Order

1. Product Narrative Team defines the wording guardrails.
2. Default Experience Team confirms which default routes and panels carry that wording.
3. Capability Preservation Team verifies what stays advanced and where it remains reachable.
4. Memory, Continuity, and Trust Team normalizes continuity/approval language so demo and docs do not drift.
5. Demo Scenario Team locks the one best product story.
6. UI Testing Agentry Team verifies the story in the browser.
7. Release and Ops Team proves the same story survives clean startup and live proof.

Dependency rule:

- no team may declare success on its lane if it created drift for a later lane
- route/copy changes must be re-shared with QA, demo, and docs before acceptance

---

## Immediate Acceptance Outputs

The Product Manager expects these concrete outputs next:

1. a wording-drift inventory for default product surfaces
2. a route-and-feature retention table for advanced power
3. a canonical partner demo script with fallback
4. a browser verification checklist mapped to that demo
5. a clean environment proof checklist for the same demo

---

## PM Acceptance Rule

This execution brief is successful only if the teams can honestly say:

- the default Mycelis story is obvious
- advanced power was preserved
- the demo is believable
- the workflows were tested as product workflows, not just technical paths

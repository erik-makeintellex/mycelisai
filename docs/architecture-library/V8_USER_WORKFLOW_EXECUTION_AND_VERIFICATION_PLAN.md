# V8 User Workflow Execution and Verification Plan

> Status: ACTIVE
> Last Updated: 2026-04-03
> Owner: Product Management / Delivery Coordination
> Purpose: Coordinate the dev team around real user workflows so Mycelis proves user-shaped delivery, governed engine management, MCP usability, and workflow-complete product outcomes rather than only immediate interaction mechanics.

## Why This Plan Exists

Mycelis is not meant to dictate one house operating model.

Users need to be able to:

- define teams the way they want delivery to work
- shape delivery lanes around their own operating style
- associate MCP capabilities when they satisfy the API and framework contracts
- request and generate content without the framework over-governing ordinary intent

Mycelis should act as a governed lattice for execution, coordination, and trust.

Mycelis should not act like a product that decides how users must structure delivery or what kinds of content utilization are acceptable by default.

This plan exists to keep implementation, UX, runtime, docs, and testing aligned to that posture.

## Core Product Rule

Mycelis is a webbing lattice framework for governed execution.

That means:

- the framework protects engine behavior, mutation risk, external exposure, and durable execution boundaries
- the framework does not over-author normal user intent, team shape, content requests, or standards-compliant MCP utilization
- users define delivery structure
- governance protects execution integrity, not creative or operational expression by default

## Non-Negotiable Workflow Rules

1. Do not assume one canonical team structure for users.
2. Do not imply that Mycelis knows the “right” way a user should organize delivery.
3. Do not block content creation or collaboration just because it involves specialists, models, or MCP, unless policy/risk rules require it.
4. Do not treat MCP association as suspicious by default when it passes API and framework rules.
5. Do not test only that a UI interaction fired; test that the user’s actual workflow outcome remained legible, trustworthy, and complete.
6. Do not flatten advanced capability; keep it reachable as framework depth rather than default-path clutter.

## Workflow Principles

### 1. User-Shaped Delivery

Users must be able to:

- create organizations that reflect their own delivery structure
- use Soma to shape teams, lanes, and working patterns without being forced into preset methodology
- move between guided default workflows and deeper structure when they want it

Success condition:

- Mycelis feels like it supports the user’s operating model instead of replacing it.

### 2. Engine Governance, Not Intent Governance

The governance layer should primarily protect:

- mutating execution
- durable side effects
- risky external operations
- capability, cost, and data-boundary posture

The governance layer should not primarily police:

- ordinary drafting
- ideation
- low-risk collaboration
- standards-compliant content generation
- standards-compliant MCP usage

Success condition:

- users feel protected from risky execution, not controlled in how they think or work.

### 3. MCP Association by Contract

If an MCP association:

- satisfies API rules
- satisfies framework compatibility rules
- stays within current governance policy

then Mycelis should use it as an intention-usability pass rather than introducing extra paternal friction.

Success condition:

- MCP-backed work feels usable and legible, not preemptively obstructed.

### 4. Workflow-Complete Proof

Every major test lane must prove:

- entry
- action
- visible result
- continuity after refresh/re-entry
- trust behavior when mutation or external risk is involved

Success condition:

- we prove complete operator workflows, not isolated clicks.

## Team Structure

### 1. Product Management / Delivery Coordination

Responsibilities:

- own workflow priority order
- reject “interaction-only” proof that does not verify the real workflow outcome
- keep the user-shaped delivery rule stable across docs, runtime, and UI

Required outputs:

- canonical workflow order
- workflow acceptance gates
- blocker classification by `product`, `runtime`, `environment`, `test`, or `docs`

### 2. User Workflow and Interaction Team

Responsibilities:

- implement guided default workflows without over-prescribing team structure
- preserve the user’s freedom to shape delivery lanes
- keep route flows centered on visible result, not only successful interaction

Required outputs:

- dashboard and organization workflow implementation
- guided team-shaping surfaces that remain user-directed
- route/browser proof that validates real workflow completion

### 3. Runtime and Governance Boundary Team

Responsibilities:

- protect engine integrity, mutation safety, and external-boundary trust
- avoid turning governance into blanket intent control
- keep policy-bound approval scoped to real execution risk

Required outputs:

- runtime policy behavior that distinguishes content intent from risky execution
- bounded approval posture for mutation/external/cost-sensitive paths
- no regression toward universal approval theater

### 4. Team Definition and Delivery Modeling Team

Responsibilities:

- keep organization/team modeling flexible enough for user-shaped delivery
- prevent UI or templates from implying one “correct” delivery structure
- ensure Soma team design supports user intent instead of imposing product dogma

Required outputs:

- team-design workflow expectations
- verification of user-defined team/lane flexibility
- docs that describe Mycelis as a framework for delivery structure, not a replacement for it

### 5. MCP and Integration Usability Team

Responsibilities:

- keep MCP association and toolset behavior aligned to framework/API rules
- avoid extra friction once an integration satisfies the supported contract
- preserve security and policy boundaries without making the product feel hostile to capability use

Required outputs:

- workflow expectations for MCP setup and association
- tests that distinguish valid framework-governed use from actually risky or invalid use
- user-facing wording that explains capability posture without paternalism

### 6. Content and Collaboration Team

Responsibilities:

- keep content requests visibly useful
- keep collaboration readable without over-framing or over-controlling the user
- ensure generated output returns as inline value or clear artifact/result

Required outputs:

- workflow-complete content and artifact proof
- specialist/model collaboration posture that feels helpful, not bureaucratic
- alignment with the ask-class/output contract lane

### 7. QA / Workflow Verification Team

Responsibilities:

- test end-to-end user workflows, not just immediate interaction
- require continuity/re-entry proof where workflow meaning depends on it
- classify whether failures are workflow gaps or merely harness drift

Required outputs:

- workflow-by-workflow evidence
- stricter outcome assertions where tests currently stop too early
- final workflow verdicts by lane

### 8. Docs and Trust Language Team

Responsibilities:

- keep wording neutral and empowering
- avoid language that overstates product control over the user’s methods
- keep docs aligned to the framework posture

Required outputs:

- wording review for workflow docs, test docs, and visible UI copy
- no “house style” drift in default-path product language

## Execution Order

### Wave 1: Workflow Inventory and Coverage Truth Map

Goal:

- inventory every primary user workflow
- map each workflow to current tests and docs
- identify where current tests prove only immediate interaction

Required outputs:

- canonical workflow list
- coverage matrix
- gap list classified as `product`, `runtime`, `test`, or `docs`

### Wave 2: Workflow-Complete Test Tightening

Goal:

- upgrade tests so they verify workflow result, continuity, and trust outcome

Required outputs:

- strengthened browser proof for current default workflows
- improved unit/component assertions where UI summaries carry workflow meaning
- no major user workflow that stops at “button clicked” proof

### Wave 3: User-Shaped Team and MCP Validation

Goal:

- verify that team definition and MCP association behave like flexible framework features rather than locked product opinions

Required outputs:

- workflow proof for team-shaping paths
- workflow proof for MCP association posture
- explicit confirmation that standards-compliant use is not over-gated

### Wave 4: Trust and Release Verification

Goal:

- prove that the full workflow set reads as trustworthy and product-legible

Required outputs:

- stable mocked browser proof
- live governed browser proof
- manual trust pass
- workflow verdict with strict next actions

## Canonical User Workflow Set

### Workflow A: Enter through Central Soma

User goal:

- understand the product and get into real work quickly

The workflow is complete only when:

- the user understands Soma is persistent across contexts
- organization entry or return is obvious
- the next meaningful action is visible

### Workflow B: Create or Reopen an AI Organization

User goal:

- enter a governed working context without friction or jargon

The workflow is complete only when:

- creation or re-entry succeeds
- the user lands in the correct organization
- the organization context remains clear

### Workflow C: Shape Delivery with User-Defined Teams

User goal:

- use Soma to define delivery structure in the user’s own terms

The workflow is complete only when:

- the team-design path is reachable
- the user can express their own delivery intent
- the result reads as support for the user’s model, not imposition of ours

### Workflow D: Ask for Direct Content or Drafting

User goal:

- get useful inline value without unnecessary governance friction

The workflow is complete only when:

- the result lands in `answer`
- the answer is visibly useful
- no mutation-first or approval-first drift appears for ordinary content asks

### Workflow E: Ask for Artifact or Durable Output

User goal:

- receive durable output that is clearly understandable

The workflow is complete only when:

- the output posture matches the ask
- the user sees clear returned-output framing
- artifacts are visible and referenced clearly

### Workflow F: Ask for Specialist or Collaborative Support

User goal:

- benefit from specialist/model/cross-agent help without unnecessary ceremony

The workflow is complete only when:

- specialist involvement is visible
- the answer still reads as one coherent outcome
- collaboration does not feel over-governed when policy does not require it

### Workflow G: Run a Governed Mutation

User goal:

- make a real change with clear trust boundaries

The workflow is complete only when:

- mutation lands in `proposal`
- approval posture is understandable
- confirm/cancel behavior is explicit
- durable proof exists after execution

### Workflow H: Associate MCP Capability

User goal:

- attach usable capability that passes framework and API rules

The workflow is complete only when:

- the association path is understandable
- standards-compliant association is not over-blocked
- risk boundaries remain intact where policy requires them

### Workflow I: Re-enter After Refresh, Navigation, or Interruption

User goal:

- keep confidence in the current working context

The workflow is complete only when:

- context survives refresh/re-entry
- the last meaningful result remains legible
- the user does not have to reconstruct the workflow from scratch

### Workflow J: Manage organization permissions without forcing enterprise IAM into the base release

User goal:

- keep organization access clear in the default product while still enabling enterprise user management where the deployment wants it

The workflow is complete only when:

- base release keeps People & Access centered on organization roles and collaboration groups
- enterprise user-directory management stays layered instead of becoming the default release surface
- owner access can manage the enterprise directory when that layer is enabled
- non-owner enterprise roles do not silently inherit user-directory mutation control

### Workflow K: Instantiate a native Mycelis team for target output

User goal:

- ask for a meaningful output and have Mycelis manifest a bounded delivery structure to produce it

The workflow is complete only when:

- the product makes it clear that Mycelis is instantiating a native managed team rather than only calling a tool
- the user can understand the target output the team is responsible for
- the result comes back with visible lineage to the managed team
- at least one bounded target-output proof exists for the release, preferably image generation

### Workflow L: Instantiate an external workflow contract without blurring it into a native team

User goal:

- target an external workflow system such as `n8n` or a comparable service without losing governance, readability, or result normalization

The workflow is complete only when:

- the product makes it clear that the target is an external workflow contract, not a native Mycelis team
- governance and capability posture remain visible
- returned result or artifact is normalized back into Mycelis cleanly
- the release contract keeps this path separate from native team manifestation

## Verification Standard

A test is not sufficient if it proves only:

- a control rendered
- a click handler fired
- a route changed
- a terminal badge appeared without a meaningful workflow result

A workflow test is sufficient only when it proves:

- the user could start the workflow
- the workflow returned the expected value/trust posture
- the result remained understandable after the next realistic step

## Immediate Execution Stack

### Slice 1: Canonical workflow inventory and ownership alignment

Files expected to change:

- `docs/architecture-library/V8_USER_WORKFLOW_EXECUTION_AND_VERIFICATION_PLAN.md`
- `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
- `docs/architecture-library/V8_DEV_TEAM_EXECUTION_AND_UI_ENGAGEMENT_PLAN.md`
- `V8_DEV_STATE.md`
- `interface/lib/docsManifest.ts`

Minimum proof:

- `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`
- `cd interface; npx tsc --noEmit`

### Slice 2: Workflow-complete browser proof tightening

Focus:

- identify the highest-signal workflows where current browser tests stop at immediate interaction
- tighten those tests to assert real user outcome and continuity

Minimum proof:

- targeted `interface.e2e` specs for touched workflows
- any matching component tests where summary/result framing changes

### Slice 3: Team-shaping and MCP usability verification

Focus:

- verify that team definition and MCP association stay user-directed and framework-governed rather than over-controlled

Minimum proof:

- targeted browser coverage
- docs/state sync

## Acceptance

This plan is successful when:

- the team executes from user workflows rather than narrow interaction fragments
- tests prove real workflow outcomes and continuity
- Mycelis reads as a governed framework for user-defined delivery
- engine protection remains strong without over-governing content, collaboration, or standards-compliant MCP use

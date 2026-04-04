# V8 User Intent Workflow Review and Instantiation Plan

> Status: ACTIVE
> Last Updated: 2026-04-03
> Owner: Product Management / Delivery Coordination
> Purpose: Review the full target architecture against the bounded release and define one canonical workflow-review and instantiation contract for all major user-intent classes, including internal team instantiation and external workflow-contract instantiation.

## Why This Plan Exists

Mycelis already has the architectural pieces for:

- direct Soma answers
- specialist consultation
- governed mutation
- governed artifact delivery
- team design and manifestation
- image/content generation
- external capability attachment through MCP and future workflow services

What is still missing is one release-grade contract that answers:

- which kind of user intent is being handled
- who should own execution for that intent
- whether the result should stay inline, become an artifact, instantiate a team, or instantiate an external workflow contract
- how the product proves the result without collapsing different execution modes into one blurry “automation” story

This plan exists to close that gap.

## Architecture Review Summary

### Current bounded release truth

The current V8.1 release architecture already supports:

- Soma-primary intent intake
- ask-class and output-contract classification
- policy-bounded governed mutation
- governed artifact return
- bounded team/mission manifestation through `intent/negotiate` and `intent/commit`
- specialist and media/image capability under bounded governance
- curated MCP-first integration posture

### Full target truth

The V8.2 target extends that into:

- editable automations
- distributed execution
- stronger capability-backed actuation
- richer workflow composition
- external contract-driven execution surfaces

### Release interpretation

Initial release must not pretend the full V8.2 automation platform is already shipped.

Initial release must, however, make these boundaries explicit and testable:

1. direct intent handling
2. internal managed team instantiation
3. external workflow-contract instantiation

## Core Product Rule

Mycelis is a governed lattice for execution.

That means:

- user intent stays primary
- the framework governs execution integrity, mutation risk, external boundaries, and result lineage
- Mycelis should not flatten all intent into chat
- Mycelis should not flatten all execution into one generic automation bucket

## Canonical User-Intent Classes

### 1. Direct answer

Use when:

- the user wants explanation, drafting, planning, review, comparison, or other inline value

Execution owner:

- Soma directly, with optional bounded specialist help

Expected result:

- `answer`

### 2. Specialist consultation

Use when:

- the ask benefits from a visible specialist contribution but does not require new durable execution structure

Execution owner:

- Soma plus one or more specialists

Expected result:

- `answer`

### 3. Governed mutation

Use when:

- the ask changes files, systems, durable runtime state, permissions, or other governed resources

Execution owner:

- Soma under governed execution

Expected result:

- `proposal` then `execution_result`

### 4. Governed artifact

Use when:

- the user wants a durable file, saved media, export, or packaged deliverable

Execution owner:

- Soma, specialist path, or bounded generation path

Expected result:

- `answer` with artifact cue or `proposal` when policy/risk requires it

### 5. Team-instantiated managed output

Use when:

- the user wants Mycelis to manifest a bounded delivery structure to produce a target output
- the work benefits from Team Lead plus specialist coordination rather than one-turn direct handling

Release example:

- instantiate a creative delivery team to generate an image asset and return it as a visible artifact/result

Execution owner:

- Soma manifests or activates an internal managed team through the organization runtime

Expected result:

- a visible team-instantiation contract
- a readable execution/result path
- returned target output with lineage to the managed team

### 6. External workflow-contract instantiation

Use when:

- the user wants a durable external workflow or external execution graph to be invoked or instantiated through a supported contract
- the target execution logically belongs to an external workflow system such as `n8n`, ComfyUI, or a comparable service

Execution owner:

- Soma binds intent to an external workflow contract instead of pretending the external graph is a native Mycelis team

Expected result:

- a clear contract surface showing the external target
- explicit governance and capability posture
- normalized result or returned artifact back into Mycelis

## Separation Contract

### Internal team instantiation

Internal team instantiation means:

- Mycelis manifests or activates its own managed team structure
- Team Lead, specialists, and result lineage stay inside the Mycelis organization model
- the user can inspect the resulting team/work/result relationship as native Mycelis execution

Release-proof example:

- a bounded image-generation team is instantiated from user intent
- the team produces an image artifact
- the image returns through Soma with readable output framing

### External workflow-contract instantiation

External workflow-contract instantiation means:

- Mycelis binds the user’s intent to a supported external contract
- the external service owns the downstream graph or workflow definition
- Mycelis owns policy review, invocation posture, result normalization, and operator legibility

Release-proof example:

- the operator chooses or targets an `n8n`-style workflow contract
- Mycelis shows the target workflow contract clearly
- the invocation/result path is normalized back into Mycelis without pretending that `n8n` itself is a Team Lead or Department

### Non-negotiable boundary

Do not blur:

- native Mycelis teams
- external workflow services
- single-turn content generation
- long-running external workflow orchestration

## Workflow Review Steps For Every Intent

Every major user-intent path must be reviewed through the same sequence:

1. classify the ask into one of the canonical intent classes
2. decide the correct execution owner:
   - Soma direct
   - Soma plus specialist
   - governed mutation
   - managed team instantiation
   - external workflow-contract instantiation
3. decide the visible terminal posture:
   - `answer`
   - `proposal`
   - `execution_result`
   - `blocker`
4. decide whether the result must stay inline, return as artifact, instantiate a team, or instantiate an external workflow contract
5. make the default-visible explanation clear enough that the operator understands:
   - what is happening
   - who is doing it
   - why approval is or is not required
   - where the output will appear
6. return visible inline value or a clearly referenced output/result
7. preserve continuity and auditability after refresh, re-entry, or later inspection

## Team Structure

### 1. Product Management / Delivery Coordination

Responsibilities:

- keep the intent-class matrix canonical
- reject execution paths that blur internal team instantiation with external contract invocation
- keep initial-release scope honest against V8.1 and V8.2

### 2. Intent Classification and Runtime Contract Team

Responsibilities:

- map intent classes to runtime execution ownership
- keep ask-class, agent-type, and output posture aligned
- prevent team/workflow instantiation from drifting into generic chat or generic approval behavior

### 3. Team Instantiation and Delivery Modeling Team

Responsibilities:

- define the release-proof path for native Mycelis team instantiation
- ensure a user can ask for target output and have Mycelis manifest a bounded delivery team to produce it
- keep team-instantiated output legible in UI, audit, and result surfaces

### 4. External Workflow Contract Team

Responsibilities:

- define the boundary for `n8n`-style and similar external workflow systems
- make the external contract explicit and inspectable
- keep result normalization clear without pretending the external workflow is a native Mycelis team

### 5. Content and Media Proof Team

Responsibilities:

- provide one image-generation proof path that exercises target-output delivery
- keep inline/media/artifact return understandable

### 6. QA / Workflow Verification Team

Responsibilities:

- add workflow review and browser-proof steps for:
  - team-instantiated output
  - external workflow-contract instantiation
- require visible result proof, not only invocation proof

## Initial Release Deliverables

### REQUIRED: Canonical workflow-review matrix

Must exist in docs and testing guidance for all intent classes above.

### REQUIRED: Native team-instantiation proof path

Initial release must prove at least one bounded target-output scenario where:

- the user asks for a target result
- Mycelis instantiates or activates a managed team
- the team returns a visible result
- the result remains legible as native Mycelis execution

Preferred proof:

- image-generation team for a bounded creative request

### REQUIRED: External workflow-contract separation

Initial release must define the external workflow-contract path explicitly so:

- `n8n` and similar services are treated as external contract targets
- Mycelis owns governance, invocation posture, and normalization
- the user can understand the difference from internal team instantiation

### NEXT: External workflow-contract runnable proof

Initial release should target one bounded runnable proof where a supported external workflow contract is invoked or proposed through a normalized surface.

## Acceptance Criteria

This lane is successful only when:

- every major user-intent class has workflow-review steps
- internal team instantiation is proven as distinct from direct answer and direct artifact generation
- external workflow-contract instantiation is proven as distinct from native Mycelis team manifestation
- at least one team-instantiated target-output workflow is captured as release proof
- testing docs require visible result proof, not only request submission or proposal appearance

# V8 Universal Soma and Context Model PRD

> Status: ACTIVE
> Last Updated: 2026-03-29
> Owner: Product Management / Delivery Coordination
> Purpose: Define the canonical product target where Soma and Council are universal operating entities, while AI Organizations, deployments, teams, and projects remain governed working contexts.
> Depends On: `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`, `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/user/memory.md`, `docs/user/resources.md`

## 1. Why this PRD exists

Mycelis has reached a point where the bounded V8.1 operator flow is useful, but it still teaches the wrong top-level product model in one important way:

- the current UI makes Soma feel organization-scoped
- the intended product is that Soma and Council are universal entities
- AI Organizations, deployments, teams, logs, sources, and runs are scoped contexts Soma can see, compare, and act within

This PRD corrects that model.

It does not replace the current bounded release contract overnight. It defines the next canonical product target and the delivery sequence required to reach it without removing advanced platform depth.

## 2. Canonical correction

### 2.1 Universal entities

- **Soma** is the persistent operator-facing counterpart across the platform.
- **Council** is the persistent advisory and specialist layer behind Soma.
- Soma and Council are not recreated per project, per AI Organization, or per deployment.

### 2.2 Scoped contexts

- **AI Organizations** are durable operating contexts Soma can enter, resume, compare, and govern.
- **Projects**, **deployments**, **teams**, **runs**, **logs**, **memory records**, and **sources** are scoped objects within or across those contexts.
- Execution remains explicitly scoped even when Soma is universal.

### 2.3 Product rule

The product must not imply "one Soma per organization."

The product must instead imply:

```text
One Soma
  -> many organizations
  -> many deployments
  -> many teams
  -> many runs, logs, memories, and sources
  -> one governed cross-context operating system
```

## 3. Design principles

### 3.1 Keep the platform, simplify the entry

Mycelis must become easier to understand without amputating what makes it valuable.

Required:

- simple default path
- universal Soma home
- explicit context entry and switching
- preserved advanced power
- preserved governed execution
- preserved memory and audit depth

Disallowed:

- flattening the product into generic chat
- removing advanced routes just to improve demos
- hiding structure so deeply that the platform feels fake

### 3.2 Scope execution, not identity

Soma identity is global.
Execution scope is local and governed.

That means Soma may:

- inspect multiple organizations
- compare planning state across contexts
- read logs and continuity signals across contexts
- route work toward the correct team or deployment

But mutating execution must still declare:

- target organization
- target team or deployment when relevant
- approval posture
- capability risk
- resulting artifacts and audit lineage

### 3.3 Deliver value before mechanics

Soma should first deliver:

- answers
- plans
- summaries
- comparisons
- content

Then use:

- proposals
- approvals
- artifact generation
- specialist/model collaboration

when the requested outcome or policy requires it.

## 4. Product target

### 4.1 Central Soma home

The top-level product entry should evolve toward a universal Soma home.

Central Soma home must provide:

- one primary conversational entry point
- recent organizations and deployments
- recent activity and approvals
- continuity summary
- creation path for new AI Organizations
- easy entry into existing contexts

### 4.2 Context switching

Operators must be able to:

- stay with one persistent Soma
- switch between organizations without feeling like they entered a different assistant
- understand which context is currently active
- ask cross-context questions without losing clarity

Required context indicators:

- current organization
- current deployment or workspace when relevant
- whether the prompt is global or scoped
- whether the next action will mutate a scoped context

### 4.3 Council visibility

Council remains universal specialist depth behind Soma.

Default posture:

- Council does not replace Soma as the front door
- Council does not appear as a flat multi-bot chooser
- Council collaboration is visible when it materially improves trust or understanding

Advanced posture:

- operators can inspect which specialists contributed
- operators can inspect why a specialist/model/tool path was chosen
- specialist collaboration remains policy-aware and auditable

### 4.4 Memory, logs, and continuity

Universal Soma should be able to use:

- durable scoped memory
- temporary planning continuity
- trace and audit records
- execution logs
- deployment and service health signals

But those stores must remain distinct.

Rules:

- not all chat becomes durable memory
- trace/audit does not automatically become semantic memory
- scoped team knowledge must stay scoped unless intentionally promoted
- cross-context reasoning must respect visibility boundaries

## 5. Governance and collaboration model

### 5.1 Approval posture

Do not treat all model-to-model, MCP, or cross-agent collaboration as approval-gated by default.

Approval must remain configurable based on:

- capability risk
- cost
- external exposure
- mutation impact
- integration trust level
- user or organization governance profile

### 5.2 Content and artifact rule

When a user asks for content, Soma must produce either:

- visible inline content
- or a clearly referenced generated artifact

It is not sufficient to only show tool intent.

### 5.3 Universal visibility with bounded mutation

Universal Soma may observe broadly.
Mutating execution must remain explicit and scoped.

That means the operator should always be able to answer:

- what context Soma is acting in
- what tool/model/capability is being used
- whether approval is required and why
- what artifact or result was produced

## 6. Delivery model

### Phase 1: Contract correction

Deliverables:

- canonical PRD for universal Soma and scoped contexts
- matching state/index/docs-manifest sync
- explicit note that current organization-scoped Soma UI is transitional

### Phase 2: Broken operator contract repair

Deliverables:

- working settings contract end to end
- current UI/testing/docs no longer teaching false settings behavior
- immediate live blockers on creation/settings closed

### Phase 3: Central Soma introduction

Deliverables:

- dashboard evolves from pure organization entry into Central Soma home
- existing AI Organization creation remains reachable
- current organization workspace becomes a scoped context entered through Soma

### Phase 4: Cross-context trust and visibility

Deliverables:

- global recent activity and approval surfaces
- cross-context continuity summary
- clear scope labels for answer vs proposal vs execution

### Phase 5: Advanced reveal without platform loss

Deliverables:

- preserved advanced routes for memory, resources, system, and deeper structure
- council/specialist contribution inspection
- scoped logs and memory inspection that remain coherent with the default product story

## 7. Acceptance criteria

This PRD is materially delivered only when all of the following are true:

- operators experience Soma as one persistent counterpart across contexts
- AI Organizations are clearly presented as work contexts, not separate Soma identities
- governed execution remains scoped and auditable
- content and artifact delivery feel direct and valuable
- settings and continuity surfaces actually persist
- advanced platform depth remains reachable without polluting the default story

## 8. Immediate next actuation

The first code slice under this PRD should:

1. fix the user-settings contract so settings really persist and reload
2. update active docs/state surfaces so they stop silently teaching an org-scoped Soma identity as the long-term model
3. prepare the dashboard/home evolution toward a Central Soma surface instead of a pure organization-launch screen

## 9. Compatibility note

The current bounded V8.1 operator contract remains valid for the release candidate:

- AI Organization creation
- organization-scoped workspace entry
- Soma-primary interaction within an organization

But that bounded contract should now be treated as a transitional release surface, not the final product identity.

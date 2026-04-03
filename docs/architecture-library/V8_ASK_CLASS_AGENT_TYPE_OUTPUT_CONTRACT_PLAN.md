# V8 Ask-Class, Agent-Type, and Output Contract Plan

> Status: ACTIVE
> Last Updated: 2026-04-02
> Owner: Product Management / Delivery Coordination
> Purpose: Establish one canonical machine-readable contract for mapping ask classes to agent types, routing posture, and output contracts so Mycelis stops relying on fragmented prompt, template, and UI assumptions.

## Why This Lane Exists

Today the repo has partial enforcement, not one authoritative assertion layer.

Current truth is split across:

- architecture docs for Agent Type Profiles and Response Contract inheritance
- provider-role routing
- standing-team prompt/template routing rules
- chat runtime answer-vs-proposal mode selection
- UI/store terminal-state handling

That is enough to work, but not enough to guarantee that the right ask always routes to the right agent posture and returns the right class of output through one shared contract.

## Target Outcome

This lane is successful when all of the following are true:

- every operator ask class resolves through one machine-readable contract
- the contract names the default and allowed agent types for that ask
- the contract names the default and allowed output contracts for that ask
- Soma, direct council chat, approval/proposal UX, and browser proof all use the same contract
- advanced depth remains available without letting routing/output meaning drift between runtime, docs, and UI

## Current Local Repo Truth

### What Is Already Enforced

1. Agent Type Profiles and Response Contract inheritance are already canonical architecture concepts.
2. Provider routing already resolves by role and scope.
3. Standing-team prompt/template routing already says which specialist should handle code, architecture, creative, and sentry asks.
4. Chat runtime already enforces `answer` vs `proposal` mode based on governed mutation detection.
5. UI/store layers already enforce the final visible terminal-state model:
   - `answer`
   - `proposal`
   - `execution_result`
   - `blocker`

### What Is Missing

The repo does not yet have one central machine-readable registry that says:

- what ask class a request belongs to
- which agent types are default or allowed for that ask
- which output contract is expected
- whether the result should be inline content, a governed artifact, an execution result, or a blocker
- whether approval is auto-allowed, optional, or required for that ask class

That missing registry is the highest-value gap.

## Non-Negotiable Rules

1. Do not flatten the system into a small intent router with no specialist depth.
2. Do not replace policy-bound approval logic with prompt-only conventions.
3. Do not let the UI invent ask classes that runtime does not understand.
4. Do not let runtime emit output modes that UI/browser proof does not explicitly cover.
5. Do not make direct content asks feel like tool activity without visible value return.

## Team Structure

### Product Management / Delivery Coordination

Responsibilities:

- own the phase order
- keep contract naming stable
- prevent drift between docs, runtime, UI, and tests

Required outputs:

- canonical ask taxonomy
- canonical slice order
- blocker classification and checkpoint packaging

### Runtime and Governance Team

Responsibilities:

- implement the shared ask-contract registry
- resolve ask classes into agent/routing/output posture
- keep approval and mutation integrity unchanged while consolidating the contract

Required outputs:

- machine-readable ask contract in Go
- runtime helper for contract resolution
- proof that governed mutation remains proposal-first

### Interface and Workflow Team

Responsibilities:

- consume the shared contract where operator-visible behavior depends on it
- keep terminal-state presentation aligned with runtime meaning
- avoid UI-only route/output assumptions

Required outputs:

- UI/store alignment to the shared contract
- no contradictory route/output wording in default surfaces

### Product Narrative and Trust Team

Responsibilities:

- keep ask-class and output-class language operator-legible
- align answer/proposal/artifact/result wording with the shared contract

Required outputs:

- stable operator-facing wording for contract-driven output states

### QA / UI Testing Agentry Team

Responsibilities:

- prove the contract at runtime and UI levels
- verify the right ask class reaches the right terminal state
- classify failures as `product`, `runtime`, `environment`, `test`, or `docs`

Required outputs:

- focused runtime tests
- focused store/component tests
- browser proof for contract-sensitive default paths

## Execution Order

### Phase 0: Truth Mapping

Status target: COMPLETE

Goals:

- freeze the initial ask taxonomy and output taxonomy from local repo truth
- identify every current enforcement seam already in code
- identify the single first runtime insertion point for the shared contract

Required outputs:

- canonical ask classes for V8.1 / MVP
- canonical output classes for V8.1 / MVP
- named first runtime insertion point

### Phase 1: Shared Runtime Contract

Status target: ACTIVE -> IN_REVIEW

Goals:

- introduce one central machine-readable registry for ask classes
- bind each ask class to:
  - default agent type or target posture
  - allowed agent types
  - default output contract
  - allowed output contracts
  - approval posture
  - artifact durability expectation when relevant

Safest first implementation slice:

- add a shared contract type and starter registry in Go
- cover only the currently proven bounded classes first:
  - direct answer
  - governed mutation
  - governed artifact
  - blocker
- make Soma chat resolution read from that registry before final mode selection
- keep council chat on the same contract for shared output meaning, without redesigning broader multi-agent orchestration yet

Current checkpoint:

- the shared Go registry now exists for `direct_answer` and `governed_mutation`
- primary Soma chat and direct council chat now resolve template/mode selection through that shared contract instead of duplicating those defaults inline
- audit context now records the resolved `ask_class` on those bounded answer/proposal paths
- focused protocol/runtime tests now prove the registry contract and the existing proposal-integrity behavior remains green
- the shared registry now also covers `governed_artifact` and `specialist_consultation`
- chat payloads now carry `ask_class` through the runtime/store contract so UI surfaces no longer have to infer artifact-vs-specialist answers indirectly
- the first default-visible UI proof is now in place: Mission Control chat labels artifact-bearing answers as `Artifact result` and consultation-shaped answers as `Specialist support`
- focused proof for this slice is green across protocol/runtime/store/component gates plus the broader managed `core.test` and `interface.test` runs
- the next browser-proof slice now asserts those cues in the stable mocked spec and asserts live `ask_class` payload values for direct-answer and governed-mutation paths in the compose-backed governed browser spec
- stable and live browser proof for that browser slice are now green after rebuilding the compose runtime against the current Core image

Immediate next slice:

- carry the same `ask_class` assertion into broader workspace/browser proof beyond the governed-chat lane so more default-path flows verify both the handler route and the returned output class
- decide whether `artifact_reference` should remain payload-only or also gain a stronger operator-facing summary contract in the store/UI layer

### Phase 2: UI and Store Alignment

Status target: NEXT

Goals:

- remove any UI-only assumptions about routing or terminal state meaning
- keep store and chat surfaces aligned to the shared contract
- ensure guided default paths still read simply

Acceptance:

- Mission Control chat, organization workspace chat, and proposal/result surfaces stay aligned with runtime contract meaning

### Phase 3: Test and Proof Consolidation

Status target: NEXT

Goals:

- add one focused runtime suite for ask-class resolution
- add one focused store/component suite for terminal-state expectations
- add one route/browser proof covering representative ask classes

Acceptance:

- the same ask-class examples are proven in runtime tests and UI/browser tests

## Initial Ask Taxonomy

This is the bounded initial machine-readable set for the first slice:

- `direct_answer`
- `governed_mutation`
- `governed_artifact`
- `specialist_consultation`
- `execution_blocker`

## Initial Output Contract Taxonomy

- `answer`
- `proposal`
- `execution_result`
- `blocker`
- `artifact_reference`

`artifact_reference` should be treated as a result shape layered inside the visible terminal states, not a replacement for the terminal-state model.

## First Slice Contract Shape

The first runtime registry should carry fields equivalent to:

- `ask_class`
- `default_agent_target`
- `allowed_agent_targets`
- `default_execution_mode`
- `allowed_execution_modes`
- `requires_confirmation`
- `approval_posture`
- `artifact_posture`
- `description`

## First Slice File Targets

Expected first code slice:

- `core/pkg/protocol/` or another shared Go contract package for the registry types
- `core/internal/server/cognitive.go` for Soma chat resolution
- `core/internal/server/cognitive_test.go` for runtime proof
- `interface/store/cortexStoreMissionChatSlice.ts` and related store helpers only if response shape changes
- focused UI/browser tests only where visible terminal-state behavior changes

## Proof Standard

Minimum proof for the first implementation slice:

- focused Go runtime tests for ask resolution
- existing proposal-integrity and answer-vs-proposal tests stay green
- focused store/component tests if the response contract changes
- browser proof for representative `answer`, `proposal`, and `blocker` behavior if default-path wording or routing presentation changes
- docs/state sync in the same slice

## Immediate Next Slice

Slice 2 for this lane is:

- extend the shared contract beyond `direct_answer` and `governed_mutation` into the next bounded ask classes with visible trust impact
- define how `specialist_consultation` and `governed_artifact` should be asserted without weakening the current terminal-state model
- add the first UI/store/browser proof that checks both `who handled it` and `what class of output came back`

This is the highest-value next move because the runtime foundation now exists and the biggest remaining gap is proof and adoption beyond mode-only validation.

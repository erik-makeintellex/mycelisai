# V8 Structured Team Asks and Lane Routing Plan

> Status: ACTIVE
> Last Updated: 2026-04-04
> Owner: Runtime / Coordination / Workflow
> Purpose: Introduce one canonical structured-ask contract for Soma-to-team delegation so team routing, lane roles, validation proof, and future instantiation behavior stop relying on free-form task strings.

## Why This Lane Exists

Mycelis already has:

- canonical team command and signal lanes
- governed `delegate_task`
- ask-class contracts for operator-facing chat
- runtime organization and team manifests

What it does not yet have is a typed contract for what Soma is asking a team to do.

Right now, team delegation is still too string-heavy. That makes it harder to:

- route asks by kind
- preserve ownership boundaries
- express validation expectations
- instantiate reusable team lane roles
- explain delegation cleanly in UI and audit surfaces

Important boundary:

- structured asks should constrain mission fit, owned scope, and proof expectations
- they should not over-constrain the engaged model into one rigid response format unless the mission explicitly requires that format

## Target Outcome

This lane is successful when:

- team-directed asks use one machine-readable contract
- the contract distinguishes ask kind from lane role
- Soma can express goal, scope, constraints, exit criteria, and evidence needs without inventing ad hoc payload shapes
- the command lane preserves structured asks end to end
- team instantiation can later bind stable default lane roles onto the same contract

## Non-Negotiable Rules

1. Do not break the existing team command bus or proposal/governance spine.
2. Do not replace governed execution with prompt-only delegation conventions.
3. Do not let UI invent team-ask semantics the runtime does not understand.
4. Do not force every ask into broad multi-team orchestration when a bounded team ask is enough.
5. Do not blur ordinary Soma memory, customer context, company knowledge, and future method learning.

## Canonical Contract

The initial structured team-ask contract should carry:

- `schema_version`
- `ask_kind`
- `lane_role`
- `goal`
- `owned_scope`
- `constraints`
- `required_capabilities`
- `approval_posture`
- `exit_criteria`
- `evidence_required`
- `context`

Bounded initial ask kinds:

- `general`
- `research`
- `implementation`
- `validation`
- `coordination`
- `review`

Bounded initial lane roles:

- `coordinator`
- `researcher`
- `implementer`
- `validator`
- `reviewer`

Rule:

- `ask_kind` describes what the ask is for
- `lane_role` describes the intended owner posture inside the team

They are related, but not identical.

## Team Structure

### Soma orchestration lane

Responsibilities:

- classify the ask
- choose direct answer vs specialist consultation vs team delegation
- emit a structured team ask when work crosses into team execution

### Runtime / protocol lane

Responsibilities:

- define the shared structured team-ask type
- normalize `delegate_task` onto that contract
- keep command-lane transport backward-compatible

### Team-instantiation lane

Responsibilities:

- bind reusable default lane roles into instantiated teams
- keep bundle/runtime organization truth separate from one-off delegation payloads

### UI / operator lane

Responsibilities:

- show legible delegation summaries
- later expose ask kind, lane role, and proof in inspectable detail without turning the default UX into a workflow console

### QA lane

Responsibilities:

- prove command-lane preservation
- prove delegate normalization
- prove no regression in existing governed delegation behavior

## Execution Order

### Phase 0: Truth mapping

Status target: COMPLETE

Result:

- the first safe insertion point is `delegate_task` normalization before publish to `swarm.team.{team_id}.internal.command`

### Phase 1: Structured ask contract

Status target: ACTIVE

Goals:

- add a shared protocol type for structured team asks
- normalize string/object `delegate_task` calls into that contract
- keep team command-lane transport backward-compatible

Required proof:

- protocol tests for normalization/defaults
- delegate normalization tests
- signal publish tests proving structured payload on the command lane
- team command normalization tests proving the payload survives into the internal trigger lane

### Phase 2: Instantiated lane-role defaults

Status target: NEXT

Goals:

- bind default lane-role templates onto instantiated team definitions
- make research-first and validation-aware routing part of reusable team instantiation rather than only prompt habit

### Phase 3: UI/operator exposure

Status target: ACTIVE

Goals:

- expose structured delegation summaries in inspectable UI surfaces
- keep default operator surfaces simple while allowing advanced delegation inspection

### Phase 4: Reflection and promotion

Status target: NEXT

Goals:

- capture recurrent team-ask patterns as reviewed method candidates
- keep project-context workflow patterns separate from portable team-method patterns

## Initial File Targets

First implementation slice:

- `core/pkg/protocol/`
- `core/internal/swarm/internal_tools_handlers_coordination.go`
- `core/internal/swarm/internal_tools_registration_coordination.go`
- `core/internal/swarm/internal_tools_delegate_test.go`
- `core/internal/swarm/internal_tools_signal_test.go`
- `core/internal/swarm/team_signals_test.go`

## Validation Standard

Minimum proof for Phase 1:

- `cd core && go test ./pkg/protocol ./internal/swarm -count=1`

No slice in this lane is complete unless:

- runtime contract and docs use the same vocabulary
- command-lane preservation is proven
- broader operator/UI changes are left out unless they are also explicitly proven

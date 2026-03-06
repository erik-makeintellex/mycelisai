# Standardization Review - 2026-03-06

## Scope

This review captures the current standardization posture across:

- architecture delivery
- operational workflows
- logging
- error handling
- development language boundaries

It is based on the active runtime and the latest delivery work around memory
restart hardening, gated execution, and multi-agent team instantiation.

## Findings

### 1. Standing team architecture is now durable, but coordination behavior is uneven

The live team topology currently includes:

- `prime-architect`
- `prime-development`
- `agui-design-architect`

These teams are now source-controlled and manifest-backed under
`core/config/teams`, which is the correct durability model.

The remaining standardization gap is behavioral: architect and AGUI now answer
central sync requests more reliably than development, so the transport and
manifest model are correct but reply discipline is still incomplete.

Required standard:

- all standing teams must be manifest-backed if they are part of the delivery
  system rather than ad hoc experiments
- standing teams must return clean, operator-readable briefs during central sync

### 2. Gated delivery is defined, but enforcement evidence is still concentrated in one record

The active next-target workflow is explicitly phase-gated from `P0` through
`P4`. That is the correct control model, but the gate discipline is currently
documented most concretely in the dedicated delivery program document and review
addenda. It still needs to be folded into the broader architecture corpus so
the control model is not fragmented.

Required standard:

- phase order must be explicit in delivery docs
- downstream work must remain blocked until the prior gate passes
- every phase must carry acceptance, rollback, and evidence requirements

### 3. Error handling improved in operations, but the standard needs to be made explicit

Recent task-layer fixes corrected two important operator-facing failures:

- missing environment dependencies now stop with actionable guidance
- database reset no longer proceeds behind a dead PostgreSQL bridge and report
  false success

That is the right direction, but the repo still needs a single explicit error
handling standard across operational code and runtime services.

Required standard:

- fail fast on missing prerequisites and broken connectivity
- never print success text after an underlying command has failed
- downgrade transient bridge noise to warning-level operator output unless
  health actually degrades
- use actionable remediation text for operator-facing failures

### 4. Logging is partially standardized by behavior, not yet by one published contract

The current system behavior suggests three logging classes:

- product/runtime logs from Go services
- operator logs from Python invoke tasks
- transport noise from external tools such as `kubectl port-forward`

The review found the right operational interpretation, but not yet one durable
repo-level contract that defines severity, shape, and ownership across those
classes.

Required standard:

- Go runtime logs must be structured and context-rich
- Python task logs must be concise, phase-oriented, and operator-readable
- external tool noise must not be reclassified as service failure without
  corroborating health evidence

### 5. Language ownership is mostly clear in practice, but it should be explicit

Current usage is converging on the right split:

- Go for core runtime, services, transport, orchestration logic, and durable
  product behavior
- TypeScript for frontend and workflow-composer UI surfaces
- Python for task automation, management scripting, testing harnesses, and
  repository support
- SQL for schema and migration logic

This needs to be treated as a written standard so implementation choices stay
consistent as the team set grows.

Required standard:

- no new core product logic should default to Python if it belongs in Go
- Python is the standard language for app management scripting across hosting
  platforms
- PowerShell must not be the implementation language for app-tied management
  workflows; it is only acceptable as a thin local invocation wrapper when the
  host platform requires it
- Python remains allowed for invoke tasks, local automation, management flows,
  and tests
- frontend interaction patterns remain TypeScript-owned

## Standardization Contract

### Language by aspect

#### Go

Primary language for:

- `core/`
- service orchestration logic
- bus and transport behavior
- team execution runtime
- long-lived backend workflows

Rules:

- prefer explicit error returns over implicit failure
- wrap errors with context
- avoid `panic` for recoverable runtime paths
- keep operationally relevant identifiers in logs

#### TypeScript

Primary language for:

- `interface/`
- workflow-composer onboarding
- user-facing orchestration controls
- frontend diagnostics and interaction state

Rules:

- treat backend state as authoritative
- surface recoverable errors clearly in UI state
- avoid inventing alternate workflow semantics in the client

#### Python

Allowed for:

- `ops/`
- app management scripting
- invoke task entrypoints
- testing harnesses
- local support scripts

Rules:

- Python is not the primary product implementation language
- Python is the default implementation language for cross-platform management
  workflows tied to the app
- avoid embedding management logic in PowerShell unless it is strictly a local
  platform wrapper around Python-owned behavior
- operator tasks must prefer clear task-phase output over raw tracebacks
- scripts must fail with actionable remediation when prerequisites are missing

#### SQL

Primary language for:

- migrations
- schema lifecycle
- durable database contracts

Rules:

- migration errors are hard failures unless a specific non-fatal path is
  intentionally designed and documented

## Logging Standard

### Go runtime

Go services must emit structured logs with enough context to diagnose:

- component
- workflow or request identifier
- team or bus subject where relevant
- severity
- error cause

Avoid:

- ad hoc console printing for service logic
- ambiguous messages without identifiers

### Python operations

Task logs should be human-readable and phase-based:

- announce the phase
- report the concrete action
- stop on hard failure
- print remediation guidance when the operator can act

Avoid:

- raw dependency tracebacks as the primary UX
- success text after failure
- noisy repeated output that obscures the failing phase

### External bridge tooling

Logs from `kubectl port-forward` and similar bridge tools must be interpreted as
transport diagnostics unless supported by health failures or repeated runtime
impact.

## Error Handling Standard

### Preconditions

- validate environment dependencies before starting work
- validate bridge availability before database operations
- validate required process state before claiming teardown or startup success

### Failures

- return or raise immediately on hard precondition failure
- wrap low-level failures with operator-relevant context
- distinguish warnings from hard failures explicitly

### Recovery

- transient reconnect behavior is acceptable for bus and bridge layers
- recovery claims must be supported by health checks, not assumptions

## Delivery Status

Standardization is improved but not complete.

What is standardized now:

- gated next-target delivery exists and is explicit
- operator entrypoints are normalized to `uv run inv ...` and local virtualenv
  invocation
- memory restart sequencing is closer to operational truth
- standing architect, development, and AGUI teams are manifest-backed
- Python usage is now bounded to tasks and testing by policy

What remains open:

- fold review addenda back into the base architecture docs
- stabilize Core startup timing during full memory restart
- publish one canonical logging and error-handling contract in the base docs
- harden clean reply behavior for all standing teams during architecture sync

## Recommended Next Actions

1. Persist the three standing teams as manifest-backed definitions.
2. Fold this review and the AGI review addendum into the core architecture
   documents.
3. Close the remaining `lifecycle.memory-restart` Core startup flake before
   advancing the delivery program gate.

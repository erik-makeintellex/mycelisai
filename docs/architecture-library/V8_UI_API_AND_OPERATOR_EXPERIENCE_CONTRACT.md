# V8 UI API and Operator Experience Contract
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

Status: canonical V8 operator/UI contract.

This index keeps the current UI/API contract stable and delegates durable detail to:
- [V8 UI Operator Workspace Contract](V8_UI_OPERATOR_WORKSPACE_CONTRACT.md)
- [V8 UI API Screen Contracts](V8_UI_API_SCREEN_CONTRACTS.md)
- [V8 UI Delivery And Governance Contract](V8_UI_DELIVERY_AND_GOVERNANCE_CONTRACT.md)
- [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md)

Compatibility note: [V8 Universal Soma And Context Model PRD](V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md) defines the next product correction where Soma and Council are universal entities and AI Organizations are governed work contexts.

## 1. Why this document exists

V8 UI must feel like creating and operating an AI Organization, not a generic chat app with architecture labels. This contract protects terminology, source-of-truth layering, screen state, and proof expectations.

The default product experience must feel like creating and operating an AI Organization.

## 2. Canonical terminology

- AI Organization: instantiated operator-visible work context
- Soma: primary operator counterpart and orchestrator
- Team Leads: operational coordinators for focused work
- Advisors/Council: advisory and governance support
- Departments/Specialists: scoped execution structure
- AI Engine Settings: provider and routing controls
- Memory & Continuity: retained knowledge, outputs, identity, and response posture

UI copy should use these terms directly unless a screen is explicitly compatibility/developer-only.

## 3. Product guardrails

### 3.1 Anti-generic-chat rule

The product must not open as a raw assistant shell. First-run creates or explores AI Organizations; re-entry opens the last or selected organization and leads to Soma.

### 3.2 Progressive disclosure rule

Default UI shows organization identity, Soma, purpose, starter path, Team Lead support, recent activity, memory/continuity, and safe guided settings. Advanced UI may reveal inheritance, provider policy, templates, capability posture, and runtime details.

### 3.3 Two explicit UX and control layers

Default operator surface:
- create/re-enter AI Organization
- Soma workspace
- intent-driven interaction
- guided teams/groups
- activity, memory, resources, and settings

Default Operator Surface:

Advanced architecture/runtime surface:
- config origin
- inheritance and overrides
- policy/capability posture
- deployment influence
- runtime availability

Advanced Architecture / Runtime Surface:

### 3.4 Source-of-truth layers

The product must keep these layers distinct:
1. guided UI settings
2. bundle/file configuration
3. deployment/env overrides
4. runtime state
5. state and architecture docs

Deployment/env overrides wire runtime endpoints and profiles; they do not replace organization truth. deployment/env overrides must not replace bundle-defined runtime organization truth. advanced UI may explain deployment/env influence, but it must not become a second source of runtime truth.

### 3.5 Primary orchestration flow

```text
User -> Soma -> routing/council -> team leads -> departments/specialists -> reviews/memory/activity -> Soma
```

### 3.6 API normalization rule

Every screen expects normalized envelopes with stable IDs, labels, role/type metadata, loading/error/empty states, and no raw backend leakage.

### 3.7 Inspectable agency rule

The UI must make agency inspectable without making the default operator path noisy.

Default surfaces show:
- direct answer, proposal, execution result, blocker, or recovery state
- why Soma stayed direct or escalated
- what team, tool, source, model, or deployment context mattered
- what output, artifact, retained result, or audit event was produced

Advanced surfaces show:
- plan/action/source/result details
- capability and approval posture
- run, group, artifact, log, checkout, deployment-root, and execution-root evidence
- memory versus deployment-context boundaries
- capability manifests for MCP tools, custom connectors, local scripts, external APIs, and plugins when they affect execution

Hidden autonomy is not acceptable for mutating work, external calls, private context use, artifact creation, or team/group execution.

## 4. Screen and flow contracts

### 4.1 First-run flow

The product must not open into a raw assistant conversation as its primary identity.

### 4.2 Create AI Organization flow

Create ends with an instantiated organization and an Open Soma Workspace action.

### 4.3 Choose template vs empty start

Template and empty-start paths both produce runtime organization truth.

### 4.4 AI Organization home and header contract

The header keeps organization identity, status, and navigation legible.

### 4.5 Soma-primary workspace behavior

Soma owns answers, proposals, execution results, blockers, and recovery.

The dashboard Soma surface and the AI Organization Soma surface use one shared operating model. Intent suggestions live inside Soma. After meaningful actions, the interface shows a causal summary that connects what Soma understood, what teams/tools were coordinated, what outputs were produced, what state changed, and what the user can do next.

### 4.6 Advisor, Department, and Specialist visibility rules

Advisors, Departments, and Specialists stay visible without becoming a flat default chat roster.

### 4.7 Advanced-mode boundaries

Advanced mode exposes inheritance, config origin, groups, resources/MCP, deep memory, runs/bus state, settings, auth, and docs without becoming required for first use.

Current screen contract families:
- first-run and organization creation
- AI Organization home/re-entry
- Soma workspace
- direct answer
- proposal/cancel/execute
- teams/groups and retained outputs
- resources/Connected Tools
- memory and activity
- settings and advanced boundaries

See [V8 UI API Screen Contracts](V8_UI_API_SCREEN_CONTRACTS.md).

managed exchange items and artifact schemas remain inspectable. The UI includes inspect-only managed exchange surfaces for Channels, Threads, and Recent Artifacts. It also shows inspect-only trust, sensitivity, and review labels for managed exchange items where they matter operationally. managed exchange security labels stay inspectable in advanced surfaces without leaking forbidden/internal-only security implementation detail into the default Soma-first UX.

## 5. API/UI contract mapping by screen and action

UI-facing routes must normalize:
- `answer`
- `proposal`
- `execution_result`
- `blocker`
- `loading`
- `empty`
- `error`

Backend/API changes require the plan block in [Testing](../TESTING.md#backendapi---ui-target-plan).

## 6. Delivery proof

Use [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) for full browser proof and [Remote User Testing](../REMOTE_USER_TESTING.md) for delivered-topology human walkthroughs.

Minimum product proof covers:
- organization entry/re-entry
- Soma direct answer
- governed proposal cancel/execute
- team/group creation and retained output review
- inspectable escalation rationale, source/tool use, and produced output
- deployment/execution-root visibility when release proof or recovery depends on it
- capability manifest visibility for promoted MCP/custom/API/script/plugin execution paths
- settings persistence or clear blocker
- activity/audit review
- recovery from dependency failure

## 7. Current boundary

Do not expand this file into another route matrix. Add durable details to the split docs and update this index only when the contract set changes.

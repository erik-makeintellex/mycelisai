# V8 UI API and Operator Experience Contract

> Status: Canonical V8 PRD
> Last Updated: 2026-03-21
> Purpose: Define the exact V8 operator and user-facing flows so UI implementation reflects AI Organization behavior instead of collapsing into generic chat UX.
> Depends On: `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md`, `docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md`

## 1. Why this document exists

V8 introduces AI Organizations, Team Leads, Advisors, Departments, Specialists, AI Engine Settings, and Learning & Context as first-class product concepts.

Those terms are not decorative renames for a generic assistant UI.

This document defines the operator-facing product contract that must sit between architecture vocabulary and implementation.
Its job is to prevent the interface from drifting into:
- a blank "ask anything" assistant shell
- a raw model-picker workflow
- a chat app that only adds organization words in the margins
- a settings-first experience that hides the AI Organization mental model

The default product experience must feel like creating and operating an AI Organization.

## 2. Canonical terminology

- **AI Organization**: the instantiated, operator-visible organization object. It is a live organization, not a template and not a raw provider session.
- **Team Lead**: the primary operator-facing lead for an AI Organization. The default workspace opens to the Team Lead, not to an anonymous assistant.
- **Advisors**: high-level advisory roles attached to the organization. They are visible as part of the organization structure, but they are not the default conversation surface.
- **Departments**: grouped execution areas within the AI Organization. Departments organize Specialist work and scoped operating context.
- **Specialists**: individual role-focused workers inside a Department or a governed roster. They are visible as organizational members, not presented as a flat buddy list by default.
- **AI Engine Settings**: provider-policy and model-routing controls that shape how the organization runs. These are real configuration surfaces, but they are advanced by default.
- **Learning & Context**: continuity, learning visibility, reviewed memory posture, tone, and behavioral defaults for the AI Organization. These settings exist, but they should not overwhelm first-run creation.

Terminology rule:
- UI copy should use these names directly.
- Implementation labels should not replace them with generic terms such as "assistant", "bot", "workspace chat", or "AI settings" unless the screen is explicitly showing a compatibility or developer-only surface.

## 3. Product guardrails

### 3.1 Anti-generic-chat rule

The product must not open into a raw assistant conversation as its primary identity.

Default entry posture:
- if no AI Organization exists, show first-run organization creation guidance
- if one or more AI Organizations exist, show organization selection or the last active AI Organization
- once an AI Organization is opened, the primary action is "Open Team Lead Workspace", not "New Chat"

Disallowed default behaviors:
- empty landing page with only a chat box and placeholder assistant text
- top-level provider/model picker as the primary first-run decision
- flat list of every Specialist as if the product were a multi-bot messenger
- exposing raw advanced config before the operator has chosen or created an AI Organization

### 3.2 Progressive disclosure rule

Beginner operators should see:
- AI Organization identity
- Team Lead identity
- organization purpose
- starter template or empty-start choice
- clear success path into the Team Lead workspace

Advanced operators may inspect:
- Advisor roster
- Department structure
- Specialist roster
- AI Engine Settings
- Learning & Context
- deeper org configuration

But those advanced surfaces should be hidden until the operator intentionally opens them.

### 3.3 Two explicit UX and control layers

Default Operator Surface:
- create AI Organization
- Team Lead-first workspace
- intent-driven interaction
- Advisors, Departments, Automations, Recent Activity, and Learning & Context
- AI Engine Settings and Response Style as guided, bounded controls

Advanced Architecture / Runtime Surface:
- separate and non-default
- organization defaults and inheritance visibility
- department overrides and Specialist role bindings
- automation definitions
- capability posture
- response-style inheritance
- bundle/config source truth
- deployment/env influence
- runtime availability and distributed execution posture later

Advanced-surface rules:
- it must remain separate from the default Team Lead-first flow
- it must make inheritance legible
- it must make config origin legible
- it must not expose raw jargon without context
- it must not undermine the simplicity of the default operator experience

### 3.4 Source-of-truth layers

The product must keep these layers distinct:

1. guided UI settings
   - safe, bounded operator controls for the current release surface
2. bundle/file configuration
   - reproducible organization defaults, inheritance inputs, and automation truth
3. deployment/env overrides
   - environment-specific provider/media/runtime wiring only
4. runtime state
   - the live resolved organization, execution posture, and service availability
5. state and architecture docs
   - README, V8.1, V8.2, and `V8_DEV_STATE.md` explain target, release, and implementation truth

Critical rule:
- deployment/env overrides must not replace bundle-defined runtime organization truth
- advanced UI may explain deployment/env influence, but it must not become a second source of runtime truth

### 3.5 API normalization rule

Every screen in this PRD expects normalized API envelopes.

Required envelope posture:
- stable object IDs
- human-readable labels
- role/type metadata
- explicit loading/error/empty states
- no raw backend implementation leakage into the default UX

## 4. Screen and flow contracts

### 4.1 First-run flow

**Purpose**

Help a new operator understand that Mycelis creates AI Organizations, not disposable chat threads.

**User sees**

- a first-run welcome surface
- one short sentence defining an AI Organization
- two primary actions: `Create AI Organization` and `Explore Templates`
- one secondary action: `Start Empty`
- a compact preview of what the operator will define:
  - Team Lead
  - Advisors
  - Departments
  - Specialists
  - AI Engine Settings
  - Learning & Context

**User actions**

- start the creation flow
- browse starter templates
- choose an empty AI Organization start
- reopen a previously created AI Organization if one already exists

**API data required**

- `GET /api/v1/organizations?view=summary`
  - returns existing AI Organization summaries for resume/select behavior
- `GET /api/v1/templates?view=starter`
  - returns starter template cards with name, purpose, and structural summary
- `GET /api/v1/system/bootstrap-status`
  - returns service readiness needed to create or open an AI Organization

**Hidden advanced details**

- raw provider lists
- full Advisor roster defaults
- Department/Specialist editable structure
- AI Engine Settings internals
- Learning & Context internals

**Success state**

- operator enters the Create AI Organization flow with a chosen start mode
- or lands on an existing AI Organization home

**Empty state**

- no organizations yet: show first-run guidance and starter templates

**Loading state**

- show organization/template placeholders and "Preparing AI Organization setup..."

**Error state**

- if bootstrap/API readiness is unavailable, show a blocker that explains creation cannot continue and links to recovery guidance

### 4.2 Create AI Organization flow

**Purpose**

Guide the operator through creating a real AI Organization with a clear structure and identity.

**User sees**

- a multi-step or single-page guided builder titled `Create AI Organization`
- required fields:
  - AI Organization name
  - purpose/mission
  - Team Lead posture
- optional but visible structure summaries:
  - Advisors
  - Departments
  - Specialists
- advanced sections collapsed by default:
  - AI Engine Settings
  - Learning & Context

**User actions**

- name the organization
- describe its purpose
- select template-based or empty-based creation
- review organization structure before creation
- create the AI Organization

**API data required**

- `POST /api/v1/organizations/drafts`
  - starts a draft organization object
- `PATCH /api/v1/organizations/drafts/{draft_id}`
  - updates name, purpose, Team Lead posture, and chosen start mode
- `POST /api/v1/organizations`
  - creates the live AI Organization from the validated draft

**Hidden advanced details**

- raw YAML/config representation
- provider-policy override graph
- internal inheritance/preference debug data

**Success state**

- a live AI Organization exists and the operator is routed to its home screen

**Empty state**

- when no templates are available, creation still supports empty/manual setup without degrading into chat-first UX

**Loading state**

- show step-level saving/creating indicators with draft preservation

**Error state**

- validation errors appear inline next to the affected field
- create failure returns a blocker explaining what failed and whether the draft was preserved

### 4.3 Choose template vs empty start

**Purpose**

Make the start-mode decision explicit without confusing templates with live organizations.

**User sees**

- a side-by-side choice:
  - `Start from Template`
  - `Start Empty`
- template cards that summarize:
  - intended organization type
  - Team Lead posture
  - Advisor presence
  - Department count
  - Specialist count
  - AI Engine Settings summary
  - Learning & Context summary
- an `Empty AI Organization` card that clearly states the operator will define structure manually

**User actions**

- inspect a template
- compare templates
- choose a template
- choose empty start

**API data required**

- `GET /api/v1/templates?view=organization-starters`
  - lists starter templates with organization-facing summaries
- `GET /api/v1/templates/{template_id}`
  - returns the selected template detail for preview
- `PATCH /api/v1/organizations/drafts/{draft_id}/start-mode`
  - stores `template` or `empty`

**Hidden advanced details**

- raw template schema
- internal provider-policy scopes
- hidden compatibility metadata

**Success state**

- draft now records a clear start mode and, if applicable, a selected template

**Empty state**

- if no starter templates exist, the screen explains that empty start remains available and is still an AI Organization path

**Loading state**

- template cards use skeletons; preview panes show a loading placeholder

**Error state**

- template fetch failure preserves empty-start availability and shows a non-destructive error

### 4.4 AI Organization home and header contract

**Purpose**

Provide a stable home screen that makes the AI Organization legible before the operator enters deeper interaction.

**User sees**

- organization header showing:
  - AI Organization name
  - purpose
  - Team Lead identity
  - current status/readiness
- organization summary panels:
  - Advisors summary
  - Departments summary
  - Specialists summary
  - AI Engine Settings summary
  - Learning & Context summary
- primary actions:
  - `Open Team Lead Workspace`
  - `Review Organization Structure`
  - `Open Advanced Settings`

**User actions**

- open Team Lead workspace
- inspect role structure
- rename or update purpose where allowed
- open advanced config surfaces intentionally

**API data required**

- `GET /api/v1/organizations/{organization_id}/home`
  - returns the normalized organization home payload
- `PATCH /api/v1/organizations/{organization_id}`
  - updates editable organization metadata such as display name or purpose

**Hidden advanced details**

- internal IDs beyond what deep links require
- provider routing tables
- raw continuity/memory policy documents

**Success state**

- organization header and structure render clearly and route the operator into the correct next action

**Empty state**

- if the organization has no Advisors, Departments, or Specialists yet, show explicit "Not configured yet" states rather than omitting those sections

**Loading state**

- header skeleton plus section placeholders

**Error state**

- one recoverable organization-home error block with retry, not a blank screen

### 4.5 Team Lead-first workspace behavior

**Purpose**

Make the Team Lead the default operating surface for the AI Organization.

**User sees**

- a Team Lead workspace header bound to the current AI Organization
- Team Lead name/role and organization purpose
- recent organization context
- a conversation or command surface framed as working with the Team Lead for this AI Organization
- optional roster sidebar or drawer for Advisors, Departments, and Specialists

**User actions**

- send a message or instruction to the Team Lead
- inspect whether Advisors or Specialists were engaged
- open organization context panels
- navigate back to organization home

**API data required**

- `GET /api/v1/organizations/{organization_id}/workspace`
  - returns Team Lead workspace metadata and recent interaction history
- `POST /api/v1/organizations/{organization_id}/workspace/messages`
  - sends operator input to the Team Lead workspace
- `GET /api/v1/organizations/{organization_id}/roster`
  - fetches organization members for side-panel visibility

**Hidden advanced details**

- raw provider/model names unless explicitly requested in advanced diagnostics
- internal chain-of-thought or ungoverned tool traces
- direct Specialist session launch controls as the default action

**Success state**

- operator receives an answer, proposal, execution result, or blocker in the Team Lead workspace while remaining inside the AI Organization frame

**Empty state**

- if no prior activity exists, show a Team Lead-specific starter prompt and organization-aware examples, not a generic assistant empty state

**Loading state**

- message send state is scoped to the Team Lead thread

**Error state**

- structured workspace blocker explains whether the issue is organization readiness, policy, tooling, or transport

### 4.6 Advisor, Department, and Specialist visibility rules

**Purpose**

Keep organization structure visible without overwhelming the operator or flattening roles into a generic chat roster.

**User sees**

- Advisors shown as strategic/oversight roles
- Departments shown as grouped execution areas
- Specialists shown inside the relevant Department or inspection drawer
- member counts and concise role summaries

**User actions**

- inspect Advisors
- inspect Departments
- inspect Specialists
- optionally route into a deeper member detail view when the product later supports it

**API data required**

- `GET /api/v1/organizations/{organization_id}/roster`
  - returns Advisors, Departments, Specialists, visibility metadata, and summaries

**Hidden advanced details**

- raw policy assignments
- hidden system roles not intended for operator management
- internal-only orchestration agents

**Success state**

- operator can understand who belongs to the AI Organization and how the structure is grouped

**Empty state**

- missing Advisors, Departments, or Specialists show explicit placeholders with creation or configuration guidance

**Loading state**

- roster groups load independently so the home/workspace shell stays usable

**Error state**

- role-group fetch failures degrade to scoped section errors instead of taking down the whole screen

### 4.7 Advanced-mode boundaries

**Purpose**

Protect the default AI Organization experience from collapsing into a config dashboard while still supporting serious operator understanding and control later.

**User sees**

- an explicit `Advanced` entrypoint, not always-open raw settings
- advanced sections that may later expose:
  - organization defaults and inheritance visibility
  - Department overrides
  - Specialist role bindings
  - detailed automation definitions
  - capability posture
  - Response Style inheritance
  - bundle/config source truth
  - deployment/env influence
  - runtime availability and distributed execution posture when that later ships

**User actions**

- intentionally enter advanced mode
- inspect or edit organization defaults where allowed
- inspect or edit Department overrides and Specialist role bindings where allowed
- inspect automations, capability posture, and runtime influence in more detail
- return to the standard organization surfaces

**API data required**

- future advanced endpoints must be split by concern and return normalized, origin-aware data
- advanced APIs must make the difference legible between:
  - guided UI settings
  - bundle/file configuration
  - deployment/env influence
  - live runtime state
- advanced APIs must not dump raw YAML, raw secrets, or opaque provider internals into the default product contract

**Hidden advanced details**

- secret material
- low-level provider auth and endpoint wiring where those belong in file/env/config only
- host-specific runtime plumbing
- cluster or distributed node plumbing unless intentionally promoted later

**Success state**

- advanced inspection or editing makes inheritance and config origin easier to understand without breaking the operator's mental model of the AI Organization

**Empty state**

- absent optional advanced values show governed defaults rather than blank raw forms

**Loading state**

- advanced sections load lazily after the operator opens them

**Error state**

- advanced-setting failures remain confined to the advanced surface and do not collapse the Team Lead workspace or organization home

## 5. API/UI contract mapping by screen and action

| Screen / Flow | Primary UI actions | Required API contract | Required data returned |
| --- | --- | --- | --- |
| First-run | create, explore templates, resume | `GET /api/v1/organizations?view=summary`, `GET /api/v1/templates?view=starter`, `GET /api/v1/system/bootstrap-status` | org summaries, starter templates, readiness state |
| Create AI Organization | save draft, create organization | `POST /api/v1/organizations/drafts`, `PATCH /api/v1/organizations/drafts/{draft_id}`, `POST /api/v1/organizations` | draft id, validation state, created organization summary |
| Template vs empty | select start mode, inspect template | `GET /api/v1/templates?view=organization-starters`, `GET /api/v1/templates/{template_id}`, `PATCH /api/v1/organizations/drafts/{draft_id}/start-mode` | template summaries, selected-template detail, draft start mode |
| Organization home | open Team Lead workspace, inspect structure | `GET /api/v1/organizations/{organization_id}/home`, `PATCH /api/v1/organizations/{organization_id}` | header, summary cards, editable org metadata |
| Team Lead workspace | send message, inspect roster | `GET /api/v1/organizations/{organization_id}/workspace`, `POST /api/v1/organizations/{organization_id}/workspace/messages`, `GET /api/v1/organizations/{organization_id}/roster` | Team Lead context, message history, terminal result payloads, roster summary |
| Visibility rules | inspect Advisors, Departments, Specialists | `GET /api/v1/organizations/{organization_id}/roster` | grouped member lists with summaries and visibility metadata |
| Advanced architecture / runtime surface | inspect or edit advanced configuration later | future normalized endpoints grouped by organization defaults, overrides, automations, capability posture, config origin, and runtime influence | origin-aware summaries, inheritance context, validation results, and scoped persisted summaries |

## 6. UX rules that implementation must not violate

1. The operator must always understand which AI Organization they are operating.
2. The default conversation surface must always be the Team Lead workspace for that AI Organization.
3. Advisors, Departments, and Specialists must appear as organization structure, not as a flat chat-contact list by default.
4. Template selection must be framed as choosing an AI Organization starting blueprint, not as selecting a chat persona.
5. AI Engine Settings and Learning & Context must be real product concepts, but they must stay behind progressive disclosure until the operator intentionally opens advanced mode.
6. Empty-start flows must still feel like creating an AI Organization, not like launching a blank assistant session.
7. Every screen must define purpose, visible structure, user actions, API data requirements, and explicit loading/empty/error/success states before UI implementation begins.

## 7. Delivery implication

This document is the canonical PRD for the V8 operator experience slice.

Implementation work that touches:
- first-run onboarding
- organization creation
- template selection
- organization home/header
- Team Lead workspace behavior
- role visibility
- advanced settings boundaries
- API contracts for those screens

must align with this document, the V8 runtime contracts, and the V8 config/bootstrap model in the same slice.

V8.1 extension note:
- Loop Profiles, Runtime Capabilities, Agent Type runtime truth, and the first bounded `Automations` surface are defined in `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md`.

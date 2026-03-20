# Mycelis V8 - Development State

> Updated: 2026-03-19
> Canonical state file for active V8 grading and delivery tracking
> References: `README.md`, `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md`, `V7_DEV_STATE.md` (legacy migration input)

---

## Purpose

`V8_DEV_STATE.md` is now the authoritative global state file for active V8 grading.

Use this file to track:
- current V8 delivery status
- accepted marker transitions
- gate evidence
- blockers and dependencies
- immediate next actions

Use `V7_DEV_STATE.md` only as a migration input and historical checkpoint source while V7 planning artifacts are being migrated.

## Feature Status Legend

- `REQUIRED`: must exist for target delivery or gate pass
- `NEXT`: highest-priority upcoming slice
- `ACTIVE`: currently in development
- `IN_REVIEW`: implemented and awaiting validation/review
- `COMPLETE`: delivered and accepted
- `BLOCKED`: cannot advance until a dependency or defect is resolved

## V8 Delivery Program Snapshot

```text
V8-0  Migration baseline and architecture alignment                 [ACTIVE]
V8-1  Canonical contract definition                                [COMPLETE]
V8-2  Config and bootstrap model planning                          [COMPLETE]
V8-3  Backend primitive refactor                                   [REQUIRED]
V8-4  Frontend/operator refactor                                   [REQUIRED]
V8-5  Documentation and naming migration                           [ACTIVE]
V8-6  Verification and release hardening                           [REQUIRED]
```

## Current V8 Grading Baseline

### 1. Architecture alignment and migration inventory

Status:
1. `ACTIVE` V8 alignment is now carried by the README directive, the V8 runtime contracts, and the V8 bootstrap-planning surface instead of a separate tracker artifact.
2. `COMPLETE` current fixed-organism assumptions vs V8 target model have been mapped at the planning level.
3. `COMPLETE` backend/frontend/docs hotspots have been identified for the first migration wave.
4. `ACTIVE` carry those alignment conclusions into explicit V7-bootstrap migration rules and de-hardcoding work now that the canonical plan exists.

Primary references:
- `README.md`
- `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
- `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`

Key grading conclusion:
- Mycelis already has the execution substrate needed for V8, but still exposes one hardcoded Soma/Council organizational expression as if it were the platform itself.

### 2. Provider and service architecture

Status:
1. `ACTIVE` V8 explicitly requires support for local, hosted, and hybrid model-provider infrastructure.
2. `COMPLETE` provider abstraction is confirmed as an existing core subsystem, not a plugin afterthought.
3. `ACTIVE` instantiated-organization provider policy now drives runtime routing with inherited organization, kernel/council-role, team, and agent scope resolution.
4. `NEXT` extend provider-policy grading beyond the standing-team bridge so future instantiated organizations can declare richer approval/data-boundary posture without fallback-era assumptions.

Required V8 provider support:
- local/open-source: Ollama, LM Studio, vLLM, custom OpenAI-compatible endpoints
- hosted/commercial: OpenAI, Anthropic, Gemini, other OpenAI-compatible providers

Required routing posture:
- role-level routing
- council-level routing
- inception-level policy boundaries
- local-first and hybrid deployment support
- explicit data-boundary and approval posture

### 3. State and documentation migration

Status:
1. `ACTIVE` canonical grading is moving from `V7_DEV_STATE.md` to `V8_DEV_STATE.md`.
2. `ACTIVE` active non-archive docs are being updated to reference the V8 state file.
3. `ACTIVE` top documentation entrypoints now link the detailed framework-memory surfaces needed for restart/onboarding (`README.md`, `docs/README.md`, `docs/archive/README.md`).
4. `COMPLETE` clearly stale dated architecture-review addenda were removed from `docs/architecture/`.
5. `REQUIRED` V7 architecture-library documents remain authoritative migration inputs until V8 replacements exist.

State-file rules:
1. all new status transitions, blocker updates, and gate evidence belong in `V8_DEV_STATE.md`
2. `V7_DEV_STATE.md` remains readable for historical migration evidence, but should not be the primary grading target for new work
3. active docs should describe V7 documents as migration inputs and V8 state as the live delivery scoreboard

### 4. Runtime and UI migration posture

Status:
1. `COMPLETE` canonical V8 contracts for `Inception`, `Soma Kernel`, `Central Council`, provider-policy scope, and identity/continuity state now exist.
2. `COMPLETE` the config/bootstrap model now documents configuration sources, template entry points, scope inheritance, precedence, and V7-to-V8 migration rules.
3. `REQUIRED` replace standing-team bootstrap assumptions with configurable organization-resolution contracts.
4. `REQUIRED` keep Workspace simple and Soma-first while making the operator model kernel-aware instead of fixed-identity-bound.
5. `REQUIRED` continue the full UI retheme and density-reduction effort under V8 delivery targets, not legacy V7 framing.

### 5. Testing and gate posture

Status:
1. `REQUIRED` every V8 migration slice must still carry backend, UI, and docs proof in the same delivery window.
2. `ACTIVE` architecture and state-file changes must be validated by doc-surface checks and in-app docs alignment where applicable.
3. `ACTIVE` clean-run discipline is now part of the delivery contract for runtime and integration-style checks: stop prior services, verify ports/processes are clear, sweep stray compiled Go binaries from prior runs, start only the minimum required stack, run the check, and tear the stack back down unless explicitly needed.
4. `REQUIRED` runtime/API changes must include explicit UI review/test targets and focused live-flow evidence when user-visible behavior is touched.

### 6. Implementation-slice cleanup and convergence rule

Status:
1. `REQUIRED` every implementation slice that edits runtime, config, UI, or supporting logic must include a touched-file cleanup/convergence pass.
2. `REQUIRED` chunk completion for code-editing slices must include a focused review-team sweep covering contract alignment, dead-code scan, integration continuity, and test alignment.
3. `REQUIRED` deferred cleanup that cannot be completed safely in the same slice must be reported explicitly in the chunk result.

Implementation rule:
1. review each changed file for stale fixed-Soma/fixed-Council assumptions, dead branches, outdated comments, duplicated helpers, obsolete config paths, and naming drift
2. remove safe-to-remove stale code in the touched files as part of the same slice
3. keep the review scope to touched files and directly adjacent files needed for correctness
4. commit cleanup changes in the same chunk when they are part of the same logical convergence work

## Active V8 Queue

```text
Task 001  V8 alignment and migration inventory                      [IN_REVIEW]
Task 002  Inception / kernel / council contract definition          [COMPLETE]
Task 003  Provider-policy scope contract                            [COMPLETE]
Task 004  Config and bootstrap model planning                       [COMPLETE]
Task 005  Standing-team bootstrap de-hardcoding plan                [ACTIVE]
Task 006  Workspace/UI kernel-aware refactor plan                   [ACTIVE]
Task 007  V8 docs/state migration and grading discipline            [ACTIVE]
Task 008  Planning-integration validation pass                      [COMPLETE]
Task 009  Next-execution/governance guidance migration              [NEXT]
```

## Current Checkpoint (2026-03-19)

Delivery updates in this checkpoint:
1. `COMPLETE` reviewed the new root `README.md` and adopted V8 as the active development/grading target.
2. `COMPLETE` established `V8_DEV_STATE.md` as the canonical state file for new work.
3. `ACTIVE` migrated active non-archive documentation references from `V7_DEV_STATE.md` to `V8_DEV_STATE.md` where those docs define current execution discipline.
4. `COMPLETE` preserved V7 architecture-library documents as migration inputs rather than rewriting them into premature V8 specifications.
5. `COMPLETE` documented how V7 bootstrap sources, implicit behaviors, fixed Soma/Council posture, and runtime-state coupling roll forward into the explicit V8 configuration/bootstrap model.
6. `COMPLETE` validated cross-doc planning consistency so README, the architecture-library index, runtime contracts, bootstrap model, docs manifest, and tests all reference the canonical V7->V8 migration contract.
7. `ACTIVE` landed the next Task 005 implementation cut: startup now instantiates runtime organization truth directly from self-contained bundle data in `core/config/templates/*.yaml` and fails closed unless a valid bootstrap bundle is available.
8. `COMPLETE` promoted clean-run testing discipline into the active testing/operations contract so runtime and integration checks must not stack on unknown local processes.
9. `COMPLETE` extended the clean-run contract so compiled Go services from prior `go build`, `go run`, or manual binary launches are explicitly detected and terminated before the next runtime or integration check.
10. `BLOCKED` the latest lifecycle/doc hardening commits are local-only until GitHub SSH-agent/key access is restored; branch publication is currently blocked by `Permission denied (publickey)` during `git push`.
11. `COMPLETE` refreshed top documentation entrypoints so restart/onboarding flow now points directly at the detailed framework-memory surfaces (`README.md`, `docs/README.md`, `docs/archive/README.md`).
12. `COMPLETE` removed clearly stale dated review docs from `docs/architecture/` to reduce documentation clutter without disturbing active migration inputs.
13. `COMPLETE` doc-surface validation is green again after the documentation cleanup pass.
14. `COMPLETE` wired provider-policy inheritance from the instantiated runtime organization into live Soma routing so organization defaults, kernel/council role defaults, team defaults, and agent overrides now resolve through one policy-bounded path.
15. `COMPLETE` retired the `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` runtime compatibility path so provider routing now resolves only from the instantiated organization policy and never from startup env maps.
16. `COMPLETE` retired the remaining no-bundle bootstrap fallback: normal startup now has one truth path only, `template -> instantiation -> runtime organization`, while mirrored `config/teams/*.yaml` packaging remains compatibility input rather than startup truth.
17. `COMPLETE` operator startup requirements are now explicit: Core requires at least one valid bootstrap bundle under `config/templates/`, and `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must be set whenever more than one bundle is present.
18. `COMPLETE` defined a dedicated V8 UI/API/operator PRD so first-run, AI Organization creation, template-vs-empty start, organization home, Team Lead-first workspace behavior, role visibility, advanced-mode boundaries, and screen-to-API mapping now have one canonical contract.
19. `COMPLETE` exposed the new V8 UI/API/operator contract through the architecture-library index, the in-app docs manifest, and doc-test enforcement so future UI slices do not drift back toward generic-chat UX.
20. `COMPLETE` implemented the first bounded V8 UI slice: `/dashboard` now opens with `Create AI Organization`, offers `Start from template` vs `Start empty`, uses user-facing AI Organization terms, and routes successful creation into a dedicated AI Organization home instead of generic workspace chat.
21. `COMPLETE` added the minimal backend contract for the new flow: starter template listing from bundle-backed organization templates, AI Organization creation requests, recent-organization summaries, and organization-home loading for the success path.
22. `COMPLETE` hardened the AI Organization entry UX before live GUI sign-off: starter-template and recent-organization failures are now decoupled, retry/recovery paths preserve any still-valid actions, architecture/dev wording is removed from operator-facing copy, and the organization home makes the Team Lead more concrete.
23. `COMPLETE` added bounded browser coverage for the AI Organization entry flow so the dominant create path, template path, empty-start path, partial-failure recovery, and forbidden-copy checks now run in Playwright before wider UI slices proceed.
24. `COMPLETE` turned the post-create organization page into the first Team Lead-first workspace shell: organization context stays visible, the Team Lead now has a concrete identity/presence block, primary next actions drive the active workspace focus, and advanced controls remain summarized instead of becoming the default surface.
25. `COMPLETE` introduced the first guided Team Lead interaction workflow: the organization workspace now offers structured Team Lead starting actions, sends a minimal workspace action request to the backend, and renders a shaped guidance response without expanding into generic chat or advisor orchestration.
26. `COMPLETE` hardened the guided Team Lead workflow for bounded production use: action failures now surface clear retry guidance without dropping the AI Organization frame, partial guidance payloads normalize into readable Team Lead sections, duplicate submits stay disabled while loading, and browser coverage now exercises guided-action failure recovery.
27. `COMPLETE` added inspect-only Advisor and Department visibility to the Team Lead workspace so the AI Organization structure now feels visible and understandable without displacing the Team Lead as the primary operating counterpart.
28. `COMPLETE` added inspect-only AI Engine Settings and Memory & Personality surfaces to the Team Lead workspace so operators can understand the current organization posture without opening advanced controls or seeing architecture jargon.
29. `COMPLETE` added inspect-only Advisor and Department detail views to the Team Lead workspace and wired both Team Lead and support-column actions into those focused views without displacing the Team Lead as the primary counterpart.
30. `COMPLETE` added an inspect-only AI Engine Settings detail view with scoped model-assignment visibility so operators can inspect organization-wide defaults, team defaults, and specific-role override status without opening advanced editing.
31. `COMPLETE` introduced the first bounded AI Engine edit capability at the organization level: the Team Lead workspace now offers a guided `Change AI Engine` flow with curated operator-facing options, backend validation of allowed values only, immediate workspace refresh, and bounded retry handling without exposing raw provider/model identifiers or advanced configuration panels.
32. `COMPLETE` added controlled Department-level AI Engine override with inheritance clarity: Department details now show whether each Team is following the organization default or using an override, operators can apply a guided Department-specific engine choice or revert cleanly, and organization-level changes continue to flow only to Departments that still inherit the default.
33. `COMPLETE` added the first bounded organization-level Response Contract milestone: the workspace now exposes a safe default `Response Style` contract with curated tone/structure/detail profiles, backend validation of allowed values only, immediate summary refresh after updates, and bounded retry handling without exposing raw prompt text or advanced policy controls.
34. `COMPLETE` added inspect-only Agent Type Profiles to Department details so the Team workspace now exposes the missing inheritance layer between Team defaults and individual agent instances, including clear user-facing visibility into type-level AI Engine bindings and Response Style bindings without opening agent-instance editing.
35. `COMPLETE` added controlled Agent Type AI Engine binding in Department details so operators can pin a curated AI Engine to a role type, cleanly return that role type to the Team default, and preserve inheritance clarity between Team-level engine choices and type-specific specialist behavior.
36. `COMPLETE` added controlled Agent Type Response Style binding in Department details so operators can pin a curated Response Style to a role type, return that role type to the Organization / Team default, and preserve inheritance clarity between the organization-wide Response Style and type-specific specialist output behavior.
37. `COMPLETE` aligned the root page with the V8.1 living AI Organization model so product-facing entry messaging now centers AI Organizations, Team Lead-guided operation, continuous reviews/checks/updates, and the post-creation workspace instead of legacy swarm-console framing.

Evidence:
1. README directive review completed against `README.md`
2. runtime-contract alignment confirmed against `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
3. bootstrap-planning alignment confirmed against `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
4. active documentation references updated to point at `V8_DEV_STATE.md`
5. V7-to-V8 bootstrap migration narrative committed in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
6. planning-integration validation sweep recorded in this chunk plus `tests/test_docs_links.py`
7. template-bundle loader, runtime-organization instantiation, fail-closed startup selection logic, and the `v8-migration-standing-team-bridge` bundle landed in `core/internal/bootstrap/template_bundle.go`, `core/internal/bootstrap/template_bundle_test.go`, `core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup.go`, and mirrored chart config packaging
8. clean-run testing discipline is now documented in `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md`
9. lifecycle cleanup/status now sweep and report stray compiled Go services in `ops/lifecycle.py` with focused regression coverage in `tests/test_lifecycle_tasks.py`
10. compiled-service inspection now fails closed: `lifecycle.status` reports unknown inspection state and `lifecycle.down` blocks runtime/integration testing when local process inspection cannot verify cleanup
11. the current local checkpoint series now includes the bounded Team Lead workspace hardening and visibility slices on this branch, with SSH publication still blocked until local key/agent state is repaired
12. documentation cleanup refreshed `README.md`, `docs/README.md`, and `docs/archive/README.md`, and removed `docs/architecture/AGI_ARCHITECTURE_REVIEW_2026-03-06.md` plus `docs/architecture/STANDARDIZATION_REVIEW_2026-03-06.md`
13. validation: `uv run pytest tests/test_docs_links.py -q` -> `23 passed`
14. instantiated-organization provider policy now resolves through `core/internal/swarm/provider_policy.go`, is carried by `core/internal/bootstrap/template_bundle.go`, is applied during startup in `core/cmd/server/main.go`, and is exercised by focused bootstrap/swarm coverage
15. the standing-team bridge bundle now declares a conservative provider-policy default in both `core/config/templates/v8-migration-standing-team-bridge.yaml` and `charts/mycelis-core/config/templates/v8-migration-standing-team-bridge.yaml` so local and charted startup follow the same instantiated-organization routing path as tests
16. startup now fails closed when the bootstrap bundle set is missing or invalid, when `MYCELIS_BOOTSTRAP_TEMPLATE_ID` requests a bundle that is absent, and when multiple bundles exist without an explicit selection; runtime provider routing ignores legacy env-map inputs and startup truth now remains bundle-only in code/tests/state (`core/internal/bootstrap/template_bundle.go`, `core/cmd/server/bootstrap_startup.go`, `core/cmd/server/main.go`, `core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup_test.go`)
17. dedicated V8 UI/API/operator contract PRD now lives in `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md` and defines first-run flow, AI Organization creation, Team Lead-first workspace behavior, role visibility, advanced-mode boundaries, and screen-to-API mapping
18. architecture-library discovery and in-app docs exposure now include the V8 UI/API/operator contract through `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `interface/lib/docsManifest.ts`, and `tests/test_docs_links.py`
19. Windows `lifecycle.down` CIM timeout/inspection failure remains separate tooling debt; it is not part of the retired bootstrap fallback surface and still tracks under the lifecycle hardening work until resolved
20. the first V8 creation-flow slice now lives in `interface/app/(app)/dashboard/page.tsx`, `interface/components/organizations/CreateOrganizationEntry.tsx`, `interface/app/(app)/organizations/[id]/page.tsx`, and `interface/components/organizations/OrganizationContextShell.tsx`
21. minimal AI Organization starter/create/home APIs now live in `core/internal/server/templates.go`, `core/internal/server/organizations.go`, and `core/internal/server/admin.go`, with focused backend coverage in `core/internal/server/organizations_test.go`
22. route/API exposure docs now include the V8 AI Organization entry flow through `docs/API_REFERENCE.md`, and frontend regression coverage lives in `interface/__tests__/pages/DashboardPage.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/__tests__/shell/ZoneA_Rail.test.tsx`
23. the entry flow now keeps recent-organization resume and starter-template setup resilient under partial API failure, exposes retry/recovery actions in-place, removes operator-visible dev/architecture copy leaks, and strengthens Team Lead status in the AI Organization home (`interface/components/organizations/CreateOrganizationEntry.tsx`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/DashboardPage.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`)
24. bounded browser automation for the entry flow now lives in `interface/e2e/specs/v8-organization-entry.spec.ts` and covers dominant AI Organization entry framing, template selection, empty start, organization-home landing, retry/recovery under partial failure, and visible forbidden-copy enforcement
25. the organization page now functions as a Team Lead-first workspace shell rather than a static landing state: `interface/components/organizations/OrganizationContextShell.tsx` keeps the AI Organization header visible, adds a concrete Team Lead identity block plus next-action controls, and switches the active workspace focus between planning, advisor review, department review, AI Engine Settings summary, and Memory & Personality summary with focused page and browser coverage
26. the first Team Lead interaction workflow now lives in `core/internal/server/organizations.go`, `interface/components/organizations/TeamLeadInteractionPanel.tsx`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/organizations/TeamLeadInteractionPanel.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; it intentionally stops at guided Team Lead responses and does not yet implement advisor orchestration, raw agent selection, advanced configuration panels, or a full chat system
27. guided Team Lead resilience now adds readable fallback shaping in `core/internal/server/organizations.go`, failure/retry and malformed-response rendering in `interface/components/organizations/TeamLeadInteractionPanel.tsx`, focused backend/frontend tests in `core/internal/server/organizations_test.go`, `interface/__tests__/organizations/TeamLeadInteractionPanel.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and bounded browser retry coverage in `interface/e2e/specs/v8-organization-entry.spec.ts`; advisor orchestration, raw agent selection, advanced configuration panels, and a full chat system remain intentionally out of scope
28. inspect-only Advisor and Department visibility now lives in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; the Team Lead remains the primary workspace counterpart while Advisors and Departments are visible as supporting structure, and full advisor orchestration, raw agent selection, and department management remain intentionally out of scope
29. inspect-only AI Engine Settings and Memory & Personality workspace surfaces now live in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; they explain current engine/profile and memory/personality posture in operator language while keeping advanced provider, capability, memory, and personality controls intentionally out of scope
30. inspect-only Advisor and Department detail views now live in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Team Lead action buttons and support-column buttons both open the same focused workspace views while advisor orchestration, department editing, and raw agent selection remain intentionally out of scope
31. inspect-only AI Engine Settings detail now lives in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; it exposes organization-wide AI engine posture, team-default visibility, and specific-role override visibility in operator language while keeping advanced editing, provider-policy terms, and runtime capability controls intentionally out of scope
32. the first bounded AI Engine edit path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/lib/organizations.ts`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; operators can choose from curated organization-level AI Engine profiles, invalid values are rejected server-side, retry stays in-place on failure, and team/role-level editing plus advanced provider/capability controls remain intentionally out of scope
33. controlled Department-level AI Engine inheritance now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/lib/organizations.ts`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details expose inherited-versus-overridden AI Engine state, guided Department overrides can be set and reverted, organization-level AI Engine changes continue to update inheriting Departments only, and agent-level overrides plus advanced configuration remain intentionally out of scope
34. the first bounded Response Contract path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/lib/organizations.ts`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; the workspace exposes `Response Style` as a safe organization-wide default with curated options, invalid values are rejected server-side, retry stays in-place on failure, and agent-level response overrides plus raw prompt/policy editing remain intentionally out of scope
35. inspect-only Agent Type Profile runtime truth now lives in `core/internal/server/organizations.go`, `core/internal/server/organizations_test.go`, `interface/lib/organizations.ts`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details now show user-facing agent type names, helps-with summaries, Team-default versus type-specific AI Engine bindings, and Organization/Team-default versus type-specific Response Style bindings without exposing raw model IDs, raw prompt text, or agent-instance editing
36. the first bounded Agent Type AI Engine binding path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/lib/organizations.ts`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details now let operators apply a curated AI Engine to an Agent Type, retry in place on failure, return that role type to the Team default, and preserve Team-default inheritance for other role types without exposing raw model IDs, raw provider terms, or agent-instance overrides
37. the first bounded Agent Type Response Style binding path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/lib/organizations.ts`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details now let operators apply a curated Response Style to an Agent Type, retry in place on failure, return that role type to the Organization / Team default, and preserve inherited organization-wide Response Style behavior for other role types without exposing raw prompt text, raw policy text, or agent-instance overrides
38. `interface/app/(marketing)/page.tsx` and `interface/__tests__/pages/LandingPage.test.tsx` now present and enforce the V8.1 root-page story around AI Organizations, Team Lead-guided work, recent activity, safe guided control, and post-creation workspace expectations without leaking internal architecture terms or old console/chat framing

### 6. V8 contract shell introduction

Status:
1. `COMPLETE` `docs/architecture-library/V8_RUNTIME_CONTRACTS.md` now exists as the new V8 runtime contract shell.
2. `COMPLETE` the initial V8 runtime contract set is now fully defined.

### 7. Inception contract definition

Status:
1. `COMPLETE` the Inception contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes the top-level organization contract for V8.

### 8. Soma Kernel contract definition

Status:
1. `COMPLETE` the Soma Kernel contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes the configurable coordination/runtime layer for an Inception.

### 9. Central Council contract definition

Status:
1. `COMPLETE` the Central Council contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes configurable advisory composition rather than a fixed built-in council pattern.

### 10. Provider Policy contract definition

Status:
1. `COMPLETE` the Provider Policy contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` team-level and agent-level configuration scope are now captured explicitly as first-class contract needs.

### 11. Identity and Continuity State contract definition

Status:
1. `COMPLETE` the Identity and Continuity State contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` the initial V8 runtime contract set is now complete.
3. `COMPLETE` the program has already moved into the config/bootstrap planning phase before backend refactor work begins.

### 12. V8 config/bootstrap planning shell introduction

Status:
1. `COMPLETE` `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` now exists as the planning shell for V8 config/bootstrap behavior.
2. `COMPLETE` the shell now has initial structure for user-facing and bootstrap-model planning slices.

### 13. User Concept Layer definition

Status:
1. `COMPLETE` the `User Concept Layer` is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` architecture-to-user concept mapping is now documented for the V8 UI mental model.
3. `COMPLETE` the beginner-friendly AI-organization mental model is now part of the bootstrap planning surface.

### 14. Bootstrap resolution flow definition

Status:
1. `COMPLETE` the `Bootstrap resolution flow` is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` team-level defaults, agent-level overrides, and beginner-versus-advanced bootstrap entry modes are now explicit in the V8 planning model.
3. `COMPLETE` the staged organization-resolution path is now documented for later implementation work.

### 15. Scope inheritance definition

Status:
1. `COMPLETE` the `Scope inheritance` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` the inheritance chain from `Inception -> Soma Kernel -> Central Council roles -> Team defaults -> Agent overrides` is now documented.
3. `COMPLETE` team and agent configuration are now explicitly treated as first-class scopes in bootstrap planning.

### 16. Precedence rules definition

Status:
1. `COMPLETE` the `Precedence rules` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` source precedence, scope precedence, policy-before-override handling, and conflict resolution are now part of the V8 bootstrap planning model.
3. `COMPLETE` the planning model now has enough structure to define how organizations are created before bootstrap activation.

### 17. Template and instantiation entry points definition

Status:
1. `COMPLETE` the `Template and instantiation entry points` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` V8 now distinguishes template-based creation, empty/manual creation, operator/API creation, and config-file/bootstrap creation as canonical organization-entry modes.
3. `COMPLETE` templates are now documented as reusable organization blueprints that feed bootstrap resolution with default kernel, council, team, and policy shape.
4. `COMPLETE` the `Migration from V7 bootstrap assumptions` section now explains how fixed V7 startup assumptions collapse into explicit V8 configuration sources, inheritance, and precedence rules.

### 18. Standing-team bootstrap de-hardcoding plan

Status:
1. `ACTIVE` Chunk 4.2 captured the standing-team de-hardcoding plan inside `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `ACTIVE` runtime/config references to fixed `prime-*` teams, hard-coded council rosters, kernel defaults, and provider wiring have documented migration paths into template definitions, bootstrap resolution, scoped inheritance, and provider-policy configuration.
3. `ACTIVE` Helm/runtime concerns (env injection for `MYCELIS_API_KEY`, 8080/8081 port normalization, config-file mounts, container storage) are now described alongside the bootstrap plan so infrastructure slices can align with runtime refactors.
4. `COMPLETE` current-state scan now records that the no-bundle startup dependency is gone: `core/internal/bootstrap/template_bundle.go` now fails closed when bundles are missing, env-map provider overrides are retired, and mirrored Helm/config packaging of `config/teams/*.yaml` no longer participates in startup truth.
5. `ACTIVE` the bridge now owns runtime startup truth through native bundle content: `core/internal/bootstrap/template_bundle.go` loads `core/config/templates/*.yaml` bundles, instantiates a runtime organization directly from bundle-defined org/team/agent data, and fails closed when no valid bootstrap bundle is available.

Evidence:
1. Plan section `## Standing-team bootstrap de-hardcoding plan` (2026-03-13) now explicitly calls out the active files and assumptions discovered in Chunk 4.2 (Helm values/env/ports, ops bridge tasks, bootstrap services, template APIs, team manifests, runtime storage expectations).
2. README review confirmed no additional guidance changes were required after adding the plan (`README.md`, 2026-03-13).
3. `v8-migration-standing-team-bridge` now validates as a self-contained V8 bundle through the bootstrap loader path without requiring standing-team manifest refs (`core/config/templates/v8-migration-standing-team-bridge.yaml`, `core/internal/bootstrap/template_bundle_test.go`).
4. startup selection regression coverage now explicitly proves bundle-instantiated runtime organizations, missing-bundle failure, invalid-bundle failure, missing-requested-bundle failure, and registry continuity with bundle presence (`core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup_test.go`, `core/internal/swarm/registry_test.go`).

Next steps:
1. Promote generated per-organization bootstrap bundles so startup stays bundle-only without depending on the fixed standing-team bridge asset forever.
2. Replace bootstrap seeding logic with template-instantiation + scope-aware inheritance.
3. Promote provider-policy scopes and Helm env/port/mount/storage alignment into actionable runtime slices with tests.

### 19. Cluster/runtime bootstrap contract alignment

Status:
1. `COMPLETE` Helm deployment now provisions the configuration bundle (`cognitive.yaml`, `policy.yaml`, standing-team YAMLs) via a ConfigMap volume so Pods read deterministic bootstrap inputs instead of whatever was baked in the image.
2. `COMPLETE` `MYCELIS_API_KEY` is required and injected through a Kubernetes Secret created by the chart (or supplied via `coreAuth.existingSecret`); `ops.k8s.deploy` now refuses to proceed when the key is missing.
3. `COMPLETE` Core HTTP port contract is unified on `8080` (`core/cmd/server` default, Helm `PORT` env, Service/bridge forwarding, ops defaults).
4. `COMPLETE` Charts and ops scripts document/mount the writable storage contract (`/data` PVC for artifacts + `$MYCELIS_WORKSPACE` under `/data/workspace`), and the ConfigMap keeps read-only bootstrap files under `/app/config`.

Evidence:
1. Helm templates: `charts/mycelis-core/templates/deployment.yaml`, `configmap-config.yaml`, `_helpers.tpl`, `core-auth-secret.yaml`, updated `values.yaml`, plus new `config/` assets.
2. Ops automation: `ops/k8s.py` enforces API key injection and 8080 port-forward; `ops/config.py` defaults shift to 8080; `ops/interface.py` startup guidance updated.
3. Runtime entrypoint: `core/cmd/server/main.go` now defaults to `PORT=8080`.
4. Validation: `uv run pytest tests/test_docs_links.py -q`.

Next steps:
1. Wire provider-policy scopes and template-instantiation flow into the runtime without depending on standing-team tables.
2. Extend Helm/ops surface so template bundles are generated from the new template serialization path instead of the baked defaults once Task 005 code work lands.

### 20. Testing/QA alignment review

Status:
1. `COMPLETE` browser QA test plan for Workspace chat now references the V8 contract (inline Soma chat, terminal states, direct-first routing) instead of the outdated V7 framing.
2. `COMPLETE` manual testing expectations now list the canonical outcomes (`answer`, `proposal`, `execution_result`, `blocker`) and reinforce the V8 inline-governance posture.

Evidence:
1. `tests/ui/browser_qa_plan_workspace_chat.md` updated on 2026-03-13 to remove V7 references and describe the V8 workspace, happy paths, and edge cases.

Next steps:
1. Keep automated Playwright suites aligned with the same inline-chat expectations once Workspace refactors land.

### 21. V8.1 living organization architecture definition

Status:
1. `COMPLETE` `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` now defines the canonical V8.1 architecture contract for Loop Profiles, Runtime Capabilities, promoted Response Contract inheritance, promoted Agent Type Profiles, and bounded Automations visibility.
2. `COMPLETE` the architecture-library index, in-app docs manifest, and doc-test contract now expose V8.1 as canonical architecture truth rather than leaving `v8-1.md` as a loose planning draft.
3. `ACTIVE` the first shippable V8.1 state is now moving through concrete slices: a bounded read-only Review Loop execution path exists, while bundle/config loop definitions, capability contract surfaces, and read-only Automations visibility remain next.

Evidence:
1. `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` now carries the canonical V8.1 PRD for Loop Profiles, Runtime Capabilities, promoted Response Contract and Agent Type Profile runtime truth, safety rules, testing requirements, and initial release definition.
2. `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `interface/lib/docsManifest.ts`, and `tests/test_docs_links.py` now index, expose, and enforce the V8.1 architecture doc as a canonical surface.
3. `docs/architecture-library/V8_RUNTIME_CONTRACTS.md` and `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md` now cross-link the V8.1 architecture extension so runtime and operator docs do not drift into a parallel truth.

### 22. First read-only Review Loop execution

Status:
1. `COMPLETE` a minimal backend Loop Executor now supports the first bounded Review Loop path inside `core/internal/server`, including Loop Profile loading, owner resolution, structured review output, and in-memory result storage.
2. `COMPLETE` manual internal triggering now exists through `/api/v1/internal/organizations/{id}/loops/{loopId}/trigger`, and recent loop results are inspectable through `/api/v1/internal/organizations/{id}/loops/results` for debugging.
3. `COMPLETE` the first loop execution remains safely bounded: review-only, no actuation, no external calls, no filesystem/hardware access, and no UI exposure of raw loop internals.
4. `COMPLETE` interval-backed Review Loops now execute automatically through a minimal in-process scheduler that only runs profiles with `interval_seconds`, prevents overlap, and stays stoppable during shutdown.
5. `COMPLETE` read-only Recent Activity visibility now surfaces the latest review/check/update outcomes inside the Team Lead workspace without exposing loop controls, raw logs, or technical internals.
6. `COMPLETE` the first event-driven Review Loop milestone now reacts to bounded internal organization events (`organization created`, Team Lead guidance completion, organization AI Engine change, and Response Style change), executes read-only reviews, records safe activity results, and shares overlap protection with scheduled execution.
7. `NEXT` promote loop definitions into bundle/config truth and surface bounded Automations visibility in the Team Lead workspace without widening into unsafe execution.

Evidence:
1. `core/internal/server/review_loops.go` now defines the first V8.1 Review Loop framework, default loop profiles, team/agent-type owner resolution, structured findings/suggestions/status output, and read-only result logging.
2. `core/internal/server/admin.go` now registers the internal trigger/debug routes, and `core/internal/server/organizations.go` seeds default review loops when new organizations are created.
3. `core/internal/server/review_loops_test.go` now proves successful execution, owner resolution, structured output, invalid-loop rejection, stored result visibility, and read-only preservation of organization state.
4. `core/internal/server/review_loop_scheduler.go` and `core/internal/server/review_loop_scheduler_test.go` now add the first bounded scheduled-loop runner with interval-based execution, invalid-config rejection, overlap protection, stoppable lifecycle wiring, and result/failure logging.
5. `GET /api/v1/organizations/{id}/loop-activity` now exposes safe user-facing activity summaries, and `interface/components/organizations/OrganizationContextShell.tsx` now renders them in a non-intrusive `Recent Activity` support panel with lightweight polling, empty-state handling, and failure-safe fallback.
6. `core/internal/server/review_loops.go`, `core/internal/server/organizations.go`, `core/internal/server/review_loops_test.go`, and `core/internal/server/organizations_test.go` now add bounded event-driven review execution for allowed internal organization events, safe failure logging into Recent Activity, and shared overlap protection so reactive reviews stay read-only and operator-visible without exposing raw event-bus internals.

## Immediate Next Actions

1. `COMPLETE` run the planning-integration validation pass so README, the architecture-library index, docs manifests, and doc-tests all confirm the new V7-to-V8 bootstrap migration contract.
2. `NEXT` update the next-execution and governance guidance so delivery slices are expressed as V8 migration slices rather than only V7 holdovers.
3. `REQUIRED` apply the touched-file cleanup/convergence rule and review-team sweep to the first backend/runtime refactor chunk and all later code-editing slices.
4. `REQUIRED` validate doc-surface integrity after the state-file migration (`docs links`, `docs manifest`, and in-app docs visibility).
5. `REQUIRED` keep all new implementation/testing checkpoints in `V8_DEV_STATE.md` going forward.
6. `BLOCKED` restore SSH-agent/key access and push the latest local lifecycle/doc/state commits to the remote branch.
7. `NEXT` continue the documentation authority cleanup so active entrypoints stay lean while compatibility docs and archive material remain intentionally separated.
8. `NEXT` promote generated per-organization bootstrap bundles so startup remains bundle-only without relying on the fixed standing-team bridge asset long term.
9. `NEXT` extend the Team Lead workspace from guided starting actions into real request/history API flows and deeper structure surfaces while keeping advisor orchestration, raw agent selection, and advanced configuration behind intentional later slices.
10. `NEXT` add bounded Team Lead response history and operator-visible continuity inside the AI Organization workspace without widening into generic chat or exposing raw agent-selection controls.
11. `NEXT` define the first bundle/config contract slice for Loop Profiles and Runtime Capabilities so V8.1 execution surfaces exist as safe, inspectable configuration before live execution is introduced.
12. `NEXT` add read-only `Automations` visibility to the Team Lead workspace using the V8.1 user-facing terms `Automations`, `Watchers`, and `Reviews` without exposing advanced controls or enabling execution.
13. `NEXT` connect the first Review Loop backend results to bounded operator visibility so Automations can show review outcomes without exposing internal loop mechanics or enabling actuation.
14. `NEXT` replace hard-coded/default interval loop seeding with bundle-defined Loop Profiles so scheduled execution stays reproducible and organization-specific.
15. `NEXT` widen Recent Activity into the first bounded `Automations` operator surface so Reviews, Checks, and Updates feel continuous without exposing configuration complexity by default.



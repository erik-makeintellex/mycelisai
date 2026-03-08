# Verification & Testing Protocol

Mycelis employs a **5-Tier Testing Strategy** covering backend handlers, frontend components, end-to-end flows, integration tests, and governance smoke tests.

Latest clean verification baseline (2026-03-07):
- `uv run inv ci.baseline` -> pass (logging/topic/line gates + core tests + interface build + typecheck + vitest)
- `uv run pytest tests/test_docs_links.py -q` -> pass (`13` passed)
- `cd core && go test -p 1 ./internal/artifacts ./internal/cognitive -count=1` -> pass
- `cd core && go test -p 1 ./internal/server -run "TestHandleSaveArtifactToFolder|TestHandleMe|TestHandleUpdateSettings_AssistantNamePersists" -count=1` -> pass
- `cd core && go test -p 1 ./internal/swarm -run "TestInternalToolRegistry_LocalCommand_Registered|TestInternalToolRegistry_LocalCommand_RejectsShellSnippet" -count=1` -> pass
- `cd interface && npx vitest run __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/FocusModeToggle.test.tsx __tests__/settings/UsersPage.test.tsx __tests__/system/SystemQuickChecks.test.tsx __tests__/lib/labels.test.ts __tests__/pages/SettingsPage.test.tsx __tests__/shell/ShellLayout.test.tsx --reporter=dot` -> pass (`78` passed)
- `cd interface && npx tsc --noEmit` -> pass
- Worktree state after cleanup commits: clean
- Last full E2E sweep remains 2026-03-02: `cd interface && npx playwright test --reporter=dot` -> pass (`51` passed, `4` skipped)
- Latest focused Playwright harness slice (2026-03-07):
  - `uv run inv interface.e2e --project=chromium --spec=e2e/specs/accessibility.spec.ts` -> pass (`3` passed)
  - `uv run inv test.e2e --project=mobile-chromium --spec=e2e/specs/mobile.spec.ts` -> pass (`3` passed)
  - `uv run inv interface.e2e --live-backend --project=chromium --spec=e2e/specs/workspace-live-backend.spec.ts` -> pass (`1` passed)
- Latest focused lifecycle/task slice: `$env:PYTHONPATH='.'; uv run pytest tests/test_db_tasks.py tests/test_lifecycle_tasks.py tests/test_ci_tasks.py tests/test_logging_tasks.py -q` -> pass (`26` tests)
- Latest focused Go runtime slice: `cd core && go test ./internal/mcp ./internal/swarm ./pkg/protocol -count=1` -> pass
- Latest destructive gate validation: `uv run inv lifecycle.memory-restart --frontend` -> pass

## Quick Reference

```bash
# Run from scratch/ root.
# Primary runner: uv run inv ...
# Compatibility probe: uvx --from invoke inv -l
# Unsupported bare alias: uvx inv ...
uv run inv core.test             # Go unit tests (all packages)
uv run inv interface.test        # Vitest unit tests (jsdom)
uv run inv interface.e2e         # Playwright E2E tests (Playwright starts/stops the Next.js server; Invoke clears stale UI listeners)
uv run inv interface.e2e --live-backend --spec=e2e/specs/workspace-live-backend.spec.ts  # Real Core-backed Workspace UI contract
uv run inv core.smoke            # Governance smoke tests
uv run inv ci.test               # Blocking Go + Vitest validation
uv run inv interface.check       # HTTP smoke test against running dev server
uv run inv logging.check-schema  # Event schema + docs coverage gate
uv run inv logging.check-topics  # Hardcoded swarm topic gate
uv run inv quality.max-lines --limit 350  # Hot-path max-lines gate with legacy caps
uv run inv lifecycle.memory-restart --frontend          # Full memory reset + post-restart memory probes
uv run inv ci.entrypoint-check   # Verify uv / uvx runner matrix
uv run inv ci.baseline           # Canonical strict baseline (docs/logging/topics/line gates + core + interface)
```

Runner matrix:
- `uv run inv ...` is the supported path for real task execution and testing.
- `uvx --from invoke inv -l` is a lightweight compatibility probe only.
- `uvx inv ...` is expected to fail and is checked as a negative control by `uv run inv ci.entrypoint-check`.

Signal/channel standard:
- When tests touch NATS channel behavior, use the canonical subject families and source metadata defined in `docs/architecture/NATS_SIGNAL_STANDARD_V7.md`.
- Development-only infrastructure subjects are not part of product orchestration and should stay out of authoritative runtime tests unless the test is explicitly exercising dev-only behavior.
- Current focused runtime check: `cd core && go test ./internal/swarm ./pkg/protocol -count=1`
- Current focused UI check: `cd interface && npx vitest run __tests__/dashboard/SignalContext.test.tsx __tests__/lib/signalNormalize.test.ts --reporter=dot`
- Current focused Workspace chat contract check: `cd interface && npx vitest run __tests__/dashboard/MissionControlChat.test.tsx __tests__/lib/labels.test.ts --reporter=dot`
- Current focused execution feedback check: `cd interface && npx vitest run __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx --reporter=dot`
- Current focused Workspace failure-model check: `cd interface && npx vitest run __tests__/lib/missionChatFailure.test.ts __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/dashboard/MissionControlChat.test.tsx __tests__/store/useCortexStore.test.ts --reporter=dot`
- Current focused Launch Crew contract check: `cd interface && npx vitest run __tests__/workspace/LaunchCrewModal.test.tsx __tests__/store/useCortexStore.test.ts --reporter=dot`
- Current focused Launch Crew browser proof: `uv run inv interface.e2e --project=chromium --spec=e2e/specs/proposals.spec.ts` (proposal outcome + blocker recovery)
- Current focused Launch Crew live confirm proof: `uv run inv interface.e2e --live-backend --project=chromium --spec=e2e/specs/proposals.spec.ts` (stubbed proposal display + real `/api/v1/intent/confirm-action` round-trip)
- Current focused team-sync contract check: `$env:PYTHONPATH='.'; uv run pytest tests/test_misc_tasks.py -q`
- Current focused README navigation check: `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`
- Current docs/task drift rule: canonical docs must not contain executable bare `uvx inv ...` examples outside explicit negative-control guidance.

UI delivery contract:
- Use `docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md` as the authoritative map for UI terminal states and backend transaction expectations.
- A UI test is incomplete if it proves rendering but does not prove the backend effect or intentionally blocked state behind that render.

---

## Product Delivery Proof (Required)

For execution-facing UI work, tests must prove product behavior, not only component mechanics.

Every changed UI path must document and test:
1. the initiating user interaction
2. the expected terminal UI state: `answer`, `proposal`, `execution_result`, or `blocker`
3. the backend effect caused by the frontend: HTTP call, DB mutation, run/event creation, or NATS interaction
4. the failure path and recovery affordance

Minimum proof requirements by path:

| UI Path | UI Proof | Backend/Transaction Proof |
| --- | --- | --- |
| Workspace / Soma chat | response lands in one valid terminal state, not planning-only output | `/api/v1/chat` call occurs and returned payload is classified correctly |
| Direct council chat | specialist answer or structured blocker card renders | `/api/v1/council/{member}/chat` path is exercised and timeout/failure behavior is mapped |
| Prime-team sync | architecture directives publish to canonical team lanes and operator-visible replies are collected | `swarm.team.{team}.internal.command` publishes and `signal.status`/`signal.result` replies are observed |
| Launch Crew / guided manifestation | proposal or activation result is visible | proposal/confirm endpoints produce identifiers and mutation state |
| Workflow composer | invalid graph blocks, valid graph proposes or activates | validate/compile/activate endpoints called with expected payloads |
| Runs / timeline / chain | operator can inspect run state and outcome | run/event/chain/conversation queries return and are rendered consistently |
| System / degraded mode | recovery guidance is visible, not only colored status | health/status responses map to explicit degraded-state actions |

Required test layers for UI-affecting delivery:
- component tests: terminal state rendering and action affordances
- integration tests: request/response mapping between UI and backend
- product-flow tests: user journey reaches a real outcome
- backend transaction tests: route/DB/NATS side effects match UI claims
- failure tests: timeout, rejection, degraded dependency, retry/reroute flow

For operator-facing blocker work, a passing test set must prove:
- the store preserves raw diagnostics
- the shared failure model classifies the error once
- Workspace blocker card, degraded banner, and status drawer all render from that shared model

Disallowed testing posture:
- proving only that a button renders
- proving only that a fetch function was called without validating resulting user state
- treating planning-only content as a passing success state
- validating backend code in isolation when the UI contract is the feature being changed

---

## Tier 1: Backend Unit Tests (Go)

**Goal:** Verify API handler logic, nil guards, validation, and SQL interactions.
**Speed:** < 5s (mocked DB via `go-sqlmock`).

### Location

| File | Coverage |
|------|----------|
| `core/internal/server/governance_test.go` | Policy CRUD, pending approvals, resolve approve/reject |
| `core/internal/server/telemetry_test.go` | Runtime telemetry, trust threshold GET/PUT/range |
| `core/internal/server/mission_test.go` | Mission CRUD, intent commit TX, negotiate, sensor configs, blueprint extraction |
| `core/internal/server/mcp_test.go` | MCP install validation, list, delete, tool call, library |
| `core/internal/server/mcp_toolsets_test.go` | MCP tool set list/create/update/delete paths + nil guards |
| `core/internal/server/memory_search_test.go` | Memory search, sitreps, sensors |
| `core/internal/server/proposals_test.go` | Proposal CRUD, approve/reject, conflict detection |
| `core/internal/server/identity_test.go` | Identity, teams, settings |
| `core/internal/server/catalogue_test.go` | Agent catalogue CRUD |
| `core/internal/server/artifacts_test.go` | Artifact storage |
| `core/internal/server/cognitive_test.go` | Cognitive router config |
| `core/internal/server/testhelpers_test.go` | Shared test helpers (`newTestServer`, `withDB`, `withGuard`, `doRequest`, `assertStatus`, `assertJSON`) |
| `core/internal/cognitive/middleware_test.go` | LLM retry, schema validation, timeouts |

### Patterns

- **Partial server construction:** Use `newTestServer(opts...)` with option functions (`withDB`, `withGuard`) to build only the dependencies a handler needs.
- **SQL mocking:** `go-sqlmock` for all DB interactions. Use `ExpectQuery`/`ExpectExec` with regex patterns.
- **Path params:** Use `http.NewServeMux` with Go 1.22+ route patterns (`"GET /path/{id}"`) for `r.PathValue()` support.
- **Nil guards:** Every handler that touches optional infrastructure (Guard, Overseer, MCP, Cognitive, Mem) has a nil-guard test returning 503.

### Running

```bash
uv run inv core.test                          # All packages
go test -v ./internal/server/...           # Server handlers only
go test -v -run TestHandleGovernance ./internal/server/...  # Single test pattern
go test -v ./internal/mcp/ -count=1        # MCP service/library/executor/toolset suites
go test -v -run TestHandleMCP ./internal/server/... -count=1
go test -v -run TestHandleUpdateToolSet ./internal/server/... -count=1
go test -v -run TestScoped ./internal/swarm/... -count=1
```

---

## Tier 2: Frontend Unit Tests (Vitest)

**Goal:** Verify component rendering, store interactions, UI transaction mapping, and terminal delivery states in jsdom.
**Speed:** < 10s.

### Location

| Directory | Coverage |
|-----------|----------|
| `interface/__tests__/shell/` | ShellLayout, ZoneA_Rail, GovernanceModal |
| `interface/__tests__/dashboard/` | TelemetryRow, ActiveMissionsBar, SensorLibrary, ManifestationPanel, Streams, TeamsSummaryCard, SignalContext |
| `interface/__tests__/` | Forge, Approvals, Console, CommandPage, CommandRail |

### Infrastructure

| File | Purpose |
|------|---------|
| `interface/__tests__/setup.ts` | Global setup: `mockFetch()`, `MockEventSource`, `next/navigation` mock |
| `interface/__tests__/mocks/reactflow.ts` | ReactFlow jsdom mock (ResizeObserver, components, hooks, enums) |
| `interface/vitest.config.ts` | jsdom environment, `@/` alias, excludes `e2e/**` |

### Patterns

- **API mocking:** Use `mockFetch` from `setup.ts` — no MSW needed. Call `mockFetch.mockResolvedValueOnce({...})`.
- **Store state:** Set Zustand state directly via `useCortexStore.setState({...})`.
- **ReactFlow components:** Import mock via `vi.mock('reactflow', () => import('../mocks/reactflow'))`.
- **Next.js navigation:** `usePathname` and `useRouter` mocked globally in `setup.ts`.
- **Terminal state assertions:** For execution-facing surfaces, assert the final delivery state (`answer`, `proposal`, `execution_result`, `blocker`) instead of only intermediate loading behavior.
- **Transaction assertions:** When a component triggers an API call, assert the expected request target/payload and the resulting user-visible outcome.

### Running

```bash
uv run inv interface.test                     # All Vitest tests
npx vitest run --reporter=verbose          # Verbose output (from interface/)
npx vitest run __tests__/shell/            # Single directory
```

---

## Tier 3: End-to-End Tests (Playwright)

**Goal:** Verify full user journeys through the running application.
**Speed:** 30s-2min (Playwright owns the Interface server; start Core separately only for live-backend specs).

For execution-facing UI work, Playwright coverage should prefer user stories with real closure:
- direct answer returned
- proposal created and confirmable
- run created and inspectable
- structured blocker with recovery path

### Location

| Spec | Coverage |
|------|----------|
| `interface/e2e/specs/missions.spec.ts` | Dashboard load, nav rail, telemetry, mission cards, dark mode |
| `interface/e2e/specs/governance.spec.ts` | Approvals page, policy tab, pending section |
| `interface/e2e/specs/catalogue.spec.ts` | Catalogue page, agent cards, create button |
| `interface/e2e/specs/settings.spec.ts` | Settings page, MCP registry |
| `interface/e2e/specs/layout.spec.ts` | Shell structure, zone rendering |
| `interface/e2e/specs/navigation.spec.ts` | Route transitions, active states |
| `interface/e2e/specs/trust_economy.spec.ts` | Trust slider, threshold updates |
| `interface/e2e/specs/telemetry.spec.ts` | Telemetry dashboard, metric cards |
| `interface/e2e/specs/memory.spec.ts` | Memory explorer, search |
| `interface/e2e/specs/proposals.spec.ts` | Proposal CRUD flow |
| `interface/e2e/specs/teams.spec.ts` | Team management, roster |
| `interface/e2e/specs/wiring-edit.spec.ts` | Neural wiring, agent edit/delete |
| `interface/e2e/specs/v7-operational-ux.spec.ts` | Execution-facing degraded-state, recovery, status drawer, and operator UX checks |
| `interface/e2e/specs/mobile.spec.ts` | Mobile viewport smoke coverage under the dedicated mobile Playwright project |
| `interface/e2e/specs/accessibility.spec.ts` | Axe-backed accessibility baseline for key operator surfaces |
| `interface/e2e/specs/workspace-live-backend.spec.ts` | Real Workspace contract coverage against live `/api/v1/services/status` and `/api/v1/council/members` traffic |

### Configuration

- **Config:** `interface/playwright.config.ts`
- **Base URL:** `http://127.0.0.1:3000` by default (`INTERFACE_HOST` / `INTERFACE_PORT` override supported)
- **Browser Projects:** `chromium`, `firefox`, `webkit`, `mobile-chromium`
- **Server Lifecycle:** Playwright `webServer` starts/stops the Next.js app for local and CI E2E runs
- **Task Cleanup:** `uv run inv interface.e2e` stops any stale listener on `:3000` before and after each run
- **Live Backend Mode:** `uv run inv interface.e2e --live-backend ...` loads proxy auth env and enables specs that require a real Core backend
- **Accessibility Gate:** `@axe-core/playwright` is a required dev dependency; accessibility specs must fail when violated, not skip because the package is missing
- **Dark mode compliance:** Every spec includes `no bg-white` assertion

### Running

```bash
# Core is only required for specs that hit the real backend instead of route stubs.
uv run inv core.run          # Optional: start live backend coverage in a separate terminal

# Run E2E tests
uv run inv interface.e2e                     # All specs
uv run inv interface.e2e --live-backend --spec=e2e/specs/workspace-live-backend.spec.ts
uv run inv interface.e2e --project=firefox
uv run inv interface.e2e --project=mobile-chromium --spec=e2e/specs/mobile.spec.ts
npx playwright test --project=chromium    # From interface/
npx playwright test e2e/specs/missions.spec.ts  # Single spec
npx playwright test e2e/specs/accessibility.spec.ts --project=chromium
npx playwright show-report                # View HTML report
```

---

## Tier 4: Integration Tests (Go)

**Goal:** Verify the real model (Ollama/OpenAI) understands prompts and schemas.
**Speed:** 1s-30s (depends on model).
**Build Tag:** `//go:build integration` (skipped by default `go test`)

### Location

- `core/tests/agent_interaction_test.go`

### Running

```bash
# Requires: Ollama running + model pulled
# ollama pull qwen2.5-coder:7b-instruct

# Windows (PowerShell)
$env:OLLAMA_HOST="http://192.168.50.156:11434"; go test -v -tags=integration ./tests/...

# Linux/Mac
OLLAMA_HOST=http://192.168.50.156:11434 go test -v -tags=integration ./tests/...
```

---

## Tier 5: Governance Smoke Tests (System)

**Goal:** Verify the Gatekeeper blocks dangerous actions and routes approvals correctly.

### Protocol

1. Start the Core: `uv run inv core.run`
2. Inject poison: Send a message with intent `k8s.delete.pod`
3. Verify block: Check logs for "Gatekeeper DENIED"
4. Inject require approval: Send `payment.create` with amount `100`
5. Verify inbox: Check `/approvals` for the pending request

### Running

```bash
uv run inv core.smoke
```

---

## Memory Restart Validation

Use this when memory path behavior is suspect (stale stream, sitrep/read model drift, failed local state recovery).

```bash
uv run inv lifecycle.memory-restart --build --frontend
```

Expected command outcomes:
- stack teardown/restart completes
- forward-only migration set applies cleanly (`001_init_memory.sql` + `*.up.sql`)
- health probe passes
- memory probes return HTTP 200:
  - `/api/v1/memory/stream`
  - `/api/v1/memory/sitreps?limit=1`
- if Core fails before binding `:8081`, inspect `workspace/logs/core-startup.log`

---

## CI Pipelines

Three GitHub Actions workflows enforce quality on every push/PR to `main` and `develop`:

| Workflow | File | What it does |
|----------|------|-------------|
| **Core CI** | `.github/workflows/core-ci.yaml` | Go test with coverage + GolangCI-Lint v1.64.5 + binary build |
| **Interface CI** | `.github/workflows/interface-ci.yaml` | npm lint + `tsc --noEmit` + Vitest + production build |
| **E2E CI** | `.github/workflows/e2e-ci.yaml` | Build Core binary + Next.js, start Core, let Playwright own the UI server, run browser matrix, upload results on failure |

### CI Checks

- **Go:** `go test -v -coverprofile=coverage.out ./...`
- **Go Lint:** GolangCI-Lint v1.64.5
- **TypeScript:** `npx tsc --noEmit` (strict type checking)
- **Frontend Lint:** `npm run lint` (ESLint)
- **Frontend Tests:** `npm run test` (Vitest)
- **E2E:** Playwright browser matrix (`chromium`, `firefox`, `webkit`, `mobile-chromium`) with axe accessibility baseline, test results uploaded as artifact on failure

---

## Adding New Tests

### UI Delivery Test Checklist

Before adding or modifying an execution-facing UI feature, answer these in the test file or PR notes:
1. what user interaction starts the flow?
2. what terminal UI state is expected?
3. what backend transaction proves the UI actually caused the intended effect?
4. what failure state should the user see?
5. what recovery action should remain available?

### Backend Handler Test

1. Create `core/internal/server/<handler>_test.go`
2. Import test helpers: `newTestServer`, `withDB`, `withGuard`, `doRequest`, `assertStatus`, `assertJSON`
3. Build a minimal server with only required dependencies
4. Use `go-sqlmock` for database expectations
5. Test: happy path, validation errors (400), nil guards (503), not found (404)

```go
func TestHandleMyEndpoint(t *testing.T) {
    dbOpt, mock := withDB(t)
    s := newTestServer(dbOpt)

    mock.ExpectQuery(`SELECT .+ FROM my_table`).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("uuid", "test"))

    rr := doRequest(t, s.handleMyEndpoint, "GET", "/my-endpoint", "")
    assertStatus(t, rr, http.StatusOK)
}
```

### Frontend Component Test

1. Create `interface/__tests__/<area>/<Component>.test.tsx`
2. Mock external dependencies (`vi.mock(...)`)
3. Set Zustand state if needed (`useCortexStore.setState({...})`)
4. Use `mockFetch` for API calls

```tsx
import { render, screen } from '@testing-library/react'
import { MyComponent } from '@/components/area/MyComponent'

vi.mock('reactflow', () => import('../mocks/reactflow'))

test('renders content', () => {
    render(<MyComponent />)
    expect(screen.getByText('Expected')).toBeInTheDocument()
})
```

### E2E Spec

1. Create `interface/e2e/specs/<feature>.spec.ts`
2. Use `page.goto()` + `waitForLoadState('domcontentloaded')` for deterministic hydration checks (avoid long-lived stream flake from `networkidle`)
3. Always include dark mode compliance check (`no bg-white`)

```typescript
import { test, expect } from '@playwright/test'

test.describe('Feature Page', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/feature')
        await page.waitForLoadState('domcontentloaded')
    })

    test('page loads without errors', async ({ page }) => {
        await expect(page.locator('text=Feature')).toBeVisible()
    })

    test('no bg-white leak', async ({ page }) => {
        const body = await page.content()
        expect(body).not.toContain('bg-white')
    })
})
```



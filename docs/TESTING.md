# Verification & Testing Protocol

Mycelis employs a **5-Tier Testing Strategy** covering backend handlers, frontend components, end-to-end flows, integration tests, and governance smoke tests.

## Quick Reference

```bash
# Run from scratch/ root — always use uvx inv, never raw commands
uvx inv core.test             # Go unit tests (all packages)
uvx inv interface.test        # Vitest unit tests (jsdom)
uvx inv interface.e2e         # Playwright E2E tests (requires running servers)
uvx inv core.smoke            # Governance smoke tests
uvx inv interface.check       # HTTP smoke test against running dev server
```

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
uvx inv core.test                          # All packages
go test -v ./internal/server/...           # Server handlers only
go test -v -run TestHandleGovernance ./internal/server/...  # Single test pattern
go test -v ./internal/mcp/ -count=1        # MCP service/library/executor/toolset suites
go test -v -run TestHandleMCP ./internal/server/... -count=1
go test -v -run TestHandleUpdateToolSet ./internal/server/... -count=1
go test -v -run TestScoped ./internal/swarm/... -count=1
```

> Note: `go test ./...` currently includes an unrelated root-package conflict (`core/probe.go` and `core/probe_test.go` both declare `main`).

---

## Tier 2: Frontend Unit Tests (Vitest)

**Goal:** Verify component rendering, store interactions, and UI logic in jsdom.
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

### Running

```bash
uvx inv interface.test                     # All Vitest tests
npx vitest run --reporter=verbose          # Verbose output (from interface/)
npx vitest run __tests__/shell/            # Single directory
```

---

## Tier 3: End-to-End Tests (Playwright)

**Goal:** Verify full user journeys through the running application.
**Speed:** 30s-2min (requires running Core + Interface servers).

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

### Configuration

- **Config:** `interface/playwright.config.ts`
- **Base URL:** `http://localhost:3000`
- **Browser:** Chromium (headless in CI)
- **Dark mode compliance:** Every spec includes `no bg-white` assertion

### Running

```bash
# Start servers first
uvx inv core.run          # Terminal 1
uvx inv interface.dev     # Terminal 2

# Run E2E tests
uvx inv interface.e2e                     # All specs
npx playwright test --project=chromium    # From interface/
npx playwright test e2e/specs/missions.spec.ts  # Single spec
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

1. Start the Core: `uvx inv core.run`
2. Inject poison: Send a message with intent `k8s.delete.pod`
3. Verify block: Check logs for "Gatekeeper DENIED"
4. Inject require approval: Send `payment.create` with amount `100`
5. Verify inbox: Check `/approvals` for the pending request

### Running

```bash
uvx inv core.smoke
```

---

## CI Pipelines

Three GitHub Actions workflows enforce quality on every push/PR to `main` and `develop`:

| Workflow | File | What it does |
|----------|------|-------------|
| **Core CI** | `.github/workflows/core-ci.yaml` | Go test with coverage + GolangCI-Lint v1.64.5 + binary build |
| **Interface CI** | `.github/workflows/interface-ci.yaml` | npm lint + `tsc --noEmit` + Vitest + production build |
| **E2E CI** | `.github/workflows/e2e-ci.yaml` | Build Core binary + Next.js, start both servers, run Playwright, upload results on failure |

### CI Checks

- **Go:** `go test -v -coverprofile=coverage.out ./...`
- **Go Lint:** GolangCI-Lint v1.64.5
- **TypeScript:** `npx tsc --noEmit` (strict type checking)
- **Frontend Lint:** `npm run lint` (ESLint)
- **Frontend Tests:** `npm run test` (Vitest)
- **E2E:** Playwright with Chromium, test results uploaded as artifact on failure

---

## Adding New Tests

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
2. Use `page.goto()` + `waitForLoadState('networkidle')`
3. Always include dark mode compliance check (`no bg-white`)

```typescript
import { test, expect } from '@playwright/test'

test.describe('Feature Page', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/feature')
        await page.waitForLoadState('networkidle')
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

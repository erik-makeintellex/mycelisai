# Interface E2E (`interface/e2e`)

> Navigation: [Interface AGENTS.md](../AGENTS.md) | [Repo AGENTS.md](../../AGENTS.md)

## Owns / does not own
- `interface/e2e/specs/**` (81 specs), `interface/e2e/support/**` (fixtures, live-journey helpers), `interface/e2e/global-setup.ts`, `interface/playwright.config.ts` -> `mycelis-e2e-proof` (`.claude/agents/`, local-only and gitignored). Distinct from `mycelis-interface`, which owns product code, not the e2e harness.
- Browser proof is run only under an exclusive lease from the lead; this role adds missing assertions and triages spec failures, it does not casually run the full suite against the shared stack.

## Contracts
- `uv run inv interface.e2e` is a single Playwright owner: it holds a PID/session lease across managed, external and live-backend modes before port selection through final cleanup, so parallel agents fail fast instead of sharing reports or launching duplicate Interface processes. Never run it in parallel with another agent or run.
- Delivered-UI proof targets `http://127.0.0.1:3000` with `--server-mode=external`; never a convenient alternate port.
- Sign-in comes from `.env` via `global-setup.ts`: it reads `MYCELIS_LOCAL_ADMIN_USERNAME`/`MYCELIS_LOCAL_ADMIN_PASSWORD`, and fails fast with a named remedy (`uv run inv auth.dev-key --admin-password=generate|<password>`, then `compose.up`) if the password is empty or does not match `MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256`. `MYCELIS_API_KEY` is not a login password.
- `global-setup.ts` waits on the URL **pathname** (`/dashboard` vs `/login?error=...`), not on `waitForURL` resolving immediately, because the login URL itself already ends in `/dashboard` in its `next` query param; it then asserts the session cookie is actually set before saving storage state. A sign-in regression must fail here, not surface as 100+ tests failing on the login page.
- Label every result by lane: mocked vs live-backend, managed vs external, headed vs headless, project/viewport. Mocked results never certify live behaviour.
- Live fixtures: owner + execution scope + exact claims; purge must assert terminal status, empty warnings and absence of fixture files/work/claims. Never clear shared NATS or the whole DB/workspace.

## Gates
- `uv run inv interface.e2e ...` only under the lead's lease; report exact pass/fail/skip/flake counts and artifact paths, and preserve failure artifacts before rerunning.
- No dedicated Go/Python unit gate; this area is Playwright/TypeScript only.

## Gotchas
- Legacy spec line-cap exceptions (`ops/quality_legacy_caps.txt`): `interface/e2e/specs/mcp-connected-tools.spec.ts` (510), `interface/e2e/specs/docs-and-runs.spec.ts` (388), `interface/e2e/support/groups-workspace.ts` (398) — don't grow them further, split only downward.
- `interface.e2e stop()` kills only repo-local next-server processes whose cwd resolves under the repo checkout (`ops/interface_runtime.py` `_repo_local_interface_processes_for_pids`); it leaves a foreign listener on the same port untouched and prints "Port N is held by non-repo process(es)" instead. Do not assume `stop()` is safe to run against a port you don't own.
- A leaked next-server (one `stop()` didn't catch because its cwd differs from the current checkout, for example a stale linked worktree) is a known harness gap; report the exact port and PID rather than killing it by hand.

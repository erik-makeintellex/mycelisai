# Interface (TypeScript/Next.js)

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Interface README](README.md)

## Owns / does not own
- `interface/app/**`, `interface/components/**`, `interface/store/**`, `interface/lib/**`, `interface/proxy.ts`, `interface/__tests__/**`, `interface/types/**` -> `mycelis-interface` (`.claude/agents/`, local-only and gitignored).
- `interface/e2e/**` and `interface/playwright.config.ts` -> `mycelis-e2e-proof`; not this area's writer. See [`interface/e2e/AGENTS.md`](e2e/AGENTS.md) for its distinct rules.
- `interface/lib/docsManifest.ts` *entries* (which docs are exposed) are proposed by `mycelis-docs-steward`; `mycelis-interface` applies them when asked, not unilaterally.
- `mycelis-ux-reviewer` is read-only here: it advises on PRD human-first/UX fit, it never edits.
- **The `feature/u1-human-first-ui` worktree is the owner's lane and is off limits.** It currently carries a broad, uncommitted rewrite across most of `interface/`. Before touching any interface file, check `git status` and `git diff dev...HEAD` in that worktree and treat every file it changes (committed or not) as off limits; list them explicitly as off limits in your own packet.

## Contracts
- PRD sections "Primary User Experience", "Human-First Composition Contract", "Cross-Device Delivery Contract", "Information Architecture" ([`docs/architecture-library/MYCELIS_CANONICAL_PRD.md`](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md)) define the UI contract.
- Outcome Health labels are exactly `Healthy`/`Waiting`/`Running`/`Degraded`/`Blocked`/`Completed`/`Archived` — no synonyms anywhere in user-facing copy.
- No runtime vocabulary (NATS, run ids, MCP ids, mission, actuation, cortex, swarm, bus) outside Inspect/admin surfaces.
- The browser never orchestrates internal services or decides authority; show normalized error states from the BFF, never raw backend text or stack traces.
- One primary scroll owner per route; overlays never narrow Soma; the composer is always reachable; 44px touch targets; no hover-only actions.

## Gates
- `uv run inv interface.lint`, `uv run inv interface.typecheck`, `uv run inv interface.test` (from a checkout with `interface/node_modules` present), plus the affected Vitest files under `interface/__tests__/**`.
- Browser/Playwright proof is requested from `mycelis-e2e-proof` or the lead; this area's writer does not hold the Playwright lease.

## Gotchas
- Files target 350 lines, hard cap 385; several legacy files already exceed it under `ops/quality_legacy_caps.txt` (for example `interface/components/organizations/OrganizationContextShell.tsx` at 2358, `interface/components/settings/BrainsPage.tsx` at 552) — split legacy oversized files only when you are already changing them, and only downward.
- Linked worktrees created before P3 (`723f76d6`) can be missing `interface/node_modules` and `.env`; they default to the main checkout's tool cache (`MYCELIS_PROJECT_CACHE_ROOT`) but may still need their own `npm install`.
- `interface.e2e` is a single-owner PID/session lease across managed, external and live-backend modes — never run it from here in parallel with another agent.

# Mycelis development team

Tracked Claude Code agent roster. The repository rules in `AGENTS.md` win on any conflict. Product truth lives in the canonical PRD (`docs/architecture-library/MYCELIS_CANONICAL_PRD.md`); live status, the current plan and delivery findings live in `.state/V8_DEV_STATE.md`, starting with its Resume Guide and "Delivery Targets And Teams". This README is an index, not an authority document: it carries no plan or findings of its own.

Writers follow [`WRITER_BRIEF.md`](WRITER_BRIEF.md). The lead gives each writer a transient packet (a scratch file, not tracked); the outcome is recorded in the scoreboard.

## Roster

The **lead** is the main session. It is the only agent that changes Git topology, integrates, runs lifecycle/Compose/Playwright proofs (or leases them) and edits the scoreboard/PRD. The owner runs `git push`.

| Agent | Aspect | Writes | Model |
| --- | --- | --- | --- |
| mycelis-architect | Contract freeze, architecture QA | nothing (read-only) | opus |
| mycelis-core-authority | Identity, authn/authz, RBAC, invocation ledger, C2 admission | Go identity/auth/invocation | opus |
| mycelis-core-execution | Dispatch/outbox, binding, supervisor, team-work, output finalization | Go execution spine, framework-runs, `proto/` | opus |
| mycelis-schema | Current-schema DDL, installer/integrity tests | `core/migrations` (single writer) | opus |
| mycelis-interface | UI, stores, BFF/session routes, Vitest | `interface/` except e2e | sonnet |
| mycelis-ux-reviewer | Human-first contract, vocabulary, layout review | nothing (read-only) | sonnet |
| mycelis-e2e-proof | Playwright specs/fixtures; runs browser proof under lease | `interface/e2e` | sonnet |
| mycelis-platform-ops | Invoke tasks, Compose, Helm, CI, topology config | `ops/`, `deploy/`, `charts/`, `sdk/`, CI | sonnet |
| mycelis-ai-runtime | Model/provider posture, cognitive config, capability evals | `core/config`, cognitive, evals | sonnet |
| mycelis-security-qa | Independent adversarial QA, evidence review | nothing (read-only) | opus |
| mycelis-docs-steward | Docs sync, README TOC, docs manifest, docs gates | `README.md`, `docs/` (not PRD) | sonnet |

Model choice follows the tiers in `AGENTS.md` ("Development Model Routing" and "Context Execution"):
- **T3** (architecture, security QA): architect, security-qa. Opus / `gpt-6-astra`.
- **T2** (hard coding): core-authority, core-execution, schema. Opus / `gpt-6-sol` high.
- **T1** (routine): interface, e2e-proof, platform-ops, ai-runtime, docs-steward, ux-reviewer. Sonnet / `gpt-6-sol` medium.
- **T0** (inventory helpers): spawn ad hoc on Haiku / `gpt-6-luna`, or the local model for summarizing supplied text.
- A review-heavy architect task that is mostly inventory runs on T1 (Sonnet); keep Opus for the decision itself.

## Slice protocol (every slice)

1. **Lead: execution shape.** Check the clean tree, `dev` state and scoreboard row. Write the owners, owned paths and proof gate. Create one `feature/*` branch, plus a separate worktree when slices run in parallel.
2. **Architect: contract packet.** Freezes scope, non-goals, invariants, negative cases and the proof gate. Required before any authority, identity, execution or schema code.
3. **Schema (if needed).** The single writer lands the DDL and integrity tests first, then the lead runs isolated first boot.
4. **Writers implement.** Only in their owned paths, test-first, with the smallest focused tests.
5. **Review in parallel.** The UX reviewer covers UI slices, e2e-proof covers browser evidence (exclusive lease), and security-qa covers authority, session, execution and deployment slices.
6. **Docs.** Docs-steward syncs the owning docs. The lead updates the scoreboard.
7. **Lead: integrate.** Commit the proven state, merge to `dev`, rerun the affected gates on merged `dev`, retire the branch/worktree, then close the agents.

Parallelism rule: run slices in parallel only when their writers' paths are disjoint. Schema changes, Playwright runs and Compose lifecycle actions are always serialized through the lead.

## Delivery-target teams

Canonical text: `AGENTS.md` "Team Orchestration And Messaging Contract" (owner rule, 2026-09-28).
- At most **3 agents live per delivery target**. The lead is not counted.
- Before spawning, declare the target's team (at most 3 roles from this roster, with tiers) in the "Delivery Targets And Teams" table in `.state/V8_DEV_STATE.md`.
- Continuations reuse that row's team: resume a live agent that holds the context, or respawn the same role with a fresh packet.
- Close a row by listing the agents actually used.

## Shared services

The shared stack is the `mycelis-home-*` Compose services in WSL Docker plus the Windows Ollama root reached through the relay. `vllm-node` is optional and stopped. Agents never start, stop or recreate these unless the lead leases it in writing.

# Architecture Library

> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

This library holds the single canonical product/architecture authority plus a small number of scoped, bounded supporting docs. It does not hold historical, versioned, or superseded architecture files; those are deleted, not archived, and Git history preserves them.

- [Mycelis Canonical PRD](MYCELIS_CANONICAL_PRD.md) — the single source of product, architecture, UX, runtime, MVP, and release-gate truth. Everything else in this library supports it and must not restate or fork it.
- [Market Position And Workflow Targets](MARKET_POSITION_AND_WORKFLOW_TARGETS.md) — competitive positioning against Paperclip (what Mycelis leads on, what to borrow, the interop option) and the 12-workflow modern-UX target model with current scores and top-5 changes.
- [Truthful Delivery And Next Architecture](TRUTHFUL_DELIVERY_AND_NEXT_ARCHITECTURE.md) — the truthfulness invariant, placeholder-audit status (fixed vs remaining), the live journey probe as the delivery metric, and the planned M1/H1/B1 architecture with owner decisions recorded.
- [G4/E10 Invocation Contract](G4_E10_DURABLE_INVOCATION.md) — the bounded counting-capability durable-invocation acceptance contract.
- [Post-G4 Delivery And GUI Plan](POST_G4_DELIVERY_AND_GUI_PLAN.md) — sequencing for Compose browser review, B2 deployment, and later framework gates.

Live implementation truth is tracked separately in [`.state/V8_DEV_STATE.md`](../../.state/V8_DEV_STATE.md), not in this library.
- [B1 Token Budgets Contract](B1_TOKEN_BUDGETS_CONTRACT.md) — per-execution token budgets keyed to model defaults (local 14B: 64k per execution, 256k per run, 2M per team per day), admin overrides, hard stops with `token_budget_exhausted`, and the usage ledger.
- [MCPS Operator MCP Scope Contract](MCPS_OPERATOR_MCP_SCOPE_CONTRACT.md) — authority, resolution, audit and argument redaction for the operator-direct MCP tool call route (draft; owner answers pending).

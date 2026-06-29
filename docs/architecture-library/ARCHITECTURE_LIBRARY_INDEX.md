# Architecture Docs Index
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical
> Last Updated: 2026-06-28
> Purpose: Keep architecture authority singular and current.

## TOC

- [Canonical Architecture](#canonical-architecture)
- [Supporting Docs](#supporting-docs)
- [Cleanup Rule](#cleanup-rule)

## Canonical Architecture

The active architecture and product source of truth is:

- [Mycelis Canonical PRD](MYCELIS_CANONICAL_PRD.md)

This PRD owns product thesis, target UX, runtime architecture, governance, outcomes, capabilities, recovery, MVP scope, P0 delivery plan, and release gates.

## Supporting Docs

Supporting docs are allowed when they serve implementation or user operation without becoming parallel doctrine:

- [Docs Home](../README.md)
- [Operations](../architecture/OPERATIONS.md)
- [Backend](../architecture/BACKEND.md)
- [Frontend](../architecture/FRONTEND.md)
- [Architecture Overview](../architecture/OVERVIEW.md)
- [API Reference](../API_REFERENCE.md)
- [Testing](../TESTING.md)
- [Worker Library / Hermes-Compatible Execution Source Map](WORKER_LIBRARY_SOURCE_MAP.md)
- [User Docs Home](../user/README.md)
- [Active Development State](../../.state/V8_DEV_STATE.md)

## Cleanup Rule

Do not restore split V7, V8.2, or V8.3 architecture documents. If an old document contained current truth, promote the requirement into [Mycelis Canonical PRD](MYCELIS_CANONICAL_PRD.md). Otherwise, leave it in Git history.

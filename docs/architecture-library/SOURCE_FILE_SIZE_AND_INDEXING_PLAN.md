# Source File Size And Indexing Plan
> Navigation: [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [Project README](../../README.md)

> Status: ACTIVE
> Last Updated: 2026-04-30
> Purpose: Make the 300-line source/documentation limit executable without hiding the existing legacy backlog.

## Rule

No source, test, documentation, chart, or configuration file should grow beyond 300 lines.

When a file approaches the limit, split by ownership:
- route/index file for public entrypoints and registration
- focused implementation modules for behavior
- focused test files for one contract family
- focused docs for one durable concept
- generated files are allowed only when generation is documented and the source-of-truth file remains small

## Current Enforcement

`uv run inv quality.max-lines --limit 300` is the source-code ratchet. It reports legacy oversize files that remain inside their recorded cap and fails when a file grows past its cap or a new uncapped file exceeds 300 lines.

This keeps releases moving while preventing new growth. Do not raise a legacy cap as a convenience fix. Split the file instead.

## Documentation Indexing

Canonical docs belong in `docs/architecture-library/` when they define durable target, delivery, governance, UI, execution, or runtime meaning.

Use index pages instead of monoliths:
- repo orientation: `README.md`
- architecture authority: `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
- live implementation state: `.state/V8_DEV_STATE.md`
- historical material: `docs/archive/README.md`

Docs meant for user-facing `/docs` must also be registered in `interface/lib/docsManifest.ts`.

## Priority Backlog

The largest active over-300 files found in this review are:
- `interface/components/organizations/OrganizationContextShell.tsx`
- `core/internal/server/organizations.go`
- `core/internal/server/cognitive.go`
- `interface/components/dashboard/MissionControlChat.tsx`
- `tests/test_docs_links.py`
- `.state/V7_DEV_STATE.md`
- `.state/V8_DEV_STATE.md`
- `docs/WORKFLOWS.md`
- `docs/architecture/OPERATIONS.md`
- `README.md`

Treat these as decomposition targets, not as examples to copy.

## Decomposition Order

1. Split user-facing UI components first, starting with organization and mission chat surfaces.
2. Split server handlers by route family and service boundary.
3. Split large test files by contract area and move shared fixtures into small helpers.
4. Split docs into index plus bounded concept pages.
5. Replace state-file growth with dated state entries or smaller linked state records when the live scoreboard crosses 300 lines.

## Acceptance

A cleanup slice is complete only when:
- no touched file exceeds 300 lines unless it was already legacy-capped and did not grow
- new files are under 300 lines
- indexes link to any new canonical docs
- `uv run inv quality.max-lines --limit 300` passes
- the final report names remaining legacy oversize files relevant to the touched area

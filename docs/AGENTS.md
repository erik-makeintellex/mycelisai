# Docs

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Docs Home](README.md)

## Owns / does not own
- `docs/**` except `docs/architecture-library/MYCELIS_CANONICAL_PRD.md` and `.state/V8_DEV_STATE.md`, `ops/README.md`, and docs entries in `interface/lib/docsManifest.ts` -> `mycelis-docs-steward` (`.claude/agents/`, local-only and gitignored).
- `MYCELIS_CANONICAL_PRD.md` and `.state/V8_DEV_STATE.md` are lead-owned; docs-steward edits only the exact text the lead or `mycelis-architect` hands over, never independently.
- `docs/TESTING.md` and `docs/architecture/OPERATIONS.md` are at their line cap (the "Task runner tightening" target is CLOSED); edits to them must be net-zero or shrink.
- `architecture/` (the user-shared root-level entrypoint, not under `docs/`) follows the same docs-steward ownership per root `AGENTS.md` "Canonical Docs Location".

## Contracts
- [`docs/architecture-library/MYCELIS_CANONICAL_PRD.md`](architecture-library/MYCELIS_CANONICAL_PRD.md) is the single product/architecture/UX/runtime/MVP/release-gate authority; do not create a second one.
- `docs/architecture-library/` holds exactly nine files today, enforced by `tests/test_docs_links.py::test_product_architecture_library_has_one_prd_and_scoped_supporting_docs`: adding or removing a file there requires updating that test's expected list in the same change.
- Document behavior that exists, with its limits; never convert fixture/mocked proof into a live claim. Remove superseded text rather than archiving it — one authority per topic (root `AGENTS.md` "Documentation Synchronization Contract").
- User-facing docs under `docs/user/*` must all appear in `interface/lib/docsManifest.ts` (`tests/test_docs_links.py::test_all_user_docs_are_exposed_in_help_manifest`).

## Gates
- `uv run --no-sync pytest -q tests/test_docs_links.py tests/test_canonical_workspace_docs.py tests/test_trusted_outcome_docs.py`, plus `tests/test_documentation_layout_contract.py` for any doc-location change.
- `git diff --check` for trailing-whitespace/EOL issues.
- Every link in a file this test set covers must resolve on disk: `test_all_active_documentation_links_resolve` walks all of `docs/**/*.md` and `architecture/**/*.md` plus a fixed list of top-level READMEs and every area `AGENTS.md` (`_area_agents_files()` in `tests/test_docs_links.py` and `tests/test_tasks_root.py`), so keep every link and `uv run inv` task name in an `AGENTS.md` resolvable.

## Gotchas
- Stale V7/V8.2/V8.3 architecture docs must stay deleted, not archived (`tests/test_docs_links.py::test_old_architecture_docs_are_deleted_not_archived_or_exposed` names each forbidden path).
- `.state/V8_DEV_STATE.md` and `MYCELIS_CANONICAL_PRD.md` may never be loose at repo root; they must stay at their exact tracked locations (`tests/test_documentation_layout_contract.py`).
- README-style pages (`README.md`, `docs/README.md`, `architecture/README.md`, `ops/README.md`, `core/README.md`, `interface/README.md`, `core/internal/registry/README.md`) each need their own TOC heading and, except the root, a "Project README"/"Navigation:" line at the top, enforced by `tests/test_docs_links.py::test_readme_style_pages_expose_project_navigation_and_tocs`. `AGENTS.md` files are not in that test's `required` set, so they are exempt from it, but are not separately exempted by name.

# MYC-RC-003 Proof Envelope Delivery Plan
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8.3 Convergence PRD](mycelis_soma_updated_prd_v8_3_current_state_convergence.md)

> Status: `ACTIVE`
> Date: 2026-05-29
> Purpose: deliver the minimum universal proof envelope for retained outputs, beginning with Codex task packages.

## Goal

`MYC-RC-003` makes retained outputs inspectable as proof-backed results, not just files, cards, or event payloads.

The first implementation slice is intentionally narrow:

1. Codex package files remain workspace files.
2. Each file gets an artifact-backed or artifact-compatible identity.
3. Each `OutputRef` carries enough proof metadata for readback.
4. The UI can show proof status, checksum, storage location, and source run/contract without exposing raw internals by default.

## Solution Architecture

### ProofEnvelope

The minimum proof envelope is a projection attached to an output. It does not need a new table in the first slice if existing artifact/event storage can carry the fields.

Minimum fields:

- `proof_id`
- `output_ref_id`
- `storage_ref`
- `source_run_id`
- `source_contract_id`
- `execution_status`
- `path_boundary_status`
- `checksum`
- `checksum_algorithm`
- `bytes`
- `content_type`
- `created_at`
- `readback_status`
- `recovery_hint`

Accepted statuses:

- `verified`
- `degraded`
- `missing`
- `blocked`

### Backend Projection

The backend must construct proof envelopes where the output is created, before the response and event payload are emitted.

Initial owner files:

- `core/pkg/protocol/execution_summary.go`
- `core/internal/server/templates_codex_package.go`
- `core/internal/server/confirm_action_outputs.go`
- `core/internal/server/templates_confirm_action.go`
- `core/internal/server/templates_codex_package_test.go`
- `core/internal/artifacts/service.go`

First-slice behavior:

- Codex package finalization records checksum, byte count, content type, path boundary result, source run, and source contract for every package file.
- Package files are persisted through the artifact service inside the confirm-action transaction, so proof rows commit or roll back with the mission run and proof status.
- `OutputRef` includes the persisted `artifact_id` for each package file.
- `OutputRef` includes `open_url` when a stable artifact/read URL exists.
- `OutputRef` includes the real proof artifact ID for every non-proof package item.
- The proof file gets role `proof` and links back to the package/run/contract.
- `artifact.created` event payload includes proof envelope fields or a nested `proof_envelopes` list keyed by `output_ref_id`.

### UI Projection

The UI should keep the current package-first workbench but add inspectable proof affordances.

Initial owner files:

- `interface/store/cortexStoreTypesChat.ts`
- `interface/store/cortexStoreOutputRefs.ts`
- `interface/components/soma/OutputWorkbench.tsx`
- `interface/components/dashboard/MissionControlMessageBubble.tsx`
- `interface/__tests__/dashboard/MissionControlChat.package-output.test.tsx`
- `interface/__tests__/store/useCortexStore.confirm-proposal.execution.test.ts`
- `interface/e2e/specs/codex-package-retention.spec.ts`

First-slice behavior:

- Package workbench rows show proof status and checksum presence in compact form.
- Inspect details show storage ref, checksum, source run, source contract, proof artifact, and readback status.
- Existing artifact/package compatibility stays intact until all output types use `OutputRef`.

## Delivery Team

| Team | Status | Owner Scope | Output |
| --- | --- | --- | --- |
| Backend Proof Team | `ACTIVE` | Package finalization, proof envelope construction, artifact/readback linkage | Protocol structs, backend projection, handler tests |
| UI Proof Team | `ACTIVE` | Output Workbench proof readback and store normalization | UI proof detail rendering, store tests, package card test |
| Validation Team | `ACTIVE` | Windows-managed proof lane while WSL is blocked | Focused and full Chromium proof, docs/state evidence |
| Docs / State Team | `ACTIVE` | Keep convergence, testing, README/index, and state synchronized | Updated docs and explicit close-out evidence |

## Delivery Sequence

1. `ACTIVE` finalize proof envelope contract and owner boundaries.
2. `NEXT` implement backend proof envelope projection for Codex package files.
   - add transactional artifact storage, such as `StoreTx`
   - persist the five Codex package files as artifact rows
   - carry artifact IDs into `packageOutputFile`, `OutputRef`, and package event payloads
3. `NEXT` add UI proof readback in Output Workbench.
   - preserve explicit backend proof fields
   - infer degraded legacy proof only for older package messages
   - show checksum, boundary, readback, source run/contract, and open/download affordances
4. `NEXT` extend tests to assert checksum, path boundary, source run/contract, output ref, proof artifact, and readback status.
5. `NEXT` run Windows-managed proof:
   - `uv run inv core.test`
   - `uv run inv interface.test`
   - `uv run inv interface.typecheck`
   - `uv run inv interface.build`
   - `uv run inv interface.e2e --project=chromium --spec=e2e/specs/codex-package-retention.spec.ts`
   - `uv run inv interface.e2e --project=chromium --workers=1`
   - `uv run pytest tests/test_docs_links.py tests/test_documentation_layout_contract.py -q`
   - `uv run inv quality.max-lines --limit 300`
   - `git diff --check`

## Acceptance

`MYC-RC-003A` is accepted when:

- every Codex package file has a proof envelope in the confirm-action response or event payload
- every non-proof output ref links to proof
- checksum and byte count are asserted by backend tests
- path-boundary verification is explicit and test-covered
- the package workbench exposes proof details without requiring raw JSON
- Windows-managed Chromium package retention proof and full Chromium proof are green from committed state

## Explicit Non-Goals

- Do not add a broad recovery table in this slice.
- Do not require WSL proof while the local WSL/Docker lane remains unstable.
- Do not convert every media, MCP, team, and template output before Codex package proof is accepted.
- Do not remove legacy `artifacts` or `package` compatibility fields yet.

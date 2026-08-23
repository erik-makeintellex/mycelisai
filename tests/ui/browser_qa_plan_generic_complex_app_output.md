# Browser QA Test Plan: Generic Complex App Output

## Application
Mycelis Soma workspace and retained output surfaces:
- `/dashboard`
- `/organizations/[id]`
- `/resources?tab=workspace`
- `/groups?panel=outputs`
- `/runs/[id]`

## Purpose

Turn the failed voxel/game proof into a reusable live GUI acceptance pass for any complex app-like output. The proof should validate the operator journey, not a particular genre or asset set:

Soma asks and shapes the outcome, the operator approves async team work, the candidate reaches review, the retained output opens from the UI, browser interaction changes visible state, and any failure appears as one clear recovery path.

## Existing Playwright Coverage To Reuse

- `interface/e2e/specs/trusted-outcome-journey-live.spec.ts` -> `Trusted Outcome Journey live smoke > proves the source-stack Ask to Revisit path with durable proof readback`
- `interface/e2e/specs/soma-natural-delivery-routing-live.spec.ts` -> `Natural Soma delivery routing > turns an application outcome ask into governed team delivery`
- `interface/e2e/specs/soma-browser-game-p0-live.spec.ts` -> `Live Soma P0 browser game delivery > creates and opens a playable console-era browser game through Soma`
- `interface/e2e/specs/team-output-content-live.spec.ts` -> `Live teams produce reviewable content outputs > creates a team through Soma and verifies its playable generated output in the GUI`
- `interface/e2e/specs/resources-workspace-files.spec.ts` -> `Resources workspace files > lets an operator browse, read, create, and write through filesystem MCP`
- `interface/e2e/specs/resources-workspace-files.spec.ts` -> `Resources workspace files > writes a retained workspace output through the live filesystem MCP`
- `interface/e2e/specs/soma-worker-profile-lineage-live.spec.ts` -> `Live Soma Worker Profile lineage > pins A, B, and rolled-back A to exact runtime teams`
- `interface/e2e/specs/active-work-api.spec.ts` -> `Active work TeamWorkItem API contract > live API records bounded ask output or degradation as durable active work`
- `interface/e2e/specs/active-work-ask-live.spec.ts` -> `Active work Ask Team live GUI proof > submits a bounded ask from /teams without blocking the operator`
- `interface/e2e/specs/ui-finalization-browser-package-retry.spec.ts` -> `UI finalization first-demo degraded retry proof > mocked package failure preserves failed proof, then retry retains package metadata`
- `interface/scripts/validate-generated-output.mjs` validates loaded app output with `load`, `no_page_errors`, `no_failed_local_assets`, and an interaction probe such as `click`, `key_press`, or `key_hold`.

## Live Acceptance Prompt

Use a generic app request, not voxel/game-specific wording:

```text
Build a small self-contained browser application package that helps an operator review a complex result. It must open from Mycelis, include clear visible state, include at least one documented control, and visibly change state after interaction. Return a retained project_package with entrypoint, folder, files, validation, and proof. If validation fails, report the trusted failure and one safe recovery action instead of claiming completion.
```

## Pass Gates

### 1. Ask And Approval

- Soma returns `mode=proposal`.
- Proposal names the expected retained `project_package`.
- Approval is conversational (`approve` or equivalent), not a hidden out-of-band action.
- Confirm response for async team delivery is running, not falsely verified.
- The UI remains usable after approval; the Soma composer is enabled while team work runs.

### 2. Candidate Review State

- A correlated `TeamWorkItem` exists for the returned `run_id`.
- Acceptable intermediate state includes `reviewing`.
- Terminal success is `output_ready`.
- Terminal failure may be `degraded` or `needs_operator` only when the UI shows the cause and one safe next action.
- Raw internal output, stack noise, or silent timeout is a blocker.

### 3. Retained Output Package

- Output ref kind is `project_package`.
- Output includes `entrypoint`, `folder`, file list, validation notes, and proof ref.
- Required support files include `README.md`, `PROOF.md`, and `project-package.json` when the runtime path supports package metadata.
- The package appears in Soma review, Resources workspace browsing, Groups outputs, and the Run receipt.

### 4. UI Open And Browser Interaction

- The output opens from a visible Mycelis control such as `Open app` or `Open file`.
- The opened page loads without page errors.
- Local/same-origin assets do not fail.
- A visible documented control exists.
- Interaction changes observable state: visual screenshot hash, text, value, or URL.
- For canvas-heavy apps, compare before/after canvas or body screenshots rather than relying on DOM text only.

### 5. Failure And Recovery

- Failed validation must not produce a trusted success card.
- UI names what failed, what remains trusted, what proof is invalid, and the safe next action.
- Retry must preserve failed-run proof and then retain package metadata on success.
- If a live forced failure is unavailable, use the mocked retry spec as regression coverage and record that live failure injection remains unproven.

## Evidence To Capture

- Screenshot of Soma proposal and approval instruction.
- Confirm response summary: `run_id`, `execution_state`, `run_status`, `verified`, and `dispatch_state` if present.
- Team work readback: `work_item_id`, `team_id`, `run_id`, `state`, `degradation_state`, `output_refs`, `proof_refs`, and recovery options.
- Screenshot of the Soma review lane with the latest output.
- Screenshot or video of the opened output before and after interaction.
- Validator report from `interface/scripts/validate-generated-output.mjs` when available.
- Screenshots or API readbacks from Resources, Groups outputs, and Run receipt.
- Cleanup result for the QA organization, fixture scope, team, run, and retained workspace folder.

## Suggested Focused Commands

```powershell
uv run inv interface.e2e --server-mode=start --project=chromium --workers=1 --grep "Trusted Outcome Journey live smoke"
uv run inv interface.e2e --server-mode=start --project=chromium --workers=1 --grep "Natural Soma delivery routing"
uv run inv interface.e2e --server-mode=start --project=chromium --workers=1 --grep "Active work Ask Team live GUI proof"
uv run inv interface.e2e --server-mode=start --project=chromium --workers=1 --grep "Resources workspace files"
```

## Next Proof Order After Runtime And UX Fixes

The latest retained Playwright result is a passing `groups.spec.ts` shard, so treat it as partial retained-output UI coverage only. It is not the generic complex-app acceptance proof.

Run from a clean local source stack:

```powershell
git status --short --branch
uv run inv lifecycle.down
uv run inv compose.infra-up
uv run inv compose.infra-health
uv run inv db.migrate
uv run inv core.compile
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
$env:PLAYWRIGHT_LIVE_BACKEND="1"
$env:PLAYWRIGHT_TEAM_WORK_GUI_LIVE="1"
$env:PLAYWRIGHT_TEAM_WORK_API="1"
```

Then run proof serially and stop at the first unexpected failure:

```powershell
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-natural-delivery-routing-live.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/trusted-outcome-journey-live.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-browser-game-p0-live.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/team-output-content-live.spec.ts
uv run inv interface.e2e --headed --server-mode=start --project=chromium --workers=1 --spec=e2e/specs/ui-finalization-browser-package-retry.spec.ts
uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/active-work-api.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/active-work-ask-live.spec.ts
```

Expected states:
- Natural delivery and Trusted Outcome must pass with proposal approval returning `execution_state=running`, the composer enabled during work, a correlated `TeamWorkItem`, terminal `output_ready`, retained `project_package`, and visible state change after opening the app. `degraded`, `needs_operator`, `result_contract_unsatisfied`, or `runtime_validation_deadline_exceeded` is an honest failure boundary but not an acceptance pass for this fix.
- P0 browser game is the complex-output stress gate. It must prove nonblank generated graphics, controls, movement/interaction, support files, and package metadata; a blank canvas or unchanged visual signature blocks release.
- Team output content must prove retained output visibility from Dashboard, Groups, Resources, opened file, folder reveal, and run-conversation proof.
- Mocked retry must pass unless a live forced-failure spec replaces it; until then, record live failure injection as unproven and retain this mocked regression.
- Active-work API may pass with `output_ready` or an expected bounded degradation (`nats_offline`, `team_response_timeout`, or `team_response_unreadable`) that sets `needs_operator=true`; the GUI ask proof must show queued/running continuity, output-ready review, or degraded recovery without blocking the operator.

Inspect artifacts after each command:
- `interface/test-results/playwright-results.json`, `interface/test-results/playwright-results.xml`, and `.last-run.json`.
- Per-spec folders under `interface/test-results/`, including screenshots, traces, attached retained package evidence, and any before/after interaction captures.
- Retained package files referenced by the run: `index.html`, `README.md`, `PROOF.md`, `project-package.json`, and `validation-report.json` when emitted by `interface/scripts/validate-generated-output.mjs`.
- API readbacks for the accepted `run_id`: TeamWorkItem, status events, proof artifacts, execution contract, run events/receipt, Groups outputs, and Resources workspace folder.

Cleanup requirements:
- Prefer specs with `createQAFixtureScope` and verify their `purgeDeliveryFixture` cleanup completed for the owned organization, team, run, group, and workspace folder.
- For specs without fixture ownership, remove only the generated folder or entrypoint named by that spec; do not delete unrelated retained workspace files.
- Before rerunning after a failure, clear stale Soma/team context for the failed QA organization so older successful output cannot satisfy the new ask.

Required environment depends on the selected live lane:
- `PLAYWRIGHT_LIVE_BACKEND=1` for Soma/Core live delivery specs.
- `PLAYWRIGHT_TEAM_WORK_GUI_LIVE=1` for `/teams` ask GUI proof.
- `PLAYWRIGHT_TEAM_WORK_API=1` or `PLAYWRIGHT_ACTIVE_WORK_API_LIVE=1` for active-work API degradation proof.
- `MYCELIS_API_KEY` when the live Core requires authenticated API access.

## Cleanup

- Prefer specs that create a QA fixture scope through `createQAFixtureScope` and clean it with `purgeDeliveryFixture`.
- Remove retained generated folders when a spec does not own fixture cleanup.
- Clear stale Soma/team runtime context before a fresh UX proof if prior failed output appears in the workspace.
- Do not delete unrelated local workspace files or retained user outputs.

## Release Blockers

- Soma reports completion before the app exists or before validation proof exists.
- Candidate never reaches `reviewing`, `output_ready`, `degraded`, or `needs_operator` with a readable explanation.
- Output cannot be opened from the UI.
- Browser interaction does not change visible state.
- Failure UI lacks one safe recovery action.
- Resources, Groups, or Run receipt cannot revisit the retained output.

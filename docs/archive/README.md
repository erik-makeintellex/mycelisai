# Archive Documentation

This folder contains historical plans and previous-state documents.

Archive docs are:
1. reference-only historical context
2. not implementation authority
3. potentially outdated against the active V8 migration model

Active implementation sources are:
1. `README.md`
2. `V8_DEV_STATE.md`
3. `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
4. `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
5. `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
6. `docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md`
7. `docs/architecture/OPERATIONS.md`

Migration rule:
- V7 assets may still be useful as migration inputs, but archived docs should not shape active implementation unless they are intentionally re-promoted through the current V8 documentation chain.

Recently archived (superseded by current execution and migration docs):
1. `ia-v7-step-01.md`
2. `v7-step-01-ui.md`
3. `UI_OPTIMAL_WORKFLOW_PRDS_V7.md`
4. `UI_OPTIMAL_ENGAGEMENT_ACTUATION_REVIEW_V7.md`
5. `V7_IMPLEMENTATION_PLAN_2026-03-07.md`
6. purged UI lane boards, execution boards, and extension-of-self planning docs now replaced by the architecture library

Draft archive rule:
- superseded root-level notes or draft PRDs should move under `docs/archive/drafts/` with a superseded header instead of remaining loose at repo root

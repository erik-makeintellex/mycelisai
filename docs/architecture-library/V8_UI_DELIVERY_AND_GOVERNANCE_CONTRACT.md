# V8 UI Delivery And Governance Contract
> Navigation: [V8 UI/API Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) | [Testing](../TESTING.md)

Status: canonical UI delivery/governance detail.

## Governance UX

Protected or mutating actions must:
- show proposal state
- explain what will happen
- allow cancel
- execute only after approval or policy allow
- show execution result or blocker
- leave audit/activity evidence

## Output Delivery

Outputs must be:
- inline when small and safe
- artifact-linked when file/media sized
- retained after refresh/reload when persistence is claimed
- clear about missing media/provider blockers

## Error And Recovery

UI errors should be normalized:
- no raw stack/provider noise
- visible retry or recovery guidance
- preserved user context where possible
- activity/run evidence when tied to execution

## Proof

Minimum proof for governance or delivery changes:

```bash
uv run inv interface.test
uv run inv interface.typecheck
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-ui-testing-agentry.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts
```

Use focused tests when only a smaller route or component changed, then rerun the owning gate before close-out.

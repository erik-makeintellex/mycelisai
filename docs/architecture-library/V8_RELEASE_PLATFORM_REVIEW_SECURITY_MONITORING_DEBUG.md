# V8 Release Platform Review: Security, Monitoring, and Debug

> Status: Canonical
> Last Updated: 2026-04-04
> Purpose: keep the release platform review for security, monitoring, and debug/release proof on one shared channel.

## Release Posture

- `IN_REVIEW` the current release platform is operationally coherent after a fresh-cluster reset and live governed browser proof.
- `COMPLETE` fresh-cluster validation is green for local Kubernetes reset, local stack bring-up, deep health probes, and the governed live browser contract.
- `IN_REVIEW` documentation and code commentary are being aligned so operators and future engineers read the same release truth.

## Shared Platform Truth

The release platform is now defined by three surfaces that must agree:

1. Security/governance truth
   - proposal-first mutation
   - approval posture derived from user governance profile plus capability risk
   - audit lineage for proposal, confirm/cancel, execution, capability use, artifacts, and channel writes
2. Monitoring/operations truth
   - `lifecycle.status` for operator posture
   - `lifecycle.health` for authenticated endpoint proof
   - `db.migrate` as forward bootstrap, not replay-on-every-run
   - `k8s.reset` / `k8s.bridge` / `k8s.status` as the local-cluster recovery path
3. Debug/release-proof truth
   - the governed live browser contract must prove the real `/api/v1/chat` and `/api/v1/intent/confirm-action` flow
   - filesystem side-effect checks must bind to the backend workspace actually used by Core
   - browser proof should reuse managed task surfaces rather than a custom one-off harness

## Team Review

### Security Team

What is strong:
- governed mutation requests enter proposal mode before execution
- approval metadata is visible to the UI and tied to capability risk
- audit records now exist for the main proposal/confirm/execute lineage
- the inspect-only Activity Log keeps governance visible without dumping raw backend logs into the default workflow

What needed alignment:
- legacy governance docs still over-emphasized trust-score-only behavior instead of the current user-governance-profile plus approval-policy model
- release docs needed one shared place to say that the current target is constrained, inspectable autonomy rather than trustless auto-execution

Residual release risk:
- the free-node release is enterprise-capable in governance posture, not full enterprise IAM
- multi-user delegated approvers and stronger policy administration remain future work

### Monitoring Team

What is strong:
- `lifecycle.status` and `lifecycle.health` now expose distinct posture-vs-proof surfaces
- `k8s.reset` is a valid fresh-cluster recovery path and was proven in the latest release review pass
- `db.migrate` now behaves like a bootstrap helper instead of replaying a partially non-idempotent schema stack every run

What needed alignment:
- release docs had to be explicit that `db.migrate` is a forward-bootstrap tool and that `db.reset` is the clean rebuild path
- operators need the discipline note that startup/bootstrap tasks should be run sequentially when proving a fresh stack rather than racing lifecycle and migration work in parallel

Residual release risk:
- the bridge layer is still the flakiest local surface on Windows, even though the cluster recovery path is now good
- `lifecycle.status` and `lifecycle.health` are complementary; neither should be treated as a substitute for the other during release proof

### Debug / Release Platform Team

What is strong:
- the live governed browser spec is now green against a freshly reset cluster
- the debug path now uses the same managed `interface.e2e` surface as the release proof
- the browser spec can target the backend workspace explicitly when the spec checkout and Core checkout differ

What needed alignment:
- docs needed to say plainly that filesystem-proof browser specs may require `MYCELIS_BACKEND_WORKSPACE_ROOT` or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT`
- the code needed a small explicit note explaining why the live spec cannot assume the backend workspace lives under the same checkout as the spec

Residual release risk:
- local operator debugging can still drift if browser proof is run from one worktree against a backend started from another without the workspace-root override
- release proof should continue to prefer the managed built-server path and serial worker count over ad hoc local runner tweaks

## Release Operator Checklist

Run this order for platform-facing release proof:

1. `uv run inv lifecycle.down`
2. `uv run inv k8s.reset`
3. `uv run inv lifecycle.up --frontend`
4. `uv run inv lifecycle.status`
5. `uv run inv db.migrate`
6. `uv run inv lifecycle.health`
7. `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`

If the live browser spec runs from a different checkout than the backend:

- set `MYCELIS_BACKEND_WORKSPACE_ROOT` or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT` to the backend's real `core/workspace` directory

## Documentation Alignment Rule

When the release platform behavior changes, update these surfaces in the same slice:

- `V8_DEV_STATE.md`
- `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
- `docs/architecture/OPERATIONS.md`
- `docs/TESTING.md`
- `ops/README.md`
- `interface/lib/docsManifest.ts` when the new canonical doc should appear in `/docs`

## Code Commentary Targets

Use code comments only when the release behavior would otherwise look arbitrary:

- explain why informational summary prompts stay answer-first even when they mention the workspace
- explain why the live browser spec accepts an explicit backend-workspace override
- explain why `db.migrate` skips replay once the schema is already bootstrapped

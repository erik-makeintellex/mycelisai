# V8 UI Testing Agentry Product Contract
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

## Purpose

This document defines what the UI testing agentry must prove before a V8 workspace or AI Organization change is treated as ready for review.

The goal is not only to prove that screens render. The goal is to prove that an operator can enter the product, work through the expected Soma-first path, observe governed actions, recover from failures, and keep trust in the system.

Companion artifacts:

- Stable mocked browser proof: `interface/e2e/specs/v8-ui-testing-agentry.spec.ts`
- Live backend browser proof: `interface/e2e/specs/soma-governance-live.spec.ts`
- Workspace continuity manual trust pass: `tests/ui/browser_qa_plan_workspace_chat.md`

## Required Proof Surfaces

The UI testing agentry must cover these operator-visible actions.

| Surface | Required operator action | Required proof |
| --- | --- | --- |
| AI Organization entry | Open an existing AI Organization and land in the Soma-first workspace | Soma is clearly primary and the organization frame stays visible |
| Direct Soma answer | Ask a non-mutating question in the main Soma conversation | Response lands in `answer` state with readable content |
| Governed mutation | Ask for a file or other mutating action | Response lands in `proposal` state with approval context and capability risk |
| Proposal cancellation | Cancel a pending proposed action | Proposal is neutralized and the UI states that no action executed |
| Approval execution | Confirm a pending proposed action | UI shows execution proof or an explicit bounded blocker |
| Cold-start recovery | Trigger a first-query transient failure | Transient startup failure is surfaced gracefully with a retry path, and retry recovers to a valid terminal state |
| Continuity | Refresh or re-enter the same AI Organization | Prior Soma context remains legible and scoped to that organization |
| Audit visibility | Open Automations -> Approvals -> Audit | Inspect-only `Activity Log` shows recent governance and execution events |
| Large content resilience | Ask for oversized markdown or code output | Content stays inside a bounded reading surface with horizontal overflow handling |
| Wording hygiene | Review workspace and support surfaces | No developer-only or architecture-leaking language appears in the normal path |

## Evidence Model

The UI testing agentry must produce findings in a format that is easy to triage.

Required output fields:

- `Severity`
- `Surface`
- `User action`
- `Expected`
- `Observed`
- `Why it matters`

Evidence requirements:

- Every criticism must include one visible piece of evidence.
- Every positive claim must include one specific reason trust was earned.
- When something feels off, classify it as one of:
  - `clarity`
  - `hierarchy`
  - `continuity`
  - `causality`
  - `navigation`
  - `trust`

## Automation Split

No single test lane is enough on its own.

1. Stable mocked browser proof
   - Purpose: prove the UI contract without backend flake
   - Command:
     - `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`

2. Live backend governance proof
   - Purpose: prove the real `/api/v1/chat` and confirm-action contract
   - Command:
     - `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`

3. Manual trust and disruption pass
   - Purpose: verify interruption, recovery feel, wording quality, and operator trust
   - Prompt baseline:
     - `tests/ui/browser_qa_plan_workspace_chat.md`

## Release Gate

The UI testing agentry must fail the release recommendation if any of the following are true:

- Soma is present in layout but not clearly primary in behavior
- the first-path workspace request fails without a clear retry and recovery path
- a mutating request bypasses the proposal and approval contract
- cancelling a proposal leaves uncertainty about whether execution still happened
- the operator cannot refresh or re-enter with confidence
- audit visibility is missing or only exposes raw backend noise
- default-path UI copy leaks developer-facing architecture terms

Passing this contract means the product behavior is believable, governed, and explainable from the operator's point of view.

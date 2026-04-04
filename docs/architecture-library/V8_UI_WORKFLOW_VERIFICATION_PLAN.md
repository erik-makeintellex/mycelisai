# V8 UI Workflow Verification Plan

> Status: ACTIVE
> Last Updated: 2026-04-03
> Purpose: Give the UI testing team a single release-verification plan grounded in the current V8 product surface, expected API calls, and terminal UI states.

## Current Release Read

- `IN_REVIEW`: build, typecheck, unit/integration coverage, live governed-browser proof, and service health proof
- `IN_REVIEW`: stable mocked browser proof for transient Soma cold-start recovery via graceful blocker plus retry
- `COMPLETE`: governed proposal flow, confirm/cancel flow, live `/api/v1/chat` path, audit/approvals surface, and fresh-cluster bring-up proof

## Team Structure

- `QA Lead`: owns the verification board, evidence capture, and final release verdict
- `Stable Browser Owner`: owns mocked browser proof and terminal-state correctness
- `Live Backend Owner`: owns real `/api/v1/chat`, `/api/v1/intent/confirm-action`, and filesystem-backed mutation proof
- `API Observer`: records request/response evidence and checks envelope normalization
- `UX Reviewer`: checks operator wording, continuity, readability, and degraded-state trust behavior
- `Release Recorder`: updates `V8_DEV_STATE.md` and attaches proof links/screenshots/logs

## Test Order

1. Run repo gates before browser work.
2. Run stable mocked browser proof.
3. Run live governed-browser proof.
4. Run manual trust pass only after automated gates are green or understood.
5. Record terminal outcome for every workflow: `answer`, `proposal`, `execution_result`, or `blocker`.

## Mandatory Repo Gates

- `uv run inv ci.build`
- `uv run inv ci.baseline`
- `uv run inv ci.service-check --live-backend`
- `uv run pytest tests/test_docs_links.py -q`

## Workflow Matrix

### 1. AI Organization Entry

- Route: `/dashboard`
- Owner: `Stable Browser Owner`
- Expected calls:
  - `GET /api/v1/templates?view=organization-starters`
  - `GET /api/v1/organizations?view=summary`
  - `POST /api/v1/organizations`
- Expected UI:
  - creation surface stays Soma-first and organization-first
  - successful creation routes into `/organizations/[id]`
  - failure keeps recent organizations and retry actions visible
- Acceptance evidence:
  - one successful empty-start or starter-template creation
  - one handled retry path

### 2. Organization Workspace Bootstrap

- Route: `/organizations/[id]`
- Owners: `Stable Browser Owner`, `UX Reviewer`
- Expected calls:
  - `GET /api/v1/organizations/{id}/home`
  - `GET /api/v1/organizations/{id}/loop-activity`
  - `GET /api/v1/organizations/{id}/automations`
  - `GET /api/v1/organizations/{id}/learning-insights`
  - `GET /api/v1/user/me`
  - `GET /api/v1/services/status`
- Expected UI:
  - Soma is clearly primary
  - Recent Activity, Memory & Continuity, Quick Checks, and support panels remain readable
  - no generic-assistant or backend-framework wording leaks
- Acceptance evidence:
  - page opens with organization context intact
  - support panels remain visible during workspace interactions

### 3. Soma Direct Answer Flow

- Route: `/organizations/[id]`
- Owners: `Stable Browser Owner`, `Live Backend Owner`, `API Observer`
- Expected calls:
  - `POST /api/v1/chat`
- Expected UI:
  - informational prompts resolve to `answer`
  - answer text renders inline in the main Soma conversation
  - no mutation controls appear for purely informational prompts
- Acceptance evidence:
  - mocked stable pass
  - live backend pass
  - page reload preserves the previous answer

### 4. Cold-Start Recovery

- Route: `/organizations/[id]`
- Owners: `Stable Browser Owner`, `Live Backend Owner`
- Expected calls:
  - first `POST /api/v1/chat` may fail transiently
  - retry path issues the next `POST /api/v1/chat` and recovers cleanly
- Expected UI:
  - first failure is bounded with clear retry guidance
  - retry returns the workspace to a valid terminal state
  - no lingering degraded Soma banner after recovery
- Acceptance evidence:
  - stable mocked proof of graceful blocker plus successful retry recovery
  - live lane should still avoid persistent failure for the same flow

### 5. Governed Mutation Proposal

- Routes: `/organizations/[id]`, `/automations?tab=approvals`
- Owners: `Stable Browser Owner`, `Live Backend Owner`, `API Observer`
- Expected calls:
  - `POST /api/v1/chat`
  - `POST /api/v1/intent/cancel-action`
  - `POST /api/v1/intent/confirm-action`
- Expected UI:
  - mutating request resolves to `proposal`
  - proposal card shows risk, approval posture, and capability context
  - cancel resolves explicitly without hidden execution
  - confirm resolves to `execution_result` or a bounded pending-proof state
- Acceptance evidence:
  - cancel path visible and immediate
  - live confirm produces run-linked proof and expected filesystem/result evidence

### 6. Approval and Audit Inspection

- Route: `/automations?tab=approvals`
- Owners: `Stable Browser Owner`, `UX Reviewer`
- Expected calls:
  - `GET /api/v1/governance/pending`
  - `GET /api/v1/audit`
- Expected UI:
  - Approvals tab is inspectable
  - Audit/Activity Log is inspect-only, not raw-log-first
  - recent actions, approval context, capability usage, and status are readable
- Acceptance evidence:
  - audit entries render with actor, action, timestamp, and result status
  - default UI is not overwhelmed by raw event detail

### 7. Organization Configuration Flows

- Route: `/organizations/[id]`
- Owners: `Stable Browser Owner`, `API Observer`
- Expected calls:
  - `PATCH /api/v1/organizations/{id}/ai-engine`
  - `PATCH /api/v1/organizations/{id}/response-contract`
  - `PATCH /api/v1/organizations/{id}/departments/{departmentId}/ai-engine`
  - `PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/ai-engine`
  - `PATCH /api/v1/organizations/{id}/departments/{departmentId}/agent-types/{agentTypeId}/response-contract`
- Expected UI:
  - guided change flow
  - inheritance/override labels stay accurate
  - failure shows retry guidance without dropping organization context
- Acceptance evidence:
  - org-level update
  - department override and revert
  - agent-type override and revert

### 8. System, Runs, Resources, and Docs Reachability

- Routes:
  - `/system`
  - `/runs`
  - `/runs/[id]`
  - `/resources`
  - `/docs`
- Owners: `Stable Browser Owner`, `API Observer`
- Expected calls:
  - `/api/v1/telemetry/compute`
  - `/api/v1/cognitive/status`
  - `/api/v1/runs/{id}/events`
  - resource/settings endpoints already covered by page suites
  - docs manifest/document endpoints
- Expected UI:
  - operator-readable system health
  - run timeline/event detail available
  - docs browser resolves canonical docs
- Acceptance evidence:
  - each route loads without broken navigation or hydration failures

### 9. Settings persistence and guided-default integrity

- Route: `/settings`
- Owners: `Stable Browser Owner`, `UX Reviewer`, `API Observer`
- Expected calls:
  - `GET /api/v1/user/settings`
  - `PUT /api/v1/user/settings`
- Expected UI:
  - settings still lead with a guided setup path instead of a raw admin panel
  - assistant identity and theme can be updated from the default profile path
  - saved values persist after reload
  - advanced sections remain intentional follow-on surfaces rather than the default first step
- Acceptance evidence:
  - assistant name saved and visible again after reload
  - theme saved, applied, and still selected after reload
  - guided setup framing remains intact while performing the workflow

### 10. Layered organization permissions and enterprise user management

- Route: `/settings`
- Owners: `Stable Browser Owner`, `UX Reviewer`, `API Observer`
- Expected calls:
  - `GET /api/v1/user/me`
  - `GET /api/v1/groups`
  - `GET /api/v1/groups/monitor`
  - optional enterprise owner proof may also exercise `POST /api/v1/user/settings` when the access layer is switched in a controlled test fixture
- Expected UI:
  - People & Access stays readable as an organization-access workflow, not a raw IAM console
  - base release keeps visible role/access framing plus collaboration groups in the default path
  - base release does not expose enterprise user-directory mutation controls by default
  - enterprise access management can unlock the user-directory layer without replacing the organization-access layer
  - enterprise non-owner roles do not receive directory mutation controls
- Acceptance evidence:
  - base-release proof shows `Base release layer`, organization access framing, collaboration groups, and no user-directory CRUD surface
  - enterprise-owner proof shows directory controls only when the enterprise layer is enabled
  - enterprise non-owner proof shows the directory layer as read-only or withheld rather than silently granting control

### 11. Native team instantiation for target output

- Route: `/organizations/[id]`, advanced team-design surface as needed
- Owners: `Stable Browser Owner`, `Live Backend Owner`, `API Observer`
- Expected calls:
  - current bounded release proof may begin with `POST /api/v1/organizations/{id}/workspace/actions`
  - future deeper proof may extend into `POST /api/v1/intent/negotiate`
  - `POST /api/v1/intent/commit` applies only when a real instantiation/activation confirmation path is present
  - artifact/result endpoints needed by the target output path
- Expected UI:
  - the operator can understand that Mycelis is instantiating a native managed team
  - the target output stays explicit in the proposal or activation framing
  - the returned result stays legible as team-owned output rather than generic residue
- Acceptance evidence:
  - one bounded release-proof path where the team-design workflow makes native managed-team ownership explicit for the target output
  - preferred proof is image-generation delivery with visible artifact/result return

### 12. External workflow-contract instantiation

- Route: `/organizations/[id]`, advanced integration/workflow surface as needed
- Owners: `Stable Browser Owner`, `Live Backend Owner`, `API Observer`
- Expected calls:
  - current bounded release proof may begin with `POST /api/v1/organizations/{id}/workspace/actions`
  - contract-specific integration inspection or invocation route for the supported external workflow surface
  - normalized result/artifact route back into Mycelis
- Expected UI:
  - the operator can understand that the target is an external workflow contract such as `n8n`, not a native Mycelis team
  - governance, capability, and result posture stay visible
  - normalized output returns into Mycelis without blurring ownership
- Acceptance evidence:
  - one bounded contract-level proof for external workflow instantiation posture in the guided organization workflow
  - when runnable proof exists, one normalized-result path captured as release evidence

## Manual Trust Pass

- Interrupt a streamed or active chat with refresh.
- Use browser Back/Forward across dashboard and organization routes.
- Force one handled mutation failure outside workspace permissions.
- Check that recent history and organization framing remain intact.
- Confirm labels remain operator-centric and never expose implementation internals like store names or framework jargon.

## Evidence Standard

- screenshot for each workflow terminal state
- captured request URL and method for every API-backed workflow
- note whether the terminal state was `answer`, `proposal`, `execution_result`, or `blocker`
- for failures, attach screenshot plus error-context/log excerpt and classify as `product`, `test`, or `environment`

## Exit Criteria

- repo gates green
- stable mocked browser proof green or explicitly blocked with written release call
- live governed-browser proof green
- every critical workflow above has current evidence
- release recorder updates `V8_DEV_STATE.md` with the final verdict

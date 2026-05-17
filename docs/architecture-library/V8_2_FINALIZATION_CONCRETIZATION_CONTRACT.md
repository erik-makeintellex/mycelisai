# V8.2 Finalization Concretization Contract
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md)

> Status: Canonical working contract
> Last Updated: 2026-05-16
> Module Boundary: Soma experience, runtime/capability, governance/trust, deployment/proof, QA/embodiment
> Purpose: Pin the first proofable slice, runtime objects, UI states, degradation semantics, deployment-trust ownership, and slice close-out template required before further architecture expansion.

## First Demo Slice

The first releasable demonstration is intentionally narrow:

```text
Operator asks Soma:
"Create a small playable browser game package with a README and validation notes."

Soma evaluates intent
-> creates an ExecutionContract
-> shows a governed proposal if mutation is required
-> operator approves
-> governed run invokes the file/artifact capability
-> durable project-package output appears
-> operator opens the game and proof
-> failure path shows retry, invalid proof, trusted remainder, and operator attention
```

Acceptance requires success and degraded-path proof in browser, not only API or unit proof.

## ExecutionContract

Durable, versioned runtime handshake object.

Required fields:
- `contract_id`
- `schema_version`
- `intent_summary`
- `operator_context_ref`
- `execution_shape`: direct, governed_capability, team_work, automation, review
- `requested_actions`
- `capability_requirements`
- `team_requirements`
- `governance_posture`
- `approval_requirements`
- `expected_outputs`
- `expected_proof`
- `risk_class`
- `recovery_posture`
- `degradation_behavior`
- `confidence_requirements`
- `run_id`
- `created_at`
- `updated_at`

## ProofArtifact

Durable operator trust object.

Required fields:
- `proof_id`
- `schema_version`
- `run_id`
- `contract_id`
- `execution_status`
- `output_refs`
- `evidence_refs`
- `audit_refs`
- `capability_refs`
- `team_refs`
- `validation_source`
- `proof_quality`
- `degradation_state`
- `trusted_remainder`
- `invalidated_evidence`
- `recovery_options`
- `confidence_provenance`
- `created_at`
- `updated_at`

## CapabilityManifestState

Capability manifest persistence is `REQUIRED` P0 trust infrastructure.

Required fields:
- `capability_id`
- `manifest_version`
- `display_name`
- `purpose`
- `source`
- `risk_class`
- `approval_posture`
- `allowed_roles`
- `input_schema_ref`
- `output_schema_ref`
- `health`
- `last_probe_status`
- `last_probe_at`
- `failure_posture`
- `recovery_posture`
- `audit_policy`
- `secret_ref_policy`
- `owner`
- `updated_at`

The Runtime/Capability lane owns persistence. Connected Tools shows the capability view. `System -> Deployments` may summarize health when it affects deployment trust.

## ConfidenceProvenance

Add room now; do not add scoring yet.

Required fields:
- `validation_source`: runtime, test, operator_review, capability_return, council_review
- `evidence_strength`: none, weak, adequate, strong
- `cross_check_status`: not_checked, agrees, disagrees, unavailable
- `proof_quality`: incomplete, partial, verified, degraded
- `review_lineage`
- `disagreement_markers`
- `last_validated_at`

## UI Response States

| State | Visible UI expression |
| --- | --- |
| `direct_answer` | Answer card with source/proof when retained. |
| `proposal` | Governed proposal card with approve/cancel and proof expectations. |
| `awaiting_approval` | Pending approval chip and blocked mutation summary. |
| `execution_result` | Operator trust package with output, run, proof, and next step. |
| `blocker` | Blocker card naming missing provider, permission, context, or capability. |
| `recovery_state` | Recovery card with retry, continue, stop, or request-input action. |
| `degraded_execution` | Degraded proof card naming trusted remainder and invalid evidence. |
| `retry_required` | Retry chip plus safe retry scope. |
| `partial_completion` | Partial output card plus remaining work and proof limits. |

These states must be visible as cards, chips, or trust indicators, not only backend classifications.

## Degraded Lifecycle

Canonical degraded run states:

| State | Meaning |
| --- | --- |
| `partially_complete` | Some outputs are valid, but the requested work is incomplete. |
| `proof_invalidated` | Output or evidence cannot support the claimed result. |
| `retryable` | Same contract can rerun safely with bounded changes. |
| `blocked_by_provider` | AI, search, media, or external provider is unavailable or timed out. |
| `blocked_by_capability` | Tool, MCP, script, filesystem, or API capability failed. |
| `needs_operator_attention` | Human choice, credential, approval, or clarification is required. |
| `safe_to_continue` | Trusted remainder can remain while failed portion reruns. |
| `must_rerun` | Proof/output cannot be trusted without rerunning the contract. |

Every degraded result must answer what succeeded, what failed, what remains trusted, what proof is invalid, what can continue safely, and what requires operator action.

## Deployment Trust Ownership

`System -> Deployments` owns detailed deployment trust:
- deployment root
- execution root
- workspace root
- artifact root
- current commit/image/chart
- endpoint posture
- runtime health
- proof lane
- recovery posture

Other surfaces may show summaries only. Resources can show connected capability context, but deployment trust does not live there.

## Team Prominence Rule

MVP default remains Soma-first. Teams stay compact, Soma-directed, and outcome-bound.

Do not promote teams as the visible center until:
- one bounded role-specific team task completes within product timeout
- output is retained and proof-linked
- degraded timeout behavior is visible
- team can be inspected, steered, paused, resumed, or archived

## Slice Close-Out Template

Every finalization slice must record:
- owning lane
- status marker
- exact user scenario
- touched runtime surfaces
- touched UI surfaces
- touched API surfaces
- emitted events
- ExecutionContract shape
- output object shape
- ProofArtifact shape
- UI response states proven
- degraded/recovery behavior
- validation commands
- browser proof
- deployment proof when runtime/deploy changes
- docs changed
- docs reviewed unchanged
- state update
- remaining blocker

No slice closes on implementation alone. It closes on visible output, proof, recovery posture, and documentation/state agreement.

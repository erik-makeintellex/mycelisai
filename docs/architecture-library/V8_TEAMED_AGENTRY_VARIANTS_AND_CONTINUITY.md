# V8 Teamed Agentry Variants And Continuity
> Navigation: [V8 Teamed Agentry Workflow Advantage](V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md)

Status: product/architecture detail.

## Variant 1: Direct Soma / Single Agent

Use for direct answers, summarization, simple analysis, and small non-mutating tasks.

Continuity: retain answer/activity only when useful.

## Variant 2: Single Agent With Deep Context

Use when one role can complete the work but needs memory, resources, or longer context.

Continuity: retain objective, context sources, and output.

## Variant 3: Lead Plus Specialist Pair

Use when coordination and one specialist role are enough.

Continuity: retain lead decision, specialist output, and review status.

## Variant 4: Compact Delivery Team

Use when work benefits from several roles but should remain one visible lane.

Continuity: retain team roster, plan, decisions, artifacts, and blockers.

## Variant 5: Multi-Lane Coordinated Team Bundle

Use when broad work naturally splits into independent tracks with later synthesis.

Continuity: retain lane objective, owner, status, output, and synthesis decision.

## Persistence Rule

Persist:
- objective
- plan/lane shape
- owners/roles
- decisions and approvals
- artifacts/outputs
- blockers and next steps

Do not persist:
- raw transient telemetry
- unreviewed private data
- secrets
- accidental model scratch text

## Reboot Proof

After restart or refresh, the operator should be able to see what happened, what remains, and where retained outputs live.

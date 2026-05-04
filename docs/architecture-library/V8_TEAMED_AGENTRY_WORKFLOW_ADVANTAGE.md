# V8 Teamed Agentry Workflow Advantage
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8 Workflow Variants](V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md)

Status: product/architecture positioning.

Detailed variants and continuity notes now live in [V8 Teamed Agentry Variants And Continuity](V8_TEAMED_AGENTRY_VARIANTS_AND_CONTINUITY.md).

## Why This Exists

Teamed agentry is valuable when work needs visible decomposition, review, specialization, continuity, or trust separation. It is not a requirement that every prompt spawn a large team.

## The Core Distinction

Single-agent flows optimize for speed and simplicity. Teamed flows optimize for structured work, reviewable lanes, and durable outputs.

Mycelis should choose the smallest workflow shape that can honestly satisfy the request.

## Workflow Variants

Default variants:
1. direct Soma / single agent
2. single agent with deep context
3. lead plus specialist pair
4. compact delivery team
5. multi-lane coordinated team bundle

See [V8 Teamed Agentry Variants And Continuity](V8_TEAMED_AGENTRY_VARIANTS_AND_CONTINUITY.md).

## When A Single Agent Is The Right Tool

Use direct Soma or a single agent when:
- the request is narrow
- no persistent team state is needed
- no specialist separation is useful
- no approval/review workflow is required
- output can be delivered directly

## Where Teamed Agentry Shows A True Win

Use a team or lanes when:
- the work has distinct research/build/review roles
- trust separation matters
- multiple artifacts must be coordinated
- recovery/reboot continuity matters
- the operator needs to review partial outputs
- the request is broad enough to split into smaller tracks

## Complex Workflow Patterns Where Teams Matter

Patterns:
- divergent research followed by convergent synthesis
- planning, build, and verification separation
- multi-artifact delivery packages
- trust-split workflows
- broad requests split into lanes
- recovery-critical execution
- deliberate internal tension for review

## Plan Design And Reboot Continuity

Plans should persist enough state for review and resume:
- objective
- lane/team shape
- ownership
- decisions
- outputs/artifacts
- blockers
- next action

Do not persist private/transient data as memory without review.

## Operator Guidance

Soma should explain why it is using direct answer, one compact team, or multiple lanes. The operator should see the plan shape and retained results without navigating raw orchestration internals.

## Product Positioning Rule

Mycelis wins by making the right amount of organization visible. Oversized teams are as much a failure as hiding real multi-role work behind one opaque answer.

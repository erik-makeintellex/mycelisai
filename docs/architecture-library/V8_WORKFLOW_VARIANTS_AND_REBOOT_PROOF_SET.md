# V8 Workflow Variants And Reboot Proof Set
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-04-15
> Purpose: Provide one concrete demonstration set that shows the difference between direct Soma, a compact team, and a multi-lane workflow for the same objective, while also proving that important plans survive a full environment reboot through the correct continuity surfaces.

## TOC

- [Why This Proof Set Exists](#why-this-proof-set-exists)
- [Automation Companions](#automation-companions)
- [The Shared Objective](#the-shared-objective)
- [Variant A: Direct Soma](#variant-a-direct-soma)
- [Variant B: Compact Team](#variant-b-compact-team)
- [Variant C: Multi-Lane Workflow](#variant-c-multi-lane-workflow)
- [Reboot Continuity Pass](#reboot-continuity-pass)
- [Evidence To Capture](#evidence-to-capture)
- [Pass Criteria](#pass-criteria)

## Why This Proof Set Exists

Mycelis should be able to demonstrate:
- when one direct Soma answer is enough
- when a compact team is a better fit
- when a broad request should become several lanes
- how a plan remains resumable after a full environment restart

The point is not "more agents."
The point is that the workflow shape should change the operator outcome in an understandable way.

## Automation Companions

Use the automated browser proof together with the manual reboot pass:
- `interface/e2e/specs/workflow-output.direct.spec.ts`, `interface/e2e/specs/workflow-output.compact-team.spec.ts`, `interface/e2e/specs/workflow-output.multi-lane.spec.ts`, and `interface/e2e/specs/workflow-output.reload-review.spec.ts` for stable mocked comparison of direct Soma, compact team packaging, multi-lane retained outputs, and resume from the retained package
- `interface/e2e/specs/workflow-variants-live-backend.spec.ts` for the live backend contract of direct Soma, compact-team vs multi-team shaping, and retained-output review after archive/reload
- `tests/ui/browser_qa_workflow_variants_reboot.md` for the full operator-facing reboot and trust pass

## The Shared Objective

Use one objective across all variants:

```text
Prepare a self-hosted release-readiness package for a Windows operator lane that uses a Windows GPU host for AI. I need a quick recommendation, a deployable validation plan, and a reviewable package I can resume after a reboot.
```

This objective is good because it naturally supports:
- a short direct answer
- a bounded plan package
- a broader multi-lane workflow with planning, validation, and review outputs

## Variant A: Direct Soma

Goal:
- show that a single direct answer is enough when the user mainly needs one recommendation

Prompt shape:

```text
Give me the shortest practical recommendation for how to validate this Windows self-hosted release lane.
```

Expected behavior:
- Soma returns one direct `answer`
- the output stays inline
- no team or temporary workflow group is required
- the answer names the recommended runtime and the highest-value next validation step

What this proves:
- Mycelis does not force a team when one clear answer is enough

## Variant B: Compact Team

Goal:
- show that planning, production, and review can be separated without turning the work into a giant lane bundle

Prompt shape:

```text
Create the smallest useful team to produce a release-readiness package for this Windows self-hosted lane. I want a named lead, a concise validation checklist, a deployment recommendation, and a risk review. Keep the final package retained so I can resume after a reboot.
```

Expected behavior:
- Soma proposes or launches a compact team
- the team shape is legible: Team Lead, Architect Prime, Focused Builder, optional Reviewer
- the output contract is explicit
- at least one retained artifact exists, such as:
  - validation checklist
  - deployment recommendation
  - risk summary

What this proves:
- the system can turn one request into a managed delivery package
- role separation produces a more inspectable result than one long answer
- the plan is not trapped only in the conversation

## Variant C: Multi-Lane Workflow

Goal:
- show the real teamed-agentry win: broad work becomes several compact lanes with visible handoffs and retained outputs

Prompt shape:

```text
This is broad. Split it into compact lanes for planning, deployment validation, and review. Keep each lane small, name the lead for each lane, retain the output contract for every lane, and make the whole plan resumable after a reboot.
```

Expected lane structure:
- Planning lane
  - defines deploy posture, scope, acceptance criteria, and output contract
- Validation lane
  - produces runtime/test checklist and environment proof sequence
- Review lane
  - identifies risks, missing evidence, and recovery notes

Expected behavior:
- Soma explains why the work was split
- each lane has a readable purpose
- each lane has a visible lead or coordination owner
- retained outputs are attached to the lane or temporary workflow group
- the operator can inspect the outputs without reading raw logs only

What this proves:
- Mycelis handles broad work through coordination, not through one overloaded thread
- the operator can see handoffs and outputs clearly
- the plan structure is durable enough to resume later

## Reboot Continuity Pass

Run this after Variant B or Variant C.

Goal:
- prove that important plans survive a full environment restart without being flattened into semantic memory

Required setup:
- at least one retained plan artifact exists
- the lane or temporary workflow group is visible
- the relevant run or output trace exists

Steps:

1. Create the plan through Variant B or Variant C.
2. Confirm the output contract and retained artifacts are visible.
3. Stop the environment cleanly or restart the runtime host.
4. Bring the environment back up.
5. Re-open the same AI Organization, temporary workflow group, or retained-output surface.
6. Ask Soma to resume from the retained plan:

```text
Resume the release-readiness work from the retained package and show me what is already done, what remains, and which lane or lead owns the next step.
```

Expected behavior:
- the retained plan summary or artifact is still available
- archived or active group outputs remain reviewable
- the run/history surfaces still explain what happened
- Soma can resume from the durable plan package instead of pretending the work never existed
- temporary continuity and retained artifacts help recovery without incorrectly promoting the whole plan into durable semantic memory

What this proves:
- plan continuity survives a reboot
- the recovery path is operational, not only conversational
- Mycelis keeps the right boundaries between:
  - temporary continuity
  - retained artifacts
  - trace and audit
  - long-term semantic memory

## Evidence To Capture

For each variant:
- screenshot of the user prompt
- screenshot of the terminal state
- screenshot of the output shape

For Variant A:
- direct answer screenshot

For Variant B:
- team or proposal screenshot
- retained artifact screenshot
- visible lead and output contract screenshot

For Variant C:
- lane bundle or temporary workflow group screenshot
- screenshot that shows separate lane outputs or named lane responsibilities
- screenshot that shows retained outputs remain visible

For reboot continuity:
- screenshot before restart showing retained plan package
- screenshot after restart showing the same package or lane still available
- screenshot of the resume response showing what is done, what remains, and who owns the next step

## Pass Criteria

This proof set passes only when all of the following are true:

1. Variant A stays direct and does not create unnecessary orchestration.
2. Variant B creates a compact team with a clear output package.
3. Variant C becomes several compact lanes instead of one oversized team.
4. The difference between the three workflow shapes is visible to a normal operator.
5. At least one retained plan package survives a full environment reboot.
6. Resume behavior uses retained outputs, continuity, and trace correctly instead of pretending all planning belongs in long-term semantic memory.

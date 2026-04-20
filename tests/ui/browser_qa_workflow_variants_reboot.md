# Browser QA Test Plan: Workflow Variants And Reboot Resume

## Application
Mycelis Cortex V8 operator workflow surfaces:
- Soma workspace
- Teams / guided team creation
- Groups / temporary workflow review
- retained outputs and resume path after restart

## Purpose

This manual pass proves one shared user objective across three workflow shapes:
- direct Soma
- compact team
- multi-lane workflow

It also proves that an important plan can survive a full environment reboot through retained outputs, continuity, and trace surfaces instead of living only in one chat thread.

Use this together with:
- `docs/architecture-library/V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md`
- `docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`
- `interface/e2e/specs/workflow-output.direct.spec.ts`
- `interface/e2e/specs/workflow-output.compact-team.spec.ts`
- `interface/e2e/specs/workflow-output.multi-lane.spec.ts`
- `interface/e2e/specs/workflow-output.reload-review.spec.ts`

## Shared Objective

Use this exact objective or a close equivalent:

```text
Prepare a self-hosted release-readiness package for a Windows operator lane that uses a Windows GPU host for AI. I need a quick recommendation, a deployable validation plan, and a reviewable package I can resume after a reboot.
```

## Pass A: Direct Soma

### Goal

Prove that Mycelis does not create orchestration when one direct answer is enough.

### Prompt

```text
Give me the shortest practical recommendation for how to validate this Windows self-hosted release lane.
```

### Expected

- terminal state is `answer`
- output stays inline in Soma
- no team or temporary workflow group is created
- response is concrete enough to act on

### Evidence

- screenshot of prompt
- screenshot of direct answer
- short note: did the system resist unnecessary orchestration?

## Pass B: Compact Team

### Goal

Prove that one managed delivery package is clearer than one long answer.

### Prompt

```text
Create the smallest useful team to produce a release-readiness package for this Windows self-hosted lane. I want a named lead, a concise validation checklist, a deployment recommendation, and a risk review. Keep the final package retained so I can resume after a reboot.
```

### Expected

- Soma proposes or launches a compact team
- team shape is legible
- a lead is visible
- output contract is explicit
- at least one retained artifact or package exists

### Evidence

- screenshot of team proposal or launched team
- screenshot of visible lead and target outputs
- screenshot of retained output or artifact summary
- short note: did the compact team create useful separation of planning, making, and review?

## Pass C: Multi-Lane Workflow

### Goal

Prove that broad work becomes several compact lanes instead of one giant roster or one overloaded thread.

### Prompt

```text
This is broad. Split it into compact lanes for planning, deployment validation, and review. Keep each lane small, name the lead for each lane, retain the output contract for every lane, and make the whole plan resumable after a reboot.
```

### Expected

- Soma explains why the work is split
- lanes are distinct and readable
- each lane has a visible purpose or owner
- retained outputs remain attached to the workflow or group

### Evidence

- screenshot of lane split or temporary workflow group
- screenshot showing lane outputs or responsibilities
- screenshot showing retained outputs remain visible
- short note: was the multi-lane structure genuinely clearer than one compact team would have been?

## Pass D: Reboot Resume

### Goal

Prove that the plan survives a full environment restart.

### Setup

Complete Pass B or Pass C first.

### Steps

1. Confirm a retained plan artifact, checklist, or output package exists.
2. Confirm the related team, group, or output surface is visible.
3. Restart the environment cleanly.
4. Re-open the same AI Organization and relevant workflow surface.
5. Ask Soma:

```text
Resume the release-readiness work from the retained package and show me what is already done, what remains, and which lane or lead owns the next step.
```

### Expected

- retained plan package still exists after restart
- the same group or team outputs are still reviewable
- Soma can resume from the retained package
- the system does not behave as if the work was lost
- the resume story uses retained outputs and continuity, not a fake fresh answer

### Evidence

- screenshot before restart showing retained plan package
- screenshot after restart showing the same package or output surface
- screenshot of the resume response
- short note: did the resume feel operational or did it feel like rebuilding context from scratch?

## Failure Classifications

Use these tags when something feels off:
- `clarity`
- `hierarchy`
- `continuity`
- `causality`
- `navigation`
- `trust`
- `workflow-shaping`
- `resume-path`

## Best Output Shape

Return findings in this format:

- **Severity**
- **Surface**
- **User action**
- **Expected**
- **Observed**
- **Why it matters**

For every criticism, include one visible piece of evidence.
For every positive claim, include one visible reason trust was earned.

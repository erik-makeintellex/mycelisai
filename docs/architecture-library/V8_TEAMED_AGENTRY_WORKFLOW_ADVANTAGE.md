# V8 Teamed Agentry Workflow Advantage
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-04-15
> Purpose: Define when Mycelis should use direct Soma or a single context-rich agent, when it should instantiate a compact team, and which workflow shapes show a real product advantage for teamed agentry over a singular agent with more training or more context.

## TOC

- [Why This Exists](#why-this-exists)
- [The Core Distinction](#the-core-distinction)
- [Workflow Variants](#workflow-variants)
- [When A Single Agent Is The Right Tool](#when-a-single-agent-is-the-right-tool)
- [Where Teamed Agentry Shows A True Win](#where-teamed-agentry-shows-a-true-win)
- [Complex Workflow Patterns Where Teams Matter](#complex-workflow-patterns-where-teams-matter)
- [Plan Design And Reboot Continuity](#plan-design-and-reboot-continuity)
- [Operator Guidance](#operator-guidance)
- [Product Positioning Rule](#product-positioning-rule)

## Why This Exists

Mycelis should not imply that a team is automatically better than one very capable agent.

Often the right answer is:
- one direct Soma response
- one context-rich reasoning pass
- one governed proposal

The product advantage appears when the work is not just "harder thinking."
It appears when the work benefits from:
- role separation
- parallel lanes
- visible review
- explicit handoffs
- different output contracts
- durable coordination over time

That is the boundary this document defines.

## The Core Distinction

A stronger single agent mainly improves:
- local reasoning depth
- style consistency
- memory use inside one thread
- performance on one continuous output

Teamed agentry mainly improves:
- decomposition across distinct responsibilities
- separation between planning, making, reviewing, and approving
- parallel exploration of alternatives
- recovery when one lane fails or stalls
- multi-output delivery with inspectable ownership
- continuity across longer and messier execution paths

The practical question is not:
- "Could one powerful agent do this eventually?"

The practical question is:
- "Does the operator get a better, more governable, more inspectable result when the work is split into explicit roles or lanes?"

## Workflow Variants

### Variant 1: Direct Soma / Single Agent

Use this when the work is one bounded response with one main reasoning line.

Examples:
- answer a product question
- summarize a document set
- draft one short memo
- explain a system state
- propose one next action

### Variant 2: Single Agent With Deep Context

Use this when the work still has one main output, but it depends on a lot of context or one long thread.

Examples:
- synthesize deployment context into one recommendation
- rewrite a policy draft with organization-specific constraints
- compare two architectures using prior discussion history
- extend one code patch in the same local area

This is still not a team problem. It is a continuity and context problem.

### Variant 3: Lead Plus Specialist Pair

Use this when the work has one visible owner but one secondary specialty materially changes the quality of the result.

Examples:
- Team Lead plus security reviewer for a deployment change
- Team Lead plus data analyst for a metrics interpretation brief
- Team Lead plus media specialist for a prompt pack and review

This is the smallest teamed workflow that creates visible separation of concerns.

### Variant 4: Compact Delivery Team

Use this when the work needs planning, production, and verification as distinct roles.

Default shape:
- Team Lead
- Architect Prime
- Focused Builder
- optional Reviewer / Tester
- optional Domain Specialist

Examples:
- implementation patch plus test proof
- launch brief plus risks and acceptance checklist
- deployment plan plus operator runbook plus recovery notes

This is where Mycelis starts to show a meaningful difference from a single contextualized agent.

### Variant 5: Multi-Lane Coordinated Team Bundle

Use this when the work has multiple deliverables or conflicting modes of work that should not be collapsed into one thread.

Examples:
- planning lane, implementation lane, validation lane
- research lane, synthesis lane, review lane
- media concept lane, generation lane, approval/publish lane

This is the strongest native teamed-agentry posture because the advantage is structural, not cosmetic.

## When A Single Agent Is The Right Tool

Prefer a single agent or direct Soma when:
- the user wants one answer, not a managed delivery process
- the task has one dominant output and no real handoff boundary
- extra roles would only restate the same reasoning in different voices
- review can happen inline instead of as a distinct lane
- the work is short enough that orchestration overhead would dominate
- the context load is high but still coherent inside one thread

Examples where a stronger single agent usually wins:
- "Explain this architecture decision."
- "Rewrite this paragraph for a customer."
- "Summarize these notes into one brief."
- "Give me the next two debugging steps."
- "Compare these two model providers for this exact setup."

In these cases, a team often adds ceremony without adding clarity.

## Where Teamed Agentry Shows A True Win

Teamed agentry shows a real win when at least one of these is true:

1. The work needs different role incentives.
   Example: one lane should generate options while another lane should attack risk or verify claims.

2. The work benefits from parallel exploration.
   Example: compare several implementation or research paths at once, then synthesize.

3. The work has more than one output contract.
   Example: deliver code, release notes, test evidence, and operator guidance together.

4. The work has explicit handoffs.
   Example: planning should complete before implementation, and implementation should complete before governed review.

5. The work spans different trust or governance postures.
   Example: research can be broad, but mutation approval and publication must be reviewed separately.

6. The work needs durable observability.
   Example: the operator should be able to inspect who produced which output and where a lane stalled.

7. The work is long-running enough that recovery matters.
   Example: a lane can fail, pause, or be retried without losing the whole mission structure.

## Complex Workflow Patterns Where Teams Matter

These are the strongest examples of teamed agentry producing something a single agent with more context usually does less reliably.

### 1. Divergent Research Followed By Convergent Synthesis

Pattern:
- multiple research lanes gather different evidence or viewpoints
- a synthesis lane reconciles findings
- a review lane checks whether the synthesis overfit one source or ignored contradictions

Why the team wins:
- parallel evidence gathering is explicit
- contradictions stay visible instead of being silently averaged away
- synthesis becomes an inspectable step rather than an invisible internal merge

### 2. Planning, Build, And Verification Separation

Pattern:
- Architect Prime defines the plan and output contract
- builder lane produces the artifact
- reviewer or tester lane checks acceptance criteria and regressions

Why the team wins:
- planning does not silently mutate while implementation is happening
- verification has a distinct responsibility instead of being self-grading
- the operator can inspect whether failure came from bad planning, bad build work, or bad validation

### 3. Multi-Artifact Delivery Packages

Pattern:
- one lane prepares implementation or core content
- one lane produces operator-facing explanation or packaging
- one lane validates completeness and consistency

Examples:
- product brief + deck outline + FAQ + rollout checklist
- code patch + migration notes + operator recovery guide
- media concept + generated assets + usage instructions

Why the team wins:
- outputs can be specialized without collapsing into one muddled artifact
- ownership of each output is visible
- package completeness is easier to verify

### 4. Trust-Split Workflows

Pattern:
- one lane can gather and analyze
- one lane can propose mutation or publication
- one lane can review trust, security, or policy impact

Why the team wins:
- safe and unsafe activities do not blur together
- approval boundaries remain legible
- the operator can see that exploration, action, and approval are different stages

### 5. Broad Requests That Should Become Lanes Instead Of A Giant Roster

Pattern:
- one broad ask is decomposed into several compact teams
- each lane has a lead and output contract
- Soma coordinates the bundle

Why the team wins:
- the system stays inspectable
- each lane remains small enough to understand
- complexity is handled by coordination, not by one overloaded thread or one bloated roster

### 6. Recovery-Critical Execution

Pattern:
- a run persists lane state, intermediate outputs, blockers, and retained artifacts
- one lane can pause or fail without erasing the whole mission
- another lane can resume, review, or replace the stalled work

Why the team wins:
- continuity becomes operational instead of purely conversational
- failure can be localized
- the operator can restart or reroute one lane without redoing the entire job

### 7. Deliberate Internal Tension

Pattern:
- one role is rewarded for possibility generation
- one role is rewarded for skepticism or constraint checking
- one role integrates toward a practical recommendation

Why the team wins:
- this produces structured productive tension
- it is stronger than asking one agent to "be creative and critical at the same time"
- the operator can inspect the disagreement instead of only seeing the final blended answer

## Plan Design And Reboot Continuity

The workflow advantage should also survive a full environment reboot.

That means a plan must not live only as:
- one transient chat turn
- one hidden internal chain of thought
- one unstructured memory blob

The durable design rule is:
- keep the plan shape visible
- keep the outputs retained
- keep the handoffs inspectable
- keep restart-safe working continuity separate from long-term learned memory

### What Should Persist

For any workflow that matters beyond one short answer, preserve:
- the requested outcome
- the chosen workflow shape
- the output contract
- the lane or team ownership
- the latest retained outputs
- the run or review trace needed to resume

In Mycelis, that persistence can live across several surfaces:
- temporary continuity for restart-safe in-flight planning checkpoints
- retained artifacts for durable plan summaries, briefs, checklists, and output packages
- run timelines for causal execution history
- team or group review surfaces for lane ownership and archived output review
- conversation templates for reusable ask shapes
- governed context stores when the plan or lesson should influence future work beyond one run

### Continuity By Workflow Variant

#### Direct Soma / Single Agent

Persist:
- short scoped session continuity
- an intentional summary or artifact if the answer should survive beyond the thread

Best use:
- small asks where reconstructing the context after reboot is cheap

Weakness:
- if the thread was carrying the whole plan implicitly, recovery becomes fragile

#### Single Agent With Deep Context

Persist:
- summary artifact
- acceptance criteria
- any reusable facts promoted into the correct memory layer

Best use:
- one complex output with one main reasoning line

Weakness:
- a reboot can leave too much of the structure trapped in one conversational state unless the plan was externalized

#### Compact Delivery Team

Persist:
- named lead
- explicit output contract
- retained outputs
- run trace
- temporary continuity checkpoints for in-flight work

Why this is stronger:
- the plan is no longer only inside one thread
- the lead, builder, and reviewer boundaries remain visible after restart
- recovery can resume from the last known output and lane state

#### Multi-Lane Coordinated Team Bundle

Persist:
- lane boundaries
- handoff rules
- retained outputs per lane
- archived temporary workflow groups when used
- causal run history and review state

Why this is strongest:
- after reboot, the operator can see which lane was complete, which lane was blocked, and which outputs already exist
- recovery can restart one lane without rebuilding the whole mission from scratch

### Practical Plan-Persistence Rule

When the plan matters enough that you would be upset to lose it after a full environment reboot, do not leave it only in chat.

Promote the plan into at least one durable surface:
- a retained artifact such as a brief, checklist, plan summary, or output contract
- a temporary workflow group with visible retained outputs
- a conversation template when the ask shape should be reused
- governed memory or context only when the content truly belongs in durable recall rather than one mission's working state

### The Important Boundary

Do not flatten all planning into semantic memory.

Keep these distinct:
- temporary continuity: restart-safe working state
- retained artifacts: durable outputs and plan packages
- semantic memory: intentionally promoted reusable meaning
- trace and audit: what happened and why

That separation is part of the product advantage.
A single context-rich agent often keeps the plan mostly inside one conversation.
A teamed execution system can keep the plan as explicit operational state that remains inspectable and resumable after reboot.

## Operator Guidance

Teach the product like this:

- use direct Soma when the user mainly needs one good answer
- use a compact team when the user needs a managed delivery process
- use multi-lane orchestration when the work has distinct outputs, handoffs, or review boundaries

The visible operator explanation should be:
- what lanes exist
- why each lane exists
- what each lane is expected to output
- how the lanes connect

The visible operator explanation should not be:
- "more agents means more intelligence"
- "a team is always better than one strong model"
- "the system spawned many roles because the prompt sounded impressive"

## Product Positioning Rule

Mycelis should position teamed agentry as a workflow and governance advantage, not as a magical collective mind.

The strongest product claim is:
- one persistent Soma can decide when a direct answer is enough
- when the work truly benefits from role separation, parallel lanes, or governed handoffs, Mycelis can instantiate a compact team or lane bundle that remains inspectable, recoverable, and output-oriented

That is the difference between:
- a singular agent with more context
- and an organization-shaped execution system.

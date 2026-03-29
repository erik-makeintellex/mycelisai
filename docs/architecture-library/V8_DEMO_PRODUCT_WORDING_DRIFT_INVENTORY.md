# V8 Demo Product Wording Drift Inventory

> Status: ACTIVE
> Last Updated: 2026-03-29
> Owner: Product Narrative Team with Default Experience Team and Memory, Continuity, and Trust Team
> Source Plan: `V8_DEMO_PRODUCT_STRIKE_TEAM_PLAN.md`
> Source Brief: `V8_DEMO_PRODUCT_EXECUTION_BRIEF.md`

---

## Purpose

This inventory is the first concrete wording-drift output for the demo-product strike team.

It exists to answer:

- where Mycelis already sounds like the product it is
- where older wording still makes the platform feel more technical, more abstract, or less coherent than it should
- which wording changes are safe to make immediately
- which naming drift should stay internal until a deliberate convergence slice is scheduled

---

## Product Language Standard

Default product surfaces should now reinforce this story:

1. create an AI Organization
2. work with Soma
3. see recent activity and approvals
4. understand what the organization retained
5. keep advanced power available without making the default story feel like a tool console

Preferred default terms:

- `AI Organization`
- `Soma`
- `Governed Execution`
- `Recent Activity`
- `Approvals`
- `Memory & Continuity`
- `Retained knowledge`
- `Retained patterns`

Terms to reduce on default product surfaces:

- `Learning`
- `Learning & Context`
- `reviewed learning`
- `generic chat`
- implementation-heavy architecture labels where a clear operator term already exists

---

## What Is Already Strong

- [page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(marketing)/page.tsx) already reads like a product, not a lab console.
- [CreateOrganizationEntry.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/CreateOrganizationEntry.tsx) already frames the product as AI Organization first, not blank-chat first.
- [OrganizationContextShell.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/OrganizationContextShell.tsx) already presents `Memory & Continuity` and retained patterns clearly in many user-facing panels.
- [core-concepts.md](d:/MakeIntellex/Projects/mycelisai/scratch/docs/user/core-concepts.md) and [soma-chat.md](d:/MakeIntellex/Projects/mycelisai/scratch/docs/user/soma-chat.md) are already closer to current product truth than older top-level docs.

---

## Priority Drift

## Priority 1: README operator language still lags the shipped product

File:
- [README.md](d:/MakeIntellex/Projects/mycelisai/scratch/README.md)

Observed drift:
- `learning visibility`
- `Recent Activity, Automations, Learning, Advisors, and Departments`
- `Advisors, Departments, Automations, Recent Activity, and Learning & Context`
- translation table entry `Identity / Continuity State -> Learning & Context`
- translation table entry `Learning Loops / reviewed learning -> What the Organization is Learning`

Why this matters:
- README is still the top narrative surface for technical evaluators.
- If README uses older wording while the product uses `Memory & Continuity`, the platform feels less deliberate and less productized.

Required direction:
- use `Memory & Continuity` for the visible operator concept
- use `What the Organization Is Retaining` or similar where the meaning is specifically about visible retained patterns
- keep deeper architecture language only where the section is explicitly architectural

Recommended handling:
- safe to update immediately in the next product-language slice

---

## Priority 2: organization-home backend-facing error text still says `Learning`

File:
- [OrganizationContextShell.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/OrganizationContextShell.tsx)

Observed drift:
- `"Learning updates unavailable"` is still used in the fetch/load path for organization retention insights

Why this matters:
- even if this string is not always the final visible headline, it increases the chance of drift in future UI fallbacks and test expectations

Required direction:
- align fallback/error wording to `Memory & Continuity` or retained-pattern language

Recommended handling:
- safe to update immediately for user-facing strings
- internal route names or API names do not need forced renaming in the same slice

---

## Priority 3: internal implementation names still carry `Learning` semantics

Files:
- [OrganizationContextShell.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/OrganizationContextShell.tsx)

Observed drift:
- `OrganizationLearningInsightItem`
- `learningInsights`
- `loadLearningInsights`
- `LearningVisibilityPanel`
- `learningHeadline`
- `learningWhyItMatters`
- `LearningStrengthBadge`

Why this matters:
- this is real naming drift between implementation and product language
- future contributors may accidentally reintroduce old wording if internals and product labels stay split for too long

Why this is not first-line copy work:
- these names are deeply wired into types, tests, and API shapes
- renaming blindly would create noisy churn and higher regression risk than a pure copy pass

Required direction:
- document these as convergence targets
- keep user-facing labels aligned first
- schedule a deliberate convergence slice if the backend/API contract is ready for renaming

Recommended handling:
- do not rename these blindly in the demo-product copy slice
- acceptable to add a brief clarifying code comment in a later cleanup slice if the internal name remains intentionally legacy for compatibility

---

## Priority 4: some top-level product text still leans on contrast-with-chat framing

Files:
- [README.md](d:/MakeIntellex/Projects/mycelisai/scratch/README.md)
- [CreateOrganizationEntry.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/CreateOrganizationEntry.tsx)
- [page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(marketing)/page.tsx)

Observed drift:
- repeated phrases like `generic assistant`, `generic chat`, or `blank assistant session`

Why this matters:
- some contrast is useful, especially for demo framing
- too much contrast language can make the product sound reactive instead of confident

Required direction:
- keep one or two strategic contrast statements
- prefer positive product language over repeated anti-chat language

Recommended handling:
- reduce repetition where possible
- keep the strongest contrast statement on the landing page and one support surface

---

## Route-by-Route Review

## Landing

File:
- [page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(marketing)/page.tsx)

Current status:
- mostly aligned

Notes:
- product framing is strong
- `Memory & Continuity` is already correct
- anti-chat contrast is present but may be slightly repetitive

Action:
- refine, not rewrite

## AI Organization entry

File:
- [CreateOrganizationEntry.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/CreateOrganizationEntry.tsx)

Current status:
- aligned

Notes:
- strong AI Organization-first framing
- good product tone
- `Memory & Continuity` already appears correctly

Action:
- preserve current direction
- only trim repeated anti-chat phrasing if it starts to feel overexplained

## Organization home

File:
- [OrganizationContextShell.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/OrganizationContextShell.tsx)

Current status:
- mostly aligned at the visible panel level

Notes:
- user-facing panels already say `What the Organization Is Retaining` and `Memory & Continuity`
- internal naming and a few fallback strings still lag behind

Action:
- fix visible drift first
- schedule internal convergence deliberately if needed

## User docs

Files:
- [core-concepts.md](d:/MakeIntellex/Projects/mycelisai/scratch/docs/user/core-concepts.md)
- [soma-chat.md](d:/MakeIntellex/Projects/mycelisai/scratch/docs/user/soma-chat.md)

Current status:
- aligned enough for the current slice

Action:
- no rewrite required right now

---

## Team Actions

## Product Narrative Team

- update [README.md](d:/MakeIntellex/Projects/mycelisai/scratch/README.md) to match shipped product language
- reduce repeated anti-chat phrasing where positive product framing can do the job better

## Default Experience Team

- preserve the current direction in landing and AI Organization entry
- only approve copy changes that strengthen confidence and legibility without increasing density

## Memory, Continuity, and Trust Team

- normalize user-facing error/fallback wording in [OrganizationContextShell.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/OrganizationContextShell.tsx)
- define which internal `Learning` names are compatibility debt versus acceptable runtime terminology

## Capability Preservation Team

- confirm that wording cleanup does not accidentally hide advanced memory, RAG, or system concepts from advanced docs and routes

## UI Testing Agentry Team

- verify that landing, AI Organization entry, and organization home consistently say `Memory & Continuity` / retained-knowledge language where expected
- flag any fallback UI that still surfaces old `Learning` terminology during degraded states

---

## Acceptance Rule

This inventory is complete only when:

- the team agrees on which wording drift is user-facing and must be fixed now
- the team agrees on which naming drift is internal and should be handled in a later convergence slice
- README, landing, organization entry, and organization home stop telling different stories about continuity and retained knowledge

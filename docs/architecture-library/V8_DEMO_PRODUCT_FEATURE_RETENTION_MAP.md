# V8 Demo Product Feature Retention Map

> Status: ACTIVE
> Last Updated: 2026-03-29
> Owner: Capability Preservation Team
> Source Plan: `V8_DEMO_PRODUCT_STRIKE_TEAM_PLAN.md`
> Source Brief: `V8_DEMO_PRODUCT_EXECUTION_BRIEF.md`

---

## Purpose

This map proves that demo-product simplification does not require feature reduction.

It exists to answer:

- which capabilities stay in the default product story
- which capabilities intentionally move behind Advanced mode
- where retained advanced power actually lives today
- what the teams must protect while making Mycelis more legible

---

## Preservation Rule

The default experience may become simpler.

The platform may not become weaker by accident.

For this lane:

- default surfaces must stay product-first
- advanced surfaces must stay reachable
- docs must continue to carry the deeper system story
- no retained capability should become orphaned, hidden without explanation, or silently removed

---

## Default-Surface Power That Must Stay Visible

## AI Organization creation

Surface:
- [CreateOrganizationEntry.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/CreateOrganizationEntry.tsx)

Retained capability:
- create from starter template
- create empty organization
- reopen recent organization
- bounded first-run explanation for AI Engine and Memory & Continuity posture

Preservation rule:
- keep this as the first obvious product action
- do not collapse it into a generic prompt box or advanced setup wizard

## Organization home and Soma workspace

Surface:
- [OrganizationContextShell.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/OrganizationContextShell.tsx)

Retained capability:
- direct-answer Soma planning
- recent activity visibility
- automations visibility
- retained knowledge / continuity visibility
- guided AI Engine and Response Style tuning
- team-design mode inside the same workspace

Preservation rule:
- simplify wording if needed
- do not remove the guided tuning and continuity proof already present here

## Governed mutation flow

Surfaces:
- [MissionControlChat.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/dashboard/MissionControlChat.tsx)
- proposal/approval surfaces already in the governed chat path

Retained capability:
- direct answer for low-risk/non-mutating work
- proposal-first mutation
- confirm / cancel
- execution proof in the same workflow

Preservation rule:
- this remains one of the core differentiators of the product story

---

## Advanced-Surface Retention Table

| Capability area | Retained home | Current proof | Default visibility rule |
| --- | --- | --- | --- |
| Connected tools and MCP access | [resources/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/resources/page.tsx) -> `Connected Tools` | advanced-gated Resources route | hidden by default, intentionally reachable |
| Workspace file access | [resources/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/resources/page.tsx) -> `Workspace Files` | advanced-gated Resources route | hidden by default, intentionally reachable |
| Managed exchange inspection | [resources/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/resources/page.tsx) -> `Exchange` | advanced-gated Resources route | hidden by default, intentionally reachable |
| Global AI Engine setup | [resources/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/resources/page.tsx) and [settings/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/settings/page.tsx) -> `AI Engines` | advanced tab access still present | hidden by default, intentionally reachable |
| Role library / reusable roles | [resources/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/resources/page.tsx) -> `Role Library` | advanced-gated Resources route | hidden by default, intentionally reachable |
| Deep memory inspection | [memory/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/memory/page.tsx) | advanced-gated Memory route | hidden by default, intentionally reachable |
| System diagnostics and recovery | [system/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/system/page.tsx) | advanced-gated System route | hidden by default, intentionally reachable |
| Shared teams and workflow builder | [automations/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/automations/page.tsx) -> `Shared Teams`, `Workflow Builder` | advanced tabs still wired | hidden by default, intentionally reachable |
| Advanced settings for AI engines and tools | [settings/page.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/app/(app)/settings/page.tsx) | advanced tabs still wired | hidden by default, intentionally reachable |
| Advanced routing / broadcast controls | [MissionControlChat.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/dashboard/MissionControlChat.tsx) | advanced routing and broadcast mode remain in chat | hidden by default, intentionally reachable |

---

## Retained Capability Notes

## Navigation-level retention

Surface:
- [ZoneA_Rail.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/shell/ZoneA_Rail.tsx)

Current truth:
- primary navigation stays simple
- `Resources`, `Memory`, and `System` appear only when Advanced mode is on

Why this is correct:
- it protects the partner/funder product story
- it does not remove advanced power

## Advanced-mode explanation

Surfaces:
- [AdvancedModeGate.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/shared/AdvancedModeGate.tsx)
- advanced-gated route pages under `Resources`, `Memory`, and `System`

Current truth:
- hidden surfaces are explained as intentionally tucked away, not missing

Why this matters:
- advanced users should feel guided, not blocked

## Chat-level advanced power

Surface:
- [MissionControlChat.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/dashboard/MissionControlChat.tsx)

Current truth:
- advanced routing metadata, direct targeting, and broadcast mode still exist
- these are not forced into the default operator path

Why this matters:
- high-power runtime behavior remains available without making the product feel like a routing console

## Team-design power

Surface:
- [TeamLeadInteractionPanel.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/organizations/TeamLeadInteractionPanel.tsx)

Current truth:
- team creation and delivery-lane shaping remain available inside the organization workspace
- this power stays nested under Soma instead of becoming a separate intimidating front door

Why this matters:
- we are simplifying the entry story, not reducing the organization-building capability

---

## Features That Must Not Be Lost During Demo Simplification

- starter-template creation
- empty organization creation
- organization re-entry
- direct-answer Soma path
- governed proposal / confirm / cancel path
- recent activity visibility
- retained knowledge / continuity visibility
- AI Engine and Response Style controls
- team-design mode
- connected tools and workspace file access
- deep memory inspection
- system diagnostics and recovery
- advanced routing / broadcast behavior
- docs access for deep architecture explanation

---

## Risk Flags

## Risk 1: simplification by concealment

Failure mode:
- a surface is removed from default view but not given a clear retained home

Guard:
- every hidden capability must have a named route, tab, or doc location

## Risk 2: copy-only product cleanup hides capability meaning

Failure mode:
- wording cleanup removes references to control, continuity, or inspection that users actually need

Guard:
- Narrative changes must be reviewed by Capability Preservation Team before acceptance

## Risk 3: advanced routes become stale because they are no longer demo-visible

Failure mode:
- advanced tabs remain in code but stop being tested or described

Guard:
- UI Testing Agentry Team must explicitly verify advanced reachability, not only the default flow

---

## Team Actions

## Capability Preservation Team

- use this document as the retained-home checklist
- reject any demo-product change that weakens a retained capability without a deliberate replacement

## Product Narrative Team

- simplify visible wording without implying that advanced capabilities are gone

## Default Experience Team

- preserve the simple navigation split already present in [ZoneA_Rail.tsx](d:/MakeIntellex/Projects/mycelisai/scratch/interface/components/shell/ZoneA_Rail.tsx)

## UI Testing Agentry Team

- verify both:
  - default product legibility
  - advanced route reachability after enabling Advanced mode

## Release and Ops Team

- keep docs and demo guidance honest about which features are advanced-only versus unavailable

---

## Acceptance Rule

This map is complete only if the team can answer all three questions clearly:

1. What stays in the default product story?
2. What remains available only in Advanced mode or deep docs?
3. Where does each retained advanced capability live right now?

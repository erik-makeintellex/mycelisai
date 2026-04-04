# V8 MVP Investor Demo and Governed Capability Plan

> Status: ACTIVE
> Last Updated: 2026-04-04
> Owner: Product Management / Delivery Coordination
> Supporting Teams: Demo Scenario Team, UI Testing Agentry Team, Runtime and Governance Team, MCP and Integration Usability Team, Release and Ops Team
> Purpose: Define the canonical MVP investor workflow story for Soma, governed execution, MCP-powered input/output, securable web research, and inspectable context-security controls.

---

## Why This Plan Exists

The investor story now needs to prove more than:

- Soma can answer
- Soma can request approval
- a file can be created

It also needs to prove:

- Soma is a strong initial working counterpart before any extra tooling is added
- MCP extends Mycelis into a governed input/output platform instead of a closed assistant
- external and research capability can be useful without becoming an uncontrolled breach path
- security posture is inspectable and configurable, not hand-waved

This plan keeps that story grounded in what the current product can actually demonstrate.

---

## Core Investor Message

Mycelis is a governed AI Organization product where:

1. Soma gives immediate planning and review value as the initial working agent
2. risky or mutating actions become explicit, reviewable workflows
3. connected tools extend Soma into real-world input/output work through MCP
4. web research and external context stay useful, but remain policy-bounded and inspectable
5. advanced capability is preserved without polluting the default product path

---

## Non-Negotiable Demo Rules

1. Start with product value before integration depth.
2. Show Soma’s baseline usefulness before showing any MCP expansion.
3. Present MCP as governed capability expansion, not as a developer plumbing exercise.
4. Show web access as bounded and securable, not unrestricted browser autonomy.
5. Make security visible through policy, trust labels, sensitivity, review posture, and approval boundaries.
6. Do not claim capabilities that are not supported by the current product proof.

---

## Canonical Workflow Stack

## Workflow 1: Soma as the initial working agent

Surface:
- `/dashboard`
- `/organizations/{id}`

Goal:
- show that the product is useful before any external capability is added

Operator actions:
- create `Northstar Labs`
- ask: `Review this AI Organization and recommend the first operating priority.`

What this proves:
- the initial agent can review, plan, and orient the organization directly
- the default path is product-shaped rather than setup-shaped
- Soma starts as a working counterpart, not merely a routing shell

Expected state:
- `answer`

## Workflow 2: Governed artifact creation

Surface:
- `/organizations/{id}`

Operator action:
- ask: `Create a kickoff brief in the workspace called northstar_kickoff.md summarizing the first operating priority and next steps.`

What this proves:
- Soma can move from planning into governed mutation
- artifact creation is visible and reviewable
- approval/control is part of the workflow, not hidden system behavior

Expected states:
- `proposal`
- `execution_result` after confirm

## Workflow 3: MCP as the input/output expansion layer

Surface:
- `/resources?tab=tools`

Operator goal:
- show how Mycelis expands beyond the initial agent through connected tools

What to show:
- `Connected Tools`
- `Current Group MCP Config`
- curated library posture
- installed vs library separation

Key message:
- Soma is useful by default
- MCP adds governed input/output reach: files, fetch/research, memory, GitHub, and future business systems
- Mycelis does not require raw arbitrary tool hookup to become useful

Current proof posture:
- local-first curated entries can install directly in the current user-owned group when policy allows
- remote/high-risk entries can return an approval boundary instead of silently installing

## Workflow 4: Governed web research

Surface:
- `/resources?tab=tools`
- `/resources?tab=exchange`
- Soma workspace when the environment is prepared for the research ask

Operator goal:
- prove that web/external context can inform work without becoming an unbounded browser hole

Primary operator story:
1. show that research capability is available through connected tools (`fetch` / research-capable MCP path)
2. issue a bounded research-style ask through Soma when the environment is prepared
3. inspect resulting exchange visibility for research outputs

Canonical research prompt:
- `Using connected research tools, gather a short market note on AI organization workflow products and summarize the top three signals for Northstar Labs.`

What this must prove:
- web/external input can help Soma produce a better answer
- research outputs can be normalized into managed exchange
- research does not bypass trust labeling or review posture

If live research is not the chosen demo path:
- still show the research-capability posture through Connected Tools and Exchange
- narrate that the product already classifies and secures research outputs through exchange labels and review posture

## Workflow 5: Context security and governance visibility

Surface:
- `/automations?tab=approvals`
- `/resources?tab=exchange`

Operator goal:
- show that external/contextual capability is controlled and inspectable

What to show:
- policy editing / policy posture
- audit/activity visibility
- exchange labels such as:
  - `sensitivity_class`
  - `trust_class`
  - `review_required`
  - `capability_id`
- approval boundary messaging for higher-risk MCP entries

Key message:
- Mycelis does not just “have web access”
- Mycelis lets operators decide how external capability is admitted and how its outputs are trusted

## Workflow 6: Optional advanced reveal

Surface:
- `/memory`
- `/resources`
- `/system`

Goal:
- show retained platform depth after the core investor story is already understood

Key message:
- advanced power is preserved
- advanced power is intentionally inspectable instead of being forced into the first-run workflow

---

## Security Story the Demo Must Land

The investor should walk away with these exact conclusions:

1. Soma can work effectively before external integrations are added.
2. MCP is how Mycelis expands input/output capability in a governed way.
3. Web or external research is not a hidden unrestricted power; it is policy-bounded capability.
4. Outputs from external/research capability carry context-security meaning, not just raw content.
5. Remote/high-risk additions can require approval, while low-risk local-first configuration can stay usable.

Current concrete security proof in product/docs:

- local-first curated MCP install posture
- explicit approval boundary for remote or still-sensitive MCP entries
- managed exchange labels for trust, sensitivity, and review posture
- audit visibility for proposal, execution, capability use, and result lineage
- policy surface for governance review

---

## Demo Environment Prep

Before the investor demo:

1. Confirm the stack is healthy in the chosen environment.
2. Confirm organization creation works.
3. Confirm governed proposal/confirm flow is healthy.
4. Confirm Resources loads:
   - Connected Tools
   - Exchange
   - Workspace Files
5. Confirm MCP library inspection/install messaging is healthy.
6. Confirm exchange inspection can show research-style trust/sensitivity labels.
7. Decide whether the live demo will include:
   - full live research ask through Soma
   - or the safer inspectable-capability reveal only
8. Rehearse one fallback where live research is skipped but governed capability and security posture are still demonstrated.

---

## Recommended Demo Sequence

1. Category and value
   - landing -> dashboard -> create organization
2. Initial agent value
   - Soma planning/review answer
3. Governed action
   - proposal -> confirm -> artifact result
4. Continuity
   - refresh/re-entry, recent activity, retained value
5. Connected tools
   - MCP library/current-group posture
6. Research + context security
   - research capability -> exchange trust/sensitivity/review labels
7. Policy/control reveal
   - approvals/policy/audit
8. Optional advanced depth
   - memory/system/resources

This order matters. The investor should understand the product before seeing advanced integration depth.

---

## Verification Pack Required

Docs and planning proof:
- `docs/architecture-library/V8_PARTNER_DEMO_SCRIPT.md`
- `docs/architecture-library/V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md`
- `docs/architecture/MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md`
- `docs/architecture-library/V8_USER_WORKFLOW_EXECUTION_AND_VERIFICATION_PLAN.md`

Focused validation:

```bash
cd interface && npx vitest run __tests__/settings/MCPLibraryBrowser.test.tsx __tests__/settings/MCPToolRegistry.test.tsx __tests__/resources/ExchangeInspector.test.tsx __tests__/pages/ResourcesPage.test.tsx __tests__/automations/ApprovalsTab.test.tsx __tests__/dashboard/ProposedActionBlock.test.tsx __tests__/dashboard/MissionControlChat.test.tsx __tests__/pages/SettingsPage.test.tsx --reporter=dot
cd interface && npx tsc --noEmit
uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts
uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-organization-entry.spec.ts
uv run inv interface.e2e --project=chromium --spec=e2e/specs/settings.spec.ts
uv run inv interface.e2e --project=chromium --spec=e2e/specs/governance.spec.ts
$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q
```

Optional deeper proof when preparing a live high-stakes environment:

```bash
cd core && go test ./internal/mcp ./internal/exchange ./internal/server -count=1
uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts
```

---

## Acceptance Rule

The MVP investor lane is ready only when all of the following are true:

1. the investor story proves Soma’s initial value before external tools are introduced
2. governed mutation is clear and confidence-building
3. MCP reads as a practical input/output expansion layer
4. web/external research is framed as useful and securable
5. context security is visible through trust/sensitivity/review posture, not buried in backend internals
6. the advanced reveal strengthens confidence without obscuring the default product story

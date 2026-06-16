# User Docs Home
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

Use this page when you want to operate Mycelis through the product instead of reading backend or implementation details.

## TOC

- [Start Here](#start-here)
- [Core Workflows](#core-workflows)
- [Setup And Recovery](#setup-and-recovery)
- [Advanced User Surfaces](#advanced-user-surfaces)

## Start Here

For most user testing and daily operation, start with Soma:

1. Sign in through `/login` with the local owner credential or enabled enterprise SSO.
2. Confirm `/dashboard` shows the signed-in Soma operating environment, including Access, Identity, and Scope.
3. Open the Soma workspace.
4. Ask for the outcome you want.
5. Let Soma propose the smallest useful team or direct answer path.
6. Review visible outputs in chat, team lead workspaces, or retained artifacts.
7. Use Settings, Resources, System, or Memory only when the workflow calls for them.

The normal product path is not "configure every backend component first." It is "sign in, tell Soma the goal, review the proposed execution shape, and inspect outputs."

## New-User Setup Checklist

Before accepting a new-user browser pass, verify these from the product:

1. Login is clear: local owner login works, SSO availability or domain restriction copy is understandable, and failures are actionable.
2. Dashboard orientation is clear: Soma input is primary, Active Work is compact by default, and any backlog opens through `Teams`.
3. Provider/search readiness is visible: Settings, Resources, or System tells the operator whether Soma can use the intended local or hosted engine/search path.
4. Capabilities are understandable: `Resources -> Capabilities` shows what Soma can use, what needs repair, **Add MCP Server**, required env vars by name, and recent MCP activity when tools are used.
5. Output roots are known: `MYCELIS_WORKSPACE` is where generated files, project packages, browser games, and filesystem MCP writes land; `MYCELIS_ARTIFACT_ROOT` is where file-backed artifacts and cached media land.
6. Canonical demo is repeatable: a retained demo output, such as a project package, opens from the browser, survives refresh/reload, and links to run/proof evidence.
7. Team proof is honest: one bounded team ask reaches readable `output_ready` or a visible `degraded` timeout/offline/unreadable state with recovery guidance.

Good first prompts:

- "Give me a short readiness check for this Mycelis environment."
- "Create a retained demo package I can open in the browser and include proof I can revisit."
- "Ask the active delivery team for one bounded status update; if it times out, show the degraded proof and recovery path."

Generated project packages, browser games, workspace files, and filesystem MCP writes land under the configured `MYCELIS_WORKSPACE`. File-backed artifacts and cached media land under `MYCELIS_ARTIFACT_ROOT`. Open **Resources -> Output Files** to browse generated content or use **Open folder** on an output card to open its local folder directly. Open **System -> Deployments** in Advanced mode when you need to confirm the exact runtime paths before or after execution.

Concrete requests Soma should understand when the matching capability is configured:

- "Search the web for the latest changes in self-hosted AI agent products, summarize the findings, and cite sources."
- "Create a small temporary team for this review and bring the retained output back here."
- "Ask the active teams for blockers and tell me which workflow needs attention first."
- "Use host data under `workspace/shared-sources` and list which files shaped the answer."
- "Review the current MCP tool structure and recommend which agents should have which tools."
- "Review my request against prior context, infer the action, and ask me to confirm before you execute."
- "Review private service or private data boundaries, name the protection reason, and ask before using credentials, customer data, or recurring behavior."

## Core Workflows

- [Using Soma Chat](soma-chat.md): central Soma chat, concrete search/team/host-data prompts, generated outputs, team handoffs, and failure recovery.
- [Workflow Variants And Plan Memory](workflow-variants-and-plan-memory.md): when direct Soma is enough, when a team matters, and how to keep plans through a reboot.
- [Teams](teams.md): compact team creation, lead-centered workflow, and broad-ask lane splitting.
- [Core Concepts](core-concepts.md): operator-language explanation of Soma, Council, teams, memory, and governance.
- [Resources](resources.md): private/context content, capability readiness, MCP tool structure, output files, AI engines, deployment context, and tool activity.
- [Memory](memory.md): semantic search, retained knowledge, reflection context, and continuity boundaries.
- [Settings And Access](settings-access.md): profile, People & Access, auth-provider posture, access-denied recovery, and connected-tool/search management boundaries.
- [Authentication Modes](auth-modes.md): local owner auth, break-glass recovery, OIDC/OAuth, SAML, Entra ID, Google Workspace, GitHub, and future SCIM enablement.

## Setup And Recovery

- [Deployment Method Selection](deployment-methods.md): choose Docker Compose, Rancher K3s or k3d local Kubernetes, enterprise self-hosted Kubernetes, or edge/binary deployment by target environment.
- [System Status & Recovery](system-status-recovery.md): health checks, degraded states, and recovery expectations.
- [Governance & Trust](governance-trust.md): approval posture, risk classes, and audit visibility.
- [Run Timeline](run-timeline.md): reading workflow activity, message-bus summaries, execution timelines, and outcome history.

## Advanced User Surfaces

- [Automations](automations.md): inspect active reviews, checks, triggers, and approval-facing automation behavior.
- [Meta-Agent & Blueprints](meta-agent-blueprint.md): advanced blueprint and mission planning language when you need graph-level detail.

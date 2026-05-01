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

1. Open the Soma workspace.
2. Ask for the outcome you want.
3. Let Soma propose the smallest useful team or direct answer path.
4. Review visible outputs in chat, team lead workspaces, or retained artifacts.
5. Use Settings, Resources, or Memory only when the workflow calls for them.

The normal product path is not "configure every backend component first." It is "tell Soma the goal, review the proposed execution shape, and inspect outputs."

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
- [Resources](resources.md): private/context content, MCP tool structure, workspace files, AI engines, deployment context, and Connected Tools activity.
- [Memory](memory.md): semantic search, retained knowledge, reflection context, and continuity boundaries.
- [Settings And Access](settings-access.md): profile, People & Access, auth-provider posture, access-denied recovery, and connected-tool/search management boundaries.

## Setup And Recovery

- [Deployment Method Selection](deployment-methods.md): choose Docker Compose, local `k3d`, enterprise self-hosted Kubernetes, or edge/binary deployment by target environment.
- [System Status & Recovery](system-status-recovery.md): health checks, degraded states, and recovery expectations.
- [Governance & Trust](governance-trust.md): approval posture, risk classes, and audit visibility.
- [Run Timeline](run-timeline.md): reading workflow activity, message-bus summaries, execution timelines, and outcome history.

## Advanced User Surfaces

- [Automations](automations.md): inspect active reviews, checks, triggers, and approval-facing automation behavior.
- [Meta-Agent & Blueprints](meta-agent-blueprint.md): advanced blueprint and mission planning language when you need graph-level detail.

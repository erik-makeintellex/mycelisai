# Mycelis User & Operator Guide
> Navigation: [Project README](../README.md) | [Docs Home](README.md) | [User Docs Home](user/README.md)

Mycelis is an AI Organization platform where you collaborate with **Soma** to plan, confirm, and materialize governed deliverables without managing raw backend infrastructure.

## TOC

- [1. Start With Soma & Deliverable Starters](#1-start-with-soma--deliverable-starters)
- [2. Governed Confirmation & 12 Deliverable Bundles](#2-governed-confirmation--12-deliverable-bundles)
- [3. Durable Files, Live HTML5 Preview, Follow-Up Queueing & Deployments Hub](#3-durable-files-live-html5-preview-follow-up-queueing--deployments-hub)
- [4. AI Organizations, Ideal Delivery Reviewer & Learning Loops](#4-ai-organizations-ideal-delivery-reviewer--learning-loops)
- [5. Trusted-Memory Arbitration & Promotion](#5-trusted-memory-arbitration--promotion)
- [6. Multi-Role RBAC, Web Agentry Presets & Navigation Modes](#6-multi-role-rbac-web-agentry-presets--navigation-modes)

## 1. Start With Soma & Deliverable Starters

When you open the Soma workspace (`/dashboard` or `/organizations/[id]`), Soma's **Start with Soma** suggestion bar surfaces **Popular deliverable starters** alongside collapsible **Advanced orchestration & council playbooks**:

- **Marketing Launch Kit** (`marketing_launch_bundle`)
- **Playable Puzzle Game / Street Fighter II Tetris** (`interactive_game_bundle`)
- **Customer Onboarding Kit** (`customer_onboarding_bundle`)
- **Interactive KPI Dashboard** (`kpi_dashboard_bundle`)
- **Product Spec & Wireframe** (`product_wireframe_bundle`)
- **Engineering Task Pack** (`codex_task_bundle`)

Selecting a starter or typing a casual or structured goal routes through Soma:
- **Direct Answer (`answer`):** Read-only questions, runtime state checks, and specialist playbook consultations return immediately in chat.
- **Governed Proposal (`proposal`):** Actions that write files, launch team workflows, or bind MCP additions present an outcome-first proposal card (`<50ms` deterministic fast-path) showing plain-English intent, planned deliverables/teams, and **Approve & Execute** / **Cancel** buttons. Technical telemetry stays tucked behind **`[View Proof Details]`**.
- **Stackable Work Queue (`+ Queue` / `Run Next`):** Click **`+ Queue`** next to the prompt input (or keep typing while Soma is working) to stack follow-up work items into the **Stacked Work Queue for Soma** tray. When you click **Approve & Execute** on an active proposal (or click **`Run Next`** in the queue tray), Soma automatically drains and dispatches the next stacked work item.

## 2. Governed Confirmation & 12 Deliverable Bundles

Clicking **Approve & Execute** (`POST /api/v1/intent/confirm-action`) executes the governed action, manifests any planned teams (with an automatic **Ideal Delivery Reviewer** `council-sentry` and companion `<team_id>-review-team` for complex projects), and materializes verified multi-file deliverable packages on disk:

| Deliverable Package | Template / Kind ID | Materialized Files |
| --- | --- | --- |
| Developer Project / Web App | `dev_project_bundle` | `index.html` (runnable HTML5 app), `README.md`, `implementation_notes.md`, `proof.md` |
| Interacting Multi-App Suite | `multi_app_bundle` | `index.html` (Dual-App Workbench), `app_interaction_contract.json`, `multi_team_architecture.md`, `proof.md` |
| Standard / Custom MCP Addition | `mcp_addition_bundle` | `mcp_binding_manifest.yaml`, `mcp_integration_guide.md`, optional `custom_mcp_server_scaffold.py`, `proof.md` |
| Team Execution Plan | `team_execution_bundle` | `team_execution_plan.md`, `next_steps_action_matrix.md`, `team_handoff_summary.json`, `proof.md` |
| Chained Multi-Team Pipeline | `chained_pipeline_bundle` | `pipeline_manifest.yaml`, `channel_handoff_contract.md`, `downstream_verification_checklist.md`, `proof.md` |
| Multi-Team Mission Blueprint | `mission_blueprint_bundle` | `mission_blueprint.yaml`, `governance_and_approval_gates.md`, `team_handoff_and_verification.md`, `proof.md` |
| Marketing Launch Kit | `marketing_launch_bundle` | `positioning_brief.md`, `launch_email_and_copy.md`, `customer_faq.md`, `proof.md` |
| Playable Browser Game | `interactive_game_bundle` | `index.html` (Sliding-Tile Puzzle or Street Fighter II Tetris), design & QA docs, `proof.md` |
| Customer Onboarding Kit | `customer_onboarding_bundle` | `welcome_and_kickoff_guide.md`, `first_30_days_checklist.md`, `escalation_and_faq.md`, `proof.md` |
| Interactive KPI Dashboard | `kpi_dashboard_bundle` | `index.html`, `metrics_definitions.md`, `data_sources_and_governance.md`, `proof.md` |
| Product Spec & Wireframe | `product_wireframe_bundle` | `index.html`, `product_spec.md`, `user_flows_and_acceptance.md`, `proof.md` |
| Engineering Task Pack | `codex_task_bundle` | `task.md`, `validation-checklist.md`, `affected-files.md`, `recommended-slices.md`, `proof.md` |

Every confirmed package records an **`OutputProofEnvelope`** verifying SHA-256 checksum, byte count, workspace path-boundary safety, artifact ID linkage, and source run/contract lineage.

## 3. Durable Files, Live HTML5 Preview, Follow-Up Queueing & Deployments Hub

- **Output Workbench (`OutputWorkbench.tsx` in Chat):** Displays the 5-stage package verification strip, `View Team & Outputs ->`, `Download / View Files`, and **`+ Stack Follow-Up`** (`output-workbench-queue-followup`) to queue a next-iteration enhancement directly into Soma's work queue.
- **Durable Files (`/resources`):** Inspect retained deliverables with package folder badges, **`Play / Live Preview`** (`<iframe sandbox="allow-scripts">`), **`Ask Soma to Refine`** inline iteration, **`+ Queue Follow-Up`** (`queue-followup-deliverable-*`), and **`Export Bundle`**.
- **Deployments & Team Hub (`/deployments`):** Click **`+ Start a New Team in Soma`** (`deployments-start-new-team`) to spin up a new goal-named team and execution package, inspect standing and runtime teams, verify the **Ideal Reviewer & Delivery Review Gate**, trigger 1-click actions (`Work with Team in Soma`, `Plan & Execute Next Steps`, `Have Another Team Build Interacting App ->`, `Chain This Team's Output`), open the inline **Deliverable Preview Drawer** (`DeploymentOutputPreviewDrawer` with the **Review Team Audit Certificate**), and bind **Standard Deployment MCPs** or **Custom Developed MCPs**. Every confirmed team or deliverable run records structured `conversation_turns` (and synthesizes fallback turns from `mission_events`) so `/runs/{id}` always displays the full execution conversation log.

## 4. AI Organizations, Ideal Delivery Reviewer & Learning Loops

All 4 company starter templates (`engineering-starter`, `marketing-starter`, `research-starter`, and `operations-starter`) render in a compact 2-column selector with a full **Target Delivery & Execution Breakdown** and instantiate with bundle-defined **`review`** and **`learning`** loop profiles (`GET /api/v1/organizations/{id}/learning-insights`).

## 5. Trusted-Memory Arbitration & Promotion

The **Memory & Continuity** surface (`/memory`) and `GET /api/v1/memory/search` enforce retrieval-time **Trusted-Memory Arbitration** across five canonical memory layers (`GOVERNED_DOCTRINE` > `PROJECT_MEMORY` > `AGENT_MEMORY` > `SOMA_MEMORY` > `REFLECTION_MEMORY`). Operators can promote `candidate` memories (`POST /api/v1/memory/promote`) and resolve `contradicted` memory halts (`POST /api/v1/memory/resolve-conflict`).

## 6. Multi-Role RBAC, Web Agentry Presets & Navigation Modes

- **Multi-Role RBAC & OIDC/JWT Federation:** Enforces `owner`, `admin`, `operator`, and `viewer` roles across API keys and federated OIDC/JWT Bearer tokens (`MYCELIS_OIDC_*`).
- **1-Click Web Agentry Service Presets (`/settings` -> `AI Engines`):** Pre-filled templates and secret-reference (`api_key_env`) wiring for `OpenAI`, `Anthropic Claude`, `Google Gemini`, `OpenRouter`, `Groq Cloud`, `DeepSeek`, `Mistral AI`, and `xAI Grok` alongside self-hosted `Ollama` / `vLLM`.
- **Calm Default vs. Advanced Mode:** Default Operator Mode shows `Soma Work` (`/dashboard`), `Teams` (`/deployments`), `Durable Files` (`/resources`), and `Settings` (`/settings`); Advanced Mode unlocks `/activity`, `/runs`, `/memory`, `/groups`, `/automations`, `/system`, and `/docs`.

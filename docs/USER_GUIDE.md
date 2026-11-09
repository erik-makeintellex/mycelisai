# Mycelis User & Operator Guide
> Navigation: [Project README](../README.md) | [Docs Home](README.md) | [User Docs Home](user/README.md)

Mycelis is an AI Organization platform where you collaborate with **Soma** to plan, confirm, and materialize governed deliverables without managing raw backend infrastructure.

## TOC

- [1. Start With Soma & Deliverable Starters](#1-start-with-soma--deliverable-starters)
- [2. Governed Confirmation & Deliverable Bundles](#2-governed-confirmation--deliverable-bundles)
- [3. Durable Files, Live HTML5 Preview & Iteration](#3-durable-files-live-html5-preview--iteration)
- [4. AI Organizations & Learning Loops](#4-ai-organizations--learning-loops)
- [5. Trusted-Memory Arbitration & Promotion](#5-trusted-memory-arbitration--promotion)
- [6. Multi-Role RBAC & Navigation Modes](#6-multi-role-rbac--navigation-modes)

## 1. Start With Soma & Deliverable Starters

When you open the Soma workspace (`/dashboard` or `/organizations/[id]`), Soma's **Start with Soma** suggestion bar surfaces **Popular deliverable starters** designed for both non-expert users and operators:

- **Marketing Launch Kit** (`marketing_launch_bundle`)
- **Playable Puzzle Game** (`interactive_game_bundle`)
- **Customer Onboarding Kit** (`customer_onboarding_bundle`)
- **Interactive KPI Dashboard** (`kpi_dashboard_bundle`)
- **Product Spec & Wireframe** (`product_wireframe_bundle`)
- **Engineering Task Pack** (`codex_task_bundle`)

Selecting a starter or typing a natural-language goal routes through Soma:
- **Direct Answer (`answer`):** Read-only questions, summaries, and explanations return immediately in chat.
- **Governed Proposal (`proposal`):** Actions that write files, launch team workflows, or mutate state present an outcome-first proposal card showing plain-English intent, planned deliverables, and **Confirm** / **Cancel** buttons. Technical telemetry (`run_id` UUIDs, token counts, capability IDs, and risk badges) stays tucked behind **`[View Proof Details]`**.

## 2. Governed Confirmation & Deliverable Bundles

Clicking **Confirm** (`POST /api/v1/intent/confirm-action`) executes the governed action and materializes a real multi-file deliverable bundle on disk under the workspace output block:

| Deliverable Starter | Template ID | Materialized Files |
| --- | --- | --- |
| Marketing Launch Kit | `marketing_launch_bundle` | `positioning_brief.md`, `launch_email_and_copy.md`, `customer_faq.md`, `proof.md` |
| Playable Puzzle Game | `interactive_game_bundle` | `index.html`, `game_design_and_rules.md`, `qa_and_controls.md`, `proof.md` |
| Customer Onboarding Kit | `customer_onboarding_bundle` | `welcome_and_kickoff_guide.md`, `first_30_days_checklist.md`, `escalation_and_faq.md`, `proof.md` |
| Interactive KPI Dashboard | `kpi_dashboard_bundle` | `index.html`, `metrics_definitions.md`, `data_sources_and_governance.md`, `proof.md` |
| Product Spec & Wireframe | `product_wireframe_bundle` | `index.html`, `product_spec.md`, `user_flows_and_acceptance.md`, `proof.md` |
| Engineering Task Pack | `codex_task_bundle` | `task.md`, `validation-checklist.md`, `affected-files.md`, `recommended-slices.md`, `proof.md` |
| General Soma Draft | fallback draft | `deliverables/soma-draft.md`, `proof.md` |

Every confirmed package records an **`OutputProofEnvelope`** verifying SHA-256 checksum, byte count, workspace path-boundary safety, artifact ID linkage, and source run/contract lineage.

## 3. Durable Files, Live HTML5 Preview & Iteration

Navigate to **Durable Files** (`/resources`, backed by `DurableFilesOperatorView`) to browse, inspect, iterate on, and export retained deliverables:

- **Preview / Open / Download / Export Bundle:** Inspect Markdown, JSON, and text deliverables with full proof metadata, download individual artifacts, or click **`Export Bundle`** to download a combined workspace deliverables bundle (`workspace-deliverables-bundle.md`).
- **`Play / Live Preview` (Interactive HTML5 Sandbox):** When inspecting an HTML5 deliverable (`index.html` from a **Playable Puzzle Game**, **Interactive KPI Dashboard**, or **Product Spec & Wireframe**), click **`Play / Live Preview`** to toggle into a sandboxed `<iframe sandbox="allow-scripts">`. You can play the puzzle game, filter KPI charts, or switch wireframe tabs directly in-app while scripts remain isolated from host credentials.
- **`Ask Soma to Refine` (Inline Iteration):** Click **`Ask Soma to Refine`** on any deliverable card to open an inline revision input and send targeted follow-up instructions to Soma without leaving `/resources`.

## 4. AI Organizations & Learning Loops

All 4 company starter templates (`engineering-starter`, `marketing-starter`, `research-starter`, and `operations-starter`) instantiate with bundle-defined **`review`** and **`learning`** loop profiles:

- **Review Loops:** Periodically evaluate team outputs against organization quality gates.
- **Learning Loops (`GET /api/v1/organizations/{id}/learning-insights`):** Capture candidate lessons and reflection insights from completed runs for operator-reviewed promotion—never silently rewriting core prompts or doctrine.

## 5. Trusted-Memory Arbitration & Promotion

The **Memory & Continuity** surface (`/memory`) and `GET /api/v1/memory/search` enforce retrieval-time **Trusted-Memory Arbitration** across five canonical memory layers:

1. `GOVERNED_DOCTRINE` (highest precedence, operator-governed rules and constitution)
2. `PROJECT_MEMORY` (curated organization and project facts)
3. `AGENT_MEMORY` (specialist and role-scoped working memory)
4. `SOMA_MEMORY` (conversational continuity and session context)
5. `REFLECTION_MEMORY` (candidate insights from learning loops)

In `/memory` (`MemoryDetailPanel`), operators can filter by memory layer, promote `candidate` memories to verified `PROJECT_MEMORY` or `GOVERNED_DOCTRINE` (`POST /api/v1/memory/promote`), and resolve `contradicted` memory halts (`Approve & Clear Conflict` or `Mark Superseded` via `POST /api/v1/memory/resolve-conflict`).

## 6. Multi-Role RBAC & Navigation Modes

Mycelis enforces multi-role RBAC across the UI and API (`owner`, `admin`, `operator`, `viewer`):

- **`owner` / `admin`:** Full access to AI Engine Settings, user/role administration, template cloning/applying, and governed action confirmation.
- **`operator`:** Can collaborate with Soma, confirm standard deliverable proposals, and manage active teams/files.
- **`viewer`:** Read-only inspection of workspaces, deliverables, and activity timelines.

**Calm Default vs. Advanced Mode:**
- **Default Operator Rail (`advancedMode === false`):** Shows `Soma Work` (`/dashboard`), `Teams` (`/teams`), `Durable Files` (`/resources`), and `Settings` (`/settings`).
- **Advanced Mode (`advancedMode === true`):** Unlocks `Activity History` (`/activity`), `Runs` (`/runs`), `Memory & Continuity` (`/memory`), `Groups` (`/groups`), `Deployments` (`/deployments`), `Automations` (`/automations`), `System` (`/system`), and `Docs` (`/docs`). All UI components comply with the strict `<= 300` line modularity standard (`0` UI component legacy caps remaining).

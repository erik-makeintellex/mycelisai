# Mycelis User & Operator Guide
> Navigation: [Project README](../README.md) | [Docs Home](README.md) | [User Docs Home](user/README.md)

Mycelis is an AI Organization platform where you collaborate with **Soma** to plan, confirm, and materialize governed deliverables without managing raw backend infrastructure.

## TOC

- [1. Start With Soma & Deliverable Starters](#1-start-with-soma--deliverable-starters)
- [2. Governed Confirmation & Deliverable Bundles](#2-governed-confirmation--deliverable-bundles)
- [3. Durable Files & Live HTML5 Preview](#3-durable-files--live-html5-preview)
- [4. AI Organizations & Learning Loops](#4-ai-organizations--learning-loops)
- [5. Trusted-Memory Arbitration](#5-trusted-memory-arbitration)
- [6. Multi-Role RBAC & Navigation Modes](#6-multi-role-rbac--navigation-modes)

## 1. Start With Soma & Deliverable Starters

When you open the Soma workspace (`/dashboard` or `/organizations/[id]`), Soma's **Start with Soma** suggestion bar surfaces **Popular deliverable starters** designed for both non-expert users and operators:

- **Marketing Launch Kit** (`marketing_launch_bundle`)
- **Playable Puzzle Game** (`interactive_game_bundle`)
- **Customer Onboarding Kit** (`customer_onboarding_bundle`)
- **Interactive KPI Dashboard** (`kpi_dashboard_bundle`)
- **Engineering Task Pack** (`codex_task_bundle`)

Selecting a starter or typing a natural-language goal routes through Soma:
- **Direct Answer (`answer`):** Read-only questions, summaries, and explanations return immediately in chat.
- **Governed Proposal (`proposal`):** Actions that write files, launch team workflows, or mutate state present an outcome-first proposal card showing plain-English intent, planned deliverables, and **Confirm** / **Cancel** buttons. Technical telemetry (`run_id` UUIDs, token counts, capability IDs, and risk badges) stays tucked behind **`[View Proof Details]`**.

## 2. Governed Confirmation & Deliverable Bundles

Clicking **Confirm** (`POST /api/v1/intent/confirm-action`) executes the governed action and materializes a real multi-file deliverable bundle on disk under the workspace output block:

| Deliverable Starter | Template ID | Materialized Files |
| --- | --- | --- |
| Marketing Launch Kit | `marketing_launch_bundle` | `campaign-brief.md`, `landing-page.html`, `channel-plan.md`, `proof.md` |
| Playable Puzzle Game | `interactive_game_bundle` | `index.html`, `game-design.md`, `proof.md` |
| Customer Onboarding Kit | `customer_onboarding_bundle` | `welcome-guide.md`, `onboarding-checklist.html`, `escalation-playbook.md`, `proof.md` |
| Interactive KPI Dashboard | `kpi_dashboard_bundle` | `index.html`, `executive-summary.md`, `metrics-snapshot.json`, `proof.md` |
| Engineering Task Pack | `codex_task_bundle` | `task.md`, `validation-checklist.md`, `affected-files.md`, `recommended-slices.md`, `proof.md` |
| General Soma Draft | fallback draft | `deliverables/soma-draft.md`, `proof.md` |

Every confirmed package records an **`OutputProofEnvelope`** verifying SHA-256 checksum, byte count, workspace path-boundary safety, artifact ID linkage, and source run/contract lineage.

## 3. Durable Files & Live HTML5 Preview

Navigate to **Durable Files** (`/resources`, backed by `DurableFilesOperatorView`) to browse, inspect, and interact with retained deliverables:

- **Preview / Open / Download:** Inspect Markdown, JSON, and text deliverables with full proof metadata (`OutputProofEnvelope` verification badge, SHA-256 hash, byte size, and creation timestamp).
- **`Play / Live Preview` (Interactive HTML5 Sandbox):** When inspecting an HTML5 deliverable (such as `index.html` from a **Playable Puzzle Game** or **Interactive KPI Dashboard**), click **`Play / Live Preview`** in the modal header to toggle from source view into a sandboxed `<iframe sandbox="allow-scripts">`. You can play the puzzle game or filter live KPI charts directly in-app while scripts remain isolated from host credentials and parent storage.

## 4. AI Organizations & Learning Loops

All 4 company starter templates (`engineering-starter`, `marketing-starter`, `research-starter`, and `operations-starter`) instantiate with bundle-defined **`review`** and **`learning`** loop profiles:

- **Review Loops:** Periodically evaluate team outputs against organization quality gates.
- **Learning Loops (`GET /api/v1/organizations/{id}/learning-insights`):** Capture candidate lessons and reflection insights from completed runs for operator-reviewed promotion—never silently rewriting core prompts or doctrine.

## 5. Trusted-Memory Arbitration

The **Memory & Continuity** surface (`/memory`) and `GET /api/v1/memory/search` enforce retrieval-time **Trusted-Memory Arbitration** across five canonical memory layers:

1. `GOVERNED_DOCTRINE` (highest precedence, operator-governed rules and constitution)
2. `PROJECT_MEMORY` (curated organization and project facts)
3. `AGENT_MEMORY` (specialist and role-scoped working memory)
4. `SOMA_MEMORY` (conversational continuity and session context)
5. `REFLECTION_MEMORY` (candidate insights from learning loops)

In `/memory`, operators can filter by memory layer, inspect verification badges, and review automatic **contradiction warnings** whenever lower-precedence memories conflict with `GOVERNED_DOCTRINE`.

## 6. Multi-Role RBAC & Navigation Modes

Mycelis enforces multi-role RBAC across the UI and API (`owner`, `admin`, `operator`, `viewer`):

- **`owner` / `admin`:** Full access to AI Engine Settings, user/role administration, template cloning/applying, and governed action confirmation.
- **`operator`:** Can collaborate with Soma, confirm standard deliverable proposals, and manage active teams/files.
- **`viewer`:** Read-only inspection of workspaces, deliverables, and activity timelines.

**Calm Default vs. Advanced Mode:**
- **Default Operator Rail (`advancedMode === false`):** Shows `Soma Work` (`/dashboard`), `Teams` (`/teams`), `Durable Files` (`/resources`), and `Settings` (`/settings`).
- **Advanced Mode (`advancedMode === true`):** Unlocks `Activity History` (`/activity`), `Runs` (`/runs`), `Memory & Continuity` (`/memory`), `Groups` (`/groups`), `Deployments` (`/deployments`), `Automations` (`/automations`), `System` (`/system`), and `Docs` (`/docs`). All UI components comply with the strict `<= 300` line modularity standard (`0` UI component legacy caps remaining).

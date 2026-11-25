# Deployments

> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> One page answers "what is running, where, with which models and tools, under which approval policy, since when, and what did it produce". Open `/deployments` any time you want that answer; every user can see it, and it never changes anything on its own.

---

## TOC

- [Overview](#overview)
- [The Four Kinds](#the-four-kinds)
- [Reading A Row](#reading-a-row)
- [The Detail Pane](#the-detail-pane)
- [Source Notices](#source-notices)
- [View Config (Advanced)](#view-config-advanced)
- [Clone And Apply Config (Advanced)](#clone-and-apply-config-advanced)
- [Where Team Creation Moved](#where-team-creation-moved)

## Overview

`/deployments` lists every standing team, mission team, group, and planned department in one place, and serves as both the live status view and the actionable team & delivery hub. Click **`+ Start a New Team in Soma`** (`deployments-start-new-team`) in the header to launch a new team directly in Soma with a prefilled goal prompt, or pick a filter chip to narrow by kind or state, then select a row to open its detail pane, inspect outputs inline, add MCP capabilities, or dispatch follow-on work through Soma.

`/teams` now redirects here. Deployments is in the main navigation rail for every user, not just Advanced mode.

## The Four Kinds

| Kind badge | Meaning |
| --- | --- |
| Standing team | A team loaded from the boot bundle. It runs continuously and is not tied to one mission. |
| Mission team | A team created for one mission. It is scoped to that mission and its lifecycle follows the mission. |
| Group | A governed work scope (work mode, approval policy, optional expiry) that can include teams and people. Only visible to operators with group access; see [Source Notices](#source-notices) when it is hidden. |
| Planned | A department from an organization's template or starter that has not launched yet. Nothing is running for it; the operator still has to approve a launch. |

## Reading A Row

Each row shows the deployment's name, its kind badge, lifecycle state (running, idle, planned, paused, archived, expired, or error), how many members it has, which provider/model combinations are actually in use, and when it last had activity.

## The Detail Pane

Selecting a row opens a detail pane with:

- scope and origin (where the deployment lives and how it was created)
- members, each with role, provider/model, and tools
- **Universal Ideal Delivery Reviewer & Delivery Review Gate:** every standing and runtime team includes an Ideal Reviewer (`council-sentry`, `IDEAL REVIEWER · DELIVERY GATE`), and complex project/deliverable asks automatically coordinate a companion `<team_id>-review-team`
- **Actionable Team Engagement CTAs:**
  - `Work with Team in Soma` and `Plan & Execute Next Steps` (`team_execution_bundle`)
  - `Have Another Team Build Interacting App ->` (`multi_app_bundle`)
  - `View Deliverables` and `Chain This Team's Output to Another Team in Soma ->` (`chained_pipeline_bundle`)
- **Generated Team Outputs & Inline Preview Drawer:** retained outputs show a count and proof status (or **unknown** when Core has not reported a summary yet — never shown as zero, because zero and unknown mean different things), and clicking `View / Open` opens `DeploymentOutputPreviewDrawer` with playable HTML5 `<iframe sandbox="allow-scripts">` / source preview and the **Review Team Audit Certificate** (`VERIFIED BY IDEAL REVIEWER & DELIVERY REVIEW TEAM`)
- **MCP Additions (`Standard Deployment MCPs` & `Custom Developed MCPs`):** 1-click `+ Add to Deployment` and `+ Scaffold & Bind Custom MCP` (`mcp_addition_bundle`)
- work mode and approval posture, written as one plain-language sentence
- since and last-activity timestamps

Advanced mode adds origin, scope, and provider-origin detail (for example, whether a member's model came from the manifest, a routing profile, or a fallback).

## Source Notices

Deployments is built from several sources (standing teams, mission teams, groups, planned departments, and retained outputs). If one of them is degraded, unavailable, or restricted, a small notice appears above the list, for example "Groups hidden: requires admin access." The rest of the page keeps working; a notice never blocks you from seeing the deployments that did load. If Core itself cannot be reached, the page shows a clear failure state with a Retry action instead of an empty list.

## View Config (Advanced)

With Advanced mode on, a "View config" button appears in the detail pane. It opens an inline panel showing the deployment's underlying configuration as read-only, verbatim text: its team manifest, group definition, or organization template (a planned department also resolves to its organization's config, and the panel labels itself "Organization config" so that's clear). It is hidden in simple mode because most operators only need to know what is running, not its raw configuration.

The panel has a YAML/JSON toggle, Copy and Download buttons (downloads as `<kind>-<id>.yaml` or `.json`), and a source line naming where the document came from (file, database, or runtime, plus the file path when there is one). A "Hidden values" list names every path the server redacted before rendering; secrets such as API keys are never shown, while environment-variable references (`${env:NAME}`) stay visible so you can tell what's configured without seeing the value. If the deployment's config needs admin access you don't have, the panel says so instead of showing an error; if Core can't be reached, it shows a Retry action.

## Clone And Apply Config (Advanced)

With Advanced mode on, the View config panel can also add two buttons that let you write configuration, not just read it. Both are governed: nothing changes until you explicitly approve a preview.

Deployments only tracks team, group, and organization configs, so template configs aren't opened from here. To view or write a template's config, go to organization setup, pick "Start from template," and use the "View config" action on a starter card (also Advanced mode only): file and bundle starters offer **Clone as custom template**, and Custom (operator) starters offer **Edit and apply**. Groups still open their View config panel from Deployments as described above.

- **Clone as custom template** appears on a file or bundle template. Name the clone, preview it, review the diff and validation, then approve. The clone is stored as an operator template and shows up as a **Custom** starter (see [Core Concepts](core-concepts.md#creating-an-ai-organization)) the next time you set up an AI Organization.
- **Edit and apply** appears on a group or on a template you already cloned. It opens a plain-text editor with the current document, and always previews before it lets you approve: the preview shows which paths changed, whether the document is valid, and disables Approve when the change is blocked.

If you edit the document again after previewing, the preview is no longer valid for what you approve, so the panel asks you to preview again before Approve is enabled. If someone else changed the stored document in the meantime, approving tells you the document changed and previews it again automatically so you can see the current version before deciding.

Secrets can never be written back through either flow: a document that still contains a hidden-value marker, or a literal secret where a reference belongs, is rejected before anything is stored, and the panel lists exactly which paths were rejected. If you do not have the role for this change, or the server cannot process governed writes right now, the panel says so plainly instead of showing a raw error.

## Where Team Creation Moved

Deployments is for reviewing what already exists. To create a new team, use the guided workflow at `/teams/create` or ask Soma directly; see [Teams](teams.md) for the team-shape guidance that still applies.

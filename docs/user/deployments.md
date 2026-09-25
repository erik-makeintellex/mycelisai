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
- [Where Team Creation Moved](#where-team-creation-moved)

## Overview

`/deployments` lists every standing team, mission team, group, and planned department in one place. Pick a filter chip to narrow by kind or state, then select a row to open its detail pane. The page is read-only: it explains what is running, it does not launch or stop anything.

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
- work mode and approval posture, written as one plain-language sentence
- since and last-activity timestamps
- retained outputs: a count and a proof status, or **unknown** when Core has not reported a summary yet (this is never shown as zero, because zero and unknown mean different things)

Advanced mode adds origin, scope, and provider-origin detail (for example, whether a member's model came from the manifest, a routing profile, or a fallback).

## Source Notices

Deployments is built from several sources (standing teams, mission teams, groups, planned departments, and retained outputs). If one of them is degraded, unavailable, or restricted, a small notice appears above the list, for example "Groups hidden: requires admin access." The rest of the page keeps working; a notice never blocks you from seeing the deployments that did load. If Core itself cannot be reached, the page shows a clear failure state with a Retry action instead of an empty list.

## View Config (Advanced)

With Advanced mode on, a "View config" action appears in the detail pane. It opens the deployment's underlying configuration (its team manifest, group definition, or organization template). It is hidden in simple mode because most operators only need to know what is running, not its raw configuration.

## Where Team Creation Moved

Deployments is for reviewing what already exists. To create a new team, use the guided workflow at `/teams/create` or ask Soma directly; see [Teams](teams.md) for the team-shape guidance that still applies.

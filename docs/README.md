# Documentation
> Navigation: [Project README](../README.md)

Use this page as the repo documentation map. It separates product docs, repo/operator docs, the single architecture PRD, and release/testing support.

The in-app Help area should stay user-first. Its default path should help someone talk with Soma, shape a request, approve work when needed, open outcomes, recover failures, revisit outputs, and use groups/resources/settings before exposing contributor or architecture-review material.

## Docs TOC

- [User Guidance](#user-guidance)
- [Repo Guidance](#repo-guidance)
- [Architecture Contracts](#architecture-contracts)
- [Testing And Release](#testing-and-release)
- [Agent And Maintainer Notes](#agent-and-maintainer-notes)

## User Guidance

These are the best entry points for someone using Mycelis through the product or in-app docs browser:

- **User Docs Home**: `./user/README.md`
- **Deployment Method Selection**: `./user/deployment-methods.md`
- **Core Concepts**: `./user/core-concepts.md`
- **Using Soma Chat**: `./user/soma-chat.md`
- **Workflow Variants And Plan Memory**: `./user/workflow-variants-and-plan-memory.md`
- **Teams**: `./user/teams.md`
- **Governance & Trust**: `./user/governance-trust.md`
- **Automations**: `./user/automations.md`
- **Resources**: `./user/resources.md`
- **Memory**: `./user/memory.md`
- **Settings And Access**: `./user/settings-access.md`
- **Authentication Modes**: `./user/auth-modes.md`
- **System Status & Recovery**: `./user/system-status-recovery.md`
- **Run Timeline**: `./user/run-timeline.md`

User guidance rules:
- lead with the action a user is trying to complete
- preserve Soma as a conversational counterpart for shaping work, not just a launcher for registered jobs
- keep proof, recovery, and output access close to the workflow they explain
- move implementation contracts, API details, raw topology, and old planning language into repo or architecture docs instead of the default user path
- when a UI slice changes an operator workflow, update the matching user doc and the in-app docs manifest in the same slice

## Repo Guidance

These are the active contributor support surfaces for changing or reviewing the repo. Product and architecture authority stays in the Canonical PRD unless a support doc is explicitly referenced from it.

- **Repository Entry Point**: `../README.md`
- **Operations**: `./architecture/OPERATIONS.md`
- **Local Dev Workflow**: `./LOCAL_DEV_WORKFLOW.md`
- **Cognitive Architecture Reference**: `./COGNITIVE_ARCHITECTURE.md`
- **API Reference**: `./API_REFERENCE.md`
- **Logging Standard**: `./logging.md`
- **Ops README**: `../ops/README.md`
- **Core README**: `../core/README.md`
- **Interface README**: `../interface/README.md`

## Architecture Contracts

The active architecture library is intentionally singular. It is not a holding area for old version notes, doctrine fragments, or temporary execution plans.

- **Architecture Docs Index**: `./architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
- **Mycelis Canonical PRD**: `./architecture-library/MYCELIS_CANONICAL_PRD.md`
- **Worker Library Source Map**: `./architecture-library/WORKER_LIBRARY_SOURCE_MAP.md`

## Testing And Release

Use these when the goal is verification, release proof, or workflow-complete validation:

- **Testing Guide**: `./TESTING.md`
- **Remote User Testing Runbook**: `./REMOTE_USER_TESTING.md`
- **Mycelis Canonical PRD**: `./architecture-library/MYCELIS_CANONICAL_PRD.md`
- **Governance System**: `./governance.md`
- **Licensing & Editions**: `./licensing.md`

## Agent And Maintainer Notes

These files are useful for Codex, maintainers, and architecture reviewers, but they should not be treated as ordinary user-facing product docs:

- **Repository Standards**: `../AGENTS.md`
- **Active Development State**: `../.state/V8_DEV_STATE.md`
- **Mycelis Canonical PRD**: `./architecture-library/MYCELIS_CANONICAL_PRD.md`

Keep these out of the normal operator path unless the task is repo maintenance, architecture review, or delivery-state inspection.

Guidance rules:
- user guidance should stay focused on using the product, visible outputs, and recovery paths, not implementation internals
- repo guidance should point to operating and implementation contracts, not old planning notes
- architecture guidance should remain singular, current, and directly referenced
- agent/maintainer notes should stay out of the default in-app operator docs unless they are intentionally surfaced for architecture review
- testing guidance should point to durable verification contracts rather than temporary execution notes
- if a removed doc contains a still-needed requirement, promote that requirement into the canonical PRD instead of restoring the old doc
- each subjective UI cleanup step should close with the docs touched, the docs reviewed unchanged, and the proof that the in-app Help entry still opens

# Market Position And Workflow Targets

Status: reference, informs delivery priority. Authority: [canonical PRD](MYCELIS_CANONICAL_PRD.md). Evidence: [scoreboard](../../.state/V8_DEV_STATE.md).
Source: mycelis-architect research round, 2026-09-27 (`ux/paperclip-comparison.md`, `ux/modern-workflows-judgment.md`), and the external repository [github.com/paperclipai/paperclip](https://github.com/paperclipai/paperclip) (MIT, launched March 2026).

This doc records competitive positioning and the workflow-quality bar the product is judged against. It adds no authority beyond the canonical PRD; it explains why the PRD's Trust, Conversation-First, and Delivery sections carry the weight they do, and it feeds the [truthful-delivery and next-architecture doc](TRUTHFUL_DELIVERY_AND_NEXT_ARCHITECTURE.md).

## Paperclip, in one line

Paperclip is "orchestration for zero-human companies": an org chart of BYO agents (Claude Code, Codex, Cursor, HTTP, plugins) with goals, projects, tasks, budgets, board approvals, an audit log, and company export/import. It explicitly says it is not a chatbot, not an agent framework, and not a workflow builder. Not yet shipped: memory/knowledge, a CEO chat interface, work queues, self-organization, ticket integrations.

## Where Mycelis leads

1. **A conversational front door with truthful delivery.** Soma is a chat-first interface with an inline proposal, preview, approve, and result loop; Paperclip has no chat surface at all (a CEO chat is on its roadmap). This only counts once outputs are real: the live probe (see the delivery doc) is the proof, not the claim.
2. **Maintained memory with citations.** Deployment context and memory exist in Mycelis today; recall and citations are landing through M1 (see the delivery doc). This is on Paperclip's roadmap and unshipped. It is the single biggest differentiator once M1 ships.
3. **Governed trust and sovereignty.** Tiered approvers, posture policies, purpose-bound tokens, fail-closed policy, proof readback, declared-tool enforcement, and a local-first root model with per-profile routing and LOCAL_ONLY/LEAVES_ORG boundaries. Paperclip has board approvals and an audit log, but no data-sovereignty story; models come from the agents the operator brings in.

## Where Paperclip leads (and what to borrow)

| Area | Paperclip | Mycelis today | Action |
| --- | --- | --- | --- |
| Cost control | Per-agent/project/model token and cost tracking with **hard-stop budgets** | Token rate only, no budgets UX | Borrow: B1 (token budgets keyed to model defaults, admin-modifiable, hard stops) |
| Work structure | Goal -> project -> task with ancestry and blocker dependencies | Outcomes/groups/teams; weak goal ancestry, no dependency graph in the UX | Borrow: goal ancestry on work, plus simple blocker dependencies |
| Scheduling/autonomy | Heartbeats, DB queue, atomic checkout, resumed context | A 5-step "Automation Chain" behind Advanced, not conversational | Borrow: heartbeat-style recurring work exposed conversationally ("every Monday...") |
| Portability | Company export/import with secret scrubbing | Organization starters and posture templates, no full export/import | Borrow: organization export/import with secret scrubbing |
| Delegation | Up/down an org chart | `delegate_task` is queue-only | H1 closes this gap with real handoffs, provenance, and ack |
| Ecosystem | 43K+ stars, many BYO adapters | Early, native swarm agents only | See interop below |

Paperclip also leads on skill evals (a skill studio) and multi-tenancy (many companies per deployment); these are lower priority than the four borrows above.

## Interop, not rivalry

Paperclip treats any heartbeat-capable HTTP agent as "hireable." A **Mycelis HTTP adapter** would let a Paperclip company hire a governed Mycelis team (Soma plus memory plus proof), giving Mycelis reach into a large existing community while keeping its own strengths. The reverse direction — Mycelis admitting Paperclip-style external agents — is the blocked C2 lane in the canonical PRD's runtime architecture; C2's contract gates (run identity, replay, approval, stop) are exactly what makes admitting external agents safe, so C2 is the prerequisite, not a separate project.

## The modern workflow target model

External 2025-26 AI-workspace products set a bar: everything starts in conversation; plans/previews show inline with one-click approve; background work reports progress and notifies on completion; results are first-class (library, search, open, revise, version); connectors/skills come from a gallery with least-privilege consent; recurring work is set up in natural language; there is one approvals inbox; a command palette exists; advanced configuration sits behind progressive disclosure.

Scored 0-4 (Actuation: 0 impossible/dead end, 4 modern; Wording: 0 opaque, 4 plain), against live dev evidence (`ux/modern-workflows-judgment.md`, dev @244af327):

| # | Workflow | Act. | Word. | Real? |
| --- | --- | --- | --- | --- |
| 1 | Ask a question | 4 | 2 | Real |
| 2 | Get a deliverable made | 1 | 1 | Fake (template content, blank viewer) |
| 3 | Revise a result | unverified | 3 | unverified |
| 4 | Find and reuse past results | 1 | 1 | Dead end (no results library) |
| 5 | Launch a team on a goal | 0 (dev) | 1 | Broken (`token_wrong_purpose`) |
| 6 | Recurring/scheduled work | 1 | 0 | Unverified, not conversational |
| 7 | Approve or deny waiting work | 1 | 1 | False success (fixed in UX1) |
| 8 | Connect a tool or skill | 1 | 1 | Partly name-only |
| 9 | Choose or check the AI engine | 2 | 1 | Real, but buried |
| 10 | Manage people and access | 2 | 1 | Real |
| 11 | Get help in context | 3 | 3 | Real |
| 12 | Know what's happening | 2 | 1 | Real |

Average: 1.6 actuation, 1.4 wording across the 11 scored workflows. Only "ask a question" and "get help" meet the modern bar. Missing modern primitives confirmed by code scan: no command palette, no completion/approval notifications, no results search.

## Top 5 changes, by leverage

1. **Truthful delivery end to end.** Nothing else matters if outputs are fake. See the [delivery doc](TRUTHFUL_DELIVERY_AND_NEXT_ARCHITECTURE.md) for status.
2. **One Results library.** Deliverables lists every retained output from Core's proof/output records, searchable and openable, no MCP prerequisite.
3. **Conversation-first setup for teams, schedules, and connectors.** Soma proposes "a team for X," "every Monday do Y," "connect Z to team W" as normal previews with approve; forms/chains move behind Advanced.
4. **An approvals inbox and notifications.** One rail entry with a badge, toasts/notifications on completion or when approval is needed.
5. **A vocabulary and progressive-disclosure pass** over Capabilities, AI Engines, Automations, and Groups: plain summary first, raw tables behind Advanced; split Help from contributor docs; add a command palette.

## Citations

External source: [github.com/paperclipai/paperclip](https://github.com/paperclipai/paperclip). Descriptions above are summarized, not reproduced, from its README and third-party coverage (jimmysong.io, zeabur blog).

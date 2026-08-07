# Run Timeline
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

A run is one execution instance supporting an Outcome. Most users should follow work from Soma and the Outcome; the run receipt and event timeline are Admin tools for proof, recovery, and support.

## When To Open A Run

Open a run when you need to:

- verify what execution occurred
- understand why work is still running or degraded
- inspect retained output and proof references
- recover a failed or incomplete attempt
- provide diagnostics to an operator or developer

Use the Outcome or deliverable view when you only need the result.

## Opening A Run

Run links may appear in Soma details, Outcome proof, Activity, Groups workflow logs, or recovery information. A direct `/runs/{run_id}` URL is bookmarkable for authorized users.

The `/runs` list and `/activity` are available through **Admin tools**. Activity provides the cleaner operating summary; the dedicated run page provides the full receipt and timeline.

## Read The Receipt First

The run page begins with a receipt that translates runtime evidence into operator questions:

| Receipt field | What it answers |
| --- | --- |
| What happened | The useful execution outcome or failure summary |
| What to trust | Whether output and proof are reliable, provisional, missing, or invalid |
| Next step | Review, wait, retry, recover, or inspect further |
| Output refs | Durable deliverables linked to this run |
| Proof refs | Validation, audit, or proof records supporting the result |

Use **Inspect receipt evidence** for the approved work mode, expected deliverable, lifecycle guidance, runtime identifiers, and exact references.

## Run Status

| Status | Meaning |
| --- | --- |
| Pending | Accepted but not yet executing |
| Running | Execution is active |
| Completed | Execution reached a successful terminal state |
| Degraded | Work settled with missing, failed, or uncertain requirements that need recovery |
| Failed | Execution terminated without a usable result |

Run status supports Outcome health but does not replace it. For example, a run can complete while an Outcome remains degraded because a required deliverable or validation record is missing.

## Missing Output Is Not Completion

When an approved contract requires a retained deliverable but no output reference exists, the receipt shows **Run needs output recovery**. The recorded execution remains available, but the missing deliverable must not be presented as trusted completion.

Recover or rerun the owning work from Soma, the Outcome review surface, Groups, or the run recovery action.

## Reading The Timeline

The timeline is an ordered reconstruction of execution events. While a run is active, the view refreshes for new evidence.

Each event can include:

- event type and severity
- source component or agent
- timestamp
- concise summary
- expandable structured payload

Common event families include execution started/completed/degraded, tool invocation, artifact creation, memory/context use, proof creation, and recovery actions. Exact payloads remain Inspect detail because they are implementation evidence, not the primary product result.

## Trust And Recovery

For a degraded or failed run, confirm these in order:

1. What work actually completed?
2. Which deliverables can still be opened?
3. Which proof is missing or invalid?
4. Is an external mutation uncertain?
5. Can Soma retry safely with the retained contract and context?
6. Does an operator need to change a dependency, permission, or request?

Do not infer exactly-once external mutation merely because message delivery was deduplicated. When the external result is ambiguous, follow the explicit recovery guidance instead of silently rerunning.

## Navigation

- **Soma** returns to `/dashboard` for conversation and steering.
- **Activity** opens the Admin tools summary of active and recent execution.
- **Groups** opens producing-team workflow logs and retained outputs.
- **Resources** opens durable files and project packages.
- **Details** or **Inspect** exposes runtime identifiers and raw evidence only when needed.

See also:

- [Using Soma Chat](soma-chat.md)
- [Governance & Trust](governance-trust.md)
- [System Status & Recovery](system-status-recovery.md)

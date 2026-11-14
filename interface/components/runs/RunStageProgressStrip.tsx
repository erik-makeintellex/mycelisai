"use client";

import React from "react";
import type { MissionEvent } from "@/store/useCortexStore";

export type RunStageStatus = "completed" | "active" | "pending" | "failed";

export interface RunLifecycleStage {
    id: string;
    label: string;
    status: RunStageStatus;
    timestamp?: string;
    summary?: string;
    actor?: string;
}

export function buildRunLifecycleStages(events: MissionEvent[]): RunLifecycleStage[] {
    const stages: RunLifecycleStage[] = [
        { id: "intent_received", label: "Intent Received", status: "pending" },
        { id: "proposal_generated", label: "Proposal Generated", status: "pending" },
        { id: "approval_required", label: "Approval Gate", status: "pending" },
        { id: "approval_granted", label: "Approval Granted", status: "pending" },
        { id: "mission_running", label: "Mission Running", status: "pending" },
        { id: "tool_action_execution", label: "Tool / Action Execution", status: "pending" },
        { id: "package_generated", label: "Package Generated", status: "pending" },
        { id: "proof_generated", label: "Proof Verified", status: "pending" },
        { id: "recovery_next_actions", label: "Recovery & Next Actions", status: "pending" },
    ];
    if (!events || events.length === 0) return stages;

    let intentEvt: MissionEvent | undefined = events[0];
    let proposalEvt: MissionEvent | undefined;
    let approvalReq: MissionEvent | undefined;
    let approvalGrant: MissionEvent | undefined;
    let approvalDeny: MissionEvent | undefined;
    let runningEvt: MissionEvent | undefined;
    let toolInvoked: MissionEvent | undefined;
    let toolDone: MissionEvent | undefined;
    let toolFailed: MissionEvent | undefined;
    let artifactEvt: MissionEvent | undefined;
    let proofEvt: MissionEvent | undefined;
    let completedEvt: MissionEvent | undefined;
    let failedEvt: MissionEvent | undefined;
    let invokedCount = 0;
    let doneCount = 0;

    for (const ev of events) {
        const t = ev.event_type;
        if (t === "mission.started" || t === "trigger.fired") {
            intentEvt = ev;
            if (!runningEvt) runningEvt = ev;
        } else if (t === "team.spawned" || t === "template.cloned" || t === "proposal.generated") {
            if (!proposalEvt) proposalEvt = ev;
            runningEvt = ev;
        } else if (t === "approval.requested") {
            approvalReq = ev;
            if (!proposalEvt) proposalEvt = ev;
        } else if (t === "approval.granted" || t === "config.applied") {
            approvalGrant = ev;
        } else if (t === "approval.denied") {
            approvalDeny = ev;
        } else if (t === "agent.started" || t === "agent.heartbeat") {
            runningEvt = ev;
        } else if (t === "tool.invoked") {
            invokedCount++;
            toolInvoked = ev;
            if (!proposalEvt) proposalEvt = ev;
            if (!runningEvt) runningEvt = ev;
        } else if (t === "tool.completed") {
            doneCount++;
            toolDone = ev;
        } else if (t === "tool.failed") {
            toolFailed = ev;
        } else if (t === "artifact.created") {
            artifactEvt = ev;
            if (ev.payload?.proof_id || ev.payload?.proof_path) proofEvt = ev;
        } else if (t === "memory.stored") {
            if (!proofEvt) proofEvt = ev;
        } else if (t === "mission.completed") {
            completedEvt = ev;
            if (!proofEvt) proofEvt = ev;
        } else if (t === "mission.failed" || t === "mission.cancelled") {
            failedEvt = ev;
        }
    }

    const apply = (idx: number, ev: MissionEvent | undefined, status: RunStageStatus, summary: string) => {
        stages[idx].status = status;
        stages[idx].summary = summary;
        if (ev) {
            stages[idx].timestamp = ev.emitted_at;
            stages[idx].actor = ev.source_agent || ev.source_team;
        }
    };

    apply(0, intentEvt, "completed", "Mission intent accepted.");
    apply(1, proposalEvt ?? intentEvt, "completed", "Proposal and execution scope prepared.");

    if (approvalDeny) {
        apply(2, approvalReq ?? approvalDeny, "completed", "Approval gate evaluated.");
        apply(3, approvalDeny, "failed", "Approval denied.");
    } else if (approvalReq && !approvalGrant && !completedEvt && !toolInvoked && !artifactEvt) {
        apply(2, approvalReq, "active", "Awaiting approval.");
    } else if (approvalGrant || toolInvoked || toolDone || artifactEvt || completedEvt || runningEvt) {
        apply(2, approvalReq ?? approvalGrant ?? runningEvt ?? intentEvt, "completed", "Governance gate cleared.");
        apply(3, approvalGrant ?? runningEvt ?? intentEvt, "completed", "Execution authorized.");
    }

    if (!approvalDeny) {
        if (failedEvt && !toolFailed) apply(4, failedEvt, "failed", "Mission run failed.");
        else if (completedEvt || artifactEvt || toolInvoked || toolDone || toolFailed) apply(4, runningEvt ?? intentEvt, "completed", "Mission runtime active.");
        else if (runningEvt && stages[2].status !== "active") apply(4, runningEvt, "active", "Mission running.");
    }

    if (toolFailed) apply(5, toolFailed, "failed", "Tool action failed.");
    else if (completedEvt && (toolDone || toolInvoked || artifactEvt)) apply(5, toolDone ?? toolInvoked ?? artifactEvt ?? completedEvt, "completed", "Tool actions completed.");
    else if (toolInvoked && (invokedCount > doneCount || (!completedEvt && !artifactEvt))) apply(5, toolDone ?? toolInvoked, "active", "Executing tool action.");
    else if (toolDone || artifactEvt) apply(5, toolDone ?? artifactEvt, "completed", "Tool execution completed.");

    if (artifactEvt || completedEvt) apply(6, artifactEvt ?? completedEvt, "completed", "Delivery package generated.");
    if (completedEvt || proofEvt) apply(7, proofEvt ?? completedEvt, "completed", "Execution proof verified.");
    else if (failedEvt || toolFailed || approvalDeny) apply(7, failedEvt ?? toolFailed ?? approvalDeny, "failed", "Proof verification halted.");

    if (approvalDeny || toolFailed || failedEvt) apply(8, approvalDeny ?? toolFailed ?? failedEvt, "failed", "Recovery required: inspect failed step.");
    else if (completedEvt) apply(8, completedEvt, "completed", "Next actions ready.");

    return stages;
}

function stageBadgeClass(status: RunStageStatus): string {
    if (status === "completed") return "bg-cortex-success/15 text-cortex-success border-cortex-success/30";
    if (status === "active") return "bg-cortex-primary/15 text-cortex-primary border-cortex-primary/40";
    if (status === "failed") return "bg-cortex-danger/15 text-cortex-danger border-cortex-danger/30";
    return "bg-cortex-surface text-cortex-text-muted/60 border-cortex-border";
}

export default function RunStageProgressStrip({ events }: { events: MissionEvent[] }) {
    const stages = buildRunLifecycleStages(events);
    return (
        <div data-testid="run-stage-progress-strip" className="border-b border-cortex-border bg-cortex-surface/40 px-4 py-2.5">
            <div className="max-w-2xl mx-auto flex flex-wrap items-center gap-1.5">
                {stages.map((stage, idx) => (
                    <span
                        key={stage.id}
                        data-testid={`run-stage-${stage.id}`}
                        data-stage-status={stage.status}
                        title={stage.summary || stage.label}
                        className={`inline-flex items-center gap-1 rounded border px-2 py-0.5 text-[10px] font-mono ${stageBadgeClass(stage.status)}`}
                    >
                        <span className="opacity-60">{idx + 1}.</span>
                        <span>{stage.label}</span>
                    </span>
                ))}
            </div>
        </div>
    );
}

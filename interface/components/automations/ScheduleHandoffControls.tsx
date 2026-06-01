import type React from "react";
import {
    formatScheduleHandoffState,
    scheduleHandoffTone,
} from "@/components/automations/scheduleHandoffState";

export function ScheduleHandoffActions({
    disabled,
    onResolve,
}: {
    disabled: boolean;
    onResolve: (status: "approved" | "rejected" | "cancelled") => Promise<void>;
}) {
    return (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded border border-amber-400/30 bg-amber-400/5 px-3 py-2">
            <p className="text-xs leading-5 text-cortex-text-main">
                This cadence handoff is waiting for an operator decision. Approval records trust intent only; it does not execute the mission.
            </p>
            <div className="flex flex-wrap gap-2">
                <ScheduleHandoffButton tone="approve" disabled={disabled} onClick={() => onResolve("approved")}>
                    Approve handoff
                </ScheduleHandoffButton>
                <ScheduleHandoffButton tone="reject" disabled={disabled} onClick={() => onResolve("rejected")}>
                    Reject
                </ScheduleHandoffButton>
                <ScheduleHandoffButton tone="cancel" disabled={disabled} onClick={() => onResolve("cancelled")}>
                    Cancel
                </ScheduleHandoffButton>
            </div>
        </div>
    );
}

export function ScheduleHandoffBadge({ ruleID, state }: { ruleID: string; state: string }) {
    const tone = scheduleHandoffTone(state);
    const className =
        tone === "success"
            ? "bg-cortex-success/10 text-cortex-success"
            : tone === "danger"
                ? "bg-red-400/10 text-red-300"
                : tone === "pending"
                    ? "bg-amber-400/10 text-amber-300"
                    : "bg-cortex-bg/70 text-cortex-text-main";

    return (
        <span
            data-testid={`schedule-rule-badge-handoff-${ruleID}`}
            className={`flex-shrink-0 px-1.5 py-0.5 text-[10px] rounded font-mono ${className}`}
        >
            handoff {formatScheduleHandoffState(state)}
        </span>
    );
}

function ScheduleHandoffButton({
    children,
    disabled,
    onClick,
    tone,
}: {
    children: React.ReactNode;
    disabled: boolean;
    onClick: () => void;
    tone: "approve" | "reject" | "cancel";
}) {
    const className =
        tone === "approve"
            ? "bg-cortex-success/10 text-cortex-success"
            : tone === "reject"
                ? "bg-red-400/10 text-red-300"
                : "border border-cortex-border text-cortex-text-muted";
    return (
        <button
            type="button"
            disabled={disabled}
            onClick={onClick}
            className={`rounded px-2.5 py-1.5 text-xs font-medium disabled:cursor-wait disabled:opacity-50 ${className}`}
        >
            {children}
        </button>
    );
}

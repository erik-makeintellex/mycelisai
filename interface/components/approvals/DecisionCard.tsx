"use client";

import { Check, X, AlertTriangle, Clock } from "lucide-react";
import type { PendingApproval } from "@/store/useCortexStore";

interface DecisionCardProps {
    approval: PendingApproval;
    onResolve: (id: string, approved: boolean) => void;
}

export function DecisionCard({ approval, onResolve }: DecisionCardProps) {
    const timeLeft = getTimeRemaining(approval.expires_at);
    const isUrgent = timeLeft.minutes < 5;

    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-xl shadow-lg overflow-hidden hover:border-cortex-primary/40 transition-colors">
            {/* Header */}
            <div className="px-4 py-3 border-b border-cortex-border flex items-start justify-between">
                <div className="flex items-center gap-3 min-w-0">
                    <div className="p-1.5 rounded-md border border-cortex-border bg-cortex-bg text-cortex-warning flex-shrink-0">
                        <AlertTriangle size={16} />
                    </div>
                    <div className="min-w-0">
                        <h3 className="text-sm font-semibold text-cortex-text-main truncate">
                            {approval.reason}
                        </h3>
                        <p className="text-xs text-cortex-text-muted">
                            Agent{" "}
                            <span className="font-mono text-cortex-primary">
                                {approval.source_agent}
                            </span>{" "}
                            / Team{" "}
                            <span className="font-mono text-cortex-text-main">
                                {approval.team_id}
                            </span>
                        </p>
                    </div>
                </div>

                {/* Expiry Badge */}
                <span
                    className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full flex items-center gap-1 flex-shrink-0 ${
                        isUrgent
                            ? "bg-cortex-danger/20 text-cortex-danger"
                            : "bg-cortex-warning/20 text-cortex-warning"
                    }`}
                >
                    <Clock size={10} />
                    {timeLeft.label}
                </span>
            </div>

            {/* Content */}
            <div className="p-4">
                <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-2">
                        <span className="text-[10px] font-mono uppercase text-cortex-text-muted">
                            Intent
                        </span>
                        <span className="text-xs font-mono text-cortex-primary bg-cortex-bg px-2 py-0.5 rounded border border-cortex-border">
                            {approval.intent}
                        </span>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="text-[10px] font-mono uppercase text-cortex-text-muted">
                            Submitted
                        </span>
                        <span className="text-xs font-mono text-cortex-text-main">
                            {formatTimestamp(approval.timestamp)}
                        </span>
                    </div>
                </div>
            </div>

            {/* Actions */}
            <div className="px-4 py-3 bg-cortex-bg/50 border-t border-cortex-border flex justify-end gap-3">
                <button
                    onClick={() => onResolve(approval.id, false)}
                    className="px-3 py-1.5 text-xs font-medium text-cortex-danger hover:bg-cortex-danger/10 border border-cortex-danger/30 rounded-lg flex items-center gap-1.5 transition-colors"
                >
                    <X size={14} />
                    Reject
                </button>
                <button
                    onClick={() => onResolve(approval.id, true)}
                    className="px-3 py-1.5 text-xs font-medium text-cortex-bg bg-cortex-success hover:bg-cortex-success/90 rounded-lg flex items-center gap-1.5 shadow-sm transition-colors"
                >
                    <Check size={14} />
                    Approve
                </button>
            </div>
        </div>
    );
}

function formatTimestamp(ts: string): string {
    try {
        return new Date(ts).toLocaleString();
    } catch {
        return ts;
    }
}

function getTimeRemaining(expiresAt: string): { minutes: number; label: string } {
    try {
        const diff = new Date(expiresAt).getTime() - Date.now();
        if (diff <= 0) return { minutes: 0, label: "EXPIRED" };
        const mins = Math.floor(diff / 60000);
        if (mins < 60) return { minutes: mins, label: `${mins}m left` };
        const hours = Math.floor(mins / 60);
        return { minutes: mins, label: `${hours}h ${mins % 60}m left` };
    } catch {
        return { minutes: 999, label: "unknown" };
    }
}

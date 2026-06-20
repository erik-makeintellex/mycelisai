"use client";

import { AlertTriangle, Zap } from "lucide-react";
import { responseStateToneClass } from "@/components/soma/SomaCausalSummaryState";
import { toolLabel } from "@/lib/labels";
import type { ChatMessage } from "@/store/useCortexStore";

type ProposalBusScope = NonNullable<ChatMessage["proposal"]>["bus_scope"];

function busScopeLabel(scope?: ProposalBusScope) {
    if (scope === "current_team") return "Current team route";
    if (scope === "multi_team") return "Multiple teams route";
    if (scope === "global") return "Organization-wide route";
    if (scope === "none") return "No team route needed";
    return null;
}

export default function MissionControlThreadStateCard({ msg }: { msg: ChatMessage }) {
    const state = msg.ui_response_state ?? msg.execution_summary?.ui_response_state;
    const route = busScopeLabel(msg.proposal?.bus_scope);
    const proposalTools = msg.proposal?.tools?.filter(Boolean) ?? [];
    const hasProposalMeta = Boolean(msg.proposal && (route || proposalTools.length || msg.proposal.nats_subjects?.length));

    if (!state && !hasProposalMeta) return null;

    const label = state?.label ?? state?.kind?.replace(/_/g, " ") ?? "Structured response";
    const detail = state?.detail ?? (
        msg.proposal
            ? "Soma prepared governed work for review before anything runs."
            : "Soma returned structured outcome state for this reply."
    );

    return (
        <div className={`rounded-lg border px-3 py-2 ${responseStateToneClass(state?.tone)}`} data-testid="soma-thread-state-card">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                    {state?.tone === "danger" || state?.tone === "warning" ? (
                        <AlertTriangle className="h-3.5 w-3.5" />
                    ) : (
                        <Zap className="h-3.5 w-3.5" />
                    )}
                    <span className="text-[10px] font-mono font-bold uppercase tracking-[0.16em]">
                        {label}
                    </span>
                </div>
                {route ? (
                    <span className="rounded border border-current/20 px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase">
                        {route}
                    </span>
                ) : null}
            </div>
            {detail ? <p className="mt-1 text-xs leading-5 text-cortex-text-main">{detail}</p> : null}
            {proposalTools.length ? (
                <p className="mt-1 text-[11px] leading-5 text-cortex-text-muted">
                    Uses {proposalTools.map(toolLabel).join(", ")} after approval.
                </p>
            ) : null}
        </div>
    );
}

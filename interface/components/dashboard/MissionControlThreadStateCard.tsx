"use client";

import { AlertTriangle, ExternalLink, Zap } from "lucide-react";
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
    const threadEvents = (msg.thread_events ?? (msg.thread_event ? [msg.thread_event] : [])).slice(-3);
    const route = busScopeLabel(msg.proposal?.bus_scope);
    const proposalTools = msg.proposal?.tools?.filter(Boolean) ?? [];
    const hasProposalMeta = Boolean(msg.proposal && (route || proposalTools.length || msg.proposal.nats_subjects?.length));
    const hasStateBlock = Boolean(state || hasProposalMeta);

    if (!state && !hasProposalMeta && threadEvents.length === 0) return null;

    const label = state?.label ?? state?.kind?.replace(/_/g, " ") ?? "Structured response";
    const detail = state?.detail ?? (
        msg.proposal
            ? "Soma prepared governed work for review before anything runs."
            : "Soma returned structured outcome state for this reply."
    );

    return (
        <div className={`rounded-md border px-2.5 py-1.5 text-[11px] ${responseStateToneClass(state?.tone)}`} data-testid="soma-thread-state-card">
            {hasStateBlock ? (
                <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                        {state?.tone === "danger" || state?.tone === "warning" ? (
                            <AlertTriangle className="h-3 w-3" />
                        ) : (
                            <Zap className="h-3 w-3" />
                        )}
                        <span className="text-[9px] font-mono font-bold uppercase tracking-[0.14em]">
                            {label}
                        </span>
                    </div>
                    {route ? (
                        <span className="rounded border border-current/20 px-1.5 py-0.5 text-[8px] font-mono font-bold uppercase">
                            {route}
                        </span>
                    ) : null}
                </div>
            ) : null}
            {hasStateBlock && detail ? <p className="mt-0.5 line-clamp-2 text-[11px] leading-4 text-cortex-text-main">{detail}</p> : null}
            {proposalTools.length ? (
                <p className="mt-0.5 line-clamp-1 text-[10px] leading-4 text-cortex-text-muted">
                    Uses {proposalTools.map(toolLabel).join(", ")} after approval.
                </p>
            ) : null}
            {threadEvents.length ? (
                <div className="mt-1.5 space-y-1">
                    {threadEvents.map((event, index) => (
                        <div key={event.id ?? `${event.kind}-${index}`} className="rounded-md border border-current/15 bg-cortex-bg/35 px-2 py-1">
                            <div className="flex flex-wrap items-center justify-between gap-2">
                                <span className="text-[10px] font-semibold text-cortex-text-main">
                                    {event.label || event.title}
                                </span>
                                {event.status ? (
                                    <span className="rounded border border-current/20 px-1.5 py-0.5 text-[8px] font-mono font-bold uppercase">
                                        {event.status}
                                    </span>
                                ) : null}
                            </div>
                            {event.detail ? <p className="mt-0.5 line-clamp-1 text-[10px] leading-4 text-cortex-text-muted">{event.detail}</p> : null}
                            {event.href ? (
                                <a href={event.href} className="mt-0.5 inline-flex items-center gap-1 text-[10px] font-semibold text-cortex-primary hover:underline">
                                    {event.href_label ?? "Open proof"}
                                    <ExternalLink className="h-2.5 w-2.5" />
                                </a>
                            ) : null}
                        </div>
                    ))}
                </div>
            ) : null}
        </div>
    );
}

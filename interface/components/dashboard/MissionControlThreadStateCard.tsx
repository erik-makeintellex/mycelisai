"use client";

import { AlertTriangle, ExternalLink, Zap } from "lucide-react";
import { responseStateToneClass } from "@/components/soma/SomaCausalSummaryState";
import { requestSomaOutputContinuation } from "@/components/soma/outputContinuation";
import { outputCanvasHref } from "@/lib/outputPackageModel";
import { packageRecoveryPresentation } from "@/lib/teamWorkRecoveryPresentation";
import { progressForChatMessage, progressForThreadEvent } from "@/lib/teamWorkProgress";
import type { ChatMessage } from "@/store/useCortexStore";

function currentSomaHref() {
    if (typeof window === "undefined") return "/dashboard";
    return `${window.location.pathname}${window.location.search}${window.location.hash}`;
}

export default function MissionControlThreadStateCard({ msg }: { msg: ChatMessage }) {
    if (msg.proposal && !(msg.thread_event || msg.thread_events?.length)) return null;

    const state = msg.ui_response_state ?? msg.execution_summary?.ui_response_state;
    const progress = progressForChatMessage(msg);
    const threadEvents = (msg.thread_events ?? (msg.thread_event ? [msg.thread_event] : [])).slice(-1);
    const hasProposalMeta = false;
    const hasStateBlock = Boolean(state || hasProposalMeta);

    if (!state && !hasProposalMeta && threadEvents.length === 0) return null;

    const label = state?.label ?? state?.kind?.replace(/_/g, " ") ?? "Structured response";
    const detail = state?.detail ?? (
        msg.proposal
            ? "Soma prepared governed work for review before anything runs."
            : "Soma returned structured outcome state for this reply."
    );

    return (
        <div className={`rounded-lg border px-3 py-2.5 text-sm ${responseStateToneClass(state?.tone)}`} data-testid="soma-thread-state-card">
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
                </div>
            ) : null}
            {hasStateBlock && detail ? <p className="mt-0.5 line-clamp-2 text-[11px] leading-4 text-cortex-text-main">{detail}</p> : null}
            {threadEvents.length ? (
                <div className={hasStateBlock ? "mt-2 border-t border-current/10 pt-2" : ""}>
                    {threadEvents.map((event, index) => {
                        const eventProgress = progressForThreadEvent(event);
                        const needsDirection = event.kind === "attention_required";
                        const packageRecovery = needsDirection ? packageRecoveryPresentation(event) : null;
                        const eventLabel = packageRecovery
                            ? "Output is not playable yet"
                            : needsDirection ? "Soma needs your direction" : event.label || event.title;
                        const eventDetail = needsDirection
                            ? packageRecovery
                                ? packageRecovery.detail
                                : "This work stopped before a usable result was produced. Nothing new should be trusted yet."
                            : event.detail;
                        const canvasHref = event.kind === "result_ready" && event.href
                            ? outputCanvasHref({
                                label: event.href_label ?? eventLabel ?? "Retained output",
                                url: event.href,
                                storagePath: event.target_reference,
                                returnTo: "/dashboard",
                                teamId: event.team_id,
                                runId: event.run_id ?? msg.run_id,
                                workItemId: event.work_item_id,
                            })
                            : null;
                        return (
                        <div key={event.id ?? `${event.kind}-${index}`}>
                            <p className="mb-1 font-mono text-[9px] font-bold uppercase tracking-[0.14em] text-cortex-text-muted">
                                {eventProgress.label}
                            </p>
                            <div className="flex flex-wrap items-center justify-between gap-2">
                                <span className="text-sm font-semibold text-cortex-text-main">
                                    {eventLabel}
                                </span>
                            </div>
                            {eventDetail ? <p className="mt-1 text-sm leading-5 text-cortex-text-muted">{eventDetail}</p> : null}
                            {packageRecovery ? (
                                <div className="mt-2">
                                    <p className="text-sm leading-5 text-cortex-text-main">{packageRecovery.trust}</p>
                                    <button
                                        type="button"
                                        onClick={() => requestSomaOutputContinuation({
                                            title: event.label || event.title || "Retained package",
                                            reference: event.target_reference,
                                            proof: event.href,
                                            sourceLabel: "unverified retained candidate",
                                            teamId: event.team_id,
                                            runId: event.run_id ?? msg.run_id,
                                            workItemId: event.work_item_id,
                                        })}
                                        className="mt-2 inline-flex rounded-md border border-cortex-warning/30 bg-cortex-warning/10 px-2.5 py-1.5 text-xs font-semibold text-cortex-text-main hover:bg-cortex-warning/15"
                                    >
                                        {packageRecovery.actionLabel}
                                    </button>
                                </div>
                            ) : needsDirection ? (
                                <p className="mt-2 text-sm leading-5 text-cortex-text-main">
                                    Tell Soma to try again, use another available service, or change the request.
                                </p>
                            ) : null}
                            {event.href && event.kind === "result_ready" ? (
                                <a
                                    href={canvasHref ?? event.href}
                                    onClick={canvasHref ? (clickEvent) => {
                                        clickEvent.preventDefault();
                                        const destination = outputCanvasHref({
                                            label: event.href_label ?? eventLabel ?? "Retained output",
                                            url: event.href,
                                            storagePath: event.target_reference,
                                            returnTo: currentSomaHref(),
                                            teamId: event.team_id,
                                            runId: event.run_id ?? msg.run_id,
                                            workItemId: event.work_item_id,
                                        });
                                        if (destination) window.location.assign(destination);
                                    } : undefined}
                                    className="mt-2 inline-flex items-center gap-1 text-sm font-semibold text-cortex-primary hover:underline"
                                >
                                    {event.href_label ?? "Open proof"}
                                    <ExternalLink className="h-2.5 w-2.5" />
                                </a>
                            ) : null}
                            {needsDirection && (event.detail || event.href) ? (
                                <details className="mt-2 border-t border-current/10 pt-2 text-xs text-cortex-text-muted">
                                    <summary className="cursor-pointer font-semibold">What happened</summary>
                                    {event.detail ? <p className="mt-1 leading-4">{event.detail}</p> : null}
                                    {event.href ? (
                                        <a href={event.href} className="mt-1 inline-flex items-center gap-1 text-cortex-primary hover:underline">
                                            View technical details
                                            <ExternalLink className="h-2.5 w-2.5" />
                                        </a>
                                    ) : null}
                                </details>
                            ) : null}
                        </div>
                        );
                    })}
                </div>
            ) : progress ? (
                <p className="text-sm font-semibold text-cortex-text-main">{progress.label}</p>
            ) : null}
        </div>
    );
}

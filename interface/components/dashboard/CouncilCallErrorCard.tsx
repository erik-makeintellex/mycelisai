"use client";

import React, { useState } from "react";
import { AlertTriangle, ChevronDown, ChevronUp, Copy, RotateCcw } from "lucide-react";
import type { MissionChatFailure } from "@/lib/missionChatFailure";

export default function CouncilCallErrorCard({
    failure,
    onRetry,
    onSwitchToSoma,
    onContinueWithSoma,
}: {
    failure: MissionChatFailure;
    onRetry: () => void;
    onSwitchToSoma: () => void;
    onContinueWithSoma: () => void;
}) {
    const isSoma = failure.routeKind === "workspace";
    const showSetupAction = failure.type === "setup_required" && Boolean(failure.setupPath);
    const [detailsOpen, setDetailsOpen] = useState(false);

    return (
        <div className="m-3 max-w-[min(100%,720px)] rounded-2xl border border-cortex-warning/35 bg-cortex-warning/10 p-4 shadow-sm">
            <div className="flex items-start gap-2">
                <AlertTriangle className="mt-1 h-4 w-4 shrink-0 text-cortex-warning" />
                <div className="min-w-0">
                    <p className="text-[11px] font-bold uppercase tracking-[0.14em] text-cortex-warning">Operational alert</p>
                    <h3 className="mt-1 text-base font-semibold text-cortex-text-main">{failure.bannerLabel}</h3>
                    <p className="mt-1 text-sm leading-6 text-cortex-text-muted">{failure.summary}</p>
                </div>
            </div>

            <div className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
                <div className="rounded-xl border border-cortex-border bg-cortex-bg/60 px-3 py-2">
                    <div className="text-[10px] font-bold uppercase tracking-[0.12em] text-cortex-text-muted">Still safe</div>
                    <p className="mt-1 text-cortex-text-main">No hidden work was started from this failed request.</p>
                </div>
                <div className="rounded-xl border border-cortex-border bg-cortex-bg/60 px-3 py-2">
                    <div className="text-[10px] font-bold uppercase tracking-[0.12em] text-cortex-text-muted">Safe next</div>
                    <p className="mt-1 text-cortex-text-main">{failure.recommendedAction}</p>
                </div>
            </div>

            <div className="mt-3 flex flex-wrap gap-2">
                <button onClick={onRetry} className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/15">
                    <RotateCcw className="w-3 h-3" />
                    Retry
                </button>
                {showSetupAction && (
                    <button
                        onClick={() => window.location.assign(failure.setupPath!)}
                        className="inline-flex h-8 items-center rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/15"
                    >
                        Open Settings
                    </button>
                )}
                {!isSoma && (
                    <button onClick={onSwitchToSoma} className="inline-flex h-8 items-center rounded-lg border border-cortex-primary/35 px-2.5 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10">
                        Switch to Soma
                    </button>
                )}
                {!isSoma && (
                    <button onClick={onContinueWithSoma} className="inline-flex h-8 items-center rounded-lg border border-cortex-border px-2.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/35">
                        Continue with Soma Only
                    </button>
                )}
                <button
                    type="button"
                    onClick={() => setDetailsOpen((open) => !open)}
                    className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-border px-2.5 text-xs font-semibold text-cortex-text-muted hover:border-cortex-primary/35"
                    aria-expanded={detailsOpen}
                >
                    {detailsOpen ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
                    Details
                </button>
            </div>
            {detailsOpen ? (
                <div className="mt-3 rounded-xl border border-cortex-border bg-cortex-bg/70 px-3 py-2 text-xs text-cortex-text-muted">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className="font-semibold text-cortex-text-main">Diagnostics</span>
                        <button
                            onClick={() => navigator.clipboard.writeText(failure.diagnostics)}
                            className="inline-flex items-center gap-1 text-cortex-primary hover:underline"
                        >
                            <Copy className="h-3 w-3" />
                            Copy
                        </button>
                    </div>
                    <div className="mt-2 grid gap-1 sm:grid-cols-2">
                        <span>Target: {failure.targetLabel}</span>
                        <span>Type: {failure.type}</span>
                    </div>
                </div>
            ) : null}
        </div>
    );
}

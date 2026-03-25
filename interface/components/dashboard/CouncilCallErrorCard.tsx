"use client";

import React from "react";
import { AlertTriangle, Copy, RotateCcw } from "lucide-react";
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

    return (
        <div className="m-3 rounded-xl border border-cortex-danger/30 bg-cortex-danger/10 p-3 space-y-3">
            <div className="flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 text-cortex-danger mt-0.5" />
                <div>
                    <p className="text-xs font-mono font-bold text-cortex-danger uppercase tracking-wide">{failure.title}</p>
                    <div className="text-[11px] font-mono text-cortex-text-main mt-1 space-y-1">
                        <div><span className="text-cortex-text-muted">Agent:</span> {failure.targetLabel}</div>
                        <div><span className="text-cortex-text-muted">Failure:</span> {failure.type}</div>
                    </div>
                </div>
            </div>

            <div className="space-y-1 text-[11px] font-mono text-cortex-text-muted">
                <div>{failure.summary}</div>
                <div className="text-cortex-text-main/80">{failure.recommendedAction}</div>
            </div>

            <div className="flex flex-wrap gap-2">
                <button onClick={onRetry} className="px-2 py-1 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/10 flex items-center gap-1">
                    <RotateCcw className="w-3 h-3" />
                    Retry
                </button>
                {showSetupAction && (
                    <button
                        onClick={() => window.location.assign(failure.setupPath!)}
                        className="px-2 py-1 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/10"
                    >
                        Open Settings
                    </button>
                )}
                {!isSoma && (
                    <button onClick={onSwitchToSoma} className="px-2 py-1 rounded border border-cortex-warning/30 text-cortex-warning text-[10px] font-mono hover:bg-cortex-warning/10">
                        Switch to Soma
                    </button>
                )}
                {!isSoma && (
                    <button onClick={onContinueWithSoma} className="px-2 py-1 rounded border border-cortex-border text-cortex-text-main text-[10px] font-mono hover:bg-cortex-border">
                        Continue with Soma Only
                    </button>
                )}
                <button
                    onClick={() => navigator.clipboard.writeText(failure.diagnostics)}
                    className="px-2 py-1 rounded border border-cortex-border text-cortex-text-muted text-[10px] font-mono hover:bg-cortex-border flex items-center gap-1"
                >
                    <Copy className="w-3 h-3" />
                    Copy Diagnostics
                </button>
            </div>
        </div>
    );
}

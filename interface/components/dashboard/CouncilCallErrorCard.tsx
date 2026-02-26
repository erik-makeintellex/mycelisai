"use client";

import React from "react";
import { AlertTriangle, Copy, RotateCcw } from "lucide-react";

type FailureType = "timeout" | "unreachable" | "server_error" | "unknown";

function classifyFailure(message: string): FailureType {
    const lower = message.toLowerCase();
    if (lower.includes("timeout")) return "timeout";
    if (lower.includes("unreachable") || lower.includes("failed to fetch")) return "unreachable";
    if (lower.includes("500") || lower.includes("server")) return "server_error";
    return "unknown";
}

function reasonFor(type: FailureType): string {
    if (type === "timeout") return "The council member did not respond before the request deadline.";
    if (type === "unreachable") return "The council member service is currently unreachable from this client.";
    if (type === "server_error") return "The council member service returned an internal error.";
    return "The request failed unexpectedly. Check system status for runtime health.";
}

export default function CouncilCallErrorCard({
    member,
    errorMessage,
    onRetry,
    onSwitchToSoma,
    onContinueWithSoma,
}: {
    member: string;
    errorMessage: string;
    onRetry: () => void;
    onSwitchToSoma: () => void;
    onContinueWithSoma: () => void;
}) {
    const failure = classifyFailure(errorMessage);

    return (
        <div className="m-3 rounded-xl border border-cortex-danger/30 bg-cortex-danger/10 p-3 space-y-3">
            <div className="flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 text-cortex-danger mt-0.5" />
                <div>
                    <p className="text-xs font-mono font-bold text-cortex-danger uppercase tracking-wide">Council Call Failed</p>
                    <div className="text-[11px] font-mono text-cortex-text-main mt-1 space-y-1">
                        <div><span className="text-cortex-text-muted">Agent:</span> {member}</div>
                        <div><span className="text-cortex-text-muted">Failure:</span> {failure}</div>
                    </div>
                </div>
            </div>

            <div className="text-[11px] font-mono text-cortex-text-muted">{reasonFor(failure)}</div>

            <div className="flex flex-wrap gap-2">
                <button onClick={onRetry} className="px-2 py-1 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/10 flex items-center gap-1">
                    <RotateCcw className="w-3 h-3" />
                    Retry
                </button>
                <button onClick={onSwitchToSoma} className="px-2 py-1 rounded border border-cortex-warning/30 text-cortex-warning text-[10px] font-mono hover:bg-cortex-warning/10">
                    Switch to Soma
                </button>
                <button onClick={onContinueWithSoma} className="px-2 py-1 rounded border border-cortex-border text-cortex-text-main text-[10px] font-mono hover:bg-cortex-border">
                    Continue with Soma Only
                </button>
                <button
                    onClick={() => navigator.clipboard.writeText(errorMessage)}
                    className="px-2 py-1 rounded border border-cortex-border text-cortex-text-muted text-[10px] font-mono hover:bg-cortex-border flex items-center gap-1"
                >
                    <Copy className="w-3 h-3" />
                    Copy Diagnostics
                </button>
            </div>
        </div>
    );
}


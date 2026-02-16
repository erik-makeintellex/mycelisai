"use client";

import React, { useRef, useEffect, useState } from "react";
import { Flame } from "lucide-react";
import { useCortexStore, type StreamSignal } from "@/store/useCortexStore";

// ── Direction Classification ─────────────────────────────────

type SignalDirection = "input" | "output" | "internal";

function classifyDirection(signal: StreamSignal): SignalDirection {
    const topic = signal.topic ?? "";
    const type = signal.type ?? "";

    if (topic.includes(".input.") || type === "sensor" || type === "sensor_data") {
        return "input";
    }
    if (type === "artifact" || type === "output") {
        return "output";
    }
    return "internal";
}

const DIRECTION_BORDER_COLOR: Record<SignalDirection, string> = {
    input: "border-l-cortex-info",
    output: "border-l-cortex-success",
    internal: "border-l-cortex-text-muted",
};

// ── Type Dot Color ───────────────────────────────────────────

function typeDotColor(type?: string): string {
    switch (type) {
        case "cognitive":
        case "thought":
        case "intent":
            return "bg-cortex-primary";
        case "artifact":
        case "output":
        case "task_complete":
            return "bg-cortex-success";
        case "error":
            return "bg-cortex-danger";
        case "governance_halt":
        case "governance":
            return "bg-cortex-warning";
        default:
            return "bg-cortex-text-muted";
    }
}

// ── Time Formatting ──────────────────────────────────────────

function formatHHMMSS(timestamp?: string): string {
    if (!timestamp) return new Date().toLocaleTimeString("en-GB", { hour12: false });
    try {
        const d = new Date(timestamp);
        if (isNaN(d.getTime())) return "now";
        return d.toLocaleTimeString("en-GB", { hour12: false });
    } catch {
        return "now";
    }
}

// ── Signal Row ───────────────────────────────────────────────

function SignalRow({ signal }: { signal: StreamSignal }) {
    const direction = classifyDirection(signal);
    const time = formatHHMMSS(signal.timestamp);
    const source = signal.source ?? "system";
    const message = (signal.message ?? JSON.stringify(signal.payload ?? {})).slice(0, 80);
    const dot = typeDotColor(signal.type);

    return (
        <div
            className={`flex items-center gap-1.5 py-1 px-2 border-b border-cortex-border/30 border-l-2 ${DIRECTION_BORDER_COLOR[direction]} hover:bg-cortex-surface/40 transition-colors`}
        >
            {/* Timestamp */}
            <span className="text-[10px] font-mono text-cortex-text-muted flex-shrink-0 w-[52px]">
                {time}
            </span>

            {/* Source */}
            <span className="text-[10px] font-bold text-cortex-text-main truncate max-w-[80px] flex-shrink-0">
                {source}
            </span>

            {/* Type dot */}
            <span className={`w-1.5 h-1.5 rounded-full inline-block flex-shrink-0 ${dot}`} />

            {/* Message preview */}
            <span className="text-[10px] text-cortex-text-main truncate flex-1 font-mono" title={message}>
                {message}
            </span>
        </div>
    );
}

// ── HotMemoryPanel ───────────────────────────────────────────

export default function HotMemoryPanel() {
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const scrollRef = useRef<HTMLDivElement>(null);
    const [paused, setPaused] = useState(false);

    // Auto-scroll to top on new logs (unless user is hovering)
    useEffect(() => {
        if (!paused && scrollRef.current) {
            scrollRef.current.scrollTop = 0;
        }
    }, [streamLogs.length, paused]);

    return (
        <div className="h-full flex flex-col">
            {/* Sub-header */}
            <div className="h-10 flex items-center justify-between px-3 border-b border-cortex-border bg-cortex-surface/50 flex-shrink-0">
                <div className="flex items-center gap-2">
                    <Flame className="w-3.5 h-3.5 text-cortex-danger" />
                    <span className="text-[10px] font-bold uppercase tracking-widest text-cortex-text-muted">
                        Hot
                    </span>
                </div>
                <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-danger/15 text-cortex-danger">
                    {streamLogs.length} live
                </span>
            </div>

            {/* Scrollable event list */}
            <div
                ref={scrollRef}
                className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border min-h-0"
                onMouseEnter={() => setPaused(true)}
                onMouseLeave={() => setPaused(false)}
            >
                {streamLogs.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full gap-2">
                        <Flame className="w-8 h-8 text-cortex-text-muted opacity-20" />
                        <span className="text-[10px] font-mono text-cortex-text-muted">
                            Waiting for signals...
                        </span>
                    </div>
                ) : (
                    streamLogs.map((signal, index) => (
                        <SignalRow key={`${signal.timestamp}-${index}`} signal={signal} />
                    ))
                )}
            </div>

            {/* Footer */}
            <div className="h-7 flex items-center justify-center border-t border-cortex-border/50 bg-cortex-surface/30 flex-shrink-0">
                <span className="text-[9px] font-mono text-cortex-text-muted">
                    {streamLogs.length} buffered
                </span>
            </div>
        </div>
    );
}

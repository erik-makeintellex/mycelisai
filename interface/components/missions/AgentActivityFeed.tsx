"use client";

import React, { useMemo, useState, useRef, useEffect } from "react";
import { Activity, Brain, Package, AlertTriangle, Radio } from "lucide-react";
import { useCortexStore, type StreamSignal } from "@/store/useCortexStore";

type FilterType = "all" | "cognitive" | "artifact" | "error";

interface AgentActivityFeedProps {
    missionId: string;
}

const FILTER_CHIPS: { key: FilterType; label: string }[] = [
    { key: "all", label: "All" },
    { key: "cognitive", label: "Cognitive" },
    { key: "artifact", label: "Artifact" },
    { key: "error", label: "Error" },
];

function signalCategory(signal: StreamSignal): FilterType {
    const t = signal.type?.toLowerCase() ?? "";
    if (t === "error" || signal.level === "error") return "error";
    if (t === "artifact" || t === "output") return "artifact";
    if (t === "thought" || t === "cognitive") return "cognitive";
    return "all";
}

function typeColor(category: FilterType): string {
    switch (category) {
        case "cognitive":
            return "bg-cortex-primary";
        case "artifact":
            return "bg-cortex-success";
        case "error":
            return "bg-cortex-danger";
        default:
            return "bg-cortex-text-muted/50";
    }
}

function formatTime(ts?: string): string {
    if (!ts) return "--:--:--";
    try {
        return new Date(ts).toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
            hour12: false,
        });
    } catch {
        return "--:--:--";
    }
}

export default function AgentActivityFeed({ missionId }: AgentActivityFeedProps) {
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const [filter, setFilter] = useState<FilterType>("all");
    const scrollRef = useRef<HTMLDivElement>(null);
    const prevCountRef = useRef(0);

    const filteredLogs = useMemo(() => {
        const logs = [...streamLogs]; // Already most-recent-first from store
        if (filter === "all") return logs;
        return logs.filter((log) => signalCategory(log) === filter);
    }, [streamLogs, filter]);

    // Auto-scroll to top when new items arrive
    useEffect(() => {
        if (streamLogs.length > prevCountRef.current && scrollRef.current) {
            scrollRef.current.scrollTop = 0;
        }
        prevCountRef.current = streamLogs.length;
    }, [streamLogs.length]);

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            {/* Header */}
            <div className="p-3 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between flex-shrink-0">
                <h3 className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-widest flex items-center gap-2">
                    <Activity className="w-3.5 h-3.5" />
                    Activity Feed
                </h3>
                <span className="text-[10px] font-mono text-cortex-text-muted/60">
                    {filteredLogs.length} signals
                </span>
            </div>

            {/* Filter chips */}
            <div className="px-4 py-2 border-b border-cortex-border/50 flex items-center gap-1.5 flex-shrink-0">
                {FILTER_CHIPS.map((chip) => (
                    <button
                        key={chip.key}
                        onClick={() => setFilter(chip.key)}
                        className={`px-2.5 py-1 rounded text-[10px] font-mono transition-all ${
                            filter === chip.key
                                ? "bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30"
                                : "text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-border/50 border border-transparent"
                        }`}
                    >
                        {chip.label}
                    </button>
                ))}
            </div>

            {/* Feed */}
            <div
                ref={scrollRef}
                className="flex-1 overflow-y-auto p-4 space-y-1"
            >
                {filteredLogs.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full">
                        <Radio className="w-10 h-10 opacity-20 text-cortex-text-muted mb-3" />
                        <p className="font-mono text-xs text-cortex-text-muted/60">
                            Waiting for agent signals...
                        </p>
                    </div>
                ) : (
                    filteredLogs.map((log, i) => {
                        const category = signalCategory(log);
                        const message = log.message ?? "";
                        const truncated =
                            message.length > 120
                                ? message.slice(0, 120) + "..."
                                : message;

                        return (
                            <div
                                key={`${log.timestamp}-${i}`}
                                className="flex items-start gap-2.5 px-2 py-1.5 rounded hover:bg-cortex-surface/50 transition-colors group"
                            >
                                {/* Timestamp */}
                                <span className="text-[10px] font-mono text-cortex-text-muted/50 flex-shrink-0 w-16 mt-0.5">
                                    {formatTime(log.timestamp)}
                                </span>

                                {/* Source */}
                                <span className="text-[11px] font-mono font-bold text-cortex-text-main flex-shrink-0 w-28 truncate mt-0.5" title={log.source}>
                                    {log.source ?? "system"}
                                </span>

                                {/* Type dot */}
                                <span
                                    className={`w-2 h-2 rounded-full flex-shrink-0 mt-1.5 ${typeColor(category)}`}
                                    title={category}
                                />

                                {/* Message */}
                                <span className="text-xs text-cortex-text-muted flex-1 leading-relaxed">
                                    {truncated || (
                                        <span className="italic text-cortex-text-muted/40">
                                            (no message)
                                        </span>
                                    )}
                                </span>
                            </div>
                        );
                    })
                )}
            </div>

            {/* Gradient fade */}
            <div className="pointer-events-none absolute inset-x-0 bottom-0 h-6 bg-gradient-to-t from-cortex-bg to-transparent" />
        </div>
    );
}

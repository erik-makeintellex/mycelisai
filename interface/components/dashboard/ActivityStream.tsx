"use client";

import React, { useEffect, useRef } from "react";
import { useSignalStream, Signal } from "./SignalContext";
import { Terminal, Brain, Cpu, MessageSquare, Shield } from "lucide-react";

export default function ActivityStream() {
    const { signals } = useSignalStream();
    const scrollRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [signals]);

    return (
        <div className="h-full flex flex-col bg-cortex-bg border-l border-cortex-border relative">
            <div className="p-3 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur top-0 z-10 flex justify-between items-center">
                <h3 className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-widest flex items-center gap-2">
                    <Terminal className="w-3 h-3" />
                    Activity Stream
                </h3>
                <span className="text-[10px] text-cortex-text-muted font-mono">
                    LIVE FEED
                </span>
            </div>

            <div
                className="flex-1 overflow-y-auto p-4 space-y-4"
                ref={scrollRef}
            >
                {signals.length === 0 && (
                    <div className="text-center text-cortex-text-muted/50 text-xs font-mono mt-10">
                        Awaiting Neural Signals...
                    </div>
                )}
                {signals.map((sig, i) => (
                    <ActivityItem key={i} signal={sig} />
                ))}
            </div>

            {/* Gradient fade at bottom */}
            <div className="pointer-events-none absolute inset-x-0 bottom-0 h-8 bg-gradient-to-t from-cortex-bg to-transparent z-0" />
        </div>
    );
}

function ActivityItem({ signal }: { signal: Signal }) {
    let content = signal.message || "";
    let type = signal.type || "info";

    if (content.trim().startsWith("{")) {
        try {
            const parsed = JSON.parse(content);
            if (parsed.type) type = parsed.type;
            if (parsed.content) content = parsed.content;
        } catch {
            // keep raw
        }
    }

    const timestamp = signal.timestamp
        ? new Date(signal.timestamp).toLocaleTimeString([], { hour12: false })
        : "";

    if (type === "thought" || signal.topic?.includes("thought")) {
        return (
            <div className="flex gap-3 text-xs opacity-80 hover:opacity-100 transition-opacity">
                <div className="mt-1 min-w-[16px]">
                    <Brain className="w-4 h-4 text-purple-400" />
                </div>
                <div className="flex-1 space-y-1">
                    <div className="text-purple-300/60 font-mono text-[10px]">
                        {timestamp} &bull; CORTEX
                    </div>
                    <div className="text-cortex-text-main/80 leading-relaxed font-serif italic border-l-2 border-purple-500/20 pl-2">
                        {content}
                    </div>
                </div>
            </div>
        );
    }

    if (type === "tool_call" || signal.topic?.includes("tool")) {
        return (
            <div className="flex gap-3 text-xs">
                <div className="mt-1 min-w-[16px]">
                    <Cpu className="w-4 h-4 text-cortex-success" />
                </div>
                <div className="flex-1 space-y-1">
                    <div className="text-cortex-success/60 font-mono text-[10px]">
                        {timestamp} &bull; TOOL EXECUTION
                    </div>
                    <div className="bg-cortex-surface border border-cortex-border rounded p-2 font-mono text-cortex-success">
                        {content}
                    </div>
                </div>
            </div>
        );
    }

    if (type === "user_input" || signal.topic?.includes("command")) {
        return (
            <div className="flex gap-3 text-xs">
                <div className="mt-1 min-w-[16px]">
                    <MessageSquare className="w-4 h-4 text-blue-400" />
                </div>
                <div className="flex-1 space-y-1">
                    <div className="text-blue-500/60 font-mono text-[10px]">
                        {timestamp} &bull; COMMAND
                    </div>
                    <div className="text-cortex-text-main font-medium">
                        {content}
                    </div>
                </div>
            </div>
        );
    }

    // Default
    return (
        <div className="flex gap-3 text-xs text-cortex-text-muted">
            <div className="mt-1 min-w-[16px]">
                <Shield className="w-4 h-4 text-cortex-text-muted/50" />
            </div>
            <div className="flex-1">
                <span className="font-mono text-[10px] mr-2 text-cortex-text-muted/50">
                    {timestamp}
                </span>
                <span className="text-cortex-text-muted">{content}</span>
                {signal.topic && (
                    <div className="text-[9px] text-cortex-text-muted/40 font-mono mt-0.5">
                        {signal.topic}
                    </div>
                )}
            </div>
        </div>
    );
}

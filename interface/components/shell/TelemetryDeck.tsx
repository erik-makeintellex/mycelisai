"use client";

import { Terminal, Activity, ChevronUp, ChevronDown } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";

export function TelemetryDeck() {
    const [isExpanded, setIsExpanded] = useState(true);

    return (
        <div className={cn(
            "border-t border-zinc-200 bg-zinc-50 transition-all duration-300 flex flex-col shadow-[0_-5px_15px_-5px_rgba(0,0,0,0.05)]",
            isExpanded ? "h-64" : "h-12"
        )}>
            {/* Header / Toggle */}
            <div
                className="h-12 px-4 flex items-center justify-between border-b border-zinc-200 cursor-pointer hover:bg-zinc-100"
                onClick={() => setIsExpanded(!isExpanded)}
            >
                <div className="flex items-center gap-2 text-zinc-600">
                    <Activity className="w-4 h-4 text-emerald-500 animate-pulse" />
                    <span className="text-xs font-mono font-medium tracking-wide">TELEMETRY STREAM</span>
                    <span className="text-xs text-zinc-400 ml-2">● LIVE</span>
                </div>

                <button className="text-zinc-400 hover:text-zinc-600">
                    {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronUp className="w-4 h-4" />}
                </button>
            </div>

            {/* Content (Terminal / Logs) */}
            {isExpanded && (
                <div className="flex-1 overflow-hidden flex">
                    {/* Sidebar for Deck */}
                    <div className="w-48 border-r border-zinc-200 bg-zinc-100 p-2 space-y-1">
                        <div className="p-2 bg-white border border-zinc-200 rounded text-xs text-zinc-700 font-mono cursor-pointer shadow-sm">
                            scip.global
                        </div>
                        <div className="p-2 hover:bg-zinc-200/50 rounded text-xs text-zinc-500 font-mono cursor-pointer">
                            sys.infra
                        </div>
                    </div>

                    {/* Log Output Area (Kept dark for terminal aesthetic, but framed in light) */}
                    <div className="flex-1 bg-zinc-950 p-4 font-mono text-xs text-zinc-400 overflow-y-auto">
                        <div className="pb-1"><span className="text-zinc-600">[10:00:01]</span> <span className="text-blue-400">INFO</span>  Core Initialized version v0.6.0</div>
                        <div className="pb-1"><span className="text-zinc-600">[10:00:02]</span> <span className="text-emerald-400">SCIP</span>  Listener attached to nats://localhost:4222</div>
                        <div className="pb-1"><span className="text-zinc-600">[10:00:05]</span> <span className="text-purple-400">BOOT</span>  Agent Registry synced (3 agents)</div>
                        <div className="mt-2 text-zinc-600 animate-pulse">_</div>
                    </div>
                </div>
            )}
        </div>
    );
}

"use client";

import React, { useState, useCallback } from "react";
import { Brain, ChevronDown, ChevronRight, Radio } from "lucide-react";
import HotMemoryPanel from "./HotMemoryPanel";
import WarmMemoryPanel from "./WarmMemoryPanel";
import ColdMemoryPanel from "./ColdMemoryPanel";
import { useCortexStore } from "@/store/useCortexStore";

// ── MemoryExplorer ────────────────────────────────────────────

export default function MemoryExplorer() {
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const [coldSearchQuery, setColdSearchQuery] = useState<string | undefined>(undefined);
    const [signalExpanded, setSignalExpanded] = useState(false);

    const handleSearchRelated = useCallback((query: string) => {
        setColdSearchQuery(query);
    }, []);

    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main">
            {/* Header */}
            <header className="h-12 border-b border-cortex-border flex items-center justify-between px-4 bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-2.5">
                    <Brain className="w-4 h-4 text-cortex-success" />
                    <div>
                        <span className="font-mono font-bold text-sm text-cortex-text-main">Memory</span>
                        <span className="hidden md:inline text-[10px] font-mono text-cortex-text-muted ml-3">
                            Retained knowledge and conversations from your swarm
                        </span>
                    </div>
                </div>
            </header>

            {/* Main two-column layout: Recent Work (40%) | Search (60%) */}
            <div
                className="flex-1 min-h-0"
                style={{
                    display: "grid",
                    gridTemplateColumns: "2fr 3fr",
                    gap: 0,
                    overflow: "hidden",
                }}
            >
                {/* Col 1: Recent Work (Warm — sitreps + artifacts) */}
                <div className="h-full overflow-hidden border-r border-cortex-border flex flex-col">
                    <div className="h-8 flex items-center px-3 border-b border-cortex-border/50 bg-cortex-surface/30 flex-shrink-0">
                        <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted font-bold">
                            Recent Work
                        </span>
                    </div>
                    <div className="flex-1 overflow-hidden">
                        <WarmMemoryPanel onSearchRelated={handleSearchRelated} />
                    </div>
                </div>

                {/* Col 2: Semantic Search (Cold — pgvector) */}
                <div className="h-full overflow-hidden flex flex-col">
                    <div className="h-8 flex items-center px-3 border-b border-cortex-border/50 bg-cortex-surface/30 flex-shrink-0">
                        <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted font-bold">
                            Search Memory
                        </span>
                    </div>
                    <div className="flex-1 overflow-hidden">
                        <ColdMemoryPanel searchQuery={coldSearchQuery} />
                    </div>
                </div>
            </div>

            {/* Advanced Mode only: Signal Stream (collapsible) */}
            {advancedMode && (
                <div className="flex-shrink-0 border-t border-cortex-border bg-cortex-surface/30">
                    <button
                        onClick={() => setSignalExpanded(!signalExpanded)}
                        className="w-full h-8 flex items-center gap-2 px-3 hover:bg-cortex-bg/40 transition-colors"
                    >
                        {signalExpanded ? (
                            <ChevronDown className="w-3.5 h-3.5 text-cortex-text-muted" />
                        ) : (
                            <ChevronRight className="w-3.5 h-3.5 text-cortex-text-muted" />
                        )}
                        <Radio className="w-3 h-3 text-cortex-text-muted" />
                        <span className="text-[9px] font-mono uppercase tracking-widest text-cortex-text-muted font-bold">
                            Signal Stream
                        </span>
                    </button>
                    {signalExpanded && (
                        <div style={{ height: 220 }} className="border-t border-cortex-border/50 overflow-hidden">
                            <HotMemoryPanel />
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

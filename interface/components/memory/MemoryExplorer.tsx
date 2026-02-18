"use client";

import React, { useState, useCallback } from "react";
import { Brain, Flame, Database, Snowflake } from "lucide-react";
import HotMemoryPanel from "./HotMemoryPanel";
import WarmMemoryPanel from "./WarmMemoryPanel";
import ColdMemoryPanel from "./ColdMemoryPanel";

// ── Tier Legend Chip ──────────────────────────────────────────

function TierChip({
    icon: Icon,
    label,
    colorClass,
}: {
    icon: React.ElementType;
    label: string;
    colorClass: string;
}) {
    return (
        <div className={`flex items-center gap-1.5 text-[9px] font-mono uppercase px-2 py-1 rounded ${colorClass}`}>
            <Icon className="w-3 h-3" />
            {label}
        </div>
    );
}

// ── MemoryExplorer ───────────────────────────────────────────

export default function MemoryExplorer() {
    const [coldSearchQuery, setColdSearchQuery] = useState<string | undefined>(undefined);

    const handleSearchRelated = useCallback((query: string) => {
        setColdSearchQuery(query);
    }, []);

    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main">
            {/* Header */}
            <header className="h-12 border-b border-cortex-border flex items-center justify-between px-4 bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <h1 className="font-mono font-bold text-lg text-cortex-success flex items-center gap-2">
                    <Brain className="w-4 h-4" />
                    THREE-TIER MEMORY
                </h1>
                <div className="flex items-center gap-2">
                    <TierChip
                        icon={Flame}
                        label="Hot"
                        colorClass="bg-cortex-danger/15 text-cortex-danger"
                    />
                    <TierChip
                        icon={Database}
                        label="Warm"
                        colorClass="bg-cortex-warning/15 text-cortex-warning"
                    />
                    <TierChip
                        icon={Snowflake}
                        label="Cold"
                        colorClass="bg-cortex-info/15 text-cortex-info"
                    />
                </div>
            </header>

            {/* Three-column grid */}
            <div
                className="flex-1 min-h-0"
                style={{
                    display: "grid",
                    gridTemplateColumns: "repeat(3, minmax(240px, 1fr))",
                    gap: 0,
                }}
            >
                {/* Col 1: Hot Memory */}
                <div className="h-full overflow-hidden border-r border-cortex-border">
                    <HotMemoryPanel />
                </div>

                {/* Col 2: Warm Memory */}
                <div className="h-full overflow-hidden border-r border-cortex-border">
                    <WarmMemoryPanel onSearchRelated={handleSearchRelated} />
                </div>

                {/* Col 3: Cold Memory */}
                <div className="h-full overflow-hidden">
                    <ColdMemoryPanel searchQuery={coldSearchQuery} />
                </div>
            </div>
        </div>
    );
}

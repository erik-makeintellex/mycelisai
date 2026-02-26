"use client";

import React from "react";
import { Maximize2, Minimize2 } from "lucide-react";

export default function FocusModeToggle({
    focused,
    onToggle,
}: {
    focused: boolean;
    onToggle: () => void;
}) {
    return (
        <button
            onClick={onToggle}
            className={`px-2 py-1 rounded border text-[10px] font-mono transition-colors flex items-center gap-1 ${
                focused
                    ? "border-cortex-primary/30 text-cortex-primary bg-cortex-primary/10"
                    : "border-cortex-border text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-border"
            }`}
            title="Toggle Focus Mode (F)"
        >
            {focused ? <Minimize2 className="w-3 h-3" /> : <Maximize2 className="w-3 h-3" />}
            {focused ? "Focus On" : "Focus Off"}
        </button>
    );
}


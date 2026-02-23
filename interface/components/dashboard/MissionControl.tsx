"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import { SignalProvider, useSignalStream } from "./SignalContext";
import { Layers, Settings, Plus } from "lucide-react";
import TelemetryRow from "./TelemetryRow";
import MissionControlChat from "./MissionControlChat";
import OpsOverview from "./OpsOverview";
import SignalDetailDrawer from "../stream/SignalDetailDrawer";
import ModeRibbon from "./ModeRibbon";
import LaunchCrewModal from "../workspace/LaunchCrewModal";

const STORAGE_KEY = "workspace-split";
const STORAGE_KEY_LEGACY = "mission-control-split";
const DEFAULT_RATIO = 0.55; // 55% chat, 45% ops
const MIN_RATIO = 0.25;
const MAX_RATIO = 0.80;

function clamp(v: number, min: number, max: number) {
    return Math.max(min, Math.min(max, v));
}

function loadRatio(): number {
    if (typeof window === "undefined") return DEFAULT_RATIO;
    // Try new key first; fall back to legacy key for migration
    const saved = localStorage.getItem(STORAGE_KEY) ?? localStorage.getItem(STORAGE_KEY_LEGACY);
    if (saved) {
        const n = parseFloat(saved);
        if (!isNaN(n)) return clamp(n, MIN_RATIO, MAX_RATIO);
    }
    return DEFAULT_RATIO;
}

export default function MissionControlLayout() {
    return (
        <SignalProvider>
            <DashboardGrid />
        </SignalProvider>
    );
}

function DashboardGrid() {
    const { isConnected } = useSignalStream();
    const containerRef = useRef<HTMLDivElement>(null);
    const [ratio, setRatio] = useState(DEFAULT_RATIO);
    const [isLaunchCrewOpen, setIsLaunchCrewOpen] = useState(false);
    const dragging = useRef(false);

    // Load persisted ratio on mount
    useEffect(() => {
        setRatio(loadRatio());
    }, []);

    // Persist ratio on change
    useEffect(() => {
        localStorage.setItem(STORAGE_KEY, ratio.toString());
    }, [ratio]);

    const onPointerDown = useCallback((e: React.PointerEvent) => {
        e.preventDefault();
        dragging.current = true;
        (e.target as HTMLElement).setPointerCapture(e.pointerId);
    }, []);

    const onPointerMove = useCallback((e: React.PointerEvent) => {
        if (!dragging.current || !containerRef.current) return;
        const rect = containerRef.current.getBoundingClientRect();
        const y = e.clientY - rect.top;
        const newRatio = clamp(y / rect.height, MIN_RATIO, MAX_RATIO);
        setRatio(newRatio);
    }, []);

    const onPointerUp = useCallback(() => {
        dragging.current = false;
    }, []);

    const topPercent = `${(ratio * 100).toFixed(2)}%`;
    const bottomPercent = `${((1 - ratio) * 100).toFixed(2)}%`;

    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main relative">
            {/* Header */}
            <header className="h-12 border-b border-cortex-border flex items-center justify-between px-4 bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-4">
                    <h1 className="font-mono font-bold text-lg text-cortex-success flex items-center gap-2">
                        <Layers className="w-4 h-4" />
                        Workspace
                    </h1>
                    <span className="text-xs text-cortex-text-muted font-mono">
                        SIGNAL:{" "}
                        {isConnected ? (
                            <span className="text-cortex-success">LIVE</span>
                        ) : (
                            <span className="text-cortex-danger">OFFLINE</span>
                        )}
                    </span>
                </div>
                <div className="flex items-center gap-3">
                    <button
                        onClick={() => setIsLaunchCrewOpen(true)}
                        className="px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 hover:bg-cortex-primary/20 text-cortex-primary text-sm font-mono flex items-center gap-2 transition-all"
                    >
                        <Plus className="w-3 h-3" />
                        Launch Crew
                    </button>
                    <div className="h-4 w-px bg-cortex-border" />
                    <a
                        href="/settings"
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        <Settings className="w-4 h-4" />
                    </a>
                </div>
            </header>

            {/* Mode Ribbon */}
            <ModeRibbon />

            {/* Telemetry Row */}
            <TelemetryRow />

            {/* Resizable vertical split: Chat (top) | Ops Overview (bottom) */}
            <div
                ref={containerRef}
                className="flex-1 flex flex-col overflow-hidden"
                style={{ minHeight: 0 }}
                onPointerMove={onPointerMove}
                onPointerUp={onPointerUp}
            >
                {/* Top: Soma / Council Chat */}
                <div className="overflow-hidden" style={{ height: topPercent, minHeight: 0 }}>
                    <MissionControlChat />
                </div>

                {/* Resize Handle */}
                <div
                    onPointerDown={onPointerDown}
                    className="flex-shrink-0 flex items-center justify-center group"
                    style={{
                        height: 6,
                        cursor: "row-resize",
                        background: "#27272a",
                        touchAction: "none",
                        userSelect: "none",
                    }}
                >
                    <div className="w-10 h-0.5 rounded-full bg-cortex-text-muted/40 group-hover:bg-cortex-primary transition-colors" />
                </div>

                {/* Bottom: Ops Overview */}
                <div className="overflow-auto" style={{ height: bottomPercent, minHeight: 0 }}>
                    <OpsOverview />
                </div>
            </div>

            {/* Signal Detail Drawer */}
            <SignalDetailDrawer />

            {/* Launch Crew Modal */}
            {isLaunchCrewOpen && (
                <LaunchCrewModal onClose={() => setIsLaunchCrewOpen(false)} />
            )}
        </div>
    );
}

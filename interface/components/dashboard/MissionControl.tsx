"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { SignalProvider, useSignalStream } from "./SignalContext";
import { Activity, Settings, Plus } from "lucide-react";
import TelemetryRow from "./TelemetryRow";
import ActiveMissionsBar from "./ActiveMissionsBar";
import PriorityStream from "./PriorityStream";
import ActivityStream from "./ActivityStream";
import SensorLibrary from "./SensorLibrary";
import ManifestationPanel from "./ManifestationPanel";

export default function MissionControlLayout() {
    return (
        <SignalProvider>
            <DashboardGrid />
        </SignalProvider>
    );
}

function DashboardGrid() {
    const { isConnected } = useSignalStream();
    const router = useRouter();

    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main">
            {/* Header */}
            <header className="h-12 border-b border-cortex-border flex items-center justify-between px-4 bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-4">
                    <h1 className="font-mono font-bold text-lg text-cortex-success flex items-center gap-2">
                        <Activity className="w-4 h-4" />
                        MISSION CONTROL
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
                        onClick={() => router.push("/wiring")}
                        className="px-3 py-1.5 rounded bg-cortex-primary/10 border border-cortex-primary/30 hover:bg-cortex-primary/20 text-cortex-primary text-xs font-mono flex items-center gap-2 transition-all"
                    >
                        <Plus className="w-3 h-3" />
                        NEW MISSION
                    </button>
                    <div className="h-4 w-px bg-cortex-border" />
                    <button
                        onClick={() => router.push("/settings")}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        <Settings className="w-4 h-4" />
                    </button>
                </div>
            </header>

            {/* Telemetry Row — Live Compute Cartography */}
            <TelemetryRow />

            {/* Active Missions Bar — compact running state */}
            <ActiveMissionsBar />

            {/* Main Content — 3-column responsive grid */}
            <div
                className="flex-1 overflow-hidden"
                style={{
                    display: "grid",
                    gridTemplateColumns: "minmax(200px, 280px) 1fr minmax(280px, 384px)",
                    minHeight: 0,
                }}
            >
                {/* Left: Sensor Library (grouped subscriptions) */}
                <div className="h-full overflow-hidden border-r border-cortex-border">
                    <SensorLibrary />
                </div>

                {/* Center: PriorityStream + ManifestationPanel (stacked) */}
                <div className="h-full flex flex-col min-w-0 overflow-hidden">
                    <div className="flex-1 min-h-0 overflow-hidden">
                        <PriorityStream />
                    </div>
                    <div className="h-64 flex-shrink-0 border-t border-cortex-border overflow-hidden">
                        <ManifestationPanel />
                    </div>
                </div>

                {/* Right: Activity Stream (full feed) */}
                <div className="h-full overflow-hidden">
                    <ActivityStream />
                </div>
            </div>
        </div>
    );
}

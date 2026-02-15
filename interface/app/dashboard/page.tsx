"use client";

import LogStream from "@/components/hud/LogStream";
import SystemStatus from "@/components/SystemStatus";
import { Vitality } from "@/components/hud/Vitality";
import MatrixGrid from "@/components/matrix/MatrixGrid";
import NetworkMap from "@/components/registry/NetworkMap";

export default function DashboardPage() {
    return (
        <div className="h-full grid grid-rows-[auto_1fr] gap-4 p-6 bg-cortex-bg text-cortex-text-main">
            {/* Top Row: KPI Deck (System Status) */}
            <div className="h-24">
                <SystemStatus />
            </div>

            {/* Main Grid: Telemetry, Matrix, Logs */}
            <div className="grid grid-cols-12 gap-4 h-full min-h-0">

                {/* Left Column: Network & Matrix (4 cols) */}
                <div className="col-span-12 lg:col-span-3 flex flex-col gap-4 min-h-0">
                    <div className="flex-1 min-h-0">
                        <MatrixGrid />
                    </div>
                    <div className="flex-1 min-h-0">
                        <NetworkMap />
                    </div>
                </div>

                {/* Middle Column: Vitality (5 cols) */}
                <div className="col-span-12 lg:col-span-6 h-full min-h-0">
                    <Vitality />
                </div>

                {/* Right Column: Logs (3 cols) */}
                <div className="col-span-12 lg:col-span-3 h-full min-h-0">
                    <LogStream />
                </div>
            </div>
        </div>
    );
}

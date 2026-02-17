"use client";

import { useState } from "react";
import { BarChart3 } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import SystemStatus from "@/components/SystemStatus";
import TelemetryRow from "@/components/dashboard/TelemetryRow";
import CognitiveStatusPanel from "@/components/workspace/CognitiveStatusPanel";
import TeamRoster from "@/components/dashboard/TeamRoster";
import AgentPanel from "@/components/dashboard/AgentPanel";
import LogStream from "@/components/hud/LogStream";

export default function DashboardPage() {
    const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const selectedTeam = teamsDetail.find((t) => t.id === selectedTeamId) ?? null;

    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main">
            {/* ── Header: System Identity + Heartbeat ── */}
            <header className="h-12 border-b border-cortex-border flex items-center justify-between px-6 bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-3">
                    <BarChart3 className="w-4 h-4 text-cortex-primary" />
                    <h1 className="font-mono font-bold text-sm text-cortex-text-main tracking-wide">
                        SYSTEM OBSERVATORY
                    </h1>
                </div>
                <SystemStatus />
            </header>

            {/* ── Telemetry Strip: Resource Metabolism ── */}
            <TelemetryRow />

            {/* ── Diagnostic Grid ── */}
            <div className="flex-1 grid grid-cols-12 gap-4 p-4 min-h-0 overflow-hidden">

                {/* Left: Cognitive Readiness + Team Roster
                     Stacked — engine health on top, team selector below.
                     Select a team here to populate the center panel. */}
                <div className="col-span-12 lg:col-span-3 flex flex-col gap-3 min-h-0 overflow-hidden">
                    <CognitiveStatusPanel />
                    <div className="flex-1 bg-cortex-bg border border-cortex-border rounded-xl overflow-hidden min-h-0">
                        <TeamRoster
                            selectedId={selectedTeamId}
                            onSelect={setSelectedTeamId}
                        />
                    </div>
                </div>

                {/* Center: Agent Panel
                     Populated when a team is selected from the roster.
                     Agent rows with checkboxes, click-to-expand details,
                     bulk actions on selection. */}
                <div className="col-span-12 lg:col-span-5 min-h-0">
                    <AgentPanel team={selectedTeam} />
                </div>

                {/* Right: Event Stream
                     Historical log entries from the Archivist memory stream. */}
                <div className="col-span-12 lg:col-span-4 min-h-0">
                    <LogStream />
                </div>
            </div>
        </div>
    );
}

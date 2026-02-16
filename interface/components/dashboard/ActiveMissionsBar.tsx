"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Radar, ExternalLink } from "lucide-react";
import { useCortexStore, type Mission } from "@/store/useCortexStore";

const STATUS_DOT: Record<string, string> = {
    active: "bg-cortex-success animate-pulse",
    completed: "bg-cortex-text-muted",
    failed: "bg-cortex-danger",
};

function MissionChip({ mission }: { mission: Mission }) {
    const router = useRouter();

    return (
        <button
            onClick={() => router.push(`/missions/${mission.id}/teams`)}
            className="flex items-center gap-2 px-2.5 py-1 rounded-md bg-cortex-bg border border-cortex-border hover:border-cortex-primary/40 transition-colors group"
        >
            <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${STATUS_DOT[mission.status] ?? STATUS_DOT.active}`} />
            <span className="text-[10px] font-mono font-bold text-cortex-text-main truncate max-w-[120px]">
                {mission.id}
            </span>
            <span className="text-[9px] font-mono text-cortex-text-muted uppercase">
                {mission.status}
            </span>
            <span className="text-[9px] font-mono text-cortex-text-muted">
                {mission.teams}T/{mission.agents}A
            </span>
            <ExternalLink className="w-2.5 h-2.5 text-cortex-text-muted opacity-0 group-hover:opacity-100 transition-opacity" />
        </button>
    );
}

export default function ActiveMissionsBar() {
    const missions = useCortexStore((s) => s.missions);
    const isFetching = useCortexStore((s) => s.isFetchingMissions);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);

    // Also check live in-memory mission
    const activeMissionId = useCortexStore((s) => s.activeMissionId);
    const missionStatus = useCortexStore((s) => s.missionStatus);
    const blueprint = useCortexStore((s) => s.blueprint);

    useEffect(() => {
        fetchMissions();
        const interval = setInterval(fetchMissions, 15000);
        return () => clearInterval(interval);
    }, [fetchMissions]);

    const liveMission: Mission | null =
        missionStatus === "active" && activeMissionId && blueprint
            ? {
                  id: activeMissionId,
                  intent: blueprint.intent,
                  status: "active",
                  teams: blueprint.teams.length,
                  agents: blueprint.teams.reduce((sum, t) => sum + t.agents.length, 0),
              }
            : null;

    const allMissions = liveMission
        ? [liveMission, ...missions.filter((m) => m.id !== liveMission.id)]
        : missions;

    const activeCount = allMissions.filter((m) => m.status === "active").length;

    return (
        <div className="flex items-center gap-3 px-4 py-2 border-b border-cortex-border bg-cortex-surface/30" data-testid="active-missions-bar">
            <div className="flex items-center gap-1.5 flex-shrink-0">
                <Radar className="w-3.5 h-3.5 text-cortex-primary" />
                <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider">
                    Missions
                </span>
                {activeCount > 0 && (
                    <span className="text-[9px] font-mono text-cortex-success bg-cortex-success/10 px-1 py-0.5 rounded">
                        {activeCount} LIVE
                    </span>
                )}
            </div>

            <div className="h-4 w-px bg-cortex-border flex-shrink-0" />

            <div className="flex-1 overflow-x-auto scrollbar-none">
                <div className="flex items-center gap-2">
                    {isFetching && allMissions.length === 0 && (
                        <span className="text-[10px] font-mono text-cortex-text-muted animate-pulse">
                            Loading...
                        </span>
                    )}
                    {!isFetching && allMissions.length === 0 && (
                        <span className="text-[10px] font-mono text-cortex-text-muted">
                            No active missions
                        </span>
                    )}
                    {allMissions.map((mission) => (
                        <MissionChip key={mission.id} mission={mission} />
                    ))}
                </div>
            </div>
        </div>
    );
}

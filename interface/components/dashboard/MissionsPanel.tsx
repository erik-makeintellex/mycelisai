"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useCortexStore } from "@/store/useCortexStore";
import { Radar, ArrowRight, Loader2 } from "lucide-react";

export default function MissionsPanel() {
    const router = useRouter();
    const missions = useCortexStore((s) => s.missions);
    const isFetching = useCortexStore((s) => s.isFetchingMissions);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);

    // Also check if there's a live mission from the wiring page
    const activeMissionId = useCortexStore((s) => s.activeMissionId);
    const missionStatus = useCortexStore((s) => s.missionStatus);
    const blueprint = useCortexStore((s) => s.blueprint);

    useEffect(() => {
        fetchMissions();
    }, [fetchMissions]);

    // Build a combined list: API missions + any live in-memory mission
    const liveMission =
        missionStatus === "active" && activeMissionId && blueprint
            ? {
                  id: activeMissionId,
                  intent: blueprint.intent,
                  status: "active" as const,
                  teams: blueprint.teams.length,
                  agents: blueprint.teams.reduce(
                      (sum, t) => sum + t.agents.length,
                      0
                  ),
              }
            : null;

    const allMissions = liveMission
        ? [liveMission, ...missions.filter((m) => m.id !== liveMission.id)]
        : missions;

    if (isFetching) {
        return (
            <div className="h-full flex items-center justify-center">
                <Loader2 className="w-6 h-6 text-cortex-primary animate-spin" />
            </div>
        );
    }

    // UX-03: Zero Active Missions empty state
    if (allMissions.length === 0) {
        return (
            <div className="h-full flex items-center justify-center p-8">
                <div className="text-center max-w-sm">
                    <div className="w-16 h-16 mx-auto mb-6 rounded-full bg-cortex-primary/10 flex items-center justify-center">
                        <Radar className="w-8 h-8 text-cortex-primary/60" />
                    </div>
                    <h2 className="text-lg font-mono font-bold text-cortex-text-main mb-2">
                        Zero Active Missions
                    </h2>
                    <p className="text-sm text-cortex-text-muted mb-6 leading-relaxed">
                        Awaiting Architect Directives. Navigate to Neural
                        Wiring to design and deploy your first mission
                        blueprint.
                    </p>
                    <button
                        onClick={() => router.push("/wiring")}
                        className="px-4 py-2 rounded-lg bg-cortex-primary/10 border border-cortex-primary/30 hover:bg-cortex-primary/20 text-cortex-primary text-sm font-mono flex items-center gap-2 mx-auto transition-all"
                    >
                        Open Neural Wiring
                        <ArrowRight className="w-4 h-4" />
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="h-full overflow-y-auto p-6">
            <h2 className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-widest mb-4">
                Active Missions
            </h2>
            <div className="space-y-3">
                {allMissions.map((mission) => (
                    <button
                        key={mission.id}
                        onClick={() => router.push("/wiring")}
                        className="w-full text-left bg-cortex-surface rounded-xl border border-cortex-border hover:border-cortex-primary/40 p-4 transition-all group"
                    >
                        <div className="flex items-start justify-between mb-2">
                            <div className="flex items-center gap-2">
                                <span
                                    className={`w-2 h-2 rounded-full ${
                                        mission.status === "active"
                                            ? "bg-cortex-success animate-pulse"
                                            : mission.status === "failed"
                                              ? "bg-red-500"
                                              : "bg-cortex-text-muted"
                                    }`}
                                />
                                <span className="font-mono text-sm font-bold text-cortex-text-main">
                                    {mission.id}
                                </span>
                            </div>
                            <span className="text-[10px] font-mono text-cortex-text-muted uppercase">
                                {mission.status}
                            </span>
                        </div>
                        <p className="text-sm text-cortex-text-muted mb-3 line-clamp-2">
                            {mission.intent}
                        </p>
                        <div className="flex gap-3 text-[11px] font-mono text-cortex-text-muted">
                            <span>
                                {mission.teams} team
                                {mission.teams !== 1 ? "s" : ""}
                            </span>
                            <span className="text-cortex-border">|</span>
                            <span>
                                {mission.agents} agent
                                {mission.agents !== 1 ? "s" : ""}
                            </span>
                        </div>
                    </button>
                ))}
            </div>
        </div>
    );
}

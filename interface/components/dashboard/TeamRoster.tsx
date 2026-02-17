"use client";

import { useEffect } from "react";
import { useCortexStore } from "@/store/useCortexStore";
import { Users, Circle } from "lucide-react";

interface TeamRosterProps {
    selectedId: string | null;
    onSelect: (id: string) => void;
}

export default function TeamRoster({ selectedId, onSelect }: TeamRosterProps) {
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const isFetching = useCortexStore((s) => s.isFetchingTeamsDetail);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);

    useEffect(() => {
        fetchTeamsDetail();
        const interval = setInterval(fetchTeamsDetail, 10000);
        return () => clearInterval(interval);
    }, [fetchTeamsDetail]);

    if (isFetching && teamsDetail.length === 0) {
        return (
            <div className="p-3">
                <span className="text-[10px] text-cortex-text-muted font-mono animate-pulse">
                    Loading teams...
                </span>
            </div>
        );
    }

    return (
        <div className="flex flex-col">
            {/* Header */}
            <div className="px-4 py-2.5 flex items-center justify-between">
                <h3 className="text-[10px] font-bold text-cortex-text-muted flex items-center gap-2 font-mono uppercase tracking-wider">
                    <Users size={10} className="text-cortex-primary" />
                    Teams ({teamsDetail.length})
                </h3>
            </div>

            {/* Team List */}
            <div className="flex-1 overflow-y-auto px-2 pb-2 space-y-0.5 custom-scrollbar">
                {teamsDetail.length === 0 && (
                    <div className="text-center py-6 text-cortex-text-muted text-[10px] italic font-mono">
                        No teams deployed.
                    </div>
                )}

                {teamsDetail.map((team) => {
                    const isSelected = selectedId === team.id;
                    const busyCount = team.agents.filter((a) => a.status === 2).length;
                    const hasError = team.agents.some((a) => a.status === 3);

                    return (
                        <button
                            key={team.id}
                            onClick={() => onSelect(team.id)}
                            className={`w-full text-left px-3 py-2 rounded-lg transition-colors flex items-center gap-2.5 group ${
                                isSelected
                                    ? "bg-cortex-primary/10 border border-cortex-primary/30"
                                    : "hover:bg-cortex-surface border border-transparent"
                            }`}
                        >
                            {/* Status dot */}
                            <Circle
                                size={6}
                                className={`flex-shrink-0 fill-current ${
                                    hasError ? "text-red-400" :
                                    busyCount > 0 ? "text-cortex-primary" :
                                    team.agents.length > 0 ? "text-cortex-success" :
                                    "text-cortex-text-muted"
                                }`}
                            />

                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-1.5">
                                    <span className={`text-[11px] font-mono font-semibold truncate ${
                                        isSelected ? "text-cortex-text-main" : "text-cortex-text-main/80"
                                    }`}>
                                        {team.name}
                                    </span>
                                </div>
                                <span className="text-[9px] text-cortex-text-muted font-mono">
                                    {team.type} · {team.agents.length} agent{team.agents.length !== 1 ? "s" : ""}
                                </span>
                            </div>
                        </button>
                    );
                })}
            </div>
        </div>
    );
}

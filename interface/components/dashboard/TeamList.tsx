"use client";

import React, { useEffect, useState } from "react";
import { Users, Terminal, Shield, Zap } from "lucide-react";
import { useSignalStream } from "./SignalContext";

interface Team {
    id: string;
    name: string;
    type: string;
    role?: string;
}

export default function TeamList() {
    const [teams, setTeams] = useState<Team[]>([]);
    const [loading, setLoading] = useState(true);
    const { signals } = useSignalStream();

    useEffect(() => {
        fetch("/api/v1/teams")
            .then((res) => res.json())
            .then((data) => {
                setTeams(data);
                setLoading(false);
            })
            .catch(() => {
                setLoading(false);
            });
    }, []);

    if (loading) {
        return (
            <div className="p-4 text-cortex-text-muted font-mono text-sm animate-pulse">
                Scanning Neural Fabric...
            </div>
        );
    }

    if (teams.length === 0) {
        return (
            <div className="p-8 text-center border-2 border-dashed border-cortex-border rounded-lg">
                <Users className="w-8 h-8 mx-auto text-cortex-text-muted/50 mb-2" />
                <p className="text-cortex-text-muted font-mono text-sm">
                    NO ACTIVE TEAMS DETECTED
                </p>
            </div>
        );
    }

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {teams.map((team) => {
                const recentActivity = signals.find(
                    (s) =>
                        s.source?.includes(team.id) ||
                        (s.payload &&
                            JSON.stringify(s.payload).includes(team.id))
                );

                return (
                    <div
                        key={team.id}
                        className="bg-cortex-surface border border-cortex-border rounded-xl p-4 hover:border-cortex-primary/40 transition-colors group"
                    >
                        <div className="flex justify-between items-start mb-2">
                            <div className="flex items-center gap-2">
                                <div
                                    className={`p-1.5 rounded bg-cortex-bg ${recentActivity ? "text-cortex-success" : "text-cortex-text-muted"}`}
                                >
                                    {team.type === "action" ? (
                                        <Zap className="w-4 h-4" />
                                    ) : (
                                        <Shield className="w-4 h-4" />
                                    )}
                                </div>
                                <div>
                                    <h3 className="font-bold text-cortex-text-main text-sm">
                                        {team.name}
                                    </h3>
                                    <span className="text-[10px] uppercase tracking-wider text-cortex-text-muted font-mono block">
                                        {team.id}
                                    </span>
                                </div>
                            </div>
                            <div className="flex items-center gap-1">
                                <span
                                    className={`w-1.5 h-1.5 rounded-full ${recentActivity ? "bg-cortex-success animate-pulse" : "bg-cortex-border"}`}
                                />
                                <span className="text-[10px] text-cortex-text-muted font-mono">
                                    {recentActivity ? "BUSY" : "IDLE"}
                                </span>
                            </div>
                        </div>

                        <div className="mt-3 flex gap-2">
                            <div className="bg-cortex-bg px-2 py-1 rounded text-xs font-mono text-cortex-text-muted flex items-center gap-1">
                                <Users className="w-3 h-3" /> 1
                            </div>
                            <div className="bg-cortex-bg px-2 py-1 rounded text-xs font-mono text-cortex-text-muted flex items-center gap-1">
                                <Terminal className="w-3 h-3" /> CMD
                            </div>
                        </div>
                    </div>
                );
            })}
        </div>
    );
}

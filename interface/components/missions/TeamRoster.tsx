"use client";

import React, { useMemo } from "react";
import { Users, Bot, Clock, Radio } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

interface TeamRosterProps {
    missionId: string;
    teamCount: number;
    agentCount: number;
}

interface DerivedAgent {
    id: string;
    lastSeen: string;
    signalCount: number;
}

export default function TeamRoster({
    missionId,
    teamCount,
    agentCount,
}: TeamRosterProps) {
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const missions = useCortexStore((s) => s.missions);

    const mission = useMemo(
        () => missions.find((m) => m.id === missionId),
        [missions, missionId]
    );

    const isActive = mission?.status === "active";

    // Derive unique agents from stream logs that have a source
    const derivedAgents = useMemo(() => {
        const agentMap = new Map<string, DerivedAgent>();

        streamLogs.forEach((log) => {
            if (!log.source) return;

            const existing = agentMap.get(log.source);
            if (existing) {
                existing.signalCount += 1;
                if (
                    log.timestamp &&
                    log.timestamp > existing.lastSeen
                ) {
                    existing.lastSeen = log.timestamp;
                }
            } else {
                agentMap.set(log.source, {
                    id: log.source,
                    lastSeen: log.timestamp ?? new Date().toISOString(),
                    signalCount: 1,
                });
            }
        });

        return Array.from(agentMap.values()).sort(
            (a, b) => b.signalCount - a.signalCount
        );
    }, [streamLogs]);

    const formatTimestamp = (ts: string) => {
        try {
            return new Date(ts).toLocaleTimeString([], {
                hour: "2-digit",
                minute: "2-digit",
                second: "2-digit",
                hour12: false,
            });
        } catch {
            return "--:--:--";
        }
    };

    // Deterministic color from agent ID
    const agentColor = (id: string): string => {
        const colors = [
            "bg-cortex-primary",
            "bg-cortex-success",
            "bg-cortex-info",
            "bg-cortex-warning",
            "bg-cortex-danger",
        ];
        let hash = 0;
        for (let i = 0; i < id.length; i++) {
            hash = id.charCodeAt(i) + ((hash << 5) - hash);
        }
        return colors[Math.abs(hash) % colors.length];
    };

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            {/* Section header */}
            <div className="p-3 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center gap-2">
                <Users className="w-3.5 h-3.5 text-cortex-text-muted" />
                <h3 className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-widest">
                    Team Roster
                </h3>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {/* Summary card */}
                <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4">
                    <div className="flex items-center justify-between mb-3">
                        <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30">
                            Overview
                        </span>
                        <div className="flex items-center gap-1.5">
                            <span
                                className={`w-2 h-2 rounded-full ${
                                    isActive
                                        ? "bg-cortex-success animate-pulse"
                                        : "bg-cortex-text-muted/40"
                                }`}
                            />
                            <span className="text-[10px] font-mono text-cortex-text-muted">
                                {isActive ? "ACTIVE" : "OFFLINE"}
                            </span>
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-3">
                        <div className="flex items-center gap-2">
                            <div className="p-1.5 rounded bg-cortex-primary/10">
                                <Users className="w-3.5 h-3.5 text-cortex-primary" />
                            </div>
                            <div>
                                <p className="text-lg font-bold font-mono text-cortex-text-main leading-none">
                                    {teamCount}
                                </p>
                                <p className="text-[10px] font-mono text-cortex-text-muted">
                                    Teams
                                </p>
                            </div>
                        </div>
                        <div className="flex items-center gap-2">
                            <div className="p-1.5 rounded bg-cortex-success/10">
                                <Bot className="w-3.5 h-3.5 text-cortex-success" />
                            </div>
                            <div>
                                <p className="text-lg font-bold font-mono text-cortex-text-main leading-none">
                                    {agentCount}
                                </p>
                                <p className="text-[10px] font-mono text-cortex-text-muted">
                                    Agents
                                </p>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Recent agents section */}
                <div>
                    <div className="flex items-center gap-2 mb-3">
                        <Radio className="w-3 h-3 text-cortex-text-muted" />
                        <h4 className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-widest">
                            Recent Agents
                        </h4>
                        {derivedAgents.length > 0 && (
                            <span className="text-[9px] font-mono text-cortex-text-muted/60 ml-auto">
                                {derivedAgents.length} detected
                            </span>
                        )}
                    </div>

                    {derivedAgents.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-8">
                            <Bot className="w-8 h-8 opacity-20 text-cortex-text-muted mb-2" />
                            <p className="font-mono text-xs text-cortex-text-muted/60">
                                No agent signals detected
                            </p>
                        </div>
                    ) : (
                        <div className="space-y-1">
                            {derivedAgents.map((agent) => (
                                <div
                                    key={agent.id}
                                    className="flex items-center gap-2.5 px-3 py-2 rounded-lg hover:bg-cortex-surface/80 transition-colors group"
                                >
                                    {/* Color dot */}
                                    <span
                                        className={`w-2 h-2 rounded-full flex-shrink-0 ${agentColor(
                                            agent.id
                                        )}`}
                                    />

                                    {/* Agent name */}
                                    <span className="text-xs font-mono text-cortex-text-main truncate flex-1">
                                        {agent.id}
                                    </span>

                                    {/* Last seen */}
                                    <span className="text-[9px] font-mono text-cortex-text-muted/60 flex items-center gap-1 flex-shrink-0">
                                        <Clock className="w-2.5 h-2.5" />
                                        {formatTimestamp(agent.lastSeen)}
                                    </span>

                                    {/* Signal count */}
                                    <span className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-cortex-border/50 text-cortex-text-muted flex-shrink-0">
                                        {agent.signalCount}
                                    </span>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

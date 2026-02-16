"use client";

import React, { useEffect } from "react";
import { Users, Bot, Package } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

function MetricPill({ icon: Icon, label, value, color }: {
    icon: any;
    label: string;
    value: number;
    color: string;
}) {
    return (
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-cortex-bg border border-cortex-border">
            <Icon className={`w-3.5 h-3.5 ${color}`} />
            <span className="text-[10px] font-mono text-cortex-text-muted uppercase">{label}</span>
            <span className="text-sm font-mono font-bold text-cortex-text-main">{value}</span>
        </div>
    );
}

export default function TeamsSummaryCard() {
    const missions = useCortexStore((s) => s.missions);
    const artifacts = useCortexStore((s) => s.artifacts);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);
    const fetchArtifacts = useCortexStore((s) => s.fetchArtifacts);

    useEffect(() => {
        fetchMissions();
        fetchArtifacts();
    }, [fetchMissions, fetchArtifacts]);

    const activeMissions = missions.filter((m) => m.status === "active");
    const totalTeams = activeMissions.reduce((sum, m) => sum + m.teams, 0);
    const totalAgents = activeMissions.reduce((sum, m) => sum + m.agents, 0);
    const recentOutputs = artifacts.length;

    return (
        <div className="flex items-center gap-3 px-4 py-1.5 border-b border-cortex-border bg-cortex-surface/20" data-testid="teams-summary">
            <MetricPill icon={Users} label="Teams" value={totalTeams} color="text-cortex-primary" />
            <MetricPill icon={Bot} label="Agents" value={totalAgents} color="text-cortex-info" />
            <MetricPill icon={Package} label="Outputs" value={recentOutputs} color="text-cortex-success" />
        </div>
    );
}

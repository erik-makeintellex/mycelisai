"use client";

import React, { useMemo } from "react";
import {
    Users,
    Bot,
    Package,
    Activity,
    Clock,
    CheckCircle,
    XCircle,
    Hourglass,
    Archive,
} from "lucide-react";
import {
    useCortexStore,
    type ArtifactStatus,
} from "@/store/useCortexStore";

interface MissionSummaryTabProps {
    missionId: string;
    teamCount: number;
    agentCount: number;
}

interface MetricCardProps {
    icon: React.ReactNode;
    label: string;
    value: number;
    accent: string;
}

function MetricCard({ icon, label, value, accent }: MetricCardProps) {
    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4">
            <div className="flex items-center justify-between mb-3">
                <div className={`p-2 rounded ${accent}`}>{icon}</div>
                <span className="text-2xl font-bold font-mono text-cortex-text-main">
                    {value}
                </span>
            </div>
            <p className="text-[10px] font-mono text-cortex-text-muted uppercase tracking-widest">
                {label}
            </p>
        </div>
    );
}

function statusIcon(status: ArtifactStatus) {
    switch (status) {
        case "pending":
            return <Hourglass className="w-3 h-3 text-cortex-warning" />;
        case "approved":
            return <CheckCircle className="w-3 h-3 text-cortex-success" />;
        case "rejected":
            return <XCircle className="w-3 h-3 text-cortex-danger" />;
        case "archived":
            return <Archive className="w-3 h-3 text-cortex-text-muted" />;
    }
}

function statusColor(status: ArtifactStatus): string {
    switch (status) {
        case "pending":
            return "text-cortex-warning";
        case "approved":
            return "text-cortex-success";
        case "rejected":
            return "text-cortex-danger";
        case "archived":
            return "text-cortex-text-muted";
    }
}

export default function MissionSummaryTab({
    missionId,
    teamCount,
    agentCount,
}: MissionSummaryTabProps) {
    const artifacts = useCortexStore((s) => s.artifacts);
    const streamLogs = useCortexStore((s) => s.streamLogs);

    const recentArtifacts = useMemo(() => {
        return [...artifacts]
            .sort(
                (a, b) =>
                    new Date(b.created_at).getTime() -
                    new Date(a.created_at).getTime()
            )
            .slice(0, 5);
    }, [artifacts]);

    const statusBreakdown = useMemo(() => {
        const counts: Record<ArtifactStatus, number> = {
            pending: 0,
            approved: 0,
            rejected: 0,
            archived: 0,
        };
        artifacts.forEach((a) => {
            if (counts[a.status] !== undefined) {
                counts[a.status]++;
            }
        });
        return counts;
    }, [artifacts]);

    const formatDate = (ts: string) => {
        try {
            return new Date(ts).toLocaleString([], {
                dateStyle: "short",
                timeStyle: "short",
            });
        } catch {
            return ts;
        }
    };

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            {/* Header */}
            <div className="p-3 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center gap-2 flex-shrink-0">
                <Activity className="w-3.5 h-3.5 text-cortex-text-muted" />
                <h3 className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-widest">
                    Mission Summary
                </h3>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-6">
                {/* 2x2 Metric cards */}
                <div className="grid grid-cols-2 gap-3">
                    <MetricCard
                        icon={<Users className="w-4 h-4 text-cortex-primary" />}
                        label="Teams"
                        value={teamCount}
                        accent="bg-cortex-primary/10"
                    />
                    <MetricCard
                        icon={<Bot className="w-4 h-4 text-cortex-success" />}
                        label="Agents"
                        value={agentCount}
                        accent="bg-cortex-success/10"
                    />
                    <MetricCard
                        icon={<Package className="w-4 h-4 text-cortex-info" />}
                        label="Artifacts"
                        value={artifacts.length}
                        accent="bg-cortex-info/10"
                    />
                    <MetricCard
                        icon={<Activity className="w-4 h-4 text-cortex-warning" />}
                        label="Signals"
                        value={streamLogs.length}
                        accent="bg-cortex-warning/10"
                    />
                </div>

                {/* Recent artifacts */}
                <div>
                    <h4 className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-widest mb-3">
                        Recent Artifacts
                    </h4>
                    {recentArtifacts.length === 0 ? (
                        <div className="flex flex-col items-center py-6">
                            <Package className="w-8 h-8 opacity-20 text-cortex-text-muted mb-2" />
                            <p className="font-mono text-xs text-cortex-text-muted/60">
                                No artifacts yet
                            </p>
                        </div>
                    ) : (
                        <div className="space-y-1">
                            {recentArtifacts.map((artifact) => (
                                <div
                                    key={artifact.id}
                                    className="flex items-center gap-2.5 px-3 py-2 rounded-lg hover:bg-cortex-surface/80 transition-colors"
                                >
                                    {statusIcon(artifact.status)}
                                    <span className="text-xs font-mono text-cortex-text-main truncate flex-1">
                                        {artifact.title}
                                    </span>
                                    <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-bg text-cortex-text-muted border border-cortex-border flex-shrink-0">
                                        {artifact.artifact_type}
                                    </span>
                                    <span className="text-[9px] font-mono text-cortex-text-muted/50 flex items-center gap-1 flex-shrink-0">
                                        <Clock className="w-2.5 h-2.5" />
                                        {formatDate(artifact.created_at)}
                                    </span>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* Status breakdown */}
                <div>
                    <h4 className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-widest mb-3">
                        Status Breakdown
                    </h4>
                    <div className="bg-cortex-surface border border-cortex-border rounded-xl p-4">
                        <div className="space-y-3">
                            {(
                                Object.entries(statusBreakdown) as [
                                    ArtifactStatus,
                                    number
                                ][]
                            ).map(([status, count]) => {
                                const total = artifacts.length || 1;
                                const pct = (count / total) * 100;
                                return (
                                    <div key={status}>
                                        <div className="flex items-center justify-between mb-1">
                                            <div className="flex items-center gap-1.5">
                                                {statusIcon(status)}
                                                <span
                                                    className={`text-[10px] font-mono uppercase ${statusColor(
                                                        status
                                                    )}`}
                                                >
                                                    {status}
                                                </span>
                                            </div>
                                            <span className="text-xs font-mono font-bold text-cortex-text-main">
                                                {count}
                                            </span>
                                        </div>
                                        <div className="h-1.5 bg-cortex-bg rounded-full overflow-hidden">
                                            <div
                                                className={`h-full rounded-full transition-all duration-500 ${
                                                    status === "pending"
                                                        ? "bg-cortex-warning"
                                                        : status === "approved"
                                                        ? "bg-cortex-success"
                                                        : status === "rejected"
                                                        ? "bg-cortex-danger"
                                                        : "bg-cortex-text-muted/40"
                                                }`}
                                                style={{ width: `${pct}%` }}
                                            />
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

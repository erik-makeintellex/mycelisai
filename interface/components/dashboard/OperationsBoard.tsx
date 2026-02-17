"use client";

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
    ChevronDown,
    ChevronRight,
    AlertTriangle,
    ShieldAlert,
    CheckCircle,
    Package,
    Users,
    Radar,
    ExternalLink,
} from "lucide-react";
import { useCortexStore, type TeamDetailEntry } from "@/store/useCortexStore";
import { useSignalStream, type Signal } from "./SignalContext";

// ── Priority Alert types ──────────────────────────────────────

const PRIORITY_TYPES = new Set(["governance_halt", "error", "task_complete", "artifact"]);

const ALERT_CONFIG: Record<string, { icon: React.ComponentType<{ className?: string }>; color: string; label: string }> = {
    governance_halt: { icon: ShieldAlert, color: "text-cortex-warning", label: "GOVERNANCE" },
    error:          { icon: AlertTriangle, color: "text-cortex-danger", label: "ERROR" },
    task_complete:  { icon: CheckCircle, color: "text-cortex-success", label: "COMPLETE" },
    artifact:       { icon: Package, color: "text-cortex-info", label: "ARTIFACT" },
};

// ── Status helpers ────────────────────────────────────────────

function aggregateStatus(agents: { status: number }[]): "error" | "busy" | "idle" | "offline" {
    if (agents.length === 0) return "offline";
    if (agents.some((a) => a.status === 3)) return "error";
    if (agents.some((a) => a.status === 2)) return "busy";
    if (agents.some((a) => a.status === 1)) return "idle";
    return "offline";
}

const STATUS_DOT: Record<string, string> = {
    error:   "bg-cortex-danger",
    busy:    "bg-cortex-info animate-pulse",
    idle:    "bg-cortex-success",
    offline: "bg-cortex-text-muted/40",
};

const STATUS_LABEL: Record<string, string> = {
    error:   "ERROR",
    busy:    "BUSY",
    idle:    "IDLE",
    offline: "OFFLINE",
};

// ── Section Header ────────────────────────────────────────────

function SectionHeader({ title, count, children }: {
    title: string;
    count: number;
    children?: React.ReactNode;
}) {
    return (
        <div className="flex items-center gap-2 px-4 py-2 border-b border-cortex-border/50">
            <span className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
                {title}
            </span>
            <span className="text-[9px] font-mono text-cortex-text-muted/60">
                {count}
            </span>
            {children}
        </div>
    );
}

// ── Priority Alerts Section ───────────────────────────────────

function PriorityAlerts() {
    const { signals } = useSignalStream();
    const [collapsed, setCollapsed] = useState(false);

    const alerts = signals.filter((s) => PRIORITY_TYPES.has(s.type)).slice(0, 10);

    if (alerts.length === 0) return null;

    return (
        <div className="border-b border-cortex-border">
            <button
                onClick={() => setCollapsed(!collapsed)}
                className="w-full flex items-center gap-2 px-4 py-2 hover:bg-cortex-bg/30 transition-colors"
            >
                {collapsed ? (
                    <ChevronRight className="w-3 h-3 text-cortex-text-muted" />
                ) : (
                    <ChevronDown className="w-3 h-3 text-cortex-text-muted" />
                )}
                <ShieldAlert className="w-3.5 h-3.5 text-cortex-warning" />
                <span className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-warning">
                    Priority Alerts
                </span>
                <span className="text-[9px] font-mono text-cortex-warning bg-cortex-warning/10 px-1.5 py-0.5 rounded">
                    {alerts.length}
                </span>
            </button>

            {!collapsed && (
                <div className="max-h-48 overflow-y-auto">
                    {alerts.map((signal, i) => (
                        <AlertRow key={`${signal.type}-${signal.timestamp}-${i}`} signal={signal} />
                    ))}
                </div>
            )}
        </div>
    );
}

function AlertRow({ signal }: { signal: Signal }) {
    const config = ALERT_CONFIG[signal.type] ?? ALERT_CONFIG.error;
    const Icon = config.icon;
    const time = signal.timestamp
        ? new Date(signal.timestamp).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })
        : "";

    return (
        <div className="px-4 py-1.5 flex items-start gap-2 hover:bg-cortex-bg/20 transition-colors">
            <Icon className={`w-3 h-3 mt-0.5 flex-shrink-0 ${config.color}`} />
            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                    <span className={`text-[8px] font-mono font-bold uppercase ${config.color}`}>
                        {config.label}
                    </span>
                    {signal.source && (
                        <span className="text-[8px] font-mono text-cortex-text-muted truncate">
                            {signal.source}
                        </span>
                    )}
                    <span className="ml-auto text-[8px] font-mono text-cortex-text-muted/60 whitespace-nowrap">
                        {time}
                    </span>
                </div>
                {signal.message && (
                    <p className="text-[10px] text-cortex-text-main mt-0.5 line-clamp-1">
                        {signal.message}
                    </p>
                )}
            </div>
        </div>
    );
}

// ── Standing Workloads Section ────────────────────────────────

function WorkloadRow({ team }: { team: TeamDetailEntry }) {
    const status = aggregateStatus(team.agents);
    const onlineCount = team.agents.filter((a) => a.status >= 1 && a.status <= 2).length;

    return (
        <div className="px-4 py-2 flex items-center gap-3 hover:bg-cortex-bg/20 transition-colors">
            <span className={`w-2 h-2 rounded-full flex-shrink-0 ${STATUS_DOT[status]}`} />
            <Users className="w-3.5 h-3.5 text-cortex-text-muted flex-shrink-0" />
            <span className="text-xs font-mono font-bold text-cortex-text-main truncate flex-1">
                {team.name}
            </span>
            <span className="text-[9px] font-mono text-cortex-text-muted">
                {onlineCount}/{team.agents.length} agents
            </span>
            <span className={`text-[8px] font-mono font-bold uppercase ${STATUS_DOT[status].replace("bg-", "text-").replace(" animate-pulse", "")}`}>
                {STATUS_LABEL[status]}
            </span>
        </div>
    );
}

// ── Missions Section ──────────────────────────────────────────

function MissionRow({ mission, missionTeams }: {
    mission: { id: string; intent: string; status: string; teams: number; agents: number };
    missionTeams: TeamDetailEntry[];
}) {
    const router = useRouter();
    const allAgents = missionTeams.flatMap((t) => t.agents);
    const status = allAgents.length > 0 ? aggregateStatus(allAgents) : "offline";

    const STATUS_BADGE: Record<string, string> = {
        active:    "bg-cortex-success/10 text-cortex-success",
        completed: "bg-cortex-text-muted/10 text-cortex-text-muted",
        failed:    "bg-cortex-danger/10 text-cortex-danger",
    };

    return (
        <button
            onClick={() => router.push(`/missions/${mission.id}/teams`)}
            className="w-full px-4 py-2.5 flex items-center gap-3 hover:bg-cortex-bg/30 transition-colors group text-left"
        >
            <span className={`w-2 h-2 rounded-full flex-shrink-0 ${STATUS_DOT[status]}`} />
            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                    <span className="text-xs font-mono font-bold text-cortex-text-main truncate">
                        {mission.id}
                    </span>
                    <span className={`text-[8px] font-mono font-bold uppercase px-1.5 py-0.5 rounded ${STATUS_BADGE[mission.status] ?? STATUS_BADGE.active}`}>
                        {mission.status}
                    </span>
                </div>
                {mission.intent && (
                    <p className="text-[10px] text-cortex-text-muted mt-0.5 line-clamp-1">
                        {mission.intent}
                    </p>
                )}
            </div>
            <span className="text-[9px] font-mono text-cortex-text-muted whitespace-nowrap">
                {mission.teams}T/{mission.agents}A
            </span>
            <ExternalLink className="w-3 h-3 text-cortex-text-muted opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" />
        </button>
    );
}

// ── Main OperationsBoard ──────────────────────────────────────

export default function OperationsBoard() {
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const missions = useCortexStore((s) => s.missions);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);

    useEffect(() => {
        fetchTeamsDetail();
        fetchMissions();
        const teamsInterval = setInterval(fetchTeamsDetail, 10000);
        const missionsInterval = setInterval(fetchMissions, 15000);
        return () => {
            clearInterval(teamsInterval);
            clearInterval(missionsInterval);
        };
    }, [fetchTeamsDetail, fetchMissions]);

    const standingTeams = teamsDetail.filter((t) => t.type === "standing");
    const missionTeams = teamsDetail.filter((t) => t.type === "mission");

    // Group mission teams by mission_id for cross-reference
    const missionTeamMap = new Map<string, TeamDetailEntry[]>();
    for (const t of missionTeams) {
        if (t.mission_id) {
            const existing = missionTeamMap.get(t.mission_id) ?? [];
            existing.push(t);
            missionTeamMap.set(t.mission_id, existing);
        }
    }

    // Sort missions: active first, then completed, then failed
    const sortedMissions = [...missions].sort((a, b) => {
        const order: Record<string, number> = { active: 0, completed: 1, failed: 2 };
        return (order[a.status] ?? 1) - (order[b.status] ?? 1);
    });

    return (
        <div className="h-full flex flex-col bg-cortex-surface" data-testid="operations-board">
            {/* Priority alerts — collapses to zero when empty */}
            <PriorityAlerts />

            {/* Standing Workloads */}
            <div className="border-b border-cortex-border">
                <SectionHeader title="Standing Workloads" count={standingTeams.length} />
                {standingTeams.length === 0 ? (
                    <div className="px-4 py-4 text-center">
                        <p className="text-[10px] font-mono text-cortex-text-muted">No standing workloads</p>
                    </div>
                ) : (
                    standingTeams.map((team) => (
                        <WorkloadRow key={team.id} team={team} />
                    ))
                )}
            </div>

            {/* Missions */}
            <div className="flex-1 min-h-0 flex flex-col">
                <SectionHeader title="Missions" count={sortedMissions.length}>
                    {sortedMissions.filter((m) => m.status === "active").length > 0 && (
                        <span className="text-[9px] font-mono text-cortex-success bg-cortex-success/10 px-1.5 py-0.5 rounded">
                            {sortedMissions.filter((m) => m.status === "active").length} LIVE
                        </span>
                    )}
                </SectionHeader>
                <div className="flex-1 overflow-y-auto">
                    {sortedMissions.length === 0 ? (
                        <div className="px-4 py-8 text-center">
                            <Radar className="w-6 h-6 mx-auto mb-2 text-cortex-text-muted opacity-20" />
                            <p className="text-[10px] font-mono text-cortex-text-muted">No active missions</p>
                            <p className="text-[9px] font-mono text-cortex-text-muted/50 mt-1">
                                Wire one from the architect
                            </p>
                        </div>
                    ) : (
                        sortedMissions.map((mission) => (
                            <MissionRow
                                key={mission.id}
                                mission={mission}
                                missionTeams={missionTeamMap.get(mission.id) ?? []}
                            />
                        ))
                    )}
                </div>
            </div>
        </div>
    );
}

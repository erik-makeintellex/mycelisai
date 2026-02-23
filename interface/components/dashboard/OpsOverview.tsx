"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import {
    ExternalLink,
    Cpu,
    Image,
    Circle,
    Radio,
    Users,
    Radar,
    ShieldAlert,
    AlertTriangle,
    CheckCircle,
    Package,
    Wrench,
    ChevronDown,
    ChevronRight,
    Plus,
    Zap,
    Activity,
} from "lucide-react";
import { useCortexStore, type TeamDetailEntry, type MissionRun } from "@/store/useCortexStore";
import { useSignalStream, type Signal } from "./SignalContext";
import { streamSignalToDetail } from "@/lib/signalNormalize";
import { registerOpsWidget, getOpsWidgets } from "@/lib/opsWidgetRegistry";

// ── Shared Helpers ───────────────────────────────────────────

function SectionCard({
    title,
    subtitle,
    icon: Icon,
    href,
    children,
}: {
    title: string;
    subtitle: string;
    icon: React.ElementType;
    href: string;
    children: React.ReactNode;
}) {
    const router = useRouter();

    return (
        <div className="border border-cortex-border rounded-lg overflow-hidden">
            <div className="flex items-center gap-2.5 px-3.5 py-2.5 bg-cortex-bg/50">
                <Icon className="w-4 h-4 text-cortex-text-muted" />
                <div className="flex-1 min-w-0">
                    <span className="text-[13px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted block">
                        {title}
                    </span>
                    <span className="text-xs font-mono text-cortex-text-muted/60 block truncate">
                        {subtitle}
                    </span>
                </div>
                <button
                    onClick={() => router.push(href)}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors flex-shrink-0"
                    title={`Open ${title}`}
                >
                    <ExternalLink className="w-3.5 h-3.5" />
                </button>
            </div>
            {children}
        </div>
    );
}

// ── System Status ────────────────────────────────────────────

function SystemStatus() {
    const cognitiveStatus = useCortexStore((s) => s.cognitiveStatus);
    const fetchCognitiveStatus = useCortexStore((s) => s.fetchCognitiveStatus);
    const sensors = useCortexStore((s) => s.sensorFeeds);
    const fetchSensors = useCortexStore((s) => s.fetchSensors);

    useEffect(() => {
        fetchCognitiveStatus();
        fetchSensors();
        const cogInterval = setInterval(fetchCognitiveStatus, 15000);
        const sensorInterval = setInterval(fetchSensors, 60000);
        return () => {
            clearInterval(cogInterval);
            clearInterval(sensorInterval);
        };
    }, [fetchCognitiveStatus, fetchSensors]);

    const text = cognitiveStatus?.text;
    const media = cognitiveStatus?.media;
    const onlineSensors = sensors.filter((s) => s.status === "online").length;

    return (
        <SectionCard title="System" subtitle="LLM engines and sensor feeds powering the organism" icon={Cpu} href="/settings/brain">
            <div className="px-3.5 py-2.5 space-y-2">
                <EngineRow
                    label="Text"
                    sublabel={text?.model ?? "—"}
                    online={text?.status === "online"}
                />
                <EngineRow
                    label="Media"
                    sublabel={media?.model ?? "—"}
                    online={media?.status === "online"}
                />
                {sensors.length > 0 && (
                    <div className="flex items-center gap-2 pt-0.5">
                        <Radio className="w-3.5 h-3.5 text-cortex-text-muted" />
                        <span className="text-xs font-mono text-cortex-text-main">
                            Sensors
                        </span>
                        <span className="ml-auto text-xs font-mono text-cortex-text-muted">
                            {onlineSensors}/{sensors.length} online
                        </span>
                    </div>
                )}
            </div>
        </SectionCard>
    );
}

function EngineRow({
    label,
    sublabel,
    online,
}: {
    label: string;
    sublabel: string;
    online: boolean;
}) {
    return (
        <div className="flex items-center gap-2">
            <Circle
                className={`w-2 h-2 flex-shrink-0 ${
                    online
                        ? "fill-cortex-success text-cortex-success"
                        : "fill-cortex-danger text-cortex-danger"
                }`}
            />
            <span className="text-[13px] font-mono font-semibold text-cortex-text-main">
                {label}
            </span>
            <span className="text-xs font-mono text-cortex-text-muted truncate ml-auto">
                {sublabel}
            </span>
        </div>
    );
}

// ── Priority Alerts ──────────────────────────────────────────

const PRIORITY_TYPES = new Set(["governance_halt", "error", "task_complete", "artifact"]);

const ALERT_CONFIG: Record<
    string,
    { icon: React.ElementType; color: string; label: string }
> = {
    governance_halt: { icon: ShieldAlert, color: "text-cortex-warning", label: "GOV" },
    error: { icon: AlertTriangle, color: "text-cortex-danger", label: "ERR" },
    task_complete: { icon: CheckCircle, color: "text-cortex-success", label: "DONE" },
    artifact: { icon: Package, color: "text-cortex-info", label: "ART" },
};

function PriorityAlerts() {
    const { signals } = useSignalStream();
    const selectSignalDetail = useCortexStore((s) => s.selectSignalDetail);
    const [collapsed, setCollapsed] = React.useState(false);

    const alerts = signals.filter((s) => PRIORITY_TYPES.has(s.type)).slice(0, 8);

    if (alerts.length === 0) return null;

    return (
        <div className="border border-cortex-border rounded-lg overflow-hidden">
            <button
                onClick={() => setCollapsed(!collapsed)}
                className="w-full flex items-center gap-2.5 px-3.5 py-2.5 bg-cortex-bg/50 hover:bg-cortex-bg/80 transition-colors"
            >
                {collapsed ? (
                    <ChevronRight className="w-3.5 h-3.5 text-cortex-text-muted" />
                ) : (
                    <ChevronDown className="w-3.5 h-3.5 text-cortex-text-muted" />
                )}
                <ShieldAlert className="w-4 h-4 text-cortex-warning" />
                <span className="text-[13px] font-mono font-bold uppercase tracking-widest text-cortex-warning flex-1 text-left">
                    Alerts
                </span>
                <span className="text-xs font-mono text-cortex-warning bg-cortex-warning/10 px-2 py-0.5 rounded">
                    {alerts.length}
                </span>
            </button>

            {!collapsed && (
                <div className="max-h-40 overflow-y-auto">
                    {alerts.map((signal, i) => {
                        const config = ALERT_CONFIG[signal.type] ?? ALERT_CONFIG.error;
                        const Icon = config.icon;
                        return (
                            <div
                                key={`${signal.type}-${signal.timestamp}-${i}`}
                                className="px-3.5 py-2 flex items-center gap-2.5 hover:bg-cortex-bg/20 transition-colors cursor-pointer"
                                onClick={() => selectSignalDetail(streamSignalToDetail(signal))}
                            >
                                <Icon className={`w-3.5 h-3.5 flex-shrink-0 ${config.color}`} />
                                <span className={`text-xs font-mono font-bold ${config.color}`}>
                                    {config.label}
                                </span>
                                <span className="text-xs font-mono text-cortex-text-main truncate flex-1">
                                    {signal.message || signal.source || "—"}
                                </span>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
}

// ── Missions ─────────────────────────────────────────────────

function aggregateStatus(agents: { status: number }[]): "error" | "busy" | "idle" | "offline" {
    if (agents.length === 0) return "offline";
    if (agents.some((a) => a.status === 3)) return "error";
    if (agents.some((a) => a.status === 2)) return "busy";
    if (agents.some((a) => a.status === 1)) return "idle";
    return "offline";
}

const STATUS_DOT: Record<string, string> = {
    error: "bg-cortex-danger",
    busy: "bg-cortex-info animate-pulse",
    idle: "bg-cortex-success",
    offline: "bg-cortex-text-muted/40",
};

function MissionsSection() {
    const missions = useCortexStore((s) => s.missions);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);

    useEffect(() => {
        fetchMissions();
        fetchTeamsDetail();
        const mInterval = setInterval(fetchMissions, 15000);
        const tInterval = setInterval(fetchTeamsDetail, 10000);
        return () => {
            clearInterval(mInterval);
            clearInterval(tInterval);
        };
    }, [fetchMissions, fetchTeamsDetail]);

    const missionTeams = teamsDetail.filter((t) => t.type === "mission");
    const missionTeamMap = new Map<string, TeamDetailEntry[]>();
    for (const t of missionTeams) {
        if (t.mission_id) {
            const existing = missionTeamMap.get(t.mission_id) ?? [];
            existing.push(t);
            missionTeamMap.set(t.mission_id, existing);
        }
    }

    const sortedMissions = [...missions].sort((a, b) => {
        const order: Record<string, number> = { active: 0, completed: 1, failed: 2 };
        return (order[a.status] ?? 1) - (order[b.status] ?? 1);
    });

    const STATUS_BADGE: Record<string, string> = {
        active: "bg-cortex-success/10 text-cortex-success",
        completed: "bg-cortex-text-muted/10 text-cortex-text-muted",
        failed: "bg-cortex-danger/10 text-cortex-danger",
    };

    return (
        <SectionCard title="Missions" subtitle="Active agent swarms executing user-defined objectives" icon={Radar} href="/automations?tab=active">
            {sortedMissions.length === 0 ? (
                <div className="px-4 py-5 text-center">
                    <p className="text-sm font-mono text-cortex-text-muted">No active missions</p>
                    <p className="text-xs font-mono text-cortex-text-muted/50 mt-1">Launch a crew from Workspace, or ask Soma</p>
                </div>
            ) : (
                <div className="max-h-64 overflow-y-auto">
                    {sortedMissions.map((mission) => {
                        const teams = missionTeamMap.get(mission.id) ?? [];
                        const allAgents = teams.flatMap((t) => t.agents);
                        const status = allAgents.length > 0 ? aggregateStatus(allAgents) : "offline";

                        return (
                            <div
                                key={mission.id}
                                className="w-full px-4 py-2.5 flex items-center gap-3 text-left"
                            >
                                <span className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${STATUS_DOT[status]}`} />
                                <span className="text-[13px] font-mono font-bold text-cortex-text-main truncate flex-1">
                                    {mission.id}
                                </span>
                                <span className={`text-xs font-mono font-bold uppercase px-2 py-0.5 rounded ${STATUS_BADGE[mission.status] ?? STATUS_BADGE.active}`}>
                                    {mission.status}
                                </span>
                                <span className="text-xs font-mono text-cortex-text-muted">
                                    {mission.teams}T / {mission.agents}A
                                </span>
                            </div>
                        );
                    })}
                </div>
            )}
        </SectionCard>
    );
}

// ── Standing Teams ───────────────────────────────────────────

function TeamsSection() {
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const standingTeams = teamsDetail.filter((t) => t.type === "standing");

    if (standingTeams.length === 0) return null;

    return (
        <SectionCard title="Teams" subtitle="Standing teams always online — Soma, Council, and more" icon={Users} href="/automations?tab=teams">
            <div className="max-h-44 overflow-y-auto">
                {standingTeams.map((team) => {
                    const status = aggregateStatus(team.agents);
                    const onlineCount = team.agents.filter((a) => a.status >= 1 && a.status <= 2).length;

                    return (
                        <div
                            key={team.id}
                            className="px-3.5 py-2 flex items-center gap-2.5"
                        >
                            <span className={`w-2 h-2 rounded-full flex-shrink-0 ${STATUS_DOT[status]}`} />
                            <span className="text-[13px] font-mono font-bold text-cortex-text-main truncate flex-1">
                                {team.name}
                            </span>
                            <span className="text-xs font-mono text-cortex-text-muted">
                                {onlineCount}/{team.agents.length}
                            </span>
                        </div>
                    );
                })}
            </div>
        </SectionCard>
    );
}

// ── MCP Tools ────────────────────────────────────────────────

const RECOMMENDED_SERVERS = ["brave-search", "github"];

function MCPToolsSection() {
    const router = useRouter();
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);

    useEffect(() => {
        fetchMCPServers();
    }, [fetchMCPServers]);

    const connectedCount = mcpServers.filter((s) => s.status === "connected").length;
    const installedNames = new Set(mcpServers.map((s) => s.name));
    const missingRecommended = RECOMMENDED_SERVERS.filter((n) => !installedNames.has(n));

    return (
        <SectionCard title="MCP Tools" subtitle="External tool servers agents can use — files, web, APIs" icon={Wrench} href="/settings/tools">
            <div className="px-3.5 py-2.5 space-y-2">
                {mcpServers.length === 0 ? (
                    <div>
                        <p className="text-xs font-mono text-cortex-text-muted">
                            No tools installed
                        </p>
                        <p className="text-xs font-mono text-cortex-text-muted/50 mt-1">
                            Start the backend to auto-install filesystem + fetch
                        </p>
                    </div>
                ) : (
                    <>
                        {mcpServers.map((srv) => (
                            <div key={srv.id} className="flex items-center gap-2.5">
                                <Circle
                                    className={`w-2 h-2 flex-shrink-0 ${
                                        srv.status === "connected"
                                            ? "fill-cortex-success text-cortex-success"
                                            : "fill-cortex-danger text-cortex-danger"
                                    }`}
                                />
                                <span className="text-[13px] font-mono text-cortex-text-main truncate flex-1">
                                    {srv.name}
                                </span>
                                {srv.tools && (
                                    <span className="text-xs font-mono text-cortex-text-muted">
                                        {srv.tools.length}t
                                    </span>
                                )}
                            </div>
                        ))}
                        <div className="text-xs font-mono text-cortex-text-muted pt-0.5">
                            {connectedCount}/{mcpServers.length} connected
                        </div>
                    </>
                )}

                {missingRecommended.length > 0 && (
                    <div className="mt-2 pt-2 border-t border-cortex-border/50">
                        <p className="text-xs font-mono font-bold uppercase tracking-widest text-cortex-text-muted mb-1.5">
                            Recommended
                        </p>
                        {missingRecommended.map((name) => (
                            <button
                                key={name}
                                onClick={() => router.push("/settings/tools")}
                                className="w-full flex items-center gap-2 py-1 text-left hover:text-cortex-primary transition-colors"
                            >
                                <Plus className="w-3.5 h-3.5 text-cortex-text-muted" />
                                <span className="text-xs font-mono text-cortex-text-muted hover:text-cortex-primary">
                                    {name}
                                </span>
                            </button>
                        ))}
                    </div>
                )}
            </div>
        </SectionCard>
    );
}

// ── Recent Runs ───────────────────────────────────────────────

function runStatusDot(status: MissionRun['status']): string {
    switch (status) {
        case 'running':   return 'bg-cortex-primary animate-pulse';
        case 'completed': return 'bg-cortex-success';
        case 'failed':    return 'bg-cortex-danger';
        default:          return 'bg-cortex-text-muted/40';
    }
}

function runStatusLabel(status: MissionRun['status']): string {
    switch (status) {
        case 'running':   return 'text-cortex-primary';
        case 'completed': return 'text-cortex-success';
        case 'failed':    return 'text-cortex-danger';
        default:          return 'text-cortex-text-muted';
    }
}

function timeAgo(iso: string): string {
    const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
    if (diff < 60)   return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    return `${Math.floor(diff / 3600)}h ago`;
}

function RecentRunsSection() {
    const router = useRouter();
    const recentRuns = useCortexStore((s) => s.recentRuns);
    const fetchRecentRuns = useCortexStore((s) => s.fetchRecentRuns);

    useEffect(() => {
        fetchRecentRuns();
        const interval = setInterval(fetchRecentRuns, 10000);
        return () => clearInterval(interval);
    }, [fetchRecentRuns]);

    const active = recentRuns.filter((r) => r.status === 'running');

    return (
        <SectionCard
            title="Runs"
            subtitle={active.length > 0 ? `${active.length} active` : "Mission execution history"}
            icon={Activity}
            href="/runs"
        >
            {recentRuns.length === 0 ? (
                <div className="px-4 py-5 text-center">
                    <p className="text-sm font-mono text-cortex-text-muted">No runs yet</p>
                    <p className="text-xs font-mono text-cortex-text-muted/50 mt-1">
                        Ask Soma to launch a crew to start a run
                    </p>
                </div>
            ) : (
                <div className="max-h-48 overflow-y-auto">
                    {recentRuns.slice(0, 10).map((run) => (
                        <button
                            key={run.id}
                            onClick={() => router.push(`/runs/${run.id}`)}
                            className="w-full px-4 py-2 flex items-center gap-3 text-left hover:bg-cortex-border/30 transition-colors group"
                        >
                            <span className={`w-2 h-2 rounded-full flex-shrink-0 ${runStatusDot(run.status)}`} />
                            <span className="text-[11px] font-mono text-cortex-text-main group-hover:text-cortex-primary transition-colors flex-1 truncate">
                                {run.id.slice(0, 8)}...
                            </span>
                            <span className={`text-[9px] font-mono font-bold uppercase ${runStatusLabel(run.status)}`}>
                                {run.status}
                            </span>
                            <span className="text-[9px] font-mono text-cortex-text-muted/60 flex-shrink-0">
                                {timeAgo(run.started_at)}
                            </span>
                            <Zap className="w-2.5 h-2.5 text-cortex-text-muted/40 group-hover:text-cortex-primary transition-colors flex-shrink-0" />
                        </button>
                    ))}
                </div>
            )}
        </SectionCard>
    );
}

// ── Widget Registration ───────────────────────────────────────
//
// Built-in widgets are registered here using multiples of 10 so
// third-party widgets can slot between them (e.g. order: 15).
// To add a new widget: create a component above, call registerOpsWidget().
// OpsOverview renders all registered widgets automatically.

registerOpsWidget({ id: "system",    order: 10, layout: "grid",      Component: SystemStatus });
registerOpsWidget({ id: "alerts",    order: 20, layout: "grid",      Component: PriorityAlerts });
registerOpsWidget({ id: "teams",     order: 30, layout: "grid",      Component: TeamsSection });
registerOpsWidget({ id: "mcp",       order: 40, layout: "grid",      Component: MCPToolsSection });
registerOpsWidget({ id: "missions",  order: 50, layout: "fullWidth", Component: MissionsSection });
registerOpsWidget({ id: "runs",      order: 60, layout: "fullWidth", Component: RecentRunsSection });

// ── Main Component ───────────────────────────────────────────

export default function OpsOverview() {
    const gridWidgets = getOpsWidgets().filter((w) => w.layout === "grid");
    const fullWidthWidgets = getOpsWidgets().filter((w) => w.layout === "fullWidth");

    return (
        <div className="h-full overflow-hidden bg-cortex-surface flex flex-col">
            <div className="flex-1 overflow-y-auto p-3 scrollbar-thin scrollbar-thumb-cortex-border space-y-3">
                {/* Top row: compact status cards in auto-fit grid */}
                <div className="grid grid-cols-[repeat(auto-fit,minmax(240px,1fr))] gap-3">
                    {gridWidgets.map(({ id, Component }) => (
                        <Component key={id} />
                    ))}
                </div>
                {/* Full-width sections: Missions + Runs and any future additions */}
                {fullWidthWidgets.map(({ id, Component }) => (
                    <Component key={id} />
                ))}
            </div>
        </div>
    );
}

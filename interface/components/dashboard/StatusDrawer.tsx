"use client";

import React, { useEffect, useMemo } from "react";
import Link from "next/link";
import { X, Brain, Shield, Wifi, Database, Radio, Users, AlertTriangle, RefreshCw } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import {
    CORE_RECOVERY_HINT,
    CORE_UNREACHABLE_SUMMARY,
    CORE_UNREACHABLE_TITLE,
    NOT_CHECKED_CORE_UNREACHABLE,
    selectCoreReachability,
} from "@/lib/coreReachability";

type Health = "healthy" | "degraded" | "failure" | "offline" | "info";

const HEALTH_CLASS: Record<Health, string> = {
    healthy: "text-cortex-success border-cortex-success/30 bg-cortex-success/10",
    degraded: "text-cortex-warning border-cortex-warning/30 bg-cortex-warning/10",
    failure: "text-cortex-danger border-cortex-danger/30 bg-cortex-danger/10",
    offline: "text-cortex-text-muted border-cortex-border bg-cortex-surface/40",
    info: "text-cortex-primary border-cortex-primary/30 bg-cortex-primary/10",
};

function statusFromService(status?: "online" | "offline" | "degraded"): Health {
    if (!status) return "offline";
    if (status === "online") return "healthy";
    if (status === "degraded") return "degraded";
    return "failure";
}

export default function StatusDrawer() {
    const isOpen = useCortexStore((s) => s.isStatusDrawerOpen);
    const setOpen = useCortexStore((s) => s.setStatusDrawerOpen);
    const councilMembers = useCortexStore((s) => s.councilMembers);
    const fetchCouncilMembers = useCortexStore((s) => s.fetchCouncilMembers);
    const missionChat = useCortexStore((s) => s.missionChat);
    const missionChatFailure = useCortexStore((s) => s.missionChatFailure);
    const isStreamConnected = useCortexStore((s) => s.isStreamConnected);
    const streamConnectionState = useCortexStore((s) => s.streamConnectionState);
    const activeBrain = useCortexStore((s) => s.activeBrain);
    const governanceMode = useCortexStore((s) => s.governanceMode);
    const missions = useCortexStore((s) => s.missions);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);
    const services = useCortexStore((s) => s.servicesStatus);
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);
    const isFetchingServices = useCortexStore((s) => s.isFetchingServicesStatus);
    const reachability = useCortexStore(selectCoreReachability);
    const coreDown = reachability === "unreachable";

    useEffect(() => {
        if (!isOpen) return;
        fetchCouncilMembers();
        fetchMissions();
        fetchServicesStatus();
    }, [isOpen, fetchCouncilMembers, fetchMissions, fetchServicesStatus]);

    const serviceMap = useMemo(() => {
        const map = new Map<string, (typeof services)[number]>();
        services.forEach((s) => map.set(s.name, s));
        return map;
    }, [services]);

    const unchecked = reachability === "unknown" ? "checking…" : "unknown";
    const svcValue = (name: string, preferDetail = false) => {
        if (coreDown) return NOT_CHECKED_CORE_UNREACHABLE;
        const svc = serviceMap.get(name);
        return (preferDetail ? svc?.detail ?? svc?.status : svc?.status) ?? unchecked;
    };
    const svcHealth = (name: string): Health => (coreDown ? "offline" : statusFromService(serviceMap.get(name)?.status));

    const sseHealth: Health = isStreamConnected
        ? "healthy"
        : streamConnectionState === "connecting"
          ? "info"
          : "failure";
    const sseLabel = isStreamConnected
        ? "live"
        : streamConnectionState === "connecting"
          ? "connecting"
          : "offline";
    const govHealth: Health = governanceMode === "strict" ? "failure" : governanceMode === "active" ? "degraded" : "healthy";

    const activeMissions = missions.filter((m) => m.status === "active").length;
    const ollamaSvc = serviceMap.get("ollama");
    const brainValue = activeBrain
        ? `${activeBrain.provider_name ?? activeBrain.provider_id} (${activeBrain.location})`
        : !coreDown && ollamaSvc?.status === "online"
          ? ollamaSvc.detail || "ollama (online)"
          : "none";
    const brainHealth: Health = activeBrain
        ? activeBrain.location === "remote"
            ? "degraded"
            : "healthy"
        : !coreDown && ollamaSvc?.status === "online"
          ? "healthy"
          : "offline";

    const councilHealth = (memberId: string): Health => {
        if (missionChatFailure && missionChatFailure.targetId === memberId) return "failure";
        const lastMsg = [...missionChat]
            .reverse()
            .find((m) => m.source_node === memberId && m.role === "council");
        if (!lastMsg) return "info";
        const txt = lastMsg.content.toLowerCase();
        if (txt.includes("unreachable") || txt.includes("timeout") || txt.includes("error")) return "degraded";
        return "healthy";
    };

    if (!isOpen) return null;

    return (
        <>
            <button
                className="fixed inset-0 bg-black/40 z-[60]"
                onClick={() => setOpen(false)}
                aria-label="Close status drawer backdrop"
            />
            <aside
                role="dialog"
                aria-label="System status drawer"
                className="fixed right-0 top-0 bottom-0 w-[380px] bg-cortex-surface border-l border-cortex-border z-[61] flex flex-col"
            >
                <div className="h-12 px-4 border-b border-cortex-border flex items-center justify-between">
                    <div className="text-sm font-mono font-bold text-cortex-text-main">System Status</div>
                    <button
                        onClick={() => setOpen(false)}
                        aria-label="Close status drawer"
                        className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted"
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>
                <div className="flex-1 overflow-y-auto p-4 space-y-4">
                    {coreDown && (
                        <div role="alert" data-testid="status-drawer-root-cause" className={`rounded-lg border px-3 py-3 text-xs ${HEALTH_CLASS.failure}`}>
                            <div className="flex items-center gap-2 font-mono font-bold uppercase tracking-wide">
                                <AlertTriangle className="w-3.5 h-3.5" />
                                {CORE_UNREACHABLE_TITLE}
                            </div>
                            <p className="mt-1.5 text-cortex-text-main">{CORE_UNREACHABLE_SUMMARY}</p>
                            <p className="mt-1.5 text-cortex-text-muted">{CORE_RECOVERY_HINT}</p>
                            <button
                                type="button"
                                onClick={() => void fetchServicesStatus()}
                                className="mt-2 inline-flex items-center gap-1 rounded border border-cortex-danger/40 px-2 py-1 font-mono hover:bg-cortex-danger/15"
                            >
                                <RefreshCw className={`w-3 h-3 ${isFetchingServices ? "animate-spin" : ""}`} />
                                Retry
                            </button>
                        </div>
                    )}
                    <StatusRow
                        icon={Radio}
                        label="Council Reachability"
                        value={coreDown ? NOT_CHECKED_CORE_UNREACHABLE : `${councilMembers.length} members`}
                        health={coreDown ? "offline" : councilMembers.length > 0 ? "healthy" : "degraded"}
                    />
                    <div className="space-y-2 pl-7">
                        {coreDown ? null : councilMembers.length === 0 ? (
                            <div className={`rounded-md border px-2 py-1.5 text-xs font-mono ${HEALTH_CLASS.degraded}`}>
                                No council members reachable
                            </div>
                        ) : (
                            councilMembers.map((m) => (
                                <div key={m.id} className={`rounded-md border px-2 py-1.5 text-xs font-mono ${HEALTH_CLASS[councilHealth(m.id)]}`}>
                                    {m.id}
                                </div>
                            ))
                        )}
                    </div>
                    <StatusRow
                        icon={Wifi}
                        label="NATS Connection"
                        value={svcValue("nats")}
                        subvalue={serviceMap.get("nats")?.address}
                        health={svcHealth("nats")}
                    />
                    <StatusRow icon={Users} label="Group Bus" value={svcValue("groups_bus", true)} health={svcHealth("groups_bus")} />
                    <StatusRow icon={Radio} label="SSE Stream" value={sseLabel} health={sseHealth} />
                    <StatusRow
                        icon={Database}
                        label="Database"
                        value={svcValue("postgres")}
                        subvalue={serviceMap.get("postgres")?.address}
                        health={svcHealth("postgres")}
                    />
                    <StatusRow
                        icon={Brain}
                        label="Active Brain"
                        value={brainValue}
                        subvalue={serviceMap.get("ollama")?.address}
                        health={brainHealth}
                    />
                    <StatusRow
                        icon={Shield}
                        label="Governance"
                        value={coreDown ? NOT_CHECKED_CORE_UNREACHABLE : governanceMode}
                        health={coreDown ? "offline" : govHealth}
                    />
                    <StatusRow
                        icon={Users}
                        label="Active Missions"
                        value={coreDown ? NOT_CHECKED_CORE_UNREACHABLE : String(activeMissions)}
                        health={coreDown ? "offline" : activeMissions > 0 ? "info" : "offline"}
                    />
                    {missionChatFailure && (
                        <div className="rounded-md border border-cortex-danger/40 bg-cortex-danger/10 px-3 py-2 text-xs font-mono text-cortex-danger flex items-start gap-2">
                            <AlertTriangle className="w-3.5 h-3.5 mt-0.5" />
                            <div>Last workspace failure: {missionChatFailure.targetLabel} ({missionChatFailure.type})</div>
                        </div>
                    )}
                    <div className="pt-2 border-t border-cortex-border space-y-2" data-testid="status-drawer-quick-actions">
                        <div className="text-[10px] uppercase font-mono tracking-wide text-cortex-text-muted">Quick Management</div>
                        <div className="grid grid-cols-2 gap-1.5">
                            {[
                                { href: "/system", label: "Service Book & Health" },
                                { href: "/deployments", label: "Active Teams & Config" },
                                { href: "/resources", label: "Deliverables & Tools" },
                                { href: "/settings", label: "AI Engines & Access" },
                            ].map((item) => (
                                <Link
                                    key={item.href}
                                    href={item.href}
                                    onClick={() => setOpen(false)}
                                    className="rounded-md border border-cortex-border bg-cortex-bg/60 px-2.5 py-1.5 text-[11px] font-mono text-cortex-text-main hover:border-cortex-primary/40 hover:text-cortex-primary transition-colors truncate"
                                >
                                    {item.label}
                                </Link>
                            ))}
                        </div>
                    </div>
                </div>
            </aside>
        </>
    );
}

function StatusRow({
    icon: Icon,
    label,
    value,
    subvalue,
    health,
}: {
    icon: React.ElementType;
    label: string;
    value: string;
    subvalue?: string;
    health: Health;
}) {
    return (
        <div className={`rounded-lg border px-3 py-2 ${HEALTH_CLASS[health]}`}>
            <div className="flex items-center gap-2">
                <Icon className="w-3.5 h-3.5" />
                <span className="text-[10px] uppercase font-mono tracking-wide">{label}</span>
            </div>
            <div className="text-xs font-mono mt-1">{value}</div>
            {subvalue ? <div className="text-[10px] font-mono text-cortex-text-muted mt-0.5 truncate" title={subvalue}>{subvalue}</div> : null}
        </div>
    );
}

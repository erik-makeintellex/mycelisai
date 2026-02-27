"use client";

import React, { useEffect, useMemo } from "react";
import { X, Brain, Shield, Wifi, Database, Radio, Users, AlertTriangle } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

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
    const missionChatError = useCortexStore((s) => s.missionChatError);
    const councilTarget = useCortexStore((s) => s.councilTarget);
    const isStreamConnected = useCortexStore((s) => s.isStreamConnected);
    const activeBrain = useCortexStore((s) => s.activeBrain);
    const governanceMode = useCortexStore((s) => s.governanceMode);
    const missions = useCortexStore((s) => s.missions);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);
    const services = useCortexStore((s) => s.servicesStatus);
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);

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

    const natsHealth = statusFromService(serviceMap.get("nats")?.status);
    const dbHealth = statusFromService(serviceMap.get("postgres")?.status);
    const sseHealth: Health = isStreamConnected ? "healthy" : "failure";
    const govHealth: Health = governanceMode === "strict" ? "failure" : governanceMode === "active" ? "degraded" : "healthy";

    const activeMissions = missions.filter((m) => m.status === "active").length;
    const failingMember = useMemo(() => {
        if (!missionChatError) return null;
        const lastFailure = [...missionChat]
            .reverse()
            .find((m) =>
                m.role === "council" &&
                typeof m.source_node === "string" &&
                typeof m.content === "string" &&
                /(unreachable|timeout|error|failed)/i.test(m.content)
            );
        return lastFailure?.source_node ?? councilTarget;
    }, [missionChatError, missionChat, councilTarget]);

    const councilHealth = (memberId: string): Health => {
        if (failingMember && memberId === failingMember) return "failure";
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
                    <StatusRow icon={Radio} label="Council Reachability" value={`${councilMembers.length} members`} health={councilMembers.length > 0 ? "healthy" : "degraded"} />
                    <div className="space-y-2 pl-7">
                        {councilMembers.length === 0 ? (
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
                    <StatusRow icon={Wifi} label="NATS Connection" value={serviceMap.get("nats")?.status ?? "unknown"} health={natsHealth} />
                    <StatusRow icon={Radio} label="SSE Stream" value={isStreamConnected ? "live" : "offline"} health={sseHealth} />
                    <StatusRow icon={Database} label="Database" value={serviceMap.get("postgres")?.status ?? "unknown"} health={dbHealth} />
                    <StatusRow
                        icon={Brain}
                        label="Active Brain"
                        value={activeBrain ? `${activeBrain.provider_name ?? activeBrain.provider_id} (${activeBrain.location})` : "none"}
                        health={activeBrain ? (activeBrain.location === "remote" ? "degraded" : "healthy") : "offline"}
                    />
                    <StatusRow icon={Shield} label="Governance" value={governanceMode} health={govHealth} />
                    <StatusRow icon={Users} label="Active Missions" value={String(activeMissions)} health={activeMissions > 0 ? "info" : "offline"} />
                    {failingMember && (
                        <div className="rounded-md border border-cortex-danger/40 bg-cortex-danger/10 px-3 py-2 text-xs font-mono text-cortex-danger flex items-start gap-2">
                            <AlertTriangle className="w-3.5 h-3.5 mt-0.5" />
                            <div>Last council failure: {failingMember}</div>
                        </div>
                    )}
                </div>
            </aside>
        </>
    );
}

function StatusRow({
    icon: Icon,
    label,
    value,
    health,
}: {
    icon: React.ElementType;
    label: string;
    value: string;
    health: Health;
}) {
    return (
        <div className={`rounded-lg border px-3 py-2 ${HEALTH_CLASS[health]}`}>
            <div className="flex items-center gap-2">
                <Icon className="w-3.5 h-3.5" />
                <span className="text-[10px] uppercase font-mono tracking-wide">{label}</span>
            </div>
            <div className="text-xs font-mono mt-1">{value}</div>
        </div>
    );
}

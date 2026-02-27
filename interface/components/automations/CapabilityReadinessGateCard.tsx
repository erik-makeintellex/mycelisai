"use client";

import React, { useCallback, useEffect, useMemo } from "react";
import { CheckCircle2, AlertTriangle, XCircle, RefreshCw } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import type { ReadinessSnapshot } from "@/lib/workflowContracts";

type GateStatus = "healthy" | "degraded" | "failure";

function toGateStatus(ready: boolean, degraded = false): GateStatus {
    if (!ready) return "failure";
    if (degraded) return "degraded";
    return "healthy";
}

function statusStyles(status: GateStatus): string {
    if (status === "healthy") return "text-cortex-success border-cortex-success/30 bg-cortex-success/10";
    if (status === "degraded") return "text-cortex-warning border-cortex-warning/30 bg-cortex-warning/10";
    return "text-cortex-danger border-cortex-danger/30 bg-cortex-danger/10";
}

function statusIcon(status: GateStatus) {
    if (status === "healthy") return <CheckCircle2 className="w-3.5 h-3.5" />;
    if (status === "degraded") return <AlertTriangle className="w-3.5 h-3.5" />;
    return <XCircle className="w-3.5 h-3.5" />;
}

interface CapabilityReadinessGateCardProps {
    onSnapshotChange?: (snapshot: ReadinessSnapshot) => void;
}

export default function CapabilityReadinessGateCard({ onSnapshotChange }: CapabilityReadinessGateCardProps) {
    const servicesStatus = useCortexStore((s) => s.servicesStatus);
    const isFetchingServicesStatus = useCortexStore((s) => s.isFetchingServicesStatus);
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);
    const isStreamConnected = useCortexStore((s) => s.isStreamConnected);
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);
    const governanceMode = useCortexStore((s) => s.governanceMode);
    const missionProfiles = useCortexStore((s) => s.missionProfiles);
    const activeBrain = useCortexStore((s) => s.activeBrain);

    const serviceMap = useMemo(() => new Map(servicesStatus.map((svc) => [svc.name, svc.status])), [servicesStatus]);

    const snapshot = useMemo<ReadinessSnapshot>(() => {
        const natsReady = serviceMap.get("nats") === "online";
        const dbReady = serviceMap.get("postgres") === "online";
        const sseReady = isStreamConnected;
        const providerReady = Boolean(activeBrain) || missionProfiles.length > 0;
        const mcpReady = mcpServers.some((srv) => srv.status === "connected" || srv.status === "installed");
        const governanceReady = governanceMode !== "strict";
        const blockers: string[] = [];
        if (!providerReady) blockers.push("No provider profile or active brain detected");
        if (!mcpReady) blockers.push("No MCP server is connected");
        if (!natsReady) blockers.push("NATS transport is unavailable");
        if (!sseReady) blockers.push("SSE stream is disconnected");
        if (!dbReady) blockers.push("Database status is offline");
        return {
            providerReady,
            mcpReady,
            governanceReady,
            natsReady,
            sseReady,
            dbReady,
            blockers,
        };
    }, [serviceMap, isStreamConnected, activeBrain, missionProfiles, mcpServers, governanceMode]);

    useEffect(() => {
        onSnapshotChange?.(snapshot);
    }, [onSnapshotChange, snapshot]);

    useEffect(() => {
        fetchServicesStatus();
        fetchMCPServers();
    }, [fetchServicesStatus, fetchMCPServers]);

    const refresh = useCallback(async () => {
        await Promise.all([fetchServicesStatus(), fetchMCPServers()]);
    }, [fetchServicesStatus, fetchMCPServers]);

    const checks = [
        { label: "Provider profile", status: toGateStatus(snapshot.providerReady) },
        { label: "MCP capability", status: toGateStatus(snapshot.mcpReady) },
        { label: "NATS bus", status: toGateStatus(snapshot.natsReady) },
        { label: "SSE stream", status: toGateStatus(snapshot.sseReady) },
        { label: "Database", status: toGateStatus(snapshot.dbReady) },
        { label: "Governance mode", status: toGateStatus(snapshot.governanceReady, governanceMode === "active") },
    ] as const;

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4 space-y-3">
            <div className="flex items-center justify-between gap-2">
                <div>
                    <h3 className="text-sm font-semibold text-cortex-text-main">Capability Readiness Gate</h3>
                    <p className="text-[11px] text-cortex-text-muted mt-1">
                        Validate provider, bus, stream, and tool access before launch.
                    </p>
                </div>
                <button
                    onClick={refresh}
                    className="px-2 py-1 rounded border border-cortex-border text-[10px] font-mono hover:bg-cortex-border flex items-center gap-1"
                >
                    <RefreshCw className={`w-3 h-3 ${isFetchingServicesStatus ? "animate-spin" : ""}`} />
                    Refresh
                </button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                {checks.map((check) => (
                    <div key={check.label} className={`rounded-md border px-2.5 py-2 text-[11px] font-mono flex items-center gap-2 ${statusStyles(check.status)}`}>
                        {statusIcon(check.status)}
                        {check.label}
                    </div>
                ))}
            </div>
            {snapshot.blockers.length > 0 ? (
                <div className="rounded-md border border-cortex-warning/30 bg-cortex-warning/10 p-2.5">
                    <p className="text-[11px] font-semibold text-cortex-warning mb-1">Launch blockers</p>
                    <ul className="text-[11px] text-cortex-text-main space-y-0.5 list-disc list-inside">
                        {snapshot.blockers.map((b) => <li key={b}>{b}</li>)}
                    </ul>
                </div>
            ) : (
                <div className="rounded-md border border-cortex-success/30 bg-cortex-success/10 p-2.5">
                    <p className="text-[11px] font-semibold text-cortex-success">Ready to launch</p>
                    <p className="text-[11px] text-cortex-text-main mt-0.5">No blocking capability gaps detected.</p>
                </div>
            )}
        </div>
    );
}

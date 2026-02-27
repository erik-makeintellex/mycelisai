"use client";

import React, { useMemo } from "react";
import { AlertTriangle, RefreshCw } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

export default function DegradedModeBanner() {
    const services = useCortexStore((s) => s.servicesStatus);
    const loading = useCortexStore((s) => s.isFetchingServicesStatus);
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);
    const missionChatError = useCortexStore((s) => s.missionChatError);
    const isStreamConnected = useCortexStore((s) => s.isStreamConnected);
    const setCouncilTarget = useCortexStore((s) => s.setCouncilTarget);
    const fetchCouncilMembers = useCortexStore((s) => s.fetchCouncilMembers);
    const setStatusDrawerOpen = useCortexStore((s) => s.setStatusDrawerOpen);
    const initializeStream = useCortexStore((s) => s.initializeStream);
    const fetchMissions = useCortexStore((s) => s.fetchMissions);

    const retryAll = async () => {
        await fetchServicesStatus();
        fetchCouncilMembers();
        fetchMissions();
        if (!isStreamConnected) {
            initializeStream();
        }
    };

    const reasons = useMemo(() => {
        const r: string[] = [];
        const m = new Map(services.map((s) => [s.name, s.status]));
        if (m.get("nats") && m.get("nats") !== "online") r.push(`NATS ${m.get("nats")}`);
        if (m.get("postgres") && m.get("postgres") !== "online") r.push(`Database ${m.get("postgres")}`);
        if (!isStreamConnected) r.push("SSE stream offline");
        if (missionChatError) r.push("Council call failure");
        return r;
    }, [services, isStreamConnected, missionChatError]);

    const degraded = reasons.length > 0;
    if (!degraded) return null;

    return (
        <div className="h-9 border-b border-cortex-warning/30 bg-cortex-warning/10 px-4 flex items-center justify-between gap-3 flex-shrink-0">
            <div className="flex items-center gap-2 min-w-0">
                <AlertTriangle className="w-3.5 h-3.5 text-cortex-warning flex-shrink-0" />
                <p className="text-[11px] font-mono text-cortex-warning truncate">
                    System in Degraded Mode — {reasons[0]}.
                </p>
            </div>
            <div className="flex items-center gap-1.5 flex-shrink-0">
                <button
                    onClick={retryAll}
                    className="px-2 py-1 rounded border border-cortex-warning/30 text-cortex-warning text-[10px] font-mono hover:bg-cortex-warning/20 transition-colors flex items-center gap-1"
                >
                    <RefreshCw className={`w-3 h-3 ${loading ? "animate-spin" : ""}`} />
                    Retry
                </button>
                <button
                    onClick={() => {
                        setCouncilTarget("admin");
                        fetchCouncilMembers();
                    }}
                    className="px-2 py-1 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/15 transition-colors"
                >
                    Switch to Soma
                </button>
                <button
                    onClick={() => setStatusDrawerOpen(true)}
                    className="px-2 py-1 rounded border border-cortex-border text-cortex-text-main text-[10px] font-mono hover:bg-cortex-border transition-colors"
                >
                    Open Status
                </button>
            </div>
        </div>
    );
}

"use client";

import React, { useEffect } from "react";
import { AlertTriangle, CheckCircle2, Loader2, RefreshCw, ShieldCheck, Sliders, XCircle } from "lucide-react";
import { useCortexStore, type ServiceHealthStatus } from "@/store/useCortexStore";

const HUMAN_SERVICE_LABELS: Record<string, string> = {
    nats: "Event Bus & Team Coordination",
    postgres: "Workspace Memory & Storage",
    cognitive: "AI Reasoning Engine",
    reactive: "Automation & Trigger Engine",
    groups_bus: "Group Collaboration Bus",
};

export default function SystemStatusSummaryCard() {
    const services = useCortexStore((s) => s.servicesStatus) ?? [];
    const loading = useCortexStore((s) => s.isFetchingServicesStatus) ?? false;
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);
    const assistantName = useCortexStore((s) => s.assistantName) || "Soma";

    useEffect(() => {
        fetchServicesStatus?.();
    }, [fetchServicesStatus]);

    const total = services.length;
    const online = services.filter((s: ServiceHealthStatus) => s.status === "online").length;
    const allOnline = total > 0 && online === total;
    const anyOnline = online > 0;

    const headline = loading && total === 0
        ? "System Status: Checking active services..."
        : allOnline
        ? `System Status: Online (${online}/${total} services active)`
        : anyOnline
        ? `System Status: Partially Active (${online}/${total} services active)`
        : "System Status: Offline — Core services are currently unreachable";

    const summary = allOnline
        ? `${assistantName}, standing teams, durable memory, and governance checks are online and ready for work.`
        : anyOnline
        ? `Some workspace capabilities are active (${online} of ${total} services online), while degraded services may pause background automations.`
        : `The workspace cannot reach backend services right now. Open Advanced Diagnostics for recovery steps or check your local runtime.`;

    return (
        <div
            className="h-full overflow-y-auto bg-cortex-bg px-6 py-8 flex items-center justify-center"
            data-testid="system-status-summary-card"
        >
            <div className="max-w-2xl w-full rounded-2xl border border-cortex-border bg-cortex-surface p-6 space-y-5 shadow-sm">
                <div className="flex items-start justify-between gap-4">
                    <div className="space-y-2">
                        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-wider text-cortex-primary">
                            <ShieldCheck className="w-3.5 h-3.5" />
                            System Readiness Summary
                        </div>
                        <div className="flex items-center gap-2.5">
                            {loading && total === 0 ? (
                                <Loader2 className="w-5 h-5 text-cortex-text-muted animate-spin flex-shrink-0" />
                            ) : allOnline ? (
                                <span className="w-3 h-3 rounded-full bg-cortex-success animate-pulse flex-shrink-0" />
                            ) : anyOnline ? (
                                <span className="w-3 h-3 rounded-full bg-cortex-warning animate-pulse flex-shrink-0" />
                            ) : (
                                <span className="w-3 h-3 rounded-full bg-cortex-danger flex-shrink-0" />
                            )}
                            <h1 className="text-xl font-bold text-cortex-text-main tracking-tight">
                                {headline}
                            </h1>
                        </div>
                        <p className="text-sm leading-relaxed text-cortex-text-muted">
                            {summary}
                        </p>
                    </div>

                    <button
                        type="button"
                        onClick={() => fetchServicesStatus()}
                        className="p-2 rounded-lg border border-cortex-border bg-cortex-bg text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        title="Refresh system status"
                    >
                        <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                    </button>
                </div>

                {total > 0 && (
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 pt-2 border-t border-cortex-border">
                        {services.map((svc: ServiceHealthStatus) => {
                            const Icon =
                                svc.status === "online"
                                    ? CheckCircle2
                                    : svc.status === "degraded"
                                    ? AlertTriangle
                                    : XCircle;
                            const tone =
                                svc.status === "online"
                                    ? "text-cortex-success border-cortex-success/25 bg-cortex-success/5"
                                    : svc.status === "degraded"
                                    ? "text-cortex-warning border-cortex-warning/25 bg-cortex-warning/5"
                                    : "text-cortex-danger border-cortex-danger/25 bg-cortex-danger/5";
                            return (
                                <div
                                    key={svc.name}
                                    className={`flex items-center justify-between rounded-lg border px-3 py-2 text-xs ${tone}`}
                                >
                                    <span className="font-medium text-cortex-text-main">
                                        {HUMAN_SERVICE_LABELS[svc.name] ?? svc.name}
                                    </span>
                                    <span className="inline-flex items-center gap-1 font-mono uppercase text-[10px]">
                                        <Icon className="w-3.5 h-3.5" />
                                        {svc.status}
                                    </span>
                                </div>
                            );
                        })}
                    </div>
                )}

                <div className="pt-2 flex flex-wrap items-center justify-between gap-3 border-t border-cortex-border">
                    <span className="text-xs text-cortex-text-muted">
                        Need deep telemetry, event bus checks, or service restart commands?
                    </span>
                    <button
                        type="button"
                        onClick={toggleAdvancedMode}
                        data-testid="open-advanced-diagnostics"
                        className="inline-flex items-center gap-2 rounded-xl border border-cortex-primary/35 bg-cortex-primary/10 px-4 py-2 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                    >
                        <Sliders className="w-3.5 h-3.5" />
                        Open Advanced Diagnostics
                    </button>
                </div>
            </div>
        </div>
    );
}

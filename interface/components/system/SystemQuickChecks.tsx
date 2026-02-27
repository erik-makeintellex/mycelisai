"use client";

import React, { useCallback, useMemo, useState } from "react";
import { CheckCircle2, AlertTriangle, XCircle, RefreshCw, Copy, Check } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

type CheckStatus = "healthy" | "degraded" | "failure" | "unknown";

interface CheckItem {
    id: string;
    label: string;
    status: CheckStatus;
    checkedAt?: Date;
}

function statusBadge(status: CheckStatus): string {
    if (status === "healthy") return "text-cortex-success border-cortex-success/30 bg-cortex-success/10";
    if (status === "degraded") return "text-cortex-warning border-cortex-warning/30 bg-cortex-warning/10";
    if (status === "failure") return "text-cortex-danger border-cortex-danger/30 bg-cortex-danger/10";
    return "text-cortex-text-muted border-cortex-border bg-cortex-surface/50";
}

function statusIcon(status: CheckStatus) {
    if (status === "healthy") return <CheckCircle2 className="w-3.5 h-3.5 text-cortex-success" />;
    if (status === "degraded") return <AlertTriangle className="w-3.5 h-3.5 text-cortex-warning" />;
    if (status === "failure") return <XCircle className="w-3.5 h-3.5 text-cortex-danger" />;
    return <AlertTriangle className="w-3.5 h-3.5 text-cortex-text-muted" />;
}

export default function SystemQuickChecks() {
    const servicesStatus = useCortexStore((s) => s.servicesStatus);
    const isFetchingServicesStatus = useCortexStore((s) => s.isFetchingServicesStatus);
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);
    const isStreamConnected = useCortexStore((s) => s.isStreamConnected);

    const [checkedAt, setCheckedAt] = useState<Record<string, Date | undefined>>({});
    const [busy, setBusy] = useState<string | null>(null);
    const [copied, setCopied] = useState<string | null>(null);

    const statusByService = useMemo(() => {
        return new Map(servicesStatus.map((svc) => [svc.name, svc.status]));
    }, [servicesStatus]);

    const checks = useMemo<CheckItem[]>(() => {
        const mapStatus = (name: string): CheckStatus => {
            const state = statusByService.get(name);
            if (state === "online") return "healthy";
            if (state === "degraded") return "degraded";
            if (state === "offline") return "failure";
            return "unknown";
        };
        return [
            { id: "nats", label: "NATS connected", status: mapStatus("nats"), checkedAt: checkedAt.nats },
            { id: "postgres", label: "Database reachable", status: mapStatus("postgres"), checkedAt: checkedAt.postgres },
            { id: "sse", label: "SSE stream live", status: isStreamConnected ? "healthy" : "failure", checkedAt: checkedAt.sse },
            { id: "triggers", label: "Trigger engine active", status: mapStatus("reactive"), checkedAt: checkedAt.triggers },
            { id: "scheduler", label: "Scheduler state", status: "degraded", checkedAt: checkedAt.scheduler },
        ];
    }, [statusByService, checkedAt, isStreamConnected]);

    const runCheck = useCallback(async (id: string) => {
        setBusy(id);
        try {
            if (id !== "scheduler") {
                await fetchServicesStatus();
            }
            setCheckedAt((prev) => ({ ...prev, [id]: new Date() }));
        } catch {
            setCheckedAt((prev) => ({ ...prev, [id]: new Date() }));
        } finally {
            setBusy(null);
        }
    }, [fetchServicesStatus]);

    const summary = useMemo(() => {
        const ok = checks.filter((c) => c.status === "healthy").length;
        return `${ok}/${checks.length} checks healthy`;
    }, [checks]);

    const copySnippet = useCallback((check: CheckItem) => {
        const snippet = JSON.stringify(
            {
                check: check.id,
                label: check.label,
                status: check.status,
                checked_at: check.checkedAt?.toISOString() ?? null,
            },
            null,
            2
        );
        navigator.clipboard.writeText(snippet);
        setCopied(check.id);
        setTimeout(() => setCopied((prev) => (prev === check.id ? null : prev)), 1200);
    }, []);

    return (
        <section className="rounded-xl border border-cortex-border bg-cortex-surface p-4 space-y-3">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-cortex-text-main">Quick Checks</h3>
                <span className="text-[10px] font-mono text-cortex-text-muted">{summary}</span>
            </div>
            <div className="space-y-2">
                {checks.map((check) => (
                    <div key={check.id} className={`rounded-md border px-3 py-2 ${statusBadge(check.status)}`}>
                        <div className="flex items-center justify-between gap-2">
                            <div className="flex items-center gap-2">
                                {statusIcon(check.status)}
                                <span className="text-xs font-mono">{check.label}</span>
                            </div>
                            <button
                                onClick={() => runCheck(check.id)}
                                className="px-2 py-1 rounded border border-cortex-border text-[10px] font-mono hover:bg-cortex-border flex items-center gap-1"
                            >
                                <RefreshCw className={`w-3 h-3 ${busy === check.id || isFetchingServicesStatus ? "animate-spin" : ""}`} />
                                Run Check
                            </button>
                            <button
                                onClick={() => copySnippet(check)}
                                className="px-2 py-1 rounded border border-cortex-border text-[10px] font-mono hover:bg-cortex-border flex items-center gap-1"
                            >
                                {copied === check.id ? <Check className="w-3 h-3 text-cortex-success" /> : <Copy className="w-3 h-3" />}
                                Copy
                            </button>
                        </div>
                        <div className="text-[10px] font-mono text-cortex-text-muted mt-1">
                            Last checked: {check.checkedAt ? check.checkedAt.toLocaleTimeString() : "never"}
                        </div>
                    </div>
                ))}
            </div>
        </section>
    );
}

"use client";

import React, { useCallback, useMemo, useState } from "react";
import { CheckCircle2, AlertTriangle, XCircle, RefreshCw, Copy, Check } from "lucide-react";

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
    const [checks, setChecks] = useState<CheckItem[]>([
        { id: "nats", label: "NATS connected", status: "unknown" },
        { id: "postgres", label: "Database reachable", status: "unknown" },
        { id: "sse", label: "SSE stream live", status: "unknown" },
        { id: "triggers", label: "Trigger engine active", status: "unknown" },
        { id: "scheduler", label: "Scheduler state", status: "unknown" },
    ]);
    const [busy, setBusy] = useState<string | null>(null);
    const [copied, setCopied] = useState<string | null>(null);

    const runCheck = useCallback(async (id: string) => {
        setBusy(id);
        try {
            if (id === "sse") {
                const res = await fetch("/api/v1/telemetry/compute");
                const ok = res.ok;
                setChecks((prev) => prev.map((c) => (c.id === id ? { ...c, status: ok ? "healthy" : "failure", checkedAt: new Date() } : c)));
                return;
            }
            const res = await fetch("/api/v1/services/status");
            if (!res.ok) {
                setChecks((prev) => prev.map((c) => (c.id === id ? { ...c, status: "failure", checkedAt: new Date() } : c)));
                return;
            }
            const body = await res.json();
            const data: Array<{ name: string; status: "online" | "offline" | "degraded" }> = body.data ?? [];
            const map = new Map(data.map((d) => [d.name, d.status]));
            let status: CheckStatus = "unknown";
            if (id === "scheduler") {
                status = "degraded";
            } else if (id === "triggers") {
                const s = map.get("reactive");
                status = s === "online" ? "healthy" : s === "degraded" ? "degraded" : "failure";
            } else {
                const s = map.get(id);
                status = s === "online" ? "healthy" : s === "degraded" ? "degraded" : s === "offline" ? "failure" : "unknown";
            }
            setChecks((prev) => prev.map((c) => (c.id === id ? { ...c, status, checkedAt: new Date() } : c)));
        } catch {
            setChecks((prev) => prev.map((c) => (c.id === id ? { ...c, status: "failure", checkedAt: new Date() } : c)));
        } finally {
            setBusy(null);
        }
    }, []);

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
                                <RefreshCw className={`w-3 h-3 ${busy === check.id ? "animate-spin" : ""}`} />
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

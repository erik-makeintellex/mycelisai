"use client";

import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Check, Copy, Loader2, RefreshCw, ShieldCheck, AlertTriangle } from "lucide-react";

interface DeploymentTrustSnapshot {
    deployment_root: string;
    execution_root: string;
    workspace_root: string;
    artifact_root: string;
    current_commit: string;
    image_tag: string;
    chart_version: string;
    deployment_lane: string;
    endpoint_posture: string;
    runtime_health: {
        status: string;
        online: number;
        degraded: number;
        offline: number;
        total: number;
    };
    proof_lane: string;
    recovery_posture: string;
    checked_at: string;
}

interface DeploymentTrustResponse {
    ok?: boolean;
    data?: DeploymentTrustSnapshot;
}

const TRUST_ROWS: Array<{ key: keyof DeploymentTrustSnapshot; label: string }> = [
    { key: "deployment_root", label: "Deployment Root" },
    { key: "execution_root", label: "Execution Root" },
    { key: "workspace_root", label: "Workspace Root" },
    { key: "artifact_root", label: "Artifact Root" },
    { key: "current_commit", label: "Current Commit" },
    { key: "image_tag", label: "Image Tag" },
    { key: "chart_version", label: "Chart Version" },
    { key: "deployment_lane", label: "Deployment Lane" },
    { key: "endpoint_posture", label: "Endpoint Posture" },
    { key: "proof_lane", label: "Proof Lane" },
    { key: "recovery_posture", label: "Recovery Posture" },
];

function statusClass(status: string): string {
    if (status === "online") return "text-cortex-success border-cortex-success/30 bg-cortex-success/10";
    if (status === "degraded") return "text-cortex-warning border-cortex-warning/30 bg-cortex-warning/10";
    if (status === "offline") return "text-cortex-danger border-cortex-danger/30 bg-cortex-danger/10";
    return "text-cortex-text-muted border-cortex-border bg-cortex-surface/50";
}

function CopyButton({ value, id }: { value: string; id: string }) {
    const [copied, setCopied] = useState(false);
    const copy = useCallback(() => {
        navigator.clipboard.writeText(value);
        setCopied(true);
        setTimeout(() => setCopied(false), 1200);
    }, [value]);

    return (
        <button
            onClick={copy}
            title={`Copy ${id}`}
            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors flex-shrink-0"
        >
            {copied ? <Check className="w-3.5 h-3.5 text-cortex-success" /> : <Copy className="w-3.5 h-3.5" />}
        </button>
    );
}

export default function SystemDeploymentsPanel() {
    const [snapshot, setSnapshot] = useState<DeploymentTrustSnapshot | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchSnapshot = useCallback(async () => {
        setLoading(true);
        try {
            const res = await fetch("/api/v1/system/deployments/trust");
            const payload = (await res.json().catch(() => ({}))) as DeploymentTrustResponse;
            if (!res.ok || !payload.ok || !payload.data) {
                throw new Error("Deployment trust snapshot unavailable");
            }
            setSnapshot(payload.data);
            setError(null);
        } catch {
            setError("Deployment trust snapshot unavailable");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchSnapshot();
    }, [fetchSnapshot]);

    const health = snapshot?.runtime_health;
    const checkedAt = useMemo(() => {
        if (!snapshot?.checked_at) return null;
        return new Date(snapshot.checked_at);
    }, [snapshot?.checked_at]);

    return (
        <div className="p-6 max-w-5xl mx-auto space-y-4">
            <div className="flex items-center justify-between gap-3">
                <div className="flex items-center gap-2">
                    <ShieldCheck className="w-5 h-5 text-cortex-text-muted" />
                    <div>
                        <h3 className="text-sm font-semibold text-cortex-text-main">Deployment Trust</h3>
                        <p className="text-xs text-cortex-text-muted">Known runtime posture only; unavailable fields stay unknown.</p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    {checkedAt && <span className="text-[10px] text-cortex-text-muted">Checked: {checkedAt.toLocaleTimeString()}</span>}
                    <button
                        onClick={fetchSnapshot}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                    </button>
                </div>
            </div>

            {error && (
                <div className="rounded-lg border border-cortex-danger/30 bg-cortex-danger/10 text-cortex-danger px-3 py-2 text-xs flex items-center gap-2">
                    <AlertTriangle className="w-4 h-4" />
                    {error}
                </div>
            )}

            {loading && !snapshot ? (
                <div className="flex items-center justify-center py-10 text-cortex-text-muted text-xs">
                    <Loader2 className="w-4 h-4 animate-spin mr-2" /> Loading deployment posture...
                </div>
            ) : snapshot ? (
                <>
                    <div className={`rounded-lg border px-4 py-3 flex items-center justify-between gap-3 ${statusClass(health?.status ?? "unknown")}`}>
                        <div>
                            <p className="text-[10px] uppercase font-mono text-cortex-text-muted">Runtime Health</p>
                            <p className="text-sm font-semibold text-cortex-text-main">{health?.status ?? "unknown"}</p>
                        </div>
                        <p className="text-xs font-mono">
                            {health ? `${health.online}/${health.total} online, ${health.degraded} degraded, ${health.offline} offline` : "unknown"}
                        </p>
                    </div>

                    <div className="rounded-xl border border-cortex-border bg-cortex-surface overflow-hidden">
                        {TRUST_ROWS.map(({ key, label }) => {
                            const raw = snapshot[key];
                            const value = typeof raw === "string" ? raw : "unknown";
                            const muted = value === "unknown" || value === "unavailable";
                            return (
                                <div key={key} className="grid grid-cols-[150px_minmax(0,1fr)_32px] items-center gap-2 border-b border-cortex-border last:border-b-0 px-3 py-2">
                                    <span className="text-[10px] uppercase font-mono text-cortex-text-muted">{label}</span>
                                    <code className={`text-xs font-mono truncate ${muted ? "text-cortex-text-muted" : "text-cortex-text-main"}`}>{value}</code>
                                    <CopyButton value={value} id={String(key)} />
                                </div>
                            );
                        })}
                    </div>
                </>
            ) : null}
        </div>
    );
}

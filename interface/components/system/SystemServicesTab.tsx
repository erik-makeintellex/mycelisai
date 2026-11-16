"use client";

import React, { useEffect, useState } from "react";
import { AlertTriangle, Check, CheckCircle, Copy, Loader2, RefreshCw, XCircle } from "lucide-react";
import { useCortexStore, type ServiceHealthStatus } from "@/store/useCortexStore";
import ServiceAddressBookPanel from "./ServiceAddressBookPanel";

const SERVICE_LABELS: Record<string, string> = {
    nats: "NATS JetStream",
    postgres: "PostgreSQL + pgvector",
    cognitive: "Cognitive Engine",
    ollama: "Local AI Provider (Ollama)",
    reactive: "Reactive Engine",
    comms: "External Notifications (Optional)",
    groups_bus: "Group Collaboration Bus",
};

const SERVICE_COMMANDS: Record<string, { up: string; down: string; restart: string }> = {
    nats: {
        up: "uv run inv k8s.bridge",
        down: "uv run inv lifecycle.down",
        restart: "uv run inv k8s.bridge",
    },
    postgres: {
        up: "uv run inv k8s.bridge",
        down: "uv run inv lifecycle.down",
        restart: "uv run inv k8s.bridge",
    },
    cognitive: {
        up: "uv run inv lifecycle.up",
        down: "uv run inv lifecycle.down",
        restart: "uv run inv core.restart",
    },
    reactive: {
        up: "uv run inv lifecycle.up",
        down: "uv run inv lifecycle.down",
        restart: "uv run inv core.restart",
    },
    groups_bus: {
        up: "uv run inv lifecycle.up --frontend",
        down: "uv run inv lifecycle.down",
        restart: "uv run inv core.restart",
    },
};

function CopyButton({ text }: { text: string }) {
    const [copied, setCopied] = useState(false);
    const copy = () => {
        navigator.clipboard.writeText(text).then(() => {
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
        });
    };
    return (
        <button
            onClick={copy}
            className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors"
            title="Copy"
        >
            {copied ? <Check className="w-3 h-3 text-cortex-success" /> : <Copy className="w-3 h-3" />}
        </button>
    );
}

function ServiceCard({ svc }: { svc: ServiceHealthStatus }) {
    const cmds = SERVICE_COMMANDS[svc.name];
    const statusCls =
        svc.status === "online"
            ? "text-cortex-success border-cortex-success/30 bg-cortex-success/5"
            : svc.status === "degraded"
            ? "text-cortex-warning border-cortex-warning/30 bg-cortex-warning/5"
            : "text-cortex-danger border-cortex-danger/30 bg-cortex-danger/5";

    const statusDot =
        svc.status === "online"
            ? "bg-cortex-success animate-pulse"
            : svc.status === "degraded"
            ? "bg-cortex-warning animate-pulse"
            : "bg-cortex-danger";

    const Icon =
        svc.status === "online"
            ? CheckCircle
            : svc.status === "degraded"
            ? AlertTriangle
            : XCircle;

    return (
        <div className={`rounded-lg border p-4 space-y-3 ${statusCls}`}>
            <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2">
                    <span className={`w-2 h-2 rounded-full flex-shrink-0 ${statusDot}`} />
                    <span className="text-sm font-semibold text-cortex-text-main">
                        {SERVICE_LABELS[svc.name] ?? svc.name}
                    </span>
                </div>
                <div className="flex items-center gap-1">
                    <Icon className="w-4 h-4" />
                    <span className="text-xs font-mono uppercase">{svc.status}</span>
                </div>
            </div>

            {svc.detail && (
                <p className="text-xs text-cortex-text-muted">{svc.detail}</p>
            )}
            {svc.address && (
                <p data-testid={`service-card-address-${svc.name}`} className="text-[11px] font-mono text-cortex-primary truncate" title={svc.address}>
                    addr: {svc.address}
                </p>
            )}
            {svc.warning && (
                <p data-testid={`service-card-warning-${svc.name}`} className="text-[10px] font-mono text-cortex-warning">
                    {svc.warning}
                </p>
            )}
            {svc.latency_ms !== undefined && svc.latency_ms > 0 && (
                <p className="text-[10px] text-cortex-text-muted font-mono">{svc.latency_ms}ms latency</p>
            )}

            {cmds && svc.status !== "online" && (
                <div className="space-y-1 pt-1 border-t border-current/20">
                    <p className="text-[10px] text-cortex-text-muted uppercase tracking-wider">Restart command</p>
                    <div className="flex items-center gap-1 bg-cortex-bg rounded px-2 py-1 border border-cortex-border">
                        <code className="text-[10px] text-cortex-primary font-mono flex-1">{cmds.restart}</code>
                        <CopyButton text={cmds.restart} />
                    </div>
                </div>
            )}
        </div>
    );
}

export default function SystemServicesTab() {
    const services = useCortexStore((s) => s.servicesStatus);
    const loading = useCortexStore((s) => s.isFetchingServicesStatus);
    const lastCheckedRaw = useCortexStore((s) => s.servicesStatusUpdatedAt);
    const fetchStatus = useCortexStore((s) => s.fetchServicesStatus);

    useEffect(() => {
        fetchStatus();
        const interval = setInterval(() => {
            fetchStatus();
        }, 10000);
        return () => clearInterval(interval);
    }, [fetchStatus]);

    const lastChecked = lastCheckedRaw ? new Date(lastCheckedRaw) : null;
    const online = services.filter((s) => s.status === "online").length;
    const total = services.length;

    return (
        <div className="p-6 max-w-4xl mx-auto space-y-4">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <span className={`text-sm font-semibold ${
                        online === total && total > 0 ? "text-cortex-success" :
                        online === 0 ? "text-cortex-danger" : "text-cortex-warning"
                    }`}>
                        {loading ? "Checking…" : `${online}/${total} services online`}
                    </span>
                </div>
                <div className="flex items-center gap-2">
                    {lastChecked && (
                        <span className="text-[10px] text-cortex-text-muted">
                            Last checked: {lastChecked.toLocaleTimeString()}
                        </span>
                    )}
                    <button
                        onClick={() => fetchStatus()}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                    </button>
                </div>
            </div>

            {loading ? (
                <div className="flex items-center justify-center py-10 text-cortex-text-muted text-xs">
                    <Loader2 className="w-4 h-4 animate-spin mr-2" /> Probing services…
                </div>
            ) : services.length === 0 ? (
                <div className="text-center py-10 text-cortex-text-muted text-xs">
                    Core API unreachable — run <code className="text-cortex-primary">uv run inv lifecycle.up</code> to start.
                </div>
            ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    {services.map((svc) => <ServiceCard key={svc.name} svc={svc} />)}
                </div>
            )}

            <ServiceAddressBookPanel />

            <div className="rounded-lg border border-cortex-border bg-cortex-surface/50 p-4 space-y-2">
                <h4 className="text-[10px] uppercase tracking-wider text-cortex-text-muted font-semibold">Lifecycle Commands</h4>
                {[
                    { label: "Start all", cmd: "uv run inv lifecycle.up --build --frontend" },
                    { label: "Stop all", cmd: "uv run inv lifecycle.down" },
                    { label: "Restart", cmd: "uv run inv lifecycle.restart --build --frontend" },
                    { label: "Status", cmd: "uv run inv lifecycle.status" },
                    { label: "Health check", cmd: "uv run inv lifecycle.health" },
                ].map(({ label, cmd }) => (
                    <div key={cmd} className="flex items-center gap-2">
                        <span className="text-[10px] text-cortex-text-muted w-24 flex-shrink-0">{label}</span>
                        <div className="flex-1 flex items-center gap-1 bg-cortex-bg rounded px-2 py-1 border border-cortex-border">
                            <code className="text-[10px] text-cortex-primary font-mono flex-1">{cmd}</code>
                            <CopyButton text={cmd} />
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

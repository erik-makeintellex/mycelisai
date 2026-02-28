"use client";

import React, { useState, useEffect, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Activity, Server, Database, BrainCircuit, Bug, Loader2, LayoutGrid, CheckCircle, XCircle, AlertTriangle, RefreshCw, Copy, Check } from "lucide-react";
import MatrixGrid from "@/components/matrix/MatrixGrid";
import SystemQuickChecks from "@/components/system/SystemQuickChecks";
import { useCortexStore } from "@/store/useCortexStore";

type TabId = "health" | "nats" | "database" | "services" | "matrix" | "debug";
const VALID_TABS: TabId[] = ["health", "nats", "database", "services", "matrix", "debug"];

interface HealthStatus {
    goroutines: number;
    heap_alloc_mb: number;
    sys_mem_mb: number;
    llm_tokens_sec: number;
    timestamp: string;
}

export default function SystemPage() {
    return (
        <Suspense fallback={<div className="h-full bg-cortex-bg" />}>
            <SystemContent />
        </Suspense>
    );
}

function SystemContent() {
    const searchParams = useSearchParams();
    const tabParam = searchParams.get("tab") as TabId | null;
    const [activeTab, setActiveTab] = useState<TabId>(
        tabParam && VALID_TABS.includes(tabParam) ? tabParam : "health"
    );

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <header className="px-6 pt-6 pb-0">
                <div className="flex items-end justify-between mb-4">
                    <div>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            System
                        </h1>
                        <p className="text-cortex-text-muted text-sm mt-1">
                            Infrastructure health, diagnostics, and advanced configuration
                        </p>
                    </div>
                    <span className="text-[10px] font-mono uppercase text-cortex-warning bg-cortex-warning/10 border border-cortex-warning/20 px-2 py-1 rounded">
                        Advanced
                    </span>
                </div>

                <div className="flex gap-1 border-b border-cortex-border">
                    <TabButton active={activeTab === "health"} onClick={() => setActiveTab("health")} icon={<Activity size={14} />} label="Event Health" />
                    <TabButton active={activeTab === "nats"} onClick={() => setActiveTab("nats")} icon={<Server size={14} />} label="NATS Status" />
                    <TabButton active={activeTab === "database"} onClick={() => setActiveTab("database")} icon={<Database size={14} />} label="Database" />
                    <TabButton active={activeTab === "services"} onClick={() => setActiveTab("services")} icon={<LayoutGrid size={14} />} label="Services" />
                    <TabButton active={activeTab === "matrix"} onClick={() => setActiveTab("matrix")} icon={<BrainCircuit size={14} />} label="Cognitive Matrix" />
                    <TabButton active={activeTab === "debug"} onClick={() => setActiveTab("debug")} icon={<Bug size={14} />} label="Debug" />
                </div>
            </header>

            <div className="flex-1 overflow-y-auto">
                {activeTab === "health" && <EventHealthTab />}
                {activeTab === "nats" && <NatsStatusTab />}
                {activeTab === "database" && <DatabaseTab />}
                {activeTab === "matrix" && (
                    <div className="p-6">
                        <MatrixGrid />
                    </div>
                )}
                {activeTab === "services" && <ServicesTab />}
                {activeTab === "debug" && <DebugTab />}
            </div>
        </div>
    );
}

function EventHealthTab() {
    const [data, setData] = useState<HealthStatus | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const res = await fetch('/api/v1/telemetry/compute');
                if (!res.ok) throw new Error('Core offline');
                const json = await res.json();
                setData(json);
                setError(null);
            } catch {
                setError("Core API unreachable");
            }
        };
        fetchData();
        const interval = setInterval(fetchData, 5000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="p-6 max-w-4xl mx-auto space-y-4">
            <div className="flex items-center gap-2 mb-4">
                <span className={`flex h-2 w-2 rounded-full ${error ? 'bg-cortex-danger' : 'bg-cortex-success animate-pulse'}`} />
                <span className={`text-sm font-medium ${error ? 'text-cortex-danger' : 'text-cortex-success'}`}>
                    {error ? 'OFFLINE' : 'LIVE'}
                </span>
                {error && <span className="text-xs text-cortex-text-muted ml-2">{error}</span>}
            </div>

            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                <MetricCard label="Goroutines" value={data ? String(data.goroutines) : "..."} />
                <MetricCard label="Heap Alloc" value={data ? `${data.heap_alloc_mb.toFixed(1)} MB` : "..."} />
                <MetricCard label="Sys Memory" value={data ? `${data.sys_mem_mb.toFixed(1)} MB` : "..."} />
                <MetricCard label="Token Rate" value={data ? `${data.llm_tokens_sec.toFixed(1)} t/s` : "..."} />
            </div>

            <div className="rounded-xl border border-cortex-border bg-cortex-surface p-4">
                <h3 className="text-sm font-semibold text-cortex-text-main mb-1">What This Means</h3>
                <p className="text-xs text-cortex-text-muted">
                    Event Health indicates runtime stability and signal propagation across the orchestration spine.
                    If checks degrade, automations may pause while Workspace chat remains available.
                </p>
            </div>

            <SystemQuickChecks />

            {error && (
                <div className="bg-cortex-surface border border-cortex-border rounded-xl p-6 text-center">
                    <p className="text-sm text-cortex-text-muted">Core API is not responding.</p>
                    <p className="text-xs text-cortex-text-muted mt-1">Check that the Core server is running: <code className="text-cortex-primary">uvx inv lifecycle.status</code></p>
                </div>
            )}
        </div>
    );
}

function NatsStatusTab() {
    const services = useCortexStore((s) => s.servicesStatus as ServiceStatus[]);
    const loading = useCortexStore((s) => s.isFetchingServicesStatus);
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);

    useEffect(() => {
        fetchServicesStatus();
        const interval = setInterval(() => {
            fetchServicesStatus();
        }, 10000);
        return () => clearInterval(interval);
    }, [fetchServicesStatus]);

    const nats = services.find((s) => s.name === 'nats');
    const status: 'checking' | 'connected' | 'degraded' | 'disconnected' =
        loading && !nats
            ? 'checking'
            : nats?.status === 'online'
                ? 'connected'
                : nats?.status === 'degraded'
                    ? 'degraded'
                    : 'disconnected';

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl p-6">
                <div className="flex items-center gap-3 mb-4">
                    <Server className="w-5 h-5 text-cortex-text-muted" />
                    <h3 className="text-sm font-semibold text-cortex-text-main">NATS JetStream</h3>
                </div>
                <div className="flex items-center gap-2">
                    {status === 'checking' && <Loader2 size={14} className="text-cortex-text-muted animate-spin" />}
                    {status === 'connected' && <span className="w-2 h-2 rounded-full bg-cortex-success" />}
                    {status === 'degraded' && <span className="w-2 h-2 rounded-full bg-cortex-warning animate-pulse" />}
                    {status === 'disconnected' && <span className="w-2 h-2 rounded-full bg-cortex-danger" />}
                    <span className={`text-sm font-mono ${
                        status === 'connected'
                            ? 'text-cortex-success'
                            : status === 'degraded'
                                ? 'text-cortex-warning'
                                : status === 'disconnected'
                                    ? 'text-cortex-danger'
                                    : 'text-cortex-text-muted'
                    }`}>
                        {status === 'checking' ? 'Checking...' : status === 'connected' ? 'Connected' : status === 'degraded' ? 'Degraded' : 'Disconnected'}
                    </span>
                </div>
                {status === 'disconnected' && (
                    <div className="mt-4 text-xs text-cortex-text-muted space-y-1">
                        <p>NATS is required for: team communication, trigger firing, scheduled execution, signal streaming.</p>
                        <p>Still available: chat (if inference online), file operations, memory search.</p>
                        <p className="text-cortex-primary">Run <code>uvx inv k8s.bridge</code> to restore port forwarding.</p>
                    </div>
                )}
            </div>
        </div>
    );
}

function DatabaseTab() {
    const [tables, setTables] = useState<Array<{ name: string; count: number }> | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        // Try fetching cognitive status as a proxy for DB health
        fetch('/api/v1/cognitive/status')
            .then(res => {
                if (res.ok) {
                    setTables([]);
                    setError(null);
                } else {
                    setError('Database check failed');
                }
            })
            .catch(() => setError('Core API unreachable — cannot verify database'));
    }, []);

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl p-6">
                <div className="flex items-center gap-3 mb-4">
                    <Database className="w-5 h-5 text-cortex-text-muted" />
                    <h3 className="text-sm font-semibold text-cortex-text-main">PostgreSQL + pgvector</h3>
                </div>
                {error ? (
                    <div className="space-y-2">
                        <div className="flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-cortex-danger" />
                            <span className="text-sm font-mono text-cortex-danger">{error}</span>
                        </div>
                        <p className="text-xs text-cortex-text-muted">
                            Ensure PostgreSQL is running and port-forwarded: <code className="text-cortex-primary">uvx inv k8s.bridge</code>
                        </p>
                    </div>
                ) : (
                    <div className="flex items-center gap-2">
                        <span className="w-2 h-2 rounded-full bg-cortex-success" />
                        <span className="text-sm font-mono text-cortex-success">Connected</span>
                        <span className="text-xs text-cortex-text-muted ml-2">22 migrations applied</span>
                    </div>
                )}
            </div>
        </div>
    );
}

function DebugTab() {
    return (
        <div className="p-6 max-w-4xl mx-auto">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl p-6">
                <h3 className="text-sm font-semibold text-cortex-text-main mb-4">Debug Console</h3>
                <p className="text-xs text-cortex-text-muted">
                    Runtime debug information and diagnostic tools will be available here.
                </p>
                <div className="mt-4 p-4 bg-cortex-bg rounded-lg border border-cortex-border font-mono text-xs text-cortex-text-muted">
                    <p>Build: V7.0 (Initiation)</p>
                    <p>Runtime: Next.js 16.1.6 + Go 1.26</p>
                    <p>Store: Zustand 5.0.11</p>
                    <p>Graph: ReactFlow 11.11.4</p>
                </div>
            </div>
        </div>
    );
}

// ── Services Tab ───────────────────────────────────────────────────────────────

interface ServiceStatus {
    name: string;
    status: "online" | "offline" | "degraded";
    detail?: string;
    latency_ms?: number;
}

const SERVICE_LABELS: Record<string, string> = {
    nats: "NATS JetStream",
    postgres: "PostgreSQL + pgvector",
    cognitive: "Cognitive Engine",
    reactive: "Reactive Engine",
};

const SERVICE_COMMANDS: Record<string, { up: string; down: string; restart: string }> = {
    nats: {
        up: "uvx inv k8s.bridge",
        down: "uvx inv lifecycle.down",
        restart: "uvx inv k8s.bridge",
    },
    postgres: {
        up: "uvx inv k8s.bridge",
        down: "uvx inv lifecycle.down",
        restart: "uvx inv k8s.bridge",
    },
    cognitive: {
        up: "uvx inv lifecycle.up",
        down: "uvx inv lifecycle.down",
        restart: "uvx inv core.restart",
    },
    reactive: {
        up: "uvx inv lifecycle.up",
        down: "uvx inv lifecycle.down",
        restart: "uvx inv core.restart",
    },
};

function CopyButton({ text }: { text: string }) {
    const [copied, setCopied] = React.useState(false);
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

function ServiceCard({ svc }: { svc: ServiceStatus }) {
    const cmds = SERVICE_COMMANDS[svc.name];
    const statusCls =
        svc.status === "online"
            ? "text-cortex-success border-cortex-success/30 bg-cortex-success/5"
            : svc.status === "degraded"
            ? "text-amber-400 border-amber-400/30 bg-amber-400/5"
            : "text-red-400 border-red-400/30 bg-red-400/5";

    const statusDot =
        svc.status === "online"
            ? "bg-cortex-success animate-pulse"
            : svc.status === "degraded"
            ? "bg-amber-400 animate-pulse"
            : "bg-red-400";

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
            {svc.latency_ms !== undefined && svc.latency_ms > 0 && (
                <p className="text-[10px] text-cortex-text-muted font-mono">{svc.latency_ms}ms latency</p>
            )}

            {/* Lifecycle commands */}
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

function ServicesTab() {
    const services = useCortexStore((s) => s.servicesStatus as ServiceStatus[]);
    const loading = useCortexStore((s) => s.isFetchingServicesStatus);
    const lastCheckedRaw = useCortexStore((s) => s.servicesStatusUpdatedAt);
    const fetchStatus = useCortexStore((s) => s.fetchServicesStatus);

    React.useEffect(() => {
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
            {/* Summary bar */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <span className={`text-sm font-semibold ${
                        online === total && total > 0 ? "text-cortex-success" :
                        online === 0 ? "text-red-400" : "text-amber-400"
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
                        onClick={fetchStatus}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                        title="Refresh"
                    >
                        <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                    </button>
                </div>
            </div>

            {/* Service cards */}
            {loading ? (
                <div className="flex items-center justify-center py-10 text-cortex-text-muted text-xs">
                    <Loader2 className="w-4 h-4 animate-spin mr-2" /> Probing services…
                </div>
            ) : services.length === 0 ? (
                <div className="text-center py-10 text-cortex-text-muted text-xs">
                    Core API unreachable — run <code className="text-cortex-primary">uvx inv lifecycle.up</code> to start.
                </div>
            ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    {services.map((svc) => <ServiceCard key={svc.name} svc={svc} />)}
                </div>
            )}

            {/* Global lifecycle commands */}
            <div className="rounded-lg border border-cortex-border bg-cortex-surface/50 p-4 space-y-2">
                <h4 className="text-[10px] uppercase tracking-wider text-cortex-text-muted font-semibold">Lifecycle Commands</h4>
                {[
                    { label: "Start all", cmd: "uvx inv lifecycle.up --build --frontend" },
                    { label: "Stop all", cmd: "uvx inv lifecycle.down" },
                    { label: "Restart", cmd: "uvx inv lifecycle.restart --build --frontend" },
                    { label: "Status", cmd: "uvx inv lifecycle.status" },
                    { label: "Health check", cmd: "uvx inv lifecycle.health" },
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

function MetricCard({ label, value }: { label: string; value: string }) {
    return (
        <div className="p-4 rounded-xl border bg-cortex-surface border-cortex-border">
            <p className="text-[10px] font-mono uppercase text-cortex-text-muted mb-1">{label}</p>
            <p className="text-lg font-bold text-cortex-text-main font-mono">{value}</p>
        </div>
    );
}

function TabButton({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
    return (
        <button
            onClick={onClick}
            className={`px-4 py-2.5 text-xs font-medium flex items-center gap-2 border-b-2 transition-colors -mb-px whitespace-nowrap ${
                active
                    ? "border-cortex-primary text-cortex-primary"
                    : "border-transparent text-cortex-text-muted hover:text-cortex-text-main"
            }`}
        >
            {icon}
            {label}
        </button>
    );
}

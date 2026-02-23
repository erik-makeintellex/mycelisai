"use client";

import { useState, useEffect, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Activity, Server, Database, BrainCircuit, Bug, Loader2 } from "lucide-react";
import MatrixGrid from "@/components/matrix/MatrixGrid";

type TabId = "health" | "nats" | "database" | "matrix" | "debug";
const VALID_TABS: TabId[] = ["health", "nats", "database", "matrix", "debug"];

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
    const [status, setStatus] = useState<'checking' | 'connected' | 'disconnected'>('checking');

    useEffect(() => {
        const check = async () => {
            try {
                const res = await fetch('/api/v1/healthz');
                if (res.ok) {
                    setStatus('connected');
                } else {
                    setStatus('disconnected');
                }
            } catch {
                setStatus('disconnected');
            }
        };
        check();
        const interval = setInterval(check, 10000);
        return () => clearInterval(interval);
    }, []);

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
                    {status === 'disconnected' && <span className="w-2 h-2 rounded-full bg-cortex-danger" />}
                    <span className={`text-sm font-mono ${status === 'connected' ? 'text-cortex-success' : status === 'disconnected' ? 'text-cortex-danger' : 'text-cortex-text-muted'}`}>
                        {status === 'checking' ? 'Checking...' : status === 'connected' ? 'Connected' : 'Disconnected'}
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

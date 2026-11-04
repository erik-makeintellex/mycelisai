"use client";

import React, { useState, useEffect, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Activity, Server, Database, Loader2, LayoutGrid } from "lucide-react";
import SystemQuickChecks from "@/components/system/SystemQuickChecks";
import SystemStatusSummaryCard from "@/components/system/SystemStatusSummaryCard";
import SystemServicesTab from "@/components/system/SystemServicesTab";
import { useCortexStore, type ServiceHealthStatus } from "@/store/useCortexStore";

type TabId = "health" | "nats" | "database" | "services";
const VALID_TABS: TabId[] = ["health", "nats", "database", "services"];

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
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const tabParam = (searchParams?.get("tab") as TabId | null) ?? null;
    const [activeTab, setActiveTab] = useState<TabId>(
        tabParam && VALID_TABS.includes(tabParam) ? tabParam : "health"
    );

    if (!advancedMode) {
        return <SystemStatusSummaryCard />;
    }

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <header className="px-6 pt-6 pb-0">
                <div className="flex items-end justify-between mb-4">
                    <div>
                        <h1 className="text-2xl font-bold text-cortex-text-main tracking-tight">
                            System
                        </h1>
                        <p className="text-cortex-text-muted text-sm mt-1">
                            Service health, event flow, storage checks, and recovery guidance
                        </p>
                    </div>
                    <span className="text-[10px] font-mono uppercase text-cortex-warning bg-cortex-warning/10 border border-cortex-warning/20 px-2 py-1 rounded">
                        Advanced
                    </span>
                </div>

                <div className="flex gap-1 border-b border-cortex-border">
                    <TabButton active={activeTab === "health"} onClick={() => setActiveTab("health")} icon={<Activity size={14} />} label="Runtime Health" />
                    <TabButton active={activeTab === "nats"} onClick={() => setActiveTab("nats")} icon={<Server size={14} />} label="Event Bus" />
                    <TabButton active={activeTab === "database"} onClick={() => setActiveTab("database")} icon={<Database size={14} />} label="Storage" />
                    <TabButton active={activeTab === "services"} onClick={() => setActiveTab("services")} icon={<LayoutGrid size={14} />} label="Services" />
                </div>
            </header>

            <div className="flex-1 overflow-y-auto">
                {activeTab === "health" && <EventHealthTab />}
                {activeTab === "nats" && <NatsStatusTab />}
                {activeTab === "database" && <DatabaseTab />}
                {activeTab === "services" && <SystemServicesTab />}
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
                <h3 className="text-sm font-semibold text-cortex-text-main mb-1">What this means</h3>
                <p className="text-xs text-cortex-text-muted">
                    Runtime health tracks whether the main backend is responsive enough for organization activity, approvals, and recent history to stay dependable.
                    If checks degrade, automations may pause while the rest of the workspace falls back to recovery guidance.
                </p>
            </div>

            <SystemQuickChecks />

            {error && (
                <div className="bg-cortex-surface border border-cortex-border rounded-xl p-6 text-center">
                    <p className="text-sm text-cortex-text-muted">Core API is not responding.</p>
                    <p className="text-xs text-cortex-text-muted mt-1">Check the current stack status with <code className="text-cortex-primary">uv run inv lifecycle.status</code>.</p>
                </div>
            )}
        </div>
    );
}

function NatsStatusTab() {
    const services = useCortexStore((s) => s.servicesStatus);
    const loading = useCortexStore((s) => s.isFetchingServicesStatus);
    const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);

    useEffect(() => {
        fetchServicesStatus();
        const interval = setInterval(() => {
            fetchServicesStatus();
        }, 10000);
        return () => clearInterval(interval);
    }, [fetchServicesStatus]);

    const nats = services.find((s: ServiceHealthStatus) => s.name === 'nats');
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
                    <h3 className="text-sm font-semibold text-cortex-text-main">Event Bus</h3>
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
                        <p>The event bus powers team coordination, trigger processing, approvals, and live activity updates.</p>
                        <p>Some workspace views may still load, but automations and live orchestration will degrade until this path is restored.</p>
                        <p className="text-cortex-primary">Run <code>uv run inv k8s.bridge</code> to restore the local bridge.</p>
                    </div>
                )}
            </div>
        </div>
    );
}

function DatabaseTab() {
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetch('/api/v1/cognitive/status')
            .then(res => {
                if (res.ok) {
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
                    <h3 className="text-sm font-semibold text-cortex-text-main">Storage</h3>
                </div>
                {error ? (
                    <div className="space-y-2">
                        <div className="flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-cortex-danger" />
                            <span className="text-sm font-mono text-cortex-danger">{error}</span>
                        </div>
                        <p className="text-xs text-cortex-text-muted">
                            Ensure the storage services are running and bridged with <code className="text-cortex-primary">uv run inv k8s.bridge</code>.
                        </p>
                    </div>
                ) : (
                    <div className="flex items-center gap-2">
                        <span className="w-2 h-2 rounded-full bg-cortex-success" />
                        <span className="text-sm font-mono text-cortex-success">Connected</span>
                        <span className="text-xs text-cortex-text-muted ml-2">Core storage is responding to health checks.</span>
                    </div>
                )}
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

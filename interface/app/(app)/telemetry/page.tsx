"use client";

import { useEffect, useState } from 'react';
import { Activity, Server, Cpu, Zap, Box } from 'lucide-react';

interface TelemetrySnapshot {
    goroutines: number;
    heap_alloc_mb: number;
    sys_mem_mb: number;
    llm_tokens_sec: number;
    timestamp: string;
}

export default function TelemetryPage() {
    const [data, setData] = useState<TelemetrySnapshot | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const res = await fetch('/api/v1/telemetry/compute');
                if (!res.ok) throw new Error('Failed to fetch telemetry');
                const json = await res.json();
                setData(json);
                setError(null);
            } catch (err) {
                console.error(err);
                setError("Connection lost");
            } finally {
                // fetch complete
            }
        };

        fetchData();
        const interval = setInterval(fetchData, 1000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="p-8 max-w-7xl mx-auto space-y-8">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-cortex-text-main">System Telemetry</h1>
                    <p className="text-cortex-text-muted">Real-time infrastructure monitoring and node status.</p>
                </div>
                <div className="flex items-center gap-2">
                    <span className={`flex h-2 w-2 rounded-full ${error ? 'bg-red-500' : 'bg-emerald-500 animate-pulse'}`}></span>
                    <span className={`text-sm font-medium ${error ? 'text-red-400' : 'text-emerald-400'}`}>
                        {error ? 'OFFLINE' : 'LIVE'}
                    </span>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <StatusCard
                    icon={Server}
                    label="Goroutines"
                    value={data ? data.goroutines.toString() : "..."}
                />
                <StatusCard
                    icon={Box}
                    label="Heap Alloc"
                    value={data ? `${data.heap_alloc_mb.toFixed(1)} MB` : "..."}
                />
                <StatusCard
                    icon={Cpu}
                    label="Sys Memory"
                    value={data ? `${data.sys_mem_mb.toFixed(1)} MB` : "..."}
                />
                <StatusCard
                    icon={Zap}
                    label="Token Rate"
                    value={data ? `${data.llm_tokens_sec.toFixed(1)} t/s` : "..."}
                />
            </div>

            {/* Placeholder for future charts */}
            <div className="bg-cortex-surface rounded-lg border border-cortex-border h-96 flex items-center justify-center text-cortex-text-muted">
                <div className="text-center">
                    <Activity className="w-12 h-12 mx-auto mb-4 opacity-20" />
                    <p>Detailed Metrics Visualization</p>
                    <p className="text-xs mt-1 opacity-60">Waiting for Prometheus/Grafana integration</p>
                </div>
            </div>
        </div>
    );
}

function StatusCard({ icon: Icon, label, value }: { icon: any, label: string, value: string }) {
    return (
        <div className="p-4 rounded-lg border flex items-center justify-between bg-cortex-surface border-cortex-border">
            <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-cortex-bg text-cortex-text-muted">
                    <Icon className="w-5 h-5" />
                </div>
                <span className="font-medium text-cortex-text-muted">{label}</span>
            </div>
            <span className="text-lg font-bold text-cortex-text-main font-mono">{value}</span>
        </div>
    );
}

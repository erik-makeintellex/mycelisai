"use client";

import React, { useEffect, useState } from 'react';
import { Activity, Server, Database, Shield, Cpu, Zap, Box } from 'lucide-react';

interface TelemetrySnapshot {
    goroutines: number;
    heap_alloc_mb: number;
    sys_mem_mb: number;
    llm_tokens_sec: number;
    timestamp: string;
}

export default function TelemetryPage() {
    const [data, setData] = useState<TelemetrySnapshot | null>(null);
    const [loading, setLoading] = useState(true);
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
                setLoading(false);
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
                    <h1 className="text-2xl font-bold text-zinc-900">System Telemetry</h1>
                    <p className="text-zinc-500">Real-time infrastructure monitoring and node status.</p>
                </div>
                <div className="flex items-center gap-2">
                    <span className={`flex h-2 w-2 rounded-full ${error ? 'bg-red-500' : 'bg-emerald-500 animate-pulse'}`}></span>
                    <span className={`text-sm font-medium ${error ? 'text-red-600' : 'text-emerald-600'}`}>
                        {error ? 'OFFLINE' : 'LIVE'}
                    </span>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <StatusCard 
                    icon={Server} 
                    label="Goroutines" 
                    value={data ? data.goroutines.toString() : "..."} 
                    status="success" 
                />
                <StatusCard 
                    icon={Box} 
                    label="Heap Alloc" 
                    value={data ? `${data.heap_alloc_mb.toFixed(1)} MB` : "..."} 
                    status="success" 
                />
                <StatusCard 
                    icon={Cpu} 
                    label="Sys Memory" 
                    value={data ? `${data.sys_mem_mb.toFixed(1)} MB` : "..."} 
                    status="warning" 
                />
                <StatusCard 
                    icon={Zap} 
                    label="Token Rate" 
                    value={data ? `${data.llm_tokens_sec.toFixed(1)} t/s` : "..."} 
                    status="success" 
                />
            </div>

            {/* Placeholder for future charts */}
            <div className="bg-zinc-50 rounded-lg border border-zinc-200 h-96 flex items-center justify-center text-zinc-400">
                <div className="text-center">
                    <Activity className="w-12 h-12 mx-auto mb-4 opacity-20" />
                    <p>Detailed Metrics Visualization</p>
                    <p className="text-xs mt-1 opacity-60">Waiting for Prometheus/Grafana integration</p>
                </div>
            </div>
        </div>
    );
}

function StatusCard({ icon: Icon, label, value, status }: { icon: any, label: string, value: string, status: 'success' | 'warning' | 'error' }) {
    const colors = {
        success: 'text-emerald-600 bg-emerald-50 border-emerald-100',
        warning: 'text-amber-600 bg-amber-50 border-amber-100',
        error: 'text-red-600 bg-red-50 border-red-100'
    };

    return (
        <div className={`p-4 rounded-lg border flex items-center justify-between bg-white border-zinc-200`}>
            <div className="flex items-center gap-3">
                <div className={`p-2 rounded-md bg-zinc-100 text-zinc-600`}>
                    <Icon className="w-5 h-5" />
                </div>
                <span className="font-medium text-zinc-600">{label}</span>
            </div>
            <span className="text-lg font-bold text-zinc-900 font-mono">{value}</span>
        </div>
    );
}

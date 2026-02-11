"use client";

import React from 'react';
import { Activity, Server, Database, Shield } from 'lucide-react';

export default function TelemetryPage() {
    return (
        <div className="p-8 max-w-7xl mx-auto space-y-8">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-zinc-900">System Telemetry</h1>
                    <p className="text-zinc-500">Real-time infrastructure monitoring and node status.</p>
                </div>
                <div className="flex items-center gap-2">
                    <span className="flex h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></span>
                    <span className="text-sm font-medium text-emerald-600">LIVE</span>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <StatusCard icon={Server} label="Core Nodes" value="3/3" status="success" />
                <StatusCard icon={Database} label="Postgres" value="Healthy" status="success" />
                <StatusCard icon={Activity} label="NATS Bus" value="Connected" status="success" />
                <StatusCard icon={Shield} label="Guard" value="Active" status="success" />
            </div>

            <div className="bg-zinc-50 rounded-lg border border-zinc-200 h-96 flex items-center justify-center text-zinc-400">
                <p>Telemetry Visualization Placeholder (Grafana / Prometheus Embed)</p>
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
        <div className={`p-4 rounded-lg border ${colors[status]} flex items-center justify-between`}>
            <div className="flex items-center gap-3">
                <div className={`p-2 rounded-md bg-white/50`}>
                    <Icon className="w-5 h-5" />
                </div>
                <span className="font-medium">{label}</span>
            </div>
            <span className="text-lg font-bold">{value}</span>
        </div>
    );
}

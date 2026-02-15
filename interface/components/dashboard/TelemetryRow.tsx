"use client";

import React, { useEffect, useState, useCallback } from 'react';
import { Cpu, MemoryStick, Zap, Activity } from 'lucide-react';

interface TelemetryData {
    goroutines: number;
    heap_alloc_mb: number;
    sys_mem_mb: number;
    llm_tokens_sec: number;
    timestamp: string;
}

interface MetricCardProps {
    icon: React.ComponentType<{ className?: string }>;
    label: string;
    value: string;
    unit: string;
    history: number[];
    color: string;
}

function Sparkline({ data, color }: { data: number[]; color: string }) {
    if (data.length < 2) return null;

    const max = Math.max(...data, 1);
    const min = Math.min(...data, 0);
    const range = max - min || 1;
    const w = 80;
    const h = 20;

    const points = data
        .map((v, i) => {
            const x = (i / (data.length - 1)) * w;
            const y = h - ((v - min) / range) * h;
            return `${x},${y}`;
        })
        .join(' ');

    return (
        <svg width={w} height={h} className="opacity-60">
            <polyline
                points={points}
                fill="none"
                stroke={color}
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}

function MetricCard({ icon: Icon, label, value, unit, history, color }: MetricCardProps) {
    return (
        <div className="flex-1 px-3 py-2 bg-cortex-surface border border-cortex-border rounded-lg flex items-center gap-3" data-testid="telemetry-card">
            <div className="flex flex-col flex-1 min-w-0">
                <div className="flex items-center gap-1.5 mb-0.5">
                    <Icon className="w-3 h-3 text-cortex-text-muted" />
                    <span className="text-[9px] font-mono uppercase text-cortex-text-muted">{label}</span>
                </div>
                <div className="flex items-baseline gap-1">
                    <span className="text-lg font-mono font-bold text-cortex-text-main tabular-nums">{value}</span>
                    <span className="text-[9px] font-mono text-cortex-text-muted">{unit}</span>
                </div>
            </div>
            <Sparkline data={history} color={color} />
        </div>
    );
}

export default function TelemetryRow() {
    const [current, setCurrent] = useState<TelemetryData | null>(null);
    const [history, setHistory] = useState<TelemetryData[]>([]);

    const poll = useCallback(async () => {
        try {
            const res = await fetch('/api/v1/telemetry/compute');
            if (res.ok) {
                const data: TelemetryData = await res.json();
                setCurrent(data);
                setHistory((prev) => [...prev, data].slice(-20));
            }
        } catch { /* degraded mode */ }
    }, []);

    useEffect(() => {
        poll();
        const interval = setInterval(poll, 5000);
        return () => clearInterval(interval);
    }, [poll]);

    if (!current) {
        return (
            <div className="flex gap-3 px-4 py-2 border-b border-cortex-border">
                {[1, 2, 3, 4].map((i) => (
                    <div key={i} className="flex-1 h-14 bg-cortex-surface border border-cortex-border rounded-lg animate-pulse" />
                ))}
            </div>
        );
    }

    return (
        <div className="flex gap-3 px-4 py-2 border-b border-cortex-border">
            <MetricCard
                icon={Cpu}
                label="Goroutines"
                value={current.goroutines.toString()}
                unit=""
                history={history.map((h) => h.goroutines)}
                color="#7367F0"
            />
            <MetricCard
                icon={MemoryStick}
                label="Heap"
                value={current.heap_alloc_mb.toFixed(1)}
                unit="MB"
                history={history.map((h) => h.heap_alloc_mb)}
                color="#00CFE8"
            />
            <MetricCard
                icon={Activity}
                label="System"
                value={current.sys_mem_mb.toFixed(0)}
                unit="MB"
                history={history.map((h) => h.sys_mem_mb)}
                color="#28C76F"
            />
            <MetricCard
                icon={Zap}
                label="LLM Tok/s"
                value={current.llm_tokens_sec.toFixed(1)}
                unit="t/s"
                history={history.map((h) => h.llm_tokens_sec)}
                color="#FF9F43"
            />
        </div>
    );
}

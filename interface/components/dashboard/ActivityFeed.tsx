"use client"

import { CheckCircle, AlertTriangle, Activity, Server, Loader2 } from "lucide-react"

// Types for Real Data
export interface ServiceHealth {
    name: string
    status: "online" | "degraded" | "offline"
    meta?: string
    host?: string
}

export interface LogEntry {
    timestamp: string
    actor: string
    message: string
    level: "info" | "warn" | "error" | "debug"
}

export function InfrastructureStatus() {
    // In a real implementation, this would come from a useSystemHealth() hook
    // Defaulting to what we expect the "Real" initial state to be
    const services: ServiceHealth[] = [
        { name: "Core Brain", status: "online", meta: "v6.1.0", host: "localhost:8080" },
        { name: "NATS JetStream", status: "online", meta: "Cluster", host: "localhost:4222" },
        { name: "MCP Bridge", status: "online", meta: "Active", host: "localhost" },
    ]

    return (
        <div className="bg-white border border-zinc-200 rounded-lg shadow-sm flex flex-col h-full">
            <div className="h-10 border-b border-zinc-200 flex items-center px-4 bg-zinc-50/50 rounded-t-lg justify-between">
                <h3 className="text-xs font-semibold text-slate-700 uppercase tracking-wide">Infrastructure</h3>
                <span className="text-[10px] text-zinc-400 font-mono">localhost</span>
            </div>
            <div className="p-2 space-y-1">
                {services.map((svc) => (
                    <div key={svc.name} className="flex items-center justify-between p-2 hover:bg-zinc-50 rounded-md transition-colors group cursor-default">
                        <div className="flex items-center gap-3">
                            {svc.status === "online" && <CheckCircle size={14} className="text-emerald-500" />}
                            {svc.status === "degraded" && <Loader2 size={14} className="text-amber-500 animate-spin" />}
                            {svc.status === "offline" && <AlertTriangle size={14} className="text-rose-500" />}

                            <div className="flex flex-col">
                                <span className="text-sm font-medium text-slate-700 leading-none">{svc.name}</span>
                                <span className="text-[10px] text-zinc-400 mt-1 font-mono">{svc.host}</span>
                            </div>
                        </div>
                        <div className="text-xs text-zinc-500 font-medium">
                            {svc.meta}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    )
}

export function CortexStream() {
    // In Phase 3, this will be connected to useLogStream()
    // For now, we show the "Waiting" state instead of fake data
    const logs: LogEntry[] = []

    return (
        <div className="bg-white border border-zinc-200 rounded-lg shadow-sm flex flex-col h-full">
            <div className="h-10 border-b border-zinc-200 flex items-center px-4 bg-zinc-50/50 rounded-t-lg justify-between">
                <h3 className="text-xs font-semibold text-slate-700 uppercase tracking-wide">Cortex Stream</h3>
                <div className="flex items-center gap-2">
                    <span className="text-[10px] text-emerald-600 font-medium flex items-center gap-1">
                        <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                        Live
                    </span>
                    <Activity size={12} className="text-zinc-400" />
                </div>
            </div>
            <div className="p-0 flex-1 overflow-y-auto max-h-[300px] flex flex-col">
                {logs.length === 0 ? (
                    <div className="flex-1 flex flex-col items-center justify-center text-zinc-400 gap-2 min-h-[150px]">
                        <Server size={24} className="opacity-20" />
                        <span className="text-xs font-medium">Waiting for telemetry...</span>
                    </div>
                ) : (
                    logs.map((log, i) => (
                        <div key={i} className="flex gap-3 items-start px-4 py-2 border-b border-zinc-100 last:border-0 hover:bg-zinc-50 font-mono text-xs">
                            <span className="text-zinc-400 shrink-0">[{log.timestamp}]</span>
                            <span className="font-semibold shrink-0 w-24 text-slate-700">[{log.actor}]</span>
                            <span className="text-slate-600 break-words">{log.message}</span>
                        </div>
                    ))
                )}
            </div>
        </div>
    )
}

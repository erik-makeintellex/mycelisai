"use client";

import { useEffect, useState } from "react";
import { Terminal, Maximize2 } from "lucide-react";

export default function LogStream() {
    const [logs, setLogs] = useState<string[]>([]);

    useEffect(() => {
        // Mock simulation
        const timer = setInterval(() => {
            const msgs = [
                "[INFO] Gatekeeper syncing policy...",
                "[DEBUG] NATS heartbeart received (RTT=4ms)",
                "[INFO] Agent 'marketing-01' reported status: IDLE",
                "[WARN] Rate limit approaching for tenant 'default'",
                "[INFO] Metric collected: cpu_usage=45%",
                "[DEBUG] Reconciling deployment state...",
            ];
            const msg = msgs[Math.floor(Math.random() * msgs.length)];
            const ts = new Date().toISOString().split('T')[1].split('.')[0];
            setLogs(prev => [`${ts} ${msg}`, ...prev].slice(0, 100));
        }, 800);
        return () => clearInterval(timer);
    }, []);

    return (
        <div className="bg-[#0b101a] rounded-xl border border-slate-800 h-full flex flex-col overflow-hidden relative font-mono text-xs">
            {/* Terminal Header */}
            <div className="bg-slate-900/80 px-4 py-2 flex items-center justify-between border-b border-slate-800">
                <div className="flex items-center gap-2 text-slate-500">
                    <Terminal size={12} className="text-blue-500" />
                    <span className="font-bold tracking-wider text-[10px]">CORTEX.LOGS.STREAM</span>
                </div>
                <div className="flex gap-1.5">
                    <div className="w-2 h-2 rounded-full bg-slate-700"></div>
                    <div className="w-2 h-2 rounded-full bg-slate-700"></div>
                </div>
            </div>

            {/* Terminal Body */}
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
                <div className="flex flex-col-reverse min-h-full">
                    {logs.map((log, i) => {
                        const parts = log.split(" ");
                        const ts = parts[0];
                        const level = parts[1];
                        const content = parts.slice(2).join(" ");

                        const levelColor =
                            level.includes("WARN") ? "text-amber-400" :
                                level.includes("ERROR") ? "text-red-400" :
                                    level.includes("DEBUG") ? "text-slate-500" :
                                        "text-blue-400";

                        return (
                            <div key={i} className="mb-1 flex gap-3 hover:bg-slate-800/30 px-2 py-0.5 rounded -mx-2 transition-colors group">
                                <span className="text-slate-600 select-none w-16 tabular-nums opacity-50 group-hover:opacity-100 transition-opacity">{ts}</span>
                                <span className={`font-bold w-12 ${levelColor}`}>{level.replace(/[\[\]]/g, "")}</span>
                                <span className="text-slate-300">{content}</span>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
}

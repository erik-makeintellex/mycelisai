"use client";

import { useEffect, useState } from "react";
import { Terminal } from "lucide-react";

interface LogEntry {
    id: string;
    trace_id: string;
    timestamp: string;
    level: string;
    source: string;
    intent: string;
    message: string;
    context: Record<string, unknown>;
}

export default function LogStream() {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [lastFetch, setLastFetch] = useState(Date.now());

    // Polling Loop
    useEffect(() => {
        const timer = setInterval(() => {
            fetch("/api/v1/memory/stream")
                .then(res => res.json())
                .then((data: LogEntry[]) => {
                    if (Array.isArray(data)) {
                        setLogs(data); // Replace buffer with real state
                    }
                })
                .catch(err => console.error("Log fetch failed", err));
            setLastFetch(Date.now());
        }, 2000); // Poll every 2s

        return () => clearInterval(timer);
    }, []);

    return (
        <div className="bg-[#0b101a] rounded-xl border border-slate-800 h-full flex flex-col overflow-hidden relative font-mono text-xs">
            {/* Terminal Header */}
            <div className="bg-slate-900/80 px-4 py-2 flex items-center justify-between border-b border-slate-800">
                <div className="flex items-center gap-2 text-slate-500">
                    <Terminal size={12} className="text-blue-500" />
                    <span className="font-bold tracking-wider text-[10px]">CORTEX.LOGS.STREAM</span>
                    <span className="text-[10px] text-slate-600 animate-pulse">
                        (LIVE: {logs.length} events)
                    </span>
                </div>
                <div className="flex gap-1.5">
                    <div className="w-2 h-2 rounded-full bg-slate-700"></div>
                </div>
            </div>

            {/* Terminal Body */}
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
                <div className="flex flex-col gap-1">
                    {logs.map((log) => {
                        const levelColor =
                            log.level === "WARN" ? "text-amber-400" :
                                log.level === "ERROR" ? "text-red-400" :
                                    log.level === "DEBUG" ? "text-slate-500" :
                                        "text-blue-400";

                        const ts = new Date(log.timestamp).toLocaleTimeString();

                        return (
                            <div key={log.id} className="flex gap-3 hover:bg-slate-800/30 px-2 py-0.5 rounded -mx-2 transition-colors group border-l-2 border-transparent hover:border-slate-700">
                                <span className="text-slate-600 select-none w-20 tabular-nums opacity-50 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                                    {ts}
                                </span>
                                <span className="text-slate-500 w-24 truncate" title={log.source}>
                                    [{log.source}]
                                </span>
                                <span className={`font-bold w-12 ${levelColor}`}>
                                    {log.level}
                                </span>
                                <span className="text-slate-300 truncate" title={log.message}>
                                    {log.message}
                                </span>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
}

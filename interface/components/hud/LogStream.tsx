"use client";

import { useEffect, useState } from "react";
import { Terminal } from "lucide-react";
import { useCortexStore, type LogEntry } from "@/store/useCortexStore";
import { logEntryToDetail } from "@/lib/signalNormalize";

export default function LogStream() {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const selectSignalDetail = useCortexStore((s) => s.selectSignalDetail);

    useEffect(() => {
        const timer = setInterval(() => {
            fetch("/api/v1/memory/stream")
                .then(res => res.json())
                .then((data: LogEntry[]) => {
                    if (Array.isArray(data)) {
                        setLogs(data);
                    }
                })
                .catch(err => console.error("Log fetch failed", err));
        }, 5000);

        return () => clearInterval(timer);
    }, []);

    return (
        <div className="bg-cortex-bg rounded-xl border border-cortex-border h-full flex flex-col overflow-hidden relative font-mono text-xs">
            {/* Terminal Header */}
            <div className="bg-cortex-surface/80 px-4 py-2 flex items-center justify-between border-b border-cortex-border">
                <div className="flex items-center gap-2 text-cortex-text-muted">
                    <Terminal size={12} className="text-cortex-info" />
                    <span className="font-bold tracking-wider text-[10px]">CORTEX.LOGS.STREAM</span>
                    <span className="text-[10px] text-cortex-text-muted/60 animate-pulse">
                        (LIVE: {logs.length} events)
                    </span>
                </div>
                <div className="flex gap-1.5">
                    <div className="w-2 h-2 rounded-full bg-cortex-border"></div>
                </div>
            </div>

            {/* Terminal Body */}
            <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
                <div className="flex flex-col gap-1">
                    {logs.map((log, index) => {
                        const levelColor =
                            log.level === "WARN" ? "text-amber-400" :
                                log.level === "ERROR" ? "text-red-400" :
                                    log.level === "DEBUG" ? "text-cortex-text-muted" :
                                        "text-cortex-info";

                        const ts = new Date(log.timestamp).toLocaleTimeString();

                        return (
                            <div
                                key={`${log.id}-${index}`}
                                className="flex gap-3 hover:bg-cortex-surface/30 px-2 py-0.5 rounded -mx-2 transition-colors group border-l-2 border-transparent hover:border-cortex-border cursor-pointer"
                                onClick={() => selectSignalDetail(logEntryToDetail(log))}
                            >
                                <span className="text-cortex-text-muted/50 select-none w-20 tabular-nums group-hover:text-cortex-text-muted transition-opacity whitespace-nowrap">
                                    {ts}
                                </span>
                                <span className="text-cortex-text-muted w-24 truncate" title={log.source}>
                                    [{log.source}]
                                </span>
                                <span className={`font-bold w-12 ${levelColor}`}>
                                    {log.level}
                                </span>
                                <span className="text-cortex-text-main truncate" title={log.message}>
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

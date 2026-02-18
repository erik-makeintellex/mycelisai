"use client";

import { useEffect, useState } from "react";

export default function SystemStatus() {
    const [status, setStatus] = useState<"loading" | "online" | "offline">("loading");

    useEffect(() => {
        const check = async () => {
            try {
                const res = await fetch("/api/healthz");
                setStatus(res.ok ? "online" : "offline");
            } catch {
                setStatus("offline");
            }
        };
        check();
        const interval = setInterval(check, 5000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="flex items-center gap-3 bg-cortex-surface px-4 py-1.5 rounded-full border border-cortex-border shadow-sm">
            <div className="relative flex h-2.5 w-2.5">
                {status === "online" && <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-cortex-success opacity-75"></span>}
                <span className={`relative inline-flex rounded-full h-2.5 w-2.5 ${
                    status === "online" ? "bg-cortex-success" :
                    status === "offline" ? "bg-red-500" : "bg-amber-400"
                }`}></span>
            </div>
            <span className={`text-[11px] font-bold tracking-widest font-mono ${
                status === "online" ? "text-cortex-success" :
                status === "offline" ? "text-red-400" : "text-amber-400"
            }`}>
                {status === "loading" ? "SYNCING" : status === "online" ? "CORE ONLINE" : "CORE OFFLINE"}
            </span>
        </div>
    );
}

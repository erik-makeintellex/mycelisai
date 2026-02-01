"use client";

import { useEffect, useState } from "react";
import axios from "axios";

export default function SystemStatus() {
    const [status, setStatus] = useState<"loading" | "online" | "offline">("loading");

    useEffect(() => {
        const check = async () => {
            try {
                await axios.get("/api/healthz");
                setStatus("online");
            } catch (e) {
                setStatus("offline");
            }
        };
        check();
        const interval = setInterval(check, 5000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="flex items-center gap-3 bg-slate-800 px-4 py-1.5 rounded-full border border-slate-700 shadow-sm">
            <div className="relative flex h-2.5 w-2.5">
                {status === "online" && <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>}
                <span className={`relative inline-flex rounded-full h-2.5 w-2.5 ${status === "online" ? "bg-emerald-500" :
                        status === "offline" ? "bg-red-500" : "bg-amber-400"
                    }`}></span>
            </div>
            <span className={`text-[11px] font-bold tracking-widest ${status === "online" ? "text-emerald-400" :
                    status === "offline" ? "text-red-400" : "text-amber-400"
                }`}>
                {status === "loading" ? "SYNCING" : status === "online" ? "ONLINE" : "OFFLINE"}
            </span>
        </div>
    );
}

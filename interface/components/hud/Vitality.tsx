
"use client";

import { useEffect, useState } from "react";
import { Activity } from "lucide-react";

export function Vitality() {
    const [pulse, setPulse] = useState<"alive" | "dead" | "loading">("loading");

    useEffect(() => {
        const checkHealth = async () => {
            try {
                const res = await fetch("/api/v1/health");
                if (res.ok) {
                    setPulse("alive");
                } else {
                    setPulse("dead");
                }
            } catch (e) {
                setPulse("dead");
            }
        };

        checkHealth();
        const interval = setInterval(checkHealth, 5000);
        return () => clearInterval(interval);
    }, []);

    const color =
        pulse === "alive" ? "text-green-500" :
            pulse === "dead" ? "text-red-500" : "text-gray-500";

    return (
        <div className="flex items-center gap-2 px-4 py-2 bg-slate-900 border border-slate-800 rounded-md">
            <Activity className={`w-5 h-5 ${color} ${pulse === "alive" ? "animate-pulse" : ""}`} />
            <span className={`text-xs font-mono font-bold ${color}`}>
                {pulse === "alive" ? "SYSTEM_ONLINE" : "NO_CARRIER"}
            </span>
        </div>
    );
}

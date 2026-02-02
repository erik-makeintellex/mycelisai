"use client";

import { useEffect, useState } from "react";

interface ModelHealthProps {
    modelId: string;
}

export function ModelHealth({ modelId }: ModelHealthProps) {
    const [status, setStatus] = useState<"ok" | "error" | "loading">("loading");

    // Mock health check for now, can be real API later
    // Logic: For MVP, assume "local-qwen" is OK if Pulse is OK.
    useEffect(() => {
        // Determine status based on Model ID or future API
        if (modelId.startsWith("local")) {
            setStatus("ok");
        } else {
            setStatus("ok"); // Assume Cloud is OK too for now
        }
    }, [modelId]);

    const color =
        status === "ok" ? "bg-green-500 shadow-[0_0_8px_#22c55e]" :
            status === "error" ? "bg-red-500 shadow-[0_0_8px_#ef4444]" : "bg-yellow-500";

    return (
        <div className={`w-2 h-2 rounded-full ${color}`} title={status === "ok" ? "Online" : "Offline"}></div>
    );
}

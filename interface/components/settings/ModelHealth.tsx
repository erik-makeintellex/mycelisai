"use client";

import { useEffect, useState } from "react";

interface ModelHealthProps {
    modelId: string;
}

export function ModelHealth({ modelId }: ModelHealthProps) {
    // Mock health check for now, logic is synchronous
    const status = modelId.startsWith("local") ? "ok" : "ok"; // Assume Cloud is OK too for now

    const color =
        status === "ok" ? "bg-green-500 shadow-[0_0_8px_#22c55e]" :
            status === "error" ? "bg-red-500 shadow-[0_0_8px_#ef4444]" : "bg-yellow-500";

    return (
        <div className={`w-2 h-2 rounded-full ${color}`} title={status === "ok" ? "Online" : "Offline"}></div>
    );
}

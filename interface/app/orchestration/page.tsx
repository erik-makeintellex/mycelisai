"use client"

import { Canvas } from "@/components/loom/Canvas"

export default function NetworkPage() {
    return (
        <div className="w-full h-full flex flex-col">
            <div className="h-14 border-b border-[rgb(var(--border))] flex items-center px-6 bg-[rgb(var(--surface))]">
                <h1 className="text-lg font-semibold text-[rgb(var(--foreground))]">The Loom</h1>
                <div className="ml-auto text-xs text-zinc-500 font-mono">
                    Network Topology / Active
                </div>
            </div>
            <div className="flex-1 relative">
                <Canvas />
            </div>
        </div>
    )
}

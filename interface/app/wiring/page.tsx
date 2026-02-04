"use client";

import WireGraph from "@/components/wiring/WireGraph";

export default function WiringPage() {
    return (
        <div className="h-full flex flex-col">
            <header className="h-16 border-b border-zinc-200 bg-white px-6 flex items-center justify-between">
                <h1 className="font-mono font-bold text-zinc-900">Neural Wiring</h1>
                <div className="text-xs text-zinc-500 font-mono">
                    <span className="w-2 h-2 bg-green-500 rounded-full inline-block mr-2" />
                    Live Signal Graph
                </div>
            </header>
            <div className="flex-1 bg-zinc-50 relative">
                <WireGraph />
            </div>
        </div>
    );
}

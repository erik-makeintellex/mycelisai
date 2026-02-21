"use client";

import dynamic from 'next/dynamic';

const Workspace = dynamic(() => import('@/components/workspace/Workspace'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading workspace...</span>
        </div>
    ),
});

export default function WiringPage() {
    return (
        <div className="h-full flex flex-col bg-cortex-bg text-cortex-text-main">
            <header className="px-6 py-3 border-b border-cortex-border flex justify-between items-center flex-shrink-0">
                <div>
                    <h1 className="text-xl font-bold bg-gradient-to-r from-cortex-primary to-cortex-info bg-clip-text text-transparent">
                        Neural Wiring
                    </h1>
                    <p className="text-cortex-text-muted text-xs font-mono mt-0.5">
                        Negotiate intent into executable DAGs
                    </p>
                </div>
                <div className="text-xs text-cortex-text-muted font-mono flex items-center gap-2">
                    <span className="w-2 h-2 bg-cortex-success rounded-full inline-block shadow-[0_0_6px_rgba(16,185,129,0.5)]" />
                    Live Signal Graph
                </div>
            </header>

            <main className="flex-1 overflow-hidden">
                <Workspace />
            </main>
        </div>
    );
}

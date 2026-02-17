"use client";

import dynamic from 'next/dynamic';

const MemoryExplorer = dynamic(() => import('@/components/memory/MemoryExplorer'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading memory explorer...</span>
        </div>
    ),
});

export default function MemoryRoute() {
    return <MemoryExplorer />;
}

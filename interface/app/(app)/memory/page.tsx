"use client";

import dynamic from 'next/dynamic';
import AdvancedModeRoute from "@/components/shared/AdvancedModeRoute";

const MemoryExplorer = dynamic(() => import('@/components/memory/MemoryExplorer'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading memory explorer...</span>
        </div>
    ),
});

export default function MemoryRoute() {
    return (
        <AdvancedModeRoute
            title="Memory is in Admin tools"
            summary="Soma should surface useful learning and artifacts in the work layer. Open Memory when an admin needs deeper retained-context inspection."
        >
            <MemoryExplorer />
        </AdvancedModeRoute>
    );
}

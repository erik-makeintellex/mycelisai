"use client";

import dynamic from 'next/dynamic';
import { useCortexStore } from "@/store/useCortexStore";
import AdvancedModeGate from "@/components/shared/AdvancedModeGate";

const MemoryExplorer = dynamic(() => import('@/components/memory/MemoryExplorer'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading memory explorer...</span>
        </div>
    ),
});

export default function MemoryRoute() {
    const advancedMode = useCortexStore((s) => s.advancedMode);

    if (!advancedMode) {
        return (
            <AdvancedModeGate
                title="Memory is an Advanced support view"
                summary="Soma should surface useful learning and artifacts in the work layer. Open Memory when an admin needs deeper retained-context inspection."
            />
        );
    }

    return <MemoryExplorer />;
}

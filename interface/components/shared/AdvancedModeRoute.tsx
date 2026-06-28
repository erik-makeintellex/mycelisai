"use client";

import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import AdvancedModeGate from "@/components/shared/AdvancedModeGate";
import { useCortexStore } from "@/store/useCortexStore";

export default function AdvancedModeRoute({
    children,
    title,
    summary,
    returnHref,
    returnLabel,
}: {
    children: ReactNode;
    title: string;
    summary: string;
    returnHref?: string;
    returnLabel?: string;
}) {
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);
    const [advancedFromQuery, setAdvancedFromQuery] = useState(false);
    const [hasMounted, setHasMounted] = useState(false);

    useEffect(() => {
        setHasMounted(true);
        const requested = new URLSearchParams(window.location.search).get("advanced") === "1";
        setAdvancedFromQuery(requested);
        if (requested && !advancedMode) {
            toggleAdvancedMode();
        }
    }, [advancedMode, toggleAdvancedMode]);

    if (!hasMounted) {
        return (
            <div className="flex h-full items-center justify-center bg-cortex-bg px-6 py-10">
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm font-medium text-cortex-text-muted">
                    Checking admin tools...
                </div>
            </div>
        );
    }

    if (!advancedMode && !advancedFromQuery) {
        return (
            <AdvancedModeGate
                title={title}
                summary={summary}
                returnHref={returnHref}
                returnLabel={returnLabel}
            />
        );
    }

    return <>{children}</>;
}

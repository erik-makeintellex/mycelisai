"use client";

import { useEffect } from "react";
import type { ReactNode } from "react";
import AdvancedModeGate from "@/components/shared/AdvancedModeGate";
import { useCortexStore } from "@/store/useCortexStore";
import { useBrowserSearch, useClientReady } from "@/lib/browserLocation";

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
    const hasMounted = useClientReady();
    const search = useBrowserSearch();
    const advancedFromQuery = new URLSearchParams(search).get("advanced") === "1";

    useEffect(() => {
        if (advancedFromQuery && !advancedMode) {
            toggleAdvancedMode();
        }
    }, [advancedFromQuery, advancedMode, toggleAdvancedMode]);

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

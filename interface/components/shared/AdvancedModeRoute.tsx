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

    useEffect(() => {
        const requested = new URLSearchParams(window.location.search).get("advanced") === "1";
        setAdvancedFromQuery(requested);
        if (requested && !advancedMode) {
            toggleAdvancedMode();
        }
    }, [advancedMode, toggleAdvancedMode]);

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

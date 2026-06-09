"use client";

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

    if (!advancedMode) {
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

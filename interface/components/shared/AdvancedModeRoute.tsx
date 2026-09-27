"use client";

import { useEffect } from "react";
import type { ReactNode } from "react";
import AdvancedModeGate from "@/components/shared/AdvancedModeGate";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
import { useCortexStore } from "@/store/useCortexStore";
import { useBrowserSearch, useClientReady } from "@/lib/browserLocation";
import { useIsAdmin } from "@/lib/useIsAdmin";

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
    const { isAdmin, checked } = useIsAdmin();

    useEffect(() => {
        if (advancedFromQuery && !advancedMode) {
            toggleAdvancedMode();
        }
    }, [advancedFromQuery, advancedMode, toggleAdvancedMode]);

    if (!hasMounted || !checked) {
        return (
            <div className="flex h-full items-center justify-center bg-cortex-bg px-6 py-10">
                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm font-medium text-cortex-text-muted">
                    Checking admin tools...
                </div>
            </div>
        );
    }

    // A standard user is never told to "Turn on Admin tools" — the toggle
    // is a client-side convenience for admins, not a real permission grant,
    // and it must not be reachable by a non-admin even via ?advanced=1.
    if (!isAdmin) {
        return (
            <div className="flex h-full items-center justify-center bg-cortex-bg px-6 py-10">
                <div className="w-full max-w-xl">
                    <InlineBlockerNotice code="admin_required" httpStatus={403} viewerIsAdmin={false} />
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

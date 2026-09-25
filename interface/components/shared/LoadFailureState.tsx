"use client";

import { AlertTriangle, RefreshCw } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import {
    CORE_RECOVERY_HINT,
    CORE_UNREACHABLE_TITLE,
    selectCoreReachability,
} from "@/lib/coreReachability";

/**
 * Visible load-failure state that replaces an empty state when data could not be fetched.
 * When Core itself is unreachable, it leads with that root cause and the documented recovery command.
 */
export default function LoadFailureState({
    title,
    detail,
    onRetry,
    retrying = false,
    testId = "load-failure-state",
}: {
    title: string;
    detail?: string | null;
    onRetry?: () => void;
    retrying?: boolean;
    testId?: string;
}) {
    const reachability = useCortexStore(selectCoreReachability);
    const coreDown = reachability === "unreachable";

    return (
        <div
            role="alert"
            data-testid={testId}
            className="rounded-2xl border border-cortex-danger/40 bg-cortex-danger/10 px-4 py-4 text-sm"
        >
            <div className="flex items-start gap-3">
                <AlertTriangle className="mt-0.5 h-4 w-4 flex-shrink-0 text-cortex-danger" />
                <div className="min-w-0 flex-1 space-y-1">
                    <p className="font-semibold text-cortex-text-main">{title}</p>
                    {coreDown ? (
                        <p className="text-cortex-text-muted">
                            {CORE_UNREACHABLE_TITLE}. {CORE_RECOVERY_HINT}
                        </p>
                    ) : detail ? (
                        <p className="text-cortex-text-muted">{detail}</p>
                    ) : null}
                </div>
                {onRetry ? (
                    <button
                        type="button"
                        onClick={onRetry}
                        className="flex flex-shrink-0 items-center gap-1 rounded-lg border border-cortex-danger/40 px-3 py-1.5 text-xs font-medium text-cortex-danger hover:bg-cortex-danger/15"
                    >
                        <RefreshCw className={`h-3.5 w-3.5 ${retrying ? "animate-spin" : ""}`} />
                        Retry
                    </button>
                ) : null}
            </div>
        </div>
    );
}

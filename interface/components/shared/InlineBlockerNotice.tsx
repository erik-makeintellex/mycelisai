"use client";

import type { ReactNode } from "react";
import { AlertTriangle } from "lucide-react";
import { blockerCopy, resolveBlockerCopy } from "@/lib/blockerCopy";

/**
 * A small, dependency-free blocker box for non-U1 surfaces (ApprovalsTab,
 * BrainsPage). `OperationalAlert` is U1-owned and does not exist on this
 * branch; this covers the same shape — `role="alert"`, a plain title, one
 * primary action — until U1 lands and callers can migrate.
 */
export default function InlineBlockerNotice({
    code,
    httpStatus,
    viewerIsAdmin = true,
    reason,
    onRetry,
    retryLabel,
    onDismiss,
    children,
}: {
    code?: string;
    httpStatus?: number;
    viewerIsAdmin?: boolean;
    reason?: string;
    onRetry?: () => void;
    retryLabel?: string;
    onDismiss?: () => void;
    children?: ReactNode;
}) {
    // blockerCopy() never resolves viewerIsAdmin itself (per the U1
    // contract); resolve it here so admins actually see the admin variant.
    const copy = resolveBlockerCopy(blockerCopy({ code: code ?? "request_failed", httpStatus, viewerIsAdmin, reason }), viewerIsAdmin);
    return (
        <div role="alert" className="rounded-xl border border-amber-400/30 bg-amber-400/10 p-3">
            <div className="flex items-start gap-2">
                <AlertTriangle aria-hidden="true" className="mt-0.5 h-4 w-4 flex-shrink-0 text-amber-400" />
                <div className="flex-1">
                    <p className="text-sm font-semibold text-cortex-text-main">{copy.title}</p>
                    <p className="mt-0.5 text-xs text-cortex-text-muted">{copy.whatHappened}</p>
                </div>
                {onDismiss ? (
                    <button onClick={onDismiss} className="text-xs text-cortex-text-muted hover:text-cortex-text-main">
                        Dismiss
                    </button>
                ) : null}
            </div>
            {children}
            {onRetry ? (
                <button
                    onClick={onRetry}
                    className="mt-3 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-1.5 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/15"
                >
                    {retryLabel ?? copy.nextAction.label}
                </button>
            ) : null}
        </div>
    );
}

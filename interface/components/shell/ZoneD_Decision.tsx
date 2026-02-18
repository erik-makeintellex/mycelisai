"use client";

import React, { useState, useCallback } from 'react';
import { AlertCircle, Check, X, Loader2 } from 'lucide-react';
import { useSignalStream, type Signal } from '../dashboard/SignalContext';

interface GovernanceRequest {
    request_id: string;
    message: string;
    source?: string;
    timestamp?: string;
    payload: Record<string, unknown>;
}

/** Extract a GovernanceRequest from a governance signal */
function toGovernanceRequest(signal: Signal): GovernanceRequest {
    return {
        request_id: signal.payload?.request_id ?? signal.payload?.id ?? 'unknown',
        message: signal.message ?? signal.payload?.description ?? 'Pending governance action',
        source: signal.source,
        timestamp: signal.timestamp,
        payload: signal.payload ?? {},
    };
}

export function ZoneD() {
    const { signals } = useSignalStream();
    const [dismissed, setDismissed] = useState<Set<string>>(new Set());
    const [loading, setLoading] = useState<string | null>(null);

    // Filter pending governance requests, excluding ones already acted on
    const pendingRequests: GovernanceRequest[] = signals
        .filter(
            (s) =>
                s.type === 'governance' &&
                s.payload?.status === 'pending'
        )
        .map(toGovernanceRequest)
        .filter((req) => !dismissed.has(req.request_id));

    // Take the first (most recent) pending request to display
    const activeRequest = pendingRequests.length > 0 ? pendingRequests[0] : null;

    const handleAction = useCallback(async (requestId: string, approved: boolean) => {
        setLoading(requestId);
        try {
            await fetch(`/admin/approvals/${requestId}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ approved }),
            });
        } catch (err) {
            console.error('Governance action failed:', err);
        } finally {
            // Remove from local state regardless of outcome
            setDismissed((prev) => new Set(prev).add(requestId));
            setLoading(null);
        }
    }, []);

    // If no pending requests, render nothing
    if (!activeRequest) return null;

    const isLoading = loading === activeRequest.request_id;

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[9999] flex items-center justify-center p-4">
            <div className="bg-white dark:bg-[rgb(var(--surface))] rounded-xl shadow-2xl max-w-md w-full overflow-hidden border border-red-100 dark:border-red-900/50">
                {/* Header */}
                <div className="bg-red-50 dark:bg-[rgb(var(--status-critical-bg))] p-4 flex items-center border-b border-red-100 dark:border-red-900/50">
                    <div className="w-10 h-10 bg-red-100 dark:bg-red-900/40 rounded-full flex items-center justify-center mr-3">
                        <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400" />
                    </div>
                    <div>
                        <h3 className="text-sm font-bold text-red-900 dark:text-red-300 uppercase">Governance Intercept</h3>
                        <p className="text-xs text-red-700 dark:text-red-400">Human Approval Required</p>
                    </div>
                    {pendingRequests.length > 1 && (
                        <span className="ml-auto text-[10px] font-bold text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/40 px-2 py-0.5 rounded-full">
                            +{pendingRequests.length - 1} more
                        </span>
                    )}
                </div>

                {/* Body */}
                <div className="p-6">
                    <p className="text-zinc-700 dark:text-zinc-300 font-medium mb-1">
                        Request ID: <span className="font-mono text-sm">{activeRequest.request_id}</span>
                    </p>
                    {activeRequest.source && (
                        <p className="text-[11px] text-zinc-400 mb-2">
                            Source: {activeRequest.source}
                        </p>
                    )}
                    <p className="text-sm text-zinc-500 dark:text-zinc-400 mb-6 bg-zinc-50 dark:bg-zinc-800/50 p-3 rounded border border-zinc-100 dark:border-zinc-700 font-mono">
                        {activeRequest.message}
                    </p>

                    <div className="flex gap-3">
                        <button
                            onClick={() => handleAction(activeRequest.request_id, false)}
                            disabled={isLoading}
                            className="flex-1 flex items-center justify-center py-2.5 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-600 text-zinc-700 dark:text-zinc-300 rounded-lg hover:bg-zinc-50 dark:hover:bg-zinc-700 font-medium transition-colors disabled:opacity-50"
                        >
                            {isLoading ? (
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                            ) : (
                                <X className="w-4 h-4 mr-2" />
                            )}
                            Deny
                        </button>
                        <button
                            onClick={() => handleAction(activeRequest.request_id, true)}
                            disabled={isLoading}
                            className="flex-1 flex items-center justify-center py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium transition-colors shadow-sm disabled:opacity-50"
                        >
                            {isLoading ? (
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                            ) : (
                                <Check className="w-4 h-4 mr-2" />
                            )}
                            Approve
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}

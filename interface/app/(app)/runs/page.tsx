"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Activity, Zap, Loader2 } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";

function timeAgo(iso: string): string {
    const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
    if (diff < 60)   return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    return `${Math.floor(diff / 3600)}h ago`;
}

function statusColor(status: string): string {
    switch (status) {
        case 'running':   return 'text-cortex-primary';
        case 'completed': return 'text-cortex-success';
        case 'failed':    return 'text-cortex-danger';
        default:          return 'text-cortex-text-muted';
    }
}

function statusDot(status: string): string {
    switch (status) {
        case 'running':   return 'bg-cortex-primary animate-pulse';
        case 'completed': return 'bg-cortex-success';
        case 'failed':    return 'bg-cortex-danger';
        default:          return 'bg-cortex-text-muted/40';
    }
}

export default function RunsPage() {
    const router = useRouter();
    const recentRuns = useCortexStore((s) => s.recentRuns);
    const isFetchingRuns = useCortexStore((s) => s.isFetchingRuns);
    const fetchRecentRuns = useCortexStore((s) => s.fetchRecentRuns);

    useEffect(() => {
        fetchRecentRuns();
    }, [fetchRecentRuns]);

    const activeRuns = recentRuns.filter((r) => r.status === 'running');

    return (
        <div className="min-h-screen bg-cortex-bg text-cortex-text-main">
            {/* Header */}
            <div className="sticky top-0 z-10 bg-cortex-bg border-b border-cortex-border px-4 py-3 flex items-center gap-3">
                <a
                    href="/dashboard"
                    className="flex items-center gap-1.5 text-cortex-text-muted hover:text-cortex-primary transition-colors text-[11px] font-mono"
                >
                    <ArrowLeft className="w-3.5 h-3.5" />
                    Workspace
                </a>

                <div className="w-px h-4 bg-cortex-border" />

                <div className="flex items-center gap-2">
                    <Activity className="w-3.5 h-3.5 text-cortex-primary" />
                    <span className="text-[11px] font-mono text-cortex-text-main font-bold uppercase tracking-widest">
                        Runs
                    </span>
                </div>

                {activeRuns.length > 0 && (
                    <span className="flex items-center gap-1 text-[9px] font-mono font-bold px-2 py-0.5 rounded-full bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/30">
                        <span className="w-1.5 h-1.5 rounded-full bg-cortex-primary animate-pulse" />
                        {activeRuns.length} active
                    </span>
                )}
            </div>

            {/* Body */}
            <div className="max-w-2xl mx-auto px-4 py-6">
                {isFetchingRuns && recentRuns.length === 0 && (
                    <div className="flex items-center justify-center py-12 gap-2 text-cortex-text-muted">
                        <Loader2 className="w-4 h-4 animate-spin" />
                        <span className="text-sm font-mono">Loading runs...</span>
                    </div>
                )}

                {!isFetchingRuns && recentRuns.length === 0 && (
                    <div className="flex flex-col items-center justify-center py-12 gap-2">
                        <Activity className="w-10 h-10 text-cortex-text-muted/20" />
                        <p className="text-sm font-mono text-cortex-text-muted">No runs yet</p>
                        <p className="text-[10px] font-mono text-cortex-text-muted/60">
                            Ask Soma to launch a crew to create your first run
                        </p>
                    </div>
                )}

                {recentRuns.length > 0 && (
                    <div className="border border-cortex-border rounded-lg overflow-hidden">
                        {recentRuns.map((run, i) => (
                            <button
                                key={run.id}
                                onClick={() => router.push(`/runs/${run.id}`)}
                                className={`w-full px-4 py-3 flex items-center gap-3 text-left hover:bg-cortex-border/30 transition-colors group ${
                                    i < recentRuns.length - 1 ? 'border-b border-cortex-border/50' : ''
                                }`}
                            >
                                <span className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${statusDot(run.status)}`} />

                                <div className="flex-1 min-w-0">
                                    <div className="text-[11px] font-mono font-bold text-cortex-text-main group-hover:text-cortex-primary transition-colors truncate">
                                        {run.id}
                                    </div>
                                    <div className="text-[9px] font-mono text-cortex-text-muted/60 mt-0.5 truncate">
                                        mission: {run.mission_id.slice(0, 16)}...
                                    </div>
                                </div>

                                <span className={`text-[9px] font-mono font-bold uppercase flex-shrink-0 ${statusColor(run.status)}`}>
                                    {run.status}
                                </span>

                                <span className="text-[9px] font-mono text-cortex-text-muted/60 flex-shrink-0">
                                    {timeAgo(run.started_at)}
                                </span>

                                <Zap className="w-3 h-3 text-cortex-text-muted/30 group-hover:text-cortex-primary transition-colors flex-shrink-0" />
                            </button>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}

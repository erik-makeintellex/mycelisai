"use client";

import dynamic from 'next/dynamic';
import { use, useEffect, useState } from 'react';
import { ArrowLeft, Zap, MessageSquare, List } from 'lucide-react';

const RunTimeline = dynamic(
    () => import('@/components/runs/RunTimeline'),
    { ssr: false }
);

const ConversationLog = dynamic(
    () => import('@/components/runs/ConversationLog'),
    { ssr: false }
);

type Tab = 'conversation' | 'events';

export default function RunPage({ params }: { params: Promise<{ id: string }> }) {
    const { id } = use(params);
    const [activeTab, setActiveTab] = useState<Tab>('conversation');
    const [runStatus, setRunStatus] = useState<string | undefined>(undefined);

    // Fetch run metadata to determine status (for ConversationLog polling + interjection visibility)
    useEffect(() => {
        let cancelled = false;
        const fetchStatus = async () => {
            try {
                const res = await fetch(`/api/v1/runs/${id}/events`);
                if (!res.ok) return;
                const body = await res.json();
                const events = body.data ?? body ?? [];
                const hasCompleted = events.some((e: { event_type: string }) => e.event_type === 'mission.completed');
                const hasFailed = events.some((e: { event_type: string }) => e.event_type === 'mission.failed');
                const hasCancelled = events.some((e: { event_type: string }) => e.event_type === 'mission.cancelled');
                if (!cancelled) {
                    if (hasCompleted) setRunStatus('completed');
                    else if (hasFailed) setRunStatus('failed');
                    else if (hasCancelled) setRunStatus('cancelled');
                    else setRunStatus('running');
                }
            } catch {
                if (!cancelled) setRunStatus('running');
            }
        };
        fetchStatus();
        const interval = setInterval(() => {
            if (runStatus === 'running' || runStatus === undefined) {
                fetchStatus();
            }
        }, 10000);
        return () => {
            cancelled = true;
            clearInterval(interval);
        };
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [id]);

    const tabs: { key: Tab; label: string; icon: React.ReactNode }[] = [
        { key: 'conversation', label: 'Conversation', icon: <MessageSquare className="w-3.5 h-3.5" /> },
        { key: 'events', label: 'Events', icon: <List className="w-3.5 h-3.5" /> },
    ];

    // Events tab renders RunTimeline with its own full layout (header, auto-refresh, etc.)
    if (activeTab === 'events') {
        return (
            <div className="min-h-screen bg-cortex-bg text-cortex-text-main">
                {/* Persistent tab bar above RunTimeline */}
                <div className="sticky top-0 z-20 bg-cortex-bg border-b border-cortex-border px-4 py-2 flex items-center gap-3">
                    <a
                        href="/dashboard"
                        className="flex items-center gap-1.5 text-cortex-text-muted hover:text-cortex-primary transition-colors text-[11px] font-mono"
                    >
                        <ArrowLeft className="w-3.5 h-3.5" />
                        Workspace
                    </a>
                    <div className="w-px h-4 bg-cortex-border" />
                    <div className="flex items-center gap-2">
                        <Zap className="w-3.5 h-3.5 text-cortex-primary" />
                        <span className="text-[11px] font-mono text-cortex-text-main font-bold">
                            Run: {id.slice(0, 8)}...
                        </span>
                    </div>
                    <div className="ml-auto flex items-center gap-0.5 bg-cortex-surface rounded-md border border-cortex-border p-0.5">
                        {tabs.map((tab) => (
                            <button
                                key={tab.key}
                                onClick={() => setActiveTab(tab.key)}
                                className={`flex items-center gap-1.5 px-3 py-1 rounded text-[10px] font-mono font-bold transition-colors ${
                                    activeTab === tab.key
                                        ? 'bg-cortex-primary/15 text-cortex-primary'
                                        : 'text-cortex-text-muted hover:text-cortex-text-main'
                                }`}
                            >
                                {tab.icon}
                                {tab.label}
                            </button>
                        ))}
                    </div>
                </div>
                <RunTimeline runId={id} />
            </div>
        );
    }

    // Conversation tab
    return (
        <div className="min-h-screen bg-cortex-bg text-cortex-text-main">
            {/* Header with tab bar */}
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
                    <Zap className="w-3.5 h-3.5 text-cortex-primary" />
                    <span className="text-[11px] font-mono text-cortex-text-main font-bold">
                        Run: {id.slice(0, 8)}...
                    </span>
                </div>

                {runStatus && (
                    <span className={`text-[9px] font-mono font-bold px-2 py-0.5 rounded-full border ${
                        runStatus === 'completed'
                            ? 'bg-cortex-success/15 text-cortex-success border-cortex-success/30'
                            : runStatus === 'failed'
                                ? 'bg-red-500/15 text-red-400 border-red-500/30'
                                : runStatus === 'cancelled'
                                    ? 'bg-amber-500/15 text-amber-400 border-amber-500/30'
                                    : 'bg-cortex-primary/10 text-cortex-primary border-cortex-primary/30'
                    }`}>
                        {runStatus === 'running' && (
                            <span className="inline-block w-1.5 h-1.5 rounded-full bg-cortex-primary animate-pulse mr-1 align-middle" />
                        )}
                        {runStatus}
                    </span>
                )}

                <div className="ml-auto flex items-center gap-0.5 bg-cortex-surface rounded-md border border-cortex-border p-0.5">
                    {tabs.map((tab) => (
                        <button
                            key={tab.key}
                            onClick={() => setActiveTab(tab.key)}
                            className={`flex items-center gap-1.5 px-3 py-1 rounded text-[10px] font-mono font-bold transition-colors ${
                                activeTab === tab.key
                                    ? 'bg-cortex-primary/15 text-cortex-primary'
                                    : 'text-cortex-text-muted hover:text-cortex-text-main'
                            }`}
                        >
                            {tab.icon}
                            {tab.label}
                        </button>
                    ))}
                </div>
            </div>

            {/* Conversation content */}
            <div className="max-w-2xl mx-auto px-4 py-6">
                <ConversationLog runId={id} runStatus={runStatus} />
            </div>
        </div>
    );
}

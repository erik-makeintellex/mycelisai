"use client";

import React, { useEffect, useRef, useState } from "react";
import { ArrowLeft, Loader2, RefreshCw, Zap } from "lucide-react";
import type { MissionEvent } from "@/store/useCortexStore";
import EventCard from "./EventCard";

// ── Terminal event types — stop polling ───────────────────────

const TERMINAL_EVENTS = new Set([
    "mission.completed",
    "mission.failed",
    "mission.cancelled",
]);

// ── Status badge ──────────────────────────────────────────────

function StatusBadge({ events }: { events: MissionEvent[] }) {
    const last = events[events.length - 1];
    if (!last) return null;

    if (last.event_type === "mission.completed") {
        return (
            <span className="text-[9px] font-mono font-bold px-2 py-0.5 rounded-full bg-cortex-success/15 text-cortex-success border border-cortex-success/30">
                completed
            </span>
        );
    }
    if (last.event_type === "mission.failed") {
        return (
            <span className="text-[9px] font-mono font-bold px-2 py-0.5 rounded-full bg-cortex-danger/15 text-cortex-danger border border-cortex-danger/30">
                failed
            </span>
        );
    }
    const isTerminal = events.some((e) => TERMINAL_EVENTS.has(e.event_type));
    if (isTerminal) {
        return (
            <span className="text-[9px] font-mono font-bold px-2 py-0.5 rounded-full bg-cortex-border text-cortex-text-muted border border-cortex-border">
                stopped
            </span>
        );
    }
    return (
        <span className="flex items-center gap-1.5 text-[9px] font-mono font-bold px-2 py-0.5 rounded-full bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/30">
            <span className="w-1.5 h-1.5 rounded-full bg-cortex-primary animate-pulse" />
            running
        </span>
    );
}

// ── RunTimeline ───────────────────────────────────────────────

interface Props {
    runId: string;
}

export default function RunTimeline({ runId }: Props) {
    const [events, setEvents] = useState<MissionEvent[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [startedAt, setStartedAt] = useState<string | null>(null);
    const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

    const isTerminal = events.some((e) => TERMINAL_EVENTS.has(e.event_type));

    const fetchEvents = async () => {
        try {
            const res = await fetch(`/api/v1/runs/${runId}/events`);
            if (!res.ok) {
                setError(`Failed to load events (${res.status})`);
                return;
            }
            const body = await res.json();
            const list: MissionEvent[] = body.data ?? body ?? [];
            setEvents(list);
            if (list.length > 0 && !startedAt) {
                setStartedAt(list[0].emitted_at);
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Network error');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchEvents();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [runId]);

    // Auto-poll every 5s while not terminal
    useEffect(() => {
        if (isTerminal) {
            if (pollingRef.current) clearInterval(pollingRef.current);
            return;
        }
        pollingRef.current = setInterval(fetchEvents, 5000);
        return () => {
            if (pollingRef.current) clearInterval(pollingRef.current);
        };
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isTerminal, runId]);

    function startedAgo(): string {
        if (!startedAt) return '';
        const diff = Math.floor((Date.now() - new Date(startedAt).getTime()) / 1000);
        if (diff < 60)   return `started ${diff}s ago`;
        if (diff < 3600) return `started ${Math.floor(diff / 60)}m ago`;
        return `started ${Math.floor(diff / 3600)}h ago`;
    }

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
                    <Zap className="w-3.5 h-3.5 text-cortex-primary" />
                    <span className="text-[11px] font-mono text-cortex-text-main font-bold">
                        Run: {runId.slice(0, 8)}...
                    </span>
                </div>

                {events.length > 0 && <StatusBadge events={events} />}

                {startedAt && (
                    <span className="text-[10px] font-mono text-cortex-text-muted/60 ml-1">
                        {startedAgo()}
                    </span>
                )}

                <div className="ml-auto flex items-center gap-2">
                    {!isTerminal && !loading && (
                        <span className="text-[9px] font-mono text-cortex-primary/60 flex items-center gap-1">
                            <RefreshCw className="w-2.5 h-2.5 animate-spin" style={{ animationDuration: '3s' }} />
                            auto-refresh
                        </span>
                    )}
                    <button
                        onClick={fetchEvents}
                        className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-primary transition-colors"
                        title="Refresh events"
                    >
                        <RefreshCw className="w-3.5 h-3.5" />
                    </button>
                </div>
            </div>

            {/* Body */}
            <div className="max-w-2xl mx-auto px-4 py-6">
                {loading && (
                    <div className="flex items-center justify-center py-12 gap-2 text-cortex-text-muted">
                        <Loader2 className="w-4 h-4 animate-spin" />
                        <span className="text-sm font-mono">Loading events...</span>
                    </div>
                )}

                {error && !loading && (
                    <div className="flex items-center justify-center py-12">
                        <div className="text-center space-y-2">
                            <p className="text-sm font-mono text-cortex-danger">{error}</p>
                            <button
                                onClick={fetchEvents}
                                className="text-[10px] font-mono text-cortex-primary hover:underline"
                            >
                                Retry
                            </button>
                        </div>
                    </div>
                )}

                {!loading && !error && events.length === 0 && (
                    <div className="flex flex-col items-center justify-center py-12 gap-2 text-cortex-text-muted">
                        <Zap className="w-8 h-8 opacity-20" />
                        <p className="text-sm font-mono">No events yet — mission may still be starting up</p>
                        <p className="text-[10px] font-mono opacity-60">Auto-refreshing every 5s</p>
                    </div>
                )}

                {!loading && events.length > 0 && (
                    <div>
                        {events.map((event, i) => (
                            <EventCard
                                key={event.id}
                                event={event}
                                isLast={i === events.length - 1}
                            />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}

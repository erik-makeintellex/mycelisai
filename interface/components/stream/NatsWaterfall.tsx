"use client";

import React, { useRef, useEffect, useState } from 'react';
import { Activity, Radio, ChevronUp, ChevronDown, ArrowDownLeft, ArrowUpRight, RotateCw } from 'lucide-react';
import { useCortexStore, type StreamSignal } from '@/store/useCortexStore';
import { streamSignalToDetail } from '@/lib/signalNormalize';

// ── Signal Direction Classification ─────────────────────────

type SignalDirection = 'input' | 'output' | 'internal';

/** Classify a signal as input (ingress), output (egress), or internal */
function classifyDirection(type?: string): SignalDirection {
    switch (type) {
        // Inputs: data flowing INTO the system
        case 'heartbeat':
        case 'sensor_data':
        case 'user_input':
        case 'command':
        case 'connected':
            return 'input';

        // Outputs: data flowing OUT of the system
        case 'artifact':
        case 'output':
        case 'tool_call':
        case 'actuation':
        case 'task_complete':
            return 'output';

        // Internal: processing within the system
        case 'thought':
        case 'cognitive':
        case 'error':
        case 'governance':
        case 'governance_halt':
        case 'memory':
        case 'intent':
        default:
            return 'internal';
    }
}

const DIRECTION_ICONS: Record<SignalDirection, React.ElementType> = {
    input: ArrowDownLeft,
    output: ArrowUpRight,
    internal: RotateCw,
};

const DIRECTION_COLORS: Record<SignalDirection, string> = {
    input: 'text-cyan-400',
    output: 'text-cortex-success',
    internal: 'text-cortex-text-muted',
};

type FilterMode = 'all' | 'input' | 'output' | 'internal';

/** Spectrum color rules per signal type */
function spectrumColor(type?: string): { dot: string; text: string; glow: string } {
    switch (type) {
        case 'connected':
        case 'heartbeat':
            return { dot: 'bg-cortex-text-muted', text: 'text-cortex-text-muted', glow: '' };
        case 'thought':
        case 'intent':
        case 'cognitive':
            return {
                dot: 'bg-cortex-info',
                text: 'text-cortex-info',
                glow: 'shadow-[0_0_8px_rgba(0,207,232,0.3)]',
            };
        case 'artifact':
        case 'output':
        case 'task_complete':
            return {
                dot: 'bg-cortex-success',
                text: 'text-cortex-success',
                glow: 'shadow-[0_0_8px_rgba(40,199,111,0.3)]',
            };
        case 'error':
            return {
                dot: 'bg-cortex-danger',
                text: 'text-cortex-danger',
                glow: 'shadow-[0_0_8px_rgba(234,84,85,0.4)]',
            };
        case 'governance':
        case 'governance_halt':
            return {
                dot: 'bg-cortex-warning',
                text: 'text-cortex-warning',
                glow: 'shadow-[0_0_8px_rgba(255,159,67,0.3)]',
            };
        case 'memory':
            return { dot: 'bg-cortex-primary', text: 'text-cortex-primary', glow: '' };
        case 'tool_call':
        case 'actuation':
            return {
                dot: 'bg-cortex-success',
                text: 'text-cortex-success',
                glow: '',
            };
        case 'sensor_data':
        case 'user_input':
        case 'command':
            return {
                dot: 'bg-cyan-400',
                text: 'text-cyan-400',
                glow: '',
            };
        default:
            return { dot: 'bg-cortex-text-muted', text: 'text-cortex-text-muted', glow: '' };
    }
}

/** Relative timestamp label */
function formatTime(timestamp?: string): string {
    if (!timestamp) return 'now';
    try {
        const date = new Date(timestamp);
        if (isNaN(date.getTime())) return timestamp;
        const diffSec = Math.floor((Date.now() - date.getTime()) / 1000);
        if (diffSec < 5) return 'now';
        if (diffSec < 60) return `${diffSec}s`;
        if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m`;
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
        return timestamp;
    }
}

function SignalRow({ signal }: { signal: StreamSignal }) {
    const selectSignalDetail = useCortexStore((s) => s.selectSignalDetail);
    const time = formatTime(signal.timestamp);
    const source = signal.source ?? 'system';
    const message = signal.message ?? JSON.stringify(signal.payload ?? {});
    const colors = spectrumColor(signal.type);
    const direction = classifyDirection(signal.type);
    const DirIcon = DIRECTION_ICONS[direction];
    const dirColor = DIRECTION_COLORS[direction];

    return (
        <div
            className={`flex items-center gap-2 px-4 py-1.5 hover:bg-cortex-surface/50 transition-colors font-mono cursor-pointer ${colors.glow}`}
            onClick={() => selectSignalDetail(streamSignalToDetail(signal))}
        >
            {/* Direction indicator */}
            <DirIcon className={`w-3 h-3 flex-shrink-0 ${dirColor}`} />

            {/* Signal type dot */}
            <span className={`inline-block w-1.5 h-1.5 rounded-full flex-shrink-0 ${colors.dot}`} />

            {/* Type label */}
            <span className={`text-[10px] font-bold uppercase tracking-wide w-16 flex-shrink-0 ${colors.text}`}>
                {signal.type ?? 'event'}
            </span>

            {/* Source */}
            <span className="text-[10px] text-cortex-text-muted w-24 flex-shrink-0 truncate">
                {source}
            </span>

            {/* Message */}
            <span className="text-[11px] text-cortex-text-main flex-1 truncate" title={message}>
                {message}
            </span>

            {/* Timestamp */}
            <span className="text-[9px] text-cortex-text-muted flex-shrink-0 w-10 text-right">
                {time}
            </span>
        </div>
    );
}

export function NatsWaterfall() {
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const isConnected = useCortexStore((s) => s.isStreamConnected);
    const initializeStream = useCortexStore((s) => s.initializeStream);
    const scrollRef = useRef<HTMLDivElement>(null);
    const [isExpanded, setIsExpanded] = useState(true);
    const [filter, setFilter] = useState<FilterMode>('all');

    // Initialize SSE on mount
    useEffect(() => {
        initializeStream();
    }, [initializeStream]);

    // Auto-scroll to top on new signals
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = 0;
        }
    }, [streamLogs.length]);

    // Apply direction filter
    const filteredLogs = filter === 'all'
        ? streamLogs
        : streamLogs.filter((s) => classifyDirection(s.type) === filter);

    // Counts for filter badges
    const inputCount = streamLogs.filter((s) => classifyDirection(s.type) === 'input').length;
    const outputCount = streamLogs.filter((s) => classifyDirection(s.type) === 'output').length;
    const errorCount = streamLogs.filter((s) => s.type === 'error').length;

    return (
        <div className={`border-t border-cortex-border bg-cortex-bg flex flex-col flex-shrink-0 transition-all duration-200 ${
            isExpanded ? 'h-52' : 'h-9'
        }`}>
            {/* Header bar — always visible, acts as toggle */}
            <div className="h-9 flex items-center px-4 bg-cortex-surface/60 flex-shrink-0 w-full">
                <button
                    onClick={() => setIsExpanded(!isExpanded)}
                    className="flex items-center hover:bg-cortex-surface transition-colors rounded px-1 -ml-1"
                >
                    <Radio className="w-3.5 h-3.5 text-cortex-text-muted mr-2" />
                    <span className="text-[10px] font-bold uppercase tracking-widest text-cortex-text-muted">
                        Spectrum
                    </span>
                    <span className="ml-2 text-cortex-text-muted">
                        {isExpanded ? (
                            <ChevronDown className="w-3.5 h-3.5" />
                        ) : (
                            <ChevronUp className="w-3.5 h-3.5" />
                        )}
                    </span>
                </button>

                <span className="ml-3 flex items-center gap-1.5">
                    <span
                        className={`inline-block w-1.5 h-1.5 rounded-full transition-all ${
                            isConnected
                                ? 'bg-cortex-success shadow-[0_0_4px_rgba(40,199,111,0.5)]'
                                : 'bg-cortex-danger shadow-[0_0_4px_rgba(234,84,85,0.5)]'
                        }`}
                    />
                    <span className="text-[9px] font-mono text-cortex-text-muted">
                        {isConnected ? 'LIVE' : 'OFF'}
                    </span>
                </span>

                {/* I/O Direction Filters */}
                {isExpanded && (
                    <div className="ml-4 flex items-center gap-1">
                        {([
                            { mode: 'all' as FilterMode, label: 'ALL', count: streamLogs.length },
                            { mode: 'input' as FilterMode, label: 'IN', count: inputCount },
                            { mode: 'output' as FilterMode, label: 'OUT', count: outputCount },
                            { mode: 'internal' as FilterMode, label: 'INT', count: streamLogs.length - inputCount - outputCount },
                        ] as const).map(({ mode, label, count }) => (
                            <button
                                key={mode}
                                onClick={() => setFilter(mode)}
                                className={`text-[9px] font-mono px-1.5 py-0.5 rounded transition-colors ${
                                    filter === mode
                                        ? 'bg-cortex-primary/20 text-cortex-primary'
                                        : 'text-cortex-text-muted hover:text-cortex-text-main'
                                }`}
                            >
                                {label} {count > 0 && <span className="opacity-60">{count}</span>}
                            </button>
                        ))}
                    </div>
                )}

                {errorCount > 0 && (
                    <span className="text-[9px] font-mono text-cortex-danger ml-auto">
                        {errorCount} err
                    </span>
                )}
            </div>

            {/* Signal list — scrollable rows */}
            {isExpanded && (
                <div ref={scrollRef} className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border min-h-0">
                    {filteredLogs.length === 0 ? (
                        <div className="flex items-center justify-center h-full text-cortex-text-muted">
                            <Activity className="w-4 h-4 mr-2 opacity-30" />
                            <span className="text-[10px] font-mono">
                                {filter === 'all' ? 'Awaiting signals...' : `No ${filter} signals`}
                            </span>
                        </div>
                    ) : (
                        <div className="divide-y divide-cortex-border/30">
                            {filteredLogs.map((signal, index) => (
                                <SignalRow key={`${signal.timestamp}-${index}`} signal={signal} />
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

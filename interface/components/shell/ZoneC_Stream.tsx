"use client";

import React, { useRef, useEffect } from 'react';
import { Activity, Radio } from 'lucide-react';
import { useSignalStream, type Signal } from '../dashboard/SignalContext';

/** Spectrum color rules per signal type */
function spectrumColor(type?: string): { dot: string; text: string; glow: string } {
    switch (type) {
        case 'connected':
        case 'heartbeat':
            return {
                dot: 'bg-blue-400/60',
                text: 'text-blue-300',
                glow: '',
            };
        case 'thought':
        case 'intent':
        case 'cognitive':
            return {
                dot: 'bg-cyan-400',
                text: 'text-cyan-300',
                glow: 'shadow-[0_0_8px_rgba(34,211,238,0.3)]',
            };
        case 'artifact':
        case 'output':
            return {
                dot: 'bg-green-400',
                text: 'text-green-300',
                glow: 'shadow-[0_0_8px_rgba(74,222,128,0.3)]',
            };
        case 'error':
            return {
                dot: 'bg-red-500',
                text: 'text-red-400',
                glow: 'shadow-[0_0_8px_rgba(239,68,68,0.4)]',
            };
        case 'governance':
            return {
                dot: 'bg-amber-500',
                text: 'text-amber-400',
                glow: 'shadow-[0_0_8px_rgba(245,158,11,0.3)]',
            };
        case 'memory':
            return {
                dot: 'bg-purple-400',
                text: 'text-purple-300',
                glow: '',
            };
        default:
            return {
                dot: 'bg-zinc-500',
                text: 'text-zinc-400',
                glow: '',
            };
    }
}

/** Format a timestamp into a relative label */
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

export function ZoneC() {
    const { isConnected, signals } = useSignalStream();
    const scrollRef = useRef<HTMLDivElement>(null);

    // Auto-scroll to top when new signals arrive
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = 0;
        }
    }, [signals.length]);

    return (
        <div className="hidden lg:flex w-80 bg-zinc-950 border-l border-zinc-800 flex-col z-40 flex-shrink-0">
            {/* Header */}
            <div className="h-14 flex items-center px-4 border-b border-zinc-800 bg-zinc-950">
                <Radio className="w-4 h-4 text-zinc-500 mr-2" />
                <span className="text-[11px] font-bold uppercase tracking-widest text-zinc-500">
                    Spectrum
                </span>
                <span className="ml-auto flex items-center gap-1.5">
                    <span
                        className={`inline-block w-2 h-2 rounded-full transition-all ${
                            isConnected
                                ? 'bg-green-500 shadow-[0_0_6px_rgba(34,197,94,0.5)] animate-pulse'
                                : 'bg-red-500 shadow-[0_0_6px_rgba(239,68,68,0.5)]'
                        }`}
                    />
                    <span className="text-[10px] font-mono text-zinc-500">
                        {isConnected ? 'LIVE' : 'OFF'}
                    </span>
                </span>
            </div>

            {/* Signal count bar */}
            <div className="px-4 py-1.5 border-b border-zinc-800/50 flex items-center justify-between">
                <span className="text-[9px] font-mono text-zinc-600 uppercase">
                    {signals.length} signals captured
                </span>
                {signals.filter(s => s.type === 'error').length > 0 && (
                    <span className="text-[9px] font-mono text-red-500">
                        {signals.filter(s => s.type === 'error').length} err
                    </span>
                )}
            </div>

            {/* Stream List */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-800">
                {signals.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-zinc-600">
                        <Activity className="w-8 h-8 mb-2 opacity-20" />
                        <p className="text-[10px] font-mono">Awaiting signals...</p>
                    </div>
                ) : (
                    <div className="divide-y divide-zinc-800/50">
                        {signals.map((signal, index) => (
                            <SpectrumItem key={`${signal.timestamp}-${index}`} signal={signal} />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}

function SpectrumItem({ signal }: { signal: Signal }) {
    const time = formatTime(signal.timestamp);
    const source = signal.source ?? 'system';
    const message = signal.message ?? JSON.stringify(signal.payload ?? {});
    const colors = spectrumColor(signal.type);

    return (
        <div className={`px-4 py-2.5 hover:bg-zinc-900/50 transition-colors ${colors.glow}`}>
            {/* Top row: type dot + source + time */}
            <div className="flex items-center gap-2 mb-1">
                <span className={`inline-block w-1.5 h-1.5 rounded-full flex-shrink-0 ${colors.dot}`} />
                <span className={`text-[10px] font-mono font-bold uppercase tracking-wide ${colors.text}`}>
                    {signal.type ?? 'event'}
                </span>
                <span className="text-[9px] font-mono text-zinc-600 ml-auto flex-shrink-0">
                    {source}
                </span>
                <span className="text-[9px] font-mono text-zinc-700 flex-shrink-0">
                    {time}
                </span>
            </div>
            {/* Message body */}
            <p className="text-[11px] text-zinc-400 leading-snug pl-3.5 font-mono truncate" title={message}>
                {message}
            </p>
        </div>
    );
}

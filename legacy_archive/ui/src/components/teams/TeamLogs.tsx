'use client';

import { useRef, useEffect } from 'react';
import { useEventStream } from '@/hooks/useEventStream';

interface TeamLogsProps {
    channel: string;
}

export default function TeamLogs({ channel }: TeamLogsProps) {
    const { events, stats } = useEventStream(channel);
    const scrollRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [events]);

    return (
        <div className="flex flex-col h-full bg-zinc-950 rounded-lg border border-zinc-800 overflow-hidden font-mono text-sm">
            <div className="flex items-center justify-between px-4 py-2 bg-zinc-900 border-b border-zinc-800">
                <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${stats.isConnected ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`} />
                    <span className="text-zinc-400">Live Logs: {channel}</span>
                </div>
                <div className="text-xs text-zinc-500">
                    {stats.eventsPerSecond} eps | {stats.totalEvents} total
                </div>
            </div>
            <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-2">
                {events.length === 0 && (
                    <div className="text-zinc-600 italic">Waiting for events...</div>
                )}
                {events.map((event, i) => (
                    <div key={i} className="flex gap-2">
                        <span className="text-zinc-500 shrink-0">[{new Date(event.timestamp * 1000).toLocaleTimeString()}]</span>
                        <span className="text-blue-400 shrink-0">{event.source}</span>
                        <span className="text-zinc-300 break-all">
                            {typeof event.payload === 'string' ? event.payload : JSON.stringify(event.payload)}
                        </span>
                    </div>
                ))}
            </div>
        </div>
    );
}

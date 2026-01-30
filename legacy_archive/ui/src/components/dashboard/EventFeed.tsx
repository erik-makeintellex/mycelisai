'use client';

import { motion, AnimatePresence } from 'framer-motion';
import { Terminal } from 'lucide-react';

interface Event {
    timestamp: number;
    event_type: string;
    payload: any;
}

interface EventFeedProps {
    events: Event[];
    channel?: string;
}

export default function EventFeed({ events, channel = 'sensors' }: EventFeedProps) {
    const formatPayload = (payload: any) => {
        try {
            return JSON.stringify(payload, null, 2);
        } catch {
            return String(payload);
        }
    };

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="border border-[--border-subtle] rounded-xl bg-[#0d1117] shadow-lg overflow-hidden"
        >
            <div className="p-4 border-b border-[--border-subtle] bg-[--bg-secondary] flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <Terminal className="h-5 w-5 text-[--accent-info]" />
                    <h3 className="text-lg font-semibold text-[--text-primary]">Live Event Feed</h3>
                </div>
                <span className="text-xs text-[--text-secondary] font-mono bg-[--accent-info]/10 px-2 py-1 rounded">
                    {channel}
                </span>
            </div>
            <div className="h-96 overflow-y-auto p-4 space-y-1 font-mono text-sm">
                {events.length === 0 ? (
                    <p className="text-[--text-secondary] italic">Waiting for events...</p>
                ) : (
                    <AnimatePresence initial={false}>
                        {events.map((event, idx) => (
                            <motion.div
                                key={`${event.timestamp}-${idx}`}
                                initial={{ opacity: 0, x: -20 }}
                                animate={{ opacity: 1, x: 0 }}
                                exit={{ opacity: 0 }}
                                transition={{ duration: 0.3 }}
                                className="flex gap-4 text-[--text-secondary] py-2 border-b border-[--border-subtle]/30 last:border-0 hover:bg-[--bg-secondary]/30 transition-colors"
                            >
                                <span className="text-[--text-secondary] opacity-70 min-w-[80px]">
                                    {new Date(event.timestamp * 1000).toLocaleTimeString()}
                                </span>
                                <span className="text-[--accent-success] font-bold min-w-[120px]">
                                    {event.event_type}
                                </span>
                                <pre className="text-[--text-mono] text-xs overflow-x-auto flex-1">
                                    {formatPayload(event.payload)}
                                </pre>
                            </motion.div>
                        ))}
                    </AnimatePresence>
                )}
            </div>
        </motion.div>
    );
}

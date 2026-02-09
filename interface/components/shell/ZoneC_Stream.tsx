import React from 'react';
import { Activity } from 'lucide-react';

export function ZoneC() {
    // Placeholder mock data - Real version will pull from NATS/UniversalRenderer
    return (
        <div className="hidden lg:flex w-80 bg-zinc-50 border-l border-zinc-200 flex-col z-40 flex-shrink-0">
            {/* Header */}
            <div className="h-14 flex items-center px-4 border-b border-zinc-200 bg-white">
                <Activity className="w-4 h-4 text-zinc-400 mr-2" />
                <span className="text-xs font-semibold uppercase tracking-wider text-zinc-500">Activity Stream</span>
            </div>

            {/* Stream List */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                <StreamItem time="10:42 AM" source="Cognitive" message="Analyzed user intent: PROVISION_INFRA" />
                <StreamItem time="10:41 AM" source="Memory" message="Projected event to Cortex DB" />
                <StreamItem time="10:40 AM" source="Guard" message="ALLOWED: device.boot (ghost-01)" />
            </div>
        </div>
    );
}

function StreamItem({ time, source, message }: { time: string; source: string; message: string }) {
    return (
        <div className="flex flex-col gap-1">
            <div className="flex items-center justify-between">
                <span className="text-[10px] font-mono text-zinc-400">{time}</span>
                <span className="text-[10px] font-bold text-indigo-600 bg-indigo-50 px-1 py-0.5 rounded">{source}</span>
            </div>
            <p className="text-xs text-zinc-600 leading-snug font-medium">
                {message}
            </p>
        </div>
    );
}

"use client";
import React from 'react';
import { UniversalRenderer } from '../shared/UniversalRenderer';
import { MOCK_STREAM } from '@/lib/mock';

export const ActivityStream = () => {
    return (
        <aside className="w-80 border-l border-zinc-200 bg-zinc-50 flex flex-col h-full">
            <div className="h-14 border-b border-zinc-200 flex items-center px-4 bg-white">
                <h2 className="text-xs font-bold uppercase tracking-wider text-zinc-500">Activity Stream</h2>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {MOCK_STREAM.map((env, i) => (
                    <div key={i} className="flex flex-col gap-1">
                        <div className="flex items-center space-x-2">
                            <span className="text-[10px] bg-zinc-200 px-1 rounded text-zinc-600 font-mono">{env.source}</span>
                            <span className="text-[10px] text-zinc-400">{new Date(env.timestamp).toLocaleTimeString()}</span>
                        </div>
                        <UniversalRenderer message={env} />
                    </div>
                ))}
            </div>
        </aside>
    );
};

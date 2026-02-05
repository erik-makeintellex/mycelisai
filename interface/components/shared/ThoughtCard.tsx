import React from 'react';
import { ThoughtContent } from '@/lib/types/protocol';

export const ThoughtCard: React.FC<{ data: ThoughtContent }> = ({ data }) => {
    return (
        <div className="border-l-2 border-zinc-300 pl-3 py-1 my-2">
            <div className="flex justify-between items-center text-xs text-zinc-500 mb-1">
                <span className="font-mono">Thinking...</span>
                <span className="bg-zinc-100 px-1 rounded text-[10px]">{data.model}</span>
            </div>
            <p className="text-sm font-medium text-zinc-700">{data.summary}</p>
            {data.detail && (
                <p className="text-xs text-zinc-500 mt-1">{data.detail}</p>
            )}
        </div>
    );
};

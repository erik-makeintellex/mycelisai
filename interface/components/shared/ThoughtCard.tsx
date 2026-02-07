import React, { useState } from 'react';
import { Brain, ChevronDown, ChevronRight, Zap } from 'lucide-react';
import { ThoughtContent } from '@/lib/types/protocol';

export function ThoughtCard({ content }: { content: ThoughtContent }) {
    const [expanded, setExpanded] = useState(false);

    return (
        <div className="border border-indigo-100 bg-indigo-50/50 rounded-lg p-3 hover:border-indigo-200 transition-colors">
            <div
                className="flex items-center cursor-pointer select-none"
                onClick={() => setExpanded(!expanded)}
            >
                <div className="mr-2 text-indigo-500">
                    {expanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                </div>
                <Brain size={16} className="text-indigo-600 mr-2" />
                <span className="text-sm font-semibold text-indigo-900 flex-1 truncate">
                    {content.summary}
                </span>
                <span className="text-[10px] uppercase font-bold text-indigo-400 bg-white px-1.5 py-0.5 rounded border border-indigo-100">
                    {content.model}
                </span>
            </div>

            {expanded && (
                <div className="mt-2 pl-8 text-xs text-indigo-800 font-mono whitespace-pre-wrap leading-relaxed border-l-2 border-indigo-200 ml-1.5">
                    {content.detail}
                </div>
            )}
        </div>
    );
}

import React from 'react';
import { MetricContent } from '@/lib/types/protocol';

export const MetricPill: React.FC<{ data: MetricContent }> = ({ data }) => {
    const statusColors = {
        nominal: 'bg-emerald-50 text-emerald-700 border-emerald-200',
        warning: 'bg-amber-50 text-amber-700 border-amber-200',
        critical: 'bg-rose-50 text-rose-700 border-rose-200',
    };

    return (
        <div className={`flex items-center justify-between px-3 py-2 rounded border ${statusColors[data.status] || 'bg-zinc-50 border-zinc-200'} my-1`}>
            <span className="text-xs font-medium">{data.label}</span>
            <div className="flex items-baseline space-x-1">
                <span className="text-sm font-bold">{data.value}</span>
                <span className="text-[10px] opacity-70">{data.unit}</span>
            </div>
        </div>
    );
};

import React from 'react';
import { MetricContent } from '@/lib/types/protocol';
import { Gauge, AlertTriangle, CheckCircle } from 'lucide-react';

export function MetricPill({ content }: { content: MetricContent }) {
    const statusColors = {
        nominal: "bg-emerald-50 text-emerald-700 border-emerald-100",
        warning: "bg-amber-50 text-amber-700 border-amber-100",
        critical: "bg-red-50 text-red-700 border-red-100"
    };

    const Icon = content.status === 'nominal' ? CheckCircle :
        content.status === 'warning' ? AlertTriangle : Gauge;

    return (
        <div className={`
        inline-flex items-center gap-2 px-3 py-1.5 rounded-full border
        shadow-sm hover:shadow-md transition-all
        ${statusColors[content.status]}
    `}>
            <Icon size={14} />
            <span className="text-xs font-medium text-opacity-80">{content.label}</span>
            <span className="text-sm font-bold ml-1">{content.value}</span>
            <span className="text-[10px] uppercase opacity-75">{content.unit}</span>
        </div>
    );
}

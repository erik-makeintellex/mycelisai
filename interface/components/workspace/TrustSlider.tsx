"use client";

import React from 'react';
import { Shield } from 'lucide-react';
import { useCortexStore } from '@/store/useCortexStore';

export default function TrustSlider() {
    const threshold = useCortexStore((s) => s.trustThreshold);
    const setThreshold = useCortexStore((s) => s.setTrustThreshold);
    const isSyncing = useCortexStore((s) => s.isSyncingThreshold);

    const levelLabel = threshold >= 0.8 ? 'STRICT' : threshold >= 0.5 ? 'MODERATE' : 'PERMISSIVE';
    const levelColor = threshold >= 0.8 ? 'text-cortex-danger' : threshold >= 0.5 ? 'text-cortex-warning' : 'text-cortex-success';

    return (
        <div className="px-3 py-2 border-b border-cortex-border bg-cortex-bg/50 flex items-center gap-2">
            <Shield className={`w-3.5 h-3.5 flex-shrink-0 ${isSyncing ? 'text-cortex-warning animate-pulse' : 'text-cortex-text-muted'}`} />
            <span className="text-[9px] font-mono text-cortex-text-muted uppercase whitespace-nowrap">
                Autonomy
            </span>
            <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={threshold}
                onChange={(e) => setThreshold(parseFloat(e.target.value))}
                className="flex-1 h-1 accent-cortex-primary cursor-pointer"
            />
            <span className="text-[10px] font-mono text-cortex-text-main w-7 text-right tabular-nums">
                {threshold.toFixed(2)}
            </span>
            <span className={`text-[8px] font-mono font-bold uppercase ${levelColor}`}>
                {levelLabel}
            </span>
        </div>
    );
}

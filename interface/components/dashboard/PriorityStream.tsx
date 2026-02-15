"use client";

import React from 'react';
import { AlertTriangle, CheckCircle, ShieldAlert, Package } from 'lucide-react';
import { useSignalStream } from './SignalContext';

const PRIORITY_TYPES = new Set(['governance_halt', 'error', 'task_complete', 'artifact']);

const typeConfig: Record<string, { icon: React.ComponentType<{ className?: string }>; color: string; label: string }> = {
    governance_halt: { icon: ShieldAlert, color: 'text-cortex-warning', label: 'GOVERNANCE' },
    error: { icon: AlertTriangle, color: 'text-cortex-danger', label: 'ERROR' },
    task_complete: { icon: CheckCircle, color: 'text-cortex-success', label: 'COMPLETE' },
    artifact: { icon: Package, color: 'text-cortex-info', label: 'ARTIFACT' },
};

export default function PriorityStream() {
    const { signals } = useSignalStream();

    const prioritySignals = signals.filter((s) => PRIORITY_TYPES.has(s.type));

    return (
        <div className="h-full flex flex-col bg-cortex-surface" data-testid="priority-stream">
            {/* Header */}
            <div className="px-3 py-2.5 border-b border-cortex-border flex items-center gap-2">
                <ShieldAlert className="w-3.5 h-3.5 text-cortex-warning" />
                <span className="text-[10px] font-mono font-bold uppercase text-cortex-text-muted tracking-wider">
                    Priority Stream
                </span>
                <span className="ml-auto text-[9px] font-mono text-cortex-text-muted">
                    {prioritySignals.length}
                </span>
            </div>

            {/* Signal list */}
            <div className="flex-1 overflow-y-auto">
                {prioritySignals.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-32 text-cortex-text-muted">
                        <ShieldAlert className="w-6 h-6 mb-2 opacity-20" />
                        <p className="text-[10px] font-mono">No priority events</p>
                    </div>
                ) : (
                    prioritySignals.map((signal, i) => {
                        const config = typeConfig[signal.type] ?? typeConfig.error;
                        const Icon = config.icon;
                        const relTime = signal.timestamp
                            ? new Date(signal.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
                            : '';

                        return (
                            <div
                                key={`${signal.type}-${signal.timestamp}-${i}`}
                                className="px-3 py-2 border-b border-cortex-border/50 hover:bg-cortex-bg/30 transition-colors"
                            >
                                <div className="flex items-start gap-2">
                                    <Icon className={`w-3.5 h-3.5 mt-0.5 flex-shrink-0 ${config.color}`} />
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2">
                                            <span className={`text-[8px] font-mono font-bold uppercase ${config.color}`}>
                                                {config.label}
                                            </span>
                                            {signal.source && (
                                                <span className="text-[8px] font-mono text-cortex-text-muted truncate">
                                                    {signal.source}
                                                </span>
                                            )}
                                            <span className="ml-auto text-[8px] font-mono text-cortex-text-muted/60 whitespace-nowrap">
                                                {relTime}
                                            </span>
                                        </div>
                                        {signal.message && (
                                            <p className="text-[10px] text-cortex-text-main mt-0.5 line-clamp-2">
                                                {signal.message}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            </div>
                        );
                    })
                )}
            </div>
        </div>
    );
}

"use client";

import React, { useState } from 'react';
import { X, ChevronDown, ChevronRight, Clock, Radio, Hash, Fingerprint, Shield } from 'lucide-react';
import { useCortexStore, type SignalDetail } from '@/store/useCortexStore';

// ── Signal type color mapping (mirrors NatsWaterfall spectrumColor) ──

function typeColor(type?: string): { badge: string; dot: string } {
    switch (type) {
        case 'connected':
        case 'heartbeat':
            return { badge: 'bg-cortex-text-muted/20 text-cortex-text-muted', dot: 'bg-cortex-text-muted' };
        case 'thought':
        case 'intent':
        case 'cognitive':
            return { badge: 'bg-cortex-info/15 text-cortex-info', dot: 'bg-cortex-info' };
        case 'artifact':
        case 'output':
        case 'task_complete':
            return { badge: 'bg-cortex-success/15 text-cortex-success', dot: 'bg-cortex-success' };
        case 'error':
        case 'ERROR':
            return { badge: 'bg-cortex-danger/15 text-cortex-danger', dot: 'bg-cortex-danger' };
        case 'governance':
        case 'governance_halt':
            return { badge: 'bg-cortex-warning/15 text-cortex-warning', dot: 'bg-cortex-warning' };
        case 'memory':
            return { badge: 'bg-cortex-primary/15 text-cortex-primary', dot: 'bg-cortex-primary' };
        case 'sensor_data':
        case 'user_input':
        case 'command':
            return { badge: 'bg-cyan-400/15 text-cyan-400', dot: 'bg-cyan-400' };
        case 'tool_call':
        case 'actuation':
            return { badge: 'bg-cortex-success/15 text-cortex-success', dot: 'bg-cortex-success' };
        default:
            return { badge: 'bg-cortex-text-muted/20 text-cortex-text-muted', dot: 'bg-cortex-text-muted' };
    }
}

function levelColor(level?: string): string {
    switch (level) {
        case 'ERROR': return 'text-cortex-danger';
        case 'WARN':  return 'text-cortex-warning';
        case 'DEBUG': return 'text-cortex-text-muted';
        default:      return 'text-cortex-info';
    }
}

function trustColor(score: number): string {
    if (score >= 0.8) return 'text-cortex-success';
    if (score >= 0.5) return 'text-cortex-warning';
    return 'text-cortex-danger';
}

function formatTimestamp(ts: string): string {
    try {
        const d = new Date(ts);
        if (isNaN(d.getTime())) return ts;
        return d.toLocaleString([], {
            year: 'numeric', month: 'short', day: 'numeric',
            hour: '2-digit', minute: '2-digit', second: '2-digit',
        });
    } catch {
        return ts;
    }
}

// ── Metadata Row ─────────────────────────────────────────────

function MetaRow({ icon: Icon, label, value, valueClass }: {
    icon: React.ComponentType<{ className?: string }>;
    label: string;
    value: string;
    valueClass?: string;
}) {
    return (
        <div className="flex items-center gap-3 py-1.5">
            <Icon className="w-3.5 h-3.5 text-cortex-text-muted flex-shrink-0" />
            <span className="text-[10px] font-mono uppercase tracking-wide text-cortex-text-muted w-20 flex-shrink-0">
                {label}
            </span>
            <span className={`text-xs font-mono truncate ${valueClass ?? 'text-cortex-text-main'}`}>
                {value}
            </span>
        </div>
    );
}

// ── Main Drawer ──────────────────────────────────────────────

export default function SignalDetailDrawer() {
    const detail = useCortexStore((s) => s.selectedSignalDetail);
    const selectSignalDetail = useCortexStore((s) => s.selectSignalDetail);
    const [rawExpanded, setRawExpanded] = useState(false);

    if (!detail) return null;

    const colors = typeColor(detail.type);
    const rawData = detail.payload ?? detail.context;
    const hasRawData = rawData != null && (typeof rawData !== 'object' || Object.keys(rawData).length > 0);

    return (
        <div className="absolute right-0 top-0 bottom-0 w-[480px] z-40 bg-cortex-surface border-l border-cortex-border shadow-2xl flex flex-col">
            {/* Header */}
            <div className="h-12 flex items-center justify-between px-4 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-3 min-w-0">
                    <span className={`w-2 h-2 rounded-full flex-shrink-0 ${colors.dot}`} />
                    <span className={`text-[9px] font-mono font-bold uppercase px-2 py-0.5 rounded ${colors.badge}`}>
                        {detail.type}
                    </span>
                    <span className="text-xs font-mono text-cortex-text-main truncate">
                        {detail.source}
                    </span>
                </div>
                <button
                    onClick={() => selectSignalDetail(null)}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors flex-shrink-0"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Scrollable body */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 min-h-0">
                {/* Metadata grid */}
                <div className="bg-cortex-bg rounded-lg p-3 border border-cortex-border/50">
                    <MetaRow icon={Clock} label="Time" value={formatTimestamp(detail.timestamp)} />

                    {detail.level && (
                        <MetaRow icon={Shield} label="Level" value={detail.level} valueClass={levelColor(detail.level)} />
                    )}

                    <MetaRow icon={Radio} label="Source" value={detail.source} />

                    {detail.topic && (
                        <MetaRow icon={Hash} label="Topic" value={detail.topic} />
                    )}

                    {detail.trace_id && (
                        <MetaRow icon={Fingerprint} label="Trace" value={detail.trace_id} />
                    )}

                    {detail.intent && (
                        <MetaRow icon={Radio} label="Intent" value={detail.intent} />
                    )}

                    {detail.trust_score != null && (
                        <div className="flex items-center gap-3 py-1.5">
                            <Shield className="w-3.5 h-3.5 text-cortex-text-muted flex-shrink-0" />
                            <span className="text-[10px] font-mono uppercase tracking-wide text-cortex-text-muted w-20 flex-shrink-0">
                                Trust
                            </span>
                            <span className={`text-xs font-mono font-bold ${trustColor(detail.trust_score)}`}>
                                {detail.trust_score.toFixed(2)}
                            </span>
                        </div>
                    )}
                </div>

                {/* Message */}
                <div>
                    <h3 className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted mb-2">
                        Message
                    </h3>
                    <div className="bg-cortex-bg rounded-lg p-3 border border-cortex-border/50">
                        <p className="text-xs font-mono text-cortex-text-main whitespace-pre-wrap break-words leading-relaxed">
                            {detail.message || '(empty)'}
                        </p>
                    </div>
                </div>

                {/* Raw Data (collapsible) */}
                {hasRawData && (
                    <div>
                        <button
                            onClick={() => setRawExpanded(!rawExpanded)}
                            className="flex items-center gap-1.5 text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted hover:text-cortex-text-main transition-colors mb-2"
                        >
                            {rawExpanded ? (
                                <ChevronDown className="w-3 h-3" />
                            ) : (
                                <ChevronRight className="w-3 h-3" />
                            )}
                            Raw Data
                        </button>
                        {rawExpanded && (
                            <div className="bg-cortex-bg rounded-lg p-3 border border-cortex-border/50 max-h-64 overflow-y-auto">
                                <pre className="text-[10px] font-mono text-cortex-text-main whitespace-pre-wrap break-words">
                                    {JSON.stringify(rawData, null, 2)}
                                </pre>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}

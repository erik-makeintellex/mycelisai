"use client";

import React from 'react';
import { Package, Clock } from 'lucide-react';
import { useCortexStore, type CTSEnvelope } from '@/store/useCortexStore';

function ArtifactCard({ envelope }: { envelope: CTSEnvelope }) {
    const selectArtifact = useCortexStore((s) => s.selectArtifact);

    return (
        <button
            onClick={() => selectArtifact(envelope)}
            className="flex-shrink-0 w-56 bg-cortex-bg border border-cortex-primary/30 rounded-lg p-3 text-left hover:border-cortex-primary/60 hover:shadow-[0_0_12px_rgba(115,103,240,0.2)] transition-all duration-200 group"
        >
            <div className="flex items-center gap-2 mb-2">
                <Package className="w-3.5 h-3.5 text-cortex-primary" />
                <span className="text-[10px] font-mono font-bold text-cortex-text-main truncate">
                    {envelope.payload.title ?? envelope.source}
                </span>
                <span className="ml-auto text-[8px] font-mono text-cortex-text-muted uppercase px-1 py-0.5 rounded bg-cortex-warning/10 text-cortex-warning">
                    pending
                </span>
            </div>

            <p className="text-[10px] font-mono text-cortex-text-muted leading-relaxed line-clamp-2">
                {envelope.payload.content.slice(0, 120)}
                {envelope.payload.content.length > 120 ? '...' : ''}
            </p>

            <div className="flex items-center gap-1.5 mt-2">
                <Clock className="w-2.5 h-2.5 text-cortex-text-muted/60" />
                <span className="text-[8px] font-mono text-cortex-text-muted/60">
                    {new Date(envelope.timestamp).toLocaleTimeString([], {
                        hour: '2-digit',
                        minute: '2-digit',
                        second: '2-digit',
                    })}
                </span>
                {envelope.proof && (
                    <span className={`ml-auto text-[8px] font-mono px-1 py-0.5 rounded ${
                        envelope.proof.pass
                            ? 'bg-cortex-success/10 text-cortex-success'
                            : 'bg-cortex-danger/10 text-cortex-danger'
                    }`}>
                        {envelope.proof.pass ? 'verified' : 'failed'}
                    </span>
                )}
            </div>
        </button>
    );
}

export default function DeliverablesTray() {
    const pendingArtifacts = useCortexStore((s) => s.pendingArtifacts);

    if (pendingArtifacts.length === 0) return null;

    return (
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-40 max-w-[80%]">
            <div className="bg-cortex-surface border border-cortex-primary/30 rounded-t-xl rounded-b-lg px-4 py-3 shadow-lg animate-pulse-subtle">
                {/* Header */}
                <div className="flex items-center gap-2 mb-2.5">
                    <div className="w-2 h-2 rounded-full bg-cortex-success shadow-[0_0_6px_rgba(40,199,111,0.5)] animate-pulse" />
                    <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider">
                        Pending Deliverables
                    </span>
                    <span className="text-[10px] font-mono text-cortex-primary font-bold">
                        {pendingArtifacts.length}
                    </span>
                </div>

                {/* Scrollable card row */}
                <div className="flex gap-3 overflow-x-auto pb-1 scrollbar-thin scrollbar-thumb-cortex-border">
                    {pendingArtifacts.map((envelope) => (
                        <ArtifactCard key={envelope.id} envelope={envelope} />
                    ))}
                </div>
            </div>
        </div>
    );
}

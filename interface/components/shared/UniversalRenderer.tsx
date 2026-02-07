import React from 'react';
import { Envelope } from '@/lib/types/protocol';
import { ThoughtCard } from './ThoughtCard';
import { MetricPill } from './MetricPill';
import { ArtifactCard } from './ArtifactCard';

export function UniversalRenderer({ envelope }: { envelope: Envelope<any> }) {
    switch (envelope.type) {
        case 'thought':
            return <ThoughtCard content={envelope.content} />;
        case 'metric':
            return <MetricPill content={envelope.content} />;
        case 'artifact':
            return <ArtifactCard content={envelope.content} />;
        default:
            return (
                <div className="p-2 bg-red-50 border border-red-200 rounded text-red-600 text-xs font-mono">
                    Unknown Envelope Type: {envelope.type}
                </div>
            );
    }
}

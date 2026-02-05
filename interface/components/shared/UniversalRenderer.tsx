import React from 'react';
import { Envelope, ThoughtContent, MetricContent, ArtifactContent, GovernanceContent } from '@/lib/types/protocol';
import { ThoughtCard } from './ThoughtCard';
import { MetricPill } from './MetricPill';
import { ArtifactCard } from './ArtifactCard';
import { ApprovalCard } from './ApprovalCard';

const JsonFallback: React.FC<{ data: unknown }> = ({ data }) => (
    <div className="bg-zinc-100 p-2 rounded my-1 text-[10px] font-mono overflow-x-auto text-zinc-500">
        <pre>{JSON.stringify(data, null, 2)}</pre>
    </div>
);

export const UniversalRenderer: React.FC<{ message: Envelope }> = ({ message }) => {
    switch (message.type) {
        case 'thought': return <ThoughtCard data={message.content as ThoughtContent} />;
        case 'metric': return <MetricPill data={message.content as MetricContent} />;
        case 'artifact': return <ArtifactCard data={message.content as ArtifactContent} />;
        case 'governance': return <ApprovalCard data={message.content as GovernanceContent} />;
        default: return <JsonFallback data={message} />;
    }
};

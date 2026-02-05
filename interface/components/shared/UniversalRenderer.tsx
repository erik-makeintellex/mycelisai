import React from 'react';
import { Envelope } from '@/lib/types/protocol';
import { ThoughtCard } from './ThoughtCard';
import { MetricPill } from './MetricPill';
import { ArtifactCard } from './ArtifactCard';
import { ApprovalCard } from './ApprovalCard';

const JsonFallback: React.FC<{ data: any }> = ({ data }) => (
    <div className="bg-zinc-100 p-2 rounded my-1 text-[10px] font-mono overflow-x-auto text-zinc-500">
        <pre>{JSON.stringify(data, null, 2)}</pre>
    </div>
);

export const UniversalRenderer: React.FC<{ message: Envelope }> = ({ message }) => {
    switch (message.type) {
        case 'thought': return <ThoughtCard data={message.content} />;
        case 'metric': return <MetricPill data={message.content} />;
        case 'artifact': return <ArtifactCard data={message.content} />;
        case 'governance': return <ApprovalCard data={message.content} />;
        default: return <JsonFallback data={message} />;
    }
};

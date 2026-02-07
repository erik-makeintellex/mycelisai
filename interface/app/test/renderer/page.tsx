import React from 'react';
import { UniversalRenderer } from '@/components/shared/UniversalRenderer';
import { Envelope } from '@/lib/types/protocol';

export default function RendererTestPage() {
    const mockData: Envelope<any>[] = [
        {
            type: 'thought',
            source: 'cognitive-engine',
            timestamp: new Date().toISOString(),
            content: {
                summary: "Analyzing infrastructure requirements",
                detail: "Detected 3 pending nodes. Evaluating load balancing strategy across GPU cluster. Recommending vertical scaling for 'core-db'.",
                model: "gpt-4-turbo"
            }
        },
        {
            type: 'metric',
            source: 'monitor-01',
            timestamp: new Date().toISOString(),
            content: {
                label: "Core CPU",
                value: "42",
                unit: "%",
                status: "nominal"
            }
        },
        {
            type: 'metric',
            source: 'monitor-01',
            timestamp: new Date().toISOString(),
            content: {
                label: "Memory Pressure",
                value: "89",
                unit: "%",
                status: "warning"
            }
        },
        {
            type: 'artifact',
            source: 'codegen',
            timestamp: new Date().toISOString(),
            content: {
                id: "art-123",
                title: "deployment.yaml",
                mime_type: "application/yaml",
                preview: "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: cortex-core",
                uri: "mem://art-123"
            }
        }
    ];

    return (
        <div className="p-8 space-y-8 max-w-2xl mx-auto">
            <div className="border-b border-zinc-200 pb-4">
                <h1 className="text-2xl font-bold text-zinc-900">Universal Renderer Verification</h1>
                <p className="text-zinc-500">Harness for validating Data Envelope visualization.</p>
            </div>

            <div className="space-y-4">
                {mockData.map((env, i) => (
                    <div key={i} className="animate-in fade-in slide-in-from-bottom-4 duration-500" style={{ animationDelay: `${i * 100}ms` }}>
                        <div className="text-[10px] font-mono text-zinc-400 mb-1 ml-1 uppercase tracking-wider">
                            {env.type} | {env.source}
                        </div>
                        <UniversalRenderer envelope={env} />
                    </div>
                ))}
            </div>
        </div>
    );
}

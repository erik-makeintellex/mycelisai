"use client";

import dynamic from 'next/dynamic';

const MCPToolRegistry = dynamic(() => import('@/components/settings/MCPToolRegistry'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading tool registry...</span>
        </div>
    ),
});

export default function ToolsRoute() {
    return <MCPToolRegistry />;
}

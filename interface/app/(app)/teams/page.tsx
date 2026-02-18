"use client";

import dynamic from 'next/dynamic';

const TeamsPage = dynamic(() => import('@/components/teams/TeamsPage'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading teams...</span>
        </div>
    ),
});

export default function TeamsRoute() {
    return <TeamsPage />;
}

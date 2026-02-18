"use client";

import dynamic from 'next/dynamic';

const TeamActuationView = dynamic(() => import('@/components/missions/TeamActuationView'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading team view...</span>
        </div>
    ),
});

export default function MissionTeamsRoute({ params }: { params: Promise<{ id: string }> }) {
    return <TeamActuationView paramsPromise={params} />;
}

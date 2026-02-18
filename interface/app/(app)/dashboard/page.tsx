"use client";

import dynamic from 'next/dynamic';

const MissionControlLayout = dynamic(
    () => import('@/components/dashboard/MissionControl'),
    {
        ssr: false,
        loading: () => (
            <div className="h-full flex items-center justify-center bg-cortex-bg">
                <span className="text-cortex-text-muted text-xs font-mono">Loading Mission Control...</span>
            </div>
        ),
    }
);

export default function DashboardPage() {
    return <MissionControlLayout />;
}

"use client";

import dynamic from 'next/dynamic';
import { use } from 'react';

const RunTimeline = dynamic(
    () => import('@/components/runs/RunTimeline'),
    { ssr: false }
);

export default function RunPage({ params }: { params: Promise<{ id: string }> }) {
    const { id } = use(params);
    return <RunTimeline runId={id} />;
}

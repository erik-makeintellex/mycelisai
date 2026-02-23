import dynamic from 'next/dynamic';

const RunTimeline = dynamic(
    () => import('@/components/runs/RunTimeline'),
    { ssr: false }
);

export default function RunPage({ params }: { params: { id: string } }) {
    return <RunTimeline runId={params.id} />;
}

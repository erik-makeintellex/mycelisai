"use client";

import dynamic from 'next/dynamic';

const CataloguePage = dynamic(() => import('@/components/catalogue/CataloguePage'), {
    ssr: false,
    loading: () => (
        <div className="h-full flex items-center justify-center bg-cortex-bg">
            <span className="text-cortex-text-muted text-xs font-mono">Loading catalogue...</span>
        </div>
    ),
});

export default function CatalogueRoute() {
    return <CataloguePage />;
}

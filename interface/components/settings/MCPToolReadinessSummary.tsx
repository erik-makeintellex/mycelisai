"use client";

import type { CapabilityManifest } from "@/store/useCortexStore";

export function CapabilityReadinessSummary({
    capabilities,
    error,
    isLoading,
    onOpenAccess,
    onOpenCatalog,
    usingFallback,
}: {
    capabilities: CapabilityManifest[];
    error: string | null;
    isLoading: boolean;
    onOpenAccess: () => void;
    onOpenCatalog: () => void;
    usingFallback: boolean;
}) {
    const ready = capabilities.filter(isCapabilityReady);
    const repair = capabilities.filter((capability) => !isCapabilityReady(capability));
    const readyNames = ready.slice(0, 3).map((capability) => capability.name);
    const repairNames = repair.slice(0, 3).map((capability) => capability.name);

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                        Capability overview
                    </p>
                    <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                        What Soma can use right now
                    </p>
                    <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                        {usingFallback
                            ? "Using connected tools and search status because the capability registry is unavailable."
                            : "Open Catalog only when you need full cards, bindings, and attention details."}
                    </p>
                </div>
                <button
                    type="button"
                    onClick={onOpenCatalog}
                    className="self-start rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-xs font-semibold text-cortex-text-muted transition hover:text-cortex-text-main"
                >
                    Open catalog
                </button>
            </div>
            {error ? (
                <p className="mt-3 rounded-lg border border-cortex-warning/25 bg-cortex-warning/10 px-3 py-2 text-xs text-cortex-warning">
                    {error}
                </p>
            ) : null}
            <div className="mt-4 flex flex-wrap gap-2">
                <SummaryChip label="Ready" count={ready.length} onClick={onOpenCatalog} />
                <SummaryChip
                    label="Needs attention"
                    count={repair.length}
                    onClick={onOpenCatalog}
                    tone={repair.length ? "warning" : "success"}
                />
                <SummaryChip label="Available to add" count={1} onClick={onOpenAccess} />
            </div>
            <div className="mt-3 divide-y divide-cortex-border overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg/50">
                <ReadinessRow
                    action="View catalog"
                    count={ready.length}
                    empty={isLoading ? "Checking capabilities..." : "No ready capabilities visible yet."}
                    items={readyNames}
                    onAction={onOpenCatalog}
                    title="Ready"
                    summary="Soma can use these when the current scope allows it."
                />
                <ReadinessRow
                    action={repair.length ? "Review attention" : "View catalog"}
                    count={repair.length}
                    empty="No visible blockers."
                    items={repairNames}
                    onAction={onOpenCatalog}
                    title="Needs attention"
                    summary={repair.length
                        ? "Review these before relying on the related work."
                        : "No capability attention work is visible right now."}
                    tone={repair.length ? "warning" : "success"}
                />
                <ReadinessRow
                    action="Open access"
                    count={1}
                    empty="Search sources, service data connections, data mounts, and scoped permissions."
                    items={[]}
                    onAction={onOpenAccess}
                    title="Available to add"
                    summary="Connect more sources or permissions when Soma needs new reach."
                />
            </div>
        </div>
    );
}

export function isCapabilityReady(capability: CapabilityManifest): boolean {
    const status = capability.availability_status?.toLowerCase();
    return !status || status === "available" || status === "connected" || status === "ready" || status === "online";
}

function SummaryChip({
    count,
    label,
    onClick,
    tone = "neutral",
}: {
    count: number;
    label: string;
    onClick: () => void;
    tone?: "neutral" | "success" | "warning";
}) {
    const toneClass = tone === "warning"
        ? "border-cortex-warning/30 bg-cortex-warning/10 text-cortex-warning"
        : tone === "success"
        ? "border-cortex-success/30 bg-cortex-success/10 text-cortex-success"
        : "border-cortex-border bg-cortex-bg text-cortex-text-main";
    return (
        <button
            type="button"
            onClick={onClick}
            className={`inline-flex min-h-8 items-center gap-2 rounded-full border px-3 py-1 text-xs font-semibold transition hover:border-cortex-primary/40 ${toneClass}`}
        >
            <span>{label}</span>
            <span className="rounded-full border border-current/20 px-1.5 py-0.5 font-mono text-[10px]">
                {count}
            </span>
        </button>
    );
}

function ReadinessRow({
    action,
    count,
    empty,
    items,
    onAction,
    summary,
    title,
    tone = "neutral",
}: {
    action: string;
    count: number;
    empty: string;
    items: string[];
    onAction: () => void;
    summary: string;
    title: string;
    tone?: "neutral" | "success" | "warning";
}) {
    const countClass = tone === "warning"
        ? "border-cortex-warning/30 bg-cortex-warning/10 text-cortex-warning"
        : tone === "success"
        ? "border-cortex-success/30 bg-cortex-success/10 text-cortex-success"
        : "border-cortex-border bg-cortex-bg text-cortex-text-muted";
    return (
        <section className="grid gap-3 px-3 py-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
            <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                    <p className="text-xs font-semibold text-cortex-text-main">{title}</p>
                    <span className={`rounded-full border px-2 py-0.5 text-[10px] font-mono ${countClass}`}>
                        {count}
                    </span>
                </div>
                <p className="mt-1 text-xs leading-5 text-cortex-text-muted">{summary}</p>
                <p className="mt-1 truncate text-xs text-cortex-text-main">
                    {items.length ? items.join(", ") : empty}
                </p>
            </div>
            <button
                type="button"
                onClick={onAction}
                className="justify-self-start rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-xs font-semibold text-cortex-text-muted transition hover:text-cortex-text-main md:justify-self-end"
            >
                {action}
            </button>
        </section>
    );
}

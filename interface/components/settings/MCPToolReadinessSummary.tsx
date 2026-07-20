"use client";

import type { CapabilityManifest } from "@/store/useCortexStore";
import { CAPABILITY_ORIGINS, capabilityOrigin } from "./MCPToolCapabilityOrigin";

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
    const originCounts = new Map(
        CAPABILITY_ORIGINS.map((origin) => [
            origin.id,
            capabilities.filter((capability) => capabilityOrigin(capability) === origin.id).length,
        ]),
    );

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
            <div className="mt-3 flex gap-2 overflow-x-auto pb-1" aria-label="Capability origin summary">
                {CAPABILITY_ORIGINS.map((origin) => (
                    <button
                        key={origin.id}
                        type="button"
                        aria-label={`${origin.label}: ${isLoading ? "checking" : originCounts.get(origin.id) ?? 0}`}
                        onClick={onOpenCatalog}
                        className="inline-flex min-h-9 shrink-0 items-center gap-2 rounded-full border border-cortex-border bg-cortex-bg/50 px-3 py-1.5 text-left transition hover:border-cortex-primary/40"
                    >
                        <span className="text-xs font-semibold text-cortex-text-main">{origin.label}</span>
                        <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-1 font-mono text-[10px] text-cortex-text-main">
                            {isLoading ? "…" : originCounts.get(origin.id) ?? 0}
                        </span>
                    </button>
                ))}
            </div>
            <dl className="mt-2 grid gap-x-4 gap-y-1 border-t border-cortex-border pt-2 sm:grid-cols-2">
                {CAPABILITY_ORIGINS.map((origin) => (
                    <div key={origin.id} className="flex min-w-0 gap-1 text-[11px] leading-4">
                        <dt className="shrink-0 font-semibold text-cortex-text-main">{origin.label}:</dt>
                        <dd className="text-cortex-text-muted">{origin.summary}</dd>
                    </div>
                ))}
            </dl>
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

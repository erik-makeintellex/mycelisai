"use client";

import { useState, type ReactNode } from "react";
import { Database, Search, ShieldCheck, SlidersHorizontal } from "lucide-react";
import type { CapabilityManifest, SearchCapabilityStatus } from "@/store/useCortexStore";
import { CapabilityRegistryPanel } from "./MCPToolCapabilityRegistry";
import { ConnectedToolsWorkflowCard, SearchCapabilityCard, SomaToolPromptCard, WebAccessSetupCard } from "./MCPToolGuidance";
import { SearchSourceRegistryCard } from "./SearchSourceRegistryCard";
import { useSearchSourceRegistry } from "./MCPToolRegistrySearchSources";
import { MCPToolSetLayersStorePanel } from "./MCPToolSetLayersPanel";
import { MCPServiceConnectionGuide } from "./MCPServiceConnectionGuide";

type OverviewFocus = "readiness" | "catalog" | "access" | "inspect";

type SearchSourceRegistryController = ReturnType<typeof useSearchSourceRegistry>;

export function MCPToolRegistryOverview({
    capabilities,
    isFetchingCapabilities,
    capabilitiesError,
    usingCapabilityFallback,
    searchCapability,
    isFetchingSearchCapability,
    searchCapabilityError,
    searchSourceRegistry,
    searchSourceCreateRequest,
    isStreamConnected,
    onAddWebCapability,
}: {
    capabilities: CapabilityManifest[];
    isFetchingCapabilities: boolean;
    capabilitiesError: string | null;
    usingCapabilityFallback: boolean;
    searchCapability: SearchCapabilityStatus | null;
    isFetchingSearchCapability: boolean;
    searchCapabilityError: string | null;
    searchSourceRegistry: SearchSourceRegistryController;
    searchSourceCreateRequest: { nonce: number; sourceType?: string } | null;
    isStreamConnected: boolean;
    onAddWebCapability: () => void;
}) {
    const [activeFocus, setActiveFocus] = useState<OverviewFocus>("readiness");
    const availableCount = capabilities.filter(isCapabilityReady).length;
    const repairCount = capabilities.length - availableCount;

    return (
        <div className="flex flex-col gap-4 p-6 max-w-4xl mx-auto">
            <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
                <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div>
                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                            Capability focus
                        </p>
                        <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                            Choose what you want to check.
                        </p>
                    </div>
                    <div className="grid gap-2 sm:grid-cols-4">
                        <FocusButton
                            active={activeFocus === "readiness"}
                            icon={<Search className="h-3.5 w-3.5" />}
                            label="Readiness"
                            detail="Search and web"
                            onClick={() => setActiveFocus("readiness")}
                        />
                        <FocusButton
                            active={activeFocus === "catalog"}
                            icon={<ShieldCheck className="h-3.5 w-3.5" />}
                            label="Catalog"
                            detail={`${availableCount} ready / ${repairCount} repair`}
                            onClick={() => setActiveFocus("catalog")}
                        />
                        <FocusButton
                            active={activeFocus === "access"}
                            icon={<SlidersHorizontal className="h-3.5 w-3.5" />}
                            label="Access"
                            detail="Sources, scopes, data"
                            onClick={() => setActiveFocus("access")}
                        />
                        <FocusButton
                            active={activeFocus === "inspect"}
                            icon={<Database className="h-3.5 w-3.5" />}
                            label="Inspect"
                            detail="Refs and examples"
                            onClick={() => setActiveFocus("inspect")}
                        />
                    </div>
                </div>
            </div>

            {activeFocus === "readiness" && (
                <div className="grid gap-4">
                    <WebAccessSetupCard
                        status={searchCapability}
                        isLoading={isFetchingSearchCapability}
                        error={searchCapabilityError}
                        onAddWebCapability={onAddWebCapability}
                    />
                    <CapabilityReadinessSummary
                        capabilities={capabilities}
                        isLoading={isFetchingCapabilities}
                        error={capabilitiesError}
                        usingFallback={usingCapabilityFallback}
                        onOpenCatalog={() => setActiveFocus("catalog")}
                        onOpenAccess={() => setActiveFocus("access")}
                    />
                </div>
            )}

            {activeFocus === "catalog" && (
                <CapabilityRegistryPanel
                    capabilities={capabilities}
                    isLoading={isFetchingCapabilities}
                    error={capabilitiesError}
                    usingFallback={usingCapabilityFallback}
                />
            )}

            {activeFocus === "access" && (
                <div className="grid gap-4">
                    <SearchSourceRegistryCard
                        sources={searchSourceRegistry.visibleSearchSources}
                        isLoading={searchSourceRegistry.isFetchingSearchSources}
                        addSupported={searchSourceRegistry.searchSourceRegistrySupported}
                        error={searchSourceRegistry.searchSourcesError}
                        addNotice={searchSourceRegistry.searchSourceNotice}
                        isAdding={searchSourceRegistry.isAddingSearchSource}
                        openCreateRequest={searchSourceCreateRequest}
                        onAddSearchSource={searchSourceRegistry.addSearchSource}
                        onDeleteSearchSource={searchSourceRegistry.deleteSearchSource}
                        onUpdateSearchSource={searchSourceRegistry.updateSearchSource}
                    />
                    <MCPServiceConnectionGuide />
                    <MCPToolSetLayersStorePanel />
                </div>
            )}

            {activeFocus === "inspect" && (
                <div className="grid gap-4">
                    <SearchCapabilityCard
                        status={searchCapability}
                        isLoading={isFetchingSearchCapability}
                        error={searchCapabilityError}
                    />
                    <SomaToolPromptCard />
                    <ConnectedToolsWorkflowCard isStreamConnected={isStreamConnected} />
                </div>
            )}
        </div>
    );
}

function FocusButton({
    active,
    detail,
    icon,
    label,
    onClick,
}: {
    active: boolean;
    detail: string;
    icon: ReactNode;
    label: string;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`rounded-xl border px-3 py-2 text-left transition ${
                active
                    ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg text-cortex-text-muted hover:text-cortex-text-main"
            }`}
        >
            <span className="flex items-center gap-2 text-xs font-semibold">
                {icon}
                {label}
            </span>
            <span className="mt-1 block text-[10px] leading-4 text-cortex-text-muted">{detail}</span>
        </button>
    );
}

function CapabilityReadinessSummary({
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
    const readyNames = ready.slice(0, 4).map((capability) => capability.name);
    const repairNames = repair.slice(0, 3).map((capability) => capability.name);

    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
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
                            : "Open Catalog only when you need full cards, bindings, and repair detail."}
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
            <div className="mt-4 grid gap-3 md:grid-cols-3">
                <ReadinessLane
                    title="Can use now"
                    count={ready.length}
                    empty={isLoading ? "Checking capabilities..." : "No ready capabilities visible yet."}
                    items={readyNames}
                />
                <ReadinessLane
                    title="Needs repair"
                    count={repair.length}
                    empty="No capability blockers visible."
                    items={repairNames}
                    tone={repair.length ? "warning" : "success"}
                />
                <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-3">
                    <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                        Can request/add
                    </p>
                    <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                        Add connector opens the curated library. Use Access for search sources, service connections, data mounts, and scoped permissions.
                    </p>
                    <button
                        type="button"
                        onClick={onOpenAccess}
                        className="mt-3 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2 text-xs font-semibold text-cortex-primary transition hover:bg-cortex-primary/20"
                    >
                        Open access
                    </button>
                </div>
            </div>
        </div>
    );
}

function ReadinessLane({
    count,
    empty,
    items,
    title,
    tone = "neutral",
}: {
    count: number;
    empty: string;
    items: string[];
    title: string;
    tone?: "neutral" | "success" | "warning";
}) {
    const toneClass = tone === "warning"
        ? "border-cortex-warning/25 bg-cortex-warning/10"
        : tone === "success"
        ? "border-cortex-success/25 bg-cortex-success/10"
        : "border-cortex-border bg-cortex-bg/60";
    return (
        <section className={`rounded-lg border px-3 py-3 ${toneClass}`}>
            <div className="flex items-center justify-between gap-2">
                <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                    {title}
                </p>
                <span className="rounded-full border border-cortex-border bg-cortex-bg px-2 py-1 text-[10px] font-mono text-cortex-text-muted">
                    {count}
                </span>
            </div>
            {items.length ? (
                <ul className="mt-2 space-y-1">
                    {items.map((item) => (
                        <li key={item} className="truncate text-xs font-semibold text-cortex-text-main">
                            {item}
                        </li>
                    ))}
                </ul>
            ) : (
                <p className="mt-2 text-xs leading-5 text-cortex-text-muted">{empty}</p>
            )}
        </section>
    );
}

function isCapabilityReady(capability: CapabilityManifest): boolean {
    const status = capability.availability_status?.toLowerCase();
    return !status || status === "available" || status === "connected" || status === "ready" || status === "online";
}

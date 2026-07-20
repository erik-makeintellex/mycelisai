"use client";

import { useState, type ReactNode } from "react";
import { Database, Search, ShieldCheck, SlidersHorizontal } from "lucide-react";
import type { CapabilityManifest, SearchCapabilityStatus } from "@/store/useCortexStore";
import { CapabilityRegistryPanel } from "./MCPToolCapabilityRegistry";
import { InputSourceRegistryCard } from "./InputSourceRegistryCard";
import type { useInputSourceRegistry } from "./InputSourceRegistry";
import { ConnectedToolsWorkflowCard, SearchCapabilityCard, SomaToolPromptCard, WebAccessSetupCard } from "./MCPToolGuidance";
import { SearchSourceRegistryCard } from "./SearchSourceRegistryCard";
import { useSearchSourceRegistry } from "./MCPToolRegistrySearchSources";
import { MCPToolSetLayersStorePanel } from "./MCPToolSetLayersPanel";
import { MCPServiceConnectionGuide } from "./MCPServiceConnectionGuide";
import { CapabilityReadinessSummary, isCapabilityReady } from "./MCPToolReadinessSummary";

type OverviewFocus = "readiness" | "catalog" | "access" | "inspect";
type AccessFocus = "search" | "inputs" | "services";

type SearchSourceRegistryController = ReturnType<typeof useSearchSourceRegistry>;
type InputSourceRegistryController = ReturnType<typeof useInputSourceRegistry>;

export function MCPToolRegistryOverview({
    capabilities,
    isFetchingCapabilities,
    capabilitiesError,
    usingCapabilityFallback,
    searchCapability,
    isFetchingSearchCapability,
    searchCapabilityError,
    searchSourceRegistry,
    inputSourceRegistry,
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
    inputSourceRegistry: InputSourceRegistryController;
    searchSourceCreateRequest: { nonce: number; sourceType?: string } | null;
    isStreamConnected: boolean;
    onAddWebCapability: () => void;
}) {
    const [activeFocus, setActiveFocus] = useState<OverviewFocus>(
        searchSourceCreateRequest ? "access" : "readiness",
    );
    const [activeAccessFocus, setActiveAccessFocus] = useState<AccessFocus>("search");
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
                    <div className="flex flex-wrap gap-2">
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
                    <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                            <div>
                                <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                                    Capability permissions
                                </p>
                                <p className="mt-1 text-sm font-semibold text-cortex-text-main">
                                    Choose the access job.
                                </p>
                                <p className="mt-1 text-xs leading-5 text-cortex-text-muted">
                                    Search sources, live event feeds, and service connections stay separate so setup does not become one long form.
                                </p>
                            </div>
                            <div className="flex flex-wrap gap-2">
                                <AccessButton
                                    active={activeAccessFocus === "search"}
                                    label="Search sources"
                                    detail={`${searchSourceRegistry.visibleSearchSources.length} configured`}
                                    onClick={() => setActiveAccessFocus("search")}
                                />
                                <AccessButton
                                    active={activeAccessFocus === "inputs"}
                                    label="Live inputs"
                                    detail={`${inputSourceRegistry.sources.length} feeds`}
                                    onClick={() => setActiveAccessFocus("inputs")}
                                />
                                <AccessButton
                                    active={activeAccessFocus === "services"}
                                    label="Service connections"
                                    detail="Tools and scopes"
                                    onClick={() => setActiveAccessFocus("services")}
                                />
                            </div>
                        </div>
                    </div>
                    {activeAccessFocus === "search" && (
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
                    )}
                    {activeAccessFocus === "inputs" && (
                        <InputSourceRegistryCard
                            bufferError={inputSourceRegistry.bufferError}
                            bufferView={inputSourceRegistry.bufferView}
                            isFetchingBuffer={inputSourceRegistry.isFetchingBuffer}
                            isLoading={inputSourceRegistry.isFetchingSources}
                            registrySupported={inputSourceRegistry.registrySupported}
                            selectedSourceId={inputSourceRegistry.selectedSourceId}
                            sources={inputSourceRegistry.sources}
                            sourcesError={inputSourceRegistry.sourcesError}
                            onRefresh={inputSourceRegistry.fetchInputSources}
                            onSelectSource={inputSourceRegistry.fetchSourceBuffer}
                        />
                    )}
                    {activeAccessFocus === "services" && (
                        <>
                            <MCPServiceConnectionGuide />
                            <MCPToolSetLayersStorePanel />
                        </>
                    )}
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

function AccessButton({
    active,
    detail,
    label,
    onClick,
}: {
    active: boolean;
    detail: string;
    label: string;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`rounded-lg border px-3 py-2 text-left transition ${
                active
                    ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg text-cortex-text-muted hover:text-cortex-text-main"
            }`}
        >
            <span className="block text-xs font-semibold">{label}</span>
            <span className="block text-[10px] leading-4">{detail}</span>
        </button>
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
            className={`inline-flex min-h-9 items-center gap-2 rounded-full border px-3 py-1.5 text-left transition ${
                active
                    ? "border-cortex-primary/40 bg-cortex-primary/10 text-cortex-text-main"
                    : "border-cortex-border bg-cortex-bg text-cortex-text-muted hover:text-cortex-text-main"
            }`}
        >
            {icon}
            <span className="text-xs font-semibold">
                {label}
            </span>
            <span className="hidden text-[10px] leading-4 text-cortex-text-muted lg:inline">{detail}</span>
        </button>
    );
}

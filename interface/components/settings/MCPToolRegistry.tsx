"use client";

import React, { useEffect, useMemo, useState } from "react";
import { BookOpen, ShieldCheck, Wrench } from "lucide-react";
import MCPServerCard, { type MCPRecentActivity } from "./MCPServerCard";
import { useCortexStore } from "@/store/useCortexStore";
import { MCPLibraryBrowserBody } from "./MCPLibraryBrowser";
import { CapabilityRegistryPanel } from "./MCPToolCapabilityRegistry";
import { formatActivityScope, useMCPRecentActivity } from "./MCPToolRegistryActivity";
import { deriveFallbackCapabilities } from "./MCPToolRegistryCapabilities";
import { ConnectedToolsWorkflowCard, SearchCapabilityCard, SomaToolPromptCard, WebAccessSetupCard } from "./MCPToolGuidance";
import { MCPToolSetLayersStorePanel } from "./MCPToolSetLayersPanel";
import { MCPInstallNotice, MCPRegistryEmptyBanner, MCPRegistryEmptyHero, MCPRegistryErrorBanner } from "./MCPToolRegistryNotices";

type Tab = "installed" | "library";

export default function MCPToolRegistry() {
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const isFetching = useCortexStore((s) => s.isFetchingMCPServers);
    const mcpServersError = useCortexStore((s) => s.mcpServersError);
    const mcpActivity = useCortexStore((s) => s.mcpActivity);
    const isFetchingActivity = useCortexStore((s) => s.isFetchingMCPActivity);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);
    const fetchMCPActivity = useCortexStore((s) => s.fetchMCPActivity);
    const fetchSearchCapability = useCortexStore((s) => s.fetchSearchCapability);
    const deleteMCPServer = useCortexStore((s) => s.deleteMCPServer);
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const isStreamConnected = useCortexStore((s) => s.isStreamConnected);
    const initializeStream = useCortexStore((s) => s.initializeStream);
    const searchCapability = useCortexStore((s) => s.searchCapability);
    const isFetchingSearchCapability = useCortexStore((s) => s.isFetchingSearchCapability);
    const searchCapabilityError = useCortexStore((s) => s.searchCapabilityError);
    const capabilities = useCortexStore((s) => s.capabilities);
    const isFetchingCapabilities = useCortexStore((s) => s.isFetchingCapabilities);
    const capabilitiesError = useCortexStore((s) => s.capabilitiesError);
    const fetchCapabilities = useCortexStore((s) => s.fetchCapabilities);
    const fetchMCPToolSets = useCortexStore((s) => s.fetchMCPToolSets);

    const [activeTab, setActiveTab] = useState<Tab>("installed");
    const [installNotice, setInstallNotice] = useState<string | null>(null);
    const [librarySearchQuery, setLibrarySearchQuery] = useState("");
    const [showTopology, setShowTopology] = useState(false);
    const isRegistryErrorState = !isFetching && Boolean(mcpServersError);
    const isEmptyInstalledState = !isFetching && !mcpServersError && mcpServers.length === 0;

    useEffect(() => {
        fetchMCPServers();
        fetchMCPActivity();
        fetchSearchCapability();
        fetchCapabilities();
        fetchMCPToolSets();
    }, [fetchCapabilities, fetchMCPActivity, fetchMCPServers, fetchMCPToolSets, fetchSearchCapability]);

    useEffect(() => {
        initializeStream();
    }, [initializeStream]);

    const recentActivity = useMCPRecentActivity(mcpServers, mcpActivity, streamLogs);

    const recentActivityByServer = useMemo(() => {
        const grouped = new Map<string, MCPRecentActivity[]>();
        for (const activity of recentActivity) {
            if (!activity.serverId) continue;
            const current = grouped.get(activity.serverId) ?? [];
            current.push(activity);
            grouped.set(activity.serverId, current);
        }
        return grouped;
    }, [recentActivity]);

    const fallbackCapabilities = useMemo(
        () => deriveFallbackCapabilities(mcpServers, searchCapability, searchCapabilityError),
        [mcpServers, searchCapability, searchCapabilityError],
    );
    const visibleCapabilities = capabilities.length > 0 ? capabilities : fallbackCapabilities;
    const usingCapabilityFallback = capabilities.length === 0 && fallbackCapabilities.length > 0;

    function handleInstalled(name: string) {
        setInstallNotice(`Installed ${name}. Check the connected server card and live MCP activity below.`);
        setActiveTab("installed");
    }

    function handleAddWebCapability() {
        setLibrarySearchQuery("fetch");
        setActiveTab("library");
    }

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            <div className="h-12 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between px-6 flex-shrink-0">
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2.5">
                        <Wrench className="w-4 h-4 text-cortex-success" />
                        <span className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-wider">
                            Capabilities
                        </span>
                    </div>
                    <div className="flex items-center gap-1 bg-cortex-bg rounded-lg p-0.5 border border-cortex-border">
                        <button
                            onClick={() => setActiveTab("installed")}
                            className={`px-3 py-1 rounded-md text-[10px] font-mono font-bold uppercase tracking-wider transition-colors ${
                                activeTab === "installed"
                                    ? "bg-cortex-surface text-cortex-text-main shadow-sm"
                                    : "text-cortex-text-muted hover:text-cortex-text-main"
                            }`}
                        >
                            Installed
                            {mcpServers.length > 0 && (
                                <span className="ml-1.5 text-[9px] px-1 py-0.5 rounded bg-cortex-info/10 text-cortex-info">
                                    {mcpServers.length}
                                </span>
                            )}
                        </button>
                        <button
                            onClick={() => setActiveTab("library")}
                            className={`flex items-center gap-1 px-3 py-1 rounded-md text-[10px] font-mono font-bold uppercase tracking-wider transition-colors ${
                                activeTab === "library"
                                    ? "bg-cortex-surface text-cortex-text-main shadow-sm"
                                    : "text-cortex-text-muted hover:text-cortex-text-main"
                            }`}
                        >
                            <BookOpen className="w-3 h-3" />
                            Add MCP
                        </button>
                    </div>
                </div>

                {activeTab === "installed" && (
                    <button
                        onClick={() => setActiveTab("library")}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-cortex-primary/10 border border-cortex-primary/30 text-xs font-mono font-bold text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                    >
                        <BookOpen className="w-3.5 h-3.5" />
                        Request capability
                    </button>
                )}
            </div>

            <div className="flex-1 overflow-y-auto">
                {activeTab === "installed" && (
                    <div className="flex flex-col gap-4 p-6 max-w-4xl mx-auto">
                        <WebAccessSetupCard
                            status={searchCapability}
                            isLoading={isFetchingSearchCapability}
                            error={searchCapabilityError}
                            onAddWebCapability={handleAddWebCapability}
                        />
                        <SomaToolPromptCard />
                        <ConnectedToolsWorkflowCard isStreamConnected={isStreamConnected} />
                        <CapabilityRegistryPanel
                            capabilities={visibleCapabilities}
                            isLoading={isFetchingCapabilities}
                            error={capabilitiesError}
                            usingFallback={usingCapabilityFallback}
                        />
                        <MCPToolSetLayersStorePanel />

                        <SearchCapabilityCard
                            status={searchCapability}
                            isLoading={isFetchingSearchCapability}
                            error={searchCapabilityError}
                        />

                        {isEmptyInstalledState && <MCPRegistryEmptyBanner />}

                        {isRegistryErrorState && <MCPRegistryErrorBanner error={mcpServersError} />}

                        {installNotice && <MCPInstallNotice message={installNotice} />}

                        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
                            <div className="flex items-center justify-between gap-2">
                                <div>
                                    <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                                        Recent MCP Activity
                                    </p>
                                    <p className="mt-1 text-xs text-cortex-text-muted">
                                        Persisted recent MCP usage from Managed Exchange, plus live in-session runtime signals when connected.
                                    </p>
                                </div>
                                <span className="rounded-full border border-cortex-border bg-cortex-bg px-2 py-1 text-[10px] font-mono text-cortex-text-muted">
                                    {recentActivity.length} visible
                                </span>
                            </div>
                            {recentActivity.length === 0 ? (
                                <p className="mt-3 text-xs text-cortex-text-muted">
                                    {isFetchingActivity
                                        ? "Loading recent MCP activity..."
                                        : "No recent MCP activity yet. Install a tool server, ask Soma to use it, and recent tool use will appear here."}
                                </p>
                            ) : (
                                <div className="mt-3 space-y-2">
                                    {recentActivity.slice(0, 5).map((activity) => (
                                        <div key={activity.id} className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-2">
                                            <div className="flex items-center justify-between gap-2">
                                                <span className="text-[11px] font-semibold text-cortex-text-main">
                                                    {activity.serverName} · {activity.toolName}
                                                </span>
                                                <span className="text-[10px] font-mono uppercase text-cortex-info">
                                                    {activity.state}
                                                </span>
                                            </div>
                                            <p className="mt-1 text-[10px] text-cortex-text-muted">{activity.message}</p>
                                            {formatActivityScope(activity) && (
                                                <p className="mt-1 text-[10px] font-mono text-cortex-text-muted">
                                                    {formatActivityScope(activity)}
                                                </p>
                                            )}
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {isFetching && mcpServers.length === 0 && (
                            <>
                                {[1, 2, 3].map((i) => (
                                    <div key={i} className="h-16 rounded-xl bg-cortex-surface border border-cortex-border animate-pulse" />
                                ))}
                            </>
                        )}

                        {isEmptyInstalledState && <MCPRegistryEmptyHero onRequest={() => setActiveTab("library")} />}

                        {mcpServers.length > 0 && (
                            <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
                                <div className="flex items-center justify-between gap-2">
                                    <div className="flex items-center gap-2">
                                        <ShieldCheck className="h-4 w-4 text-cortex-primary" />
                                        <div>
                                            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                                                Advanced Inspect
                                            </p>
                                            <p className="mt-1 text-xs text-cortex-text-muted">
                                                Raw MCP topology is available when you need transport, command, secret refs, tools, and activity detail.
                                            </p>
                                        </div>
                                    </div>
                                    <button
                                        type="button"
                                        onClick={() => setShowTopology((value) => !value)}
                                        className="rounded border border-cortex-border bg-cortex-bg px-3 py-1.5 text-[10px] font-mono font-bold uppercase text-cortex-text-muted hover:text-cortex-text-main"
                                    >
                                        {showTopology ? "Hide topology" : "Inspect MCP topology"}
                                    </button>
                                </div>
                                {showTopology && (
                                    <div className="mt-3 space-y-3">
                                        <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                                            MCP Server Drill-Down
                                        </p>
                                        {mcpServers.map((server) => (
                                            <MCPServerCard
                                                key={server.id}
                                                server={server}
                                                onDelete={deleteMCPServer}
                                                onEdit={() => setActiveTab("library")}
                                                recentActivity={recentActivityByServer.get(server.id) ?? []}
                                            />
                                        ))}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                )}

                {activeTab === "library" && (
                    <MCPLibraryBrowserBody onInstalled={handleInstalled} initialSearchQuery={librarySearchQuery} />
                )}
            </div>
        </div>
    );
}

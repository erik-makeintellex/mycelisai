"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Wrench, BookOpen, Activity, Radio } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import MCPServerCard, { type MCPRecentActivity } from "./MCPServerCard";
import { MCPLibraryBrowserBody } from "./MCPLibraryBrowser";

type Tab = "installed" | "library";

export default function MCPToolRegistry() {
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const isFetching = useCortexStore((s) => s.isFetchingMCPServers);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);
    const deleteMCPServer = useCortexStore((s) => s.deleteMCPServer);
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const isStreamConnected = useCortexStore((s) => s.isStreamConnected);
    const initializeStream = useCortexStore((s) => s.initializeStream);

    const [activeTab, setActiveTab] = useState<Tab>("installed");
    const [installNotice, setInstallNotice] = useState<string | null>(null);

    useEffect(() => {
        fetchMCPServers();
    }, [fetchMCPServers]);

    useEffect(() => {
        initializeStream();
    }, [initializeStream]);

    const recentActivity = useMemo<MCPRecentActivity[]>(() => {
        const serverNames = new Map(mcpServers.map((server) => [server.id, server.name]));
        return streamLogs
            .filter((signal) => signal.source_kind === "mcp")
            .map((signal, index) => {
                const serverId = typeof signal.payload?.server_id === "string" ? signal.payload.server_id : undefined;
                const toolName = typeof signal.payload?.tool === "string" ? signal.payload.tool : "unknown_tool";
                const state = typeof signal.payload?.state === "string" ? signal.payload.state : "activity";
                const preview = typeof signal.payload?.result_preview === "string"
                    ? signal.payload.result_preview
                    : typeof signal.payload?.error === "string"
                    ? signal.payload.error
                    : signal.message ?? "Agent MCP activity recorded.";
                return {
                    id: `${signal.timestamp ?? "mcp"}-${serverId ?? "server"}-${toolName}-${index}`,
                    serverId,
                    serverName: serverId ? (serverNames.get(serverId) ?? serverId) : "mcp",
                    toolName,
                    state,
                    message: preview,
                    timestamp: signal.timestamp ?? new Date().toISOString(),
                };
            })
            .slice(0, 12);
    }, [mcpServers, streamLogs]);

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

    function handleInstalled(name: string) {
        setInstallNotice(`Installed ${name}. Check the connected server card and live MCP activity below.`);
        setActiveTab("installed");
    }

    return (
        <div className="h-full flex flex-col bg-cortex-bg">
            {/* Header Bar */}
            <div className="h-12 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex items-center justify-between px-6 flex-shrink-0">
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2.5">
                        <Wrench className="w-4 h-4 text-cortex-success" />
                        <span className="text-xs font-mono font-bold text-cortex-text-muted uppercase tracking-wider">
                            MCP Tool Registry
                        </span>
                    </div>

                    {/* Tabs */}
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
                            Library
                        </button>
                    </div>
                </div>

                {activeTab === "installed" && (
                    <button
                        onClick={() => setActiveTab("library")}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-cortex-primary/10 border border-cortex-primary/30 text-xs font-mono font-bold text-cortex-primary hover:bg-cortex-primary/20 transition-colors"
                    >
                        <BookOpen className="w-3.5 h-3.5" />
                        BROWSE LIBRARY
                    </button>
                )}
            </div>

            {/* Body */}
            <div className="flex-1 overflow-y-auto">
                {activeTab === "installed" && (
                    <div className="flex flex-col gap-4 p-6 max-w-4xl mx-auto">
                        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
                            <div className="flex items-center gap-2">
                                <Activity className="w-4 h-4 text-cortex-primary" />
                                <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                                    Connected Tools Workflow
                                </p>
                            </div>
                            <div className="mt-3 grid gap-3 md:grid-cols-3">
                                <WorkflowStep title="1. Add">
                                    Install a curated MCP server from Library instead of wiring raw config by hand.
                                </WorkflowStep>
                                <WorkflowStep title="2. Verify">
                                    Confirm the server is connected and that its discovered tools match what agents should use.
                                </WorkflowStep>
                                <WorkflowStep title="3. Watch">
                                    Live MCP activity below shows which server and tool agents are using right now.
                                </WorkflowStep>
                            </div>
                            <div className="mt-3 flex items-center gap-2 text-[11px] text-cortex-text-muted">
                                <Radio className={`w-3.5 h-3.5 ${isStreamConnected ? "text-cortex-success" : "text-cortex-warning"}`} />
                                {isStreamConnected
                                    ? "Live activity stream connected."
                                    : "Live activity stream is reconnecting. Recent MCP use will appear once the stream is online."}
                            </div>
                        </div>

                        {installNotice && (
                            <div className="rounded-xl border border-cortex-success/25 bg-cortex-success/10 px-4 py-3">
                                <p className="text-xs font-mono leading-5 text-cortex-text-main">{installNotice}</p>
                            </div>
                        )}

                        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-4">
                            <div className="flex items-center justify-between gap-2">
                                <div>
                                    <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                                        Recent MCP Activity
                                    </p>
                                    <p className="mt-1 text-xs text-cortex-text-muted">
                                        Agent-visible MCP invocations and outcomes from the live runtime signal bus.
                                    </p>
                                </div>
                                <span className="rounded-full border border-cortex-border bg-cortex-bg px-2 py-1 text-[10px] font-mono text-cortex-text-muted">
                                    {recentActivity.length} visible
                                </span>
                            </div>
                            {recentActivity.length === 0 ? (
                                <p className="mt-3 text-xs text-cortex-text-muted">
                                    No live MCP activity yet. Install a tool server, ask Soma to use it, and recent tool use will appear here.
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
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        {/* Loading Skeletons */}
                        {isFetching && mcpServers.length === 0 && (
                            <>
                                {[1, 2, 3].map((i) => (
                                    <div key={i} className="h-16 rounded-xl bg-cortex-surface border border-cortex-border animate-pulse" />
                                ))}
                            </>
                        )}

                        {/* Empty State */}
                        {!isFetching && mcpServers.length === 0 && (
                            <div className="flex flex-col items-center justify-center py-24 text-cortex-text-muted">
                                <Wrench className="w-12 h-12 mb-3 opacity-20" />
                                <p className="text-sm font-mono">No MCP servers installed.</p>
                                <p className="text-[10px] font-mono mt-1 opacity-50">
                                    Browse the Library tab to add approved tool servers.
                                </p>
                                <button
                                    onClick={() => setActiveTab("library")}
                                    className="mt-4 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-1.5 text-[10px] font-mono font-bold text-cortex-primary transition-colors hover:bg-cortex-primary/20"
                                >
                                    OPEN LIBRARY
                                </button>
                            </div>
                        )}

                        {/* Server Cards */}
                        {mcpServers.map((server) => (
                            <MCPServerCard
                                key={server.id}
                                server={server}
                                onDelete={deleteMCPServer}
                                recentActivity={recentActivityByServer.get(server.id) ?? []}
                            />
                        ))}
                    </div>
                )}

                {activeTab === "library" && <MCPLibraryBrowserBody onInstalled={handleInstalled} />}
            </div>
        </div>
    );
}

function WorkflowStep({ title, children }: { title: string; children: React.ReactNode }) {
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-3">
            <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-primary">{title}</p>
            <p className="mt-2 text-xs leading-5 text-cortex-text-main">{children}</p>
        </div>
    );
}

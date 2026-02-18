"use client";

import React, { useEffect, useState } from "react";
import { Wrench, Plus, BookOpen } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import MCPServerCard from "./MCPServerCard";
import MCPInstallModal from "./MCPInstallModal";
import MCPLibraryBrowser from "./MCPLibraryBrowser";

type Tab = "installed" | "library";

export default function MCPToolRegistry() {
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const isFetching = useCortexStore((s) => s.isFetchingMCPServers);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);
    const installMCPServer = useCortexStore((s) => s.installMCPServer);
    const deleteMCPServer = useCortexStore((s) => s.deleteMCPServer);

    const [isInstallModalOpen, setIsInstallModalOpen] = useState(false);
    const [activeTab, setActiveTab] = useState<Tab>("installed");

    useEffect(() => {
        fetchMCPServers();
    }, [fetchMCPServers]);

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
                        onClick={() => setIsInstallModalOpen(true)}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-cortex-success/10 border border-cortex-success/30 text-xs font-mono font-bold text-cortex-success hover:bg-cortex-success/20 transition-colors"
                    >
                        <Plus className="w-3.5 h-3.5" />
                        INSTALL
                    </button>
                )}
            </div>

            {/* Body */}
            <div className="flex-1 overflow-y-auto">
                {activeTab === "installed" && (
                    <div className="flex flex-col gap-4 p-6 max-w-4xl mx-auto">
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
                                    Click &quot;+ Install&quot; or browse the Library tab.
                                </p>
                            </div>
                        )}

                        {/* Server Cards */}
                        {mcpServers.map((server) => (
                            <MCPServerCard
                                key={server.id}
                                server={server}
                                onDelete={deleteMCPServer}
                            />
                        ))}
                    </div>
                )}

                {activeTab === "library" && <MCPLibraryBrowser />}
            </div>

            {/* Install Modal */}
            <MCPInstallModal
                isOpen={isInstallModalOpen}
                onClose={() => setIsInstallModalOpen(false)}
                onInstall={installMCPServer}
            />
        </div>
    );
}

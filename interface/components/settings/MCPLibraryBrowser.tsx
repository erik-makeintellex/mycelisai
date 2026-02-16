"use client";

import React, { useEffect, useState } from "react";
import { Download, Search, Tag, Loader2 } from "lucide-react";
import { useCortexStore, type MCPLibraryEntry, type MCPLibraryCategory } from "@/store/useCortexStore";

interface EnvModalProps {
    entry: MCPLibraryEntry;
    onInstall: (env: Record<string, string>) => void;
    onClose: () => void;
}

function EnvConfigModal({ entry, onInstall, onClose }: EnvModalProps) {
    const requiredEnv = entry.env ? Object.keys(entry.env) : [];
    const [envValues, setEnvValues] = useState<Record<string, string>>(() => {
        const init: Record<string, string> = {};
        requiredEnv.forEach((k) => { init[k] = ""; });
        return init;
    });

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="bg-cortex-surface border border-cortex-border rounded-xl w-full max-w-md p-6 shadow-2xl">
                <h3 className="text-sm font-mono font-bold text-cortex-text-main mb-1">
                    Configure {entry.name}
                </h3>
                <p className="text-[10px] font-mono text-cortex-text-muted mb-4">
                    Set required environment variables before installing.
                </p>

                <div className="flex flex-col gap-3 mb-5">
                    {requiredEnv.map((key) => (
                        <label key={key} className="flex flex-col gap-1">
                            <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider">
                                {key}
                            </span>
                            <input
                                type="text"
                                value={envValues[key]}
                                onChange={(e) => setEnvValues((v) => ({ ...v, [key]: e.target.value }))}
                                placeholder={`Enter ${key}`}
                                className="bg-cortex-bg border border-cortex-border rounded-lg px-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/40 focus:outline-none focus:ring-1 focus:ring-cortex-primary/50"
                            />
                        </label>
                    ))}
                </div>

                <div className="flex justify-end gap-2">
                    <button
                        onClick={onClose}
                        className="px-3 py-1.5 rounded-lg text-xs font-mono text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={() => onInstall(envValues)}
                        className="px-4 py-1.5 rounded-lg bg-cortex-success/10 border border-cortex-success/30 text-xs font-mono font-bold text-cortex-success hover:bg-cortex-success/20 transition-colors"
                    >
                        Install
                    </button>
                </div>
            </div>
        </div>
    );
}

export default function MCPLibraryBrowser() {
    const library = useCortexStore((s) => s.mcpLibrary);
    const isFetching = useCortexStore((s) => s.isFetchingMCPLibrary);
    const fetchLibrary = useCortexStore((s) => s.fetchMCPLibrary);
    const installFromLibrary = useCortexStore((s) => s.installFromLibrary);
    const mcpServers = useCortexStore((s) => s.mcpServers);

    const [searchQuery, setSearchQuery] = useState("");
    const [installingName, setInstallingName] = useState<string | null>(null);
    const [envModalEntry, setEnvModalEntry] = useState<MCPLibraryEntry | null>(null);

    useEffect(() => {
        fetchLibrary();
    }, [fetchLibrary]);

    const installedNames = new Set(mcpServers.map((s) => s.name));

    const handleInstallClick = (entry: MCPLibraryEntry) => {
        const hasRequiredEnv = entry.env && Object.keys(entry.env).length > 0;
        if (hasRequiredEnv) {
            setEnvModalEntry(entry);
        } else {
            doInstall(entry.name);
        }
    };

    const doInstall = async (name: string, env?: Record<string, string>) => {
        setInstallingName(name);
        await installFromLibrary(name, env);
        setInstallingName(null);
        setEnvModalEntry(null);
    };

    // Filter logic
    const filterEntries = (categories: MCPLibraryCategory[]) => {
        if (!searchQuery.trim()) return categories;
        const q = searchQuery.toLowerCase();
        return categories
            .map((cat) => ({
                ...cat,
                servers: cat.servers.filter(
                    (s) =>
                        s.name.toLowerCase().includes(q) ||
                        s.description.toLowerCase().includes(q) ||
                        s.tags.some((t) => t.toLowerCase().includes(q))
                ),
            }))
            .filter((cat) => cat.servers.length > 0);
    };

    const filtered = filterEntries(library);

    return (
        <div className="flex flex-col gap-4 p-6 max-w-4xl mx-auto">
            {/* Search */}
            <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-cortex-text-muted/50" />
                <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search by name, description, or tag..."
                    className="w-full bg-cortex-surface border border-cortex-border rounded-lg pl-9 pr-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/40 focus:outline-none focus:ring-1 focus:ring-cortex-primary/50"
                />
            </div>

            {/* Loading */}
            {isFetching && library.length === 0 && (
                <div className="flex items-center justify-center py-12">
                    <Loader2 className="w-5 h-5 text-cortex-text-muted animate-spin" />
                </div>
            )}

            {/* Categories */}
            {filtered.map((cat) => (
                <div key={cat.name}>
                    <h3 className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-3">
                        {cat.name}
                    </h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {cat.servers.map((entry) => {
                            const isInstalled = installedNames.has(entry.name);
                            const isInstalling = installingName === entry.name;
                            return (
                                <div
                                    key={entry.name}
                                    className="bg-cortex-surface border border-cortex-border rounded-xl p-4 flex flex-col gap-2"
                                >
                                    <div className="flex items-start justify-between">
                                        <div>
                                            <h4 className="text-xs font-mono font-bold text-cortex-text-main">
                                                {entry.name}
                                            </h4>
                                            <p className="text-[10px] font-mono text-cortex-text-muted mt-0.5 leading-relaxed">
                                                {entry.description}
                                            </p>
                                        </div>
                                        <button
                                            onClick={() => handleInstallClick(entry)}
                                            disabled={isInstalled || isInstalling}
                                            className={`flex-shrink-0 flex items-center gap-1 px-2.5 py-1 rounded-lg text-[10px] font-mono font-bold transition-colors ${
                                                isInstalled
                                                    ? "bg-cortex-success/10 text-cortex-success/60 border border-cortex-success/20 cursor-default"
                                                    : isInstalling
                                                    ? "bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20 cursor-wait"
                                                    : "bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/30 hover:bg-cortex-primary/20"
                                            }`}
                                        >
                                            {isInstalled ? (
                                                "INSTALLED"
                                            ) : isInstalling ? (
                                                <><Loader2 className="w-3 h-3 animate-spin" /> INSTALLING</>
                                            ) : (
                                                <><Download className="w-3 h-3" /> INSTALL</>
                                            )}
                                        </button>
                                    </div>
                                    <div className="flex flex-wrap gap-1">
                                        {entry.tags.map((tag) => (
                                            <span
                                                key={tag}
                                                className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-cortex-bg text-[9px] font-mono text-cortex-text-muted border border-cortex-border"
                                            >
                                                <Tag className="w-2.5 h-2.5" />
                                                {tag}
                                            </span>
                                        ))}
                                    </div>
                                    {entry.env && Object.keys(entry.env).length > 0 && (
                                        <p className="text-[9px] font-mono text-cortex-warning/70">
                                            Requires: {Object.keys(entry.env).join(", ")}
                                        </p>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>
            ))}

            {/* Empty filtered */}
            {!isFetching && filtered.length === 0 && library.length > 0 && (
                <div className="flex flex-col items-center justify-center py-12 text-cortex-text-muted">
                    <Search className="w-8 h-8 mb-2 opacity-20" />
                    <p className="text-xs font-mono">No matches for &quot;{searchQuery}&quot;</p>
                </div>
            )}

            {/* Env Config Modal */}
            {envModalEntry && (
                <EnvConfigModal
                    entry={envModalEntry}
                    onInstall={(env) => doInstall(envModalEntry.name, env)}
                    onClose={() => setEnvModalEntry(null)}
                />
            )}
        </div>
    );
}

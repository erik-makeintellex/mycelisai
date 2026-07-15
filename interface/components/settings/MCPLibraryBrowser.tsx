"use client";

import React, { useEffect, useState } from "react";
import { Download, Search, Tag, Loader2, Settings2 } from "lucide-react";
import { useCortexStore, type MCPLibraryEntry, type MCPLibraryCategory } from "@/store/useCortexStore";
import { EnvConfigModal } from "./MCPLibraryEnvConfigModal";

export default function MCPLibraryBrowser() {
    return <MCPLibraryBrowserBody />;
}

interface MCPLibraryBrowserProps {
    onInstalled?: (name: string) => void;
    initialSearchQuery?: string;
}

export function MCPLibraryBrowserBody({ onInstalled, initialSearchQuery = "" }: MCPLibraryBrowserProps = {}) {
    const library = useCortexStore((s) => s.mcpLibrary);
    const isFetching = useCortexStore((s) => s.isFetchingMCPLibrary);
    const fetchLibrary = useCortexStore((s) => s.fetchMCPLibrary);
    const installFromLibrary = useCortexStore((s) => s.installFromLibrary);
    const mcpServers = useCortexStore((s) => s.mcpServers);

    const [searchQuery, setSearchQuery] = useState(initialSearchQuery);
    const [installingName, setInstallingName] = useState<string | null>(null);
    const [envModalEntry, setEnvModalEntry] = useState<MCPLibraryEntry | null>(null);
    const [installMessage, setInstallMessage] = useState<string | null>(null);

    useEffect(() => {
        fetchLibrary();
    }, [fetchLibrary]);

    const installedNames = new Set(mcpServers.map((s) => s.name));

    const handleInstallClick = (entry: MCPLibraryEntry, isInstalled = false) => {
        const hasRequiredEnv = (entry.environment_variables && entry.environment_variables.length > 0) || (entry.env && Object.keys(entry.env).length > 0);
        if (hasRequiredEnv) {
            setEnvModalEntry(entry);
        } else if (isInstalled) {
            setInstallMessage(`${entry.title ?? entry.name} has no required fields here. Use Servers for status, Access for scopes, and this Library entry to reapply the curated shape when needed.`);
        } else {
            doInstall(entry.name);
        }
    };

    const doInstall = async (name: string, env?: Record<string, string>) => {
        setInstallingName(name);
        setInstallMessage(null);
        const result = await installFromLibrary(name, env);
        if (result.message) {
            setInstallMessage(result.message);
        }
        setInstallingName(null);
        if (result.ok) {
            setEnvModalEntry(null);
            onInstalled?.(name);
        }
    };

    const filterEntries = (categories: MCPLibraryCategory[]) => {
        if (!searchQuery.trim()) return categories;
        const q = searchQuery.toLowerCase();
        return categories
            .map((cat) => ({
                ...cat,
                servers: cat.servers.filter((s) =>
                    s.name.toLowerCase().includes(q) || s.description.toLowerCase().includes(q) || s.tags.some((t) => t.toLowerCase().includes(q))),
            }))
            .filter((cat) => cat.servers.length > 0);
    };

    const filtered = filterEntries(library);

    return (
        <div className="flex flex-col gap-4 p-6 max-w-4xl mx-auto">
            <div className="rounded-xl border border-cortex-success/25 bg-cortex-success/10 px-4 py-3">
                <p className="text-[10px] font-mono font-bold uppercase tracking-wider text-cortex-success">
                    Add capability connector
                </p>
                <p className="mt-1 text-xs leading-5 text-cortex-text-main">
                    Choose an optional connector Soma can use. Search itself may already be built in; connectors add explicit URL reading,
                    service access, private data sources, or team tools. Installed connectors can be configured or reapplied from their cards.
                </p>
            </div>

            {installMessage && (
                <div className="rounded-xl border border-cortex-border bg-cortex-surface px-4 py-3">
                    <p className="text-xs font-mono leading-5 text-cortex-text-main">{installMessage}</p>
                </div>
            )}

            <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-cortex-text-muted/50" />
                <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search MCP servers by name, description, or tag..."
                    className="w-full bg-cortex-surface border border-cortex-border rounded-lg pl-9 pr-3 py-2 text-xs font-mono text-cortex-text-main placeholder:text-cortex-text-muted/40 focus:outline-none focus:ring-1 focus:ring-cortex-primary/50"
                />
            </div>

            {isFetching && library.length === 0 && (
                <div className="flex items-center justify-center py-12">
                    <Loader2 className="w-5 h-5 text-cortex-text-muted animate-spin" />
                </div>
            )}

            {filtered.map((cat) => (
                <div key={cat.name}>
                    <h3 className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wider mb-3">
                        {cat.name}
                    </h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {cat.servers.map((entry) => {
                            const isInstalled = installedNames.has(entry.name);
                            const isInstalling = installingName === entry.name;
                            const primaryPackage = entry.packages?.[0];
                            const versionPolicy = primaryPackage?.version || entry.version;
                            return (
                                <div
                                    key={entry.name}
                                    className="bg-cortex-surface border border-cortex-border rounded-xl p-4 flex flex-col gap-2"
                                >
                                    <div className="flex items-start justify-between">
                                        <div>
                                            <h4 className="text-xs font-mono font-bold text-cortex-text-main">
                                                {entry.title ?? entry.name}
                                            </h4>
                                            <p className="text-[9px] font-mono uppercase tracking-wider text-cortex-text-muted mt-0.5">
                                                {entry.name}{entry.version ? ` · v${entry.version}` : ""}
                                            </p>
                                            <p className="text-[10px] font-mono text-cortex-text-muted mt-0.5 leading-relaxed">
                                                {entry.description}
                                            </p>
                                        </div>
                                        <button
                                            onClick={() => handleInstallClick(entry, isInstalled)}
                                            disabled={isInstalling}
                                            className={`flex-shrink-0 flex items-center gap-1 px-2.5 py-1 rounded-lg text-[10px] font-mono font-bold transition-colors ${
                                                isInstalled
                                                    ? "bg-cortex-success/10 text-cortex-success border border-cortex-success/20 hover:bg-cortex-success/20"
                                                    : isInstalling
                                                    ? "bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20 cursor-wait"
                                                    : "bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/30 hover:bg-cortex-primary/20"
                                            }`}
                                        >
                                            {isInstalled ? (
                                                <><Settings2 className="w-3 h-3" /> CONFIGURE</>
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
                                    {primaryPackage && (
                                        <div className="text-[9px] font-mono text-cortex-text-muted leading-relaxed">
                                            Package: {primaryPackage.identifier}
                                            {primaryPackage.version ? ` @ ${primaryPackage.version}` : ""}
                                            {primaryPackage.transport?.type ? ` · ${primaryPackage.transport.type}` : ""}
                                        </div>
                                    )}
                                    {versionPolicy && (
                                        <div className="text-[9px] font-mono text-cortex-text-muted leading-relaxed">
                                            Version policy: {versionPolicy === "latest" ? "latest (curated upstream tracking)" : versionPolicy}
                                        </div>
                                    )}
                                    <ConfigurationSummary entry={entry} isInstalled={isInstalled} />
                                    <div className="text-[9px] font-mono text-cortex-text-muted leading-relaxed">
                                        Capability binding: {entry.tool_set ?? entry.name} · source mcp · outputs normalize through Managed Exchange.
                                    </div>
                                    {(entry.repository || entry.homepage) && (
                                        <div className="flex flex-wrap gap-2 text-[9px] font-mono text-cortex-text-muted leading-relaxed">
                                            {entry.repository && (
                                                <a
                                                    href={entry.repository}
                                                    target="_blank"
                                                    rel="noreferrer"
                                                    className="text-cortex-primary hover:underline"
                                                >
                                                    Repository
                                                </a>
                                            )}
                                            {entry.homepage && (
                                                <a
                                                    href={entry.homepage}
                                                    target="_blank"
                                                    rel="noreferrer"
                                                    className="text-cortex-primary hover:underline"
                                                >
                                                    Homepage
                                                </a>
                                            )}
                                        </div>
                                    )}
                                    {((entry.environment_variables && entry.environment_variables.length > 0) || (entry.env && Object.keys(entry.env).length > 0)) && (
                                        <p className="text-[9px] font-mono text-cortex-warning/70">
                                            Requires: {(entry.environment_variables?.map((spec) => spec.name) ?? Object.keys(entry.env ?? {})).join(", ")}
                                        </p>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>
            ))}

            {!isFetching && filtered.length === 0 && library.length > 0 && (
                <div className="flex flex-col items-center justify-center py-12 text-cortex-text-muted">
                    <Search className="w-8 h-8 mb-2 opacity-20" />
                    <p className="text-xs font-mono">No matches for &quot;{searchQuery}&quot;</p>
                </div>
            )}

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

function ConfigurationSummary({ entry, isInstalled }: { entry: MCPLibraryEntry; isInstalled: boolean }) {
    const hasEnv = Boolean((entry.environment_variables && entry.environment_variables.length > 0) || (entry.env && Object.keys(entry.env).length > 0));
    const title = configurationTitle(entry, hasEnv);
    const detail = configurationDetail(entry, hasEnv, isInstalled);
    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg/60 px-3 py-2">
            <div className="flex flex-wrap items-center gap-2">
                <span className="text-[9px] font-mono font-bold uppercase tracking-wider text-cortex-text-muted">
                    Configure
                </span>
                {entry.multiple_connections && (
                    <span className="rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-2 py-0.5 text-[9px] font-mono text-cortex-primary">
                        multiple named connections
                    </span>
                )}
                {entry.connection_resource && (
                    <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[9px] font-mono text-cortex-text-muted">
                        {entry.connection_resource}
                    </span>
                )}
            </div>
            <p className="mt-1 text-[10px] font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-1 text-[10px] leading-4 text-cortex-text-muted">{detail}</p>
            {entry.configuration_hint && (
                <p className="mt-2 text-[10px] leading-4 text-cortex-text-muted">{entry.configuration_hint}</p>
            )}
        </div>
    );
}

function configurationTitle(entry: MCPLibraryEntry, hasEnv: boolean): string {
    if (entry.configuration_kind === "connection_profiles") {
        return "Use named data-source connections";
    }
    if (entry.name === "fetch") {
        return "No provider setup required";
    }
    if (hasEnv) {
        return "Requires configuration before use";
    }
    return "Ready to install with default configuration";
}

function configurationDetail(entry: MCPLibraryEntry, hasEnv: boolean, isInstalled: boolean): string {
    if (entry.configuration_kind === "connection_profiles") {
        return isInstalled
            ? "Configure adds or reapplies one connection profile. Use a distinct profile per database and scope it in Access."
            : "Install/configure one connection profile at a time; each profile should be named and scoped so Soma knows which database it may use.";
    }
    if (entry.name === "fetch") {
        return "Fetch reads a URL the user or team already supplied. It is separate from Soma public web search.";
    }
    if (hasEnv) {
        return "Secret values are entered as deployment-managed references or local environment values, then redacted from normal UI.";
    }
    return "Use Servers to inspect status and Access to grant this connector to Everyone, a Group, or a Host.";
}

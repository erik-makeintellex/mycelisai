"use client";

import React, { useEffect, useState } from "react";
import { X, Wrench, Server, Zap, ChevronDown, ChevronRight, Play, Loader2 } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import { toolLabel, WORKSPACE_LABELS } from "@/lib/labels";

export default function ToolsPalette() {
    const isOpen = useCortexStore((s) => s.isToolsPaletteOpen);
    const toggle = useCortexStore((s) => s.toggleToolsPalette);
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const mcpTools = useCortexStore((s) => s.mcpTools);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);
    const fetchMCPTools = useCortexStore((s) => s.fetchMCPTools);

    const [expandedServer, setExpandedServer] = useState<string | null>(null);
    const [executingTool, setExecutingTool] = useState<string | null>(null);
    const [toolResult, setToolResult] = useState<string | null>(null);

    useEffect(() => {
        if (isOpen) {
            fetchMCPServers();
            fetchMCPTools();
        }
    }, [isOpen, fetchMCPServers, fetchMCPTools]);

    if (!isOpen) return null;

    // Group tools by source: internal vs MCP servers
    const internalTools = [
        "consult_council", "delegate_task", "search_memory", "list_teams",
        "list_missions", "get_system_status", "list_available_tools",
        "generate_blueprint", "list_catalogue", "remember", "recall",
        "store_artifact", "generate_image", "publish_signal", "read_signals",
        "read_file", "write_file",
    ];

    async function handleCallTool(serverId: string, toolName: string) {
        setExecutingTool(`${serverId}:${toolName}`);
        setToolResult(null);
        try {
            const res = await fetch(`/api/v1/mcp/servers/${serverId}/tools/${toolName}/call`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ arguments: {} }),
            });
            const text = await res.text();
            setToolResult(text);
        } catch (err: any) {
            setToolResult(`Error: ${err.message}`);
        } finally {
            setExecutingTool(null);
        }
    }

    return (
        <div className="absolute left-0 top-0 bottom-0 w-80 z-40 bg-cortex-surface border-r border-cortex-border shadow-2xl flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between px-3 py-2.5 border-b border-cortex-border">
                <div className="flex items-center gap-2">
                    <Wrench className="w-4 h-4 text-cortex-primary" />
                    <span className="text-xs font-mono font-bold text-cortex-text-main uppercase">
                        {WORKSPACE_LABELS.toolRegistry}
                    </span>
                    <span className="text-[9px] font-mono text-cortex-text-muted px-1.5 py-0.5 bg-cortex-bg rounded border border-cortex-border">
                        {internalTools.length + mcpTools.length}
                    </span>
                </div>
                <button
                    onClick={toggle}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Scrollable content */}
            <div className="flex-1 overflow-y-auto">
                {/* Internal Tools */}
                <div className="p-2">
                    <div
                        className="flex items-center gap-2 px-2 py-1.5 cursor-pointer hover:bg-cortex-bg rounded transition-colors"
                        onClick={() => setExpandedServer(expandedServer === "internal" ? null : "internal")}
                    >
                        {expandedServer === "internal" ? (
                            <ChevronDown className="w-3 h-3 text-cortex-text-muted" />
                        ) : (
                            <ChevronRight className="w-3 h-3 text-cortex-text-muted" />
                        )}
                        <Zap className="w-3.5 h-3.5 text-cortex-success" />
                        <span className="text-[11px] font-mono font-bold text-cortex-text-main">
                            {WORKSPACE_LABELS.internalTools}
                        </span>
                        <span className="text-[9px] font-mono text-cortex-text-muted ml-auto">
                            {internalTools.length}
                        </span>
                    </div>
                    {expandedServer === "internal" && (
                        <div className="ml-5 mt-1 space-y-0.5">
                            {internalTools.map((name) => (
                                <div
                                    key={name}
                                    className="flex items-center gap-2 px-2 py-1 rounded hover:bg-cortex-bg transition-colors group"
                                >
                                    <div className="w-1.5 h-1.5 rounded-full bg-cortex-success flex-shrink-0" />
                                    <div className="min-w-0 flex-1">
                                        <span className="text-[10px] font-mono text-cortex-text-main truncate block">
                                            {toolLabel(name)}
                                        </span>
                                        <span className="text-[8px] font-mono text-cortex-text-muted/50 truncate block">
                                            {name}
                                        </span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* MCP Servers */}
                {mcpServers.map((server) => (
                    <div key={server.id} className="p-2 pt-0">
                        <div
                            className="flex items-center gap-2 px-2 py-1.5 cursor-pointer hover:bg-cortex-bg rounded transition-colors"
                            onClick={() =>
                                setExpandedServer(expandedServer === server.id ? null : server.id)
                            }
                        >
                            {expandedServer === server.id ? (
                                <ChevronDown className="w-3 h-3 text-cortex-text-muted" />
                            ) : (
                                <ChevronRight className="w-3 h-3 text-cortex-text-muted" />
                            )}
                            <Server className="w-3.5 h-3.5 text-cortex-primary" />
                            <span className="text-[11px] font-mono font-bold text-cortex-text-main truncate">
                                {server.name}
                            </span>
                            <div className="flex items-center gap-1.5 ml-auto flex-shrink-0">
                                <span className="text-[9px] font-mono text-cortex-text-muted">
                                    {server.tools?.length ?? 0}
                                </span>
                                <div
                                    className={`w-1.5 h-1.5 rounded-full ${
                                        server.status === "connected"
                                            ? "bg-cortex-success"
                                            : "bg-cortex-danger"
                                    }`}
                                />
                            </div>
                        </div>
                        {expandedServer === server.id && server.tools && (
                            <div className="ml-5 mt-1 space-y-0.5">
                                {server.tools.map((tool) => (
                                    <div
                                        key={tool.name}
                                        className="flex items-center gap-2 px-2 py-1 rounded hover:bg-cortex-bg transition-colors group"
                                    >
                                        <div className="w-1.5 h-1.5 rounded-full bg-cortex-primary flex-shrink-0" />
                                        <div className="min-w-0 flex-1">
                                            <span className="text-[10px] font-mono text-cortex-text-main truncate block">
                                                {tool.name}
                                            </span>
                                            {tool.description && (
                                                <span className="text-[9px] text-cortex-text-muted truncate block">
                                                    {tool.description}
                                                </span>
                                            )}
                                        </div>
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleCallTool(server.id, tool.name);
                                            }}
                                            className="p-0.5 rounded opacity-0 group-hover:opacity-100 hover:bg-cortex-primary/20 text-cortex-primary transition-all flex-shrink-0"
                                            title="Execute tool"
                                        >
                                            {executingTool === `${server.id}:${tool.name}` ? (
                                                <Loader2 className="w-3 h-3 animate-spin" />
                                            ) : (
                                                <Play className="w-3 h-3" />
                                            )}
                                        </button>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                ))}

                {mcpServers.length === 0 && (
                    <div className="px-4 py-6 text-center">
                        <Server className="w-8 h-8 mx-auto mb-2 text-cortex-text-muted opacity-30" />
                        <p className="text-[10px] font-mono text-cortex-text-muted">
                            No MCP servers installed
                        </p>
                        <p className="text-[9px] font-mono text-cortex-text-muted mt-0.5">
                            Install from Settings &rarr; Tools
                        </p>
                    </div>
                )}
            </div>

            {/* Tool result footer */}
            {toolResult && (
                <div className="border-t border-cortex-border p-2 max-h-32 overflow-y-auto">
                    <div className="flex items-center justify-between mb-1">
                        <span className="text-[9px] font-mono font-bold text-cortex-text-muted uppercase">
                            Result
                        </span>
                        <button
                            onClick={() => setToolResult(null)}
                            className="text-cortex-text-muted hover:text-cortex-text-main"
                        >
                            <X className="w-3 h-3" />
                        </button>
                    </div>
                    <pre className="text-[9px] font-mono text-cortex-text-main whitespace-pre-wrap break-all bg-cortex-bg rounded p-1.5 border border-cortex-border">
                        {toolResult}
                    </pre>
                </div>
            )}
        </div>
    );
}

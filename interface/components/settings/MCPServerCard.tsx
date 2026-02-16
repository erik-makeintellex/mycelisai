"use client";

import React, { useState } from "react";
import { ChevronRight, Trash2, Server, Wrench } from "lucide-react";
import type { MCPServerWithTools } from "@/store/useCortexStore";

// ── Status Dot Color ──────────────────────────────────────────

function statusDotClass(status: string, error?: string): string {
    if (error) return "bg-cortex-danger shadow-[0_0_6px_var(--color-cortex-danger)]";
    if (status === "connected") return "bg-cortex-success shadow-[0_0_6px_var(--color-cortex-success)]";
    return "bg-cortex-warning shadow-[0_0_6px_var(--color-cortex-warning)]";
}

// ── Props ─────────────────────────────────────────────────────

interface MCPServerCardProps {
    server: MCPServerWithTools;
    onDelete: (id: string) => void;
}

export default function MCPServerCard({ server, onDelete }: MCPServerCardProps) {
    const [isExpanded, setIsExpanded] = useState(false);
    const [isConfirmingDelete, setIsConfirmingDelete] = useState(false);

    const toolCount = server.tools?.length ?? 0;

    function handleDelete(e: React.MouseEvent) {
        e.stopPropagation();
        if (isConfirmingDelete) {
            onDelete(server.id);
            setIsConfirmingDelete(false);
        } else {
            setIsConfirmingDelete(true);
            // Auto-reset confirmation after 3 seconds
            setTimeout(() => setIsConfirmingDelete(false), 3000);
        }
    }

    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-xl overflow-hidden transition-all">
            {/* Header Row */}
            <button
                onClick={() => setIsExpanded((prev) => !prev)}
                className="w-full flex items-center gap-3 p-4 hover:bg-cortex-bg/40 transition-colors cursor-pointer"
            >
                {/* Status Dot */}
                <span className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${statusDotClass(server.status, server.error)}`} />

                {/* Server Name */}
                <Server className="w-4 h-4 text-cortex-text-muted flex-shrink-0" />
                <span className="text-sm font-bold text-cortex-text-main truncate">
                    {server.name}
                </span>

                {/* Transport Badge */}
                <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20 flex-shrink-0">
                    {server.transport}
                </span>

                {/* Tool Count Chip */}
                <span className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-cortex-info/10 text-cortex-info border border-cortex-info/20 flex-shrink-0">
                    {toolCount} tool{toolCount !== 1 ? "s" : ""}
                </span>

                {/* Spacer */}
                <span className="flex-1" />

                {/* Delete Button */}
                <span
                    role="button"
                    tabIndex={0}
                    onClick={handleDelete}
                    onKeyDown={(e) => { if (e.key === "Enter") handleDelete(e as unknown as React.MouseEvent); }}
                    className={`p-1.5 rounded-md transition-colors flex-shrink-0 ${
                        isConfirmingDelete
                            ? "bg-cortex-danger/20 text-cortex-danger"
                            : "hover:bg-cortex-danger/10 text-cortex-text-muted hover:text-cortex-danger"
                    }`}
                    title={isConfirmingDelete ? "Click again to confirm" : "Delete server"}
                >
                    <Trash2 className="w-3.5 h-3.5" />
                </span>

                {/* Expand Chevron */}
                <ChevronRight className={`w-4 h-4 text-cortex-text-muted transition-transform flex-shrink-0 ${
                    isExpanded ? "rotate-90" : ""
                }`} />
            </button>

            {/* Error Message */}
            {server.error && (
                <div className="px-4 pb-2 -mt-1">
                    <p className="text-[11px] font-mono text-cortex-danger bg-cortex-danger/5 border border-cortex-danger/20 rounded-md px-2.5 py-1.5">
                        {server.error}
                    </p>
                </div>
            )}

            {/* Expanded: Tool List */}
            {isExpanded && (
                <div className="px-4 pb-4 border-t border-cortex-border/50">
                    {toolCount === 0 ? (
                        <div className="flex flex-col items-center justify-center py-6 text-cortex-text-muted">
                            <Wrench className="w-6 h-6 mb-1.5 opacity-20" />
                            <p className="text-[10px] font-mono">No tools registered</p>
                        </div>
                    ) : (
                        <div className="pt-3 space-y-0.5">
                            {server.tools.map((tool, idx) => {
                                const isLast = idx === server.tools.length - 1;
                                const prefix = isLast ? "\u2514\u2500" : "\u251C\u2500";

                                return (
                                    <div key={tool.id} className="flex items-start gap-2 group">
                                        <span className="text-cortex-text-muted font-mono text-xs leading-5 select-none flex-shrink-0 w-5">
                                            {prefix}
                                        </span>
                                        <div className="min-w-0 py-0.5">
                                            <span className="text-xs font-bold font-mono text-cortex-text-main">
                                                {tool.name}
                                            </span>
                                            {tool.description && (
                                                <p className="text-[10px] font-mono text-cortex-text-muted leading-tight mt-0.5">
                                                    {tool.description}
                                                </p>
                                            )}
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

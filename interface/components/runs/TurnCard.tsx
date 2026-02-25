"use client";

import React, { useState } from "react";
import { ChevronDown, ChevronRight, AlertTriangle, Bot, User, Wrench, Terminal, MessageSquare } from "lucide-react";
import type { ConversationTurn } from "@/types/conversations";
import { TURN_ROLE_STYLES } from "@/types/conversations";
import { brainDisplayName } from "@/lib/labels";

function relativeTime(iso: string): string {
    const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
    if (diff < 60) return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    return `${Math.floor(diff / 3600)}h ago`;
}

function roleIcon(role: ConversationTurn["role"]): React.ReactNode {
    switch (role) {
        case "system":
            return <Terminal className="w-3 h-3" />;
        case "user":
            return <User className="w-3 h-3" />;
        case "assistant":
            return <Bot className="w-3 h-3" />;
        case "tool_call":
            return <Wrench className="w-3 h-3" />;
        case "tool_result":
            return <Wrench className="w-3 h-3" />;
        case "interjection":
            return <AlertTriangle className="w-3 h-3" />;
        default:
            return <MessageSquare className="w-3 h-3" />;
    }
}

interface Props {
    turn: ConversationTurn;
}

export default function TurnCard({ turn }: Props) {
    const style = TURN_ROLE_STYLES[turn.role] ?? TURN_ROLE_STYLES.system;
    const [expanded, setExpanded] = useState(turn.role !== "system");
    const [argsExpanded, setArgsExpanded] = useState(false);

    const isSystem = turn.role === "system";
    const isToolCall = turn.role === "tool_call";
    const isToolResult = turn.role === "tool_result";
    const isInterjection = turn.role === "interjection";

    return (
        <div
            className={`border-l-2 ${style.border} bg-cortex-surface rounded-r-md mb-2 ${isToolResult ? "ml-6" : ""}`}
        >
            {/* Header row */}
            <div
                className={`flex items-center gap-2 px-3 py-2 ${isSystem ? "cursor-pointer select-none" : ""}`}
                onClick={isSystem ? () => setExpanded((p) => !p) : undefined}
            >
                {/* Role icon + label */}
                <span className={`flex items-center gap-1.5 text-[10px] font-mono font-bold ${style.labelColor}`}>
                    {roleIcon(turn.role)}
                    {isInterjection ? (
                        <span className="text-red-400 font-bold uppercase tracking-wide">
                            Operator Interjection
                        </span>
                    ) : (
                        style.label
                    )}
                </span>

                {/* Agent chip */}
                {turn.agent_id && (
                    <span className="text-[9px] font-mono text-cortex-text-muted bg-cortex-bg border border-cortex-border px-1.5 py-0.5 rounded">
                        {turn.agent_id}
                    </span>
                )}

                {/* Tool name badge (for tool_call / tool_result) */}
                {(isToolCall || isToolResult) && turn.tool_name && (
                    <span className="text-[9px] font-mono text-violet-300 bg-violet-500/10 border border-violet-500/30 px-1.5 py-0.5 rounded">
                        {turn.tool_name}
                    </span>
                )}

                {/* Provider/model badge (for assistant) */}
                {turn.role === "assistant" && (turn.provider_id || turn.model_used) && (
                    <span className="text-[9px] font-mono text-cortex-text-muted bg-cortex-bg border border-cortex-border px-1.5 py-0.5 rounded">
                        {turn.provider_id ? brainDisplayName(turn.provider_id) : ""}
                        {turn.provider_id && turn.model_used ? " / " : ""}
                        {turn.model_used ?? ""}
                    </span>
                )}

                {/* Consultation badge (for tool_result) */}
                {isToolResult && turn.consultation_of && (
                    <span className="text-[9px] font-mono text-cyan-300 bg-cyan-500/10 border border-cyan-500/30 px-1.5 py-0.5 rounded">
                        consulted: {turn.consultation_of}
                    </span>
                )}

                {/* Expand/collapse toggle for system messages */}
                {isSystem && (
                    <span className="text-cortex-text-muted ml-auto">
                        {expanded
                            ? <ChevronDown className="w-3 h-3" />
                            : <ChevronRight className="w-3 h-3" />
                        }
                    </span>
                )}

                {/* Timestamp */}
                <span className="text-[9px] font-mono text-cortex-text-muted/60 ml-auto">
                    {relativeTime(turn.created_at)}
                </span>
            </div>

            {/* Content body */}
            {expanded && (
                <div className="px-3 pb-2">
                    <div className="text-[11px] font-mono text-cortex-text-main leading-relaxed whitespace-pre-wrap break-words">
                        {turn.content}
                    </div>

                    {/* Tool args (collapsed by default for tool_call) */}
                    {isToolCall && turn.tool_args && Object.keys(turn.tool_args).length > 0 && (
                        <div className="mt-1.5">
                            <button
                                onClick={() => setArgsExpanded((p) => !p)}
                                className="flex items-center gap-1 text-[9px] font-mono text-cortex-text-muted hover:text-cortex-primary transition-colors"
                            >
                                {argsExpanded
                                    ? <ChevronDown className="w-2.5 h-2.5" />
                                    : <ChevronRight className="w-2.5 h-2.5" />
                                }
                                arguments
                            </button>
                            {argsExpanded && (
                                <pre className="mt-1 bg-cortex-bg border border-cortex-border rounded p-2 text-[9px] font-mono text-cortex-text-muted overflow-x-auto max-h-40 overflow-y-auto leading-relaxed">
                                    {JSON.stringify(turn.tool_args, null, 2)}
                                </pre>
                            )}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

"use client";

import React, { useState } from 'react';
import { X, ChevronDown, ChevronRight, Wrench, Radio } from 'lucide-react';
import type { TeamDetailEntry, TeamDetailAgentEntry } from '@/store/useCortexStore';

const statusLabel: Record<number, { text: string; color: string }> = {
    0: { text: 'Offline', color: 'bg-cortex-text-muted/40' },
    1: { text: 'Idle', color: 'bg-cortex-success' },
    2: { text: 'Busy', color: 'bg-cortex-primary animate-pulse' },
    3: { text: 'Error', color: 'bg-red-500' },
};

function relativeTime(ts: string): string {
    if (!ts) return 'never';
    const diff = Date.now() - new Date(ts).getTime();
    if (diff < 5000) return 'just now';
    if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`;
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    return `${Math.floor(diff / 3600000)}h ago`;
}

interface TeamDetailDrawerProps {
    team: TeamDetailEntry;
    onClose: () => void;
}

export default function TeamDetailDrawer({ team, onClose }: TeamDetailDrawerProps) {
    const [expandedAgent, setExpandedAgent] = useState<string | null>(null);

    return (
        <div className="absolute right-0 top-0 bottom-0 w-[480px] z-40 bg-cortex-surface border-l border-cortex-border shadow-2xl flex flex-col">
            {/* Header */}
            <div className="h-12 flex items-center justify-between px-4 border-b border-cortex-border bg-cortex-surface/50 backdrop-blur-sm flex-shrink-0">
                <div className="flex items-center gap-2">
                    <span className="text-xs font-mono font-bold text-cortex-text-main uppercase truncate">
                        {team.name}
                    </span>
                    <span className="text-[9px] font-mono uppercase px-1.5 py-0.5 rounded bg-cortex-primary/15 text-cortex-primary border border-cortex-primary/30">
                        {team.role}
                    </span>
                    <span
                        className={`text-[9px] font-mono uppercase px-1.5 py-0.5 rounded ${
                            team.type === 'standing'
                                ? 'bg-cortex-text-muted/10 text-cortex-text-muted border border-cortex-border'
                                : 'bg-cortex-success/10 text-cortex-success border border-cortex-success/30'
                        }`}
                    >
                        {team.type}
                    </span>
                </div>
                <button
                    onClick={onClose}
                    className="p-1 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>

            {/* Scrollable body */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {/* Mission context */}
                {team.type === 'mission' && team.mission_id && (
                    <div className="rounded-lg bg-cortex-bg border border-cortex-border p-3 space-y-1">
                        <div className="text-[10px] font-mono uppercase text-cortex-text-muted">
                            Mission
                        </div>
                        <div className="text-[10px] font-mono text-cortex-primary truncate">
                            {team.mission_id}
                        </div>
                        {team.mission_intent && (
                            <div className="text-xs font-mono text-cortex-text-main">
                                {team.mission_intent}
                            </div>
                        )}
                    </div>
                )}

                {/* Input topics */}
                {team.inputs.length > 0 && (
                    <div>
                        <div className="text-[10px] font-mono uppercase text-cortex-text-muted mb-1.5">
                            Input Topics
                        </div>
                        <div className="flex flex-wrap gap-1">
                            {team.inputs.map((t) => (
                                <span
                                    key={t}
                                    className="inline-flex items-center gap-1 text-[10px] font-mono px-1.5 py-0.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-muted"
                                >
                                    <Radio className="w-2.5 h-2.5" />
                                    {t}
                                </span>
                            ))}
                        </div>
                    </div>
                )}

                {/* Delivery topics */}
                {team.deliveries.length > 0 && (
                    <div>
                        <div className="text-[10px] font-mono uppercase text-cortex-text-muted mb-1.5">
                            Delivery Topics
                        </div>
                        <div className="flex flex-wrap gap-1">
                            {team.deliveries.map((t) => (
                                <span
                                    key={t}
                                    className="inline-flex items-center gap-1 text-[10px] font-mono px-1.5 py-0.5 rounded bg-cortex-success/10 border border-cortex-success/20 text-cortex-success"
                                >
                                    <Radio className="w-2.5 h-2.5" />
                                    {t}
                                </span>
                            ))}
                        </div>
                    </div>
                )}

                {/* Agent Roster */}
                <div>
                    <div className="text-[10px] font-mono uppercase text-cortex-text-muted mb-2">
                        Agent Roster ({team.agents.length})
                    </div>
                    <div className="space-y-1">
                        {team.agents.map((agent) => (
                            <AgentRow
                                key={agent.id}
                                agent={agent}
                                isExpanded={expandedAgent === agent.id}
                                onToggle={() =>
                                    setExpandedAgent(expandedAgent === agent.id ? null : agent.id)
                                }
                            />
                        ))}
                        {team.agents.length === 0 && (
                            <div className="text-xs font-mono text-cortex-text-muted/50 py-4 text-center">
                                No agents in this team
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

function AgentRow({
    agent,
    isExpanded,
    onToggle,
}: {
    agent: TeamDetailAgentEntry;
    isExpanded: boolean;
    onToggle: () => void;
}) {
    const st = statusLabel[agent.status] ?? statusLabel[0];

    return (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg overflow-hidden">
            {/* Summary row */}
            <button
                onClick={onToggle}
                className="w-full flex items-center gap-2 px-3 py-2 text-left hover:bg-cortex-surface/50 transition-colors"
            >
                {isExpanded ? (
                    <ChevronDown className="w-3 h-3 text-cortex-text-muted flex-shrink-0" />
                ) : (
                    <ChevronRight className="w-3 h-3 text-cortex-text-muted flex-shrink-0" />
                )}
                <span className={`w-2 h-2 rounded-full flex-shrink-0 ${st.color}`} />
                <span className="text-xs font-mono font-bold text-cortex-text-main truncate flex-1">
                    {agent.id}
                </span>
                <span className="text-[9px] font-mono text-cortex-text-muted">
                    {agent.role}
                </span>
                {agent.model && (
                    <span className="text-[8px] font-mono px-1 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20 truncate max-w-[100px]">
                        {agent.model}
                    </span>
                )}
                {agent.tools.length > 0 && (
                    <span className="text-[8px] font-mono px-1 py-0.5 rounded bg-cortex-surface text-cortex-text-muted border border-cortex-border flex items-center gap-0.5">
                        <Wrench className="w-2.5 h-2.5" />
                        {agent.tools.length}
                    </span>
                )}
                <span className="text-[8px] font-mono text-cortex-text-muted/60">
                    {relativeTime(agent.last_heartbeat)}
                </span>
            </button>

            {/* Expanded detail */}
            {isExpanded && (
                <div className="px-3 pb-3 pt-1 border-t border-cortex-border/50 space-y-3">
                    {/* System prompt */}
                    {agent.system_prompt && (
                        <div>
                            <div className="text-[9px] font-mono uppercase text-cortex-text-muted mb-1">
                                System Prompt
                            </div>
                            <div className="text-[10px] font-mono text-cortex-text-main bg-cortex-surface rounded p-2 max-h-32 overflow-y-auto whitespace-pre-wrap border border-cortex-border/50">
                                {agent.system_prompt}
                            </div>
                        </div>
                    )}

                    {/* Tools */}
                    {agent.tools.length > 0 && (
                        <div>
                            <div className="text-[9px] font-mono uppercase text-cortex-text-muted mb-1">
                                Tools ({agent.tools.length})
                            </div>
                            <div className="flex flex-wrap gap-1">
                                {agent.tools.map((t) => (
                                    <span
                                        key={t}
                                        className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/20"
                                    >
                                        {t}
                                    </span>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Model */}
                    {agent.model && (
                        <div className="flex items-center gap-2">
                            <span className="text-[9px] font-mono uppercase text-cortex-text-muted">
                                Model:
                            </span>
                            <span className="text-[10px] font-mono text-cortex-text-main">
                                {agent.model}
                            </span>
                        </div>
                    )}

                    {/* Status */}
                    <div className="flex items-center gap-2">
                        <span className="text-[9px] font-mono uppercase text-cortex-text-muted">
                            Status:
                        </span>
                        <span className={`w-2 h-2 rounded-full ${st.color}`} />
                        <span className="text-[10px] font-mono text-cortex-text-main">
                            {st.text}
                        </span>
                    </div>
                </div>
            )}
        </div>
    );
}

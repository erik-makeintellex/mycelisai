"use client";

import { useState } from "react";
import type { TeamDetailEntry, TeamDetailAgentEntry } from "@/store/useCortexStore";
import { Bot, Circle, ChevronDown, ChevronRight, Wrench, Clock, Cpu, Square, AlertTriangle } from "lucide-react";

const AGENT_STATUS: Record<number, { label: string; color: string; dot: string }> = {
    0: { label: "OFFLINE", color: "text-cortex-text-muted", dot: "text-cortex-text-muted" },
    1: { label: "IDLE", color: "text-cortex-success", dot: "text-cortex-success" },
    2: { label: "BUSY", color: "text-cortex-primary", dot: "text-cortex-primary" },
    3: { label: "ERROR", color: "text-red-400", dot: "text-red-400" },
};

function relativeTime(ts: string): string {
    if (!ts) return "never";
    const diff = Date.now() - new Date(ts).getTime();
    if (diff < 0) return "just now";
    const secs = Math.floor(diff / 1000);
    if (secs < 60) return `${secs}s ago`;
    const mins = Math.floor(secs / 60);
    if (mins < 60) return `${mins}m ago`;
    const hrs = Math.floor(mins / 60);
    return `${hrs}h ago`;
}

interface AgentPanelProps {
    team: TeamDetailEntry | null;
}

export default function AgentPanel({ team }: AgentPanelProps) {
    const [expandedId, setExpandedId] = useState<string | null>(null);
    const [checked, setChecked] = useState<Set<string>>(new Set());

    const toggleExpand = (id: string) =>
        setExpandedId((prev) => (prev === id ? null : id));

    const toggleCheck = (id: string) => {
        setChecked((prev) => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });
    };

    const toggleAll = () => {
        if (!team) return;
        if (checked.size === team.agents.length) {
            setChecked(new Set());
        } else {
            setChecked(new Set(team.agents.map((a) => a.id)));
        }
    };

    // Empty state — no team selected
    if (!team) {
        return (
            <div className="bg-cortex-bg border border-cortex-border rounded-xl h-full flex flex-col items-center justify-center gap-3">
                <Bot size={28} className="text-cortex-text-muted/30" />
                <span className="text-xs text-cortex-text-muted font-mono">
                    Select a team to inspect agents
                </span>
            </div>
        );
    }

    const allChecked = team.agents.length > 0 && checked.size === team.agents.length;
    const someChecked = checked.size > 0;

    return (
        <div className="bg-cortex-bg border border-cortex-border rounded-xl h-full flex flex-col overflow-hidden">
            {/* Header */}
            <div className="px-4 py-3 border-b border-cortex-border flex items-center justify-between flex-shrink-0">
                <div className="flex items-center gap-3 min-w-0">
                    <h3 className="text-xs font-bold text-cortex-text-main font-mono truncate">
                        {team.name}
                    </h3>
                    <span className={`text-[9px] font-mono px-1.5 py-0.5 rounded border flex-shrink-0 ${
                        team.type === "standing"
                            ? "text-cortex-text-muted border-cortex-border bg-cortex-surface"
                            : "text-cortex-primary border-cortex-primary/30 bg-cortex-primary/10"
                    }`}>
                        {team.type.toUpperCase()}
                    </span>
                    <span className="text-[10px] text-cortex-text-muted font-mono flex-shrink-0">
                        {team.agents.length} agent{team.agents.length !== 1 ? "s" : ""}
                    </span>
                </div>

                {/* Bulk action bar — visible when checkboxes selected */}
                {someChecked && (
                    <div className="flex items-center gap-2 flex-shrink-0">
                        <span className="text-[9px] text-cortex-text-muted font-mono">
                            {checked.size} selected
                        </span>
                        <button className="flex items-center gap-1 px-2 py-1 rounded text-[10px] font-mono font-bold bg-red-500/10 text-red-400 border border-red-500/20 hover:bg-red-500/20 transition-colors">
                            <Square size={9} />
                            STOP
                        </button>
                    </div>
                )}
            </div>

            {/* Mission intent */}
            {team.mission_intent && (
                <div className="px-4 py-2 border-b border-cortex-border/50 bg-cortex-surface/30">
                    <p className="text-[10px] text-cortex-text-muted font-mono truncate">
                        {team.mission_intent}
                    </p>
                </div>
            )}

            {/* Column header */}
            <div className="flex items-center gap-3 px-4 py-1.5 border-b border-cortex-border/50 bg-cortex-surface/20 flex-shrink-0">
                <input
                    type="checkbox"
                    checked={allChecked}
                    onChange={toggleAll}
                    className="w-3.5 h-3.5 rounded border-cortex-border accent-cortex-primary flex-shrink-0"
                />
                <span className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider flex-1">Agent</span>
                <span className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider w-24 text-right">Role</span>
                <span className="text-[9px] font-mono text-cortex-text-muted uppercase tracking-wider w-16 text-right">Status</span>
            </div>

            {/* Agent rows */}
            <div className="flex-1 overflow-y-auto custom-scrollbar">
                {team.agents.length === 0 && (
                    <div className="text-center py-8 text-cortex-text-muted text-xs italic font-mono">
                        No agents in this team.
                    </div>
                )}

                {team.agents.map((agent) => {
                    const st = AGENT_STATUS[agent.status] ?? AGENT_STATUS[0];
                    const isExpanded = expandedId === agent.id;
                    const isChecked = checked.has(agent.id);

                    return (
                        <div key={agent.id} className="border-b border-cortex-border/30 last:border-b-0">
                            {/* Agent Row */}
                            <div className="flex items-center gap-3 px-4 py-2 hover:bg-cortex-surface/20 transition-colors">
                                <input
                                    type="checkbox"
                                    checked={isChecked}
                                    onChange={() => toggleCheck(agent.id)}
                                    className="w-3.5 h-3.5 rounded border-cortex-border accent-cortex-primary flex-shrink-0"
                                />

                                <button
                                    onClick={() => toggleExpand(agent.id)}
                                    className="flex items-center gap-2 flex-1 min-w-0 text-left"
                                >
                                    {isExpanded
                                        ? <ChevronDown size={12} className="text-cortex-text-muted flex-shrink-0" />
                                        : <ChevronRight size={12} className="text-cortex-text-muted flex-shrink-0" />
                                    }
                                    <span className="text-[11px] font-mono font-semibold text-cortex-text-main truncate">
                                        {agent.id}
                                    </span>
                                </button>

                                <span className="text-[10px] font-mono text-cortex-text-muted w-24 text-right truncate">
                                    {agent.role}
                                </span>

                                <div className="flex items-center gap-1.5 w-16 justify-end flex-shrink-0">
                                    <Circle size={6} className={`fill-current ${st.dot}`} />
                                    <span className={`text-[9px] font-mono font-bold ${st.color}`}>
                                        {st.label}
                                    </span>
                                </div>
                            </div>

                            {/* Expanded detail */}
                            {isExpanded && (
                                <AgentDetail agent={agent} />
                            )}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

function AgentDetail({ agent }: { agent: TeamDetailAgentEntry }) {
    return (
        <div className="px-4 py-3 bg-cortex-surface/30 border-t border-cortex-border/30 ml-10 space-y-2.5">
            {/* Model */}
            <div className="flex items-center gap-2">
                <Cpu size={10} className="text-cortex-text-muted flex-shrink-0" />
                <span className="text-[9px] font-mono text-cortex-text-muted uppercase w-14">Model</span>
                <span className="text-[10px] font-mono text-cortex-text-main">
                    {agent.model || "—"}
                </span>
            </div>

            {/* Last heartbeat */}
            <div className="flex items-center gap-2">
                <Clock size={10} className="text-cortex-text-muted flex-shrink-0" />
                <span className="text-[9px] font-mono text-cortex-text-muted uppercase w-14">Pulse</span>
                <span className="text-[10px] font-mono text-cortex-text-main">
                    {relativeTime(agent.last_heartbeat)}
                </span>
            </div>

            {/* Tools */}
            {agent.tools && agent.tools.length > 0 && (
                <div className="flex items-start gap-2">
                    <Wrench size={10} className="text-cortex-text-muted flex-shrink-0 mt-0.5" />
                    <span className="text-[9px] font-mono text-cortex-text-muted uppercase w-14 flex-shrink-0">Tools</span>
                    <div className="flex flex-wrap gap-1">
                        {agent.tools.map((tool) => (
                            <span
                                key={tool}
                                className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-muted"
                            >
                                {tool}
                            </span>
                        ))}
                    </div>
                </div>
            )}

            {/* System prompt (truncated) */}
            {agent.system_prompt && (
                <div className="flex items-start gap-2">
                    <AlertTriangle size={10} className="text-cortex-text-muted flex-shrink-0 mt-0.5" />
                    <span className="text-[9px] font-mono text-cortex-text-muted uppercase w-14 flex-shrink-0">Prompt</span>
                    <p className="text-[10px] font-mono text-cortex-text-muted/80 line-clamp-3 leading-relaxed">
                        {agent.system_prompt}
                    </p>
                </div>
            )}
        </div>
    );
}

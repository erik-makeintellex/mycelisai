"use client";

import React, { useRef, useEffect, useMemo } from 'react';
import { ArrowLeft, Users, MessageSquare, FileCheck, Bot, BarChart3 } from 'lucide-react';
import { useCortexStore, type StreamSignal } from '@/store/useCortexStore';
import { ChartRenderer, type MycelisChartSpec } from '@/components/charts';

interface SquadRoomProps {
    teamId: string;
}

/** Resolve the team name and member agent IDs from the blueprint + nodes */
function useTeamInfo(teamId: string) {
    const blueprint = useCortexStore((s) => s.blueprint);
    const nodes = useCortexStore((s) => s.nodes);

    return useMemo(() => {
        // Find the team label node (e.g. "team-0-label") to get the display name
        const labelNode = nodes.find((n) => n.id === `${teamId}-label`);
        const teamName = labelNode?.data?.label ?? teamId;

        // Collect agent IDs that are children of this team group node
        const agentIds = nodes
            .filter((n) => n.parentNode === teamId && n.type === 'agentNode')
            .map((n) => n.data?.label as string)
            .filter(Boolean);

        // Also get agent labels from the blueprint for richer display
        const teamIndex = parseInt(teamId.replace('team-', ''), 10);
        const blueprintTeam = blueprint?.teams?.[teamIndex];

        return { teamName, agentIds, agentCount: agentIds.length, blueprintTeam };
    }, [teamId, blueprint, nodes]);
}

/** Color rules for debate feed entries */
function debateColor(type?: string): { badge: string; text: string } {
    switch (type) {
        case 'thought':
        case 'cognitive':
            return { badge: 'bg-cyan-500/20 text-cyan-400', text: 'text-cyan-300/80' };
        case 'artifact':
        case 'output':
            return { badge: 'bg-emerald-500/20 text-emerald-400', text: 'text-emerald-300/80' };
        case 'error':
            return { badge: 'bg-red-500/20 text-red-400', text: 'text-red-300/80' };
        case 'governance':
            return { badge: 'bg-amber-500/20 text-amber-400', text: 'text-amber-300/80' };
        default:
            return { badge: 'bg-cortex-text-muted/20 text-cortex-text-muted', text: 'text-cortex-text-muted' };
    }
}

function DebateEntry({ signal }: { signal: StreamSignal }) {
    const colors = debateColor(signal.type);
    const message = signal.message ?? JSON.stringify(signal.payload ?? {});

    return (
        <div className="px-4 py-3 border-b border-cortex-border/50 hover:bg-cortex-bg/40 transition-colors">
            <div className="flex items-center gap-2 mb-1.5">
                <Bot className="w-3 h-3 text-cortex-text-muted flex-shrink-0" />
                <span className="text-[11px] font-mono font-bold text-cortex-text-main">
                    {signal.source ?? 'unknown'}
                </span>
                <span className={`text-[9px] font-mono font-bold px-1.5 py-0.5 rounded ${colors.badge}`}>
                    {signal.type ?? 'event'}
                </span>
                <span className="text-[9px] font-mono text-cortex-text-muted/60 ml-auto">
                    {signal.timestamp
                        ? new Date(signal.timestamp).toLocaleTimeString([], {
                              hour: '2-digit',
                              minute: '2-digit',
                              second: '2-digit',
                          })
                        : 'now'}
                </span>
            </div>
            <p className={`text-[11px] font-mono leading-relaxed pl-5 ${colors.text}`}>
                {message}
            </p>
        </div>
    );
}

function ProofCard({ signal }: { signal: StreamSignal }) {
    const message = signal.message ?? JSON.stringify(signal.payload ?? {});

    // Attempt to detect chart spec in signal payload
    const chartSpec = useMemo(() => {
        try {
            const parsed = JSON.parse(message);
            if (parsed.chart_type && parsed.data && Array.isArray(parsed.data)) {
                return parsed as MycelisChartSpec;
            }
        } catch { /* not a chart */ }
        return null;
    }, [message]);

    if (chartSpec) {
        return (
            <div className="mx-3 mb-3 rounded-lg border border-emerald-700/40 bg-emerald-950/20 p-3">
                <div className="flex items-center gap-2 mb-2">
                    <BarChart3 className="w-3.5 h-3.5 text-emerald-400" />
                    <span className="text-[10px] font-mono font-bold text-emerald-400 uppercase">
                        Chart: {chartSpec.title}
                    </span>
                    <span className="text-[10px] font-mono text-cortex-text-muted ml-auto">
                        {signal.source ?? 'agent'}
                    </span>
                </div>
                <div className="bg-cortex-bg rounded p-2 h-32 overflow-hidden">
                    <ChartRenderer spec={chartSpec} compact={true} />
                </div>
            </div>
        );
    }

    return (
        <div className="mx-3 mb-3 rounded-lg border border-emerald-700/40 bg-emerald-950/20 p-3">
            <div className="flex items-center gap-2 mb-2">
                <FileCheck className="w-3.5 h-3.5 text-emerald-400" />
                <span className="text-[10px] font-mono font-bold text-emerald-400 uppercase">
                    Artifact
                </span>
                <span className="text-[10px] font-mono text-cortex-text-muted ml-auto">
                    {signal.source ?? 'agent'}
                </span>
            </div>
            <p className="text-[11px] font-mono text-emerald-200/70 leading-relaxed whitespace-pre-wrap">
                {message}
            </p>
        </div>
    );
}

export default function SquadRoom({ teamId }: SquadRoomProps) {
    const exitSquadRoom = useCortexStore((s) => s.exitSquadRoom);
    const streamLogs = useCortexStore((s) => s.streamLogs);
    const { teamName, agentIds, agentCount, blueprintTeam } = useTeamInfo(teamId);
    const debateRef = useRef<HTMLDivElement>(null);

    // Filter stream logs to only this team's agents
    const teamLogs = useMemo(
        () => streamLogs.filter((s) => s.source && agentIds.includes(s.source)),
        [streamLogs, agentIds],
    );

    // Split: debate = thoughts/errors/governance, proofs = artifacts/outputs
    const debateSignals = useMemo(
        () => teamLogs.filter((s) => s.type !== 'artifact' && s.type !== 'output'),
        [teamLogs],
    );
    const proofSignals = useMemo(
        () => teamLogs.filter((s) => s.type === 'artifact' || s.type === 'output'),
        [teamLogs],
    );

    // Auto-scroll debate feed
    useEffect(() => {
        if (debateRef.current) {
            debateRef.current.scrollTop = 0;
        }
    }, [debateSignals.length]);

    return (
        <div className="h-full flex flex-col bg-cortex-surface">
            {/* Header */}
            <div className="h-14 flex items-center px-4 border-b border-cortex-border bg-cortex-bg/50 flex-shrink-0">
                <button
                    onClick={exitSquadRoom}
                    className="flex items-center gap-1.5 text-cortex-text-muted hover:text-cortex-text-main transition-colors mr-4"
                >
                    <ArrowLeft className="w-4 h-4" />
                    <span className="text-[10px] font-mono uppercase tracking-wide">Back</span>
                </button>

                <div className="h-5 w-px bg-cortex-border mr-4" />

                <Users className="w-4 h-4 text-cortex-info mr-2" />
                <span className="text-sm font-bold text-cortex-text-main tracking-wide">
                    {teamName}
                </span>

                <span className="ml-3 text-[10px] font-mono text-cortex-text-muted">
                    {agentCount} agent{agentCount !== 1 ? 's' : ''}
                </span>

                {blueprintTeam?.role && (
                    <span className="ml-3 text-[10px] font-mono text-cortex-info/70 px-1.5 py-0.5 rounded bg-cortex-info/10 border border-cortex-info/20">
                        {blueprintTeam.role}
                    </span>
                )}

                <span className="ml-auto text-[9px] font-mono text-cortex-text-muted/60 uppercase">
                    Squad Room
                </span>
            </div>

            {/* Agent roster bar */}
            <div className="px-4 py-2 border-b border-cortex-border/50 flex items-center gap-3 flex-shrink-0">
                {agentIds.map((id) => (
                    <div key={id} className="flex items-center gap-1.5">
                        <span className="w-1.5 h-1.5 rounded-full bg-cortex-success shadow-[0_0_4px_rgba(40,199,111,0.5)]" />
                        <span className="text-[10px] font-mono text-cortex-text-muted">{id}</span>
                    </div>
                ))}
                {agentIds.length === 0 && (
                    <span className="text-[10px] font-mono text-cortex-text-muted/60 italic">
                        No agents resolved for this team
                    </span>
                )}
            </div>

            {/* Split view: Debate Feed + Proof Viewer */}
            <div className="flex-1 grid grid-cols-[1fr_340px] min-h-0">
                {/* Left: Debate Feed */}
                <div className="flex flex-col border-r border-cortex-border min-h-0">
                    <div className="px-4 py-2 border-b border-cortex-border/50 flex items-center gap-2 flex-shrink-0">
                        <MessageSquare className="w-3.5 h-3.5 text-cortex-info" />
                        <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wide">
                            Internal Debate
                        </span>
                        <span className="text-[9px] font-mono text-cortex-text-muted/60 ml-auto">
                            {debateSignals.length} signals
                        </span>
                    </div>
                    <div ref={debateRef} className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border">
                        {debateSignals.length === 0 ? (
                            <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                                <MessageSquare className="w-8 h-8 mb-2 opacity-20" />
                                <p className="text-[10px] font-mono">Awaiting internal chatter...</p>
                                <p className="text-[9px] font-mono opacity-60 mt-1">
                                    Agent thoughts will appear here
                                </p>
                            </div>
                        ) : (
                            debateSignals.map((signal, i) => (
                                <DebateEntry key={`${signal.timestamp}-${i}`} signal={signal} />
                            ))
                        )}
                    </div>
                </div>

                {/* Right: Proof Viewer */}
                <div className="flex flex-col min-h-0">
                    <div className="px-4 py-2 border-b border-cortex-border/50 flex items-center gap-2 flex-shrink-0">
                        <FileCheck className="w-3.5 h-3.5 text-cortex-success" />
                        <span className="text-[10px] font-mono font-bold text-cortex-text-muted uppercase tracking-wide">
                            Proof of Work
                        </span>
                        <span className="text-[9px] font-mono text-cortex-text-muted/60 ml-auto">
                            {proofSignals.length} artifacts
                        </span>
                    </div>
                    <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-cortex-border pt-3">
                        {proofSignals.length === 0 ? (
                            <div className="flex flex-col items-center justify-center h-full text-cortex-text-muted">
                                <FileCheck className="w-8 h-8 mb-2 opacity-20" />
                                <p className="text-[10px] font-mono">No artifacts yet</p>
                                <p className="text-[9px] font-mono opacity-60 mt-1">
                                    Verified outputs appear here
                                </p>
                            </div>
                        ) : (
                            proofSignals.map((signal, i) => (
                                <ProofCard key={`proof-${signal.timestamp}-${i}`} signal={signal} />
                            ))
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

"use client";

import React from 'react';
import { Users, Radio, Cpu, Eye, BookOpen, Activity } from 'lucide-react';
import type { TeamDetailEntry } from '@/store/useCortexStore';

const roleConfig: Record<string, { color: string; icon: typeof Cpu }> = {
    action: { color: 'border-l-cortex-primary', icon: Cpu },
    expression: { color: 'border-l-cortex-success', icon: Activity },
    cognitive: { color: 'border-l-purple-500', icon: Cpu },
    sensory: { color: 'border-l-cyan-500', icon: Eye },
    actuation: { color: 'border-l-cortex-success', icon: Activity },
    ledger: { color: 'border-l-gray-500', icon: BookOpen },
};

function getAggregateStatus(agents: TeamDetailEntry['agents']): 'online' | 'busy' | 'offline' | 'error' {
    if (agents.length === 0) return 'offline';
    const hasError = agents.some((a) => a.status === 3);
    if (hasError) return 'error';
    const hasBusy = agents.some((a) => a.status === 2);
    if (hasBusy) return 'busy';
    const hasOnline = agents.some((a) => a.status >= 1);
    if (hasOnline) return 'online';
    return 'offline';
}

const statusDot: Record<string, string> = {
    online: 'bg-cortex-success',
    busy: 'bg-cortex-primary animate-pulse',
    offline: 'bg-cortex-text-muted/40',
    error: 'bg-red-500',
};

interface TeamCardProps {
    team: TeamDetailEntry;
    onClick: () => void;
    isSelected: boolean;
}

export default function TeamCard({ team, onClick, isSelected }: TeamCardProps) {
    const cfg = roleConfig[team.role] ?? roleConfig.action;
    const Icon = cfg.icon;
    const status = getAggregateStatus(team.agents);
    const onlineCount = team.agents.filter((a) => a.status >= 1).length;

    return (
        <button
            onClick={onClick}
            className={`w-full text-left rounded-xl border-l-4 ${cfg.color} border border-cortex-border bg-cortex-surface p-4 transition-all duration-200 hover:bg-cortex-bg hover:border-cortex-text-muted/50 ${isSelected ? 'ring-1 ring-cortex-primary' : ''}`}
        >
            {/* Header row */}
            <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                    <span className={`w-2 h-2 rounded-full ${statusDot[status]}`} />
                    <Icon className="w-4 h-4 text-cortex-text-muted" />
                    <span className="text-sm font-mono font-bold text-cortex-text-main truncate">
                        {team.name}
                    </span>
                </div>
                <span
                    className={`text-[9px] font-mono uppercase px-1.5 py-0.5 rounded ${
                        team.type === 'standing'
                            ? 'bg-cortex-primary/10 text-cortex-primary border border-cortex-primary/30'
                            : 'bg-cortex-success/10 text-cortex-success border border-cortex-success/30'
                    }`}
                >
                    {team.type}
                </span>
            </div>

            {/* Agent count */}
            <div className="flex items-center gap-1.5 mb-2">
                <Users className="w-3 h-3 text-cortex-text-muted" />
                <span className="text-[10px] font-mono text-cortex-text-muted">
                    {onlineCount}/{team.agents.length} agent{team.agents.length !== 1 ? 's' : ''} online
                </span>
            </div>

            {/* Delivery topics (max 2) */}
            {team.deliveries.length > 0 && (
                <div className="space-y-0.5 mb-2">
                    {team.deliveries.slice(0, 2).map((d) => (
                        <div key={d} className="flex items-center gap-1">
                            <Radio className="w-2.5 h-2.5 text-cortex-text-muted/60" />
                            <span className="text-[9px] font-mono text-cortex-text-muted/80 truncate">
                                {d}
                            </span>
                        </div>
                    ))}
                    {team.deliveries.length > 2 && (
                        <span className="text-[9px] font-mono text-cortex-text-muted/50">
                            +{team.deliveries.length - 2} more
                        </span>
                    )}
                </div>
            )}

            {/* Mission intent (only for mission teams) */}
            {team.type === 'mission' && team.mission_intent && (
                <div className="pt-2 border-t border-cortex-border/50">
                    <span className="text-[9px] font-mono text-cortex-success/80 line-clamp-2">
                        {team.mission_intent}
                    </span>
                </div>
            )}
        </button>
    );
}

"use client";

import React from 'react';
import { Users, Radio, Cpu, Eye, BookOpen, Activity, MessageSquare, Route, Network, ScrollText } from 'lucide-react';
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

function statusBadge(status: 'online' | 'busy' | 'offline' | 'error'): string {
    if (status === 'online') return 'bg-cortex-success/10 text-cortex-success border-cortex-success/30';
    if (status === 'busy') return 'bg-cortex-primary/10 text-cortex-primary border-cortex-primary/30';
    if (status === 'error') return 'bg-cortex-danger/10 text-cortex-danger border-cortex-danger/30';
    return 'bg-cortex-surface text-cortex-text-muted border-cortex-border';
}

function statusLabel(status: 'online' | 'busy' | 'offline' | 'error'): string {
    if (status === 'online') return 'Healthy';
    if (status === 'busy') return 'Degraded';
    if (status === 'error') return 'Offline';
    return 'Offline';
}

function relativeTime(ts?: string): string {
    if (!ts) return 'never';
    const diff = Date.now() - new Date(ts).getTime();
    if (diff < 5000) return 'just now';
    if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`;
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    return `${Math.floor(diff / 3600000)}h ago`;
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
    const lastHeartbeat = team.agents
        .map((a) => a.last_heartbeat)
        .filter(Boolean)
        .sort((a, b) => new Date(b).getTime() - new Date(a).getTime())[0];
    const summary = team.mission_intent || `${team.role} team handling ${team.type} workflows`;

    const onKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onClick();
        }
    };

    return (
        <div
            role="button"
            tabIndex={0}
            onClick={onClick}
            onKeyDown={onKeyDown}
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

            <div className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded border text-[9px] font-mono uppercase ${statusBadge(status)}`}>
                <span className={`w-1.5 h-1.5 rounded-full ${statusDot[status]}`} />
                {statusLabel(status)}
            </div>

            {/* Agent count */}
            <div className="flex items-center gap-1.5 mt-2 mb-1">
                <Users className="w-3 h-3 text-cortex-text-muted" />
                <span className="text-[10px] font-mono text-cortex-text-muted">
                    {onlineCount}/{team.agents.length} agent{team.agents.length !== 1 ? 's' : ''} online
                </span>
            </div>
            <div className="text-[10px] font-mono text-cortex-text-muted mb-2">
                Last heartbeat: {relativeTime(lastHeartbeat)}
            </div>

            <p className="text-[10px] text-cortex-text-main/80 line-clamp-1 mb-2">{summary}</p>

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

            <div className="pt-2 mt-2 border-t border-cortex-border/40 grid grid-cols-2 gap-1">
                <a href="/dashboard" onClick={(e) => e.stopPropagation()} className="text-[9px] font-mono px-1.5 py-1 rounded border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg inline-flex items-center gap-1" data-testid={`team-${team.id}-open-chat`}>
                    <MessageSquare className="w-3 h-3" />
                    Open chat
                </a>
                <a href="/runs" onClick={(e) => e.stopPropagation()} className="text-[9px] font-mono px-1.5 py-1 rounded border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg inline-flex items-center gap-1" data-testid={`team-${team.id}-view-runs`}>
                    <Route className="w-3 h-3" />
                    View runs
                </a>
                <a href="/automations?tab=wiring" onClick={(e) => e.stopPropagation()} className="text-[9px] font-mono px-1.5 py-1 rounded border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg inline-flex items-center gap-1" data-testid={`team-${team.id}-view-wiring`}>
                    <Network className="w-3 h-3" />
                    View wiring
                </a>
                <a href="/system?tab=debug" onClick={(e) => e.stopPropagation()} className="text-[9px] font-mono px-1.5 py-1 rounded border border-cortex-border text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg inline-flex items-center gap-1" data-testid={`team-${team.id}-view-logs`}>
                    <ScrollText className="w-3 h-3" />
                    View logs
                </a>
            </div>
        </div>
    );
}

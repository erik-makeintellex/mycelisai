"use client";

import React, { useEffect } from 'react';
import { Users, User, Radio, Brain, Zap, BookOpen, AlertCircle } from 'lucide-react';
import { useCortexStore, type TeamDetail, type TeamAgent } from '@/store/useCortexStore';

const STATUS_DOT: Record<number, string> = {
    0: 'bg-cortex-text-muted/40',           // offline → gray
    1: 'bg-cortex-success',                  // idle → green
    2: 'bg-cortex-info animate-pulse',       // busy → blue pulse
    3: 'bg-cortex-danger',                   // error → red
};

const STATUS_LABEL: Record<number, string> = {
    0: 'offline',
    1: 'idle',
    2: 'busy',
    3: 'error',
};

const ROLE_ICON: Record<string, React.ComponentType<{ className?: string }>> = {
    cognitive: Brain,
    sensory: Radio,
    actuation: Zap,
    ledger: BookOpen,
    observer: Users,
};

function relativeTime(timestamp: string): string {
    if (!timestamp) return '—';
    const diff = Date.now() - new Date(timestamp).getTime();
    if (diff < 0) return 'just now';
    const seconds = Math.floor(diff / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    return `${hours}h ago`;
}

function AgentRow({ agent }: { agent: TeamAgent }) {
    return (
        <div className="flex items-center gap-2 px-2 py-1 rounded hover:bg-cortex-bg/50 transition-colors">
            <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${STATUS_DOT[agent.status] ?? STATUS_DOT[0]}`} />
            <span className="text-[10px] font-mono text-cortex-text-main truncate flex-1">
                {agent.name || agent.id.slice(0, 12)}
            </span>
            <span className="text-[9px] font-mono text-cortex-text-muted">
                {relativeTime(agent.last_heartbeat)}
            </span>
        </div>
    );
}

function TeamCard({ team }: { team: TeamDetail }) {
    const RoleIcon = ROLE_ICON[team.role] ?? Users;
    const onlineCount = team.agents.filter((a) => a.status >= 1 && a.status <= 2).length;

    return (
        <div className="bg-cortex-surface border border-cortex-border rounded-lg overflow-hidden" data-testid="team-card">
            {/* Team header */}
            <div className="px-3 py-2 border-b border-cortex-border/50 flex items-center gap-2">
                <RoleIcon className="w-3.5 h-3.5 text-cortex-text-muted" />
                <span className="text-xs font-mono font-bold text-cortex-text-main truncate flex-1">
                    {team.name}
                </span>
                <span className="text-[9px] font-mono text-cortex-text-muted">
                    {onlineCount}/{team.agents.length}
                </span>
                <span className={`w-1.5 h-1.5 rounded-full ${onlineCount > 0 ? 'bg-cortex-success' : 'bg-cortex-text-muted/40'}`} />
            </div>

            {/* Agent list */}
            {team.agents.length > 0 ? (
                <div className="py-1">
                    {team.agents.map((agent) => (
                        <AgentRow key={agent.id} agent={agent} />
                    ))}
                </div>
            ) : (
                <div className="px-3 py-2">
                    <span className="text-[9px] font-mono text-cortex-text-muted">No agents registered</span>
                </div>
            )}
        </div>
    );
}

export default function TeamExplorer() {
    const teamRoster = useCortexStore((s) => s.teamRoster);
    const isFetchingTeamRoster = useCortexStore((s) => s.isFetchingTeamRoster);
    const fetchTeamDetails = useCortexStore((s) => s.fetchTeamDetails);

    useEffect(() => {
        fetchTeamDetails();
        const interval = setInterval(fetchTeamDetails, 15000);
        return () => clearInterval(interval);
    }, [fetchTeamDetails]);

    // Loading skeleton
    if (isFetchingTeamRoster && teamRoster.length === 0) {
        return (
            <div className="h-full overflow-y-auto p-3 space-y-3" data-testid="team-explorer">
                {[1, 2, 3].map((i) => (
                    <div key={i} className="h-24 bg-cortex-surface border border-cortex-border rounded-lg animate-pulse" />
                ))}
            </div>
        );
    }

    // Empty state
    if (teamRoster.length === 0) {
        return (
            <div className="h-full flex flex-col items-center justify-center text-cortex-text-muted" data-testid="team-explorer">
                <Users className="w-8 h-8 mb-2 opacity-20" />
                <p className="text-[10px] font-mono text-center">No active teams</p>
                <p className="text-[9px] font-mono mt-1 opacity-60">
                    Create a mission from the Wiring page
                </p>
            </div>
        );
    }

    const totalAgents = teamRoster.reduce((sum, t) => sum + t.agents.length, 0);
    const activeAgents = teamRoster.reduce(
        (sum, t) => sum + t.agents.filter((a) => a.status >= 1 && a.status <= 2).length,
        0,
    );

    return (
        <div className="h-full flex flex-col" data-testid="team-explorer">
            {/* Summary bar */}
            <div className="px-3 py-1.5 border-b border-cortex-border flex items-center gap-3 flex-shrink-0">
                <span className="text-[9px] font-mono text-cortex-text-muted">
                    {teamRoster.length} team{teamRoster.length !== 1 ? 's' : ''}
                </span>
                <span className="text-[9px] font-mono text-cortex-text-muted">
                    {activeAgents}/{totalAgents} agents online
                </span>
            </div>

            {/* Team cards */}
            <div className="flex-1 overflow-y-auto p-3 space-y-3 scrollbar-thin scrollbar-thumb-cortex-border">
                {teamRoster.map((team) => (
                    <TeamCard key={team.id} team={team} />
                ))}
            </div>
        </div>
    );
}

"use client";

import React, { useEffect, useCallback, useMemo } from 'react';
import { Users, RefreshCw } from 'lucide-react';
import { useCortexStore, type TeamsFilter, type TeamDetailEntry } from '@/store/useCortexStore';
import TeamCard from './TeamCard';
import TeamDetailDrawer from './TeamDetailDrawer';

const FILTERS: { value: TeamsFilter; label: string }[] = [
    { value: 'all', label: 'All Teams' },
    { value: 'standing', label: 'Standing' },
    { value: 'mission', label: 'Mission' },
];

export default function TeamsPage() {
    const teamsDetail = useCortexStore((s) => s.teamsDetail);
    const isFetching = useCortexStore((s) => s.isFetchingTeamsDetail);
    const selectedTeamId = useCortexStore((s) => s.selectedTeamId);
    const isDrawerOpen = useCortexStore((s) => s.isTeamDrawerOpen);
    const teamsFilter = useCortexStore((s) => s.teamsFilter);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const selectTeam = useCortexStore((s) => s.selectTeam);
    const setTeamsFilter = useCortexStore((s) => s.setTeamsFilter);

    // Fetch on mount + auto-refresh every 15s
    useEffect(() => {
        fetchTeamsDetail();
        const interval = setInterval(fetchTeamsDetail, 15000);
        return () => clearInterval(interval);
    }, [fetchTeamsDetail]);

    const handleFilterChange = useCallback(
        (e: React.ChangeEvent<HTMLSelectElement>) => {
            setTeamsFilter(e.target.value as TeamsFilter);
        },
        [setTeamsFilter],
    );

    // Filter + sort: standing first, then by name
    const filteredTeams = useMemo(() => {
        let teams = teamsDetail;
        if (teamsFilter !== 'all') {
            teams = teams.filter((t) => t.type === teamsFilter);
        }
        return [...teams].sort((a, b) => {
            if (a.type !== b.type) return a.type === 'standing' ? -1 : 1;
            return a.name.localeCompare(b.name);
        });
    }, [teamsDetail, teamsFilter]);

    const selectedTeam: TeamDetailEntry | null = useMemo(() => {
        if (!selectedTeamId) return null;
        return teamsDetail.find((t) => t.id === selectedTeamId) ?? null;
    }, [selectedTeamId, teamsDetail]);

    const totalAgents = teamsDetail.reduce((sum, t) => sum + t.agents.length, 0);
    const onlineAgents = teamsDetail.reduce(
        (sum, t) => sum + t.agents.filter((a) => a.status >= 1).length,
        0,
    );

    return (
        <div className="h-full flex flex-col bg-cortex-bg relative">
            {/* Header */}
            <div className="px-6 py-4 border-b border-cortex-border bg-cortex-surface/50 flex items-center justify-between flex-shrink-0">
                <div className="flex items-center gap-3">
                    <Users className="w-5 h-5 text-cortex-primary" />
                    <h1 className="text-sm font-mono font-bold text-cortex-text-main uppercase tracking-wider">
                        Team Management
                    </h1>
                    <span className="text-[10px] font-mono text-cortex-text-muted">
                        {filteredTeams.length} team{filteredTeams.length !== 1 ? 's' : ''}
                    </span>
                    <span className="text-[10px] font-mono text-cortex-success">
                        {onlineAgents}/{totalAgents} agents online
                    </span>
                </div>

                <div className="flex items-center gap-3">
                    {/* Filter */}
                    <select
                        value={teamsFilter}
                        onChange={handleFilterChange}
                        className="bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main focus:outline-none focus:border-cortex-primary transition-colors appearance-none"
                    >
                        {FILTERS.map((f) => (
                            <option key={f.value} value={f.value}>
                                {f.label}
                            </option>
                        ))}
                    </select>

                    {/* Refresh */}
                    <button
                        onClick={fetchTeamsDetail}
                        disabled={isFetching}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors disabled:opacity-50"
                    >
                        <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
                    </button>
                </div>
            </div>

            {/* Grid */}
            <div className="flex-1 overflow-y-auto p-6">
                {filteredTeams.length > 0 ? (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                        {filteredTeams.map((team) => (
                            <TeamCard
                                key={team.id}
                                team={team}
                                onClick={() => selectTeam(team.id)}
                                isSelected={selectedTeamId === team.id}
                            />
                        ))}
                    </div>
                ) : (
                    <div className="flex flex-col items-center justify-center h-full text-center text-cortex-text-muted">
                        <Users className="w-12 h-12 mb-3 opacity-20" />
                        <p className="text-sm font-mono">No teams found</p>
                        <p className="text-xs font-mono mt-1 opacity-60">
                            {teamsFilter !== 'all'
                                ? `No ${teamsFilter} teams — try changing the filter`
                                : 'Start the Core server to see standing teams'}
                        </p>
                    </div>
                )}
            </div>

            {/* Detail Drawer */}
            {isDrawerOpen && selectedTeam && (
                <TeamDetailDrawer team={selectedTeam} onClose={() => selectTeam(null)} />
            )}
        </div>
    );
}

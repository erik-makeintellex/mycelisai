"use client";

import React, { useEffect, useCallback, useMemo, useState } from 'react';
import Link from 'next/link';
import { Bot, BookOpen, RefreshCw, Sparkles, Users } from 'lucide-react';
import {
    useCortexStore,
    type CatalogueAgent,
    type TeamsFilter,
    type TeamDetailEntry,
} from '@/store/useCortexStore';
import AgentEditorDrawer from '@/components/catalogue/AgentEditorDrawer';
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
    const catalogueAgents = useCortexStore((s) => s.catalogueAgents);
    const isFetchingCatalogue = useCortexStore((s) => s.isFetchingCatalogue);
    const selectedTeamId = useCortexStore((s) => s.selectedTeamId);
    const isDrawerOpen = useCortexStore((s) => s.isTeamDrawerOpen);
    const teamsFilter = useCortexStore((s) => s.teamsFilter);
    const fetchTeamsDetail = useCortexStore((s) => s.fetchTeamsDetail);
    const fetchCatalogue = useCortexStore((s) => s.fetchCatalogue);
    const createCatalogueAgent = useCortexStore((s) => s.createCatalogueAgent);
    const updateCatalogueAgent = useCortexStore((s) => s.updateCatalogueAgent);
    const selectTeam = useCortexStore((s) => s.selectTeam);
    const setTeamsFilter = useCortexStore((s) => s.setTeamsFilter);
    const [isTemplateDrawerOpen, setIsTemplateDrawerOpen] = useState(false);
    const [editingTemplate, setEditingTemplate] = useState<CatalogueAgent | null>(null);

    useEffect(() => {
        fetchTeamsDetail();
        fetchCatalogue();
        const interval = setInterval(fetchTeamsDetail, 15000);
        return () => clearInterval(interval);
    }, [fetchCatalogue, fetchTeamsDetail]);

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
    const sortedTemplates = useMemo(
        () => [...catalogueAgents].sort((a, b) => Date.parse(b.updated_at) - Date.parse(a.updated_at)),
        [catalogueAgents],
    );
    const highlightedTemplates = sortedTemplates.slice(0, 4);
    const templateCoverage = useMemo(() => {
        const coverage = new Map<string, CatalogueAgent[]>();
        sortedTemplates.forEach((agent) => {
            const keys = agent.outputs.length > 0 ? agent.outputs : [agent.role];
            keys.forEach((key) => {
                const trimmed = key.trim();
                if (!trimmed) {
                    return;
                }
                const current = coverage.get(trimmed) ?? [];
                current.push(agent);
                coverage.set(trimmed, current);
            });
        });
        return Array.from(coverage.entries()).slice(0, 6);
    }, [sortedTemplates]);

    const openTemplateDrawer = useCallback((agent: CatalogueAgent | null) => {
        setEditingTemplate(agent);
        setIsTemplateDrawerOpen(true);
    }, []);

    const closeTemplateDrawer = useCallback(() => {
        setEditingTemplate(null);
        setIsTemplateDrawerOpen(false);
    }, []);

    const handleTemplateSave = useCallback(
        (data: Partial<CatalogueAgent>) => {
            if (editingTemplate) {
                void updateCatalogueAgent(editingTemplate.id, data);
            } else {
                void createCatalogueAgent(data);
            }
            closeTemplateDrawer();
        },
        [closeTemplateDrawer, createCatalogueAgent, editingTemplate, updateCatalogueAgent],
    );

    return (
        <div className="h-full flex flex-col bg-cortex-bg relative">
            <div className="px-6 py-4 border-b border-cortex-border bg-cortex-surface/50 flex items-center justify-between flex-shrink-0">
                <div className="flex items-start gap-3">
                    <Users className="w-5 h-5 text-cortex-primary mt-0.5" />
                    <div>
                        <div className="flex flex-wrap items-center gap-3">
                            <h1 className="text-sm font-mono font-bold text-cortex-text-main uppercase tracking-wider">
                                Team Lead Workspaces
                            </h1>
                            <span className="text-[10px] font-mono text-cortex-text-muted">
                                {filteredTeams.length} team{filteredTeams.length !== 1 ? 's' : ''}
                            </span>
                            <span className="text-[10px] font-mono text-cortex-success">
                                {onlineAgents}/{totalAgents} agents online
                            </span>
                        </div>
                        <p className="mt-1 text-xs text-cortex-text-muted">
                            Review live teams here, open focused lead workspaces, and define which reusable team-member templates Soma should apply when specializing a new lane.
                        </p>
                    </div>
                </div>

                <div className="flex items-center gap-3">
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

                    <button
                        onClick={fetchTeamsDetail}
                        disabled={isFetching}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors disabled:opacity-50"
                    >
                        <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-6 space-y-4">
                <div className="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <div className="flex items-start gap-3">
                            <div className="rounded-full border border-cortex-primary/20 bg-cortex-primary/10 p-2 text-cortex-primary">
                                <Sparkles className="h-4 w-4" />
                            </div>
                            <div className="flex-1">
                                <div className="flex flex-wrap items-center gap-2">
                                    <p className="text-sm font-semibold text-cortex-text-main">Specialize new teams through Soma</p>
                                    <span className="rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-2.5 py-1 text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">
                                        Soma-first
                                    </span>
                                </div>
                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                    Start new teams through a guided creation flow, then hand the finished ask to Soma. Return here after launch to open the team lead, review the current roster, and maintain the member templates Soma should reuse on future team builds.
                                </p>
                                <div className="mt-4 flex flex-wrap gap-2">
                                    <Link
                                        href="/teams/create"
                                        className="inline-flex items-center justify-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10"
                                    >
                                        Open guided team creation
                                    </Link>
                                    <Link
                                        href="/dashboard"
                                        className="inline-flex items-center justify-center rounded-2xl border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-main hover:bg-cortex-border"
                                    >
                                        Open Soma workspace
                                    </Link>
                                    <Link
                                        href="/groups"
                                        className="inline-flex items-center justify-center rounded-2xl border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-main hover:bg-cortex-border"
                                    >
                                        Open groups workspace
                                    </Link>
                                </div>
                                <div className="mt-4 grid gap-3 md:grid-cols-3">
                                    <QuickFact
                                        label="Guide the request"
                                        value="Use the guided team-creation flow to define the lane, lead posture, and expected outputs before handoff."
                                    />
                                    <QuickFact
                                        label="Work through the lead"
                                        value="Each team opens around one focused lead counterpart instead of another generic global chat."
                                    />
                                    <QuickFact
                                        label="Templates guide member choice"
                                        value="Soma can reuse the configured team-member templates below when a lane needs specific agent behavior."
                                    />
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                        <div className="flex items-start justify-between gap-3">
                            <div className="flex items-start gap-3">
                                <div className="rounded-full border border-cortex-success/20 bg-cortex-success/10 p-2 text-cortex-success">
                                    <BookOpen className="h-4 w-4" />
                                </div>
                                <div>
                                    <div className="flex flex-wrap items-center gap-2">
                                        <p className="text-sm font-semibold text-cortex-text-main">Soma team-member templates</p>
                                        {isFetchingCatalogue && (
                                            <span className="text-[10px] font-mono text-cortex-text-muted animate-pulse">
                                                loading...
                                            </span>
                                        )}
                                    </div>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                        By default, team members inherit the shared organization model policy. Give Soma reusable templates when certain work should prefer a specific specialist role, model, toolset, or output contract.
                                    </p>
                                </div>
                            </div>
                            <button
                                type="button"
                                onClick={() => openTemplateDrawer(null)}
                                className="inline-flex items-center justify-center rounded-2xl border border-cortex-success/30 px-3 py-2 text-sm font-semibold text-cortex-success hover:bg-cortex-success/10"
                            >
                                New template
                            </button>
                        </div>

                        <div className="mt-4 space-y-3">
                            {highlightedTemplates.length > 0 ? (
                                highlightedTemplates.map((agent) => (
                                    <button
                                        key={agent.id}
                                        type="button"
                                        onClick={() => openTemplateDrawer(agent)}
                                        className="w-full rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-left transition-colors hover:border-cortex-primary/30 hover:bg-cortex-surface"
                                    >
                                        <div className="flex items-start justify-between gap-3">
                                            <div>
                                                <div className="flex flex-wrap items-center gap-2">
                                                    <p className="text-sm font-semibold text-cortex-text-main">{agent.name}</p>
                                                    <span className="rounded-full border border-cortex-border px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
                                                        {agent.role}
                                                    </span>
                                                </div>
                                                <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                                    {agent.system_prompt?.trim() || "Reusable specialist template for future team-member creation."}
                                                </p>
                                            </div>
                                            <Bot className="mt-0.5 h-4 w-4 text-cortex-text-muted" />
                                        </div>
                                        <div className="mt-3 flex flex-wrap gap-2 text-[10px] font-mono">
                                            <span className="rounded-full border border-cortex-primary/20 bg-cortex-primary/10 px-2 py-0.5 text-cortex-primary">
                                                {agent.model?.trim() || "inherits org model"}
                                            </span>
                                            <span className="rounded-full border border-cortex-border px-2 py-0.5 text-cortex-text-muted">
                                                {agent.tools.length} tool{agent.tools.length !== 1 ? "s" : ""}
                                            </span>
                                            <span className="rounded-full border border-cortex-border px-2 py-0.5 text-cortex-text-muted">
                                                {agent.outputs.length > 0
                                                    ? `outputs: ${agent.outputs.slice(0, 2).join(", ")}${agent.outputs.length > 2 ? "..." : ""}`
                                                    : "general lane support"}
                                            </span>
                                        </div>
                                    </button>
                                ))
                            ) : (
                                <div className="rounded-2xl border border-dashed border-cortex-border bg-cortex-bg px-4 py-4">
                                    <p className="text-sm font-semibold text-cortex-text-main">No team-member templates yet</p>
                                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                        Create reusable agent templates here so Soma knows which kinds of specialists to reach for when a new team needs writers, coders, researchers, or reviewers.
                                    </p>
                                </div>
                            )}
                        </div>

                        <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-4">
                            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">What Soma can match by work type</p>
                            <div className="mt-3 space-y-2">
                                {templateCoverage.length > 0 ? (
                                    templateCoverage.map(([kind, agents]) => (
                                        <div key={kind} className="flex flex-col gap-1 rounded-2xl border border-cortex-border/80 bg-cortex-surface px-3 py-2 sm:flex-row sm:items-center sm:justify-between">
                                            <span className="text-sm font-medium text-cortex-text-main">{kind}</span>
                                            <span className="text-xs text-cortex-text-muted">
                                                {agents.slice(0, 2).map((agent) => agent.name).join(", ")}
                                                {agents.length > 2 ? ` +${agents.length - 2} more` : ""}
                                            </span>
                                        </div>
                                    ))
                                ) : (
                                    <p className="text-sm leading-6 text-cortex-text-muted">
                                        Once you add templates, this page will show the work types Soma can map to specific team-member defaults.
                                    </p>
                                )}
                            </div>
                            <div className="mt-4 flex flex-wrap gap-2">
                                <Link
                                    href="/resources?tab=roles"
                                    className="inline-flex items-center justify-center rounded-2xl border border-cortex-border px-4 py-2 text-sm font-semibold text-cortex-text-main hover:bg-cortex-border"
                                >
                                    Open full role library
                                </Link>
                                <Link
                                    href="/dashboard"
                                    className="inline-flex items-center justify-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10"
                                >
                                    Ask Soma to use these defaults
                                </Link>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-4">
                    <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                        <div>
                            <p className="text-sm font-semibold text-cortex-text-main">Groups have their own workspace now.</p>
                            <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                                Keep this page focused on team leads. Open Groups when you want temporary or standing collaboration lanes, output review, and broadcast coordination.
                            </p>
                        </div>
                        <Link href="/groups" className="inline-flex items-center justify-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10">
                            Open groups workspace
                        </Link>
                    </div>
                </div>
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
                        <div className="mt-3 flex items-center gap-2">
                            <button onClick={fetchTeamsDetail} className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/10">
                                Refresh
                            </button>
                            <a href="/teams/create" className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/10">
                                Guided team creation
                            </a>
                            <a href="/dashboard" className="px-2.5 py-1.5 rounded border border-cortex-border text-cortex-text-main text-[10px] font-mono hover:bg-cortex-border">
                                Open Soma workspace
                            </a>
                        </div>
                    </div>
                )}
            </div>

            {/* Detail Drawer */}
            {isDrawerOpen && selectedTeam && (
                <TeamDetailDrawer team={selectedTeam} onClose={() => selectTeam(null)} />
            )}
            {isTemplateDrawerOpen && (
                <AgentEditorDrawer
                    agent={editingTemplate}
                    onClose={closeTemplateDrawer}
                    onSave={handleTemplateSave}
                />
            )}
        </div>
    );
}

function QuickFact({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3">
            <p className="text-[10px] font-mono uppercase tracking-[0.18em] text-cortex-text-muted">{label}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-main">{value}</p>
        </div>
    );
}

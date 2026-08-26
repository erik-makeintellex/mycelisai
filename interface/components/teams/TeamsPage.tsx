"use client";

import React, { useEffect, useCallback, useMemo, useState, useSyncExternalStore } from "react";
import { RefreshCw, Users } from "lucide-react";
import {
  useCortexStore,
  type CatalogueAgent,
  type TeamsFilter,
  type TeamDetailEntry,
} from "@/store/useCortexStore";
import AgentEditorDrawer from "@/components/catalogue/AgentEditorDrawer";
import { customizeWorkerProfilePrompt, newWorkerProfilePrompt, requestSomaPromptHandoff } from "@/components/soma/somaPromptHandoff";
import { recoveryReviewQueueItems } from "@/components/recovery/recoveryQueue";
import { ActiveWorkLane } from "./ActiveWorkLane";
import TeamCard from "./TeamCard";
import TeamDetailDrawer from "./TeamDetailDrawer";
import {
  TeamsOutputCollaboration,
  TeamsReviewContextNote,
  TeamsSetupPanels,
  TeamsWorkReviewIntro,
} from "./TeamsPageSections";
import { useDurableTeamWork } from "@/components/soma/useDurableTeamWork";
import { mergeTeamWorkItems, useTeamWorkActionHandler } from "./useTeamWorkActionHandler";
import { prioritizeRequestedWorkItem } from "./teamsPageWorkReview";
import { useBrowserSearch } from "@/lib/browserLocation";

const FILTERS: { value: TeamsFilter; label: string }[] = [
  { value: "all", label: "All Teams" }, { value: "standing", label: "Standing" }, { value: "mission", label: "Mission" },
];
const subscribeToHydration = () => () => {};

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
  const selectTeam = useCortexStore((s) => s.selectTeam);
  const setTeamsFilter = useCortexStore((s) => s.setTeamsFilter);
  const durableWorkRefreshVersion = useCortexStore(
    (s) => s.durableWorkRefreshVersion,
  );
  const isInteractive = useSyncExternalStore(
    subscribeToHydration,
    () => true,
    () => false,
  );
  const [isTemplateDrawerOpen, setIsTemplateDrawerOpen] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<CatalogueAgent | null>(null);
  const activeWorkActions = useTeamWorkActionHandler(selectTeam);
  const teamsSearchParams = new URLSearchParams(useBrowserSearch());
  const isWorkReviewView = teamsSearchParams.get("view") === "work";
  const requestedWorkItemId = teamsSearchParams.get("work_item_id") || null;

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

  const filteredTeams = useMemo(() => {
    let teams = teamsDetail;
    if (teamsFilter !== "all") {
      teams = teams.filter((t) => t.type === teamsFilter);
    }
    return [...teams].sort((a, b) => {
      if (a.type !== b.type) return a.type === "standing" ? -1 : 1;
      return a.name.localeCompare(b.name);
    });
  }, [teamsDetail, teamsFilter]);

  const selectedTeam: TeamDetailEntry | null = useMemo(() => {
    if (!selectedTeamId) return null;
    return teamsDetail.find((t) => t.id === selectedTeamId) ?? null;
  }, [selectedTeamId, teamsDetail]);

  const totalAgents = teamsDetail.reduce((sum, t) => sum + t.agents.length, 0);
  const onlineAgents = teamsDetail.reduce((sum, t) => sum + t.agents.filter((a) => a.status >= 1).length, 0);
  const sortedTemplates = useMemo(
    () => catalogueAgents
      .filter((agent) => agent.source === "built_in" || agent.locked)
      .sort((a, b) => Date.parse(b.updated_at) - Date.parse(a.updated_at)),
    [catalogueAgents],
  );
  const highlightedTemplates = sortedTemplates.slice(0, 4);
  const activeTeamWork = useDurableTeamWork({
    teams: filteredTeams,
    refreshVersion: durableWorkRefreshVersion + activeWorkActions.activeWorkRefreshVersion,
    maxTeams: 12,
  });
  const activeWorkItems = mergeTeamWorkItems(
    activeTeamWork.items,
    activeWorkActions.submittedTeamWorkItems,
  );
  const recoveryReviewItems = recoveryReviewQueueItems(activeWorkItems);
  const laneItems = isWorkReviewView
    ? prioritizeRequestedWorkItem(recoveryReviewItems, requestedWorkItemId)
    : activeWorkItems;
  const requestedWorkItem = requestedWorkItemId
    ? recoveryReviewItems.find((item) => item.id === requestedWorkItemId) ?? null
    : null;
  const activeWorkLane = (
    <ActiveWorkLane
      title={isWorkReviewView ? "Recovery and review" : "Active work lane"}
      items={laneItems}
      emptyMessage={isWorkReviewView && activeWorkItems.length > 0 && recoveryReviewItems.length === 0
        ? "No recovery or review items need operator attention right now."
        : activeTeamWork.emptyMessage}
      statusLabel={requestedWorkItem
        ? `Opened "${requestedWorkItem.title}" from Outcome Vault.`
        : activeWorkActions.activeWorkActionNotice ?? activeTeamWork.statusLabel}
      degradedMessage={
        activeWorkActions.activeWorkActionError ?? activeTeamWork.degradedMessage
      }
      onAction={activeWorkActions.handleActiveWorkAction}
      onVerifyExternalOutcome={activeWorkActions.handleExternalOutcomeVerification}
      onTeamAsk={activeWorkActions.handleTeamAsk}
      purpose={isWorkReviewView ? "review" : "active"}
      moreItemsHref="/teams?view=work"
    />
  );

  const openTemplateDrawer = useCallback((agent: CatalogueAgent | null) => {
    setEditingTemplate(agent);
    setIsTemplateDrawerOpen(true);
  }, []);

  const closeTemplateDrawer = useCallback(() => {
    setEditingTemplate(null);
    setIsTemplateDrawerOpen(false);
  }, []);

  return (
    <div className="h-full flex flex-col bg-cortex-bg relative">
      <div className="relative flex flex-shrink-0 flex-col gap-2 border-b border-cortex-border bg-cortex-surface/50 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:py-4">
        <div className={`min-w-0 ${isWorkReviewView ? "pr-12 sm:pr-0" : ""}`}>
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <Users className="h-5 w-5 shrink-0 text-cortex-primary" />
              <h1 className="text-sm font-mono font-bold text-cortex-text-main uppercase tracking-wider">
                {isWorkReviewView ? "Recovery and Review" : "Team Lead Workspaces"}
              </h1>
              <span className="text-[10px] font-mono text-cortex-text-muted">
                {filteredTeams.length} team
                {filteredTeams.length !== 1 ? "s" : ""}
              </span>
              <span className="text-[10px] font-mono text-cortex-success" aria-label={`${onlineAgents} of ${totalAgents} agents online`}>
                {isWorkReviewView
                  ? `${onlineAgents}/${totalAgents} online`
                  : `${onlineAgents}/${totalAgents} agents online`}
              </span>
            </div>
            <p className="mt-1 text-xs text-cortex-text-muted">
              {isWorkReviewView
                ? "Review work that needs a decision, recovery, or output check."
                : "Review live teams here, open focused lead workspaces, and define which worker profiles Soma may apply when specializing a new lane."}
            </p>
        </div>

        <div className={isWorkReviewView
          ? "absolute right-3 top-2 flex items-center gap-2 sm:static sm:self-auto"
          : "flex items-center gap-2 self-end sm:self-auto"}>
          {!isWorkReviewView ? (
            <select
              value={teamsFilter}
              onChange={handleFilterChange}
              disabled={!isInteractive}
              aria-label="Filter teams"
              className="bg-cortex-bg border border-cortex-border rounded px-2.5 py-1.5 text-xs font-mono text-cortex-text-main focus:outline-none focus:border-cortex-primary transition-colors appearance-none"
            >
              {FILTERS.map((f) => (
                <option key={f.value} value={f.value}>
                  {f.label}
                </option>
              ))}
            </select>
          ) : null}

          <button
            onClick={fetchTeamsDetail}
            disabled={isFetching}
            aria-label="Refresh teams"
            className="inline-flex min-h-11 min-w-11 items-center justify-center rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors disabled:opacity-50 sm:min-h-8 sm:min-w-8"
          >
            <RefreshCw
              className={`w-4 h-4 ${isFetching ? "animate-spin" : ""}`}
            />
          </button>
        </div>
      </div>

      <div className="flex-1 space-y-3 overflow-y-auto p-3 sm:space-y-4 sm:p-6">
        {isWorkReviewView ? (
          <>
            <TeamsWorkReviewIntro />
            {activeWorkLane}
            <TeamsReviewContextNote />
          </>
        ) : (
          <>
            <TeamsSetupPanels
                highlightedTemplates={highlightedTemplates}
                isFetchingCatalogue={isFetchingCatalogue}
                onNewTemplate={() => requestSomaPromptHandoff(newWorkerProfilePrompt())}
                onEditTemplate={openTemplateDrawer}
            />
            <TeamsOutputCollaboration />
            {activeWorkLane}
          </>
        )}
        {!isWorkReviewView && filteredTeams.length > 0 ? (
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
        ) : !isWorkReviewView ? (
          <div className="flex flex-col items-center justify-center h-full text-center text-cortex-text-muted">
            <Users className="w-12 h-12 mb-3 opacity-20" />
            <p className="text-sm font-mono">No teams found</p>
            <p className="text-xs font-mono mt-1 opacity-60">
              {teamsFilter !== "all"
                ? `No ${teamsFilter} teams — try changing the filter`
                : "Start the Core server to see standing teams"}
            </p>
            <div className="mt-3 flex items-center gap-2">
              <button
                onClick={fetchTeamsDetail}
                className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/10"
              >
                Refresh
              </button>
              <a
                href="/teams/create"
                className="px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-[10px] font-mono hover:bg-cortex-primary/10"
              >
                Guided team creation
              </a>
              <a
                href="/dashboard"
                className="px-2.5 py-1.5 rounded border border-cortex-border text-cortex-text-main text-[10px] font-mono hover:bg-cortex-border"
              >
                Open Soma workspace
              </a>
            </div>
          </div>
        ) : null}
      </div>

      {isDrawerOpen && selectedTeam && (
        <TeamDetailDrawer team={selectedTeam} onClose={() => selectTeam(null)} />
      )}
      {isTemplateDrawerOpen && (
        <AgentEditorDrawer
          agent={editingTemplate}
          onClose={closeTemplateDrawer}
          onSave={() => undefined}
          onCustomize={(agent) => requestSomaPromptHandoff(customizeWorkerProfilePrompt(agent.name))}
        />
      )}
    </div>
  );
}

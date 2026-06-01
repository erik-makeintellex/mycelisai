"use client";

import type React from "react";
import { Activity, CheckSquare, Layers2, ListChecks, Wrench, X } from "lucide-react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import { ActiveWorkLane } from "@/components/teams/ActiveWorkLane";
import { useTeamWorkActionHandler } from "@/components/teams/useTeamWorkActionHandler";
import type { ChatMessage, TeamInteraction, TeamWorkItem } from "@/store/useCortexStore";
import { useCortexStore } from "@/store/useCortexStore";
import {
  mergeOutputWorkbenchItems,
  OutputWorkbench,
  outputWorkbenchItems,
  projectPackageOutputs,
  teamOutputProjectPackages,
  teamOutputWorkbenchItems,
} from "./OutputWorkbench";
import { SomaCausalSummary } from "./SomaCausalSummary";
import { SomaEvidencePanel, type SomaEvidenceItem } from "./SomaEvidencePanel";
import { SomaHeader } from "./SomaHeader";
import { DEFAULT_SOMA_SUGGESTIONS, type SomaSuggestion } from "./SomaSuggestionBar";
import { SomaTeamContextSwitcher } from "./SomaTeamContextSwitcher";
import { SomaWorkspaceFrame } from "./SomaWorkspaceFrame";
import { useDurableTeamWork } from "./useDurableTeamWork";

function lastSomaMessage(messages: ChatMessage[]) {
  return [...messages]
    .reverse()
    .find((message) => message.role !== "user" && message.role !== "system");
}

export function SomaOperatingSurface({
  organizationId,
  organizationName,
  activeMode,
  focusedTeamId,
  governancePosture,
  evidenceItems,
  activeWorkSlot,
  outputSlot,
  trustSlot,
  contextSlot,
  suggestions = DEFAULT_SOMA_SUGGESTIONS,
}: {
  organizationId?: string;
  organizationName?: string | null;
  activeMode?: string | null;
  focusedTeamId?: string | null;
  governancePosture?: string;
  evidenceItems?: SomaEvidenceItem[];
  activeWorkSlot?: React.ReactNode;
  outputSlot?: React.ReactNode;
  trustSlot?: React.ReactNode;
  contextSlot?: React.ReactNode;
  suggestions?: readonly SomaSuggestion[];
}) {
  const missionChat = useCortexStore((state) => state.missionChat);
  const teamsDetail = useCortexStore((state) => state.teamsDetail);
  const durableWorkRefreshVersion = useCortexStore((state) => state.durableWorkRefreshVersion);
  const selectTeam = useCortexStore((state) => state.selectTeam);
  const selectedTeamId = useCortexStore((state) => state.selectedTeamId);
  const activeWorkActions = useTeamWorkActionHandler(selectTeam);
  const evidence = evidenceItems ?? defaultEvidence;
  const latestSoma = lastSomaMessage(missionChat);
  const effectiveFocusedTeamId = focusedTeamId || selectedTeamId || null;
  const focusedTeam = effectiveFocusedTeamId
    ? teamsDetail.find((team) => team.id === effectiveFocusedTeamId) ?? null
    : null;
  const outputItems = outputWorkbenchItems(latestSoma?.execution_summary, latestSoma?.artifacts);
  const projectPackages = projectPackageOutputs(latestSoma?.execution_summary?.outputs);
  const teamWork = useDurableTeamWork({
    teams: teamsDetail,
    focusedTeamId: effectiveFocusedTeamId,
    refreshVersion: durableWorkRefreshVersion + activeWorkActions.activeWorkRefreshVersion,
  });
  const teamOutputItems = teamOutputWorkbenchItems(teamWork.outputRefs);
  const teamProjectPackages = teamOutputProjectPackages(teamWork.outputRefs);
  const mergedOutputItems = mergeOutputWorkbenchItems(outputItems, teamOutputItems);
  const mergedProjectPackages = [...projectPackages, ...teamProjectPackages];
  const somaHomeWorkItems = prioritizeSomaHomeWorkItems(teamWork.items).filter(
    (item) => item.state !== "archived",
  );
  const displayedMode = activeMode ?? (focusedTeam ? `Focused team: ${focusedTeam.name}` : null);

  const clearFocusedContext = () => {
    selectTeam(null);
    if (typeof window !== "undefined" && window.location.pathname === "/dashboard") {
      window.history.replaceState(null, "", "/dashboard");
    }
  };

  const focusTeamContext = (teamId: string) => {
    selectTeam(teamId);
    if (typeof window !== "undefined" && window.location.pathname === "/dashboard") {
      window.history.replaceState(null, "", `/dashboard?team_id=${encodeURIComponent(teamId)}`);
    }
  };

  const handleActiveWorkAction = async (item: TeamWorkItem, action: TeamInteraction) => {
    await activeWorkActions.handleActiveWorkAction(item, action);
    if (action.action === "inspect") {
      const teamId = item.teamIds[0] ?? item.id;
      if (typeof window !== "undefined" && window.location.pathname === "/dashboard") {
        window.history.replaceState(null, "", `/dashboard?team_id=${encodeURIComponent(teamId)}`);
      }
    }
  };

  return (
    <section
      className="overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface shadow-[0_18px_40px_rgba(148,163,184,0.16)]"
      data-testid="soma-operating-surface"
    >
      <SomaHeader
        organizationName={organizationName}
        activeMode={displayedMode}
        governancePosture={governancePosture}
      />
      <div className="p-4 lg:p-5">
        <SomaContextFocusBar
          teamName={focusedTeam?.name}
          teamId={effectiveFocusedTeamId}
          onClear={clearFocusedContext}
        />
        <SomaTeamContextSwitcher
          teams={teamsDetail}
          workItems={teamWork.items}
          focusedTeamId={effectiveFocusedTeamId}
          onRootSelect={clearFocusedContext}
          onTeamSelect={focusTeamContext}
        />
        <SomaWorkspaceFrame
          expression={(
            <div
              data-testid="central-soma-chat-frame"
              className="h-[64vh] min-h-[460px] overflow-hidden rounded-xl border border-cortex-border bg-cortex-bg lg:h-full lg:min-h-0"
            >
              <MissionControlChat
                simpleMode
                autoFocus
                organizationId={organizationId}
                focusedTeamId={effectiveFocusedTeamId}
                suggestions={suggestions}
              />
            </div>
          )}
          activeWork={activeWorkSlot ?? (
            <ActiveWorkLane
              title="Active work"
              items={somaHomeWorkItems}
              emptyMessage={activeMode && teamWork.items.length === 0
                ? `${activeMode} is the current workspace lane. ${teamWork.emptyMessage}`
                : teamWork.emptyMessage}
              statusLabel={teamWork.statusLabel}
              degradedMessage={activeWorkActions.activeWorkActionError ?? teamWork.degradedMessage}
              onAction={handleActiveWorkAction}
              onTeamAsk={activeWorkActions.handleTeamAsk}
              frame={false}
              maxVisibleItems={effectiveFocusedTeamId ? 6 : 3}
              totalItemCount={teamWork.items.length}
              moreItemsHref="/teams"
            />
          )}
          trust={trustSlot ?? <SomaCausalSummary messages={missionChat} />}
          output={outputSlot ?? (
            <OutputWorkbench
              outputs={mergedOutputItems}
              projectPackages={mergedProjectPackages}
              emptyMessage={teamWork.status === "loading"
                ? "Checking for retained Soma and team outputs."
                : "Soma has not returned a retained output package yet."}
            />
          )}
          context={contextSlot ?? <SomaEvidencePanel items={evidence} compact />}
        />
      </div>
    </section>
  );
}

function SomaContextFocusBar({
  teamName,
  teamId,
  onClear,
}: {
  teamName?: string | null;
  teamId?: string | null;
  onClear: () => void;
}) {
  const isFocused = Boolean(teamId);
  return (
    <div
      className="mb-3 flex flex-col gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-muted lg:flex-row lg:items-center lg:justify-between"
      data-testid="soma-context-focus-bar"
    >
      <div className="flex min-w-0 items-start gap-2">
        <span className="mt-0.5 text-cortex-primary">
          <Layers2 className="h-4 w-4" />
        </span>
        <div className="min-w-0">
          <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-cortex-primary">
            {isFocused ? "Focused team context" : "Root Soma context"}
          </p>
          <p className="mt-1 leading-5">
            {isFocused
              ? `${teamName || teamId} has its own chat, active work, and retained output focus. Soma can still reference other team contexts when you ask.`
              : "Select a running or executed team to switch this workbench into that team's chat, active work, and output lane."}
          </p>
        </div>
      </div>
      {isFocused ? (
        <button
          type="button"
          onClick={onClear}
          className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-cortex-border px-2.5 py-1.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/30"
        >
          <X className="h-3.5 w-3.5" />
          Soma root
        </button>
      ) : null}
    </div>
  );
}

const defaultEvidence: SomaEvidenceItem[] = [
  {
    title: "Approval queue",
    detail: "Review gated actions, risk decisions, and pending confirmations.",
    href: "/approvals",
    icon: <CheckSquare className="h-4 w-4" />,
  },
  {
    title: "Activity and runs",
    detail: "See progress, events, and recent outcomes behind Soma actions.",
    href: "/activity",
    icon: <ListChecks className="h-4 w-4" />,
  },
  {
    title: "Learning and context",
    detail: "Inspect retained patterns, artifacts, and continuity evidence.",
    href: "/memory",
    icon: <Activity className="h-4 w-4" />,
  },
  {
    title: "Tool readiness",
    detail: "Check connected tools, search configuration, and MCP capability status.",
    href: "/resources?tab=tools",
    icon: <Wrench className="h-4 w-4" />,
  },
];

const statePriority: Record<TeamWorkItem["state"], number> = {
  needs_operator: 0,
  degraded: 1,
  running: 2,
  reviewing: 3,
  queued: 4,
  output_ready: 5,
  paused: 6,
  briefed: 7,
  new: 8,
  archived: 9,
};

export function prioritizeSomaHomeWorkItems(items: TeamWorkItem[]) {
  return [...items].sort((left, right) => {
    const leftPriority = itemPriority(left);
    const rightPriority = itemPriority(right);
    if (leftPriority !== rightPriority) return leftPriority - rightPriority;
    const leftTime = left.updatedAt ? Date.parse(left.updatedAt) : 0;
    const rightTime = right.updatedAt ? Date.parse(right.updatedAt) : 0;
    return rightTime - leftTime;
  });
}

function itemPriority(item: TeamWorkItem) {
  const sourcePenalty = item.source === "projection" ? 20 : 0;
  return (item.needsOperator ? -1 : statePriority[item.state]) + sourcePenalty;
}

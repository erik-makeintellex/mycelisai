"use client";

import type React from "react";
import { Activity, CheckSquare, ListChecks, Wrench } from "lucide-react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import { ActiveWorkLane } from "@/components/teams/ActiveWorkLane";
import {
  mergeTeamWorkItems,
  useTeamWorkActionHandler,
} from "@/components/teams/useTeamWorkActionHandler";
import type { ChatMessage, TeamInteraction, TeamWorkItem } from "@/store/useCortexStore";
import { useCortexStore } from "@/store/useCortexStore";
import {
  actionableOutputWorkbenchItems,
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

function somaMessagesNewestFirst(messages: ChatMessage[]) {
  return [...messages]
    .reverse()
    .filter((message) => (
      message.role !== "user"
      && (message.role !== "system" || Boolean(message.execution_summary || message.artifacts?.length))
    ));
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
  const somaMessages = somaMessagesNewestFirst(missionChat);
  const effectiveFocusedTeamId = focusedTeamId || selectedTeamId || null;
  const focusedTeam = effectiveFocusedTeamId
    ? teamsDetail.find((team) => team.id === effectiveFocusedTeamId) ?? null
    : null;
  const outputItems = actionableOutputWorkbenchItems(mergeOutputWorkbenchItems(
    ...somaMessages.map((message) => outputWorkbenchItems(message.execution_summary, message.artifacts)),
  ));
  const projectPackages = somaMessages.flatMap((message) => (
    projectPackageOutputs(message.execution_summary?.outputs)
  ));
  const teamWork = useDurableTeamWork({
    teams: teamsDetail,
    focusedTeamId: effectiveFocusedTeamId,
    refreshVersion: durableWorkRefreshVersion + activeWorkActions.activeWorkRefreshVersion,
  });
  const teamOutputItems = teamOutputWorkbenchItems(teamWork.outputRefs);
  const teamProjectPackages = teamOutputProjectPackages(teamWork.outputRefs);
  const preferFocusedOutputs = Boolean(effectiveFocusedTeamId);
  const mergedOutputItems = preferFocusedOutputs
    ? mergeOutputWorkbenchItems(teamOutputItems, outputItems)
    : mergeOutputWorkbenchItems(outputItems, teamOutputItems);
  const mergedProjectPackages = preferFocusedOutputs
    ? [...teamProjectPackages, ...projectPackages]
    : [...projectPackages, ...teamProjectPackages];
  const activeWorkItems = mergeTeamWorkItems(
    teamWork.items,
    activeWorkActions.submittedTeamWorkItems,
  );
  const somaHomeWorkItems = prioritizeSomaHomeWorkItems(activeWorkItems).filter(
    (item) => item.state !== "archived",
  );
  const attentionWorkCount = somaHomeWorkItems.filter(needsWorkAttention).length;
  const displayedMode = activeMode ?? (focusedTeam ? focusedTeam.name : null);
  const hasWorkContextChoices = Boolean(effectiveFocusedTeamId)
    || activeWorkItems.length > 0
    || teamWork.outputRefs.length > 0;
  const hasWorkReviewContent = Boolean(effectiveFocusedTeamId)
    || somaHomeWorkItems.length > 0
    || teamWork.status === "loading"
    || Boolean(activeWorkActions.activeWorkActionNotice)
    || Boolean(activeWorkActions.activeWorkActionError);
  const hasOutputReviewContent = mergedOutputItems.length > 0 || mergedProjectPackages.length > 0;
  const hasTrustReviewContent = missionChat.length > 0;
  const hasContextReviewContent = Boolean(effectiveFocusedTeamId) || hasWorkContextChoices;

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
        {hasWorkContextChoices ? (
          <SomaTeamContextSwitcher
            teams={teamsDetail}
            workItems={teamWork.items}
            focusedTeamId={effectiveFocusedTeamId}
            onRootSelect={clearFocusedContext}
            onTeamSelect={focusTeamContext}
          />
        ) : null}
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
            hasWorkReviewContent ? (
              <ActiveWorkLane
                title="Work to review"
                items={somaHomeWorkItems}
                emptyMessage={displayedMode && teamWork.items.length === 0
                  ? `Soma is focused on ${displayedMode}. ${teamWork.emptyMessage}`
                  : teamWork.emptyMessage}
                statusLabel={activeWorkActions.activeWorkActionNotice ?? teamWork.statusLabel}
                degradedMessage={activeWorkActions.activeWorkActionError ?? teamWork.degradedMessage}
                onAction={handleActiveWorkAction}
                onTeamAsk={activeWorkActions.handleTeamAsk}
                frame={false}
                maxVisibleItems={effectiveFocusedTeamId ? 6 : 3}
                totalItemCount={activeWorkItems.length}
                moreItemsHref="/teams"
              />
            ) : undefined
          )}
          trust={trustSlot ?? (hasTrustReviewContent ? <SomaCausalSummary messages={missionChat} /> : undefined)}
          output={outputSlot ?? (
            hasOutputReviewContent ? (
              <OutputWorkbench
                outputs={mergedOutputItems}
                projectPackages={mergedProjectPackages}
                emptyMessage={teamWork.status === "loading"
                  ? "Checking for saved Soma and team outputs."
                  : "No finished output is ready yet."}
              />
            ) : undefined
          )}
          context={contextSlot ?? (hasContextReviewContent ? <SomaEvidencePanel items={evidence} compact /> : undefined)}
          primaryPanel={attentionWorkCount > 0 && !hasOutputReviewContent ? "work" : undefined}
          reviewCount={attentionWorkCount > 0 && !hasOutputReviewContent ? attentionWorkCount : undefined}
          showOutputDigest
        />
      </div>
    </section>
  );
}

const defaultEvidence: SomaEvidenceItem[] = [
  {
    title: "Approval queue",
    detail: "Approve or reject actions that need your decision.",
    href: "/approvals",
    icon: <CheckSquare className="h-4 w-4" />,
  },
  {
    title: "Activity and runs",
    detail: "See recent progress and what changed.",
    href: "/activity",
    icon: <ListChecks className="h-4 w-4" />,
  },
  {
    title: "Learning and context",
    detail: "Review saved patterns and useful context.",
    href: "/memory",
    icon: <Activity className="h-4 w-4" />,
  },
  {
    title: "Tool setup",
    detail: "Check connected tools and search setup.",
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

function needsWorkAttention(item: TeamWorkItem) {
  return item.needsOperator || ["needs_operator", "degraded", "running", "reviewing", "queued"].includes(item.state);
}

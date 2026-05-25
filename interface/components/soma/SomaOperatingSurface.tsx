"use client";

import type React from "react";
import { Activity, CheckSquare, ListChecks, Wrench } from "lucide-react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import { ActiveWorkLane } from "@/components/teams/ActiveWorkLane";
import { useTeamWorkActionHandler } from "@/components/teams/useTeamWorkActionHandler";
import type { ChatMessage, TeamWorkItem } from "@/store/useCortexStore";
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
  const activeWorkActions = useTeamWorkActionHandler(selectTeam);
  const evidence = evidenceItems ?? defaultEvidence;
  const latestSoma = lastSomaMessage(missionChat);
  const outputItems = outputWorkbenchItems(latestSoma?.execution_summary, latestSoma?.artifacts);
  const projectPackages = projectPackageOutputs(latestSoma?.execution_summary?.outputs);
  const teamWork = useDurableTeamWork({
    teams: teamsDetail,
    focusedTeamId,
    refreshVersion: durableWorkRefreshVersion + activeWorkActions.activeWorkRefreshVersion,
  });
  const teamOutputItems = teamOutputWorkbenchItems(teamWork.outputRefs);
  const teamProjectPackages = teamOutputProjectPackages(teamWork.outputRefs);
  const mergedOutputItems = mergeOutputWorkbenchItems(outputItems, teamOutputItems);
  const mergedProjectPackages = [...projectPackages, ...teamProjectPackages];
  const somaHomeWorkItems = prioritizeSomaHomeWorkItems(teamWork.items).filter(
    (item) => item.state !== "archived",
  );

  return (
    <section
      className="overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface shadow-[0_18px_40px_rgba(148,163,184,0.16)]"
      data-testid="soma-operating-surface"
    >
      <SomaHeader
        organizationName={organizationName}
        activeMode={activeMode}
        governancePosture={governancePosture}
      />
      <div className="p-4 lg:p-5">
        <SomaWorkspaceFrame
          expression={(
            <div
              data-testid="central-soma-chat-frame"
              className="h-[68vh] min-h-[500px] max-h-[760px] overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg"
            >
              <MissionControlChat
                simpleMode
                autoFocus
                organizationId={organizationId}
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
              onAction={activeWorkActions.handleActiveWorkAction}
              onTeamAsk={activeWorkActions.handleTeamAsk}
              frame={false}
              maxVisibleItems={focusedTeamId ? 6 : 3}
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
          context={contextSlot ?? <SomaEvidencePanel items={evidence} />}
        />
      </div>
    </section>
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

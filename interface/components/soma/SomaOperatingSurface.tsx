"use client";

import type React from "react";
import { Activity, CheckSquare, ListChecks, Wrench } from "lucide-react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import { recoveryReviewQueueItems } from "@/components/recovery/recoveryQueue";
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
import { outputWorkbenchDigest } from "./OutputWorkbenchDigest";
import { SomaActionShelf } from "./SomaActionShelf";
import { SomaCausalSummary } from "./SomaCausalSummary";
import { SomaEvidencePanel, type SomaEvidenceItem } from "./SomaEvidencePanel";
import { SomaOutcomeVaultPanel } from "./SomaOutcomeVaultPanel";
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
  const sendMissionChat = useCortexStore((state) => state.sendMissionChat);
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
  const activeWorkItems = mergeTeamWorkItems(teamWork.items, activeWorkActions.submittedTeamWorkItems);
  const somaHomeWorkItems = recoveryReviewQueueItems(activeWorkItems);
  const attentionWorkCount = somaHomeWorkItems.length;
  const unresolvedWorkReviewCount = somaHomeWorkItems.filter((item) => item.state !== "output_ready").length;
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
  const scopeCopy = organizationName ? `Ready to coordinate work for ${organizationName}` : "Ready to help create or resume an AI Organization";
  const activeWorkNode = activeWorkSlot ?? (
    hasWorkReviewContent ? (
      <ActiveWorkLane
        title="Recovery and review"
        items={somaHomeWorkItems}
        emptyMessage={displayedMode && teamWork.items.length === 0
          ? `Soma is focused on ${displayedMode}. ${teamWork.emptyMessage}`
          : teamWork.emptyMessage}
        statusLabel={activeWorkActions.activeWorkActionNotice ?? teamWork.statusLabel}
        degradedMessage={activeWorkActions.activeWorkActionError ?? teamWork.degradedMessage}
        onAction={handleActiveWorkAction}
        onTeamAsk={activeWorkActions.handleTeamAsk}
        frame={false}
        purpose="review"
        maxVisibleItems={effectiveFocusedTeamId ? 6 : 3}
        totalItemCount={somaHomeWorkItems.length}
        moreItemsHref="/teams?view=work"
      />
    ) : undefined
  );
  const outputNode = outputSlot ?? (
    hasOutputReviewContent ? (
      <OutputWorkbench
        outputs={mergedOutputItems}
        projectPackages={mergedProjectPackages}
        emptyMessage={teamWork.status === "loading"
          ? "Checking for saved Soma and team outputs."
          : "No finished output is ready yet."}
      />
    ) : undefined
  );
  const latestOutputDigest = outputWorkbenchDigest({
    outputs: mergedOutputItems,
    projectPackages: mergedProjectPackages,
  });
  const trustNode = trustSlot ?? (hasTrustReviewContent ? <SomaCausalSummary messages={missionChat} /> : undefined);
  const contextNode = contextSlot ?? (hasContextReviewContent ? <SomaEvidencePanel items={evidence} compact /> : undefined);

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

  async function handleActiveWorkAction(item: TeamWorkItem, action: TeamInteraction) {
    await activeWorkActions.handleActiveWorkAction(item, action);
    if (action.action === "inspect") {
      const teamId = item.teamIds[0] ?? item.id;
      if (typeof window !== "undefined" && window.location.pathname === "/dashboard") {
        window.history.replaceState(null, "", `/dashboard?team_id=${encodeURIComponent(teamId)}`);
      }
    }
  }

  const handlePinnedAction = (prompt: string) => {
    void sendMissionChat(prompt);
  };

  return (
    <section
      className="overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface shadow-[0_18px_40px_rgba(0,0,0,0.18)]"
      data-testid="soma-operating-surface"
    >
      <SomaActionShelf onRunAction={handlePinnedAction} />
      <div className="border-b border-cortex-border px-5 py-5 lg:px-6">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="min-w-0">
            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-bold uppercase tracking-[0.18em] text-cortex-primary">
              Soma
            </div>
            <h1 className="mt-3 text-3xl font-semibold tracking-tight text-cortex-text-main">
              What do you want Soma to do?
            </h1>
            <p className="mt-1 text-base text-cortex-text-muted">{scopeCopy}</p>
          </div>
          <div className="flex flex-wrap items-center gap-2 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
            <span className="font-semibold text-cortex-text-main">Ready</span>
            {displayedMode ? <span>Mode: {displayedMode}</span> : null}
            <span>{governancePosture ?? "Governed execution enabled"}</span>
          </div>
        </div>
      </div>
      <div className="p-4 lg:p-6">
        {hasWorkContextChoices ? (
          <SomaTeamContextSwitcher
            teams={teamsDetail}
            workItems={teamWork.items}
            focusedTeamId={effectiveFocusedTeamId}
            onRootSelect={clearFocusedContext}
            onTeamSelect={focusTeamContext}
          />
        ) : null}
        <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_380px] 2xl:grid-cols-[minmax(0,1fr)_420px]">
          <div className="min-w-0 overflow-hidden rounded-3xl border border-cortex-border bg-cortex-bg shadow-sm">
            <div className="border-b border-cortex-border px-6 py-5">
              <h2 className="text-xl font-semibold tracking-tight text-cortex-text-main">Talk to Soma</h2>
              <p className="mt-1 text-sm text-cortex-text-muted">
                Ask a question, trigger a workflow, review a proposal, or open the result.
              </p>
            </div>
            <div className="p-4 lg:p-5">
              <SomaWorkspaceFrame
                expression={(
                  <div
                    data-testid="central-soma-chat-frame"
                    className="h-[46vh] min-h-[180px] max-h-[520px] overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg"
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
                activeWork={activeWorkNode}
                trust={trustNode}
                output={outputNode}
                context={contextNode}
                primaryPanel={attentionWorkCount > 0 && !hasOutputReviewContent ? "work" : undefined}
                recoveryReviewCount={hasOutputReviewContent ? unresolvedWorkReviewCount : 0}
                reviewCount={attentionWorkCount > 0 && !hasOutputReviewContent ? attentionWorkCount : undefined}
                showOutputDigest
              />
            </div>
          </div>
          <SomaOutcomeVaultPanel
            operationCount={somaHomeWorkItems.length}
            latestOutput={latestOutputDigest}
            recoveryCount={unresolvedWorkReviewCount}
          />
        </div>
      </div>
    </section>
  );
}

const defaultEvidence: SomaEvidenceItem[] = [
  {
    title: "Outcome approvals",
    detail: "Decide what Soma can run next.",
    href: "/approvals",
    icon: <CheckSquare className="h-4 w-4" />,
  },
  {
    title: "Outcome history",
    detail: "Review what happened and what changed.",
    href: "/activity",
    icon: <ListChecks className="h-4 w-4" />,
  },
  {
    title: "Outcome context",
    detail: "Review saved patterns that shape future work.",
    href: "/memory",
    icon: <Activity className="h-4 w-4" />,
  },
  {
    title: "What Soma can use",
    detail: "Check capability readiness and repair paths.",
    href: "/resources?tab=tools",
    icon: <Wrench className="h-4 w-4" />,
  },
];

export function prioritizeSomaHomeWorkItems(items: TeamWorkItem[]) {
  return recoveryReviewQueueItems(items);
}

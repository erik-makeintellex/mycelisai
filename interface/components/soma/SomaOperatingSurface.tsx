"use client";

import type React from "react";
import { useState } from "react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import { ActiveWorkLane } from "@/components/teams/ActiveWorkLane";
import { mergeTeamWorkItems, useTeamWorkActionHandler } from "@/components/teams/useTeamWorkActionHandler";
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
import {
  defaultSomaEvidence,
  outcomeProjectSummaryFromWork,
  prioritizeSomaHomeWorkItems,
  railAlertsFromWorkItems,
} from "./SomaOperatingSurfaceSupport";
import { SomaOutcomeVaultHeaderButton, SomaOutcomeVaultOverlay } from "./SomaOutcomeVaultOverlay";
import { DEFAULT_SOMA_SUGGESTIONS, type SomaSuggestion } from "./SomaSuggestionBar";
import { SomaTeamContextSwitcher } from "./SomaTeamContextSwitcher";
import { SomaWorkspaceFrame } from "./SomaWorkspaceFrame";
import { useDurableTeamWork } from "./useDurableTeamWork";
import { useOutcomeProjectSummary } from "./useOutcomeProjects";

export { prioritizeSomaHomeWorkItems } from "./SomaOperatingSurfaceSupport";

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
  const evidence = evidenceItems ?? defaultSomaEvidence;
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
  const somaHomeWorkItems = prioritizeSomaHomeWorkItems(activeWorkItems);
  const outcomeVaultAlerts = railAlertsFromWorkItems(somaHomeWorkItems);
  const projectedOutcomeProjectSummary = outcomeProjectSummaryFromWork({ teams: teamsDetail, focusedTeamId: effectiveFocusedTeamId, workItems: activeWorkItems, outputRefs: teamWork.outputRefs });
  const durableOutcomeProjectSummary = useOutcomeProjectSummary({
    teams: teamsDetail,
    focusedTeamId: effectiveFocusedTeamId,
    refreshKey: durableWorkRefreshVersion + activeWorkActions.activeWorkRefreshVersion,
  });
  const outcomeProjectSummary = durableOutcomeProjectSummary ?? projectedOutcomeProjectSummary;
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
  const [vaultOpen, setVaultOpen] = useState(false);

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
      className="flex h-full min-h-0 flex-col overflow-hidden rounded-3xl border border-cortex-border bg-cortex-surface shadow-[0_18px_40px_rgba(0,0,0,0.18)]"
      data-testid="soma-operating-surface"
    >
      <SomaActionShelf onRunAction={handlePinnedAction} />
      <div className="border-b border-cortex-border bg-cortex-bg/65 px-4 py-2.5 lg:px-5">
        <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center">
            <div className="inline-flex w-fit items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-[0.16em] text-cortex-primary">
              Soma
            </div>
            <div className="min-w-0">
              <h1 className="truncate text-base font-semibold tracking-tight text-cortex-text-main">
                What do you want Soma to do?
              </h1>
              <p className="truncate text-xs text-cortex-text-muted">{scopeCopy}</p>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2 text-xs text-cortex-text-muted">
            <span className="rounded-full border border-cortex-success/25 bg-cortex-success/10 px-2.5 py-1 font-semibold text-cortex-success">
              Ready
            </span>
            {displayedMode ? (
              <span className="rounded-full border border-cortex-border bg-cortex-surface px-2.5 py-1">
                Mode: {displayedMode}
              </span>
            ) : null}
            <span className="rounded-full border border-cortex-border bg-cortex-surface px-2.5 py-1">
              {governancePosture ?? "Governed execution enabled"}
            </span>
          </div>
        </div>
      </div>
      <div className="min-h-0 flex-1 p-3 lg:p-4">
        {hasWorkContextChoices ? (
          <SomaTeamContextSwitcher
            teams={teamsDetail}
            workItems={teamWork.items}
            focusedTeamId={effectiveFocusedTeamId}
            onRootSelect={clearFocusedContext}
            onTeamSelect={focusTeamContext}
          />
        ) : null}
        <div className="relative">
          <div className="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-2xl border border-cortex-border bg-cortex-bg shadow-sm">
            <div className="flex flex-col gap-2 border-b border-cortex-border px-4 py-2.5 sm:flex-row sm:items-center sm:justify-between">
              <div className="min-w-0">
                <h2 className="text-base font-semibold tracking-tight text-cortex-text-main">Talk to Soma</h2>
                <p className="truncate text-xs text-cortex-text-muted">
                  Ask, approve, inspect proof, or turn an outcome into retained work.
                </p>
              </div>
              <div className="flex shrink-0 flex-wrap items-center gap-2 text-[11px] font-semibold text-cortex-text-muted">
                <span className="rounded-full border border-cortex-border bg-cortex-surface px-2.5 py-1">Threaded</span>
                <span className="rounded-full border border-cortex-border bg-cortex-surface px-2.5 py-1">Governed</span>
                {!vaultOpen ? (
                  <SomaOutcomeVaultHeaderButton
                    attentionCount={attentionWorkCount + (hasOutputReviewContent ? 1 : 0)}
                    onOpen={() => setVaultOpen(true)}
                  />
                ) : null}
              </div>
            </div>
            <div className="min-h-0 flex-1 p-2 lg:p-3">
              <SomaWorkspaceFrame
                expression={(
                  <div
                    data-testid="central-soma-chat-frame"
                    className="h-[52vh] min-h-[300px] overflow-hidden rounded-xl border border-cortex-border bg-cortex-bg lg:h-full lg:min-h-[300px] 2xl:min-h-[500px]"
                  >
                    <MissionControlChat simpleMode autoFocus organizationId={organizationId} focusedTeamId={effectiveFocusedTeamId} suggestions={suggestions} />
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
          <SomaOutcomeVaultOverlay
            open={vaultOpen}
            operationCount={somaHomeWorkItems.length}
            latestOutput={latestOutputDigest}
            projectSummary={outcomeProjectSummary}
            recoveryCount={unresolvedWorkReviewCount}
            alerts={outcomeVaultAlerts}
            onClose={() => setVaultOpen(false)}
          />
        </div>
      </div>
    </section>
  );
}

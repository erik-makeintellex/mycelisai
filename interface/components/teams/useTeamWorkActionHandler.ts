"use client";

import { useCallback, useState } from "react";
import type { TeamInteraction, TeamWorkItem, TeamWorkItemState } from "@/store/useCortexStore";
import { postTeamWorkAction, postTeamWorkAsk, type TeamWorkAskResult } from "./teamWorkActions";

export function useTeamWorkActionHandler(
  selectTeam: (teamId: string | null) => void,
) {
  const [activeWorkRefreshVersion, setActiveWorkRefreshVersion] = useState(0);
  const [activeWorkActionError, setActiveWorkActionError] = useState<string | null>(null);
  const [activeWorkActionNotice, setActiveWorkActionNotice] = useState<string | null>(null);
  const [submittedTeamWorkItems, setSubmittedTeamWorkItems] = useState<TeamWorkItem[]>([]);

  const handleActiveWorkAction = useCallback(
    async (item: TeamWorkItem, action: TeamInteraction) => {
      if (action.action === "inspect") {
        selectTeam(item.teamIds[0] ?? item.id);
        return;
      }
      setActiveWorkActionError(null);
      setActiveWorkActionNotice(null);
      try {
        await postTeamWorkAction(item, action.action, actionSummary(item, action));
        setActiveWorkActionNotice(actionNotice(item, action));
        setActiveWorkRefreshVersion((version) => version + 1);
      } catch (error) {
        setActiveWorkActionError(
          error instanceof Error ? error.message : "Team work action failed.",
        );
      }
    },
    [selectTeam],
  );

  const handleTeamAsk = useCallback(async (item: TeamWorkItem, message: string) => {
    setActiveWorkActionError(null);
    setActiveWorkActionNotice(null);
    try {
      const result = await postTeamWorkAsk(item, message);
      setActiveWorkActionNotice(teamAskNotice(item, result));
      setSubmittedTeamWorkItems((items) =>
        upsertTeamWorkItem(items, submittedTeamAskItem(item, result)),
      );
      if (result.state === "degraded") {
        setActiveWorkActionError(teamAskDegradedMessage(result));
      }
      setActiveWorkRefreshVersion((version) => version + 1);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Team ask failed.";
      setActiveWorkActionError(message);
      throw new Error(message);
    }
  }, []);

  return {
    activeWorkRefreshVersion,
    activeWorkActionError,
    activeWorkActionNotice,
    submittedTeamWorkItems,
    handleActiveWorkAction,
    handleTeamAsk,
  };
}

export function mergeTeamWorkItems(
  durableItems: TeamWorkItem[],
  submittedItems: TeamWorkItem[],
) {
  const durableIds = new Set(durableItems.map((item) => item.id));
  return [
    ...submittedItems.filter((item) => !durableIds.has(item.id)),
    ...durableItems,
  ];
}

function upsertTeamWorkItem(items: TeamWorkItem[], next: TeamWorkItem) {
  const existingIndex = items.findIndex((item) => item.id === next.id);
  if (existingIndex === -1) return [next, ...items];
  return items.map((item, index) => (index === existingIndex ? next : item));
}

function submittedTeamAskItem(sourceItem: TeamWorkItem, result: TeamWorkAskResult): TeamWorkItem {
  const state = submittedTeamAskState(result.state);
  const outputRefs = result.outputRefs ?? [];
  const proofRefs = result.proofRefs ?? [];
  const auditRefs = result.auditRefs ?? [];
  return {
    id: result.workItemId ?? `${sourceItem.id}-submitted-ask`,
    title: result.objective ?? `Continue ${sourceItem.title}`,
    description: teamAskDescription(sourceItem, result),
    state,
    ownerLabel: "Soma",
    scopeLabel: "Delegated work",
    updatedAt: new Date().toISOString(),
    teamIds: [result.teamId ?? sourceItem.teamIds[0] ?? sourceItem.id],
    interactions: [],
    source: "durable",
    sourceLabel: state === "output_ready" ? "Team output ready" : "Queued team ask",
    runId: result.runId,
    outputRefs,
    outputCount: outputRefs.length || undefined,
    proofRefs,
    auditRefs,
    fallbackReason: state === "degraded" ? teamAskDegradedMessage(result) : undefined,
    nextAction: submittedTeamAskNextAction(state, outputRefs.length, result),
    recoveryOptions: state === "degraded"
      ? ["Review degraded delivery and retry from retained context."]
      : undefined,
  };
}

function submittedTeamAskState(state?: string): TeamWorkItemState {
  if (state === "output_ready" || state === "degraded" || state === "running" || state === "queued") {
    return state;
  }
  return "queued";
}

function teamAskNotice(item: TeamWorkItem, result: TeamWorkAskResult) {
  if (result.state === "degraded") {
    return "Team ask recorded with degraded delivery. Review the work item for recovery.";
  }
  if (result.state === "output_ready") {
    return "Team response is ready and retained in Active Work.";
  }
  if (result.accepted || result.state === "queued") {
    return "Team ask queued. You can keep working; output will appear in Active Work.";
  }
  const teamLabel = item.teamIds[0] ?? "team";
  return `Soma sent the ask to ${teamLabel}; Active Work will refresh with the result.`;
}

function teamAskDescription(item: TeamWorkItem, result: TeamWorkAskResult) {
  const parts = [
    result.eventHeadline,
    result.eventDetails,
    result.reply ? `Reply: ${result.reply}` : null,
    !result.eventHeadline && !result.eventDetails ? teamAskNotice(item, result) : null,
  ].filter(Boolean);
  return parts.join(" ");
}

function teamAskDegradedMessage(result: TeamWorkAskResult) {
  const reason = result.degradationState ?? result.dispatchState ?? "delivery degraded";
  return `Team ask was recorded but degraded: ${reason}.`;
}

function submittedTeamAskNextAction(
  state: TeamWorkItemState,
  outputCount: number,
  result: TeamWorkAskResult,
) {
  if (result.eventNextAction) return result.eventNextAction;
  if (state === "output_ready") {
    if (outputCount > 0) {
      return "Review retained output and proof.";
    }
    return "Review the team response and proof.";
  }
  if (state === "queued" || state === "running") {
    return "Wait for team output or degradation proof.";
  }
  if (state === "degraded") {
    return "Review recovery before retrying.";
  }
  return undefined;
}

function actionSummary(item: TeamWorkItem, action: TeamInteraction) {
  if (action.action === "steer") {
    return `Operator requested steering for "${item.title}". Continue from the current objective and ask Soma for clarified guidance if needed.`;
  }
  if (action.action === "recover") {
    return `Operator requested recovery for "${item.title}" using retained context, outputs, proof, and audit refs.`;
  }
  if (action.action === "archive") {
    return `Operator archived "${item.title}" from active review. Retained proof remains inspectable in history.`;
  }
  return undefined;
}

function actionNotice(item: TeamWorkItem, action: TeamInteraction) {
  if (action.action === "archive") {
    return "Cleared from Review Queue. Retained proof remains available in history.";
  }
  if (action.action === "recover") {
    return "Recovery requested. Watch this lane for a new output, proof, or blocker.";
  }
  if (action.action === "pause") {
    return "Work paused. Resume or archive it when ready.";
  }
  if (action.action === "resume" || action.action === "start_work") {
    return "Work queued. Watch this lane for status and retained output.";
  }
  if (action.action === "steer") {
    return "Steering recorded. The team can continue from the retained work state.";
  }
  return `Updated ${item.title}.`;
}

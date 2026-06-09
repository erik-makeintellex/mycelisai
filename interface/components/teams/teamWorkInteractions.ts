import type {
  TeamInteraction,
  TeamWorkItemState,
} from "@/store/useCortexStore";

export function durableInteractions({
  teamId,
  workItemId,
  state,
  runId,
  needsOperator = false,
  executionShape,
}: {
  teamId: string;
  workItemId: string;
  state: TeamWorkItemState;
  runId?: string | null;
  needsOperator?: boolean;
  executionShape?: string | null;
}): TeamInteraction[] {
  const inspectHref = runId ? `/runs/${encodeURIComponent(runId)}` : "/teams?view=work";
  const inspectLabel = runId ? "Open run" : "Open details";
  const isActive = state === "running" || state === "reviewing";
  const canSteer = state !== "archived";
  const canRecover = state === "degraded" || state === "needs_operator";
  const isTeamSetup = executionShape === "create_team";
  const canStart =
    !isTeamSetup && (state === "briefed" || state === "queued" || state === "new");
  const canResume = state === "paused";
  return [
    { action: "inspect", label: inspectLabel, href: inspectHref, audited: true },
    {
      action: "steer",
      label: needsOperator ? "Reply to team" : "Ask for changes",
      disabled: !canSteer,
      disabledReason: canSteer ? undefined : "Archived work cannot be steered.",
      audited: true,
    },
    {
      action: "start_work",
      label: "Start task",
      disabled: !canStart,
      disabledReason: startDisabledReason(canStart, isTeamSetup, workItemId, state),
      audited: true,
    },
    {
      action: "pause",
      label: "Pause",
      disabled: !isActive,
      disabledReason: isActive ? undefined : "Pause is available after work is running.",
      audited: true,
    },
    {
      action: "resume",
      label: "Resume",
      disabled: !canResume,
      disabledReason: canResume ? undefined : "Resume is available for paused work.",
      audited: true,
    },
    {
      action: "recover",
      label: "Retry recovery",
      disabled: !canRecover,
      disabledReason: canRecover
        ? undefined
        : "Recovery is available for degraded or operator-needed work.",
      audited: true,
    },
    {
      action: "archive",
      label: "Clear from review",
      disabled: state === "archived",
      disabledReason: state === "archived" ? "This work item is already archived." : undefined,
      audited: true,
    },
  ];
}

function startDisabledReason(
  canStart: boolean,
  isTeamSetup: boolean,
  workItemId: string,
  state: TeamWorkItemState,
) {
  if (canStart) return undefined;
  if (isTeamSetup) {
    return "Team setup records need a delegated work item before they can start.";
  }
  return `Work item ${workItemId} is already ${state.replace("_", " ")}.`;
}

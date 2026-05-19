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
  const leadHref = `/dashboard?team_id=${encodeURIComponent(teamId)}`;
  const inspectHref = runId ? `/runs/${encodeURIComponent(runId)}` : "/teams";
  const isActive = state === "running" || state === "reviewing";
  const isTeamSetup = executionShape === "create_team";
  const canStart =
    !isTeamSetup && (state === "briefed" || state === "queued" || state === "new");
  const canResume = state === "paused";
  return [
    { action: "inspect", label: "Inspect", href: inspectHref, audited: true },
    { action: "steer", label: needsOperator ? "Respond" : "Steer", href: leadHref, audited: true },
    {
      action: "start_work",
      label: "Start",
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
      action: "archive",
      label: "Archive",
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

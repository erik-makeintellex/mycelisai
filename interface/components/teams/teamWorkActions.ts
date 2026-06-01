import type { TeamInteractionAction, TeamWorkItem } from "@/store/useCortexStore";

type TeamWorkActionResponse = {
  data?: unknown;
  error?: string;
};

export type TeamWorkAskResult = {
  accepted: boolean;
  dispatchState?: string;
  workItemId?: string;
  teamId?: string;
  state?: string;
  objective?: string;
  degradationState?: string;
};

export async function postTeamWorkAction(
  item: TeamWorkItem,
  action: TeamInteractionAction,
  summary?: string,
): Promise<void> {
  const teamId = item.teamIds[0];
  if (!teamId) {
    throw new Error("Team work item is missing its team id.");
  }

  const response = await fetch(
    `/api/v1/teams/${encodeURIComponent(teamId)}/work/${encodeURIComponent(item.id)}/actions`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        action,
        source_kind: "workspace_ui",
        source_channel: "teams.active_work",
        actor_ref: "operator",
        payload_kind: "team_work_action",
        summary,
      }),
    },
  );
  const payload = await response.json().catch(() => ({})) as TeamWorkActionResponse;
  if (!response.ok) {
    throw new Error(payload.error || `Could not ${action.replace("_", " ")} team work.`);
  }
}

export async function postTeamWorkAsk(
  item: TeamWorkItem,
  message: string,
): Promise<TeamWorkAskResult> {
  const teamId = item.teamIds[0];
  if (!teamId) {
    throw new Error("Team work item is missing its team id.");
  }

  const response = await fetch(
    `/api/v1/teams/${encodeURIComponent(teamId)}/work/ask`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        message,
        async: true,
        summary: `Operator asked ${teamId} to continue work on "${item.title}".`,
        actor_ref: "operator",
        timeout_seconds: 60,
        expected_outputs: ["Team response or retained output"],
        expected_proof: ["Team response event or degraded timeout proof"],
        payload: {
          source_work_item_id: item.id,
          source_work_state: item.state,
        },
      }),
    },
  );
  const payload = await response.json().catch(() => ({})) as TeamWorkActionResponse;
  if (!response.ok) {
    throw new Error(payload.error || "Could not ask the team for more work.");
  }
  return normalizeTeamWorkAskResult(payload.data);
}

function normalizeTeamWorkAskResult(data: unknown): TeamWorkAskResult {
  if (!isRecord(data)) {
    return { accepted: false };
  }
  const workItem = isRecord(data.work_item) ? data.work_item : {};
  return {
    accepted: data.accepted === true,
    dispatchState: stringValue(data.dispatch_state),
    workItemId: stringValue(workItem.work_item_id),
    teamId: stringValue(workItem.team_id),
    state: stringValue(workItem.state),
    objective: stringValue(workItem.objective),
    degradationState: stringValue(workItem.degradation_state),
  };
}

function stringValue(value: unknown) {
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

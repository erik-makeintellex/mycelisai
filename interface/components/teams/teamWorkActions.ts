import type { TeamInteractionAction, TeamWorkItem } from "@/store/useCortexStore";

type TeamWorkActionResponse = {
  data?: unknown;
  error?: string;
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

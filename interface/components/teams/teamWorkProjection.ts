import {
  type TeamDetailEntry,
  type TeamInteraction,
  type TeamWorkItem,
  type TeamWorkItemState,
} from "@/store/useCortexStore";

export function projectTeamWorkItem(team: TeamDetailEntry): TeamWorkItem {
  const hasAgents = team.agents.length > 0;
  const hasError = team.agents.some((agent) => agent.status >= 3);
  const hasRunning = team.agents.some((agent) => agent.status === 2);
  const hasOnline = team.agents.some((agent) => agent.status >= 1);
  const hasDeliveries = team.deliveries.length > 0;
  const state: TeamWorkItemState = hasError
    ? "degraded"
    : hasRunning
      ? "running"
      : hasDeliveries
        ? "output_ready"
        : hasOnline
          ? "queued"
          : hasAgents
            ? "queued"
            : "new";
  const leadHref = `/dashboard?team_id=${encodeURIComponent(team.id)}`;
  const interactions: TeamInteraction[] = [
    { action: "inspect", label: "Inspect", audited: true },
    { action: "steer", label: "Steer", href: leadHref, audited: true },
    { action: "start_work", label: "Start", href: leadHref, audited: true },
    {
      action: "pause",
      label: "Pause",
      disabled: true,
      disabledReason: "Team pause is not exposed by the Teams API yet.",
    },
    {
      action: "resume",
      label: "Resume",
      disabled: true,
      disabledReason: "Team resume is not exposed by the Teams API yet.",
    },
    {
      action: "archive",
      label: "Archive",
      disabled: true,
      disabledReason: "Archive temporary collaboration records from Groups.",
    },
  ];
  const models = unique(team.agents.map((agent) => agent.model).filter(Boolean));
  const tools = unique(team.agents.flatMap((agent) => agent.tools));
  const promptCount = team.agents.filter((agent) =>
    Boolean(agent.system_prompt?.trim()),
  ).length;
  return {
    id: team.id,
    title: team.name,
    description:
      team.mission_intent ||
      `${team.name} lead lane with ${team.agents.length} agent${team.agents.length === 1 ? "" : "s"}.`,
    state,
    ownerLabel: `${team.name} lead`,
    scopeLabel: team.type === "mission" ? "Mission team" : "Standing team",
    updatedAt: latestHeartbeat(team),
    outputCount: team.deliveries.length || undefined,
    teamIds: [team.id],
    interactions,
    advanced: {
      inputs: team.inputs,
      deliveries: team.deliveries,
      modelIds: models,
      toolIds: tools,
      promptCount,
    },
  };
}

function latestHeartbeat(team: TeamDetailEntry): string | null {
  const latest = team.agents
    .map((agent) => Date.parse(agent.last_heartbeat))
    .filter(Number.isFinite)
    .sort((a, b) => b - a)[0];
  return typeof latest === "number" ? new Date(latest).toISOString() : null;
}

function unique(values: string[]): string[] {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
}

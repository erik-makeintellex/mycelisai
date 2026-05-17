import {
  type TeamDetailEntry,
  type TeamOutputRef,
  type TeamInteraction,
  type TeamWorkItem,
  type TeamWorkItemState,
} from "@/store/useCortexStore";

type TeamWorkAPIRecord = {
  work_item_id?: unknown;
  team_id?: unknown;
  run_id?: unknown;
  intent_proof_id?: unknown;
  contract_id?: unknown;
  proof_id?: unknown;
  objective?: unknown;
  scope?: unknown;
  owner?: unknown;
  execution_shape?: unknown;
  expected_outputs?: unknown;
  expected_proof?: unknown;
  capability_requirements?: unknown;
  governance_posture?: unknown;
  state?: unknown;
  last_event?: unknown;
  needs_operator?: unknown;
  degradation_state?: unknown;
  recovery_options?: unknown;
  output_refs?: unknown;
  proof_refs?: unknown;
  audit_refs?: unknown;
  created_at?: unknown;
  updated_at?: unknown;
  version?: unknown;
};

type TeamStatusEventAPIRecord = {
  headline?: unknown;
  details?: unknown;
  next_action?: unknown;
};

const durableStates = new Set<TeamWorkItemState>([
  "new",
  "briefed",
  "queued",
  "running",
  "reviewing",
  "paused",
  "output_ready",
  "degraded",
  "needs_operator",
  "archived",
]);

export function projectTeamWorkItem(team: TeamDetailEntry): TeamWorkItem {
  const state: TeamWorkItemState = "degraded";
  const leadHref = `/dashboard?team_id=${encodeURIComponent(team.id)}`;
  const interactions: TeamInteraction[] = [
    { action: "inspect", label: "Inspect", href: "/teams", audited: true },
    { action: "steer", label: "Steer", href: leadHref, audited: true },
    {
      action: "start_work",
      label: "Start",
      href: leadHref,
      audited: true,
    },
    {
      action: "pause",
      label: "Pause",
      disabled: true,
      disabledReason: "Durable work state is not loaded for this projected team.",
    },
    {
      action: "resume",
      label: "Resume",
      disabled: true,
      disabledReason: "Durable work state is not loaded for this projected team.",
    },
    {
      action: "archive",
      label: "Archive",
      disabled: true,
      disabledReason: "Archive requires a durable work item.",
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
      team.mission_intent
        ? `${team.mission_intent} This is an inspectable roster projection until durable team-work state loads.`
        : `${team.name} exists, but durable team-work state is not loaded here yet.`,
    state,
    ownerLabel: `${team.name} lead`,
    scopeLabel: "Projection fallback",
    updatedAt: latestHeartbeat(team),
    teamIds: [team.id],
    interactions,
    source: "projection",
    sourceLabel: "Projection fallback",
    fallbackReason: "Durable TeamWorkItem records were unavailable, so this row is only an inspectable team roster projection.",
    advanced: {
      inputs: team.inputs,
      deliveries: team.deliveries,
      modelIds: models,
      toolIds: tools,
      promptCount,
    },
  };
}

export function mapDurableTeamWorkItem(raw: TeamWorkAPIRecord, team?: TeamDetailEntry): TeamWorkItem | null {
  const workItemId = stringValue(raw.work_item_id);
  const teamId = stringValue(raw.team_id) ?? team?.id;
  const objective = stringValue(raw.objective);
  if (!workItemId || !teamId || !objective) return null;

  const state = teamWorkState(raw.state);
  const outputRefs = outputRefArray(raw.output_refs, teamId, workItemId);
  const lastEvent = objectValue<TeamStatusEventAPIRecord>(raw.last_event);
  const runId = stringValue(raw.run_id);
  const expectedOutputs = stringArray(raw.expected_outputs);
  const expectedProof = stringArray(raw.expected_proof);
  const recoveryOptions = stringArray(raw.recovery_options);
  const proofRefs = stringArray(raw.proof_refs);
  const auditRefs = stringArray(raw.audit_refs);
  const nextAction = stringValue(lastEvent?.next_action);
  const description = [
    stringValue(lastEvent?.headline),
    stringValue(lastEvent?.details),
    stringValue(raw.degradation_state) ? `Degraded: ${stringValue(raw.degradation_state)}` : null,
    nextAction ? `Next: ${nextAction}` : null,
  ].filter(Boolean).join(" ");

  return {
    id: workItemId,
    title: objective,
    description: description || expectedOutputs.map((item) => `Output: ${item}`).join(" "),
    state,
    ownerLabel: stringValue(raw.owner) || (team ? `${team.name} lead` : "Team lead"),
    scopeLabel: executionShapeLabel(stringValue(raw.execution_shape)),
    updatedAt: stringValue(raw.updated_at) ?? stringValue(raw.created_at),
    outputCount: outputRefs.length || undefined,
    teamIds: [teamId],
    interactions: durableInteractions(teamId, workItemId, state, runId, raw.needs_operator === true),
    source: "durable",
    sourceLabel: "Durable team work",
    runId: runId ?? undefined,
    outputRefs,
    proofRefs,
    auditRefs,
    needsOperator: raw.needs_operator === true,
    nextAction: nextAction ?? undefined,
    recoveryOptions,
    advanced: {
      expectedOutputs,
      expectedProof,
      capabilityIds: stringArray(raw.capability_requirements),
      policyRef: stringValue(raw.governance_posture) ?? undefined,
      executionShape: stringValue(raw.execution_shape) ? [stringValue(raw.execution_shape) as string] : [],
    },
  };
}

export function teamOutputRefsFromItems(items: TeamWorkItem[]): TeamOutputRef[] {
  const seen = new Set<string>();
  return items
    .filter((item) => item.source === "durable")
    .flatMap((item) => item.outputRefs ?? [])
    .filter((output) => {
      const key = output.output_id || `${output.team_id}-${output.work_item_id}-${output.label}-${output.storage_ref ?? ""}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    });
}

export function parseTeamWorkAPIItems(payload: unknown): TeamWorkAPIRecord[] {
  const data = objectValue<{ data?: unknown }>(payload)?.data ?? payload;
  return Array.isArray(data) ? data.filter(isRecord) : [];
}

function durableInteractions(
  teamId: string,
  workItemId: string,
  state: TeamWorkItemState,
  runId?: string | null,
  needsOperator = false,
): TeamInteraction[] {
  const leadHref = `/dashboard?team_id=${encodeURIComponent(teamId)}`;
  const inspectHref = runId ? `/runs/${encodeURIComponent(runId)}` : "/teams";
  const isActive = state === "running" || state === "reviewing";
  const canStart = state === "briefed" || state === "queued" || state === "new";
  const canResume = state === "paused";
  return [
    { action: "inspect", label: "Inspect", href: inspectHref, audited: true },
    { action: "steer", label: needsOperator ? "Respond" : "Steer", href: leadHref, audited: true },
    {
      action: "start_work",
      label: "Start",
      href: canStart ? leadHref : undefined,
      disabled: !canStart,
      disabledReason: canStart ? undefined : `Work item ${workItemId} is already ${state.replace("_", " ")}.`,
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
      href: canResume ? leadHref : undefined,
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

function teamWorkState(value: unknown): TeamWorkItemState {
  const state = stringValue(value) as TeamWorkItemState | null;
  return state && durableStates.has(state) ? state : "degraded";
}

function executionShapeLabel(shape?: string | null) {
  if (shape === "create_team") return "Team setup";
  if (shape === "deliverable") return "Deliverable work";
  if (shape === "delegated_work") return "Delegated work";
  return "Durable team work";
}

function outputRefArray(value: unknown, teamId: string, workItemId: string): TeamOutputRef[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item, index) => ({
    output_id: stringValue(item.output_id) ?? `${workItemId}-output-${index}`,
    team_id: stringValue(item.team_id) ?? teamId,
    work_item_id: stringValue(item.work_item_id) ?? workItemId,
    run_id: stringValue(item.run_id) ?? undefined,
    kind: stringValue(item.kind) ?? "file",
    label: stringValue(item.label) ?? "Team output",
    storage_ref: stringValue(item.storage_ref) ?? undefined,
    entrypoint: stringValue(item.entrypoint) ?? undefined,
    validation_ref: stringValue(item.validation_ref) ?? undefined,
    proof_ref: stringValue(item.proof_ref) ?? undefined,
    contract_id: stringValue(item.contract_id) ?? undefined,
    proof_id: stringValue(item.proof_id) ?? undefined,
    audit_refs: stringArray(item.audit_refs),
    created_at: stringValue(item.created_at) ?? undefined,
  }));
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

function stringValue(value: unknown): string | null {
  return typeof value === "string" && value.trim() ? value.trim() : null;
}

function stringArray(value: unknown): string[] {
  return Array.isArray(value)
    ? unique(value.map((item) => typeof item === "string" ? item : "").filter(Boolean))
    : [];
}

function objectValue<T extends object>(value: unknown): T | null {
  return isRecord(value) ? value as T : null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

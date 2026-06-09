import type {
  TeamInteractionAction,
  TeamOutputRef,
  TeamWorkItem,
} from "@/store/useCortexStore";

type TeamWorkActionResponse = {
  data?: unknown;
  error?: string;
};

export type TeamWorkAskResult = {
  accepted: boolean;
  dispatchState?: string;
  workItemId?: string;
  teamId?: string;
  runId?: string;
  state?: string;
  objective?: string;
  degradationState?: string;
  reply?: string;
  eventHeadline?: string;
  eventDetails?: string;
  eventNextAction?: string;
  outputRefs?: TeamOutputRef[];
  proofRefs?: string[];
  auditRefs?: string[];
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
  const event = isRecord(data.event) ? data.event : {};
  const workItemId = stringValue(workItem.work_item_id);
  const teamId = stringValue(workItem.team_id);
  return {
    accepted: data.accepted === true,
    dispatchState: stringValue(data.dispatch_state),
    workItemId,
    teamId,
    runId: stringValue(workItem.run_id) ?? stringValue(data.run_id),
    state: stringValue(workItem.state),
    objective: stringValue(workItem.objective),
    degradationState: stringValue(workItem.degradation_state),
    reply: stringValue(data.reply),
    eventHeadline: stringValue(event.headline),
    eventDetails: stringValue(event.details),
    eventNextAction: stringValue(event.next_action),
    outputRefs: outputRefArray(workItem.output_refs, teamId, workItemId),
    proofRefs: stringArray(workItem.proof_refs),
    auditRefs: stringArray(workItem.audit_refs),
  };
}

function stringValue(value: unknown) {
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}

function stringArray(value: unknown) {
  if (!Array.isArray(value)) return [];
  return unique(
    value
      .map((item) => stringValue(item))
      .filter((item): item is string => Boolean(item)),
  );
}

function outputRefArray(
  value: unknown,
  teamId?: string,
  workItemId?: string,
): TeamOutputRef[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item, index) => ({
    output_id: stringValue(item.output_id) ?? `${workItemId ?? "work"}-output-${index}`,
    team_id: stringValue(item.team_id) ?? teamId ?? "",
    work_item_id: stringValue(item.work_item_id) ?? workItemId ?? "",
    run_id: stringValue(item.run_id),
    kind: stringValue(item.kind) ?? "file",
    label: stringValue(item.label) ?? "Team output",
    storage_ref: stringValue(item.storage_ref),
    entrypoint: stringValue(item.entrypoint),
    validation_ref: stringValue(item.validation_ref),
    proof_ref: stringValue(item.proof_ref),
    contract_id: stringValue(item.contract_id),
    proof_id: stringValue(item.proof_id),
    audit_refs: stringArray(item.audit_refs),
    created_at: stringValue(item.created_at) ?? stringValue(item.updated_at),
  }));
}

function unique(values: string[]) {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

import type {
  CapabilityManifest,
  MCPActivityEntry,
  MCPServerWithTools,
  MissionEvent,
  MissionRun,
  SearchCapabilityStatus,
  TeamWorkItem,
} from "@/store/useCortexStore";
import { compactId, compactStrings } from "./recoveryQueueProjectionUtils";
import {
  projectCapabilityRecoveryItems,
  projectMCPActivityRecoveryItems,
  projectMCPServerRecoveryItems,
  projectRegistryErrorItems,
  projectSearchRecoveryItems,
} from "./recoveryQueueCapabilityProjection";

export {
  projectCapabilityRecoveryItems,
  projectSearchRecoveryItems,
} from "./recoveryQueueCapabilityProjection";

export type RecoveryQueueSource =
  | "team_work"
  | "run"
  | "mission_event"
  | "mcp_server"
  | "mcp_activity"
  | "search"
  | "capability"
  | "registry";

export type RecoveryQueueSeverity =
  | "needs_operator"
  | "failed"
  | "blocked"
  | "degraded";

export interface RecoveryQueueItem {
  id: string;
  source: RecoveryQueueSource;
  severity: RecoveryQueueSeverity;
  title: string;
  detail: string;
  updatedAt?: string | null;
  runId?: string;
  teamIds?: string[];
  agentId?: string;
  actionLabel?: string;
  actionHref?: string;
  evidenceRefs?: string[];
  recoveryOptions?: string[];
}

export interface RecoveryQueueProjectionInput {
  teamWorkItems?: TeamWorkItem[];
  recentRuns?: MissionRun[];
  missionEvents?: MissionEvent[] | null;
  mcpServers?: MCPServerWithTools[];
  mcpActivity?: MCPActivityEntry[];
  mcpServersError?: string | null;
  mcpToolSetsError?: string | null;
  searchCapability?: SearchCapabilityStatus | null;
  searchCapabilityError?: string | null;
  capabilities?: CapabilityManifest[];
  capabilitiesError?: string | null;
  limit?: number;
}

const DEFAULT_LIMIT = 20;
const severityRank: Record<RecoveryQueueSeverity, number> = {
  needs_operator: 0,
  failed: 1,
  blocked: 2,
  degraded: 3,
};

export function projectRecoveryQueue(input: RecoveryQueueProjectionInput): RecoveryQueueItem[] {
  const items = [
    ...projectTeamWorkRecoveryItems(input.teamWorkItems ?? []),
    ...projectRunRecoveryItems(input.recentRuns ?? []),
    ...projectMissionEventRecoveryItems(input.missionEvents ?? []),
    ...projectMCPServerRecoveryItems(input.mcpServers ?? [], input.mcpServersError),
    ...projectMCPActivityRecoveryItems(input.mcpActivity ?? []),
    ...projectSearchRecoveryItems(input.searchCapability, input.searchCapabilityError),
    ...projectCapabilityRecoveryItems(input.capabilities ?? [], input.capabilitiesError),
    ...projectRegistryErrorItems(input),
  ];
  return sortRecoveryQueue(items).slice(0, input.limit ?? DEFAULT_LIMIT);
}

export function projectTeamWorkRecoveryItems(items: TeamWorkItem[]): RecoveryQueueItem[] {
  return items
    .filter((item) => item.state === "degraded" || item.state === "needs_operator" || item.needsOperator)
    .map((item) => {
      const inspect = item.interactions.find((action) => action.href && !action.disabled);
      const severity: RecoveryQueueSeverity =
        item.state === "needs_operator" || item.needsOperator ? "needs_operator" : "degraded";
      return {
        id: `team-work:${item.id}`,
        source: "team_work",
        severity,
        title: item.title,
        detail: firstText([
          item.description,
          item.fallbackReason,
          item.nextAction,
          "Team work needs operator recovery.",
        ]),
        updatedAt: item.updatedAt,
        runId: item.runId,
        teamIds: item.teamIds,
        actionLabel: item.nextAction ?? item.recoveryOptions?.[0] ?? inspect?.label,
        actionHref: inspect?.href ?? (item.runId ? `/runs/${encodeURIComponent(item.runId)}` : undefined),
        evidenceRefs: compactStrings([
          ...(item.proofRefs ?? []),
          ...(item.auditRefs ?? []),
          ...(item.outputRefs ?? []).map((output) => output.output_id),
        ]),
        recoveryOptions: item.recoveryOptions,
      };
    });
}

export function projectRunRecoveryItems(runs: MissionRun[]): RecoveryQueueItem[] {
  return runs
    .filter((run) => run.status === "failed")
    .map((run) => ({
      id: `run:${run.id}`,
      source: "run",
      severity: "failed",
      title: `Run ${compactId(run.id)} failed`,
      detail: textFromMetadata(run.metadata) ?? "The run stopped before it produced a trusted completed result.",
      updatedAt: run.completed_at ?? run.started_at,
      runId: run.id,
      actionLabel: "Open run",
      actionHref: `/runs/${encodeURIComponent(run.id)}`,
      evidenceRefs: compactStrings([run.metadata?.audit_event_id, run.metadata?.proof_id]),
    }));
}

export function projectMissionEventRecoveryItems(events: MissionEvent[]): RecoveryQueueItem[] {
  return events
    .filter(isRecoveryMissionEvent)
    .map((event) => {
      const severity = eventSeverity(event);
      return {
        id: `mission-event:${event.id}`,
        source: "mission_event",
        severity,
        title: missionEventTitle(event),
        detail: payloadText(event, ["operator_summary", "summary", "message", "error", "reason"])
          ?? "Mission event reports a recoverable failure or degraded state.",
        updatedAt: event.emitted_at,
        runId: event.run_id,
        teamIds: compactStrings([event.source_team, stringPayload(event, "team_id")]),
        agentId: event.source_agent,
        actionLabel: "Open run",
        actionHref: `/runs/${encodeURIComponent(event.run_id)}`,
        evidenceRefs: compactStrings([
          event.audit_event_id,
          stringPayload(event, "proof_id"),
          stringPayload(event, "intent_proof_id"),
          stringPayload(event, "contract_id"),
        ]),
      };
    });
}

export function sortRecoveryQueue(items: RecoveryQueueItem[]): RecoveryQueueItem[] {
  return items
    .map((item, index) => ({ item, index }))
    .sort((left, right) => {
      const severity = severityRank[left.item.severity] - severityRank[right.item.severity];
      if (severity !== 0) return severity;
      const time = timestamp(right.item.updatedAt) - timestamp(left.item.updatedAt);
      if (time !== 0) return time;
      return left.index - right.index;
    })
    .map(({ item }) => item);
}

function isRecoveryMissionEvent(event: MissionEvent) {
  return /fail|degrad|block|halt|cancel/i.test(event.event_type)
    || /fail|error|critical|degrad/i.test(event.severity);
}

function eventSeverity(event: MissionEvent): RecoveryQueueSeverity {
  if (/degrad/i.test(`${event.event_type} ${event.severity}`)) return "degraded";
  if (/block|halt/i.test(event.event_type)) return "blocked";
  return "failed";
}

function missionEventTitle(event: MissionEvent) {
  if (event.event_type === "mission.failed") return `Run ${compactId(event.run_id)} failed`;
  if (/degrad/i.test(event.event_type)) return `Run ${compactId(event.run_id)} degraded`;
  return `${event.event_type} needs recovery`;
}

function payloadText(event: MissionEvent, keys: string[]) {
  for (const key of keys) {
    const value = stringPayload(event, key);
    if (value) return value;
  }
  return null;
}

function stringPayload(event: MissionEvent, key: string): string | undefined {
  const value = event.payload?.[key];
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}

function textFromMetadata(metadata?: Record<string, unknown>) {
  if (!metadata) return null;
  for (const key of ["operator_summary", "summary", "message", "error", "reason"]) {
    const value = metadata[key];
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return null;
}

function firstText(values: Array<string | null | undefined>) {
  return values.find((value) => typeof value === "string" && value.trim())?.trim() ?? "";
}

function timestamp(value?: string | null) {
  if (!value) return 0;
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

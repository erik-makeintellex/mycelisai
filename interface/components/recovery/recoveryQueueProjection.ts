import type {
  CapabilityManifest,
  MCPActivityEntry,
  MCPServerWithTools,
  MissionEvent,
  MissionRun,
  SearchCapabilityStatus,
  TeamWorkItem,
} from "@/store/useCortexStore";

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
const CAPABILITIES_HREF = "/resources?tab=tools";

const severityRank: Record<RecoveryQueueSeverity, number> = {
  needs_operator: 0,
  failed: 1,
  blocked: 2,
  degraded: 3,
};

const healthyCapabilityStates = new Set(["available", "connected", "ready", "online"]);
const failedServerStates = new Set(["degraded", "unavailable", "offline", "disconnected", "failed", "error"]);

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

export function projectMCPServerRecoveryItems(
  servers: MCPServerWithTools[],
  mcpServersError?: string | null,
): RecoveryQueueItem[] {
  if (mcpServersError) return [];
  return servers
    .filter((server) => Boolean(server.error) || failedServerStates.has(server.status.toLowerCase()))
    .map((server) => ({
      id: `mcp-server:${server.id}`,
      source: "mcp_server",
      severity: server.status.toLowerCase() === "degraded" ? "degraded" : "blocked",
      title: `${server.name} MCP server needs repair`,
      detail: server.error ?? `Server status is ${server.status}.`,
      updatedAt: server.created_at,
      actionLabel: "Open capabilities",
      actionHref: CAPABILITIES_HREF,
      evidenceRefs: compactStrings([server.command, server.url, ...(server.capability_ids ?? [])]),
    }));
}

export function projectMCPActivityRecoveryItems(entries: MCPActivityEntry[]): RecoveryQueueItem[] {
  return entries
    .filter((entry) => /fail|error|degrad|block|denied|rejected/i.test(entry.state))
    .map((entry) => ({
      id: `mcp-activity:${entry.id}`,
      source: "mcp_activity",
      severity: /degrad/i.test(entry.state) ? "degraded" : "failed",
      title: `${entry.server_name}: ${entry.tool_name} needs recovery`,
      detail: entry.message || entry.summary || "Recent MCP activity reported a blocker.",
      updatedAt: entry.timestamp,
      runId: entry.run_id,
      teamIds: compactStrings([entry.team_id]),
      agentId: entry.agent_id,
      actionLabel: entry.run_id ? "Open run" : "Open capabilities",
      actionHref: entry.run_id ? `/runs/${encodeURIComponent(entry.run_id)}` : CAPABILITIES_HREF,
    }));
}

export function projectSearchRecoveryItems(
  status?: SearchCapabilityStatus | null,
  error?: string | null,
): RecoveryQueueItem[] {
  if (error) {
    return [registryItem("search", "Search capability status unavailable", error)];
  }
  if (!status) return [];
  if (status.enabled && status.configured && !status.blocker) return [];
  return [{
    id: "search:capability",
    source: "search",
    severity: "blocked",
    title: "Soma search needs configuration",
    detail: status.blocker?.message
      ?? status.next_actions?.[0]
      ?? "Search is not enabled or not configured.",
    actionLabel: status.blocker?.next_action ?? status.next_actions?.[0] ?? "Open capabilities",
    actionHref: CAPABILITIES_HREF,
    evidenceRefs: compactStrings([status.provider, status.soma_tool_name]),
    recoveryOptions: compactStrings([status.blocker?.next_action, ...(status.next_actions ?? [])]),
  }];
}

export function projectCapabilityRecoveryItems(
  capabilities: CapabilityManifest[],
  error?: string | null,
): RecoveryQueueItem[] {
  const registryItems = error ? [registryItem("capabilities", "Capability registry unavailable", error)] : [];
  return [
    ...registryItems,
    ...capabilities
      .filter((capability) => capability.review_required || !isCapabilityAvailable(capability))
      .map((capability) => {
        const status = capability.availability_status?.toLowerCase();
        const needsReview = capability.review_required === true;
        return {
          id: `capability:${capability.id}`,
          source: "capability" as const,
          severity: needsReview
            ? "needs_operator" as const
            : status === "degraded"
            ? "degraded" as const
            : "blocked" as const,
          title: `${capability.name} capability needs recovery`,
          detail: capability.fallback_behavior
            ?? capability.description
            ?? `Capability status is ${capability.availability_status ?? "not available"}.`,
          actionLabel: needsReview ? "Review capability" : "Open capabilities",
          actionHref: "/settings/tools",
          evidenceRefs: compactStrings([
            capability.id,
            capability.bound_server_name,
            capability.bound_tool_name,
            capability.provider,
            ...(capability.config_refs ?? []),
            ...(capability.secret_refs ?? []),
          ]),
          recoveryOptions: compactStrings([capability.fallback_behavior]),
        };
      }),
  ];
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

function projectRegistryErrorItems(input: RecoveryQueueProjectionInput): RecoveryQueueItem[] {
  return [
    input.mcpServersError
      ? registryItem("mcp-servers", "MCP registry unavailable", input.mcpServersError)
      : null,
    input.mcpToolSetsError
      ? registryItem("mcp-toolsets", "MCP access layers unavailable", input.mcpToolSetsError)
      : null,
  ].filter(Boolean) as RecoveryQueueItem[];
}

function registryItem(id: string, title: string, detail: string): RecoveryQueueItem {
  return {
    id: `registry:${id}`,
    source: "registry",
    severity: "blocked",
    title,
    detail,
    actionLabel: "Open capabilities",
    actionHref: CAPABILITIES_HREF,
  };
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

function compactStrings(values: unknown[]): string[] {
  return Array.from(new Set(values.filter((value): value is string =>
    typeof value === "string" && value.trim().length > 0,
  ).map((value) => value.trim())));
}

function isCapabilityAvailable(capability: CapabilityManifest): boolean {
  const status = capability.availability_status?.toLowerCase();
  if (!status) return true;
  return healthyCapabilityStates.has(status);
}

function timestamp(value?: string | null) {
  if (!value) return 0;
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function compactId(id: string) {
  return id.length <= 12 ? id : `${id.slice(0, 8)}...`;
}

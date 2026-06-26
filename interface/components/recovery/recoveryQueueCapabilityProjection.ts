import type {
  CapabilityManifest,
  MCPActivityEntry,
  MCPServerWithTools,
  SearchCapabilityStatus,
} from "@/store/useCortexStore";
import type {
  RecoveryQueueItem,
  RecoveryQueueProjectionInput,
} from "./recoveryQueueProjection";
import { compactStrings } from "./recoveryQueueProjectionUtils";

const CAPABILITIES_HREF = "/resources?tab=tools";
const healthyCapabilityStates = new Set(["available", "connected", "ready", "online"]);
const failedServerStates = new Set(["degraded", "unavailable", "offline", "disconnected", "failed", "error"]);

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
  if (!status || (status.enabled && status.configured && !status.blocker)) return [];
  return [{
    id: "search:capability",
    source: "search",
    severity: "blocked",
    title: "Soma search needs configuration",
    detail: status.blocker?.message ?? status.next_actions?.[0] ?? "Search is not enabled or not configured.",
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
          severity: needsReview ? "needs_operator" as const : status === "degraded" ? "degraded" as const : "blocked" as const,
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

export function projectRegistryErrorItems(input: RecoveryQueueProjectionInput): RecoveryQueueItem[] {
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

function isCapabilityAvailable(capability: CapabilityManifest): boolean {
  const status = capability.availability_status?.toLowerCase();
  return !status || healthyCapabilityStates.has(status);
}

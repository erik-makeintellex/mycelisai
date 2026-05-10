import { useMemo } from "react";
import type { MCPActivityEntry, MCPServerWithTools, StreamSignal } from "@/store/useCortexStore";
import type { MCPRecentActivity } from "./MCPServerCard";

export function useMCPRecentActivity(
    mcpServers: MCPServerWithTools[],
    mcpActivity: MCPActivityEntry[],
    streamLogs: StreamSignal[],
): MCPRecentActivity[] {
    return useMemo(() => {
        const serverNames = new Map(mcpServers.map((server) => [server.id, server.name]));
        const persistedActivity = mcpActivity.map((entry) => ({
            id: entry.id,
            serverId: entry.server_id,
            serverName: entry.server_id ? (serverNames.get(entry.server_id) ?? entry.server_name) : entry.server_name,
            toolName: entry.tool_name,
            state: entry.state,
            message: entry.message || entry.summary,
            timestamp: entry.timestamp,
            runId: entry.run_id,
            teamId: entry.team_id,
            agentId: entry.agent_id,
        }));
        const liveActivity = streamLogs
            .filter((signal) => signal.source_kind === "mcp")
            .map((signal, index) => {
                const serverId = typeof signal.payload?.server_id === "string" ? signal.payload.server_id : undefined;
                const toolName = typeof signal.payload?.tool === "string" ? signal.payload.tool : "unknown_tool";
                const state = typeof signal.payload?.state === "string" ? signal.payload.state : "activity";
                const preview = typeof signal.payload?.result_preview === "string"
                    ? signal.payload.result_preview
                    : typeof signal.payload?.error === "string"
                    ? signal.payload.error
                    : signal.message ?? "Agent MCP activity recorded.";
                return {
                    id: `${signal.timestamp ?? "mcp"}-${serverId ?? "server"}-${toolName}-${index}`,
                    serverId,
                    serverName: serverId ? (serverNames.get(serverId) ?? serverId) : "mcp",
                    toolName,
                    state,
                    message: preview,
                    timestamp: signal.timestamp ?? new Date().toISOString(),
                    runId: signal.run_id,
                    teamId: signal.team_id,
                    agentId: signal.agent_id,
                };
            });
        const merged = [...liveActivity, ...persistedActivity];
        const deduped = new Map<string, MCPRecentActivity>();
        for (const activity of merged) {
            const key = `${activity.timestamp}|${activity.serverId ?? activity.serverName}|${activity.toolName}|${activity.state}|${activity.message}`;
            if (!deduped.has(key)) deduped.set(key, activity);
        }
        return Array.from(deduped.values())
            .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
            .slice(0, 12);
    }, [mcpActivity, mcpServers, streamLogs]);
}

export function formatActivityScope(activity: MCPRecentActivity): string {
    const parts: string[] = [];
    if (activity.teamId) parts.push(`Team ${activity.teamId}`);
    if (activity.agentId) parts.push(`Agent ${activity.agentId}`);
    if (activity.runId) parts.push(`Run ${activity.runId}`);
    return parts.join(" · ");
}

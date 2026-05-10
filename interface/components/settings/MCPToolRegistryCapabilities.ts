import type { CapabilityManifest, MCPServerWithTools, MCPTool, SearchCapabilityStatus } from "@/store/useCortexStore";

export function deriveFallbackCapabilities(
    servers: MCPServerWithTools[],
    searchCapability: SearchCapabilityStatus | null,
    searchCapabilityError: string | null,
): CapabilityManifest[] {
    const derived: CapabilityManifest[] = [];
    if (searchCapability || searchCapabilityError) {
        const ready = Boolean(searchCapability?.enabled && searchCapability?.configured);
        derived.push({
            id: "search.web",
            name: "Mycelis Search",
            description: searchCapability?.blocker?.message ?? "Governed search routed through the active Mycelis search provider.",
            source: "builtin",
            category: "research",
            risk: searchCapability?.supports_public_web ? "medium" : "low",
            approval: "optional",
            outputs: ["SearchResult", "ResearchSummary"],
            writes: ["exchange.search.results", "artifacts.research", "run evidence"],
            allowed_roles: ["soma", "research_lead", "reviewer"],
            audit: "required",
            health_check: true,
            availability_status: ready ? "available" : "unavailable",
            fallback_behavior: searchCapability?.blocker?.next_action ?? searchCapabilityError ?? "Return a concrete search capability blocker.",
            provider: searchCapability?.provider ?? "unknown",
            bound_tool_name: searchCapability?.soma_tool_name ?? "web_search",
        });
    }

    for (const server of servers) {
        if (server.tools.length === 0) {
            derived.push(capabilityFromServer(server));
            continue;
        }
        for (const tool of server.tools) {
            derived.push(capabilityFromTool(server, tool));
        }
    }
    return derived;
}

function capabilityFromServer(server: MCPServerWithTools): CapabilityManifest {
    return {
        id: `mcp.${slugifyCapabilityPart(server.name)}`,
        name: server.name,
        description: server.error ?? "MCP server is installed but has not exposed tools yet.",
        source: "mcp",
        category: inferCapabilityCategory(server.name),
        risk: server.error ? "medium" : "low",
        approval: "policy_resolved",
        outputs: ["MCPResult"],
        writes: ["Managed Exchange", "run evidence"],
        allowed_roles: ["soma"],
        audit: "required",
        health_check: true,
        availability_status: server.status === "connected" && !server.error ? "available" : "unavailable",
        fallback_behavior: server.error ?? "Report that the MCP server has no callable tools.",
        server_or_package: server.command ?? server.url ?? server.transport,
        bound_server_id: server.id,
        bound_server_name: server.name,
    };
}

function capabilityFromTool(server: MCPServerWithTools, tool: MCPTool): CapabilityManifest {
    const name = `${server.name}: ${tool.name}`;
    const mutating = isMutatingTool(tool.name);
    return {
        id: tool.capability_id ?? `mcp.${slugifyCapabilityPart(server.name)}.${slugifyCapabilityPart(tool.name)}`,
        name,
        description: tool.description,
        source: "mcp",
        category: inferCapabilityCategory(`${server.name} ${tool.name} ${tool.description ?? ""}`),
        risk: mutating ? "high" : inferToolRisk(tool.name),
        approval: mutating ? "required" : "policy_resolved",
        inputs: Object.keys(tool.input_schema ?? {}),
        outputs: ["MCPResult", "ToolResult"],
        writes: mutating ? ["workspace files", "Managed Exchange", "run evidence", "audit event"] : ["Managed Exchange", "run evidence"],
        allowed_roles: ["soma", "team_agent"],
        audit: "required",
        health_check: true,
        availability_status: server.status === "connected" && !server.error ? "available" : "unavailable",
        fallback_behavior: server.error ?? "Return a capability blocker and keep the run recoverable.",
        server_or_package: server.command ?? server.url ?? server.transport,
        bound_server_id: server.id,
        bound_server_name: server.name,
        bound_tool_id: tool.id,
        bound_tool_name: tool.name,
    };
}

function isMutatingTool(name: string): boolean {
    return /write|delete|remove|update|create|patch|move|rename|execute|run/i.test(name);
}

function inferToolRisk(name: string): "low" | "medium" {
    return /search|fetch|http|web|request|browse/i.test(name) ? "medium" : "low";
}

function inferCapabilityCategory(text: string): string {
    if (/search|fetch|web|browser|http/i.test(text)) return "research";
    if (/file|filesystem|directory|path/i.test(text)) return "filesystem";
    if (/github|git|repo/i.test(text)) return "development";
    if (/memory|vector|knowledge/i.test(text)) return "memory";
    return "tooling";
}

function slugifyCapabilityPart(value: string): string {
    return value.toLowerCase().replace(/[^a-z0-9]+/g, ".").replace(/^\.+|\.+$/g, "") || "capability";
}

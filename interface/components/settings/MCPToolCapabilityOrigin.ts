import type { CapabilityManifest } from "@/store/useCortexStore";

export type CapabilityOrigin = "builtin" | "host" | "mcp" | "connector";

export const CAPABILITY_ORIGINS: Array<{
    id: CapabilityOrigin;
    label: string;
    summary: string;
}> = [
    {
        id: "builtin",
        label: "Built-in runtime",
        summary: "Implemented inside Mycelis. No MCP server is involved.",
    },
    {
        id: "host",
        label: "Host / CLI",
        summary: "Allowlisted commands or scripts exposed by the machine or container running Core.",
    },
    {
        id: "mcp",
        label: "MCP",
        summary: "Tools provided by an installed and configured MCP server.",
    },
    {
        id: "connector",
        label: "Connector",
        summary: "External APIs, plugins, or provider connections that do not use MCP.",
    },
];

export function capabilityOrigin(capability: CapabilityManifest): CapabilityOrigin {
    const source = capability.source.toLowerCase();
    const id = capability.id.toLowerCase();

    if (
        source === "mcp"
        || source.startsWith("mcp_")
        || id.startsWith("mcp:")
        || id.startsWith("mcp.")
        || Boolean(capability.bound_server_id || capability.bound_server_name)
    ) {
        return "mcp";
    }
    if (
        source === "hostcmd_allowlist"
        || source === "local_script"
        || source === "python"
        || id.startsWith("hostcmd:")
    ) {
        return "host";
    }
    if (
        source === "builtin"
        || source === "internal_tool"
        || source === "code_context"
        || source.startsWith("internal_")
        || id.startsWith("code_context")
        || id.startsWith("code-context")
    ) {
        return "builtin";
    }
    return "connector";
}

export function capabilityOriginLabel(capability: CapabilityManifest): string {
    const origin = capabilityOrigin(capability);
    if (origin === "mcp") {
        const server = capability.bound_server_name ?? capability.bound_server_id;
        return server ? `MCP · ${server}` : "MCP server";
    }
    return CAPABILITY_ORIGINS.find((item) => item.id === origin)?.label ?? "Connector";
}

import { describe, expect, it } from "vitest";
import type { CapabilityManifest } from "@/store/useCortexStore";
import { capabilityOrigin, capabilityOriginLabel } from "@/components/settings/MCPToolCapabilityOrigin";

const baseCapability: CapabilityManifest = {
    id: "internal_tool:list_teams",
    name: "List teams",
    source: "internal_tool",
    category: "coordination",
    risk: "low",
    approval: "none",
};

describe("capability origin", () => {
    it("keeps built-in, host command, MCP, and connector provenance distinct", () => {
        const hostCapability = {
            ...baseCapability,
            id: "hostcmd:hostname",
            source: "hostcmd_allowlist",
        };
        const mcpCapability = {
            ...baseCapability,
            id: "mcp:filesystem:read_file",
            source: "mcp",
            bound_server_name: "filesystem",
        };
        const connectorCapability = {
            ...baseCapability,
            id: "openapi:billing",
            source: "external_api",
        };

        expect(capabilityOrigin(baseCapability)).toBe("builtin");
        expect(capabilityOrigin(hostCapability)).toBe("host");
        expect(capabilityOrigin(mcpCapability)).toBe("mcp");
        expect(capabilityOrigin(connectorCapability)).toBe("connector");
        expect(capabilityOriginLabel(hostCapability)).toBe("Host / CLI");
        expect(capabilityOriginLabel(mcpCapability)).toBe("MCP · filesystem");
    });
});

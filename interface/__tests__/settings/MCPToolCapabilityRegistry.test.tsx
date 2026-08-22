import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { CapabilityRegistryPanel } from "@/components/settings/MCPToolCapabilityRegistry";
import { webResearchCapability } from "./MCPToolRegistry.testData";

describe("CapabilityRegistryPanel", () => {
    it("separates host commands from MCP tools in a bounded catalog", () => {
        render(
            <CapabilityRegistryPanel
                capabilities={[
                    {
                        ...webResearchCapability,
                        id: "hostcmd:hostname",
                        name: "Host command: hostname",
                        source: "hostcmd_allowlist",
                    },
                    {
                        ...webResearchCapability,
                        id: "mcp:filesystem:read_file",
                        name: "Read file",
                        source: "mcp",
                        bound_server_name: "filesystem",
                    },
                ]}
                error={null}
                isLoading={false}
                usingFallback={false}
            />,
        );

        const filters = screen.getByLabelText("Capability origin filters");
        expect(screen.getByTestId("capability-catalog-list").className).toContain("overflow-y-auto");
        expect(screen.getAllByText("Host / CLI").length).toBeGreaterThan(0);
        expect(screen.getByText("MCP · filesystem")).toBeDefined();

        fireEvent.click(within(filters).getByRole("button", { name: /Host \/ CLI.*1/i }));
        expect(screen.getByText("Host command: hostname")).toBeDefined();
        expect(screen.queryByText("Read file")).toBeNull();

        fireEvent.click(within(filters).getByRole("button", { name: /^MCP.*1/i }));
        expect(screen.getByText("Read file")).toBeDefined();
        expect(screen.queryByText("Host command: hostname")).toBeNull();
    });
});

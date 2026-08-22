import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { CapabilityManifest } from "@/store/useCortexStore";
import { CapabilityReadinessSummary } from "@/components/settings/MCPToolReadinessSummary";
import { webResearchCapability } from "./MCPToolRegistry.testData";

describe("CapabilityReadinessSummary", () => {
    it("summarizes capability origins without listing the full inventory", () => {
        const readyCapabilities = makeCapabilities("Ready capability", 8, "available");
        const repairCapabilities = makeCapabilities("Repair capability", 4, "degraded");
        const openAccess = vi.fn();
        const openCatalog = vi.fn();

        render(
            <CapabilityReadinessSummary
                capabilities={[...readyCapabilities, ...repairCapabilities]}
                error={null}
                isLoading={false}
                onOpenAccess={openAccess}
                onOpenCatalog={openCatalog}
                usingFallback={false}
            />,
        );

        expect(screen.queryByText(/Ready capability 1/)).toBeNull();
        expect(screen.queryByText(/Repair capability 1/)).toBeNull();
        expect(screen.getByRole("button", { name: "Built-in runtime: 12" })).toBeDefined();
        expect(screen.getByRole("button", { name: "Host / CLI: 0" })).toBeDefined();
        expect(screen.getByRole("button", { name: "MCP: 0" })).toBeDefined();
        expect(screen.getByRole("button", { name: "Connector: 0" })).toBeDefined();
        expect(screen.getByText(/No MCP server is involved/i)).toBeDefined();
        expect(screen.getByText(/machine or container running Core/i)).toBeDefined();
        expect(screen.queryByText(/approval optional/i)).toBeNull();

        fireEvent.click(screen.getByRole("button", { name: /Available to add/i }));
        fireEvent.click(screen.getByRole("button", { name: "Built-in runtime: 12" }));

        expect(openAccess).toHaveBeenCalledTimes(1);
        expect(openCatalog).toHaveBeenCalledTimes(1);
    });
});

function makeCapabilities(label: string, count: number, status: string): CapabilityManifest[] {
    return Array.from({ length: count }, (_, index) => ({
        ...webResearchCapability,
        id: `${label.toLowerCase().replaceAll(" ", "-")}-${index + 1}`,
        name: `${label} ${index + 1}`,
        availability_status: status,
    }));
}

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { CapabilityManifest } from "@/store/useCortexStore";
import { CapabilityReadinessSummary } from "@/components/settings/MCPToolReadinessSummary";
import { webResearchCapability } from "./MCPToolRegistry.testData";

describe("CapabilityReadinessSummary", () => {
    it("keeps the default readiness overview compact when many capabilities exist", () => {
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

        expect(screen.getByText(/Ready capability 1/)).toBeDefined();
        expect(screen.getByText(/Ready capability 3/)).toBeDefined();
        expect(screen.queryByText(/Ready capability 4/)).toBeNull();
        expect(screen.getByText(/Repair capability 1/)).toBeDefined();
        expect(screen.getByText(/Repair capability 3/)).toBeDefined();
        expect(screen.queryByText(/Repair capability 4/)).toBeNull();
        expect(screen.queryByText(/approval optional/i)).toBeNull();

        fireEvent.click(screen.getByRole("button", { name: /Available to add/i }));
        fireEvent.click(screen.getByRole("button", { name: /Ready/i }));

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

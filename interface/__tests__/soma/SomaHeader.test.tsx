import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { SomaHeader } from "@/components/soma/SomaHeader";
import { useCortexStore } from "@/store/useCortexStore";

describe("SomaHeader", () => {
    beforeEach(() => {
        useCortexStore.setState({
            servicesStatus: [],
            servicesStatusFailure: null,
            servicesStatusUpdatedAt: null,
            streamConnectionState: "online",
        });
    });

    it("does not claim Ready or governed execution when Core is unreachable", () => {
        useCortexStore.setState({ servicesStatusFailure: { kind: "unreachable", at: new Date().toISOString() } });

        render(<SomaHeader organizationName="Northstar" />);

        expect(screen.getByTestId("soma-header-status").getAttribute("data-reachability")).toBe("unreachable");
        expect(screen.getByText("Core API unreachable")).toBeDefined();
        expect(screen.getByText(/uv run inv lifecycle\.up --frontend/)).toBeDefined();
        expect(screen.queryByText("Ready")).toBeNull();
        expect(screen.queryByText("Governed execution enabled")).toBeNull();
        expect(screen.getByText(/Soma cannot take requests/)).toBeDefined();
    });

    it("shows a checking state before the first health check", () => {
        render(<SomaHeader />);

        expect(screen.getByText(/Checking Core connection/)).toBeDefined();
        expect(screen.queryByText("Governed execution enabled")).toBeNull();
    });

    it("shows Ready and governed execution once Core reports healthy services", () => {
        useCortexStore.setState({
            servicesStatusUpdatedAt: new Date().toISOString(),
            servicesStatus: [
                { name: "nats", status: "online" },
                { name: "postgres", status: "online" },
                { name: "cognitive", status: "online" },
            ],
        });

        render(<SomaHeader organizationName="Northstar" />);

        expect(screen.getByText("Ready")).toBeDefined();
        expect(screen.getByText("Governed execution enabled")).toBeDefined();
        expect(screen.getByText("Ready to coordinate work for Northstar")).toBeDefined();
    });
});

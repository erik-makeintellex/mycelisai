import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

vi.mock("reactflow", async () => {
    const mock = await import("../mocks/reactflow");
    return mock;
});

import { MockEventSource } from "../setup";
import SystemQuickChecks from "@/components/system/SystemQuickChecks";
import { useCortexStore } from "@/store/useCortexStore";

function resetStore() {
    useCortexStore.setState({
        servicesStatus: [
            { name: "nats", status: "offline" },
            { name: "postgres", status: "online" },
            { name: "reactive", status: "degraded" },
        ],
        isFetchingServicesStatus: false,
        fetchServicesStatus: vi.fn().mockResolvedValue(undefined),
        isStreamConnected: false,
        streamConnectionState: "offline",
    });
}

describe("SystemQuickChecks", () => {
    beforeEach(() => {
        MockEventSource.reset();
        resetStore();
        Object.defineProperty(navigator, "clipboard", {
            value: { writeText: vi.fn() },
            configurable: true,
        });
    });

    it("renders all quick checks and runs service-backed checks", async () => {
        const fetchServicesStatus = vi.fn().mockResolvedValue(undefined);
        useCortexStore.setState({ fetchServicesStatus });

        render(<SystemQuickChecks />);
        expect(screen.getByText("Quick Checks")).toBeDefined();
        expect(screen.getByTestId("quick-check-nats")).toBeDefined();
        expect(screen.getByTestId("quick-check-postgres")).toBeDefined();
        expect(screen.getByTestId("quick-check-sse")).toBeDefined();
        expect(screen.getByTestId("quick-check-triggers")).toBeDefined();
        expect(screen.getByTestId("quick-check-scheduler")).toBeDefined();

        const natsRow = screen.getByTestId("quick-check-nats");
        expect(natsRow.textContent).toContain("Last checked: never");

        fireEvent.click(screen.getByTestId("quick-check-nats-run"));
        await waitFor(() => expect(fetchServicesStatus).toHaveBeenCalledTimes(1));
        await waitFor(() => expect(natsRow.textContent).not.toContain("Last checked: never"));
    });

    it("runs SSE check and marks the row healthy on successful open", async () => {
        render(<SystemQuickChecks />);

        const sseRow = screen.getByTestId("quick-check-sse");
        expect(sseRow.className).toContain("text-cortex-danger");

        fireEvent.click(screen.getByTestId("quick-check-sse-run"));

        await waitFor(() => {
            expect(sseRow.className).toContain("text-cortex-success");
            expect(sseRow.textContent).not.toContain("Last checked: never");
        });
    });

    it("shows the SSE row as connecting while startup is still establishing the stream", () => {
        useCortexStore.setState({
            isStreamConnected: false,
            streamConnectionState: "connecting",
        });

        render(<SystemQuickChecks />);

        const sseRow = screen.getByTestId("quick-check-sse");
        expect(sseRow.className).toContain("text-cortex-warning");
    });

    it("copies diagnostic snippet for a check", () => {
        render(<SystemQuickChecks />);

        fireEvent.click(screen.getByTestId("quick-check-scheduler-copy"));
        const writeText = navigator.clipboard.writeText as ReturnType<typeof vi.fn>;
        expect(writeText).toHaveBeenCalledTimes(1);

        const snippet = writeText.mock.calls[0][0] as string;
        const payload = JSON.parse(snippet);
        expect(payload.check).toBe("scheduler");
        expect(payload.label).toBe("Automation timing");
        expect(payload.status).toBe("degraded");
    });
});
